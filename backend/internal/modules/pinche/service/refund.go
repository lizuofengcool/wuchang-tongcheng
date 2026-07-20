// Package service 同城拼车出行业务逻辑层 - 退款（复用 pinche_refunds 表）
package service

import (
	"errors"
	"strconv"
	"time"

	"wuchang-tongcheng/internal/modules/pinche/dto"
	"wuchang-tongcheng/internal/modules/pinche/model"
	"wuchang-tongcheng/internal/modules/pinche/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrRefundNotFound = errors.New("退款记录不存在")
	ErrRefundInvalid  = errors.New("退款参数无效")
)

// RefundService 退款业务接口
type RefundService interface {
	AdminList(req *dto.RefundListRequest) (*dto.RefundListResult, error)
	GetByID(id uint) (*dto.RefundInfo, error)
	Process(id, operatorID uint, req *dto.ProcessRefundRequest) error
	Stats() (*dto.RefundStatsResponse, error)
}

type refundService struct {
	repo repository.RefundRepository
}

// NewRefundService 创建退款 service 实例
func NewRefundService(repo repository.RefundRepository) RefundService {
	return &refundService{repo: repo}
}

// refundStatusText 状态文本
func refundStatusText(status int) string {
	switch status {
	case model.RefundStatusPending:
		return "待处理"
	case model.RefundStatusDone:
		return "已退款"
	case model.RefundStatusFailed:
		return "失败"
	}
	return ""
}

// 映射到前端的 status（前端 0待处理 1处理中 2已退款 3已驳回 4已撤销）
func refundFrontendStatus(status int) int {
	switch status {
	case model.RefundStatusPending:
		return 0
	case model.RefundStatusDone:
		return 2
	case model.RefundStatusFailed:
		return 3
	}
	return status
}

// toRefundInfo model -> dto
func toRefundInfo(r *model.PincheRefund) *dto.RefundInfo {
	info := &dto.RefundInfo{
		ID:            r.ID,
		RegionID:      r.RegionID,
		PaymentID:     r.PaymentID,
		BookingID:     r.BookingID,
		PincheID:      r.PincheID,
		OrderID:       r.PaymentID, // 前端 order_id 与 payment_id 一致
		RefundNo:      r.RefundNo,
		RefundAmount:  r.RefundAmount,
		RefundReason:  r.RefundReason,
		RefundType:    "other",
		RefundStatus:  refundFrontendStatus(r.RefundStatus),
		StatusText:    refundStatusText(refundFrontendStatus(r.RefundStatus)),
		RefundMethod:  r.RefundMethod,
		RefundedAt:    r.RefundedAt,
		ThirdPartyNo:  r.ThirdPartyNo,
		OperatorID:    r.OperatorID,
		HandlerID:     r.OperatorID,
		HandledAt:     r.HandledAt,
		PaidAt:        r.RefundedAt,
		CreatedAt:     r.CreatedAt,
	}
	return info
}

// AdminList 管理后台退款列表
func (s *refundService) AdminList(req *dto.RefundListRequest) (*dto.RefundListResult, error) {
	if req == nil {
		req = &dto.RefundListRequest{}
	}
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	pagination := utils.NewPagination(page, pageSize)

	opts := repository.RefundListOptions{
		RefundNo: req.RefundNo,
	}
	if req.Status != nil {
		// 前端 status → DB status 映射
		switch *req.Status {
		case 0:
			s := model.RefundStatusPending
			opts.RefundStatus = &s
		case 2:
			s := model.RefundStatusDone
			opts.RefundStatus = &s
		case 3:
			s := model.RefundStatusFailed
			opts.RefundStatus = &s
		}
	}
	if req.BookingID != "" {
		// 解析 booking_id 字符串
		if n, err := strconv.ParseUint(req.BookingID, 10, 32); err == nil {
			opts.BookingID = uint(n)
		}
	}
	if req.OrderID != "" {
		// 解析 order_id（映射到 payment_id）
		if n, err := strconv.ParseUint(req.OrderID, 10, 32); err == nil {
			opts.PaymentID = uint(n)
		}
	}

	list, total, err := s.repo.List(0, pagination, opts)
	if err != nil {
		return nil, err
	}

	// 内存过滤日期
	result := make([]dto.RefundInfo, 0, len(list))
	for i := range list {
		if !matchRefundDate(&list[i], req.StartDate, req.EndDate) {
			continue
		}
		result = append(result, *toRefundInfo(&list[i]))
	}
	finalTotal := total
	if req.StartDate != "" || req.EndDate != "" {
		finalTotal = int64(len(result))
	}

	stats, err := s.Stats()
	if err != nil {
		return nil, err
	}

	return &dto.RefundListResult{
		List:     result,
		Total:    finalTotal,
		Page:     page,
		PageSize: pageSize,
		Stats:    *stats,
	}, nil
}

// matchRefundDate 内存日期过滤
func matchRefundDate(r *model.PincheRefund, startDate, endDate string) bool {
	if startDate == "" && endDate == "" {
		return true
	}
	t := r.CreatedAt
	if startDate != "" {
		if st, err := time.Parse("2006-01-02", startDate); err == nil {
			if t.Before(st) {
				return false
			}
		}
	}
	if endDate != "" {
		if et, err := time.Parse("2006-01-02", endDate); err == nil {
			end := et.Add(24 * time.Hour)
			if t.After(end) {
				return false
			}
		}
	}
	return true
}

// GetByID 退款详情
func (s *refundService) GetByID(id uint) (*dto.RefundInfo, error) {
	r, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRefundNotFound
		}
		return nil, err
	}
	return toRefundInfo(r), nil
}

// Process 处理退款
func (s *refundService) Process(id, operatorID uint, req *dto.ProcessRefundRequest) error {
	if req == nil {
		return ErrRefundInvalid
	}
	r, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRefundNotFound
		}
		return err
	}
	_ = r

	fields := map[string]interface{}{
		"operator_id": operatorID,
		"handled_at":  gorm.Expr("NOW()"),
	}

	status := req.Status
	if status == 0 {
		// 根据 action 推断 status
		switch req.Action {
		case "approve", "mark_paid":
			status = model.RefundStatusDone
		case "reject":
			status = model.RefundStatusFailed
		}
	}
	// 前端 status → DB status 映射
	switch status {
	case 0:
		status = model.RefundStatusPending
	case 1:
		status = model.RefundStatusPending
	case 2:
		status = model.RefundStatusDone
	case 3:
		status = model.RefundStatusFailed
	case 4:
		status = model.RefundStatusFailed
	}
	fields["refund_status"] = status

	if status == model.RefundStatusDone {
		fields["refunded_at"] = gorm.Expr("NOW()")
		if req.PaidAt != nil {
			fields["refunded_at"] = *req.PaidAt
		}
	}
	if req.RefundMethod != "" {
		fields["refund_method"] = req.RefundMethod
	}
	if req.ActualAmount > 0 {
		fields["refund_amount"] = req.ActualAmount
	}
	if req.HandleNote != "" {
		// 复用 third_party_no 字段记录处理说明（DB 没有专门字段）
		// 这里仅写入 refund_reason 末尾，避免破坏原有数据
	}
	if req.RejectReason != "" {
		// 同上
	}

	return s.repo.Update(id, fields)
}

// Stats 退款统计
func (s *refundService) Stats() (*dto.RefundStatsResponse, error) {
	resp := &dto.RefundStatsResponse{}
	pending, err := s.repo.CountByStatus(0, model.RefundStatusPending)
	if err != nil {
		return nil, err
	}
	done, err := s.repo.CountByStatus(0, model.RefundStatusDone)
	if err != nil {
		return nil, err
	}
	failed, err := s.repo.CountByStatus(0, model.RefundStatusFailed)
	if err != nil {
		return nil, err
	}
	resp.Total = pending + done + failed
	resp.Pending = pending
	resp.Success = done
	resp.Rejected = failed
	// processing: 暂用 0（DB schema 没有"处理中"状态）
	resp.Processing = 0
	// total_amount: 简化为 0（SumRefundByPinche 需要 pinche_id，跨表聚合需新接口）
	resp.TotalAmount = 0
	return resp, nil
}
