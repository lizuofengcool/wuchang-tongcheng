package oauth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"wuchang-tongcheng/internal/pkg/config"
)

// ===== MockProvider =====

func TestMockProvider_GetUserInfo(t *testing.T) {
	cases := []struct {
		code    string
		wantOID string
		wantNick string
		wantErr bool
	}{
		{"mock:o_123", "o_123", "微信用户", false},
		{"mock:o_456:张三", "o_456", "张三", false},
		{"mock::nick", "", "", true},        // 空 openid
		{"notmock:o_1", "", "", true},        // 前缀不符
		{"", "", "", true},                   // 空码
	}
	for _, c := range cases {
		info, err := MockProvider{}.GetUserInfo(context.Background(), c.code)
		if c.wantErr {
			if err == nil {
				t.Fatalf("code=%q: expected error, got %+v", c.code, info)
			}
			continue
		}
		if err != nil {
			t.Fatalf("code=%q: unexpected err %v", c.code, err)
		}
		if info.OpenID != c.wantOID {
			t.Fatalf("code=%q: openid=%q want %q", c.code, info.OpenID, c.wantOID)
		}
		if info.Nickname != c.wantNick {
			t.Fatalf("code=%q: nickname=%q want %q", c.code, info.Nickname, c.wantNick)
		}
		if info.Provider != "wechat" {
			t.Fatalf("code=%q: provider=%q want wechat", c.code, info.Provider)
		}
		if info.UnionID != "mock_union_"+c.wantOID {
			t.Fatalf("code=%q: unionid=%q", c.code, info.UnionID)
		}
	}
}

func TestMockProvider_NameAvailable(t *testing.T) {
	if (MockProvider{}).Name() != "wechat" {
		t.Fatal("name should be wechat")
	}
	if !(MockProvider{}).IsAvailable() {
		t.Fatal("mock provider should be available")
	}
}

// ===== NoopProvider =====

func TestNoopProvider_GetUserInfo(t *testing.T) {
	if (NoopProvider{}).IsAvailable() {
		t.Fatal("noop should not be available")
	}
	_, err := NoopProvider{}.GetUserInfo(context.Background(), "any")
	if !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("err=%v want ErrNotConfigured", err)
	}
}

// ===== resolveWeChatProvider / NewService =====

func TestResolveWeChatProvider(t *testing.T) {
	cases := []struct {
		cfg  *config.WeChatConfig
		want string // "nil" | "mock" | "wechat"
	}{
		{nil, "nil"},
		{&config.WeChatConfig{}, "nil"},                                                  // provider 空
		{&config.WeChatConfig{Provider: "mock"}, "mock"},                                 // mock
		{&config.WeChatConfig{Provider: "MOCK"}, "mock"},                                 // 大小写不敏感
		{&config.WeChatConfig{Provider: "wechat"}, "nil"},                                // 缺 AppID/Secret
		{&config.WeChatConfig{Provider: "wechat", AppID: "your-id", AppSecret: "s"}, "nil"}, // AppID 占位
		{&config.WeChatConfig{Provider: "wechat", AppID: "id", AppSecret: "your-s"}, "nil"}, // Secret 占位
		{&config.WeChatConfig{Provider: "wechat", AppID: "wx123", AppSecret: "secret"}, "wechat"},
	}
	for i, c := range cases {
		p := resolveWeChatProvider(c.cfg)
		got := "nil"
		switch p.(type) {
		case MockProvider:
			got = "mock"
		case *WeChatProvider:
			got = "wechat"
		case nil:
			got = "nil"
		}
		if got != c.want {
			t.Fatalf("case %d: got=%s want=%s", i, got, c.want)
		}
	}
}

func TestNewService_NilConfig(t *testing.T) {
	s := NewService(nil)
	if s.HasProvider("wechat") {
		t.Fatal("nil config should register no providers")
	}
	_, err := s.Login(context.Background(), "wechat", "code")
	if !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("err=%v want ErrNotConfigured", err)
	}
}

func TestNewService_MockConfig(t *testing.T) {
	s := NewService(&config.OAuthConfig{WeChat: config.WeChatConfig{Provider: "mock"}})
	if !s.HasProvider("wechat") {
		t.Fatal("mock config should register wechat provider")
	}
	info, err := s.Login(context.Background(), "wechat", "mock:o_abc:nick")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if info.OpenID != "o_abc" || info.Nickname != "nick" {
		t.Fatalf("unexpected info %+v", info)
	}
}

func TestService_Login_EmptyCode(t *testing.T) {
	s := NewService(&config.OAuthConfig{WeChat: config.WeChatConfig{Provider: "mock"}})
	_, err := s.Login(context.Background(), "wechat", "")
	if !errors.Is(err, ErrInvalidCode) {
		t.Fatalf("err=%v want ErrInvalidCode", err)
	}
}

func TestService_Login_UnknownProvider(t *testing.T) {
	s := NewService(&config.OAuthConfig{WeChat: config.WeChatConfig{Provider: "mock"}})
	_, err := s.Login(context.Background(), "github", "mock:x")
	if !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("err=%v want ErrNotConfigured", err)
	}
}

// ===== WeChatProvider =====

func TestWeChatProvider_IsAvailable(t *testing.T) {
	cases := []struct {
		appID, appSecret string
		want             bool
	}{
		{"", "", false},
		{"wx123", "", false},
		{"", "secret", false},
		{"your-appid", "secret", false},
		{"wx123", "your-secret", false},
		{"wx123", "secret", true},
	}
	for i, c := range cases {
		p := NewWeChatProvider(&config.WeChatConfig{AppID: c.appID, AppSecret: c.appSecret})
		if got := p.IsAvailable(); got != c.want {
			t.Fatalf("case %d: IsAvailable=%v want %v", i, got, c.want)
		}
	}
}

func TestWeChatProvider_GetUserInfo_NotConfigured(t *testing.T) {
	p := NewWeChatProvider(&config.WeChatConfig{})
	_, err := p.GetUserInfo(context.Background(), "code")
	if !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("err=%v want ErrNotConfigured", err)
	}
}

func TestWeChatProvider_GetUserInfo_EmptyCode(t *testing.T) {
	p := NewWeChatProvider(&config.WeChatConfig{AppID: "wx", AppSecret: "s"})
	_, err := p.GetUserInfo(context.Background(), "")
	if !errors.Is(err, ErrInvalidCode) {
		t.Fatalf("err=%v want ErrInvalidCode", err)
	}
}

// newWeChatTestServer 启动一个模拟微信 OAuth 两段式接口的 httptest.Server，
// 返回 server 与控制开关（accessOK / userOK 控制各阶段是否返回成功）
func newWeChatTestServer(t *testing.T, accessOK, userOK bool) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/sns/oauth2/access_token", func(w http.ResponseWriter, r *http.Request) {
		if !accessOK {
			w.Write([]byte(`{"errcode":40029,"errmsg":"invalid code"}`))
			return
		}
		w.Write([]byte(`{
			"access_token":"ACCESS_TOKEN_X",
			"expires_in":7200,
			"refresh_token":"RT",
			"openid":"OPENID_001",
			"scope":"snsapi_login",
			"unionid":"UNIONID_001"
		}`))
	})
	mux.HandleFunc("/sns/userinfo", func(w http.ResponseWriter, r *http.Request) {
		if !userOK {
			w.Write([]byte(`{"errcode":48001,"errmsg":"api unauthorized"}`))
			return
		}
		// 校验入参
		if r.URL.Query().Get("openid") != "OPENID_001" {
			w.Write([]byte(`{"errcode":40003,"errmsg":"invalid openid"}`))
			return
		}
		w.Write([]byte(`{
			"openid":"OPENID_001",
			"nickname":"微信用户",
			"sex":1,
			"headimgurl":"https://wx.qlogo.cn/avatar.png",
			"unionid":"UNIONID_001"
		}`))
	})
	return httptest.NewServer(mux)
}

func TestWeChatProvider_GetUserInfo_Success(t *testing.T) {
	srv := newWeChatTestServer(t, true, true)
	defer srv.Close()

	p := NewWeChatProvider(&config.WeChatConfig{AppID: "wx123", AppSecret: "secret"})
	p.withHTTPClient(srv.Client(), srv.URL+"/sns/oauth2/access_token", srv.URL+"/sns/userinfo")

	info, err := p.GetUserInfo(context.Background(), "CODE_FROM_REDIRECT")
	if err != nil {
		t.Fatalf("GetUserInfo: %v", err)
	}
	if info.Provider != "wechat" {
		t.Fatalf("provider=%q", info.Provider)
	}
	if info.OpenID != "OPENID_001" {
		t.Fatalf("openid=%q", info.OpenID)
	}
	if info.UnionID != "UNIONID_001" {
		t.Fatalf("unionid=%q", info.UnionID)
	}
	if info.Nickname != "微信用户" {
		t.Fatalf("nickname=%q", info.Nickname)
	}
	if info.Avatar != "https://wx.qlogo.cn/avatar.png" {
		t.Fatalf("avatar=%q", info.Avatar)
	}
}

func TestWeChatProvider_GetUserInfo_BadCode(t *testing.T) {
	srv := newWeChatTestServer(t, false, true)
	defer srv.Close()

	p := NewWeChatProvider(&config.WeChatConfig{AppID: "wx123", AppSecret: "secret"})
	p.withHTTPClient(srv.Client(), srv.URL+"/sns/oauth2/access_token", srv.URL+"/sns/userinfo")

	_, err := p.GetUserInfo(context.Background(), "bad")
	if err == nil {
		t.Fatal("expected error for bad code")
	}
	if !strings.Contains(err.Error(), "exchange code failed") {
		t.Fatalf("err=%v should mention exchange code failed", err)
	}
}

func TestWeChatProvider_GetUserInfo_UserInfoErr(t *testing.T) {
	srv := newWeChatTestServer(t, true, false)
	defer srv.Close()

	p := NewWeChatProvider(&config.WeChatConfig{AppID: "wx123", AppSecret: "secret"})
	p.withHTTPClient(srv.Client(), srv.URL+"/sns/oauth2/access_token", srv.URL+"/sns/userinfo")

	_, err := p.GetUserInfo(context.Background(), "ok")
	if err == nil {
		t.Fatal("expected error from userinfo")
	}
	if !strings.Contains(err.Error(), "fetch userinfo failed") {
		t.Fatalf("err=%v should mention fetch userinfo failed", err)
	}
}

func TestWeChatProvider_GetUserInfo_NetworkError(t *testing.T) {
	// 指向一个已关闭的端口模拟网络错误
	p := NewWeChatProvider(&config.WeChatConfig{AppID: "wx123", AppSecret: "secret"})
	p.withHTTPClient(&http.Client{Timeout: wechatTimeout}, "http://127.0.0.1:1/sns/oauth2/access_token", "http://127.0.0.1:1/sns/userinfo")

	_, err := p.GetUserInfo(context.Background(), "ok")
	if err == nil {
		t.Fatal("expected network error")
	}
	if !strings.Contains(err.Error(), "request failed") {
		t.Fatalf("err=%v should mention request failed", err)
	}
}

func TestWeChatProvider_Name(t *testing.T) {
	if NewWeChatProvider(&config.WeChatConfig{}).Name() != "wechat" {
		t.Fatal("name should be wechat")
	}
}

// 编译期断言：MockProvider / NoopProvider 实现 Provider 接口
var (
	_ Provider = MockProvider{}
	_ Provider = NoopProvider{}
)
