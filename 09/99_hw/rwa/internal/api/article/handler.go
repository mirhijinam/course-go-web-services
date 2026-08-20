package article

import (
	"rwa/internal/repository/article"
	"rwa/internal/repository/user"
)

type Handler struct {
	Repo     article.Repo
	UserRepo user.Repo
}

func New(repo article.Repo, userRepo user.Repo) Handler {
	return Handler{
		Repo:     repo,
		UserRepo: userRepo,
	}
}
