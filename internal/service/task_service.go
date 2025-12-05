package service

import (
	"context"

	"video-transcript/internal/model"
	"video-transcript/internal/repository"
)

// TaskService defines business logic for tasks.
type TaskService interface {
	Create(ctx context.Context, t *model.Task) error
	GetByID(ctx context.Context, id int64) (*model.Task, error)
	ListByUser(ctx context.Context, userID int64, limit, offset int) ([]*model.Task, error)
	UpdateStatus(ctx context.Context, id int64, status model.TaskStatus, outputURL *string, durationSec *float64, errorMessage *string) error
	UpdateTranscript(ctx context.Context, id int64, status model.TaskStatus, transcriptText *string, transcriptJSON []byte) error
}

type taskService struct {
	repo repository.TaskRepository
}

// NewTaskService creates a new TaskService.
func NewTaskService(repo repository.TaskRepository) TaskService {
	return &taskService{repo: repo}
}

func (s *taskService) Create(ctx context.Context, t *model.Task) error {
	return s.repo.Create(ctx, t)
}

func (s *taskService) GetByID(ctx context.Context, id int64) (*model.Task, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *taskService) ListByUser(ctx context.Context, userID int64, limit, offset int) ([]*model.Task, error) {
	return s.repo.ListByUser(ctx, userID, limit, offset)
}

func (s *taskService) UpdateStatus(ctx context.Context, id int64, status model.TaskStatus, outputURL *string, durationSec *float64, errorMessage *string) error {
	return s.repo.UpdateStatus(ctx, id, status, outputURL, durationSec, errorMessage)
}

func (s *taskService) UpdateTranscript(ctx context.Context, id int64, status model.TaskStatus, transcriptText *string, transcriptJSON []byte) error {
	return s.repo.UpdateTranscript(ctx, id, status, transcriptText, transcriptJSON)
}
