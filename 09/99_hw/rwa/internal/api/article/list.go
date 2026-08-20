package article

import (
	"encoding/json"
	"net/http"
	"rwa/internal/models"
)

type ListResponse struct {
	Articles      []models.Article `json:"articles"`
	ArticlesCount int              `json:"articlesCount"`
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	queryParams := r.URL.Query()

	var p models.ArticleListParams
	if len(queryParams) != 0 {
		if len(queryParams["author"]) > 0 {
			p.Author = new(queryParams["author"][0])
		}
		if len(queryParams["tag"]) > 0 {
			p.Tag = new(queryParams["tag"][0])
		}
	}

	articles, err := h.Repo.List(ctx, p)
	if err != nil {
		http.Error(w, "list articles error 2 "+err.Error(), http.StatusInternalServerError)
	}

	resp := ListResponse{
		Articles:      articles,
		ArticlesCount: len(articles),
	}

	response, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "list articles error 3 "+err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
