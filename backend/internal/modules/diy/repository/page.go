// Package repository DIY 前端页面中台数据访问层 - 页面（page 子域）
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
package repository

import (
	"wuchang-tongcheng/internal/modules/diy/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// PageListOptions 页面列表过滤条件
type PageListOptions struct {
	UserID  uint
	Type    string
	Status  *int // nil=全部
	Keyword string
}

// PageRepository 页面仓储接口
type PageRepository interface {
	// CRUD
	Create(p *model.Page) error
	FindByID(id uint) (*model.Page, error)
	Update(p *model.Page) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	// 查询
	FindBySlug(regionID uint, slug string) (*model.Page, error)
	List(regionID uint, opts PageListOptions, pagination *utils.Pagination) ([]model.Page, int64, error)
	ListByUser(userID uint, opts PageListOptions, pagination *utils.Pagination) ([]model.Page, int64, error)
	ListByType(regionID uint, pageType string, pagination *utils.Pagination) ([]model.Page, int64, error)
	ListByRegion(regionID uint, pagination *utils.Pagination) ([]model.Page, int64, error)

	// 状态
	UpdateStatus(id uint, status int) error
}

type pageRepository struct {
	db *gorm.DB
}

// NewPageRepository 创建页面仓储实例
func NewPageRepository(db *gorm.DB) PageRepository {
	return &pageRepository{db: db}
}

func (r *pageRepository) Create(p *model.Page) error {
	return r.db.Create(p).Error
}

func (r *pageRepository) FindByID(id uint) (*model.Page, error) {
	var p model.Page
	if err := r.db.First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *pageRepository) Update(p *model.Page) error {
	return r.db.Save(p).Error
}

func (r *pageRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Page{}).Where("id = ?", id).Updates(fields).Error
}

func (r *pageRepository) Delete(id uint) error {
	return r.db.Delete(&model.Page{}, id).Error
}

func (r *pageRepository) FindBySlug(regionID uint, slug string) (*model.Page, error) {
	var p model.Page
	query := r.db.Model(&model.Page{}).Where("slug = ? AND status = ?", slug, model.PageStatusPublish)
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if err := query.First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *pageRepository) List(regionID uint, opts PageListOptions, pagination *utils.Pagination) ([]model.Page, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Page
	var total int64

	query := r.db.Model(&model.Page{})
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if opts.UserID > 0 {
		query = query.Where("user_id = ?", opts.UserID)
	}
	if opts.Type != "" {
		query = query.Where("type = ?", opts.Type)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("title ILIKE ? OR slug ILIKE ?", like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("updated_at DESC, id DESC").
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *pageRepository) ListByUser(userID uint, opts PageListOptions, pagination *utils.Pagination) ([]model.Page, int64, error) {
	opts.UserID = userID
	return r.List(0, opts, pagination)
}

func (r *pageRepository) ListByType(regionID uint, pageType string, pagination *utils.Pagination) ([]model.Page, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Page
	var total int64

	query := r.db.Model(&model.Page{}).Where("type = ?", pageType)
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("updated_at DESC, id DESC").
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *pageRepository) ListByRegion(regionID uint, pagination *utils.Pagination) ([]model.Page, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Page
	var total int64

	query := r.db.Model(&model.Page{})
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("updated_at DESC, id DESC").
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *pageRepository) UpdateStatus(id uint, status int) error {
	return r.db.Model(&model.Page{}).Where("id = ?", id).
		Update("status", status).Error
}
