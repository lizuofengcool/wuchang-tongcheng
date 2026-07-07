// aliyun.go 阿里云 STS AssumeRole Provider 实现
//
// 通过 sts.aliyuncs.com RPC API 申请临时访问凭据，遵循项目既有风格：
//   - 标准库 net/http，无新外部依赖（与 pkg/amap / pkg/sms 一致）
//   - HMAC-SHA1 签名（阿里云 RPC API v1 规则，与 pkg/sms/aliyun.go 同算法）
//   - AccessKey/SecretKey/RoleArn 任一缺失或占位时由 resolveProvider 降级 NoopProvider
//
// STS 直传与 RAM 角色模型：后端用主账号/子账号 AK/SK 调用 AssumeRole 扮演一个配置了
// OSS 写权限策略的 RAM 角色，拿到临时凭据后下发给前端；前端凭据过期后再来换取。
package sts

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"wuchang-tongcheng/internal/pkg/config"
)

const (
	aliyunSTSEndpoint   = "https://sts.aliyuncs.com"
	aliyunSTSAPIVersion = "2015-04-01"
	aliyunSTSRegionID   = "cn-hangzhou"
	aliyunSTSHTTPMethod = "POST"
	aliyunSTSTimeout    = 5 * time.Second

	// STS DurationSeconds 取值范围 900~3600，默认 3600（1 小时）
	defaultDurationSeconds = 3600
	minDurationSeconds     = 900
	maxDurationSeconds     = 3600

	// 占位值前缀（与 storage/sms 判定一致）
	placeholderPrefix = "your-"
)

// AliyunProvider 阿里云 STS 服务商
type AliyunProvider struct {
	accessKey       string
	secretKey       string
	roleArn         string
	roleSessionName string
	durationSeconds int
	// OSS 落地信息（透传给前端，构造 OSS 客户端用）
	bucket       string
	region       string
	endpoint     string
	objectPrefix string
	// 可注入，便于测试用 httptest.Server
	httpClient *http.Client
	endpointURL string // STS API 端点（默认 https://sts.aliyuncs.com）
}

// NewAliyunProvider 构造阿里云 STS Provider
func NewAliyunProvider(cfg *config.STSConfig) *AliyunProvider {
	dur := cfg.DurationSeconds
	if dur == 0 {
		dur = defaultDurationSeconds
	}
	if dur < minDurationSeconds {
		dur = minDurationSeconds
	}
	if dur > maxDurationSeconds {
		dur = maxDurationSeconds
	}
	session := cfg.RoleSessionName
	if session == "" {
		session = "wuchang-upload"
	}
	return &AliyunProvider{
		accessKey:       cfg.AccessKey,
		secretKey:       cfg.SecretKey,
		roleArn:         cfg.RoleArn,
		roleSessionName: session,
		durationSeconds: dur,
		bucket:          cfg.Bucket,
		region:          cfg.Region,
		endpoint:        cfg.Endpoint,
		objectPrefix:    normalizeObjectPrefix(cfg.ObjectPrefix),
		httpClient:      &http.Client{Timeout: aliyunSTSTimeout},
		endpointURL:     aliyunSTSEndpoint,
	}
}

// IsAvailable 必要配置是否齐全且非占位值
func (p *AliyunProvider) IsAvailable() bool {
	return p.accessKey != "" && !strings.HasPrefix(p.accessKey, placeholderPrefix) &&
		p.secretKey != "" && !strings.HasPrefix(p.secretKey, placeholderPrefix) &&
		p.roleArn != "" && !strings.HasPrefix(p.roleArn, placeholderPrefix)
}

// AssumeRole 调用 STS AssumeRole 申请临时凭据
func (p *AliyunProvider) AssumeRole(ctx context.Context) (*STSCredentials, error) {
	if !p.IsAvailable() {
		return nil, ErrNotConfigured
	}
	params := map[string]string{
		// 业务参数
		"RoleArn":         p.roleArn,
		"RoleSessionName": p.roleSessionName,
		"DurationSeconds": fmt.Sprintf("%d", p.durationSeconds),
		// 公共参数
		"Action":           "AssumeRole",
		"Version":          aliyunSTSAPIVersion,
		"RegionId":         aliyunSTSRegionID,
		"Format":           "JSON",
		"SignatureMethod":  "HMAC-SHA1",
		"SignatureVersion": "1.0",
		"SignatureNonce":   randNonce(),
		"Timestamp":        time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		"AccessKeyId":      p.accessKey,
	}
	params["Signature"] = p.sign(params)

	creds, err := p.doRequest(ctx, params)
	if err != nil {
		return nil, err
	}

	return &STSCredentials{
		Credentials:  *creds,
		Bucket:       p.bucket,
		Region:       p.region,
		Endpoint:     p.endpoint,
		ObjectPrefix: p.objectPrefix,
		ExpiresIn:    p.durationSeconds,
	}, nil
}

// sign 计算阿里云 RPC API 签名（与 pkg/sms/aliyun.go 同算法）
//
//	StringToSign = HTTPMethod + "&" + percentEncode("/") + "&" + percentEncode(CanonicalizedQueryString)
//	Signature    = base64(HMAC-SHA1(StringToSign, SecretKey + "&"))
func (p *AliyunProvider) sign(params map[string]string) string {
	canonical := canonicalQueryString(params)
	stringToSign := aliyunSTSHTTPMethod + "&" + percentEncode("/") + "&" + percentEncode(canonical)
	mac := hmac.New(sha1.New, []byte(p.secretKey+"&"))
	mac.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// canonicalQueryString 按 key 字典序构造 key=value&... 并对 k/v 做 RFC3986 编码
func canonicalQueryString(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for i, k := range keys {
		if i > 0 {
			b.WriteByte('&')
		}
		b.WriteString(percentEncode(k))
		b.WriteByte('=')
		b.WriteString(percentEncode(params[k]))
	}
	return b.String()
}

// percentEncode RFC 3986 百分号编码（A-Za-z0-9-_.~ 不编码，空格→%20）
func percentEncode(s string) string {
	var b strings.Builder
	for _, r := range s {
		if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') ||
			r == '-' || r == '_' || r == '.' || r == '~' {
			b.WriteRune(r)
			continue
		}
		for _, bt := range []byte(string(r)) {
			fmt.Fprintf(&b, "%%%02X", bt)
		}
	}
	return b.String()
}

// randNonce 生成 16 字节随机十六进制串作为 SignatureNonce
func randNonce() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand 失败极少见；退化为时间戳保证唯一性
		return fmt.Sprintf("%x", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

// doRequest 发送 POST 表单请求并解析 AssumeRole 响应
func (p *AliyunProvider) doRequest(ctx context.Context, params map[string]string) (*Credentials, error) {
	form := make(url.Values, len(params))
	for k, v := range params {
		form.Set(k, v)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpointURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("sts: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := p.httpClient
	if client == nil {
		client = &http.Client{Timeout: aliyunSTSTimeout}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sts: request failed: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("sts: read body: %w", err)
	}
	// 阿里云错误响应：{ "Code": "...", "Message": "...", "RequestId": "..." }
	var errResp struct {
		Code      string `json:"Code"`
		Message   string `json:"Message"`
		RequestId string `json:"RequestId"`
	}
	if jerr := json.Unmarshal(body, &errResp); jerr == nil && errResp.Code != "" && errResp.Code != "Success" {
		return nil, fmt.Errorf("sts: AssumeRole failed: code=%s message=%s", errResp.Code, errResp.Message)
	}
	// 成功响应：{ "Credentials": {...}, "RequestId": "..." }
	var okResp struct {
		Credentials struct {
			AccessKeyID     string `json:"AccessKeyId"`
			AccessKeySecret string `json:"AccessKeySecret"`
			SecurityToken   string `json:"SecurityToken"`
			Expiration      string `json:"Expiration"`
		} `json:"Credentials"`
		RequestId string `json:"RequestId"`
	}
	if err := json.Unmarshal(body, &okResp); err != nil {
		return nil, fmt.Errorf("sts: decode response: %w (body=%s)", err, string(body))
	}
	if okResp.Credentials.AccessKeyID == "" || okResp.Credentials.SecurityToken == "" {
		return nil, fmt.Errorf("sts: empty credentials in response (body=%s)", string(body))
	}
	return &Credentials{
		AccessKeyID:     okResp.Credentials.AccessKeyID,
		AccessKeySecret: okResp.Credentials.AccessKeySecret,
		SecurityToken:   okResp.Credentials.SecurityToken,
		Expiration:      okResp.Credentials.Expiration,
	}, nil
}

// normalizeObjectPrefix 规范化对象 key 前缀：非空则补齐尾部 /
func normalizeObjectPrefix(prefix string) string {
	if prefix == "" {
		return ""
	}
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		return ""
	}
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	return prefix
}

// withHTTPClient 内部测试辅助：注入 http client 与 STS 端点，返回自身便于链式调用
func (p *AliyunProvider) withHTTPClient(c *http.Client, endpoint string) *AliyunProvider {
	p.httpClient = c
	p.endpointURL = endpoint
	return p
}

// 编译期断言：AliyunProvider 实现 Provider 接口
var _ Provider = (*AliyunProvider)(nil)
