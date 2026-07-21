// Package repository 同城商城数据访问层 - 物流
package repository

import (
	"wuchang-tongcheng/internal/modules/mall/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// LogisticsRepository 物流仓储接口
type LogisticsRepository interface {
	Create(l *model.Logistics) error
	FindByID(id uint) (*model.Logistics, error)
	FindByOrderID(orderID uint) (*model.Logistics, error)
	FindByTrackingNo(trackingNo string) (*model.Logistics, error)
	Update(l *model.Logistics) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	List(opts LogisticsListOptions, pagination *utils.Pagination) ([]model.Logistics, int64, error)
	ListByUser(userID uint, pagination *utils.Pagination) ([]model.Logistics, int64, error)
	ListByShop(shopID uint, pagination *utils.Pagination) ([]model.Logistics, int64, error)

	UpdateStatus(id uint, status int, fields map[string]interface{}) error
	UpdateTraces(id uint, traces model.JSONB) error
}

// LogisticsListOptions 物流列表过滤条件
type LogisticsListOptions struct {
	OrderID     uint
	OrderNo     string
	ShopID      uint
	UserID      uint
	TrackingNo  string
	CompanyCode string
	Status      *int
	Keyword     string
	StartDate   string
	EndDate     string
	RegionID    uint
}

type logisticsRepository struct {
	db *gorm.DB
}

// NewLogisticsRepository 创建物流仓储实例
func NewLogisticsRepository(db *gorm.DB) LogisticsRepository {
	return &logisticsRepository{db: db}
}

func (r *logisticsRepository) Create(l *model.Logistics) error {
	return r.db.Create(l).Error
}

func (r *logisticsRepository) FindByID(id uint) (*model.Logistics, error) {
	var l model.Logistics
	if err := r.db.First(&l, id).Error; err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *logisticsRepository) FindByOrderID(orderID uint) (*model.Logistics, error) {
	var l model.Logistics
	if err := r.db.Where("order_id = ?", orderID).First(&l).Error; err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *logisticsRepository) FindByTrackingNo(trackingNo string) (*model.Logistics, error) {
	var l model.Logistics
	if err := r.db.Where("tracking_no = ?", trackingNo).First(&l).Error; err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *logisticsRepository) Update(l *model.Logistics) error {
	return r.db.Save(l).Error
}

func (r *logisticsRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Logistics{}).Where("id = ?", id).Updates(fields).Error
}

func (r *logisticsRepository) Delete(id uint) error {
	return r.db.Delete(&model.Logistics{}, id).Error
}

func (r *logisticsRepository) List(opts LogisticsListOptions, pagination *utils.Pagination) ([]model.Logistics, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Logistics
	var total int64

	query := r.db.Model(&model.Logistics{})
	if opts.OrderID > 0 {
		query = query.Where("order_id = ?", opts.OrderID)
	}
	if opts.OrderNo != "" {
		query = query.Where("order_no = ?", opts.OrderNo)
	}
	if opts.ShopID > 0 {
		query = query.Where("shop_id = ?", opts.ShopID)
	}
	if opts.UserID > 0 {
		query = query.Where("user_id = ?", opts.UserID)
	}
	if opts.TrackingNo != "" {
		query = query.Where("tracking_no = ?", opts.TrackingNo)
	}
	if opts.CompanyCode != "" {
		query = query.Where("company_code = ?", opts.CompanyCode)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.RegionID > 0 {
		query = query.Where("region_id = ?", opts.RegionID)
	}
	if opts.StartDate != "" {
		query = query.Where("created_at >= ?", opts.StartDate)
	}
	if opts.EndDate != "" {
		query = query.Where("created_at <= ?", opts.EndDate)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("tracking_no ILIKE ? OR order_no ILIKE ? OR company ILIKE ? OR receiver_name ILIKE ?", like, like, like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *logisticsRepository) ListByUser(userID uint, pagination *utils.Pagination) ([]model.Logistics, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Logistics
	var total int64

	query := r.db.Model(&model.Logistics{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *logisticsRepository) ListByShop(shopID uint, pagination *utils.Pagination) ([]model.Logistics, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Logistics
	var total int64

	query := r.db.Model(&model.Logistics{}).Where("shop_id = ?", shopID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *logisticsRepository) UpdateStatus(id uint, status int, fields map[string]interface{}) error {
	if fields == nil {
		fields = map[string]interface{}{}
	}
	fields["status"] = status
	return r.db.Model(&model.Logistics{}).Where("id = ?", id).Updates(fields).Error
}

func (r *logisticsRepository) UpdateTraces(id uint, traces model.JSONB) error {
	return r.db.Model(&model.Logistics{}).Where("id = ?", id).
		UpdateColumn("traces", traces).Error
}
