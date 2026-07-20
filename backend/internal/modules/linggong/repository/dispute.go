// Package repository 同城零工兼职数据访问层 - 纠纷
package repository

import (
	"wuchang-tongcheng/internal/modules/linggong/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// DisputeRepository 纠纷仓储接口
type DisputeRepository interface {
	Create(d *model.LinggongDispute) error
	FindByID(id uint) (*model.LinggongDispute, error)
	FindByDisputeNo(no string) (*model.LinggongDispute, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	List(regionID uint, pagination *utils.Pagination, opts DisputeListOptions) ([]model.LinggongDispute, int64, error)
	ListByLinggong(linggongID uint, pagination *utils.Pagination) ([]model.LinggongDispute, int64, error)
	ListByApplicant(applicantID uint, pagination *utils.Pagination) ([]model.LinggongDispute, int64, error)
	ListByRespondent(respondentID uint, pagination *utils.Pagination) ([]model.LinggongDispute, int64, error)
}

// DisputeListOptions 纠纷列表过滤条件
type DisputeListOptions struct {
	LinggongID   uint
	DisputeType  string
	ApplicantID  uint
	RespondentID uint
	Status       *int
	Keyword      string
}

type disputeRepository struct {
	db *gorm.DB
}

// NewDisputeRepository 创建纠纷仓储实例
func NewDisputeRepository(db *gorm.DB) DisputeRepository {
	return &disputeRepository{db: db}
}

func (r *disputeRepository) Create(d *model.LinggongDispute) error {
	return r.db.Create(d).Error
}

func (r *disputeRepository) FindByID(id uint) (*model.LinggongDispute, error) {
	var d model.LinggongDispute
	if err := r.db.First(&d, id).Error; err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *disputeRepository) FindByDisputeNo(no string) (*model.LinggongDispute, error) {
	var d model.LinggongDispute
	if err := r.db.Where("dispute_no = ?", no).First(&d).Error; err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *disputeRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.LinggongDispute{}).Where("id = ?", id).Updates(fields).Error
}

func (r *disputeRepository) Delete(id uint) error {
	return r.db.Delete(&model.LinggongDispute{}, id).Error
}

func (r *disputeRepository) List(regionID uint, pagination *utils.Pagination, opts DisputeListOptions) ([]model.LinggongDispute, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LinggongDispute
	var total int64

	query := r.db.Model(&model.LinggongDispute{})
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if opts.LinggongID > 0 {
		query = query.Where("linggong_id = ?", opts.LinggongID)
	}
	if opts.DisputeType != "" {
		query = query.Where("dispute_type = ?", opts.DisputeType)
	}
	if opts.ApplicantID > 0 {
		query = query.Where("applicant_id = ?", opts.ApplicantID)
	}
	if opts.RespondentID > 0 {
		query = query.Where("respondent_id = ?", opts.RespondentID)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("dispute_no ILIKE ? OR title ILIKE ?", like, like)
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

func (r *disputeRepository) ListByLinggong(linggongID uint, pagination *utils.Pagination) ([]model.LinggongDispute, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LinggongDispute
	var total int64
	query := r.db.Model(&model.LinggongDispute{}).Where("linggong_id = ?", linggongID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *disputeRepository) ListByApplicant(applicantID uint, pagination *utils.Pagination) ([]model.LinggongDispute, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LinggongDispute
	var total int64
	query := r.db.Model(&model.LinggongDispute{}).Where("applicant_id = ?", applicantID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *disputeRepository) ListByRespondent(respondentID uint, pagination *utils.Pagination) ([]model.LinggongDispute, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LinggongDispute
	var total int64
	query := r.db.Model(&model.LinggongDispute{}).Where("respondent_id = ?", respondentID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
