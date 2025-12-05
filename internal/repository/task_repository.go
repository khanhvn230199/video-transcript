package repository

import (
	"context"
	"database/sql"
	"errors"

	"video-transcript/internal/model"

	"go.uber.org/zap"
)

// TaskRepository defines operations for tasks.
type TaskRepository interface {
	Create(ctx context.Context, t *model.Task) error
	GetByID(ctx context.Context, id int64) (*model.Task, error)
	ListByUser(ctx context.Context, userID int64, limit, offset int) ([]*model.Task, error)
	UpdateStatus(ctx context.Context, id int64, status model.TaskStatus, outputURL *string, durationSec *float64, errorMessage *string) error
	UpdateTranscript(ctx context.Context, id int64, status model.TaskStatus, transcriptText *string, transcriptJSON []byte) error
}

type taskRepository struct {
	db *sql.DB
}

// NewTaskRepository returns a concrete implementation of TaskRepository.
func NewTaskRepository(db *sql.DB) TaskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) Create(ctx context.Context, t *model.Task) error {
	query := `
		INSERT INTO tasks (task_type, status_task, input_text, input_url, output_url, transcript_text, transcript_json, duration_sec, error_message, user_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at
	`
	// Xử lý transcript_json: nếu nil hoặc rỗng thì truyền NULL
	var transcriptJSON interface{}
	if len(t.TranscriptJSON) == 0 {
		transcriptJSON = nil
	} else {
		transcriptJSON = t.TranscriptJSON
	}

	return r.db.
		QueryRowContext(
			ctx,
			query,
			t.TaskType,
			t.Status,
			t.InputText,
			t.InputURL,
			t.OutputURL,
			t.TranscriptText,
			transcriptJSON,
			t.DurationSec,
			t.ErrorMessage,
			t.UserID,
		).
		Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
}

func (r *taskRepository) GetByID(ctx context.Context, id int64) (*model.Task, error) {
	query := `
		SELECT id, task_type, status_task, input_text, input_url, output_url, transcript_text, transcript_json, duration_sec, error_message, user_id, created_at, updated_at
		FROM tasks
		WHERE id = $1
	`
	t := &model.Task{}
	err := r.db.
		QueryRowContext(ctx, query, id).
		Scan(
			&t.ID,
			&t.TaskType,
			&t.Status,
			&t.InputText,
			&t.InputURL,
			&t.OutputURL,
			&t.TranscriptText,
			&t.TranscriptJSON,
			&t.DurationSec,
			&t.ErrorMessage,
			&t.UserID,
			&t.CreatedAt,
			&t.UpdatedAt,
		)
	if err != nil {
		if err == sql.ErrNoRows {
			zap.S().Infow("task not found", "id", id)
			return nil, errors.New("task not found")
		}
		zap.S().Errorw("get task by id failed", "id", id, "error", err)
		return nil, err
	}
	return t, nil
}

func (r *taskRepository) ListByUser(ctx context.Context, userID int64, limit, offset int) ([]*model.Task, error) {
	query := `
		SELECT id, task_type, status_task, input_text, input_url, output_url, transcript_text, transcript_json, duration_sec, error_message, user_id, created_at, updated_at
		FROM tasks
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		zap.S().Errorw("list tasks by user failed", "user_id", userID, "error", err)
		return nil, err
	}
	defer rows.Close()

	tasks := make([]*model.Task, 0)
	for rows.Next() {
		t := &model.Task{}
		if err := rows.Scan(
			&t.ID,
			&t.TaskType,
			&t.Status,
			&t.InputText,
			&t.InputURL,
			&t.OutputURL,
			&t.TranscriptText,
			&t.TranscriptJSON,
			&t.DurationSec,
			&t.ErrorMessage,
			&t.UserID,
			&t.CreatedAt,
			&t.UpdatedAt,
		); err != nil {
			zap.S().Errorw("scan task failed", "user_id", userID, "error", err)
			continue
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (r *taskRepository) UpdateStatus(ctx context.Context, id int64, status model.TaskStatus, outputURL *string, durationSec *float64, errorMessage *string) error {
	query := `
		UPDATE tasks
		SET status_task = $2,
			output_url = $3,
			duration_sec = $4,
			error_message = $5,
			updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at
	`
	var updatedAt sql.NullTime
	if err := r.db.QueryRowContext(ctx, query, id, status, outputURL, durationSec, errorMessage).Scan(&updatedAt); err != nil {
		zap.S().Errorw("update task status failed", "id", id, "error", err)
		return err
	}
	return nil
}

func (r *taskRepository) UpdateTranscript(ctx context.Context, id int64, status model.TaskStatus, transcriptText *string, transcriptJSON []byte) error {
	query := `
		UPDATE tasks
		SET transcript_text = $2,
			transcript_json = $3,
			status_task = $4,
			updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at
	`
	// Xử lý transcript_json: nếu nil hoặc rỗng thì truyền NULL
	var transcriptJSONVal interface{}
	if len(transcriptJSON) == 0 {
		transcriptJSONVal = nil
	} else {
		transcriptJSONVal = transcriptJSON
	}

	var updatedAt sql.NullTime
	if err := r.db.QueryRowContext(ctx, query, id, transcriptText, transcriptJSONVal, status).Scan(&updatedAt); err != nil {
		zap.S().Errorw("update task transcript failed", "id", id, "error", err)
		return err
	}
	return nil
}
