package app

import (
	"database/sql"

	"github.com/gin-gonic/gin"

	"video-transcript/internal/handler"
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
	userRepo := repository.NewUserRepository(db)
	userSvc := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userSvc)
	authHandler := handler.NewAuthHandler(userSvc)

	r := gin.Default()

	// Auth routes (public)
	authHandler.RegisterRoutes(r)

	userHandler.RegisterRoutes(r)

	return &App{
		Engine:      r,
		UserHandler: userHandler,
	}
}

// Run chạy HTTP server.
func (a *App) Run(addr string) error {
	return a.Engine.Run(addr)
}
