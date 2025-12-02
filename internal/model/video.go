package model

import "time"

// Video represents a row in the `videos` table.
type Video struct {
	ID          int64     `db:"id" json:"id"`
	UserID      int64     `db:"user_id" json:"user_id"`
	LinkVideo   string    `db:"link_video" json:"link_video"`
	NameFile    string    `db:"name_file" json:"name_file"`
	Description *string   `db:"description" json:"description,omitempty"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}
