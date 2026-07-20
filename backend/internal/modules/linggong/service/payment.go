// Package service 同城零工兼职业务逻辑层 - 薪资支付
// 对标兼职猫日结 + 支付中台：薪资日结 + T+0/T+1/T+7 多种结算方式
// 4 维数据隔离（region_id + user_id）
package service

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"wuchang-tongcheng/internal/modules/linggong/dto"
	"wuchang-tongcheng/internal/modules/linggong/model"
	"wuchang-tongcheng/internal/modules/linggong/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrPaymentNotFound      = errors.New("支付记录不存在")
	ErrPaymentNoPermission  = errors.New("无权操作此支付记录")
	ErrPaymentStatusInvalid = errors.New("支付状态不允许此操作")
)

// PaymentService 薪资支付业务接口
type PaymentService interface {
	// C 端
	Create(regionID uint, userID uint, req *dto.CreatePaymentRequest) (*dto.PaymentInfo, error)
	Update(id uint, operatorID uint, req *dto.UpdatePaymentRequest) error
	GetByID(id uint) (*dto.PaymentInfo, error)
	GetByPaymentNo(no string) (*dto.PaymentInfo, error)
	List(regionID uint, req *dto.PaymentListRequest) (*utils.Pagination, []dto.PaymentInfo, error)
	ListByLinggong(linggongID uint, page, pageSize int) (*utils.Pagination, []dto.PaymentInfo, error)
	ListByEmployer(employerID uint, page, pageSize int) (*utils.Pagination, []dto.PaymentInfo, error)
	ListByWorker(workerID uint, page, pageSize int) (*utils.Pagination, []dto.PaymentInfo, error)

	// 状态变更/结算
	UpdateStatus(id uint, req *dto.PaymentStatusUpdateRequest) error
	Settle(id uint, req *dto.PaymentSettleRequest) error

	// M 端管理
	AdminList(req *dto.PaymentAdminListRequest) (*utils.Pagination, []dto.PaymentInfo, error)
}

type paymentService struct {
	repo repository.PaymentRepository
}

// NewPaymentService 创建薪资支付 service 实例
func NewPaymentService(repo repository.PaymentRepository) PaymentService {
	return &paymentService{repo: repo}
}

// paymentStatusText 支付状态文本
func paymentStatusText(s int) string {
	switch s {
	case model.PaymentStatusPending:
		return "待支付"
	case model.PaymentStatusProcessing:
		return "处理中"
	case model.PaymentStatusPaid:
		return "已支付"
	case model.PaymentStatusFailed:
		return "支付失败"
	case model.PaymentStatusRefunded:
		return "已退款"
	case model.PaymentStatusDisputed:
		return "争议中"
	case model.PaymentStatusCanceled:
		return "已取消"
	case model.PaymentStatusPartial:
		return "部分支付"
	}
	return ""
}

// paymentMethodText 支付方式文本
func paymentMethodText(m string) string {
	switch m {
	case model.PaymentMethodWechat:
		return "微信支付"
	case model.PaymentMethodAlipay:
		return "支付宝"
	case model.PaymentMethodBank:
		return "银行卡"
	case model.PaymentMethodCash:
		return "现金"
	case model.PaymentMethodBalance:
		return "余额支付"
	case model.PaymentMethodEscrow:
		return "担保支付"
	}
	return ""
}

// paymentTypeText 支付类型文本
func paymentTypeText(t string) string {
	switch t {
	case model.PaymentTypeSalary:
		return "工资"
	case model.PaymentTypeBonus:
		return "奖金"
	case model.PaymentTypeOvertime:
		return "加班费"
	case model.PaymentTypeAllowance:
		return "补贴"
	case model.PaymentTypeReimburse:
		return "报销"
	case model.PaymentTypePenalty:
		return "罚款"
	case model.PaymentTypeRefund:
		return "退款"
	case model.PaymentTypeDeposit:
		return "押金"
	}
	return ""
}

// paymentSettlementStatusText 结算状态文本
func paymentSettlementStatusText(s int) string {
	switch s {
	case model.SettlementStatusPending:
		return "待结算"
	case model.SettlementStatusSettling:
		return "结算中"
	case model.SettlementStatusSettled:
		return "已结算"
	case model.SettlementStatusFailed:
		return "结算失败"
	}
	return ""
}

// toPaymentInfo model -> dto
func toPaymentInfo(p *model.LinggongPayment) *dto.PaymentInfo {
	return &dto.PaymentInfo{
		ID:                   p.ID,
		PaymentNo:            p.PaymentNo,
		LinggongID:           p.LinggongID,
		TaskID:               p.TaskID,
		ApplicationID:        p.ApplicationID,
		ContractID:           p.ContractID,
		EmployerID:           p.EmployerID,
		EmployerName:         p.EmployerName,
		WorkerID:             p.WorkerID,
		WorkerName:           p.WorkerName,
		WorkerPhone:          p.WorkerPhone,
		WorkerBankAccount:    p.WorkerBankAccount,
		WorkerAlipay:         p.WorkerAlipay,
		WorkerWechat:         p.WorkerWechat,
		PaymentType:          p.PaymentType,
		PaymentTypeText:      paymentTypeText(p.PaymentType),
		Amount:               p.Amount,
		WorkHours:            p.WorkHours,
		WorkDays:             p.WorkDays,
		TaskCount:            p.TaskCount,
		UnitPrice:            p.UnitPrice,
		Quantity:             p.Quantity,
		Settlement:           p.Settlement,
		SettlementText:       settlementText(p.Settlement),
		SettlementStatus:     p.SettlementStatus,
		SettlementStatusText: paymentSettlementStatusText(p.SettlementStatus),
		SettlementAt:         p.SettlementAt,
		DueAt:                p.DueAt,
		PayMethod:            p.PayMethod,
		PayMethodText:        paymentMethodText(p.PayMethod),
		PayTradeNo:           p.PayTradeNo,
		PayChannel:           p.PayChannel,
		PayeeName:            p.PayeeName,
		PlatformFee:          p.PlatformFee,
		TaxAmount:            p.TaxAmount,
		ActualAmount:         p.ActualAmount,
		Status:               p.Status,
		StatusText:           paymentStatusText(p.Status),
		FailedReason:         p.FailedReason,
		PaidAt:               p.PaidAt,
		ConfirmedAt:          p.ConfirmedAt,
		CanceledAt:           p.CanceledAt,
		WorkStartDate:        p.WorkStartDate,
		WorkEndDate:          p.WorkEndDate,
		EvidenceImages:       p.EvidenceImages,
		Remark:               p.Remark,
		InvoiceURL:           p.InvoiceURL,
		RegionID:             p.RegionID,
		CreatedAt:            p.CreatedAt,
		UpdatedAt:            p.UpdatedAt,
	}
}

// genPaymentNo 生成支付单号：LP + yyyyMMddHHmmss + 6 位随机
func genPaymentNo() string {
	return fmt.Sprintf("LP%s%06d", time.Now().Format("20060102150405"), rand.Intn(1000000))
}

// calcDueAt 计算应结时间（基于结算周期）
func calcDueAt(settlement string) *time.Time {
	now := time.Now()
	var due time.Time
	switch settlement {
	case model.SettlementT0:
		due = now
	case model.SettlementT1:
		due = now.Add(24 * time.Hour)
	case model.SettlementT3:
		due = now.Add(3 * 24 * time.Hour)
	case model.SettlementT7:
		due = now.Add(7 * 24 * time.Hour)
	case model.SettlementM1:
		due = now.AddDate(0, 1, 0)
	case model.SettlementProject:
		due = now.Add(30 * 24 * time.Hour)
	default:
		due = now.Add(24 * time.Hour)
	}
	return &due
}

// ===== C 端 =====

// Create 创建支付记录
func (s *paymentService) Create(regionID uint, userID uint, req *dto.CreatePaymentRequest) (*dto.PaymentInfo, error) {
	p := &model.LinggongPayment{
		PaymentNo:         genPaymentNo(),
		LinggongID:        req.LinggongID,
		TaskID:            req.TaskID,
		ApplicationID:     req.ApplicationID,
		ContractID:        req.ContractID,
		EmployerID:        req.EmployerID,
		EmployerName:      req.EmployerName,
		WorkerID:          req.WorkerID,
		WorkerName:        req.WorkerName,
		WorkerPhone:       req.WorkerPhone,
		WorkerBankAccount: req.WorkerBankAccount,
		WorkerAlipay:      req.WorkerAlipay,
		WorkerWechat:      req.WorkerWechat,
		PaymentType:       req.PaymentType,
		Amount:            req.Amount,
		WorkHours:         req.WorkHours,
		WorkDays:          req.WorkDays,
		TaskCount:         req.TaskCount,
		UnitPrice:         req.UnitPrice,
		Quantity:          req.Quantity,
		Settlement:        req.Settlement,
		SettlementStatus:  model.SettlementStatusPending,
		PayMethod:         req.PayMethod,
		PayChannel:        req.PayChannel,
		PayeeName:         req.PayeeName,
		WorkStartDate:     req.WorkStartDate,
		WorkEndDate:       req.WorkEndDate,
		Remark:            req.Remark,
		InvoiceURL:        req.InvoiceURL,
		Status:            model.PaymentStatusPending,
	}
	p.RegionID = regionID

	if p.PaymentType == "" {
		p.PaymentType = model.PaymentTypeSalary
	}
	if p.Settlement == "" {
		p.Settlement = model.SettlementT1
	}
	if p.PayMethod == "" {
		p.PayMethod = model.PaymentMethodWechat
	}

	p.DueAt = calcDueAt(p.Settlement)
	p.ActualAmount = p.Amount - p.PlatformFee - p.TaxAmount

	if req.EvidenceImages != nil {
		if jb, err := model.FromJSON(req.EvidenceImages); err == nil {
			p.EvidenceImages = jb
		}
	}

	_ = userID

	if err := s.repo.Create(p); err != nil {
		return nil, err
	}
	return toPaymentInfo(p), nil
}

// Update 更新支付记录（仅待支付状态可改）
func (s *paymentService) Update(id uint, operatorID uint, req *dto.UpdatePaymentRequest) error {
	p, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPaymentNotFound
		}
		return err
	}
	if p.Status != model.PaymentStatusPending {
		return ErrPaymentStatusInvalid
	}

	fields := map[string]interface{}{}
	if req.Amount != nil {
		fields["amount"] = *req.Amount
		fields["actual_amount"] = *req.Amount - p.PlatformFee - p.TaxAmount
	}
	if req.WorkHours != nil {
		fields["work_hours"] = *req.WorkHours
	}
	if req.WorkDays != nil {
		fields["work_days"] = *req.WorkDays
	}
	if req.TaskCount != nil {
		fields["task_count"] = *req.TaskCount
	}
	if req.UnitPrice != nil {
		fields["unit_price"] = *req.UnitPrice
	}
	if req.Quantity != nil {
		fields["quantity"] = *req.Quantity
	}
	if req.Settlement != nil {
		fields["settlement"] = *req.Settlement
		fields["due_at"] = calcDueAt(*req.Settlement)
	}
	if req.PayMethod != nil {
		fields["pay_method"] = *req.PayMethod
	}
	if req.PayeeName != nil {
		fields["payee_name"] = *req.PayeeName
	}
	if req.Remark != nil {
		fields["remark"] = *req.Remark
	}
	if req.InvoiceURL != nil {
		fields["invoice_url"] = *req.InvoiceURL
	}
	if req.EvidenceImages != nil {
		if jb, err := model.FromJSON(req.EvidenceImages); err == nil {
			fields["evidence_images"] = jb
		}
	}

	_ = operatorID

	if len(fields) == 0 {
		return nil
	}
	return s.repo.Update(id, fields)
}

// GetByID 获取支付详情
func (s *paymentService) GetByID(id uint) (*dto.PaymentInfo, error) {
	p, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPaymentNotFound
		}
		return nil, err
	}
	return toPaymentInfo(p), nil
}

// GetByPaymentNo 按支付单号查询
func (s *paymentService) GetByPaymentNo(no string) (*dto.PaymentInfo, error) {
	p, err := s.repo.FindByPaymentNo(no)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPaymentNotFound
		}
		return nil, err
	}
	return toPaymentInfo(p), nil
}

// List C 端支付列表
func (s *paymentService) List(regionID uint, req *dto.PaymentListRequest) (*utils.Pagination, []dto.PaymentInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.PaymentListOptions{
		LinggongID:       req.LinggongID,
		TaskID:           req.TaskID,
		ApplicationID:    req.ApplicationID,
		EmployerID:       req.EmployerID,
		WorkerID:         req.WorkerID,
		PaymentType:      req.PaymentType,
		Settlement:       req.Settlement,
		Status:           req.Status,
		SettlementStatus: req.SettlementStatus,
		Keyword:          req.Keyword,
	}
	list, total, err := s.repo.List(regionID, pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.PaymentInfo, 0, len(list))
	for i := range list {
		result = append(result, *toPaymentInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListByLinggong 按岗位反查
func (s *paymentService) ListByLinggong(linggongID uint, page, pageSize int) (*utils.Pagination, []dto.PaymentInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByLinggong(linggongID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.PaymentInfo, 0, len(list))
	for i := range list {
		result = append(result, *toPaymentInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListByEmployer 按雇主反查
func (s *paymentService) ListByEmployer(employerID uint, page, pageSize int) (*utils.Pagination, []dto.PaymentInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByEmployer(employerID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.PaymentInfo, 0, len(list))
	for i := range list {
		result = append(result, *toPaymentInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListByWorker 按求职者反查
func (s *paymentService) ListByWorker(workerID uint, page, pageSize int) (*utils.Pagination, []dto.PaymentInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByWorker(workerID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.PaymentInfo, 0, len(list))
	for i := range list {
		result = append(result, *toPaymentInfo(&list[i]))
	}
	return pagination, result, nil
}

// UpdateStatus 更新支付状态（支付回调/手动确认）
func (s *paymentService) UpdateStatus(id uint, req *dto.PaymentStatusUpdateRequest) error {
	p, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPaymentNotFound
		}
		return err
	}

	if p.Status == model.PaymentStatusCanceled || p.Status == model.PaymentStatusRefunded {
		return ErrPaymentStatusInvalid
	}

	fields := map[string]interface{}{
		"status":        req.Status,
		"failed_reason": req.FailedReason,
	}
	if req.PayTradeNo != "" {
		fields["pay_trade_no"] = req.PayTradeNo
	}

	now := time.Now()
	switch req.Status {
	case model.PaymentStatusPaid:
		fields["paid_at"] = &now
	case model.PaymentStatusCanceled:
		fields["canceled_at"] = &now
	}

	return s.repo.Update(id, fields)
}

// Settle 结算薪资
func (s *paymentService) Settle(id uint, req *dto.PaymentSettleRequest) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPaymentNotFound
		}
		return err
	}

	fields := map[string]interface{}{
		"settlement_status": req.SettlementStatus,
	}
	now := time.Now()
	if req.SettlementStatus == model.SettlementStatusSettled {
		fields["settlement_at"] = &now
	}

	return s.repo.Update(id, fields)
}

// ===== M 端管理 =====

// AdminList M 端支付列表（跨地区）
func (s *paymentService) AdminList(req *dto.PaymentAdminListRequest) (*utils.Pagination, []dto.PaymentInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.PaymentAdminListOptions{
		RegionID:    req.RegionID,
		LinggongID:  req.LinggongID,
		EmployerID:  req.EmployerID,
		WorkerID:    req.WorkerID,
		PaymentType: req.PaymentType,
		Status:      req.Status,
		Keyword:     req.Keyword,
	}
	list, total, err := s.repo.AdminList(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.PaymentInfo, 0, len(list))
	for i := range list {
		result = append(result, *toPaymentInfo(&list[i]))
	}
	return pagination, result, nil
}
