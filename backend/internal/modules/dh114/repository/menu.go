// Package repository 同城114数据访问层 - 菜单/服务项目
// 菜品/服务/价格/图片/招牌
package repository

import (
	"wuchang-tongcheng/internal/modules/dh114/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// MenuRepository 菜单/服务项目仓储接口
type MenuRepository interface {
	Create(m *model.Dh114Menu) error
	FindByID(id uint) (*model.Dh114Menu, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(query MenuListQuery, pagination *utils.Pagination) ([]model.Dh114Menu, int64, error)
	ListByDh114(dh114ID uint, onlyActive bool) ([]model.Dh114Menu, error)
	ListSignature(dh114ID uint) ([]model.Dh114Menu, error)
	IncrOrderCount(id uint, count int) error
	ReplaceMenus(dh114ID uint, menus []model.Dh114Menu) error
}

// MenuListQuery 菜单列表查询
type MenuListQuery struct {
	Dh114ID    uint
	MenuType   string
	Status     *int
	IsSignature *bool
	Keyword    string
}

type menuRepository struct {
	db *gorm.DB
}

// NewMenuRepository 创建菜单/服务项目仓储实例
func NewMenuRepository(db *gorm.DB) MenuRepository {
	return &menuRepository{db: db}
}

func (r *menuRepository) Create(m *model.Dh114Menu) error {
	return r.db.Create(m).Error
}

func (r *menuRepository) FindByID(id uint) (*model.Dh114Menu, error) {
	var m model.Dh114Menu
	if err := r.db.First(&m, id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *menuRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Dh114Menu{}).Where("id = ?", id).Updates(fields).Error
}

func (r *menuRepository) Delete(id uint) error {
	return r.db.Delete(&model.Dh114Menu{}, id).Error
}

func (r *menuRepository) List(query MenuListQuery, pagination *utils.Pagination) ([]model.Dh114Menu, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 50)
	}
	var list []model.Dh114Menu
	var total int64

	q := r.db.Model(&model.Dh114Menu{})
	if query.Dh114ID > 0 {
		q = q.Where("dh114_id = ?", query.Dh114ID)
	}
	if query.MenuType != "" {
		q = q.Where("menu_type = ?", query.MenuType)
	}
	if query.Status != nil {
		q = q.Where("status = ?", *query.Status)
	}
	if query.IsSignature != nil {
		q = q.Where("is_signature = ?", *query.IsSignature)
	}
	if query.Keyword != "" {
		like := "%" + query.Keyword + "%"
		q = q.Where("name ILIKE ? OR description ILIKE ?", like, like)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("is_signature DESC, sort ASC, order_count DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *menuRepository) ListByDh114(dh114ID uint, onlyActive bool) ([]model.Dh114Menu, error) {
	var list []model.Dh114Menu
	q := r.db.Where("dh114_id = ?", dh114ID)
	if onlyActive {
		q = q.Where("status = ?", 1)
	}
	if err := q.Order("is_signature DESC, sort ASC, order_count DESC, id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *menuRepository) ListSignature(dh114ID uint) ([]model.Dh114Menu, error) {
	var list []model.Dh114Menu
	if err := r.db.Where("dh114_id = ? AND is_signature = ? AND status = ?", dh114ID, true, 1).
		Order("sort ASC, id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *menuRepository) IncrOrderCount(id uint, count int) error {
	return r.db.Model(&model.Dh114Menu{}).Where("id = ?", id).
		UpdateColumn("order_count", gorm.Expr("order_count + ?", count)).Error
}

func (r *menuRepository) ReplaceMenus(dh114ID uint, menus []model.Dh114Menu) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	if err := tx.Where("dh114_id = ?", dh114ID).Delete(&model.Dh114Menu{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	for i := range menus {
		menus[i].Dh114ID = dh114ID
		if menus[i].Sort == 0 {
			menus[i].Sort = i
		}
		if err := tx.Create(&menus[i]).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}
