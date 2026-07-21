// Package repository 多租户分站数据访问层 - 域名
package repository

import (
	"wuchang-tongcheng/internal/modules/tenant/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// DomainListOptions 域名列表过滤条件
type DomainListOptions struct {
	StationID uint
	Domain    string
	SSLStatus string
}

// DomainRepository 域名仓储接口
type DomainRepository interface {
	Create(d *model.Domain) error
	FindByID(id uint) (*model.Domain, error)
	FindByDomain(domain string) (*model.Domain, error)
	Update(d *model.Domain) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(pagination *utils.Pagination, opts DomainListOptions) ([]model.Domain, int64, error)
	ListByStation(stationID uint) ([]model.Domain, error)
	ClearPrimary(stationID uint) error
}

type domainRepository struct {
	db *gorm.DB
}

// NewDomainRepository 创建域名仓储实例
func NewDomainRepository(db *gorm.DB) DomainRepository {
	return &domainRepository{db: db}
}

func (r *domainRepository) Create(d *model.Domain) error {
	return r.db.Create(d).Error
}

func (r *domainRepository) FindByID(id uint) (*model.Domain, error) {
	var d model.Domain
	if err := r.db.First(&d, id).Error; err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *domainRepository) FindByDomain(domain string) (*model.Domain, error) {
	var d model.Domain
	if err := r.db.Where("domain = ?", domain).First(&d).Error; err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *domainRepository) Update(d *model.Domain) error {
	return r.db.Save(d).Error
}

func (r *domainRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Domain{}).Where("id = ?", id).Updates(fields).Error
}

func (r *domainRepository) Delete(id uint) error {
	return r.db.Delete(&model.Domain{}, id).Error
}

func (r *domainRepository) List(pagination *utils.Pagination, opts DomainListOptions) ([]model.Domain, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 20)
	}
	var list []model.Domain
	var total int64

	query := r.db.Model(&model.Domain{})
	if opts.StationID > 0 {
		query = query.Where("station_id = ?", opts.StationID)
	}
	if opts.Domain != "" {
		query = query.Where("domain ILIKE ?", "%"+opts.Domain+"%")
	}
	if opts.SSLStatus != "" {
		query = query.Where("ssl_status = ?", opts.SSLStatus)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("is_primary DESC, created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *domainRepository) ListByStation(stationID uint) ([]model.Domain, error) {
	var list []model.Domain
	if err := r.db.Where("station_id = ?", stationID).
		Order("is_primary DESC, created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// ClearPrimary 清除指定分站下所有域名的主域名标记（用于切换主域名）
func (r *domainRepository) ClearPrimary(stationID uint) error {
	return r.db.Model(&model.Domain{}).
		Where("station_id = ? AND is_primary = true", stationID).
		Update("is_primary", false).Error
}
