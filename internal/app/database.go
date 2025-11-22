package app

import (
	"os"
	"video-transcript/internal/model"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// InitDB initializes the database connection
func InitDB() (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	// Get database URL from environment variable
	dbURL := os.Getenv("DATABASE_URL")
	dbType := os.Getenv("DB_TYPE") // "postgres" or "sqlite"

	if dbType == "postgres" && dbURL != "" {
		db, err = gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	} else {
		// Default to SQLite for development
		db, err = gorm.Open(sqlite.Open("video_transcript.db"), &gorm.Config{})
	}

	if err != nil {
		return nil, err
	}

	return db, nil
}

// AutoMigrate runs database migrations
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{},
	)
}
