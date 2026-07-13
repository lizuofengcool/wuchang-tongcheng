// Package service 地区服务单元测试。
// 使用内存 mock repository，覆盖缓存键生成、DTO 转换、层级计算、
// 最大层级限制、父地区校验、编码去重、树形构建等核心业务逻辑，不依赖 DB/Redis。
package service

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"wuchang-tongcheng/internal/modules/region/dto"
	"wuchang-tongcheng/internal/modules/region/model"
)

// mockRegionRepo 内存 mock，实现 RegionRepository 接口
type mockRegionRepo struct {
	byID        map[uint]*model.Region
	byParent    map[uint][]model.Region // parentID -> children
	byCode      map[string]*model.Region
	all         []model.Region
	nextID      uint
	createErr   error
	findErr     error
	deleteErr   error
	updateErr   error
	findCodeErr error
}

func newMockRegionRepo() *mockRegionRepo {
	return &mockRegionRepo{
		byID:     make(map[uint]*model.Region),
		byParent: make(map[uint][]model.Region),
		byCode:   make(map[string]*model.Region),
		nextID:   1,
	}
}

func (m *mockRegionRepo) Create(region *model.Region) error {
	if m.createErr != nil {
		return m.createErr
	}
	region.ID = m.nextID
	m.nextID++
	cp := *region
	m.byID[region.ID] = &cp
	m.byCode[region.Code] = &cp
	m.byParent[region.ParentID] = append(m.byParent[region.ParentID], cp)
	m.all = append(m.all, cp)
	return nil
}

func (m *mockRegionRepo) FindByID(id uint) (*model.Region, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	r, ok := m.byID[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *r
	return &cp, nil
}

func (m *mockRegionRepo) FindByCode(code string) (*model.Region, error) {
	if m.findCodeErr != nil {
		return nil, m.findCodeErr
	}
	r, ok := m.byCode[code]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *r
	return &cp, nil
}

func (m *mockRegionRepo) FindByParentID(parentID uint) ([]model.Region, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	out := make([]model.Region, 0, len(m.byParent[parentID]))
	out = append(out, m.byParent[parentID]...)
	return out, nil
}

func (m *mockRegionRepo) FindAll() ([]model.Region, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	out := make([]model.Region, 0, len(m.all))
	out = append(out, m.all...)
	return out, nil
}

func (m *mockRegionRepo) Update(region *model.Region) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if _, ok := m.byID[region.ID]; !ok {
		return gorm.ErrRecordNotFound
	}
	cp := *region
	m.byID[region.ID] = &cp
	return nil
}

func (m *mockRegionRepo) UpdateFields(id uint, fields map[string]interface{}) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	r, ok := m.byID[id]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	if v, ok := fields["name"]; ok {
		r.Name = v.(string)
	}
	if v, ok := fields["sort"]; ok {
		r.Sort = v.(int)
	}
	if v, ok := fields["status"]; ok {
		r.Status = v.(int)
	}
	return nil
}

func (m *mockRegionRepo) Delete(id uint) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if _, ok := m.byID[id]; !ok {
		return gorm.ErrRecordNotFound
	}
	delete(m.byID, id)
	return nil
}

// ===== 纯函数 =====

func TestRegionCacheKeyTree(t *testing.T) {
	assert.Equal(t, "cache:region:tree", regionCacheKeyTree())
}

func TestRegionCacheKeyByID(t *testing.T) {
	assert.Equal(t, "cache:region:id:1", regionCacheKeyByID(1))
	assert.Equal(t, "cache:region:id:0", regionCacheKeyByID(0))
	assert.Equal(t, "cache:region:id:8888", regionCacheKeyByID(8888))
}

func TestRegionCacheKeyByParent(t *testing.T) {
	assert.Equal(t, "cache:region:parent:1", regionCacheKeyByParent(1))
	assert.Equal(t, "cache:region:parent:0", regionCacheKeyByParent(0))
	assert.Equal(t, "cache:region:parent:100", regionCacheKeyByParent(100))
}

func TestToRegionInfo(t *testing.T) {
	r := &model.Region{
		Name:     "五常市",
		Code:     "230182",
		ParentID: 2,
		Level:    3,
		Sort:     10,
		Status:   1,
	}
	r.ID = 5

	info := toRegionInfo(r)
	assert.Equal(t, uint(5), info.ID)
	assert.Equal(t, "五常市", info.Name)
	assert.Equal(t, "230182", info.Code)
	assert.Equal(t, uint(2), info.ParentID)
	assert.Equal(t, 3, info.Level)
	assert.Equal(t, 10, info.Sort)
	assert.Equal(t, 1, info.Status)
}

func TestToRegionInfo_NilSafe(t *testing.T) {
	// 即使 ID 等嵌入字段未设置也不应 panic
	r := &model.Region{Name: "空地区"}
	info := toRegionInfo(r)
	assert.Equal(t, "空地区", info.Name)
	assert.Equal(t, uint(0), info.ID)
	assert.Equal(t, 0, info.Level)
	assert.Equal(t, 0, info.Status)
}

// ===== 构造函数 =====

func TestNewRegionService(t *testing.T) {
	repo := newMockRegionRepo()
	svc := NewRegionService(repo)
	assert.NotNil(t, svc)
	// 应返回可用的 RegionService 接口实例
	_, ok := svc.(*regionService)
	assert.True(t, ok, "返回值应为 *regionService 类型")
}

// ===== Create =====

func TestRegionService_Create_TopLevel(t *testing.T) {
	repo := newMockRegionRepo()
	svc := NewRegionService(repo)

	info, err := svc.Create(&dto.CreateRegionRequest{
		Name:   "黑龙江省",
		Code:   "230000",
		Sort:   5,
		Status: 1,
	})
	require.NoError(t, err)
	assert.Equal(t, "黑龙江省", info.Name)
	assert.Equal(t, "230000", info.Code)
	assert.Equal(t, 1, info.Level, "ParentID=0 应为顶级，level=1")
	assert.Equal(t, uint(0), info.ParentID)
	assert.Equal(t, 5, info.Sort)
	assert.Equal(t, 1, info.Status)
}

func TestRegionService_Create_DefaultStatusToOne(t *testing.T) {
	repo := newMockRegionRepo()
	svc := NewRegionService(repo)

	// Status=0 时应默认填充为 1
	info, err := svc.Create(&dto.CreateRegionRequest{
		Name:   "哈尔滨市",
		Code:   "230100",
		Status: 0,
	})
	require.NoError(t, err)
	assert.Equal(t, 1, info.Status, "Status=0 应默认填充为 1")
}

func TestRegionService_Create_ChildLevel(t *testing.T) {
	repo := newMockRegionRepo()
	svc := NewRegionService(repo)

	// 先建顶级
	parent, err := svc.Create(&dto.CreateRegionRequest{Name: "黑龙江省", Code: "230000", Status: 1})
	require.NoError(t, err)
	require.Equal(t, 1, parent.Level)

	// 再建子级，level 应为父级+1=2
	child, err := svc.Create(&dto.CreateRegionRequest{
		Name:     "哈尔滨市",
		Code:     "230100",
		ParentID: parent.ID,
		Status:   1,
	})
	require.NoError(t, err)
	assert.Equal(t, 2, child.Level, "子级 level 应为父级 level+1")
	assert.Equal(t, parent.ID, child.ParentID)
}

func TestRegionService_Create_GrandchildLevel(t *testing.T) {
	repo := newMockRegionRepo()
	svc := NewRegionService(repo)

	root, _ := svc.Create(&dto.CreateRegionRequest{Name: "黑龙江省", Code: "230000", Status: 1})
	child, _ := svc.Create(&dto.CreateRegionRequest{Name: "哈尔滨市", Code: "230100", ParentID: root.ID, Status: 1})
	require.Equal(t, 2, child.Level)

	// 第三层（最大允许层级）
	grandchild, err := svc.Create(&dto.CreateRegionRequest{
		Name:     "五常市",
		Code:     "230182",
		ParentID: child.ID,
		Status:   1,
	})
	require.NoError(t, err)
	assert.Equal(t, 3, grandchild.Level)
}

func TestRegionService_Create_ExceedMaxLevel(t *testing.T) {
	repo := newMockRegionRepo()
	svc := NewRegionService(repo)

	root, _ := svc.Create(&dto.CreateRegionRequest{Name: "黑龙江省", Code: "230000", Status: 1})
	child, _ := svc.Create(&dto.CreateRegionRequest{Name: "哈尔滨市", Code: "230100", ParentID: root.ID, Status: 1})
	grandchild, _ := svc.Create(&dto.CreateRegionRequest{Name: "五常市", Code: "230182", ParentID: child.ID, Status: 1})
	require.Equal(t, 3, grandchild.Level)

	// 第四层应被拒绝（MaxRegionLevel=3）
	_, err := svc.Create(&dto.CreateRegionRequest{
		Name:     "山河镇",
		Code:     "230182100",
		ParentID: grandchild.ID,
		Status:   1,
	})
	assert.ErrorIs(t, err, ErrRegionMaxLevel)
}

func TestRegionService_Create_ParentNotFound(t *testing.T) {
	repo := newMockRegionRepo()
	svc := NewRegionService(repo)

	_, err := svc.Create(&dto.CreateRegionRequest{
		Name:     "孤儿地区",
		Code:     "999999",
		ParentID: 9999,
		Status:   1,
	})
	assert.ErrorIs(t, err, ErrRegionParentInvalid)
}

func TestRegionService_Create_CodeExists(t *testing.T) {
	repo := newMockRegionRepo()
	svc := NewRegionService(repo)

	_, err := svc.Create(&dto.CreateRegionRequest{Name: "黑龙江省", Code: "230000", Status: 1})
	require.NoError(t, err)

	// 重复编码应被拒绝
	_, err = svc.Create(&dto.CreateRegionRequest{Name: "重名", Code: "230000", Status: 1})
	assert.ErrorIs(t, err, ErrRegionCodeExists)
}

func TestRegionService_Create_RepoError(t *testing.T) {
	repo := newMockRegionRepo()
	boom := errors.New("db down")
	repo.createErr = boom
	svc := NewRegionService(repo)

	_, err := svc.Create(&dto.CreateRegionRequest{Name: "x", Code: "y", Status: 1})
	assert.ErrorIs(t, err, boom)
}

func TestRegionService_Create_ParentFindError(t *testing.T) {
	repo := newMockRegionRepo()
	repo.findErr = errors.New("connection refused")
	svc := NewRegionService(repo)

	// 父地区查询失败（非 ErrRecordNotFound）应原样透传错误
	_, err := svc.Create(&dto.CreateRegionRequest{
		Name:     "x",
		Code:     "y",
		ParentID: 5,
		Status:   1,
	})
	assert.EqualError(t, err, "connection refused")
}

func TestRegionService_Create_FindCodeError(t *testing.T) {
	repo := newMockRegionRepo()
	repo.findCodeErr = errors.New("code query failed")
	svc := NewRegionService(repo)

	// 编码查询失败（非 ErrRecordNotFound）应原样透传错误
	_, err := svc.Create(&dto.CreateRegionRequest{
		Name:   "x",
		Code:   "y",
		Status: 1,
	})
	assert.EqualError(t, err, "code query failed")
}

// ===== GetTree =====

func TestRegionService_GetTree_Empty(t *testing.T) {
	repo := newMockRegionRepo()
	svc := NewRegionService(repo)

	tree, err := svc.GetTree()
	require.NoError(t, err)
	assert.Empty(t, tree)
}

func TestRegionService_GetTree_MultiLevel(t *testing.T) {
	repo := newMockRegionRepo()
	// 直接塞入 mock 数据：3 层树
	//   root1 (id=1, parent=0, level=1)
	//     child1 (id=2, parent=1, level=2)
	//       grandchild1 (id=3, parent=2, level=3)
	//   root2 (id=4, parent=0, level=1)
	// ID 在嵌入的 BaseModel 中，构造后单独赋值
	regions := []model.Region{
		{Name: "root1", Code: "r1", ParentID: 0, Level: 1, Sort: 10},
		{Name: "child1", Code: "c1", ParentID: 1, Level: 2, Sort: 5},
		{Name: "grandchild1", Code: "g1", ParentID: 2, Level: 3, Sort: 0},
		{Name: "root2", Code: "r2", ParentID: 0, Level: 1, Sort: 20},
	}
	for i := range regions {
		regions[i].ID = uint(i + 1)
	}
	repo.all = regions
	svc := NewRegionService(repo)

	tree, err := svc.GetTree()
	require.NoError(t, err)
	require.Len(t, tree, 2, "根节点应有 2 个")

	// mock 不排序，按塞入顺序
	assert.Equal(t, "root1", tree[0].Name)
	assert.Equal(t, "root2", tree[1].Name)

	require.Len(t, tree[0].Children, 1)
	assert.Equal(t, "child1", tree[0].Children[0].Name)

	require.Len(t, tree[0].Children[0].Children, 1)
	assert.Equal(t, "grandchild1", tree[0].Children[0].Children[0].Name)

	// root2 无子节点
	assert.Empty(t, tree[1].Children)
}

func TestRegionService_GetTree_RepoError(t *testing.T) {
	repo := newMockRegionRepo()
	repo.findErr = errors.New("db error")
	svc := NewRegionService(repo)

	_, err := svc.GetTree()
	assert.EqualError(t, err, "db error")
}

// ===== GetAll =====

func TestRegionService_GetAll_Empty(t *testing.T) {
	repo := newMockRegionRepo()
	svc := NewRegionService(repo)

	list, err := svc.GetAll()
	require.NoError(t, err)
	assert.Empty(t, list)
}

func TestRegionService_GetAll_FlatList(t *testing.T) {
	repo := newMockRegionRepo()
	regions := []model.Region{
		{Name: "a", Code: "ca", ParentID: 0, Level: 1, Sort: 1, Status: 1},
		{Name: "b", Code: "cb", ParentID: 0, Level: 1, Sort: 2, Status: 1},
		{Name: "c", Code: "cc", ParentID: 1, Level: 2, Sort: 0, Status: 1},
	}
	for i := range regions {
		regions[i].ID = uint(i + 1)
	}
	repo.all = regions
	svc := NewRegionService(repo)

	list, err := svc.GetAll()
	require.NoError(t, err)
	require.Len(t, list, 3)
	assert.Equal(t, "a", list[0].Name)
	assert.Equal(t, "b", list[1].Name)
	assert.Equal(t, "c", list[2].Name)
	// 平铺列表应包含层级信息
	assert.Equal(t, 1, list[0].Level)
	assert.Equal(t, 2, list[2].Level)
}

func TestRegionService_GetAll_RepoError(t *testing.T) {
	repo := newMockRegionRepo()
	repo.findErr = errors.New("db error")
	svc := NewRegionService(repo)

	_, err := svc.GetAll()
	assert.EqualError(t, err, "db error")
}

// ===== GetByParentID =====

func TestRegionService_GetByParentID_Empty(t *testing.T) {
	repo := newMockRegionRepo()
	svc := NewRegionService(repo)

	list, err := svc.GetByParentID(0)
	require.NoError(t, err)
	assert.Empty(t, list)
}

func TestRegionService_GetByParentID_Success(t *testing.T) {
	repo := newMockRegionRepo()
	parent := model.Region{Name: "parent", Code: "p", ParentID: 0, Level: 1, Sort: 0}
	parent.ID = 1
	child1 := model.Region{Name: "child1", Code: "c1", ParentID: 1, Level: 2, Sort: 1}
	child1.ID = 2
	child2 := model.Region{Name: "child2", Code: "c2", ParentID: 1, Level: 2, Sort: 2}
	child2.ID = 3
	repo.byParent[1] = []model.Region{child1, child2}
	svc := NewRegionService(repo)

	list, err := svc.GetByParentID(1)
	require.NoError(t, err)
	require.Len(t, list, 2)
	assert.Equal(t, "child1", list[0].Name)
	assert.Equal(t, "child2", list[1].Name)
}

func TestRegionService_GetByParentID_RepoError(t *testing.T) {
	repo := newMockRegionRepo()
	repo.findErr = errors.New("db error")
	svc := NewRegionService(repo)

	_, err := svc.GetByParentID(0)
	assert.EqualError(t, err, "db error")
}

// ===== GetByID =====

func TestRegionService_GetByID_NotFound(t *testing.T) {
	repo := newMockRegionRepo()
	svc := NewRegionService(repo)

	_, err := svc.GetByID(9999)
	assert.ErrorIs(t, err, ErrRegionNotFound)
}

func TestRegionService_GetByID_Success(t *testing.T) {
	repo := newMockRegionRepo()
	r := model.Region{Name: "五常市", Code: "230182", ParentID: 1, Level: 3, Sort: 5, Status: 1}
	r.ID = 2
	repo.byID[2] = &r
	svc := NewRegionService(repo)

	info, err := svc.GetByID(2)
	require.NoError(t, err)
	assert.Equal(t, "五常市", info.Name)
	assert.Equal(t, "230182", info.Code)
	assert.Equal(t, 3, info.Level)
}

func TestRegionService_GetByID_RepoError(t *testing.T) {
	repo := newMockRegionRepo()
	repo.findErr = errors.New("db error")
	svc := NewRegionService(repo)

	_, err := svc.GetByID(1)
	assert.EqualError(t, err, "db error")
}

// ===== Delete =====

func TestRegionService_Delete_NotFound(t *testing.T) {
	repo := newMockRegionRepo()
	svc := NewRegionService(repo)

	err := svc.Delete(9999)
	assert.ErrorIs(t, err, ErrRegionNotFound)
}

func TestRegionService_Delete_HasChildren(t *testing.T) {
	repo := newMockRegionRepo()
	parent := model.Region{Name: "parent", Code: "p", ParentID: 0, Level: 1, Sort: 0}
	parent.ID = 1
	child := model.Region{Name: "child", Code: "c", ParentID: 1, Level: 2}
	child.ID = 2
	repo.byID[1] = &parent
	repo.byParent[0] = []model.Region{parent}
	repo.byParent[1] = []model.Region{child}
	svc := NewRegionService(repo)

	err := svc.Delete(1)
	assert.ErrorIs(t, err, ErrRegionHasChildren)
}

func TestRegionService_Delete_Leaf(t *testing.T) {
	repo := newMockRegionRepo()
	leaf := model.Region{Name: "leaf", Code: "l", ParentID: 0, Level: 1, Sort: 0}
	leaf.ID = 1
	repo.byID[1] = &leaf
	repo.byParent[0] = []model.Region{leaf}
	svc := NewRegionService(repo)

	err := svc.Delete(1)
	require.NoError(t, err)
	// 二次删除应返回 NotFound
	err = svc.Delete(1)
	assert.ErrorIs(t, err, ErrRegionNotFound)
}

func TestRegionService_Delete_FindChildrenError(t *testing.T) {
	repo := newMockRegionRepo()
	leaf := model.Region{Name: "leaf", Code: "l", ParentID: 0, Level: 1, Sort: 0}
	leaf.ID = 1
	repo.byID[1] = &leaf
	// findErr 影响 FindByParentID
	repo.findErr = errors.New("db error")
	svc := NewRegionService(repo)

	err := svc.Delete(1)
	assert.EqualError(t, err, "db error")
}

func TestRegionService_Delete_RepoDeleteError(t *testing.T) {
	repo := newMockRegionRepo()
	leaf := model.Region{Name: "leaf", Code: "l", ParentID: 0, Level: 1, Sort: 0}
	leaf.ID = 1
	repo.byID[1] = &leaf
	repo.deleteErr = errors.New("delete failed")
	svc := NewRegionService(repo)

	err := svc.Delete(1)
	assert.EqualError(t, err, "delete failed")
}

// ===== Update =====

func TestRegionService_Update_NotFound(t *testing.T) {
	repo := newMockRegionRepo()
	svc := NewRegionService(repo)

	err := svc.Update(9999, &dto.UpdateRegionRequest{Name: "x"})
	assert.ErrorIs(t, err, ErrRegionNotFound)
}

func TestRegionService_Update_Success(t *testing.T) {
	repo := newMockRegionRepo()
	old := model.Region{Name: "old", Code: "c", Sort: 0, Status: 1}
	old.ID = 1
	repo.byID[1] = &old
	svc := NewRegionService(repo)

	err := svc.Update(1, &dto.UpdateRegionRequest{
		Name:   "new",
		Sort:   5,
		Status: 1,
	})
	require.NoError(t, err)
	// 校验字段更新
	r, _ := repo.FindByID(1)
	assert.Equal(t, "new", r.Name)
	assert.Equal(t, 5, r.Sort)
	assert.Equal(t, 1, r.Status)
}

func TestRegionService_Update_OnlyName(t *testing.T) {
	repo := newMockRegionRepo()
	old := model.Region{Name: "old", Sort: 3, Status: 1}
	old.ID = 1
	repo.byID[1] = &old
	svc := NewRegionService(repo)

	// 仅传 name，sort 始终被更新（req.Sort=0 覆盖原值）
	err := svc.Update(1, &dto.UpdateRegionRequest{Name: "only-name"})
	require.NoError(t, err)
	r, _ := repo.FindByID(1)
	assert.Equal(t, "only-name", r.Name)
	assert.Equal(t, 0, r.Sort, "sort 始终被更新（req.Sort=0 覆盖原值）")
}

func TestRegionService_Update_StatusNotChangedWhenInvalid(t *testing.T) {
	repo := newMockRegionRepo()
	old := model.Region{Name: "old", Sort: 1, Status: 1}
	old.ID = 1
	repo.byID[1] = &old
	svc := NewRegionService(repo)

	// Status 不在 {0,1} 范围（如 2）时不应更新 status 字段
	// 注：UpdateRegionRequest binding 限制了 status 只能是 0 或 1，service 层只判断 0/1
	err := svc.Update(1, &dto.UpdateRegionRequest{Name: "n", Sort: 1, Status: 2})
	require.NoError(t, err)
	r, _ := repo.FindByID(1)
	assert.Equal(t, 1, r.Status, "status 非 0/1 时不应被更新，保留原值")
}

func TestRegionService_Update_RepoError(t *testing.T) {
	repo := newMockRegionRepo()
	old := model.Region{Name: "old", Status: 1}
	old.ID = 1
	repo.byID[1] = &old
	repo.updateErr = errors.New("update failed")
	svc := NewRegionService(repo)

	err := svc.Update(1, &dto.UpdateRegionRequest{Name: "x"})
	assert.EqualError(t, err, "update failed")
}
