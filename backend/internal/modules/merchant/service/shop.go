// Package service 商户中台业务逻辑层 - 店铺
// 依据架构设计 4.4：商家入驻/认领/店铺管理/信用分调整/等级管理
package service

import (
	"errors"
	"time"

	"wuchang-tongcheng/internal/modules/merchant/dto"
	"wuchang-tongcheng/internal/modules/merchant/model"
	"wuchang-tongcheng/internal/modules/merchant/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrShopNotFound      = errors.New("店铺不存在")
	ErrShopNoPermission  = errors.New("无权操作此店铺")
	ErrShopStatusInvalid = errors.New("店铺状态不允许此操作")
	ErrShopAlreadyOwned  = errors.New("店铺已被认领")
)

// ShopService 店铺业务接口
type ShopService interface {
	// C 端
	Apply(regionID uint, ownerID uint, req *dto.CreateShopRequest) (*dto.ShopInfo, error)
	Update(id uint, operatorID uint, req *dto.UpdateShopRequest) error
	GetByID(id uint) (*dto.ShopInfo, error)
	List(regionID uint, req *dto.ShopListRequest) (*utils.Pagination, []dto.ShopInfo, error)
	Search(regionID uint, req *dto.ShopListRequest) (*utils.Pagination, []dto.ShopInfo, error)
	ListMine(ownerID uint, page, pageSize int) (*utils.Pagination, []dto.ShopInfo, error)
	Claim(shopID, userID uint) (*dto.ShopInfo, error)

	// 信用分/等级
	UpdateCreditScore(id uint, delta int, reason string) error
	UpdateLevel(id uint, level int) error

	// M 端管理
	AdminList(req *dto.ShopAdminListRequest) (*utils.Pagination, []dto.ShopInfo, error)
	AdminGetByID(id uint) (*dto.ShopInfo, error)
	UpdateStatus(id uint, status int) error
}

type shopService struct {
	repo repository.ShopRepository
}

// NewShopService 创建店铺 service 实例
func NewShopService(repo repository.ShopRepository) ShopService {
	return &shopService{repo: repo}
}

// shopStatusText 店铺状态文本
func shopStatusText(status int) string {
	switch status {
	case model.ShopStatusPending:
		return "审核中"
	case model.ShopStatusActive:
		return "正常"
	case model.ShopStatusStopped:
		return "停用"
	}
	return ""
}

// Apply 商户入驻
func (s *shopService) Apply(regionID uint, ownerID uint, req *dto.CreateShopRequest) (*dto.ShopInfo, error) {
	shop := &model.Shop{
		OwnerID:     ownerID,
		Name:        req.Name,
		Logo:        req.Logo,
		Intro:       req.Intro,
		CategoryID:  req.CategoryID,
		Status:      model.ShopStatusPending,
		CreditScore: 100,
		Level:       1,
	}
	shop.RegionID = regionID
	if err := s.repo.Create(shop); err != nil {
		return nil, err
	}
	return s.toInfo(shop), nil
}

// Update 更新店铺
func (s *shopService) Update(id uint, operatorID uint, req *dto.UpdateShopRequest) error {
	shop, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrShopNotFound
		}
		return err
	}
	if shop.OwnerID != operatorID {
		return ErrShopNoPermission
	}
	fields := make(map[string]interface{})
	if req.Name != nil {
		fields["name"] = *req.Name
	}
	if req.Logo != nil {
		fields["logo"] = *req.Logo
	}
	if req.Intro != nil {
		fields["intro"] = *req.Intro
	}
	if req.CategoryID != nil {
		fields["category_id"] = *req.CategoryID
	}
	if len(fields) == 0 {
		return nil
	}
	return s.repo.UpdateFields(id, fields)
}

// GetByID 店铺详情
func (s *shopService) GetByID(id uint) (*dto.ShopInfo, error) {
	shop, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrShopNotFound
		}
		return nil, err
	}
	return s.toInfo(shop), nil
}

// List 店铺列表
func (s *shopService) List(regionID uint, req *dto.ShopListRequest) (*utils.Pagination, []dto.ShopInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.ShopListOptions{
		Keyword:    req.Keyword,
		CategoryID:  req.CategoryID,
		OwnerID:     req.OwnerID,
		Status:      req.Status,
	}
	list, total, err := s.repo.List(regionID, pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	infos := make([]dto.ShopInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *s.toInfo(&list[i]))
	}
	return pagination, infos, nil
}

// Search 搜索店铺
func (s *shopService) Search(regionID uint, req *dto.ShopListRequest) (*utils.Pagination, []dto.ShopInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	list, total, err := s.repo.Search(regionID, pagination, req.Keyword)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	infos := make([]dto.ShopInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *s.toInfo(&list[i]))
	}
	return pagination, infos, nil
}

// ListMine 我的店铺
func (s *shopService) ListMine(ownerID uint, page, pageSize int) (*utils.Pagination, []dto.ShopInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, err := s.repo.FindByOwnerID(ownerID)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = int64(len(list))
	infos := make([]dto.ShopInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *s.toInfo(&list[i]))
	}
	return pagination, infos, nil
}

// Claim 认领店铺（适用于"无主"店铺的认领场景，目前实现为：将已有店铺 owner_id 改为 userID）
// 注意：若店铺已有 owner，则返回 ErrShopAlreadyOwned
func (s *shopService) Claim(shopID, userID uint) (*dto.ShopInfo, error) {
	shop, err := s.repo.FindByID(shopID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrShopNotFound
		}
		return nil, err
	}
	if shop.OwnerID != 0 {
		return nil, ErrShopAlreadyOwned
	}
	if err := s.repo.UpdateFields(shopID, map[string]interface{}{
		"owner_id":   userID,
		"status":     model.ShopStatusActive,
		"settled_at": time.Now(),
	}); err != nil {
		return nil, err
	}
	now := time.Now()
	shop.OwnerID = userID
	shop.Status = model.ShopStatusActive
	shop.SettledAt = &now
	return s.toInfo(shop), nil
}

// UpdateCreditScore 调整信用分
func (s *shopService) UpdateCreditScore(id uint, delta int, reason string) error {
	_, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrShopNotFound
		}
		return err
	}
	return s.repo.UpdateCreditScore(id, delta)
}

// UpdateLevel 调整等级
func (s *shopService) UpdateLevel(id uint, level int) error {
	_, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrShopNotFound
		}
		return err
	}
	return s.repo.UpdateLevel(id, level)
}

// AdminList 管理后台列表
func (s *shopService) AdminList(req *dto.ShopAdminListRequest) (*utils.Pagination, []dto.ShopInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.ShopAdminListOptions{
		RegionID:   req.RegionID,
		OwnerID:    req.OwnerID,
		CategoryID: req.CategoryID,
		Status:     req.Status,
		Keyword:    req.Keyword,
	}
	list, total, err := s.repo.AdminList(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	infos := make([]dto.ShopInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *s.toInfo(&list[i]))
	}
	return pagination, infos, nil
}

// AdminGetByID 管理后台详情
func (s *shopService) AdminGetByID(id uint) (*dto.ShopInfo, error) {
	return s.GetByID(id)
}

// UpdateStatus 更新状态
func (s *shopService) UpdateStatus(id uint, status int) error {
	if status != model.ShopStatusPending && status != model.ShopStatusActive && status != model.ShopStatusStopped {
		return ErrShopStatusInvalid
	}
	_, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrShopNotFound
		}
		return err
	}
	return s.repo.UpdateFields(id, map[string]interface{}{"status": status})
}

// toInfo 模型转 DTO
func (s *shopService) toInfo(m *model.Shop) *dto.ShopInfo {
	return &dto.ShopInfo{
		ID:          m.ID,
		RegionID:    m.RegionID,
		OwnerID:     m.OwnerID,
		Name:        m.Name,
		Logo:        m.Logo,
		Intro:       m.Intro,
		CategoryID:  m.CategoryID,
		Status:      m.Status,
		StatusText:  shopStatusText(m.Status),
		CreditScore: m.CreditScore,
		Level:       m.Level,
		SettledAt:   m.SettledAt,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}
