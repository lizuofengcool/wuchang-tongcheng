// Package middleware 限流中间件
// 基于 Redis INCR + EXPIRE 实现的固定窗口限流，按 IP + 路由分组计数
// Redis 不可用时优雅降级（直接放行，不阻塞业务）
package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"wuchang-tongcheng/internal/core/response"
	redispkg "wuchang-tongcheng/internal/pkg/redis"
	"wuchang-tongcheng/internal/pkg/utils"

	"github.com/gin-gonic/gin"
)

// RateLimit 基于 Redis 的固定窗口限流中间件
//
// 参数：
//   - maxCount: 窗口内允许的最大请求数
//   - windowSec: 窗口时长（秒）
//   - keyPrefix: 限流键前缀（用于区分不同业务，如 "login"、"news"）
//
// 限流维度：IP + keyPrefix，超限返回 429 Too Many Requests
// Redis 不可用时直接放行（fail-open，保证业务可用）
func RateLimit(maxCount int, windowSec int, keyPrefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Redis 不可用，降级放行
		if !redispkg.IsAvailable() {
			c.Next()
			return
		}

		ip := c.ClientIP()
		// 限流键：ratelimit:{prefix}:{ip}
		key := fmt.Sprintf("ratelimit:%s:%s", keyPrefix, ip)

		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		client := redispkg.GetClient()
		// INCR 原子自增
		count, err := client.Incr(ctx, key).Result()
		if err != nil {
			// Redis 操作异常，降级放行
			c.Next()
			return
		}

		// 首次访问时设置过期时间（固定窗口）
		if count == 1 {
			_ = client.Expire(ctx, key, time.Duration(windowSec)*time.Second).Err()
		}

		// 超过阈值，拒绝
		if count > int64(maxCount) {
			// 剩余等待时间（窗口剩余 TTL）
			ttl, _ := client.TTL(ctx, key).Result()
			if ttl < 0 {
				ttl = time.Duration(windowSec) * time.Second
			}
			c.Header("Retry-After", strconv.Itoa(int(ttl.Seconds())))
			c.Header("X-RateLimit-Limit", strconv.Itoa(maxCount))
			c.Header("X-RateLimit-Remaining", "0")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, response.Fail(utils.CodeTooManyRequests, "请求过于频繁，请稍后再试"))
			return
		}

		// 设置响应头，暴露限流信息
		c.Header("X-RateLimit-Limit", strconv.Itoa(maxCount))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(maxCount-int(count)))

		c.Next()
	}
}
