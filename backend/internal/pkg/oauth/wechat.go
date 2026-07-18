// wechat.go 微信开放平台网站应用 OAuth2 实现
//
// 流程（前端在微信扫码授权后拿到 code，POST 到后端 /login/oauth/wechat）：
//  1. code → access_token + openid（+unionid）
//     GET https://api.weixin.qq.com/sns/oauth2/access_token?appid=&secret=&code=&grant_type=authorization_code
//  2. access_token + openid → 用户资料（nickname + headimgurl + unionid）
//     GET https://api.weixin.qq.com/sns/userinfo?access_token=&openid=
//
// 遵循项目风格：标准库 net/http，无新外部依赖（与 pkg/amap / pkg/sms / pkg/sts 一致）。
// AppID/AppSecret 任一缺失或占位（your-）时由 resolveWeChatProvider 不注册（Login 返回 ErrNotConfigured）。
package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"wuchang-tongcheng/internal/pkg/config"
)

const (
	wechatAccessTokenURL = "https://api.weixin.qq.com/sns/oauth2/access_token"
	wechatUserInfoURL    = "https://api.weixin.qq.com/sns/userinfo"
	wechatTimeout        = 5 * time.Second
	wechatPlaceholder    = "your-" // 与 storage/sms/sts 占位判定一致
)

// WeChatProvider 微信开放平台网站应用 OAuth
type WeChatProvider struct {
	appID     string
	appSecret string
	// 可注入，便于测试用 httptest.Server
	httpClient     *http.Client
	accessTokenURL string
	userInfoURL    string
}

// NewWeChatProvider 构造微信 OAuth Provider
func NewWeChatProvider(cfg *config.WeChatConfig) *WeChatProvider {
	return &WeChatProvider{
		appID:          cfg.AppID,
		appSecret:      cfg.AppSecret,
		httpClient:     &http.Client{Timeout: wechatTimeout},
		accessTokenURL: wechatAccessTokenURL,
		userInfoURL:    wechatUserInfoURL,
	}
}

// IsAvailable AppID/AppSecret 齐全且非占位值
func (p *WeChatProvider) IsAvailable() bool {
	return p.appID != "" && !strings.HasPrefix(p.appID, wechatPlaceholder) &&
		p.appSecret != "" && !strings.HasPrefix(p.appSecret, wechatPlaceholder)
}

// Name 提供商名称
func (p *WeChatProvider) Name() string { return "wechat" }

// GetUserInfo code → access_token → 用户资料
func (p *WeChatProvider) GetUserInfo(ctx context.Context, code string) (*UserInfo, error) {
	if code == "" {
		return nil, ErrInvalidCode
	}
	if !p.IsAvailable() {
		return nil, ErrNotConfigured
	}
	tok, err := p.exchangeCode(ctx, code)
	if err != nil {
		return nil, err
	}
	info, err := p.fetchUserInfo(ctx, tok.AccessToken, tok.OpenID)
	if err != nil {
		return nil, err
	}
	return &UserInfo{
		Provider: p.Name(),
		OpenID:   info.OpenID,
		UnionID:  info.UnionID,
		Nickname: info.Nickname,
		Avatar:   info.HeadImgURL,
	}, nil
}

// wechatToken access_token 接口响应字段
type wechatToken struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenID       string `json:"openid"`
	UnionID      string `json:"unionid"`
	Scope        string `json:"scope"`
}

// wechatUserinfo userinfo 接口响应字段
type wechatUserinfo struct {
	OpenID     string   `json:"openid"`
	Nickname   string   `json:"nickname"`
	Sex        int      `json:"sex"`
	Province   string   `json:"province"`
	City       string   `json:"city"`
	Country    string   `json:"country"`
	HeadImgURL string   `json:"headimgurl"`
	Privilege  []string `json:"privilege"`
	UnionID    string   `json:"unionid"`
}

// exchangeCode 第 1 步：code 换 access_token + openid
func (p *WeChatProvider) exchangeCode(ctx context.Context, code string) (*wechatToken, error) {
	q := url.Values{}
	q.Set("appid", p.appID)
	q.Set("secret", p.appSecret)
	q.Set("code", code)
	q.Set("grant_type", "authorization_code")

	body, err := p.doGet(ctx, p.accessTokenURL, q.Encode())
	if err != nil {
		return nil, err
	}
	if ec := parseWeChatErrcode(body); ec != 0 {
		return nil, fmt.Errorf("oauth wechat: exchange code failed: errcode=%d errmsg=%s", ec, parseWeChatErrmsg(body))
	}
	var tok wechatToken
	if err := json.Unmarshal(body, &tok); err != nil {
		return nil, fmt.Errorf("oauth wechat: decode access_token response: %w (body=%s)", err, string(body))
	}
	if tok.AccessToken == "" || tok.OpenID == "" {
		return nil, fmt.Errorf("oauth wechat: empty access_token/openid in response (body=%s)", string(body))
	}
	return &tok, nil
}

// fetchUserInfo 第 2 步：access_token + openid 换用户资料
func (p *WeChatProvider) fetchUserInfo(ctx context.Context, accessToken, openID string) (*wechatUserinfo, error) {
	q := url.Values{}
	q.Set("access_token", accessToken)
	q.Set("openid", openID)

	body, err := p.doGet(ctx, p.userInfoURL, q.Encode())
	if err != nil {
		return nil, err
	}
	if ec := parseWeChatErrcode(body); ec != 0 {
		return nil, fmt.Errorf("oauth wechat: fetch userinfo failed: errcode=%d errmsg=%s", ec, parseWeChatErrmsg(body))
	}
	var info wechatUserinfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf("oauth wechat: decode userinfo response: %w (body=%s)", err, string(body))
	}
	if info.OpenID == "" {
		return nil, fmt.Errorf("oauth wechat: empty openid in userinfo (body=%s)", string(body))
	}
	return &info, nil
}

// doGet 发送 GET 请求并返回响应体
func (p *WeChatProvider) doGet(ctx context.Context, baseURL, rawQuery string) ([]byte, error) {
	fullURL := baseURL
	if rawQuery != "" {
		fullURL = baseURL + "?" + rawQuery
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("oauth wechat: build request: %w", err)
	}
	client := p.httpClient
	if client == nil {
		client = &http.Client{Timeout: wechatTimeout}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("oauth wechat: request failed: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("oauth wechat: read body: %w", err)
	}
	return body, nil
}

// parseWeChatErrcode 从响应体解析 errcode（无则返回 0；微信成功响应不含 errcode）
func parseWeChatErrcode(body []byte) int {
	var e struct {
		ErrCode int `json:"errcode"`
	}
	_ = json.Unmarshal(body, &e)
	return e.ErrCode
}

// parseWeChatErrmsg 从响应体解析 errmsg
func parseWeChatErrmsg(body []byte) string {
	var e struct {
		ErrMsg string `json:"errmsg"`
	}
	_ = json.Unmarshal(body, &e)
	return e.ErrMsg
}

// withHTTPClient 内部测试辅助：注入 http client 与端点，返回自身便于链式调用
func (p *WeChatProvider) withHTTPClient(c *http.Client, accessTokenURL, userInfoURL string) *WeChatProvider {
	p.httpClient = c
	if accessTokenURL != "" {
		p.accessTokenURL = accessTokenURL
	}
	if userInfoURL != "" {
		p.userInfoURL = userInfoURL
	}
	return p
}

// 编译期断言：WeChatProvider 实现 Provider 接口
var _ Provider = (*WeChatProvider)(nil)
