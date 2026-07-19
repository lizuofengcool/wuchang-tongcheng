// Package service 同城车辆买卖业务逻辑层 - 统计 + 推荐 + 担保 + 合同
// CarStatistic 统计（RegionBaseModel）
// CarRecommendation 推荐（RegionBaseModel）
// CarEscrow 担保交易（RegionBaseModel）
// CarContract 合同（RegionBaseModel）
package service

import (
	"errors"
	"time"

	"wuchang-tongcheng/internal/modules/car/dto"
	"wuchang-tongcheng/internal/modules/car/model"
	"wuchang-tongcheng/internal/modules/car/repository"
	"wuchang-tongcheng/internal/pkg/utils"
)

var (
	ErrStatisticNotFound = errors.New("统计记录不存在")

	ErrRecommendationNotFound = errors.New("推荐记录不存在")

	ErrEscrowNotFound      = errors.New("担保交易不存在")
	ErrEscrowStatusInvalid = errors.New("担保状态不允许此操作")
	ErrEscrowNoPermission  = errors.New("无权操作此担保交易")

	ErrContractNotFound      = errors.New("合同不存在")
	ErrContractStatusInvalid = errors.New("合同状态不允许此操作")
	ErrContractNoPermission  = errors.New("无权操作此合同")
)

// ===== StatisticService 统计业务接口 =====

// StatisticService 统计业务接口
type StatisticService interface {
	// 列表
	List(req *dto.StatisticListRequest) (*utils.Pagination, []dto.StatisticInfo, error)
	// 平台总览
	Overview() (*dto.OverviewResponse, error)
	// 卖家总览
	SellerOverview(userID uint) (*dto.SellerOverviewResponse, error)
	// 热门车源
	HotCars(limit int) ([]dto.HotItemResponse, error)
	// 价格趋势
	PriceTrend(regionID uint, startDate, endDate string) ([]dto.PriceTrendResponse, error)
}

type statisticService struct {
	statRepo repository.StatisticRepository
}

// NewStatisticService 创建统计 service 实例
func NewStatisticService(statRepo repository.StatisticRepository) StatisticService {
	return &statisticService{statRepo: statRepo}
}

// toStatisticInfo model -> dto
func toStatisticInfo(s *model.CarStatistic) *dto.StatisticInfo {
	return &dto.StatisticInfo{
		ID:              s.ID,
		StatDate:        s.StatDate,
		StatType:        s.StatType,
		TargetID:        s.TargetID,
		TargetName:      s.TargetName,
		ImpressionCount: s.ImpressionCount,
		ClickCount:      s.ClickCount,
		FavCount:        s.FavCount,
		ContactCount:    s.ContactCount,
		TestDriveCount:  s.TestDriveCount,
		DealCount:       s.DealCount,
		ConversionRate:  s.ConversionRate,
		AvgPrice:        s.AvgPrice,
		AvgDealDays:     s.AvgDealDays,
		RegionID:        s.RegionID,
	}
}

// parseStatDateRange 解析日期范围字符串（YYYY-MM-DD）
func parseStatDateRange(startDate, endDate string) (*time.Time, *time.Time) {
	var st, et *time.Time
	if startDate != "" {
		if t, err := time.ParseInLocation("2006-01-02", startDate, time.Local); err == nil {
			st = &t
		}
	}
	if endDate != "" {
		if t, err := time.ParseInLocation("2006-01-02", endDate, time.Local); err == nil {
			// endDate 包含整天：+24h-1s
			end := t.Add(24*time.Hour - time.Second)
			et = &end
		}
	}
	return st, et
}

// List 统计列表
func (s *statisticService) List(req *dto.StatisticListRequest) (*utils.Pagination, []dto.StatisticInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	st, et := parseStatDateRange(req.StartDate, req.EndDate)
	opts := repository.StatisticListOptions{
		StatType:  req.StatType,
		TargetID:  req.TargetID,
		StartDate: st,
		EndDate:   et,
	}
	list, total, err := s.statRepo.List(opts, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	infos := make([]dto.StatisticInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toStatisticInfo(&list[i]))
	}
	return pagination, infos, nil
}

// Overview 平台总览统计（基于统计表聚合）
func (s *statisticService) Overview() (*dto.OverviewResponse, error) {
	resp := &dto.OverviewResponse{}
	// 使用 StatType=platform 的统计聚合
	summary, err := s.statRepo.SumByType(model.StatTypePlatform, nil, nil)
	if err == nil && summary != nil {
		resp.TotalDealings = int64(summary.DealCount)
		resp.TotalAmount = summary.AvgPrice // MVP：AvgPrice 字段在 platform 类型下复用为累计成交额
	}
	return resp, nil
}

// SellerOverview 卖家总览（基于统计表 dealer 类型）
func (s *statisticService) SellerOverview(userID uint) (*dto.SellerOverviewResponse, error) {
	resp := &dto.SellerOverviewResponse{
		UserID: userID,
	}
	// 从统计表查 StatType=dealer, TargetID=userID 的最近一条
	list, err := s.statRepo.ListByTarget(model.StatTypeDealer, userID, nil, nil)
	if err != nil || len(list) == 0 {
		return resp, nil
	}
	// 取最近一条作为概览
	latest := list[0]
	resp.TotalViews = int64(latest.ImpressionCount)
	resp.TotalFavs = int64(latest.FavCount)
	resp.TotalContacts = int64(latest.ContactCount)
	resp.TotalTestDrives = int64(latest.TestDriveCount)
	resp.TotalDeals = int64(latest.DealCount)
	resp.TotalAmount = latest.AvgPrice // MVP：AvgPrice 字段在 dealer 类型下复用为累计成交额
	resp.AvgDealDays = latest.AvgDealDays
	if latest.ImpressionCount > 0 {
		resp.ConversionRate = float64(latest.DealCount) / float64(latest.ImpressionCount)
	}
	return resp, nil
}

// HotCars 热门车源（按成交数+点击数排序）
func (s *statisticService) HotCars(limit int) ([]dto.HotItemResponse, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	list, err := s.statRepo.ListTopByType(model.StatTypeCar, limit, nil, nil)
	if err != nil {
		return nil, err
	}
	result := make([]dto.HotItemResponse, 0, len(list))
	for i := range list {
		s := &list[i]
		result = append(result, dto.HotItemResponse{
			CarID:      s.TargetID,
			Title:      s.TargetName,
			ViewCount:  s.ImpressionCount,
			FavCount:   s.FavCount,
			ContactCount: s.ContactCount,
		})
	}
	return result, nil
}

// PriceTrend 价格趋势
func (s *statisticService) PriceTrend(regionID uint, startDate, endDate string) ([]dto.PriceTrendResponse, error) {
	st, et := parseStatDateRange(startDate, endDate)
	list, err := s.statRepo.ListByRegion(regionID, model.StatTypeRegion, st, et)
	if err != nil {
		return nil, err
	}
	result := make([]dto.PriceTrendResponse, 0, len(list))
	for i := range list {
		s := &list[i]
		result = append(result, dto.PriceTrendResponse{
			Date:      s.StatDate.Format("2006-01-02"),
			AvgPrice:  s.AvgPrice,
			DealCount: s.DealCount,
		})
	}
	return result, nil
}

// ===== RecommendationService 推荐业务接口 =====

// RecommendationService 推荐业务接口
type RecommendationService interface {
	// C 端
	ListByUser(userID uint, recType string, page, pageSize int) (*utils.Pagination, []dto.RecommendationInfo, error)
	ListByCarID(carID uint, page, pageSize int) (*utils.Pagination, []dto.RecommendationInfo, error)
	MarkClicked(id uint, userID uint) error
	MarkContacted(id uint, userID uint) error
	MarkDismissed(id uint, userID uint) error
	// M 端
	AdminList(req *dto.RecommendationListRequest) (*utils.Pagination, []dto.RecommendationInfo, error)
	Delete(id uint) error
}

type recommendationService struct {
	repo repository.RecommendationRepository
}

// NewRecommendationService 创建推荐 service 实例
func NewRecommendationService(repo repository.RecommendationRepository) RecommendationService {
	return &recommendationService{repo: repo}
}

// recStatusText 推荐状态文本
func recStatusText(status int) string {
	switch status {
	case model.RecStatusPending:
		return "待展示"
	case model.RecStatusShown:
		return "已展示"
	case model.RecStatusClicked:
		return "已点击"
	case model.RecStatusContacted:
		return "已联系"
	case model.RecStatusDismissed:
		return "已忽略"
	case model.RecStatusExpired:
		return "已过期"
	}
	return ""
}

// toRecommendationInfo model -> dto
func toRecommendationInfo(r *model.CarRecommendation) *dto.RecommendationInfo {
	return &dto.RecommendationInfo{
		ID:             r.ID,
		UserID:         r.UserID,
		CarID:          r.CarID,
		RecType:        r.RecType,
		Source:         r.Source,
		Score:          r.Score,
		Reason:         r.Reason,
		PriceMatch:     r.PriceMatch,
		BrandMatch:     r.BrandMatch,
		TypeMatch:      r.TypeMatch,
		ConditionMatch: r.ConditionMatch,
		Status:         r.Status,
		StatusText:     recStatusText(r.Status),
		ClickedAt:      r.ClickedAt,
		ContactedAt:    r.ContactedAt,
		ViewedAt:       r.ViewedAt,
		CreatedAt:      r.CreatedAt,
	}
}

// ListByUser 用户推荐列表
func (s *recommendationService) ListByUser(userID uint, recType string, page, pageSize int) (*utils.Pagination, []dto.RecommendationInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByUser(userID, recType, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	infos := make([]dto.RecommendationInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toRecommendationInfo(&list[i]))
	}
	return pagination, infos, nil
}

// ListByCarID 按车源反查推荐记录
func (s *recommendationService) ListByCarID(carID uint, page, pageSize int) (*utils.Pagination, []dto.RecommendationInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByCarID(carID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	infos := make([]dto.RecommendationInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toRecommendationInfo(&list[i]))
	}
	return pagination, infos, nil
}

// MarkClicked 标记已点击
func (s *recommendationService) MarkClicked(id uint, userID uint) error {
	rec, err := s.repo.FindByID(id)
	if err != nil {
		return ErrRecommendationNotFound
	}
	if rec.UserID != userID {
		return ErrRecommendationNotFound
	}
	now := time.Now()
	return s.repo.UpdateStatus(id, model.RecStatusClicked, map[string]interface{}{
		"clicked_at": &now,
	})
}

// MarkContacted 标记已联系
func (s *recommendationService) MarkContacted(id uint, userID uint) error {
	rec, err := s.repo.FindByID(id)
	if err != nil {
		return ErrRecommendationNotFound
	}
	if rec.UserID != userID {
		return ErrRecommendationNotFound
	}
	now := time.Now()
	return s.repo.UpdateStatus(id, model.RecStatusContacted, map[string]interface{}{
		"contacted_at": &now,
	})
}

// MarkDismissed 标记已忽略
func (s *recommendationService) MarkDismissed(id uint, userID uint) error {
	rec, err := s.repo.FindByID(id)
	if err != nil {
		return ErrRecommendationNotFound
	}
	if rec.UserID != userID {
		return ErrRecommendationNotFound
	}
	now := time.Now()
	return s.repo.UpdateStatus(id, model.RecStatusDismissed, map[string]interface{}{
		"dismissed_at": &now,
	})
}

// AdminList M 端推荐列表
func (s *recommendationService) AdminList(req *dto.RecommendationListRequest) (*utils.Pagination, []dto.RecommendationInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.RecommendationListOptions{
		UserID:  req.UserID,
		CarID:   req.CarID,
		RecType: req.RecType,
		Source:  req.Source,
		Status:  req.Status,
	}
	list, total, err := s.repo.List(opts, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	infos := make([]dto.RecommendationInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toRecommendationInfo(&list[i]))
	}
	return pagination, infos, nil
}

// Delete M 端删除推荐
func (s *recommendationService) Delete(id uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		return ErrRecommendationNotFound
	}
	return s.repo.Update(id, map[string]interface{}{"status": model.RecStatusExpired})
}

// ===== EscrowService 担保交易业务接口 =====

// EscrowService 担保交易业务接口
type EscrowService interface {
	// C 端
	GetByID(id uint, userID uint) (*dto.EscrowInfo, error)
	GetByEscrowNo(escrowNo string, userID uint) (*dto.EscrowInfo, error)
	List(userID uint, req *dto.EscrowListRequest) (*utils.Pagination, []dto.EscrowInfo, error)
	ListByCarID(carID uint, page, pageSize int) (*utils.Pagination, []dto.EscrowInfo, error)
	Action(id uint, userID uint, req *dto.EscrowActionRequest) error

	// M 端
	AdminList(req *dto.EscrowListRequest) (*utils.Pagination, []dto.EscrowInfo, error)
	AdminGetByID(id uint) (*dto.EscrowInfo, error)
	UpdateStatus(id uint, status int) error
}

type escrowService struct {
	repo repository.EscrowRepository
}

// NewEscrowService 创建担保交易 service 实例
func NewEscrowService(repo repository.EscrowRepository) EscrowService {
	return &escrowService{repo: repo}
}

// escrowStatusText 担保状态文本
func escrowStatusText(status int) string {
	switch status {
	case model.EscrowStatusPending:
		return "待支付"
	case model.EscrowStatusPaid:
		return "已支付"
	case model.EscrowStatusReleased:
		return "已放款"
	case model.EscrowStatusRefunded:
		return "已退款"
	case model.EscrowStatusDisputed:
		return "争议中"
	case model.EscrowStatusArbitrated:
		return "已仲裁"
	case model.EscrowStatusCanceled:
		return "已取消"
	}
	return ""
}

// toEscrowInfo model -> dto
func toEscrowInfo(e *model.CarEscrow) *dto.EscrowInfo {
	return &dto.EscrowInfo{
		ID:                e.ID,
		EscrowNo:          e.EscrowNo,
		EscrowType:        e.EscrowType,
		CarID:             e.CarID,
		ListingID:         e.ListingID,
		ContractID:        e.ContractID,
		PayerID:           e.PayerID,
		PayeeID:           e.PayeeID,
		DealerID:          e.DealerID,
		Amount:            e.Amount,
		PlatformFee:       e.PlatformFee,
		DealerFee:         e.DealerFee,
		PayeeAmount:       e.PayeeAmount,
		Status:            e.Status,
		StatusText:        escrowStatusText(e.Status),
		PayMethod:         e.PayMethod,
		PayTradeNo:        e.PayTradeNo,
		PaidAt:            e.PaidAt,
		FrozenAt:          e.FrozenAt,
		ReleaseAt:         e.ReleaseAt,
		RefundedAt:        e.RefundedAt,
		AutoReleaseAt:     e.AutoReleaseAt,
		DisputeReason:     e.DisputeReason,
		ArbitrationResult: e.ArbitrationResult,
		CompletedAt:       e.CompletedAt,
		RegionID:          e.RegionID,
		CreatedAt:         e.CreatedAt,
		UpdatedAt:         e.UpdatedAt,
	}
}

// GetByID 查询担保详情（C 端校验归属）
func (s *escrowService) GetByID(id uint, userID uint) (*dto.EscrowInfo, error) {
	e, err := s.repo.FindByID(id)
	if err != nil {
		return nil, ErrEscrowNotFound
	}
	// userID == 0 表示 M 端强制操作
	if userID != 0 && e.PayerID != userID && e.PayeeID != userID {
		return nil, ErrEscrowNoPermission
	}
	return toEscrowInfo(e), nil
}

// GetByEscrowNo 按单号查询
func (s *escrowService) GetByEscrowNo(escrowNo string, userID uint) (*dto.EscrowInfo, error) {
	e, err := s.repo.FindByEscrowNo(escrowNo)
	if err != nil {
		return nil, ErrEscrowNotFound
	}
	if userID != 0 && e.PayerID != userID && e.PayeeID != userID {
		return nil, ErrEscrowNoPermission
	}
	return toEscrowInfo(e), nil
}

// List C 端列表（按用户）
func (s *escrowService) List(userID uint, req *dto.EscrowListRequest) (*utils.Pagination, []dto.EscrowInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.EscrowListOptions{
		UserID:     userID,
		Role:       "all",
		CarID:      req.CarID,
		EscrowType: req.EscrowType,
		Status:     req.Status,
	}
	list, total, err := s.repo.List(opts, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	infos := make([]dto.EscrowInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toEscrowInfo(&list[i]))
	}
	return pagination, infos, nil
}

// ListByCarID 按车源反查
func (s *escrowService) ListByCarID(carID uint, page, pageSize int) (*utils.Pagination, []dto.EscrowInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByCarID(carID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	infos := make([]dto.EscrowInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toEscrowInfo(&list[i]))
	}
	return pagination, infos, nil
}

// Action 担保操作（放款/退款/争议/仲裁）
func (s *escrowService) Action(id uint, userID uint, req *dto.EscrowActionRequest) error {
	e, err := s.repo.FindByID(id)
	if err != nil {
		return ErrEscrowNotFound
	}
	// userID == 0 表示 M 端强制操作
	if userID != 0 && e.PayerID != userID && e.PayeeID != userID {
		return ErrEscrowNoPermission
	}
	now := time.Now()
	switch req.Action {
	case "release":
		// 放款：仅已支付状态可放款
		if e.Status != model.EscrowStatusPaid {
			return ErrEscrowStatusInvalid
		}
		return s.repo.UpdateStatus(id, model.EscrowStatusReleased, map[string]interface{}{
			"release_at":   &now,
			"completed_at": &now,
		})
	case "refund":
		// 退款：仅已支付状态可退款
		if e.Status != model.EscrowStatusPaid {
			return ErrEscrowStatusInvalid
		}
		return s.repo.UpdateStatus(id, model.EscrowStatusRefunded, map[string]interface{}{
			"refunded_at":  &now,
			"completed_at": &now,
		})
	case "dispute":
		// 争议：仅已支付状态可发起争议
		if e.Status != model.EscrowStatusPaid {
			return ErrEscrowStatusInvalid
		}
		return s.repo.UpdateStatus(id, model.EscrowStatusDisputed, map[string]interface{}{
			"dispute_reason": req.DisputeReason,
		})
	case "arbitrate":
		// 仲裁：仅争议中状态可仲裁（M 端操作）
		if e.Status != model.EscrowStatusDisputed {
			return ErrEscrowStatusInvalid
		}
		return s.repo.UpdateStatus(id, model.EscrowStatusArbitrated, map[string]interface{}{
			"arbitration_result": req.ArbitrationResult,
			"completed_at":       &now,
		})
	}
	return ErrEscrowStatusInvalid
}

// AdminList M 端列表
func (s *escrowService) AdminList(req *dto.EscrowListRequest) (*utils.Pagination, []dto.EscrowInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.EscrowAdminListOptions{
		CarID:      req.CarID,
		DealerID:   req.DealerID,
		EscrowType: req.EscrowType,
		Status:     req.Status,
	}
	// M 端可按付款方/收款方 ID 过滤（任一非 0 即转为 UserID）
	if req.PayerID > 0 {
		opts.UserID = req.PayerID
	} else if req.PayeeID > 0 {
		opts.UserID = req.PayeeID
	}
	list, total, err := s.repo.AdminList(opts, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	infos := make([]dto.EscrowInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toEscrowInfo(&list[i]))
	}
	return pagination, infos, nil
}

// AdminGetByID M 端查询
func (s *escrowService) AdminGetByID(id uint) (*dto.EscrowInfo, error) {
	e, err := s.repo.FindByID(id)
	if err != nil {
		return nil, ErrEscrowNotFound
	}
	return toEscrowInfo(e), nil
}

// UpdateStatus M 端更新状态
func (s *escrowService) UpdateStatus(id uint, status int) error {
	if _, err := s.repo.FindByID(id); err != nil {
		return ErrEscrowNotFound
	}
	return s.repo.UpdateStatus(id, status, nil)
}

// ===== ContractService 合同业务接口 =====

// ContractService 合同业务接口
type ContractService interface {
	// C 端
	GetByID(id uint, userID uint) (*dto.ContractInfo, error)
	GetByContractNo(contractNo string, userID uint) (*dto.ContractInfo, error)
	List(userID uint, req *dto.ContractListRequest) (*utils.Pagination, []dto.ContractInfo, error)
	ListByCarID(carID uint, page, pageSize int) (*utils.Pagination, []dto.ContractInfo, error)
	Sign(id uint, operatorID uint) error
	Terminate(id uint, operatorID uint, reason string) error

	// M 端
	AdminList(req *dto.ContractListRequest) (*utils.Pagination, []dto.ContractInfo, error)
	AdminGetByID(id uint) (*dto.ContractInfo, error)
	UpdateStatus(id uint, status int) error
}

type contractService struct {
	repo repository.ContractRepository
}

// NewContractService 创建合同 service 实例
func NewContractService(repo repository.ContractRepository) ContractService {
	return &contractService{repo: repo}
}

// contractStatusText 合同状态文本
func contractStatusText(status int) string {
	switch status {
	case model.ContractStatusDraft:
		return "草稿"
	case model.ContractStatusPending:
		return "待签署"
	case model.ContractStatusSigned:
		return "已签署"
	case model.ContractStatusEffective:
		return "已生效"
	case model.ContractStatusFulfilling:
		return "履行中"
	case model.ContractStatusCompleted:
		return "已完成"
	case model.ContractStatusTerminated:
		return "已终止"
	case model.ContractStatusCanceled:
		return "已取消"
	}
	return ""
}

// toContractInfo model -> dto
func toContractInfo(c *model.CarContract) *dto.ContractInfo {
	info := &dto.ContractInfo{
		ID:            c.ID,
		ContractNo:    c.ContractNo,
		ContractType:  c.ContractType,
		CarID:         c.CarID,
		ListingID:     c.ListingID,
		SellerID:      c.SellerID,
		SellerName:    c.SellerName,
		SellerPhone:   c.SellerPhone,
		SellerIDCard:  c.SellerIDCard,
		BuyerID:       c.BuyerID,
		BuyerName:     c.BuyerName,
		BuyerPhone:    c.BuyerPhone,
		BuyerIDCard:   c.BuyerIDCard,
		AgentID:       c.AgentID,
		AgentName:     c.AgentName,
		DealPrice:     c.DealPrice,
		Deposit:       c.Deposit,
		PaymentMethod: c.PaymentMethod,
		LoanAmount:    c.LoanAmount,
		LoanPeriods:   c.LoanPeriods,
		TransferFee:   c.TransferFee,
		ServiceFee:    c.ServiceFee,
		ContractURL:   c.ContractURL,
		SignedAt:      c.SignedAt,
		EffectiveAt:   c.EffectiveAt,
		ExpiredAt:     c.ExpiredAt,
		TerminatedAt:  c.TerminatedAt,
		Status:        c.Status,
		StatusText:    contractStatusText(c.Status),
		Remark:        c.Remark,
		RegionID:      c.RegionID,
		CreatedAt:     c.CreatedAt,
		UpdatedAt:     c.UpdatedAt,
	}
	if c.Attachments != nil {
		info.Attachments = c.Attachments
	}
	return info
}

// GetByID 查询合同详情
func (s *contractService) GetByID(id uint, userID uint) (*dto.ContractInfo, error) {
	c, err := s.repo.FindByID(id)
	if err != nil {
		return nil, ErrContractNotFound
	}
	if userID != 0 && c.SellerID != userID && c.BuyerID != userID {
		return nil, ErrContractNoPermission
	}
	return toContractInfo(c), nil
}

// GetByContractNo 按合同号查询
func (s *contractService) GetByContractNo(contractNo string, userID uint) (*dto.ContractInfo, error) {
	c, err := s.repo.FindByContractNo(contractNo)
	if err != nil {
		return nil, ErrContractNotFound
	}
	if userID != 0 && c.SellerID != userID && c.BuyerID != userID {
		return nil, ErrContractNoPermission
	}
	return toContractInfo(c), nil
}

// List C 端合同列表
func (s *contractService) List(userID uint, req *dto.ContractListRequest) (*utils.Pagination, []dto.ContractInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.ContractListOptions{
		UserID:       userID,
		Role:         "all",
		CarID:        req.CarID,
		ContractType: req.ContractType,
		Status:       req.Status,
	}
	list, total, err := s.repo.List(opts, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	infos := make([]dto.ContractInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toContractInfo(&list[i]))
	}
	return pagination, infos, nil
}

// ListByCarID 按车源反查合同
func (s *contractService) ListByCarID(carID uint, page, pageSize int) (*utils.Pagination, []dto.ContractInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByCarID(carID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	infos := make([]dto.ContractInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toContractInfo(&list[i]))
	}
	return pagination, infos, nil
}

// Sign 签署合同
func (s *contractService) Sign(id uint, operatorID uint) error {
	c, err := s.repo.FindByID(id)
	if err != nil {
		return ErrContractNotFound
	}
	if operatorID != 0 && c.SellerID != operatorID && c.BuyerID != operatorID {
		return ErrContractNoPermission
	}
	// 仅草稿/待签状态可签
	if c.Status != model.ContractStatusDraft && c.Status != model.ContractStatusPending {
		return ErrContractStatusInvalid
	}
	now := time.Now()
	// 签署后直接进入生效状态
	return s.repo.UpdateStatus(id, model.ContractStatusEffective, map[string]interface{}{
		"signed_at":    &now,
		"effective_at": &now,
	})
}

// Terminate 终止合同
func (s *contractService) Terminate(id uint, operatorID uint, reason string) error {
	c, err := s.repo.FindByID(id)
	if err != nil {
		return ErrContractNotFound
	}
	if operatorID != 0 && c.SellerID != operatorID && c.BuyerID != operatorID {
		return ErrContractNoPermission
	}
	// 已完成/已取消的不能终止
	if c.Status == model.ContractStatusCompleted || c.Status == model.ContractStatusCanceled {
		return ErrContractStatusInvalid
	}
	now := time.Now()
	return s.repo.UpdateStatus(id, model.ContractStatusTerminated, map[string]interface{}{
		"terminated_at": &now,
		"remark":        reason,
	})
}

// AdminList M 端合同列表
func (s *contractService) AdminList(req *dto.ContractListRequest) (*utils.Pagination, []dto.ContractInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.ContractAdminListOptions{
		CarID:        req.CarID,
		SellerID:     req.SellerID,
		BuyerID:      req.BuyerID,
		ContractType: req.ContractType,
		Status:       req.Status,
	}
	list, total, err := s.repo.AdminList(opts, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	infos := make([]dto.ContractInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toContractInfo(&list[i]))
	}
	return pagination, infos, nil
}

// AdminGetByID M 端查询
func (s *contractService) AdminGetByID(id uint) (*dto.ContractInfo, error) {
	c, err := s.repo.FindByID(id)
	if err != nil {
		return nil, ErrContractNotFound
	}
	return toContractInfo(c), nil
}

// UpdateStatus M 端更新状态
func (s *contractService) UpdateStatus(id uint, status int) error {
	if _, err := s.repo.FindByID(id); err != nil {
		return ErrContractNotFound
	}
	return s.repo.UpdateStatus(id, status, nil)
}
