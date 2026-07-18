// Package sms 验证码存储实现
//
// CodeStore 接口两种实现：
//   - redisCodeStore：Redis 可用时使用，JSON 序列化 codeEntry，TTL 由 Redis 管理
//   - memoryCodeStore：Redis 不可用时降级，sync.Map + 惰性过期清理
package sms

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	redispkg "wuchang-tongcheng/internal/pkg/redis"

	"github.com/redis/go-redis/v9"
)

// codeEntry 验证码条目
type codeEntry struct {
	Code      string `json:"code"`
	Attempts  int    `json:"attempts"`
	ExpiresAt int64  `json:"expires_at"` // unix nano
}

// CodeStore 验证码存储接口
type CodeStore interface {
	// Set 写入验证码（覆盖已有），重置尝试次数，TTL 由 ttl 指定
	Set(ctx context.Context, phone, code string, ttl time.Duration) error
	// Verify 校验验证码：成功删除（一次性）；失败累计 attempts，超 maxAttempts 删除并返回 ErrTooManyAttempts
	Verify(ctx context.Context, phone, code string, maxAttempts int) error
}

// NewCodeStore 工厂：Redis 可用 → redisCodeStore；否则 → memoryCodeStore
func NewCodeStore() CodeStore {
	if redispkg.IsAvailable() {
		return &redisCodeStore{client: redispkg.GetClient()}
	}
	return newMemoryCodeStore()
}

const smsKeyPrefix = "sms:code:"

func smsKey(phone string) string { return smsKeyPrefix + phone }

// ===== Redis 实现 =====

type redisCodeStore struct {
	client *redis.Client
}

// Set 写入验证码条目，TTL 交由 Redis 管理
func (s *redisCodeStore) Set(ctx context.Context, phone, code string, ttl time.Duration) error {
	entry := codeEntry{Code: code, Attempts: 0, ExpiresAt: time.Now().Add(ttl).UnixNano()}
	raw, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	return s.client.Set(ctx, smsKey(phone), raw, ttl).Err()
}

// Verify 校验，成功删除；失败累加 attempts 写回（保留剩余 TTL）
func (s *redisCodeStore) Verify(ctx context.Context, phone, code string, maxAttempts int) error {
	key := smsKey(phone)
	raw, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return ErrCodeNotFound
		}
		return err
	}
	var entry codeEntry
	if err := json.Unmarshal([]byte(raw), &entry); err != nil {
		// 数据损坏，清理后按不存在处理
		_ = s.client.Del(ctx, key).Err()
		return ErrCodeNotFound
	}
	if time.Now().UnixNano() >= entry.ExpiresAt {
		_ = s.client.Del(ctx, key).Err()
		return ErrCodeNotFound
	}
	if entry.Attempts >= maxAttempts {
		_ = s.client.Del(ctx, key).Err()
		return ErrTooManyAttempts
	}
	if entry.Code == code {
		_ = s.client.Del(ctx, key).Err()
		return nil
	}
	// 错误：累加尝试次数，写回（保留剩余 TTL）
	entry.Attempts++
	raw2, _ := json.Marshal(entry)
	ttl, err := s.client.TTL(ctx, key).Result()
	if err != nil || ttl <= 0 {
		ttl = 60 * time.Second
	}
	_ = s.client.Set(ctx, key, raw2, ttl).Err()
	return ErrCodeInvalid
}

// ===== 内存实现（Redis 不可用时降级）=====

type memoryCodeStore struct {
	mu      sync.Mutex
	entries map[string]codeEntry
}

func newMemoryCodeStore() *memoryCodeStore {
	return &memoryCodeStore{entries: make(map[string]codeEntry)}
}

// Set 写入验证码条目（覆盖已有）
func (s *memoryCodeStore) Set(ctx context.Context, phone, code string, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[smsKey(phone)] = codeEntry{Code: code, Attempts: 0, ExpiresAt: time.Now().Add(ttl).UnixNano()}
	return nil
}

// Verify 校验，惰性清理过期条目
func (s *memoryCodeStore) Verify(ctx context.Context, phone, code string, maxAttempts int) error {
	key := smsKey(phone)
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.entries[key]
	if !ok || time.Now().UnixNano() >= entry.ExpiresAt {
		delete(s.entries, key)
		return ErrCodeNotFound
	}
	if entry.Attempts >= maxAttempts {
		delete(s.entries, key)
		return ErrTooManyAttempts
	}
	if entry.Code == code {
		delete(s.entries, key)
		return nil
	}
	entry.Attempts++
	s.entries[key] = entry
	return ErrCodeInvalid
}
