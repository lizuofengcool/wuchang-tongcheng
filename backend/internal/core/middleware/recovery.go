package middleware

import (
	"net/http"

	"wuchang-tongcheng/internal/core/response"

	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

// Recovery 崩溃恢复中间件
func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
				)

				c.JSON(http.StatusOK, response.ServerError("服务器内部错误"))
				c.Abort()
			}
		}()
		c.Next()
	}
}
