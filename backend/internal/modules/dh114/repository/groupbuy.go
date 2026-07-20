// Package repository 同城114数据访问层 - 团购
package repository

import (
	"wuchang-tongcheng/internal/modules/dh114/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// GroupbuyRepository 团购仓储接口
type GroupbuyRepository interface {
	Create(g *model.Dh114Groupbuy) error
	FindByID(id uint) (*model.Dh114Groupbuy, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(regionID uint, query GroupbuyListQuery, pagination *utils.Pagination) ([]model.Dh114Groupbuy, int64, error)
	ListByDh114(regionID uint, dh114ID uint, pagination *utils.Pagination) ([]model.Dh114Groupbuy, int64, error)
	ListHot(regionID uint, pagination *utils.Pagination) ([]model.Dh114Groupbuy, int64, error)
	AdminList(query GroupbuyAdminListQuery, pagination *utils.Pagination) ([]model.Dh114Groupbuy, int64, error)
	IncrViewCount(id uint) error
	IncrFavCount(id uint) error
	IncrSoldCount(id uint, count int) error
}

// GroupbuyListQuery 团购列表查询
type GroupbuyListQuery struct {
	Dh114ID   uint
	Status    *int
	Featured  *bool
	MinPrice  float64
	MaxPrice  float64
	Keyword   string
	Sort      string
}

// GroupbuyAdminListQuery 管理后台团购列表查询
type GroupbuyAdminListQuery struct {
	Dh114ID     uint
	Status      *int
	AuditStatus *int
	Keyword     string
}

type groupbuyRepository struct {
	db *gorm.DB
}

// NewGroupbuyRepository 创建团购仓储实例
func NewGroupbuyRepository(db *gorm.DB) GroupbuyRepository {
	return &groupbuyRepository{db: db}
}

func (r *groupbuyRepository) Create(g *model.Dh114Groupbuy) error {
	return r.db.Create(g).Error
}

func (r *groupbuyRepository) FindByID(id uint) (*model.Dh114Groupbuy, error) {
	var g model.Dh114Groupbuy
	if err := r.db.First(&g, id).Error; err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *groupbuyRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Dh114Groupbuy{}).Where("id = ?", id).Updates(fields).Error
}

func (r *groupbuyRepository) Delete(id uint) error {
	return r.db.Delete(&model.Dh114Groupbuy{}, id).Error
}

func (r *groupbuyRepository) List(regionID uint, query GroupbuyListQuery, pagination *utils.Pagination) ([]model.Dh114Groupbuy, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Dh114Groupbuy
	var total int64

	q := r.db.Model(&model.Dh114Groupbuy{})
	if regionID > 0 {
		q = q.Where("region_id = ?", regionID)
	}
	if query.Dh114ID > 0 {
		q = q.Where("dh114_id = ?", query.Dh114ID)
	}
	if query.Status != nil {
		q = q.Where("status = ?", *query.Status)
	} else {
		q = q.Where("status = ?", model.GroupbuyStatusPublished)
	}
	if query.Featured != nil {
		q = q.Where("featured = ?", *query.Featured)
	}
	if query.MinPrice > 0 {
		q = q.Where("groupbuy_price >= ?", query.MinPrice)
	}
	if query.MaxPrice > 0 {
		q = q.Where("groupbuy_price <= ?", query.MaxPrice)
	}
	if query.Keyword != "" {
		like := "%" + query.Keyword + "%"
		q = q.Where("title ILIKE ? OR description ILIKE ?", like, like)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderClause := "published_at DESC, id DESC"
	switch query.Sort {
	case "price_asc":
		orderClause = "groupbuy_price ASC, id DESC"
	case "price_desc":
		orderClause = "groupbuy_price DESC, id DESC"
	case "sold_desc":
		orderClause = "sold_count DESC, id DESC"
	case "popular":
		orderClause = "view_count DESC, id DESC"
	}
	orderClause = "featured DESC, " + orderClause

	if err := q.Scopes(utils.Paginate(pagination)).Order(orderClause).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *groupbuyRepository) ListByDh114(regionID uint, dh114ID uint, pagination *utils.Pagination) ([]model.Dh114Groupbuy, int64, error) {
	return r.List(regionID, GroupbuyListQuery{
		Dh114ID: dh114ID,
		Status:  intPtrDh114(model.GroupbuyStatusPublished),
	}, pagination)
}

func (r *groupbuyRepository) ListHot(regionID uint, pagination *utils.Pagination) ([]model.Dh114Groupbuy, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Dh114Groupbuy
	var total int64

	q := r.db.Model(&model.Dh114Groupbuy{}).
		Where("status = ?", model.GroupbuyStatusPublished).
		Where("audit_status = ?", model.AuditApproved)
	if regionID > 0 {
		q = q.Where("region_id = ?", regionID)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("featured DESC, sold_count DESC, view_count DESC, id DESC").
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *groupbuyRepository) AdminList(query GroupbuyAdminListQuery, pagination *utils.Pagination) ([]model.Dh114Groupbuy, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Dh114Groupbuy
	var total int64

	q := r.db.Model(&model.Dh114Groupbuy{})
	if query.Dh114ID > 0 {
		q = q.Where("dh114_id = ?", query.Dh114ID)
	}
	if query.Status != nil {
		q = q.Where("status = ?", *query.Status)
	}
	if query.AuditStatus != nil {
		q = q.Where("audit_status = ?", *query.AuditStatus)
	}
	if query.Keyword != "" {
		like := "%" + query.Keyword + "%"
		q = q.Where("title ILIKE ? OR groupbuy_no ILIKE ?", like, like)
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

func (r *groupbuyRepository) IncrViewCount(id uint) error {
	return r.db.Model(&model.Dh114Groupbuy{}).Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

func (r *groupbuyRepository) IncrFavCount(id uint) error {
	return r.db.Model(&model.Dh114Groupbuy{}).Where("id = ?", id).
		UpdateColumn("fav_count", gorm.Expr("fav_count + 1")).Error
}

func (r *groupbuyRepository) IncrSoldCount(id uint, count int) error {
	return r.db.Model(&model.Dh114Groupbuy{}).Where("id = ?", id).
		UpdateColumn("sold_count", gorm.Expr("sold_count + ?", count)).Error
}
