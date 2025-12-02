package service

import (
	"context"
	"database/sql"
	"errors"

	"golang.org/x/crypto/bcrypt"

	"video-transcript/internal/model"
	"video-transcript/internal/repository"
)

// UserService defines business logic for users.
type UserService interface {
	Create(ctx context.Context, u *model.User) error
	Authenticate(ctx context.Context, email, password string) (*model.User, error)
	GetByID(ctx context.Context, id int64) (*model.User, error)
	List(ctx context.Context) ([]*model.User, error)
	Update(ctx context.Context, u *model.User) error
	Delete(ctx context.Context, id int64) error

	GetByEmail(ctx context.Context, email string) (*model.User, error)
}

type userService struct {
	repo repository.UserRepository
}

// NewUserService creates a new UserService.
func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) Create(ctx context.Context, u *model.User) error {
	// place for validations, business rules, etc.
	if u.Role == "" {
		u.Role = "user"
	}

	// u.PasswordHash hiện đang chứa plaintext password (từ handler)
	if u.PasswordHash == "" {
		return errors.New("password is required")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(u.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hashed)

	return s.repo.Create(ctx, u)
}

// Authenticate kiểm tra email + password, trả về user nếu đúng.
func (s *userService) Authenticate(ctx context.Context, email, password string) (*model.User, error) {
	u, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("email not found")
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	return u, nil
}

func (s *userService) GetByID(ctx context.Context, id int64) (*model.User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *userService) List(ctx context.Context) ([]*model.User, error) {
	return s.repo.List(ctx)
}

func (s *userService) Update(ctx context.Context, u *model.User) error {
	return s.repo.Update(ctx, u)
}

func (s *userService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func (s *userService) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	return s.repo.GetByEmail(ctx, email)
}
