package user

import (
	"rwa/internal/repository/session"
	"rwa/internal/repository/user"
)

type Handler struct {
	Repo    user.Repo
	Session session.Repo
}

func New(repo user.Repo, session session.Repo) Handler {
	return Handler{
		Repo:    repo,
		Session: session,
	}
}
