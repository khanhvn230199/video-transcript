package main

import (
	"context"
	"database/sql"
	"flag"
	"log"

	_ "github.com/lib/pq"
	"go.uber.org/zap"

	speakClient "github.com/deepgram/deepgram-go-sdk/v3/pkg/client/speak"

	"video-transcript/internal/app"
	"video-transcript/internal/config"
	"video-transcript/internal/uploads"
)

func main() {
	// Init Deepgram client first - this will register klog flags
	// This must be done before flag.Parse() to avoid conflicts
	speakClient.Init(speakClient.InitLib{
		LogLevel: speakClient.LogLevelTrace,
	})

	// Disable klog logging (flags already registered by speakClient.Init)
	flag.Set("logtostderr", "false")
	flag.Set("alsologtostderr", "false")
	flag.Parse()

	// Setup global zap logger so zap.S() in other packages actually logs.
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}
	zap.ReplaceGlobals(logger)
	defer logger.Sync()

	db, err := sql.Open("postgres", config.SvcCfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping db: %v", err)
	}

	// Init R2 client (dùng cho upload).
	if err := uploads.InitR2(context.Background()); err != nil {
		log.Fatalf("failed to init R2: %v", err)
	}

	appInstance := app.NewApp(db)

	addr := ":8080"
	log.Printf("listening on %s", addr)
	if err := appInstance.Run(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
