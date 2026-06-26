// Package service 同城头条业务逻辑层
package service

import (
	"errors"
	"time"

	"wuchang-tongcheng/internal/modules/news/dto"
	"wuchang-tongcheng/internal/modules/news/model"
	"wuchang-tongcheng/internal/modules/news/repository"
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
}

type newsService struct {
	newsRepo repository.NewsRepository
}

// NewNewsService 创建头条服务
func NewNewsService(newsRepo repository.NewsRepository) NewsService {
	return &newsService{newsRepo: newsRepo}
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
	return s.newsRepo.UpdateFields(id, fields)
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
	return s.newsRepo.Delete(id)
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
