// Package repository love 相亲交友数据访问层 - 认证
package repository

import (
	"wuchang-tongcheng/internal/modules/love/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// LoveVerificationRepository 认证仓储接口
type LoveVerificationRepository interface {
	Create(v *model.LoveVerification) error
	FindByID(id uint) (*model.LoveVerification, error)
	FindByVerifyNo(no string) (*model.LoveVerification, error)
	FindByLoveIDAndType(loveID uint, verifyType string) (*model.LoveVerification, error)
	ListByLoveID(loveID uint) ([]model.LoveVerification, error)
	ListByUserID(userID uint) ([]model.LoveVerification, error)
	Update(v *model.LoveVerification) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(pagination *utils.Pagination, opts LoveVerificationListOptions) ([]model.LoveVerification, int64, error)
	UpdateStatus(id uint, status int, reason string, verifiedBy uint, verifiedName string) error
	CountPending() (int64, error)
	CountApprovedByLoveID(loveID uint) (int64, error)
}

// LoveVerificationListOptions 认证列表过滤
type LoveVerificationListOptions struct {
	UserID uint
	LoveID uint
	Type   string
	Status *int
}

type loveVerificationRepository struct {
	db *gorm.DB
}

// NewLoveVerificationRepository 创建认证仓储
func NewLoveVerificationRepository(db *gorm.DB) LoveVerificationRepository {
	return &loveVerificationRepository{db: db}
}

func (r *loveVerificationRepository) Create(v *model.LoveVerification) error {
	return r.db.Create(v).Error
}

func (r *loveVerificationRepository) FindByID(id uint) (*model.LoveVerification, error) {
	var v model.LoveVerification
	if err := r.db.First(&v, id).Error; err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *loveVerificationRepository) FindByVerifyNo(no string) (*model.LoveVerification, error) {
	var v model.LoveVerification
	if err := r.db.Where("verify_no = ?", no).First(&v).Error; err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *loveVerificationRepository) FindByLoveIDAndType(loveID uint, verifyType string) (*model.LoveVerification, error) {
	var v model.LoveVerification
	if err := r.db.Where("love_id = ? AND type = ?", loveID, verifyType).First(&v).Error; err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *loveVerificationRepository) ListByLoveID(loveID uint) ([]model.LoveVerification, error) {
	var list []model.LoveVerification
	err := r.db.Where("love_id = ?", loveID).Order("id DESC").Find(&list).Error
	return list, err
}

func (r *loveVerificationRepository) ListByUserID(userID uint) ([]model.LoveVerification, error) {
	var list []model.LoveVerification
	err := r.db.Where("user_id = ?", userID).Order("id DESC").Find(&list).Error
	return list, err
}

func (r *loveVerificationRepository) Update(v *model.LoveVerification) error {
	return r.db.Save(v).Error
}

func (r *loveVerificationRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.LoveVerification{}).Where("id = ?", id).Updates(fields).Error
}

func (r *loveVerificationRepository) Delete(id uint) error {
	return r.db.Delete(&model.LoveVerification{}, id).Error
}

func (r *loveVerificationRepository) List(pagination *utils.Pagination, opts LoveVerificationListOptions) ([]model.LoveVerification, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LoveVerification
	var total int64

	query := r.db.Model(&model.LoveVerification{})
	if opts.UserID > 0 {
		query = query.Where("user_id = ?", opts.UserID)
	}
	if opts.LoveID > 0 {
		query = query.Where("love_id = ?", opts.LoveID)
	}
	if opts.Type != "" {
		query = query.Where("type = ?", opts.Type)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).Order("id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *loveVerificationRepository) UpdateStatus(id uint, status int, reason string, verifiedBy uint, verifiedName string) error {
	updates := map[string]interface{}{
		"status":         status,
		"reject_reason":  reason,
		"verified_by":    verifiedBy,
		"verified_name":  verifiedName,
	}
	if status == model.VerifyStatusApproved {
		updates["verified_at"] = gorm.Expr("NOW()")
	}
	return r.db.Model(&model.LoveVerification{}).Where("id = ?", id).Updates(updates).Error
}

func (r *loveVerificationRepository) CountPending() (int64, error) {
	var count int64
	err := r.db.Model(&model.LoveVerification{}).Where("status = ?", model.VerifyStatusPending).Count(&count).Error
	return count, err
}

func (r *loveVerificationRepository) CountApprovedByLoveID(loveID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.LoveVerification{}).Where("love_id = ? AND status = ?", loveID, model.VerifyStatusApproved).Count(&count).Error
	return count, err
}
