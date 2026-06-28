// Package service 同城分类信息业务逻辑层
package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"wuchang-tongcheng/internal/modules/news/dto"
	"wuchang-tongcheng/internal/modules/news/indexer"
	"wuchang-tongcheng/internal/modules/news/model"
	"wuchang-tongcheng/internal/modules/news/repository"
	"wuchang-tongcheng/internal/pkg/es"
	rediscache "wuchang-tongcheng/internal/pkg/redis"
	"wuchang-tongcheng/internal/pkg/utils"
	wspkg "wuchang-tongcheng/internal/pkg/ws"

	"gorm.io/gorm"
)

var (
	ErrNewsNotFound     = errors.New("信息不存在")
	ErrNewsNoPermission = errors.New("无权操作此信息")
	ErrCommentNotFound  = errors.New("评论不存在")
)

const (
	newsCachePrefix = "cache:news:"
	newsCacheTTL    = 60 * time.Second
)

func newsCacheKeyList(regionID uint, req *dto.NewsListRequest) string {
	return fmt.Sprintf(newsCachePrefix+"list:%d:%d:%d:%d:%d:%s:%s:%.0f:%.0f:%v:%s",
		regionID, req.CategoryID, req.Status, req.Page, req.PageSize,
		req.ListingType, req.Keyword, req.MinPrice, req.MaxPrice, req.IsUrgent, req.Sort)
}

func invalidateNewsCache() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = rediscache.DelByPrefix(ctx, newsCachePrefix)
}

// NewsService 分类信息业务逻辑接口
type NewsService interface {
	Create(regionID uint, authorID uint, authorName string, req *dto.CreateNewsRequest) (*dto.NewsInfo, error)
	Update(id uint, operatorID uint, req *dto.UpdateNewsRequest) error
	Delete(id uint, operatorID uint) error
	GetByID(id uint) (*dto.NewsInfo, error)
	List(regionID uint, req *dto.NewsListRequest) (*utils.Pagination, []dto.NewsInfo, error)
	Search(regionID uint, req *dto.NewsSearchRequest) (*utils.Pagination, []dto.NewsInfo, error)
	Like(userID, newsID uint) (*dto.LikeResponse, error)
	LikeStatus(userID, newsID uint) (*dto.LikeResponse, error)
	// 收藏
	Fav(userID, newsID uint) (*dto.FavResponse, error)
	FavStatus(userID, newsID uint) (*dto.FavResponse, error)
	// 评论
	CreateComment(newsID uint, userID uint, userName string, avatar string, req *dto.CreateCommentRequest) (*dto.CommentInfo, error)
	ListComments(newsID uint, page, pageSize int) ([]dto.CommentInfo, int64, error)
	DeleteComment(id uint, userID uint) error
	// 消息
	ListMessages(userID uint, page, pageSize int) ([]dto.MessageInfo, int64, error)
	UnreadCount(userID uint) (int64, error)
	MarkRead(userID uint, ids []uint) error
}

type newsService struct {
	newsRepo repository.NewsRepository
	indexer  indexer.Indexer
}

func NewNewsService(newsRepo repository.NewsRepository, idx indexer.Indexer) NewsService {
	if idx == nil {
		idx = indexer.NoopIndexer{}
	}
	return &newsService{newsRepo: newsRepo, indexer: idx}
}

func toNewsInfo(n *model.News) *dto.NewsInfo {
	return &dto.NewsInfo{
		ID:          n.ID,
		Title:       n.Title,
		Content:     n.Content,
		CoverImage:  n.CoverImage,
		Images:      n.Images,
		Summary:     n.Summary,
		AuthorID:    n.AuthorID,
		AuthorName:  n.AuthorName,
		CategoryID:  n.CategoryID,
		Tags:        n.Tags,
		Price:       n.Price,
		PriceUnit:   n.PriceUnit,
		ListingType: n.ListingType,
		Condition:   n.Condition,
		ContactPhone:  n.ContactPhone,
		ContactWechat: n.ContactWechat,
		ContactQQ:     n.ContactQQ,
		Address:       n.Address,
		Latitude:      n.Latitude,
		Longitude:     n.Longitude,
		IsUrgent:      n.IsUrgent,
		ExpiryTime:    n.ExpiryTime,
		ViewCount:     n.ViewCount,
		LikeCount:     n.LikeCount,
		FavCount:      n.FavCount,
		CommentCount:  n.CommentCount,
		Status:        n.Status,
		RegionID:      n.RegionID,
		PublishedAt:   n.PublishedAt,
		CreatedAt:     n.CreatedAt,
		UpdatedAt:     n.UpdatedAt,
	}
}

func toCommentInfo(c *model.NewsComment) *dto.CommentInfo {
	return &dto.CommentInfo{
		ID:        c.ID,
		NewsID:    c.NewsID,
		UserID:    c.UserID,
		UserName:  c.UserName,
		Avatar:    c.Avatar,
		Content:   c.Content,
		ParentID:  c.ParentID,
		ReplyTo:   c.ReplyTo,
		CreatedAt: c.CreatedAt,
	}
}

func toMessageInfo(m *model.Message) *dto.MessageInfo {
	return &dto.MessageInfo{
		ID:         m.ID,
		FromUserID: m.FromUserID,
		ToUserID:   m.ToUserID,
		NewsID:     m.NewsID,
		Type:       m.Type,
		Content:    m.Content,
		IsRead:     m.IsRead,
		CreatedAt:  m.CreatedAt,
	}
}

// Create 创建分类信息
func (s *newsService) Create(regionID uint, authorID uint, authorName string, req *dto.CreateNewsRequest) (*dto.NewsInfo, error) {
	expireDays := req.ExpireDays
	if expireDays <= 0 {
		expireDays = 30
	}
	expiryTime := time.Now().AddDate(0, 0, expireDays)

	news := &model.News{
		Title:      req.Title,
		Content:    req.Content,
		CoverImage: req.CoverImage,
		Images:     req.Images,
		Summary:    req.Summary,
		AuthorID:   authorID,
		AuthorName: authorName,
		CategoryID: req.CategoryID,
		Tags:       req.Tags,
		Status:     req.Status,

		Price:       req.Price,
		PriceUnit:   req.PriceUnit,
		ListingType: req.ListingType,
		Condition:   req.Condition,

		ContactPhone:  req.ContactPhone,
		ContactWechat: req.ContactWechat,
		ContactQQ:     req.ContactQQ,

		Address:   req.Address,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,

		ExpiryTime: &expiryTime,
	}
	news.RegionID = regionID

	if req.PriceUnit == "" {
		news.PriceUnit = "元"
	}
	if req.ListingType == "" {
		news.ListingType = model.ListingTypeSell
	}
	if req.Condition == "" {
		news.Condition = model.ConditionUsed
	}

	if req.Status == 1 {
		now := time.Now()
		news.PublishedAt = &now
	}

	if err := s.newsRepo.Create(news); err != nil {
		return nil, err
	}
	invalidateNewsCache()
	s.indexer.OnIndex(news)
	return toNewsInfo(news), nil
}

// Update 更新分类信息
func (s *newsService) Update(id uint, operatorID uint, req *dto.UpdateNewsRequest) error {
	news, err := s.newsRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNewsNotFound
		}
		return err
	}
	if news.AuthorID != operatorID {
		return ErrNewsNoPermission
	}

	fields := map[string]interface{}{}
	if req.Title != "" {
		fields["title"] = req.Title
	}
	if req.Content != "" {
		fields["content"] = req.Content
	}
	if req.CoverImage != "" {
		fields["cover_image"] = req.CoverImage
	}
	if req.Images != "" {
		fields["images"] = req.Images
	}
	if req.Summary != "" {
		fields["summary"] = req.Summary
	}
	fields["category_id"] = req.CategoryID
	if req.Tags != "" {
		fields["tags"] = req.Tags
	}
	if req.ListingType != "" {
		fields["listing_type"] = req.ListingType
	}
	if req.Condition != "" {
		fields["condition"] = req.Condition
	}
	fields["price"] = req.Price
	if req.PriceUnit != "" {
		fields["price_unit"] = req.PriceUnit
	}
	if req.ContactPhone != "" {
		fields["contact_phone"] = req.ContactPhone
	}
	if req.ContactWechat != "" {
		fields["contact_wechat"] = req.ContactWechat
	}
	if req.ContactQQ != "" {
		fields["contact_qq"] = req.ContactQQ
	}
	if req.Address != "" {
		fields["address"] = req.Address
	}
	fields["latitude"] = req.Latitude
	fields["longitude"] = req.Longitude
	fields["is_urgent"] = req.IsUrgent
	if req.ExpiryTime != nil {
		fields["expiry_time"] = req.ExpiryTime
	}

	if req.Status == 1 && news.Status != 1 {
		now := time.Now()
		fields["status"] = 1
		fields["published_at"] = &now
	} else if req.Status >= 0 && req.Status <= 3 {
		fields["status"] = req.Status
	}

	if len(fields) == 0 {
		return nil
	}
	if err := s.newsRepo.UpdateFields(id, fields); err != nil {
		return err
	}
	invalidateNewsCache()
	if updated, err := s.newsRepo.FindByID(id); err == nil {
		s.indexer.OnIndex(updated)
	}
	return nil
}

// Delete 删除分类信息
func (s *newsService) Delete(id uint, operatorID uint) error {
	news, err := s.newsRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNewsNotFound
		}
		return err
	}
	if news.AuthorID != operatorID {
		return ErrNewsNoPermission
	}
	if err := s.newsRepo.Delete(id); err != nil {
		return err
	}
	invalidateNewsCache()
	s.indexer.OnDelete(id)
	return nil
}

// GetByID 获取详情（同时增加浏览量）
func (s *newsService) GetByID(id uint) (*dto.NewsInfo, error) {
	news, err := s.newsRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNewsNotFound
		}
		return nil, err
	}
	_ = s.newsRepo.IncrViewCount(id)
	news.ViewCount++
	return toNewsInfo(news), nil
}

type newsListCache struct {
	Pagination utils.Pagination `json:"pagination"`
	List       []dto.NewsInfo   `json:"list"`
}

// List 分类信息列表
func (s *newsService) List(regionID uint, req *dto.NewsListRequest) (*utils.Pagination, []dto.NewsInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)

	if req.Keyword == "" && req.MinPrice == 0 && req.MaxPrice == 0 && req.IsUrgent == nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if req.ListingType == "" {
			var cached newsListCache
			if hit, _ := rediscache.GetJSON(ctx, newsCacheKeyList(regionID, req), &cached); hit {
				return &cached.Pagination, cached.List, nil
			}
		}

		list, total, err := s.newsRepo.List(regionID, pagination, req.CategoryID, req.Status, req.ListingType, "", req.MinPrice, req.MaxPrice, req.IsUrgent, req.Sort)
		if err != nil {
			return nil, nil, err
		}
		pagination.Total = total
		result := make([]dto.NewsInfo, 0, len(list))
		for i := range list {
			result = append(result, *toNewsInfo(&list[i]))
		}
		if req.ListingType == "" {
			_ = rediscache.SetJSON(ctx, newsCacheKeyList(regionID, req), newsListCache{Pagination: *pagination, List: result}, newsCacheTTL)
		}
		return pagination, result, nil
	}

	list, total, err := s.newsRepo.List(regionID, pagination, req.CategoryID, req.Status, req.ListingType, req.Keyword, req.MinPrice, req.MaxPrice, req.IsUrgent, req.Sort)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.NewsInfo, 0, len(list))
	for i := range list {
		result = append(result, *toNewsInfo(&list[i]))
	}
	return pagination, result, nil
}

// Like 点赞 toggle
func (s *newsService) Like(userID, newsID uint) (*dto.LikeResponse, error) {
	news, err := s.newsRepo.FindByID(newsID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNewsNotFound
		}
		return nil, err
	}

	exists, err := s.newsRepo.LikeExists(userID, newsID)
	if err != nil {
		return nil, err
	}

	if exists {
		_ = s.newsRepo.DeleteLike(userID, newsID)
		_ = s.newsRepo.DecrLikeCount(newsID)
		news.LikeCount--
		if news.LikeCount < 0 {
			news.LikeCount = 0
		}
		return &dto.LikeResponse{Liked: false, LikeCount: news.LikeCount}, nil
	}

	_ = s.newsRepo.CreateLike(&model.NewsLike{UserID: userID, NewsID: newsID})
	_ = s.newsRepo.IncrLikeCount(newsID)
	news.LikeCount++

	// 通知作者
	if hub := wspkg.GetHub(); hub != nil && news.AuthorID != 0 && news.AuthorID != userID {
		hub.SendToUser(news.AuthorID, &wspkg.Message{
			Type: wspkg.TypeLike,
			Data: wspkg.LikeNotification{
				NewsID:    newsID,
				NewsTitle: news.Title,
				Liked:     true,
				LikeCount: news.LikeCount,
			},
		})
		// 记录消息
		_ = s.newsRepo.CreateMessage(&model.Message{
			FromUserID: userID,
			ToUserID:   news.AuthorID,
			NewsID:     &newsID,
			Type:       "like",
			Content:    fmt.Sprintf("有人赞了你的信息「%s」", news.Title),
		})
	}
	return &dto.LikeResponse{Liked: true, LikeCount: news.LikeCount}, nil
}

func (s *newsService) LikeStatus(userID, newsID uint) (*dto.LikeResponse, error) {
	news, err := s.newsRepo.FindByID(newsID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNewsNotFound
		}
		return nil, err
	}
	liked, err := s.newsRepo.LikeExists(userID, newsID)
	if err != nil {
		return nil, err
	}
	return &dto.LikeResponse{Liked: liked, LikeCount: news.LikeCount}, nil
}

// ====== 收藏 ======

func (s *newsService) Fav(userID, newsID uint) (*dto.FavResponse, error) {
	news, err := s.newsRepo.FindByID(newsID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNewsNotFound
		}
		return nil, err
	}

	exists, err := s.newsRepo.FavExists(userID, newsID)
	if err != nil {
		return nil, err
	}
	if exists {
		_ = s.newsRepo.DeleteFav(userID, newsID)
		_ = s.newsRepo.DecrFavCount(newsID)
		news.FavCount--
		if news.FavCount < 0 {
			news.FavCount = 0
		}
		return &dto.FavResponse{Faved: false, FavCount: news.FavCount}, nil
	}

	_ = s.newsRepo.CreateFav(&model.NewsFavorite{UserID: userID, NewsID: newsID})
	_ = s.newsRepo.IncrFavCount(newsID)
	news.FavCount++
	return &dto.FavResponse{Faved: true, FavCount: news.FavCount}, nil
}

func (s *newsService) FavStatus(userID, newsID uint) (*dto.FavResponse, error) {
	news, err := s.newsRepo.FindByID(newsID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNewsNotFound
		}
		return nil, err
	}
	faved, err := s.newsRepo.FavExists(userID, newsID)
	if err != nil {
		return nil, err
	}
	return &dto.FavResponse{Faved: faved, FavCount: news.FavCount}, nil
}

// ====== 评论 ======

func (s *newsService) CreateComment(newsID uint, userID uint, userName string, avatar string, req *dto.CreateCommentRequest) (*dto.CommentInfo, error) {
	// 校验信息存在
	news, err := s.newsRepo.FindByID(newsID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNewsNotFound
		}
		return nil, err
	}

	comment := &model.NewsComment{
		NewsID:   newsID,
		UserID:   userID,
		UserName: userName,
		Avatar:   avatar,
		Content:  req.Content,
		ParentID: req.ParentID,
		ReplyTo:  req.ReplyTo,
		Status:   1,
	}

	if err := s.newsRepo.CreateComment(comment); err != nil {
		return nil, err
	}
	_ = s.newsRepo.IncrCommentCount(newsID)

	// 通知作者
	if news.AuthorID != 0 && news.AuthorID != userID {
		_ = s.newsRepo.CreateMessage(&model.Message{
			FromUserID: userID,
			ToUserID:   news.AuthorID,
			NewsID:     &newsID,
			Type:       "comment",
			Content:    fmt.Sprintf("%s 评论了你的信息「%s」：%s", userName, news.Title, req.Content),
		})
	}

	return toCommentInfo(comment), nil
}

func (s *newsService) ListComments(newsID uint, page, pageSize int) ([]dto.CommentInfo, int64, error) {
	list, total, err := s.newsRepo.ListComments(newsID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.CommentInfo, 0, len(list))
	for i := range list {
		result = append(result, *toCommentInfo(&list[i]))
	}
	return result, total, nil
}

func (s *newsService) DeleteComment(id uint, userID uint) error {
	return s.newsRepo.DeleteComment(id)
}

// ====== 消息 ======

func (s *newsService) ListMessages(userID uint, page, pageSize int) ([]dto.MessageInfo, int64, error) {
	list, total, err := s.newsRepo.ListMessages(userID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.MessageInfo, 0, len(list))
	for i := range list {
		result = append(result, *toMessageInfo(&list[i]))
	}
	return result, total, nil
}

func (s *newsService) UnreadCount(userID uint) (int64, error) {
	return s.newsRepo.UnreadCount(userID)
}

func (s *newsService) MarkRead(userID uint, ids []uint) error {
	return s.newsRepo.MarkRead(userID, ids)
}

// ====== 搜索 ======

func (s *newsService) Search(regionID uint, req *dto.NewsSearchRequest) (*utils.Pagination, []dto.NewsInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)

	if es.IsAvailable() && req.Keyword != "" {
		return s.searchByES(regionID, req, pagination)
	}
	list, total, err := s.newsRepo.List(regionID, pagination, req.CategoryID, 1, req.ListingType, req.Keyword, 0, 0, nil, "")
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.NewsInfo, 0, len(list))
	for i := range list {
		result = append(result, *toNewsInfo(&list[i]))
	}
	return pagination, result, nil
}

func (s *newsService) searchByES(regionID uint, req *dto.NewsSearchRequest, pagination *utils.Pagination) (*utils.Pagination, []dto.NewsInfo, error) {
	queryObj := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{
						"multi_match": map[string]interface{}{
							"query":  req.Keyword,
							"fields": []string{"title^3", "summary^2", "tags^2", "content"},
						},
					},
				},
				"filter": s.buildESFilters(regionID, req),
			},
		},
		"sort": []map[string]interface{}{
			{"published_at": map[string]string{"order": "desc"}},
			{"_score": map[string]string{"order": "desc"}},
		},
	}
	body, err := json.Marshal(queryObj)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal es query failed: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res, err := es.SearchByQuery(ctx, indexer.IndexName, string(body), pagination.Offset(), pagination.PageSize)
	if err != nil {
		list, total, dbErr := s.newsRepo.List(regionID, pagination, req.CategoryID, 1, req.ListingType, req.Keyword, 0, 0, nil, "")
		if dbErr != nil {
			return nil, nil, dbErr
		}
		pagination.Total = total
		result := make([]dto.NewsInfo, 0, len(list))
		for i := range list {
			result = append(result, *toNewsInfo(&list[i]))
		}
		return pagination, result, nil
	}
	pagination.Total = res.Total

	ids := make([]uint, 0, len(res.Hits))
	for _, hit := range res.Hits {
		if idVal, ok := hit["id"]; ok {
			var id uint
			switch v := idVal.(type) {
			case float64:
				id = uint(v)
			case json.Number:
				if n, err := v.Int64(); err == nil {
					id = uint(n)
				}
			}
			if id > 0 {
				ids = append(ids, id)
			}
		}
	}
	if len(ids) == 0 {
		return pagination, []dto.NewsInfo{}, nil
	}

	newsList, err := s.newsRepo.FindByIDs(ids)
	if err != nil {
		return nil, nil, err
	}
	newsMap := make(map[uint]*model.News, len(newsList))
	for i := range newsList {
		newsMap[newsList[i].ID] = &newsList[i]
	}
	result := make([]dto.NewsInfo, 0, len(ids))
	for _, id := range ids {
		if n, ok := newsMap[id]; ok {
			result = append(result, *toNewsInfo(n))
		}
	}
	return pagination, result, nil
}

func (s *newsService) buildESFilters(regionID uint, req *dto.NewsSearchRequest) []map[string]interface{} {
	filters := []map[string]interface{}{
		{"term": map[string]interface{}{"status": 1}},
	}
	if regionID > 0 {
		filters = append(filters, map[string]interface{}{
			"term": map[string]interface{}{"region_id": regionID},
		})
	}
	if req.CategoryID > 0 {
		filters = append(filters, map[string]interface{}{
			"term": map[string]interface{}{"category_id": req.CategoryID},
		})
	}
	if req.ListingType != "" {
		filters = append(filters, map[string]interface{}{
			"term": map[string]interface{}{"listing_type": req.ListingType},
		})
	}
	return filters
}