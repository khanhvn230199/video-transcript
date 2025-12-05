package service

import (
	"context"

	"video-transcript/internal/model"
	"video-transcript/internal/repository"
)

// VideoService defines business logic for videos.
type VideoService interface {
	Create(ctx context.Context, v *model.Video) error
	GetByID(ctx context.Context, id int64) (*model.Video, error)
	GetVideoByUserIDAndURL(ctx context.Context, userID int64, url string) ([]*model.Video, error)
}

type videoService struct {
	repo repository.VideoRepository
}

// NewVideoService creates a new VideoService.
func NewVideoService(repo repository.VideoRepository) VideoService {
	return &videoService{repo: repo}
}

func (s *videoService) Create(ctx context.Context, v *model.Video) error {
	return s.repo.Create(ctx, v)
}

func (s *videoService) GetByID(ctx context.Context, id int64) (*model.Video, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *videoService) GetVideoByUserIDAndURL(ctx context.Context, userID int64, url string) ([]*model.Video, error) {
	return s.repo.GetVideoByUserIDAndURL(ctx, userID, url)
}
