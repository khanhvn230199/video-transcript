package app

import (
	"database/sql"

	"github.com/gin-gonic/gin"

	"video-transcript/internal/handler"
	"video-transcript/internal/middleware"
	"video-transcript/internal/repository"
	"video-transcript/internal/service"
)

// App giữ toàn bộ wiring cho HTTP server.
type App struct {
	Engine      *gin.Engine
	UserHandler *handler.UserHandler
}

// NewApp khởi tạo repository, service, handler và router.
func NewApp(db *sql.DB) *App {
	// init repositories
	userRepo := repository.NewUserRepository(db)
	videoRepo := repository.NewVideoRepository(db)
	taskRepo := repository.NewTaskRepository(db)

	// init services
	userSvc := service.NewUserService(userRepo)
	videoSvc := service.NewVideoService(videoRepo)
	taskSvc := service.NewTaskService(taskRepo)

	// init handlers
	userHandler := handler.NewUserHandler(userSvc)
	authHandler := handler.NewAuthHandler(userSvc)
	uploadHandler := handler.NewUploadHandler(videoSvc)
	deepgramHandler := handler.NewDeepgramHandler(videoSvc, taskSvc)
	taskHandler := handler.NewTaskHandler(taskSvc)

	r := gin.Default()

	// Global middleware
	r.Use(middleware.CORSMiddleware())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// Auth routes (public)
	authHandler.RegisterRoutes(r)

	// Upload routes (public)
	uploadHandler.RegisterRoutes(r)

	userHandler.RegisterRoutes(r)

	deepgramHandler.RegisterRoutes(r)
	taskHandler.RegisterRoutes(r)

	return &App{
		Engine:      r,
		UserHandler: userHandler,
	}
}

// Run chạy HTTP server.
func (a *App) Run(addr string) error {
	return a.Engine.Run(addr)
}
