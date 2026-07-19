// Package service 支付中台扩展业务逻辑层
// 依据 012_pay_full.sql：交易流水/渠道/商户/回调/争议仲裁/统计
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

// 扩展错误
var (
	ErrTransactionNotFound = errors.New("交易流水不存在")
	ErrChannelNotFound      = errors.New("支付渠道不存在")
	ErrMerchantNotFound     = errors.New("商户不存在")
	ErrCallbackNotFound    = errors.New("回调记录不存在")
	ErrDisputeNotFound     = errors.New("担保争议不存在")
	ErrDisputeAlreadyResolved = errors.New("争议已仲裁")
)

// PayExtendService 支付中台扩展业务接口
type PayExtendService interface {
	// 交易流水
	ListTransactions(regionID uint, req *dto.TransactionListRequest) ([]dto.TransactionInfo, int64, error)
	GetTransaction(txnNo string) (*dto.TransactionInfo, error)

	// 渠道
	CreateChannel(regionID uint, req *dto.CreateChannelRequest) (*dto.ChannelInfo, error)
	UpdateChannel(id uint, req *dto.UpdateChannelRequest) error
	DeleteChannel(id uint) error
	ListChannels(code string, page, pageSize int) ([]dto.ChannelInfo, int64, error)
	GetChannel(id uint) (*dto.ChannelInfo, error)

	// 商户
	CreateMerchant(regionID uint, req *dto.CreateMerchantRequest) (*dto.MerchantInfo, error)
	AuditMerchant(req *dto.AuditMerchantRequest) error
	ListMerchants(status int, page, pageSize int) ([]dto.MerchantInfo, int64, error)
	GetMerchant(id uint) (*dto.MerchantInfo, error)

	// 回调
	RecordCallback(regionID uint, req *dto.RecordCallbackRequest) (*dto.CallbackInfo, error)
	ListCallbacks(orderNo, channel string, status int, page, pageSize int) ([]dto.CallbackInfo, int64, error)
	GetCallback(id uint) (*dto.CallbackInfo, error)

	// 担保争议
	CreateDispute(regionID, userID uint, req *dto.EscrowDisputeRequest) error
	ListDisputes(page, pageSize int) ([]dto.EscrowInfo, int64, error)
	ArbitrateDispute(arbitratorID uint, req *dto.EscrowArbitrateRequest) error

	// 统计
	Statistics() (*dto.PayStatisticsResponse, error)
}

type payExtendService struct {
	repo   repository.PayRepository
	extRepo repository.PayExtendRepository
}

// NewPayExtendService 创建扩展 service 实例
func NewPayExtendService(repo repository.PayRepository, extRepo repository.PayExtendRepository) PayExtendService {
	return &payExtendService{repo: repo, extRepo: extRepo}
}

// ===== 交易流水 =====

// ListTransactions 交易流水列表
func (s *payExtendService) ListTransactions(regionID uint, req *dto.TransactionListRequest) ([]dto.TransactionInfo, int64, error) {
	q := &repository.ListTransactionsQuery{
		RegionID: regionID,
		UserID:   req.UserID,
		OrderNo:  req.OrderNo,
		Channel:  req.Channel,
		Status:   req.Status,
		Page:     req.Page,
		PageSize: req.PageSize,
	}
	list, total, err := s.extRepo.ListTransactions(q)
	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.TransactionInfo, 0, len(list))
	for i := range list {
		result = append(result, *toTransactionInfo(&list[i]))
	}
	return result, total, nil
}

// GetTransaction 查询交易流水
func (s *payExtendService) GetTransaction(txnNo string) (*dto.TransactionInfo, error) {
	t, err := s.extRepo.FindTransactionByNo(txnNo)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTransactionNotFound
		}
		return nil, err
	}
	return toTransactionInfo(t), nil
}

// ===== 渠道 =====

// CreateChannel 创建支付渠道
func (s *payExtendService) CreateChannel(regionID uint, req *dto.CreateChannelRequest) (*dto.ChannelInfo, error) {
	status := req.Status
	if status == 0 {
		status = 1
	}
	c := &model.Channel{
		ChannelCode: req.ChannelCode,
		ChannelName: req.ChannelName,
		MerchantNo:  req.MerchantNo,
		AppID:       req.AppID,
		SecretKey:   req.SecretKey,
		PublicKey:   req.PublicKey,
		PrivateKey:  req.PrivateKey,
		CallbackURL: req.CallbackURL,
		NotifyURL:   req.NotifyURL,
		FeeRate:     req.FeeRate,
		FeeFixed:    req.FeeFixed,
		Config:      defaultJSON(req.Config),
		Sort:        req.Sort,
		Status:      status,
	}
	c.RegionID = regionID
	if err := s.extRepo.CreateChannel(c); err != nil {
		return nil, err
	}
	return toChannelInfo(c), nil
}

// UpdateChannel 更新渠道
func (s *payExtendService) UpdateChannel(id uint, req *dto.UpdateChannelRequest) error {
	c, err := s.extRepo.FindChannelByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrChannelNotFound
		}
		return err
	}
	_ = c
	fields := map[string]interface{}{
		"channel_name": req.ChannelName,
		"merchant_no":  req.MerchantNo,
		"app_id":       req.AppID,
		"callback_url": req.CallbackURL,
		"notify_url":   req.NotifyURL,
		"fee_rate":     req.FeeRate,
		"fee_fixed":    req.FeeFixed,
		"sort":         req.Sort,
		"status":       req.Status,
	}
	if req.SecretKey != "" {
		fields["secret_key"] = req.SecretKey
	}
	if req.PublicKey != "" {
		fields["public_key"] = req.PublicKey
	}
	if req.PrivateKey != "" {
		fields["private_key"] = req.PrivateKey
	}
	if req.Config != "" {
		fields["config"] = req.Config
	}
	return s.extRepo.UpdateChannelFields(id, fields)
}

// DeleteChannel 删除渠道
func (s *payExtendService) DeleteChannel(id uint) error {
	if _, err := s.extRepo.FindChannelByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrChannelNotFound
		}
		return err
	}
	return s.extRepo.DeleteChannel(id)
}

// ListChannels 渠道列表
func (s *payExtendService) ListChannels(code string, page, pageSize int) ([]dto.ChannelInfo, int64, error) {
	list, total, err := s.extRepo.ListChannels(code, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.ChannelInfo, 0, len(list))
	for i := range list {
		result = append(result, *toChannelInfo(&list[i]))
	}
	return result, total, nil
}

// GetChannel 查询渠道
func (s *payExtendService) GetChannel(id uint) (*dto.ChannelInfo, error) {
	c, err := s.extRepo.FindChannelByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrChannelNotFound
		}
		return nil, err
	}
	return toChannelInfo(c), nil
}

// ===== 商户 =====

// CreateMerchant 创建商户
func (s *payExtendService) CreateMerchant(regionID uint, req *dto.CreateMerchantRequest) (*dto.MerchantInfo, error) {
	cycle := req.SettlementCycle
	if cycle == "" {
		cycle = model.PeriodT1
	}
	m := &model.Merchant{
		MerchantNo:      req.MerchantNo,
		MerchantName:    req.MerchantName,
		UserID:          req.UserID,
		ContactName:     req.ContactName,
		ContactPhone:    req.ContactPhone,
		FeeRate:         req.FeeRate,
		SettlementCycle: cycle,
		BusinessLicense: req.BusinessLicense,
		BusinessScope:   req.BusinessScope,
		BankAccount:     defaultJSON(req.BankAccount),
		Status:          model.MerchantStatusPending,
	}
	m.RegionID = regionID
	if err := s.extRepo.CreateMerchant(m); err != nil {
		return nil, err
	}
	return toMerchantInfo(m), nil
}

// AuditMerchant 审核商户
func (s *payExtendService) AuditMerchant(req *dto.AuditMerchantRequest) error {
	m, err := s.extRepo.FindMerchantByID(req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMerchantNotFound
		}
		return err
	}
	_ = m
	return s.extRepo.UpdateMerchantFields(req.ID, map[string]interface{}{
		"status":       req.Status,
		"audit_remark": req.AuditRemark,
	})
}

// ListMerchants 商户列表
func (s *payExtendService) ListMerchants(status int, page, pageSize int) ([]dto.MerchantInfo, int64, error) {
	list, total, err := s.extRepo.ListMerchants(status, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.MerchantInfo, 0, len(list))
	for i := range list {
		result = append(result, *toMerchantInfo(&list[i]))
	}
	return result, total, nil
}

// GetMerchant 查询商户
func (s *payExtendService) GetMerchant(id uint) (*dto.MerchantInfo, error) {
	m, err := s.extRepo.FindMerchantByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMerchantNotFound
		}
		return nil, err
	}
	return toMerchantInfo(m), nil
}

// ===== 回调 =====

// RecordCallback 记录三方回调
func (s *payExtendService) RecordCallback(regionID uint, req *dto.RecordCallbackRequest) (*dto.CallbackInfo, error) {
	c := &model.Callback{
		OrderNo:      req.OrderNo,
		Channel:      req.Channel,
		ThirdPartyNo: req.ThirdPartyNo,
		NotifyType:   req.NotifyType,
		RawData:      req.RawData,
		ParsedData:   defaultJSON(req.ParsedData),
		Signature:    req.Signature,
		Status:       model.CallbackStatusPending,
	}
	c.RegionID = regionID
	if err := s.extRepo.CreateCallback(c); err != nil {
		return nil, err
	}
	return toCallbackInfo(c), nil
}

// ListCallbacks 回调列表
func (s *payExtendService) ListCallbacks(orderNo, channel string, status int, page, pageSize int) ([]dto.CallbackInfo, int64, error) {
	list, total, err := s.extRepo.ListCallbacks(orderNo, channel, status, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.CallbackInfo, 0, len(list))
	for i := range list {
		result = append(result, *toCallbackInfo(&list[i]))
	}
	return result, total, nil
}

// GetCallback 查询回调
func (s *payExtendService) GetCallback(id uint) (*dto.CallbackInfo, error) {
	c, err := s.extRepo.FindCallbackByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCallbackNotFound
		}
		return nil, err
	}
	return toCallbackInfo(c), nil
}

// ===== 担保争议 =====

// CreateDispute 发起担保争议
func (s *payExtendService) CreateDispute(regionID, userID uint, req *dto.EscrowDisputeRequest) error {
	o, err := s.repo.FindOrderByNo(req.OrderNo)
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
	if escrow.DisputeStatus == model.DisputeStatusArbitrated {
		return ErrDisputeAlreadyResolved
	}
	// 买家本人才能发起争议
	if escrow.UserID != userID {
		return errors.New("无权发起争议")
	}
	now := time.Now()
	return s.repo.UpdateEscrowFields(escrow.ID, map[string]interface{}{
		"dispute_status": model.DisputeStatusActive,
		"dispute_reason": req.Reason,
		"arbitrator_id":  0,
		"arbitrated_at":  &now,
	})
}

// ListDisputes 争议列表（M 端）
func (s *payExtendService) ListDisputes(page, pageSize int) ([]dto.EscrowInfo, int64, error) {
	list, total, err := s.extRepo.ListDisputedEscrows(page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.EscrowInfo, 0, len(list))
	for i := range list {
		result = append(result, *toDisputeInfo(&list[i]))
	}
	return result, total, nil
}

// ArbitrateDispute 仲裁争议（M 端）
func (s *payExtendService) ArbitrateDispute(arbitratorID uint, req *dto.EscrowArbitrateRequest) error {
	o, err := s.repo.FindOrderByNo(req.OrderNo)
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
	if escrow.DisputeStatus == model.DisputeStatusArbitrated {
		return ErrDisputeAlreadyResolved
	}
	now := time.Now()
	if err := s.repo.UpdateEscrowFields(escrow.ID, map[string]interface{}{
		"dispute_status":     model.DisputeStatusArbitrated,
		"arbitrator_id":      arbitratorID,
		"arbitration_remark": req.Remark,
		"arbitrated_at":      &now,
	}); err != nil {
		return err
	}

	// 根据仲裁结果执行放款/退款/分账
	switch req.Result {
	case "release":
		// 放款给商家
		_ = s.repo.UpdateEscrowFields(escrow.ID, map[string]interface{}{
			"status":     model.EscrowStatusReleased,
			"release_at": &now,
		})
		if escrow.MerchantID > 0 {
			if mAcc, err := s.repo.GetOrCreateAccount(escrow.MerchantID, o.RegionID); err == nil {
				_ = s.repo.UpdateAccountFields(mAcc.ID, map[string]interface{}{
					"balance":      mAcc.Balance + escrow.Amount,
					"total_income": mAcc.TotalIncome + escrow.Amount,
				})
			}
		}
	case "refund":
		// 全额退款给买家
		_ = s.repo.UpdateEscrowFields(escrow.ID, map[string]interface{}{
			"status": model.EscrowStatusRefunded,
		})
		if acc, err := s.repo.GetOrCreateAccount(o.UserID, o.RegionID); err == nil {
			_ = s.repo.UpdateAccountFields(acc.ID, map[string]interface{}{
				"balance":       acc.Balance + escrow.Amount,
				"frozen_amount":  acc.FrozenAmount - escrow.Amount,
				"total_income":   acc.TotalIncome + escrow.Amount,
			})
		}
	case "split":
		// 分账：买家按比例，剩余给商家
		ratio := req.BuyerRatio
		if ratio <= 0 {
			ratio = 0.5
		}
		if ratio > 1 {
			ratio = 1
		}
		buyerAmount := escrow.Amount * ratio
		merchantAmount := escrow.Amount - buyerAmount
		_ = s.repo.UpdateEscrowFields(escrow.ID, map[string]interface{}{
			"status":     model.EscrowStatusPartRefund,
			"release_at": &now,
		})
		if buyerAmount > 0 {
			if acc, err := s.repo.GetOrCreateAccount(o.UserID, o.RegionID); err == nil {
				_ = s.repo.UpdateAccountFields(acc.ID, map[string]interface{}{
					"balance":       acc.Balance + buyerAmount,
					"frozen_amount": acc.FrozenAmount - escrow.Amount,
				})
			}
		}
		if merchantAmount > 0 && escrow.MerchantID > 0 {
			if mAcc, err := s.repo.GetOrCreateAccount(escrow.MerchantID, o.RegionID); err == nil {
				_ = s.repo.UpdateAccountFields(mAcc.ID, map[string]interface{}{
					"balance":      mAcc.Balance + merchantAmount,
					"total_income": mAcc.TotalIncome + merchantAmount,
				})
			}
		}
	}
	return nil
}

// ===== 统计 =====

// Statistics 支付总览统计
func (s *payExtendService) Statistics() (*dto.PayStatisticsResponse, error) {
	resp := &dto.PayStatisticsResponse{
		ChannelStats: []dto.ChannelStatItem{},
	}
	totalOrders, _ := s.extRepo.StatTotalOrders()
	totalAmount, _ := s.extRepo.StatTotalAmount()
	refundAmount, _ := s.extRepo.StatRefundAmount()
	escrowAmount, _ := s.extRepo.StatEscrowAmount()
	todayCount, _ := s.extRepo.StatTodayCount()
	todayAmount, _ := s.extRepo.StatTodayAmount()
	successRate, _ := s.extRepo.StatSuccessRate()
	refundRate, _ := s.extRepo.StatRefundRate()
	rows, _ := s.extRepo.StatByChannel()

	resp.TotalOrders = totalOrders
	resp.TotalAmount = totalAmount
	resp.RefundAmount = refundAmount
	resp.EscrowAmount = escrowAmount
	resp.TodayCount = todayCount
	resp.TodayAmount = todayAmount
	resp.SuccessRate = successRate
	resp.RefundRate = refundRate
	for _, r := range rows {
		resp.ChannelStats = append(resp.ChannelStats, dto.ChannelStatItem{
			Channel: r.Channel,
			Count:   r.Count,
			Amount:  r.Amount,
		})
	}
	return resp, nil
}

// ===== 工具函数 =====

// generateTransactionNo 生成交易流水号
func generateTransactionNo(channel string) string {
	prefix := "TXN"
	switch channel {
	case model.ChannelWechat:
		prefix = "WX"
	case model.ChannelAlipay:
		prefix = "AL"
	case model.ChannelUnionPay:
		prefix = "UP"
	case model.ChannelStripe:
		prefix = "ST"
	}
	return fmt.Sprintf("%s%s%s", prefix, time.Now().Format("20060102150405"), utils.RandomNumber(8))
}

// toTransactionInfo model → dto
func toTransactionInfo(t *model.Transaction) *dto.TransactionInfo {
	return &dto.TransactionInfo{
		ID:            t.ID,
		TransactionNo: t.TransactionNo,
		OrderID:       t.OrderID,
		OrderNo:       t.OrderNo,
		UserID:        t.UserID,
		Channel:       t.Channel,
		ThirdPartyNo:  t.ThirdPartyNo,
		Amount:        t.Amount,
		Fee:           t.Fee,
		Status:        t.Status,
		ChannelResp:   t.ChannelResp,
		PaidAt:        t.PaidAt,
		RegionID:      t.RegionID,
		CreatedAt:     t.CreatedAt,
	}
}

// toChannelInfo model → dto
func toChannelInfo(c *model.Channel) *dto.ChannelInfo {
	return &dto.ChannelInfo{
		ID:          c.ID,
		ChannelCode: c.ChannelCode,
		ChannelName: c.ChannelName,
		MerchantNo:  c.MerchantNo,
		AppID:       c.AppID,
		CallbackURL: c.CallbackURL,
		NotifyURL:   c.NotifyURL,
		FeeRate:     c.FeeRate,
		FeeFixed:    c.FeeFixed,
		Status:      c.Status,
		Sort:        c.Sort,
		CreatedAt:   c.CreatedAt,
	}
}

// toMerchantInfo model → dto
func toMerchantInfo(m *model.Merchant) *dto.MerchantInfo {
	return &dto.MerchantInfo{
		ID:              m.ID,
		MerchantNo:      m.MerchantNo,
		MerchantName:    m.MerchantName,
		UserID:          m.UserID,
		ContactName:     m.ContactName,
		ContactPhone:    m.ContactPhone,
		FeeRate:         m.FeeRate,
		SettlementCycle: m.SettlementCycle,
		BusinessLicense: m.BusinessLicense,
		BusinessScope:   m.BusinessScope,
		BankAccount:     m.BankAccount,
		Status:          m.Status,
		AuditRemark:     m.AuditRemark,
		CreatedAt:       m.CreatedAt,
	}
}

// toCallbackInfo model → dto
func toCallbackInfo(c *model.Callback) *dto.CallbackInfo {
	return &dto.CallbackInfo{
		ID:           c.ID,
		OrderNo:      c.OrderNo,
		Channel:      c.Channel,
		ThirdPartyNo: c.ThirdPartyNo,
		NotifyType:   c.NotifyType,
		RawData:      c.RawData,
		ParsedData:   c.ParsedData,
		Signature:    c.Signature,
		Status:       c.Status,
		ProcessCount: c.ProcessCount,
		ErrorMsg:     c.ErrorMsg,
		ProcessedAt:  c.ProcessedAt,
		CreatedAt:    c.CreatedAt,
	}
}

// toDisputeInfo model → dto（争议信息）
func toDisputeInfo(e *model.EscrowAccount) *dto.EscrowInfo {
	return &dto.EscrowInfo{
		ID:            e.ID,
		OrderID:       e.OrderID,
		OrderNo:       e.OrderNo,
		UserID:        e.UserID,
		MerchantID:    e.MerchantID,
		Amount:        e.Amount,
		Status:        e.Status,
		FrozenAt:      e.FrozenAt,
		ReleaseAt:     e.ReleaseAt,
		AutoReleaseAt: e.AutoReleaseAt,
	}
}
