package main

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

func (srv *MyApi) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.URL.Path {
	case "/user/profile":
		srv.HandlerProfile(w, req)
	case "/user/create":
		srv.HandlerCreate(w, req)
	default:
		SendError(w, http.StatusNotFound, "unknown method")
		return
	}
}

func (srv *OtherApi) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.URL.Path {
	case "/user/create":
		srv.HandlerCreate(w, req)
	default:
		SendError(w, http.StatusNotFound, "unknown method")
		return
	}
}

func (srv *MyApi) HandlerProfile(w http.ResponseWriter, r *http.Request) {

	var (
		params ProfileParams
		err    error
	)

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

	{

		Login := getQueries.Get("login")
		if Login == "" {
			Login = postQueries["login"]
		}

		params.Login = Login

		if params.Login == "" {
			SendError(w, http.StatusBadRequest, "login must be not empty")
			return
		}

	}

	var (
		res *User
	)

	res, err = srv.Profile(r.Context(), params)
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

func (srv *MyApi) HandlerCreate(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" && r.Method != "" {
		SendError(w, http.StatusNotAcceptable, "bad method")
		return
	}

	if r.Header.Get("X-Auth") != "100500" {
		SendError(w, http.StatusForbidden, "unauthorized")
		return
	}

	var (
		params CreateParams
		err    error
	)

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

	{
		Age := postQueries["age"]

		AgeInt, err := strconv.Atoi(Age)
		if err != nil {
			SendError(w, http.StatusBadRequest, "age must be int")
			return
		}
		params.Age = AgeInt

		if params.Age < 0 {
			SendError(w, http.StatusBadRequest, "age must be >= 0")
			return
		}

		if params.Age > 128 {
			SendError(w, http.StatusBadRequest, "age must be <= 128")
			return
		}
	}

	{
		Login := postQueries["login"]

		params.Login = Login

		if params.Login == "" {
			SendError(w, http.StatusBadRequest, "login must be not empty")
			return
		}

		if len(params.Login) < 10 {
			SendError(w, http.StatusBadRequest, "login len must be >= 10")
			return
		}

	}

	{
		Name := postQueries["full_name"]

		params.Name = Name

	}

	{
		Status := postQueries["status"]

		params.Status = Status

		if params.Status == "" {
			params.Status = "user"
		}

		enumMap := map[string]struct{}{

			"user": {},

			"moderator": {},

			"admin": {},
		}
		if _, ok := enumMap[params.Status]; !ok {
			SendError(w, http.StatusBadRequest, "status must be one of [user moderator admin]")
			return
		}

	}

	var (
		res *NewUser
	)

	res, err = srv.Create(r.Context(), params)
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

func (srv *OtherApi) HandlerCreate(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" && r.Method != "" {
		SendError(w, http.StatusNotAcceptable, "bad method")
		return
	}

	if r.Header.Get("X-Auth") != "100500" {
		SendError(w, http.StatusForbidden, "unauthorized")
		return
	}

	var (
		params OtherCreateParams
		err    error
	)

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

	{
		Class := postQueries["class"]

		params.Class = Class

		if params.Class == "" {
			params.Class = "warrior"
		}

		enumMap := map[string]struct{}{

			"warrior": {},

			"sorcerer": {},

			"rouge": {},
		}
		if _, ok := enumMap[params.Class]; !ok {
			SendError(w, http.StatusBadRequest, "class must be one of [warrior sorcerer rouge]")
			return
		}

	}

	{
		Level := postQueries["level"]

		LevelInt, err := strconv.Atoi(Level)
		if err != nil {
			SendError(w, http.StatusBadRequest, "level must be int")
			return
		}
		params.Level = LevelInt

		if params.Level < 1 {
			SendError(w, http.StatusBadRequest, "level must be >= 1")
			return
		}

		if params.Level > 50 {
			SendError(w, http.StatusBadRequest, "level must be <= 50")
			return
		}
	}

	{
		Name := postQueries["account_name"]

		params.Name = Name

	}

	{
		Username := postQueries["username"]

		params.Username = Username

		if params.Username == "" {
			SendError(w, http.StatusBadRequest, "username must be not empty")
			return
		}

		if len(params.Username) < 3 {
			SendError(w, http.StatusBadRequest, "username len must be >= 3")
			return
		}

	}

	var (
		res *OtherUser
	)

	res, err = srv.Create(r.Context(), params)
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
