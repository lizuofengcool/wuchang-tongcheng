// Package repository 同城114数据访问层 - 商户认证
// 营业执照 + 实地认证 + 品牌授权
package repository

import (
	"wuchang-tongcheng/internal/modules/dh114/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// VerificationRepository 商户认证仓储接口
type VerificationRepository interface {
	Create(v *model.Dh114Verification) error
	FindByID(id uint) (*model.Dh114Verification, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(query VerificationListQuery, pagination *utils.Pagination) ([]model.Dh114Verification, int64, error)
	ListByDh114(dh114ID uint, pagination *utils.Pagination) ([]model.Dh114Verification, int64, error)
	ListByUser(userID uint, pagination *utils.Pagination) ([]model.Dh114Verification, int64, error)
	FindLatestByDh114(dh114ID uint, vType string) (*model.Dh114Verification, error)
	UpdateAudit(id uint, status int, auditRemark string, auditedBy uint, auditedAt interface{}, validUntil interface{}) error
}

// VerificationListQuery 认证列表查询
type VerificationListQuery struct {
	Dh114ID          uint
	UserID           uint
	VerificationType string
	Status           *int
	Keyword          string
}

type verificationRepository struct {
	db *gorm.DB
}

// NewVerificationRepository 创建商户认证仓储实例
func NewVerificationRepository(db *gorm.DB) VerificationRepository {
	return &verificationRepository{db: db}
}

func (r *verificationRepository) Create(v *model.Dh114Verification) error {
	return r.db.Create(v).Error
}

func (r *verificationRepository) FindByID(id uint) (*model.Dh114Verification, error) {
	var v model.Dh114Verification
	if err := r.db.First(&v, id).Error; err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *verificationRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Dh114Verification{}).Where("id = ?", id).Updates(fields).Error
}

func (r *verificationRepository) Delete(id uint) error {
	return r.db.Delete(&model.Dh114Verification{}, id).Error
}

func (r *verificationRepository) List(query VerificationListQuery, pagination *utils.Pagination) ([]model.Dh114Verification, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Dh114Verification
	var total int64

	q := r.db.Model(&model.Dh114Verification{})
	if query.Dh114ID > 0 {
		q = q.Where("dh114_id = ?", query.Dh114ID)
	}
	if query.UserID > 0 {
		q = q.Where("user_id = ?", query.UserID)
	}
	if query.VerificationType != "" {
		q = q.Where("verification_type = ?", query.VerificationType)
	}
	if query.Status != nil {
		q = q.Where("status = ?", *query.Status)
	}
	if query.Keyword != "" {
		like := "%" + query.Keyword + "%"
		q = q.Where("verification_no ILIKE ? OR business_name ILIKE ? OR license_no ILIKE ? OR legal_person ILIKE ? OR brand_name ILIKE ?", like, like, like, like, like)
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

func (r *verificationRepository) ListByDh114(dh114ID uint, pagination *utils.Pagination) ([]model.Dh114Verification, int64, error) {
	return r.List(VerificationListQuery{Dh114ID: dh114ID}, pagination)
}

func (r *verificationRepository) ListByUser(userID uint, pagination *utils.Pagination) ([]model.Dh114Verification, int64, error) {
	return r.List(VerificationListQuery{UserID: userID}, pagination)
}

func (r *verificationRepository) FindLatestByDh114(dh114ID uint, vType string) (*model.Dh114Verification, error) {
	var v model.Dh114Verification
	q := r.db.Where("dh114_id = ?", dh114ID)
	if vType != "" {
		q = q.Where("verification_type = ?", vType)
	}
	if err := q.Order("created_at DESC, id DESC").First(&v).Error; err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *verificationRepository) UpdateAudit(id uint, status int, auditRemark string, auditedBy uint, auditedAt interface{}, validUntil interface{}) error {
	fields := map[string]interface{}{
		"status":       status,
		"audit_remark": auditRemark,
		"audited_by":   auditedBy,
		"audited_at":   auditedAt,
	}
	if validUntil != nil {
		fields["valid_until"] = validUntil
	}
	return r.db.Model(&model.Dh114Verification{}).Where("id = ?", id).Updates(fields).Error
}
