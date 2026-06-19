package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
)

// тут вы пишете код
// обращаю ваше внимание - в этом задании запрещены глобальные переменные

type ResponseError struct {
	Error any `json:"error"`
}

type ResponseOK struct {
	Result any `json:"response"`
}

func ParamsFromReqBody(r *http.Request) (map[string]any, error) {
	params := make(map[string]any)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	defer func() { _ = r.Body.Close() }()
	err = json.Unmarshal(body, &params)
	if err != nil {
		return nil, err
	}
	return params, nil
}

func SendError(w http.ResponseWriter, errMsg string, errCode int) {
	errJSON, _ := json.Marshal(ResponseError{errMsg})
	http.Error(w, string(errJSON), errCode)
}

func SendOK(w http.ResponseWriter, resp any) {
	w.WriteHeader(http.StatusOK)

	respJSON, err := json.Marshal(resp)
	if err != nil {
		SendError(w, "internal error", http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(respJSON); err != nil {
		SendError(w, "internal error", http.StatusInternalServerError)
	}
}

type DbExplorer struct {
	DB     *sql.DB
	Mux    *http.ServeMux
	Tables []string
}

func NewDbExplorer(db *sql.DB) (DbExplorer, error) {
	e := DbExplorer{DB: db}
	e.initMultiplexer()
	e.initTables()
	return e, nil
}

func (e *DbExplorer) initTables() {
	tableRows, err := e.DB.Query("SHOW TABLES")
	if err != nil {
		return
	}
	defer func() { _ = tableRows.Close() }()

	tables := make([]string, 0)
	for tableRows.Next() {
		var t string
		if err := tableRows.Scan(&t); err != nil {
			return
		}
		tables = append(tables, t)
	}

	e.Tables = tables
}

func (e *DbExplorer) initMultiplexer() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", e.TableList)
	mux.HandleFunc("GET /{table}", e.Records)
	mux.HandleFunc("PUT /{table}/", e.RowCreate)
	mux.HandleFunc("GET /{table}/{id}", e.Record)
	mux.HandleFunc("POST /{table}/{id}", e.RecordUpdate)
	mux.HandleFunc("DELETE /{table}/{id}", e.RecordDelete)

	e.Mux = mux
}

func (e *DbExplorer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	e.Mux.ServeHTTP(w, r)
}

func (e *DbExplorer) TableList(w http.ResponseWriter, _ *http.Request) {
	SendOK(w, ResponseOK{
		map[string]any{"tables": e.Tables},
	})
}

func (e *DbExplorer) Records(w http.ResponseWriter, r *http.Request) {
	table := r.PathValue("table")
	if !slices.Contains(e.Tables, table) {
		SendError(w, "unknown table", http.StatusNotFound)
		return
	}

	params := r.URL.Query()

	limit, err := strconv.Atoi(params.Get("limit"))
	if err != nil {
		limit = 5
	}

	offset, err := strconv.Atoi(params.Get("offset"))
	if err != nil {
		offset = 0
	}

	query := fmt.Sprintf("SELECT * FROM %s LIMIT ? OFFSET ?", table)
	rows, err := e.DB.Query(query, limit, offset)
	if err != nil {
		SendError(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer func() { _ = rows.Close() }()

	columns, err := rows.Columns()
	if err != nil {
		SendError(w, "internal error", http.StatusInternalServerError)
		return
	}

	values := make([]any, len(columns))
	for i := range values {
		values[i] = new(any)
	}

	records := make([]map[string]any, 0)
	for rows.Next() {
		if err = rows.Scan(values...); err != nil {
			SendError(w, "internal error", http.StatusInternalServerError)
			return
		}

		record := make(map[string]any)
		for i, col := range columns {
			x := *(values[i].(*any))
			switch v := x.(type) {
			case []byte:
				record[col] = string(v)
			case nil:
				record[col] = nil
			default:
				record[col] = v
			}
		}

		records = append(records, record)
	}

	SendOK(w, ResponseOK{map[string]any{
		"records": records,
	}})
}

type columnInfo struct {
	name       string
	typ        string
	collation  sql.NullString
	null       string
	key        string
	defaultVal sql.NullString
	extra      string
	privileges string
	comment    string
}

func (info columnInfo) nullVal() any {
	if info.null == "NO" {
		switch {
		case strings.Contains(info.typ, "varchar") || strings.Contains(info.typ, "text"):
			return ""
		case strings.Contains(info.typ, "int"):
			return 0
		}
	}
	return nil
}

type ColumnsInfo map[string]columnInfo

func (c ColumnsInfo) pk() columnInfo {
	for _, v := range c {
		if v.key == "PRI" {
			return v
		}
	}
	return columnInfo{}
}

func (ci ColumnsInfo) autoIncColumns() ColumnsInfo {
	autoIncCols := ColumnsInfo{}
	for _, v := range ci {
		if v.extra == "auto_increment" {
			autoIncCols[v.name] = v
		}
	}
	return autoIncCols
}

func (e *DbExplorer) AllColumnsInfo(table string) (ColumnsInfo, error) {
	queryColumnsInfo := fmt.Sprintf("SHOW FULL COLUMNS FROM `%s`", table)
	rows, err := e.DB.Query(queryColumnsInfo)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	cols := make(map[string]columnInfo)

	colInfo := columnInfo{}
	for rows.Next() {
		if err = rows.Scan(
			&colInfo.name,
			&colInfo.typ,
			&colInfo.collation,
			&colInfo.null,
			&colInfo.key,
			&colInfo.defaultVal,
			&colInfo.extra,
			&colInfo.privileges,
			&colInfo.comment,
		); err != nil {
			return nil, err
		}
		cols[colInfo.name] = colInfo
	}

	return cols, nil
}

func (e *DbExplorer) Record(w http.ResponseWriter, r *http.Request) {
	table := r.PathValue("table")
	if !slices.Contains(e.Tables, table) {
		SendError(w, "unknown table", http.StatusNotFound)
		return
	}

	id := r.PathValue("id")

	cols, err := e.AllColumnsInfo(table)
	if err != nil {
		SendError(w, "internal error", http.StatusInternalServerError)
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE %s = ?", table, cols.pk().name)
	row, err := e.DB.Query(query, id)
	if err != nil {
		SendError(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer func() { _ = row.Close() }()

	columns, err := row.Columns()
	if err != nil {
		SendError(w, "internal error", http.StatusInternalServerError)
		return
	}

	values := make([]any, len(columns))
	for i := range values {
		values[i] = new(any)
	}

	var record map[string]any
	for row.Next() {
		if err = row.Scan(values...); err != nil {
			SendError(w, "internal error", http.StatusInternalServerError)
			return
		}
		record = make(map[string]any)
		for i, col := range columns {
			x := *(values[i].(*any))
			switch v := x.(type) {
			case []byte:
				record[col] = string(v)
			case nil:
				record[col] = nil
			default:
				record[col] = v
			}
		}
	}
	defer func() { _ = row.Close() }()

	if len(record) == 0 {
		SendError(w, "record not found", http.StatusNotFound)
		return
	}

	SendOK(w, ResponseOK{map[string]any{
		"record": record,
	}})
}

func (e *DbExplorer) RowCreate(w http.ResponseWriter, r *http.Request) {
	table := r.PathValue("table")
	if !slices.Contains(e.Tables, table) {
		SendError(w, "unknown table", http.StatusNotFound)
		return
	}

	colsInfo, err := e.AllColumnsInfo(table)
	if err != nil {
		SendError(w, "internal error", http.StatusInternalServerError)
		return
	}

	params, err := ParamsFromReqBody(r)
	if err != nil {
		SendError(w, "internal error", http.StatusInternalServerError)
		return
	}

	colsToCreate := make([]string, 0)
	placeholdersForVals := make([]string, 0)
	valsToCreate := make([]interface{}, 0)

	// Добавление колонок, пришедших в теле запроса.
	for colName, val := range params {
		_, isAutoInc := colsInfo.autoIncColumns()[colName]
		if isAutoInc {
			continue
		}

		_, isExist := colsInfo[colName]
		if !isExist {
			continue
		}

		colsToCreate = append(colsToCreate, colName)
		valsToCreate = append(valsToCreate, val)
		placeholdersForVals = append(placeholdersForVals, "?")
	}

	// Добавление колонок, не пришедших в теле запроса.
	for name, info := range colsInfo {
		if _, ok := params[name]; !ok && !info.defaultVal.Valid {
			colsToCreate = append(colsToCreate, name)
			placeholdersForVals = append(placeholdersForVals, "?")
			valsToCreate = append(valsToCreate, info.nullVal())
		}
	}

	if len(colsToCreate) == 0 {
		SendError(w, "no valid columns to insert", http.StatusBadRequest)
		return
	}

	columnsString := strings.Join(colsToCreate, ", ")
	placeholdersString := strings.Join(placeholdersForVals, ", ")

	query := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s)", table, columnsString, placeholdersString)
	res, err := e.DB.Exec(query, valsToCreate...)
	if err != nil {
		fmt.Println("err", err)
		fmt.Println(columnsString)
		fmt.Println(placeholdersString)
		fmt.Printf("%#v\n", valsToCreate)

		SendError(w, "internal error", http.StatusInternalServerError)
		return
	}

	id, err := res.LastInsertId()
	if err != nil {
		SendError(w, "internal error", http.StatusInternalServerError)
		return
	}

	SendOK(w, ResponseOK{map[string]any{
		colsInfo.pk().name: id,
	}})
}

func (e *DbExplorer) RecordUpdate(w http.ResponseWriter, r *http.Request) {
	table := r.PathValue("table")
	if !slices.Contains(e.Tables, table) {
		SendError(w, "unknown table", http.StatusNotFound)
		return
	}

	id := r.PathValue("id")

	params, err := ParamsFromReqBody(r)
	if err != nil {
		SendError(w, "internal error", http.StatusInternalServerError)
		return
	}

	colsInfo, err := e.AllColumnsInfo(table)
	if err != nil {
		SendError(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Валидация значений колонок, пришедших в теле запроса.
	for _, info := range colsInfo {
		if _, ok := params[info.name]; ok {
			switch params[info.name].(type) {
			case string:
				if !strings.Contains(info.typ, "varchar") && !strings.Contains(info.typ, "text") {
					SendError(w, fmt.Sprintf("field %s have invalid type", info.name), http.StatusBadRequest)
					return
				}
			case float64, int:
				if !strings.Contains(info.typ, "int") {
					SendError(w, fmt.Sprintf("field %s have invalid type", info.name), http.StatusBadRequest)
					return
				}
			case nil:
				if info.null == "NO" {
					SendError(w, fmt.Sprintf("field %s have invalid type", info.name), http.StatusBadRequest)
					return
				}
			}
		}
	}

	// В теле запроса не должно быть PK.
	if _, ok := params[colsInfo.pk().name]; ok {
		SendError(w, fmt.Sprintf("field %s have invalid type", colsInfo.pk().name), http.StatusBadRequest)
		return
	}

	updateStringBuilder := strings.Builder{}
	for col, val := range params {
		var valString string
		switch val.(type) {
		case string:
			valString = fmt.Sprintf("'%s'", val)
		case int, float64:
			valString = fmt.Sprintf("%v", val)
		case nil:
			valString = "null"
		default:
			valString = "unexpected_value"
		}
		if _, ok := colsInfo.autoIncColumns()[col]; ok {
			valString = "null"
		}
		updateStringBuilder.WriteString(fmt.Sprintf("%s = %s, ", col, valString))
	}

	updateString := updateStringBuilder.String()
	updateString = updateString[:len(updateString)-2]

	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s = ?", table, updateString, colsInfo.pk().name)
	res, err := e.DB.Exec(query, id)
	if err != nil {
		SendError(w, "internal error", http.StatusInternalServerError)
		return
	}

	resCnt, err := res.RowsAffected()
	if err != nil {
		SendError(w, "internal error", http.StatusInternalServerError)
		return
	}

	if resCnt == 0 {
		SendError(w, "record not found", http.StatusNotFound)
		return
	}

	SendOK(w, ResponseOK{
		map[string]any{"updated": resCnt},
	})
}

func (e *DbExplorer) RecordDelete(w http.ResponseWriter, r *http.Request) {
	table := r.PathValue("table")
	if !slices.Contains(e.Tables, table) {
		SendError(w, "unknown table", http.StatusNotFound)
		return
	}

	id := r.PathValue("id")

	colsInfo, err := e.AllColumnsInfo(table)
	if err != nil {
		SendError(w, "internal error", http.StatusInternalServerError)
		return
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE %s = ?", table, colsInfo.pk().name)
	res, err := e.DB.Exec(query, id)
	if err != nil {
		SendError(w, "internal error", http.StatusInternalServerError)
		return
	}

	resCnt, err := res.RowsAffected()
	if err != nil {
		SendError(w, "internal error", http.StatusInternalServerError)
		return
	}

	SendOK(w, ResponseOK{
		map[string]any{"deleted": resCnt},
	})

}
