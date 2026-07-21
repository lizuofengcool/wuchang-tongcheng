// Package service 多租户分站业务逻辑层 - 员工
// 职责：员工 CRUD / 权限分配 / 角色管理
package service

import (
	"errors"

	"wuchang-tongcheng/internal/modules/tenant/dto"
	"wuchang-tongcheng/internal/modules/tenant/model"
	"wuchang-tongcheng/internal/modules/tenant/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrStaffNotFound       = errors.New("员工不存在")
	ErrStaffExists         = errors.New("该用户已是本分站员工")
	ErrStaffRoleInvalid    = errors.New("员工角色无效")
	ErrStaffStatusInvalid  = errors.New("员工状态不允许此操作")
	ErrStaffStationNotFnd  = errors.New("所属分站不存在")
)

// StaffService 员工业务接口
type StaffService interface {
	Create(req *dto.CreateStaffRequest) (*dto.StaffInfo, error)
	Update(id uint, req *dto.UpdateStaffRequest) error
	Delete(id uint) error
	GetByID(id uint) (*dto.StaffInfo, error)
	List(req *dto.StaffListRequest) (*utils.Pagination, []dto.StaffInfo, error)
	ListByStation(stationID uint) ([]dto.StaffInfo, error)
	ListByUser(userID uint) ([]dto.StaffInfo, error)
}

type staffService struct {
	repo        repository.StaffRepository
	stationRepo repository.StationRepository
}

// NewStaffService 创建员工 service 实例
func NewStaffService(repo repository.StaffRepository, stationRepo repository.StationRepository) StaffService {
	return &staffService{repo: repo, stationRepo: stationRepo}
}

// staffRoleText 角色文本
func staffRoleText(role string) string {
	switch role {
	case model.StaffRoleOperator:
		return "运营员"
	case model.StaffRoleManager:
		return "管理员"
	}
	return ""
}

// staffStatusText 状态文本
func staffStatusText(status int) string {
	switch status {
	case model.StaffStatusDisabled:
		return "已停用"
	case model.StaffStatusEnabled:
		return "已启用"
	}
	return ""
}

// toStaffInfo model -> dto
func toStaffInfo(s *model.Staff) *dto.StaffInfo {
	return &dto.StaffInfo{
		ID:          s.ID,
		StationID:   s.StationID,
		UserID:      s.UserID,
		Role:        s.Role,
		RoleText:    staffRoleText(s.Role),
		Permissions: configToAny(s.Permissions),
		Status:      s.Status,
		StatusText:  staffStatusText(s.Status),
		CreatedAt:   s.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   s.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// Create 创建员工
func (s *staffService) Create(req *dto.CreateStaffRequest) (*dto.StaffInfo, error) {
	// 校验分站存在
	if _, err := s.stationRepo.FindByID(req.StationID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrStaffStationNotFnd
		}
		return nil, err
	}
	// 校验同一分站不重复添加同一用户
	if existing, err := s.repo.FindByStationAndUser(req.StationID, req.UserID); err == nil && existing != nil {
		return nil, ErrStaffExists
	}

	status := req.Status
	if status == 0 {
		status = model.StaffStatusEnabled
	}

	staff := &model.Staff{
		StationID:   req.StationID,
		UserID:      req.UserID,
		Role:        req.Role,
		Permissions: parseConfig(req.Permissions),
		Status:      status,
	}
	if err := s.repo.Create(staff); err != nil {
		return nil, err
	}
	return toStaffInfo(staff), nil
}

// Update 更新员工（角色/权限/状态）
func (s *staffService) Update(id uint, req *dto.UpdateStaffRequest) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrStaffNotFound
		}
		return err
	}

	fields := make(map[string]interface{})
	if req.Role != nil {
		fields["role"] = *req.Role
	}
	if req.Permissions != nil {
		fields["permissions"] = parseConfig(req.Permissions)
	}
	if req.Status != nil {
		fields["status"] = *req.Status
	}

	if len(fields) == 0 {
		return nil
	}
	return s.repo.UpdateFields(id, fields)
}

// Delete 删除员工
func (s *staffService) Delete(id uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrStaffNotFound
		}
		return err
	}
	return s.repo.Delete(id)
}

// GetByID 获取员工详情
func (s *staffService) GetByID(id uint) (*dto.StaffInfo, error) {
	staff, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrStaffNotFound
		}
		return nil, err
	}
	return toStaffInfo(staff), nil
}

// List 员工列表
func (s *staffService) List(req *dto.StaffListRequest) (*utils.Pagination, []dto.StaffInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.StaffListOptions{
		StationID: req.StationID,
		UserID:    req.UserID,
		Role:      req.Role,
		Status:    req.Status,
	}
	list, total, err := s.repo.List(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.StaffInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toStaffInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// ListByStation 按分站查询员工
func (s *staffService) ListByStation(stationID uint) ([]dto.StaffInfo, error) {
	list, err := s.repo.ListByStation(stationID)
	if err != nil {
		return nil, err
	}
	infos := make([]dto.StaffInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toStaffInfo(&list[i]))
	}
	return infos, nil
}

// ListByUser 按用户查询其所属分站员工记录
func (s *staffService) ListByUser(userID uint) ([]dto.StaffInfo, error) {
	list, err := s.repo.ListByUser(userID)
	if err != nil {
		return nil, err
	}
	infos := make([]dto.StaffInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toStaffInfo(&list[i]))
	}
	return infos, nil
}
