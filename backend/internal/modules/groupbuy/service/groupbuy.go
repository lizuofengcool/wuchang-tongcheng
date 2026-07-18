// Package service 团购优惠券业务逻辑层
package service

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"wuchang-tongcheng/internal/modules/groupbuy/dto"
	"wuchang-tongcheng/internal/modules/groupbuy/model"
	"wuchang-tongcheng/internal/modules/groupbuy/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrGroupBuyNotFound  = errors.New("团购商品不存在")
	ErrGroupBuyNotActive = errors.New("团购商品未上架或未通过审核")
	ErrGroupBuyEnded     = errors.New("团购已结束")
	ErrStockInsufficient = errors.New("库存不足")
	ErrPerLimitExceeded  = errors.New("超过每人限购数量")
	ErrCouponNotFound    = errors.New("优惠券不存在")
	ErrCouponDisabled    = errors.New("优惠券已禁用")
	ErrCouponExpired     = errors.New("优惠券已过期")
	ErrCouponNotActive   = errors.New("优惠券尚未开始或已结束")
	ErrCouponNotReceived = errors.New("优惠券未领取或已使用")
	ErrCouponLimit       = errors.New("超过优惠券每人限领数量")
	ErrCouponNotMatch    = errors.New("优惠券不适用于该团购")
	ErrOrderNotFound     = errors.New("订单不存在")
	ErrOrderNoPermission = errors.New("无权操作此订单")
	ErrOrderStatus       = errors.New("订单状态不允许此操作")
)

// Service 团购模块统一业务服务
type Service interface {
	// 团购商品
	CreateGroupBuy(regionID, userID uint, req *dto.CreateGroupBuyRequest) (*dto.GroupBuyInfo, error)
	UpdateGroupBuy(id uint, req *dto.UpdateGroupBuyRequest) error
	DeleteGroupBuy(id uint) error
	GetGroupBuy(id uint) (*dto.GroupBuyInfo, error)
	ListGroupBuy(regionID uint, req *dto.GroupBuyListRequest) (*utils.Pagination, []dto.GroupBuyInfo, error)
	UpdateGroupBuyStatus(id uint, status int) error
	AuditGroupBuy(id uint, auditStatus int) error
	// 优惠券
	CreateCoupon(regionID uint, req *dto.CreateCouponRequest) (*dto.CouponInfo, error)
	UpdateCoupon(id uint, req *dto.UpdateCouponRequest) error
	DeleteCoupon(id uint) error
	ListCoupon(regionID uint, req *dto.CouponListRequest) (*utils.Pagination, []dto.CouponInfo, error)
	AvailableCoupons(regionID uint, req *dto.CouponListRequest) (*utils.Pagination, []dto.CouponInfo, error)
	ReceiveCoupon(regionID, userID, couponID uint) (*dto.UserCouponInfo, error)
	MyCoupons(userID uint, req *dto.CouponListRequest) (*utils.Pagination, []dto.UserCouponInfo, error)
	// 订单
	CreateOrder(regionID, userID uint, req *dto.CreateOrderRequest) (*dto.OrderInfo, error)
	GetOrder(id, userID uint) (*dto.OrderInfo, error)
	MyOrders(userID uint, req *dto.OrderListRequest) (*utils.Pagination, []dto.OrderInfo, error)
	CancelOrder(id, userID uint) error
	VerifyOrder(id uint, verifyCode string) error
	AdminOrderList(regionID uint, req *dto.OrderListRequest) (*utils.Pagination, []dto.OrderInfo, error)
}

type service struct {
	gbRepo     repository.GroupBuyRepository
	couponRepo repository.CouponRepository
	ucRepo     repository.UserCouponRepository
	orderRepo  repository.OrderRepository
}

// NewService 创建团购服务
func NewService(gbRepo repository.GroupBuyRepository, couponRepo repository.CouponRepository, ucRepo repository.UserCouponRepository, orderRepo repository.OrderRepository) Service {
	return &service{
		gbRepo:     gbRepo,
		couponRepo: couponRepo,
		ucRepo:     ucRepo,
		orderRepo:  orderRepo,
	}
}

// ===== 转换函数 =====

func toGroupBuyInfo(g *model.GroupBuy) *dto.GroupBuyInfo {
	return &dto.GroupBuyInfo{
		ID:            g.ID,
		Title:         g.Title,
		Description:   g.Description,
		Cover:         g.Cover,
		OriginalPrice: g.OriginalPrice,
		GroupBuyPrice: g.GroupBuyPrice,
		Stock:         g.Stock,
		SoldCount:     g.SoldCount,
		PerLimit:      g.PerLimit,
		StartTime:     g.StartTime,
		EndTime:       g.EndTime,
		Status:        g.Status,
		AuditStatus:   g.AuditStatus,
		IsRecommend:   g.IsRecommend,
		Sort:          g.Sort,
		ShopID:        g.ShopID,
		UserID:        g.UserID,
		CreatedAt:     g.CreatedAt,
	}
}

func toCouponInfo(c *model.Coupon) *dto.CouponInfo {
	return &dto.CouponInfo{
		ID:            c.ID,
		Name:          c.Name,
		Type:          c.Type,
		Value:         c.Value,
		MinAmount:     c.MinAmount,
		Scope:         c.Scope,
		ScopeID:       c.ScopeID,
		TotalCount:    c.TotalCount,
		ReceivedCount: c.ReceivedCount,
		UsedCount:     c.UsedCount,
		PerLimit:      c.PerLimit,
		StartTime:     c.StartTime,
		EndTime:       c.EndTime,
		ValidityType:  c.ValidityType,
		ValidDays:     c.ValidDays,
		Status:        c.Status,
		CreatedAt:     c.CreatedAt,
	}
}

func toUserCouponInfo(uc *model.UserCoupon, c *model.Coupon) *dto.UserCouponInfo {
	info := &dto.UserCouponInfo{
		ID:         uc.ID,
		UserID:     uc.UserID,
		CouponID:   uc.CouponID,
		Status:     uc.Status,
		ReceivedAt: uc.ReceivedAt,
		UsedAt:     uc.UsedAt,
		ExpireAt:   uc.ExpireAt,
		OrderID:    uc.OrderID,
	}
	if c != nil {
		info.Coupon = toCouponInfo(c)
	}
	return info
}

func toOrderInfo(o *model.GroupBuyOrder, g *model.GroupBuy) *dto.OrderInfo {
	info := &dto.OrderInfo{
		ID:             o.ID,
		OrderNo:        o.OrderNo,
		UserID:         o.UserID,
		GroupBuyID:     o.GroupBuyID,
		Quantity:       o.Quantity,
		UnitPrice:      o.UnitPrice,
		TotalPrice:     o.TotalPrice,
		CouponID:       o.CouponID,
		DiscountAmount: o.DiscountAmount,
		PayAmount:      o.PayAmount,
		PayStatus:      o.PayStatus,
		PayAt:          o.PayAt,
		VerifyStatus:   o.VerifyStatus,
		VerifyCode:     o.VerifyCode,
		VerifyAt:       o.VerifyAt,
		Status:         o.Status,
		CreatedAt:      o.CreatedAt,
	}
	if g != nil {
		info.GroupBuy = toGroupBuyInfo(g)
	}
	return info
}

// ===== 团购商品 =====

// CreateGroupBuy 创建团购商品
func (s *service) CreateGroupBuy(regionID, userID uint, req *dto.CreateGroupBuyRequest) (*dto.GroupBuyInfo, error) {
	gb := &model.GroupBuy{
		Title:         req.Title,
		Description:   req.Description,
		Cover:         req.Cover,
		OriginalPrice: req.OriginalPrice,
		GroupBuyPrice: req.GroupBuyPrice,
		Stock:         req.Stock,
		SoldCount:     0,
		PerLimit:      req.PerLimit,
		StartTime:     req.StartTime,
		EndTime:       req.EndTime,
		Status:        0, // 默认下架，需审核通过后上架
		IsRecommend:   req.IsRecommend,
		Sort:          req.Sort,
		ShopID:        req.ShopID,
		UserID:        userID,
		AuditStatus:   0, // 默认待审核
	}
	if gb.PerLimit <= 0 {
		gb.PerLimit = 1
	}
	gb.RegionID = regionID

	if err := s.gbRepo.Create(gb); err != nil {
		return nil, err
	}
	return toGroupBuyInfo(gb), nil
}

// UpdateGroupBuy 更新团购商品
func (s *service) UpdateGroupBuy(id uint, req *dto.UpdateGroupBuyRequest) error {
	gb, err := s.gbRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrGroupBuyNotFound
		}
		return err
	}

	fields := map[string]interface{}{}
	if req.Title != "" {
		fields["title"] = req.Title
	}
	if req.Description != "" {
		fields["description"] = req.Description
	}
	if req.Cover != "" {
		fields["cover"] = req.Cover
	}
	if req.OriginalPrice > 0 {
		fields["original_price"] = req.OriginalPrice
	}
	if req.GroupBuyPrice > 0 {
		fields["groupbuy_price"] = req.GroupBuyPrice
	}
	if req.Stock > 0 {
		fields["stock"] = req.Stock
	}
	if req.PerLimit > 0 {
		fields["per_limit"] = req.PerLimit
	}
	if req.StartTime != nil {
		fields["start_time"] = req.StartTime
	}
	if req.EndTime != nil {
		fields["end_time"] = req.EndTime
	}
	fields["is_recommend"] = req.IsRecommend
	fields["sort"] = req.Sort
	// 编辑后重回待审核
	if len(fields) > 0 && gb.AuditStatus == 1 {
		fields["audit_status"] = 0
	}

	if len(fields) == 0 {
		return nil
	}
	return s.gbRepo.UpdateFields(id, fields)
}

// DeleteGroupBuy 删除团购商品
func (s *service) DeleteGroupBuy(id uint) error {
	if _, err := s.gbRepo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrGroupBuyNotFound
		}
		return err
	}
	return s.gbRepo.Delete(id)
}

// GetGroupBuy 获取团购详情
func (s *service) GetGroupBuy(id uint) (*dto.GroupBuyInfo, error) {
	gb, err := s.gbRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGroupBuyNotFound
		}
		return nil, err
	}
	return toGroupBuyInfo(gb), nil
}

// ListGroupBuy 团购列表
func (s *service) ListGroupBuy(regionID uint, req *dto.GroupBuyListRequest) (*utils.Pagination, []dto.GroupBuyInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	list, total, err := s.gbRepo.List(regionID, pagination, req.Keyword, req.Status, req.IsRecommend, req.ShopID)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total

	result := make([]dto.GroupBuyInfo, 0, len(list))
	for i := range list {
		result = append(result, *toGroupBuyInfo(&list[i]))
	}
	return pagination, result, nil
}

// UpdateGroupBuyStatus 上架/下架
func (s *service) UpdateGroupBuyStatus(id uint, status int) error {
	if _, err := s.gbRepo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrGroupBuyNotFound
		}
		return err
	}
	return s.gbRepo.UpdateFields(id, map[string]interface{}{"status": status})
}

// AuditGroupBuy 审核团购
func (s *service) AuditGroupBuy(id uint, auditStatus int) error {
	if _, err := s.gbRepo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrGroupBuyNotFound
		}
		return err
	}
	return s.gbRepo.UpdateFields(id, map[string]interface{}{"audit_status": auditStatus})
}

// ===== 优惠券 =====

// CreateCoupon 创建优惠券
func (s *service) CreateCoupon(regionID uint, req *dto.CreateCouponRequest) (*dto.CouponInfo, error) {
	c := &model.Coupon{
		Name:         req.Name,
		Type:         req.Type,
		Value:        req.Value,
		MinAmount:    req.MinAmount,
		Scope:        req.Scope,
		ScopeID:      req.ScopeID,
		TotalCount:   req.TotalCount,
		PerLimit:     req.PerLimit,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
		ValidityType: req.ValidityType,
		ValidDays:    req.ValidDays,
		Status:       req.Status,
	}
	if c.PerLimit <= 0 {
		c.PerLimit = 1
	}
	if c.Status == 0 {
		c.Status = 1
	}
	c.RegionID = regionID

	if err := s.couponRepo.Create(c); err != nil {
		return nil, err
	}
	return toCouponInfo(c), nil
}

// UpdateCoupon 更新优惠券
func (s *service) UpdateCoupon(id uint, req *dto.UpdateCouponRequest) error {
	if _, err := s.couponRepo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCouponNotFound
		}
		return err
	}

	fields := map[string]interface{}{}
	if req.Name != "" {
		fields["name"] = req.Name
	}
	if req.Type >= 1 && req.Type <= 3 {
		fields["type"] = req.Type
	}
	if req.Value > 0 {
		fields["value"] = req.Value
	}
	if req.MinAmount > 0 {
		fields["min_amount"] = req.MinAmount
	}
	if req.Scope >= 0 && req.Scope <= 2 {
		fields["scope"] = req.Scope
		fields["scope_id"] = req.ScopeID
	}
	if req.TotalCount > 0 {
		fields["total_count"] = req.TotalCount
	}
	if req.PerLimit > 0 {
		fields["per_limit"] = req.PerLimit
	}
	if req.StartTime != nil {
		fields["start_time"] = req.StartTime
	}
	if req.EndTime != nil {
		fields["end_time"] = req.EndTime
	}
	if req.ValidityType == 0 || req.ValidityType == 1 {
		fields["validity_type"] = req.ValidityType
		fields["valid_days"] = req.ValidDays
	}
	if req.Status == 0 || req.Status == 1 {
		fields["status"] = req.Status
	}

	if len(fields) == 0 {
		return nil
	}
	return s.couponRepo.UpdateFields(id, fields)
}

// DeleteCoupon 删除优惠券
func (s *service) DeleteCoupon(id uint) error {
	if _, err := s.couponRepo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCouponNotFound
		}
		return err
	}
	return s.couponRepo.Delete(id)
}

// ListCoupon 优惠券列表（管理端）
func (s *service) ListCoupon(regionID uint, req *dto.CouponListRequest) (*utils.Pagination, []dto.CouponInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	list, total, err := s.couponRepo.List(regionID, pagination, req.Status, req.Type)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total

	result := make([]dto.CouponInfo, 0, len(list))
	for i := range list {
		result = append(result, *toCouponInfo(&list[i]))
	}
	return pagination, result, nil
}

// AvailableCoupons 可领取的优惠券列表（公开）
func (s *service) AvailableCoupons(regionID uint, req *dto.CouponListRequest) (*utils.Pagination, []dto.CouponInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	list, total, err := s.couponRepo.AvailableList(regionID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total

	result := make([]dto.CouponInfo, 0, len(list))
	for i := range list {
		result = append(result, *toCouponInfo(&list[i]))
	}
	return pagination, result, nil
}

// ReceiveCoupon 领取优惠券
func (s *service) ReceiveCoupon(regionID, userID, couponID uint) (*dto.UserCouponInfo, error) {
	c, err := s.couponRepo.FindByID(couponID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCouponNotFound
		}
		return nil, err
	}
	if c.Status != 1 {
		return nil, ErrCouponDisabled
	}
	now := time.Now()
	if c.StartTime != nil && now.Before(*c.StartTime) {
		return nil, ErrCouponNotActive
	}
	if c.EndTime != nil && now.After(*c.EndTime) {
		return nil, ErrCouponExpired
	}
	// 达到发放总量
	if c.TotalCount > 0 && c.ReceivedCount >= c.TotalCount {
		return nil, ErrCouponLimit
	}
	// 每人限领
	if c.PerLimit > 0 {
		count, err := s.ucRepo.CountByUserAndCoupon(userID, couponID)
		if err != nil {
			return nil, err
		}
		if count >= int64(c.PerLimit) {
			return nil, ErrCouponLimit
		}
	}

	// 计算过期时间
	var expireAt *time.Time
	if c.ValidityType == 1 && c.ValidDays > 0 {
		end := now.AddDate(0, 0, c.ValidDays)
		expireAt = &end
	} else if c.EndTime != nil {
		expireAt = c.EndTime
	}

	uc := &model.UserCoupon{
		UserID:     userID,
		CouponID:   couponID,
		Status:     0,
		ReceivedAt: now,
		ExpireAt:   expireAt,
	}
	uc.RegionID = regionID

	if err := s.ucRepo.Create(uc); err != nil {
		return nil, err
	}
	// 优惠券已领取数+1
	_ = s.couponRepo.IncrReceivedCount(couponID)

	return toUserCouponInfo(uc, c), nil
}

// MyCoupons 我的优惠券
func (s *service) MyCoupons(userID uint, req *dto.CouponListRequest) (*utils.Pagination, []dto.UserCouponInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	list, total, err := s.ucRepo.ListByUser(userID, pagination, req.Status)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total

	result := make([]dto.UserCouponInfo, 0, len(list))
	for i := range list {
		c, _ := s.couponRepo.FindByID(list[i].CouponID)
		result = append(result, *toUserCouponInfo(&list[i], c))
	}
	return pagination, result, nil
}

// ===== 订单 =====

// CreateOrder 创建团购订单
func (s *service) CreateOrder(regionID, userID uint, req *dto.CreateOrderRequest) (*dto.OrderInfo, error) {
	gb, err := s.gbRepo.FindByID(req.GroupBuyID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGroupBuyNotFound
		}
		return nil, err
	}
	// 校验上架且审核通过
	if gb.Status != 1 || gb.AuditStatus != 1 {
		return nil, ErrGroupBuyNotActive
	}
	// 校验时间范围
	now := time.Now()
	if gb.StartTime != nil && now.Before(*gb.StartTime) {
		return nil, ErrGroupBuyNotActive
	}
	if gb.EndTime != nil && now.After(*gb.EndTime) {
		return nil, ErrGroupBuyEnded
	}
	// 校验库存
	if gb.Stock < req.Quantity {
		return nil, ErrStockInsufficient
	}
	// 校验每人限购
	if gb.PerLimit > 0 {
		bought, err := s.orderRepo.SumQuantityByUserAndGroupBuy(userID, req.GroupBuyID)
		if err != nil {
			return nil, err
		}
		if bought+int64(req.Quantity) > int64(gb.PerLimit) {
			return nil, ErrPerLimitExceeded
		}
	}

	unitPrice := gb.GroupBuyPrice
	totalPrice := unitPrice * float64(req.Quantity)
	discount := float64(0)
	couponID := uint(0)
	userCouponID := uint(0)

	// 优惠券处理
	if req.UserCouponID > 0 {
		uc, err := s.ucRepo.FindByID(req.UserCouponID)
		if err != nil {
			return nil, ErrCouponNotReceived
		}
		if uc.UserID != userID || uc.Status != 0 {
			return nil, ErrCouponNotReceived
		}
		// 校验未过期
		if uc.ExpireAt != nil && now.After(*uc.ExpireAt) {
			return nil, ErrCouponExpired
		}
		c, err := s.couponRepo.FindByID(uc.CouponID)
		if err != nil || c.Status != 1 {
			return nil, ErrCouponDisabled
		}
		// 校验最低消费
		if totalPrice < c.MinAmount {
			return nil, ErrCouponNotMatch
		}
		// 校验适用范围
		if c.Scope == 1 && (gb.ShopID == 0 || c.ScopeID != gb.ShopID) {
			return nil, ErrCouponNotMatch
		}
		// 计算优惠金额
		switch c.Type {
		case 1: // 满减
			discount = c.Value
		case 2: // 折扣（value为折扣率，如0.8表示8折）
			if c.Value > 0 && c.Value < 1 {
				discount = totalPrice * (1 - c.Value)
			}
		case 3: // 代金券
			discount = c.Value
		}
		if discount > totalPrice {
			discount = totalPrice
		}
		couponID = c.ID
		userCouponID = uc.ID
	}

	payAmount := totalPrice - discount
	if payAmount < 0 {
		payAmount = 0
	}

	payAt := now
	order := &model.GroupBuyOrder{
		OrderNo:        genOrderNo(),
		UserID:         userID,
		GroupBuyID:     req.GroupBuyID,
		Quantity:       req.Quantity,
		UnitPrice:      unitPrice,
		TotalPrice:     totalPrice,
		CouponID:       couponID,
		UserCouponID:   userCouponID,
		DiscountAmount: discount,
		PayAmount:      payAmount,
		PayStatus:      1, // 已支付（无独立支付接口，下单即支付）
		PayAt:          &payAt,
		VerifyStatus:   0,
		VerifyCode:     genVerifyCode(),
		Status:         1, // 已支付
	}
	order.RegionID = regionID

	if err := s.orderRepo.CreateOrderInTransaction(order, req.GroupBuyID, userCouponID, req.Quantity); err != nil {
		return nil, err
	}
	return toOrderInfo(order, gb), nil
}

// GetOrder 获取订单详情
func (s *service) GetOrder(id, userID uint) (*dto.OrderInfo, error) {
	order, err := s.orderRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}
	// 普通用户只能查看自己的订单（userID=0 表示管理员查询，不校验）
	if userID > 0 && order.UserID != userID {
		return nil, ErrOrderNoPermission
	}
	gb, _ := s.gbRepo.FindByID(order.GroupBuyID)
	return toOrderInfo(order, gb), nil
}

// MyOrders 我的订单列表
func (s *service) MyOrders(userID uint, req *dto.OrderListRequest) (*utils.Pagination, []dto.OrderInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	req.UserID = userID
	list, total, err := s.orderRepo.List(0, pagination, req.UserID, req.GroupBuyID, req.Status, req.PayStatus)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total

	result := make([]dto.OrderInfo, 0, len(list))
	for i := range list {
		gb, _ := s.gbRepo.FindByID(list[i].GroupBuyID)
		result = append(result, *toOrderInfo(&list[i], gb))
	}
	return pagination, result, nil
}

// CancelOrder 取消订单
func (s *service) CancelOrder(id, userID uint) error {
	order, err := s.orderRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrOrderNotFound
		}
		return err
	}
	if order.UserID != userID {
		return ErrOrderNoPermission
	}
	if order.Status != 1 {
		return ErrOrderStatus
	}
	return s.orderRepo.CancelOrderInTransaction(order)
}

// VerifyOrder 核销订单
func (s *service) VerifyOrder(id uint, verifyCode string) error {
	order, err := s.orderRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrOrderNotFound
		}
		return err
	}
	if order.Status != 1 {
		return ErrOrderStatus
	}
	return s.orderRepo.VerifyOrder(id, verifyCode)
}

// AdminOrderList 管理端订单列表
func (s *service) AdminOrderList(regionID uint, req *dto.OrderListRequest) (*utils.Pagination, []dto.OrderInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	list, total, err := s.orderRepo.List(regionID, pagination, req.UserID, req.GroupBuyID, req.Status, req.PayStatus)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total

	result := make([]dto.OrderInfo, 0, len(list))
	for i := range list {
		gb, _ := s.gbRepo.FindByID(list[i].GroupBuyID)
		result = append(result, *toOrderInfo(&list[i], gb))
	}
	return pagination, result, nil
}

// ===== 工具函数 =====

// genOrderNo 生成订单号：GB + 时间 + 随机数
func genOrderNo() string {
	return fmt.Sprintf("GB%s%04d", time.Now().Format("20060102150405"), rand.Intn(10000))
}

// genVerifyCode 生成8位核销码
func genVerifyCode() string {
	return fmt.Sprintf("%08d", rand.Intn(100000000))
}
