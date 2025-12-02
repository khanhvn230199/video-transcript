package repository

import (
	"context"
	"database/sql"

	"video-transcript/internal/model"
)

// VideoRepository defines operations for videos.
type VideoRepository interface {
	Create(ctx context.Context, v *model.Video) error
}

type videoRepository struct {
	db *sql.DB
}

// NewVideoRepository returns a concrete implementation of VideoRepository.
func NewVideoRepository(db *sql.DB) VideoRepository {
	return &videoRepository{db: db}
}

func (r *videoRepository) Create(ctx context.Context, v *model.Video) error {
	query := `
		INSERT INTO videos (user_id, link_video, name_file, description)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`
	return r.db.
		QueryRowContext(
			ctx,
			query,
			v.UserID,
			v.LinkVideo,
			v.NameFile,
			v.Description,
		).
		Scan(&v.ID, &v.CreatedAt, &v.UpdatedAt)
}
