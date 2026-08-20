package db

import (
	"encoding/json"
	"fmt"
	"rwa/internal/models"
	"time"
)

func (db *DB) UserCreate(u models.User) (*models.User, error) {
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()

	d, err := json.Marshal(u)
	if err != nil {
		return nil, err
	}

	db.profiles[u.ID] = d

	return new(u), nil
}

func (db *DB) UserUpdate(u models.User) (*models.User, error) {
	u.UpdatedAt = time.Now()

	d, err := json.Marshal(u)
	if err != nil {
		return nil, err
	}

	db.profiles[u.ID] = d

	return new(u), nil
}

func (db *DB) UserGetByID(id string) (*models.User, error) {
	row, ok := db.profiles[id]
	if !ok {
		return nil, fmt.Errorf("user not found")
	}

	var u models.User
	err := json.Unmarshal(row, &u)
	if err != nil {
		return nil, err
	}

	return new(u), nil
}

func (db *DB) UserGetByEmail(email string) (*models.User, error) {
	for _, row := range db.profiles {
		var u models.User
		err := json.Unmarshal(row, &u)
		if err != nil {
			return nil, err
		}
		if u.Email == email {
			return new(u), nil
		}
	}

	return nil, fmt.Errorf("user not found")
}

func (db *DB) UserGetByUsername(username string) (*models.User, error) {
	for _, row := range db.profiles {
		var u models.User
		err := json.Unmarshal(row, &u)
		if err != nil {
			return nil, err
		}
		if u.Username == username {
			return new(u), nil
		}
	}

	return nil, fmt.Errorf("user not found")
}
