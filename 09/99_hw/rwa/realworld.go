package main

import (
	"net/http"
	apiarticle "rwa/internal/api/article"
	"rwa/internal/api/middleware"
	apiuser "rwa/internal/api/user"
	"rwa/internal/db"
	"rwa/internal/repository/article"
	"rwa/internal/repository/session"
	"rwa/internal/repository/user"
)

// сюда писать код

func helloWorld(w http.ResponseWriter, _ *http.Request) {
	w.Write([]byte("Hello world"))
}

func GetApp() http.Handler {
	db := db.New()

	userRepo := user.New(db)
	sessionRepo := session.New(db)
	articleRepo := article.New(db)

	userHandler := apiuser.New(userRepo, sessionRepo)
	articleHandler := apiarticle.New(articleRepo, userRepo)

	publicMux := http.NewServeMux()
	publicMux.HandleFunc("GET /", helloWorld)

	privateMux := http.NewServeMux()

	privateMux.HandleFunc("POST /api/users", userHandler.Register)
	privateMux.HandleFunc("PUT /api/user", userHandler.UpdateCurrent)
	privateMux.HandleFunc("POST /api/users/login", userHandler.Login)
	privateMux.HandleFunc("GET /api/user", userHandler.Current)
	privateMux.HandleFunc("POST /api/user/logout", userHandler.CurrentLogout)

	privateMux.HandleFunc("POST /api/articles", articleHandler.Create)
	privateMux.HandleFunc("GET /api/articles", articleHandler.List)

	mux := http.NewServeMux()

	mux.Handle("/", publicMux)
	mux.Handle("/api/", middleware.Auth(sessionRepo, privateMux))

	return mux
}
