package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"video-transcript/internal/middleware"
	"video-transcript/internal/model"
	"video-transcript/internal/service"
)

// AuthHandler xử lý logic đăng ký / đăng nhập.
type AuthHandler struct {
	svc service.UserService
}

// NewAuthHandler tạo AuthHandler mới.
func NewAuthHandler(svc service.UserService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// RegisterRoutes đăng ký các route auth (không cần JWT).
func (h *AuthHandler) RegisterRoutes(r *gin.Engine) {
	authGroup := r.Group("/auth")
	authGroup.POST("/register", h.register)
	authGroup.POST("/login", h.login)
}

// register: đăng ký user mới (public).
func (h *AuthHandler) register(c *gin.Context) {
	var in struct {
		Email     string  `json:"email"`
		Password  string  `json:"password"`
		Name      *string `json:"name"`
		AvatarURL *string `json:"avatar_url"`
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
		Role:         "user",
		Credit:       0,
	}

	if err := h.svc.Create(c.Request.Context(), u); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":    u.ID,
		"email": u.Email,
		"name":  u.Name,
		"role":  u.Role,
	})
}

// login: nhận email + password, trả về JWT token nếu đúng.
func (h *AuthHandler) login(c *gin.Context) {
	var in struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u, err := h.svc.Authenticate(c.Request.Context(), in.Email, in.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	token, err := middleware.GenerateToken(u.ID, u.Email, u.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"id":    u.ID,
			"email": u.Email,
			"name":  u.Name,
			"role":  u.Role,
		},
	})
}
