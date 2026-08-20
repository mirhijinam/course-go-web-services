package user

import (
	"encoding/json"
	"net/http"
	"rwa/internal/models"
)

type CurrentUpdateRequest struct {
	User struct {
		Email    *string `json:"email"`
		Bio      *string `json:"bio"`
		Image    *string `json:"image"`
		Username *string `json:"username"`
	} `json:"user"`
}

func (h *Handler) UpdateCurrent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	sess, ok := ctx.Value("session").(models.Session)
	if !ok {
		http.Error(w, "update current user error 1 "+ErrSessionProblems.Error(), http.StatusInternalServerError)
		return
	}

	u, err := h.Repo.GetByID(ctx, sess.UserID)
	if err != nil {
		http.Error(w, "update current user error 2 "+err.Error(), http.StatusUnauthorized)
		return
	}

	var req CurrentUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "update current user error 3 "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.User.Bio == nil && req.User.Username == nil && req.User.Email == nil && req.User.Image == nil {
		http.Error(w, "update current user error 4", http.StatusBadRequest)
		return
	}

	userWithUpdatedFields := *u
	if req.User.Bio != nil {
		userWithUpdatedFields.Bio = *req.User.Bio
	}
	if req.User.Username != nil {
		userWithUpdatedFields.Username = *req.User.Username
	}
	if req.User.Email != nil {
		userWithUpdatedFields.Email = *req.User.Email
	}
	if req.User.Image != nil {
		userWithUpdatedFields.Image = *req.User.Email
	}

	updatedUser, err := h.Repo.Update(ctx, userWithUpdatedFields)
	if err != nil {
		http.Error(w, "update current user error 5", http.StatusBadRequest)
		return
	}

	var resp CurrentResponse
	resp.Fill(*updatedUser, sess)

	response, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "update current user error 6", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
