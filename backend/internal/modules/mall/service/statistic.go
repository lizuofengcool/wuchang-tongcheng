// Package service 同城商城业务逻辑层 - 统计
package service

import (
	"time"

	"wuchang-tongcheng/internal/modules/mall/dto"
	"wuchang-tongcheng/internal/modules/mall/model"
	"wuchang-tongcheng/internal/modules/mall/repository"
	"wuchang-tongcheng/internal/pkg/utils"
)

// StatisticService 统计业务接口
type StatisticService interface {
	// 写入（M 端定时任务）
	Upsert(regionID uint, req *dto.UpsertStatisticRequest) error

	// 查询
	List(req *dto.StatisticListRequest) (*utils.Pagination, []dto.StatisticInfo, error)
	Summary(regionID uint, req *dto.StatisticSummaryRequest) (*dto.StatisticSummary, error)

	// 热榜
	HotProducts(regionID uint, limit int) ([]repository.HotProductStat, error)
	HotShops(regionID uint, limit int) ([]repository.HotShopStat, error)
	HotCategories(regionID uint, startDate, endDate string) ([]repository.HotCategoryStat, error)

	// 平台总览
	Overview(regionID uint, startDate, endDate string) (*repository.MallOverviewStat, error)
}

type statisticService struct {
	repo repository.StatisticRepository
}

// NewStatisticService 创建统计 service 实例
func NewStatisticService(repo repository.StatisticRepository) StatisticService {
	return &statisticService{repo: repo}
}

// parseDate 解析 YYYY-MM-DD 格式日期
func parseDate(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Time{}
	}
	return t
}

// toStatisticInfo model -> dto
func toStatisticInfo(s *model.Statistic) *dto.StatisticInfo {
	info := &dto.StatisticInfo{
		ID:                  s.ID,
		StatDate:            s.StatDate.Format("2006-01-02"),
		StatType:            s.StatType,
		ShopID:              s.ShopID,
		ProductID:           s.ProductID,
		OrderCount:          s.OrderCount,
		OrderAmount:         s.OrderAmount,
		PaidOrderCount:      s.PaidOrderCount,
		PaidOrderAmount:     s.PaidOrderAmount,
		CancelledOrderCount: s.CancelledOrderCount,
		RefundCount:         s.RefundCount,
		RefundAmount:        s.RefundAmount,
		ViewCount:           s.ViewCount,
		FavoriteCount:       s.FavoriteCount,
		CartCount:           s.CartCount,
		SalesCount:          s.SalesCount,
		ReviewCount:         s.ReviewCount,
		NewReviewCount:      s.NewReviewCount,
		AvgRating:           s.AvgRating,
		GoodRate:            s.GoodRate,
		NewBuyerCount:       s.NewBuyerCount,
		ActiveBuyerCount:    s.ActiveBuyerCount,
		RepurchaseCount:     s.RepurchaseCount,
		ConversionRate:      s.ConversionRate,
		RegionID:            s.RegionID,
		CreatedAt:           s.CreatedAt,
		UpdatedAt:           s.UpdatedAt,
	}
	if s.CategoryID != nil {
		info.CategoryID = *s.CategoryID
	}
	return info
}

// Upsert 写入/更新统计（M 端定时任务调用，使用 ON CONFLICT 实现幂等）
func (s *statisticService) Upsert(regionID uint, req *dto.UpsertStatisticRequest) error {
	statDate, err := time.Parse("2006-01-02", req.StatDate)
	if err != nil {
		return err
	}

	stat := &model.Statistic{
		StatDate:              statDate,
		StatType:              req.StatType,
		ShopID:                req.ShopID,
		ProductID:             req.ProductID,
		OrderCount:            req.OrderCount,
		OrderAmount:           req.OrderAmount,
		PaidOrderCount:        req.PaidOrderCount,
		PaidOrderAmount:       req.PaidOrderAmount,
		CancelledOrderCount:   req.CancelledOrderCount,
		RefundCount:           req.RefundCount,
		RefundAmount:          req.RefundAmount,
		ViewCount:             req.ViewCount,
		FavoriteCount:         req.FavoriteCount,
		CartCount:             req.CartCount,
		SalesCount:            req.SalesCount,
		ReviewCount:           req.ReviewCount,
		NewReviewCount:        req.NewReviewCount,
		AvgRating:             req.AvgRating,
		GoodRate:              req.GoodRate,
		NewBuyerCount:         req.NewBuyerCount,
		ActiveBuyerCount:      req.ActiveBuyerCount,
		RepurchaseCount:       req.RepurchaseCount,
		ConversionRate:        req.ConversionRate,
	}
	stat.RegionID = regionID
	if req.CategoryID > 0 {
		catID := req.CategoryID
		stat.CategoryID = &catID
	}

	return s.repo.Upsert(stat)
}

// List 统计列表
func (s *statisticService) List(req *dto.StatisticListRequest) (*utils.Pagination, []dto.StatisticInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	startDate := parseDate(req.StartDate)
	endDate := parseDate(req.EndDate)

	var list []model.Statistic
	var total int64

	// 根据 ShopID/ProductID/CategoryID 选择查询方式
	if req.ShopID > 0 {
		items, err := s.repo.ListByShop(req.ShopID, startDate, endDate)
		if err != nil {
			return nil, nil, err
		}
		list = items
		total = int64(len(items))
	} else if req.ProductID > 0 {
		items, err := s.repo.ListByProduct(req.ProductID, startDate, endDate)
		if err != nil {
			return nil, nil, err
		}
		list = items
		total = int64(len(items))
	} else if req.CategoryID > 0 {
		items, err := s.repo.ListByCategory(req.CategoryID, startDate, endDate)
		if err != nil {
			return nil, nil, err
		}
		list = items
		total = int64(len(items))
	} else {
		items, err := s.repo.ListByDateRange(startDate, endDate, req.StatType)
		if err != nil {
			return nil, nil, err
		}
		list = items
		total = int64(len(items))
	}

	// 简化分页（已经在 repo 层做了过滤，这里直接返回全部）
	infos := make([]dto.StatisticInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toStatisticInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// Summary 统计汇总
func (s *statisticService) Summary(regionID uint, req *dto.StatisticSummaryRequest) (*dto.StatisticSummary, error) {
	opts := repository.StatisticSummaryOptions{
		RegionID:   regionID,
		ShopID:     req.ShopID,
		StartDate:  parseDate(req.StartDate),
		EndDate:    parseDate(req.EndDate),
		GroupBy:    req.GroupBy,
	}
	result, err := s.repo.Summary(opts)
	if err != nil {
		return nil, err
	}

	summary := &dto.StatisticSummary{
		TotalOrderCount:       result.TotalOrderCount,
		TotalOrderAmount:      result.TotalOrderAmount,
		TotalPaidAmount:       result.TotalPaidAmount,
		TotalRefundAmount:     result.TotalRefundAmount,
		TotalViewCount:        result.TotalViewCount,
		TotalFavoriteCount:    result.TotalFavoriteCount,
		TotalSalesCount:       result.TotalSalesCount,
		TotalReviewCount:      result.TotalReviewCount,
		AvgRating:             result.AvgRating,
		TotalNewBuyerCount:    result.TotalNewBuyerCount,
		TotalActiveBuyerCount: result.TotalActiveBuyerCount,
		TotalRepurchaseCount:  result.TotalRepurchaseCount,
		AvgConversionRate:     result.AvgConversionRate,
	}
	return summary, nil
}

// HotProducts 热销商品榜
func (s *statisticService) HotProducts(regionID uint, limit int) ([]repository.HotProductStat, error) {
	return s.repo.HotProducts(regionID, limit)
}

// HotShops 热门店铺榜
func (s *statisticService) HotShops(regionID uint, limit int) ([]repository.HotShopStat, error) {
	return s.repo.HotShops(regionID, limit)
}

// HotCategories 热门分类榜
func (s *statisticService) HotCategories(regionID uint, startDate, endDate string) ([]repository.HotCategoryStat, error) {
	sd := parseDate(startDate)
	ed := parseDate(endDate)
	if sd.IsZero() {
		sd = time.Now().AddDate(0, 0, -30)
	}
	if ed.IsZero() {
		ed = time.Now()
	}
	return s.repo.HotCategories(regionID, sd, ed)
}

// Overview 平台总览
func (s *statisticService) Overview(regionID uint, startDate, endDate string) (*repository.MallOverviewStat, error) {
	sd := parseDate(startDate)
	ed := parseDate(endDate)
	if sd.IsZero() {
		sd = time.Now().AddDate(0, 0, -30)
	}
	if ed.IsZero() {
		ed = time.Now()
	}
	return s.repo.Overview(regionID, sd, ed)
}
