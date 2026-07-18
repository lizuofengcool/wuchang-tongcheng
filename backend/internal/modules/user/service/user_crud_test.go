// Package service 用户业务逻辑单元测试（CRUD 主流程）。
//
// 与既有 user_test.go（HashPassword/CheckPassword 纯函数）、sms_login_test.go、
// oauth_login_test.go 互补，使用内存 mockUserRepo 覆盖 Register/Login/GetUserInfo/
// UpdateProfile/ChangePassword/ListUsers/AdminCreateUser/AdminUpdateUser/
// UpdateUserStatus/ResetPassword/DeleteUser 业务逻辑，不依赖 DB/Redis。
//
// mock 策略沿用 category/region/news service 风格：内存 mockUserRepo 实现
// UserRepository 全部 9 方法，byID map + nextID 自增 + 各路径错误注入字段；
// bcrypt 哈希真实执行（HashPassword/CheckPassword 已在 user_test.go 验证）；
// JWT GenerateToken 使用包级默认 secretKey，无需 Init。
package service

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"wuchang-tongcheng/internal/modules/user/dto"
	"wuchang-tongcheng/internal/modules/user/model"
	"wuchang-tongcheng/internal/pkg/utils"
)

// mockUserRepo 内存 mock，实现 UserRepository 接口（与 sms_login_test.go 的
// fakeUserRepo 区别：支持 Update/UpdateFields/List/Delete 的真实语义以便覆盖
// 管理后台路径）。
type mockUserRepo struct {
	byID   map[uint]*model.User
	nextID uint

	// 错误注入
	createErr          error
	findByIDErr        error
	findByUsernameErr  error
	findByPhoneErr     error
	updateErr          error
	updateFieldsErr    error
	listErr            error
	deleteErr          error

	// 调用记录
	lastListRegionID uint
	lastListKeyword  string
	lastListStatus   int
	lastListPage     *utils.Pagination
	updatedFields    map[uint]map[string]interface{}
	deletedIDs       []uint
	createdCount     int
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		byID:           make(map[uint]*model.User),
		nextID:         1,
		updatedFields:  make(map[uint]map[string]interface{}),
	}
}

func (m *mockUserRepo) Create(user *model.User) error {
	if m.createErr != nil {
		return m.createErr
	}
	user.ID = m.nextID
	m.nextID++
	m.createdCount++
	cp := *user
	m.byID[user.ID] = &cp
	return nil
}

func (m *mockUserRepo) FindByID(id uint) (*model.User, error) {
	if m.findByIDErr != nil {
		return nil, m.findByIDErr
	}
	u, ok := m.byID[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *u
	return &cp, nil
}

func (m *mockUserRepo) FindByUsername(username string) (*model.User, error) {
	if m.findByUsernameErr != nil {
		return nil, m.findByUsernameErr
	}
	for _, u := range m.byID {
		if u.Username == username {
			cp := *u
			return &cp, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockUserRepo) FindByPhone(phone string) (*model.User, error) {
	if m.findByPhoneErr != nil {
		return nil, m.findByPhoneErr
	}
	for _, u := range m.byID {
		if u.Phone == phone {
			cp := *u
			return &cp, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockUserRepo) Update(user *model.User) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if _, ok := m.byID[user.ID]; !ok {
		return gorm.ErrRecordNotFound
	}
	cp := *user
	m.byID[user.ID] = &cp
	return nil
}

func (m *mockUserRepo) UpdateFields(id uint, fields map[string]interface{}) error {
	if m.updateFieldsErr != nil {
		return m.updateFieldsErr
	}
	u, ok := m.byID[id]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	// 记录调用
	rec := make(map[string]interface{})
	for k, v := range fields {
		rec[k] = v
	}
	m.updatedFields[id] = rec
	// 应用字段
	if v, ok := fields["nickname"]; ok {
		u.Nickname = v.(string)
	}
	if v, ok := fields["avatar"]; ok {
		u.Avatar = v.(string)
	}
	if v, ok := fields["phone"]; ok {
		u.Phone = v.(string)
	}
	if v, ok := fields["email"]; ok {
		u.Email = v.(string)
	}
	if v, ok := fields["gender"]; ok {
		u.Gender = v.(int)
	}
	if v, ok := fields["status"]; ok {
		u.Status = v.(int)
	}
	if v, ok := fields["password"]; ok {
		u.Password = v.(string)
	}
	return nil
}

func (m *mockUserRepo) List(regionID uint, p *utils.Pagination, keyword string, status int) ([]model.User, int64, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	m.lastListRegionID = regionID
	m.lastListKeyword = keyword
	m.lastListStatus = status
	m.lastListPage = p
	var out []model.User
	for _, u := range m.byID {
		if regionID > 0 && u.RegionID != regionID {
			continue
		}
		if status == 0 || status == 1 {
			if u.Status != status {
				continue
			}
		}
		out = append(out, *u)
	}
	return out, int64(len(out)), nil
}

func (m *mockUserRepo) Delete(id uint) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if _, ok := m.byID[id]; !ok {
		return gorm.ErrRecordNotFound
	}
	delete(m.byID, id)
	m.deletedIDs = append(m.deletedIDs, id)
	return nil
}

// ===== 纯函数 =====

func TestToUserInfo(t *testing.T) {
	u := &model.User{
		Username: "alice",
		Nickname: "Alice",
		Avatar:   "https://x/a.png",
		Phone:    "13800138000",
		Email:    "a@b.com",
		Gender:   2,
		Status:   1,
	}
	u.ID = 9
	u.RegionID = 3

	info := toUserInfo(u)
	assert.Equal(t, uint(9), info.ID)
	assert.Equal(t, "alice", info.Username)
	assert.Equal(t, "Alice", info.Nickname)
	assert.Equal(t, "https://x/a.png", info.Avatar)
	assert.Equal(t, "13800138000", info.Phone)
	assert.Equal(t, "a@b.com", info.Email)
	assert.Equal(t, 2, info.Gender)
	assert.Equal(t, 1, info.Status)
}

func TestToUserInfo_NilSafe(t *testing.T) {
	// 即使嵌入字段未设置也不应 panic
	info := toUserInfo(&model.User{Username: "空用户"})
	assert.Equal(t, "空用户", info.Username)
	assert.Equal(t, uint(0), info.ID)
	assert.Equal(t, 0, info.Status)
}

// 注：TestGenOAuthUsername（含超长 openid 截断）已在 oauth_login_test.go 覆盖，此处不重复。

func TestRandomPassword_Length(t *testing.T) {
	pw, err := randomPassword(32)
	require.NoError(t, err)
	assert.Len(t, pw, 32, "长度应为 32")
	// 两次生成不应相同（crypto/rand）
	pw2, _ := randomPassword(32)
	assert.NotEqual(t, pw, pw2, "两次随机密码不应相同")
}

func TestRandomPassword_NonPositiveDefaultsTo32(t *testing.T) {
	pw, err := randomPassword(0)
	require.NoError(t, err)
	assert.Len(t, pw, 32, "n<=0 应默认 32")
	pw2, err := randomPassword(-5)
	require.NoError(t, err)
	assert.Len(t, pw2, 32)
}

// ===== 构造函数 =====

func TestNewUserService(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo, nil, nil, nil)
	assert.NotNil(t, svc)
	// 应返回可用的 UserService 接口实例
	_, ok := svc.(*userService)
	assert.True(t, ok, "返回值应为 *userService 类型")
}

// ===== Register =====

func TestUserService_Register_Success(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo, nil, nil, nil)

	info, err := svc.Register(1, &dto.RegisterRequest{
		Username: "alice",
		Password: "secret123",
		Nickname: "Alice",
		Phone:    "13800138000",
	})
	require.NoError(t, err)
	assert.Equal(t, "alice", info.Username)
	assert.Equal(t, "Alice", info.Nickname)
	assert.Equal(t, "13800138000", info.Phone)
	assert.Equal(t, 1, info.Status, "新注册用户状态应为 1")
	// 密码不应出现在 DTO
	assert.Equal(t, "", "", "DTO 无 Password 字段")
	// 仓库应写入一条
	stored, err := repo.FindByUsername("alice")
	require.NoError(t, err)
	assert.NotEqual(t, "secret123", stored.Password, "存储的应为哈希而非明文")
}

func TestUserService_Register_DefaultNickname(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo, nil, nil, nil)

	info, err := svc.Register(1, &dto.RegisterRequest{
		Username: "bob",
		Password: "secret123",
		// Nickname 留空
	})
	require.NoError(t, err)
	assert.Equal(t, "bob", info.Nickname, "Nickname 为空时应回退为 Username")
}

func TestUserService_Register_AlreadyExists(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo, nil, nil, nil)

	_, err := svc.Register(1, &dto.RegisterRequest{Username: "alice", Password: "secret123"})
	require.NoError(t, err)

	_, err = svc.Register(1, &dto.RegisterRequest{Username: "alice", Password: "secret123"})
	assert.ErrorIs(t, err, ErrUserAlreadyExists)
}

func TestUserService_Register_FindByUsernameOtherError(t *testing.T) {
	repo := newMockUserRepo()
	repo.findByUsernameErr = errors.New("db connection lost")
	svc := NewUserService(repo, nil, nil, nil)

	_, err := svc.Register(1, &dto.RegisterRequest{Username: "alice", Password: "secret123"})
	assert.Equal(t, "db connection lost", err.Error(), "非 NotFound 错误应原样透传")
}

func TestUserService_Register_CreateError(t *testing.T) {
	repo := newMockUserRepo()
	repo.createErr = errors.New("write failure")
	svc := NewUserService(repo, nil, nil, nil)

	_, err := svc.Register(1, &dto.RegisterRequest{Username: "alice", Password: "secret123"})
	assert.Equal(t, "write failure", err.Error())
}

// ===== Login =====

func TestUserService_Login_Success(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo, nil, nil, nil)

	// 先注册一个用户
	_, err := svc.Register(1, &dto.RegisterRequest{Username: "alice", Password: "secret123"})
	require.NoError(t, err)

	resp, err := svc.Login(&dto.LoginRequest{Username: "alice", Password: "secret123"})
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Token, "应签发非空 token")
	assert.Equal(t, 24*3600, resp.Expires)
	assert.Equal(t, "alice", resp.UserInfo.Username)
}

func TestUserService_Login_UserNotFound(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo, nil, nil, nil)

	_, err := svc.Login(&dto.LoginRequest{Username: "ghost", Password: "secret123"})
	assert.ErrorIs(t, err, ErrUserNotFound)
}

func TestUserService_Login_FindByUsernameOtherError(t *testing.T) {
	repo := newMockUserRepo()
	repo.findByUsernameErr = errors.New("db down")
	svc := NewUserService(repo, nil, nil, nil)

	_, err := svc.Login(&dto.LoginRequest{Username: "alice", Password: "secret123"})
	assert.Equal(t, "db down", err.Error())
}

func TestUserService_Login_DisabledUser(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo, nil, nil, nil)

	// 注册后禁用
	info, err := svc.Register(1, &dto.RegisterRequest{Username: "alice", Password: "secret123"})
	require.NoError(t, err)
	require.NoError(t, svc.UpdateUserStatus(info.ID, 0))

	_, err = svc.Login(&dto.LoginRequest{Username: "alice", Password: "secret123"})
	assert.ErrorIs(t, err, ErrUserDisabled)
}

func TestUserService_Login_WrongPassword(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo, nil, nil, nil)

	_, err := svc.Register(1, &dto.RegisterRequest{Username: "alice", Password: "secret123"})
	require.NoError(t, err)

	_, err = svc.Login(&dto.LoginRequest{Username: "alice", Password: "wrong-password"})
	assert.ErrorIs(t, err, ErrPasswordInvalid)
}

// ===== GetUserInfo =====

func TestUserService_GetUserInfo_Success(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo, nil, nil, nil)

	reg, err := svc.Register(1, &dto.RegisterRequest{Username: "alice", Password: "secret123"})
	require.NoError(t, err)

	info, err := svc.GetUserInfo(reg.ID)
	require.NoError(t, err)
	assert.Equal(t, "alice", info.Username)
}

func TestUserService_GetUserInfo_NotFound(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo, nil, nil, nil)

	_, err := svc.GetUserInfo(9999)
	assert.ErrorIs(t, err, ErrUserNotFound)
}

func TestUserService_GetUserInfo_OtherError(t *testing.T) {
	repo := newMockUserRepo()
	repo.findByIDErr = errors.New("boom")
	svc := NewUserService(repo, nil, nil, nil)

	_, err := svc.GetUserInfo(1)
	assert.Equal(t, "boom", err.Error())
}

// ===== UpdateProfile =====

func TestUserService_UpdateProfile_AllEmpty_NoOp(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo, nil, nil, nil)

	reg, err := svc.Register(1, &dto.RegisterRequest{Username: "alice", Password: "secret123"})
	require.NoError(t, err)

	// 全部零值 → 不调用 UpdateFields
	err = svc.UpdateProfile(reg.ID, &dto.UpdateProfileRequest{})
	require.NoError(t, err)
	assert.NotContains(t, repo.updatedFields, reg.ID, "无字段时不应调用 UpdateFields")
}

func TestUserService_UpdateProfile_PartialFields(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo, nil, nil, nil)

	reg, err := svc.Register(1, &dto.RegisterRequest{Username: "alice", Password: "secret123"})
	require.NoError(t, err)

	err = svc.UpdateProfile(reg.ID, &dto.UpdateProfileRequest{
		Nickname: "NewNick",
		Email:    "new@x.com",
	})
	require.NoError(t, err)
	fields := repo.updatedFields[reg.ID]
	assert.Equal(t, "NewNick", fields["nickname"])
	assert.Equal(t, "new@x.com", fields["email"])
	// Gender=0 + 有其他字段 → gender 也应被写入（条件 req.Gender != 0 || len(fields) > 0）
	assert.Equal(t, 0, fields["gender"])
}

func TestUserService_UpdateProfile_GenderOnly(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo, nil, nil, nil)

	reg, err := svc.Register(1, &dto.RegisterRequest{Username: "alice", Password: "secret123"})
	require.NoError(t, err)

	// 仅 Gender=2 → 应触发 UpdateFields（Gender!=0 分支）
	err = svc.UpdateProfile(reg.ID, &dto.UpdateProfileRequest{Gender: 2})
	require.NoError(t, err)
	fields := repo.updatedFields[reg.ID]
	assert.Equal(t, 2, fields["gender"])
}

func TestUserService_UpdateProfile_UpdateFieldsError(t *testing.T) {
	repo := newMockUserRepo()
	repo.updateFieldsErr = errors.New("update denied")
	svc := NewUserService(repo, nil, nil, nil)

	reg, err := svc.Register(1, &dto.RegisterRequest{Username: "alice", Password: "secret123"})
	require.NoError(t, err)

	err = svc.UpdateProfile(reg.ID, &dto.UpdateProfileRequest{Nickname: "x"})
	assert.Equal(t, "update denied", err.Error())
}

// ===== ChangePassword =====

func TestUserService_ChangePassword_Success(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo, nil, nil, nil)

	reg, err := svc.Register(1, &dto.RegisterRequest{Username: "alice", Password: "old123456"})
	require.NoError(t, err)

	err = svc.ChangePassword(reg.ID, &dto.ChangePasswordRequest{
		OldPassword: "old123456",
		NewPassword: "new123456",
	})
	require.NoError(t, err)
	// 仓库应写入新密码哈希
	fields := repo.updatedFields[reg.ID]
	assert.NotEmpty(t, fields["password"], "应写入 password 字段")
	assert.NotEqual(t, "new123456", fields["password"], "存储的应为哈希")
	// 新密码应能登录
	_, err = svc.Login(&dto.LoginRequest{Username: "alice", Password: "new123456"})
	require.NoError(t, err)
}

func TestUserService_ChangePassword_UserNotFound(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo, nil, nil, nil)

	err := svc.ChangePassword(9999, &dto.ChangePasswordRequest{
		OldPassword: "old",
		NewPassword: "new123456",
	})
	assert.ErrorIs(t, err, ErrUserNotFound)
}

func TestUserService_ChangePassword_OldPasswordWrong(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo, nil, nil, nil)

	reg, err := svc.Register(1, &dto.RegisterRequest{Username: "alice", Password: "old123456"})
	require.NoError(t, err)

	err = svc.ChangePassword(reg.ID, &dto.ChangePasswordRequest{
		OldPassword: "wrong-old",
		NewPassword: "new123456",
	})
	assert.ErrorIs(t, err, ErrOldPasswordWrong)
	// 不应调用 UpdateFields
	assert.NotContains(t, repo.updatedFields, reg.ID)
}

func TestUserService_ChangePassword_FindError(t *testing.T) {
	repo := newMockUserRepo()
	repo.findByIDErr = errors.New("db err")
	svc := NewUserService(repo, nil, nil, nil)

	err := svc.ChangePassword(1, &dto.ChangePasswordRequest{
		OldPassword: "old",
		NewPassword: "new123456",
	})
	assert.Equal(t, "db err", err.Error())
}

// ===== ListUsers =====

func TestUserService_ListUsers_Empty(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo, nil, nil, nil)

	pagination, list, err := svc.ListUsers(0, &dto.ListUsersRequest{Page: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(0), pagination.Total)
	assert.Empty(t, list)
}

func TestUserService_ListUsers_Multiple(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo, nil, nil, nil)

	_, err := svc.Register(1, &dto.RegisterRequest{Username: "alice", Password: "secret123"})
	require.NoError(t, err)
	_, err = svc.Register(1, &dto.RegisterRequest{Username: "bob", Password: "secret123"})
	require.NoError(t, err)

	// Status=-1 表示全部（见 dto ListUsersRequest 注释），仓库对 status∈{0,1} 才过滤
	pagination, list, err := svc.ListUsers(0, &dto.ListUsersRequest{Page: 1, PageSize: 10, Status: -1})
	require.NoError(t, err)
	assert.Equal(t, int64(2), pagination.Total)
	assert.Len(t, list, 2)
}

func TestUserService_ListUsers_StatusFilter(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo, nil, nil, nil)

	// 注册两个正常用户 + 禁用其中一个
	a, err := svc.Register(1, &dto.RegisterRequest{Username: "alice", Password: "secret123"})
	require.NoError(t, err)
	_, err = svc.Register(1, &dto.RegisterRequest{Username: "bob", Password: "secret123"})
	require.NoError(t, err)
	require.NoError(t, svc.UpdateUserStatus(a.ID, 0))

	// Status=1 只看正常用户 → 1 条
	_, listNormal, err := svc.ListUsers(0, &dto.ListUsersRequest{Page: 1, PageSize: 10, Status: 1})
	require.NoError(t, err)
	assert.Len(t, listNormal, 1)
	assert.Equal(t, "bob", listNormal[0].Username)
	// Status=0 只看禁用用户 → 1 条
	_, listDisabled, err := svc.ListUsers(0, &dto.ListUsersRequest{Page: 1, PageSize: 10, Status: 0})
	require.NoError(t, err)
	assert.Len(t, listDisabled, 1)
	assert.Equal(t, "alice", listDisabled[0].Username)
	// Status=-1 全部 → 2 条
	_, listAll, err := svc.ListUsers(0, &dto.ListUsersRequest{Page: 1, PageSize: 10, Status: -1})
	require.NoError(t, err)
	assert.Len(t, listAll, 2)
}

func TestUserService_ListUsers_RegionIsolation(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo, nil, nil, nil)

	// 在 region 1 和 region 2 各注册一个
	_, err := svc.Register(1, &dto.RegisterRequest{Username: "r1user", Password: "secret123"})
	require.NoError(t, err)
	_, err = svc.Register(2, &dto.RegisterRequest{Username: "r2user", Password: "secret123"})
	require.NoError(t, err)

	// regionID=1 只能看到 r1user（Status=-1 不过滤状态）
	_, list, err := svc.ListUsers(1, &dto.ListUsersRequest{Page: 1, PageSize: 10, Status: -1})
	require.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, "r1user", list[0].Username)
	assert.Equal(t, uint(1), repo.lastListRegionID, "透传到 repo 的 regionID 应为 1")
	// regionID=0 超管跨区，看到全部
	_, listAll, err := svc.ListUsers(0, &dto.ListUsersRequest{Page: 1, PageSize: 10, Status: -1})
	require.NoError(t, err)
	assert.Len(t, listAll, 2)
	assert.Equal(t, uint(0), repo.lastListRegionID, "最后一次调用 regionID 应为 0")
}

func TestUserService_ListUsers_ParamsPassthrough(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo, nil, nil, nil)

	_, _, err := svc.ListUsers(5, &dto.ListUsersRequest{Page: 2, PageSize: 20, Keyword: "ali", Status: 1})
	require.NoError(t, err)
	assert.Equal(t, uint(5), repo.lastListRegionID)
	assert.Equal(t, "ali", repo.lastListKeyword)
	assert.Equal(t, 1, repo.lastListStatus)
	require.NotNil(t, repo.lastListPage)
	assert.Equal(t, 2, repo.lastListPage.Page)
	assert.Equal(t, 20, repo.lastListPage.PageSize)
}

func TestUserService_ListUsers_DefaultPagination(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo, nil, nil, nil)

	// Page=0 / PageSize=0 → NewPagination 兜底为 1/10
	_, _, err := svc.ListUsers(0, &dto.ListUsersRequest{})
	require.NoError(t, err)
	require.NotNil(t, repo.lastListPage)
	assert.Equal(t, 1, repo.lastListPage.Page)
	assert.Equal(t, 10, repo.lastListPage.PageSize)
}

func TestUserService_ListUsers_RepoError(t *testing.T) {
	repo := newMockUserRepo()
	repo.listErr = errors.New("list failed")
	svc := NewUserService(repo, nil, nil, nil)

	_, _, err := svc.ListUsers(0, &dto.ListUsersRequest{Page: 1, PageSize: 10})
	assert.Equal(t, "list failed", err.Error())
}

// ===== AdminCreateUser =====

func TestUserService_AdminCreateUser_Success(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo, nil, nil, nil)

	info, err := svc.AdminCreateUser(1, &dto.AdminCreateUserRequest{
		Username: "admin-created",
		Password: "secret123",
		Nickname: "Admin",
		Phone:    "13900000000",
		Email:    "a@b.com",
		Gender:   1,
		Status:   1,
	})
	require.NoError(t, err)
	assert.Equal(t, "admin-created", info.Username)
	assert.Equal(t, "Admin", info.Nickname)
	assert.Equal(t, "13900000000", info.Phone)
	assert.Equal(t, "a@b.com", info.Email)
	assert.Equal(t, 1, info.Gender)
	assert.Equal(t, 1, info.Status)
	// 仓库内 region_id 应为入参
	stored, err := repo.FindByUsername("admin-created")
	require.NoError(t, err)
	assert.Equal(t, uint(1), stored.RegionID)
}

func TestUserService_AdminCreateUser_DefaultStatus(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo, nil, nil, nil)

	// Status=0 → 默认填充为 1
	info, err := svc.AdminCreateUser(1, &dto.AdminCreateUserRequest{
		Username: "u1",
		Password: "secret123",
		Status:   0,
	})
	require.NoError(t, err)
	assert.Equal(t, 1, info.Status, "Status=0 应默认填充为 1")
}

func TestUserService_AdminCreateUser_DefaultNickname(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo, nil, nil, nil)

	info, err := svc.AdminCreateUser(1, &dto.AdminCreateUserRequest{
		Username: "noname",
		Password: "secret123",
	})
	require.NoError(t, err)
	assert.Equal(t, "noname", info.Nickname, "Nickname 为空时应回退为 Username")
}

func TestUserService_AdminCreateUser_AlreadyExists(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo, nil, nil, nil)

	_, err := svc.AdminCreateUser(1, &dto.AdminCreateUserRequest{
		Username: "dupe",
		Password: "secret123",
	})
	require.NoError(t, err)

	_, err = svc.AdminCreateUser(1, &dto.AdminCreateUserRequest{
		Username: "dupe",
		Password: "secret123",
	})
	assert.ErrorIs(t, err, ErrUserAlreadyExists)
}

func TestUserService_AdminCreateUser_FindByUsernameOtherError(t *testing.T) {
	repo := newMockUserRepo()
	repo.findByUsernameErr = errors.New("conn lost")
	svc := NewUserService(repo, nil, nil, nil)

	_, err := svc.AdminCreateUser(1, &dto.AdminCreateUserRequest{
		Username: "u1",
		Password: "secret123",
	})
	assert.Equal(t, "conn lost", err.Error())
}

func TestUserService_AdminCreateUser_CreateError(t *testing.T) {
	repo := newMockUserRepo()
	repo.createErr = errors.New("write fail")
	svc := NewUserService(repo, nil, nil, nil)

	_, err := svc.AdminCreateUser(1, &dto.AdminCreateUserRequest{
		Username: "u1",
		Password: "secret123",
	})
	assert.Equal(t, "write fail", err.Error())
}

// ===== AdminUpdateUser =====

func TestUserService_AdminUpdateUser_PartialFields(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo, nil, nil, nil)

	reg, err := svc.Register(1, &dto.RegisterRequest{Username: "alice", Password: "secret123"})
	require.NoError(t, err)

	err = svc.AdminUpdateUser(reg.ID, &dto.AdminUpdateUserRequest{
		Nickname: "AdminNick",
		Phone:    "13700000000",
	})
	require.NoError(t, err)
	fields := repo.updatedFields[reg.ID]
	assert.Equal(t, "AdminNick", fields["nickname"])
	assert.Equal(t, "13700000000", fields["phone"])
	// AdminUpdateUser 总是写入 gender（无 UpdateProfile 的条件分支）
	assert.Equal(t, 0, fields["gender"])
}

func TestUserService_AdminUpdateUser_AllFields(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo, nil, nil, nil)

	reg, err := svc.Register(1, &dto.RegisterRequest{Username: "alice", Password: "secret123"})
	require.NoError(t, err)

	err = svc.AdminUpdateUser(reg.ID, &dto.AdminUpdateUserRequest{
		Nickname: "N",
		Avatar:   "A",
		Phone:    "P",
		Email:    "E",
		Gender:   2,
	})
	require.NoError(t, err)
	fields := repo.updatedFields[reg.ID]
	assert.Equal(t, "N", fields["nickname"])
	assert.Equal(t, "A", fields["avatar"])
	assert.Equal(t, "P", fields["phone"])
	assert.Equal(t, "E", fields["email"])
	assert.Equal(t, 2, fields["gender"])
}

func TestUserService_AdminUpdateUser_UpdateFieldsError(t *testing.T) {
	repo := newMockUserRepo()
	repo.updateFieldsErr = errors.New("denied")
	svc := NewUserService(repo, nil, nil, nil)

	err := svc.AdminUpdateUser(1, &dto.AdminUpdateUserRequest{Nickname: "x"})
	assert.Equal(t, "denied", err.Error())
}

// ===== UpdateUserStatus =====

func TestUserService_UpdateUserStatus_Success(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo, nil, nil, nil)

	reg, err := svc.Register(1, &dto.RegisterRequest{Username: "alice", Password: "secret123"})
	require.NoError(t, err)

	err = svc.UpdateUserStatus(reg.ID, 0)
	require.NoError(t, err)
	fields := repo.updatedFields[reg.ID]
	assert.Equal(t, 0, fields["status"])
	// 禁用后登录应被拒
	_, err = svc.Login(&dto.LoginRequest{Username: "alice", Password: "secret123"})
	assert.ErrorIs(t, err, ErrUserDisabled)

	// 再次启用
	err = svc.UpdateUserStatus(reg.ID, 1)
	require.NoError(t, err)
	_, err = svc.Login(&dto.LoginRequest{Username: "alice", Password: "secret123"})
	require.NoError(t, err)
}

func TestUserService_UpdateUserStatus_UpdateFieldsError(t *testing.T) {
	repo := newMockUserRepo()
	repo.updateFieldsErr = errors.New("nope")
	svc := NewUserService(repo, nil, nil, nil)

	err := svc.UpdateUserStatus(1, 1)
	assert.Equal(t, "nope", err.Error())
}

// ===== ResetPassword =====

func TestUserService_ResetPassword_Success(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo, nil, nil, nil)

	reg, err := svc.Register(1, &dto.RegisterRequest{Username: "alice", Password: "old123456"})
	require.NoError(t, err)

	err = svc.ResetPassword(reg.ID, &dto.ResetPasswordRequest{NewPassword: "brand-new-pwd"})
	require.NoError(t, err)
	fields := repo.updatedFields[reg.ID]
	assert.NotEmpty(t, fields["password"])
	assert.NotEqual(t, "brand-new-pwd", fields["password"], "存储的应为哈希")

	// 新密码可登录，旧密码失效
	_, err = svc.Login(&dto.LoginRequest{Username: "alice", Password: "brand-new-pwd"})
	require.NoError(t, err)
	_, err = svc.Login(&dto.LoginRequest{Username: "alice", Password: "old123456"})
	assert.ErrorIs(t, err, ErrPasswordInvalid)
}

func TestUserService_ResetPassword_UpdateFieldsError(t *testing.T) {
	repo := newMockUserRepo()
	repo.updateFieldsErr = errors.New("deny")
	svc := NewUserService(repo, nil, nil, nil)

	err := svc.ResetPassword(1, &dto.ResetPasswordRequest{NewPassword: "new123456"})
	assert.Equal(t, "deny", err.Error())
}

// ===== DeleteUser =====

func TestUserService_DeleteUser_Success(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo, nil, nil, nil)

	reg, err := svc.Register(1, &dto.RegisterRequest{Username: "alice", Password: "secret123"})
	require.NoError(t, err)

	err = svc.DeleteUser(reg.ID)
	require.NoError(t, err)
	assert.Contains(t, repo.deletedIDs, reg.ID)
	// 删除后 GetUserInfo 应返回 NotFound
	_, err = svc.GetUserInfo(reg.ID)
	assert.ErrorIs(t, err, ErrUserNotFound)
}

func TestUserService_DeleteUser_NotFound(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo, nil, nil, nil)

	err := svc.DeleteUser(9999)
	assert.ErrorIs(t, err, ErrUserNotFound)
}

func TestUserService_DeleteUser_FindError(t *testing.T) {
	repo := newMockUserRepo()
	repo.findByIDErr = errors.New("db err")
	svc := NewUserService(repo, nil, nil, nil)

	err := svc.DeleteUser(1)
	assert.Equal(t, "db err", err.Error())
}

func TestUserService_DeleteUser_DeleteError(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo, nil, nil, nil)

	reg, err := svc.Register(1, &dto.RegisterRequest{Username: "alice", Password: "secret123"})
	require.NoError(t, err)

	repo.deleteErr = errors.New("delete blocked")
	err = svc.DeleteUser(reg.ID)
	assert.Equal(t, "delete blocked", err.Error())
}
