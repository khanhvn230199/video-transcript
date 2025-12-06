package repository

import (
	"context"
	"database/sql"

	"video-transcript/internal/model"
)

// UserRepository defines CRUD operations for users.
type UserRepository interface {
	Create(ctx context.Context, u *model.User) error
	GetByID(ctx context.Context, id int64) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	List(ctx context.Context) ([]*model.User, error)
	Update(ctx context.Context, u *model.User) error
	Delete(ctx context.Context, id int64) error
}

type userRepository struct {
	db *sql.DB
}

// NewUserRepository returns a concrete implementation of UserRepository.
func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, u *model.User) error {
	query := `
		INSERT INTO users (email, password_hash, name, avatar_url, gender, dob, phone, address, role, credit)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at
	`
	return r.db.
		QueryRowContext(
			ctx,
			query,
			u.Email,
			u.PasswordHash,
			u.Name,
			u.AvatarURL,
			u.Gender,
			u.DOB,
			u.Phone,
			u.Address,
			u.Role,
			u.Credit,
		).
		Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
}

func (r *userRepository) GetByID(ctx context.Context, id int64) (*model.User, error) {
	query := `
		SELECT id, email, name, avatar_url, gender, dob, phone, address, role, credit, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	u := &model.User{}
	err := r.db.
		QueryRowContext(ctx, query, id).
		Scan(
			&u.ID,
			&u.Email,
			&u.Name,
			&u.AvatarURL,
			&u.Gender,
			&u.DOB,
			&u.Phone,
			&u.Address,
			&u.Role,
			&u.Credit,
			&u.CreatedAt,
			&u.UpdatedAt,
		)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
		SELECT id, email, password_hash, name, avatar_url, gender, dob, phone, address, role, credit, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	u := &model.User{}
	err := r.db.
		QueryRowContext(ctx, query, email).
		Scan(
			&u.ID,
			&u.Email,
			&u.PasswordHash,
			&u.Name,
			&u.AvatarURL,
			&u.Gender,
			&u.DOB,
			&u.Phone,
			&u.Address,
			&u.Role,
			&u.Credit,
			&u.CreatedAt,
			&u.UpdatedAt,
		)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *userRepository) List(ctx context.Context) ([]*model.User, error) {
	query := `
		SELECT id, email, password_hash, name, avatar_url, gender, dob, phone, address, role, credit, created_at, updated_at
		FROM users
		ORDER BY id
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		u := &model.User{}
		if err := rows.Scan(
			&u.ID,
			&u.Email,
			&u.PasswordHash,
			&u.Name,
			&u.AvatarURL,
			&u.Gender,
			&u.DOB,
			&u.Phone,
			&u.Address,
			&u.Role,
			&u.Credit,
			&u.CreatedAt,
			&u.UpdatedAt,
		); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *userRepository) Update(ctx context.Context, u *model.User) error {
	// Nếu password_hash rỗng, không update password (giữ nguyên password cũ)
	if u.PasswordHash == "" {
		query := `
			UPDATE users
			SET email = $1,
			    name = $2,
			    avatar_url = $3,
			    gender = $4,
			    dob = $5,
			    phone = $6,
			    address = $7,
			    role = $8,
			    credit = $9,
			    updated_at = NOW()
			WHERE id = $10
		`
		_, err := r.db.ExecContext(
			ctx,
			query,
			u.Email,
			u.Name,
			u.AvatarURL,
			u.Gender,
			u.DOB,
			u.Phone,
			u.Address,
			u.Role,
			u.Credit,
			u.ID,
		)
		return err
	}

	// Nếu có password mới, update cả password
	query := `
		UPDATE users
		SET email = $1,
		    password_hash = $2,
		    name = $3,
		    avatar_url = $4,
		    gender = $5,
		    dob = $6,
		    phone = $7,
		    address = $8,
		    role = $9,
		    credit = $10,
		    updated_at = NOW()
		WHERE id = $11
	`
	_, err := r.db.ExecContext(
		ctx,
		query,
		u.Email,
		u.PasswordHash,
		u.Name,
		u.AvatarURL,
		u.Gender,
		u.DOB,
		u.Phone,
		u.Address,
		u.Role,
		u.Credit,
		u.ID,
	)
	return err
}

func (r *userRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, id)
	return err
}
