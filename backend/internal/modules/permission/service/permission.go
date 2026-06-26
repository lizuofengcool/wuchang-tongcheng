// Package service 权限业务逻辑层
package service

import (
	"errors"

	"wuchang-tongcheng/internal/modules/permission/dto"
	"wuchang-tongcheng/internal/modules/permission/model"
	"wuchang-tongcheng/internal/modules/permission/repository"

	"gorm.io/gorm"
)

var (
	ErrRoleNotFound       = errors.New("角色不存在")
	ErrRoleCodeExists     = errors.New("角色编码已存在")
	ErrPermissionNotFound = errors.New("权限不存在")
	ErrPermCodeExists     = errors.New("权限编码已存在")
)

// PermissionService 权限业务逻辑接口
type PermissionService interface {
	// 角色
	CreateRole(req *dto.CreateRoleRequest) (*dto.RoleInfo, error)
	UpdateRole(id uint, req *dto.UpdateRoleRequest) error
	DeleteRole(id uint) error
	GetRoleByID(id uint) (*dto.RoleInfo, error)
	ListRoles() ([]dto.RoleInfo, error)
	// 权限
	CreatePermission(req *dto.CreatePermissionRequest) (*dto.PermissionInfo, error)
	UpdatePermission(id uint, fields map[string]interface{}) error
	DeletePermission(id uint) error
	ListPermissions() ([]dto.PermissionInfo, error)
	// 分配
	AssignRolesToUser(req *dto.AssignRolesRequest) error
	AssignPermissionsToRole(req *dto.AssignPermissionsRequest) error
	GetRolesByUserID(userID uint) ([]dto.RoleInfo, error)
	GetPermissionsByUserID(userID uint) ([]dto.PermissionInfo, error)
	GetPermissionCodesByUserID(userID uint) ([]string, error)
	GetRoleCodesByUserID(userID uint) ([]string, error)
	// 校验
	HasPermission(userID uint, permCode string) (bool, error)
}

type permissionService struct {
	repo repository.PermissionRepository
}

// NewPermissionService 创建权限服务
func NewPermissionService(repo repository.PermissionRepository) PermissionService {
	return &permissionService{repo: repo}
}

func toRoleInfo(r *model.Role) *dto.RoleInfo {
	return &dto.RoleInfo{
		ID:          r.ID,
		Name:        r.Name,
		Code:        r.Code,
		Description: r.Description,
		Sort:        r.Sort,
		Status:      r.Status,
	}
}

func toPermInfo(p *model.Permission) *dto.PermissionInfo {
	return &dto.PermissionInfo{
		ID:       p.ID,
		Name:     p.Name,
		Code:     p.Code,
		Type:     p.Type,
		ParentID: p.ParentID,
		Path:     p.Path,
		Method:   p.Method,
		Sort:     p.Sort,
		Status:   p.Status,
	}
}

// ===== 角色 =====

func (s *permissionService) CreateRole(req *dto.CreateRoleRequest) (*dto.RoleInfo, error) {
	if _, err := s.repo.FindRoleByCode(req.Code); err == nil {
		return nil, ErrRoleCodeExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	status := req.Status
	if status == 0 {
		status = 1
	}
	role := &model.Role{
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		Sort:        req.Sort,
		Status:      status,
	}
	if err := s.repo.CreateRole(role); err != nil {
		return nil, err
	}
	return toRoleInfo(role), nil
}

func (s *permissionService) UpdateRole(id uint, req *dto.UpdateRoleRequest) error {
	if _, err := s.repo.FindRoleByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRoleNotFound
		}
		return err
	}
	fields := map[string]interface{}{}
	if req.Name != "" {
		fields["name"] = req.Name
	}
	if req.Description != "" {
		fields["description"] = req.Description
	}
	fields["sort"] = req.Sort
	if req.Status == 0 || req.Status == 1 {
		fields["status"] = req.Status
	}
	return s.repo.UpdateRoleFields(id, fields)
}

func (s *permissionService) DeleteRole(id uint) error {
	if _, err := s.repo.FindRoleByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRoleNotFound
		}
		return err
	}
	return s.repo.DeleteRole(id)
}

func (s *permissionService) GetRoleByID(id uint) (*dto.RoleInfo, error) {
	role, err := s.repo.FindRoleByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRoleNotFound
		}
		return nil, err
	}
	return toRoleInfo(role), nil
}

func (s *permissionService) ListRoles() ([]dto.RoleInfo, error) {
	roles, err := s.repo.FindAllRoles()
	if err != nil {
		return nil, err
	}
	result := make([]dto.RoleInfo, 0, len(roles))
	for i := range roles {
		result = append(result, *toRoleInfo(&roles[i]))
	}
	return result, nil
}

// ===== 权限 =====

func (s *permissionService) CreatePermission(req *dto.CreatePermissionRequest) (*dto.PermissionInfo, error) {
	if _, err := s.repo.FindPermissionByCode(req.Code); err == nil {
		return nil, ErrPermCodeExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	status := req.Status
	if status == 0 {
		status = 1
	}
	p := &model.Permission{
		Name:     req.Name,
		Code:     req.Code,
		Type:     req.Type,
		ParentID: req.ParentID,
		Path:     req.Path,
		Method:   req.Method,
		Sort:     req.Sort,
		Status:   status,
	}
	if err := s.repo.CreatePermission(p); err != nil {
		return nil, err
	}
	return toPermInfo(p), nil
}

func (s *permissionService) UpdatePermission(id uint, fields map[string]interface{}) error {
	if _, err := s.repo.FindPermissionByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPermissionNotFound
		}
		return err
	}
	return s.repo.UpdatePermissionFields(id, fields)
}

func (s *permissionService) DeletePermission(id uint) error {
	if _, err := s.repo.FindPermissionByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPermissionNotFound
		}
		return err
	}
	return s.repo.DeletePermission(id)
}

func (s *permissionService) ListPermissions() ([]dto.PermissionInfo, error) {
	perms, err := s.repo.FindAllPermissions()
	if err != nil {
		return nil, err
	}
	result := make([]dto.PermissionInfo, 0, len(perms))
	for i := range perms {
		result = append(result, *toPermInfo(&perms[i]))
	}
	return result, nil
}

// ===== 分配 =====

func (s *permissionService) AssignRolesToUser(req *dto.AssignRolesRequest) error {
	return s.repo.AssignRolesToUser(req.UserID, req.RoleIDs)
}

func (s *permissionService) AssignPermissionsToRole(req *dto.AssignPermissionsRequest) error {
	if _, err := s.repo.FindRoleByID(req.RoleID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRoleNotFound
		}
		return err
	}
	return s.repo.AssignPermissionsToRole(req.RoleID, req.PermissionIDs)
}

func (s *permissionService) GetRolesByUserID(userID uint) ([]dto.RoleInfo, error) {
	roles, err := s.repo.FindRolesByUserID(userID)
	if err != nil {
		return nil, err
	}
	result := make([]dto.RoleInfo, 0, len(roles))
	for i := range roles {
		result = append(result, *toRoleInfo(&roles[i]))
	}
	return result, nil
}

func (s *permissionService) GetPermissionsByUserID(userID uint) ([]dto.PermissionInfo, error) {
	perms, err := s.repo.FindPermissionsByUserID(userID)
	if err != nil {
		return nil, err
	}
	result := make([]dto.PermissionInfo, 0, len(perms))
	for i := range perms {
		result = append(result, *toPermInfo(&perms[i]))
	}
	return result, nil
}

func (s *permissionService) GetPermissionCodesByUserID(userID uint) ([]string, error) {
	return s.repo.FindPermissionCodesByUserID(userID)
}

// GetRoleCodesByUserID 获取用户角色编码列表
func (s *permissionService) GetRoleCodesByUserID(userID uint) ([]string, error) {
	return s.repo.FindRoleCodesByUserID(userID)
}

// HasPermission 校验用户是否拥有某权限
func (s *permissionService) HasPermission(userID uint, permCode string) (bool, error) {
	codes, err := s.repo.FindPermissionCodesByUserID(userID)
	if err != nil {
		return false, err
	}
	for _, c := range codes {
		if c == permCode {
			return true, nil
		}
	}
	return false, nil
}
