package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"video-transcript/internal/helper"
	"video-transcript/internal/middleware"
	"video-transcript/internal/model"
	"video-transcript/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type DeepgramHandler struct {
	videoSvc service.VideoService
	taskSvc  service.TaskService
}

func NewDeepgramHandler(videoSvc service.VideoService, taskSvc service.TaskService) *DeepgramHandler {
	return &DeepgramHandler{videoSvc: videoSvc, taskSvc: taskSvc}
}

func (h *DeepgramHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	deepgramGroup := r.Group("/deepgram", authMiddleware)
	deepgramGroup.POST("/tts", h.DeepgramTTS)
	deepgramGroup.POST("/stt", h.DeepgramSTT)
}

func (h *DeepgramHandler) DeepgramTTS(c *gin.Context) {
	currentUser := middleware.CurrentUser(c)
	if currentUser == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID := currentUser.ID
	ctx := context.Background()
	var in struct {
		Text string `json:"text"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		zap.S().Errorw("should bind json failed", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// create task
	task := &model.Task{
		TaskType:  model.TaskTypeTTS,
		Status:    model.TaskStatusPending,
		InputText: &in.Text,
		UserID:    &userID,
	}
	if err := h.taskSvc.Create(ctx, task); err != nil {
		zap.S().Errorw("create task failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	go func() {
		url, err := helper.DeepgramTTS(ctx, fmt.Sprintf("%d", userID), in.Text)
		if err != nil {
			zap.S().Errorw("deepgram tts failed", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			errorMessage := err.Error()
			err := h.taskSvc.UpdateStatus(ctx, task.ID, model.TaskStatusFailed, nil, nil, &errorMessage)
			if err != nil {
				zap.S().Errorw("update task status failed", "id", task.ID, "error", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			return
		}
		uploadVideo := &model.Video{
			UserID:      userID,
			LinkVideo:   url,
			NameFile:    "deepgram-tts.mp3",
			Description: &in.Text,
		}
		if err := h.videoSvc.Create(ctx, uploadVideo); err != nil {
			zap.S().Errorw("create video failed", "user_id", userID, "file_url", url, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		err = h.taskSvc.UpdateStatus(ctx, task.ID, model.TaskStatusCompleted, &url, nil, nil)
		if err != nil {
			zap.S().Errorw("update task status failed", "id", task.ID, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}()

	c.JSON(http.StatusOK, gin.H{"data": task})
}

func (h *DeepgramHandler) DeepgramSTT(c *gin.Context) {
	currentUser := middleware.CurrentUser(c)
	if currentUser == nil {
		zap.S().Errorw("current user is nil")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID := currentUser.ID
	ctx := context.Background()
	var in struct {
		FileURL string `json:"file_url"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		zap.S().Errorw("should bind json failed", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// create task
	task := &model.Task{
		TaskType: model.TaskTypeSTT,
		Status:   model.TaskStatusPending,
		InputURL: &in.FileURL,
		UserID:   &userID,
	}
	if err := h.taskSvc.Create(ctx, task); err != nil {
		zap.S().Errorw("create task failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	go func() {
		res, err := helper.DeepgramSTTFromBytes(ctx, in.FileURL, "audio/mpeg")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			errorMessage := err.Error()
			err := h.taskSvc.UpdateStatus(ctx, task.ID, model.TaskStatusFailed, nil, nil, &errorMessage)
			if err != nil {
				zap.S().Errorw("update task status failed", "id", task.ID, "error", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			return
		}

		videos, err := h.videoSvc.GetVideoByUserIDAndURL(ctx, userID, in.FileURL)
		if err != nil {
			zap.S().Errorw("get video by user id and url failed", "user_id", userID, "file_url", in.FileURL, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if len(videos) == 0 {
			uploadVideo := &model.Video{
				UserID:      userID,
				LinkVideo:   in.FileURL,
				NameFile:    "",
				Description: nil,
			}
			if err := h.videoSvc.Create(ctx, uploadVideo); err != nil {
				zap.S().Errorw("create video failed", "user_id", userID, "file_url", in.FileURL, "error", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
		simpleTranscript, err := model.ConvertDeepgramToSimple(res)
		if err != nil {
			zap.S().Errorw("convert deepgram to simple transcript failed", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		transcriptJSON, err := json.Marshal(simpleTranscript)
		if err != nil {
			zap.S().Errorw("marshal simple transcript failed", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		err = h.taskSvc.UpdateTranscript(ctx, task.ID, model.TaskStatusCompleted, &simpleTranscript.TranscriptText, transcriptJSON)
		if err != nil {
			zap.S().Errorw("update task transcript failed", "id", task.ID, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}()

	c.JSON(http.StatusOK, gin.H{"data": task})
}
