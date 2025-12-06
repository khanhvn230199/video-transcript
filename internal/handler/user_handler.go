package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"video-transcript/internal/middleware"
	"video-transcript/internal/model"
	"video-transcript/internal/service"
)

// UserHandler holds dependencies for user HTTP handlers.
type UserHandler struct {
	svc      service.UserService
	videoSvc service.VideoService
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(svc service.UserService, videoSvc service.VideoService) *UserHandler {
	return &UserHandler{svc: svc, videoSvc: videoSvc}
}

// RegisterRoutes registers user routes on the given Gin router/route group.
func (h *UserHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	userGroup := r.Group("/users", authMiddleware)
	userGroup.POST("", h.createUser) // tạo user mới (admin)
	userGroup.GET("", h.listUsers)   // lấy danh sách user (admin)
	userGroup.GET("/:id", h.getUser)
	userGroup.PUT("/:id", h.updateUser)
	userGroup.DELETE("/:id", h.deleteUser)
	userGroup.GET("videos/:id", h.listVideoByUserID)
}

func (h *UserHandler) createUser(c *gin.Context) {
	currentUser := middleware.CurrentUser(c)
	if currentUser == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if currentUser.Role != "admin" {
		c.Writer.Header().Set("X-Requester-Email", currentUser.Email)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var in model.CreateUserRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u := &model.User{
		Email:        in.Email,
		PasswordHash: in.Password, // sẽ được hash trong service
		Name:         in.Name,
		AvatarURL:    in.AvatarURL,
		Gender:       in.Gender,
		DOB:          in.DOB,
		Phone:        in.Phone,
		Address:      in.Address,
		Role:         in.Role,
		Credit:       in.Credit,
	}

	if err := h.svc.Create(c.Request.Context(), u); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, u)
}

func (h *UserHandler) getUser(c *gin.Context) {
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

	// Nếu không phải admin và không phải chính mình, chỉ cho xem thông tin của mình
	if currentUser.Role != "admin" && currentUser.ID != id {
		id = currentUser.ID
	}

	u, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, u)
}

func (h *UserHandler) listUsers(c *gin.Context) {
	users, err := h.svc.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

func (h *UserHandler) updateUser(c *gin.Context) {
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

	var in model.UpdateUserRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Lấy user hiện tại từ DB
	existingUser, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	zap.S().Infow("update user", "user", in)
	// Nếu không phải admin, không cho update email, credit, và role
	isAdmin := currentUser.Role == "admin"

	u := &model.User{
		ID:           id,
		Email:        existingUser.Email,
		PasswordHash: "", // Mặc định rỗng, chỉ set nếu có password mới
		Name:         in.Name,
		AvatarURL:    in.AvatarURL,
		Gender:       in.Gender,
		DOB:          in.DOB,
		Phone:        in.Phone,
		Address:      in.Address,
		Role:         existingUser.Role,
		Credit:       existingUser.Credit,
	}

	// Chỉ update password nếu có giá trị mới
	if in.Password != "" {
		u.PasswordHash = in.Password // Sẽ được hash trong service
	}

	if isAdmin {
		if in.Email != "" {
			u.Email = in.Email
		}
		if in.Role != "" {
			u.Role = in.Role
		}
		u.Credit = in.Credit
	}

	if err := h.svc.Update(c.Request.Context(), u); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, u)
}

func (h *UserHandler) deleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *UserHandler) listVideoByUserID(c *gin.Context) {
	currentUser := middleware.CurrentUser(c)
	if currentUser == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	pageSize := 10 // Default rows per page
	page := 1      // Default page number

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if currentUser.Role != "admin" {
		id = currentUser.ID
	}

	// Parse page_size (rows per page)
	if v := c.Query("page_size"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}
	// Parse page number
	if v := c.Query("page"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			page = parsed
		}
	}

	search := c.Query("search")

	// Calculate offset from page
	offset := (page - 1) * pageSize

	videos, err := h.videoSvc.ListVideoByUserID(c.Request.Context(), id, pageSize, offset, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Calculate total pages
	totalPages := (videos.Total + pageSize - 1) / pageSize // Ceiling division
	if totalPages == 0 {
		totalPages = 1
	}

	// Update response with page-based pagination
	response := map[string]interface{}{
		"page":        page,
		"page_size":   pageSize,
		"total":       videos.Total,
		"total_pages": totalPages,
		"videos":      videos.Videos,
	}

	c.JSON(http.StatusOK, response)
}
