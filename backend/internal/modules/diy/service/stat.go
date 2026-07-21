// Package service DIY 前端页面中台业务逻辑层 - 统计（stat 子域）
package service

import (
	"errors"
	"time"

	"wuchang-tongcheng/internal/modules/diy/dto"
	"wuchang-tongcheng/internal/modules/diy/model"
	"wuchang-tongcheng/internal/modules/diy/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// StatService 统计业务接口
type StatService interface {
	// 记录
	RecordView(pageID uint) error
	RecordClick(pageID uint) error
	RecordConversion(pageID uint) error

	// 查询
	ListByPageID(pageID uint, page, pageSize int) (*utils.Pagination, []dto.StatInfo, error)
	ListByDateRange(req *dto.StatSummaryRequest) (*utils.Pagination, []dto.StatInfo, error)

	// 汇总
	SumByPageID(pageID uint) (*dto.StatSummary, error)
	SumByDateRange(req *dto.StatSummaryRequest) (*dto.StatSummary, error)
}

type statService struct {
	repo repository.StatRepository
}

// NewStatService 创建统计 service 实例
func NewStatService(repo repository.StatRepository) StatService {
	return &statService{repo: repo}
}

// toStatInfo model -> dto
func toStatInfo(s *model.PageStat) *dto.StatInfo {
	return &dto.StatInfo{
		ID:              s.ID,
		PageID:          s.PageID,
		ViewCount:       s.ViewCount,
		ClickCount:      s.ClickCount,
		ConversionCount: s.ConversionCount,
		StatDate:        s.StatDate,
		CreatedAt:       s.CreatedAt,
		UpdatedAt:       s.UpdatedAt,
	}
}

// RecordView 记录浏览
func (s *statService) RecordView(pageID uint) error {
	return s.repo.IncrViewCount(pageID, time.Now())
}

// RecordClick 记录点击
func (s *statService) RecordClick(pageID uint) error {
	return s.repo.IncrClickCount(pageID, time.Now())
}

// RecordConversion 记录转化
func (s *statService) RecordConversion(pageID uint) error {
	return s.repo.IncrConversionCount(pageID, time.Now())
}

// ListByPageID 按页面 ID 列出统计
func (s *statService) ListByPageID(pageID uint, page, pageSize int) (*utils.Pagination, []dto.StatInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByPageID(pageID, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.StatInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toStatInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// ListByDateRange 按日期范围列出统计
func (s *statService) ListByDateRange(req *dto.StatSummaryRequest) (*utils.Pagination, []dto.StatInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.StatRangeOptions{
		PageID:    req.PageID,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
	}
	list, total, err := s.repo.ListByDateRange(opts, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.StatInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toStatInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// SumByPageID 按页面 ID 汇总
func (s *statService) SumByPageID(pageID uint) (*dto.StatSummary, error) {
	viewCount, clickCount, conversionCount, err := s.repo.SumByPageID(pageID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrStatNotFound
		}
		return nil, err
	}
	return &dto.StatSummary{
		PageID:          pageID,
		TotalView:       viewCount,
		TotalClick:      clickCount,
		TotalConversion: conversionCount,
		DailyList:       []dto.StatInfo{},
	}, nil
}

// SumByDateRange 按日期范围汇总
func (s *statService) SumByDateRange(req *dto.StatSummaryRequest) (*dto.StatSummary, error) {
	opts := repository.StatRangeOptions{
		PageID:    req.PageID,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
	}
	viewCount, clickCount, conversionCount, err := s.repo.SumByDateRange(opts)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrStatNotFound
		}
		return nil, err
	}
	return &dto.StatSummary{
		PageID:          req.PageID,
		TotalView:       viewCount,
		TotalClick:      clickCount,
		TotalConversion: conversionCount,
		DailyList:       []dto.StatInfo{},
	}, nil
}
