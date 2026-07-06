// aliyun.go 阿里云短信（dysmsapi）Provider 实现
//
// 通过 dysmsapi.aliyuncs.com RPC API 发送验证码，遵循项目既有风格：
//   - 标准库 net/http，无新外部依赖（与 pkg/amap 一致）
//   - HMAC-SHA1 签名（阿里云 RPC API v1 规则）
//   - AK/SK/SignName/TemplateCode 任一缺失时由 resolveProvider 降级 NoopProvider，
//     避免引入未集成依赖或运行时配置错误

package sms

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
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
	aliyunEndpoint   = "https://dysmsapi.aliyuncs.com"
	aliyunAPIVersion = "2017-05-25"
	aliyunRegionID   = "cn-hangzhou"
	aliyunHTTPMethod = "POST" // RPC v1 签名用 HTTP 方法
	aliyunTimeout    = 5 * time.Second
)

// AliyunProvider 阿里云短信服务商
//
// 通过 dysmsapi.aliyuncs.com RPC API（HMAC-SHA1 签名）发送验证码。
// 必要配置缺失时不应被构造（resolveProvider 会降级 NoopProvider）。
type AliyunProvider struct {
	accessKey    string
	secretKey    string
	signName     string
	templateCode string
	// httpClient 可注入，便于测试用 httptest.Server
	httpClient *http.Client
	// endpoint 可注入，便于测试指向 httptest.Server
	endpoint string
}

// NewAliyunProvider 构造阿里云短信 Provider
func NewAliyunProvider(cfg *config.SMSConfig) *AliyunProvider {
	return &AliyunProvider{
		accessKey:    cfg.AccessKey,
		secretKey:    cfg.SecretKey,
		signName:     cfg.SignName,
		templateCode: cfg.TemplateCode,
		httpClient:   &http.Client{Timeout: aliyunTimeout},
		endpoint:     aliyunEndpoint,
	}
}

// Send 发送验证码短信
func (p *AliyunProvider) Send(ctx context.Context, phone, code string) error {
	if phone == "" {
		return errors.New("手机号不能为空")
	}
	if !p.IsAvailable() {
		return errors.New("aliyun sms: config incomplete (access_key/secret_key/sign_name/template_code)")
	}
	// 阿里云要求模板参数为 JSON 字符串
	tplParam, err := json.Marshal(map[string]string{"code": code})
	if err != nil {
		return fmt.Errorf("aliyun sms: marshal template param: %w", err)
	}
	params := map[string]string{
		// 业务参数
		"PhoneNumbers":  phone,
		"SignName":      p.signName,
		"TemplateCode":  p.templateCode,
		"TemplateParam": string(tplParam),
		// 公共参数
		"Action":           "SendSms",
		"Version":          aliyunAPIVersion,
		"RegionId":         aliyunRegionID,
		"Format":           "JSON",
		"SignatureMethod":  "HMAC-SHA1",
		"SignatureVersion": "1.0",
		"SignatureNonce":   randNonce(),
		"Timestamp":        time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		"AccessKeyId":      p.accessKey,
	}
	params["Signature"] = p.sign(params)
	return p.doRequest(ctx, params)
}

// IsAvailable 必要配置是否齐全（供 resolveProvider 与测试使用）
func (p *AliyunProvider) IsAvailable() bool {
	return p.accessKey != "" && p.secretKey != "" && p.signName != "" && p.templateCode != ""
}

// sign 计算阿里云 RPC API 签名
//
//	StringToSign = HTTPMethod + "&" + percentEncode("/") + "&" + percentEncode(CanonicalizedQueryString)
//	Signature    = base64(HMAC-SHA1(StringToSign, SecretKey + "&"))
func (p *AliyunProvider) sign(params map[string]string) string {
	canonical := canonicalQueryString(params)
	stringToSign := aliyunHTTPMethod + "&" + percentEncode("/") + "&" + percentEncode(canonical)
	mac := hmac.New(sha1.New, []byte(p.secretKey+"&"))
	mac.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// canonicalQueryString 按 key 字典序排列，构造 key=value&... 并对 k/v 做 RFC3986 编码
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

// percentEncode RFC 3986 百分号编码（阿里云规则：A-Za-z0-9-_.~ 不编码，其余均编码，空格→%20）
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

// doRequest 发送 POST 表单请求并解析响应
func (p *AliyunProvider) doRequest(ctx context.Context, params map[string]string) error {
	form := make(url.Values, len(params))
	for k, v := range params {
		form.Set(k, v)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("aliyun sms: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := p.httpClient
	if client == nil {
		client = &http.Client{Timeout: aliyunTimeout}
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("aliyun sms: request failed: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("aliyun sms: read body: %w", err)
	}
	var result struct {
		Code      string `json:"Code"`
		Message   string `json:"Message"`
		BizId     string `json:"BizId"`
		RequestId string `json:"RequestId"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("aliyun sms: decode response: %w (body=%s)", err, string(body))
	}
	if result.Code != "OK" {
		return fmt.Errorf("aliyun sms: send failed: code=%s message=%s", result.Code, result.Message)
	}
	return nil
}

// withHTTPClient 内部测试辅助：注入 http client 与 endpoint，返回自身便于链式调用
func (p *AliyunProvider) withHTTPClient(c *http.Client, endpoint string) *AliyunProvider {
	p.httpClient = c
	p.endpoint = endpoint
	return p
}

// 编译期断言：AliyunProvider 实现 Provider 接口
var _ Provider = (*AliyunProvider)(nil)
