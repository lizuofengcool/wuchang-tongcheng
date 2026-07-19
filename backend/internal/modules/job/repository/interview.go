// Package repository 面试邀约数据访问层
// 依据 v3.2.1 架构方案：对标 BOSS直聘多轮面试 + Offer
package repository

import (
	"wuchang-tongcheng/internal/modules/job/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// InterviewRepository 面试仓储接口
type InterviewRepository interface {
	Create(i *model.JobInterview) error
	FindByID(id uint) (*model.JobInterview, error)
	FindByInterviewNo(interviewNo string) (*model.JobInterview, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(query InterviewListQuery, pagination *utils.Pagination) ([]model.JobInterview, int64, error)
	ListByApplicationID(applicationID uint) ([]model.JobInterview, error)
	ListByApplicant(applicantID uint, pagination *utils.Pagination) ([]model.JobInterview, int64, error)
	ListByRecruiter(recruiterID uint, pagination *utils.Pagination) ([]model.JobInterview, int64, error)
	CountByApplicationID(applicationID uint) (int64, error)
	CountByStatus(userID uint, role string, status int) (int64, error)
	BatchUpdateStatus(ids []uint, status int) error
	Stats(userID uint, role string) (*model.JobInterview, map[int]int64, error)
}

// InterviewListQuery 面试列表查询
type InterviewListQuery struct {
	UserID        uint
	Role          string // applicant/recruiter/all
	Status        *int
	Result        string
	JobID         uint
	ApplicationID uint
	CompanyID     uint
	InterviewNo   string
	StartTime     interface{}
	EndTime       interface{}
}

type interviewRepository struct {
	db *gorm.DB
}

// NewInterviewRepository 创建面试仓储实例
func NewInterviewRepository(db *gorm.DB) InterviewRepository {
	return &interviewRepository{db: db}
}

func (r *interviewRepository) Create(i *model.JobInterview) error {
	return r.db.Create(i).Error
}

func (r *interviewRepository) FindByID(id uint) (*model.JobInterview, error) {
	var i model.JobInterview
	if err := r.db.First(&i, id).Error; err != nil {
		return nil, err
	}
	return &i, nil
}

func (r *interviewRepository) FindByInterviewNo(interviewNo string) (*model.JobInterview, error) {
	var i model.JobInterview
	if err := r.db.Where("interview_no = ?", interviewNo).First(&i).Error; err != nil {
		return nil, err
	}
	return &i, nil
}

func (r *interviewRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.JobInterview{}).Where("id = ?", id).Updates(fields).Error
}

func (r *interviewRepository) Delete(id uint) error {
	return r.db.Delete(&model.JobInterview{}, id).Error
}

func (r *interviewRepository) List(query InterviewListQuery, pagination *utils.Pagination) ([]model.JobInterview, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.JobInterview
	var total int64

	q := r.db.Model(&model.JobInterview{})
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
	if query.Result != "" {
		q = q.Where("result = ?", query.Result)
	}
	if query.JobID > 0 {
		q = q.Where("job_id = ?", query.JobID)
	}
	if query.ApplicationID > 0 {
		q = q.Where("application_id = ?", query.ApplicationID)
	}
	if query.CompanyID > 0 {
		q = q.Where("company_id = ?", query.CompanyID)
	}
	if query.InterviewNo != "" {
		q = q.Where("interview_no = ?", query.InterviewNo)
	}
	if query.StartTime != nil {
		q = q.Where("scheduled_at >= ?", query.StartTime)
	}
	if query.EndTime != nil {
		q = q.Where("scheduled_at <= ?", query.EndTime)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("scheduled_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *interviewRepository) ListByApplicationID(applicationID uint) ([]model.JobInterview, error) {
	var list []model.JobInterview
	if err := r.db.Where("application_id = ?", applicationID).
		Order("round ASC, id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *interviewRepository) ListByApplicant(applicantID uint, pagination *utils.Pagination) ([]model.JobInterview, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.JobInterview
	var total int64

	q := r.db.Model(&model.JobInterview{}).Where("applicant_id = ?", applicantID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("scheduled_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *interviewRepository) ListByRecruiter(recruiterID uint, pagination *utils.Pagination) ([]model.JobInterview, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.JobInterview
	var total int64

	q := r.db.Model(&model.JobInterview{}).Where("recruiter_id = ?", recruiterID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("scheduled_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *interviewRepository) CountByApplicationID(applicationID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.JobInterview{}).
		Where("application_id = ?", applicationID).
		Count(&count).Error
	return count, err
}

func (r *interviewRepository) CountByStatus(userID uint, role string, status int) (int64, error) {
	var count int64
	q := r.db.Model(&model.JobInterview{})
	switch role {
	case "applicant":
		q = q.Where("applicant_id = ?", userID)
	case "recruiter":
		q = q.Where("recruiter_id = ?", userID)
	}
	err := q.Where("status = ?", status).Count(&count).Error
	return count, err
}

func (r *interviewRepository) BatchUpdateStatus(ids []uint, status int) error {
	return r.db.Model(&model.JobInterview{}).Where("id IN ?", ids).Update("status", status).Error
}

// Stats 返回面试统计（total/pending/confirmed/completed/canceled/no_show 等）
// 第二返回值为按 status 分组的计数
func (r *interviewRepository) Stats(userID uint, role string) (*model.JobInterview, map[int]int64, error) {
	q := r.db.Model(&model.JobInterview{})
	switch role {
	case "applicant":
		q = q.Where("applicant_id = ?", userID)
	case "recruiter":
		q = q.Where("recruiter_id = ?", userID)
	}

	type groupResult struct {
		Status int   `gorm:"column:status"`
		Count  int64 `gorm:"column:count"`
	}
	var groups []groupResult
	if err := q.Select("status, COUNT(*) as count").Group("status").Scan(&groups).Error; err != nil {
		return nil, nil, err
	}
	m := make(map[int]int64, len(groups))
	for _, g := range groups {
		m[g.Status] = g.Count
	}
	return nil, m, nil
}
