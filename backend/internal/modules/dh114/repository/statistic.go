// Package repository 同城114数据访问层 - 统计
// 日统计/商户统计/分类统计
package repository

import (
	"time"

	"wuchang-tongcheng/internal/modules/dh114/model"

	"gorm.io/gorm"
)

// StatisticRepository 统计仓储接口
type StatisticRepository interface {
	Create(s *model.Dh114Statistic) error
	Update(id uint, fields map[string]interface{}) error
	FindByDateTypeTarget(statDate time.Time, statType string, dh114ID uint) (*model.Dh114Statistic, error)
	ListByDateRange(startDate, endDate time.Time, statType string) ([]model.Dh114Statistic, error)
	ListByDh114(dh114ID uint, startDate, endDate time.Time) ([]model.Dh114Statistic, error)
	ListByCategory(categoryID uint, startDate, endDate time.Time) ([]model.Dh114Statistic, error)
	Upsert(s *model.Dh114Statistic) error
	// 汇总查询
	SumByDh114(dh114ID uint, startDate, endDate time.Time) (*model.Dh114Statistic, error)
	HotBusiness(regionID uint, limit int) ([]HotBusinessStat, error)
	HotCategories(regionID uint, startDate, endDate time.Time) ([]HotCategoryStat, error)
	Overview(regionID uint, startDate, endDate time.Time) (*OverviewStat, error)
}

// HotBusinessStat 热门商户统计
type HotBusinessStat struct {
	Dh114ID     uint    `gorm:"column:dh114_id" json:"dh114_id"`
	Title       string  `gorm:"column:title" json:"title"`
	ViewCount   int64   `gorm:"column:view_count" json:"view_count"`
	FavCount    int64   `gorm:"column:fav_count" json:"fav_count"`
	CallCount   int64   `gorm:"column:call_count" json:"call_count"`
	ReviewCount int64   `gorm:"column:review_count" json:"review_count"`
	Rating      float64 `gorm:"column:rating" json:"rating"`
}

// HotCategoryStat 热门分类统计
type HotCategoryStat struct {
	CategoryID   uint   `gorm:"column:category_id" json:"category_id"`
	CategoryName string `gorm:"column:category_name" json:"category_name"`
	BusinessCount int   `gorm:"column:business_count" json:"business_count"`
	TotalViews    int64  `gorm:"column:total_views" json:"total_views"`
	TotalReviews  int64  `gorm:"column:total_reviews" json:"total_reviews"`
}

// OverviewStat 总览统计
type OverviewStat struct {
	TotalBusiness  int64   `gorm:"column:total_business" json:"total_business"`
	TotalView      int64   `gorm:"column:total_view" json:"total_view"`
	TotalFav       int64   `gorm:"column:total_fav" json:"total_fav"`
	TotalCall      int64   `gorm:"column:total_call" json:"total_call"`
	TotalReview    int64   `gorm:"column:total_review" json:"total_review"`
	TotalGroupbuy  int64   `gorm:"column:total_groupbuy" json:"total_groupbuy"`
	TotalCoupon    int64   `gorm:"column:total_coupon" json:"total_coupon"`
	TotalOrder     int64   `gorm:"column:total_order" json:"total_order"`
	TotalAmount    float64 `gorm:"column:total_amount" json:"total_amount"`
	AvgRating      float64 `gorm:"column:avg_rating" json:"avg_rating"`
}

type statisticRepository struct {
	db *gorm.DB
}

// NewStatisticRepository 创建统计仓储实例
func NewStatisticRepository(db *gorm.DB) StatisticRepository {
	return &statisticRepository{db: db}
}

func (r *statisticRepository) Create(s *model.Dh114Statistic) error {
	return r.db.Create(s).Error
}

func (r *statisticRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Dh114Statistic{}).Where("id = ?", id).Updates(fields).Error
}

func (r *statisticRepository) FindByDateTypeTarget(statDate time.Time, statType string, dh114ID uint) (*model.Dh114Statistic, error) {
	var s model.Dh114Statistic
	if err := r.db.Where("stat_date = ? AND stat_type = ? AND dh114_id = ?", statDate, statType, dh114ID).
		First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *statisticRepository) ListByDateRange(startDate, endDate time.Time, statType string) ([]model.Dh114Statistic, error) {
	var list []model.Dh114Statistic
	q := r.db.Where("stat_date BETWEEN ? AND ?", startDate, endDate)
	if statType != "" {
		q = q.Where("stat_type = ?", statType)
	}
	if err := q.Order("stat_date ASC, id ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *statisticRepository) ListByDh114(dh114ID uint, startDate, endDate time.Time) ([]model.Dh114Statistic, error) {
	var list []model.Dh114Statistic
	if err := r.db.Where("dh114_id = ? AND stat_type = ? AND stat_date BETWEEN ? AND ?", dh114ID, model.StatTypeBusiness, startDate, endDate).
		Order("stat_date ASC, id ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *statisticRepository) ListByCategory(categoryID uint, startDate, endDate time.Time) ([]model.Dh114Statistic, error) {
	var list []model.Dh114Statistic
	if err := r.db.Where("category_id = ? AND stat_type = ? AND stat_date BETWEEN ? AND ?", categoryID, model.StatTypeCategory, startDate, endDate).
		Order("stat_date ASC, id ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *statisticRepository) Upsert(s *model.Dh114Statistic) error {
	// PostgreSQL ON CONFLICT 实现幂等 upsert
	return r.db.Exec(`INSERT INTO dh114_statistics 
		(region_id, stat_date, stat_type, dh114_id, business_id, category_id,
		view_count, fav_count, call_count, share_count, contact_count, visit_count,
		review_count, new_review_count, avg_rating, good_rate,
		groupbuy_sold, groupbuy_amount, coupon_issued, coupon_used, order_count, order_amount,
		created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
		ON CONFLICT (stat_date, stat_type, dh114_id) DO UPDATE SET
		view_count = EXCLUDED.view_count,
		fav_count = EXCLUDED.fav_count,
		call_count = EXCLUDED.call_count,
		share_count = EXCLUDED.share_count,
		contact_count = EXCLUDED.contact_count,
		visit_count = EXCLUDED.visit_count,
		review_count = EXCLUDED.review_count,
		new_review_count = EXCLUDED.new_review_count,
		avg_rating = EXCLUDED.avg_rating,
		good_rate = EXCLUDED.good_rate,
		groupbuy_sold = EXCLUDED.groupbuy_sold,
		groupbuy_amount = EXCLUDED.groupbuy_amount,
		coupon_issued = EXCLUDED.coupon_issued,
		coupon_used = EXCLUDED.coupon_used,
		order_count = EXCLUDED.order_count,
		order_amount = EXCLUDED.order_amount,
		updated_at = NOW()`,
		s.RegionID, s.StatDate, s.StatType, s.Dh114ID, s.BusinessID, s.CategoryID,
		s.ViewCount, s.FavCount, s.CallCount, s.ShareCount, s.ContactCount, s.VisitCount,
		s.ReviewCount, s.NewReviewCount, s.AvgRating, s.GoodRate,
		s.GroupbuySold, s.GroupbuyAmount, s.CouponIssued, s.CouponUsed, s.OrderCount, s.OrderAmount,
	).Error
}

func (r *statisticRepository) SumByDh114(dh114ID uint, startDate, endDate time.Time) (*model.Dh114Statistic, error) {
	var s model.Dh114Statistic
	err := r.db.Model(&model.Dh114Statistic{}).
		Select("COALESCE(SUM(view_count),0) AS view_count, COALESCE(SUM(fav_count),0) AS fav_count, "+
			"COALESCE(SUM(call_count),0) AS call_count, COALESCE(SUM(share_count),0) AS share_count, "+
			"COALESCE(SUM(contact_count),0) AS contact_count, COALESCE(SUM(visit_count),0) AS visit_count, "+
			"COALESCE(SUM(review_count),0) AS review_count, COALESCE(SUM(new_review_count),0) AS new_review_count, "+
			"COALESCE(AVG(avg_rating),0) AS avg_rating, COALESCE(AVG(good_rate),0) AS good_rate, "+
			"COALESCE(SUM(groupbuy_sold),0) AS groupbuy_sold, COALESCE(SUM(groupbuy_amount),0) AS groupbuy_amount, "+
			"COALESCE(SUM(coupon_issued),0) AS coupon_issued, COALESCE(SUM(coupon_used),0) AS coupon_used, "+
			"COALESCE(SUM(order_count),0) AS order_count, COALESCE(SUM(order_amount),0) AS order_amount").
		Where("dh114_id = ? AND stat_type = ? AND stat_date BETWEEN ? AND ?", dh114ID, model.StatTypeBusiness, startDate, endDate).
		Scan(&s).Error
	return &s, err
}

func (r *statisticRepository) HotBusiness(regionID uint, limit int) ([]HotBusinessStat, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	var list []HotBusinessStat
	q := r.db.Table("dh114s AS d").
		Select("d.id AS dh114_id, d.title AS title, d.view_count AS view_count, d.fav_count AS fav_count, d.call_count AS call_count, d.review_count AS review_count, d.rating AS rating").
		Where("d.deleted_at IS NULL AND d.status = ? AND d.audit_status = ?", model.StatusPublished, model.AuditApproved)
	if regionID > 0 {
		q = q.Where("d.region_id = ?", regionID)
	}
	if err := q.Order("d.view_count DESC, d.fav_count DESC, d.id DESC").
		Limit(limit).Scan(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *statisticRepository) HotCategories(regionID uint, startDate, endDate time.Time) ([]HotCategoryStat, error) {
	var list []HotCategoryStat
	q := r.db.Table("dh114_categories AS c").
		Select("c.id AS category_id, c.name AS category_name, COUNT(d.id) AS business_count, COALESCE(SUM(d.view_count),0) AS total_views, COALESCE(SUM(d.review_count),0) AS total_reviews").
		Joins("LEFT JOIN dh114s AS d ON d.category_id = c.id AND d.deleted_at IS NULL AND d.status = 1").
		Where("c.status = ?", model.CategoryStatusPublished).
		Group("c.id, c.name").
		Order("total_views DESC, business_count DESC, c.id DESC")
	if regionID > 0 {
		q = q.Where("d.region_id = ? OR d.id IS NULL", regionID)
	}
	if err := q.Scan(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *statisticRepository) Overview(regionID uint, startDate, endDate time.Time) (*OverviewStat, error) {
	var s OverviewStat
	q := r.db.Table("dh114s AS d").
		Select("COUNT(d.id) AS total_business, COALESCE(SUM(d.view_count),0) AS total_view, COALESCE(SUM(d.fav_count),0) AS total_fav, COALESCE(SUM(d.call_count),0) AS total_call, COALESCE(SUM(d.review_count),0) AS total_review, COALESCE(AVG(d.rating),0) AS avg_rating").
		Where("d.deleted_at IS NULL AND d.status = ? AND d.audit_status = ?", model.StatusPublished, model.AuditApproved)
	if regionID > 0 {
		q = q.Where("d.region_id = ?", regionID)
	}
	if err := q.Scan(&s).Error; err != nil {
		return nil, err
	}

	// 团购/优惠券/订单统计
	var tradeStat struct {
		TotalGroupbuy int64   `gorm:"column:total_groupbuy"`
		TotalCoupon   int64   `gorm:"column:total_coupon"`
		TotalOrder    int64   `gorm:"column:total_order"`
		TotalAmount   float64 `gorm:"column:total_amount"`
	}
	r.db.Table("dh114_groupbuys AS g").
		Select("COUNT(g.id) AS total_groupbuy, 0 AS total_coupon, 0 AS total_order, COALESCE(SUM(g.sold_count * g.groupbuy_price),0) AS total_amount").
		Where("g.deleted_at IS NULL AND g.status = ?", model.GroupbuyStatusPublished).
		Scan(&tradeStat)
	s.TotalGroupbuy = tradeStat.TotalGroupbuy
	s.TotalAmount = tradeStat.TotalAmount
	return &s, nil
}
