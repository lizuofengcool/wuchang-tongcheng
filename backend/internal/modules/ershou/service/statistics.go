// Package service 数据统计业务逻辑层
// 依据 v3.2.1 架构方案：M 端运营数据 + C 端卖家数据
package service

import (
	"time"

	"wuchang-tongcheng/internal/modules/ershou/dto"
	"wuchang-tongcheng/internal/modules/ershou/model"

	"gorm.io/gorm"
)

// StatisticsService 数据统计业务接口
type StatisticsService interface {
	// Overview 平台总览（M端）
	Overview(regionID uint) (*dto.StatisticsResponse, error)
	// SellerOverview 卖家数据总览（C端）
	SellerOverview(userID uint) (*dto.SellerStatisticsResponse, error)
	// HotItems 热门商品 Top N
	HotItems(limit int) ([]dto.HotItemResponse, error)
	// PriceTrend 价格趋势（近 N 天）
	PriceTrend(brandID uint, days int) (*dto.PriceTrendResponse, error)
}

type statisticsService struct {
	db *gorm.DB
}

// NewStatisticsService 创建数据统计 service 实例
func NewStatisticsService(db *gorm.DB) StatisticsService {
	return &statisticsService{db: db}
}

// Overview 平台总览
func (s *statisticsService) Overview(regionID uint) (*dto.StatisticsResponse, error) {
	resp := &dto.StatisticsResponse{}
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	itemQuery := s.db.Model(&model.Ershou{})
	if regionID > 0 {
		itemQuery = itemQuery.Where("region_id = ?", regionID)
	}
	// 商品总数（已发布）
	itemQuery.Where("status = ?", model.StatusPublished).Count(&resp.TotalItems)
	// 今日新增商品
	s.db.Model(&model.Ershou{}).Where("created_at >= ?", todayStart).Count(&resp.TodayNewItems)
	// 活跃卖家数（去重）
	s.db.Model(&model.Ershou{}).Where("status = ?", model.StatusPublished).
		Distinct("user_id").Count(&resp.ActiveSellers)

	// 订单统计
	orderQuery := s.db.Model(&model.ErshouOrder{})
	if regionID > 0 {
		orderQuery = orderQuery.Where("region_id = ?", regionID)
	}
	orderQuery.Count(&resp.TotalOrders)
	s.db.Model(&model.ErshouOrder{}).Where("created_at >= ?", todayStart).Count(&resp.TodayOrders)

	// 成交额
	type amtResult struct {
		Total float64
	}
	var totalAmt amtResult
	s.db.Model(&model.ErshouOrder{}).
		Where("status IN ?", []int{model.OrderStatusCompleted}).
		Select("COALESCE(SUM(total_amount), 0) AS total").
		Scan(&totalAmt)
	resp.TotalAmount = totalAmt.Total

	var todayAmt amtResult
	s.db.Model(&model.ErshouOrder{}).
		Where("status IN ? AND created_at >= ?", []int{model.OrderStatusCompleted}, todayStart).
		Select("COALESCE(SUM(total_amount), 0) AS total").
		Scan(&todayAmt)
	resp.TodayAmount = todayAmt.Total

	// 完成率
	var completedCount, totalCount int64
	s.db.Model(&model.ErshouOrder{}).Where("status = ?", model.OrderStatusCompleted).Count(&completedCount)
	s.db.Model(&model.ErshouOrder{}).Count(&totalCount)
	if totalCount > 0 {
		resp.CompletedRate = float64(completedCount) / float64(totalCount) * 100
	}

	// 退款率
	var refundedCount int64
	s.db.Model(&model.ErshouOrder{}).Where("status = ?", model.OrderStatusRefunded).Count(&refundedCount)
	if totalCount > 0 {
		resp.RefundRate = float64(refundedCount) / float64(totalCount) * 100
	}

	// 平均价格
	var avgPriceResult struct {
		AvgPrice float64
	}
	s.db.Model(&model.Ershou{}).Where("status = ?", model.StatusPublished).
		Select("COALESCE(AVG(price), 0) AS avg_price").
		Scan(&avgPriceResult)
	resp.AvgPrice = avgPriceResult.AvgPrice

	return resp, nil
}

// SellerOverview 卖家数据总览
func (s *statisticsService) SellerOverview(userID uint) (*dto.SellerStatisticsResponse, error) {
	resp := &dto.SellerStatisticsResponse{UserID: userID}

	// 发布总数
	s.db.Model(&model.Ershou{}).Where("user_id = ?", userID).Count(&resp.TotalItems)
	// 售出总数（status = 已售出）
	s.db.Model(&model.Ershou{}).Where("user_id = ? AND status = ?", userID, model.StatusSold).Count(&resp.SoldItems)
	// 订单总数（作为卖家）
	s.db.Model(&model.ErshouOrder{}).Where("seller_id = ?", userID).Count(&resp.TotalOrders)
	// 完成订单数
	s.db.Model(&model.ErshouOrder{}).Where("seller_id = ? AND status = ?", userID, model.OrderStatusCompleted).Count(&resp.CompletedOrders)
	// 成交额
	type amtResult struct {
		Total float64
	}
	var amt amtResult
	s.db.Model(&model.ErshouOrder{}).
		Where("seller_id = ? AND status = ?", userID, model.OrderStatusCompleted).
		Select("COALESCE(SUM(total_amount), 0) AS total").
		Scan(&amt)
	resp.TotalAmount = amt.Total
	// 平均评分
	var rateResult struct {
		AvgRate float64
	}
	s.db.Model(&model.ErshouReview{}).
		Where("reviewee_id = ?", userID).
		Select("COALESCE(AVG(rating), 0) AS avg_rate").
		Scan(&rateResult)
	resp.AvgRating = rateResult.AvgRate
	// 粉丝数
	s.db.Model(&model.ErshouShopFollower{}).
		Joins("JOIN ers_shops ON ers_shops.id = ers_shop_followers.shop_id").
		Where("ers_shops.user_id = ?", userID).Count(&resp.Followers)
	// 转化率 = 完成订单数 / 商品总数
	if resp.TotalItems > 0 {
		resp.ConversionRate = float64(resp.CompletedOrders) / float64(resp.TotalItems) * 100
	}
	return resp, nil
}

// HotItems 热门商品 Top N（按浏览量+收藏量+订单数综合排序）
func (s *statisticsService) HotItems(limit int) ([]dto.HotItemResponse, error) {
	if limit <= 0 {
		limit = 10
	}
	var list []dto.HotItemResponse
	err := s.db.Model(&model.Ershou{}).
		Select("id AS ershou_id, title, cover_image, price, view_count, fav_count, message_count AS order_count").
		Where("status = ?", model.StatusPublished).
		Order("view_count DESC, fav_count DESC, id DESC").
		Limit(limit).
		Scan(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

// PriceTrend 价格趋势（按品牌/分类的近 N 天平均价）
func (s *statisticsService) PriceTrend(brandID uint, days int) (*dto.PriceTrendResponse, error) {
	if days <= 0 {
		days = 30
	}
	if days > 365 {
		days = 365
	}
	resp := &dto.PriceTrendResponse{}
	if brandID > 0 {
		// 取品牌名
		var b model.ErshouBrand
		if err := s.db.First(&b, brandID).Error; err == nil {
			resp.Brand = b.Name
		}
	}
	now := time.Now()
	startDate := now.AddDate(0, 0, -days)

	type trendRow struct {
		Date  string
		Price float64
	}
	var rows []trendRow
	query := s.db.Model(&model.Ershou{}).
		Select("TO_CHAR(created_at, 'YYYY-MM-DD') AS date, COALESCE(AVG(price), 0) AS price").
		Where("created_at >= ? AND status = ?", startDate, model.StatusPublished).
		Group("TO_CHAR(created_at, 'YYYY-MM-DD')").
		Order("date ASC")
	if brandID > 0 {
		query = query.Where("brand_id = ?", brandID)
	}
	if err := query.Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, r := range rows {
		resp.Dates = append(resp.Dates, r.Date)
		resp.Prices = append(resp.Prices, r.Price)
	}
	return resp, nil
}
