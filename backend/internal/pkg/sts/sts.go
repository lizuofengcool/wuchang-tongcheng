// Package sts 阿里云 OSS STS 临时凭据直传
//
// 通过阿里云 RAM STS AssumeRole API 申请短期访问凭据（AccessKeyID/AccessKeySecret/
// SecurityToken/Expiration），前端拿到后用 OSS 浏览器 SDK 或 SigV4 + x-amz-security-token
// 头直接 PUT 到 OSS，绕过后端带宽且凭据可复用上传多对象（适合批量/大文件场景）。
//
// 与预签名直传（pkg/storage PresignPut，D24）互补：
//   - 预签名：单个对象一次性 URL，无外部网络调用（本地 SigV4 签名），适合单文件偶发上传
//   - STS：一组临时 AK/SK 可在有效期内上传任意多对象，需调用 STS API（网络请求）
//
// 降级策略与项目其它第三方集成一致（amap/sms/storage）：
//   - provider=aliyun 且 AccessKey/SecretKey/RoleArn 齐全且非占位值 → AliyunProvider
//   - 任一缺失或占位（your-）→ NoopProvider，AssumeRole 返回 ErrNotConfigured，
//     file 模块据此回 501 并提示前端使用普通上传或预签名直传
package sts

import (
	"context"
	"errors"

	"wuchang-tongcheng/internal/pkg/config"
)

// ErrNotConfigured STS 未配置或配置不全，调用方应回退普通上传或预签名直传。
var ErrNotConfigured = errors.New("sts: not configured (provider=aliyun requires access_key/secret_key/role_arn)")

// Credentials STS 下发的临时访问凭据
type Credentials struct {
	AccessKeyID     string `json:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret"`
	SecurityToken   string `json:"security_token"`
	Expiration      string `json:"expiration"` // RFC3339（阿里云返回 ISO8601，原样透传）
}

// STSCredentials 临时凭据 + OSS 落地信息（前端构造 OSS 客户端所需）
type STSCredentials struct {
	Credentials
	Bucket       string `json:"bucket"`        // OSS 桶名
	Region       string `json:"region"`        // OSS 区域，如 oss-cn-hangzhou
	Endpoint     string `json:"endpoint"`      // OSS 端点，如 https://oss-cn-hangzhou.aliyuncs.com
	ObjectPrefix string `json:"object_prefix"` // 对象 key 前缀，如 uploads/2026/07/
	ExpiresIn    int    `json:"expires_in"`    // 凭据剩余有效期（秒）
}

// Provider STS 凭据提供者接口
type Provider interface {
	// AssumeRole 申请临时凭据
	AssumeRole(ctx context.Context) (*STSCredentials, error)
	// IsAvailable 配置是否齐全（供 Init 选择 provider 与测试使用）
	IsAvailable() bool
}

// 全局 provider 实例
var provider Provider

// Init 按配置初始化 STS provider
//
//	cfg == nil 或 provider 非 aliyun → NoopProvider（AssumeRole 返回 ErrNotConfigured）
//	provider=aliyun 但配置缺失/占位 → 同样降级 NoopProvider
func Init(cfg *config.STSConfig) {
	provider = resolveProvider(cfg)
}

// Get 获取全局 provider（未初始化时返回 NoopProvider，避免空指针）
func Get() Provider {
	if provider == nil {
		return &NoopProvider{}
	}
	return provider
}

// IsAvailable 全局 provider 是否可用
func IsAvailable() bool {
	return Get().IsAvailable()
}

// resolveProvider 按配置解析 provider（与 pkg/sms.resolveProvider 风格一致）
func resolveProvider(cfg *config.STSConfig) Provider {
	if cfg == nil {
		return &NoopProvider{}
	}
	switch cfg.Provider {
	case "aliyun":
		p := NewAliyunProvider(cfg)
		if !p.IsAvailable() {
			return &NoopProvider{}
		}
		return p
	default:
		return &NoopProvider{}
	}
}
