package models

import "time"

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	Bio       string    `json:"bio"`
	Image     string    `json:"image"`
	Following bool      `json:"following"`
}
