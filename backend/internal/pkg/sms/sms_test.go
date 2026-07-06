package sms

import (
	"context"
	"errors"
	"testing"
	"time"

	"wuchang-tongcheng/internal/pkg/config"
)

// TestService_SendAndVerify 验证码发送 → 正确校验通过 → 二次校验拒绝（一次性）
func TestService_SendAndVerify(t *testing.T) {
	svc := NewService(&config.SMSConfig{
		Provider: "mock", CodeLength: 6, CodeTTL: 60, MaxAttempts: 3, DevReturnCode: true,
	})
	ctx := context.Background()

	code, err := svc.SendCode(ctx, "13800138000")
	if err != nil {
		t.Fatalf("SendCode: %v", err)
	}
	if len(code) != 6 {
		t.Fatalf("code length = %d, want 6", len(code))
	}

	// 错误验证码 → ErrCodeInvalid
	if err := svc.Verify(ctx, "13800138000", "000000"); !errors.Is(err, ErrCodeInvalid) {
		t.Fatalf("wrong code err = %v, want ErrCodeInvalid", err)
	}
	// 正确验证码 → 通过
	if err := svc.Verify(ctx, "13800138000", code); err != nil {
		t.Fatalf("verify correct: %v", err)
	}
	// 二次校验 → ErrCodeNotFound（已删除）
	if err := svc.Verify(ctx, "13800138000", code); !errors.Is(err, ErrCodeNotFound) {
		t.Fatalf("reuse err = %v, want ErrCodeNotFound", err)
	}
}

// TestService_DeterministicCode 注入 rand 桩，验证码格式化正确（前导零补齐）
func TestService_DeterministicCode(t *testing.T) {
	svc := NewService(&config.SMSConfig{
		Provider: "mock", CodeLength: 6, CodeTTL: 60, DevReturnCode: true,
	})
	svc.rand = func() (int, error) { return 42, nil }

	code, err := svc.SendCode(context.Background(), "13600136000")
	if err != nil {
		t.Fatalf("SendCode: %v", err)
	}
	if code != "000042" {
		t.Fatalf("code = %q, want 000042", code)
	}
	if err := svc.Verify(context.Background(), "13600136000", "000042"); err != nil {
		t.Fatalf("verify: %v", err)
	}
}

// TestService_MaxAttempts 错误次数达上限后，即使正确验证码也被拒绝
func TestService_MaxAttempts(t *testing.T) {
	svc := NewService(&config.SMSConfig{
		Provider: "mock", CodeLength: 4, CodeTTL: 60, MaxAttempts: 2, DevReturnCode: true,
	})
	svc.rand = func() (int, error) { return 1234, nil }
	ctx := context.Background()

	code, _ := svc.SendCode(ctx, "13900139000")
	// 两次错误尝试，attempts 累加到 2
	_ = svc.Verify(ctx, "13900139000", "0000") // attempts 0→1
	_ = svc.Verify(ctx, "13900139000", "0000") // attempts 1→2
	// 第三次：attempts(2) >= maxAttempts(2) → ErrTooManyAttempts（即使验证码正确）
	err := svc.Verify(ctx, "13900139000", code)
	if !errors.Is(err, ErrTooManyAttempts) {
		t.Fatalf("after max attempts err = %v, want ErrTooManyAttempts", err)
	}
}

// TestService_EmptyPhone 空手机号拒绝
func TestService_EmptyPhone(t *testing.T) {
	svc := NewService(nil)
	if _, err := svc.SendCode(context.Background(), ""); err == nil {
		t.Fatal("expected error for empty phone")
	}
	if err := svc.Verify(context.Background(), "", "1234"); !errors.Is(err, ErrCodeInvalid) {
		t.Fatalf("empty phone verify err = %v, want ErrCodeInvalid", err)
	}
}

// TestService_NilConfigDefaults nil 配置走默认值，devReturn=false 时返回空串
func TestService_NilConfigDefaults(t *testing.T) {
	svc := NewService(nil)
	svc.rand = func() (int, error) { return 7, nil }
	code, err := svc.SendCode(context.Background(), "13700137000")
	if err != nil {
		t.Fatalf("SendCode: %v", err)
	}
	if code != "" {
		t.Fatalf("expected empty code when devReturn=false, got %q", code)
	}
	// 默认 6 位
	if err := svc.Verify(context.Background(), "13700137000", "000007"); err != nil {
		t.Fatalf("verify with default 6-digit code: %v", err)
	}
}

// TestService_VerifyUnknownPhone 未发送验证码的手机号 → ErrCodeNotFound
func TestService_VerifyUnknownPhone(t *testing.T) {
	svc := NewService(nil)
	err := svc.Verify(context.Background(), "13100131000", "123456")
	if !errors.Is(err, ErrCodeNotFound) {
		t.Fatalf("unknown phone err = %v, want ErrCodeNotFound", err)
	}
}

// TestMemoryCodeStore_Expiry 内存存储过期后校验返回 ErrCodeNotFound
func TestMemoryCodeStore_Expiry(t *testing.T) {
	s := newMemoryCodeStore()
	ctx := context.Background()

	if err := s.Set(ctx, "13000130000", "1234", 50*time.Millisecond); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if err := s.Verify(ctx, "13000130000", "1234", 3); err != nil {
		t.Fatalf("verify before expiry: %v", err)
	}
	// 重新写入并等待过期
	if err := s.Set(ctx, "13000130000", "1234", 50*time.Millisecond); err != nil {
		t.Fatalf("Set: %v", err)
	}
	time.Sleep(80 * time.Millisecond)
	if err := s.Verify(ctx, "13000130000", "1234", 3); !errors.Is(err, ErrCodeNotFound) {
		t.Fatalf("expired err = %v, want ErrCodeNotFound", err)
	}
}

// TestResolveProvider_DegradesToNoop 各种未配置/占位情形均降级 NoopProvider
func TestResolveProvider_DegradesToNoop(t *testing.T) {
	cases := []*config.SMSConfig{
		{Provider: ""},
		{Provider: "mock"},
		{Provider: "MOCK"},
		{Provider: "aliyun"},                                                        // 无 AK/SK
		{Provider: "aliyun", AccessKey: "your-ak", SecretKey: "sk"},                 // AK 占位
		{Provider: "aliyun", AccessKey: "ak", SecretKey: "your-sk"},                 // SK 占位
		{Provider: "aliyun", AccessKey: "AKIDxxxx", SecretKey: "secretxxxx"},        // 缺 SignName/TemplateCode → Noop
	}
	for i, cfg := range cases {
		if _, ok := resolveProvider(cfg).(NoopProvider); !ok {
			t.Fatalf("case %d provider=%q: expected NoopProvider", i, cfg.Provider)
		}
	}
}
