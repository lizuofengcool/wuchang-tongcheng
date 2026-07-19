// Package repository 公司信息 + 企业认证数据访问层
// 依据 v3.2.1 架构方案：对标 BOSS直聘公司主页/认证
package repository

import (
	"wuchang-tongcheng/internal/modules/job/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// CompanyRepository 公司仓储接口
type CompanyRepository interface {
	Create(c *model.JobCompany) error
	FindByID(id uint) (*model.JobCompany, error)
	FindByUserID(userID uint) (*model.JobCompany, error)
	Update(c *model.JobCompany) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(query CompanyListQuery, pagination *utils.Pagination) ([]model.JobCompany, int64, error)

	// 关注/取关
	FavExists(userID, companyID uint) (bool, error)
	CreateFav(fav *model.JobFavorite) error
	DeleteFav(userID, companyID uint) error
	IncrFollowerCount(companyID uint) error
	DecrFollowerCount(companyID uint) error

	// 统计
	IncrJobCount(companyID uint) error
	DecrJobCount(companyID uint) error
	IncrActiveJobCount(companyID uint) error
	DecrActiveJobCount(companyID uint) error
	IncrTotalHiredCount(companyID uint, n int) error
}

// CompanyListQuery 公司列表查询
type CompanyListQuery struct {
	UserID   uint
	Status   *int
	Level    *int
	Industry string
	Keyword  string
}

type companyRepository struct {
	db *gorm.DB
}

// NewCompanyRepository 创建公司仓储实例
func NewCompanyRepository(db *gorm.DB) CompanyRepository {
	return &companyRepository{db: db}
}

func (r *companyRepository) Create(c *model.JobCompany) error {
	return r.db.Create(c).Error
}

func (r *companyRepository) FindByID(id uint) (*model.JobCompany, error) {
	var c model.JobCompany
	if err := r.db.First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *companyRepository) FindByUserID(userID uint) (*model.JobCompany, error) {
	var c model.JobCompany
	if err := r.db.Where("user_id = ?", userID).First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *companyRepository) Update(c *model.JobCompany) error {
	return r.db.Save(c).Error
}

func (r *companyRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.JobCompany{}).Where("id = ?", id).Updates(fields).Error
}

func (r *companyRepository) Delete(id uint) error {
	return r.db.Delete(&model.JobCompany{}, id).Error
}

func (r *companyRepository) List(query CompanyListQuery, pagination *utils.Pagination) ([]model.JobCompany, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.JobCompany
	var total int64

	q := r.db.Model(&model.JobCompany{})
	if query.UserID > 0 {
		q = q.Where("user_id = ?", query.UserID)
	}
	if query.Status != nil {
		q = q.Where("status = ?", *query.Status)
	} else {
		q = q.Where("status = ?", model.CompanyStatusApproved)
	}
	if query.Level != nil {
		q = q.Where("level >= ?", *query.Level)
	}
	if query.Industry != "" {
		q = q.Where("industry ILIKE ?", "%"+query.Industry+"%")
	}
	if query.Keyword != "" {
		like := "%" + query.Keyword + "%"
		q = q.Where("name ILIKE ? OR short_name ILIKE ? OR description ILIKE ?", like, like, like)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("level DESC, verified_at DESC, id DESC").
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ===== 关注 =====

func (r *companyRepository) FavExists(userID, companyID uint) (bool, error) {
	var count int64
	err := r.db.Model(&model.JobFavorite{}).
		Where("user_id = ? AND company_id = ? AND favorite_type = ?", userID, companyID, model.FavoriteTypeCompany).
		Count(&count).Error
	return count > 0, err
}

func (r *companyRepository) CreateFav(fav *model.JobFavorite) error {
	return r.db.Create(fav).Error
}

func (r *companyRepository) DeleteFav(userID, companyID uint) error {
	return r.db.Where("user_id = ? AND company_id = ? AND favorite_type = ?", userID, companyID, model.FavoriteTypeCompany).
		Delete(&model.JobFavorite{}).Error
}

func (r *companyRepository) IncrFollowerCount(companyID uint) error {
	return r.db.Model(&model.JobCompany{}).Where("id = ?", companyID).
		UpdateColumn("follower_count", gorm.Expr("follower_count + 1")).Error
}

func (r *companyRepository) DecrFollowerCount(companyID uint) error {
	return r.db.Model(&model.JobCompany{}).Where("id = ? AND follower_count > 0", companyID).
		UpdateColumn("follower_count", gorm.Expr("follower_count - 1")).Error
}

// ===== 统计 =====

func (r *companyRepository) IncrJobCount(companyID uint) error {
	return r.db.Model(&model.JobCompany{}).Where("id = ?", companyID).
		UpdateColumn("job_count", gorm.Expr("job_count + 1")).Error
}

func (r *companyRepository) DecrJobCount(companyID uint) error {
	return r.db.Model(&model.JobCompany{}).Where("id = ? AND job_count > 0", companyID).
		UpdateColumn("job_count", gorm.Expr("job_count - 1")).Error
}

func (r *companyRepository) IncrActiveJobCount(companyID uint) error {
	return r.db.Model(&model.JobCompany{}).Where("id = ?", companyID).
		UpdateColumn("active_job_count", gorm.Expr("active_job_count + 1")).Error
}

func (r *companyRepository) DecrActiveJobCount(companyID uint) error {
	return r.db.Model(&model.JobCompany{}).Where("id = ? AND active_job_count > 0", companyID).
		UpdateColumn("active_job_count", gorm.Expr("active_job_count - 1")).Error
}

func (r *companyRepository) IncrTotalHiredCount(companyID uint, n int) error {
	return r.db.Model(&model.JobCompany{}).Where("id = ?", companyID).
		UpdateColumn("total_hired_count", gorm.Expr("total_hired_count + ?", n)).Error
}

// ===== 企业认证 =====

// CertificationRepository 企业认证仓储接口
type CertificationRepository interface {
	Create(c *model.JobCertification) error
	FindByID(id uint) (*model.JobCertification, error)
	Update(id uint, fields map[string]interface{}) error
	ListByCompanyID(companyID uint) ([]model.JobCertification, error)
	List(query CertificationListQuery, pagination *utils.Pagination) ([]model.JobCertification, int64, error)
}

// CertificationListQuery 认证列表查询
type CertificationListQuery struct {
	CompanyID uint
	UserID    uint
	Status    *int
	CertType  string
}

type certificationRepository struct {
	db *gorm.DB
}

// NewCertificationRepository 创建企业认证仓储实例
func NewCertificationRepository(db *gorm.DB) CertificationRepository {
	return &certificationRepository{db: db}
}

func (r *certificationRepository) Create(c *model.JobCertification) error {
	return r.db.Create(c).Error
}

func (r *certificationRepository) FindByID(id uint) (*model.JobCertification, error) {
	var c model.JobCertification
	if err := r.db.First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *certificationRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.JobCertification{}).Where("id = ?", id).Updates(fields).Error
}

func (r *certificationRepository) ListByCompanyID(companyID uint) ([]model.JobCertification, error) {
	var list []model.JobCertification
	if err := r.db.Where("company_id = ?", companyID).Order("created_at DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *certificationRepository) List(query CertificationListQuery, pagination *utils.Pagination) ([]model.JobCertification, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.JobCertification
	var total int64

	q := r.db.Model(&model.JobCertification{})
	if query.CompanyID > 0 {
		q = q.Where("company_id = ?", query.CompanyID)
	}
	if query.UserID > 0 {
		q = q.Where("user_id = ?", query.UserID)
	}
	if query.Status != nil {
		q = q.Where("status = ?", *query.Status)
	}
	if query.CertType != "" {
		q = q.Where("cert_type = ?", query.CertType)
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
