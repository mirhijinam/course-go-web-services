package user

import (
	"encoding/json"
	"net/http"
	"rwa/internal/models"
	"time"
)

type CurrentResponse struct {
	User struct {
		ID        string    `json:"ID"`
		Email     string    `json:"email"`
		CreatedAt time.Time `json:"createdAt"`
		UpdatedAt time.Time `json:"updatedAt"`
		Username  string    `json:"username"`
		Bio       string    `json:"bio"`
		Image     string    `json:"image"`
		Following bool      `json:"following"`
		Token     string    `json:"token"`
	} `json:"User"`
}

func (resp *CurrentResponse) Fill(u models.User, s models.Session) {
	resp.User.Email = u.Email
	resp.User.CreatedAt = u.CreatedAt
	resp.User.UpdatedAt = u.UpdatedAt
	resp.User.Username = u.Username
	resp.User.Bio = u.Bio
	resp.User.Image = u.Image
	resp.User.Following = u.Following
	resp.User.Token = s.ID
}

func (h *Handler) Current(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	sess, ok := ctx.Value("session").(models.Session)
	if !ok {
		http.Error(w, "current user error 1 "+ErrSessionProblems.Error(), http.StatusInternalServerError)
		return
	}

	u, err := h.Repo.GetByID(ctx, sess.UserID)
	if err != nil {
		http.Error(w, "current user error 2 "+err.Error(), http.StatusUnauthorized)
		return
	}

	var resp CurrentResponse
	resp.Fill(*u, sess)

	response, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "current user error 3 "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
