package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
)

const (
	SearchParamLimit      = "limit"
	SearchParamOffset     = "offset"
	SearchParamQuery      = "query"
	SearchParamOrderField = "order_field"
	SearchParamOrderBy    = "order_by"
)

var errWrongXMLStartElement = errors.New("wrong xml start element")

var AccessToken = "access_token"

var (
	ErrorBadAccessToken = "Bad AccessToken"
)

func SendError(w http.ResponseWriter, code int, errStr string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	errResp := SearchErrorResponse{Error: errStr}

	errRespByte, _ := json.Marshal(errResp)

	_, err := w.Write(errRespByte)
	if err != nil {
		return
	}

}

func SendOK(w http.ResponseWriter, resp any) {
	data, err := json.Marshal(resp)
	if err != nil {
		SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(data); err != nil {
		SendError(w, http.StatusInternalServerError, err.Error())
		return
	}
}

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("AccessToken") != AccessToken {
		SendError(w, http.StatusUnauthorized, ErrorBadAccessToken)
	}

	params := searchRequestParameters(r)

	if ok := validateSearchRequestParameters(params); !ok {
		SendError(w, http.StatusBadRequest, ErrorBadOrderField)
		return
	}

	file, err := os.Open("dataset.xml")
	if err != nil {
		SendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer func() { _ = file.Close() }()

	var users []User

	decoder := xml.NewDecoder(file)
	for {
		candidate, err := extractUserFromXML(decoder)
		if err != nil {
			errEOF := io.EOF
			if errors.Is(err, errEOF) {
				break
			}
			if errors.Is(err, errWrongXMLStartElement) {
				continue
			}
			SendError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if ok := validateUser(candidate, params.Query); ok {
			users = append(users, candidate)
		}
	}

	users = prepareUserList(users, params)

	SendOK(w, users)
}

func searchRequestParameters(r *http.Request) SearchRequest {
	query := strings.ToLower(r.FormValue(SearchParamQuery))

	orderField := strings.ToLower(r.FormValue(SearchParamOrderField))
	if len(orderField) == 0 {
		orderField = "name"
	}

	orderBy, err := strconv.Atoi(r.FormValue(SearchParamOrderBy))
	if err != nil {
		orderBy = OrderByAsIs
	}

	limit, err := strconv.Atoi(r.FormValue(SearchParamLimit))
	if err != nil {
		limit = 0
	}

	offset, err := strconv.Atoi(r.FormValue(SearchParamOffset))
	if err != nil {
		offset = 0
	}

	return SearchRequest{
		Query:      query,
		OrderField: orderField,
		OrderBy:    orderBy,
		Limit:      limit,
		Offset:     offset,
	}
}

func validateSearchRequestParameters(p SearchRequest) bool {
	if !slices.Contains([]string{"id", "age", "name"}, p.OrderField) {
		return false
	}
	return true
}

func extractUserFromXML(decoder *xml.Decoder) (User, error) {
	token, err := decoder.Token()
	if err != nil {
		return User{}, err
	}

	switch se := token.(type) {
	case xml.StartElement:
		if se.Name.Local == "row" {
			var row struct {
				ID        int    `xml:"id"`
				FirstName string `xml:"first_name"`
				LastName  string `xml:"last_name"`
				Age       int    `xml:"age"`
				About     string `xml:"about"`
				Gender    string `xml:"gender"`
			}

			if err := decoder.DecodeElement(&row, &se); err != nil {
				return User{}, err
			}

			candidate := User{
				Id:     row.ID,
				Age:    row.Age,
				Name:   strings.ToLower(row.FirstName + " " + row.LastName),
				About:  strings.ToLower(row.About),
				Gender: strings.ToLower(row.Gender),
			}

			return candidate, nil
		}
	}

	return User{}, errWrongXMLStartElement
}

func validateUser(u User, q string) bool {
	if q == "" {
		return true
	} else if strings.Contains(u.Name, q) || strings.Contains(u.About, q) {
		return true
	}
	return false
}

func prepareUserList(users []User, p SearchRequest) []User {
	if len(users) == 0 {
		return []User{}
	}

	slices.SortFunc(users, func(a, b User) int {
		switch p.OrderField {
		case "id":
			if a.Id > b.Id {
				return p.OrderBy
			} else if a.Id < b.Id {
				return -1 * p.OrderBy
			}
			return 0 * p.OrderBy
		case "age":
			if a.Age > b.Age {
				return 1 * p.OrderBy
			} else if a.Age < b.Age {
				return -1 * p.OrderBy
			}
			return 0 * p.OrderBy
		case "name":
			return strings.Compare(a.Name, b.Name) * p.OrderBy
		}
		return 0
	})

	users = users[p.Offset:]

	limit := p.Limit
	if limit == 0 || limit > len(users) {
		limit = len(users)
	}
	users = users[:limit]

	return users
}
