package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"video-transcript/internal/middleware"
	"video-transcript/internal/model"
	"video-transcript/internal/service"
)

// UserHandler holds dependencies for user HTTP handlers.
type UserHandler struct {
	svc service.UserService
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// RegisterRoutes registers user routes on the given Gin router/route group.
func (h *UserHandler) RegisterRoutes(r gin.IRoutes) {
	r.POST("/users", h.createUser, middleware.JWTAuth()) // tạo user mới (admin)
	r.GET("/users", h.listUsers, middleware.JWTAuth())   // lấy danh sách user (admin)
	r.GET("/users/:id", h.getUser)
	r.PUT("/users/:id", h.updateUser, middleware.JWTAuth())
	r.DELETE("/users/:id", h.deleteUser, middleware.JWTAuth())
}

func (h *UserHandler) createUser(c *gin.Context) {
	currentUser := middleware.CurrentUser(c)
	if currentUser != nil {
		c.Writer.Header().Set("X-Requester-Email", currentUser.Email)
	}

	var in struct {
		Email     string  `json:"email"`
		Password  string  `json:"password"`
		Name      *string `json:"name"`
		AvatarURL *string `json:"avatar_url"`
		Role      string  `json:"role"`
		Credit    int     `json:"credit"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u := &model.User{
		Email:        in.Email,
		PasswordHash: in.Password, // sẽ được hash trong service
		Name:         in.Name,
		AvatarURL:    in.AvatarURL,
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

	u, err := h.svc.GetByID(c.Request.Context(), currentUser.ID)
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
	if currentUser != nil && currentUser.Role != "admin" && currentUser.ID != 0 {
		// Nếu không phải admin, có thể enforce update chính mình (ví dụ).
		// Ở đây chỉ là ví dụ đơn giản, bạn có thể tùy chỉnh rule.
	}

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var in struct {
		Email     string  `json:"email"`
		Password  string  `json:"password"`
		Name      *string `json:"name"`
		AvatarURL *string `json:"avatar_url"`
		Role      string  `json:"role"`
		Credit    int     `json:"credit"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u := &model.User{
		ID:           id,
		Email:        in.Email,
		PasswordHash: in.Password, // nếu không rỗng, có thể hash lại trong service (chưa xử lý ở đây)
		Name:         in.Name,
		AvatarURL:    in.AvatarURL,
		Role:         in.Role,
		Credit:       in.Credit,
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
