package service

import (
	"context"
	"errors"
	"sync"
	"testing"

	"wuchang-tongcheng/internal/modules/user/model"
	"wuchang-tongcheng/internal/pkg/oauth"

	"gorm.io/gorm"
)

// fakeOAuthService OAuthService 桩件
type fakeOAuthService struct {
	mu          sync.Mutex
	identity    *oauth.UserInfo
	err         error
	hasProvider bool
	called      bool
}

func (f *fakeOAuthService) Login(_ context.Context, provider, code string) (*oauth.UserInfo, error) {
	f.mu.Lock()
	f.called = true
	f.mu.Unlock()
	if f.err != nil {
		return nil, f.err
	}
	if f.identity != nil {
		return f.identity, nil
	}
	// 默认返回一个固定身份，便于测试
	return &oauth.UserInfo{
		Provider: provider,
		OpenID:   "wx_openid_001",
		UnionID:  "wx_unionid_001",
		Nickname: "微信用户",
		Avatar:   "https://wx.qlogo.cn/a.png",
	}, nil
}

func (f *fakeOAuthService) HasProvider(name string) bool { return f.hasProvider }

// fakeOAuthRepo UserOAuthRepository 桩件
type fakeOAuthRepo struct {
	bindings  map[string]*model.UserOAuth // key: provider + ":" + openid
	createErr error
	created   []*model.UserOAuth
}

func newFakeOAuthRepo() *fakeOAuthRepo {
	return &fakeOAuthRepo{bindings: make(map[string]*model.UserOAuth)}
}

func (r *fakeOAuthRepo) FindByProviderOpenID(provider, openID string) (*model.UserOAuth, error) {
	if b, ok := r.bindings[provider+":"+openID]; ok {
		return b, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *fakeOAuthRepo) Create(b *model.UserOAuth) error {
	if r.createErr != nil {
		return r.createErr
	}
	r.bindings[b.Provider+":"+b.OpenID] = b
	r.created = append(r.created, b)
	return nil
}

// TestOAuthLogin_NotConfigured oauth 未注入 → ErrOAuthNotConfigured
func TestOAuthLogin_NotConfigured(t *testing.T) {
	svc := NewUserService(newFakeUserRepo(), nil, nil, nil)
	_, err := svc.OAuthLogin(context.Background(), 1, "wechat", "mock:x")
	if !errors.Is(err, ErrOAuthNotConfigured) {
		t.Fatalf("err=%v want ErrOAuthNotConfigured", err)
	}
}

// TestOAuthLogin_ProviderNotRegistered provider 未注册 → ErrOAuthNotConfigured
func TestOAuthLogin_ProviderNotRegistered(t *testing.T) {
	o := &fakeOAuthService{hasProvider: false}
	svc := NewUserService(newFakeUserRepo(), nil, newFakeOAuthRepo(), o)
	_, err := svc.OAuthLogin(context.Background(), 1, "wechat", "mock:x")
	if !errors.Is(err, ErrOAuthNotConfigured) {
		t.Fatalf("err=%v want ErrOAuthNotConfigured", err)
	}
}

// TestOAuthLogin_OAuthError oauth.Login 失败透传
func TestOAuthLogin_OAuthError(t *testing.T) {
	o := &fakeOAuthService{hasProvider: true, err: errors.New("wechat api timeout")}
	svc := NewUserService(newFakeUserRepo(), nil, newFakeOAuthRepo(), o)
	_, err := svc.OAuthLogin(context.Background(), 1, "wechat", "code")
	if err == nil {
		t.Fatal("expected oauth error")
	}
	if err.Error() != "wechat api timeout" {
		t.Fatalf("err=%v", err)
	}
}

// TestOAuthLogin_ExistingBinding 命中绑定 → 登录既有用户，不新建
func TestOAuthLogin_ExistingBinding(t *testing.T) {
	repo := newFakeUserRepo()
	exist := &model.User{}
	exist.ID = 42
	exist.Username = "wechat_existing"
	exist.Nickname = "老王"
	exist.Phone = "13800138000"
	exist.Status = 1
	repo.users["13800138000"] = exist

	oauthRepo := newFakeOAuthRepo()
	oauthRepo.bindings["wechat:wx_openid_001"] = &model.UserOAuth{
		UserID:   42,
		Provider: "wechat",
		OpenID:   "wx_openid_001",
	}

	o := &fakeOAuthService{hasProvider: true}
	svc := NewUserService(repo, nil, oauthRepo, o)
	resp, err := svc.OAuthLogin(context.Background(), 1, "wechat", "anycode")
	if err != nil {
		t.Fatalf("OAuthLogin: %v", err)
	}
	if resp.Token == "" {
		t.Fatal("expected token")
	}
	if resp.UserInfo.ID != 42 {
		t.Fatalf("user id=%d want 42", resp.UserInfo.ID)
	}
	if len(repo.created) != 0 {
		t.Fatalf("should not create user, created %d", len(repo.created))
	}
	if len(oauthRepo.created) != 0 {
		t.Fatalf("should not create binding, created %d", len(oauthRepo.created))
	}
}

// TestOAuthLogin_NewUserAutoRegister 无绑定 → 自动注册用户 + 写绑定 + 签发 token
func TestOAuthLogin_NewUserAutoRegister(t *testing.T) {
	repo := newFakeUserRepo()
	oauthRepo := newFakeOAuthRepo()
	o := &fakeOAuthService{hasProvider: true}
	svc := NewUserService(repo, nil, oauthRepo, o)

	resp, err := svc.OAuthLogin(context.Background(), 2, "wechat", "code")
	if err != nil {
		t.Fatalf("OAuthLogin: %v", err)
	}
	if resp.Token == "" {
		t.Fatal("expected token")
	}
	if resp.UserInfo.Username != "wechat_wx_openid_001" {
		t.Fatalf("username=%q want wechat_wx_openid_001", resp.UserInfo.Username)
	}
	if resp.UserInfo.Nickname != "微信用户" {
		t.Fatalf("nickname=%q want 微信用户", resp.UserInfo.Nickname)
	}
	if resp.UserInfo.Avatar != "https://wx.qlogo.cn/a.png" {
		t.Fatalf("avatar=%q", resp.UserInfo.Avatar)
	}
	if len(repo.created) != 1 {
		t.Fatalf("expected 1 created user, got %d", len(repo.created))
	}
	if len(oauthRepo.created) != 1 {
		t.Fatalf("expected 1 created binding, got %d", len(oauthRepo.created))
	}
	b := oauthRepo.created[0]
	if b.Provider != "wechat" || b.OpenID != "wx_openid_001" {
		t.Fatalf("binding=%+v", b)
	}
	if b.UnionID != "wx_unionid_001" {
		t.Fatalf("unionid=%q", b.UnionID)
	}
}

// TestOAuthLogin_NewUserEmptyNickname OAuth 返回空昵称 → 用 provider_openid 兜底
func TestOAuthLogin_NewUserEmptyNickname(t *testing.T) {
	repo := newFakeUserRepo()
	oauthRepo := newFakeOAuthRepo()
	o := &fakeOAuthService{
		hasProvider: true,
		identity: &oauth.UserInfo{
			Provider: "wechat",
			OpenID:   "o_abc",
			Nickname: "",
		},
	}
	svc := NewUserService(repo, nil, oauthRepo, o)
	resp, err := svc.OAuthLogin(context.Background(), 1, "wechat", "code")
	if err != nil {
		t.Fatalf("OAuthLogin: %v", err)
	}
	if resp.UserInfo.Nickname != "wechat_o_abc" {
		t.Fatalf("nickname=%q want wechat_o_abc", resp.UserInfo.Nickname)
	}
}

// TestOAuthLogin_DisabledUser 命中绑定但用户禁用 → ErrUserDisabled
func TestOAuthLogin_DisabledUser(t *testing.T) {
	repo := newFakeUserRepo()
	exist := &model.User{}
	exist.ID = 7
	exist.Username = "banned"
	exist.Phone = "13900139000"
	exist.Status = 0
	repo.users["13900139000"] = exist

	oauthRepo := newFakeOAuthRepo()
	oauthRepo.bindings["wechat:o_banned"] = &model.UserOAuth{UserID: 7, Provider: "wechat", OpenID: "o_banned"}

	o := &fakeOAuthService{
		hasProvider: true,
		identity:    &oauth.UserInfo{Provider: "wechat", OpenID: "o_banned", Nickname: "x"},
	}
	svc := NewUserService(repo, nil, oauthRepo, o)
	_, err := svc.OAuthLogin(context.Background(), 1, "wechat", "code")
	if !errors.Is(err, ErrUserDisabled) {
		t.Fatalf("err=%v want ErrUserDisabled", err)
	}
}

// TestOAuthLogin_BindingCreateFailed 写绑定失败 → 透传错误
func TestOAuthLogin_BindingCreateFailed(t *testing.T) {
	repo := newFakeUserRepo()
	oauthRepo := newFakeOAuthRepo()
	oauthRepo.createErr = errors.New("db write failed")
	o := &fakeOAuthService{hasProvider: true}
	svc := NewUserService(repo, nil, oauthRepo, o)
	_, err := svc.OAuthLogin(context.Background(), 1, "wechat", "code")
	if err == nil || err.Error() != "db write failed" {
		t.Fatalf("err=%v", err)
	}
}

// TestGenOAuthUsername 用户名生成 + 超长 openid 截断
func TestGenOAuthUsername(t *testing.T) {
	if got := genOAuthUsername("wechat", "o_123"); got != "wechat_o_123" {
		t.Fatalf("got=%q", got)
	}
	long := ""
	for i := 0; i < 100; i++ {
		long += "x"
	}
	got := genOAuthUsername("wechat", long)
	// "wechat_" (7) + 40 = 47，<= users.username size 50
	if len(got) != 47 {
		t.Fatalf("len=%d want 47", len(got))
	}
}
