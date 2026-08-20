package article

import (
	"encoding/json"
	"net/http"
	"rwa/internal/models"
)

type CreateRequest struct {
	Article struct {
		Title       string   `json:"title"`
		Description string   `json:"description"`
		Body        string   `json:"body"`
		TagList     []string `json:"tagList"`
	} `json:"article"`
}

type CreateResponse struct {
	Article models.Article `json:"article"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	sess, ok := ctx.Value("session").(models.Session)
	if !ok {
		http.Error(w, "create article error 1 "+ErrSessionProblems.Error(), http.StatusInternalServerError)
		return
	}

	u, err := h.UserRepo.GetByID(ctx, sess.UserID)
	if err != nil {
		http.Error(w, "create article error 2 "+err.Error(), http.StatusUnauthorized)
		return
	}

	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "create article error 3 "+err.Error(), http.StatusBadRequest)
		return
	}

	a, err := h.Repo.Create(ctx, req.Article.Title, req.Article.Description, req.Article.Body, req.Article.TagList, *u)
	if err != nil {
		http.Error(w, "create article error 4 "+err.Error(), http.StatusBadRequest)
		return
	}

	// For tests.
	a.Author.Email = ""

	resp := CreateResponse{
		Article: *a,
	}

	response, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "create article error 5 "+err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(response)

}
