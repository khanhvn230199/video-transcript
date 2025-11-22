package model

import "time"

// User represents a user in the system
type User struct {
	ID              string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Email           string    `json:"email" gorm:"uniqueIndex;not null"`
	PasswordHash    string    `json:"-" gorm:"column:password_hash;not null"`
	Name            string    `json:"name" gorm:"not null"`
	AvatarURL       string    `json:"avatar_url" gorm:"column:avatar_url"`
	DefaultLanguage string    `json:"default_language" gorm:"column:default_language;default:'en'"`
	CreditRemaining int       `json:"credit_remaining" gorm:"column:credit_remaining;default:0"`
	CreditUsed      int       `json:"credit_used" gorm:"column:credit_used;default:0"`
	CreatedAt       time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"column:updated_at"`
}

// TableName specifies the table name for User model
func (User) TableName() string {
	return "users"
}

// CreateUserRequest represents the request payload for creating a user
type CreateUserRequest struct {
	Email           string `json:"email" binding:"required,email"`
	Password        string `json:"password" binding:"required,min=6"`
	Name            string `json:"name" binding:"required"`
	AvatarURL       string `json:"avatar_url"`
	DefaultLanguage string `json:"default_language" binding:"oneof=en vi zh ja ko"`
}

// UpdateUserRequest represents the request payload for updating a user
type UpdateUserRequest struct {
	Name            string `json:"name"`
	AvatarURL       string `json:"avatar_url"`
	DefaultLanguage string `json:"default_language" binding:"omitempty,oneof=en vi zh ja ko"`
}

// LoginRequest represents the request payload for user login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}
