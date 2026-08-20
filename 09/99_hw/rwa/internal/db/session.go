package db

import (
	"encoding/json"
	"fmt"
	"rwa/internal/models"
	"time"
)

func (db *DB) SessionCreate(s models.Session) (*models.Session, error) {
	d, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}

	db.sessions[s.ID] = d

	return new(s), nil
}

func (db *DB) SessionCheck(id string) (*models.Session, error) {
	row, ok := db.sessions[id]
	if !ok {
		return nil, fmt.Errorf("session not found")
	}

	var s models.Session
	err := json.Unmarshal(row, &s)
	if err != nil {
		return nil, err
	}

	if time.Now().After(s.ExpiresAt) {
		return nil, nil
	}

	return new(s), nil
}

func (db *DB) SessionDelete(id string) error {
	delete(db.sessions, id)
	return nil
}

func (db *DB) SessionDeleteAll(uid string) error {
	for _, row := range db.sessions {
		var s models.Session
		err := json.Unmarshal(row, &s)
		if err != nil {
			return err
		}

		if uid == s.UserID {
			delete(db.sessions, s.ID)
		}
	}

	return nil
}
