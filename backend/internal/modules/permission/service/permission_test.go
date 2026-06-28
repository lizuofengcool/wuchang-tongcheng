// Package service 权限服务单元测试。
// 使用内存 mock repository，覆盖错误码映射、默认值填充、HasPermission、
// AssignPermissionsToRole 角色存在校验等核心业务逻辑，不依赖 DB。
package service

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"wuchang-tongcheng/internal/modules/permission/dto"
	permModel "wuchang-tongcheng/internal/modules/permission/model"
)

// mockRepo 内存 mock，实现 PermissionRepository 接口
type mockRepo struct {
	roles       map[uint]*permModel.Role
	perms       map[uint]*permModel.Permission
	userRoles   map[uint][]uint // userID -> []roleID
	rolePerms   map[uint][]uint // roleID -> []permID
	nextRoleID  uint
	nextPermID  uint
	// 注入错误（用于测试错误透传）
	createRoleErr     error
	createPermErr     error
	findPermCodesErr  error
	findRoleCodesErr  error
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		roles:      make(map[uint]*permModel.Role),
		perms:      make(map[uint]*permModel.Permission),
		userRoles:  make(map[uint][]uint),
		rolePerms:  make(map[uint][]uint),
		nextRoleID: 1,
		nextPermID: 1,
	}
}

// ===== 实现接口方法 =====

func (m *mockRepo) CreateRole(role *permModel.Role) error {
	if m.createRoleErr != nil {
		return m.createRoleErr
	}
	role.ID = m.nextRoleID
	m.nextRoleID++
	cp := *role
	m.roles[role.ID] = &cp
	return nil
}

func (m *mockRepo) FindRoleByID(id uint) (*permModel.Role, error) {
	r, ok := m.roles[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *r
	return &cp, nil
}

func (m *mockRepo) FindRoleByCode(code string) (*permModel.Role, error) {
	for _, r := range m.roles {
		if r.Code == code {
			cp := *r
			return &cp, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockRepo) FindAllRoles() ([]permModel.Role, error) {
	out := make([]permModel.Role, 0, len(m.roles))
	for _, r := range m.roles {
		out = append(out, *r)
	}
	return out, nil
}

func (m *mockRepo) UpdateRole(role *permModel.Role) error {
	if _, ok := m.roles[role.ID]; !ok {
		return gorm.ErrRecordNotFound
	}
	cp := *role
	m.roles[role.ID] = &cp
	return nil
}

func (m *mockRepo) UpdateRoleFields(id uint, fields map[string]interface{}) error {
	r, ok := m.roles[id]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	if v, ok := fields["name"]; ok {
		r.Name = v.(string)
	}
	if v, ok := fields["description"]; ok {
		r.Description = v.(string)
	}
	if v, ok := fields["sort"]; ok {
		r.Sort = v.(int)
	}
	if v, ok := fields["status"]; ok {
		r.Status = v.(int)
	}
	return nil
}

func (m *mockRepo) DeleteRole(id uint) error {
	if _, ok := m.roles[id]; !ok {
		return gorm.ErrRecordNotFound
	}
	delete(m.roles, id)
	delete(m.rolePerms, id)
	for uid, rids := range m.userRoles {
		var kept []uint
		for _, rid := range rids {
			if rid != id {
				kept = append(kept, rid)
			}
		}
		m.userRoles[uid] = kept
	}
	return nil
}

func (m *mockRepo) CreatePermission(p *permModel.Permission) error {
	if m.createPermErr != nil {
		return m.createPermErr
	}
	p.ID = m.nextPermID
	m.nextPermID++
	cp := *p
	m.perms[p.ID] = &cp
	return nil
}

func (m *mockRepo) FindPermissionByID(id uint) (*permModel.Permission, error) {
	p, ok := m.perms[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *p
	return &cp, nil
}

func (m *mockRepo) FindPermissionByCode(code string) (*permModel.Permission, error) {
	for _, p := range m.perms {
		if p.Code == code {
			cp := *p
			return &cp, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockRepo) FindAllPermissions() ([]permModel.Permission, error) {
	out := make([]permModel.Permission, 0, len(m.perms))
	for _, p := range m.perms {
		out = append(out, *p)
	}
	return out, nil
}

func (m *mockRepo) UpdatePermissionFields(id uint, fields map[string]interface{}) error {
	p, ok := m.perms[id]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	if v, ok := fields["name"]; ok {
		p.Name = v.(string)
	}
	if v, ok := fields["path"]; ok {
		p.Path = v.(string)
	}
	if v, ok := fields["method"]; ok {
		p.Method = v.(string)
	}
	if v, ok := fields["sort"]; ok {
		p.Sort = v.(int)
	}
	if v, ok := fields["status"]; ok {
		p.Status = v.(int)
	}
	return nil
}

func (m *mockRepo) DeletePermission(id uint) error {
	if _, ok := m.perms[id]; !ok {
		return gorm.ErrRecordNotFound
	}
	delete(m.perms, id)
	for rid, pids := range m.rolePerms {
		var kept []uint
		for _, pid := range pids {
			if pid != id {
				kept = append(kept, pid)
			}
		}
		m.rolePerms[rid] = kept
	}
	return nil
}

func (m *mockRepo) AssignRolesToUser(userID uint, roleIDs []uint) error {
	m.userRoles[userID] = append([]uint(nil), roleIDs...)
	return nil
}

func (m *mockRepo) FindRolesByUserID(userID uint) ([]permModel.Role, error) {
	rids := m.userRoles[userID]
	out := make([]permModel.Role, 0, len(rids))
	for _, rid := range rids {
		if r, ok := m.roles[rid]; ok {
			out = append(out, *r)
		}
	}
	return out, nil
}

func (m *mockRepo) ClearUserRoles(userID uint) error {
	delete(m.userRoles, userID)
	return nil
}

func (m *mockRepo) AssignPermissionsToRole(roleID uint, permIDs []uint) error {
	m.rolePerms[roleID] = append([]uint(nil), permIDs...)
	return nil
}

func (m *mockRepo) FindPermissionsByRoleID(roleID uint) ([]permModel.Permission, error) {
	pids := m.rolePerms[roleID]
	out := make([]permModel.Permission, 0, len(pids))
	for _, pid := range pids {
		if p, ok := m.perms[pid]; ok {
			out = append(out, *p)
		}
	}
	return out, nil
}

func (m *mockRepo) ClearRolePermissions(roleID uint) error {
	delete(m.rolePerms, roleID)
	return nil
}

func (m *mockRepo) FindPermissionsByUserID(userID uint) ([]permModel.Permission, error) {
	codes, err := m.FindPermissionCodesByUserID(userID)
	if err != nil {
		return nil, err
	}
	out := make([]permModel.Permission, 0, len(codes))
	for _, c := range codes {
		for _, p := range m.perms {
			if p.Code == c && p.Status == 1 {
				out = append(out, *p)
				break
			}
		}
	}
	return out, nil
}

func (m *mockRepo) FindPermissionCodesByUserID(userID uint) ([]string, error) {
	if m.findPermCodesErr != nil {
		return nil, m.findPermCodesErr
	}
	rids := m.userRoles[userID]
	seen := make(map[string]bool)
	out := make([]string, 0)
	for _, rid := range rids {
		for _, pid := range m.rolePerms[rid] {
			if p, ok := m.perms[pid]; ok && p.Status == 1 {
				if !seen[p.Code] {
					seen[p.Code] = true
					out = append(out, p.Code)
				}
			}
		}
	}
	return out, nil
}

func (m *mockRepo) FindRoleCodesByUserID(userID uint) ([]string, error) {
	if m.findRoleCodesErr != nil {
		return nil, m.findRoleCodesErr
	}
	rids := m.userRoles[userID]
	out := make([]string, 0, len(rids))
	for _, rid := range rids {
		if r, ok := m.roles[rid]; ok && r.Status == 1 {
			out = append(out, r.Code)
		}
	}
	return out, nil
}

// ========== 测试用例 ==========

// TestCreateRole_DuplicateCode 重复 code 返回 ErrRoleCodeExists
func TestCreateRole_DuplicateCode(t *testing.T) {
	repo := newMockRepo()
	svc := NewPermissionService(repo)

	_, err := svc.CreateRole(&dto.CreateRoleRequest{
		Name: "管理员", Code: "admin", Sort: 1, Status: 1,
	})
	require.NoError(t, err)

	_, err = svc.CreateRole(&dto.CreateRoleRequest{
		Name: "重复管理员", Code: "admin", Sort: 2, Status: 1,
	})
	assert.ErrorIs(t, err, ErrRoleCodeExists)
}

// TestCreateRole_DefaultStatus status=0 自动填充为 1
func TestCreateRole_DefaultStatus(t *testing.T) {
	repo := newMockRepo()
	svc := NewPermissionService(repo)

	role, err := svc.CreateRole(&dto.CreateRoleRequest{
		Name: "测试角色", Code: "test", Sort: 1, Status: 0, // 0 应被填充为 1
	})
	require.NoError(t, err)
	assert.Equal(t, 1, role.Status, "status=0 应自动填充为 1")
}

// TestCreateRole_RepoError repo 错误透传
func TestCreateRole_RepoError(t *testing.T) {
	repo := newMockRepo()
	repo.createRoleErr = errors.New("db connection lost")
	svc := NewPermissionService(repo)

	_, err := svc.CreateRole(&dto.CreateRoleRequest{
		Name: "测试", Code: "test", Status: 1,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "db connection lost")
}

// TestCreatePermission_DuplicateCode 重复 code 返回 ErrPermCodeExists
func TestCreatePermission_DuplicateCode(t *testing.T) {
	repo := newMockRepo()
	svc := NewPermissionService(repo)

	_, err := svc.CreatePermission(&dto.CreatePermissionRequest{
		Name: "用户读取", Code: "user:read", Type: 3, Sort: 1, Status: 1,
	})
	require.NoError(t, err)

	_, err = svc.CreatePermission(&dto.CreatePermissionRequest{
		Name: "用户读取重复", Code: "user:read", Type: 3, Sort: 2, Status: 1,
	})
	assert.ErrorIs(t, err, ErrPermCodeExists)
}

// TestCreatePermission_DefaultStatus status=0 自动填充为 1
func TestCreatePermission_DefaultStatus(t *testing.T) {
	repo := newMockRepo()
	svc := NewPermissionService(repo)

	p, err := svc.CreatePermission(&dto.CreatePermissionRequest{
		Name: "测试权限", Code: "test:r", Type: 3, Sort: 1, Status: 0,
	})
	require.NoError(t, err)
	assert.Equal(t, 1, p.Status)
}

// TestUpdateRole_NotFound 角色不存在返回 ErrRoleNotFound
func TestUpdateRole_NotFound(t *testing.T) {
	repo := newMockRepo()
	svc := NewPermissionService(repo)

	err := svc.UpdateRole(9999, &dto.UpdateRoleRequest{Name: "更新"})
	assert.ErrorIs(t, err, ErrRoleNotFound)
}

// TestDeleteRole_NotFound 角色不存在返回 ErrRoleNotFound
func TestDeleteRole_NotFound(t *testing.T) {
	repo := newMockRepo()
	svc := NewPermissionService(repo)

	err := svc.DeleteRole(9999)
	assert.ErrorIs(t, err, ErrRoleNotFound)
}

// TestUpdatePermission_NotFound 权限不存在返回 ErrPermissionNotFound
func TestUpdatePermission_NotFound(t *testing.T) {
	repo := newMockRepo()
	svc := NewPermissionService(repo)

	err := svc.UpdatePermission(9999, map[string]interface{}{"name": "更新"})
	assert.ErrorIs(t, err, ErrPermissionNotFound)
}

// TestDeletePermission_NotFound 权限不存在返回 ErrPermissionNotFound
func TestDeletePermission_NotFound(t *testing.T) {
	repo := newMockRepo()
	svc := NewPermissionService(repo)

	err := svc.DeletePermission(9999)
	assert.ErrorIs(t, err, ErrPermissionNotFound)
}

// TestAssignPermissionsToRole_RoleNotFound 角色不存在返回 ErrRoleNotFound
func TestAssignPermissionsToRole_RoleNotFound(t *testing.T) {
	repo := newMockRepo()
	svc := NewPermissionService(repo)

	err := svc.AssignPermissionsToRole(&dto.AssignPermissionsRequest{
		RoleID: 9999, PermissionIDs: []uint{1},
	})
	assert.ErrorIs(t, err, ErrRoleNotFound)
}

// TestAssignPermissionsToRole_Success 角色存在时透传 repo
func TestAssignPermissionsToRole_Success(t *testing.T) {
	repo := newMockRepo()
	svc := NewPermissionService(repo)

	// 先创建角色
	role, err := svc.CreateRole(&dto.CreateRoleRequest{
		Name: "测试角色", Code: "test", Status: 1,
	})
	require.NoError(t, err)

	err = svc.AssignPermissionsToRole(&dto.AssignPermissionsRequest{
		RoleID: role.ID, PermissionIDs: []uint{1, 2, 3},
	})
	require.NoError(t, err)

	perms, err := svc.GetPermissionsByRoleID(role.ID)
	require.NoError(t, err)
	assert.Empty(t, perms, "权限 ID 1/2/3 不存在，返回空但不应报错")
}

// TestHasPermission_Hit 用户拥有该权限返回 true
func TestHasPermission_Hit(t *testing.T) {
	repo := newMockRepo()
	svc := NewPermissionService(repo)

	// 构造：角色 → 权限，用户 → 角色
	role, _ := svc.CreateRole(&dto.CreateRoleRequest{Name: "编辑", Code: "editor", Status: 1})
	perm, _ := svc.CreatePermission(&dto.CreatePermissionRequest{
		Name: "用户读取", Code: "user:read", Type: 3, Status: 1,
	})
	require.NoError(t, repo.AssignPermissionsToRole(role.ID, []uint{perm.ID}))
	require.NoError(t, repo.AssignRolesToUser(100, []uint{role.ID}))

	ok, err := svc.HasPermission(100, "user:read")
	require.NoError(t, err)
	assert.True(t, ok)
}

// TestHasPermission_Miss 用户没有该权限返回 false
func TestHasPermission_Miss(t *testing.T) {
	repo := newMockRepo()
	svc := NewPermissionService(repo)

	role, _ := svc.CreateRole(&dto.CreateRoleRequest{Name: "编辑", Code: "editor", Status: 1})
	require.NoError(t, repo.AssignRolesToUser(100, []uint{role.ID}))

	ok, err := svc.HasPermission(100, "user:write")
	require.NoError(t, err)
	assert.False(t, ok)
}

// TestHasPermission_RepoError repo 错误透传
func TestHasPermission_RepoError(t *testing.T) {
	repo := newMockRepo()
	repo.findPermCodesErr = errors.New("db down")
	svc := NewPermissionService(repo)

	ok, err := svc.HasPermission(100, "user:read")
	require.Error(t, err)
	assert.False(t, ok)
	assert.Contains(t, err.Error(), "db down")
}

// TestGetMyAuth 同时返回权限码和角色码
func TestGetMyAuth(t *testing.T) {
	repo := newMockRepo()
	svc := NewPermissionService(repo)

	role1, _ := svc.CreateRole(&dto.CreateRoleRequest{Name: "编辑1", Code: "editor", Status: 1})
	role2, _ := svc.CreateRole(&dto.CreateRoleRequest{Name: "编辑2", Code: "auditor", Status: 1})
	p1, _ := svc.CreatePermission(&dto.CreatePermissionRequest{Name: "读", Code: "user:read", Type: 3, Status: 1})
	p2, _ := svc.CreatePermission(&dto.CreatePermissionRequest{Name: "写", Code: "user:write", Type: 3, Status: 1})

	require.NoError(t, repo.AssignPermissionsToRole(role1.ID, []uint{p1.ID}))
	require.NoError(t, repo.AssignPermissionsToRole(role2.ID, []uint{p2.ID}))
	require.NoError(t, repo.AssignRolesToUser(200, []uint{role1.ID, role2.ID}))

	perms, roles, err := svc.GetMyAuth(200)
	require.NoError(t, err)
	require.Len(t, perms, 2)
	require.Len(t, roles, 2)
	assert.Contains(t, perms, "user:read")
	assert.Contains(t, perms, "user:write")
	assert.Contains(t, roles, "editor")
	assert.Contains(t, roles, "auditor")
}

// TestListRoles 空列表返回非 nil slice
func TestListRoles_Empty(t *testing.T) {
	repo := newMockRepo()
	svc := NewPermissionService(repo)

	roles, err := svc.ListRoles()
	require.NoError(t, err)
	assert.NotNil(t, roles, "空列表应返回非 nil slice")
	assert.Len(t, roles, 0)
}

// TestListPermissions 空列表返回非 nil slice
func TestListPermissions_Empty(t *testing.T) {
	repo := newMockRepo()
	svc := NewPermissionService(repo)

	perms, err := svc.ListPermissions()
	require.NoError(t, err)
	assert.NotNil(t, perms)
	assert.Len(t, perms, 0)
}

// TestGetRoleByID_NotFound 返回 ErrRoleNotFound
func TestGetRoleByID_NotFound(t *testing.T) {
	repo := newMockRepo()
	svc := NewPermissionService(repo)

	_, err := svc.GetRoleByID(9999)
	assert.ErrorIs(t, err, ErrRoleNotFound)
}

// TestGetRoleByID_Success 正常返回 RoleInfo
func TestGetRoleByID_Success(t *testing.T) {
	repo := newMockRepo()
	svc := NewPermissionService(repo)

	created, err := svc.CreateRole(&dto.CreateRoleRequest{
		Name: "管理员", Code: "admin", Description: "管理员角色", Sort: 5, Status: 1,
	})
	require.NoError(t, err)

	got, err := svc.GetRoleByID(created.ID)
	require.NoError(t, err)
	assert.Equal(t, "admin", got.Code)
	assert.Equal(t, "管理员", got.Name)
	assert.Equal(t, "管理员角色", got.Description)
	assert.Equal(t, 5, got.Sort)
	assert.Equal(t, 1, got.Status)
}

// TestUpdateRole_Success 局部更新字段
func TestUpdateRole_Success(t *testing.T) {
	repo := newMockRepo()
	svc := NewPermissionService(repo)

	created, err := svc.CreateRole(&dto.CreateRoleRequest{
		Name: "原名", Code: "r1", Description: "原描述", Sort: 1, Status: 1,
	})
	require.NoError(t, err)

	err = svc.UpdateRole(created.ID, &dto.UpdateRoleRequest{
		Name:        "新名",
		Description: "新描述",
		Sort:        10,
		Status:      0,
	})
	require.NoError(t, err)

	got, err := svc.GetRoleByID(created.ID)
	require.NoError(t, err)
	assert.Equal(t, "新名", got.Name)
	assert.Equal(t, "新描述", got.Description)
	assert.Equal(t, 10, got.Sort)
	assert.Equal(t, 0, got.Status)
}

// TestUpdateRole_KeepStatusWhenOutOfRange status 非 0/1 时不更新
func TestUpdateRole_KeepStatusWhenOutOfRange(t *testing.T) {
	repo := newMockRepo()
	svc := NewPermissionService(repo)

	created, err := svc.CreateRole(&dto.CreateRoleRequest{
		Name: "测试", Code: "test", Status: 1,
	})
	require.NoError(t, err)

	// status=2 越界，不应更新
	err = svc.UpdateRole(created.ID, &dto.UpdateRoleRequest{
		Status: 2,
	})
	require.NoError(t, err)

	got, err := svc.GetRoleByID(created.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, got.Status, "status=2 越界应保持原值 1")
}

// TestDeleteRole_Success 删除角色成功
func TestDeleteRole_Success(t *testing.T) {
	repo := newMockRepo()
	svc := NewPermissionService(repo)

	created, err := svc.CreateRole(&dto.CreateRoleRequest{
		Name: "待删除", Code: "del", Status: 1,
	})
	require.NoError(t, err)

	require.NoError(t, svc.DeleteRole(created.ID))

	_, err = svc.GetRoleByID(created.ID)
	assert.ErrorIs(t, err, ErrRoleNotFound)
}
