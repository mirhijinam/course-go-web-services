package session

import (
	"rwa/internal/db"
	"rwa/internal/models"
	"time"

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

func (r *Repo) Create(uid string) (*models.Session, error) {
	draw := models.Session{
		ID:        uuid.New().String(),
		UserID:    uid,
		ExpiresAt: time.Now().Add(models.SessionLifetime),
	}

	s, err := r.DB.SessionCreate(draw)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (r *Repo) Check(id string) (*models.Session, error) {
	return r.DB.SessionCheck(id)
}

func (r *Repo) Delete(id string) error {
	return r.DB.SessionDelete(id)
}

func (r *Repo) DeleteAll(uid string) error {
	return r.DB.SessionDeleteAll(uid)
}
