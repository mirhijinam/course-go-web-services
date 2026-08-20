package middleware

import (
	"context"
	"net/http"
	"rwa/internal/repository/session"
	"strings"
)

var routesNoAuth = []string{
	"POST /api/users",
	"POST /api/users/login",
	"GET /api/articles",
}

func Auth(sr session.Repo, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, x := range routesNoAuth {
			parts := strings.Split(x, " ")
			if strings.Contains(r.Method, parts[0]) && strings.Contains(r.URL.String(), parts[1]) {
				next.ServeHTTP(w, r)
				return
			}
		}

		authHeader := r.Header.Get("Authorization")
		token := strings.TrimPrefix(authHeader, "Token ")

		s, err := sr.Check(token)
		if err != nil {
			http.Error(w, "failed to check token", http.StatusUnauthorized)
			return
		}
		if s == nil {
			http.Error(w, "wrong token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "session", *s)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
