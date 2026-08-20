package user

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"rwa/internal/models"
	"rwa/pkg/customhash"
	"time"
)

type LoginRequest struct {
	User struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	} `json:"user"`
}

type LoginResponse struct {
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

func (resp *LoginResponse) Fill(u models.User, s models.Session) {
	resp.User.Email = u.Email
	resp.User.CreatedAt = u.CreatedAt
	resp.User.UpdatedAt = u.UpdatedAt
	resp.User.Username = u.Username
	resp.User.Bio = u.Bio
	resp.User.Image = u.Image
	resp.User.Following = u.Following
	resp.User.Token = s.ID
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "login error 1 "+err.Error(), http.StatusBadRequest)
		return
	}

	u, err := h.Repo.GetByEmail(ctx, req.User.Email)
	if err != nil {
		http.Error(w, "login error 2 "+err.Error(), http.StatusInternalServerError)
		return
	}

	salt := u.Password[:8]
	reqPassHashed := customhash.Hash(req.User.Password, salt)

	reqPassHashedDecoded, _ := base64.RawStdEncoding.DecodeString(reqPassHashed)
	correctPassHashedDecoded, _ := base64.RawStdEncoding.DecodeString(u.Password)

	if !bytes.Equal(reqPassHashedDecoded, correctPassHashedDecoded) {
		http.Error(w, "login error 3 "+ErrIncorrectPassword.Error(), http.StatusUnauthorized)
		return
	}

	sess, err := h.Session.Create(u.ID)
	if err != nil {
		http.Error(w, "login error 4 "+err.Error(), http.StatusInternalServerError)
		return
	}

	var ur RegisterResponse
	ur.Fill(*u)
	ur.User.Token = sess.ID

	resp, err := json.Marshal(ur)
	if err != nil {
		http.Error(w, "login error 5 "+err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}
