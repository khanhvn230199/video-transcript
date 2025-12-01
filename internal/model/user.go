package model

import "time"

// User represents a row in the `users` table.
type User struct {
	ID           int64     `db:"id" json:"id"`
	Email        string    `db:"email" json:"email"`
	PasswordHash string    `db:"password_hash" json:"-"`
	Name         *string   `db:"name" json:"name,omitempty"`
	AvatarURL    *string   `db:"avatar_url" json:"avatar_url,omitempty"`
	Role         string    `db:"role" json:"role"`
	Credit       int       `db:"credit" json:"credit"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}
