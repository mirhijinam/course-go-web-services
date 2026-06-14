package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/template"
)

const (
	genPrefix = `// apigen:api `
)

var ErrUnsupportedMethod = fmt.Errorf("unsupported method")

var regexpAPIValidator = regexp.MustCompile("`apivalidator:\"([^\"]+)\"`")

var (
	serveHTTPMethodTpl = template.Must(template.New("serveHTTPMethodTpl").Parse(`
{{ range $key, $val := .Receivers }} 
func ({{ $val }} *{{ $key }}) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.URL.Path { 
{{ range $method := $.Methods }}
{{- if eq $method.RecvType.Name $key -}}
	case "{{ $method.HS.URL }}":
		{{ $method.RecvName }}.Handler{{ $method.Name }}(w, req) 
{{ end -}}
{{ end -}}
	default:
		SendError(w, http.StatusNotFound, "unknown method")
		return
	}
}
{{ end }}
`))

	handlersTpl = template.Must(template.New("handlersTpl").Parse(`
{{ range $method := .Methods }}
func ({{ .RecvName }} *{{ .RecvType }}) Handler{{ .Name }}(w http.ResponseWriter, r *http.Request) { 
{{ if ne .HS.MethodType "" }} 
	if r.Method != "{{ .HS.MethodType }}" && r.Method != "" {
		SendError(w, http.StatusNotAcceptable, "bad method")
		return
	}
{{ end }}

{{ if eq .HS.Auth true }}
	if r.Header.Get("X-Auth") != "100500" {
		SendError(w, http.StatusForbidden, "unauthorized")
		return
	}
{{ end }}

	var (
		params {{ .ParamsStructType }}
		err error
	)

{{ if eq $method.HS.MethodType "GET" }}
	getQueries := r.URL.Query()
{{ end }}

{{ if eq $method.HS.MethodType "POST" }}
	body, _ := io.ReadAll(r.Body)

	postQueries := make(map[string]string)
	tmpQueries := strings.Split(string(body), "&")
	for _, v := range tmpQueries {
		if v == "" {
			continue
		}
		keyValue := strings.Split(v, "=")
		postQueries[keyValue[0]] = keyValue[1]
	}
{{ end }}

{{ if eq $method.HS.MethodType "" }}
	getQueries := r.URL.Query()

	body, _ := io.ReadAll(r.Body)

	postQueries := make(map[string]string)
	tmpQueries := strings.Split(string(body), "&")
	for _, v := range tmpQueries {
		if v == "" {
			continue
		}
		keyValue := strings.Split(v, "=")
		postQueries[keyValue[0]] = keyValue[1]
	}
{{ end }}

{{ range $key, $val := .ParamsStructFields }} 
	{
	{{ if (eq $method.HS.MethodType "GET") }}
		{{ $key.Name }} := getQueries.Get("{{ $val.ParamName }}") 
	{{ end }}

	{{- if (eq $method.HS.MethodType "POST") }}
		{{- $key.Name }} := postQueries["{{ $val.ParamName }}"]
	{{ end }}

	{{ if (eq $method.HS.MethodType "") }}
		{{ $key.Name }} := getQueries.Get("{{ $val.ParamName }}") 
		if {{ $key.Name }} == "" {
			{{ $key.Name }} = postQueries["{{ $val.ParamName }}"]
		}
	{{ end }}

	{{ if eq $key.Type "string" }}
		params.{{ $key.Name }} = {{ $key.Name }}
	{{ end }}

	{{ if eq $key.Type "int" }}
		{{ $key.Name }}Int, err := strconv.Atoi({{ $key.Name }})
		if err != nil {
			SendError(w, http.StatusBadRequest, "{{ $val.ParamName }} must be int")
			return
		}
		params.{{ $key.Name }} = {{ $key.Name }}Int
	{{ end }}

	{{ if eq $val.Required true }}
		if params.{{ $key.Name }} == "" {
			SendError(w, http.StatusBadRequest, "{{ $val.ParamName }} must be not empty")
			return
		}
	{{ end }}
       

	{{ if ne $val.DefaultValue nil }} 
		if params.{{ $key.Name }} == "" {
			params.{{ $key.Name }} = "{{ $val.DefaultValue }}"
		}
	{{ end }}

	{{ if gt (len $val.Enum) 0 }}
		enumMap := map[string]struct{}{
		{{ range $val.Enum }}
			"{{ . }}": {}, 
		{{ end }}
		}
		if _, ok := enumMap[params.{{ $key.Name }}]; !ok {
			SendError(w, http.StatusBadRequest, "{{ $val.ParamName }} must be one of {{ $val.Enum }}")
			return
		}
	{{ end }}

	{{ if eq $key.Type "string" }}
	{{ if ne $val.Min nil }}
		if len(params.{{ $key.Name }}) < {{ $val.Min }} {
			SendError(w, http.StatusBadRequest, "{{ $val.ParamName }} len must be >= {{ $val.Min }}")
			return
		}
	{{ end }}
	{{ if ne $val.Max nil }}
		if len(params.{{ $key.Name }}) > {{ $val.Max }} {
			SendError(w, http.StatusBadRequest, "{{ $val.ParamName }} len must be <= {{ $val.Max }}")
			return
		}
	{{ end }}
	{{ end }}

	{{ if eq $key.Type "int" }}
	{{ if ne $val.Min nil }}
		if params.{{ $key.Name }} < {{ $val.Min }} {
			SendError(w, http.StatusBadRequest, "{{ $val.ParamName }} must be >= {{ $val.Min }}")
			return
		}
	{{ end }}
	{{ if ne $val.Max nil }}
		if params.{{ $key.Name }} > {{ $val.Max }} {
			SendError(w, http.StatusBadRequest, "{{ $val.ParamName }} must be <= {{ $val.Max }}")
			return
		}
	{{- end -}}
	{{- end -}}
	}
{{ end }}

	var (
		res {{if eq .RetTypeStar true }}*{{ end }}{{ .RetType }}
	)

	res, err = {{ .RecvName }}.{{ .Name }}(r.Context(), params)
	if err != nil {
		apiErr := ApiError{}
		if errors.As(err, &apiErr) {
			SendError(w, apiErr.HTTPStatus, apiErr.Err.Error())
		} else {
			SendError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	SendOK(w, res)
}
{{ end }}`))
)

type tpl struct {
	Receivers map[string]string
	Methods   []Method
}

type Method struct {
	Name               string
	RecvName           string
	RecvType           *ast.Ident
	RecvTypeStar       bool // unused.
	RetType            *ast.Ident
	RetTypeStar        bool
	ParamsStructType   *ast.Ident
	ParamsStructFields map[Field]FieldConditions
	HS                 HandlerSettings
}

type HandlerSettings struct {
	URL        string `json:"url"`
	Auth       bool   `json:"auth"`
	MethodType string `json:"method"`
}

type Field struct {
	Name string
	Type string
}

func (f *Field) Fill(lf *ast.Field) {
	f.Name = lf.Names[0].Name
	f.Type = lf.Type.(*ast.Ident).Name
}

type FieldConditions struct {
	ParamName    string
	Required     bool
	Enum         []any
	DefaultValue any
	Min          *int
	Max          *int
}

func (c *FieldConditions) Fill(lf *ast.Field) {
	conditionsRaw := strings.Split(regexpAPIValidator.FindStringSubmatch(lf.Tag.Value)[1], ",")

	for _, cond := range conditionsRaw {
		condKeyVal := strings.Split(cond, "=")
		switch condKeyVal[0] {
		case "required":
			c.Required = true
		case "paramname":
			c.ParamName = condKeyVal[1]
		case "enum":
			enum := strings.Split(condKeyVal[1], "|")
			if lf.Type.(*ast.Ident).Name == "string" {
				for _, v := range enum {
					c.Enum = append(c.Enum, v)
				}
			} else {
				for _, v := range enum {
					vInt, _ := strconv.Atoi(v)
					c.Enum = append(c.Enum, vInt)
				}
			}
		case "default":
			c.DefaultValue = condKeyVal[1]
		case "min":
			condVal, _ := strconv.Atoi(condKeyVal[1])
			c.Min = &condVal
		case "max":
			condVal, _ := strconv.Atoi(condKeyVal[1])
			c.Max = &condVal
		}
	}

	defaultParamName := strings.ToLower(lf.Names[0].Name)
	if c.ParamName == "" {
		c.ParamName = defaultParamName
	}
}

func ParseMethod(f *ast.FuncDecl) (Method, error) {
	recvName, recvType, _, err := receiverType(f)
	if err != nil {
		return Method{}, err
	}

	retType, retTypeIsStar, err := returnValueType(f)
	if err != nil {
		return Method{}, err
	}

	hs, err := handlerSettings(f)
	if err != nil {
		return Method{}, err
	}

	paramsStructType, paramsStructFieldList, err := paramsStruct(f)
	if err != nil {
		return Method{}, err
	}

	paramsStructFieldConditionsMap := make(map[Field]FieldConditions)
	for _, astField := range paramsStructFieldList {
		if astField == nil {
			continue
		}
		var (
			apiValidation FieldConditions
			field         Field
		)

		apiValidation.Fill(astField)
		field.Fill(astField)

		paramsStructFieldConditionsMap[field] = apiValidation
	}

	fmt.Printf("%s:\n\ttype: %T\n\tdata: %+v\n\n", f.Name.Name, f, f)
	fmt.Printf("\trecvType:\n\t\ttype: %T\n\t\tdata: %+v\n\n", recvType, recvType)
	fmt.Printf("\ths:\n\t\ttype: %T\n\t\tdata: %+v\n\n", hs, hs)
	fmt.Printf("\tparamsStructType:\n\t\ttype: %T\n\t\tdata: %+v\n\n", paramsStructType, paramsStructType)
	fmt.Printf("\tparamsStructFieldList:\n")
	for i, astField := range paramsStructFieldList {
		fmt.Printf("\t\tastField %d:\n\t\t\ttype: %T\n\t\t\tdata: %+v\n\n", i, *astField, *astField)
	}

	return Method{
		Name:               f.Name.Name,
		RecvName:           recvName,
		RecvType:           recvType,
		RetType:            retType,
		RetTypeStar:        retTypeIsStar,
		HS:                 hs,
		ParamsStructType:   paramsStructType,
		ParamsStructFields: paramsStructFieldConditionsMap,
	}, nil
}

func returnValueType(f *ast.FuncDecl) (retType *ast.Ident, isStar bool, err error) {
	if f.Type == nil || f.Type.Results == nil || len(f.Type.Results.List) == 0 || f.Recv.List[0] == nil || f.Type.Results.List[0].Type == nil {
		return nil, false, fmt.Errorf("failed to extract return value type")
	}

	ret := f.Type.Results.List[0].Type

	starExpr, ok := ret.(*ast.StarExpr)
	if ok {
		retType, ok = starExpr.X.(*ast.Ident)
		if !ok {
			return nil, false, fmt.Errorf("failed to extract pointer receiver type")
		}
		isStar = true
	} else {
		retType, ok = ret.(*ast.Ident)
		if !ok {
			return nil, false, fmt.Errorf("failed to extract value receiver type")
		}
	}

	return retType, isStar, nil
}

func receiverType(f *ast.FuncDecl) (recvName string, recvType *ast.Ident, isStar bool, err error) {
	if f.Recv == nil || len(f.Recv.List) == 0 || f.Recv.List[0] == nil {
		return "", nil, false, fmt.Errorf("failed to extract receiver type")
	}

	recv := f.Recv.List[0]

	if len(recv.Names) == 0 {
		return "", nil, false, fmt.Errorf("failed to extract receiver name")
	}

	recvName = recv.Names[0].Name

	starExpr, ok := recv.Type.(*ast.StarExpr)
	if ok {
		recvType, ok = starExpr.X.(*ast.Ident)
		if !ok {
			return "", nil, false, fmt.Errorf("failed to extract pointer receiver type")
		}
		isStar = true
	} else {
		recvType, ok = recv.Type.(*ast.Ident)
		if !ok {
			return "", nil, false, fmt.Errorf("failed to extract value receiver type")
		}
	}

	return recvName, recvType, isStar, nil
}

func handlerSettings(f *ast.FuncDecl) (hs HandlerSettings, err error) {
	if f.Doc == nil && len(f.Doc.List) == 0 {
		return HandlerSettings{}, fmt.Errorf("failed to extract content of '%s' tag", genPrefix)
	}

	var tagValue string
	for _, c := range f.Doc.List {
		if c == nil || !strings.HasPrefix(c.Text, genPrefix) {
			continue
		}

		tagValue = strings.TrimPrefix(c.Text, genPrefix)

		if err := json.Unmarshal([]byte(tagValue), &hs); err != nil {
			return HandlerSettings{}, fmt.Errorf("failed to unmarshal content of '%s' tag", genPrefix)
		}
	}

	if len(tagValue) == 0 {
		return HandlerSettings{}, fmt.Errorf("failed to find '%s' tag", genPrefix)
	}

	if hs.URL == "" {
		return HandlerSettings{}, fmt.Errorf("failed to find url from content of '%s' tag", genPrefix)
	}

	return hs, nil
}

func paramsStruct(f *ast.FuncDecl) (structType *ast.Ident, structFields []*ast.Field, err error) {
	if f.Type == nil || f.Type.Params == nil || len(f.Type.Params.List) < 2 || f.Type.Params.List[1] == nil {
		return nil, nil, fmt.Errorf("failed to extract params argument")
	}

	ps := f.Type.Params.List[1]
	ident, ok := ps.Type.(*ast.Ident)
	if !ok {
		return nil, nil, fmt.Errorf("failed to extract params argument type")
	}

	typeSpec, ok := ident.Obj.Decl.(*ast.TypeSpec)
	if !ok {
		return nil, nil, fmt.Errorf("failed to extract specification about params argument type")
	}

	structType = typeSpec.Name

	structTypeSpec, ok := typeSpec.Type.(*ast.StructType)
	if !ok {
		return nil, nil, fmt.Errorf("params argument type must be structure")
	}

	if structTypeSpec.Fields == nil {
		return nil, nil, fmt.Errorf("failed to extract params structure fields")
	}

	structFields = structTypeSpec.Fields.List

	return structType, structFields, nil
}

func main() {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, os.Args[1], nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	out, err := os.Create(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = out.Close() }()

	var buf bytes.Buffer

	fprintln(&buf, fmt.Sprintf(`package %v

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
)

var (
	_ = strconv.Itoa
	_ = strings.Split
	_ = io.ReadAll
)
`, node.Name.Name))

	RecvFuncDecls := make([]*ast.FuncDecl, 0)

	for _, n := range node.Decls {
		funcDecl, ok := n.(*ast.FuncDecl)
		if !ok || funcDecl.Doc == nil {
			continue
		}

		if funcDecl.Recv != nil {
			RecvFuncDecls = append(RecvFuncDecls, funcDecl)
		}
	}

	methods := make([]Method, 0)
	for _, funcDecl := range RecvFuncDecls {
		m, err := ParseMethod(funcDecl)
		if err != nil {
			switch {
			case errors.As(err, &ErrUnsupportedMethod):
				continue
			default:
				log.Fatal(err)
			}
		}
		methods = append(methods, m)
	}

	receivers := make(map[string]string)
	for _, m := range methods {
		receivers[m.RecvType.Name] = m.RecvName
	}

	err = serveHTTPMethodTpl.Execute(&buf, tpl{receivers, methods})
	if err != nil {
		log.Fatal(err)
	}

	err = handlersTpl.Execute(&buf, tpl{receivers, methods})
	if err != nil {
		log.Fatal(err)
	}

	fprintln(&buf, `
func SendError(w http.ResponseWriter, code int, errStr string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	data, _ := json.Marshal(CR{"error": errStr})
	if _, err := w.Write(data); err != nil {
		return
	}
}

func SendOK(w http.ResponseWriter, resp any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	data, _ := json.Marshal(CR{"error": "", "response": resp})
	if _, err := w.Write(data); err != nil {
		SendError(w, http.StatusInternalServerError, err.Error())
		return
	}
}
`)

	src, err := format.Source(buf.Bytes())
	if err != nil {
		log.Fatal(err)
	}

	if _, err := out.Write(src); err != nil {
		log.Fatal(err)
	}
}

func fprintln(w io.Writer, data ...any) {
	if _, err := fmt.Fprintln(w, data...); err != nil {
		log.Fatal(err)
	}
}
