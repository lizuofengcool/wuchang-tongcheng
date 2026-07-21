// Package repository 同城商城数据访问层 - 统计
package repository

import (
	"time"

	"wuchang-tongcheng/internal/modules/mall/model"

	"gorm.io/gorm"
)

// StatisticRepository 统计仓储接口
type StatisticRepository interface {
	Create(s *model.Statistic) error
	Update(id uint, fields map[string]interface{}) error
	FindByDateTypeTarget(statDate time.Time, statType string, shopID, productID, categoryID uint) (*model.Statistic, error)
	Upsert(s *model.Statistic) error

	ListByDateRange(startDate, endDate time.Time, statType string) ([]model.Statistic, error)
	ListByShop(shopID uint, startDate, endDate time.Time) ([]model.Statistic, error)
	ListByProduct(productID uint, startDate, endDate time.Time) ([]model.Statistic, error)
	ListByCategory(categoryID uint, startDate, endDate time.Time) ([]model.Statistic, error)

	Summary(opts StatisticSummaryOptions) (*StatisticSummaryResult, error)
	HotProducts(regionID uint, limit int) ([]HotProductStat, error)
	HotShops(regionID uint, limit int) ([]HotShopStat, error)
	HotCategories(regionID uint, startDate, endDate time.Time) ([]HotCategoryStat, error)
	Overview(regionID uint, startDate, endDate time.Time) (*MallOverviewStat, error)
}

// StatisticSummaryOptions 统计汇总选项
type StatisticSummaryOptions struct {
	RegionID   uint
	ShopID     uint
	ProductID  uint
	CategoryID uint
	StartDate  time.Time
	EndDate    time.Time
	GroupBy    string
}

// StatisticSummaryResult 统计汇总结果
type StatisticSummaryResult struct {
	TotalOrderCount      int64   `gorm:"column:total_order_count" json:"total_order_count"`
	TotalOrderAmount     float64 `gorm:"column:total_order_amount" json:"total_order_amount"`
	TotalPaidAmount      float64 `gorm:"column:total_paid_amount" json:"total_paid_amount"`
	TotalRefundAmount    float64 `gorm:"column:total_refund_amount" json:"total_refund_amount"`
	TotalViewCount       int64   `gorm:"column:total_view_count" json:"total_view_count"`
	TotalFavoriteCount   int64   `gorm:"column:total_favorite_count" json:"total_favorite_count"`
	TotalSalesCount      int64   `gorm:"column:total_sales_count" json:"total_sales_count"`
	TotalReviewCount     int64   `gorm:"column:total_review_count" json:"total_review_count"`
	AvgRating            float64 `gorm:"column:avg_rating" json:"avg_rating"`
	TotalNewBuyerCount   int64   `gorm:"column:total_new_buyer_count" json:"total_new_buyer_count"`
	TotalActiveBuyerCount int64  `gorm:"column:total_active_buyer_count" json:"total_active_buyer_count"`
	TotalRepurchaseCount int64   `gorm:"column:total_repurchase_count" json:"total_repurchase_count"`
	AvgConversionRate    float64 `gorm:"column:avg_conversion_rate" json:"avg_conversion_rate"`
}

// HotProductStat 热门商品统计
type HotProductStat struct {
	ProductID   uint    `gorm:"column:product_id" json:"product_id"`
	ProductName string  `gorm:"column:product_name" json:"product_name"`
	SalesCount  int64   `gorm:"column:sales_count" json:"sales_count"`
	ViewCount   int64   `gorm:"column:view_count" json:"view_count"`
	OrderCount  int64   `gorm:"column:order_count" json:"order_count"`
	OrderAmount float64 `gorm:"column:order_amount" json:"order_amount"`
}

// HotShopStat 热门店铺统计
type HotShopStat struct {
	ShopID      uint    `gorm:"column:shop_id" json:"shop_id"`
	ShopName    string  `gorm:"column:shop_name" json:"shop_name"`
	OrderCount  int64   `gorm:"column:order_count" json:"order_count"`
	OrderAmount float64 `gorm:"column:order_amount" json:"order_amount"`
	SalesCount  int64   `gorm:"column:sales_count" json:"sales_count"`
	Rating      float64 `gorm:"column:rating" json:"rating"`
}

// HotCategoryStat 热门分类统计
type HotCategoryStat struct {
	CategoryID    uint    `gorm:"column:category_id" json:"category_id"`
	CategoryName  string  `gorm:"column:category_name" json:"category_name"`
	ProductCount  int64   `gorm:"column:product_count" json:"product_count"`
	OrderCount    int64   `gorm:"column:order_count" json:"order_count"`
	OrderAmount   float64 `gorm:"column:order_amount" json:"order_amount"`
}

// MallOverviewStat 平台总览统计
type MallOverviewStat struct {
	TotalShop       int64   `gorm:"column:total_shop" json:"total_shop"`
	TotalProduct    int64   `gorm:"column:total_product" json:"total_product"`
	TotalOrder      int64   `gorm:"column:total_order" json:"total_order"`
	TotalAmount     float64 `gorm:"column:total_amount" json:"total_amount"`
	TotalPaidAmount float64 `gorm:"column:total_paid_amount" json:"total_paid_amount"`
	TotalRefundAmount float64 `gorm:"column:total_refund_amount" json:"total_refund_amount"`
	TotalUser       int64   `gorm:"column:total_user" json:"total_user"`
	TotalReview     int64   `gorm:"column:total_review" json:"total_review"`
	AvgRating       float64 `gorm:"column:avg_rating" json:"avg_rating"`
}

type statisticRepository struct {
	db *gorm.DB
}

// NewStatisticRepository 创建统计仓储实例
func NewStatisticRepository(db *gorm.DB) StatisticRepository {
	return &statisticRepository{db: db}
}

func (r *statisticRepository) Create(s *model.Statistic) error {
	return r.db.Create(s).Error
}

func (r *statisticRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Statistic{}).Where("id = ?", id).Updates(fields).Error
}

func (r *statisticRepository) FindByDateTypeTarget(statDate time.Time, statType string, shopID, productID, categoryID uint) (*model.Statistic, error) {
	var s model.Statistic
	if err := r.db.Where("stat_date = ? AND stat_type = ? AND shop_id = ? AND product_id = ? AND category_id = ?",
		statDate, statType, shopID, productID, categoryID).First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *statisticRepository) Upsert(s *model.Statistic) error {
	// PostgreSQL ON CONFLICT 实现幂等 upsert
	// 唯一索引为 (stat_date, stat_type, shop_id, product_id, category_id)
	return r.db.Exec(`
		INSERT INTO mall_statistics (
			region_id, stat_date, stat_type, shop_id, product_id, category_id,
			order_count, order_amount, paid_order_count, paid_order_amount,
			cancelled_order_count, refund_count, refund_amount,
			view_count, favorite_count, cart_count, sales_count,
			review_count, new_review_count, avg_rating, good_rate,
			new_buyer_count, active_buyer_count, repurchase_count, conversion_rate,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
		ON CONFLICT (stat_date, stat_type, shop_id, product_id, category_id) DO UPDATE SET
			order_count = EXCLUDED.order_count,
			order_amount = EXCLUDED.order_amount,
			paid_order_count = EXCLUDED.paid_order_count,
			paid_order_amount = EXCLUDED.paid_order_amount,
			cancelled_order_count = EXCLUDED.cancelled_order_count,
			refund_count = EXCLUDED.refund_count,
			refund_amount = EXCLUDED.refund_amount,
			view_count = EXCLUDED.view_count,
			favorite_count = EXCLUDED.favorite_count,
			cart_count = EXCLUDED.cart_count,
			sales_count = EXCLUDED.sales_count,
			review_count = EXCLUDED.review_count,
			new_review_count = EXCLUDED.new_review_count,
			avg_rating = EXCLUDED.avg_rating,
			good_rate = EXCLUDED.good_rate,
			new_buyer_count = EXCLUDED.new_buyer_count,
			active_buyer_count = EXCLUDED.active_buyer_count,
			repurchase_count = EXCLUDED.repurchase_count,
			conversion_rate = EXCLUDED.conversion_rate,
			updated_at = NOW()
	`,
		s.RegionID, s.StatDate, s.StatType, s.ShopID, s.ProductID, s.CategoryID,
		s.OrderCount, s.OrderAmount, s.PaidOrderCount, s.PaidOrderAmount,
		s.CancelledOrderCount, s.RefundCount, s.RefundAmount,
		s.ViewCount, s.FavoriteCount, s.CartCount, s.SalesCount,
		s.ReviewCount, s.NewReviewCount, s.AvgRating, s.GoodRate,
		s.NewBuyerCount, s.ActiveBuyerCount, s.RepurchaseCount, s.ConversionRate,
	).Error
}

func (r *statisticRepository) ListByDateRange(startDate, endDate time.Time, statType string) ([]model.Statistic, error) {
	var list []model.Statistic
	q := r.db.Where("stat_date BETWEEN ? AND ?", startDate, endDate)
	if statType != "" {
		q = q.Where("stat_type = ?", statType)
	}
	if err := q.Order("stat_date ASC, id ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *statisticRepository) ListByShop(shopID uint, startDate, endDate time.Time) ([]model.Statistic, error) {
	var list []model.Statistic
	if err := r.db.Where("shop_id = ? AND stat_type = ? AND stat_date BETWEEN ? AND ?", shopID, model.StatTypeShop, startDate, endDate).
		Order("stat_date ASC, id ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *statisticRepository) ListByProduct(productID uint, startDate, endDate time.Time) ([]model.Statistic, error) {
	var list []model.Statistic
	if err := r.db.Where("product_id = ? AND stat_type = ? AND stat_date BETWEEN ? AND ?", productID, model.StatTypeProduct, startDate, endDate).
		Order("stat_date ASC, id ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *statisticRepository) ListByCategory(categoryID uint, startDate, endDate time.Time) ([]model.Statistic, error) {
	var list []model.Statistic
	if err := r.db.Where("category_id = ? AND stat_type = ? AND stat_date BETWEEN ? AND ?", categoryID, model.StatTypeCategory, startDate, endDate).
		Order("stat_date ASC, id ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *statisticRepository) Summary(opts StatisticSummaryOptions) (*StatisticSummaryResult, error) {
	var result StatisticSummaryResult
	query := r.db.Model(&model.Statistic{}).Select(`
		COALESCE(SUM(order_count), 0) AS total_order_count,
		COALESCE(SUM(order_amount), 0) AS total_order_amount,
		COALESCE(SUM(paid_order_amount), 0) AS total_paid_amount,
		COALESCE(SUM(refund_amount), 0) AS total_refund_amount,
		COALESCE(SUM(view_count), 0) AS total_view_count,
		COALESCE(SUM(favorite_count), 0) AS total_favorite_count,
		COALESCE(SUM(sales_count), 0) AS total_sales_count,
		COALESCE(SUM(review_count), 0) AS total_review_count,
		COALESCE(AVG(avg_rating), 0) AS avg_rating,
		COALESCE(SUM(new_buyer_count), 0) AS total_new_buyer_count,
		COALESCE(SUM(active_buyer_count), 0) AS total_active_buyer_count,
		COALESCE(SUM(repurchase_count), 0) AS total_repurchase_count,
		COALESCE(AVG(conversion_rate), 0) AS avg_conversion_rate
	`)
	if opts.RegionID > 0 {
		query = query.Where("region_id = ?", opts.RegionID)
	}
	if opts.ShopID > 0 {
		query = query.Where("shop_id = ?", opts.ShopID)
	}
	if opts.ProductID > 0 {
		query = query.Where("product_id = ?", opts.ProductID)
	}
	if opts.CategoryID > 0 {
		query = query.Where("category_id = ?", opts.CategoryID)
	}
	if !opts.StartDate.IsZero() {
		query = query.Where("stat_date >= ?", opts.StartDate)
	}
	if !opts.EndDate.IsZero() {
		query = query.Where("stat_date <= ?", opts.EndDate)
	}
	if err := query.Scan(&result).Error; err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *statisticRepository) HotProducts(regionID uint, limit int) ([]HotProductStat, error) {
	if limit <= 0 {
		limit = 10
	}
	var list []HotProductStat
	query := r.db.Table("mall_statistics AS s").
		Select("s.product_id, p.name AS product_name, SUM(s.sales_count) AS sales_count, SUM(s.view_count) AS view_count, SUM(s.order_count) AS order_count, SUM(s.order_amount) AS order_amount").
		Joins("LEFT JOIN mall_products AS p ON p.id = s.product_id").
		Where("s.stat_type = ? AND s.product_id > 0", model.StatTypeProduct).
		Group("s.product_id, p.name").
		Order("sales_count DESC").
		Limit(limit)
	if regionID > 0 {
		query = query.Where("s.region_id = ?", regionID)
	}
	if err := query.Scan(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *statisticRepository) HotShops(regionID uint, limit int) ([]HotShopStat, error) {
	if limit <= 0 {
		limit = 10
	}
	var list []HotShopStat
	query := r.db.Table("mall_statistics AS s").
		Select("s.shop_id, sh.shop_name, SUM(s.order_count) AS order_count, SUM(s.order_amount) AS order_amount, SUM(s.sales_count) AS sales_count, MAX(sh.rating) AS rating").
		Joins("LEFT JOIN mall_shops AS sh ON sh.id = s.shop_id").
		Where("s.stat_type = ? AND s.shop_id > 0", model.StatTypeShop).
		Group("s.shop_id, sh.shop_name").
		Order("order_amount DESC").
		Limit(limit)
	if regionID > 0 {
		query = query.Where("s.region_id = ?", regionID)
	}
	if err := query.Scan(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *statisticRepository) HotCategories(regionID uint, startDate, endDate time.Time) ([]HotCategoryStat, error) {
	var list []HotCategoryStat
	query := r.db.Table("mall_statistics AS s").
		Select("s.category_id, c.name AS category_name, COUNT(DISTINCT s.product_id) AS product_count, SUM(s.order_count) AS order_count, SUM(s.order_amount) AS order_amount").
		Joins("LEFT JOIN mall_categories AS c ON c.id = s.category_id").
		Where("s.stat_type = ? AND s.category_id > 0", model.StatTypeCategory).
		Where("s.stat_date BETWEEN ? AND ?", startDate, endDate).
		Group("s.category_id, c.name").
		Order("order_amount DESC")
	if regionID > 0 {
		query = query.Where("s.region_id = ?", regionID)
	}
	if err := query.Scan(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *statisticRepository) Overview(regionID uint, startDate, endDate time.Time) (*MallOverviewStat, error) {
	var result MallOverviewStat

	// 平台总览
	query := r.db.Table("mall_statistics AS s").Select(`
		COALESCE(SUM(CASE WHEN s.stat_type = 'shop' AND s.stat_date = ? THEN 1 ELSE 0 END), 0) AS total_shop,
		COALESCE(SUM(CASE WHEN s.stat_type = 'product' AND s.stat_date = ? THEN 1 ELSE 0 END), 0) AS total_product,
		COALESCE(SUM(s.order_count), 0) AS total_order,
		COALESCE(SUM(s.order_amount), 0) AS total_amount,
		COALESCE(SUM(s.paid_order_amount), 0) AS total_paid_amount,
		COALESCE(SUM(s.refund_amount), 0) AS total_refund_amount,
		0 AS total_user,
		COALESCE(SUM(s.review_count), 0) AS total_review,
		COALESCE(AVG(s.avg_rating), 0) AS avg_rating
	`, endDate, endDate)
	if regionID > 0 {
		query = query.Where("s.region_id = ?", regionID)
	}
	query = query.Where("s.stat_date BETWEEN ? AND ?", startDate, endDate)
	if err := query.Scan(&result).Error; err != nil {
		return nil, err
	}
	return &result, nil
}
