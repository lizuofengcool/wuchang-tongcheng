// Package repository 权限仓储的集成测试。
// 使用 testcontainers 启动真实 PostgreSQL，验证 RBAC 核心行为：
//   - 角色/权限的 CRUD、唯一索引、排序
//   - 用户-角色、角色-权限的覆盖式分配
//   - 综合查询（去重、status=1 过滤、禁用权限不返回）
//   - 删除角色/权限时的关联级联清理
// 无 Docker 时自动 skip。
package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	permModel "wuchang-tongcheng/internal/modules/permission/model"
	"wuchang-tongcheng/internal/testutil/pgtest"
)

// newRepoForTest 启动 PG 容器 + 全量建表 + 构造 permissionRepository。
func newRepoForTest(t *testing.T) (*permissionRepository, *gorm.DB) {
	t.Helper()
	db := pgtest.SetupPostgres(t)
	pgtest.MigrateAll(t, db)
	return &permissionRepository{db: db}, db
}

func makeRole(code, name string, sort int) *permModel.Role {
	return &permModel.Role{
		Code:        code,
		Name:        name,
		Description: name + " desc",
		Sort:        sort,
		Status:      1,
	}
}

func makePerm(code, name string, ptype int, sort int) *permModel.Permission {
	return &permModel.Permission{
		Code:   code,
		Name:   name,
		Type:   ptype,
		Sort:   sort,
		Status: 1,
	}
}

// ========== 角色 CRUD ==========

// TestPermissionRepository_RoleCreateAndFind 创建 + 按 ID/Code 查询
func TestPermissionRepository_RoleCreateAndFind(t *testing.T) {
	repo, _ := newRepoForTest(t)

	role := makeRole("admin", "管理员", 1)
	require.NoError(t, repo.CreateRole(role))
	require.NotZero(t, role.ID)

	got, err := repo.FindRoleByID(role.ID)
	require.NoError(t, err)
	assert.Equal(t, "admin", got.Code)
	assert.Equal(t, "管理员", got.Name)
	assert.Equal(t, 1, got.Sort)
	assert.Equal(t, 1, got.Status)

	byCode, err := repo.FindRoleByCode("admin")
	require.NoError(t, err)
	assert.Equal(t, got.ID, byCode.ID)
}

// TestPermissionRepository_FindRoleByCode_NotFound 查不到返回 ErrRecordNotFound
func TestPermissionRepository_FindRoleByCode_NotFound(t *testing.T) {
	repo, _ := newRepoForTest(t)

	_, err := repo.FindRoleByCode("non-existent")
	require.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

// TestPermissionRepository_CreateRole_DuplicateCode 唯一索引：重复 code 失败
func TestPermissionRepository_CreateRole_DuplicateCode(t *testing.T) {
	repo, _ := newRepoForTest(t)

	role1 := makeRole("editor", "编辑", 2)
	require.NoError(t, repo.CreateRole(role1))

	role2 := makeRole("editor", "编辑重复", 3)
	err := repo.CreateRole(role2)
	require.Error(t, err, "重复 code 应因唯一索引失败")
}

// TestPermissionRepository_FindAllRoles 按 sort ASC, id ASC 排序
func TestPermissionRepository_FindAllRoles(t *testing.T) {
	repo, _ := newRepoForTest(t)

	require.NoError(t, repo.CreateRole(makeRole("c", "C", 3)))
	require.NoError(t, repo.CreateRole(makeRole("a", "A", 1)))
	require.NoError(t, repo.CreateRole(makeRole("b", "B", 2)))

	roles, err := repo.FindAllRoles()
	require.NoError(t, err)
	require.Len(t, roles, 3)
	assert.Equal(t, "a", roles[0].Code, "应按 sort ASC 排序")
	assert.Equal(t, "b", roles[1].Code)
	assert.Equal(t, "c", roles[2].Code)
}

// TestPermissionRepository_UpdateRoleFields 局部字段更新
func TestPermissionRepository_UpdateRoleFields(t *testing.T) {
	repo, _ := newRepoForTest(t)

	role := makeRole("auditor", "审计员", 5)
	require.NoError(t, repo.CreateRole(role))

	require.NoError(t, repo.UpdateRoleFields(role.ID, map[string]interface{}{
		"name":   "审计员V2",
		"sort":   10,
		"status": 0,
	}))

	got, err := repo.FindRoleByID(role.ID)
	require.NoError(t, err)
	assert.Equal(t, "审计员V2", got.Name)
	assert.Equal(t, "审计员 desc", got.Description, "未更新字段应保持不变")
	assert.Equal(t, 10, got.Sort)
	assert.Equal(t, 0, got.Status)
}

// TestPermissionRepository_DeleteRole_Cascade 删除角色时级联清理 UserRole 和 RolePermission
func TestPermissionRepository_DeleteRole_Cascade(t *testing.T) {
	repo, db := newRepoForTest(t)

	role := makeRole("temp", "临时角色", 9)
	require.NoError(t, repo.CreateRole(role))

	perm := makePerm("temp:read", "临时读", 3, 1)
	require.NoError(t, repo.CreatePermission(perm))

	// 建立关联
	require.NoError(t, repo.AssignPermissionsToRole(role.ID, []uint{perm.ID}))
	// 直接插入 UserRole（user_id 假值 999）
	require.NoError(t, db.Create(&permModel.UserRole{UserID: 999, RoleID: role.ID}).Error)

	// 验证关联存在
	rps, err := repo.FindPermissionsByRoleID(role.ID)
	require.NoError(t, err)
	assert.Len(t, rps, 1)

	// 删除角色
	require.NoError(t, repo.DeleteRole(role.ID))

	// 验证 RolePermission 已级联删除
	rpsAfter, err := repo.FindPermissionsByRoleID(role.ID)
	require.NoError(t, err)
	assert.Empty(t, rpsAfter, "删除角色后 RolePermission 应被级联清理")

	// 验证 UserRole 已级联删除
	var count int64
	require.NoError(t, db.Model(&permModel.UserRole{}).Where("role_id = ?", role.ID).Count(&count).Error)
	assert.Equal(t, int64(0), count, "删除角色后 UserRole 应被级联清理")

	// 角色本身也被删除
	_, err = repo.FindRoleByID(role.ID)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

	// 权限本身保留（未被级联删）
	gotPerm, err := repo.FindPermissionByID(perm.ID)
	require.NoError(t, err)
	assert.Equal(t, "temp:read", gotPerm.Code)
}

// ========== 权限 CRUD ==========

// TestPermissionRepository_PermCreateAndFind 创建 + 按 ID/Code 查询
func TestPermissionRepository_PermCreateAndFind(t *testing.T) {
	repo, _ := newRepoForTest(t)

	p := makePerm("user:read", "用户读取", 3, 1)
	require.NoError(t, repo.CreatePermission(p))
	require.NotZero(t, p.ID)

	got, err := repo.FindPermissionByID(p.ID)
	require.NoError(t, err)
	assert.Equal(t, "user:read", got.Code)
	assert.Equal(t, 3, got.Type)

	byCode, err := repo.FindPermissionByCode("user:read")
	require.NoError(t, err)
	assert.Equal(t, got.ID, byCode.ID)
}

// TestPermissionRepository_FindPermissionByCode_NotFound 查不到返回 ErrRecordNotFound
func TestPermissionRepository_FindPermissionByCode_NotFound(t *testing.T) {
	repo, _ := newRepoForTest(t)

	_, err := repo.FindPermissionByCode("non-existent")
	require.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

// TestPermissionRepository_CreatePermission_DuplicateCode 唯一索引
func TestPermissionRepository_CreatePermission_DuplicateCode(t *testing.T) {
	repo, _ := newRepoForTest(t)

	p1 := makePerm("user:write", "用户写入", 3, 2)
	require.NoError(t, repo.CreatePermission(p1))

	p2 := makePerm("user:write", "用户写入重复", 3, 3)
	err := repo.CreatePermission(p2)
	require.Error(t, err, "重复 code 应因唯一索引失败")
}

// TestPermissionRepository_FindAllPermissions 按 type ASC, sort ASC, id ASC 排序
func TestPermissionRepository_FindAllPermissions(t *testing.T) {
	repo, _ := newRepoForTest(t)

	// type=3 接口类
	require.NoError(t, repo.CreatePermission(makePerm("c:read", "C读", 3, 3)))
	require.NoError(t, repo.CreatePermission(makePerm("a:read", "A读", 3, 1)))
	// type=1 菜单类（应排在 type=3 前面）
	require.NoError(t, repo.CreatePermission(makePerm("menu:a", "菜单A", 1, 1)))

	perms, err := repo.FindAllPermissions()
	require.NoError(t, err)
	require.Len(t, perms, 3)
	assert.Equal(t, "menu:a", perms[0].Code, "type=1 应排在 type=3 前面")
	assert.Equal(t, "a:read", perms[1].Code, "type=3 内按 sort ASC")
	assert.Equal(t, "c:read", perms[2].Code)
}

// TestPermissionRepository_UpdatePermissionFields 局部字段更新
func TestPermissionRepository_UpdatePermissionFields(t *testing.T) {
	repo, _ := newRepoForTest(t)

	p := makePerm("perm:up", "待更新", 3, 5)
	require.NoError(t, repo.CreatePermission(p))

	require.NoError(t, repo.UpdatePermissionFields(p.ID, map[string]interface{}{
		"name":   "已更新",
		"path":   "/api/v1/perm",
		"method": "GET",
		"sort":   20,
	}))

	got, err := repo.FindPermissionByID(p.ID)
	require.NoError(t, err)
	assert.Equal(t, "已更新", got.Name)
	assert.Equal(t, "/api/v1/perm", got.Path)
	assert.Equal(t, "GET", got.Method)
	assert.Equal(t, 20, got.Sort)
	assert.Equal(t, 1, got.Status, "未更新字段保持不变")
}

// TestPermissionRepository_DeletePermission_Cascade 删除权限时级联清理 RolePermission
func TestPermissionRepository_DeletePermission_Cascade(t *testing.T) {
	repo, _ := newRepoForTest(t)

	role := makeRole("owner", "所有者", 1)
	require.NoError(t, repo.CreateRole(role))

	p := makePerm("del:read", "待删除读", 3, 1)
	require.NoError(t, repo.CreatePermission(p))

	require.NoError(t, repo.AssignPermissionsToRole(role.ID, []uint{p.ID}))

	require.NoError(t, repo.DeletePermission(p.ID))

	// RolePermission 已级联清理
	rps, err := repo.FindPermissionsByRoleID(role.ID)
	require.NoError(t, err)
	assert.Empty(t, rps, "删除权限后 RolePermission 应被级联清理")

	// 权限本身已删除
	_, err = repo.FindPermissionByID(p.ID)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

	// 角色保留
	gotRole, err := repo.FindRoleByID(role.ID)
	require.NoError(t, err)
	assert.Equal(t, "owner", gotRole.Code)
}

// ========== 用户-角色关联 ==========

// TestPermissionRepository_AssignRolesToUser_Overwrite 覆盖式分配：先分配 A/B，再分配 C，最终只剩 C
func TestPermissionRepository_AssignRolesToUser_Overwrite(t *testing.T) {
	repo, _ := newRepoForTest(t)

	roleA := makeRole("ra", "角色A", 1)
	roleB := makeRole("rb", "角色B", 2)
	roleC := makeRole("rc", "角色C", 3)
	require.NoError(t, repo.CreateRole(roleA))
	require.NoError(t, repo.CreateRole(roleB))
	require.NoError(t, repo.CreateRole(roleC))

	userID := uint(1001)

	// 先分配 A、B
	require.NoError(t, repo.AssignRolesToUser(userID, []uint{roleA.ID, roleB.ID}))
	roles, err := repo.FindRolesByUserID(userID)
	require.NoError(t, err)
	assert.Len(t, roles, 2)

	// 再分配 C（覆盖）
	require.NoError(t, repo.AssignRolesToUser(userID, []uint{roleC.ID}))
	roles, err = repo.FindRolesByUserID(userID)
	require.NoError(t, err)
	require.Len(t, roles, 1, "覆盖后只剩一个角色")
	assert.Equal(t, "rc", roles[0].Code)
}

// TestPermissionRepository_AssignRolesToUser_Empty 空数组等于解除所有角色
func TestPermissionRepository_AssignRolesToUser_Empty(t *testing.T) {
	repo, _ := newRepoForTest(t)

	roleA := makeRole("ea", "角色EA", 1)
	require.NoError(t, repo.CreateRole(roleA))

	userID := uint(2002)
	require.NoError(t, repo.AssignRolesToUser(userID, []uint{roleA.ID}))
	roles, err := repo.FindRolesByUserID(userID)
	require.NoError(t, err)
	assert.Len(t, roles, 1)

	// 空数组 = 清空
	require.NoError(t, repo.AssignRolesToUser(userID, []uint{}))
	roles, err = repo.FindRolesByUserID(userID)
	require.NoError(t, err)
	assert.Empty(t, roles)
}

// TestPermissionRepository_ClearUserRoles 显式清空用户角色
func TestPermissionRepository_ClearUserRoles(t *testing.T) {
	repo, _ := newRepoForTest(t)

	role := makeRole("clr", "清空测试", 1)
	require.NoError(t, repo.CreateRole(role))

	userID := uint(3003)
	require.NoError(t, repo.AssignRolesToUser(userID, []uint{role.ID}))

	require.NoError(t, repo.ClearUserRoles(userID))
	roles, err := repo.FindRolesByUserID(userID)
	require.NoError(t, err)
	assert.Empty(t, roles)
}

// ========== 角色-权限关联 ==========

// TestPermissionRepository_AssignPermissionsToRole_Overwrite 覆盖式分配
func TestPermissionRepository_AssignPermissionsToRole_Overwrite(t *testing.T) {
	repo, _ := newRepoForTest(t)

	role := makeRole("ovr", "覆盖测试", 1)
	require.NoError(t, repo.CreateRole(role))

	p1 := makePerm("ovr:r1", "权限1", 3, 1)
	p2 := makePerm("ovr:r2", "权限2", 3, 2)
	p3 := makePerm("ovr:r3", "权限3", 3, 3)
	require.NoError(t, repo.CreatePermission(p1))
	require.NoError(t, repo.CreatePermission(p2))
	require.NoError(t, repo.CreatePermission(p3))

	// 先分配 p1, p2
	require.NoError(t, repo.AssignPermissionsToRole(role.ID, []uint{p1.ID, p2.ID}))
	perms, err := repo.FindPermissionsByRoleID(role.ID)
	require.NoError(t, err)
	assert.Len(t, perms, 2)

	// 覆盖为 p3
	require.NoError(t, repo.AssignPermissionsToRole(role.ID, []uint{p3.ID}))
	perms, err = repo.FindPermissionsByRoleID(role.ID)
	require.NoError(t, err)
	require.Len(t, perms, 1)
	assert.Equal(t, "ovr:r3", perms[0].Code)
}

// TestPermissionRepository_ClearRolePermissions 显式清空角色权限
func TestPermissionRepository_ClearRolePermissions(t *testing.T) {
	repo, _ := newRepoForTest(t)

	role := makeRole("clrp", "清空权限测试", 1)
	require.NoError(t, repo.CreateRole(role))
	p := makePerm("clrp:r", "权限", 3, 1)
	require.NoError(t, repo.CreatePermission(p))

	require.NoError(t, repo.AssignPermissionsToRole(role.ID, []uint{p.ID}))
	require.NoError(t, repo.ClearRolePermissions(role.ID))

	perms, err := repo.FindPermissionsByRoleID(role.ID)
	require.NoError(t, err)
	assert.Empty(t, perms)
}

// ========== 综合查询（RBAC 核心） ==========

// TestPermissionRepository_FindPermissionsByUserID_MultiRoleDedup
// 用户拥有多个角色，权限去重后返回
func TestPermissionRepository_FindPermissionsByUserID_MultiRoleDedup(t *testing.T) {
	repo, _ := newRepoForTest(t)

	// 角色 A 拥有 perm1, perm2；角色 B 拥有 perm2, perm3
	roleA := makeRole("ra", "角色A", 1)
	roleB := makeRole("rb", "角色B", 2)
	require.NoError(t, repo.CreateRole(roleA))
	require.NoError(t, repo.CreateRole(roleB))

	p1 := makePerm("p1", "权限1", 3, 1)
	p2 := makePerm("p2", "权限2", 3, 2)
	p3 := makePerm("p3", "权限3", 3, 3)
	require.NoError(t, repo.CreatePermission(p1))
	require.NoError(t, repo.CreatePermission(p2))
	require.NoError(t, repo.CreatePermission(p3))

	require.NoError(t, repo.AssignPermissionsToRole(roleA.ID, []uint{p1.ID, p2.ID}))
	require.NoError(t, repo.AssignPermissionsToRole(roleB.ID, []uint{p2.ID, p3.ID}))

	userID := uint(4004)
	require.NoError(t, repo.AssignRolesToUser(userID, []uint{roleA.ID, roleB.ID}))

	// 综合查询应去重（perm2 在两个角色都有，但只返回一次）
	perms, err := repo.FindPermissionsByUserID(userID)
	require.NoError(t, err)
	require.Len(t, perms, 3, "perm2 应去重，最终 3 个权限")

	// 按 sort ASC 排序
	assert.Equal(t, "p1", perms[0].Code)
	assert.Equal(t, "p2", perms[1].Code)
	assert.Equal(t, "p3", perms[2].Code)
}

// TestPermissionRepository_FindPermissionCodesByUserID_ExcludesDisabled
// status=0 的权限不应返回（禁用权限不生效）
func TestPermissionRepository_FindPermissionCodesByUserID_ExcludesDisabled(t *testing.T) {
	repo, _ := newRepoForTest(t)

	role := makeRole("dis", "禁用权限测试角色", 1)
	require.NoError(t, repo.CreateRole(role))

	pEnabled := makePerm("en:r", "启用权限", 3, 1)
	pDisabled := makePerm("dis:r", "禁用权限", 3, 2)
	pDisabled.Status = 0 // 禁用
	require.NoError(t, repo.CreatePermission(pEnabled))
	require.NoError(t, repo.CreatePermission(pDisabled))

	require.NoError(t, repo.AssignPermissionsToRole(role.ID, []uint{pEnabled.ID, pDisabled.ID}))

	userID := uint(5005)
	require.NoError(t, repo.AssignRolesToUser(userID, []uint{role.ID}))

	codes, err := repo.FindPermissionCodesByUserID(userID)
	require.NoError(t, err)
	require.Len(t, codes, 1, "禁用权限不应返回")
	assert.Equal(t, "en:r", codes[0])
}

// TestPermissionRepository_FindRoleCodesByUserID_ExcludesDisabled
// status=0 的角色不应返回
func TestPermissionRepository_FindRoleCodesByUserID_ExcludesDisabled(t *testing.T) {
	repo, _ := newRepoForTest(t)

	roleOn := makeRole("ron", "启用角色", 1)
	roleOff := makeRole("roff", "禁用角色", 2)
	roleOff.Status = 0 // 禁用
	require.NoError(t, repo.CreateRole(roleOn))
	require.NoError(t, repo.CreateRole(roleOff))

	userID := uint(6006)
	require.NoError(t, repo.AssignRolesToUser(userID, []uint{roleOn.ID, roleOff.ID}))

	codes, err := repo.FindRoleCodesByUserID(userID)
	require.NoError(t, err)
	require.Len(t, codes, 1, "禁用角色不应返回")
	assert.Equal(t, "ron", codes[0])
}

// TestPermissionRepository_FindPermissionsByUserID_NoAssignment
// 用户未分配任何角色时返回空列表
func TestPermissionRepository_FindPermissionsByUserID_NoAssignment(t *testing.T) {
	repo, _ := newRepoForTest(t)

	perms, err := repo.FindPermissionsByUserID(99999)
	require.NoError(t, err)
	assert.Empty(t, perms)

	codes, err := repo.FindPermissionCodesByUserID(99999)
	require.NoError(t, err)
	assert.Empty(t, codes)

	roleCodes, err := repo.FindRoleCodesByUserID(99999)
	require.NoError(t, err)
	assert.Empty(t, roleCodes)
}
