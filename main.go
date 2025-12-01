package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"

	"video-transcript/internal/app"
	"video-transcript/internal/config"
)

func main() {
	db, err := sql.Open("postgres", config.SvcCfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping db: %v", err)
	}

	appInstance := app.NewApp(db)

	addr := ":8081"
	log.Printf("listening on %s", addr)
	if err := appInstance.Run(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
