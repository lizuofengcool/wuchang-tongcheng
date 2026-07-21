// Package repository 多租户分站数据访问层 - 配置
package repository

import (
	"wuchang-tongcheng/internal/modules/tenant/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// ConfigListOptions 配置列表过滤条件
type ConfigListOptions struct {
	StationID uint
	BizModule string
	ConfigKey string
	Keyword   string
}

// ConfigRepository 配置仓储接口
type ConfigRepository interface {
	Create(c *model.Config) error
	FindByID(id uint) (*model.Config, error)
	FindByStationAndKey(stationID uint, bizModule, configKey string) (*model.Config, error)
	Upsert(c *model.Config) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	DeleteByStation(stationID uint, bizModule string) error
	List(pagination *utils.Pagination, opts ConfigListOptions) ([]model.Config, int64, error)
	ListByStationAndModule(stationID uint, bizModule string) ([]model.Config, error)
	BatchGet(stationID uint, bizModule string, keys []string) ([]model.Config, error)
}

type configRepository struct {
	db *gorm.DB
}

// NewConfigRepository 创建配置仓储实例
func NewConfigRepository(db *gorm.DB) ConfigRepository {
	return &configRepository{db: db}
}

func (r *configRepository) Create(c *model.Config) error {
	return r.db.Create(c).Error
}

func (r *configRepository) FindByID(id uint) (*model.Config, error) {
	var c model.Config
	if err := r.db.First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *configRepository) FindByStationAndKey(stationID uint, bizModule, configKey string) (*model.Config, error) {
	var c model.Config
	if err := r.db.Where("station_id = ? AND biz_module = ? AND config_key = ?", stationID, bizModule, configKey).
		First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

// Upsert 按 station_id + biz_module + config_key 唯一键新增或更新
func (r *configRepository) Upsert(c *model.Config) error {
	existing, err := r.FindByStationAndKey(c.StationID, c.BizModule, c.ConfigKey)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return r.db.Create(c).Error
		}
		return err
	}
	c.ID = existing.ID
	return r.db.Model(&model.Config{}).Where("id = ?", existing.ID).
		Update("config_value", c.ConfigValue).Error
}

func (r *configRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Config{}).Where("id = ?", id).Updates(fields).Error
}

func (r *configRepository) Delete(id uint) error {
	return r.db.Delete(&model.Config{}, id).Error
}

func (r *configRepository) DeleteByStation(stationID uint, bizModule string) error {
	query := r.db.Where("station_id = ?", stationID)
	if bizModule != "" {
		query = query.Where("biz_module = ?", bizModule)
	}
	return query.Delete(&model.Config{}).Error
}

func (r *configRepository) List(pagination *utils.Pagination, opts ConfigListOptions) ([]model.Config, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 20)
	}
	var list []model.Config
	var total int64

	query := r.db.Model(&model.Config{})
	if opts.StationID > 0 {
		query = query.Where("station_id = ?", opts.StationID)
	}
	if opts.BizModule != "" {
		query = query.Where("biz_module = ?", opts.BizModule)
	}
	if opts.ConfigKey != "" {
		query = query.Where("config_key ILIKE ?", "%"+opts.ConfigKey+"%")
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("config_key ILIKE ? OR config_value ILIKE ?", like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("updated_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *configRepository) ListByStationAndModule(stationID uint, bizModule string) ([]model.Config, error) {
	var list []model.Config
	if err := r.db.Where("station_id = ? AND biz_module = ?", stationID, bizModule).
		Order("config_key ASC, id ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *configRepository) BatchGet(stationID uint, bizModule string, keys []string) ([]model.Config, error) {
	var list []model.Config
	query := r.db.Where("station_id = ? AND biz_module = ?", stationID, bizModule)
	if len(keys) > 0 {
		query = query.Where("config_key IN ?", keys)
	}
	if err := query.Order("config_key ASC, id ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
