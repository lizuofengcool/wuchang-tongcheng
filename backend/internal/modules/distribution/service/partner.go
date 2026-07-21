// Package service 分销合伙人中台业务逻辑层 - 合伙人
// 职责：申请加入 / 升级 / 上下级关系 / 佣金率计算
package service

import (
	"errors"
	"time"

	"wuchang-tongcheng/internal/modules/distribution/dto"
	"wuchang-tongcheng/internal/modules/distribution/model"
	"wuchang-tongcheng/internal/modules/distribution/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrPartnerNotFound      = errors.New("合伙人不存在")
	ErrPartnerExists        = errors.New("该用户已是合伙人")
	ErrPartnerStatusInvalid = errors.New("合伙人状态不允许此操作")
	ErrPartnerLevelInvalid  = errors.New("合伙人等级无效")
	ErrPartnerParentInvalid = errors.New("上级合伙人无效")
	ErrPartnerRateInvalid   = errors.New("佣金比例无效")
)

// PartnerService 合伙人业务接口
type PartnerService interface {
	Apply(regionID, userID uint, req *dto.PartnerApplyRequest) (*dto.PartnerInfo, error)
	GetByID(id uint) (*dto.PartnerInfo, error)
	GetByUserID(userID uint) (*dto.PartnerInfo, error)
	List(req *dto.PartnerListRequest) (*utils.Pagination, []dto.PartnerInfo, error)
	Tree(req *dto.PartnerTreeRequest) ([]dto.PartnerInfo, error)

	// 管理端
	Update(id uint, req *dto.PartnerUpdateRequest) error
	UpdateStatus(id uint, status int) error
	Upgrade(id uint, targetLevel int) error
	AdjustCommissionRate(id uint, rate float64) error

	// 佣金率计算（按等级自动获取）
	GetCommissionRate(partnerID uint) (float64, error)
}

type partnerService struct {
	repo     repository.PartnerRepository
	levelSvc LevelService
}

// NewPartnerService 创建合伙人 service 实例
func NewPartnerService(repo repository.PartnerRepository, levelSvc LevelService) PartnerService {
	return &partnerService{repo: repo, levelSvc: levelSvc}
}

// partnerLevelText 等级文本
func partnerLevelText(level int) string {
	switch level {
	case model.PartnerLevelNormal:
		return "普通合伙人"
	case model.PartnerLevelSenior:
		return "高级合伙人"
	case model.PartnerLevelCity:
		return "城市合伙人"
	}
	return ""
}

// partnerStatusText 状态文本
func partnerStatusText(status int) string {
	switch status {
	case model.PartnerStatusPending:
		return "待审核"
	case model.PartnerStatusActive:
		return "正常"
	case model.PartnerStatusFrozen:
		return "冻结"
	case model.PartnerStatusRejected:
		return "拒绝"
	case model.PartnerStatusQuit:
		return "退出"
	}
	return ""
}

// toPartnerInfo model -> dto
func toPartnerInfo(p *model.Partner) *dto.PartnerInfo {
	return &dto.PartnerInfo{
		ID:                p.ID,
		UserID:            p.UserID,
		ParentID:          p.ParentID,
		Level:             p.Level,
		LevelText:         partnerLevelText(p.Level),
		CommissionRate:    p.CommissionRate,
		TotalCommission:   p.TotalCommission,
		SettledCommission: p.SettledCommission,
		PendingCommission: p.PendingCommission,
		Status:            p.Status,
		StatusText:        partnerStatusText(p.Status),
		JoinedAt:          p.JoinedAt,
		RegionID:          p.RegionID,
		CreatedAt:         p.CreatedAt,
		UpdatedAt:         p.UpdatedAt,
	}
}

// Apply 申请加入合伙人
func (s *partnerService) Apply(regionID, userID uint, req *dto.PartnerApplyRequest) (*dto.PartnerInfo, error) {
	// 检查是否已是合伙人
	if existing, err := s.repo.FindByUserID(userID); err == nil && existing != nil {
		return nil, ErrPartnerExists
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 校验上级
	if req.ParentID > 0 {
		parent, err := s.repo.FindByID(req.ParentID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrPartnerParentInvalid
			}
			return nil, err
		}
		if parent.Status != model.PartnerStatusActive {
			return nil, ErrPartnerParentInvalid
		}
	}

	// 等级与佣金率
	level := req.Level
	if level == 0 {
		level = model.PartnerLevelNormal
	}
	if level < model.PartnerLevelNormal || level > model.PartnerLevelCity {
		return nil, ErrPartnerLevelInvalid
	}

	rate := req.CommissionRate
	if rate == 0 {
		// 按等级默认佣金率
		if lvl, err := s.levelSvc.GetByLevel(level); err == nil && lvl != nil {
			rate = lvl.CommissionRate
		}
	}
	if rate < 0 || rate > 1 {
		return nil, ErrPartnerRateInvalid
	}

	now := time.Now()
	p := &model.Partner{
		UserID:         userID,
		ParentID:       req.ParentID,
		Level:          level,
		CommissionRate: rate,
		Status:         model.PartnerStatusPending, // 申请后待审核
	}
	p.RegionID = regionID
	_ = now // joined_at 在审核通过时设置

	if err := s.repo.Create(p); err != nil {
		return nil, err
	}
	return toPartnerInfo(p), nil
}

// GetByID 详情
func (s *partnerService) GetByID(id uint) (*dto.PartnerInfo, error) {
	p, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPartnerNotFound
		}
		return nil, err
	}
	return toPartnerInfo(p), nil
}

// GetByUserID 按用户查询
func (s *partnerService) GetByUserID(userID uint) (*dto.PartnerInfo, error) {
	p, err := s.repo.FindByUserID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPartnerNotFound
		}
		return nil, err
	}
	return toPartnerInfo(p), nil
}

// List 列表
func (s *partnerService) List(req *dto.PartnerListRequest) (*utils.Pagination, []dto.PartnerInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.PartnerListOptions{
		UserID:   req.UserID,
		ParentID: req.ParentID,
		Level:    req.Level,
		Status:   req.Status,
		RegionID: req.RegionID,
		Keyword:  req.Keyword,
	}
	list, total, err := s.repo.List(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.PartnerInfo, 0, len(list))
	for i := range list {
		info := *toPartnerInfo(&list[i])
		if cnt, err := s.repo.CountByParent(list[i].ID); err == nil {
			info.ChildCount = int(cnt)
		}
		infos = append(infos, info)
	}
	pagination.Total = total
	return pagination, infos, nil
}

// Tree 上下级树
func (s *partnerService) Tree(req *dto.PartnerTreeRequest) ([]dto.PartnerInfo, error) {
	parentID := req.ParentID
	depth := req.Depth
	if depth <= 0 {
		depth = 2
	}
	return s.buildTree(parentID, depth)
}

// buildTree 递归构建上下级树
func (s *partnerService) buildTree(parentID uint, depth int) ([]dto.PartnerInfo, error) {
	children, err := s.repo.ListByParent(parentID)
	if err != nil {
		return nil, err
	}
	infos := make([]dto.PartnerInfo, 0, len(children))
	for i := range children {
		info := *toPartnerInfo(&children[i])
		if cnt, err := s.repo.CountByParent(children[i].ID); err == nil {
			info.ChildCount = int(cnt)
		}
		if depth > 1 {
			sub, _ := s.buildTree(children[i].ID, depth-1)
			info.Children = sub
		}
		infos = append(infos, info)
	}
	return infos, nil
}

// Update 管理端更新
func (s *partnerService) Update(id uint, req *dto.PartnerUpdateRequest) error {
	p, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPartnerNotFound
		}
		return err
	}
	fields := make(map[string]interface{})
	if req.Level != nil {
		if *req.Level < model.PartnerLevelNormal || *req.Level > model.PartnerLevelCity {
			return ErrPartnerLevelInvalid
		}
		fields["level"] = *req.Level
	}
	if req.CommissionRate != nil {
		if *req.CommissionRate < 0 || *req.CommissionRate > 1 {
			return ErrPartnerRateInvalid
		}
		fields["commission_rate"] = *req.CommissionRate
	}
	if req.ParentID != nil {
		if *req.ParentID > 0 {
			parent, err := s.repo.FindByID(*req.ParentID)
			if err != nil || parent.ID == id {
				return ErrPartnerParentInvalid
			}
		}
		fields["parent_id"] = *req.ParentID
	}
	if req.Status != nil {
		fields["status"] = *req.Status
		// 审核通过时记录加入时间
		if *req.Status == model.PartnerStatusActive && p.JoinedAt == nil {
			now := time.Now()
			fields["joined_at"] = &now
		}
	}
	if len(fields) == 0 {
		return nil
	}
	return s.repo.UpdateFields(id, fields)
}

// UpdateStatus 更新状态
func (s *partnerService) UpdateStatus(id uint, status int) error {
	p, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPartnerNotFound
		}
		return err
	}
	fields := map[string]interface{}{"status": status}
	if status == model.PartnerStatusActive && p.JoinedAt == nil {
		now := time.Now()
		fields["joined_at"] = &now
	}
	return s.repo.UpdateFields(id, fields)
}

// Upgrade 升级合伙人等级
func (s *partnerService) Upgrade(id uint, targetLevel int) error {
	if targetLevel < model.PartnerLevelNormal || targetLevel > model.PartnerLevelCity {
		return ErrPartnerLevelInvalid
	}
	p, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPartnerNotFound
		}
		return err
	}
	if targetLevel <= p.Level {
		return errors.New("目标等级必须高于当前等级")
	}
	fields := map[string]interface{}{"level": targetLevel}
	// 同步更新佣金率为目标等级的默认佣金率
	if lvl, err := s.levelSvc.GetByLevel(targetLevel); err == nil && lvl != nil && lvl.CommissionRate > 0 {
		fields["commission_rate"] = lvl.CommissionRate
	}
	return s.repo.UpdateFields(id, fields)
}

// AdjustCommissionRate 调整佣金率
func (s *partnerService) AdjustCommissionRate(id uint, rate float64) error {
	if rate < 0 || rate > 1 {
		return ErrPartnerRateInvalid
	}
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPartnerNotFound
		}
		return err
	}
	return s.repo.UpdateFields(id, map[string]interface{}{"commission_rate": rate})
}

// GetCommissionRate 获取合伙人佣金率（若为 0 则按等级默认）
func (s *partnerService) GetCommissionRate(partnerID uint) (float64, error) {
	p, err := s.repo.FindByID(partnerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, ErrPartnerNotFound
		}
		return 0, err
	}
	if p.CommissionRate > 0 {
		return p.CommissionRate, nil
	}
	if lvl, err := s.levelSvc.GetByLevel(p.Level); err == nil && lvl != nil {
		return lvl.CommissionRate, nil
	}
	return 0, nil
}
