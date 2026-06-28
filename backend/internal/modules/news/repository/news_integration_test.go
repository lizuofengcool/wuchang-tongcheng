// Package repository 同城头条仓储的集成测试。
// 使用 testcontainers 启动真实 PostgreSQL，验证 NewsRepository 的
// 地区/状态/分类/关键词过滤、分页、浏览量自增、点赞幂等流程（LikeExists/CreateLike/DeleteLike/IncrLikeCount/DecrLikeCount）、
// FindByIDs 批量查询等。
// 无 Docker 时自动 skip。
package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	newsModel "wuchang-tongcheng/internal/modules/news/model"
	"wuchang-tongcheng/internal/pkg/utils"
	"wuchang-tongcheng/internal/testutil/pgtest"
)

func newNewsRepoForTest(t *testing.T) (*newsRepository, *gorm.DB) {
	t.Helper()
	db := pgtest.SetupPostgres(t)
	pgtest.MigrateAll(t, db)
	return &newsRepository{db: db}, db
}

// makeNews 构造一条头条（status=1 已发布，regionID 指定）
func makeNews(title string, authorID uint, regionID uint, categoryID uint, status int) *newsModel.News {
	n := &newsModel.News{
		Title:       title,
		Content:     "正文 " + title,
		Summary:     title + " 摘要",
		AuthorID:    authorID,
		AuthorName:  "author",
		CategoryID:  categoryID,
		Tags:        "tag1,tag2",
		Status:      status,
	}
	n.RegionID = regionID
	return n
}

// TestNewsRepository_CreateAndFindByID 创建 + 查询
func TestNewsRepository_CreateAndFindByID(t *testing.T) {
	repo, _ := newNewsRepoForTest(t)

	n := makeNews("头条一", 1, 2, 10, 1)
	require.NoError(t, repo.Create(n))
	require.NotZero(t, n.ID)

	got, err := repo.FindByID(n.ID)
	require.NoError(t, err)
	assert.Equal(t, "头条一", got.Title)
	assert.Equal(t, uint(2), got.RegionID)
	assert.Equal(t, uint(10), got.CategoryID)
	assert.Equal(t, 1, got.Status)
}

// TestNewsRepository_List_Filters 地区/状态/分类/关键词四维过滤
func TestNewsRepository_List_Filters(t *testing.T) {
	repo, _ := newNewsRepoForTest(t)

	// 武汉市(2) 3 条已发布 + 1 条草稿；洪山区(5) 2 条已发布
	require.NoError(t, repo.Create(makeNews("武汉新闻A", 1, 2, 10, 1)))
	require.NoError(t, repo.Create(makeNews("武汉新闻B", 1, 2, 11, 1)))
	require.NoError(t, repo.Create(makeNews("武汉招聘", 1, 2, 11, 1)))
	require.NoError(t, repo.Create(makeNews("武汉草稿", 1, 2, 10, 0))) // 草稿
	require.NoError(t, repo.Create(makeNews("洪山新闻", 1, 5, 10, 1)))

	pg := utils.NewPagination(1, 10)

	// 1) 武汉市已发布 → 3 条
	list, total, err := repo.List(2, pg, uint(0), 1, "")
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, list, 3)

	// 2) 武汉市 + 分类 11 已发布 → 2 条
	pg2 := utils.NewPagination(1, 10)
	list, total, err = repo.List(2, pg2, 11, 1, "")
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, list, 2)

	// 3) 武汉市 + 关键词 "新闻"（标题匹配）→ 2 条（A、B）
	pg3 := utils.NewPagination(1, 10)
	list, total, err = repo.List(2, pg3, 0, 1, "新闻")
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, list, 2)

	// 4) 武汉市 + status=2（不在 0..2 之外触发默认 status=1）→ 3 条已发布
	//    注：repo 逻辑 status<0||status>2 时 WHERE status=1
	pg4 := utils.NewPagination(1, 10)
	list, total, err = repo.List(2, pg4, 0, 9, "")
	require.NoError(t, err)
	assert.Equal(t, int64(3), total, "status 越界默认查已发布")
	assert.Len(t, list, 3)

	// 5) 武汉市 + status=0（草稿）→ 1 条
	pg5 := utils.NewPagination(1, 10)
	list, total, err = repo.List(2, pg5, 0, 0, "")
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, list, 1)
	assert.Equal(t, "武汉草稿", list[0].Title)
}

// TestNewsRepository_IncrViewCount 浏览量自增
func TestNewsRepository_IncrViewCount(t *testing.T) {
	repo, _ := newNewsRepoForTest(t)

	n := makeNews("浏览测试", 1, 2, 10, 1)
	require.NoError(t, repo.Create(n))

	require.NoError(t, repo.IncrViewCount(n.ID))
	require.NoError(t, repo.IncrViewCount(n.ID))
	require.NoError(t, repo.IncrViewCount(n.ID))

	got, err := repo.FindByID(n.ID)
	require.NoError(t, err)
	assert.Equal(t, 3, got.ViewCount)
}

// TestNewsRepository_LikeFlow 点赞完整流程：未赞 → 赞 → 已赞 → 取消
func TestNewsRepository_LikeFlow(t *testing.T) {
	repo, _ := newNewsRepoForTest(t)

	n := makeNews("点赞测试", 100, 2, 10, 1)
	require.NoError(t, repo.Create(n))

	// 初始：无人点赞
	exists, err := repo.LikeExists(200, n.ID)
	require.NoError(t, err)
	assert.False(t, exists)

	// 点赞
	require.NoError(t, repo.CreateLike(&newsModel.NewsLike{UserID: 200, NewsID: n.ID}))
	require.NoError(t, repo.IncrLikeCount(n.ID))

	exists, err = repo.LikeExists(200, n.ID)
	require.NoError(t, err)
	assert.True(t, exists)

	got, err := repo.FindByID(n.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, got.LikeCount)

	// 重复点赞触发唯一索引冲突（业务层用 LikeExists 预判）
	err = repo.CreateLike(&newsModel.NewsLike{UserID: 200, NewsID: n.ID})
	assert.Error(t, err, "重复点赞应触发唯一索引")

	// 取消点赞
	require.NoError(t, repo.DeleteLike(200, n.ID))
	require.NoError(t, repo.DecrLikeCount(n.ID))

	exists, err = repo.LikeExists(200, n.ID)
	require.NoError(t, err)
	assert.False(t, exists)

	got, err = repo.FindByID(n.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, got.LikeCount)
}

// TestNewsRepository_DecrLikeCount_Floor 点赞数不会降到负数
// repo 实现 WHERE like_count > 0，所以已为 0 时再 DecrLikeCount 不生效
func TestNewsRepository_DecrLikeCount_Floor(t *testing.T) {
	repo, _ := newNewsRepoForTest(t)

	n := makeNews("地板测试", 1, 2, 10, 1)
	require.NoError(t, repo.Create(n))
	// like_count 初始 0

	// DecrLikeCount 在 0 时不生效（WHERE like_count > 0 命中 0 行）
	require.NoError(t, repo.DecrLikeCount(n.ID))

	got, err := repo.FindByID(n.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, got.LikeCount, "like_count 不应为负")
}

// TestNewsRepository_FindByIDs 批量查询 + 空入参
func TestNewsRepository_FindByIDs(t *testing.T) {
	repo, _ := newNewsRepoForTest(t)

	// 空入参应安全返回 nil
	list, err := repo.FindByIDs(nil)
	require.NoError(t, err)
	assert.Nil(t, list)

	n1 := makeNews("批量1", 1, 2, 10, 1)
	n2 := makeNews("批量2", 1, 2, 10, 1)
	n3 := makeNews("批量3", 1, 2, 10, 1)
	require.NoError(t, repo.Create(n1))
	require.NoError(t, repo.Create(n2))
	require.NoError(t, repo.Create(n3))

	list, err = repo.FindByIDs([]uint{n1.ID, n2.ID, n3.ID, 999999})
	require.NoError(t, err)
	assert.Len(t, list, 3, "不存在的 ID 自动忽略")

	// 验证三条都在
	idSet := map[uint]bool{n1.ID: false, n2.ID: false, n3.ID: false}
	for _, n := range list {
		if _, ok := idSet[n.ID]; ok {
			idSet[n.ID] = true
		}
	}
	for id, found := range idSet {
		assert.True(t, found, "ID %d 应在结果中", id)
	}
}

// TestNewsRepository_Delete_CascadeLikeNotAuto 删除头条不会级联删点赞记录
// （业务约束：删头条前应先删点赞记录，或外键级联由 DB 保证；当前 repo 不做级联）
func TestNewsRepository_Delete_CascadeLikeNotAuto(t *testing.T) {
	repo, db := newNewsRepoForTest(t)

	n := makeNews("级联测试", 1, 2, 10, 1)
	require.NoError(t, repo.Create(n))
	require.NoError(t, repo.CreateLike(&newsModel.NewsLike{UserID: 300, NewsID: n.ID}))

	require.NoError(t, repo.Delete(n.ID))

	// 头条软删除后查不到
	_, err := repo.FindByID(n.ID)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

	// 点赞记录仍存在（无外键级联）
	var likeCount int64
	require.NoError(t, db.Model(&newsModel.NewsLike{}).Where("news_id = ?", n.ID).Count(&likeCount).Error)
	assert.Equal(t, int64(1), likeCount, "repo 未做级联删除，点赞记录仍在")
}
