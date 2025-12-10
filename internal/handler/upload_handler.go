package handler

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"video-transcript/internal/middleware"
	"video-transcript/internal/model"
	"video-transcript/internal/service"
	"video-transcript/internal/uploads"
)

// UploadHandler xử lý upload file/audio lên server (R2) và lưu metadata vào DB.
type UploadHandler struct {
	videoSvc service.VideoService
}

// NewUploadHandler tạo UploadHandler mới.
func NewUploadHandler(videoSvc service.VideoService) *UploadHandler {
	return &UploadHandler{videoSvc: videoSvc}
}

// RegisterRoutes đăng ký route upload (yêu cầu JWT).
func (h *UploadHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	uploadGroup := r.Group("/upload", authMiddleware)
	uploadGroup.POST("", h.uploadFile)
	uploadGroup.PUT("/:id", h.updateDescriptionVideo)
}

// uploadFile nhận multipart/form-data với field "file" và lưu vào thư mục uploads.
func (h *UploadHandler) uploadFile(c *gin.Context) {
	currentUser := middleware.CurrentUser(c)
	if currentUser == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Check Content-Type
	contentType := c.GetHeader("Content-Type")
	if contentType == "" || !strings.HasPrefix(contentType, "multipart/form-data") {
		zap.S().Errorw("Content-Type must be multipart/form-data",
			"content_type", contentType,
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":    "Content-Type must be multipart/form-data",
			"received": contentType,
		})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		zap.S().Errorw("failed to get file from form",
			"error", err,
			"content_type", contentType,
			"content_length", c.GetHeader("Content-Length"),
		)

		// Provide more specific error message for common issues
		errorMsg := "failed to get file from form"
		if err.Error() == "unexpected EOF" {
			errorMsg = "file upload incomplete or corrupted. Please check: 1) File size is within limit (100MB), 2) Connection is stable, 3) File is not corrupted"
		}

		c.JSON(http.StatusBadRequest, gin.H{
			"error":   errorMsg,
			"details": err.Error(),
		})
		return
	}

	// Validate file size (100MB limit)
	const maxFileSize = 100 << 20 // 100 MB
	if file.Size > maxFileSize {
		zap.S().Errorw("file too large",
			"file_size", file.Size,
			"max_size", maxFileSize,
			"filename", file.Filename,
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":        "file too large",
			"max_size_mb":  100,
			"file_size_mb": file.Size / (1 << 20),
		})
		return
	}

	src, err := file.Open()
	if err != nil {
		zap.S().Errorw("could not open uploaded file",
			"error", err,
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not open uploaded file"})
		return
	}
	defer src.Close()

	// Tạo key duy nhất cho R2: uploads/<timestamp>-<filename>
	key := filepath.Join("uploads", fmt.Sprintf("%d-%s", time.Now().UnixNano(), filepath.Base(file.Filename)))

	url, err := uploads.UploadToR2(c.Request.Context(), key, src, file.Size, file.Header.Get("Content-Type"))
	if err != nil {
		zap.S().Errorw("upload to R2 failed",
			"error", err,
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Optional: description từ form-data (field "description").
	var description *string
	if desc := c.PostForm("description"); desc != "" {
		description = &desc
	}

	video := &model.Video{
		UserID:      currentUser.ID,
		LinkVideo:   url,
		NameFile:    file.Filename,
		Description: description,
	}

	if err := h.videoSvc.Create(c.Request.Context(), video); err != nil {
		zap.S().Errorw("could not save video metadata",
			"error", err,
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save video metadata"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":          video.ID,
		"user_id":     video.UserID,
		"link_video":  video.LinkVideo,
		"name_file":   video.NameFile,
		"description": video.Description,
		"created_at":  video.CreatedAt,
		"updated_at":  video.UpdatedAt,
	})
}

func (h *UploadHandler) updateDescriptionVideo(c *gin.Context) {
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

	var in model.UpdateDescriptionVideoRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.videoSvc.UpdateDescription(c.Request.Context(), id, in.Description); err != nil {
		zap.S().Errorw("could not update video description",
			"error", err,
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update video description"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "video description updated successfully"})
}
