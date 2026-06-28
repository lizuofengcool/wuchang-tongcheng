package storage

import (
	"strings"
	"testing"

	"wuchang-tongcheng/internal/pkg/config"
)

// TestNewQiniuStorage_RejectsInvalidConfig 验证 key 未配置/占位值/缺少必填项时
// NewQiniuStorage 返回错误（由 Init 据此降级到 local）。
func TestNewQiniuStorage_RejectsInvalidConfig(t *testing.T) {
	cases := []struct {
		name string
		cfg  *config.StorageConfig
		want string
	}{
		{
			name: "nil config",
			cfg:  nil,
			want: "config is nil",
		},
		{
			name: "空 AccessKey",
			cfg:  &config.StorageConfig{Type: "qiniu", SecretKey: "sk", Bucket: "bkt", Domain: "https://cdn.example.com"},
			want: "access_key not configured",
		},
		{
			name: "占位 AccessKey",
			cfg:  &config.StorageConfig{Type: "qiniu", AccessKey: "your-access-key", SecretKey: "sk", Bucket: "bkt", Domain: "https://cdn.example.com"},
			want: "access_key not configured",
		},
		{
			name: "占位 AccessKey 别名",
			cfg:  &config.StorageConfig{Type: "qiniu", AccessKey: "your-qiniu-ak", SecretKey: "sk", Bucket: "bkt", Domain: "https://cdn.example.com"},
			want: "access_key not configured",
		},
		{
			name: "空 SecretKey",
			cfg:  &config.StorageConfig{Type: "qiniu", AccessKey: "ak", Bucket: "bkt", Domain: "https://cdn.example.com"},
			want: "secret_key not configured",
		},
		{
			name: "占位 SecretKey",
			cfg:  &config.StorageConfig{Type: "qiniu", AccessKey: "ak", SecretKey: "your-secret-key", Bucket: "bkt", Domain: "https://cdn.example.com"},
			want: "secret_key not configured",
		},
		{
			name: "空 Bucket",
			cfg:  &config.StorageConfig{Type: "qiniu", AccessKey: "ak", SecretKey: "sk", Domain: "https://cdn.example.com"},
			want: "bucket name is required",
		},
		{
			name: "空 Domain",
			cfg:  &config.StorageConfig{Type: "qiniu", AccessKey: "ak", SecretKey: "sk", Bucket: "bkt"},
			want: "domain is required",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s, err := NewQiniuStorage(c.cfg)
			if err == nil {
				t.Errorf("期望返回错误，got nil（storage=%T）", s)
				return
			}
			if !strings.Contains(err.Error(), c.want) {
				t.Errorf("错误信息应包含 %q，got %q", c.want, err.Error())
			}
			if s != nil {
				t.Errorf("失败时应返回 nil storage，got %T", s)
			}
		})
	}
}

// TestNewQiniuStorage_ValidConfig 验证合法配置时返回非 nil 的 QiniuStorage 实例
// （不发起真实网络请求，仅校验对象构造）。
func TestNewQiniuStorage_ValidConfig(t *testing.T) {
	cfg := &config.StorageConfig{
		Type:      "qiniu",
		AccessKey: "fake-ak-1234567890",
		SecretKey: "fake-sk-0987654321",
		Bucket:    "wuchang-test",
		Domain:    "https://cdn.example.com/",
	}
	s, err := NewQiniuStorage(cfg)
	if err != nil {
		t.Fatalf("合法配置不应返回错误，got %v", err)
	}
	if s == nil {
		t.Fatal("合法配置应返回非 nil storage")
	}
	if s.bucket != "wuchang-test" {
		t.Errorf("bucket 字段错误，got %q", s.bucket)
	}
	// domain 应去除末尾斜杠
	if s.domain != "https://cdn.example.com" {
		t.Errorf("domain 字段应去除末尾斜杠，got %q", s.domain)
	}
	// HTTPS 域名应启用 UseHTTPS
	if !s.cfg.UseHTTPS {
		t.Error("HTTPS 域名应启用 UseHTTPS")
	}
	// 依赖对象均应初始化
	if s.mac == nil {
		t.Error("mac 凭据未初始化")
	}
	if s.bucketManager == nil {
		t.Error("bucketManager 未初始化")
	}
}

// TestInit_QiniuFallbackToLocal 验证 Init 在 qiniu 凭据未配置时降级到 LocalStorage。
func TestInit_QiniuFallbackToLocal(t *testing.T) {
	// 备份并重置全局 storage
	origStorage := storage
	storage = nil
	defer func() { storage = origStorage }()

	err := Init(&config.StorageConfig{
		Type:   "qiniu",
		Bucket: "", // 缺少凭据 → 降级
	})
	if err != nil {
		t.Fatalf("降级路径 Init 不应返回错误，got %v", err)
	}
	got := GetStorage()
	if got == nil {
		t.Fatal("降级后 GetStorage 不应返回 nil")
	}
	if _, ok := got.(*LocalStorage); !ok {
		t.Errorf("降级后应为 *LocalStorage，got %T", got)
	}
}

// TestInit_QiniuValid 验证 Init 在 qiniu 凭据合法时初始化为 QiniuStorage。
func TestInit_QiniuValid(t *testing.T) {
	origStorage := storage
	storage = nil
	defer func() { storage = origStorage }()

	err := Init(&config.StorageConfig{
		Type:      "qiniu",
		AccessKey: "fake-ak-1234567890",
		SecretKey: "fake-sk-0987654321",
		Bucket:    "wuchang-test",
		Domain:    "https://cdn.example.com",
	})
	if err != nil {
		t.Fatalf("合法配置 Init 不应返回错误，got %v", err)
	}
	got := GetStorage()
	if got == nil {
		t.Fatal("GetStorage 不应返回 nil")
	}
	if _, ok := got.(*QiniuStorage); !ok {
		t.Errorf("合法配置应为 *QiniuStorage，got %T", got)
	}
}

// TestQiniuStorage_Delete_InvalidURL 验证 Delete 在 URL 不匹配 domain 前缀时返回错误。
func TestQiniuStorage_Delete_InvalidURL(t *testing.T) {
	s := &QiniuStorage{
		domain: "https://cdn.example.com",
	}
	// 完全不匹配 domain
	if err := s.Delete("https://other.example.com/2026/06/x.jpg"); err == nil {
		t.Error("不匹配 domain 前缀的 URL 应返回错误")
	}
	// 仅 domain 无 objectName
	if err := s.Delete("https://cdn.example.com/"); err == nil {
		t.Error("空 objectName 应返回错误")
	}
}

// TestQiniuStorage_Delete_EmptyObject 验证 Delete 在 objectName 为空时返回错误。
func TestQiniuStorage_Delete_EmptyObject(t *testing.T) {
	s := &QiniuStorage{
		domain: "https://cdn.example.com",
	}
	// domain 后面只有一个斜杠，objectName 为空
	if err := s.Delete("https://cdn.example.com/"); err == nil {
		t.Error("空 objectName 应返回错误")
	}
}

// TestQiniuPlaceholderKeys 验证占位 key 集合覆盖常见样本。
func TestQiniuPlaceholderKeys(t *testing.T) {
	for k := range qiniuPlaceholderKeys {
		if !qiniuPlaceholderKeys[k] {
			t.Errorf("占位 key %q 应为 true", k)
		}
	}
	// 真实 key 不应在占位集合中
	if qiniuPlaceholderKeys["real-ak-abc123"] {
		t.Error("真实 key 不应在占位集合中")
	}
}
