// Package repository 用户数据访问层
package repository

import (
	"wuchang-tongcheng/internal/modules/user/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// UserRepository 用户仓储接口
type UserRepository interface {
	Create(user *model.User) error
	FindByID(id uint) (*model.User, error)
	FindByUsername(username string) (*model.User, error)
	Update(user *model.User) error
	UpdateFields(id uint, fields map[string]interface{}) error
	// 管理后台
	List(pagination *utils.Pagination, keyword string, status int) ([]model.User, int64, error)
	Delete(id uint) error
}

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓储
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// Create 创建用户
func (r *userRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

// FindByID 根据ID查询用户
func (r *userRepository) FindByID(id uint) (*model.User, error) {
	var user model.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByUsername 根据用户名查询用户
func (r *userRepository) FindByUsername(username string) (*model.User, error) {
	var user model.User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// Update 更新用户
func (r *userRepository) Update(user *model.User) error {
	return r.db.Save(user).Error
}

// UpdateFields 更新指定字段
func (r *userRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.User{}).Where("id = ?", id).Updates(fields).Error
}

// List 用户分页列表（支持用户名/昵称关键词、状态筛选）
func (r *userRepository) List(pagination *utils.Pagination, keyword string, status int) ([]model.User, int64, error) {
	var list []model.User
	var total int64

	query := r.db.Model(&model.User{})

	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("username LIKE ? OR nickname LIKE ?", like, like)
	}
	if status == 0 || status == 1 {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Scopes(utils.Paginate(pagination)).Order("id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// Delete 删除用户
func (r *userRepository) Delete(id uint) error {
	return r.db.Delete(&model.User{}, id).Error
}
