// Package repository 同城车辆买卖数据访问层 - 车况检测
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
package repository

import (
	"wuchang-tongcheng/internal/modules/car/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// InspectionRepository 车况检测仓储接口
type InspectionRepository interface {
	Create(i *model.CarInspection) error
	FindByID(id uint) (*model.CarInspection, error)
	FindByInspectionNo(inspectionNo string) (*model.CarInspection, error)
	FindByCarID(carID uint) (*model.CarInspection, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	// 列表（C 端，地区隔离）
	List(regionID uint, pagination *utils.Pagination, opts InspectionListOptions) ([]model.CarInspection, int64, error)
	// 管理后台列表（M 端，可跨地区）
	AdminList(pagination *utils.Pagination, opts InspectionAdminListOptions) ([]model.CarInspection, int64, error)
	// 检测师自己的检测单
	ListByInspector(inspectorID uint, pagination *utils.Pagination) ([]model.CarInspection, int64, error)

	// 统计
	CountByStatus(regionID uint, status int) (int64, error)
	CountByInspector(inspectorID uint) (int64, error)
}

// InspectionListOptions C 端检测单列表过滤条件
type InspectionListOptions struct {
	CarID           uint
	ListingID       uint
	InspectorID     uint
	InspectionType  string
	ConditionLevel  string
	HasAccident     *bool
	Status          *int
	Keyword         string
}

// InspectionAdminListOptions M 端检测单列表过滤条件
type InspectionAdminListOptions struct {
	RegionID        uint
	CarID           uint
	InspectorID     uint
	InspectionType  string
	ConditionLevel  string
	HasAccident     *bool
	Status          *int
	Keyword         string
}

type inspectionRepository struct {
	db *gorm.DB
}

// NewInspectionRepository 创建车况检测仓储实例
func NewInspectionRepository(db *gorm.DB) InspectionRepository {
	return &inspectionRepository{db: db}
}

// ===== CRUD =====

func (r *inspectionRepository) Create(i *model.CarInspection) error {
	return r.db.Create(i).Error
}

func (r *inspectionRepository) FindByID(id uint) (*model.CarInspection, error) {
	var i model.CarInspection
	if err := r.db.First(&i, id).Error; err != nil {
		return nil, err
	}
	return &i, nil
}

func (r *inspectionRepository) FindByInspectionNo(inspectionNo string) (*model.CarInspection, error) {
	var i model.CarInspection
	if err := r.db.Where("inspection_no = ?", inspectionNo).First(&i).Error; err != nil {
		return nil, err
	}
	return &i, nil
}

func (r *inspectionRepository) FindByCarID(carID uint) (*model.CarInspection, error) {
	var i model.CarInspection
	if err := r.db.Where("car_id = ?", carID).Order("id DESC").First(&i).Error; err != nil {
		return nil, err
	}
	return &i, nil
}

func (r *inspectionRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.CarInspection{}).Where("id = ?", id).Updates(fields).Error
}

func (r *inspectionRepository) Delete(id uint) error {
	return r.db.Delete(&model.CarInspection{}, id).Error
}

// ===== 列表查询 =====

func (r *inspectionRepository) List(regionID uint, pagination *utils.Pagination, opts InspectionListOptions) ([]model.CarInspection, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.CarInspection
	var total int64

	query := r.db.Model(&model.CarInspection{})

	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if opts.CarID > 0 {
		query = query.Where("car_id = ?", opts.CarID)
	}
	if opts.ListingID > 0 {
		query = query.Where("listing_id = ?", opts.ListingID)
	}
	if opts.InspectorID > 0 {
		query = query.Where("inspector_id = ?", opts.InspectorID)
	}
	if opts.InspectionType != "" {
		query = query.Where("inspection_type = ?", opts.InspectionType)
	}
	if opts.ConditionLevel != "" {
		query = query.Where("condition_level = ?", opts.ConditionLevel)
	}
	if opts.HasAccident != nil {
		query = query.Where("has_accident = ?", *opts.HasAccident)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("inspection_no ILIKE ? OR inspector_name ILIKE ?", like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *inspectionRepository) AdminList(pagination *utils.Pagination, opts InspectionAdminListOptions) ([]model.CarInspection, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.CarInspection
	var total int64

	query := r.db.Model(&model.CarInspection{})

	if opts.RegionID > 0 {
		query = query.Where("region_id = ?", opts.RegionID)
	}
	if opts.CarID > 0 {
		query = query.Where("car_id = ?", opts.CarID)
	}
	if opts.InspectorID > 0 {
		query = query.Where("inspector_id = ?", opts.InspectorID)
	}
	if opts.InspectionType != "" {
		query = query.Where("inspection_type = ?", opts.InspectionType)
	}
	if opts.ConditionLevel != "" {
		query = query.Where("condition_level = ?", opts.ConditionLevel)
	}
	if opts.HasAccident != nil {
		query = query.Where("has_accident = ?", *opts.HasAccident)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("inspection_no ILIKE ? OR inspector_name ILIKE ?", like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *inspectionRepository) ListByInspector(inspectorID uint, pagination *utils.Pagination) ([]model.CarInspection, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.CarInspection
	var total int64

	query := r.db.Model(&model.CarInspection{}).Where("inspector_id = ?", inspectorID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ===== 统计 =====

func (r *inspectionRepository) CountByStatus(regionID uint, status int) (int64, error) {
	var count int64
	q := r.db.Model(&model.CarInspection{}).Where("status = ?", status)
	if regionID > 0 {
		q = q.Where("region_id = ?", regionID)
	}
	if err := q.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *inspectionRepository) CountByInspector(inspectorID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&model.CarInspection{}).Where("inspector_id = ?", inspectorID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
