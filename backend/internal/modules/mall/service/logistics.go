// Package service 同城商城业务逻辑层 - 物流
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
	ErrLogisticsNotFound = errors.New("物流记录不存在")
)

// LogisticsService 物流业务接口
type LogisticsService interface {
	Create(regionID, userID uint, req *dto.CreateLogisticsRequest) (*dto.LogisticsInfo, error)
	Update(id uint, req *dto.UpdateLogisticsRequest) error
	Delete(id uint) error
	GetByID(id uint) (*dto.LogisticsInfo, error)
	GetByOrderID(orderID uint) (*dto.LogisticsInfo, error)
	GetByTrackingNo(trackingNo string) (*dto.LogisticsInfo, error)
	List(req *dto.LogisticsListRequest) (*utils.Pagination, []dto.LogisticsInfo, error)
	ListByUser(userID uint, page, pageSize int) (*utils.Pagination, []dto.LogisticsInfo, error)
	ListByShop(shopID uint, page, pageSize int) (*utils.Pagination, []dto.LogisticsInfo, error)
	UpdateStatus(id uint, status int) error
	UpdateTraces(id uint, traces interface{}) error
	UpdateStatusByCallback(req *dto.UpdateLogisticsStatusRequest) error
}

type logisticsService struct {
	repo repository.LogisticsRepository
}

// NewLogisticsService 创建物流 service 实例
func NewLogisticsService(repo repository.LogisticsRepository) LogisticsService {
	return &logisticsService{repo: repo}
}

// logisticsStatusText 物流状态文本
func logisticsStatusText(s int) string {
	switch s {
	case model.LogisticsStatusPending:
		return "待发货"
	case model.LogisticsStatusShipped:
		return "已发货"
	case model.LogisticsStatusInTransit:
		return "运输中"
	case model.LogisticsStatusDelivered:
		return "已派送"
	case model.LogisticsStatusReceived:
		return "已签收"
	case model.LogisticsStatusReturned:
		return "已退回"
	}
	return ""
}

// toLogisticsInfo model -> dto
func toLogisticsInfo(l *model.Logistics) *dto.LogisticsInfo {
	info := &dto.LogisticsInfo{
		ID:              l.ID,
		OrderID:         l.OrderID,
		OrderNo:         l.OrderNo,
		UserID:          l.UserID,
		ShopID:          l.ShopID,
		Company:         l.Company,
		CompanyCode:     l.CompanyCode,
		TrackingNo:      l.TrackingNo,
		CourierName:     l.CourierName,
		CourierPhone:    l.CourierPhone,
		SenderName:      l.SenderName,
		SenderPhone:     l.SenderPhone,
		SenderAddress:   l.SenderAddress,
		ReceiverName:    l.ReceiverName,
		ReceiverPhone:   l.ReceiverPhone,
		ReceiverAddress: l.ReceiverAddress,
		Status:          l.Status,
		StatusText:      logisticsStatusText(l.Status),
		ShippedAt:       l.ShippedAt,
		InTransitAt:     l.InTransitAt,
		DeliveredAt:     l.DeliveredAt,
		ReceivedAt:      l.ReceivedAt,
		ReturnedAt:      l.ReturnedAt,
		Weight:          l.Weight,
		Volume:          l.Volume,
		Pieces:          l.Pieces,
		Freight:         l.Freight,
		InsuredFee:      l.InsuredFee,
		CodFee:          l.CodFee,
		RegionID:        l.RegionID,
		CreatedAt:       l.CreatedAt,
		UpdatedAt:       l.UpdatedAt,
	}
	if l.Traces != nil {
		info.Traces = l.Traces
	}
	return info
}

// Create 创建物流记录
func (s *logisticsService) Create(regionID, userID uint, req *dto.CreateLogisticsRequest) (*dto.LogisticsInfo, error) {
	pieces := req.Pieces
	if pieces == 0 {
		pieces = 1
	}
	l := &model.Logistics{
		OrderID:       req.OrderID,
		Company:       req.Company,
		CompanyCode:   req.CompanyCode,
		TrackingNo:    req.TrackingNo,
		CourierName:   req.CourierName,
		CourierPhone:  req.CourierPhone,
		SenderName:    req.SenderName,
		SenderPhone:   req.SenderPhone,
		SenderAddress: req.SenderAddress,
		Weight:        req.Weight,
		Volume:        req.Volume,
		Pieces:        pieces,
		Freight:       req.Freight,
		InsuredFee:    req.InsuredFee,
		CodFee:        req.CodFee,
		Status:        model.LogisticsStatusShipped,
	}
	l.RegionID = regionID
	now := time.Now()
	l.ShippedAt = &now

	if err := s.repo.Create(l); err != nil {
		return nil, err
	}
	return toLogisticsInfo(l), nil
}

// Update 更新物流
func (s *logisticsService) Update(id uint, req *dto.UpdateLogisticsRequest) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrLogisticsNotFound
		}
		return err
	}

	fields := make(map[string]interface{})
	if req.Company != nil {
		fields["company"] = *req.Company
	}
	if req.CompanyCode != nil {
		fields["company_code"] = *req.CompanyCode
	}
	if req.TrackingNo != nil {
		fields["tracking_no"] = *req.TrackingNo
	}
	if req.CourierName != nil {
		fields["courier_name"] = *req.CourierName
	}
	if req.CourierPhone != nil {
		fields["courier_phone"] = *req.CourierPhone
	}
	if req.Status != nil {
		fields["status"] = *req.Status
		// 同时更新对应状态的时间戳
		now := time.Now()
		switch *req.Status {
		case model.LogisticsStatusShipped:
			fields["shipped_at"] = &now
		case model.LogisticsStatusInTransit:
			fields["in_transit_at"] = &now
		case model.LogisticsStatusDelivered:
			fields["delivered_at"] = &now
		case model.LogisticsStatusReceived:
			fields["received_at"] = &now
		case model.LogisticsStatusReturned:
			fields["returned_at"] = &now
		}
	}
	if req.Traces != nil {
		if b, err := model.FromJSON(req.Traces); err == nil {
			fields["traces"] = b
		}
	}

	if len(fields) == 0 {
		return nil
	}
	return s.repo.UpdateFields(id, fields)
}

// Delete 删除物流记录
func (s *logisticsService) Delete(id uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrLogisticsNotFound
		}
		return err
	}
	return s.repo.Delete(id)
}

// GetByID 获取物流详情
func (s *logisticsService) GetByID(id uint) (*dto.LogisticsInfo, error) {
	l, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrLogisticsNotFound
		}
		return nil, err
	}
	return toLogisticsInfo(l), nil
}

// GetByOrderID 按订单 ID 查询物流
func (s *logisticsService) GetByOrderID(orderID uint) (*dto.LogisticsInfo, error) {
	l, err := s.repo.FindByOrderID(orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrLogisticsNotFound
		}
		return nil, err
	}
	return toLogisticsInfo(l), nil
}

// GetByTrackingNo 按物流单号查询
func (s *logisticsService) GetByTrackingNo(trackingNo string) (*dto.LogisticsInfo, error) {
	l, err := s.repo.FindByTrackingNo(trackingNo)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrLogisticsNotFound
		}
		return nil, err
	}
	return toLogisticsInfo(l), nil
}

// List 物流列表（管理后台）
func (s *logisticsService) List(req *dto.LogisticsListRequest) (*utils.Pagination, []dto.LogisticsInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.LogisticsListOptions{
		OrderID:     req.OrderID,
		OrderNo:     req.OrderNo,
		ShopID:      req.ShopID,
		UserID:      req.UserID,
		TrackingNo:  req.TrackingNo,
		CompanyCode: req.CompanyCode,
		Status:      req.Status,
		Keyword:     req.Keyword,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		RegionID:    req.RegionID,
	}
	list, total, err := s.repo.List(opts, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.LogisticsInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toLogisticsInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// ListByUser 按用户列出
func (s *logisticsService) ListByUser(userID uint, page, pageSize int) (*utils.Pagination, []dto.LogisticsInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByUser(userID, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.LogisticsInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toLogisticsInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// ListByShop 按店铺列出
func (s *logisticsService) ListByShop(shopID uint, page, pageSize int) (*utils.Pagination, []dto.LogisticsInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByShop(shopID, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.LogisticsInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toLogisticsInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// UpdateStatus 更新物流状态
func (s *logisticsService) UpdateStatus(id uint, status int) error {
	fields := map[string]interface{}{}
	now := time.Now()
	switch status {
	case model.LogisticsStatusShipped:
		fields["shipped_at"] = &now
	case model.LogisticsStatusInTransit:
		fields["in_transit_at"] = &now
	case model.LogisticsStatusDelivered:
		fields["delivered_at"] = &now
	case model.LogisticsStatusReceived:
		fields["received_at"] = &now
	case model.LogisticsStatusReturned:
		fields["returned_at"] = &now
	}
	return s.repo.UpdateStatus(id, status, fields)
}

// UpdateTraces 更新物流轨迹
func (s *logisticsService) UpdateTraces(id uint, traces interface{}) error {
	b, err := model.FromJSON(traces)
	if err != nil {
		return err
	}
	return s.repo.UpdateTraces(id, b)
}

// UpdateStatusByCallback 物流回调更新状态
func (s *logisticsService) UpdateStatusByCallback(req *dto.UpdateLogisticsStatusRequest) error {
	var l *model.Logistics
	var err error
	if req.TrackingNo != "" {
		l, err = s.repo.FindByTrackingNo(req.TrackingNo)
	} else if req.OrderID > 0 {
		l, err = s.repo.FindByOrderID(req.OrderID)
	} else {
		return errors.New("tracking_no 或 order_id 必填其一")
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrLogisticsNotFound
		}
		return err
	}

	fields := map[string]interface{}{}
	now := time.Now()
	switch req.Status {
	case model.LogisticsStatusShipped:
		fields["shipped_at"] = &now
	case model.LogisticsStatusInTransit:
		fields["in_transit_at"] = &now
	case model.LogisticsStatusDelivered:
		fields["delivered_at"] = &now
	case model.LogisticsStatusReceived:
		fields["received_at"] = &now
	case model.LogisticsStatusReturned:
		fields["returned_at"] = &now
	}
	if len(req.Traces) > 0 {
		if b, err := model.FromJSON(req.Traces); err == nil {
			fields["traces"] = b
		}
	}
	return s.repo.UpdateStatus(l.ID, req.Status, fields)
}
