// Package module 模块注册表数据访问层
package module

import (
	"errors"

	"gorm.io/gorm"
)

// ErrModuleNotFound 模块未找到错误
var ErrModuleNotFound = errors.New("模块不存在")

// Repository 模块仓储接口
type Repository interface {
	List() ([]Module, error)
	GetByName(name string) (*Module, error)
	Create(m *Module) error
	Update(m *Module) error
	Delete(name string) error
	Enable(name string) error
	Disable(name string) error
}

type repository struct {
	db *gorm.DB
}

// NewRepository 创建模块仓储
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) List() ([]Module, error) {
	var modules []Module
	err := r.db.Order("category ASC, name ASC").Find(&modules).Error
	return modules, err
}

func (r *repository) GetByName(name string) (*Module, error) {
	var m Module
	err := r.db.Where("name = ?", name).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrModuleNotFound
		}
		return nil, err
	}
	return &m, nil
}

func (r *repository) Create(m *Module) error {
	return r.db.Create(m).Error
}

func (r *repository) Update(m *Module) error {
	return r.db.Save(m).Error
}

func (r *repository) Delete(name string) error {
	return r.db.Where("name = ?", name).Delete(&Module{}).Error
}

func (r *repository) Enable(name string) error {
	return r.db.Model(&Module{}).Where("name = ?", name).Update("enabled", true).Error
}

func (r *repository) Disable(name string) error {
	return r.db.Model(&Module{}).Where("name = ?", name).Update("enabled", false).Error
}
