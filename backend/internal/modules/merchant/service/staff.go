// Package service 商户中台业务逻辑层 - 员工
// 对标美团/大众点评商家员工管理
package service

import (
	"errors"

	"wuchang-tongcheng/internal/modules/merchant/dto"
	"wuchang-tongcheng/internal/modules/merchant/model"
	"wuchang-tongcheng/internal/modules/merchant/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrStaffNotFound      = errors.New("员工不存在")
	ErrStaffNoPermission  = errors.New("无权操作员工")
	ErrStaffAlreadyExists = errors.New("员工已存在")
	ErrStaffRoleInvalid   = errors.New("员工角色无效")
)

// StaffService 员工业务接口
type StaffService interface {
	Create(req *dto.CreateStaffRequest) (*dto.StaffInfo, error)
	Update(id uint, operatorID uint, req *dto.UpdateStaffRequest) error
	Delete(id uint, operatorID uint) error
	GetByID(id uint) (*dto.StaffInfo, error)
	List(req *dto.StaffListRequest) (*utils.Pagination, []dto.StaffInfo, error)
	ListByShop(shopID uint) ([]dto.StaffInfo, error)
	ListByUser(userID uint) ([]dto.StaffInfo, error)

	AssignPermissions(id uint, operatorID uint, permissions interface{}) error
	SwitchRole(id uint, operatorID uint, role string) error
}

type staffService struct {
	repo repository.StaffRepository
}

// NewStaffService 创建员工 service 实例
func NewStaffService(repo repository.StaffRepository) StaffService {
	return &staffService{repo: repo}
}

// staffRoleText 角色文本
func staffRoleText(role string) string {
	switch role {
	case model.StaffRoleOwner:
		return "店主"
	case model.StaffRoleManager:
		return "管理员"
	case model.StaffRoleClerk:
		return "店员"
	}
	return ""
}

// staffStatusText 状态文本
func staffStatusText(status int) string {
	switch status {
	case model.StaffStatusActive:
		return "在职"
	case model.StaffStatusStopped:
		return "停用"
	}
	return ""
}

// Create 添加员工
func (s *staffService) Create(req *dto.CreateStaffRequest) (*dto.StaffInfo, error) {
	// 检查是否已存在
	if existing, err := s.repo.FindByShopAndUser(req.ShopID, req.UserID); err == nil && existing != nil {
		return nil, ErrStaffAlreadyExists
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	role := req.Role
	if role == "" {
		role = model.StaffRoleClerk
	}

	var permissionsJSON model.JSONB
	if req.Permissions != nil {
		if b, err := model.FromJSON(req.Permissions); err == nil {
			permissionsJSON = b
		}
	}

	staff := &model.Staff{
		ShopID:      req.ShopID,
		UserID:      req.UserID,
		Role:        role,
		Permissions: permissionsJSON,
		Status:      model.StaffStatusActive,
	}
	if err := s.repo.Create(staff); err != nil {
		return nil, err
	}
	return s.toInfo(staff), nil
}

// Update 更新员工
func (s *staffService) Update(id uint, operatorID uint, req *dto.UpdateStaffRequest) error {
	staff, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrStaffNotFound
		}
		return err
	}
	fields := make(map[string]interface{})
	if req.Role != nil {
		fields["role"] = *req.Role
	}
	if req.Status != nil {
		fields["status"] = *req.Status
	}
	if req.Permissions != nil {
		if b, err := model.FromJSON(req.Permissions); err == nil {
			fields["permissions"] = b
		}
	}
	if len(fields) == 0 {
		return nil
	}
	_ = staff
	return s.repo.UpdateFields(id, fields)
}

// Delete 删除员工
func (s *staffService) Delete(id uint, operatorID uint) error {
	_, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrStaffNotFound
		}
		return err
	}
	return s.repo.Delete(id)
}

// GetByID 员工详情
func (s *staffService) GetByID(id uint) (*dto.StaffInfo, error) {
	staff, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrStaffNotFound
		}
		return nil, err
	}
	return s.toInfo(staff), nil
}

// List 员工列表
func (s *staffService) List(req *dto.StaffListRequest) (*utils.Pagination, []dto.StaffInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.StaffListOptions{
		ShopID: req.ShopID,
		UserID: req.UserID,
		Role:   req.Role,
		Status: req.Status,
	}
	list, total, err := s.repo.List(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	infos := make([]dto.StaffInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *s.toInfo(&list[i]))
	}
	return pagination, infos, nil
}

// ListByShop 按店铺查询
func (s *staffService) ListByShop(shopID uint) ([]dto.StaffInfo, error) {
	list, err := s.repo.FindByShopID(shopID)
	if err != nil {
		return nil, err
	}
	infos := make([]dto.StaffInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *s.toInfo(&list[i]))
	}
	return infos, nil
}

// ListByUser 按用户查询
func (s *staffService) ListByUser(userID uint) ([]dto.StaffInfo, error) {
	list, err := s.repo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}
	infos := make([]dto.StaffInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *s.toInfo(&list[i]))
	}
	return infos, nil
}

// AssignPermissions 权限分配
func (s *staffService) AssignPermissions(id uint, operatorID uint, permissions interface{}) error {
	staff, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrStaffNotFound
		}
		return err
	}
	if permissions == nil {
		return nil
	}
	b, err := model.FromJSON(permissions)
	if err != nil {
		return err
	}
	_ = staff
	return s.repo.UpdateFields(id, map[string]interface{}{"permissions": b})
}

// SwitchRole 角色切换
func (s *staffService) SwitchRole(id uint, operatorID uint, role string) error {
	if role != model.StaffRoleOwner && role != model.StaffRoleManager && role != model.StaffRoleClerk {
		return ErrStaffRoleInvalid
	}
	_, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrStaffNotFound
		}
		return err
	}
	return s.repo.UpdateFields(id, map[string]interface{}{"role": role})
}

// toInfo 模型转 DTO
func (s *staffService) toInfo(m *model.Staff) *dto.StaffInfo {
	var perms interface{}
	if m.Permissions != nil && len(m.Permissions) > 0 {
		var parsed interface{}
		if _ = m.Permissions.Parse(&parsed); parsed != nil {
			perms = parsed
		} else {
			perms = m.Permissions.Bytes()
		}
	}
	if perms == nil {
		perms = []interface{}{}
	}
	return &dto.StaffInfo{
		ID:          m.ID,
		ShopID:      m.ShopID,
		UserID:      m.UserID,
		Role:        m.Role,
		RoleText:    staffRoleText(m.Role),
		Permissions: perms,
		Status:      m.Status,
		StatusText:  staffStatusText(m.Status),
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}
