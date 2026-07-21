// Package service 同城商城业务逻辑层 - SKU 规格表
package service

import (
	"errors"

	"wuchang-tongcheng/internal/modules/mall/dto"
	"wuchang-tongcheng/internal/modules/mall/model"
	"wuchang-tongcheng/internal/modules/mall/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// SkuService SKU 业务接口
type SkuService interface {
	Create(regionID, userID, productID uint, req *dto.CreateSkuRequest) (*dto.SkuInfo, error)
	Update(id, userID uint, req *dto.UpdateSkuRequest) error
	Delete(id, userID uint) error
	GetByID(id uint) (*dto.SkuInfo, error)
	ListByProduct(productID uint) ([]dto.SkuInfo, error)
	ListByShop(shopID uint, req *dto.SkuListRequest) (*utils.Pagination, []dto.SkuInfo, error)

	UpdateStock(id, userID uint, stock int) error
	BatchUpdateStock(userID uint, req *dto.BatchUpdateStockRequest) error
}

type skuService struct {
	repo        repository.SkuRepository
	productRepo repository.ProductRepository
}

// NewSkuService 创建 SKU service 实例
func NewSkuService(repo repository.SkuRepository, productRepo repository.ProductRepository) SkuService {
	return &skuService{repo: repo, productRepo: productRepo}
}

func toSkuInfo(s *model.Sku) *dto.SkuInfo {
	info := &dto.SkuInfo{
		ID:            s.ID,
		ProductID:     s.ProductID,
		ShopID:        s.ShopID,
		Name:          s.Name,
		SkuCode:       s.SkuCode,
		Barcode:       s.Barcode,
		Price:         s.Price,
		OriginalPrice: s.OriginalPrice,
		CostPrice:     s.CostPrice,
		Stock:         s.Stock,
		Sales:         s.Sales,
		WarnStock:     s.WarnStock,
		Image:         s.Image,
		Sort:          s.Sort,
		Status:        s.Status,
		RegionID:      s.RegionID,
		CreatedAt:     s.CreatedAt,
		UpdatedAt:     s.UpdatedAt,
	}
	if s.Specs != nil {
		info.Specs = s.Specs
	}
	return info
}

// Create 创建 SKU
func (s *skuService) Create(regionID, userID, productID uint, req *dto.CreateSkuRequest) (*dto.SkuInfo, error) {
	p, err := s.productRepo.FindByID(productID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}
	if p.UserID != userID {
		return nil, errors.New("无权操作他人商品")
	}

	sku := &model.Sku{
		ProductID:    productID,
		ShopID:       p.ShopID,
		UserID:       userID,
		Name:         req.Name,
		SkuCode:      req.SkuCode,
		Barcode:      req.Barcode,
		Price:        req.Price,
		OriginalPrice: req.OriginalPrice,
		CostPrice:    req.CostPrice,
		Stock:        req.Stock,
		WarnStock:    req.WarnStock,
		Image:        req.Image,
		Sort:         req.Sort,
		Status:       1,
	}
	sku.RegionID = regionID
	if req.Specs != nil {
		if b, err := model.FromJSON(req.Specs); err == nil {
			sku.Specs = b
		}
	}

	if err := s.repo.Create(sku); err != nil {
		return nil, err
	}
	return toSkuInfo(sku), nil
}

// Update 更新 SKU
func (s *skuService) Update(id, userID uint, req *dto.UpdateSkuRequest) error {
	sku, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSkuNotFound
		}
		return err
	}
	if sku.UserID != userID {
		return errors.New("无权操作他人 SKU")
	}

	fields := make(map[string]interface{})
	if req.Name != nil {
		fields["name"] = *req.Name
	}
	if req.SkuCode != nil {
		fields["sku_code"] = *req.SkuCode
	}
	if req.Barcode != nil {
		fields["barcode"] = *req.Barcode
	}
	if req.Price != nil {
		fields["price"] = *req.Price
	}
	if req.OriginalPrice != nil {
		fields["original_price"] = *req.OriginalPrice
	}
	if req.CostPrice != nil {
		fields["cost_price"] = *req.CostPrice
	}
	if req.Stock != nil {
		fields["stock"] = *req.Stock
	}
	if req.WarnStock != nil {
		fields["warn_stock"] = *req.WarnStock
	}
	if req.Image != nil {
		fields["image"] = *req.Image
	}
	if req.Sort != nil {
		fields["sort"] = *req.Sort
	}
	if req.Status != nil {
		fields["status"] = *req.Status
	}
	if req.Specs != nil {
		if b, err := model.FromJSON(req.Specs); err == nil {
			fields["specs"] = b
		}
	}

	if len(fields) == 0 {
		return nil
	}
	return s.repo.UpdateFields(id, fields)
}

// Delete 删除 SKU
func (s *skuService) Delete(id, userID uint) error {
	sku, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSkuNotFound
		}
		return err
	}
	if sku.UserID != userID {
		return errors.New("无权操作他人 SKU")
	}
	return s.repo.Delete(id)
}

// GetByID 获取 SKU 详情
func (s *skuService) GetByID(id uint) (*dto.SkuInfo, error) {
	sku, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSkuNotFound
		}
		return nil, err
	}
	return toSkuInfo(sku), nil
}

// ListByProduct 按商品列出 SKU
func (s *skuService) ListByProduct(productID uint) ([]dto.SkuInfo, error) {
	list, err := s.repo.ListByProduct(productID)
	if err != nil {
		return nil, err
	}
	infos := make([]dto.SkuInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toSkuInfo(&list[i]))
	}
	return infos, nil
}

// ListByShop 按店铺列出 SKU
func (s *skuService) ListByShop(shopID uint, req *dto.SkuListRequest) (*utils.Pagination, []dto.SkuInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	list, total, err := s.repo.ListByShop(shopID, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.SkuInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toSkuInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// UpdateStock 更新 SKU 库存
func (s *skuService) UpdateStock(id, userID uint, stock int) error {
	sku, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSkuNotFound
		}
		return err
	}
	if sku.UserID != userID {
		return errors.New("无权操作他人 SKU")
	}
	return s.repo.UpdateStock(id, stock)
}

// BatchUpdateStock 批量更新库存
func (s *skuService) BatchUpdateStock(userID uint, req *dto.BatchUpdateStockRequest) error {
	items := make([]repository.SkuStockUpdateItem, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, repository.SkuStockUpdateItem{
			SkuID:    it.SkuID,
			Quantity: it.Quantity,
		})
	}
	return s.repo.BatchUpdateStock(items)
}
