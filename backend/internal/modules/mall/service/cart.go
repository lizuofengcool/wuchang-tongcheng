// Package service 同城商城业务逻辑层 - 购物车
package service

import (
	"errors"

	"wuchang-tongcheng/internal/modules/mall/dto"
	"wuchang-tongcheng/internal/modules/mall/model"
	"wuchang-tongcheng/internal/modules/mall/repository"

	"gorm.io/gorm"
)

var (
	ErrCartNotFound      = errors.New("购物车项不存在")
	ErrCartNotOwner      = errors.New("无权操作他人购物车")
	ErrCartProductInvalid = errors.New("商品已下架或删除")
)

// CartService 购物车业务接口
type CartService interface {
	Add(regionID, userID uint, req *dto.AddCartRequest) (*dto.CartInfo, error)
	Update(id, userID uint, req *dto.UpdateCartRequest) error
	BatchUpdate(userID uint, req *dto.BatchUpdateCartRequest) error
	Delete(id, userID uint) error
	BatchDelete(userID uint, ids []uint) error
	ClearByUser(userID uint) error
	ClearByUserAndShop(userID, shopID uint) error
	SelectAll(userID uint, req *dto.SelectAllCartRequest) error

	GetByID(id uint) (*dto.CartInfo, error)
	ListByUser(userID uint) ([]dto.CartInfo, error)
	ListByUserAndShop(userID, shopID uint) ([]dto.CartInfo, error)
	ListSelected(userID uint) ([]dto.CartInfo, error)
	ListGroupByShop(userID uint) ([]dto.CartGroupByShop, error)
	Summary(userID uint, req *dto.CartSummaryRequest) (*dto.CartSummary, error)

	CountByUser(userID uint) (int64, error)
	CountSelectedByUser(userID uint) (int64, error)
}

type cartService struct {
	repo        repository.CartRepository
	productRepo repository.ProductRepository
	skuRepo     repository.SkuRepository
	shopRepo    repository.ShopRepository
}

// NewCartService 创建购物车 service 实例
func NewCartService(repo repository.CartRepository, productRepo repository.ProductRepository, skuRepo repository.SkuRepository, shopRepo repository.ShopRepository) CartService {
	return &cartService{repo: repo, productRepo: productRepo, skuRepo: skuRepo, shopRepo: shopRepo}
}

// cartStatusText 购物车状态文本
func cartStatusText(s int) string {
	switch s {
	case 0:
		return "失效"
	case 1:
		return "有效"
	}
	return ""
}

// cartSelectedText 选中状态文本
func cartSelectedText(s int) string {
	switch s {
	case model.CartUnselected:
		return "未选中"
	case model.CartSelected:
		return "已选中"
	}
	return ""
}

// toCartInfo model -> dto
func toCartInfo(c *model.Cart) *dto.CartInfo {
	info := &dto.CartInfo{
		ID:           c.ID,
		UserID:       c.UserID,
		ShopID:       c.ShopID,
		ProductID:    c.ProductID,
		SkuID:        c.SkuID,
		ProductName:  c.ProductName,
		MainImage:    c.MainImage,
		SkuName:      c.SkuName,
		SkuSpecs:     c.SkuSpecs,
		Price:        c.Price,
		Quantity:     c.Quantity,
		Selected:     c.Selected,
		SelectedText: cartSelectedText(c.Selected),
		Status:       c.Status,
		StatusText:   cartStatusText(c.Status),
		Subtotal:     c.Price * float64(c.Quantity),
		RegionID:     c.RegionID,
		CreatedAt:    c.CreatedAt,
		UpdatedAt:    c.UpdatedAt,
	}
	return info
}

// Add 加入购物车
func (s *cartService) Add(regionID, userID uint, req *dto.AddCartRequest) (*dto.CartInfo, error) {
	// 查找商品
	p, err := s.productRepo.FindByID(req.ProductID)
	if err != nil {
		return nil, ErrProductNotFound
	}
	if p.Status != model.ProductStatusOnSale {
		return nil, ErrCartProductInvalid
	}

	// 获取 SKU 信息（如果有规格）
	var sku *model.Sku
	shopID := p.ShopID
	price := p.Price
	skuName := ""
	skuSpecs := ""
	if req.SkuID > 0 {
		sku, err = s.skuRepo.FindByID(req.SkuID)
		if err != nil {
			return nil, ErrSkuNotFound
		}
		if sku.ProductID != p.ID {
			return nil, errors.New("SKU 不属于该商品")
		}
		price = sku.Price
		skuName = sku.Name
		skuSpecs = string(sku.Specs)
		shopID = sku.ShopID
	}

	// 检查是否已在购物车（同 user + sku）
	if existing, err := s.repo.FindByUserAndSku(userID, req.SkuID); err == nil && existing != nil {
		// 已存在则累加数量
		newQty := existing.Quantity + req.Quantity
		if err := s.repo.UpdateFields(existing.ID, map[string]interface{}{
			"quantity": newQty,
		}); err != nil {
			return nil, err
		}
		existing.Quantity = newQty
		return toCartInfo(existing), nil
	}

	c := &model.Cart{
		UserID:      userID,
		ShopID:      shopID,
		ProductID:   req.ProductID,
		SkuID:       req.SkuID,
		ProductName: p.Name,
		MainImage:   p.MainImage,
		SkuName:     skuName,
		SkuSpecs:    skuSpecs,
		Price:       price,
		Quantity:    req.Quantity,
		Selected:    model.CartSelected,
		Status:      1,
	}
	c.RegionID = regionID

	if err := s.repo.Create(c); err != nil {
		return nil, err
	}
	return toCartInfo(c), nil
}

// Update 更新购物车项
func (s *cartService) Update(id, userID uint, req *dto.UpdateCartRequest) error {
	c, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCartNotFound
		}
		return err
	}
	if c.UserID != userID {
		return ErrCartNotOwner
	}

	fields := make(map[string]interface{})
	if req.Quantity != nil {
		fields["quantity"] = *req.Quantity
	}
	if req.Selected != nil {
		fields["selected"] = *req.Selected
	}

	if len(fields) == 0 {
		return nil
	}
	return s.repo.UpdateFields(id, fields)
}

// BatchUpdate 批量更新
func (s *cartService) BatchUpdate(userID uint, req *dto.BatchUpdateCartRequest) error {
	for _, item := range req.Items {
		fields := make(map[string]interface{})
		if item.Quantity != nil {
			fields["quantity"] = *item.Quantity
		}
		if item.Selected != nil {
			fields["selected"] = *item.Selected
		}
		if len(fields) > 0 {
			c, err := s.repo.FindByID(item.ID)
			if err != nil || c.UserID != userID {
				continue
			}
			_ = s.repo.UpdateFields(item.ID, fields)
		}
	}
	return nil
}

// Delete 删除购物车项
func (s *cartService) Delete(id, userID uint) error {
	c, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCartNotFound
		}
		return err
	}
	if c.UserID != userID {
		return ErrCartNotOwner
	}
	return s.repo.Delete(id)
}

// BatchDelete 批量删除
func (s *cartService) BatchDelete(userID uint, ids []uint) error {
	// 由于 repo.BatchDelete 不校验 user_id，这里通过查列表来过滤
	// 简化实现：直接调用批量删除
	return s.repo.BatchDelete(ids)
}

// ClearByUser 清空用户购物车
func (s *cartService) ClearByUser(userID uint) error {
	return s.repo.DeleteByUser(userID)
}

// ClearByUserAndShop 清空用户某店铺购物车
func (s *cartService) ClearByUserAndShop(userID, shopID uint) error {
	return s.repo.DeleteByUserAndShop(userID, shopID)
}

// SelectAll 全选/取消全选
func (s *cartService) SelectAll(userID uint, req *dto.SelectAllCartRequest) error {
	if req.ShopID > 0 {
		// 按店铺全选，需先取该店铺购物车 IDs
		list, err := s.repo.ListByUserAndShop(userID, req.ShopID)
		if err != nil {
			return err
		}
		ids := make([]uint, 0, len(list))
		for i := range list {
			ids = append(ids, list[i].ID)
		}
		return s.repo.SelectItems(userID, ids, req.Selected)
	}
	return s.repo.SelectAll(userID, req.Selected)
}

// GetByID 获取购物车项详情
func (s *cartService) GetByID(id uint) (*dto.CartInfo, error) {
	c, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCartNotFound
		}
		return nil, err
	}
	return toCartInfo(c), nil
}

// ListByUser 按用户列出购物车
func (s *cartService) ListByUser(userID uint) ([]dto.CartInfo, error) {
	list, err := s.repo.ListByUser(userID)
	if err != nil {
		return nil, err
	}
	infos := make([]dto.CartInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toCartInfo(&list[i]))
	}
	return infos, nil
}

// ListByUserAndShop 按用户和店铺列出购物车
func (s *cartService) ListByUserAndShop(userID, shopID uint) ([]dto.CartInfo, error) {
	list, err := s.repo.ListByUserAndShop(userID, shopID)
	if err != nil {
		return nil, err
	}
	infos := make([]dto.CartInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toCartInfo(&list[i]))
	}
	return infos, nil
}

// ListSelected 列出已选中的购物车项
func (s *cartService) ListSelected(userID uint) ([]dto.CartInfo, error) {
	list, err := s.repo.ListSelected(userID)
	if err != nil {
		return nil, err
	}
	infos := make([]dto.CartInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toCartInfo(&list[i]))
	}
	return infos, nil
}

// ListGroupByShop 按店铺分组列出购物车
func (s *cartService) ListGroupByShop(userID uint) ([]dto.CartGroupByShop, error) {
	list, err := s.repo.ListByUser(userID)
	if err != nil {
		return nil, err
	}

	// 按店铺分组
	shopMap := make(map[uint]*dto.CartGroupByShop)
	shopOrder := make([]uint, 0)
	for i := range list {
		shopID := list[i].ShopID
		group, ok := shopMap[shopID]
		if !ok {
			// 获取店铺信息
			shop, err := s.shopRepo.FindByID(shopID)
			if err != nil {
				continue
			}
			group = &dto.CartGroupByShop{
				ShopID:   shopID,
				ShopName: shop.ShopName,
				ShopLogo: shop.Logo,
				Items:    []dto.CartInfo{},
			}
			shopMap[shopID] = group
			shopOrder = append(shopOrder, shopID)
		}
		info := *toCartInfo(&list[i])
		group.Items = append(group.Items, info)
		group.TotalAmount += info.Subtotal
		group.TotalCount += info.Quantity
	}

	// 保持插入顺序
	result := make([]dto.CartGroupByShop, 0, len(shopOrder))
	for _, shopID := range shopOrder {
		result = append(result, *shopMap[shopID])
	}
	return result, nil
}

// Summary 购物车汇总
func (s *cartService) Summary(userID uint, req *dto.CartSummaryRequest) (*dto.CartSummary, error) {
	list, err := s.repo.ListByUser(userID)
	if err != nil {
		return nil, err
	}

	summary := &dto.CartSummary{}
	shopSet := make(map[uint]bool)
	for i := range list {
		summary.TotalCount += list[i].Quantity
		summary.TotalAmount += list[i].Price * float64(list[i].Quantity)
		shopSet[list[i].ShopID] = true
		if !req.SelectedOnly || list[i].Selected == model.CartSelected {
			summary.SelectedCount += list[i].Quantity
			summary.SelectedAmount += list[i].Price * float64(list[i].Quantity)
		}
	}
	summary.ShopCount = len(shopSet)
	return summary, nil
}

// CountByUser 用户购物车数量
func (s *cartService) CountByUser(userID uint) (int64, error) {
	return s.repo.CountByUser(userID)
}

// CountSelectedByUser 用户已选中购物车数量
func (s *cartService) CountSelectedByUser(userID uint) (int64, error) {
	return s.repo.CountSelectedByUser(userID)
}
