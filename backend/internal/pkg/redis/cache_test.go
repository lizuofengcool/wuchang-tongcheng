package redis

import (
	"context"
	"testing"
	"time"
)

// TestCacheDegradationWhenUnavailable 验证 Redis 未初始化时缓存助手优雅降级：
// GetJSON 返回 miss（hit=false, err=nil），SetJSON/DelByPrefix 不 panic 不报错。
// 这是业务缓存全链路降级的核心保证（CI 等无 Redis 环境下走 DB）。
func TestCacheDegradationWhenUnavailable(t *testing.T) {
	if IsAvailable() {
		t.Skip("Redis 已初始化，跳过降级路径测试")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// GetJSON 不可用时返回 miss
	var dst struct{ V int }
	hit, err := GetJSON(ctx, "any:key", &dst)
	if hit {
		t.Errorf("Redis 不可用时 GetJSON 不应命中，got hit=true")
	}
	if err != nil {
		t.Errorf("Redis 不可用时 GetJSON 不应返回错误，got err=%v", err)
	}

	// SetJSON 不可用时 no-op
	if err := SetJSON(ctx, "any:key", "value", time.Minute); err != nil {
		t.Errorf("Redis 不可用时 SetJSON 不应返回错误，got err=%v", err)
	}

	// DelByPrefix 不可用时 no-op
	if err := DelByPrefix(ctx, "any:prefix"); err != nil {
		t.Errorf("Redis 不可用时 DelByPrefix 不应返回错误，got err=%v", err)
	}
}
