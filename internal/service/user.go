package service

import (
	"errors"
	"time"
	"video-transcript/internal/model"
	"video-transcript/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

// UserService handles business logic for users
type UserService struct {
	userRepo *repository.UserRepository
}

// NewUserService creates a new user service
func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

// CreateUser creates a new user
func (s *UserService) CreateUser(req *model.CreateUserRequest) (*model.User, error) {
	// Check if user already exists
	_, err := s.userRepo.GetByEmail(req.Email)
	if err == nil {
		return nil, errors.New("email already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Set default language if not provided
	defaultLang := req.DefaultLanguage
	if defaultLang == "" {
		defaultLang = "en"
	}

	// Create user
	user := &model.User{
		Email:           req.Email,
		PasswordHash:    string(hashedPassword),
		Name:            req.Name,
		AvatarURL:       req.AvatarURL,
		DefaultLanguage: defaultLang,
		CreditRemaining: 0,
		CreditUsed:      0,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(id string) (*model.User, error) {
	return s.userRepo.GetByID(id)
}

// GetUserByEmail retrieves a user by email
func (s *UserService) GetUserByEmail(email string) (*model.User, error) {
	return s.userRepo.GetByEmail(email)
}

// UpdateUser updates a user
func (s *UserService) UpdateUser(id string, req *model.UpdateUserRequest) (*model.User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.AvatarURL != "" {
		user.AvatarURL = req.AvatarURL
	}
	if req.DefaultLanguage != "" {
		user.DefaultLanguage = req.DefaultLanguage
	}
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	return user, nil
}

// DeleteUser deletes a user
func (s *UserService) DeleteUser(id string) error {
	return s.userRepo.Delete(id)
}

// Login authenticates a user
func (s *UserService) Login(req *model.LoginRequest) (*model.User, error) {
	user, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	return user, nil
}

// UpdateCredits updates user credits
func (s *UserService) UpdateCredits(id string, creditRemaining, creditUsed int) error {
	return s.userRepo.UpdateCredits(id, creditRemaining, creditUsed)
}

// ListUsers retrieves a list of users with pagination
func (s *UserService) ListUsers(page, pageSize int) ([]model.User, int64, error) {
	offset := (page - 1) * pageSize
	return s.userRepo.List(offset, pageSize)
}
