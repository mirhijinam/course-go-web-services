package db

import (
	"encoding/json"
	"rwa/internal/models"
	"rwa/pkg/customhash"
	"slices"
	"time"
)

func (db *DB) ArticleCreate(a models.Article) (*models.Article, error) {
	a.Slug = customhash.Salt(8)
	a.CreatedAt = time.Now()
	a.UpdatedAt = time.Now()

	d, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}

	db.articles[a.Slug] = d

	return new(a), nil
}

func (db *DB) ArticleList(p models.ArticleListParams) ([]models.Article, error) {
	res := make([]models.Article, 0)
	for _, x := range db.articles {
		var a models.Article
		err := json.Unmarshal(x, &a)
		if err != nil {
			return nil, err
		}
		if p.Author != nil && a.Author.Username != *p.Author {
			continue
		}
		if p.Tag != nil && !slices.Contains(a.TagList, *p.Tag) {
			continue
		}

		res = append(res, a)
	}

	return res, nil
}
