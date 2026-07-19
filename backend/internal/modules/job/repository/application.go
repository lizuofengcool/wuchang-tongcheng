// Package repository 投递记录数据访问层
// 依据 v3.2.1 架构方案：对标 BOSS直聘 9 状态机投递
package repository

import (
	"wuchang-tongcheng/internal/modules/job/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// ApplicationRepository 投递记录仓储接口
type ApplicationRepository interface {
	Create(a *model.JobApplication) error
	FindByID(id uint) (*model.JobApplication, error)
	FindByApplicationNo(applicationNo string) (*model.JobApplication, error)
	Update(id uint, fields map[string]interface{}) error
	List(query ApplicationListQuery, pagination *utils.Pagination) ([]model.JobApplication, int64, error)
	ListByJobID(jobID uint, pagination *utils.Pagination) ([]model.JobApplication, int64, error)
	ListByApplicant(applicantID uint, pagination *utils.Pagination) ([]model.JobApplication, int64, error)
	ListByRecruiter(recruiterID uint, pagination *utils.Pagination) ([]model.JobApplication, int64, error)
	CountByJobAndApplicant(jobID, applicantID uint) (int64, error)
	CountByStatus(userID uint, role string, status int) (int64, error)
	BatchUpdateStatus(ids []uint, status int) error
	BatchUpdateFields(ids []uint, fields map[string]interface{}) error

	// 互动统计
	IncrInterviewCount(id uint) error
	SetOfferInfo(id uint, amount float64, offerAt interface{}) error
}

// ApplicationListQuery 投递列表查询
type ApplicationListQuery struct {
	UserID        uint
	Role          string // applicant/recruiter/all
	Status        *int
	JobID         uint
	CompanyID     uint
	ApplicationNo string
	Keyword       string
}

type applicationRepository struct {
	db *gorm.DB
}

// NewApplicationRepository 创建投递记录仓储实例
func NewApplicationRepository(db *gorm.DB) ApplicationRepository {
	return &applicationRepository{db: db}
}

func (r *applicationRepository) Create(a *model.JobApplication) error {
	return r.db.Create(a).Error
}

func (r *applicationRepository) FindByID(id uint) (*model.JobApplication, error) {
	var a model.JobApplication
	if err := r.db.First(&a, id).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *applicationRepository) FindByApplicationNo(applicationNo string) (*model.JobApplication, error) {
	var a model.JobApplication
	if err := r.db.Where("application_no = ?", applicationNo).First(&a).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *applicationRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.JobApplication{}).Where("id = ?", id).Updates(fields).Error
}

func (r *applicationRepository) List(query ApplicationListQuery, pagination *utils.Pagination) ([]model.JobApplication, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.JobApplication
	var total int64

	q := r.db.Model(&model.JobApplication{})
	switch query.Role {
	case "applicant":
		q = q.Where("applicant_id = ?", query.UserID)
	case "recruiter":
		q = q.Where("recruiter_id = ?", query.UserID)
	case "all":
		q = q.Where("applicant_id = ? OR recruiter_id = ?", query.UserID, query.UserID)
	}
	if query.Status != nil {
		q = q.Where("status = ?", *query.Status)
	}
	if query.JobID > 0 {
		q = q.Where("job_id = ?", query.JobID)
	}
	if query.CompanyID > 0 {
		q = q.Where("company_id = ?", query.CompanyID)
	}
	if query.ApplicationNo != "" {
		q = q.Where("application_no = ?", query.ApplicationNo)
	}
	if query.Keyword != "" {
		like := "%" + query.Keyword + "%"
		q = q.Where("application_no ILIKE ? OR position_name ILIKE ?", like, like)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *applicationRepository) ListByJobID(jobID uint, pagination *utils.Pagination) ([]model.JobApplication, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.JobApplication
	var total int64

	q := r.db.Model(&model.JobApplication{}).Where("job_id = ?", jobID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *applicationRepository) ListByApplicant(applicantID uint, pagination *utils.Pagination) ([]model.JobApplication, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.JobApplication
	var total int64

	q := r.db.Model(&model.JobApplication{}).Where("applicant_id = ?", applicantID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *applicationRepository) ListByRecruiter(recruiterID uint, pagination *utils.Pagination) ([]model.JobApplication, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.JobApplication
	var total int64

	q := r.db.Model(&model.JobApplication{}).Where("recruiter_id = ?", recruiterID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *applicationRepository) CountByJobAndApplicant(jobID, applicantID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.JobApplication{}).
		Where("job_id = ? AND applicant_id = ?", jobID, applicantID).
		Count(&count).Error
	return count, err
}

func (r *applicationRepository) CountByStatus(userID uint, role string, status int) (int64, error) {
	var count int64
	q := r.db.Model(&model.JobApplication{})
	switch role {
	case "applicant":
		q = q.Where("applicant_id = ?", userID)
	case "recruiter":
		q = q.Where("recruiter_id = ?", userID)
	}
	err := q.Where("status = ?", status).Count(&count).Error
	return count, err
}

func (r *applicationRepository) BatchUpdateStatus(ids []uint, status int) error {
	return r.db.Model(&model.JobApplication{}).Where("id IN ?", ids).Update("status", status).Error
}

func (r *applicationRepository) BatchUpdateFields(ids []uint, fields map[string]interface{}) error {
	return r.db.Model(&model.JobApplication{}).Where("id IN ?", ids).Updates(fields).Error
}

func (r *applicationRepository) IncrInterviewCount(id uint) error {
	return r.db.Model(&model.JobApplication{}).Where("id = ?", id).
		UpdateColumn("interview_count", gorm.Expr("interview_count + 1")).Error
}

func (r *applicationRepository) SetOfferInfo(id uint, amount float64, offerAt interface{}) error {
	return r.db.Model(&model.JobApplication{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"offer_amount": amount,
			"offer_at":     offerAt,
		}).Error
}
