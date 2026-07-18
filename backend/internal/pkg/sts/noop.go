// noop.go STS NoopProvider：未配置时的降级实现
package sts

import "context"

// NoopProvider 空实现，AssumeRole 永远返回 ErrNotConfigured。
// 用于开发环境或未配置 STS 时保持 file 模块可用（前端走普通上传/预签名）。
type NoopProvider struct{}

// AssumeRole 返回 ErrNotConfigured
func (p *NoopProvider) AssumeRole(_ context.Context) (*STSCredentials, error) {
	return nil, ErrNotConfigured
}

// IsAvailable 永远 false
func (p *NoopProvider) IsAvailable() bool { return false }

// 编译期断言
var _ Provider = (*NoopProvider)(nil)
