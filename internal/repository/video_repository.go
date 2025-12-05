package repository

import (
	"context"
	"database/sql"
	"errors"

	"video-transcript/internal/model"

	"go.uber.org/zap"
)

// VideoRepository defines operations for videos.
type VideoRepository interface {
	Create(ctx context.Context, v *model.Video) error
	GetByID(ctx context.Context, id int64) (*model.Video, error)
	GetVideoByUserIDAndURL(ctx context.Context, userID int64, url string) ([]*model.Video, error)
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

func (r *videoRepository) GetByID(ctx context.Context, id int64) (*model.Video, error) {
	query := `
		SELECT id, user_id, link_video, name_file, description, created_at, updated_at
		FROM videos
		WHERE id = $1
	`
	v := &model.Video{}
	err := r.db.
		QueryRowContext(ctx, query, id).
		Scan(
			&v.ID,
			&v.UserID,
			&v.LinkVideo,
			&v.NameFile,
			&v.Description,
			&v.CreatedAt,
			&v.UpdatedAt,
		)
	if err != nil {
		if err == sql.ErrNoRows {
			zap.S().Errorw("video not found", "error", err)
			return nil, errors.New("video not found")
		}
		zap.S().Errorw("get video by id failed", "error", err)
		return nil, err
	}
	return v, nil
}

func (r *videoRepository) GetVideoByUserIDAndURL(ctx context.Context, userID int64, url string) ([]*model.Video, error) {
	query := `
		SELECT id, user_id, link_video, name_file, description, created_at, updated_at
		FROM videos
		WHERE user_id = $1 AND link_video = $2
	`
	videos := []*model.Video{}
	rows, err := r.db.QueryContext(ctx, query, userID, url)
	if err != nil {
		if err == sql.ErrNoRows {
			zap.S().Errorw("video not found", "error", err)
			return nil, errors.New("video not found")
		}
		zap.S().Errorw("get video by user id and url failed", "error", err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		v := &model.Video{}
		err := rows.Scan(&v.ID, &v.UserID, &v.LinkVideo, &v.NameFile, &v.Description, &v.CreatedAt, &v.UpdatedAt)
		if err != nil {
			zap.S().Errorw("scan video by user id and url failed", "error", err)
			continue
		}
		videos = append(videos, &model.Video{
			ID:          v.ID,
			UserID:      v.UserID,
			LinkVideo:   v.LinkVideo,
			NameFile:    v.NameFile,
			Description: v.Description,
			CreatedAt:   v.CreatedAt,
			UpdatedAt:   v.UpdatedAt,
		})
	}
	if len(videos) == 0 {
		zap.S().Errorw("video not found", "error", errors.New("video not found"))
		return nil, errors.New("video not found")
	}
	return videos, nil
}
