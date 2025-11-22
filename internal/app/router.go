package app

import (
	"net/http"
	"os"
	"video-transcript/internal/handler"
	"video-transcript/internal/repository"
	"video-transcript/internal/service"

	"github.com/gin-gonic/gin"
)

// SetupRouter sets up the Gin router with all routes
func SetupRouter() *gin.Engine {
	r := gin.Default()

	// Home route
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	// Initialize repositories
	userRepo := repository.NewUserRepository(GetDB())

	// Initialize services
	userService := service.NewUserService(userRepo)

	// Initialize handlers
	userHandler := handler.NewUserHandler(userService)

	// API routes
	api := r.Group("/api")
	{
		// User routes
		users := api.Group("/users")
		{
			users.POST("", userHandler.CreateUser)
			users.GET("", userHandler.ListUsers)
			users.POST("/login", userHandler.Login)
			users.GET("/:id", userHandler.GetUser)
			users.PUT("/:id", userHandler.UpdateUser)
			users.DELETE("/:id", userHandler.DeleteUser)
		}
	}

	return r
}

// InitApp initializes the application
func InitApp() error {
	// Create uploads directory if it doesn't exist
	if err := os.MkdirAll("uploads", 0755); err != nil {
		return err
	}

	// Run database migrations
	if err := AutoMigrate(GetDB()); err != nil {
		return err
	}

	return nil
}
