// Package repository 权限数据访问层
package repository

import (
	"wuchang-tongcheng/internal/modules/permission/model"

	"gorm.io/gorm"
)

// PermissionRepository 权限仓储接口
type PermissionRepository interface {
	// 角色
	CreateRole(role *model.Role) error
	FindRoleByID(id uint) (*model.Role, error)
	FindRoleByCode(code string) (*model.Role, error)
	FindAllRoles() ([]model.Role, error)
	UpdateRole(role *model.Role) error
	UpdateRoleFields(id uint, fields map[string]interface{}) error
	DeleteRole(id uint) error
	// 权限
	CreatePermission(p *model.Permission) error
	FindPermissionByID(id uint) (*model.Permission, error)
	FindPermissionByCode(code string) (*model.Permission, error)
	FindAllPermissions() ([]model.Permission, error)
	UpdatePermissionFields(id uint, fields map[string]interface{}) error
	DeletePermission(id uint) error
	// 关联：用户角色
	AssignRolesToUser(userID uint, roleIDs []uint) error
	FindRolesByUserID(userID uint) ([]model.Role, error)
	ClearUserRoles(userID uint) error
	// 关联：角色权限
	AssignPermissionsToRole(roleID uint, permIDs []uint) error
	FindPermissionsByRoleID(roleID uint) ([]model.Permission, error)
	ClearRolePermissions(roleID uint) error
	// 综合查询
	FindPermissionsByUserID(userID uint) ([]model.Permission, error)
	FindPermissionCodesByUserID(userID uint) ([]string, error)
	FindRoleCodesByUserID(userID uint) ([]string, error)
}

type permissionRepository struct {
	db *gorm.DB
}

// NewPermissionRepository 创建权限仓储
func NewPermissionRepository(db *gorm.DB) PermissionRepository {
	return &permissionRepository{db: db}
}

// ===== 角色 =====

func (r *permissionRepository) CreateRole(role *model.Role) error {
	return r.db.Create(role).Error
}

func (r *permissionRepository) FindRoleByID(id uint) (*model.Role, error) {
	var role model.Role
	if err := r.db.First(&role, id).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *permissionRepository) FindRoleByCode(code string) (*model.Role, error) {
	var role model.Role
	if err := r.db.Where("code = ?", code).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *permissionRepository) FindAllRoles() ([]model.Role, error) {
	var roles []model.Role
	if err := r.db.Order("sort ASC, id ASC").Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *permissionRepository) UpdateRole(role *model.Role) error {
	return r.db.Save(role).Error
}

func (r *permissionRepository) UpdateRoleFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Role{}).Where("id = ?", id).Updates(fields).Error
}

func (r *permissionRepository) DeleteRole(id uint) error {
	// 事务删除角色及关联
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", id).Delete(&model.RolePermission{}).Error; err != nil {
			return err
		}
		if err := tx.Where("role_id = ?", id).Delete(&model.UserRole{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Role{}, id).Error
	})
}

// ===== 权限 =====

func (r *permissionRepository) CreatePermission(p *model.Permission) error {
	return r.db.Create(p).Error
}

func (r *permissionRepository) FindPermissionByID(id uint) (*model.Permission, error) {
	var p model.Permission
	if err := r.db.First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *permissionRepository) FindPermissionByCode(code string) (*model.Permission, error) {
	var p model.Permission
	if err := r.db.Where("code = ?", code).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *permissionRepository) FindAllPermissions() ([]model.Permission, error) {
	var perms []model.Permission
	if err := r.db.Order("type ASC, sort ASC, id ASC").Find(&perms).Error; err != nil {
		return nil, err
	}
	return perms, nil
}

func (r *permissionRepository) UpdatePermissionFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Permission{}).Where("id = ?", id).Updates(fields).Error
}

func (r *permissionRepository) DeletePermission(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("permission_id = ?", id).Delete(&model.RolePermission{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Permission{}, id).Error
	})
}

// ===== 用户-角色关联 =====

func (r *permissionRepository) AssignRolesToUser(userID uint, roleIDs []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 先清空旧角色
		if err := tx.Where("user_id = ?", userID).Delete(&model.UserRole{}).Error; err != nil {
			return err
		}
		// 批量插入新角色
		if len(roleIDs) == 0 {
			return nil
		}
		userRoles := make([]model.UserRole, 0, len(roleIDs))
		for _, rid := range roleIDs {
			userRoles = append(userRoles, model.UserRole{UserID: userID, RoleID: rid})
		}
		return tx.Create(&userRoles).Error
	})
}

func (r *permissionRepository) FindRolesByUserID(userID uint) ([]model.Role, error) {
	var roles []model.Role
	err := r.db.
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", userID).
		Order("roles.sort ASC, roles.id ASC").
		Find(&roles).Error
	if err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *permissionRepository) ClearUserRoles(userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&model.UserRole{}).Error
}

// ===== 角色-权限关联 =====

func (r *permissionRepository) AssignPermissionsToRole(roleID uint, permIDs []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleID).Delete(&model.RolePermission{}).Error; err != nil {
			return err
		}
		if len(permIDs) == 0 {
			return nil
		}
		rps := make([]model.RolePermission, 0, len(permIDs))
		for _, pid := range permIDs {
			rps = append(rps, model.RolePermission{RoleID: roleID, PermissionID: pid})
		}
		return tx.Create(&rps).Error
	})
}

func (r *permissionRepository) FindPermissionsByRoleID(roleID uint) ([]model.Permission, error) {
	var perms []model.Permission
	err := r.db.
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Where("role_permissions.role_id = ?", roleID).
		Order("permissions.sort ASC, permissions.id ASC").
		Find(&perms).Error
	if err != nil {
		return nil, err
	}
	return perms, nil
}

func (r *permissionRepository) ClearRolePermissions(roleID uint) error {
	return r.db.Where("role_id = ?", roleID).Delete(&model.RolePermission{}).Error
}

// ===== 综合查询 =====

// FindPermissionsByUserID 查询用户拥有的所有权限（经角色关联）
func (r *permissionRepository) FindPermissionsByUserID(userID uint) ([]model.Permission, error) {
	var perms []model.Permission
	err := r.db.
		Distinct("permissions.*").
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Joins("JOIN user_roles ON user_roles.role_id = role_permissions.role_id").
		Where("user_roles.user_id = ? AND permissions.status = 1", userID).
		Order("permissions.sort ASC, permissions.id ASC").
		Find(&perms).Error
	if err != nil {
		return nil, err
	}
	return perms, nil
}

// FindPermissionCodesByUserID 查询用户拥有的权限编码列表
func (r *permissionRepository) FindPermissionCodesByUserID(userID uint) ([]string, error) {
	var codes []string
	err := r.db.
		Model(&model.Permission{}).
		Distinct("permissions.code").
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Joins("JOIN user_roles ON user_roles.role_id = role_permissions.role_id").
		Where("user_roles.user_id = ? AND permissions.status = 1", userID).
		Pluck("permissions.code", &codes).Error
	if err != nil {
		return nil, err
	}
	return codes, nil
}

// FindRoleCodesByUserID 查询用户拥有的角色编码列表
func (r *permissionRepository) FindRoleCodesByUserID(userID uint) ([]string, error) {
	var codes []string
	err := r.db.
		Model(&model.Role{}).
		Distinct("roles.code").
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ? AND roles.status = 1", userID).
		Pluck("roles.code", &codes).Error
	if err != nil {
		return nil, err
	}
	return codes, nil
}
