package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"video-transcript/internal/config"
	"video-transcript/internal/model"
)

// Claims đại diện payload của JWT.
type Claims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// JWTAuth là middleware Gin để verify JWT từ header Authorization: Bearer <token>.
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing Authorization header"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid Authorization header"})
			return
		}

		tokenString := parts[1]

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.SvcCfg.JWTSecret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		// Lưu thông tin user (model.User đơn giản) vào context để handler phía sau dùng.
		currentUser := &model.User{
			ID:    claims.UserID,
			Email: claims.Email,
			Role:  claims.Role,
		}
		if claims.UserID == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Set("currentUser", currentUser)

		c.Next()
	}
}

// GenerateToken tạo JWT cho user id + email.
func GenerateToken(userID int64, email string, role string) (string, error) {
	now := time.Now()
	expMinutes := time.Duration(config.SvcCfg.JWTExpireMinute) * time.Minute

	claims := &Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   email,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expMinutes)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.SvcCfg.JWTSecret))
}

// CurrentUser lấy user từ context (được set bởi JWTAuth).
func CurrentUser(c *gin.Context) *model.User {
	v, ok := c.Get("currentUser")
	if !ok {
		return nil
	}
	u, _ := v.(*model.User)
	return u
}
