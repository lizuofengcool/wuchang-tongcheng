// Package sts 阿里云 OSS STS 临时凭据单元测试。
//
// 全部用例离线运行（httptest.Server 模拟 STS API），无外部依赖。
// 覆盖：percentEncode/canonicalQueryString 签名原语、IsAvailable 配置校验、
// resolveProvider 降级路径、NoopProvider、normalizeObjectPrefix、
// AssumeRole 成功/阿里云业务错误/网络错误/非 JSON/空凭据/签名一致性、
// DurationSeconds 钳制、RoleSessionName 默认值。
package sts

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wuchang-tongcheng/internal/pkg/config"
)

// ----- 签名原语 -----

// TestPercentEncode RFC3986 编码（与 pkg/sms 行为一致）
func TestPercentEncode(t *testing.T) {
	cases := map[string]string{
		"":       "",
		"abc123": "abc123",
		"abc-_.~": "abc-_.~", // 不编码字符
		" ":      "%20",      // 空格 → %20（非 +）
		"中":      "%E4%B8%AD",
		"a b":    "a%20b",
		"=/&":    "%3D%2F%26",
		"杭州":     "%E6%9D%AD%E5%B7%9E",
	}
	for in, want := range cases {
		assert.Equal(t, want, percentEncode(in), "percentEncode(%q)", in)
	}
}

// TestCanonicalQueryString 按 key 字典序排列且 k/v 均编码
func TestCanonicalQueryString(t *testing.T) {
	params := map[string]string{
		"Action":  "AssumeRole",
		"Version": "2015-04-01",
		"RoleArn": "acs:ram::123:role/test",
	}
	got := canonicalQueryString(params)
	// 期望按字典序：Action, RoleArn, Version
	want := "Action=AssumeRole&RoleArn=acs%3Aram%3A%3A123%3Arole%2Ftest&Version=2015-04-01"
	assert.Equal(t, want, got)
}

// TestSignatureDeterminism 相同参数（含固定 nonce/timestamp）签名一致
func TestSignatureDeterminism(t *testing.T) {
	p := &AliyunProvider{accessKey: "ak", secretKey: "sk"}
	params := map[string]string{
		"Action":           "AssumeRole",
		"Version":          "2015-04-01",
		"SignatureNonce":   "fixed-nonce",
		"Timestamp":        "2026-07-08T12:00:00Z",
		"AccessKeyId":      "ak",
		"RoleArn":          "acs:ram::123:role/test",
		"RoleSessionName":  "wuchang-upload",
		"DurationSeconds":  "3600",
	}
	sig1 := p.sign(params)
	sig2 := p.sign(params)
	assert.NotEmpty(t, sig1)
	assert.Equal(t, sig1, sig2, "signature must be deterministic for identical params")
}

// ----- IsAvailable 配置校验 -----

func TestAliyunProvider_IsAvailable(t *testing.T) {
	cases := []struct {
		name string
		cfg  *config.STSConfig
		want bool
	}{
		{"complete", &config.STSConfig{Provider: "aliyun", AccessKey: "ak", SecretKey: "sk", RoleArn: "acs:ram::1:role/r"}, true},
		{"missing_ak", &config.STSConfig{Provider: "aliyun", SecretKey: "sk", RoleArn: "acs:ram::1:role/r"}, false},
		{"missing_sk", &config.STSConfig{Provider: "aliyun", AccessKey: "ak", RoleArn: "acs:ram::1:role/r"}, false},
		{"missing_rolearn", &config.STSConfig{Provider: "aliyun", AccessKey: "ak", SecretKey: "sk"}, false},
		{"placeholder_ak", &config.STSConfig{Provider: "aliyun", AccessKey: "your-ak", SecretKey: "sk", RoleArn: "acs:ram::1:role/r"}, false},
		{"placeholder_sk", &config.STSConfig{Provider: "aliyun", AccessKey: "ak", SecretKey: "your-sk", RoleArn: "acs:ram::1:role/r"}, false},
		{"placeholder_rolearn", &config.STSConfig{Provider: "aliyun", AccessKey: "ak", SecretKey: "sk", RoleArn: "your-role"}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := NewAliyunProvider(c.cfg)
			assert.Equal(t, c.want, p.IsAvailable())
		})
	}
}

// ----- resolveProvider 降级路径 -----

func TestResolveProvider(t *testing.T) {
	t.Run("nil_cfg_returns_noop", func(t *testing.T) {
		_, ok := resolveProvider(nil).(*NoopProvider)
		assert.True(t, ok)
	})
	t.Run("empty_provider_returns_noop", func(t *testing.T) {
		_, ok := resolveProvider(&config.STSConfig{}).(*NoopProvider)
		assert.True(t, ok)
	})
	t.Run("aliyun_incomplete_returns_noop", func(t *testing.T) {
		p := resolveProvider(&config.STSConfig{Provider: "aliyun", AccessKey: "ak"})
		_, ok := p.(*NoopProvider)
		assert.True(t, ok)
		assert.False(t, p.IsAvailable())
	})
	t.Run("aliyun_complete_returns_real_provider", func(t *testing.T) {
		p := resolveProvider(&config.STSConfig{
			Provider: "aliyun", AccessKey: "ak", SecretKey: "sk",
			RoleArn: "acs:ram::1:role/r",
		})
		_, ok := p.(*AliyunProvider)
		assert.True(t, ok)
		assert.True(t, p.IsAvailable())
	})
}

// ----- NoopProvider / Get / IsAvailable -----

func TestNoopProvider(t *testing.T) {
	p := &NoopProvider{}
	assert.False(t, p.IsAvailable())
	_, err := p.AssumeRole(context.Background())
	assert.ErrorIs(t, err, ErrNotConfigured)
}

// TestGet_UninitializedReturnsNoop 未调用 Init 时 Get 返回 NoopProvider
func TestGet_UninitializedReturnsNoop(t *testing.T) {
	// 注意：本测试依赖全局 provider 状态，重置后再断言
	provider = nil
	p := Get()
	_, ok := p.(*NoopProvider)
	assert.True(t, ok)
	assert.False(t, IsAvailable())
}

// ----- normalizeObjectPrefix -----

func TestNormalizeObjectPrefix(t *testing.T) {
	cases := map[string]string{
		"":           "",
		"   ":        "",
		"uploads":    "uploads/",
		"uploads/":   "uploads/",
		"a/b":        "a/b/",
		"a/b/":       "a/b/",
		"  uploads ": "uploads/",
	}
	for in, want := range cases {
		assert.Equal(t, want, normalizeObjectPrefix(in), "normalizeObjectPrefix(%q)", in)
	}
}

// ----- DurationSeconds 钳制 + RoleSessionName 默认 -----

func TestNewAliyunProvider_Defaults(t *testing.T) {
	t.Run("duration_default_when_zero", func(t *testing.T) {
		p := NewAliyunProvider(&config.STSConfig{AccessKey: "ak", SecretKey: "sk", RoleArn: "r"})
		assert.Equal(t, 3600, p.durationSeconds)
		assert.Equal(t, "wuchang-upload", p.roleSessionName)
	})
	t.Run("duration_clamped_to_min", func(t *testing.T) {
		p := NewAliyunProvider(&config.STSConfig{AccessKey: "ak", SecretKey: "sk", RoleArn: "r", DurationSeconds: 100})
		assert.Equal(t, 900, p.durationSeconds) // 低于 900 钳到默认 3600？不，钳到 min 900
	})
	t.Run("duration_clamped_to_max", func(t *testing.T) {
		p := NewAliyunProvider(&config.STSConfig{AccessKey: "ak", SecretKey: "sk", RoleArn: "r", DurationSeconds: 99999})
		assert.Equal(t, 3600, p.durationSeconds)
	})
	t.Run("duration_in_range_kept", func(t *testing.T) {
		p := NewAliyunProvider(&config.STSConfig{AccessKey: "ak", SecretKey: "sk", RoleArn: "r", DurationSeconds: 1800})
		assert.Equal(t, 1800, p.durationSeconds)
	})
	t.Run("role_session_name_custom", func(t *testing.T) {
		p := NewAliyunProvider(&config.STSConfig{AccessKey: "ak", SecretKey: "sk", RoleArn: "r", RoleSessionName: "custom"})
		assert.Equal(t, "custom", p.roleSessionName)
	})
}

// ----- AssumeRole httptest 场景 -----

// newTestProvider 构造可用 AliyunProvider 并注入 httptest 端点
func newTestProvider(t *testing.T, srv *httptest.Server) *AliyunProvider {
	t.Helper()
	p := NewAliyunProvider(&config.STSConfig{
		Provider:    "aliyun",
		AccessKey:   "LTAI5tFakeAK",
		SecretKey:   "FakeSK1234567890",
		RoleArn:     "acs:ram::123456789:role/wuchang-upload",
		Bucket:      "wuchang-tongcheng",
		Region:      "oss-cn-hangzhou",
		Endpoint:    "https://oss-cn-hangzhou.aliyuncs.com",
		ObjectPrefix: "uploads/",
	})
	return p.withHTTPClient(srv.Client(), srv.URL)
}

// TestAssumeRole_Success 成功换取临时凭据 + 签名一致性校验
func TestAssumeRole_Success(t *testing.T) {
	var receivedForm url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		receivedForm, _ = url.ParseQuery(string(body))
		// 返回成功响应
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"Credentials": map[string]string{
				"AccessKeyId":     "STS.FakeTempAK",
				"AccessKeySecret": "FakeTempSK",
				"SecurityToken":   "CAESFakeToken==",
				"Expiration":      "2026-07-08T13:00:00Z",
			},
			"RequestId": "req-123",
		})
	}))
	defer srv.Close()

	p := newTestProvider(t, srv)
	creds, err := p.AssumeRole(context.Background())
	require.NoError(t, err)
	require.NotNil(t, creds)

	// 凭据字段
	assert.Equal(t, "STS.FakeTempAK", creds.AccessKeyID)
	assert.Equal(t, "FakeTempSK", creds.AccessKeySecret)
	assert.Equal(t, "CAESFakeToken==", creds.SecurityToken)
	assert.Equal(t, "2026-07-08T13:00:00Z", creds.Expiration)
	// OSS 落地信息透传
	assert.Equal(t, "wuchang-tongcheng", creds.Bucket)
	assert.Equal(t, "oss-cn-hangzhou", creds.Region)
	assert.Equal(t, "https://oss-cn-hangzhou.aliyuncs.com", creds.Endpoint)
	assert.Equal(t, "uploads/", creds.ObjectPrefix)
	assert.Equal(t, 3600, creds.ExpiresIn)

	// 请求参数断言
	assert.Equal(t, "AssumeRole", receivedForm.Get("Action"))
	assert.Equal(t, "2015-04-01", receivedForm.Get("Version"))
	assert.Equal(t, "acs:ram::123456789:role/wuchang-upload", receivedForm.Get("RoleArn"))
	assert.Equal(t, "wuchang-upload", receivedForm.Get("RoleSessionName"))
	assert.Equal(t, "3600", receivedForm.Get("DurationSeconds"))
	assert.Equal(t, "LTAI5tFakeAK", receivedForm.Get("AccessKeyId"))
	assert.Equal(t, "HMAC-SHA1", receivedForm.Get("SignatureMethod"))
	assert.Equal(t, "1.0", receivedForm.Get("SignatureVersion"))
	assert.NotEmpty(t, receivedForm.Get("SignatureNonce"))
	assert.NotEmpty(t, receivedForm.Get("Timestamp"))
	assert.NotEmpty(t, receivedForm.Get("Signature"))

	// 端到端签名一致性：用收到的全部参数（含随机 nonce/timestamp）重新签名应等于收到的 Signature
	receivedSig := receivedForm.Get("Signature")
	params := map[string]string{}
	for k := range receivedForm {
		params[k] = receivedForm.Get(k)
	}
	delete(params, "Signature")
	recomputed := p.sign(params)
	assert.Equal(t, receivedSig, recomputed, "signature must match server-recomputed value")
}

// TestAssumeRole_AliyunBusinessError 阿里云返回业务错误（Code != Success）
func TestAssumeRole_AliyunBusinessError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{
			"Code":      "NoPermission",
			"Message":   "You are not authorized to do this action",
			"RequestId": "req-err",
		})
	}))
	defer srv.Close()

	p := newTestProvider(t, srv)
	_, err := p.AssumeRole(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "NoPermission")
	assert.Contains(t, err.Error(), "not authorized")
}

// TestAssumeRole_NetworkError 网络错误（端点不可达）
func TestAssumeRole_NetworkError(t *testing.T) {
	p := newTestProvider(t, httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})))
	// 立即关闭服务器模拟端点不可达
	p.endpointURL = "http://127.0.0.1:0" // 不可达端口
	// 用一个极短超时客户端确保快速失败
	p.httpClient = &http.Client{Timeout: 100 * time.Millisecond}
	_, err := p.AssumeRole(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "sts: request failed")
}

// TestAssumeRole_NonJSONResponse 非 JSON 响应应返回解码错误
func TestAssumeRole_NonJSONResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("<html>not json</html>"))
	}))
	defer srv.Close()

	p := newTestProvider(t, srv)
	_, err := p.AssumeRole(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "decode response")
}

// TestAssumeRole_EmptyCredentials 成功 JSON 但凭据字段为空
func TestAssumeRole_EmptyCredentials(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"Credentials": map[string]string{
				"AccessKeyId":     "",
				"AccessKeySecret": "",
				"SecurityToken":   "",
				"Expiration":      "",
			},
			"RequestId": "req-empty",
		})
	}))
	defer srv.Close()

	p := newTestProvider(t, srv)
	_, err := p.AssumeRole(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty credentials")
}

// TestAssumeRole_NotConfigured 未配置时（IsAvailable=false）直接返回 ErrNotConfigured
func TestAssumeRole_NotConfigured(t *testing.T) {
	p := NewAliyunProvider(&config.STSConfig{Provider: "aliyun"}) // 配置不全
	_, err := p.AssumeRole(context.Background())
	assert.ErrorIs(t, err, ErrNotConfigured)
}

// TestInit_ResolvesProvider Init 按配置解析全局 provider
func TestInit_ResolvesProvider(t *testing.T) {
	t.Run("incomplete_falls_back_to_noop", func(t *testing.T) {
		Init(&config.STSConfig{Provider: "aliyun", AccessKey: "ak"})
		_, ok := Get().(*NoopProvider)
		assert.True(t, ok)
		assert.False(t, IsAvailable())
	})
	t.Run("empty_provider_falls_back_to_noop", func(t *testing.T) {
		Init(&config.STSConfig{})
		_, ok := Get().(*NoopProvider)
		assert.True(t, ok)
	})
	t.Run("complete_returns_aliyun", func(t *testing.T) {
		Init(&config.STSConfig{Provider: "aliyun", AccessKey: "ak", SecretKey: "sk", RoleArn: "r"})
		_, ok := Get().(*AliyunProvider)
		assert.True(t, ok)
		assert.True(t, IsAvailable())
	})
	// 重置全局状态避免污染其它测试
	provider = nil
}
