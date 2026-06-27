// Package service 同城头条业务逻辑层
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
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrNewsNotFound     = errors.New("头条不存在")
	ErrNewsNoPermission = errors.New("无权操作此头条")
)

// NewsService 头条业务逻辑接口
type NewsService interface {
	Create(regionID uint, authorID uint, authorName string, req *dto.CreateNewsRequest) (*dto.NewsInfo, error)
	Update(id uint, operatorID uint, req *dto.UpdateNewsRequest) error
	Delete(id uint, operatorID uint) error
	GetByID(id uint) (*dto.NewsInfo, error)
	List(regionID uint, req *dto.NewsListRequest) (*utils.Pagination, []dto.NewsInfo, error)
	// Search 全文检索（ES 优先，ES 不可用时降级到 DB LIKE）
	Search(regionID uint, req *dto.NewsSearchRequest) (*utils.Pagination, []dto.NewsInfo, error)
	// Like 点赞（幂等：已点赞则取消，未点赞则点赞）
	Like(userID, newsID uint) (*dto.LikeResponse, error)
	// LikeStatus 查询当前用户对该头条的点赞状态
	LikeStatus(userID, newsID uint) (*dto.LikeResponse, error)
}

type newsService struct {
	newsRepo repository.NewsRepository
	indexer  indexer.Indexer
}

// NewNewsService 创建头条服务
// idx 为 nil 时使用 NoopIndexer（不索引）
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
		Summary:     n.Summary,
		AuthorID:    n.AuthorID,
		AuthorName:  n.AuthorName,
		CategoryID:  n.CategoryID,
		Tags:        n.Tags,
		ViewCount:   n.ViewCount,
		LikeCount:   n.LikeCount,
		Status:      n.Status,
		PublishedAt: n.PublishedAt,
		CreatedAt:   n.CreatedAt,
	}
}

// Create 创建头条
func (s *newsService) Create(regionID uint, authorID uint, authorName string, req *dto.CreateNewsRequest) (*dto.NewsInfo, error) {
	news := &model.News{
		Title:      req.Title,
		Content:    req.Content,
		CoverImage: req.CoverImage,
		Summary:    req.Summary,
		AuthorID:   authorID,
		AuthorName: authorName,
		CategoryID: req.CategoryID,
		Tags:       req.Tags,
		Status:     req.Status,
	}
	news.RegionID = regionID

	// 如果状态为发布，设置发布时间
	if req.Status == 1 {
		now := time.Now()
		news.PublishedAt = &now
	}

	if err := s.newsRepo.Create(news); err != nil {
		return nil, err
	}
	// 触发索引同步（fire-and-forget）
	s.indexer.OnIndex(news)
	return toNewsInfo(news), nil
}

// Update 更新头条
func (s *newsService) Update(id uint, operatorID uint, req *dto.UpdateNewsRequest) error {
	news, err := s.newsRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNewsNotFound
		}
		return err
	}

	// 仅作者本人可编辑
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
	if req.Summary != "" {
		fields["summary"] = req.Summary
	}
	fields["category_id"] = req.CategoryID
	if req.Tags != "" {
		fields["tags"] = req.Tags
	}

	// 状态变更：若改为发布，且之前未发布过，则设置发布时间
	if req.Status == 1 && news.Status != 1 {
		now := time.Now()
		fields["status"] = 1
		fields["published_at"] = &now
	} else if req.Status >= 0 && req.Status <= 2 {
		fields["status"] = req.Status
	}

	if len(fields) == 0 {
		return nil
	}
	if err := s.newsRepo.UpdateFields(id, fields); err != nil {
		return err
	}
	// 重新查询最新数据触发索引同步（fire-and-forget）
	if updated, err := s.newsRepo.FindByID(id); err == nil {
		s.indexer.OnIndex(updated)
	}
	return nil
}

// Delete 删除头条
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
	// 触发索引删除（fire-and-forget）
	s.indexer.OnDelete(id)
	return nil
}

// GetByID 获取头条详情（同时增加浏览量）
func (s *newsService) GetByID(id uint) (*dto.NewsInfo, error) {
	news, err := s.newsRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNewsNotFound
		}
		return nil, err
	}

	// 异步增加浏览量（简化为同步）
	_ = s.newsRepo.IncrViewCount(id)
	news.ViewCount++

	return toNewsInfo(news), nil
}

// List 头条列表
func (s *newsService) List(regionID uint, req *dto.NewsListRequest) (*utils.Pagination, []dto.NewsInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)

	list, total, err := s.newsRepo.List(regionID, pagination, req.CategoryID, req.Status, req.Keyword)
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

// Like 点赞（幂等 toggle：已点赞则取消并 liked=false，未点赞则点赞并 liked=true）
func (s *newsService) Like(userID, newsID uint) (*dto.LikeResponse, error) {
	// 校验头条是否存在
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
		// 已点赞 -> 取消
		if err := s.newsRepo.DeleteLike(userID, newsID); err != nil {
			return nil, err
		}
		_ = s.newsRepo.DecrLikeCount(newsID)
		news.LikeCount--
		if news.LikeCount < 0 {
			news.LikeCount = 0
		}
		return &dto.LikeResponse{Liked: false, LikeCount: news.LikeCount}, nil
	}

	// 未点赞 -> 点赞
	if err := s.newsRepo.CreateLike(&model.NewsLike{UserID: userID, NewsID: newsID}); err != nil {
		return nil, err
	}
	_ = s.newsRepo.IncrLikeCount(newsID)
	news.LikeCount++
	return &dto.LikeResponse{Liked: true, LikeCount: news.LikeCount}, nil
}

// LikeStatus 查询当前用户对该头条的点赞状态
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

// Search 全文检索
// 优先走 ES（匹配 title/content/summary/tags 多字段），ES 不可用时降级到 DB LIKE
// 已发布（status=1）+ 地区隔离 + 可选 category 过滤
func (s *newsService) Search(regionID uint, req *dto.NewsSearchRequest) (*utils.Pagination, []dto.NewsInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)

	// ES 可用：走 ES multi_match 检索，再用命中的 ID 回查 DB 取完整字段
	if es.IsAvailable() && req.Keyword != "" {
		return s.searchByES(regionID, req, pagination)
	}
	// 降级：走 DB（复用 List 的 keyword LIKE 语义）
	list, total, err := s.newsRepo.List(regionID, pagination, req.CategoryID, 1, req.Keyword)
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

// searchByES 走 ES multi_match 检索，再用命中 ID 回查 DB
func (s *newsService) searchByES(regionID uint, req *dto.NewsSearchRequest, pagination *utils.Pagination) (*utils.Pagination, []dto.NewsInfo, error) {
	// 构建 ES query DSL：multi_match 跨 4 个字段
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
		// ES 出错时降级到 DB
		list, total, dbErr := s.newsRepo.List(regionID, pagination, req.CategoryID, 1, req.Keyword)
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

	// ES 返回的 _source 中包含 id 字段，提取后回查 DB
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

	// 按 ES 返回顺序重排 DB 结果（保持相关性排序）
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

// buildESFilters 构建 ES bool filter 子句
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
	return filters
}
