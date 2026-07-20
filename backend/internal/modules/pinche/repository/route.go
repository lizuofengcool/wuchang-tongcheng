// Package repository 同城拼车出行数据访问层 - 路线
package repository

import (
	"wuchang-tongcheng/internal/modules/pinche/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// RouteListOptions 路线列表过滤条件
type RouteListOptions struct {
	UserID    uint
	IsCommon  *bool
	Status    *int
	Keyword   string
}

// RouteRepository 路线仓储接口
type RouteRepository interface {
	Create(r *model.PincheRoute) error
	FindByID(id uint) (*model.PincheRoute, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(regionID uint, pagination *utils.Pagination, opts RouteListOptions) ([]model.PincheRoute, int64, error)
	ListByUser(userID uint, pagination *utils.Pagination) ([]model.PincheRoute, int64, error)
	ListCommon(regionID uint, pagination *utils.Pagination) ([]model.PincheRoute, int64, error)
	IncrUseCount(id uint) error
}

type routeRepository struct {
	db *gorm.DB
}

// NewRouteRepository 创建路线仓储实例
func NewRouteRepository(db *gorm.DB) RouteRepository {
	return &routeRepository{db: db}
}

func (r *routeRepository) Create(rt *model.PincheRoute) error {
	return r.db.Create(rt).Error
}

func (r *routeRepository) FindByID(id uint) (*model.PincheRoute, error) {
	var rt model.PincheRoute
	if err := r.db.First(&rt, id).Error; err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *routeRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.PincheRoute{}).Where("id = ?", id).Updates(fields).Error
}

func (r *routeRepository) Delete(id uint) error {
	return r.db.Delete(&model.PincheRoute{}, id).Error
}

func (r *routeRepository) List(regionID uint, pagination *utils.Pagination, opts RouteListOptions) ([]model.PincheRoute, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.PincheRoute
	var total int64

	query := r.db.Model(&model.PincheRoute{})
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if opts.UserID > 0 {
		query = query.Where("user_id = ?", opts.UserID)
	}
	if opts.IsCommon != nil {
		query = query.Where("is_common = ?", *opts.IsCommon)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("route_name ILIKE ? OR origin_address ILIKE ? OR destination_address ILIKE ?", like, like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("use_count DESC, created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *routeRepository) ListByUser(userID uint, pagination *utils.Pagination) ([]model.PincheRoute, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.PincheRoute
	var total int64

	query := r.db.Model(&model.PincheRoute{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("use_count DESC, created_at DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *routeRepository) ListCommon(regionID uint, pagination *utils.Pagination) ([]model.PincheRoute, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.PincheRoute
	var total int64

	query := r.db.Model(&model.PincheRoute{}).Where("is_common = ?", true)
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("use_count DESC, created_at DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *routeRepository) IncrUseCount(id uint) error {
	return r.db.Model(&model.PincheRoute{}).Where("id = ?", id).
		UpdateColumn("use_count", gorm.Expr("use_count + 1")).Error
}
