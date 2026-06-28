// Package middleware 限流中间件单元测试
// 无 Redis 环境下验证 fail-open 降级路径（直接放行）。
// Redis 可用时配合 testcontainers 或集成测试覆盖限流逻辑。
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	redispkg "wuchang-tongcheng/internal/pkg/redis"
)

// newRateLimitRouter 构造限流路由
func newRateLimitRouter(maxCount, windowSec int, prefix string) *gin.Engine {
	r := gin.New()
	r.Use(RateLimit(maxCount, windowSec, prefix))
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"code": 0})
	})
	return r
}

// TestRateLimit_RedisUnavailable_FailOpen 无 Redis 时降级放行
// 这是开发环境/CI 无 Redis 时的核心保障，必须验证。
func TestRateLimit_RedisUnavailable_FailOpen(t *testing.T) {
	// 当前测试环境无 Redis，IsAvailable 应返回 false
	if redispkg.IsAvailable() {
		t.Skip("本环境 Redis 可用，跳过 fail-open 测试")
	}

	r := newRateLimitRouter(5, 60, "login")

	// 连续打 100 次请求，无 Redis 时应全部放行
	for i := 0; i < 100; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code, "第 %d 次请求应放行（fail-open）", i)
		assert.Contains(t, w.Body.String(), `"code":0`)
	}
}

// TestRateLimit_RedisUnavailable_NoHeaders 降级时不设置限流响应头
func TestRateLimit_RedisUnavailable_NoHeaders(t *testing.T) {
	if redispkg.IsAvailable() {
		t.Skip("本环境 Redis 可用，跳过 fail-open 测试")
	}

	r := newRateLimitRouter(5, 60, "login")
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	// 降级路径直接 c.Next()，不应设置限流头
	assert.Empty(t, w.Header().Get("X-RateLimit-Limit"))
	assert.Empty(t, w.Header().Get("X-RateLimit-Remaining"))
	assert.Empty(t, w.Header().Get("Retry-After"))
}
