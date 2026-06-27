// Package service 用户业务逻辑层
package service

import (
	"errors"

	"wuchang-tongcheng/internal/modules/user/dto"
	"wuchang-tongcheng/internal/modules/user/model"
	"wuchang-tongcheng/internal/modules/user/repository"
	"wuchang-tongcheng/internal/pkg/jwt"
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
)

// UserService 用户业务逻辑接口
type UserService interface {
	Register(req *dto.RegisterRequest) (*dto.UserInfo, error)
	Login(req *dto.LoginRequest) (*dto.LoginResponse, error)
	GetUserInfo(userID uint) (*dto.UserInfo, error)
	UpdateProfile(userID uint, req *dto.UpdateProfileRequest) error
	ChangePassword(userID uint, req *dto.ChangePasswordRequest) error
	// 管理后台
	ListUsers(req *dto.ListUsersRequest) (*utils.Pagination, []dto.UserInfo, error)
	AdminCreateUser(req *dto.AdminCreateUserRequest) (*dto.UserInfo, error)
	AdminUpdateUser(id uint, req *dto.AdminUpdateUserRequest) error
	UpdateUserStatus(id uint, status int) error
	ResetPassword(id uint, req *dto.ResetPasswordRequest) error
	DeleteUser(id uint) error
}

type userService struct {
	userRepo repository.UserRepository
}

// NewUserService 创建用户服务
func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo: userRepo}
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

// Register 用户注册
func (s *userService) Register(req *dto.RegisterRequest) (*dto.UserInfo, error) {
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

// ListUsers 用户列表
func (s *userService) ListUsers(req *dto.ListUsersRequest) (*utils.Pagination, []dto.UserInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	list, total, err := s.userRepo.List(pagination, req.Keyword, req.Status)
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

// AdminCreateUser 管理员创建用户
func (s *userService) AdminCreateUser(req *dto.AdminCreateUserRequest) (*dto.UserInfo, error) {
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
