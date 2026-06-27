// Package redis Redis封装
// 基于go-redis的Redis连接封装，支持连接池
package redis

import (
	"context"
	"fmt"
	"time"

	"wuchang-tongcheng/internal/pkg/config"

	"github.com/redis/go-redis/v9"
)

var (
	client *redis.Client
)

// Init 初始化Redis连接
func Init(cfg *config.RedisConfig) error {
	client = redis.NewClient(&redis.Options{
		Addr:     cfg.GetAddr(),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("connect redis failed: %w", err)
	}

	return nil
}

// GetClient 获取Redis客户端
func GetClient() *redis.Client {
	if client == nil {
		panic("redis not initialized, call Init() first")
	}
	return client
}

// IsAvailable 检查 Redis 是否已初始化可用
// 用于限流等可选功能在 Redis 不可用时优雅降级
func IsAvailable() bool {
	return client != nil
}

// Close 关闭Redis连接
func Close() error {
	if client == nil {
		return nil
	}
	return client.Close()
}

// Get 获取缓存
func Get(ctx context.Context, key string) (string, error) {
	return client.Get(ctx, key).Result()
}

// Set 设置缓存
func Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return client.Set(ctx, key, value, expiration).Err()
}

// Del 删除缓存
func Del(ctx context.Context, keys ...string) error {
	return client.Del(ctx, keys...).Err()
}

// Exists 判断key是否存在
func Exists(ctx context.Context, key string) (bool, error) {
	result, err := client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// Incr 自增
func Incr(ctx context.Context, key string) (int64, error) {
	return client.Incr(ctx, key).Result()
}

// Decr 自减
func Decr(ctx context.Context, key string) (int64, error) {
	return client.Decr(ctx, key).Result()
}

// Expire 设置过期时间
func Expire(ctx context.Context, key string, expiration time.Duration) error {
	return client.Expire(ctx, key, expiration).Err()
}

// TTL 获取剩余过期时间
func TTL(ctx context.Context, key string) (time.Duration, error) {
	return client.TTL(ctx, key).Result()
}

// HGet Hash获取
func HGet(ctx context.Context, key, field string) (string, error) {
	return client.HGet(ctx, key, field).Result()
}

// HSet Hash设置
func HSet(ctx context.Context, key string, values ...interface{}) error {
	return client.HSet(ctx, key, values...).Err()
}

// HGetAll 获取所有Hash字段
func HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return client.HGetAll(ctx, key).Result()
}

// HDel 删除Hash字段
func HDel(ctx context.Context, key string, fields ...string) error {
	return client.HDel(ctx, key, fields...).Err()
}

// LPush List左推入
func LPush(ctx context.Context, key string, values ...interface{}) error {
	return client.LPush(ctx, key, values...).Err()
}

// RPop List右弹出
func RPop(ctx context.Context, key string) (string, error) {
	return client.RPop(ctx, key).Result()
}

// LLen List长度
func LLen(ctx context.Context, key string) (int64, error) {
	return client.LLen(ctx, key).Result()
}

// SAdd Set添加
func SAdd(ctx context.Context, key string, members ...interface{}) error {
	return client.SAdd(ctx, key, members...).Err()
}

// SMembers 获取所有Set成员
func SMembers(ctx context.Context, key string) ([]string, error) {
	return client.SMembers(ctx, key).Result()
}

// SIsMember 判断是否是Set成员
func SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return client.SIsMember(ctx, key, member).Result()
}

// SRem 移除Set成员
func SRem(ctx context.Context, key string, members ...interface{}) error {
	return client.SRem(ctx, key, members...).Err()
}

// ZAdd 有序集合添加
func ZAdd(ctx context.Context, key string, score float64, member interface{}) error {
	z := redis.Z{
		Score:  score,
		Member: member,
	}
	return client.ZAdd(ctx, key, z).Err()
}

// ZRange 有序集合范围查询
func ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return client.ZRange(ctx, key, start, stop).Result()
}

// ZRevRange 有序集合倒序范围查询
func ZRevRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return client.ZRevRange(ctx, key, start, stop).Result()
}

// ZScore 获取有序集合成员分数
func ZScore(ctx context.Context, key string, member string) (float64, error) {
	return client.ZScore(ctx, key, member).Result()
}
