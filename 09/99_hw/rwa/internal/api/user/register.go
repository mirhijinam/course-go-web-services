package user

import (
	"encoding/json"
	"net/http"
	"rwa/internal/models"
	"time"
)

type RegisterRequest struct {
	User struct {
		Email    string `json:"email"`
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"user"`
}

type RegisterResponse struct {
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

func (resp *RegisterResponse) Fill(u models.User) {
	resp.User.Email = u.Email
	resp.User.CreatedAt = u.CreatedAt
	resp.User.UpdatedAt = u.UpdatedAt
	resp.User.Username = u.Username
	resp.User.Bio = u.Bio
	resp.User.Image = u.Image
	resp.User.Following = u.Following
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "register error 1 "+err.Error(), http.StatusBadRequest)
		return
	}

	u, err := h.Repo.Create(ctx, req.User.Email, req.User.Username, req.User.Password)
	if err != nil {
		http.Error(w, "register error 2", http.StatusInternalServerError)
		return
	}

	_, err = h.Session.Create(u.ID)
	if err != nil {
		http.Error(w, "register error 3", http.StatusInternalServerError)
		return
	}

	var resp RegisterResponse
	resp.Fill(*u)

	response, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "register error 4 "+err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(response)
}
