package sms

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"wuchang-tongcheng/internal/pkg/config"
)

// TestPercentEncode 阿里云 RFC3986 百分号编码规则
func TestPercentEncode(t *testing.T) {
	cases := map[string]string{
		"abc123-_.~": "abc123-_.~", // 不编码字符原样保留
		" ":          "%20",        // 空格 → %20（非 +）
		"+":          "%2B",
		"/":          "%2F",
		"=":          "%3D",
		"&":          "%26",
		"中文":         "%E4%B8%AD%E6%96%87", // UTF-8 三字节
		"a b":        "a%20b",
		"a+b/c":      "a%2Bb%2Fc",
	}
	for in, want := range cases {
		if got := percentEncode(in); got != want {
			t.Errorf("percentEncode(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestCanonicalQueryString key 按字典序，k/v 均编码
func TestCanonicalQueryString(t *testing.T) {
	params := map[string]string{
		"PhoneNumbers": "13800138000",
		"Action":       "SendSms",
		"SignName":     "测试",
	}
	got := canonicalQueryString(params)
	// 期望顺序：Action, PhoneNumbers, SignName
	want := "Action=SendSms&PhoneNumbers=13800138000&SignName=" + percentEncode("测试")
	if got != want {
		t.Errorf("canonicalQueryString = %q, want %q", got, want)
	}
}

// TestAliyunProvider_IsAvailable 四项配置齐全才可用
func TestAliyunProvider_IsAvailable(t *testing.T) {
	full := &AliyunProvider{accessKey: "ak", secretKey: "sk", signName: "sign", templateCode: "tpl"}
	if !full.IsAvailable() {
		t.Fatal("full config should be available")
	}
	missing := []AliyunProvider{
		{secretKey: "sk", signName: "sign", templateCode: "tpl"}, // 无 AK
		{accessKey: "ak", signName: "sign", templateCode: "tpl"}, // 无 SK
		{accessKey: "ak", secretKey: "sk", templateCode: "tpl"},  // 无 SignName
		{accessKey: "ak", secretKey: "sk", signName: "sign"},     // 无 TemplateCode
	}
	for i, p := range missing {
		if p.IsAvailable() {
			t.Errorf("case %d: expected not available", i)
		}
	}
}

// TestAliyunProvider_SendSuccess 模拟阿里云返回 OK，验证请求参数完整且签名正确
func TestAliyunProvider_SendSuccess(t *testing.T) {
	var receivedForm url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		receivedForm = r.PostForm
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"Code":"OK","BizId":"12345","RequestId":"req-1"}`))
	}))
	defer srv.Close()

	p := NewAliyunProvider(&config.SMSConfig{
		AccessKey: "testid", SecretKey: "testsecret",
		SignName: "测试签名", TemplateCode: "SMS_123456",
	}).withHTTPClient(srv.Client(), srv.URL)

	if err := p.Send(context.Background(), "13800138000", "123456"); err != nil {
		t.Fatalf("Send: %v", err)
	}
	// 校验关键业务参数
	if got := receivedForm.Get("Action"); got != "SendSms" {
		t.Errorf("Action = %q, want SendSms", got)
	}
	if got := receivedForm.Get("PhoneNumbers"); got != "13800138000" {
		t.Errorf("PhoneNumbers = %q", got)
	}
	if got := receivedForm.Get("SignName"); got != "测试签名" {
		t.Errorf("SignName = %q", got)
	}
	if got := receivedForm.Get("TemplateCode"); got != "SMS_123456" {
		t.Errorf("TemplateCode = %q", got)
	}
	if got := receivedForm.Get("TemplateParam"); got != `{"code":"123456"}` {
		t.Errorf("TemplateParam = %q", got)
	}
	if receivedForm.Get("Signature") == "" {
		t.Error("Signature empty")
	}
	if receivedForm.Get("SignatureNonce") == "" {
		t.Error("SignatureNonce empty")
	}
	if receivedForm.Get("Timestamp") == "" {
		t.Error("Timestamp empty")
	}
	// 独立重算签名校验正确性
	if !verifyAliyunSignature(t, receivedForm, "testsecret") {
		t.Fatal("signature verification failed")
	}
}

// TestAliyunProvider_SendFailure 阿里云返回非 OK → 包装错误
func TestAliyunProvider_SendFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"Code":"isv.BUSINESS_LIMIT_CONTROL","Message":"业务限流","RequestId":"req-2"}`))
	}))
	defer srv.Close()

	p := NewAliyunProvider(&config.SMSConfig{
		AccessKey: "testid", SecretKey: "testsecret",
		SignName: "sign", TemplateCode: "SMS_1",
	}).withHTTPClient(srv.Client(), srv.URL)

	err := p.Send(context.Background(), "13800138000", "123456")
	if err == nil {
		t.Fatal("expected error for non-OK response")
	}
	if !strings.Contains(err.Error(), "isv.BUSINESS_LIMIT_CONTROL") {
		t.Errorf("err = %v, want contains isv.BUSINESS_LIMIT_CONTROL", err)
	}
}

// TestAliyunProvider_SendNetworkError 网络错误（不可达端点）→ 返回错误
func TestAliyunProvider_SendNetworkError(t *testing.T) {
	// 指向一个已关闭的 server 地址，触发连接拒绝
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Close()

	p := NewAliyunProvider(&config.SMSConfig{
		AccessKey: "testid", SecretKey: "testsecret",
		SignName: "sign", TemplateCode: "SMS_1",
	}).withHTTPClient(srv.Client(), srv.URL)

	err := p.Send(context.Background(), "13800138000", "123456")
	if err == nil {
		t.Fatal("expected network error")
	}
	if !strings.Contains(err.Error(), "request failed") {
		t.Errorf("err = %v, want contains 'request failed'", err)
	}
}

// TestAliyunProvider_SendInvalidJSON 响应非 JSON → 解码错误
func TestAliyunProvider_SendInvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<html>not json</html>`))
	}))
	defer srv.Close()

	p := NewAliyunProvider(&config.SMSConfig{
		AccessKey: "testid", SecretKey: "testsecret",
		SignName: "sign", TemplateCode: "SMS_1",
	}).withHTTPClient(srv.Client(), srv.URL)

	err := p.Send(context.Background(), "13800138000", "123456")
	if err == nil {
		t.Fatal("expected decode error")
	}
	if !strings.Contains(err.Error(), "decode response") {
		t.Errorf("err = %v, want contains 'decode response'", err)
	}
}

// TestAliyunProvider_SendEmptyPhone 空手机号拒绝
func TestAliyunProvider_SendEmptyPhone(t *testing.T) {
	p := NewAliyunProvider(&config.SMSConfig{
		AccessKey: "testid", SecretKey: "testsecret",
		SignName: "sign", TemplateCode: "SMS_1",
	})
	if err := p.Send(context.Background(), "", "123456"); err == nil {
		t.Fatal("expected error for empty phone")
	}
}

// TestAliyunProvider_SendConfigIncomplete 配置不全直接拒绝（不发起请求）
func TestAliyunProvider_SendConfigIncomplete(t *testing.T) {
	p := &AliyunProvider{accessKey: "ak", secretKey: "sk"} // 缺 SignName/TemplateCode
	err := p.Send(context.Background(), "13800138000", "123456")
	if err == nil {
		t.Fatal("expected error for incomplete config")
	}
	if !strings.Contains(err.Error(), "config incomplete") {
		t.Errorf("err = %v, want contains 'config incomplete'", err)
	}
}

// TestResolveProvider_Aliyun 四项配置齐全 → 返回 *AliyunProvider
func TestResolveProvider_Aliyun(t *testing.T) {
	cfg := &config.SMSConfig{
		Provider:     "aliyun",
		AccessKey:    "LTAIxxxx",
		SecretKey:    "secretxxxx",
		SignName:     "武昌同城",
		TemplateCode: "SMS_123456",
	}
	p, ok := resolveProvider(cfg).(*AliyunProvider)
	if !ok {
		t.Fatalf("expected *AliyunProvider, got %T", resolveProvider(cfg))
	}
	if p.signName != "武昌同城" {
		t.Errorf("signName = %q", p.signName)
	}
	if p.templateCode != "SMS_123456" {
		t.Errorf("templateCode = %q", p.templateCode)
	}
	// 大写 ALIYUN 也应识别
	cfg.Provider = "ALIYUN"
	if _, ok := resolveProvider(cfg).(*AliyunProvider); !ok {
		t.Fatal("ALIYUN (uppercase) should resolve to AliyunProvider")
	}
}

// TestAliyunProvider_SignDeterministic 相同入参签名一致；变更任一参数签名变化
func TestAliyunProvider_SignDeterministic(t *testing.T) {
	p := &AliyunProvider{accessKey: "ak", secretKey: "sk", signName: "s", templateCode: "t"}
	params := map[string]string{"A": "1", "B": "2", "SignatureNonce": "n", "Timestamp": "2024-01-01T00:00:00Z"}
	s1 := p.sign(params)
	s2 := p.sign(params)
	if s1 != s2 {
		t.Fatalf("signature not deterministic: %q vs %q", s1, s2)
	}
	// 改动一个参数
	orig := params["A"]
	params["A"] = "999"
	if p.sign(params) == s1 {
		t.Fatal("signature should change when param changes")
	}
	params["A"] = orig
	// 还原后应再次一致
	if p.sign(params) != s1 {
		t.Fatal("signature should restore after reverting param")
	}
	// 验证签名是合法 base64
	if _, err := base64.StdEncoding.DecodeString(s1); err != nil {
		t.Fatalf("signature not valid base64: %v", err)
	}
}

// verifyAliyunSignature 独立重算签名，校验请求里的 Signature 是否正确
func verifyAliyunSignature(t *testing.T, form url.Values, secretKey string) bool {
	t.Helper()
	sig := form.Get("Signature")
	// 复制一份并移除 Signature，构造 canonical
	clone := make(map[string]string, len(form))
	for k, v := range form {
		if k == "Signature" {
			continue
		}
		if len(v) > 0 {
			clone[k] = v[0]
		}
	}
	canonical := canonicalQueryString(clone)
	stringToSign := aliyunHTTPMethod + "&" + percentEncode("/") + "&" + percentEncode(canonical)
	mac := hmac.New(sha1.New, []byte(secretKey+"&"))
	mac.Write([]byte(stringToSign))
	expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return sig == expected
}
