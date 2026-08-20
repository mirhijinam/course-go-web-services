package article

import (
	"context"
	"rwa/internal/db"
	"rwa/internal/models"
	"slices"
)

type Repo struct {
	DB db.DB
}

func New(db db.DB) Repo {
	return Repo{
		DB: db,
	}
}

func (r *Repo) Create(ctx context.Context, title, desc, body string, tags []string, author models.User) (*models.Article, error) {
	draw := models.Article{
		Title:       title,
		Description: desc,
		Body:        body,
		TagList:     tags,
		Author:      author,
	}

	a, err := r.DB.ArticleCreate(draw)
	if err != nil {
		return nil, err
	}

	return a, nil
}

func (r *Repo) List(ctx context.Context, p models.ArticleListParams) ([]models.Article, error) {
	articles, err := r.DB.ArticleList(p)
	if err != nil {
		return nil, err
	}
	slices.SortFunc(articles, func(x, y models.Article) int {
		if x.CreatedAt.Before(y.CreatedAt) {
			return -1
		} else if y.CreatedAt.Before(x.CreatedAt) {
			return 1
		}
		return 0
	})

	return articles, nil
}
