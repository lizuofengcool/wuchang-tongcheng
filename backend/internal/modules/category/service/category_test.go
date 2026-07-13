// Package service 分类服务单元测试。
// 使用内存 mock repository，覆盖缓存键生成、DTO 转换、层级计算、
// 最大层级限制、父分类校验、树形构建等核心业务逻辑，不依赖 DB/Redis。
package service

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"wuchang-tongcheng/internal/modules/category/dto"
	"wuchang-tongcheng/internal/modules/category/model"
)

// mockCategoryRepo 内存 mock，实现 CategoryRepository 接口
type mockCategoryRepo struct {
	byID       map[uint]*model.Category
	byParent   map[uint][]model.Category // parentID -> children
	byRegion   []model.Category
	nextID     uint
	createErr  error
	findErr    error
	deleteErr  error
	updateErr  error
}

func newMockCategoryRepo() *mockCategoryRepo {
	return &mockCategoryRepo{
		byID:     make(map[uint]*model.Category),
		byParent: make(map[uint][]model.Category),
		nextID:   1,
	}
}

func (m *mockCategoryRepo) Create(category *model.Category) error {
	if m.createErr != nil {
		return m.createErr
	}
	category.ID = m.nextID
	m.nextID++
	cp := *category
	m.byID[category.ID] = &cp
	m.byParent[category.ParentID] = append(m.byParent[category.ParentID], cp)
	m.byRegion = append(m.byRegion, cp)
	return nil
}

func (m *mockCategoryRepo) FindByID(id uint) (*model.Category, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	c, ok := m.byID[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *c
	return &cp, nil
}

func (m *mockCategoryRepo) FindByParentID(parentID uint, regionID uint) ([]model.Category, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	out := make([]model.Category, 0, len(m.byParent[parentID]))
	out = append(out, m.byParent[parentID]...)
	return out, nil
}

func (m *mockCategoryRepo) FindByRegionID(regionID uint) ([]model.Category, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	out := make([]model.Category, 0, len(m.byRegion))
	out = append(out, m.byRegion...)
	return out, nil
}

func (m *mockCategoryRepo) Update(category *model.Category) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if _, ok := m.byID[category.ID]; !ok {
		return gorm.ErrRecordNotFound
	}
	cp := *category
	m.byID[category.ID] = &cp
	return nil
}

func (m *mockCategoryRepo) UpdateFields(id uint, fields map[string]interface{}) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	c, ok := m.byID[id]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	if v, ok := fields["name"]; ok {
		c.Name = v.(string)
	}
	if v, ok := fields["icon"]; ok {
		c.Icon = v.(string)
	}
	if v, ok := fields["sort"]; ok {
		c.Sort = v.(int)
	}
	if v, ok := fields["status"]; ok {
		c.Status = v.(int)
	}
	return nil
}

func (m *mockCategoryRepo) Delete(id uint) error {
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

func TestCategoryCacheKeyTree(t *testing.T) {
	assert.Equal(t, "cache:category:tree:1", categoryCacheKeyTree(1))
	assert.Equal(t, "cache:category:tree:999", categoryCacheKeyTree(999))
	assert.Equal(t, "cache:category:tree:0", categoryCacheKeyTree(0))
}

func TestCategoryCacheKeyByID(t *testing.T) {
	assert.Equal(t, "cache:category:id:1", categoryCacheKeyByID(1))
	assert.Equal(t, "cache:category:id:0", categoryCacheKeyByID(0))
	assert.Equal(t, "cache:category:id:8888", categoryCacheKeyByID(8888))
}

func TestCategoryCacheKeyByParent(t *testing.T) {
	assert.Equal(t, "cache:category:parent:1:2", categoryCacheKeyByParent(1, 2))
	assert.Equal(t, "cache:category:parent:0:1", categoryCacheKeyByParent(0, 1))
	assert.Equal(t, "cache:category:parent:100:200", categoryCacheKeyByParent(100, 200))
}

func TestToCategoryInfo(t *testing.T) {
	c := &model.Category{
		Name:     "二手房",
		Icon:     "icon-house",
		ParentID: 2,
		Level:    2,
		Sort:     10,
		Status:   1,
	}
	c.ID = 5
	c.RegionID = 3

	info := toCategoryInfo(c)
	assert.Equal(t, uint(5), info.ID)
	assert.Equal(t, "二手房", info.Name)
	assert.Equal(t, "icon-house", info.Icon)
	assert.Equal(t, uint(2), info.ParentID)
	assert.Equal(t, 2, info.Level)
	assert.Equal(t, 10, info.Sort)
	assert.Equal(t, 1, info.Status)
}

func TestToCategoryInfo_NilSafe(t *testing.T) {
	// 即使 RegionID 等嵌入字段未设置也不应 panic
	c := &model.Category{Name: "空分类"}
	info := toCategoryInfo(c)
	assert.Equal(t, "空分类", info.Name)
	assert.Equal(t, uint(0), info.ID)
	assert.Equal(t, 0, info.Level)
	assert.Equal(t, 0, info.Status)
}

// ===== 构造函数 =====

func TestNewCategoryService(t *testing.T) {
	repo := newMockCategoryRepo()
	svc := NewCategoryService(repo)
	assert.NotNil(t, svc)
	// 应返回可用的 CategoryService 接口实例
	_, ok := svc.(*categoryService)
	assert.True(t, ok, "返回值应为 *categoryService 类型")
}

// ===== Create =====

func TestCategoryService_Create_TopLevel(t *testing.T) {
	repo := newMockCategoryRepo()
	svc := NewCategoryService(repo)

	info, err := svc.Create(1, &dto.CreateCategoryRequest{
		Name:   "招聘",
		Icon:   "job",
		Sort:   5,
		Status: 1,
	})
	require.NoError(t, err)
	assert.Equal(t, "招聘", info.Name)
	assert.Equal(t, "job", info.Icon)
	assert.Equal(t, 1, info.Level, "ParentID=0 应为顶级，level=1")
	assert.Equal(t, uint(0), info.ParentID)
	assert.Equal(t, 5, info.Sort)
	assert.Equal(t, 1, info.Status)
}

func TestCategoryService_Create_DefaultStatusToOne(t *testing.T) {
	repo := newMockCategoryRepo()
	svc := NewCategoryService(repo)

	// Status=0 时应默认填充为 1
	info, err := svc.Create(1, &dto.CreateCategoryRequest{
		Name:   "二手房",
		Status: 0,
	})
	require.NoError(t, err)
	assert.Equal(t, 1, info.Status, "Status=0 应默认填充为 1")
}

func TestCategoryService_Create_ChildLevel(t *testing.T) {
	repo := newMockCategoryRepo()
	svc := NewCategoryService(repo)

	// 先建顶级
	parent, err := svc.Create(1, &dto.CreateCategoryRequest{Name: "招聘", Status: 1})
	require.NoError(t, err)
	require.Equal(t, 1, parent.Level)

	// 再建子级，level 应为父级+1=2
	child, err := svc.Create(1, &dto.CreateCategoryRequest{
		Name:     "IT招聘",
		ParentID: parent.ID,
		Status:   1,
	})
	require.NoError(t, err)
	assert.Equal(t, 2, child.Level, "子级 level 应为父级 level+1")
	assert.Equal(t, parent.ID, child.ParentID)
}

func TestCategoryService_Create_GrandchildLevel(t *testing.T) {
	repo := newMockCategoryRepo()
	svc := NewCategoryService(repo)

	root, _ := svc.Create(1, &dto.CreateCategoryRequest{Name: "招聘", Status: 1})
	child, _ := svc.Create(1, &dto.CreateCategoryRequest{Name: "IT招聘", ParentID: root.ID, Status: 1})
	require.Equal(t, 2, child.Level)

	// 第三层（最大允许层级）
	grandchild, err := svc.Create(1, &dto.CreateCategoryRequest{
		Name:     "Java工程师",
		ParentID: child.ID,
		Status:   1,
	})
	require.NoError(t, err)
	assert.Equal(t, 3, grandchild.Level)
}

func TestCategoryService_Create_ExceedMaxLevel(t *testing.T) {
	repo := newMockCategoryRepo()
	svc := NewCategoryService(repo)

	root, _ := svc.Create(1, &dto.CreateCategoryRequest{Name: "招聘", Status: 1})
	child, _ := svc.Create(1, &dto.CreateCategoryRequest{Name: "IT招聘", ParentID: root.ID, Status: 1})
	grandchild, _ := svc.Create(1, &dto.CreateCategoryRequest{Name: "Java", ParentID: child.ID, Status: 1})
	require.Equal(t, 3, grandchild.Level)

	// 第四层应被拒绝（MaxCategoryLevel=3）
	_, err := svc.Create(1, &dto.CreateCategoryRequest{
		Name:     "Spring Boot",
		ParentID: grandchild.ID,
		Status:   1,
	})
	assert.ErrorIs(t, err, ErrCategoryMaxLevel)
}

func TestCategoryService_Create_ParentNotFound(t *testing.T) {
	repo := newMockCategoryRepo()
	svc := NewCategoryService(repo)

	_, err := svc.Create(1, &dto.CreateCategoryRequest{
		Name:     "孤儿分类",
		ParentID: 9999,
		Status:   1,
	})
	assert.ErrorIs(t, err, ErrCategoryParentInvalid)
}

func TestCategoryService_Create_RepoError(t *testing.T) {
	repo := newMockCategoryRepo()
	boom := errors.New("db down")
	repo.createErr = boom
	svc := NewCategoryService(repo)

	_, err := svc.Create(1, &dto.CreateCategoryRequest{Name: "x", Status: 1})
	assert.ErrorIs(t, err, boom)
}

func TestCategoryService_Create_ParentFindError(t *testing.T) {
	repo := newMockCategoryRepo()
	repo.findErr = errors.New("connection refused")
	svc := NewCategoryService(repo)

	// 父分类查询失败（非 ErrRecordNotFound）应原样透传错误
	_, err := svc.Create(1, &dto.CreateCategoryRequest{
		Name:     "x",
		ParentID: 5,
		Status:   1,
	})
	assert.EqualError(t, err, "connection refused")
}

// ===== GetTree =====

func TestCategoryService_GetTree_Empty(t *testing.T) {
	repo := newMockCategoryRepo()
	svc := NewCategoryService(repo)

	tree, err := svc.GetTree(1)
	require.NoError(t, err)
	assert.Empty(t, tree)
}

func TestCategoryService_GetTree_MultiLevel(t *testing.T) {
	repo := newMockCategoryRepo()
	// 直接塞入 mock 数据：3 层树
	//   root1 (id=1, parent=0, level=1)
	//     child1 (id=2, parent=1, level=2)
	//       grandchild1 (id=3, parent=2, level=3)
	//   root2 (id=4, parent=0, level=1)
	// ID 在嵌入的 BaseModel 中，构造后单独赋值
	cats := []model.Category{
		{Name: "root1", ParentID: 0, Level: 1, Sort: 10},
		{Name: "child1", ParentID: 1, Level: 2, Sort: 5},
		{Name: "grandchild1", ParentID: 2, Level: 3, Sort: 0},
		{Name: "root2", ParentID: 0, Level: 1, Sort: 20},
	}
	for i := range cats {
		cats[i].ID = uint(i + 1)
	}
	repo.byRegion = cats
	svc := NewCategoryService(repo)

	tree, err := svc.GetTree(1)
	require.NoError(t, err)
	require.Len(t, tree, 2, "根节点应有 2 个")

	// root2 排在前面（Sort 20 > 10，但 mock 不排序，按塞入顺序）
	assert.Equal(t, "root1", tree[0].Name)
	assert.Equal(t, "root2", tree[1].Name)

	require.Len(t, tree[0].Children, 1)
	assert.Equal(t, "child1", tree[0].Children[0].Name)

	require.Len(t, tree[0].Children[0].Children, 1)
	assert.Equal(t, "grandchild1", tree[0].Children[0].Children[0].Name)

	// root2 无子节点
	assert.Empty(t, tree[1].Children)
}

// ===== GetAll =====

func TestCategoryService_GetAll_Empty(t *testing.T) {
	repo := newMockCategoryRepo()
	svc := NewCategoryService(repo)

	list, err := svc.GetAll(1)
	require.NoError(t, err)
	assert.Empty(t, list)
}

func TestCategoryService_GetAll_FlatList(t *testing.T) {
	repo := newMockCategoryRepo()
	cats := []model.Category{
		{Name: "a", ParentID: 0, Level: 1, Sort: 1, Status: 1},
		{Name: "b", ParentID: 0, Level: 1, Sort: 2, Status: 1},
		{Name: "c", ParentID: 1, Level: 2, Sort: 0, Status: 1},
	}
	for i := range cats {
		cats[i].ID = uint(i + 1)
	}
	repo.byRegion = cats
	svc := NewCategoryService(repo)

	list, err := svc.GetAll(1)
	require.NoError(t, err)
	require.Len(t, list, 3)
	assert.Equal(t, "a", list[0].Name)
	assert.Equal(t, "b", list[1].Name)
	assert.Equal(t, "c", list[2].Name)
	// 平铺列表应包含层级信息
	assert.Equal(t, 1, list[0].Level)
	assert.Equal(t, 2, list[2].Level)
}

func TestCategoryService_GetAll_RepoError(t *testing.T) {
	repo := newMockCategoryRepo()
	repo.findErr = errors.New("db error")
	svc := NewCategoryService(repo)

	_, err := svc.GetAll(1)
	assert.EqualError(t, err, "db error")
}

// ===== Delete =====

func TestCategoryService_Delete_NotFound(t *testing.T) {
	repo := newMockCategoryRepo()
	svc := NewCategoryService(repo)

	err := svc.Delete(9999)
	assert.ErrorIs(t, err, ErrCategoryNotFound)
}

func TestCategoryService_Delete_HasChildren(t *testing.T) {
	repo := newMockCategoryRepo()
	parent := model.Category{Name: "parent", ParentID: 0, Level: 1, Sort: 0}
	parent.ID = 1
	child := model.Category{Name: "child", ParentID: 1, Level: 2}
	child.ID = 2
	repo.byID[1] = &parent
	repo.byParent[0] = []model.Category{parent}
	repo.byParent[1] = []model.Category{child}
	svc := NewCategoryService(repo)

	err := svc.Delete(1)
	assert.ErrorIs(t, err, ErrCategoryHasChildren)
}

func TestCategoryService_Delete_Leaf(t *testing.T) {
	repo := newMockCategoryRepo()
	leaf := model.Category{Name: "leaf", ParentID: 0, Level: 1, Sort: 0}
	leaf.ID = 1
	repo.byID[1] = &leaf
	repo.byParent[0] = []model.Category{leaf}
	svc := NewCategoryService(repo)

	err := svc.Delete(1)
	require.NoError(t, err)
	// 二次删除应返回 NotFound
	err = svc.Delete(1)
	assert.ErrorIs(t, err, ErrCategoryNotFound)
}

// ===== Update =====

func TestCategoryService_Update_NotFound(t *testing.T) {
	repo := newMockCategoryRepo()
	svc := NewCategoryService(repo)

	err := svc.Update(9999, &dto.UpdateCategoryRequest{Name: "x"})
	assert.ErrorIs(t, err, ErrCategoryNotFound)
}

func TestCategoryService_Update_Success(t *testing.T) {
	repo := newMockCategoryRepo()
	old := model.Category{Name: "old", Icon: "old-icon", Sort: 0, Status: 1}
	old.ID = 1
	repo.byID[1] = &old
	svc := NewCategoryService(repo)

	err := svc.Update(1, &dto.UpdateCategoryRequest{
		Name:   "new",
		Icon:   "new-icon",
		Sort:   5,
		Status: 1,
	})
	require.NoError(t, err)
	// 校验字段更新
	c, _ := repo.FindByID(1)
	assert.Equal(t, "new", c.Name)
	assert.Equal(t, "new-icon", c.Icon)
	assert.Equal(t, 5, c.Sort)
	assert.Equal(t, 1, c.Status)
}

func TestCategoryService_Update_OnlyName(t *testing.T) {
	repo := newMockCategoryRepo()
	old := model.Category{Name: "old", Icon: "keep", Sort: 3, Status: 1}
	old.ID = 1
	repo.byID[1] = &old
	svc := NewCategoryService(repo)

	// 仅传 name，其它字段不传（空字符串字段不覆盖原值，sort 始终被更新为零值）
	err := svc.Update(1, &dto.UpdateCategoryRequest{Name: "only-name"})
	require.NoError(t, err)
	c, _ := repo.FindByID(1)
	assert.Equal(t, "only-name", c.Name)
	assert.Equal(t, "keep", c.Icon, "空字符串字段不应覆盖原值")
	assert.Equal(t, 0, c.Sort, "sort 始终被更新（req.Sort=0 覆盖原值）")
}
