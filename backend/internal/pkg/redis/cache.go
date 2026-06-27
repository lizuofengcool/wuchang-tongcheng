// Package redis 缓存助手
// 提供基于 JSON 序列化的 cache-aside 语义，Redis 不可用时优雅降级为 miss
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// GetJSON 按 key 取缓存并反序列化到 dst。
// 语义：Redis 不可用 / key 不存在 / 反序列化失败 → hit=false, err=nil（降级为 miss）
// 调用方应在 hit=false 时回源 DB。
func GetJSON(ctx context.Context, key string, dst interface{}) (bool, error) {
	if !IsAvailable() {
		return false, nil
	}
	raw, err := client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil // key 不存在，正常 miss
		}
		return false, nil // 其他错误也降级为 miss，不阻断业务
	}
	if uerr := json.Unmarshal([]byte(raw), dst); uerr != nil {
		return false, nil // 反序列化失败，按 miss 处理
	}
	return true, nil
}

// SetJSON 序列化 value 并写入缓存，带过期时间。
// 语义：Redis 不可用时 no-op（不报错），写缓存失败不阻断业务。
func SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if !IsAvailable() {
		return nil
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal cache value failed: %w", err)
	}
	return client.Set(ctx, key, raw, expiration).Err()
}

// DelByPrefix 按前缀批量删除缓存（SCAN + DEL，避免 KEYS 阻塞）。
// 语义：Redis 不可用时 no-op。用于写操作失效整组缓存。
func DelByPrefix(ctx context.Context, prefix string) error {
	if !IsAvailable() {
		return nil
	}
	var cursor uint64
	for {
		keys, next, err := client.Scan(ctx, cursor, prefix+"*", 100).Result()
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			if err := client.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}
	return nil
}
