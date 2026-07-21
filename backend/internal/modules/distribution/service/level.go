// Package service 分销合伙人中台业务逻辑层 - 合伙人等级
// 职责：CRUD / 自动升级判断
package service

import (
	"errors"

	"wuchang-tongcheng/internal/modules/distribution/dto"
	"wuchang-tongcheng/internal/modules/distribution/model"
	"wuchang-tongcheng/internal/modules/distribution/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrLevelNotFound      = errors.New("等级不存在")
	ErrLevelExists        = errors.New("该等级已存在")
	ErrLevelHasPartners   = errors.New("该等级下存在合伙人，无法删除")
	ErrLevelStatusInvalid = errors.New("等级状态无效")
)

// LevelService 等级业务接口
type LevelService interface {
	Create(req *dto.LevelCreateRequest) (*dto.LevelInfo, error)
	Update(id uint, req *dto.LevelUpdateRequest) error
	Delete(id uint) error
	GetByID(id uint) (*dto.LevelInfo, error)
	GetByLevel(level int) (*dto.LevelInfo, error)
	List(req *dto.LevelListRequest) (*utils.Pagination, []dto.LevelInfo, error)
	ListAll() ([]dto.LevelInfo, error)

	// 自动升级判断
	CheckUpgrade(req *dto.LevelCheckUpgradeRequest) (*dto.LevelCheckUpgradeResponse, error)
}

type levelService struct {
	repo        repository.LevelRepository
	partnerRepo repository.PartnerRepository
}

// NewLevelService 创建等级 service 实例
func NewLevelService(repo repository.LevelRepository, partnerRepo repository.PartnerRepository) LevelService {
	return &levelService{repo: repo, partnerRepo: partnerRepo}
}

// levelStatusText 状态文本
func levelStatusText(status int) string {
	switch status {
	case model.LevelStatusEnabled:
		return "启用"
	case model.LevelStatusDisabled:
		return "禁用"
	}
	return ""
}

// toLevelInfo model -> dto
func toLevelInfo(l *model.Level) *dto.LevelInfo {
	info := &dto.LevelInfo{
		ID:             l.ID,
		Level:          l.Level,
		Name:           l.Name,
		RequiredAmount: l.RequiredAmount,
		CommissionRate: l.CommissionRate,
		Status:         l.Status,
		StatusText:     levelStatusText(l.Status),
		CreatedAt:      l.CreatedAt,
		UpdatedAt:      l.UpdatedAt,
	}
	if l.ExtraBenefits != nil {
		info.ExtraBenefits = l.ExtraBenefits
	}
	return info
}

// Create 创建等级
func (s *levelService) Create(req *dto.LevelCreateRequest) (*dto.LevelInfo, error) {
	// 校验等级唯一性
	if existing, err := s.repo.FindByLevel(req.Level); err == nil && existing != nil {
		return nil, ErrLevelExists
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	l := &model.Level{
		Level:          req.Level,
		Name:           req.Name,
		RequiredAmount: req.RequiredAmount,
		CommissionRate: req.CommissionRate,
		Status:         req.Status,
	}
	if l.Status == 0 {
		l.Status = model.LevelStatusEnabled
	}
	if req.ExtraBenefits != nil {
		if b, err := model.FromJSON(req.ExtraBenefits); err == nil {
			l.ExtraBenefits = b
		}
	}
	if err := s.repo.Create(l); err != nil {
		return nil, err
	}
	return toLevelInfo(l), nil
}

// Update 更新等级
func (s *levelService) Update(id uint, req *dto.LevelUpdateRequest) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrLevelNotFound
		}
		return err
	}
	fields := make(map[string]interface{})
	if req.Name != nil {
		fields["name"] = *req.Name
	}
	if req.RequiredAmount != nil {
		fields["required_amount"] = *req.RequiredAmount
	}
	if req.CommissionRate != nil {
		fields["commission_rate"] = *req.CommissionRate
	}
	if req.Status != nil {
		fields["status"] = *req.Status
	}
	if req.ExtraBenefits != nil {
		if b, err := model.FromJSON(req.ExtraBenefits); err == nil {
			fields["extra_benefits"] = b
		}
	}
	if len(fields) == 0 {
		return nil
	}
	return s.repo.UpdateFields(id, fields)
}

// Delete 删除等级
func (s *levelService) Delete(id uint) error {
	l, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrLevelNotFound
		}
		return err
	}
	// 检查是否有合伙人使用此等级
	pagination := utils.NewPagination(1, 1)
	if list, _, _ := s.partnerRepo.ListByLevel(l.Level, pagination); len(list) > 0 {
		return ErrLevelHasPartners
	}
	return s.repo.Delete(id)
}

// GetByID 详情
func (s *levelService) GetByID(id uint) (*dto.LevelInfo, error) {
	l, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrLevelNotFound
		}
		return nil, err
	}
	return toLevelInfo(l), nil
}

// GetByLevel 按等级值查询
func (s *levelService) GetByLevel(level int) (*dto.LevelInfo, error) {
	l, err := s.repo.FindByLevel(level)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrLevelNotFound
		}
		return nil, err
	}
	return toLevelInfo(l), nil
}

// List 列表
func (s *levelService) List(req *dto.LevelListRequest) (*utils.Pagination, []dto.LevelInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.LevelListOptions{
		Level:  req.Level,
		Status: req.Status,
	}
	list, total, err := s.repo.List(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.LevelInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toLevelInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// ListAll 全部启用等级
func (s *levelService) ListAll() ([]dto.LevelInfo, error) {
	list, err := s.repo.ListAll()
	if err != nil {
		return nil, err
	}
	infos := make([]dto.LevelInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toLevelInfo(&list[i]))
	}
	return infos, nil
}

// CheckUpgrade 检查自动升级
func (s *levelService) CheckUpgrade(req *dto.LevelCheckUpgradeRequest) (*dto.LevelCheckUpgradeResponse, error) {
	p, err := s.partnerRepo.FindByID(req.PartnerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPartnerNotFound
		}
		return nil, err
	}

	resp := &dto.LevelCheckUpgradeResponse{
		PartnerID:       p.ID,
		CurrentLevel:    p.Level,
		TargetLevel:     p.Level,
		TotalCommission: p.TotalCommission,
		Upgraded:        false,
	}

	// 查询所有启用的更高等级，按等级升序
	allLevels, err := s.repo.ListAll()
	if err != nil {
		return nil, err
	}
	targetLevel := p.Level
	for _, l := range allLevels {
		if l.Level > p.Level && p.TotalCommission >= l.RequiredAmount && l.RequiredAmount > 0 {
			if l.Level > targetLevel {
				targetLevel = l.Level
			}
		}
	}
	resp.TargetLevel = targetLevel

	// 若建议等级高于当前等级，执行升级
	if targetLevel > p.Level {
		if err := s.partnerRepo.UpdateFields(p.ID, map[string]interface{}{
			"level": targetLevel,
		}); err != nil {
			return nil, err
		}
		// 同步佣金率为新等级默认佣金率
		if lvl, err := s.repo.FindByLevel(targetLevel); err == nil && lvl != nil && lvl.CommissionRate > 0 {
			_ = s.partnerRepo.UpdateFields(p.ID, map[string]interface{}{
				"commission_rate": lvl.CommissionRate,
			})
		}
		resp.Upgraded = true
	}
	return resp, nil
}
