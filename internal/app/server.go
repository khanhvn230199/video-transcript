package app

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

var db *gorm.DB

// RunServer starts the HTTP server
func RunServer() error {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// Initialize database
	var err error
	db, err = InitDB()
	if err != nil {
		return fmt.Errorf("failed to initialize database: %v", err)
	}

	// Initialize app
	if err := InitApp(); err != nil {
		return fmt.Errorf("failed to initialize app: %v", err)
	}

	// Setup router
	r := SetupRouter()

	// Start server
	port := ":8080"
	fmt.Printf("Server running on http://localhost%s\n", port)
	return r.Run(port)
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return db
}
