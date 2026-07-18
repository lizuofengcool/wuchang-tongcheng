// Package service 用户业务逻辑层
package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"

	"wuchang-tongcheng/internal/modules/user/dto"
	"wuchang-tongcheng/internal/modules/user/model"
	"wuchang-tongcheng/internal/modules/user/repository"
	"wuchang-tongcheng/internal/pkg/jwt"
	"wuchang-tongcheng/internal/pkg/oauth"
	"wuchang-tongcheng/internal/pkg/utils"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrUserAlreadyExists = errors.New("用户名已存在")
	ErrUserNotFound      = errors.New("用户不存在")
	ErrPasswordInvalid   = errors.New("密码错误")
	ErrUserDisabled      = errors.New("用户已被禁用")
	ErrOldPasswordWrong  = errors.New("原密码错误")
	ErrSMSCodeInvalid    = errors.New("验证码错误或已过期")
	ErrSMSNotConfigured  = errors.New("短信服务未启用")
	ErrOAuthNotConfigured = errors.New("第三方登录未启用")
)

// SMSService 短信验证码服务接口（由 pkg/sms.Service 实现）
type SMSService interface {
	SendCode(ctx context.Context, phone string) (string, error)
	Verify(ctx context.Context, phone, code string) error
}

// OAuthService 第三方 OAuth 登录服务接口（由 pkg/oauth.Service 实现）
type OAuthService interface {
	Login(ctx context.Context, provider, code string) (*oauth.UserInfo, error)
	HasProvider(name string) bool
}

// UserService 用户业务逻辑接口
type UserService interface {
	Register(regionID uint, req *dto.RegisterRequest) (*dto.UserInfo, error)
	Login(req *dto.LoginRequest) (*dto.LoginResponse, error)
	SendSMSCode(ctx context.Context, phone string) (*dto.SendSMSCodeResponse, error)
	LoginBySMS(ctx context.Context, regionID uint, phone, code string) (*dto.LoginResponse, error)
	// OAuthLogin 第三方 OAuth 登录：code 换取身份 → 命中绑定则登录，否则自动注册并绑定
	OAuthLogin(ctx context.Context, regionID uint, provider, code string) (*dto.LoginResponse, error)
	GetUserInfo(userID uint) (*dto.UserInfo, error)
	UpdateProfile(userID uint, req *dto.UpdateProfileRequest) error
	ChangePassword(userID uint, req *dto.ChangePasswordRequest) error
	// 管理后台
	ListUsers(regionID uint, req *dto.ListUsersRequest) (*utils.Pagination, []dto.UserInfo, error)
	AdminCreateUser(regionID uint, req *dto.AdminCreateUserRequest) (*dto.UserInfo, error)
	AdminUpdateUser(id uint, req *dto.AdminUpdateUserRequest) error
	UpdateUserStatus(id uint, status int) error
	ResetPassword(id uint, req *dto.ResetPasswordRequest) error
	DeleteUser(id uint) error
}

type userService struct {
	userRepo  repository.UserRepository
	oauthRepo repository.UserOAuthRepository
	sms       SMSService
	oauth     OAuthService
}

// NewUserService 创建用户服务
//
// sms 可为 nil：未启用短信服务时调用 SendSMSCode/LoginBySMS 返回 ErrSMSNotConfigured
// oauthRepo/oauth 可为 nil：未启用第三方登录时调用 OAuthLogin 返回 ErrOAuthNotConfigured
func NewUserService(
	userRepo repository.UserRepository,
	sms SMSService,
	oauthRepo repository.UserOAuthRepository,
	oauth OAuthService,
) UserService {
	return &userService{userRepo: userRepo, sms: sms, oauthRepo: oauthRepo, oauth: oauth}
}

// HashPassword 密码哈希
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword 校验密码
func CheckPassword(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// toUserInfo 转换为用户信息DTO
func toUserInfo(user *model.User) *dto.UserInfo {
	return &dto.UserInfo{
		ID:        user.ID,
		Username:  user.Username,
		Nickname:  user.Nickname,
		Avatar:    user.Avatar,
		Phone:     user.Phone,
		Email:     user.Email,
		Gender:    user.Gender,
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
	}
}

// Register 用户注册（按请求头 X-Region-ID 写入用户所属地区）
func (s *userService) Register(regionID uint, req *dto.RegisterRequest) (*dto.UserInfo, error) {
	// 检查用户名是否已存在
	if _, err := s.userRepo.FindByUsername(req.Username); err == nil {
		return nil, ErrUserAlreadyExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 密码哈希
	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// 昵称默认为用户名
	nickname := req.Nickname
	if nickname == "" {
		nickname = req.Username
	}

	user := &model.User{
		Username: req.Username,
		Password: hashedPassword,
		Nickname: nickname,
		Phone:    req.Phone,
		Status:   1,
	}
	user.RegionID = regionID

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return toUserInfo(user), nil
}

// Login 用户登录
func (s *userService) Login(req *dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := s.userRepo.FindByUsername(req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	// 检查用户状态
	if user.Status == 0 {
		return nil, ErrUserDisabled
	}

	// 校验密码
	if !CheckPassword(req.Password, user.Password) {
		return nil, ErrPasswordInvalid
	}

	// 生成Token
	token, err := jwt.GenerateToken(user.ID, user.Username)
	if err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		Token:    token,
		Expires:  24 * 3600,
		UserInfo: *toUserInfo(user),
	}, nil
}

// SendSMSCode 发送短信验证码
// 返回的 SendSMSCodeResponse.DevCode 仅在 mock provider + dev_return_code=true 时非空（联调用）
func (s *userService) SendSMSCode(ctx context.Context, phone string) (*dto.SendSMSCodeResponse, error) {
	if s.sms == nil {
		return nil, ErrSMSNotConfigured
	}
	devCode, err := s.sms.SendCode(ctx, phone)
	if err != nil {
		return nil, err
	}
	return &dto.SendSMSCodeResponse{DevCode: devCode}, nil
}

// LoginBySMS 短信验证码登录
// 流程：校验验证码 → 按手机号查用户 → 存在则直接签发 token；不存在则自动注册（用户名=手机号，随机密码，无法走密码登录）
func (s *userService) LoginBySMS(ctx context.Context, regionID uint, phone, code string) (*dto.LoginResponse, error) {
	if s.sms == nil {
		return nil, ErrSMSNotConfigured
	}
	// 校验验证码（成功后一次性删除）
	if err := s.sms.Verify(ctx, phone, code); err != nil {
		return nil, ErrSMSCodeInvalid
	}

	// 按手机号查找用户
	user, err := s.userRepo.FindByPhone(phone)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		// 新手机号：自动注册（用户名=手机号，随机密码占位，无法用密码登录）
		randPwd, perr := randomPassword(32)
		if perr != nil {
			return nil, perr
		}
		hashedPassword, perr := HashPassword(randPwd)
		if perr != nil {
			return nil, perr
		}
		user = &model.User{
			Username: phone,
			Password: hashedPassword,
			Nickname: phone,
			Phone:    phone,
			Status:   1,
		}
		user.RegionID = regionID
		if err := s.userRepo.Create(user); err != nil {
			return nil, err
		}
	}

	// 禁用用户拒绝登录
	if user.Status == 0 {
		return nil, ErrUserDisabled
	}

	token, err := jwt.GenerateToken(user.ID, user.Username)
	if err != nil {
		return nil, err
	}
	return &dto.LoginResponse{
		Token:    token,
		Expires:  24 * 3600,
		UserInfo: *toUserInfo(user),
	}, nil
}

// randomPassword 生成指定长度的十六进制随机字符串（用于自动注册用户的占位密码）
func randomPassword(n int) (string, error) {
	if n <= 0 {
		n = 32
	}
	b := make([]byte, n/2+1)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate random password: %w", err)
	}
	return hex.EncodeToString(b)[:n], nil
}

// genOAuthUsername 生成第三方登录自动注册用户的用户名（provider_openid，openid 超长截断）
// 受 users.username size:50 约束，openid 最多保留 40 字符
func genOAuthUsername(provider, openid string) string {
	const maxOpenID = 40
	oid := openid
	if len(oid) > maxOpenID {
		oid = oid[:maxOpenID]
	}
	return provider + "_" + oid
}

// OAuthLogin 第三方 OAuth 登录
//
// 流程：用 code 换取第三方身份 → 按 (provider, openid) 查绑定 → 命中则登录对应用户；
// 未命中则自动注册（用户名=provider_openid，随机占位密码无法走密码登录）并写入绑定，最后签发 JWT。
// 禁用用户拒绝登录。oauth/oauthRepo 未注入或 provider 未注册时返回 ErrOAuthNotConfigured。
func (s *userService) OAuthLogin(ctx context.Context, regionID uint, provider, code string) (*dto.LoginResponse, error) {
	if s.oauth == nil || s.oauthRepo == nil {
		return nil, ErrOAuthNotConfigured
	}
	if !s.oauth.HasProvider(provider) {
		return nil, ErrOAuthNotConfigured
	}

	identity, err := s.oauth.Login(ctx, provider, code)
	if err != nil {
		return nil, err
	}

	// 按 (provider, openid) 查绑定
	binding, err := s.oauthRepo.FindByProviderOpenID(provider, identity.OpenID)
	if err == nil {
		user, ferr := s.userRepo.FindByID(binding.UserID)
		if ferr != nil {
			return nil, ferr
		}
		return s.loginAndIssueToken(user)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 全新用户：自动注册 + 绑定
	randPwd, perr := randomPassword(32)
	if perr != nil {
		return nil, perr
	}
	hashedPassword, perr := HashPassword(randPwd)
	if perr != nil {
		return nil, perr
	}
	nickname := identity.Nickname
	if nickname == "" {
		nickname = provider + "_" + identity.OpenID
	}
	user := &model.User{
		Username: genOAuthUsername(provider, identity.OpenID),
		Password: hashedPassword,
		Nickname: nickname,
		Avatar:   identity.Avatar,
		Status:   1,
	}
	user.RegionID = regionID
	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}
	binding = &model.UserOAuth{
		UserID:   user.ID,
		Provider: provider,
		OpenID:   identity.OpenID,
		UnionID:  identity.UnionID,
		Nickname: identity.Nickname,
		Avatar:   identity.Avatar,
	}
	if err := s.oauthRepo.Create(binding); err != nil {
		return nil, err
	}
	return s.loginAndIssueToken(user)
}

// loginAndIssueToken 校验用户状态并签发 JWT（OAuth 登录复用）
func (s *userService) loginAndIssueToken(user *model.User) (*dto.LoginResponse, error) {
	if user.Status == 0 {
		return nil, ErrUserDisabled
	}
	token, err := jwt.GenerateToken(user.ID, user.Username)
	if err != nil {
		return nil, err
	}
	return &dto.LoginResponse{
		Token:    token,
		Expires:  24 * 3600,
		UserInfo: *toUserInfo(user),
	}, nil
}

// GetUserInfo 获取用户信息
func (s *userService) GetUserInfo(userID uint) (*dto.UserInfo, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return toUserInfo(user), nil
}

// UpdateProfile 更新个人资料
func (s *userService) UpdateProfile(userID uint, req *dto.UpdateProfileRequest) error {
	fields := map[string]interface{}{}
	if req.Nickname != "" {
		fields["nickname"] = req.Nickname
	}
	if req.Avatar != "" {
		fields["avatar"] = req.Avatar
	}
	if req.Phone != "" {
		fields["phone"] = req.Phone
	}
	if req.Email != "" {
		fields["email"] = req.Email
	}
	if req.Gender != 0 || len(fields) > 0 {
		fields["gender"] = req.Gender
	}

	if len(fields) == 0 {
		return nil
	}
	return s.userRepo.UpdateFields(userID, fields)
}

// ChangePassword 修改密码
func (s *userService) ChangePassword(userID uint, req *dto.ChangePasswordRequest) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	// 校验原密码
	if !CheckPassword(req.OldPassword, user.Password) {
		return ErrOldPasswordWrong
	}

	// 哈希新密码
	hashedPassword, err := HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	return s.userRepo.UpdateFields(userID, map[string]interface{}{
		"password": hashedPassword,
	})
}

// ===== 管理后台 =====

// ListUsers 用户列表（按地区隔离：regionID=0 表示不限制，超管跨区查看）
func (s *userService) ListUsers(regionID uint, req *dto.ListUsersRequest) (*utils.Pagination, []dto.UserInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	list, total, err := s.userRepo.List(regionID, pagination, req.Keyword, req.Status)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total

	result := make([]dto.UserInfo, 0, len(list))
	for i := range list {
		result = append(result, *toUserInfo(&list[i]))
	}
	return pagination, result, nil
}

// AdminCreateUser 管理员创建用户（按管理员所在地区写入 region_id）
func (s *userService) AdminCreateUser(regionID uint, req *dto.AdminCreateUserRequest) (*dto.UserInfo, error) {
	if _, err := s.userRepo.FindByUsername(req.Username); err == nil {
		return nil, ErrUserAlreadyExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	nickname := req.Nickname
	if nickname == "" {
		nickname = req.Username
	}
	status := req.Status
	if status == 0 {
		status = 1
	}

	user := &model.User{
		Username: req.Username,
		Password: hashedPassword,
		Nickname: nickname,
		Phone:    req.Phone,
		Email:    req.Email,
		Gender:   req.Gender,
		Status:   status,
	}
	user.RegionID = regionID

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}
	return toUserInfo(user), nil
}

// AdminUpdateUser 管理员更新用户资料
func (s *userService) AdminUpdateUser(id uint, req *dto.AdminUpdateUserRequest) error {
	fields := map[string]interface{}{}
	if req.Nickname != "" {
		fields["nickname"] = req.Nickname
	}
	if req.Avatar != "" {
		fields["avatar"] = req.Avatar
	}
	if req.Phone != "" {
		fields["phone"] = req.Phone
	}
	if req.Email != "" {
		fields["email"] = req.Email
	}
	fields["gender"] = req.Gender
	return s.userRepo.UpdateFields(id, fields)
}

// UpdateUserStatus 更新用户状态（启用/禁用）
func (s *userService) UpdateUserStatus(id uint, status int) error {
	return s.userRepo.UpdateFields(id, map[string]interface{}{"status": status})
}

// ResetPassword 管理员重置用户密码
func (s *userService) ResetPassword(id uint, req *dto.ResetPasswordRequest) error {
	hashedPassword, err := HashPassword(req.NewPassword)
	if err != nil {
		return err
	}
	return s.userRepo.UpdateFields(id, map[string]interface{}{"password": hashedPassword})
}

// DeleteUser 删除用户
func (s *userService) DeleteUser(id uint) error {
	if _, err := s.userRepo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}
	return s.userRepo.Delete(id)
}

// 引用utils避免未使用导入（保留备用）
var _ = utils.MD5
