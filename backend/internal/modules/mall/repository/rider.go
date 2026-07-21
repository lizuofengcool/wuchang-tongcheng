// Package repository 同城商城数据访问层 - 骑手
package repository

import (
	"wuchang-tongcheng/internal/modules/mall/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// RiderRepository 骑手仓储接口
type RiderRepository interface {
	Create(r *model.Rider) error
	FindByID(id uint) (*model.Rider, error)
	FindByUserID(userID uint) (*model.Rider, error)
	Update(r *model.Rider) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	List(regionID uint, pagination *utils.Pagination, opts RiderListOptions) ([]model.Rider, int64, error)
	AdminList(pagination *utils.Pagination, opts RiderAdminListOptions) ([]model.Rider, int64, error)

	// 统计 / 计数器
	IncrTotalOrders(id uint, earnings float64) error
	UpdateOnlineStatus(id uint, status int) error
	UpdateStatus(id uint, status int, reason string) error

	// 在线骑手列表（抢单池筛选用）
	ListOnlineByRegion(regionID uint) ([]model.Rider, error)
}

// RiderListOptions C 端骑手列表过滤条件
type RiderListOptions struct {
	Keyword      string
	Status       *int
	OnlineStatus *int
	ShopID       uint
	UserID       uint
}

// RiderAdminListOptions M 端管理列表过滤条件
type RiderAdminListOptions struct {
	RegionID     uint
	UserID       uint
	ShopID       uint
	Status       *int
	OnlineStatus *int
	Keyword      string
}

type riderRepository struct {
	db *gorm.DB
}

// NewRiderRepository 创建骑手仓储实例
func NewRiderRepository(db *gorm.DB) RiderRepository {
	return &riderRepository{db: db}
}

func (r *riderRepository) Create(rider *model.Rider) error {
	return r.db.Create(rider).Error
}

func (r *riderRepository) FindByID(id uint) (*model.Rider, error) {
	var rider model.Rider
	if err := r.db.First(&rider, id).Error; err != nil {
		return nil, err
	}
	return &rider, nil
}

func (r *riderRepository) FindByUserID(userID uint) (*model.Rider, error) {
	var rider model.Rider
	if err := r.db.Where("user_id = ?", userID).First(&rider).Error; err != nil {
		return nil, err
	}
	return &rider, nil
}

func (r *riderRepository) Update(rider *model.Rider) error {
	return r.db.Save(rider).Error
}

func (r *riderRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Rider{}).Where("id = ?", id).Updates(fields).Error
}

func (r *riderRepository) Delete(id uint) error {
	return r.db.Delete(&model.Rider{}, id).Error
}

func (r *riderRepository) List(regionID uint, pagination *utils.Pagination, opts RiderListOptions) ([]model.Rider, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Rider
	var total int64

	query := r.db.Model(&model.Rider{})
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.OnlineStatus != nil {
		query = query.Where("online_status = ?", *opts.OnlineStatus)
	}
	if opts.ShopID > 0 {
		query = query.Where("shop_id = ?", opts.ShopID)
	}
	if opts.UserID > 0 {
		query = query.Where("user_id = ?", opts.UserID)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("real_name ILIKE ? OR phone ILIKE ?", like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *riderRepository) AdminList(pagination *utils.Pagination, opts RiderAdminListOptions) ([]model.Rider, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Rider
	var total int64

	query := r.db.Model(&model.Rider{})
	if opts.RegionID > 0 {
		query = query.Where("region_id = ?", opts.RegionID)
	}
	if opts.UserID > 0 {
		query = query.Where("user_id = ?", opts.UserID)
	}
	if opts.ShopID > 0 {
		query = query.Where("shop_id = ?", opts.ShopID)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.OnlineStatus != nil {
		query = query.Where("online_status = ?", *opts.OnlineStatus)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("real_name ILIKE ? OR phone ILIKE ? OR id_card ILIKE ?", like, like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ===== 计数器 =====

func (r *riderRepository) IncrTotalOrders(id uint, earnings float64) error {
	return r.db.Model(&model.Rider{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"total_orders":   gorm.Expr("total_orders + 1"),
			"total_earnings": gorm.Expr("total_earnings + ?", earnings),
		}).Error
}

func (r *riderRepository) UpdateOnlineStatus(id uint, status int) error {
	return r.db.Model(&model.Rider{}).Where("id = ?", id).
		UpdateColumn("online_status", status).Error
}

func (r *riderRepository) UpdateStatus(id uint, status int, reason string) error {
	fields := map[string]interface{}{
		"status": status,
	}
	if reason != "" {
		fields["audit_reason"] = reason
	}
	return r.db.Model(&model.Rider{}).Where("id = ?", id).Updates(fields).Error
}

func (r *riderRepository) ListOnlineByRegion(regionID uint) ([]model.Rider, error) {
	var list []model.Rider
	query := r.db.Model(&model.Rider{}).Where("status = ?", model.RiderStatusApproved)
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if err := query.Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
