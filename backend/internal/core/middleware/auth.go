package middleware

import (
	"net/http"
	"strings"

	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/pkg/jwt"

	"github.com/gin-gonic/gin"
)

// 上下文中存储用户信息的Key
const (
	ContextUserID   = "user_id"
	ContextUsername = "username"
)

// Auth 鉴权中间件
// 解析JWT Token，将用户信息存入上下文；无Token视为游客放行（用于公开接口）
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 Authorization 头获取，格式: Bearer <token>
		token := c.GetHeader("Authorization")
		if token != "" {
			token = strings.TrimPrefix(token, "Bearer ")
		} else {
			// 兼容从 query 参数获取
			token = c.Query("token")
		}

		// 无Token，游客身份放行（公开接口），需要登录的接口用 AuthRequired
		if token == "" {
			c.Set(ContextUserID, uint(0))
			c.Set(ContextUsername, "guest")
			c.Next()
			return
		}

		// 解析Token
		claims, err := jwt.ParseToken(token)
		if err != nil {
			c.Set(ContextUserID, uint(0))
			c.Set(ContextUsername, "guest")
			c.Next()
			return
		}

		// 将用户信息存入上下文
		c.Set(ContextUserID, claims.UserID)
		c.Set(ContextUsername, claims.Username)
		c.Next()
	}
}

// AuthRequired 必须登录的中间件
// 放在需要登录的路由上：router.GET("/profile", middleware.AuthRequired(), handler)
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get(ContextUserID)
		if !exists || userID == uint(0) {
			c.JSON(http.StatusOK, response.Unauthorized("请先登录"))
			c.Abort()
			return
		}
		c.Next()
	}
}

// GetUserID 从上下文中获取用户ID
func GetUserID(c *gin.Context) uint {
	if value, exists := c.Get(ContextUserID); exists {
		if id, ok := value.(uint); ok {
			return id
		}
	}
	return 0
}

// GetUsername 从上下文中获取用户名
func GetUsername(c *gin.Context) string {
	if value, exists := c.Get(ContextUsername); exists {
		if name, ok := value.(string); ok {
			return name
		}
	}
	return ""
}
