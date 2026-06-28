// Package repository 用户仓储的集成测试。
// 使用 testcontainers 启动真实 PostgreSQL，验证 UserRepository 的 CRUD、
// 唯一索引、按地区/关键词/状态过滤的 List、软删除等行为。
// 无 Docker 时自动 skip。
package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	userModel "wuchang-tongcheng/internal/modules/user/model"
	"wuchang-tongcheng/internal/pkg/utils"
	"wuchang-tongcheng/internal/testutil/pgtest"
)

// newRepoForTest 启动 PG 容器 + 全量建表 + 构造 UserRepository。
// 失败/无 Docker 时会 skip，调用方无需处理 error。
func newRepoForTest(t *testing.T) (*userRepository, *gorm.DB) {
	t.Helper()
	db := pgtest.SetupPostgres(t)
	pgtest.MigrateAll(t, db)
	return &userRepository{db: db}, db
}

func makeUser(name, pwd, nick string, regionID uint, status int) *userModel.User {
	u := &userModel.User{
		Username: name,
		Password: pwd,
		Nickname: nick,
		Email:    name + "@example.com",
		Gender:   0,
		Status:   status,
	}
	u.RegionID = regionID
	return u
}

// TestUserRepository_CreateAndFindByID 创建 + 按 ID 查询 + 字段回填
func TestUserRepository_CreateAndFindByID(t *testing.T) {
	repo, _ := newRepoForTest(t)

	u := makeUser("alice", "$2a$10$hashedpwdhashedpwdhashedpwdhashed", "Alice", 2, 1)
	require.NoError(t, repo.Create(u))
	require.NotZero(t, u.ID)

	got, err := repo.FindByID(u.ID)
	require.NoError(t, err)
	assert.Equal(t, "alice", got.Username)
	assert.Equal(t, "Alice", got.Nickname)
	assert.Equal(t, uint(2), got.RegionID)
	assert.Equal(t, 1, got.Status)
	assert.False(t, got.CreatedAt.IsZero())
}

// TestUserRepository_FindByUsername_NotFound 查不到返回 ErrRecordNotFound
func TestUserRepository_FindByUsername_NotFound(t *testing.T) {
	repo, _ := newRepoForTest(t)

	_, err := repo.FindByUsername("nobody")
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

// TestUserRepository_FindByUsername_UniqueIndex 重复用户名违反唯一索引
func TestUserRepository_FindByUsername_UniqueIndex(t *testing.T) {
	repo, _ := newRepoForTest(t)

	require.NoError(t, repo.Create(makeUser("dup", "p1", "Dup1", 2, 1)))
	err := repo.Create(makeUser("dup", "p2", "Dup2", 2, 1))
	assert.Error(t, err, "重复 username 应触发唯一索引冲突")
}

// TestUserRepository_UpdateFields 部分字段更新（保留未更新字段）
func TestUserRepository_UpdateFields(t *testing.T) {
	repo, _ := newRepoForTest(t)

	u := makeUser("bob", "pwd", "Bob", 2, 1)
	require.NoError(t, repo.Create(u))

	require.NoError(t, repo.UpdateFields(u.ID, map[string]interface{}{
		"nickname": "Bob2",
		"gender":   2,
		"status":   0,
	}))

	got, err := repo.FindByID(u.ID)
	require.NoError(t, err)
	assert.Equal(t, "Bob2", got.Nickname)
	assert.Equal(t, 2, got.Gender)
	assert.Equal(t, 0, got.Status)
	// 未更新字段保留
	assert.Equal(t, "bob", got.Username)
}

// TestUserRepository_List_Filters 地区/关键词/状态三维过滤 + 分页
func TestUserRepository_List_Filters(t *testing.T) {
	repo, _ := newRepoForTest(t)

	// 武汉市(2) 4 个，洪山区(5) 2 个
	require.NoError(t, repo.Create(makeUser("wu1", "p", "Wu One", 2, 1)))
	require.NoError(t, repo.Create(makeUser("wu2", "p", "Wu Two", 2, 1)))
	require.NoError(t, repo.Create(makeUser("wu3", "p", "Alpha", 2, 0))) // 禁用
	require.NoError(t, repo.Create(makeUser("wu4", "p", "Beta Wu", 2, 1)))
	require.NoError(t, repo.Create(makeUser("hong1", "p", "Hong", 5, 1)))
	require.NoError(t, repo.Create(makeUser("hong2", "p", "Honger", 5, 1)))

	// 注：user repo 的 List 中，status==0 表示过滤"禁用"用户（WHERE status=0），
	// status==1 过滤"正常"，其他值（如 2）则不加状态过滤（返回全部）。
	pg := utils.NewPagination(1, 10)

	// 1) 仅按地区、不过滤状态：武汉市 4 条（含禁用）
	list, total, err := repo.List(2, pg, "", 2)
	require.NoError(t, err)
	assert.Equal(t, int64(4), total)
	assert.Len(t, list, 4)

	// 2) 地区 + 状态过滤：武汉市正常 3 条
	pg2 := utils.NewPagination(1, 10)
	list, total, err = repo.List(2, pg2, "", 1)
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, list, 3)
	for _, u := range list {
		assert.Equal(t, 1, u.Status)
	}

	// 3) 地区 + 关键词："wu" 匹配 username 或 nickname
	pg3 := utils.NewPagination(1, 10)
	list, total, err = repo.List(2, pg3, "wu", 0)
	require.NoError(t, err)
	// wu1/wu2/wu4(username) + Wu One/Wu Two/Beta Wu(nickname) = 3 条
	assert.Equal(t, int64(3), total)

	// 4) regionID=0 表示跨区（超管），应返回全部 6 条
	pg4 := utils.NewPagination(1, 10)
	_, total, err = repo.List(0, pg4, "", 1)
	require.NoError(t, err)
	assert.Equal(t, int64(5), total, "跨区正常用户共 5 条")
}

// TestUserRepository_List_Pagination 分页边界
func TestUserRepository_List_Pagination(t *testing.T) {
	repo, _ := newRepoForTest(t)

	for i := 0; i < 15; i++ {
		require.NoError(t, repo.Create(makeUser(
			"u"+string(rune('a'+i)), "p", "N"+string(rune('a'+i)), 2, 1)))
	}

	// pageSize=10 第 1 页应返回 10 条，total=15
	pg := utils.NewPagination(1, 10)
	list, total, err := repo.List(2, pg, "", 1)
	require.NoError(t, err)
	assert.Equal(t, int64(15), total)
	assert.Len(t, list, 10)

	// pageSize=10 第 2 页应返回 5 条
	pg2 := utils.NewPagination(2, 10)
	list, _, err = repo.List(2, pg2, "", 1)
	require.NoError(t, err)
	assert.Len(t, list, 5)
}

// TestUserRepository_Delete_SoftDelete 软删除：FindByID 查不到，但物理记录仍在
func TestUserRepository_Delete_SoftDelete(t *testing.T) {
	repo, db := newRepoForTest(t)

	u := makeUser("delme", "p", "Del", 2, 1)
	require.NoError(t, repo.Create(u))
	require.NoError(t, repo.Delete(u.ID))

	// 软删除后 repository 查不到
	_, err := repo.FindByID(u.ID)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

	// Unscoped 可查到物理记录，deleted_at 非空
	var raw userModel.User
	require.NoError(t, db.Unscoped().First(&raw, u.ID).Error)
	assert.False(t, raw.DeletedAt.Time.IsZero())
}
