package user

import (
	"context"
	"rwa/internal/db"
	"rwa/internal/models"
	"rwa/pkg/customhash"

	"github.com/google/uuid"
)

type Repo struct {
	DB db.DB
}

func New(db db.DB) Repo {
	return Repo{
		DB: db,
	}
}

func (r *Repo) Create(ctx context.Context, email, username, password string) (*models.User, error) {
	salt := customhash.Salt(8)
	hashedPass := customhash.Hash(password, salt)

	draw := models.User{
		ID:       uuid.New().String(),
		Email:    email,
		Username: username,
		Password: string(hashedPass),
	}

	u, err := r.DB.UserCreate(draw)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (r *Repo) Update(ctx context.Context, u models.User) (*models.User, error) {
	return r.DB.UserUpdate(u)
}

func (r *Repo) GetByID(ctx context.Context, id string) (*models.User, error) {
	return r.DB.UserGetByID(id)
}

func (r *Repo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return r.DB.UserGetByEmail(email)
}

func (r *Repo) GetByUsername(uname string) (*models.User, error) {
	return r.DB.UserGetByUsername(uname)
}
