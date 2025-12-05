package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	interfacesv1 "github.com/deepgram/deepgram-go-sdk/v3/pkg/api/listen/v1/rest/interfaces"
)

// TaskType represents the type of task
type TaskType string

const (
	TaskTypeSTT TaskType = "stt" // Speech-to-Text
	TaskTypeTTS TaskType = "tts" // Text-to-Speech
)

// Value implements driver.Valuer interface
func (t TaskType) Value() (driver.Value, error) {
	return string(t), nil
}

// Scan implements sql.Scanner interface
func (t *TaskType) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	str, ok := value.(string)
	if !ok {
		return errors.New("failed to scan TaskType")
	}
	*t = TaskType(str)
	return nil
}

// TaskStatus represents the status of a task
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusProcessing TaskStatus = "processing"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
)

// Value implements driver.Valuer interface
func (s TaskStatus) Value() (driver.Value, error) {
	return string(s), nil
}

// Scan implements sql.Scanner interface
func (s *TaskStatus) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	str, ok := value.(string)
	if !ok {
		return errors.New("failed to scan TaskStatus")
	}
	*s = TaskStatus(str)
	return nil
}

// Task represents a row in the `tasks` table.
type Task struct {
	ID             int64           `db:"id" json:"id"`
	TaskType       TaskType        `db:"task_type" json:"task_type"`
	Status         TaskStatus      `db:"status_task" json:"status"`
	InputText      *string         `db:"input_text" json:"input_text,omitempty"`
	InputURL       *string         `db:"input_url" json:"input_url,omitempty"`
	OutputURL      *string         `db:"output_url" json:"output_url,omitempty"`
	TranscriptText *string         `db:"transcript_text" json:"transcript_text,omitempty"`
	TranscriptJSON json.RawMessage `db:"transcript_json" json:"transcript_json,omitempty"`
	DurationSec    *float64        `db:"duration_sec" json:"duration_sec,omitempty"`
	ErrorMessage   *string         `db:"error_message" json:"error_message,omitempty"`
	UserID         *int64          `db:"user_id" json:"user_id,omitempty"`
	CreatedAt      time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time       `db:"updated_at" json:"updated_at"`
}

type SimpleWord struct {
	Word  string  `json:"word"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}

type SimpleUtterance struct {
	Start      float64 `json:"start"`
	End        float64 `json:"end"`
	Transcript string  `json:"transcript"`
}

type SimpleTranscript struct {
	TranscriptText string            `json:"transcript_text"`
	Words          []SimpleWord      `json:"words"`
	Utterances     []SimpleUtterance `json:"utterances"`
}

func ConvertDeepgramToSimple(resp *interfacesv1.PreRecordedResponse) (*SimpleTranscript, error) {
	out := &SimpleTranscript{}
	if resp.Results == nil {
		return nil, errors.New("results is nil")
	}

	if len(resp.Results.Utterances) == 0 {
		return nil, errors.New("utterances is nil")
	}

	if len(resp.Results.Channels) > 0 {
		if len(resp.Results.Channels[0].Alternatives) > 0 {
			out.TranscriptText = resp.Results.Channels[0].Alternatives[0].Transcript
		}
	}
	// Build words (flatten all utterances)
	for _, utt := range resp.Results.Utterances {
		for _, w := range utt.Words {
			out.Words = append(out.Words, SimpleWord{
				Word:  w.PunctuatedWord, // dùng chữ có punctuation
				Start: w.Start,
				End:   w.End,
			})
		}
	}

	// Build utterances
	for _, utt := range resp.Results.Utterances {
		out.Utterances = append(out.Utterances, SimpleUtterance{
			Start:      utt.Start,
			End:        utt.End,
			Transcript: utt.Transcript,
		})
	}

	return out, nil
}
