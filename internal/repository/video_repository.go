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
	ListVideoByUserID(ctx context.Context, userID int64, limit, offset int, search string) (*model.ListVideoByUserIDResponse, error)
	UpdateDescription(ctx context.Context, id int64, description *string) error
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

func (r *videoRepository) ListVideoByUserID(ctx context.Context, userID int64, limit, offset int, search string) (*model.ListVideoByUserIDResponse, error) {
	var query string
	var queryCount string
	var args []interface{}
	var countArgs []interface{}

	if search != "" {
		query = `
			SELECT id, user_id, link_video, name_file, description, created_at, updated_at
			FROM videos
			WHERE user_id = $1 AND description ILIKE $2
			ORDER BY updated_at DESC
			LIMIT $3 OFFSET $4
		`
		searchPattern := "%" + search + "%"
		args = []interface{}{userID, searchPattern, limit, offset}

		queryCount = `SELECT COUNT(*) FROM videos WHERE user_id = $1 AND description ILIKE $2`
		countArgs = []interface{}{userID, searchPattern}
	} else {
		query = `
			SELECT id, user_id, link_video, name_file, description, created_at, updated_at
			FROM videos
			WHERE user_id = $1
			ORDER BY updated_at DESC
			LIMIT $2 OFFSET $3
		`
		args = []interface{}{userID, limit, offset}

		queryCount = `SELECT COUNT(*) FROM videos WHERE user_id = $1`
		countArgs = []interface{}{userID}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		zap.S().Errorw("list video by user id failed", "user_id", userID, "error", err)
		return nil, err
	}
	defer rows.Close()

	videos := []*model.Video{}
	for rows.Next() {
		v := &model.Video{}
		err := rows.Scan(&v.ID, &v.UserID, &v.LinkVideo, &v.NameFile, &v.Description, &v.CreatedAt, &v.UpdatedAt)
		if err != nil {
			zap.S().Errorw("scan video by user id failed", "error", err)
			continue
		}
		videos = append(videos, v)
	}

	var totalCount int
	err = r.db.QueryRowContext(ctx, queryCount, countArgs...).Scan(&totalCount)
	if err != nil {
		zap.S().Errorw("get total video by user id failed", "user_id", userID, "error", err)
		return nil, err
	}

	page := (offset / limit) + 1
	if limit == 0 {
		page = 1
	}
	totalPages := (totalCount + limit - 1) / limit // Ceiling division
	if totalPages == 0 {
		totalPages = 1
	}

	return &model.ListVideoByUserIDResponse{
		Page:       page,
		PageSize:   limit,
		Total:      totalCount,
		TotalPages: totalPages,
		Videos:     videos,
	}, nil
}

func (r *videoRepository) UpdateDescription(ctx context.Context, id int64, description *string) error {
	query := `
		UPDATE videos
		SET description = $1, updated_at = NOW()
		WHERE id = $2
	`
	_, err := r.db.ExecContext(ctx, query, description, id)
	if err != nil {
		zap.S().Errorw("update video description failed", "id", id, "error", err)
		return err
	}
	return nil
}
