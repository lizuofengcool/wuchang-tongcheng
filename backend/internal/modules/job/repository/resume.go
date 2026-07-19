// Package repository 简历数据访问层
// 依据 v3.2.1 架构方案：对标 BOSS直聘简历
package repository

import (
	"wuchang-tongcheng/internal/modules/job/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// ResumeRepository 简历仓储接口
type ResumeRepository interface {
	Create(r *model.JobResume) error
	FindByID(id uint) (*model.JobResume, error)
	Update(r *model.JobResume) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(query ResumeListQuery, pagination *utils.Pagination) ([]model.JobResume, int64, error)
	ListByUser(userID uint, pagination *utils.Pagination) ([]model.JobResume, int64, error)
	GetDefault(userID uint) (*model.JobResume, error)
	SetDefault(userID, resumeID uint) error

	// 浏览/投递统计
	IncrViewCount(id uint) error
	IncrDeliverCount(id uint) error
	IncrInterviewCount(id uint) error
	IncrOfferCount(id uint) error
}

// ResumeListQuery 简历列表查询
type ResumeListQuery struct {
	UserID         uint
	Keyword        string
	EducationLevel string
	ExpectCity     string
	ExpectPosition string
	IsPublic       *bool
	Status         *int
}

type resumeRepository struct {
	db *gorm.DB
}

// NewResumeRepository 创建简历仓储实例
func NewResumeRepository(db *gorm.DB) ResumeRepository {
	return &resumeRepository{db: db}
}

func (r *resumeRepository) Create(resume *model.JobResume) error {
	return r.db.Create(resume).Error
}

func (r *resumeRepository) FindByID(id uint) (*model.JobResume, error) {
	var resume model.JobResume
	if err := r.db.First(&resume, id).Error; err != nil {
		return nil, err
	}
	return &resume, nil
}

func (r *resumeRepository) Update(resume *model.JobResume) error {
	return r.db.Save(resume).Error
}

func (r *resumeRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.JobResume{}).Where("id = ?", id).Updates(fields).Error
}

func (r *resumeRepository) Delete(id uint) error {
	return r.db.Delete(&model.JobResume{}, id).Error
}

func (r *resumeRepository) List(query ResumeListQuery, pagination *utils.Pagination) ([]model.JobResume, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.JobResume
	var total int64

	q := r.db.Model(&model.JobResume{})
	if query.UserID > 0 {
		q = q.Where("user_id = ?", query.UserID)
	}
	if query.Keyword != "" {
		like := "%" + query.Keyword + "%"
		q = q.Where("name ILIKE ? OR expect_position ILIKE ? OR current_position ILIKE ? OR major ILIKE ?", like, like, like, like)
	}
	if query.EducationLevel != "" {
		q = q.Where("education_level = ?", query.EducationLevel)
	}
	if query.ExpectCity != "" {
		q = q.Where("expect_city ILIKE ?", "%"+query.ExpectCity+"%")
	}
	if query.ExpectPosition != "" {
		q = q.Where("expect_position ILIKE ?", "%"+query.ExpectPosition+"%")
	}
	if query.IsPublic != nil {
		q = q.Where("is_public = ?", *query.IsPublic)
	}
	if query.Status != nil {
		q = q.Where("status = ?", *query.Status)
	} else {
		q = q.Where("status = ?", model.ResumeStatusPublished)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("is_default DESC, updated_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *resumeRepository) ListByUser(userID uint, pagination *utils.Pagination) ([]model.JobResume, int64, error) {
	return r.List(ResumeListQuery{UserID: userID, Status: intPtr(model.ResumeStatusDraft)}, pagination)
}

// intPtr 工具：返回 int 指针
func intPtr(v int) *int {
	return &v
}

func (r *resumeRepository) GetDefault(userID uint) (*model.JobResume, error) {
	var resume model.JobResume
	if err := r.db.Where("user_id = ? AND is_default = ?", userID, true).First(&resume).Error; err != nil {
		return nil, err
	}
	return &resume, nil
}

func (r *resumeRepository) SetDefault(userID, resumeID uint) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	// 先取消其他默认
	if err := tx.Model(&model.JobResume{}).Where("user_id = ?", userID).
		Update("is_default", false).Error; err != nil {
		tx.Rollback()
		return err
	}
	// 设置当前为默认
	if err := tx.Model(&model.JobResume{}).Where("id = ? AND user_id = ?", resumeID, userID).
		Update("is_default", true).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

// ===== 统计 =====

func (r *resumeRepository) IncrViewCount(id uint) error {
	return r.db.Model(&model.JobResume{}).Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

func (r *resumeRepository) IncrDeliverCount(id uint) error {
	return r.db.Model(&model.JobResume{}).Where("id = ?", id).
		UpdateColumn("deliver_count", gorm.Expr("deliver_count + 1")).Error
}

func (r *resumeRepository) IncrInterviewCount(id uint) error {
	return r.db.Model(&model.JobResume{}).Where("id = ?", id).
		UpdateColumn("interview_count", gorm.Expr("interview_count + 1")).Error
}

func (r *resumeRepository) IncrOfferCount(id uint) error {
	return r.db.Model(&model.JobResume{}).Where("id = ?", id).
		UpdateColumn("offer_count", gorm.Expr("offer_count + 1")).Error
}
