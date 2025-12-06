package repository

import (
	"context"
	"database/sql"
	"encoding/json"
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
	ListTaskByUserID(ctx context.Context, userID int64, limit, offset int, search string, status string) (*model.ListTaskByUserIDResponse, error)
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
	var transcriptJSON sql.NullString
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
			&transcriptJSON,
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

	// Convert sql.NullString to json.RawMessage
	if transcriptJSON.Valid {
		t.TranscriptJSON = json.RawMessage(transcriptJSON.String)
	} else {
		t.TranscriptJSON = nil
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
		var transcriptJSON sql.NullString
		if err := rows.Scan(
			&t.ID,
			&t.TaskType,
			&t.Status,
			&t.InputText,
			&t.InputURL,
			&t.OutputURL,
			&t.TranscriptText,
			&transcriptJSON,
			&t.DurationSec,
			&t.ErrorMessage,
			&t.UserID,
			&t.CreatedAt,
			&t.UpdatedAt,
		); err != nil {
			zap.S().Errorw("scan task failed", "user_id", userID, "error", err)
			continue
		}

		// Convert sql.NullString to json.RawMessage
		if transcriptJSON.Valid {
			t.TranscriptJSON = json.RawMessage(transcriptJSON.String)
		} else {
			t.TranscriptJSON = nil
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

func (r *taskRepository) ListTaskByUserID(ctx context.Context, userID int64, limit, offset int, search string, status string) (*model.ListTaskByUserIDResponse, error) {
	var query string
	var queryCount string
	var args []interface{}
	var countArgs []interface{}

	// Handle all combinations: search only, status only, both, or neither
	if search != "" && status != "" {
		// Both search and status
		query = `
			SELECT id, task_type, status_task, input_text, input_url, output_url, transcript_text, transcript_json, duration_sec, error_message, user_id, created_at, updated_at
			FROM tasks
			WHERE user_id = $1 AND task_type = $2 AND status_task = $3
			ORDER BY created_at DESC
			LIMIT $4 OFFSET $5
		`
		args = []interface{}{userID, search, status, limit, offset}
		queryCount = `SELECT COUNT(*) FROM tasks WHERE user_id = $1 AND task_type = $2 AND status_task = $3`
		countArgs = []interface{}{userID, search, status}
	} else if search != "" {
		// Only search (task_type)
		query = `
			SELECT id, task_type, status_task, input_text, input_url, output_url, transcript_text, transcript_json, duration_sec, error_message, user_id, created_at, updated_at
			FROM tasks
			WHERE user_id = $1 AND task_type = $2
			ORDER BY created_at DESC
			LIMIT $3 OFFSET $4
		`
		args = []interface{}{userID, search, limit, offset}
		queryCount = `SELECT COUNT(*) FROM tasks WHERE user_id = $1 AND task_type = $2`
		countArgs = []interface{}{userID, search}
	} else if status != "" {
		// Only status
		query = `
			SELECT id, task_type, status_task, input_text, input_url, output_url, transcript_text, transcript_json, duration_sec, error_message, user_id, created_at, updated_at
			FROM tasks
			WHERE user_id = $1 AND status_task = $2
			ORDER BY created_at DESC
			LIMIT $3 OFFSET $4
		`
		args = []interface{}{userID, status, limit, offset}
		queryCount = `SELECT COUNT(*) FROM tasks WHERE user_id = $1 AND status_task = $2`
		countArgs = []interface{}{userID, status}
	} else {
		// Neither search nor status
		query = `
			SELECT id, task_type, status_task, input_text, input_url, output_url, transcript_text, transcript_json, duration_sec, error_message, user_id, created_at, updated_at
			FROM tasks
			WHERE user_id = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3
		`
		args = []interface{}{userID, limit, offset}
		queryCount = `SELECT COUNT(*) FROM tasks WHERE user_id = $1`
		countArgs = []interface{}{userID}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		zap.S().Errorw("list tasks by user id failed", "user_id", userID, "error", err)
		return nil, err
	}
	defer rows.Close()

	tasks := make([]*model.Task, 0)
	for rows.Next() {
		t := &model.Task{}
		var transcriptJSON sql.NullString
		if err := rows.Scan(
			&t.ID,
			&t.TaskType,
			&t.Status,
			&t.InputText,
			&t.InputURL,
			&t.OutputURL,
			&t.TranscriptText,
			&transcriptJSON,
			&t.DurationSec,
			&t.ErrorMessage,
			&t.UserID,
			&t.CreatedAt,
			&t.UpdatedAt,
		); err != nil {
			zap.S().Errorw("scan task failed", "user_id", userID, "error", err)
			continue
		}

		// Convert sql.NullString to json.RawMessage
		if transcriptJSON.Valid {
			t.TranscriptJSON = json.RawMessage(transcriptJSON.String)
		} else {
			t.TranscriptJSON = nil
		}

		tasks = append(tasks, t)
	}

	var totalCount int
	err = r.db.QueryRowContext(ctx, queryCount, countArgs...).Scan(&totalCount)
	if err != nil {
		zap.S().Errorw("get total tasks by user id failed", "user_id", userID, "error", err)
		return nil, err
	}

	// Calculate page and total pages
	page := (offset / limit) + 1
	if limit == 0 {
		page = 1
	}
	totalPages := (totalCount + limit - 1) / limit // Ceiling division
	if totalPages == 0 {
		totalPages = 1
	}

	return &model.ListTaskByUserIDResponse{
		Page:       page,
		PageSize:   limit,
		Total:      totalCount,
		TotalPages: totalPages,
		Tasks:      tasks,
	}, nil
}
