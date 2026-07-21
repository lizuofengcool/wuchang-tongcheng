// Package repository LBS地图中台数据访问层 - 区域分站
package repository

import (
	"strconv"

	"wuchang-tongcheng/internal/modules/lbs/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// RegionRepository 区域仓储接口
type RegionRepository interface {
	// CRUD
	Create(r *model.Region) error
	FindByID(id uint) (*model.Region, error)
	Update(r *model.Region) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	// 列表
	List(pagination *utils.Pagination, opts RegionListOptions) ([]model.Region, int64, error)
	ListByParent(parentID uint) ([]model.Region, error)
	FindByCityCode(cityCode string) (*model.Region, error)
	FindByAdCode(adCode string) (*model.Region, error)
	FindByLocation(lat, lng float64) (*model.Region, error)
}

// RegionListOptions 区域列表过滤条件
type RegionListOptions struct {
	Keyword  string
	CityCode string
	ParentID uint
	Level    int
	Status   *int
}

type regionRepository struct {
	db *gorm.DB
}

// NewRegionRepository 创建区域仓储实例
func NewRegionRepository(db *gorm.DB) RegionRepository {
	return &regionRepository{db: db}
}

func (r *regionRepository) Create(reg *model.Region) error {
	return r.db.Create(reg).Error
}

func (r *regionRepository) FindByID(id uint) (*model.Region, error) {
	var reg model.Region
	if err := r.db.First(&reg, id).Error; err != nil {
		return nil, err
	}
	return &reg, nil
}

func (r *regionRepository) Update(reg *model.Region) error {
	return r.db.Save(reg).Error
}

func (r *regionRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Region{}).Where("id = ?", id).Updates(fields).Error
}

func (r *regionRepository) Delete(id uint) error {
	return r.db.Delete(&model.Region{}, id).Error
}

func (r *regionRepository) List(pagination *utils.Pagination, opts RegionListOptions) ([]model.Region, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Region
	var total int64

	query := r.db.Model(&model.Region{})
	if opts.CityCode != "" {
		query = query.Where("city_code = ?", opts.CityCode)
	}
	if opts.ParentID > 0 {
		query = query.Where("parent_id = ?", opts.ParentID)
	}
	if opts.Level > 0 {
		query = query.Where("level = ?", opts.Level)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("name ILIKE ?", like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("sort ASC, id ASC").
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *regionRepository) ListByParent(parentID uint) ([]model.Region, error) {
	var list []model.Region
	if err := r.db.Where("parent_id = ?", parentID).
		Order("sort ASC, id ASC").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *regionRepository) FindByCityCode(cityCode string) (*model.Region, error) {
	var reg model.Region
	if err := r.db.Where("city_code = ? AND status = ?", cityCode, model.LBSRegionStatusEnabled).
		First(&reg).Error; err != nil {
		return nil, err
	}
	return &reg, nil
}

func (r *regionRepository) FindByAdCode(adCode string) (*model.Region, error) {
	var reg model.Region
	if err := r.db.Where("ad_code = ? AND status = ?", adCode, model.LBSRegionStatusEnabled).
		First(&reg).Error; err != nil {
		return nil, err
	}
	return &reg, nil
}

// FindByLocation 根据经纬度查找分站
// 通过 center_lat/center_lng 反向最近邻匹配（DECIMAL 降级方案，不依赖 PostGIS）
// 若需精确边界判断，使用 boundary JSONB 字段配合多边形算法在 service 层处理
func (r *regionRepository) FindByLocation(lat, lng float64) (*model.Region, error) {
	var reg model.Region
	// 简化方案：取状态启用、中心点最近的区域（按经纬度平方距离排序）
	// 注意：GORM Order 不支持参数化占位符，但 lat/lng 为 float64 数值，
	// 通过 strconv.FormatFloat 转为字符串内嵌 SQL，避免注入风险
	latStr := strconv.FormatFloat(lat, 'f', 7, 64)
	lngStr := strconv.FormatFloat(lng, 'f', 7, 64)
	orderExpr := "((center_lat - " + latStr + ")*(center_lat - " + latStr + ") + " +
		"(center_lng - " + lngStr + ")*(center_lng - " + lngStr + ")) ASC"
	if err := r.db.Where("status = ? AND center_lat <> 0 AND center_lng <> 0", model.LBSRegionStatusEnabled).
		Order(orderExpr).
		First(&reg).Error; err != nil {
		return nil, err
	}
	return &reg, nil
}
