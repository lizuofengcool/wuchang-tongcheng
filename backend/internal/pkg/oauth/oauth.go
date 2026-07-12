// Package oauth 第三方 OAuth 登录
//
// 提供：
//   - Provider 接口（mock 联调不调真实 OAuth 服务端；wechat 走开放平台网站应用 OAuth2 网页授权）
//   - Service 按 provider 名查找并完成 code → 用户身份 的换取
//
// 设计要点（与 pkg/sms / pkg/sts 风格一致）：
//   - 标准库 net/http，无新外部依赖
//   - 未配置/占位值统一降级：wechat 缺 AppID/AppSecret 或占位 → 不注册，Login 返回 ErrNotConfigured
//   - mock provider 便于前端/联调：code 形如 "mock:<openid>[:<nickname>]" 直接构造身份，不访问网络
//
// 与 user 模块的边界：本包只负责 "code → 第三方身份"，user service 负责 "身份 → 本地用户 → JWT"。
package oauth

import (
	"context"
	"errors"
	"strings"

	"wuchang-tongcheng/internal/pkg/config"
)

// 错误
var (
	// ErrNotConfigured provider 未注册或配置不全，调用方应回退其它登录方式
	ErrNotConfigured = errors.New("oauth: provider not configured")
	// ErrUnknownProvider 传入了未注册的 provider 名
	ErrUnknownProvider = errors.New("oauth: unknown provider")
	// ErrInvalidCode code 为空或格式非法
	ErrInvalidCode = errors.New("oauth: invalid or empty code")
)

// UserInfo 第三方 OAuth 返回的用户身份
type UserInfo struct {
	Provider string // 提供商名，如 wechat
	OpenID   string // 第三方 OpenID（同一应用内唯一）
	UnionID  string // 联合 ID（同主体多应用唯一，可为空）
	Nickname string
	Avatar   string
}

// Provider OAuth 提供商接口
type Provider interface {
	// Name 提供商名称（如 wechat），用作 Service 的注册 key
	Name() string
	// GetUserInfo 用前端授权回调带回的 code 换取用户身份
	GetUserInfo(ctx context.Context, code string) (*UserInfo, error)
	// IsAvailable 配置是否齐全（供 Init 选择 provider 与测试使用）
	IsAvailable() bool
}

// NoopProvider 未配置占位：GetUserInfo 返回 ErrNotConfigured
type NoopProvider struct{}

func (NoopProvider) Name() string                              { return "" }
func (NoopProvider) IsAvailable() bool                         { return false }
func (NoopProvider) GetUserInfo(context.Context, string) (*UserInfo, error) {
	return nil, ErrNotConfigured
}

// MockProvider 联调用：不调用真实 OAuth 服务端
//
// code 格式：
//   - "mock:<openid>"              → 昵称默认 "微信用户"
//   - "mock:<openid>:<nickname>"   → 自定义昵称
//
// 便于前端在不接入真实微信的情况下打通登录 → JWT 全链路。
type MockProvider struct{}

func (MockProvider) Name() string      { return "wechat" }
func (MockProvider) IsAvailable() bool { return true }

func (MockProvider) GetUserInfo(_ context.Context, code string) (*UserInfo, error) {
	if !strings.HasPrefix(code, "mock:") {
		return nil, ErrInvalidCode
	}
	rest := strings.TrimPrefix(code, "mock:")
	parts := strings.SplitN(rest, ":", 2)
	openid := parts[0]
	if openid == "" {
		return nil, ErrInvalidCode
	}
	nickname := "微信用户"
	if len(parts) == 2 && parts[1] != "" {
		nickname = parts[1]
	}
	return &UserInfo{
		Provider: "wechat",
		OpenID:   openid,
		UnionID:  "mock_union_" + openid,
		Nickname: nickname,
	}, nil
}

// Service OAuth 登录服务，按 provider 名分发
type Service struct {
	providers map[string]Provider
}

// NewService 按配置创建 OAuth 服务
//
//	cfg == nil → 空 providers（所有 Login 返回 ErrNotConfigured）
//	wechat.provider = "mock" → 注册 MockProvider（联调用）
//	wechat.provider = "wechat" 且 AppID/AppSecret 齐全非占位 → 注册 WeChatProvider；缺失/占位降级不注册
func NewService(cfg *config.OAuthConfig) *Service {
	s := &Service{providers: make(map[string]Provider)}
	if cfg == nil {
		return s
	}
	if p := resolveWeChatProvider(&cfg.WeChat); p != nil {
		s.providers[p.Name()] = p
	}
	return s
}

// resolveWeChatProvider 解析微信 OAuth provider，未配置/占位返回 nil（不注册）
func resolveWeChatProvider(cfg *config.WeChatConfig) Provider {
	if cfg == nil {
		return nil
	}
	switch strings.ToLower(cfg.Provider) {
	case "mock":
		return MockProvider{}
	case "wechat":
		p := NewWeChatProvider(cfg)
		if !p.IsAvailable() {
			return nil
		}
		return p
	default: // "" → 不启用
		return nil
	}
}

// HasProvider 是否注册了指定 provider
func (s *Service) HasProvider(name string) bool {
	_, ok := s.providers[name]
	return ok
}

// Login 用 code 换取第三方用户身份
func (s *Service) Login(ctx context.Context, providerName, code string) (*UserInfo, error) {
	if code == "" {
		return nil, ErrInvalidCode
	}
	p, ok := s.providers[providerName]
	if !ok {
		return nil, ErrNotConfigured
	}
	return p.GetUserInfo(ctx, code)
}
