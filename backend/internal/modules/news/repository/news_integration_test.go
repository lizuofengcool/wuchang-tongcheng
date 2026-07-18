// Package repository 同城头条仓储的集成测试。
// 使用 testcontainers 启动真实 PostgreSQL，验证 NewsRepository 的
// 地区/状态/分类/关键词过滤、分页、浏览量自增、点赞幂等流程（LikeExists/CreateLike/DeleteLike/IncrLikeCount/DecrLikeCount）、
// FindByIDs 批量查询等。
// 无 Docker 时自动 skip。
package repository

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	newsModel "wuchang-tongcheng/internal/modules/news/model"
	"wuchang-tongcheng/internal/pkg/geo"
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

// makeNewsListing 构造一条带分类信息字段（listingType/price/isUrgent）的已发布头条，
// 用于覆盖 List 仓储层新增的 listingType / 价格区间 / 加急 过滤参数。
func makeNewsListing(title string, regionID, categoryID uint, listingType string, price float64, isUrgent bool) *newsModel.News {
	n := &newsModel.News{
		Title:       title,
		Content:     "正文 " + title,
		Summary:     title + " 摘要",
		AuthorID:    1,
		AuthorName:  "author",
		CategoryID:  categoryID,
		Tags:        "tag1,tag2",
		Status:      1,
		ListingType: listingType,
		Price:       price,
		PriceUnit:   "元",
		IsUrgent:    isUrgent,
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
	//    List 新签名：(regionID, pagination, categoryID, status, listingType, keyword, minPrice, maxPrice, isUrgent, sort)
	list, total, err := repo.List(2, pg, uint(0), 1, "", "", 0, 0, nil, "")
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, list, 3)

	// 2) 武汉市 + 分类 11 已发布 → 2 条
	pg2 := utils.NewPagination(1, 10)
	list, total, err = repo.List(2, pg2, 11, 1, "", "", 0, 0, nil, "")
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, list, 2)

	// 3) 武汉市 + 关键词 "新闻"（标题匹配）→ 2 条（A、B）
	pg3 := utils.NewPagination(1, 10)
	list, total, err = repo.List(2, pg3, 0, 1, "", "新闻", 0, 0, nil, "")
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, list, 2)

	// 4) 武汉市 + status=2（不在 0..2 之外触发默认 status=1）→ 3 条已发布
	//    注：repo 逻辑 status<0||status>2 时 WHERE status=1
	pg4 := utils.NewPagination(1, 10)
	list, total, err = repo.List(2, pg4, 0, 9, "", "", 0, 0, nil, "")
	require.NoError(t, err)
	assert.Equal(t, int64(3), total, "status 越界默认查已发布")
	assert.Len(t, list, 3)

	// 5) 武汉市 + status=0（草稿）→ 1 条
	pg5 := utils.NewPagination(1, 10)
	list, total, err = repo.List(2, pg5, 0, 0, "", "", 0, 0, nil, "")
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, list, 1)
	assert.Equal(t, "武汉草稿", list[0].Title)
}

// TestNewsRepository_List_ListingFilters 覆盖 List 仓储层新增的分类信息过滤参数：
// listingType（出售/求购/出租/服务/招聘）、minPrice/maxPrice 价格区间、isUrgent 加急。
// 这一组参数是「同城分类信息平台」升级时新增到 List 签名的，需要专门的集成测试覆盖。
func TestNewsRepository_List_ListingFilters(t *testing.T) {
	repo, _ := newNewsRepoForTest(t)

	// 武汉市(2) 全部已发布，覆盖 5 种 listingType + 不同价格 + 是否加急
	require.NoError(t, repo.Create(makeNewsListing("出售-笔记本", 2, 10, newsModel.ListingTypeSell, 3000, true)))
	require.NoError(t, repo.Create(makeNewsListing("出售-自行车", 2, 10, newsModel.ListingTypeSell, 200, false)))
	require.NoError(t, repo.Create(makeNewsListing("求购-笔记本", 2, 10, newsModel.ListingTypeBuy, 2500, false)))
	require.NoError(t, repo.Create(makeNewsListing("出租-单间", 2, 10, newsModel.ListingTypeRent, 1500, true)))
	require.NoError(t, repo.Create(makeNewsListing("服务-搬家", 2, 10, newsModel.ListingTypeService, 500, false)))

	// 1) listingType=sell → 2 条（笔记本、自行车）
	pg := utils.NewPagination(1, 10)
	list, total, err := repo.List(2, pg, 0, 1, newsModel.ListingTypeSell, "", 0, 0, nil, "")
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, list, 2)

	// 2) listingType=rent → 1 条（单间）
	pg2 := utils.NewPagination(1, 10)
	list, total, err = repo.List(2, pg2, 0, 1, newsModel.ListingTypeRent, "", 0, 0, nil, "")
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, list, 1)
	assert.Equal(t, "出租-单间", list[0].Title)

	// 3) 价格区间 [1000, 3000] → 3 条（笔记本 3000、求购笔记本 2500、单间 1500）
	pg3 := utils.NewPagination(1, 10)
	list, total, err = repo.List(2, pg3, 0, 1, "", "", 1000, 3000, nil, "")
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, list, 3)

	// 4) 仅 minPrice=1000 → 4 条（≥1000：笔记本、求购笔记本、单间、搬家 500 不算 → 3 条）
	//    注：搬家 500 < 1000 应被排除
	pg4 := utils.NewPagination(1, 10)
	list, total, err = repo.List(2, pg4, 0, 1, "", "", 1000, 0, nil, "")
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, list, 3)

	// 5) isUrgent=true → 2 条（笔记本、单间）
	urgent := true
	pg5 := utils.NewPagination(1, 10)
	list, total, err = repo.List(2, pg5, 0, 1, "", "", 0, 0, &urgent, "")
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, list, 2)

	// 6) 组合：listingType=sell + isUrgent=true → 1 条（笔记本）
	pg6 := utils.NewPagination(1, 10)
	list, total, err = repo.List(2, pg6, 0, 1, newsModel.ListingTypeSell, "", 0, 0, &urgent, "")
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, list, 1)
	assert.Equal(t, "出售-笔记本", list[0].Title)

	// 7) isUrgent=false（*bool 为 nil 表示不过滤，传 false 指针当前实现仅 *isUrgent==true 才生效，
	//    所以传 false 与 nil 等价 → 5 条全部返回）
	notUrgent := false
	pg7 := utils.NewPagination(1, 10)
	list, total, err = repo.List(2, pg7, 0, 1, "", "", 0, 0, &notUrgent, "")
	require.NoError(t, err)
	assert.Equal(t, int64(5), total, "isUrgent=false 当前实现等价于不过滤")
	assert.Len(t, list, 5)

	// 8) sort=price → 按价格升序，第一条应是「出售-自行车」(200)
	pg8 := utils.NewPagination(1, 10)
	list, total, err = repo.List(2, pg8, 0, 1, "", "", 0, 0, nil, "price")
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
	require.Len(t, list, 5)
	assert.Equal(t, "出售-自行车", list[0].Title, "sort=price 应按价格升序")
	assert.Equal(t, float64(200), list[0].Price)

	// 9) sort=price_desc → 按价格降序，第一条应是「出售-笔记本」(3000)
	pg9 := utils.NewPagination(1, 10)
	list, total, err = repo.List(2, pg9, 0, 1, "", "", 0, 0, nil, "price_desc")
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
	require.Len(t, list, 5)
	assert.Equal(t, "出售-笔记本", list[0].Title, "sort=price_desc 应按价格降序")
	assert.Equal(t, float64(3000), list[0].Price)
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

// makeNewsWithLocation 构造一条带经纬度的已发布头条，用于 ListNearby 附近查询测试。
func makeNewsWithLocation(title string, regionID, categoryID uint, lat, lng float64) *newsModel.News {
	n := &newsModel.News{
		Title:       title,
		Content:     "正文 " + title,
		Summary:     title + " 摘要",
		AuthorID:    1,
		AuthorName:  "author",
		CategoryID:  categoryID,
		Status:      1,
		ListingType: newsModel.ListingTypeSell,
		Latitude:    lat,
		Longitude:   lng,
		Address:     title + " 地址",
	}
	n.RegionID = regionID
	return n
}

// TestNewsRepository_ListNearby 附近信息查询集成测试。
// pgtest 容器为 postgres:16-alpine（未装 PostGIS），自动走 Haversine 降级路径，
// 验证：半径过滤、距离升序排序、Distance 字段回填、草稿/无坐标排除、分类过滤、半径上限钳制。
func TestNewsRepository_ListNearby(t *testing.T) {
	geo.ResetForTest()
	repo, _ := newNewsRepoForTest(t)

	// 以五常 (44.9225, 127.1500) 为中心
	const centerLat, centerLng = 44.9225, 127.1500

	// A：正中心 → 距离 0
	require.NoError(t, repo.Create(makeNewsWithLocation("中心点", 2, 10, centerLat, centerLng)))
	// B：正北 ~30 km（纬度 +0.27）→ 半径 50 km 内
	require.NoError(t, repo.Create(makeNewsWithLocation("北30km", 2, 10, centerLat+0.27, centerLng)))
	// C：正东 ~80 km（经度增量按 cos(lat) 折算）→ 半径 50 km 外
	cosLat := math.Cos(centerLat * math.Pi / 180.0)
	lngDeltaC := 80.0 / (111.0 * cosLat)
	require.NoError(t, repo.Create(makeNewsWithLocation("东80km", 2, 10, centerLat, centerLng+lngDeltaC)))
	// D：中心点但草稿 → 不应返回
	draft := makeNewsWithLocation("草稿", 2, 10, centerLat, centerLng)
	draft.Status = 0
	require.NoError(t, repo.Create(draft))
	// E：中心点但无坐标（lat/lng=0）→ 不应返回
	require.NoError(t, repo.Create(makeNewsWithLocation("无坐标", 2, 10, 0, 0)))

	// 1) 半径 50 km → 仅 A、B，按距离升序（A 在前）
	pg := utils.NewPagination(1, 10)
	list, total, err := repo.ListNearby(2, pg, centerLat, centerLng, 50, 0, "")
	require.NoError(t, err)
	require.Equal(t, int64(2), total, "50km 内应有 2 条（中心点 + 北30km）")
	require.Len(t, list, 2)
	assert.Equal(t, "中心点", list[0].Title, "中心点距离最近应排第一")
	assert.Equal(t, "北30km", list[1].Title)
	// Distance 字段应被回填
	assert.LessOrEqual(t, list[0].Distance, 1.0, "中心点距离应 < 1 km")
	assert.InDelta(t, 30.0, list[1].Distance, 5.0, "北30km 距离应约 30 km")

	// 2) 半径 100 km → A、B、C 三条
	pg2 := utils.NewPagination(1, 10)
	list, total, err = repo.ListNearby(2, pg2, centerLat, centerLng, 100, 0, "")
	require.NoError(t, err)
	require.Equal(t, int64(3), total, "100km 内应有 3 条")
	require.Len(t, list, 3)
	assert.InDelta(t, 80.0, list[2].Distance, 5.0, "东80km 距离应约 80 km，排第三")

	// 3) 半径 10 km → 仅 A
	pg3 := utils.NewPagination(1, 10)
	list, total, err = repo.ListNearby(2, pg3, centerLat, centerLng, 10, 0, "")
	require.NoError(t, err)
	require.Equal(t, int64(1), total, "10km 内应仅中心点 1 条")
	require.Len(t, list, 1)
	assert.Equal(t, "中心点", list[0].Title)

	// 4) 越界半径（>100）应钳制为 100，与 radius=100 结果一致
	pg4 := utils.NewPagination(1, 10)
	_, total4, err := repo.ListNearby(2, pg4, centerLat, centerLng, 9999, 0, "")
	require.NoError(t, err)
	assert.Equal(t, total, total4, "半径 9999 应钳制为 100，结果数一致")

	// 5) 半径 50 km + 分类过滤（构造一条 50km 内但分类不同的，应被排除）
	//    重新用 30km 北点位置但 categoryID=99，半径 50km 内但按 category=10 过滤应不含它
	//    （上面 B 用 category=10，这里再加一个 category=99 的近点）
	require.NoError(t, repo.Create(makeNewsWithLocation("北30km-分类99", 2, 99, centerLat+0.27, centerLng)))
	pg5 := utils.NewPagination(1, 10)
	list, total, err = repo.ListNearby(2, pg5, centerLat, centerLng, 50, 10, "")
	require.NoError(t, err)
	require.Equal(t, int64(2), total, "半径50km+分类10 → 仍为 A、B 两条（分类99 的被排除）")
	require.Len(t, list, 2)
}

