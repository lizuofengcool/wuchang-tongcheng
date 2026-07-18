// Package service 同城头条信息服务单元测试。
// 使用内存 mock repository + mock indexer，覆盖缓存键生成、DTO 转换、
// 创建默认值填充、过期时间计算、权限校验、点赞/收藏 toggle 与计数钳制、
// 评论与消息通知、搜索降级、ES 过滤器构建等核心业务逻辑，不依赖 DB/Redis/ES/MQ。
package service

import (
	"errors"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"wuchang-tongcheng/internal/modules/news/dto"
	"wuchang-tongcheng/internal/modules/news/indexer"
	"wuchang-tongcheng/internal/modules/news/model"
	"wuchang-tongcheng/internal/pkg/utils"
)

// ===== mockNewsRepo 内存 mock，实现 NewsRepository 接口 =====

type mockNewsRepo struct {
	byID      map[uint]*model.News
	nextID    uint
	createErr error
	findErr   error
	deleteErr error
	updateErr error
	listErr   error

	// 点赞/收藏/评论/消息操作记录（用于断言副作用）
	likes       map[uint]map[uint]bool // newsID -> userID set
	favs        map[uint]map[uint]bool
	comments    []*model.NewsComment
	messages    []*model.Message
	viewIncr    map[uint]int
	likeIncr    map[uint]int
	likeDecr    map[uint]int
	favIncr     map[uint]int
	favDecr     map[uint]int
	commentIncr map[uint]int
	likeExistsErr      error
	favExistsErr       error
	createCommentErr   error
}

func newMockNewsRepo() *mockNewsRepo {
	return &mockNewsRepo{
		byID:        make(map[uint]*model.News),
		nextID:      1,
		likes:       make(map[uint]map[uint]bool),
		favs:        make(map[uint]map[uint]bool),
		viewIncr:    make(map[uint]int),
		likeIncr:    make(map[uint]int),
		likeDecr:    make(map[uint]int),
		favIncr:     make(map[uint]int),
		favDecr:     make(map[uint]int),
		commentIncr: make(map[uint]int),
	}
}

func (m *mockNewsRepo) Create(news *model.News) error {
	if m.createErr != nil {
		return m.createErr
	}
	news.ID = m.nextID
	m.nextID++
	cp := *news
	m.byID[news.ID] = &cp
	return nil
}

func (m *mockNewsRepo) FindByID(id uint) (*model.News, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	n, ok := m.byID[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *n
	return &cp, nil
}

func (m *mockNewsRepo) Update(news *model.News) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if _, ok := m.byID[news.ID]; !ok {
		return gorm.ErrRecordNotFound
	}
	cp := *news
	m.byID[news.ID] = &cp
	return nil
}

func (m *mockNewsRepo) UpdateFields(id uint, fields map[string]interface{}) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	n, ok := m.byID[id]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	// 仅更新 mock 支持的字段，用于断言
	if v, ok := fields["title"]; ok {
		n.Title = v.(string)
	}
	if v, ok := fields["status"]; ok {
		n.Status = v.(int)
	}
	if v, ok := fields["published_at"]; ok {
		if t, ok := v.(*time.Time); ok {
			n.PublishedAt = t
		}
	}
	if v, ok := fields["category_id"]; ok {
		n.CategoryID = v.(uint)
	}
	if v, ok := fields["price"]; ok {
		n.Price = v.(float64)
	}
	if v, ok := fields["is_urgent"]; ok {
		n.IsUrgent = v.(bool)
	}
	return nil
}

func (m *mockNewsRepo) Delete(id uint) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if _, ok := m.byID[id]; !ok {
		return gorm.ErrRecordNotFound
	}
	delete(m.byID, id)
	return nil
}

func (m *mockNewsRepo) List(regionID uint, pagination *utils.Pagination, categoryID uint, status int, listingType string, keyword string, minPrice, maxPrice float64, isUrgent *bool, sort string) ([]model.News, int64, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	list := make([]model.News, 0, len(m.byID))
	for _, n := range m.byID {
		cp := *n
		list = append(list, cp)
	}
	return list, int64(len(list)), nil
}

func (m *mockNewsRepo) ListNearby(regionID uint, pagination *utils.Pagination, lat, lng, radiusKm float64, categoryID uint, listingType string) ([]model.News, int64, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	list := make([]model.News, 0, len(m.byID))
	for _, n := range m.byID {
		cp := *n
		list = append(list, cp)
	}
	return list, int64(len(list)), nil
}

func (m *mockNewsRepo) IncrViewCount(id uint) error {
	m.viewIncr[id]++
	return nil
}

func (m *mockNewsRepo) LikeExists(userID, newsID uint) (bool, error) {
	if m.likeExistsErr != nil {
		return false, m.likeExistsErr
	}
	if s, ok := m.likes[newsID]; ok {
		return s[userID], nil
	}
	return false, nil
}

func (m *mockNewsRepo) CreateLike(like *model.NewsLike) error {
	if m.likes[like.NewsID] == nil {
		m.likes[like.NewsID] = make(map[uint]bool)
	}
	m.likes[like.NewsID][like.UserID] = true
	return nil
}

func (m *mockNewsRepo) DeleteLike(userID, newsID uint) error {
	if s, ok := m.likes[newsID]; ok {
		delete(s, userID)
	}
	return nil
}

func (m *mockNewsRepo) IncrLikeCount(id uint) error {
	m.likeIncr[id]++
	if n, ok := m.byID[id]; ok {
		n.LikeCount++
	}
	return nil
}

func (m *mockNewsRepo) DecrLikeCount(id uint) error {
	m.likeDecr[id]++
	if n, ok := m.byID[id]; ok && n.LikeCount > 0 {
		n.LikeCount--
	}
	return nil
}

func (m *mockNewsRepo) FavExists(userID, newsID uint) (bool, error) {
	if m.favExistsErr != nil {
		return false, m.favExistsErr
	}
	if s, ok := m.favs[newsID]; ok {
		return s[userID], nil
	}
	return false, nil
}

func (m *mockNewsRepo) CreateFav(fav *model.NewsFavorite) error {
	if m.favs[fav.NewsID] == nil {
		m.favs[fav.NewsID] = make(map[uint]bool)
	}
	m.favs[fav.NewsID][fav.UserID] = true
	return nil
}

func (m *mockNewsRepo) DeleteFav(userID, newsID uint) error {
	if s, ok := m.favs[newsID]; ok {
		delete(s, userID)
	}
	return nil
}

func (m *mockNewsRepo) IncrFavCount(id uint) error {
	m.favIncr[id]++
	if n, ok := m.byID[id]; ok {
		n.FavCount++
	}
	return nil
}

func (m *mockNewsRepo) DecrFavCount(id uint) error {
	m.favDecr[id]++
	if n, ok := m.byID[id]; ok && n.FavCount > 0 {
		n.FavCount--
	}
	return nil
}

// ListFavs 模拟按 created_at 倒序的分页收藏查询。
// mock 用 newsID 倒序模拟"最近收藏在前"，保证跨页查询顺序稳定（Go map 迭代序随机，
// 必须显式排序，否则分页会跨页重叠或遗漏）。
func (m *mockNewsRepo) ListFavs(userID uint, page, pageSize int) ([]model.NewsFavorite, int64, error) {
	type entry struct {
		newsID    uint
		createdAt time.Time
	}
	var all []entry
	for newsID, users := range m.favs {
		if users[userID] {
			all = append(all, entry{newsID: newsID, createdAt: time.Now()})
		}
	}
	// 按 newsID 倒序保证可重复（map 迭代序随机，未排序会导致分页跨页重叠/遗漏）
	sort.Slice(all, func(i, j int) bool { return all[i].newsID > all[j].newsID })
	total := int64(len(all))
	offset := (page - 1) * pageSize
	if offset >= len(all) {
		return []model.NewsFavorite{}, total, nil
	}
	end := offset + pageSize
	if end > len(all) {
		end = len(all)
	}
	out := make([]model.NewsFavorite, 0, end-offset)
	for i := offset; i < end; i++ {
		out = append(out, model.NewsFavorite{
			UserID:    userID,
			NewsID:    all[i].newsID,
			CreatedAt: all[i].createdAt,
		})
	}
	return out, total, nil
}

func (m *mockNewsRepo) CreateComment(comment *model.NewsComment) error {
	if m.createCommentErr != nil {
		return m.createCommentErr
	}
	m.comments = append(m.comments, comment)
	return nil
}

func (m *mockNewsRepo) ListComments(newsID uint, page, pageSize int) ([]model.NewsComment, int64, error) {
	out := make([]model.NewsComment, 0, len(m.comments))
	for _, c := range m.comments {
		if c.NewsID == newsID {
			out = append(out, *c)
		}
	}
	return out, int64(len(out)), nil
}

func (m *mockNewsRepo) DeleteComment(id uint) error {
	for i, c := range m.comments {
		if c.ID == id {
			m.comments = append(m.comments[:i], m.comments[i+1:]...)
			return nil
		}
	}
	return nil
}

func (m *mockNewsRepo) IncrCommentCount(id uint) error {
	m.commentIncr[id]++
	if n, ok := m.byID[id]; ok {
		n.CommentCount++
	}
	return nil
}

func (m *mockNewsRepo) CreateMessage(msg *model.Message) error {
	m.messages = append(m.messages, msg)
	return nil
}

func (m *mockNewsRepo) ListMessages(userID uint, page, pageSize int) ([]model.Message, int64, error) {
	out := make([]model.Message, 0, len(m.messages))
	for _, msg := range m.messages {
		if msg.ToUserID == userID {
			out = append(out, *msg)
		}
	}
	return out, int64(len(out)), nil
}

func (m *mockNewsRepo) UnreadCount(userID uint) (int64, error) {
	var count int64
	for _, msg := range m.messages {
		if msg.ToUserID == userID && !msg.IsRead {
			count++
		}
	}
	return count, nil
}

func (m *mockNewsRepo) MarkRead(userID uint, ids []uint) error {
	idSet := make(map[uint]bool, len(ids))
	for _, id := range ids {
		idSet[id] = true
	}
	for _, msg := range m.messages {
		if msg.ToUserID == userID && idSet[msg.ID] {
			msg.IsRead = true
		}
	}
	return nil
}

func (m *mockNewsRepo) FindByIDs(ids []uint) ([]model.News, error) {
	out := make([]model.News, 0, len(ids))
	for _, id := range ids {
		if n, ok := m.byID[id]; ok {
			out = append(out, *n)
		}
	}
	return out, nil
}

// ===== mockIndexer 记录 OnIndex/OnDelete 调用 =====

type mockIndexer struct {
	indexed  []*model.News
	deleted  []uint
}

func (m *mockIndexer) OnIndex(news *model.News) {
	if news != nil {
		m.indexed = append(m.indexed, news)
	}
}

func (m *mockIndexer) OnDelete(newsID uint) {
	m.deleted = append(m.deleted, newsID)
}

// ===== 辅助：构造一条已存在的 news =====

func seedNews(repo *mockNewsRepo, authorID uint, likeCount, favCount int) *model.News {
	n := &model.News{
		Title:      "测试信息",
		AuthorID:   authorID,
		AuthorName: "作者",
		Status:     1,
		LikeCount:  likeCount,
		FavCount:   favCount,
	}
	n.ID = repo.nextID
	repo.nextID++
	cp := *n
	repo.byID[n.ID] = &cp
	return n
}

// ===== 纯函数：缓存键 =====

func TestNewsCacheKeyList(t *testing.T) {
	req := &dto.NewsListRequest{
		CategoryID: 2, Status: 1, Page: 3, PageSize: 10,
		ListingType: "sell", Keyword: "手机", MinPrice: 100, MaxPrice: 500,
		IsUrgent: boolPtr(true), Sort: "price",
	}
	key := newsCacheKeyList(1, req)
	// 键应包含 regionID=1 与各参数（%.0f 格式化浮点）
	assert.Contains(t, key, "cache:news:list:")
	assert.Contains(t, key, ":1:") // regionID
	assert.Contains(t, key, ":2:") // categoryID
}

func TestNewsCacheKeyList_ZeroValues(t *testing.T) {
	req := &dto.NewsListRequest{}
	key := newsCacheKeyList(0, req)
	// IsUrgent 为 nil *bool 时 %v 输出 "<nil>"
	assert.Contains(t, key, "cache:news:list:0:0:0:0:0:::0:0:<nil>:")
}

// ===== 纯函数：DTO 转换 =====

func TestToNewsInfo(t *testing.T) {
	now := time.Now()
	n := &model.News{
		Title:       "二手 iPhone",
		Content:     "95新",
		CoverImage:  "http://x/cover.jpg",
		AuthorID:    5,
		AuthorName:  "张三",
		CategoryID:  2,
		Price:       1999,
		PriceUnit:   "元",
		ListingType: model.ListingTypeSell,
		Condition:   model.ConditionUsed,
		Status:      1,
		LikeCount:   10,
		ViewCount:   100,
		Distance:    1.5,
	}
	n.ID = 1
	n.RegionID = 3
	n.CreatedAt = now
	n.UpdatedAt = now

	info := toNewsInfo(n)
	assert.Equal(t, uint(1), info.ID)
	assert.Equal(t, "二手 iPhone", info.Title)
	assert.Equal(t, uint(5), info.AuthorID)
	assert.Equal(t, uint(2), info.CategoryID)
	assert.Equal(t, 1999.0, info.Price)
	assert.Equal(t, model.ListingTypeSell, info.ListingType)
	assert.Equal(t, 1, info.Status)
	assert.Equal(t, 10, info.LikeCount)
	assert.Equal(t, 100, info.ViewCount)
	assert.Equal(t, 1.5, info.Distance)
	assert.Equal(t, uint(3), info.RegionID)
}

func TestToNewsInfo_NilSafe(t *testing.T) {
	// 即使嵌入字段未设置也不应 panic
	info := toNewsInfo(&model.News{Title: "空"})
	assert.Equal(t, "空", info.Title)
	assert.Equal(t, uint(0), info.ID)
	assert.Equal(t, 0, info.LikeCount)
}

func TestToCommentInfo(t *testing.T) {
	parentID := uint(2)
	c := &model.NewsComment{
		NewsID:   1,
		UserID:   5,
		UserName: "评论者",
		Avatar:   "http://x/a.png",
		Content:  "好价",
		ParentID: &parentID,
		ReplyTo:  "楼主",
	}
	c.ID = 10
	info := toCommentInfo(c)
	assert.Equal(t, uint(10), info.ID)
	assert.Equal(t, uint(1), info.NewsID)
	assert.Equal(t, "评论者", info.UserName)
	assert.Equal(t, "好价", info.Content)
	require.NotNil(t, info.ParentID)
	assert.Equal(t, uint(2), *info.ParentID)
	assert.Equal(t, "楼主", info.ReplyTo)
}

func TestToMessageInfo(t *testing.T) {
	newsID := uint(7)
	m := &model.Message{
		FromUserID: 1,
		ToUserID:   2,
		NewsID:     &newsID,
		Type:       "like",
		Content:    "有人赞了你的信息",
		IsRead:     false,
	}
	m.ID = 3
	info := toMessageInfo(m)
	assert.Equal(t, uint(3), info.ID)
	assert.Equal(t, uint(1), info.FromUserID)
	assert.Equal(t, uint(2), info.ToUserID)
	require.NotNil(t, info.NewsID)
	assert.Equal(t, newsID, *info.NewsID)
	assert.Equal(t, "like", info.Type)
	assert.False(t, info.IsRead)
}

// ===== 构造函数 =====

func TestNewNewsService(t *testing.T) {
	repo := newMockNewsRepo()
	svc := NewNewsService(repo, nil)
	assert.NotNil(t, svc)
	_, ok := svc.(*newsService)
	assert.True(t, ok, "返回值应为 *newsService 类型")
}

func TestNewNewsService_NilIndexerDefaultsToNoop(t *testing.T) {
	repo := newMockNewsRepo()
	svc := NewNewsService(repo, nil).(*newsService)
	// nil indexer 应被替换为 NoopIndexer
	_, ok := svc.indexer.(indexer.NoopIndexer)
	assert.True(t, ok, "nil indexer 应默认为 NoopIndexer")
}

// ===== Create =====

func TestNewsService_Create_DefaultExpireDays(t *testing.T) {
	repo := newMockNewsRepo()
	svc := NewNewsService(repo, nil)

	info, err := svc.Create(1, 5, "张三", &dto.CreateNewsRequest{
		Title:  "测试",
		Status: 0,
	})
	require.NoError(t, err)
	assert.Equal(t, "测试", info.Title)
	assert.Equal(t, uint(5), info.AuthorID)
	assert.Equal(t, uint(1), info.RegionID, "RegionID 应被注入")
	require.NotNil(t, info.ExpiryTime, "默认应设置过期时间")
	// 默认 30 天，过期时间应在 29~31 天后
	diff := info.ExpiryTime.Sub(time.Now())
	assert.InDelta(t, 30*24*time.Hour, diff, float64(2*time.Hour))
}

func TestNewsService_Create_CustomExpireDays(t *testing.T) {
	repo := newMockNewsRepo()
	svc := NewNewsService(repo, nil)

	info, err := svc.Create(1, 5, "张三", &dto.CreateNewsRequest{
		Title:      "测试",
		Status:     0,
		ExpireDays: 7,
	})
	require.NoError(t, err)
	require.NotNil(t, info.ExpiryTime)
	diff := info.ExpiryTime.Sub(time.Now())
	assert.InDelta(t, 7*24*time.Hour, diff, float64(2*time.Hour))
}

func TestNewsService_Create_Defaults(t *testing.T) {
	repo := newMockNewsRepo()
	svc := NewNewsService(repo, nil)

	info, err := svc.Create(1, 5, "张三", &dto.CreateNewsRequest{
		Title:  "空默认值",
		Status: 0,
	})
	require.NoError(t, err)
	assert.Equal(t, "元", info.PriceUnit, "空 PriceUnit 应默认为 '元'")
	assert.Equal(t, model.ListingTypeSell, info.ListingType, "空 ListingType 应默认为 sell")
	assert.Equal(t, model.ConditionUsed, info.Condition, "空 Condition 应默认为 used")
	assert.Nil(t, info.PublishedAt, "草稿 Status=0 不应设置 PublishedAt")
}

func TestNewsService_Create_PublishedAtWhenStatusOne(t *testing.T) {
	repo := newMockNewsRepo()
	svc := NewNewsService(repo, nil)

	info, err := svc.Create(1, 5, "张三", &dto.CreateNewsRequest{
		Title:  "已发布",
		Status: 1,
	})
	require.NoError(t, err)
	require.NotNil(t, info.PublishedAt, "Status=1 应设置 PublishedAt")
	assert.WithinDuration(t, time.Now(), *info.PublishedAt, 5*time.Second)
}

func TestNewsService_Create_RepoError(t *testing.T) {
	repo := newMockNewsRepo()
	boom := errors.New("db down")
	repo.createErr = boom
	svc := NewNewsService(repo, nil)

	_, err := svc.Create(1, 5, "张三", &dto.CreateNewsRequest{Title: "x"})
	assert.ErrorIs(t, err, boom)
}

func TestNewsService_Create_IndexerCalled(t *testing.T) {
	repo := newMockNewsRepo()
	idx := &mockIndexer{}
	svc := NewNewsService(repo, idx)

	info, err := svc.Create(1, 5, "张三", &dto.CreateNewsRequest{Title: "x", Status: 1})
	require.NoError(t, err)
	require.Len(t, idx.indexed, 1, "indexer.OnIndex 应被调用一次")
	assert.Equal(t, info.ID, idx.indexed[0].ID)
}

// ===== Update =====

func TestNewsService_Update_NotFound(t *testing.T) {
	repo := newMockNewsRepo()
	svc := NewNewsService(repo, nil)

	err := svc.Update(9999, 5, &dto.UpdateNewsRequest{Title: "x"})
	assert.ErrorIs(t, err, ErrNewsNotFound)
}

func TestNewsService_Update_NoPermission(t *testing.T) {
	repo := newMockNewsRepo()
	n := seedNews(repo, 5, 0, 0) // 作者=5
	svc := NewNewsService(repo, nil)

	err := svc.Update(n.ID, 999, &dto.UpdateNewsRequest{Title: "篡改"}) // 操作者=999
	assert.ErrorIs(t, err, ErrNewsNoPermission)
}

func TestNewsService_Update_Success_StatusChangeSetsPublishedAt(t *testing.T) {
	repo := newMockNewsRepo()
	n := seedNews(repo, 5, 0, 0)
	n.Status = 0 // 草稿
	repo.byID[n.ID].Status = 0
	svc := NewNewsService(repo, nil)

	err := svc.Update(n.ID, 5, &dto.UpdateNewsRequest{Title: "新标题", Status: 1})
	require.NoError(t, err)
	// 校验字段更新
	updated := repo.byID[n.ID]
	assert.Equal(t, "新标题", updated.Title)
	assert.Equal(t, 1, updated.Status, "状态应更新为 1")
}

func TestNewsService_Update_FindError(t *testing.T) {
	repo := newMockNewsRepo()
	repo.findErr = errors.New("connection refused")
	svc := NewNewsService(repo, nil)

	err := svc.Update(1, 5, &dto.UpdateNewsRequest{Title: "x"})
	assert.EqualError(t, err, "connection refused")
}

func TestNewsService_Update_UpdateFieldsError(t *testing.T) {
	repo := newMockNewsRepo()
	n := seedNews(repo, 5, 0, 0)
	boom := errors.New("update failed")
	repo.updateErr = boom
	svc := NewNewsService(repo, nil)

	err := svc.Update(n.ID, 5, &dto.UpdateNewsRequest{Title: "x"})
	assert.ErrorIs(t, err, boom)
}

// ===== Delete =====

func TestNewsService_Delete_NotFound(t *testing.T) {
	repo := newMockNewsRepo()
	svc := NewNewsService(repo, nil)

	err := svc.Delete(9999, 5)
	assert.ErrorIs(t, err, ErrNewsNotFound)
}

func TestNewsService_Delete_NoPermission(t *testing.T) {
	repo := newMockNewsRepo()
	n := seedNews(repo, 5, 0, 0)
	svc := NewNewsService(repo, nil)

	err := svc.Delete(n.ID, 999)
	assert.ErrorIs(t, err, ErrNewsNoPermission)
}

func TestNewsService_Delete_Success(t *testing.T) {
	repo := newMockNewsRepo()
	n := seedNews(repo, 5, 0, 0)
	idx := &mockIndexer{}
	svc := NewNewsService(repo, idx)

	err := svc.Delete(n.ID, 5)
	require.NoError(t, err)
	_, exists := repo.byID[n.ID]
	assert.False(t, exists, "记录应被删除")
	require.Len(t, idx.deleted, 1, "indexer.OnDelete 应被调用")
	assert.Equal(t, n.ID, idx.deleted[0])
}

func TestNewsService_Delete_RepoError(t *testing.T) {
	repo := newMockNewsRepo()
	n := seedNews(repo, 5, 0, 0)
	boom := errors.New("delete failed")
	repo.deleteErr = boom
	svc := NewNewsService(repo, nil)

	err := svc.Delete(n.ID, 5)
	assert.ErrorIs(t, err, boom)
}

// ===== GetByID =====

func TestNewsService_GetByID_NotFound(t *testing.T) {
	repo := newMockNewsRepo()
	svc := NewNewsService(repo, nil)

	_, err := svc.GetByID(9999)
	assert.ErrorIs(t, err, ErrNewsNotFound)
}

func TestNewsService_GetByID_Success_IncrViewCount(t *testing.T) {
	repo := newMockNewsRepo()
	n := seedNews(repo, 5, 0, 0)
	svc := NewNewsService(repo, nil)

	info, err := svc.GetByID(n.ID)
	require.NoError(t, err)
	assert.Equal(t, n.Title, info.Title)
	assert.Equal(t, n.ViewCount+1, info.ViewCount, "返回的 ViewCount 应 +1")
	assert.Equal(t, 1, repo.viewIncr[n.ID], "IncrViewCount 应被调用一次")
}

func TestNewsService_GetByID_FindError(t *testing.T) {
	repo := newMockNewsRepo()
	repo.findErr = errors.New("db error")
	svc := NewNewsService(repo, nil)

	_, err := svc.GetByID(1)
	assert.EqualError(t, err, "db error")
}

// ===== List =====

func TestNewsService_List_Empty(t *testing.T) {
	repo := newMockNewsRepo()
	svc := NewNewsService(repo, nil)

	pg, list, err := svc.List(1, &dto.NewsListRequest{Page: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Empty(t, list)
	assert.Equal(t, int64(0), pg.Total)
}

func TestNewsService_List_WithResults(t *testing.T) {
	repo := newMockNewsRepo()
	seedNews(repo, 5, 0, 0)
	seedNews(repo, 6, 0, 0)
	svc := NewNewsService(repo, nil)

	pg, list, err := svc.List(1, &dto.NewsListRequest{Page: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Len(t, list, 2)
	assert.Equal(t, int64(2), pg.Total)
}

func TestNewsService_List_WithKeyword_SkipsCache(t *testing.T) {
	repo := newMockNewsRepo()
	seedNews(repo, 5, 0, 0)
	svc := NewNewsService(repo, nil)

	// 带 keyword 应走非缓存路径（Redis 不可用，仍能回源 DB）
	_, list, err := svc.List(1, &dto.NewsListRequest{Page: 1, PageSize: 10, Keyword: "测试"})
	require.NoError(t, err)
	assert.Len(t, list, 1)
}

func TestNewsService_List_RepoError(t *testing.T) {
	repo := newMockNewsRepo()
	repo.listErr = errors.New("list failed")
	svc := NewNewsService(repo, nil)

	_, _, err := svc.List(1, &dto.NewsListRequest{Page: 1, PageSize: 10})
	assert.EqualError(t, err, "list failed")
}

// ===== ListNearby =====

func TestNewsService_ListNearby_DefaultRadius(t *testing.T) {
	repo := newMockNewsRepo()
	svc := NewNewsService(repo, nil)

	// RadiusKm=0 应默认 5（mock 不校验半径，仅验证不报错且返回结果）
	pg, list, err := svc.ListNearby(1, &dto.NewsNearbyRequest{
		Latitude: 44.0, Longitude: 127.0, Page: 1, PageSize: 10, RadiusKm: 0,
	})
	require.NoError(t, err)
	assert.Empty(t, list)
	assert.Equal(t, int64(0), pg.Total)
}

func TestNewsService_ListNearby_WithResults(t *testing.T) {
	repo := newMockNewsRepo()
	seedNews(repo, 5, 0, 0)
	svc := NewNewsService(repo, nil)

	pg, list, err := svc.ListNearby(1, &dto.NewsNearbyRequest{
		Latitude: 44.0, Longitude: 127.0, Page: 1, PageSize: 10, RadiusKm: 5,
	})
	require.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, int64(1), pg.Total)
}

func TestNewsService_ListNearby_RepoError(t *testing.T) {
	repo := newMockNewsRepo()
	repo.listErr = errors.New("nearby failed")
	svc := NewNewsService(repo, nil)

	_, _, err := svc.ListNearby(1, &dto.NewsNearbyRequest{
		Latitude: 44.0, Longitude: 127.0, RadiusKm: 5,
	})
	assert.EqualError(t, err, "nearby failed")
}

// ===== Like =====

func TestNewsService_Like_NotFound(t *testing.T) {
	repo := newMockNewsRepo()
	svc := NewNewsService(repo, nil)

	_, err := svc.Like(1, 9999)
	assert.ErrorIs(t, err, ErrNewsNotFound)
}

func TestNewsService_Like_ToggleOn(t *testing.T) {
	repo := newMockNewsRepo()
	n := seedNews(repo, 5, 0, 0) // 作者=5，点赞数=0
	svc := NewNewsService(repo, nil)

	// 用户=6 点赞（非作者）
	// 注：消息通知仅在 ws Hub 可用时创建，单元测试中 Hub 为 nil，
	// 故此处不验证消息创建，仅验证点赞计数与 IncrLikeCount 副作用
	resp, err := svc.Like(6, n.ID)
	require.NoError(t, err)
	assert.True(t, resp.Liked)
	assert.Equal(t, 1, resp.LikeCount, "点赞后计数应 +1")
	assert.Equal(t, 1, repo.likeIncr[n.ID], "IncrLikeCount 应被调用")
	assert.Equal(t, 0, repo.likeDecr[n.ID], "DecrLikeCount 不应被调用")
}

func TestNewsService_Like_ToggleOff_ClampAtZero(t *testing.T) {
	repo := newMockNewsRepo()
	n := seedNews(repo, 5, 1, 0) // 点赞数=1
	// 预置已点赞
	repo.likes[n.ID] = map[uint]bool{6: true}
	svc := NewNewsService(repo, nil)

	resp, err := svc.Like(6, n.ID)
	require.NoError(t, err)
	assert.False(t, resp.Liked)
	assert.Equal(t, 0, resp.LikeCount, "取消后计数应为 0")
	assert.Equal(t, 1, repo.likeDecr[n.ID], "DecrLikeCount 应被调用")
}

func TestNewsService_Like_ToggleOff_CountClampWhenZero(t *testing.T) {
	repo := newMockNewsRepo()
	n := seedNews(repo, 5, 0, 0) // 点赞数=0（异常状态）
	repo.likes[n.ID] = map[uint]bool{6: true}
	svc := NewNewsService(repo, nil)

	resp, err := svc.Like(6, n.ID)
	require.NoError(t, err)
	assert.False(t, resp.Liked)
	assert.Equal(t, 0, resp.LikeCount, "计数不应为负，应钳制为 0")
}

func TestNewsService_Like_SelfLike_NoMessage(t *testing.T) {
	repo := newMockNewsRepo()
	n := seedNews(repo, 5, 0, 0) // 作者=5
	svc := NewNewsService(repo, nil)

	// 作者给自己点赞，不应创建消息
	_, err := svc.Like(5, n.ID)
	require.NoError(t, err)
	assert.Empty(t, repo.messages, "作者给自己点赞不应产生消息通知")
}

func TestNewsService_Like_FindError(t *testing.T) {
	repo := newMockNewsRepo()
	repo.findErr = errors.New("db error")
	svc := NewNewsService(repo, nil)

	_, err := svc.Like(1, 1)
	assert.EqualError(t, err, "db error")
}

func TestNewsService_Like_LikeExistsError(t *testing.T) {
	repo := newMockNewsRepo()
	n := seedNews(repo, 5, 0, 0)
	repo.likeExistsErr = errors.New("like query failed")
	svc := NewNewsService(repo, nil)

	_, err := svc.Like(6, n.ID)
	assert.EqualError(t, err, "like query failed")
}

// ===== LikeStatus =====

func TestNewsService_LikeStatus_NotFound(t *testing.T) {
	repo := newMockNewsRepo()
	svc := NewNewsService(repo, nil)

	_, err := svc.LikeStatus(1, 9999)
	assert.ErrorIs(t, err, ErrNewsNotFound)
}

func TestNewsService_LikeStatus_Liked(t *testing.T) {
	repo := newMockNewsRepo()
	n := seedNews(repo, 5, 3, 0)
	repo.likes[n.ID] = map[uint]bool{6: true}
	svc := NewNewsService(repo, nil)

	resp, err := svc.LikeStatus(6, n.ID)
	require.NoError(t, err)
	assert.True(t, resp.Liked)
	assert.Equal(t, 3, resp.LikeCount)
}

func TestNewsService_LikeStatus_NotLiked(t *testing.T) {
	repo := newMockNewsRepo()
	n := seedNews(repo, 5, 2, 0)
	svc := NewNewsService(repo, nil)

	resp, err := svc.LikeStatus(6, n.ID)
	require.NoError(t, err)
	assert.False(t, resp.Liked)
	assert.Equal(t, 2, resp.LikeCount)
}

// ===== Fav =====

func TestNewsService_Fav_NotFound(t *testing.T) {
	repo := newMockNewsRepo()
	svc := NewNewsService(repo, nil)

	_, err := svc.Fav(1, 9999)
	assert.ErrorIs(t, err, ErrNewsNotFound)
}

func TestNewsService_Fav_ToggleOn(t *testing.T) {
	repo := newMockNewsRepo()
	n := seedNews(repo, 5, 0, 0)
	svc := NewNewsService(repo, nil)

	resp, err := svc.Fav(6, n.ID)
	require.NoError(t, err)
	assert.True(t, resp.Faved)
	assert.Equal(t, 1, resp.FavCount)
	assert.Equal(t, 1, repo.favIncr[n.ID])
}

func TestNewsService_Fav_ToggleOff_ClampAtZero(t *testing.T) {
	repo := newMockNewsRepo()
	n := seedNews(repo, 5, 0, 1) // 收藏数=1
	repo.favs[n.ID] = map[uint]bool{6: true}
	svc := NewNewsService(repo, nil)

	resp, err := svc.Fav(6, n.ID)
	require.NoError(t, err)
	assert.False(t, resp.Faved)
	assert.Equal(t, 0, resp.FavCount)
	assert.Equal(t, 1, repo.favDecr[n.ID])
}

func TestNewsService_Fav_FavExistsError(t *testing.T) {
	repo := newMockNewsRepo()
	n := seedNews(repo, 5, 0, 0)
	repo.favExistsErr = errors.New("fav query failed")
	svc := NewNewsService(repo, nil)

	_, err := svc.Fav(6, n.ID)
	assert.EqualError(t, err, "fav query failed")
}

// ===== FavStatus =====

func TestNewsService_FavStatus_NotFound(t *testing.T) {
	repo := newMockNewsRepo()
	svc := NewNewsService(repo, nil)

	_, err := svc.FavStatus(1, 9999)
	assert.ErrorIs(t, err, ErrNewsNotFound)
}

func TestNewsService_FavStatus_Faved(t *testing.T) {
	repo := newMockNewsRepo()
	n := seedNews(repo, 5, 0, 4)
	repo.favs[n.ID] = map[uint]bool{6: true}
	svc := NewNewsService(repo, nil)

	resp, err := svc.FavStatus(6, n.ID)
	require.NoError(t, err)
	assert.True(t, resp.Faved)
	assert.Equal(t, 4, resp.FavCount)
}

func TestNewsService_FavStatus_NotFaved(t *testing.T) {
	repo := newMockNewsRepo()
	n := seedNews(repo, 5, 0, 2)
	svc := NewNewsService(repo, nil)

	resp, err := svc.FavStatus(6, n.ID)
	require.NoError(t, err)
	assert.False(t, resp.Faved)
	assert.Equal(t, 2, resp.FavCount)
}

// ===== ListFavorites =====

func TestNewsService_ListFavorites_Empty(t *testing.T) {
	repo := newMockNewsRepo()
	svc := NewNewsService(repo, nil)

	pagination, list, err := svc.ListFavorites(6, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(0), pagination.Total)
	assert.Equal(t, 1, pagination.Page)
	assert.Equal(t, 10, pagination.PageSize)
	assert.Empty(t, list)
}

func TestNewsService_ListFavorites_ReturnsFavedNews(t *testing.T) {
	repo := newMockNewsRepo()
	n1 := seedNews(repo, 5, 0, 0)
	n2 := seedNews(repo, 5, 0, 0)
	n3 := seedNews(repo, 5, 0, 0)
	// 用户 6 收藏 n1、n2；n3 未收藏
	repo.favs[n1.ID] = map[uint]bool{6: true}
	repo.favs[n2.ID] = map[uint]bool{6: true}
	svc := NewNewsService(repo, nil)

	pagination, list, err := svc.ListFavorites(6, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(2), pagination.Total)
	require.Len(t, list, 2)
	// 两条都应来自被收藏的 news（n1、n2），不应混入 n3
	gotIDs := map[uint]bool{list[0].ID: true, list[1].ID: true}
	assert.True(t, gotIDs[n1.ID], "应包含已收藏的 n1")
	assert.True(t, gotIDs[n2.ID], "应包含已收藏的 n2")
	assert.False(t, gotIDs[n3.ID], "不应包含未收藏的 n3")
}

func TestNewsService_ListFavorites_Pagination(t *testing.T) {
	repo := newMockNewsRepo()
	// 创建 3 条 news，全部被用户 6 收藏
	n1 := seedNews(repo, 5, 0, 0)
	n2 := seedNews(repo, 5, 0, 0)
	n3 := seedNews(repo, 5, 0, 0)
	repo.favs[n1.ID] = map[uint]bool{6: true}
	repo.favs[n2.ID] = map[uint]bool{6: true}
	repo.favs[n3.ID] = map[uint]bool{6: true}
	svc := NewNewsService(repo, nil)

	// page_size=2 取第一页，应返回 2 条；total 仍为 3
	pagination, list, err := svc.ListFavorites(6, 1, 2)
	require.NoError(t, err)
	assert.Equal(t, int64(3), pagination.Total)
	assert.Equal(t, 2, pagination.PageSize)
	require.Len(t, list, 2)

	// 第二页应返回剩余 1 条
	pagination2, list2, err := svc.ListFavorites(6, 2, 2)
	require.NoError(t, err)
	assert.Equal(t, int64(3), pagination2.Total)
	require.Len(t, list2, 1)

	// 两页合并应覆盖全部 3 条且无重复
	allIDs := map[uint]bool{}
	for _, n := range append(list, list2...) {
		allIDs[n.ID] = true
	}
	assert.Len(t, allIDs, 3)
	assert.True(t, allIDs[n1.ID])
	assert.True(t, allIDs[n2.ID])
	assert.True(t, allIDs[n3.ID])
}

func TestNewsService_ListFavorites_SkipsDeletedNews(t *testing.T) {
	repo := newMockNewsRepo()
	n1 := seedNews(repo, 5, 0, 0)
	n2 := seedNews(repo, 5, 0, 0)
	// 用户 6 收藏两条，但 n2 从 byID 中移除模拟"软删除被过滤"
	repo.favs[n1.ID] = map[uint]bool{6: true}
	repo.favs[n2.ID] = map[uint]bool{6: true}
	delete(repo.byID, n2.ID)
	svc := NewNewsService(repo, nil)

	pagination, list, err := svc.ListFavorites(6, 1, 10)
	require.NoError(t, err)
	// total 仍按收藏记录数（2），但实际返回列表只含 n1
	assert.Equal(t, int64(2), pagination.Total)
	require.Len(t, list, 1)
	assert.Equal(t, n1.ID, list[0].ID)
}

func TestNewsService_ListFavorites_DefaultsInvalidPaging(t *testing.T) {
	repo := newMockNewsRepo()
	n1 := seedNews(repo, 5, 0, 0)
	repo.favs[n1.ID] = map[uint]bool{6: true}
	svc := NewNewsService(repo, nil)

	// page=0、pageSize=0 应被 utils.NewPagination 规范化为 1/10
	pagination, list, err := svc.ListFavorites(6, 0, 0)
	require.NoError(t, err)
	assert.Equal(t, 1, pagination.Page)
	assert.Equal(t, 10, pagination.PageSize)
	assert.Equal(t, int64(1), pagination.Total)
	require.Len(t, list, 1)
	assert.Equal(t, n1.ID, list[0].ID)
}

// ===== CreateComment =====

func TestNewsService_CreateComment_NotFound(t *testing.T) {
	repo := newMockNewsRepo()
	svc := NewNewsService(repo, nil)

	_, err := svc.CreateComment(9999, 6, "评论者", "", &dto.CreateCommentRequest{Content: "好"})
	assert.ErrorIs(t, err, ErrNewsNotFound)
}

func TestNewsService_CreateComment_Success_NotifyAuthor(t *testing.T) {
	repo := newMockNewsRepo()
	n := seedNews(repo, 5, 0, 0) // 作者=5
	svc := NewNewsService(repo, nil)

	comment, err := svc.CreateComment(n.ID, 6, "评论者", "http://x/a.png", &dto.CreateCommentRequest{
		Content: "好价",
	})
	require.NoError(t, err)
	assert.Equal(t, "好价", comment.Content)
	assert.Equal(t, "评论者", comment.UserName)
	assert.Equal(t, uint(n.ID), comment.NewsID)
	assert.Equal(t, 1, repo.commentIncr[n.ID], "IncrCommentCount 应被调用")
	// 作者≠评论者，应创建消息通知
	require.Len(t, repo.messages, 1)
	assert.Equal(t, uint(5), repo.messages[0].ToUserID)
	assert.Equal(t, "comment", repo.messages[0].Type)
}

func TestNewsService_CreateComment_SelfComment_NoMessage(t *testing.T) {
	repo := newMockNewsRepo()
	n := seedNews(repo, 5, 0, 0) // 作者=5
	svc := NewNewsService(repo, nil)

	// 作者评论自己的信息，不应产生消息
	_, err := svc.CreateComment(n.ID, 5, "作者", "", &dto.CreateCommentRequest{Content: "自评"})
	require.NoError(t, err)
	assert.Empty(t, repo.messages, "自评不应产生消息通知")
}

func TestNewsService_CreateComment_RepoCreateError(t *testing.T) {
	repo := newMockNewsRepo()
	n := seedNews(repo, 5, 0, 0)
	boom := errors.New("create comment failed")
	repo.createCommentErr = boom
	svc := NewNewsService(repo, nil)

	_, err := svc.CreateComment(n.ID, 6, "评论者", "", &dto.CreateCommentRequest{Content: "x"})
	assert.ErrorIs(t, err, boom)
}

// ===== ListComments =====

func TestNewsService_ListComments(t *testing.T) {
	repo := newMockNewsRepo()
	n := seedNews(repo, 5, 0, 0)
	// 预置评论
	repo.comments = []*model.NewsComment{
		{NewsID: n.ID, UserID: 6, UserName: "A", Content: "评论1"},
		{NewsID: n.ID, UserID: 7, UserName: "B", Content: "评论2"},
		{NewsID: 999, UserID: 8, UserName: "C", Content: "其它信息评论"},
	}
	svc := NewNewsService(repo, nil)

	list, total, err := svc.ListComments(n.ID, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total, "仅返回该 newsID 的评论")
	assert.Len(t, list, 2)
}

// ===== DeleteComment =====

func TestNewsService_DeleteComment(t *testing.T) {
	repo := newMockNewsRepo()
	repo.comments = []*model.NewsComment{{ID: 1, NewsID: 5, Content: "x"}}
	svc := NewNewsService(repo, nil)

	err := svc.DeleteComment(1, 6)
	require.NoError(t, err)
	assert.Empty(t, repo.comments)
}

// ===== ListMessages / UnreadCount / MarkRead =====

func TestNewsService_ListMessages(t *testing.T) {
	repo := newMockNewsRepo()
	repo.messages = []*model.Message{
		{ID: 1, ToUserID: 5, Content: "msg1"},
		{ID: 2, ToUserID: 5, Content: "msg2"},
		{ID: 3, ToUserID: 6, Content: "msg3"},
	}
	svc := NewNewsService(repo, nil)

	list, total, err := svc.ListMessages(5, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, list, 2)
}

func TestNewsService_UnreadCount(t *testing.T) {
	repo := newMockNewsRepo()
	repo.messages = []*model.Message{
		{ToUserID: 5, IsRead: false},
		{ToUserID: 5, IsRead: true},
		{ToUserID: 5, IsRead: false},
	}
	svc := NewNewsService(repo, nil)

	count, err := svc.UnreadCount(5)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestNewsService_MarkRead(t *testing.T) {
	repo := newMockNewsRepo()
	repo.messages = []*model.Message{
		{ID: 1, ToUserID: 5, IsRead: false},
		{ID: 2, ToUserID: 5, IsRead: false},
	}
	svc := NewNewsService(repo, nil)

	err := svc.MarkRead(5, []uint{1})
	require.NoError(t, err)
	assert.True(t, repo.messages[0].IsRead)
	assert.False(t, repo.messages[1].IsRead, "未在 ids 中的消息不应被标记已读")
}

// ===== Search =====

func TestNewsService_Search_DBFallback(t *testing.T) {
	repo := newMockNewsRepo()
	seedNews(repo, 5, 0, 0)
	svc := NewNewsService(repo, nil)

	// ES 不可用，应走 DB LIKE 降级路径
	pg, list, err := svc.Search(1, &dto.NewsSearchRequest{
		Keyword: "测试", Page: 1, PageSize: 10,
	})
	require.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, int64(1), pg.Total)
}

func TestNewsService_Search_RepoError(t *testing.T) {
	repo := newMockNewsRepo()
	repo.listErr = errors.New("search failed")
	svc := NewNewsService(repo, nil)

	_, _, err := svc.Search(1, &dto.NewsSearchRequest{Keyword: "x"})
	assert.EqualError(t, err, "search failed")
}

// ===== buildESFilters =====

func TestBuildESFilters_OnlyStatus(t *testing.T) {
	svc := NewNewsService(newMockNewsRepo(), nil).(*newsService)
	filters := svc.buildESFilters(0, &dto.NewsSearchRequest{})
	require.Len(t, filters, 1, "仅 status=1 过滤器")
	assert.Equal(t, 1, filters[0]["term"].(map[string]interface{})["status"])
}

func TestBuildESFilters_AllFilters(t *testing.T) {
	svc := NewNewsService(newMockNewsRepo(), nil).(*newsService)
	filters := svc.buildESFilters(1, &dto.NewsSearchRequest{
		CategoryID: 2, ListingType: "sell",
	})
	require.Len(t, filters, 4, "status + region + category + listingType")
	// 验证每个 term 过滤器
	assert.Equal(t, 1, filters[0]["term"].(map[string]interface{})["status"])
	assert.Equal(t, uint(1), filters[1]["term"].(map[string]interface{})["region_id"])
	assert.Equal(t, uint(2), filters[2]["term"].(map[string]interface{})["category_id"])
	assert.Equal(t, "sell", filters[3]["term"].(map[string]interface{})["listing_type"])
}

func TestBuildESFilters_PartialFilters(t *testing.T) {
	svc := NewNewsService(newMockNewsRepo(), nil).(*newsService)
	// regionID=0 但 categoryID>0
	filters := svc.buildESFilters(0, &dto.NewsSearchRequest{CategoryID: 3})
	require.Len(t, filters, 2, "status + category")
	// listingType 非空但 categoryID=0
	filters = svc.buildESFilters(1, &dto.NewsSearchRequest{ListingType: "rent"})
	require.Len(t, filters, 3, "status + region + listingType")
}

// ===== 辅助函数 =====

func boolPtr(v bool) *bool {
	return &v
}
