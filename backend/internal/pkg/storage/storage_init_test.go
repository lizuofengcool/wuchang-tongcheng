// Package storage 对象存储 Init 工厂分支测试
// 覆盖 local/默认/minio 失败/不支持类型/GetStorage 兜底等路径。
// qiniu 分支已在 qiniu_test.go 覆盖。
package storage

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wuchang-tongcheng/internal/pkg/config"
)

// backupStorage 备份并重置全局 storage，返回恢复函数
func backupStorage(t *testing.T) func() {
	t.Helper()
	orig := storage
	storage = nil
	return func() { storage = orig }
}

// TestInit_LocalExplicit 显式 type=local 成功
func TestInit_LocalExplicit(t *testing.T) {
	restore := backupStorage(t)
	defer restore()

	err := Init(&config.StorageConfig{
		Type:   "local",
		Domain: "http://localhost:8080",
	})
	require.NoError(t, err)
	got := GetStorage()
	require.NotNil(t, got)
	_, ok := got.(*LocalStorage)
	assert.True(t, ok, "type=local 应返回 *LocalStorage")
}

// TestInit_EmptyTypeDefaultsToLocal 空 type 默认走 local
func TestInit_EmptyTypeDefaultsToLocal(t *testing.T) {
	restore := backupStorage(t)
	defer restore()

	err := Init(&config.StorageConfig{
		Type: "",
	})
	require.NoError(t, err)
	got := GetStorage()
	require.NotNil(t, got)
	_, ok := got.(*LocalStorage)
	assert.True(t, ok, "空 type 应默认走 local")
}

// TestInit_UnsupportedType 不支持的类型返回错误
func TestInit_UnsupportedType(t *testing.T) {
	restore := backupStorage(t)
	defer restore()

	err := Init(&config.StorageConfig{
		Type: "azure-blob-unsupported",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported storage type")
	assert.Contains(t, err.Error(), "azure-blob-unsupported")
}

// TestInit_MinioInvalidEndpoint minio endpoint 非法时返回错误（不发起网络请求）
func TestInit_MinioInvalidEndpoint(t *testing.T) {
	restore := backupStorage(t)
	defer restore()

	// 空 domain：NewMinIOStorage 返回 "minio endpoint(domain) is required"
	err := Init(&config.StorageConfig{
		Type:   "minio",
		Domain: "",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "minio endpoint")
}

// TestInit_MinioInvalidHost host 解析失败
func TestInit_MinioInvalidHost(t *testing.T) {
	restore := backupStorage(t)
	defer restore()

	// 缺少 host 部分的 URL，触发 "invalid minio endpoint host"
	err := Init(&config.StorageConfig{
		Type:   "minio",
		Domain: "http://",
	})
	require.Error(t, err)
}

// TestGetStorage_NilFallback storage 全局为 nil 时兜底返回 LocalStorage
func TestGetStorage_NilFallback(t *testing.T) {
	restore := backupStorage(t)
	defer restore()

	// 此时 storage == nil
	got := GetStorage()
	require.NotNil(t, got, "GetStorage 不应返回 nil")
	_, ok := got.(*LocalStorage)
	assert.True(t, ok, "nil 时应兜底返回 LocalStorage")
}

// TestInit_QiniuNilConfig qiniu 类型 + nil config（Init 不应 panic）
// Init 直接调用 NewQiniuStorage(cfg)，cfg 可能为 nil 但 config 必填字段为空，
// 实际 storage.Init 收到的 cfg 来自 viper，不会 nil，但接口应安全。
func TestInit_QiniuNilConfig(t *testing.T) {
	restore := backupStorage(t)
	defer restore()

	// cfg 各字段为空，应触发 qiniu 降级到 local
	err := Init(&config.StorageConfig{
		Type: "qiniu",
	})
	require.NoError(t, err, "qiniu 占位/空配置应降级到 local，不报错")
	got := GetStorage()
	_, ok := got.(*LocalStorage)
	assert.True(t, ok, "降级后应为 LocalStorage")
}

// TestInit_LocalWithCustomBucket 自定义 bucket 路径生效
func TestInit_LocalWithCustomBucket(t *testing.T) {
	restore := backupStorage(t)
	defer restore()
	defer func() {
		// 清理临时目录
		_ = os.RemoveAll("./test-uploads-local")
	}()

	err := Init(&config.StorageConfig{
		Type:   "local",
		Bucket: "./test-uploads-local",
		Domain: "http://example.com",
	})
	require.NoError(t, err)
	got, ok := GetStorage().(*LocalStorage)
	require.True(t, ok)
	assert.Equal(t, "http://example.com", got.domain)
	assert.Equal(t, "./test-uploads-local", got.basePath)
	assert.Equal(t, "/uploads", got.urlPrefix)
}
