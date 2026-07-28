package models

import "time"

type User struct {
	ID        int64     `db:"id" json:"id"`
	Username  string    `db:"username" json:"username"`
	Email     *string   `db:"email" json:"email,omitempty"`
	Password  string    `db:"password" json:"-"`
	Role      string    `db:"role" json:"role"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}


type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}


type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email,omitempty"`
	Password string `json:"password"`
}


type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
