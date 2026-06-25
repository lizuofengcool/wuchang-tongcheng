package middleware

import (
	"net/http"

	"wuchang-tongcheng/internal/core/response"

	"github.com/gin-gonic/gin"
)

// Auth 鉴权中间件（占位实现，待后续完善）
// 验证用户Token，将用户信息存入上下文
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: 实现完整的JWT鉴权逻辑
		token := c.GetHeader("Authorization")
		if token == "" {
			token = c.Query("token")
		}

		// 临时占位：如果没有token，直接放行（后续需要完善）
		if token == "" {
			// 开发模式下放行，生产环境需要验证
			c.Set("user_id", uint(0))
			c.Set("username", "guest")
			c.Next()
			return
		}

		// TODO: 解析JWT Token，获取用户信息
		// 临时占位处理
		c.Set("user_id", uint(1))
		c.Set("username", "admin")
		c.Next()
	}
}

// AuthRequired 必须登录的中间件
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists || userID == uint(0) {
			c.JSON(http.StatusOK, response.Unauthorized("请先登录"))
			c.Abort()
			return
		}
		c.Next()
	}
}
