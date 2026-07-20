// Package repository 同城零工兼职数据访问层 - 资质证书
package repository

import (
	"wuchang-tongcheng/internal/modules/linggong/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// CertificationRepository 资质证书仓储接口
type CertificationRepository interface {
	Create(c *model.LinggongCertification) error
	FindByID(id uint) (*model.LinggongCertification, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	List(regionID uint, pagination *utils.Pagination, opts CertificationListOptions) ([]model.LinggongCertification, int64, error)
	ListByWorker(workerID uint, pagination *utils.Pagination) ([]model.LinggongCertification, int64, error)
	ListByUser(userID uint, pagination *utils.Pagination) ([]model.LinggongCertification, int64, error)
}

// CertificationListOptions 资质证书列表过滤条件
type CertificationListOptions struct {
	UserID   uint
	WorkerID uint
	CertType string
	SkillID  uint
	Status   *int
	Verified *bool
}

type certificationRepository struct {
	db *gorm.DB
}

// NewCertificationRepository 创建资质证书仓储实例
func NewCertificationRepository(db *gorm.DB) CertificationRepository {
	return &certificationRepository{db: db}
}

func (r *certificationRepository) Create(c *model.LinggongCertification) error {
	return r.db.Create(c).Error
}

func (r *certificationRepository) FindByID(id uint) (*model.LinggongCertification, error) {
	var c model.LinggongCertification
	if err := r.db.First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *certificationRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.LinggongCertification{}).Where("id = ?", id).Updates(fields).Error
}

func (r *certificationRepository) Delete(id uint) error {
	return r.db.Delete(&model.LinggongCertification{}, id).Error
}

func (r *certificationRepository) List(regionID uint, pagination *utils.Pagination, opts CertificationListOptions) ([]model.LinggongCertification, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LinggongCertification
	var total int64

	query := r.db.Model(&model.LinggongCertification{})
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if opts.UserID > 0 {
		query = query.Where("user_id = ?", opts.UserID)
	}
	if opts.WorkerID > 0 {
		query = query.Where("worker_id = ?", opts.WorkerID)
	}
	if opts.CertType != "" {
		query = query.Where("cert_type = ?", opts.CertType)
	}
	if opts.SkillID > 0 {
		query = query.Where("skill_id = ?", opts.SkillID)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.Verified != nil {
		query = query.Where("verified = ?", *opts.Verified)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *certificationRepository) ListByWorker(workerID uint, pagination *utils.Pagination) ([]model.LinggongCertification, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LinggongCertification
	var total int64
	query := r.db.Model(&model.LinggongCertification{}).Where("worker_id = ?", workerID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *certificationRepository) ListByUser(userID uint, pagination *utils.Pagination) ([]model.LinggongCertification, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LinggongCertification
	var total int64
	query := r.db.Model(&model.LinggongCertification{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
