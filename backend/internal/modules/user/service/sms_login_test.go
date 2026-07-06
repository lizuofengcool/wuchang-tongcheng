package service

import (
	"context"
	"errors"
	"sync"
	"testing"

	"wuchang-tongcheng/internal/modules/user/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// fakeSMSService SMSService 桩件
type fakeSMSService struct {
	mu       sync.Mutex
	sendErr  error
	devCode  string
	verifyOK bool
	verifyErr error
	verifyd  bool
}

func (f *fakeSMSService) SendCode(ctx context.Context, phone string) (string, error) {
	if f.sendErr != nil {
		return "", f.sendErr
	}
	return f.devCode, nil
}

func (f *fakeSMSService) Verify(ctx context.Context, phone, code string) error {
	f.mu.Lock()
	f.verifyd = true
	f.mu.Unlock()
	if f.verifyErr != nil {
		return f.verifyErr
	}
	if !f.verifyOK && code != "correct" {
		return errors.New("验证码错误")
	}
	return nil
}

// fakeUserRepo UserRepository 桩件
type fakeUserRepo struct {
	users      map[string]*model.User // key: phone
	createErr  error
	created    []*model.User
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{users: make(map[string]*model.User)}
}

func (r *fakeUserRepo) Create(user *model.User) error {
	if r.createErr != nil {
		return r.createErr
	}
	r.users[user.Phone] = user
	r.created = append(r.created, user)
	return nil
}
func (r *fakeUserRepo) FindByID(id uint) (*model.User, error) {
	for _, u := range r.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}
func (r *fakeUserRepo) FindByUsername(username string) (*model.User, error) {
	for _, u := range r.users {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}
func (r *fakeUserRepo) FindByPhone(phone string) (*model.User, error) {
	if u, ok := r.users[phone]; ok {
		return u, nil
	}
	return nil, gorm.ErrRecordNotFound
}
func (r *fakeUserRepo) Update(user *model.User) error         { return nil }
func (r *fakeUserRepo) UpdateFields(id uint, fields map[string]interface{}) error {
	return nil
}
func (r *fakeUserRepo) List(regionID uint, p *utils.Pagination, keyword string, status int) ([]model.User, int64, error) {
	return nil, 0, nil
}
func (r *fakeUserRepo) Delete(id uint) error { return nil }

// TestSendSMSCode_NilSMS sms 未启用时返回 ErrSMSNotConfigured
func TestSendSMSCode_NilSMS(t *testing.T) {
	svc := NewUserService(newFakeUserRepo(), nil)
	_, err := svc.SendSMSCode(context.Background(), "13800138000")
	if !errors.Is(err, ErrSMSNotConfigured) {
		t.Fatalf("err = %v, want ErrSMSNotConfigured", err)
	}
}

// TestSendSMSCode_OK sms 启用时透传 dev code
func TestSendSMSCode_OK(t *testing.T) {
	sms := &fakeSMSService{devCode: "123456"}
	svc := NewUserService(newFakeUserRepo(), sms)
	resp, err := svc.SendSMSCode(context.Background(), "13800138000")
	if err != nil {
		t.Fatalf("SendSMSCode: %v", err)
	}
	if resp.DevCode != "123456" {
		t.Fatalf("DevCode = %q, want 123456", resp.DevCode)
	}
}

// TestLoginBySMS_VerifyFailed 验证码错误时返回 ErrSMSCodeInvalid 且不查库
func TestLoginBySMS_VerifyFailed(t *testing.T) {
	repo := newFakeUserRepo()
	sms := &fakeSMSService{verifyErr: errors.New("bad code")}
	svc := NewUserService(repo, sms)
	_, err := svc.LoginBySMS(context.Background(), 1, "13800138000", "wrong")
	if !errors.Is(err, ErrSMSCodeInvalid) {
		t.Fatalf("err = %v, want ErrSMSCodeInvalid", err)
	}
	if len(repo.created) != 0 {
		t.Fatalf("should not create user on verify failure, created %d", len(repo.created))
	}
}

// TestLoginBySMS_NewUserAutoRegister 新手机号验证通过 → 自动注册并签发 token
func TestLoginBySMS_NewUserAutoRegister(t *testing.T) {
	repo := newFakeUserRepo()
	sms := &fakeSMSService{verifyOK: true}
	svc := NewUserService(repo, sms)
	resp, err := svc.LoginBySMS(context.Background(), 1, "13800138000", "correct")
	if err != nil {
		t.Fatalf("LoginBySMS: %v", err)
	}
	if resp.Token == "" {
		t.Fatal("expected non-empty token")
	}
	if resp.UserInfo.Phone != "13800138000" {
		t.Fatalf("phone = %q, want 13800138000", resp.UserInfo.Phone)
	}
	if resp.UserInfo.Username != "13800138000" {
		t.Fatalf("username = %q, want 13800138000", resp.UserInfo.Username)
	}
	if len(repo.created) != 1 {
		t.Fatalf("expected 1 created user, got %d", len(repo.created))
	}
}

// TestLoginBySMS_ExistingUser 老用户验证通过 → 直接登录，不新建
func TestLoginBySMS_ExistingUser(t *testing.T) {
	repo := newFakeUserRepo()
	exist := &model.User{}
	exist.ID = 42
	exist.Username = "existing"
	exist.Phone = "13800138000"
	exist.Status = 1
	repo.users["13800138000"] = exist

	sms := &fakeSMSService{verifyOK: true}
	svc := NewUserService(repo, sms)
	resp, err := svc.LoginBySMS(context.Background(), 1, "13800138000", "correct")
	if err != nil {
		t.Fatalf("LoginBySMS: %v", err)
	}
	if resp.UserInfo.ID != 42 {
		t.Fatalf("user id = %d, want 42", resp.UserInfo.ID)
	}
	if len(repo.created) != 0 {
		t.Fatalf("should not create user for existing phone, created %d", len(repo.created))
	}
}

// TestLoginBySMS_DisabledUser 禁用用户拒绝登录
func TestLoginBySMS_DisabledUser(t *testing.T) {
	repo := newFakeUserRepo()
	exist := &model.User{}
	exist.ID = 7
	exist.Username = "disabled"
	exist.Phone = "13800138000"
	exist.Status = 0
	repo.users["13800138000"] = exist

	sms := &fakeSMSService{verifyOK: true}
	svc := NewUserService(repo, sms)
	_, err := svc.LoginBySMS(context.Background(), 1, "13800138000", "correct")
	if !errors.Is(err, ErrUserDisabled) {
		t.Fatalf("err = %v, want ErrUserDisabled", err)
	}
}
