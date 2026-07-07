// Package service STS 临时凭据业务逻辑单元测试。
//
// 覆盖 GetSTSCredentials 的降级路径：未配置 STS（全局 provider 为 NoopProvider）
// 时返回 sts.ErrNotConfigured，handler 据此回 501 提示回退普通上传/预签名。
// 不依赖 DB / 真实 STS API，可在无基础设施环境离线运行。
package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"wuchang-tongcheng/internal/pkg/sts"
)

// TestGetSTSCredentials_NotConfigured 未配置 STS 时返回 ErrNotConfigured
func TestGetSTSCredentials_NotConfigured(t *testing.T) {
	svc := NewFileService(nil)

	// 测试环境未调用 sts.Init，全局 provider 为 nil → Get() 返回 NoopProvider
	resp, err := svc.GetSTSCredentials(context.Background(), 1, 100)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, sts.ErrNotConfigured)
}
