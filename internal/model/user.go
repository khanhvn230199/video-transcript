package model

import "time"

// User represents a row in the `users` table.
type User struct {
	ID           int64      `db:"id" json:"id"`
	Email        string     `db:"email" json:"email"`
	PasswordHash string     `db:"password_hash" json:"-"`
	Name         *string    `db:"name" json:"name,omitempty"`
	AvatarURL    *string    `db:"avatar_url" json:"avatar_url,omitempty"`
	Gender       *string    `db:"gender" json:"gender,omitempty"`
	DOB          *time.Time `db:"dob" json:"dob,omitempty"`
	Phone        *string    `db:"phone" json:"phone,omitempty"`
	Address      *string    `db:"address" json:"address,omitempty"`
	Role         string     `db:"role" json:"role"`
	Credit       int        `db:"credit" json:"credit"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`
}

// CreateUserRequest represents the request payload for creating a user.
type CreateUserRequest struct {
	Email     string     `json:"email" binding:"required"`
	Password  string     `json:"password" binding:"required"`
	Name      *string    `json:"name"`
	AvatarURL *string    `json:"avatar_url"`
	Gender    *string    `json:"gender"`
	DOB       *time.Time `json:"dob"`
	Phone     *string    `json:"phone"`
	Address   *string    `json:"address"`
	Role      string     `json:"role"`
	Credit    int        `json:"credit"`
}

// UpdateUserRequest represents the request payload for updating a user.
type UpdateUserRequest struct {
	Email     string     `json:"email"`
	Password  string     `json:"password"`
	Name      *string    `json:"name"`
	AvatarURL *string    `json:"avatar_url"`
	Gender    *string    `json:"gender"`
	DOB       *time.Time `json:"dob"`
	Phone     *string    `json:"phone"`
	Address   *string    `json:"address"`
	Role      string     `json:"role"`
	Credit    int        `json:"credit"`
}
