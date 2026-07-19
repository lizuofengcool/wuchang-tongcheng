// Package repository 技能/福利/职位模板/分类/薪资范围 数据访问层
// 依据 v3.2.1 架构方案：对标 BOSS直聘标准职位库/分类/技能/福利
package repository

import (
	"wuchang-tongcheng/internal/modules/job/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// ===== Skill =====

// SkillRepository 技能标签仓储接口
type SkillRepository interface {
	Create(s *model.JobSkill) error
	FindByID(id uint) (*model.JobSkill, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(query SkillListQuery, pagination *utils.Pagination) ([]model.JobSkill, int64, error)
	ListHot(limit int) ([]model.JobSkill, error)
	IncrUseCount(id uint) error
}

// SkillListQuery 技能列表查询
type SkillListQuery struct {
	Category string
	Status   *int
	Keyword  string
}

type skillRepository struct {
	db *gorm.DB
}

// NewSkillRepository 创建技能仓储实例
func NewSkillRepository(db *gorm.DB) SkillRepository {
	return &skillRepository{db: db}
}

func (r *skillRepository) Create(s *model.JobSkill) error {
	return r.db.Create(s).Error
}

func (r *skillRepository) FindByID(id uint) (*model.JobSkill, error) {
	var s model.JobSkill
	if err := r.db.First(&s, id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *skillRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.JobSkill{}).Where("id = ?", id).Updates(fields).Error
}

func (r *skillRepository) Delete(id uint) error {
	return r.db.Delete(&model.JobSkill{}, id).Error
}

func (r *skillRepository) List(query SkillListQuery, pagination *utils.Pagination) ([]model.JobSkill, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 50)
	}
	var list []model.JobSkill
	var total int64

	q := r.db.Model(&model.JobSkill{})
	if query.Category != "" {
		q = q.Where("category = ?", query.Category)
	}
	if query.Status != nil {
		q = q.Where("status = ?", *query.Status)
	} else {
		q = q.Where("status = ?", model.SkillStatusEnabled)
	}
	if query.Keyword != "" {
		q = q.Where("name ILIKE ?", "%"+query.Keyword+"%")
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("is_hot DESC, use_count DESC, sort ASC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *skillRepository) ListHot(limit int) ([]model.JobSkill, error) {
	if limit <= 0 {
		limit = 10
	}
	var list []model.JobSkill
	if err := r.db.Where("status = ? AND is_hot = ?", model.SkillStatusEnabled, true).
		Order("use_count DESC, id DESC").Limit(limit).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *skillRepository) IncrUseCount(id uint) error {
	return r.db.Model(&model.JobSkill{}).Where("id = ?", id).
		UpdateColumn("use_count", gorm.Expr("use_count + 1")).Error
}

// ===== Benefit =====

// BenefitRepository 福利标签仓储接口
type BenefitRepository interface {
	Create(b *model.JobBenefit) error
	FindByID(id uint) (*model.JobBenefit, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(query BenefitListQuery, pagination *utils.Pagination) ([]model.JobBenefit, int64, error)
	ListHot(limit int) ([]model.JobBenefit, error)
	IncrUseCount(id uint) error
}

// BenefitListQuery 福利列表查询
type BenefitListQuery struct {
	Category string
	Status   *int
	Keyword  string
}

type benefitRepository struct {
	db *gorm.DB
}

// NewBenefitRepository 创建福利仓储实例
func NewBenefitRepository(db *gorm.DB) BenefitRepository {
	return &benefitRepository{db: db}
}

func (r *benefitRepository) Create(b *model.JobBenefit) error {
	return r.db.Create(b).Error
}

func (r *benefitRepository) FindByID(id uint) (*model.JobBenefit, error) {
	var b model.JobBenefit
	if err := r.db.First(&b, id).Error; err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *benefitRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.JobBenefit{}).Where("id = ?", id).Updates(fields).Error
}

func (r *benefitRepository) Delete(id uint) error {
	return r.db.Delete(&model.JobBenefit{}, id).Error
}

func (r *benefitRepository) List(query BenefitListQuery, pagination *utils.Pagination) ([]model.JobBenefit, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 50)
	}
	var list []model.JobBenefit
	var total int64

	q := r.db.Model(&model.JobBenefit{})
	if query.Category != "" {
		q = q.Where("category = ?", query.Category)
	}
	if query.Status != nil {
		q = q.Where("status = ?", *query.Status)
	} else {
		q = q.Where("status = ?", model.BenefitStatusEnabled)
	}
	if query.Keyword != "" {
		q = q.Where("name ILIKE ?", "%"+query.Keyword+"%")
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("is_hot DESC, use_count DESC, sort ASC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *benefitRepository) ListHot(limit int) ([]model.JobBenefit, error) {
	if limit <= 0 {
		limit = 10
	}
	var list []model.JobBenefit
	if err := r.db.Where("status = ? AND is_hot = ?", model.BenefitStatusEnabled, true).
		Order("use_count DESC, id DESC").Limit(limit).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *benefitRepository) IncrUseCount(id uint) error {
	return r.db.Model(&model.JobBenefit{}).Where("id = ?", id).
		UpdateColumn("use_count", gorm.Expr("use_count + 1")).Error
}

// ===== Position 职位模板 =====

// PositionRepository 职位模板仓储接口
type PositionRepository interface {
	Create(p *model.JobPosition) error
	FindByID(id uint) (*model.JobPosition, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(query PositionListQuery, pagination *utils.Pagination) ([]model.JobPosition, int64, error)
	IncrUseCount(id uint) error
}

// PositionListQuery 职位模板列表查询
type PositionListQuery struct {
	CategoryID uint
	Status     *int
	Keyword    string
}

type positionRepository struct {
	db *gorm.DB
}

// NewPositionRepository 创建职位模板仓储实例
func NewPositionRepository(db *gorm.DB) PositionRepository {
	return &positionRepository{db: db}
}

func (r *positionRepository) Create(p *model.JobPosition) error {
	return r.db.Create(p).Error
}

func (r *positionRepository) FindByID(id uint) (*model.JobPosition, error) {
	var p model.JobPosition
	if err := r.db.First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *positionRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.JobPosition{}).Where("id = ?", id).Updates(fields).Error
}

func (r *positionRepository) Delete(id uint) error {
	return r.db.Delete(&model.JobPosition{}, id).Error
}

func (r *positionRepository) List(query PositionListQuery, pagination *utils.Pagination) ([]model.JobPosition, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 50)
	}
	var list []model.JobPosition
	var total int64

	q := r.db.Model(&model.JobPosition{})
	if query.CategoryID > 0 {
		q = q.Where("category_id = ?", query.CategoryID)
	}
	if query.Status != nil {
		q = q.Where("status = ?", *query.Status)
	} else {
		q = q.Where("status = ?", model.PositionTemplateStatusEnabled)
	}
	if query.Keyword != "" {
		q = q.Where("name ILIKE ?", "%"+query.Keyword+"%")
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("use_count DESC, sort ASC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *positionRepository) IncrUseCount(id uint) error {
	return r.db.Model(&model.JobPosition{}).Where("id = ?", id).
		UpdateColumn("use_count", gorm.Expr("use_count + 1")).Error
}

// ===== Category 职位分类 =====

// CategoryRepository 职位分类仓储接口
type CategoryRepository interface {
	Create(c *model.JobCategory) error
	FindByID(id uint) (*model.JobCategory, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(query CategoryListQuery, pagination *utils.Pagination) ([]model.JobCategory, int64, error)
	ListTree(parentID uint) ([]model.JobCategory, error)
	IncrJobCount(id uint) error
	DecrJobCount(id uint) error
}

// CategoryListQuery 分类列表查询
type CategoryListQuery struct {
	ParentID uint
	Status   *int
	Keyword  string
}

type categoryRepository struct {
	db *gorm.DB
}

// NewCategoryRepository 创建职位分类仓储实例
func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) Create(c *model.JobCategory) error {
	return r.db.Create(c).Error
}

func (r *categoryRepository) FindByID(id uint) (*model.JobCategory, error) {
	var c model.JobCategory
	if err := r.db.First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *categoryRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.JobCategory{}).Where("id = ?", id).Updates(fields).Error
}

func (r *categoryRepository) Delete(id uint) error {
	return r.db.Delete(&model.JobCategory{}, id).Error
}

func (r *categoryRepository) List(query CategoryListQuery, pagination *utils.Pagination) ([]model.JobCategory, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 100)
	}
	var list []model.JobCategory
	var total int64

	q := r.db.Model(&model.JobCategory{})
	if query.ParentID > 0 {
		q = q.Where("parent_id = ?", query.ParentID)
	}
	if query.Status != nil {
		q = q.Where("status = ?", *query.Status)
	} else {
		q = q.Where("status = ?", model.CategoryStatusEnabled)
	}
	if query.Keyword != "" {
		q = q.Where("name ILIKE ?", "%"+query.Keyword+"%")
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("sort ASC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *categoryRepository) ListTree(parentID uint) ([]model.JobCategory, error) {
	var list []model.JobCategory
	if err := r.db.Where("parent_id = ? AND status = ?", parentID, model.CategoryStatusEnabled).
		Order("sort ASC, id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *categoryRepository) IncrJobCount(id uint) error {
	return r.db.Model(&model.JobCategory{}).Where("id = ?", id).
		UpdateColumn("job_count", gorm.Expr("job_count + 1")).Error
}

func (r *categoryRepository) DecrJobCount(id uint) error {
	return r.db.Model(&model.JobCategory{}).Where("id = ? AND job_count > 0", id).
		UpdateColumn("job_count", gorm.Expr("job_count - 1")).Error
}

// ===== SalaryRange 薪资范围 =====

// SalaryRangeRepository 薪资范围仓储接口
type SalaryRangeRepository interface {
	Create(s *model.JobSalaryRange) error
	FindByID(id uint) (*model.JobSalaryRange, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(query SalaryRangeListQuery, pagination *utils.Pagination) ([]model.JobSalaryRange, int64, error)
}

// SalaryRangeListQuery 薪资范围列表查询
type SalaryRangeListQuery struct {
	Status *int
	Period string
}

type salaryRangeRepository struct {
	db *gorm.DB
}

// NewSalaryRangeRepository 创建薪资范围仓储实例
func NewSalaryRangeRepository(db *gorm.DB) SalaryRangeRepository {
	return &salaryRangeRepository{db: db}
}

func (r *salaryRangeRepository) Create(s *model.JobSalaryRange) error {
	return r.db.Create(s).Error
}

func (r *salaryRangeRepository) FindByID(id uint) (*model.JobSalaryRange, error) {
	var s model.JobSalaryRange
	if err := r.db.First(&s, id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *salaryRangeRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.JobSalaryRange{}).Where("id = ?", id).Updates(fields).Error
}

func (r *salaryRangeRepository) Delete(id uint) error {
	return r.db.Delete(&model.JobSalaryRange{}, id).Error
}

func (r *salaryRangeRepository) List(query SalaryRangeListQuery, pagination *utils.Pagination) ([]model.JobSalaryRange, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 50)
	}
	var list []model.JobSalaryRange
	var total int64

	q := r.db.Model(&model.JobSalaryRange{})
	if query.Status != nil {
		q = q.Where("status = ?", *query.Status)
	} else {
		q = q.Where("status = ?", model.SalaryRangeStatusEnabled)
	}
	if query.Period != "" {
		q = q.Where("period = ?", query.Period)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("sort ASC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
