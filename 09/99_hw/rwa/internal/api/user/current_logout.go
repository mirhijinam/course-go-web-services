package user

import (
	"net/http"
	"rwa/internal/models"
)

func (h *Handler) CurrentLogout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sess, ok := ctx.Value("session").(models.Session)
	if !ok {
		http.Error(w, "logout current user error 1 "+ErrSessionProblems.Error(), http.StatusInternalServerError)
		return
	}

	err := h.Repo.DB.SessionDeleteAll(sess.UserID)
	if err != nil {
		http.Error(w, "logout current user error 1 "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
