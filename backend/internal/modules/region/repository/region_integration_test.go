// Package repository 地区仓储的集成测试。
// 使用 testcontainers 启动真实 PostgreSQL，验证 RegionRepository 的
// 树形结构（省/市/区县）、FindByParentID 排序、FindByCode 唯一索引、软删除等。
// 无 Docker 时自动 skip。
package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	regionModel "wuchang-tongcheng/internal/modules/region/model"
	"wuchang-tongcheng/internal/testutil/pgtest"
)

func newRegionRepoForTest(t *testing.T) (*regionRepository, *gorm.DB) {
	t.Helper()
	db := pgtest.SetupPostgres(t)
	pgtest.MigrateAll(t, db)
	return &regionRepository{db: db}, db
}

// 构造三级地区树：湖北省(1) → 武汉市(2) → 武昌区/洪山区/江夏区(3)
// 返回所有创建的 region，顺序与入参一致。
func seedRegionTree(t *testing.T, repo *regionRepository) (
	province, city *regionModel.Region,
	districts [3]*regionModel.Region,
) {
	t.Helper()
	province = &regionModel.Region{Name: "湖北省", Code: "420000", Level: 1, Sort: 1, Status: 1}
	require.NoError(t, repo.Create(province))

	city = &regionModel.Region{Name: "武汉市", Code: "420100", Level: 2, ParentID: province.ID, Sort: 1, Status: 1}
	require.NoError(t, repo.Create(city))

	districts[0] = &regionModel.Region{Name: "武昌区", Code: "420106", Level: 3, ParentID: city.ID, Sort: 1, Status: 1}
	districts[1] = &regionModel.Region{Name: "洪山区", Code: "420111", Level: 3, ParentID: city.ID, Sort: 2, Status: 1}
	districts[2] = &regionModel.Region{Name: "江夏区", Code: "420115", Level: 3, ParentID: city.ID, Sort: 3, Status: 1}
	for _, d := range districts {
		require.NoError(t, repo.Create(d))
	}
	return province, city, districts
}

// TestRegionRepository_CreateAndFindByID 基础 CRUD + ID 查询
func TestRegionRepository_CreateAndFindByID(t *testing.T) {
	repo, _ := newRegionRepoForTest(t)

	r := &regionModel.Region{Name: "测试省", Code: "990000", Level: 1, Sort: 1, Status: 1}
	require.NoError(t, repo.Create(r))
	require.NotZero(t, r.ID)

	got, err := repo.FindByID(r.ID)
	require.NoError(t, err)
	assert.Equal(t, "测试省", got.Name)
	assert.Equal(t, "990000", got.Code)
	assert.Equal(t, 1, got.Level)
}

// TestRegionRepository_FindByCode 唯一索引：按编码查询 + 重复冲突
func TestRegionRepository_FindByCode(t *testing.T) {
	repo, _ := newRegionRepoForTest(t)

	require.NoError(t, repo.Create(&regionModel.Region{Name: "A省", Code: "110000", Level: 1, Status: 1}))

	got, err := repo.FindByCode("110000")
	require.NoError(t, err)
	assert.Equal(t, "A省", got.Name)

	// 重复 code 触发唯一索引
	err = repo.Create(&regionModel.Region{Name: "B省", Code: "110000", Level: 1, Status: 1})
	assert.Error(t, err, "重复 code 应触发唯一索引冲突")

	// 查不到
	_, err = repo.FindByCode("999999")
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

// TestRegionRepository_FindByParentID 树形查询 + sort 排序
func TestRegionRepository_FindByParentID(t *testing.T) {
	repo, _ := newRegionRepoForTest(t)
	_, city, districts := seedRegionTree(t, repo)

	children, err := repo.FindByParentID(city.ID)
	require.NoError(t, err)
	require.Len(t, children, 3)

	// 验证按 sort ASC 排序
	assert.Equal(t, "武昌区", children[0].Name)
	assert.Equal(t, "洪山区", children[1].Name)
	assert.Equal(t, "江夏区", children[2].Name)

	// 验证 ID 与预期一致
	for i, d := range children {
		assert.Equal(t, districts[i].ID, d.ID)
	}
}

// TestRegionRepository_FindAll 全量查询按 level/sort 排序
func TestRegionRepository_FindAll(t *testing.T) {
	repo, _ := newRegionRepoForTest(t)
	province, city, districts := seedRegionTree(t, repo)

	all, err := repo.FindAll()
	require.NoError(t, err)
	require.Len(t, all, 5)

	// 期望顺序：省(level1) → 市(level2) → 3 个区(level3, 按 sort)
	assert.Equal(t, province.ID, all[0].ID)
	assert.Equal(t, city.ID, all[1].ID)
	assert.Equal(t, districts[0].ID, all[2].ID)
	assert.Equal(t, districts[1].ID, all[3].ID)
	assert.Equal(t, districts[2].ID, all[4].ID)
}

// TestRegionRepository_UpdateFields 部分更新（改名 + 排序）
func TestRegionRepository_UpdateFields(t *testing.T) {
	repo, _ := newRegionRepoForTest(t)

	r := &regionModel.Region{Name: "旧名", Code: "880000", Level: 1, Sort: 5, Status: 1}
	require.NoError(t, repo.Create(r))

	require.NoError(t, repo.UpdateFields(r.ID, map[string]interface{}{
		"name": "新名",
		"sort": 99,
	}))

	got, err := repo.FindByID(r.ID)
	require.NoError(t, err)
	assert.Equal(t, "新名", got.Name)
	assert.Equal(t, 99, got.Sort)
	assert.Equal(t, "880000", got.Code, "未更新字段保留")
}

// TestRegionRepository_Delete_SoftDelete 软删除：FindByID 查不到，Unscoped 可见
func TestRegionRepository_Delete_SoftDelete(t *testing.T) {
	repo, db := newRegionRepoForTest(t)

	r := &regionModel.Region{Name: "待删", Code: "770000", Level: 1, Status: 1}
	require.NoError(t, repo.Create(r))
	require.NoError(t, repo.Delete(r.ID))

	_, err := repo.FindByID(r.ID)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

	var raw regionModel.Region
	require.NoError(t, db.Unscoped().First(&raw, r.ID).Error)
	assert.False(t, raw.DeletedAt.Time.IsZero())
}

// TestRegionRepository_Delete_CascadeChildrenNotAuto 删除父节点不会级联删子节点
// （业务约束：删地区前应先处理子节点，repo 不做级联）
func TestRegionRepository_Delete_CascadeChildrenNotAuto(t *testing.T) {
	repo, db := newRegionRepoForTest(t)
	_, city, _ := seedRegionTree(t, repo)

	// 删除市，子区县应仍存在（业务需上层保证不删有子的节点）
	require.NoError(t, repo.Delete(city.ID))

	var childCount int64
	require.NoError(t, db.Model(&regionModel.Region{}).Where("parent_id = ?", city.ID).Count(&childCount).Error)
	assert.Equal(t, int64(3), childCount, "repo 不做级联删除，子节点仍存在")
}
