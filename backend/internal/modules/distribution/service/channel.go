// Package service 分销合伙人中台业务逻辑层 - 推广渠道
// 职责：CRUD / 生成推广码 / 统计追踪
package service

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"wuchang-tongcheng/internal/modules/distribution/dto"
	"wuchang-tongcheng/internal/modules/distribution/model"
	"wuchang-tongcheng/internal/modules/distribution/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrChannelNotFound  = errors.New("渠道不存在")
	ErrChannelCodeExists = errors.New("渠道码已存在")
	ErrChannelNoPermission = errors.New("无权操作此渠道")
)

// ChannelService 渠道业务接口
type ChannelService interface {
	Create(partnerID uint, req *dto.ChannelCreateRequest) (*dto.ChannelInfo, error)
	Update(id, operatorPartnerID uint, req *dto.ChannelUpdateRequest) error
	Delete(id, operatorPartnerID uint) error
	GetByID(id uint) (*dto.ChannelInfo, error)
	List(req *dto.ChannelListRequest) (*utils.Pagination, []dto.ChannelInfo, error)
	ListByPartner(partnerID uint) ([]dto.ChannelInfo, error)
	Stats(partnerID uint) (*dto.ChannelStatsResponse, error)

	// 追踪（公开）
	Track(req *dto.ChannelTrackRequest) error
}

type channelService struct {
	repo repository.ChannelRepository
}

// NewChannelService 创建渠道 service 实例
func NewChannelService(repo repository.ChannelRepository) ChannelService {
	return &channelService{repo: repo}
}

// toChannelInfo model -> dto
func toChannelInfo(c *model.Channel) *dto.ChannelInfo {
	return &dto.ChannelInfo{
		ID:                c.ID,
		PartnerID:         c.PartnerID,
		Code:              c.Code,
		Name:              c.Name,
		ClickCount:        c.ClickCount,
		RegisterCount:     c.RegisterCount,
		OrderCount:        c.OrderCount,
		CommissionAmount:  c.CommissionAmount,
		CreatedAt:         c.CreatedAt,
		UpdatedAt:         c.UpdatedAt,
	}
}

// generateChannelCode 生成唯一渠道码
func generateChannelCode() string {
	return fmt.Sprintf("DC%s%06d", time.Now().Format("0102150405"), rand.Intn(1000000))
}

// Create 创建渠道
func (s *channelService) Create(partnerID uint, req *dto.ChannelCreateRequest) (*dto.ChannelInfo, error) {
	code := req.Code
	if code == "" {
		// 自动生成唯一码
		for i := 0; i < 5; i++ {
			code = generateChannelCode()
			if _, err := s.repo.FindByCode(code); errors.Is(err, gorm.ErrRecordNotFound) {
				break
			}
		}
	} else {
		// 校验唯一性
		if existing, err := s.repo.FindByCode(code); err == nil && existing != nil {
			return nil, ErrChannelCodeExists
		} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}

	c := &model.Channel{
		PartnerID: partnerID,
		Code:      code,
		Name:      req.Name,
	}
	if err := s.repo.Create(c); err != nil {
		return nil, err
	}
	return toChannelInfo(c), nil
}

// Update 更新渠道
func (s *channelService) Update(id, operatorPartnerID uint, req *dto.ChannelUpdateRequest) error {
	c, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrChannelNotFound
		}
		return err
	}
	if operatorPartnerID > 0 && c.PartnerID != operatorPartnerID {
		return ErrChannelNoPermission
	}
	fields := make(map[string]interface{})
	if req.Name != nil {
		fields["name"] = *req.Name
	}
	if len(fields) == 0 {
		return nil
	}
	return s.repo.UpdateFields(id, fields)
}

// Delete 删除渠道
func (s *channelService) Delete(id, operatorPartnerID uint) error {
	c, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrChannelNotFound
		}
		return err
	}
	if operatorPartnerID > 0 && c.PartnerID != operatorPartnerID {
		return ErrChannelNoPermission
	}
	return s.repo.Delete(id)
}

// GetByID 详情
func (s *channelService) GetByID(id uint) (*dto.ChannelInfo, error) {
	c, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrChannelNotFound
		}
		return nil, err
	}
	return toChannelInfo(c), nil
}

// List 列表
func (s *channelService) List(req *dto.ChannelListRequest) (*utils.Pagination, []dto.ChannelInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.ChannelListOptions{
		PartnerID: req.PartnerID,
		Code:      req.Code,
		Keyword:   req.Keyword,
	}
	list, total, err := s.repo.List(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.ChannelInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toChannelInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// ListByPartner 按合伙人查询
func (s *channelService) ListByPartner(partnerID uint) ([]dto.ChannelInfo, error) {
	list, err := s.repo.ListByPartner(partnerID)
	if err != nil {
		return nil, err
	}
	infos := make([]dto.ChannelInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toChannelInfo(&list[i]))
	}
	return infos, nil
}

// Stats 渠道统计
func (s *channelService) Stats(partnerID uint) (*dto.ChannelStatsResponse, error) {
	totalChannels, totalClicks, totalRegisters, totalOrders, totalCommission, err := s.repo.StatsByPartner(partnerID)
	if err != nil {
		return nil, err
	}
	return &dto.ChannelStatsResponse{
		TotalChannels:   totalChannels,
		TotalClicks:     totalClicks,
		TotalRegisters:  totalRegisters,
		TotalOrders:     totalOrders,
		TotalCommission: totalCommission,
	}, nil
}

// Track 渠道追踪（公开调用，记录点击/注册/下单）
func (s *channelService) Track(req *dto.ChannelTrackRequest) error {
	c, err := s.repo.FindByCode(req.Code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrChannelNotFound
		}
		return err
	}
	switch req.Action {
	case "click":
		return s.repo.IncrClick(c.ID)
	case "register":
		return s.repo.IncrRegister(c.ID)
	case "order":
		return s.repo.IncrOrder(c.ID)
	}
	return errors.New("不支持的动作类型")
}
