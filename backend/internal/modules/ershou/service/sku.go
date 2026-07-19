// Package service SKU 规格业务逻辑层
// 依据 v3.2.1 架构方案：对标闲鱼/转转
package service

import (
	"errors"

	"wuchang-tongcheng/internal/modules/ershou/dto"
	"wuchang-tongcheng/internal/modules/ershou/model"
	"wuchang-tongcheng/internal/modules/ershou/repository"

	"gorm.io/gorm"
)

var (
	ErrSKUNotFound       = errors.New("SKU 不存在")
	ErrSKUCodeDuplicate  = errors.New("SKU 编码已存在")
	ErrSKUStockInsufficient = errors.New("SKU 库存不足")
)

// SKUService SKU 业务接口
type SKUService interface {
	List(ershouID uint) (*dto.SKUListResponse, error)
	Create(ershouID uint, operatorID uint, req *dto.SKURequest) (*dto.SKUResponse, error)
	Update(ershouID, skuID, operatorID uint, req *dto.SKURequest) (*dto.SKUResponse, error)
	Delete(ershouID, skuID, operatorID uint) error
}

type skuService struct {
	repo      repository.SKURepository
	ershouRepo repository.ErshouRepository
}

// NewSKUService 创建 SKU service 实例
func NewSKUService(repo repository.SKURepository, ershouRepo repository.ErshouRepository) SKUService {
	return &skuService{repo: repo, ershouRepo: ershouRepo}
}

func (s *skuService) List(ershouID uint) (*dto.SKUListResponse, error) {
	list, err := s.repo.ListByErshouID(ershouID)
	if err != nil {
		return nil, err
	}
	resp := &dto.SKUListResponse{List: []dto.SKUResponse{}, Total: int64(len(list))}
	for _, sku := range list {
		resp.List = append(resp.List, *toSKUResponse(&sku))
	}
	return resp, nil
}

func (s *skuService) Create(ershouID uint, operatorID uint, req *dto.SKURequest) (*dto.SKUResponse, error) {
	e, err := s.ershouRepo.FindByID(ershouID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrErshouNotFound
		}
		return nil, err
	}
	if e.UserID != operatorID {
		return nil, ErrErshouNoPermission
	}
	// 编码重复检查
	if existing, err := s.repo.FindByCode(ershouID, req.SKUCode); err == nil && existing != nil {
		return nil, ErrSKUCodeDuplicate
	}

	attrs, _ := model.FromJSON(req.Attributes)
	sku := &model.ErshouSKU{
		ErshouID:   ershouID,
		SKUCode:    req.SKUCode,
		Name:       req.Name,
		Color:      req.Color,
		Size:       req.Size,
		Version:    req.Version,
		Price:      req.Price,
		Stock:      req.Stock,
		Image:      req.Image,
		Weight:     req.Weight,
		Barcode:    req.Barcode,
		Status:     req.Status,
		Attributes: attrs,
		Sort:       req.Sort,
	}
	if sku.Status == 0 {
		sku.Status = 1
	}
	if err := s.repo.Create(sku); err != nil {
		return nil, err
	}
	// 主表标记启用 SKU
	_ = s.ershouRepo.UpdateFields(ershouID, map[string]interface{}{"sku_enabled": true})
	return toSKUResponse(sku), nil
}

func (s *skuService) Update(ershouID, skuID, operatorID uint, req *dto.SKURequest) (*dto.SKUResponse, error) {
	e, err := s.ershouRepo.FindByID(ershouID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrErshouNotFound
		}
		return nil, err
	}
	if e.UserID != operatorID {
		return nil, ErrErshouNoPermission
	}
	sku, err := s.repo.FindByID(skuID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSKUNotFound
		}
		return nil, err
	}
	if sku.ErshouID != ershouID {
		return nil, ErrErshouNoPermission
	}

	fields := map[string]interface{}{
		"sku_code": req.SKUCode,
		"name":     req.Name,
		"color":    req.Color,
		"size":     req.Size,
		"version":  req.Version,
		"price":    req.Price,
		"stock":    req.Stock,
		"image":    req.Image,
		"weight":   req.Weight,
		"barcode":  req.Barcode,
		"status":   req.Status,
		"sort":     req.Sort,
	}
	if req.Attributes != nil {
		attrs, _ := model.FromJSON(req.Attributes)
		fields["attributes"] = attrs
	}
	if err := s.repo.Update(skuID, fields); err != nil {
		return nil, err
	}
	updated, _ := s.repo.FindByID(skuID)
	return toSKUResponse(updated), nil
}

func (s *skuService) Delete(ershouID, skuID, operatorID uint) error {
	e, err := s.ershouRepo.FindByID(ershouID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrErshouNotFound
		}
		return err
	}
	if e.UserID != operatorID {
		return ErrErshouNoPermission
	}
	sku, err := s.repo.FindByID(skuID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSKUNotFound
		}
		return err
	}
	if sku.ErshouID != ershouID {
		return ErrErshouNoPermission
	}
	return s.repo.Delete(skuID)
}

func toSKUResponse(sku *model.ErshouSKU) *dto.SKUResponse {
	resp := &dto.SKUResponse{
		ID:        sku.ID,
		ErshouID:  sku.ErshouID,
		SKUCode:   sku.SKUCode,
		Name:      sku.Name,
		Color:     sku.Color,
		Size:      sku.Size,
		Version:   sku.Version,
		Price:     sku.Price,
		Stock:     sku.Stock,
		SoldCount: sku.SoldCount,
		Image:     sku.Image,
		Weight:    sku.Weight,
		Barcode:   sku.Barcode,
		Status:    sku.Status,
		Sort:      sku.Sort,
		CreatedAt: sku.CreatedAt,
		UpdatedAt: sku.UpdatedAt,
	}
	if sku.Attributes != nil {
		var m map[string]interface{}
		_ = sku.Attributes.Parse(&m)
		resp.Attributes = m
	}
	return resp
}
