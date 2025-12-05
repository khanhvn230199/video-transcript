package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"video-transcript/internal/middleware"
	"video-transcript/internal/model"
	"video-transcript/internal/service"
)

// TaskHandler exposes task-related endpoints.
type TaskHandler struct {
	svc service.TaskService
}

// NewTaskHandler creates a new TaskHandler.
func NewTaskHandler(svc service.TaskService) *TaskHandler {
	return &TaskHandler{svc: svc}
}

// RegisterRoutes registers task routes under /tasks (JWT required).
func (h *TaskHandler) RegisterRoutes(r *gin.Engine) {
	g := r.Group("/tasks", middleware.JWTAuth())
	g.POST("", h.create)
	g.GET("", h.listByUser)
	g.GET("/:id", h.getByID)
}

type createTaskRequest struct {
	TaskType  string  `json:"task_type" binding:"required"` // "stt" | "tts"
	InputText *string `json:"input_text,omitempty"`         // For TTS
	InputURL  *string `json:"input_url,omitempty"`          // For STT
}

func (h *TaskHandler) create(c *gin.Context) {
	currentUser := middleware.CurrentUser(c)
	if currentUser == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var in createTaskRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tt := model.TaskType(in.TaskType)
	if tt != model.TaskTypeSTT && tt != model.TaskTypeTTS {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task_type"})
		return
	}

	// Basic input validation per type.
	if tt == model.TaskTypeTTS && (in.InputText == nil || *in.InputText == "") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "input_text is required for tts"})
		return
	}
	if tt == model.TaskTypeSTT && (in.InputURL == nil || *in.InputURL == "") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "input_url is required for stt"})
		return
	}

	task := &model.Task{
		TaskType:  tt,
		Status:    model.TaskStatusPending,
		InputText: in.InputText,
		InputURL:  in.InputURL,
		UserID:    &currentUser.ID,
	}

	if err := h.svc.Create(c.Request.Context(), task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"task": task})
}

func (h *TaskHandler) getByID(c *gin.Context) {
	currentUser := middleware.CurrentUser(c)
	if currentUser == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	task, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	if task.UserID != nil && *task.UserID != currentUser.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"task": task})
}

func (h *TaskHandler) listByUser(c *gin.Context) {
	currentUser := middleware.CurrentUser(c)
	if currentUser == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	limit := 20
	offset := 0
	if v := c.Query("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	if v := c.Query("offset"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	tasks, err := h.svc.ListByUser(c.Request.Context(), currentUser.ID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tasks": tasks})
}
