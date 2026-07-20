// Package service 同城零工兼职业务逻辑层 - 提现
// 对接支付中台：求职者/雇主提现 + 提现状态机 + 多种到账方式
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
	ErrWithdrawalNotFound      = errors.New("提现记录不存在")
	ErrWithdrawalNoPermission  = errors.New("无权操作此提现记录")
	ErrWithdrawalStatusInvalid = errors.New("提现状态不允许此操作")
	ErrWithdrawalAmountInvalid = errors.New("提现金额无效")
)

// WithdrawalService 提现业务接口
type WithdrawalService interface {
	// C 端
	Create(regionID uint, userID uint, userName string, userPhone string, req *dto.CreateWithdrawalRequest) (*dto.WithdrawalInfo, error)
	GetByID(id uint) (*dto.WithdrawalInfo, error)
	GetByWithdrawalNo(no string) (*dto.WithdrawalInfo, error)
	List(regionID uint, req *dto.WithdrawalListRequest) (*utils.Pagination, []dto.WithdrawalInfo, error)
	ListByUser(userID uint, page, pageSize int) (*utils.Pagination, []dto.WithdrawalInfo, error)

	// M 端管理
	Audit(id uint, reviewerID uint, reviewerName string, req *dto.WithdrawalAuditRequest) error
	UpdateStatus(id uint, status int, failedReason string) error
	AdminList(req *dto.WithdrawalListRequest) (*utils.Pagination, []dto.WithdrawalInfo, error)
}

type withdrawalService struct {
	repo repository.WithdrawalRepository
}

// NewWithdrawalService 创建提现 service 实例
func NewWithdrawalService(repo repository.WithdrawalRepository) WithdrawalService {
	return &withdrawalService{repo: repo}
}

// withdrawalStatusText 提现状态文本
func withdrawalStatusText(s int) string {
	switch s {
	case model.WithdrawalStatusPending:
		return "待审核"
	case model.WithdrawalStatusApproved:
		return "已审核"
	case model.WithdrawalStatusProcessing:
		return "处理中"
	case model.WithdrawalStatusSucceeded:
		return "已到账"
	case model.WithdrawalStatusFailed:
		return "失败"
	case model.WithdrawalStatusRejected:
		return "已驳回"
	case model.WithdrawalStatusCanceled:
		return "已取消"
	}
	return ""
}

// withdrawalMethodText 提现方式文本
func withdrawalMethodText(m string) string {
	switch m {
	case model.WithdrawalMethodWechat:
		return "微信"
	case model.WithdrawalMethodAlipay:
		return "支付宝"
	case model.WithdrawalMethodBank:
		return "银行卡"
	case model.WithdrawalMethodBalance:
		return "余额"
	}
	return ""
}

// withdrawalUserTypeText 用户类型文本
func withdrawalUserTypeText(t string) string {
	switch t {
	case model.WithdrawalUserTypeWorker:
		return "求职者"
	case model.WithdrawalUserTypeEmployer:
		return "雇主"
	}
	return ""
}

// toWithdrawalInfo model -> dto
func toWithdrawalInfo(w *model.LinggongWithdrawal) *dto.WithdrawalInfo {
	return &dto.WithdrawalInfo{
		ID:               w.ID,
		WithdrawalNo:     w.WithdrawalNo,
		UserID:           w.UserID,
		UserType:         w.UserType,
		UserTypeText:     withdrawalUserTypeText(w.UserType),
		UserName:         w.UserName,
		UserPhone:        w.UserPhone,
		Amount:           w.Amount,
		Fee:              w.Fee,
		Tax:              w.Tax,
		ActualAmount:     w.ActualAmount,
		BalanceBefore:    w.BalanceBefore,
		BalanceAfter:     w.BalanceAfter,
		Method:           w.Method,
		MethodText:       withdrawalMethodText(w.Method),
		PayeeName:        w.PayeeName,
		PayeeAccount:     w.PayeeAccount,
		PayeeBank:        w.PayeeBank,
		PayeeBankBranch:  w.PayeeBankBranch,
		BankCardNo:       w.BankCardNo,
		AlipayAccount:    w.AlipayAccount,
		WechatAccount:    w.WechatAccount,
		Status:           w.Status,
		StatusText:       withdrawalStatusText(w.Status),
		FailedReason:     w.FailedReason,
		ReviewedBy:       w.ReviewedBy,
		ReviewedByName:   w.ReviewedByName,
		ReviewedAt:       w.ReviewedAt,
		ReviewedRemark:   w.ReviewedRemark,
		PayTradeNo:       w.PayTradeNo,
		PayChannel:       w.PayChannel,
		ProcessedAt:      w.ProcessedAt,
		SucceededAt:      w.SucceededAt,
		CanceledAt:       w.CanceledAt,
		EstimatedArrival: w.EstimatedArrival,
		Remark:           w.Remark,
		RegionID:         w.RegionID,
		CreatedAt:        w.CreatedAt,
		UpdatedAt:        w.UpdatedAt,
	}
}

// genWithdrawalNo 生成提现单号：LW + yyyyMMddHHmmss + 6 位随机
func genWithdrawalNo() string {
	return fmt.Sprintf("LW%s%06d", time.Now().Format("20060102150405"), rand.Intn(1000000))
}

// calcWithdrawalFee 计算手续费（MVP 简化：固定 1% 手续费，最低 0.1，无税费）
func calcWithdrawalFee(amount float64) (fee, tax, actual float64) {
	if amount <= 0 {
		return 0, 0, 0
	}
	fee = amount * 0.01
	if fee < 0.1 {
		fee = 0.1
	}
	tax = 0
	actual = amount - fee - tax
	return
}

// ===== C 端 =====

// Create 创建提现申请
func (s *withdrawalService) Create(regionID uint, userID uint, userName string, userPhone string, req *dto.CreateWithdrawalRequest) (*dto.WithdrawalInfo, error) {
	if req.Amount <= 0 {
		return nil, ErrWithdrawalAmountInvalid
	}

	userType := req.UserType
	if userType == "" {
		userType = model.WithdrawalUserTypeWorker
	}
	method := req.Method
	if method == "" {
		method = model.WithdrawalMethodWechat
	}

	fee, tax, actual := calcWithdrawalFee(req.Amount)
	// 预计到账时间：微信/支付宝/余额 1 小时内，银行卡 T+1
	now := time.Now()
	var estimated time.Time
	switch method {
	case model.WithdrawalMethodBank:
		estimated = now.Add(24 * time.Hour)
	default:
		estimated = now.Add(1 * time.Hour)
	}

	w := &model.LinggongWithdrawal{
		WithdrawalNo:     genWithdrawalNo(),
		UserID:           userID,
		UserType:         userType,
		UserName:         userName,
		UserPhone:        userPhone,
		Amount:           req.Amount,
		Fee:              fee,
		Tax:              tax,
		ActualAmount:     actual,
		Method:           method,
		PayeeName:        req.PayeeName,
		PayeeAccount:     req.PayeeAccount,
		PayeeBank:        req.PayeeBank,
		PayeeBankBranch:  req.PayeeBankBranch,
		BankCardNo:       req.BankCardNo,
		AlipayAccount:    req.AlipayAccount,
		WechatAccount:    req.WechatAccount,
		Status:           model.WithdrawalStatusPending,
		EstimatedArrival: &estimated,
		Remark:           req.Remark,
	}
	w.RegionID = regionID

	if err := s.repo.Create(w); err != nil {
		return nil, err
	}
	return toWithdrawalInfo(w), nil
}

// GetByID 获取提现详情
func (s *withdrawalService) GetByID(id uint) (*dto.WithdrawalInfo, error) {
	w, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWithdrawalNotFound
		}
		return nil, err
	}
	return toWithdrawalInfo(w), nil
}

// GetByWithdrawalNo 按提现单号查询
func (s *withdrawalService) GetByWithdrawalNo(no string) (*dto.WithdrawalInfo, error) {
	w, err := s.repo.FindByWithdrawalNo(no)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWithdrawalNotFound
		}
		return nil, err
	}
	return toWithdrawalInfo(w), nil
}

// List C 端提现列表
func (s *withdrawalService) List(regionID uint, req *dto.WithdrawalListRequest) (*utils.Pagination, []dto.WithdrawalInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.WithdrawalListOptions{
		UserID:   req.UserID,
		UserType: req.UserType,
		Status:   req.Status,
		Method:   req.Method,
		Keyword:  req.Keyword,
	}
	list, total, err := s.repo.List(regionID, pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.WithdrawalInfo, 0, len(list))
	for i := range list {
		result = append(result, *toWithdrawalInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListByUser 按用户反查
func (s *withdrawalService) ListByUser(userID uint, page, pageSize int) (*utils.Pagination, []dto.WithdrawalInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByUser(userID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.WithdrawalInfo, 0, len(list))
	for i := range list {
		result = append(result, *toWithdrawalInfo(&list[i]))
	}
	return pagination, result, nil
}

// ===== M 端管理 =====

// Audit 审核提现（通过 → Approved / 驳回 → Rejected）
func (s *withdrawalService) Audit(id uint, reviewerID uint, reviewerName string, req *dto.WithdrawalAuditRequest) error {
	w, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrWithdrawalNotFound
		}
		return err
	}
	if w.Status != model.WithdrawalStatusPending {
		return ErrWithdrawalStatusInvalid
	}

	now := time.Now()
	fields := map[string]interface{}{
		"status":           req.Status,
		"reviewed_by":      reviewerID,
		"reviewed_by_name": reviewerName,
		"reviewed_at":      &now,
		"reviewed_remark":  req.ReviewedRemark,
	}
	if req.Status == model.WithdrawalStatusRejected {
		fields["failed_reason"] = req.FailedReason
	}

	return s.repo.Update(id, fields)
}

// UpdateStatus 更新提现状态（处理中/已到账/失败/取消）
func (s *withdrawalService) UpdateStatus(id uint, status int, failedReason string) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrWithdrawalNotFound
		}
		return err
	}
	now := time.Now()
	fields := map[string]interface{}{
		"status": status,
	}
	switch status {
	case model.WithdrawalStatusProcessing:
		fields["processed_at"] = &now
	case model.WithdrawalStatusSucceeded:
		fields["succeeded_at"] = &now
	case model.WithdrawalStatusFailed:
		fields["failed_reason"] = failedReason
		fields["processed_at"] = &now
	case model.WithdrawalStatusCanceled:
		fields["canceled_at"] = &now
	}
	return s.repo.Update(id, fields)
}

// AdminList M 端提现列表（跨地区）
func (s *withdrawalService) AdminList(req *dto.WithdrawalListRequest) (*utils.Pagination, []dto.WithdrawalInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.WithdrawalListOptions{
		UserID:   req.UserID,
		UserType: req.UserType,
		Status:   req.Status,
		Method:   req.Method,
		Keyword:  req.Keyword,
	}
	// 跨地区：regionID=0
	list, total, err := s.repo.List(0, pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.WithdrawalInfo, 0, len(list))
	for i := range list {
		result = append(result, *toWithdrawalInfo(&list[i]))
	}
	return pagination, result, nil
}
