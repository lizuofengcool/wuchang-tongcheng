// Package middleware 模块开关拦截中间件
//
// 从 URL 路径 /api/v1/:module/... 提取模块名，查询 modules 表判断模块是否启用。
// 已禁用模块的请求直接返回 403，不进入业务 handler。
//
// 缓存策略：
//   - Redis 缓存模块开关状态，key: module:enabled:{name}，TTL 60s
//   - 缓存值为 moduleStatus 的 JSON（含 DisplayName，用于错误消息）
//   - 启停模块时由 module 服务调用 redis.Del 失效缓存（单机版）
//   - Redis 不可用时降级为每次查 DB
//
// 白名单路径（系统基础模块，始终放行）：
//   /api/v1/auth、/api/v1/modules、/api/v1/region、/api/v1/setting、
//   /api/v1/file、/api/v1/permission、/api/v1/dashboard
package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"wuchang-tongcheng/internal/core/response"
	redispkg "wuchang-tongcheng/internal/pkg/redis"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 模块开关缓存键前缀与 TTL
// 注意：与 module 包 service.go 中的常量保持一致（避免循环依赖未共享定义）
const (
	moduleCheckCacheKeyPrefix = "module:enabled:"
	moduleCheckCacheTTL       = 60 * time.Second
)

// moduleCheckCacheKey 构造模块开关缓存的键
func moduleCheckCacheKey(name string) string {
	return moduleCheckCacheKeyPrefix + name
}

// 模块开关白名单：系统基础模块始终放行，不受开关拦截
// 这些模块是平台运行的基础设施，禁用会导致管理后台无法访问
var moduleCheckWhitelist = map[string]struct{}{
	"auth":       {},
	"modules":    {},
	"region":     {},
	"setting":    {},
	"file":       {},
	"permission": {},
	"dashboard":  {},
}

// moduleStatus DB 查询与缓存用的最小字段集
// 避免依赖 module 包，防止循环依赖
type moduleStatus struct {
	DisplayName string
	Enabled     bool
}

// ModuleCheck 模块开关拦截中间件
// db 参数用于查询 modules 表；nil 时不拦截（测试场景）
func ModuleCheck(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db == nil {
			c.Next()
			return
		}

		// 从 URL 路径提取模块名：/api/v1/:module/...
		moduleName := extractModuleName(c.Request.URL.Path)
		if moduleName == "" {
			// 非 /api/v1/:module/... 路径（如 /health、/ws、/swagger），放行
			c.Next()
			return
		}

		// 白名单模块始终放行
		if _, ok := moduleCheckWhitelist[moduleName]; ok {
			c.Next()
			return
		}

		// 查询模块开关状态（带 Redis 缓存）
		status, err := getModuleStatus(c.Request.Context(), db, moduleName)
		if err != nil {
			// 查询失败不阻塞请求（降级放行，避免 DB 故障导致全站不可用）
			c.Next()
			return
		}

		if !status.Enabled {
			displayName := status.DisplayName
			if displayName == "" {
				displayName = moduleName
			}
			c.JSON(http.StatusOK, response.Fail(http.StatusForbidden, fmt.Sprintf("模块未开通：%s", displayName)))
			c.Abort()
			return
		}

		c.Next()
	}
}

// extractModuleName 从 URL 路径提取模块名
// 路径格式：/api/v1/:module/... 或 /api/v1/:module
// 非 /api/v1/ 前缀路径返回空串
func extractModuleName(path string) string {
	// 标准化路径：去除尾部斜杠
	path = strings.TrimSuffix(path, "/")
	if !strings.HasPrefix(path, "/api/v1/") {
		return ""
	}
	// 去除前缀 /api/v1/
	rest := strings.TrimPrefix(path, "/api/v1/")
	// 取第一段作为模块名
	if idx := strings.Index(rest, "/"); idx >= 0 {
		return rest[:idx]
	}
	return rest
}

// getModuleStatus 查询模块开关状态（先查 Redis 缓存，miss 则查 DB 并回填）
// 查询失败或记录不存在时返回 enabled=true（不阻塞未注册的模块）
func getModuleStatus(ctx context.Context, db *gorm.DB, name string) (moduleStatus, error) {
	// 1. 先查 Redis 缓存（JSON 序列化的 moduleStatus）
	var cached moduleStatus
	if hit, _ := redispkg.GetJSON(ctx, moduleCheckCacheKey(name), &cached); hit {
		return cached, nil
	}

	// 2. 查 DB
	status, err := fetchModuleStatusFromDB(db, name)
	if err != nil {
		return moduleStatus{Enabled: true}, err
	}

	// 3. 回填缓存（Redis 不可用时 no-op）
	_ = redispkg.SetJSON(ctx, moduleCheckCacheKey(name), status, moduleCheckCacheTTL)

	return status, nil
}

// fetchModuleStatusFromDB 从 DB 查询模块状态
// 记录不存在或表不存在时返回 enabled=true（不阻塞未注册的模块）
func fetchModuleStatusFromDB(db *gorm.DB, name string) (moduleStatus, error) {
	var status moduleStatus
	err := db.Table("modules").
		Select("display_name, enabled").
		Where("name = ?", name).
		Limit(1).
		First(&status).Error
	if err != nil {
		// 记录不存在或表不存在 → 视为启用
		return moduleStatus{Enabled: true}, nil
	}
	return status, nil
}
