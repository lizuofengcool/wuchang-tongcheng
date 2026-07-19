// Package repository 担保交易 + AI 推荐数据访问层
// 依据 v3.2.1 架构方案：对标 BOSS直聘担保交易/智能推荐
package repository

import (
	"wuchang-tongcheng/internal/modules/job/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// ===== Escrow 担保交易 =====

// EscrowRepository 担保交易仓储接口
type EscrowRepository interface {
	Create(e *model.JobEscrow) error
	FindByID(id uint) (*model.JobEscrow, error)
	FindByEscrowNo(escrowNo string) (*model.JobEscrow, error)
	Update(id uint, fields map[string]interface{}) error
	List(query EscrowListQuery, pagination *utils.Pagination) ([]model.JobEscrow, int64, error)
	ListByJobID(jobID uint, pagination *utils.Pagination) ([]model.JobEscrow, int64, error)
	CountByStatus(userID uint, role string, status int) (int64, error)
}

// EscrowListQuery 担保列表查询
type EscrowListQuery struct {
	UserID    uint
	Role      string // payer/payee/all
	Status    *int
	EscrowType string
	JobID     uint
	EscrowNo  string
}

type escrowRepository struct {
	db *gorm.DB
}

// NewEscrowRepository 创建担保仓储实例
func NewEscrowRepository(db *gorm.DB) EscrowRepository {
	return &escrowRepository{db: db}
}

func (r *escrowRepository) Create(e *model.JobEscrow) error {
	return r.db.Create(e).Error
}

func (r *escrowRepository) FindByID(id uint) (*model.JobEscrow, error) {
	var e model.JobEscrow
	if err := r.db.First(&e, id).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *escrowRepository) FindByEscrowNo(escrowNo string) (*model.JobEscrow, error) {
	var e model.JobEscrow
	if err := r.db.Where("escrow_no = ?", escrowNo).First(&e).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *escrowRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.JobEscrow{}).Where("id = ?", id).Updates(fields).Error
}

func (r *escrowRepository) List(query EscrowListQuery, pagination *utils.Pagination) ([]model.JobEscrow, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.JobEscrow
	var total int64

	q := r.db.Model(&model.JobEscrow{})
	switch query.Role {
	case "payer":
		q = q.Where("payer_id = ?", query.UserID)
	case "payee":
		q = q.Where("payee_id = ?", query.UserID)
	case "all":
		q = q.Where("payer_id = ? OR payee_id = ?", query.UserID, query.UserID)
	}
	if query.Status != nil {
		q = q.Where("status = ?", *query.Status)
	}
	if query.EscrowType != "" {
		q = q.Where("escrow_type = ?", query.EscrowType)
	}
	if query.JobID > 0 {
		q = q.Where("job_id = ?", query.JobID)
	}
	if query.EscrowNo != "" {
		q = q.Where("escrow_no = ?", query.EscrowNo)
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

func (r *escrowRepository) ListByJobID(jobID uint, pagination *utils.Pagination) ([]model.JobEscrow, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.JobEscrow
	var total int64

	q := r.db.Model(&model.JobEscrow{}).Where("job_id = ?", jobID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *escrowRepository) CountByStatus(userID uint, role string, status int) (int64, error) {
	var count int64
	q := r.db.Model(&model.JobEscrow{})
	switch role {
	case "payer":
		q = q.Where("payer_id = ?", userID)
	case "payee":
		q = q.Where("payee_id = ?", userID)
	}
	err := q.Where("status = ?", status).Count(&count).Error
	return count, err
}

// ===== Recommendation AI 推荐 =====

// RecommendationRepository AI 推荐仓储接口
type RecommendationRepository interface {
	Create(r *model.JobRecommendation) error
	FindByID(id uint) (*model.JobRecommendation, error)
	Update(id uint, fields map[string]interface{}) error
	List(query RecommendationListQuery, pagination *utils.Pagination) ([]model.JobRecommendation, int64, error)
	ListByUser(userID uint, recType string, pagination *utils.Pagination) ([]model.JobRecommendation, int64, error)
	Delete(id uint) error
	BatchCreate(items []model.JobRecommendation) error
}

// RecommendationListQuery 推荐列表查询
type RecommendationListQuery struct {
	UserID   uint
	JobID    uint
	RecType  string
	Source   string
	Status   *int
}

type recommendationRepository struct {
	db *gorm.DB
}

// NewRecommendationRepository 创建 AI 推荐仓储实例
func NewRecommendationRepository(db *gorm.DB) RecommendationRepository {
	return &recommendationRepository{db: db}
}

func (r *recommendationRepository) Create(rec *model.JobRecommendation) error {
	return r.db.Create(rec).Error
}

func (r *recommendationRepository) FindByID(id uint) (*model.JobRecommendation, error) {
	var rec model.JobRecommendation
	if err := r.db.First(&rec, id).Error; err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *recommendationRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.JobRecommendation{}).Where("id = ?", id).Updates(fields).Error
}

func (r *recommendationRepository) List(query RecommendationListQuery, pagination *utils.Pagination) ([]model.JobRecommendation, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.JobRecommendation
	var total int64

	q := r.db.Model(&model.JobRecommendation{})
	if query.UserID > 0 {
		q = q.Where("user_id = ?", query.UserID)
	}
	if query.JobID > 0 {
		q = q.Where("job_id = ?", query.JobID)
	}
	if query.RecType != "" {
		q = q.Where("rec_type = ?", query.RecType)
	}
	if query.Source != "" {
		q = q.Where("source = ?", query.Source)
	}
	if query.Status != nil {
		q = q.Where("status = ?", *query.Status)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("score DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *recommendationRepository) ListByUser(userID uint, recType string, pagination *utils.Pagination) ([]model.JobRecommendation, int64, error) {
	return r.List(RecommendationListQuery{UserID: userID, RecType: recType}, pagination)
}

func (r *recommendationRepository) Delete(id uint) error {
	return r.db.Delete(&model.JobRecommendation{}, id).Error
}

func (r *recommendationRepository) BatchCreate(items []model.JobRecommendation) error {
	if len(items) == 0 {
		return nil
	}
	return r.db.Create(&items).Error
}
