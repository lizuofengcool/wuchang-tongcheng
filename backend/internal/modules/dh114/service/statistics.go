// Package service 同城114业务逻辑层 - 统计
// 依据 v3.2.1 架构方案：日统计/商户统计/分类统计
// 对标大众点评/美团 数据分析
package service

import (
	"errors"
	"time"

	"wuchang-tongcheng/internal/modules/dh114/dto"
	"wuchang-tongcheng/internal/modules/dh114/model"
	"wuchang-tongcheng/internal/modules/dh114/repository"
)

var (
	ErrStatisticNotFound = errors.New("统计记录不存在")
	ErrStatDateInvalid   = errors.New("统计日期区间无效")
)

// StatisticService 统计业务接口
type StatisticService interface {
	// 查询
	ListByDateRange(startDate, endDate time.Time, statType string) ([]dto.StatisticInfo, error)
	ListByDh114(dh114ID uint, startDate, endDate time.Time) ([]dto.StatisticInfo, error)
	ListByCategory(categoryID uint, startDate, endDate time.Time) ([]dto.StatisticInfo, error)
	SumByDh114(dh114ID uint, startDate, endDate time.Time) (*dto.StatisticInfo, error)

	// 汇总查询
	HotBusiness(regionID uint, limit int) ([]dto.HotBusinessResponse, error)
	HotCategories(regionID uint, startDate, endDate time.Time) ([]dto.HotCategoryResponse, error)
	Overview(regionID uint, startDate, endDate time.Time) (*dto.OverviewResponse, error)

	// M 端管理
	Upsert(regionID uint, req *model.Dh114Statistic) error
}

type statisticService struct {
	repo repository.StatisticRepository
}

// NewStatisticService 创建统计 service 实例
func NewStatisticService(repo repository.StatisticRepository) StatisticService {
	return &statisticService{repo: repo}
}

// statTypeText 统计类型文本
func statTypeText(t string) string {
	switch t {
	case model.StatTypeDaily:
		return "日统计"
	case model.StatTypeBusiness:
		return "商户统计"
	case model.StatTypeCategory:
		return "分类统计"
	}
	return ""
}

// toStatisticInfo model -> dto
func toStatisticInfo(s *model.Dh114Statistic) *dto.StatisticInfo {
	info := &dto.StatisticInfo{
		ID:              s.ID,
		StatDate:        s.StatDate.Format("2006-01-02"),
		StatType:        s.StatType,
		StatTypeText:    statTypeText(s.StatType),
		Dh114ID:         s.Dh114ID,
		BusinessID:      s.BusinessID,
		CategoryID:      s.CategoryID,
		ViewCount:       s.ViewCount,
		FavCount:        s.FavCount,
		CallCount:       s.CallCount,
		ShareCount:      s.ShareCount,
		ContactCount:    s.ContactCount,
		VisitCount:      s.VisitCount,
		ReviewCount:     s.ReviewCount,
		NewReviewCount:  s.NewReviewCount,
		AvgRating:       s.AvgRating,
		GoodRate:        s.GoodRate,
		GroupbuySold:    s.GroupbuySold,
		GroupbuyAmount:  s.GroupbuyAmount,
		CouponIssued:    s.CouponIssued,
		CouponUsed:      s.CouponUsed,
		OrderCount:      s.OrderCount,
		OrderAmount:     s.OrderAmount,
		RegionID:        s.RegionID,
		CreatedAt:       s.CreatedAt,
		UpdatedAt:       s.UpdatedAt,
	}
	return info
}

// ListByDateRange 按日期区间列出统计
func (s *statisticService) ListByDateRange(startDate, endDate time.Time, statType string) ([]dto.StatisticInfo, error) {
	if endDate.Before(startDate) {
		return nil, ErrStatDateInvalid
	}
	list, err := s.repo.ListByDateRange(startDate, endDate, statType)
	if err != nil {
		return nil, err
	}
	infos := make([]dto.StatisticInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toStatisticInfo(&list[i]))
	}
	return infos, nil
}

// ListByDh114 按商户列出统计
func (s *statisticService) ListByDh114(dh114ID uint, startDate, endDate time.Time) ([]dto.StatisticInfo, error) {
	if endDate.Before(startDate) {
		return nil, ErrStatDateInvalid
	}
	list, err := s.repo.ListByDh114(dh114ID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	infos := make([]dto.StatisticInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toStatisticInfo(&list[i]))
	}
	return infos, nil
}

// ListByCategory 按分类列出统计
func (s *statisticService) ListByCategory(categoryID uint, startDate, endDate time.Time) ([]dto.StatisticInfo, error) {
	if endDate.Before(startDate) {
		return nil, ErrStatDateInvalid
	}
	list, err := s.repo.ListByCategory(categoryID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	infos := make([]dto.StatisticInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toStatisticInfo(&list[i]))
	}
	return infos, nil
}

// SumByDh114 商户统计汇总
func (s *statisticService) SumByDh114(dh114ID uint, startDate, endDate time.Time) (*dto.StatisticInfo, error) {
	if endDate.Before(startDate) {
		return nil, ErrStatDateInvalid
	}
	stat, err := s.repo.SumByDh114(dh114ID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	return toStatisticInfo(stat), nil
}

// HotBusiness 热门商户
func (s *statisticService) HotBusiness(regionID uint, limit int) ([]dto.HotBusinessResponse, error) {
	if limit <= 0 {
		limit = 10
	}
	list, err := s.repo.HotBusiness(regionID, limit)
	if err != nil {
		return nil, err
	}
	infos := make([]dto.HotBusinessResponse, 0, len(list))
	for _, h := range list {
		infos = append(infos, dto.HotBusinessResponse{
			Dh114ID:     h.Dh114ID,
			Title:       h.Title,
			ViewCount:   h.ViewCount,
			FavCount:    h.FavCount,
			CallCount:   h.CallCount,
			ReviewCount: h.ReviewCount,
			Rating:      h.Rating,
		})
	}
	return infos, nil
}

// HotCategories 热门分类
func (s *statisticService) HotCategories(regionID uint, startDate, endDate time.Time) ([]dto.HotCategoryResponse, error) {
	list, err := s.repo.HotCategories(regionID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	infos := make([]dto.HotCategoryResponse, 0, len(list))
	for _, h := range list {
		infos = append(infos, dto.HotCategoryResponse{
			CategoryID:    h.CategoryID,
			CategoryName:  h.CategoryName,
			BusinessCount: h.BusinessCount,
			TotalViews:    h.TotalViews,
			TotalReviews:  h.TotalReviews,
		})
	}
	return infos, nil
}

// Overview 总览统计
func (s *statisticService) Overview(regionID uint, startDate, endDate time.Time) (*dto.OverviewResponse, error) {
	_ = startDate
	_ = endDate
	stat, err := s.repo.Overview(regionID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	return &dto.OverviewResponse{
		TotalBusiness: stat.TotalBusiness,
		TotalView:     stat.TotalView,
		TotalFav:      stat.TotalFav,
		TotalCall:     stat.TotalCall,
		TotalReview:   stat.TotalReview,
		TotalGroupbuy: stat.TotalGroupbuy,
		TotalCoupon:   stat.TotalCoupon,
		TotalOrder:    stat.TotalOrder,
		TotalAmount:   stat.TotalAmount,
		AvgRating:     stat.AvgRating,
	}, nil
}

// Upsert 写入/更新统计（M 端定时任务调用）
func (s *statisticService) Upsert(regionID uint, req *model.Dh114Statistic) error {
	if req.RegionID == 0 {
		req.RegionID = regionID
	}
	if req.StatDate.IsZero() {
		req.StatDate = time.Now()
	}
	if req.StatType == "" {
		req.StatType = model.StatTypeDaily
	}
	return s.repo.Upsert(req)
}
