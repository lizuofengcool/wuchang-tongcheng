// Package service 同城商城业务逻辑层 - 商品 SPU
package service

import (
	"errors"
	"time"

	"wuchang-tongcheng/internal/modules/mall/dto"
	"wuchang-tongcheng/internal/modules/mall/model"
	"wuchang-tongcheng/internal/modules/mall/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrProductNotFound = errors.New("商品不存在")
	ErrSkuNotFound     = errors.New("SKU 不存在")
)

// ProductService 商品业务接口
type ProductService interface {
	Create(regionID, userID, shopID uint, req *dto.CreateProductRequest) (*dto.ProductInfo, error)
	Update(id, userID uint, req *dto.UpdateProductRequest) error
	Delete(id, userID uint) error
	GetByID(id uint) (*dto.ProductInfo, error)
	List(regionID uint, req *dto.ProductListRequest) (*utils.Pagination, []dto.ProductInfo, error)
	AdminList(req *dto.ProductAdminListRequest) (*utils.Pagination, []dto.ProductInfo, error)
	Search(regionID uint, keyword string, page, pageSize int) (*utils.Pagination, []dto.ProductInfo, error)
	ListByShop(shopID uint, page, pageSize int) (*utils.Pagination, []dto.ProductInfo, error)
	ListByCategory(regionID, categoryID uint, page, pageSize int) (*utils.Pagination, []dto.ProductInfo, error)
	ListByUser(userID uint, page, pageSize int) (*utils.Pagination, []dto.ProductInfo, error)
	ListFeatured(regionID uint, limit int) ([]dto.ProductInfo, error)
	ListHot(regionID uint, limit int) ([]dto.ProductInfo, error)
	ListNew(regionID uint, limit int) ([]dto.ProductInfo, error)

	UpdateStatus(id, userID uint, status int) error
	Audit(id uint, auditStatus int, reason string) error
	UpdatePromotion(id uint, req *dto.PromotionRequest) error
	IncrViewCount(id uint) error
}

type productService struct {
	repo     repository.ProductRepository
	shopRepo repository.ShopRepository
}

// NewProductService 创建商品 service 实例
func NewProductService(repo repository.ProductRepository, shopRepo repository.ShopRepository) ProductService {
	return &productService{repo: repo, shopRepo: shopRepo}
}

func productStatusText(s int) string {
	switch s {
	case model.ProductStatusDraft:
		return "草稿"
	case model.ProductStatusOnSale:
		return "在售"
	case model.ProductStatusOffShelf:
		return "已下架"
	case model.ProductStatusSoldOut:
		return "售罄"
	case model.ProductStatusRecycled:
		return "回收站"
	}
	return ""
}

func productAuditStatusText(s int) string {
	switch s {
	case model.ProductAuditPending:
		return "待审核"
	case model.ProductAuditApproved:
		return "已通过"
	case model.ProductAuditRejected:
		return "已拒绝"
	}
	return ""
}

func toProductInfo(p *model.Product) *dto.ProductInfo {
	info := &dto.ProductInfo{
		ID:             p.ID,
		ShopID:         p.ShopID,
		UserID:         p.UserID,
		CategoryID:     p.CategoryID,
		BrandID:        p.BrandID,
		Name:           p.Name,
		Subtitle:       p.Subtitle,
		MainImage:      p.MainImage,
		Detail:         p.Detail,
		ProductType:    p.ProductType,
		Price:          p.Price,
		OriginalPrice:  p.OriginalPrice,
		MinPrice:       p.MinPrice,
		MaxPrice:       p.MaxPrice,
		Stock:          p.Stock,
		Sales:          p.Sales,
		VirtualSales:   p.VirtualSales,
		StockWarn:      p.StockWarn,
		Status:         p.Status,
		StatusText:     productStatusText(p.Status),
		AuditStatus:    p.AuditStatus,
		AuditStatusText: productAuditStatusText(p.AuditStatus),
		AuditReason:    p.AuditReason,
		PublishedAt:    p.PublishedAt,
		ViewCount:      p.ViewCount,
		FavoriteCount:  p.FavoriteCount,
		ReviewCount:    p.ReviewCount,
		Rating:         p.Rating,
		GoodRate:       p.GoodRate,
		Featured:       p.Featured,
		Recommended:    p.Recommended,
		NewArrival:     p.NewArrival,
		HotSale:        p.HotSale,
		PromotionLevel: p.PromotionLevel,
		TrafficWeight:  p.TrafficWeight,
		Sort:           p.Sort,
		FreeShipping:   p.FreeShipping,
		ShippingFee:    p.ShippingFee,
		ShippingTemplateID: p.ShippingTemplateID,
		Weight:         p.Weight,
		Volume:         p.Volume,
		RegionID:       p.RegionID,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
	if p.Images != nil {
		info.Images = p.Images
	}
	if p.Specs != nil {
		info.Specs = p.Specs
	}
	if p.Attributes != nil {
		info.Attributes = p.Attributes
	}
	if p.Tags != nil {
		info.Tags = p.Tags
	}
	if p.SkuSpecs != nil {
		info.SkuSpecs = p.SkuSpecs
	}
	return info
}

// Create 创建商品
func (s *productService) Create(regionID, userID, shopID uint, req *dto.CreateProductRequest) (*dto.ProductInfo, error) {
	p := &model.Product{
		ShopID:      shopID,
		UserID:      userID,
		CategoryID:  req.CategoryID,
		Name:        req.Name,
		Subtitle:    req.Subtitle,
		MainImage:   req.MainImage,
		Detail:      req.Detail,
		ProductType: req.ProductType,
		Price:       req.Price,
		OriginalPrice: req.OriginalPrice,
		CostPrice:   req.CostPrice,
		Stock:       req.Stock,
		StockWarn:   req.StockWarn,
		Status:      model.ProductStatusDraft,
		FreeShipping: req.FreeShipping,
		ShippingFee: req.ShippingFee,
		Weight:      req.Weight,
		Volume:      req.Volume,
	}
	p.RegionID = regionID
	if p.ProductType == "" {
		p.ProductType = model.ProductTypePhysical
	}
	now := time.Now()
	if req.Status == 1 {
		p.Status = model.ProductStatusOnSale
		p.PublishedAt = &now
	}
	if req.Images != nil {
		if b, err := model.FromJSON(req.Images); err == nil {
			p.Images = b
		}
	}
	if req.Specs != nil {
		if b, err := model.FromJSON(req.Specs); err == nil {
			p.Specs = b
		}
	}
	if req.Attributes != nil {
		if b, err := model.FromJSON(req.Attributes); err == nil {
			p.Attributes = b
		}
	}
	if req.Tags != nil {
		if b, err := model.FromJSON(req.Tags); err == nil {
			p.Tags = b
		}
	}
	if p.MinPrice == 0 && p.MaxPrice == 0 {
		p.MinPrice = p.Price
		p.MaxPrice = p.Price
	}

	if err := s.repo.Create(p); err != nil {
		return nil, err
	}
	if shopID > 0 {
		_ = s.shopRepo.IncrProductCount(shopID)
	}
	return toProductInfo(p), nil
}

// Update 更新商品
func (s *productService) Update(id, userID uint, req *dto.UpdateProductRequest) error {
	p, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrProductNotFound
		}
		return err
	}
	if p.UserID != userID {
		return errors.New("无权操作他人商品")
	}

	fields := make(map[string]interface{})
	if req.Name != nil {
		fields["name"] = *req.Name
	}
	if req.Subtitle != nil {
		fields["subtitle"] = *req.Subtitle
	}
	if req.MainImage != nil {
		fields["main_image"] = *req.MainImage
	}
	if req.Detail != nil {
		fields["detail"] = *req.Detail
	}
	if req.CategoryID != nil {
		fields["category_id"] = *req.CategoryID
	}
	if req.BrandID != nil {
		fields["brand_id"] = *req.BrandID
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
	if req.StockWarn != nil {
		fields["stock_warn"] = *req.StockWarn
	}
	if req.Sort != nil {
		fields["sort"] = *req.Sort
	}
	if req.FreeShipping != nil {
		fields["free_shipping"] = *req.FreeShipping
	}
	if req.ShippingFee != nil {
		fields["shipping_fee"] = *req.ShippingFee
	}
	if req.Weight != nil {
		fields["weight"] = *req.Weight
	}
	if req.Volume != nil {
		fields["volume"] = *req.Volume
	}
	if req.Images != nil {
		if b, err := model.FromJSON(req.Images); err == nil {
			fields["images"] = b
		}
	}
	if req.Specs != nil {
		if b, err := model.FromJSON(req.Specs); err == nil {
			fields["specs"] = b
		}
	}
	if req.Attributes != nil {
		if b, err := model.FromJSON(req.Attributes); err == nil {
			fields["attributes"] = b
		}
	}
	if req.Tags != nil {
		if b, err := model.FromJSON(req.Tags); err == nil {
			fields["tags"] = b
		}
	}

	if len(fields) == 0 {
		return nil
	}
	return s.repo.UpdateFields(id, fields)
}

// Delete 删除商品
func (s *productService) Delete(id, userID uint) error {
	p, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrProductNotFound
		}
		return err
	}
	if p.UserID != userID {
		return errors.New("无权操作他人商品")
	}
	if err := s.repo.Delete(id); err != nil {
		return err
	}
	if p.ShopID > 0 {
		_ = s.shopRepo.DecrProductCount(p.ShopID)
	}
	return nil
}

// GetByID 获取商品详情
func (s *productService) GetByID(id uint) (*dto.ProductInfo, error) {
	p, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}
	return toProductInfo(p), nil
}

// List 商品列表
func (s *productService) List(regionID uint, req *dto.ProductListRequest) (*utils.Pagination, []dto.ProductInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.ProductListOptions{
		ShopID:       req.ShopID,
		CategoryID:   req.CategoryID,
		BrandID:      req.BrandID,
		Keyword:      req.Keyword,
		ProductType:  req.ProductType,
		MinPrice:     req.MinPrice,
		MaxPrice:     req.MaxPrice,
		Featured:     req.Featured,
		Recommended:  req.Recommended,
		NewArrival:   req.NewArrival,
		HotSale:      req.HotSale,
		FreeShipping: req.FreeShipping,
		Sort:         req.Sort,
	}
	list, total, err := s.repo.List(regionID, pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.ProductInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toProductInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// AdminList 管理后台商品列表
func (s *productService) AdminList(req *dto.ProductAdminListRequest) (*utils.Pagination, []dto.ProductInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.ProductAdminListOptions{
		RegionID:    req.RegionID,
		UserID:      req.UserID,
		ShopID:      req.ShopID,
		CategoryID:  req.CategoryID,
		BrandID:     req.BrandID,
		ProductType: req.ProductType,
		Status:      req.Status,
		AuditStatus: req.AuditStatus,
		Featured:    req.Featured,
		Recommended: req.Recommended,
		Keyword:     req.Keyword,
	}
	list, total, err := s.repo.AdminList(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.ProductInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toProductInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// Search 搜索商品
func (s *productService) Search(regionID uint, keyword string, page, pageSize int) (*utils.Pagination, []dto.ProductInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.Search(regionID, pagination, keyword)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.ProductInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toProductInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// ListByShop 按店铺列出商品
func (s *productService) ListByShop(shopID uint, page, pageSize int) (*utils.Pagination, []dto.ProductInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByShop(shopID, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.ProductInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toProductInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// ListByCategory 按分类列出商品
func (s *productService) ListByCategory(regionID, categoryID uint, page, pageSize int) (*utils.Pagination, []dto.ProductInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByCategory(regionID, categoryID, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.ProductInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toProductInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// ListByUser 按用户列出商品
func (s *productService) ListByUser(userID uint, page, pageSize int) (*utils.Pagination, []dto.ProductInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByUser(userID, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.ProductInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toProductInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// ListFeatured 精选商品
func (s *productService) ListFeatured(regionID uint, limit int) ([]dto.ProductInfo, error) {
	list, err := s.repo.ListFeatured(regionID, limit)
	if err != nil {
		return nil, err
	}
	infos := make([]dto.ProductInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toProductInfo(&list[i]))
	}
	return infos, nil
}

// ListHot 热销商品
func (s *productService) ListHot(regionID uint, limit int) ([]dto.ProductInfo, error) {
	list, err := s.repo.ListHot(regionID, limit)
	if err != nil {
		return nil, err
	}
	infos := make([]dto.ProductInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toProductInfo(&list[i]))
	}
	return infos, nil
}

// ListNew 新品
func (s *productService) ListNew(regionID uint, limit int) ([]dto.ProductInfo, error) {
	list, err := s.repo.ListNew(regionID, limit)
	if err != nil {
		return nil, err
	}
	infos := make([]dto.ProductInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toProductInfo(&list[i]))
	}
	return infos, nil
}

// UpdateStatus 更新商品状态
func (s *productService) UpdateStatus(id, userID uint, status int) error {
	p, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrProductNotFound
		}
		return err
	}
	if p.UserID != userID {
		return errors.New("无权操作他人商品")
	}
	fields := map[string]interface{}{"status": status}
	if status == model.ProductStatusOnSale {
		now := time.Now()
		fields["published_at"] = &now
	}
	return s.repo.UpdateFields(id, fields)
}

// Audit 审核商品
func (s *productService) Audit(id uint, auditStatus int, reason string) error {
	fields := map[string]interface{}{
		"audit_status": auditStatus,
	}
	if reason != "" {
		fields["audit_reason"] = reason
	}
	return s.repo.UpdateFields(id, fields)
}

// UpdatePromotion 更新商品推广配置
func (s *productService) UpdatePromotion(id uint, req *dto.PromotionRequest) error {
	fields := make(map[string]interface{})
	if req.Featured != nil {
		fields["featured"] = *req.Featured
	}
	if req.PromotionLevel != nil {
		fields["promotion_level"] = *req.PromotionLevel
	}
	if req.TrafficWeight != nil {
		fields["traffic_weight"] = *req.TrafficWeight
	}
	if len(fields) == 0 {
		return nil
	}
	return s.repo.UpdateFields(id, fields)
}

// IncrViewCount 增加浏览数
func (s *productService) IncrViewCount(id uint) error {
	return s.repo.IncrViewCount(id)
}
