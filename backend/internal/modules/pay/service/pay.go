// Package service 支付财务中台精简版业务逻辑层
// 依据 ershou 模块依赖：担保交易 + 订单 + 退款 + 提现 + 结算 + 资金账户
// 暴露 PayService 接口供其他模块直接 import 调用（不通过 HTTP）
package service

import (
	"errors"
	"fmt"
	"time"

	"wuchang-tongcheng/internal/modules/pay/dto"
	"wuchang-tongcheng/internal/modules/pay/model"
	"wuchang-tongcheng/internal/modules/pay/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrOrderNotFound       = errors.New("支付订单不存在")
	ErrOrderAlreadyPaid    = errors.New("订单已支付")
	ErrOrderClosed         = errors.New("订单已关闭")
	ErrInsufficientBalance = errors.New("余额不足")
	ErrEscrowNotFound      = errors.New("担保交易不存在")
	ErrEscrowNotFrozen     = errors.New("担保交易非冻结状态")
	ErrRefundAmountExceed  = errors.New("退款金额超过订单金额")
	ErrRefundNotFound      = errors.New("退款单不存在")
	ErrWithdrawNotFound    = errors.New("提现单不存在")
	ErrWithdrawNotPending  = errors.New("提现单非待审核状态")
	ErrSettlementNotFound  = errors.New("结算单不存在")
	ErrAccountNotFound     = errors.New("资金账户不存在")
)

// PayService 支付中台业务接口
// 暴露给其他模块直接 import 调用，不通过 HTTP
type PayService interface {
	// 订单
	CreatePayment(regionID uint, userID uint, req *dto.CreatePaymentRequest) (*dto.PaymentOrderInfo, error)
	ConfirmPayment(req *dto.ConfirmPaymentRequest) error
	GetPayment(orderNo string) (*dto.PaymentOrderInfo, error)
	ListPayments(userID uint, page, pageSize int) ([]dto.PaymentOrderInfo, int64, error)

	// 担保交易
	ConfirmEscrow(orderNo string) error
	GetEscrow(orderNo string) (*dto.EscrowInfo, error)

	// 退款
	Refund(regionID uint, req *dto.RefundRequest) (*dto.RefundInfo, error)
	GetRefund(refundNo string) (*dto.RefundInfo, error)

	// 提现
	Withdraw(regionID uint, userID uint, req *dto.WithdrawRequest) (*dto.WithdrawInfo, error)
	ListMyWithdrawals(userID uint, page, pageSize int) ([]dto.WithdrawInfo, int64, error)
	ListPendingWithdrawals(page, pageSize int) ([]dto.WithdrawInfo, int64, error)
	HandleWithdrawal(req *dto.WithdrawActionRequest, handlerID uint) error

	// 结算
	Settle(regionID uint, req *dto.SettleRequest) (*dto.SettlementInfo, error)
	ListSettlements(merchantID uint, page, pageSize int) ([]dto.SettlementInfo, int64, error)

	// 资金账户
	GetAccount(userID uint) (*dto.AccountInfo, error)
}

type payService struct {
	repo repository.PayRepository
}

// NewPayService 创建 service 实例
func NewPayService(repo repository.PayRepository) PayService {
	return &payService{repo: repo}
}

// ===== 订单 =====

// CreatePayment 创建支付订单
func (s *payService) CreatePayment(regionID uint, userID uint, req *dto.CreatePaymentRequest) (*dto.PaymentOrderInfo, error) {
	expireSec := req.ExpireSec
	if expireSec <= 0 {
		expireSec = 1800
	}
	expireAt := time.Now().Add(time.Duration(expireSec) * time.Second)

	o := &model.PaymentOrder{
		OrderNo:   generateOrderNo("PAY"),
		UserID:    userID,
		BizModule: req.BizModule,
		BizID:     req.BizID,
		Title:     req.Title,
		Amount:    req.Amount,
		PayMethod: req.PayMethod,
		PayStatus: model.PayStatusPending,
		ExpireAt:  &expireAt,
		Extra:     defaultJSON(req.Extra),
	}
	o.RegionID = regionID

	if err := s.repo.CreateOrder(o); err != nil {
		return nil, err
	}
	return toPaymentOrderInfo(o), nil
}

// ConfirmPayment 确认支付（第三方回调或主动确认）
// 流程：订单标记为已支付 → 若为担保交易则创建 EscrowAccount 冻结资金
func (s *payService) ConfirmPayment(req *dto.ConfirmPaymentRequest) error {
	o, err := s.repo.FindOrderByNo(req.OrderNo)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrOrderNotFound
		}
		return err
	}
	if o.PayStatus == model.PayStatusPaid {
		return ErrOrderAlreadyPaid
	}
	if o.PayStatus == model.PayStatusClosed {
		return ErrOrderClosed
	}

	now := time.Now()
	fields := map[string]interface{}{
		"pay_status":     model.PayStatusPaid,
		"third_party_no": req.ThirdPartyNo,
		"pay_method":     req.PayMethod,
		"paid_at":        &now,
	}
	if err := s.repo.UpdateOrderFields(o.ID, fields); err != nil {
		return err
	}

	// 余额支付：扣减账户余额
	if req.PayMethod == model.PayMethodBalance {
		acc, err := s.repo.GetOrCreateAccount(o.UserID, o.RegionID)
		if err != nil {
			return err
		}
		if acc.Balance < o.Amount {
			return ErrInsufficientBalance
		}
		if err := s.repo.UpdateAccountFields(acc.ID, map[string]interface{}{
			"balance":       acc.Balance - o.Amount,
			"frozen_amount": acc.FrozenAmount + o.Amount,
			"total_expense": acc.TotalExpense + o.Amount,
		}); err != nil {
			return err
		}
	}

	// 创建担保账户（资金托管）
	escrow := &model.EscrowAccount{
		OrderID:       o.ID,
		OrderNo:       o.OrderNo,
		UserID:        o.UserID,
		Amount:        o.Amount,
		Status:        model.EscrowStatusFrozen,
		FrozenAt:      now,
		AutoReleaseAt: autoReleaseTime(now, 7), // 默认 7 天后自动放款
	}
	escrow.RegionID = o.RegionID
	if err := s.repo.CreateEscrow(escrow); err != nil {
		return err
	}
	return nil
}

// GetPayment 查询订单
func (s *payService) GetPayment(orderNo string) (*dto.PaymentOrderInfo, error) {
	o, err := s.repo.FindOrderByNo(orderNo)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}
	return toPaymentOrderInfo(o), nil
}

// ListPayments 用户订单列表
func (s *payService) ListPayments(userID uint, page, pageSize int) ([]dto.PaymentOrderInfo, int64, error) {
	list, total, err := s.repo.ListOrders(userID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.PaymentOrderInfo, 0, len(list))
	for i := range list {
		result = append(result, *toPaymentOrderInfo(&list[i]))
	}
	return result, total, nil
}

// ===== 担保交易 =====

// ConfirmEscrow 确认收货（放款给商家）
// 流程：担保账户标记已放款 → 卖家账户余额增加 → 累计收入增加
func (s *payService) ConfirmEscrow(orderNo string) error {
	o, err := s.repo.FindOrderByNo(orderNo)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrOrderNotFound
		}
		return err
	}
	escrow, err := s.repo.FindEscrowByOrderID(o.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrEscrowNotFound
		}
		return err
	}
	if escrow.Status != model.EscrowStatusFrozen {
		return ErrEscrowNotFrozen
	}

	now := time.Now()
	if err := s.repo.UpdateEscrowFields(escrow.ID, map[string]interface{}{
		"status":     model.EscrowStatusReleased,
		"release_at": &now,
	}); err != nil {
		return err
	}

	// 卖家账户增加余额（如果是余额支付，回退冻结金额）
	if o.PayMethod == model.PayMethodBalance {
		if acc, err := s.repo.GetOrCreateAccount(o.UserID, o.RegionID); err == nil {
			_ = s.repo.UpdateAccountFields(acc.ID, map[string]interface{}{
				"frozen_amount": acc.FrozenAmount - escrow.Amount,
			})
		}
	}
	// 商家账户累计收入
	if escrow.MerchantID > 0 {
		if mAcc, err := s.repo.GetOrCreateAccount(escrow.MerchantID, o.RegionID); err == nil {
			_ = s.repo.UpdateAccountFields(mAcc.ID, map[string]interface{}{
				"balance":      mAcc.Balance + escrow.Amount,
				"total_income": mAcc.TotalIncome + escrow.Amount,
			})
		}
	}
	return nil
}

// GetEscrow 查询担保账户
func (s *payService) GetEscrow(orderNo string) (*dto.EscrowInfo, error) {
	o, err := s.repo.FindOrderByNo(orderNo)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}
	escrow, err := s.repo.FindEscrowByOrderID(o.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEscrowNotFound
		}
		return nil, err
	}
	return &dto.EscrowInfo{
		ID:            escrow.ID,
		OrderID:       escrow.OrderID,
		OrderNo:       escrow.OrderNo,
		UserID:        escrow.UserID,
		MerchantID:    escrow.MerchantID,
		Amount:        escrow.Amount,
		Status:        escrow.Status,
		FrozenAt:      escrow.FrozenAt,
		ReleaseAt:     escrow.ReleaseAt,
		AutoReleaseAt: escrow.AutoReleaseAt,
	}, nil
}

// ===== 退款 =====

// Refund 退款（原路返回）
// 流程：创建退款单 → 担保账户标记退款 → 订单标记已退款
func (s *payService) Refund(regionID uint, req *dto.RefundRequest) (*dto.RefundInfo, error) {
	o, err := s.repo.FindOrderByNo(req.OrderNo)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}
	if req.Amount > o.Amount {
		return nil, ErrRefundAmountExceed
	}

	r := &model.RefundOrder{
		RefundNo:     generateOrderNo("RFD"),
		OrderID:      o.ID,
		OrderNo:      o.OrderNo,
		UserID:       o.UserID,
		Amount:       req.Amount,
		Reason:       req.Reason,
		Status:       model.RefundStatusRefunded,
		RefundMethod: o.PayMethod,
		ProcessedAt:  &[]time.Time{time.Now()}[0],
	}
	r.RegionID = regionID

	if err := s.repo.CreateRefund(r); err != nil {
		return nil, err
	}

	// 担保账户标记退款
	if escrow, err := s.repo.FindEscrowByOrderID(o.ID); err == nil {
		_ = s.repo.UpdateEscrowFields(escrow.ID, map[string]interface{}{
			"status": model.EscrowStatusRefunded,
		})
	}

	// 订单标记已退款
	newStatus := model.PayStatusRefunded
	if req.Amount < o.Amount {
		newStatus = model.PayStatusPartRefund
	}
	_ = s.repo.UpdateOrderFields(o.ID, map[string]interface{}{
		"pay_status": newStatus,
	})

	// 余额支付：退回买家账户
	if o.PayMethod == model.PayMethodBalance {
		if acc, err := s.repo.GetOrCreateAccount(o.UserID, o.RegionID); err == nil {
			_ = s.repo.UpdateAccountFields(acc.ID, map[string]interface{}{
				"balance":       acc.Balance + req.Amount,
				"frozen_amount": acc.FrozenAmount - req.Amount,
				"total_income":  acc.TotalIncome + req.Amount,
			})
		}
	}

	return toRefundInfo(r), nil
}

// GetRefund 查询退款单
func (s *payService) GetRefund(refundNo string) (*dto.RefundInfo, error) {
	r, err := s.repo.FindRefundByNo(refundNo)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRefundNotFound
		}
		return nil, err
	}
	return toRefundInfo(r), nil
}

// ===== 提现 =====

// Withdraw 提现申请
// 流程：扣减账户余额 → 创建提现单（待审核）
func (s *payService) Withdraw(regionID uint, userID uint, req *dto.WithdrawRequest) (*dto.WithdrawInfo, error) {
	acc, err := s.repo.GetOrCreateAccount(userID, regionID)
	if err != nil {
		return nil, err
	}
	if acc.Balance < req.Amount {
		return nil, ErrInsufficientBalance
	}

	// 手续费 0.1%（最低 0.1 元）
	fee := req.Amount * 0.001
	if fee < 0.1 {
		fee = 0.1
	}
	actual := req.Amount - fee

	// 扣减余额、增加冻结
	if err := s.repo.UpdateAccountFields(acc.ID, map[string]interface{}{
		"balance":       acc.Balance - req.Amount,
		"frozen_amount": acc.FrozenAmount + req.Amount,
	}); err != nil {
		return nil, err
	}

	w := &model.Withdrawal{
		WithdrawalNo: generateOrderNo("WD"),
		UserID:       userID,
		Amount:       req.Amount,
		Fee:          fee,
		ActualAmount: actual,
		BankCard: fmt.Sprintf(`{"card_no":"%s","bank":"%s","name":"%s","phone":"%s"}`,
			req.BankCardNo, req.BankName, req.HolderName, req.Phone),
		Status: model.WithdrawStatusPending,
	}
	w.RegionID = regionID

	if err := s.repo.CreateWithdrawal(w); err != nil {
		// 回滚账户
		_ = s.repo.UpdateAccountFields(acc.ID, map[string]interface{}{
			"balance":       acc.Balance,
			"frozen_amount": acc.FrozenAmount,
		})
		return nil, err
	}
	return toWithdrawInfo(w), nil
}

// ListMyWithdrawals 用户提现列表
func (s *payService) ListMyWithdrawals(userID uint, page, pageSize int) ([]dto.WithdrawInfo, int64, error) {
	list, total, err := s.repo.ListWithdrawals(userID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.WithdrawInfo, 0, len(list))
	for i := range list {
		result = append(result, *toWithdrawInfo(&list[i]))
	}
	return result, total, nil
}

// ListPendingWithdrawals 待审核提现列表（M 端）
func (s *payService) ListPendingWithdrawals(page, pageSize int) ([]dto.WithdrawInfo, int64, error) {
	list, total, err := s.repo.ListPendingWithdrawals(page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.WithdrawInfo, 0, len(list))
	for i := range list {
		result = append(result, *toWithdrawInfo(&list[i]))
	}
	return result, total, nil
}

// HandleWithdrawal 处理提现申请（M 端）
func (s *payService) HandleWithdrawal(req *dto.WithdrawActionRequest, handlerID uint) error {
	w, err := s.repo.FindWithdrawalByNo(req.WithdrawalNo)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrWithdrawNotFound
		}
		return err
	}
	if w.Status != model.WithdrawStatusPending && w.Status != model.WithdrawStatusApproved {
		return ErrWithdrawNotPending
	}

	now := time.Now()
	fields := map[string]interface{}{
		"processed_at": &now,
	}

	switch req.Action {
	case "approve":
		fields["status"] = model.WithdrawStatusApproved
	case "reject":
		fields["status"] = model.WithdrawStatusRejected
		fields["reject_reason"] = req.Reason
		// 退回冻结金额到余额
		if acc, err := s.repo.FindAccount(w.UserID); err == nil {
			_ = s.repo.UpdateAccountFields(acc.ID, map[string]interface{}{
				"balance":       acc.Balance + w.Amount,
				"frozen_amount": acc.FrozenAmount - w.Amount,
			})
		}
	case "paid":
		fields["status"] = model.WithdrawStatusPaid
		// 解冻扣除
		if acc, err := s.repo.FindAccount(w.UserID); err == nil {
			_ = s.repo.UpdateAccountFields(acc.ID, map[string]interface{}{
				"frozen_amount": acc.FrozenAmount - w.Amount,
				"total_expense": acc.TotalExpense + w.Amount,
			})
		}
	case "failed":
		fields["status"] = model.WithdrawStatusFailed
		fields["reject_reason"] = req.Reason
		// 退回冻结金额到余额
		if acc, err := s.repo.FindAccount(w.UserID); err == nil {
			_ = s.repo.UpdateAccountFields(acc.ID, map[string]interface{}{
				"balance":       acc.Balance + w.Amount,
				"frozen_amount": acc.FrozenAmount - w.Amount,
			})
		}
	}

	return s.repo.UpdateWithdrawalFields(w.ID, fields)
}

// ===== 结算 =====

// Settle 触发结算（精简版：扫描周期内已放款的担保交易）
func (s *payService) Settle(regionID uint, req *dto.SettleRequest) (*dto.SettlementInfo, error) {
	periodType := req.PeriodType
	if periodType == "" {
		periodType = model.PeriodT1
	}
	now := time.Now()
	var start, end time.Time
	switch periodType {
	case model.PeriodT1:
		start = now.AddDate(0, 0, -1)
		end = now
	case model.PeriodT7:
		start = now.AddDate(0, 0, -7)
		end = now
	case model.PeriodMonthly:
		start = now.AddDate(0, -1, 0)
		end = now
	}

	// 精简版：暂用占位数据，实际需扫描 pay_escrows 表统计
	s2 := &model.Settlement{
		SettlementNo:     generateOrderNo("STL"),
		MerchantID:       req.MerchantID,
		PeriodType:       periodType,
		PeriodStart:      start,
		PeriodEnd:        end,
		OrderCount:       0,
		TotalAmount:      0,
		Commission:       0,
		SettlementAmount: 0,
		Status:           model.SettleStatusDone,
		SettledAt:        &now,
	}
	s2.RegionID = regionID
	if err := s.repo.CreateSettlement(s2); err != nil {
		return nil, err
	}
	return toSettlementInfo(s2), nil
}

// ListSettlements 商家结算列表
func (s *payService) ListSettlements(merchantID uint, page, pageSize int) ([]dto.SettlementInfo, int64, error) {
	list, total, err := s.repo.ListSettlements(merchantID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.SettlementInfo, 0, len(list))
	for i := range list {
		result = append(result, *toSettlementInfo(&list[i]))
	}
	return result, total, nil
}

// ===== 资金账户 =====

// GetAccount 查询用户资金账户（不存在则自动创建）
func (s *payService) GetAccount(userID uint) (*dto.AccountInfo, error) {
	acc, err := s.repo.GetOrCreateAccount(userID, 1)
	if err != nil {
		return nil, ErrAccountNotFound
	}
	return &dto.AccountInfo{
		ID:           acc.ID,
		UserID:       acc.UserID,
		Balance:      acc.Balance,
		FrozenAmount: acc.FrozenAmount,
		TotalIncome:  acc.TotalIncome,
		TotalExpense: acc.TotalExpense,
		BankCards:    acc.BankCards,
	}, nil
}

// ===== 工具函数 =====

// generateOrderNo 生成订单号（前缀+时间戳+随机数）
func generateOrderNo(prefix string) string {
	return fmt.Sprintf("%s%s%s", prefix, time.Now().Format("20060102150405"), utils.RandomNumber(6))
}

// autoReleaseTime 计算自动放款时间
func autoReleaseTime(from time.Time, days int) *time.Time {
	t := from.AddDate(0, 0, days)
	return &t
}

// defaultJSON 返回合法的 JSON 字符串（空则返回 {}）
func defaultJSON(s string) string {
	if s == "" {
		return "{}"
	}
	return s
}

// toPaymentOrderInfo model → dto
func toPaymentOrderInfo(o *model.PaymentOrder) *dto.PaymentOrderInfo {
	return &dto.PaymentOrderInfo{
		ID:           o.ID,
		OrderNo:      o.OrderNo,
		UserID:       o.UserID,
		BizModule:    o.BizModule,
		BizID:        o.BizID,
		Title:        o.Title,
		Amount:       o.Amount,
		PayMethod:    o.PayMethod,
		PayStatus:    o.PayStatus,
		ThirdPartyNo: o.ThirdPartyNo,
		PaidAt:       o.PaidAt,
		ExpireAt:     o.ExpireAt,
		RegionID:     o.RegionID,
		CreatedAt:    o.CreatedAt,
		UpdatedAt:    o.UpdatedAt,
	}
}

// toRefundInfo model → dto
func toRefundInfo(r *model.RefundOrder) *dto.RefundInfo {
	return &dto.RefundInfo{
		ID:                 r.ID,
		RefundNo:           r.RefundNo,
		OrderID:            r.OrderID,
		OrderNo:            r.OrderNo,
		UserID:             r.UserID,
		Amount:             r.Amount,
		Reason:             r.Reason,
		Status:             r.Status,
		ThirdPartyRefundNo: r.ThirdPartyRefundNo,
		RefundMethod:       r.RefundMethod,
		ProcessedAt:        r.ProcessedAt,
		CreatedAt:          r.CreatedAt,
	}
}

// toWithdrawInfo model → dto
func toWithdrawInfo(w *model.Withdrawal) *dto.WithdrawInfo {
	return &dto.WithdrawInfo{
		ID:           w.ID,
		WithdrawalNo: w.WithdrawalNo,
		UserID:       w.UserID,
		Amount:       w.Amount,
		Fee:          w.Fee,
		ActualAmount: w.ActualAmount,
		BankCard:     w.BankCard,
		Status:       w.Status,
		RejectReason: w.RejectReason,
		ProcessedAt:  w.ProcessedAt,
		CreatedAt:    w.CreatedAt,
	}
}

// toSettlementInfo model → dto
func toSettlementInfo(s *model.Settlement) *dto.SettlementInfo {
	return &dto.SettlementInfo{
		ID:               s.ID,
		SettlementNo:     s.SettlementNo,
		MerchantID:       s.MerchantID,
		PeriodType:       s.PeriodType,
		PeriodStart:      s.PeriodStart,
		PeriodEnd:        s.PeriodEnd,
		OrderCount:       s.OrderCount,
		TotalAmount:      s.TotalAmount,
		Commission:       s.Commission,
		SettlementAmount: s.SettlementAmount,
		Status:           s.Status,
		SettledAt:        s.SettledAt,
		CreatedAt:        s.CreatedAt,
	}
}
