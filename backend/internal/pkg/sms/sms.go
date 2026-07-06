// Package sms 短信验证码服务
//
// 提供：
//   - Provider 接口（mock/dev 不实际发短信；aliyun 预留，AK/SK 未配置降级 mock）
//   - CodeStore 验证码存储（Redis 优先，Redis 不可用时降级内存）
//   - Service 验证码生成（crypto/rand）+ 发送 + 校验（一次性 + 尝试次数限制）
//
// 设计要点：
//   - 验证码使用 crypto/rand 生成，避免可预测
//   - 校验成功后立即删除（一次性消费）
//   - 错误累计尝试次数，超 maxAttempts 后删除并拒绝（防暴力枚举）
//   - Redis 不可用全链路降级到内存存储，业务不中断
package sms

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"wuchang-tongcheng/internal/pkg/config"
	redispkg "wuchang-tongcheng/internal/pkg/redis"
)

// 验证码校验错误
var (
	ErrCodeNotFound    = errors.New("验证码不存在或已过期")
	ErrCodeInvalid     = errors.New("验证码错误")
	ErrTooManyAttempts = errors.New("验证码尝试次数过多，请重新获取")
)

// Provider 短信发送提供商接口
type Provider interface {
	// Send 发送验证码到指定手机号。mock/dev 模式不实际发送。
	Send(ctx context.Context, phone, code string) error
}

// NoopProvider mock/dev 用：不实际发送短信
type NoopProvider struct{}

// Send no-op，始终成功
func (NoopProvider) Send(ctx context.Context, phone, code string) error { return nil }

// Service 短信验证码服务
type Service struct {
	provider    Provider
	store       CodeStore
	codeLength  int
	codeTTL     time.Duration
	maxAttempts int
	devReturn   bool // 是否在发送响应里返回验证码明文（仅 mock 联调用）
	// rand 替换点，便于测试确定性地生成验证码
	rand func() (int, error)
}

// NewService 根据配置创建短信验证码服务
//
// 配置 provider 为空或 "mock" → NoopProvider
// provider="aliyun" 但 AK/SK 未配置齐全 → 降级 NoopProvider（真实阿里云 SDK 待补齐）
// Redis 可用 → redisCodeStore；否则 → memoryCodeStore
func NewService(cfg *config.SMSConfig) *Service {
	if cfg == nil {
		cfg = &config.SMSConfig{}
	}
	codeLength := cfg.CodeLength
	if codeLength == 0 {
		codeLength = 6
	}
	codeTTL := cfg.CodeTTL
	if codeTTL == 0 {
		codeTTL = 300
	}
	maxAttempts := cfg.MaxAttempts
	if maxAttempts == 0 {
		maxAttempts = 5
	}

	ttl := time.Duration(codeTTL) * time.Second
	s := &Service{
		provider:    resolveProvider(cfg),
		store:       NewCodeStore(),
		codeLength:  codeLength,
		codeTTL:     ttl,
		maxAttempts: maxAttempts,
		devReturn:   cfg.DevReturnCode,
	}
	s.rand = s.defaultRand
	return s
}

// resolveProvider 按 config 解析短信提供商，未配置/占位值统一降级 NoopProvider
func resolveProvider(cfg *config.SMSConfig) Provider {
	switch strings.ToLower(cfg.Provider) {
	case "aliyun":
		// 阿里云短信 SDK 待补齐；AK/SK 缺失或占位时降级 Noop，避免引入未集成依赖
		if cfg.AccessKey == "" || cfg.SecretKey == "" ||
			strings.Contains(cfg.AccessKey, "your-") || strings.Contains(cfg.SecretKey, "your-") {
			return NoopProvider{}
		}
		return NoopProvider{}
	default: // "" 或 "mock"
		return NoopProvider{}
	}
}

// defaultRand 使用 crypto/rand 生成 [0, 10^codeLength) 的随机数
func (s *Service) defaultRand() (int, error) {
	max := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(s.codeLength)), nil)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return 0, err
	}
	return int(n.Int64()), nil
}

// SendCode 生成并发送验证码。
// devReturn=true（仅 mock provider 联调）时返回验证码明文，否则返回空串。
func (s *Service) SendCode(ctx context.Context, phone string) (string, error) {
	if phone == "" {
		return "", errors.New("手机号不能为空")
	}
	n, err := s.rand()
	if err != nil {
		return "", err
	}
	code := fmt.Sprintf("%0*d", s.codeLength, n)
	if err := s.store.Set(ctx, phone, code, s.codeTTL); err != nil {
		return "", err
	}
	if err := s.provider.Send(ctx, phone, code); err != nil {
		return "", err
	}
	if s.devReturn {
		return code, nil
	}
	return "", nil
}

// Verify 校验验证码。成功后立即删除（一次性）；错误累计尝试次数，超限后删除并返回 ErrTooManyAttempts。
func (s *Service) Verify(ctx context.Context, phone, code string) error {
	if phone == "" || code == "" {
		return ErrCodeInvalid
	}
	return s.store.Verify(ctx, phone, code, s.maxAttempts)
}

// 编译期断言：redispkg 仅用于 NewCodeStore 内部可用性判断，避免被误删导入
var _ = redispkg.IsAvailable
