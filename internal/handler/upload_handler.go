package handler

import (
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"

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
func (h *UploadHandler) RegisterRoutes(r gin.IRoutes) {
	r.POST("/upload", middleware.JWTAuth(), h.uploadFile)
}

// uploadFile nhận multipart/form-data với field "file" và lưu vào thư mục uploads.
func (h *UploadHandler) uploadFile(c *gin.Context) {
	currentUser := middleware.CurrentUser(c)
	if currentUser == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing file"})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not open uploaded file"})
		return
	}
	defer src.Close()

	// Tạo key duy nhất cho R2: uploads/<timestamp>-<filename>
	key := filepath.Join("uploads", fmt.Sprintf("%d-%s", time.Now().UnixNano(), filepath.Base(file.Filename)))

	url, err := uploads.UploadToR2(c.Request.Context(), key, src, file.Size, file.Header.Get("Content-Type"))
	if err != nil {
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
