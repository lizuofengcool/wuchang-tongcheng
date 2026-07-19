// Package service 数据统计 + 房贷 + 担保 + 成交 + 推荐 + VR 业务逻辑层
// 依据 v3.2.1 架构方案第五章：对标贝壳/链家
package service

import (
	"errors"
	"fmt"
	"math"
	"time"

	"wuchang-tongcheng/internal/modules/house/dto"
	"wuchang-tongcheng/internal/modules/house/model"
	"wuchang-tongcheng/internal/modules/house/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrStatNotFound        = errors.New("统计记录不存在")
	ErrMortgageNotFound    = errors.New("房贷方案不存在")
	ErrMortgageCodeExists  = errors.New("房贷方案编码已存在")
	ErrEscrowNotFound      = errors.New("担保交易不存在")
	ErrEscrowNoPermission  = errors.New("无权操作此担保交易")
	ErrEscrowStatusInvalid = errors.New("担保交易状态不允许此操作")
	ErrDealNotFound        = errors.New("成交记录不存在")
	ErrDealStatusInvalid   = errors.New("成交记录状态不允许此操作")
	ErrRecommendationNotFound = errors.New("推荐记录不存在")
	ErrVRTourNotFound      = errors.New("VR 看房不存在")
	ErrVRNoPermission      = errors.New("无权操作此 VR 看房")
)

// ===== 统计 =====

// StatisticService 数据统计业务接口
type StatisticService interface {
	GetOverview(regionID uint) (*dto.OverviewResponse, error)
	List(req *dto.StatListQuery) (*utils.Pagination, []dto.StatResponse, error)
	ListByDateRange(start, end interface{}, statType string, targetID uint) ([]dto.PriceTrendResponse, error)
	GetByID(id uint) (*dto.StatResponse, error)
	Upsert(s *model.HouseStatistic) error
	ListByType(statType string, page, pageSize int) (*utils.Pagination, []dto.StatResponse, error)
}

type statisticService struct {
	repo        repository.StatisticRepository
	houseRepo   repository.HouseRepository
	listingRepo repository.ListingRepository
	agentRepo   repository.AgentRepository
	communityRepo repository.CommunityRepository
	dealRepo    repository.DealRepository
	viewingRepo repository.ViewingRepository
	riskRepo    repository.RiskRepository
}

// NewStatisticService 创建统计 service 实例
func NewStatisticService(
	repo repository.StatisticRepository,
	houseRepo repository.HouseRepository,
	listingRepo repository.ListingRepository,
	agentRepo repository.AgentRepository,
	communityRepo repository.CommunityRepository,
	dealRepo repository.DealRepository,
	viewingRepo repository.ViewingRepository,
	riskRepo repository.RiskRepository,
) StatisticService {
	return &statisticService{
		repo:         repo,
		houseRepo:    houseRepo,
		listingRepo:  listingRepo,
		agentRepo:    agentRepo,
		communityRepo: communityRepo,
		dealRepo:     dealRepo,
		viewingRepo:  viewingRepo,
		riskRepo:     riskRepo,
	}
}

// toStatInfo model -> dto
func toStatInfo(s *model.HouseStatistic) *dto.StatResponse {
	return &dto.StatResponse{
		ID:              s.ID,
		RegionID:        s.RegionID,
		StatDate:        s.StatDate,
		StatType:        s.StatType,
		TargetID:        s.TargetID,
		TargetName:      s.TargetName,
		ImpressionCount: s.ImpressionCount,
		ClickCount:      s.ClickCount,
		FavCount:        s.FavCount,
		ContactCount:    s.ContactCount,
		ViewingCount:    s.ViewingCount,
		DealCount:       s.DealCount,
		ConversionRate:  s.ConversionRate,
		AvgSalePrice:    s.AvgSalePrice,
		AvgRentPrice:    s.AvgRentPrice,
		AvgDealDays:     s.AvgDealDays,
		CreatedAt:       s.CreatedAt,
		UpdatedAt:       s.UpdatedAt,
	}
}

// GetOverview 平台总览
func (s *statisticService) GetOverview(regionID uint) (*dto.OverviewResponse, error) {
	overview := &dto.OverviewResponse{}
	// 各仓储统计总数：本 MVP 实现简化版，依赖各仓储 List 计数
	pagination := utils.NewPagination(1, 1)

	if houses, total, err := s.houseRepo.AdminList(pagination, repository.HouseAdminListOptions{RegionID: regionID}); err == nil {
		overview.TotalHouses = total
		_ = houses
	}
	if listings, total, err := s.listingRepo.AdminList(pagination, repository.ListingAdminListOptions{RegionID: regionID}); err == nil {
		overview.TotalListings = total
		_ = listings
	}
	if agents, total, err := s.agentRepo.AdminList(pagination, repository.AgentAdminListOptions{RegionID: regionID}); err == nil {
		overview.TotalAgents = total
		_ = agents
	}
	if communities, total, err := s.communityRepo.AdminList(pagination, repository.CommunityAdminListOptions{RegionID: regionID}); err == nil {
		overview.TotalCommunities = total
		_ = communities
	}
	if deals, total, err := s.dealRepo.AdminList(pagination, repository.DealAdminListOptions{RegionID: regionID}); err == nil {
		overview.TotalDeals = total
		_ = deals
	}
	if viewings, total, err := s.viewingRepo.AdminList(pagination, repository.ViewingAdminListOptions{RegionID: regionID}); err == nil {
		overview.TotalViewings = total
		_ = viewings
	}
	// 待处理举报数
	if pendingReports, err := s.riskRepo.CountPendingReports(); err == nil {
		overview.PendingReports = pendingReports
	}

	return overview, nil
}

func (s *statisticService) List(req *dto.StatListQuery) (*utils.Pagination, []dto.StatResponse, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.StatListOptions{
		RegionID: req.RegionID,
		StatType: req.StatType,
		TargetID: req.TargetID,
	}
	if !req.StartDate.IsZero() {
		opts.StartDate = req.StartDate.Format("2006-01-02")
	}
	if !req.EndDate.IsZero() {
		opts.EndDate = req.EndDate.Format("2006-01-02")
	}

	list, total, err := s.repo.List(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.StatResponse, 0, len(list))
	for i := range list {
		result = append(result, *toStatInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListByDateRange 价格趋势
func (s *statisticService) ListByDateRange(start, end interface{}, statType string, targetID uint) ([]dto.PriceTrendResponse, error) {
	list, err := s.repo.ListByDateRange(start, end, statType, targetID)
	if err != nil {
		return nil, err
	}
	result := make([]dto.PriceTrendResponse, 0, len(list))
	for i := range list {
		result = append(result, dto.PriceTrendResponse{
			StatDate:     list[i].StatDate,
			AvgSalePrice: list[i].AvgSalePrice,
			AvgRentPrice: list[i].AvgRentPrice,
			DealCount:    list[i].DealCount,
		})
	}
	return result, nil
}

func (s *statisticService) GetByID(id uint) (*dto.StatResponse, error) {
	r, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrStatNotFound
		}
		return nil, err
	}
	return toStatInfo(r), nil
}

func (s *statisticService) Upsert(stat *model.HouseStatistic) error {
	return s.repo.UpsertByDateTypeTarget(stat)
}

func (s *statisticService) ListByType(statType string, page, pageSize int) (*utils.Pagination, []dto.StatResponse, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByType(statType, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.StatResponse, 0, len(list))
	for i := range list {
		result = append(result, *toStatInfo(&list[i]))
	}
	return pagination, result, nil
}

// ===== 房贷 =====

// MortgageService 房贷业务接口
type MortgageService interface {
	Create(req *dto.MortgageCreateRequest) (*dto.MortgageResponse, error)
	Update(id uint, req *dto.MortgageCreateRequest) error
	Delete(id uint) error
	GetByID(id uint) (*dto.MortgageResponse, error)
	List(req *dto.MortgageListQuery) (*utils.Pagination, []dto.MortgageResponse, error)
	ListAll() ([]dto.MortgageResponse, error)
	ListByLoanType(loanType string) ([]dto.MortgageResponse, error)
	ListHot(limit int) ([]dto.MortgageResponse, error)
	UpdateStatus(id uint, status int) error
	BatchUpdateStatus(ids []uint, status int) (*dto.BatchResultResponse, error)
	Calculate(req *dto.MortgageCalculateRequest) (*dto.MortgageCalculateResponse, error)
}

type mortgageService struct {
	repo repository.MortgageRepository
}

// NewMortgageService 创建房贷 service 实例
func NewMortgageService(repo repository.MortgageRepository) MortgageService {
	return &mortgageService{repo: repo}
}

// toMortgageInfo model -> dto
func toMortgageInfo(m *model.HouseMortgage) *dto.MortgageResponse {
	return &dto.MortgageResponse{
		ID:             m.ID,
		Name:           m.Name,
		Code:           m.Code,
		LoanType:       m.LoanType,
		LoanTypeText:   mortgageLoanTypeText(m.LoanType),
		MinDownPayment: m.MinDownPayment,
		MaxDownPayment: m.MaxDownPayment,
		InterestRate:   m.InterestRate,
		MaxPeriods:     m.MaxPeriods,
		MaxAmount:      m.MaxAmount,
		Description:    m.Description,
		Sort:           m.Sort,
		Status:         m.Status,
		IsHot:          m.IsHot,
		UseCount:       m.UseCount,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}

func mortgageLoanTypeText(t string) string {
	switch t {
	case model.MortgageTypeCommercial:
		return "商业贷款"
	case model.MortgageTypeProvidentFund:
		return "公积金贷款"
	case model.MortgageTypeCombined:
		return "组合贷款"
	case model.MortgageTypeFull:
		return "全款"
	}
	return ""
}

func mortgageStatusText(s int) string {
	switch s {
	case model.MortgageStatusDisabled:
		return "禁用"
	case model.MortgageStatusEnabled:
		return "启用"
	}
	return ""
}

func (s *mortgageService) Create(req *dto.MortgageCreateRequest) (*dto.MortgageResponse, error) {
	// 检查 code 唯一
	if existing, err := s.repo.FindByCode(req.Code); err == nil && existing != nil {
		return nil, ErrMortgageCodeExists
	}
	m := &model.HouseMortgage{
		Name:           req.Name,
		Code:           req.Code,
		LoanType:       req.LoanType,
		MinDownPayment: req.MinDownPayment,
		MaxDownPayment: req.MaxDownPayment,
		InterestRate:   req.InterestRate,
		MaxPeriods:     req.MaxPeriods,
		MaxAmount:      req.MaxAmount,
		Description:    req.Description,
		Sort:           req.Sort,
		Status:         req.Status,
		IsHot:          req.IsHot,
	}
	if m.LoanType == "" {
		m.LoanType = model.MortgageTypeCommercial
	}
	if m.MinDownPayment == 0 {
		m.MinDownPayment = 30
	}
	if m.MaxDownPayment == 0 {
		m.MaxDownPayment = 100
	}
	if m.MaxPeriods == 0 {
		m.MaxPeriods = 360
	}
	if m.Status == 0 {
		m.Status = model.MortgageStatusEnabled
	}
	if err := s.repo.Create(m); err != nil {
		return nil, err
	}
	return toMortgageInfo(m), nil
}

func (s *mortgageService) Update(id uint, req *dto.MortgageCreateRequest) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMortgageNotFound
		}
		return err
	}
	fields := map[string]interface{}{
		"name":             req.Name,
		"code":             req.Code,
		"loan_type":        req.LoanType,
		"min_down_payment": req.MinDownPayment,
		"max_down_payment": req.MaxDownPayment,
		"interest_rate":    req.InterestRate,
		"max_periods":      req.MaxPeriods,
		"max_amount":       req.MaxAmount,
		"description":      req.Description,
		"sort":             req.Sort,
		"status":           req.Status,
		"is_hot":           req.IsHot,
	}
	return s.repo.UpdateFields(id, fields)
}

func (s *mortgageService) Delete(id uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMortgageNotFound
		}
		return err
	}
	return s.repo.Delete(id)
}

func (s *mortgageService) GetByID(id uint) (*dto.MortgageResponse, error) {
	m, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMortgageNotFound
		}
		return nil, err
	}
	return toMortgageInfo(m), nil
}

func (s *mortgageService) List(req *dto.MortgageListQuery) (*utils.Pagination, []dto.MortgageResponse, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.MortgageListOptions{
		LoanType: req.LoanType,
		Status:   req.Status,
		IsHot:    req.IsHot,
	}
	list, total, err := s.repo.List(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.MortgageResponse, 0, len(list))
	for i := range list {
		result = append(result, *toMortgageInfo(&list[i]))
	}
	return pagination, result, nil
}

func (s *mortgageService) ListAll() ([]dto.MortgageResponse, error) {
	list, err := s.repo.ListAll()
	if err != nil {
		return nil, err
	}
	result := make([]dto.MortgageResponse, 0, len(list))
	for i := range list {
		result = append(result, *toMortgageInfo(&list[i]))
	}
	return result, nil
}

func (s *mortgageService) ListByLoanType(loanType string) ([]dto.MortgageResponse, error) {
	list, err := s.repo.ListByLoanType(loanType)
	if err != nil {
		return nil, err
	}
	result := make([]dto.MortgageResponse, 0, len(list))
	for i := range list {
		result = append(result, *toMortgageInfo(&list[i]))
	}
	return result, nil
}

func (s *mortgageService) ListHot(limit int) ([]dto.MortgageResponse, error) {
	list, err := s.repo.ListHot(limit)
	if err != nil {
		return nil, err
	}
	result := make([]dto.MortgageResponse, 0, len(list))
	for i := range list {
		result = append(result, *toMortgageInfo(&list[i]))
	}
	return result, nil
}

func (s *mortgageService) UpdateStatus(id uint, status int) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMortgageNotFound
		}
		return err
	}
	return s.repo.UpdateFields(id, map[string]interface{}{"status": status})
}

func (s *mortgageService) BatchUpdateStatus(ids []uint, status int) (*dto.BatchResultResponse, error) {
	affected, err := s.repo.BatchUpdateStatus(ids, status)
	if err != nil {
		return &dto.BatchResultResponse{
			Total: len(ids), Success: 0, Failed: len(ids), FailedIDs: ids,
		}, err
	}
	return &dto.BatchResultResponse{
		Total:   len(ids),
		Success: int(affected),
		Failed:  len(ids) - int(affected),
	}, nil
}

// Calculate 房贷计算（等额本息 + 等额本金首月）
func (s *mortgageService) Calculate(req *dto.MortgageCalculateRequest) (*dto.MortgageCalculateResponse, error) {
	if req.TotalPrice <= 0 || req.Periods <= 0 {
		return nil, errors.New("参数错误")
	}
	// 首付金额
	downPaymentAmount := req.TotalPrice * req.DownPayment / 100
	// 贷款金额
	loanAmount := req.TotalPrice - downPaymentAmount
	// 月利率
	monthlyRate := req.InterestRate / 12
	// 期数
	n := req.Periods

	// 等额本息月供：M = P * r * (1+r)^n / ((1+r)^n - 1)
	var monthlyPayment float64
	if monthlyRate > 0 {
		pow := math.Pow(1+monthlyRate, float64(n))
		monthlyPayment = loanAmount * monthlyRate * pow / (pow - 1)
	} else {
		monthlyPayment = loanAmount / float64(n)
	}
	// 还款总额
	totalPayment := monthlyPayment * float64(n)
	// 总利息
	totalInterest := totalPayment - loanAmount
	// 等额本金首月：首月本金 = loanAmount / n；首月利息 = loanAmount * monthlyRate
	monthlyPrincipal := loanAmount / float64(n)
	monthlyInterest := loanAmount * monthlyRate

	resp := &dto.MortgageCalculateResponse{
		LoanType:       req.LoanType,
		TotalPrice:     req.TotalPrice,
		DownPayment:    downPaymentAmount,
		DownPaymentPct: req.DownPayment,
		LoanAmount:     loanAmount,
		Periods:        req.Periods,
		InterestRate:   req.InterestRate,
		MonthlyPayment: monthlyPayment,
		TotalPayment:   totalPayment,
		TotalInterest:  totalInterest,
		MonthlyPrincipal: monthlyPrincipal,
		MonthlyInterest:  monthlyInterest,
	}
	return resp, nil
}

// ===== 担保交易 =====

// EscrowService 担保交易业务接口
type EscrowService interface {
	Create(regionID uint, payerID uint, req *dto.EscrowCreateRequest) (*dto.EscrowResponse, error)
	GetByID(id uint) (*dto.EscrowResponse, error)
	GetByNo(no string) (*dto.EscrowResponse, error)
	List(req *dto.EscrowListQuery) (*utils.Pagination, []dto.EscrowResponse, error)
	ListByPayer(payerID uint, page, pageSize int) (*utils.Pagination, []dto.EscrowResponse, error)
	ListByPayee(payeeID uint, page, pageSize int) (*utils.Pagination, []dto.EscrowResponse, error)
	ListByHouse(houseID uint, page, pageSize int) (*utils.Pagination, []dto.EscrowResponse, error)
	ListDisputed(page, pageSize int) (*utils.Pagination, []dto.EscrowResponse, error)
	MarkPaid(id uint, payMethod string, payTradeNo string) error
	Release(id uint) error
	Refund(id uint) error
	Dispute(id uint, userID uint, req *dto.EscrowDisputeRequest) error
	Arbitrate(id uint, handlerID uint, req *dto.EscrowArbitrateRequest) error
	Cancel(id uint, userID uint) error

	// M 端管理
	AdminList(req *dto.EscrowListQuery) (*utils.Pagination, []dto.EscrowResponse, error)
	BatchUpdateStatus(ids []uint, status int) (*dto.BatchResultResponse, error)

	// 定时任务
	ListToAutoRelease() ([]dto.EscrowResponse, error)
}

type escrowService struct {
	repo repository.EscrowRepository
}

// NewEscrowService 创建担保交易 service 实例
func NewEscrowService(repo repository.EscrowRepository) EscrowService {
	return &escrowService{repo: repo}
}

// toEscrowInfo model -> dto
func toEscrowInfo(e *model.HouseEscrow) *dto.EscrowResponse {
	return &dto.EscrowResponse{
		ID:                e.ID,
		EscrowNo:          e.EscrowNo,
		EscrowType:        e.EscrowType,
		HouseID:           e.HouseID,
		ListingID:         e.ListingID,
		ContractID:        e.ContractID,
		CommunityID:       e.CommunityID,
		PayerID:           e.PayerID,
		PayeeID:           e.PayeeID,
		AgentID:           e.AgentID,
		Amount:            e.Amount,
		PlatformFee:       e.PlatformFee,
		AgentFee:          e.AgentFee,
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

func escrowStatusText(s int) string {
	switch s {
	case model.EscrowStatusPending:
		return "待支付"
	case model.EscrowStatusPaid:
		return "已支付（冻结中）"
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

// generateEscrowNo 担保单号：ES yyyyMMddHHmmss . 000
func generateEscrowNo() string {
	return fmt.Sprintf("ES%s.%03d", time.Now().Format("20060102150405"), time.Now().Nanosecond()%1000)
}

func (s *escrowService) Create(regionID uint, payerID uint, req *dto.EscrowCreateRequest) (*dto.EscrowResponse, error) {
	e := &model.HouseEscrow{
		EscrowNo:    generateEscrowNo(),
		EscrowType:  req.EscrowType,
		HouseID:     req.HouseID,
		ListingID:   req.ListingID,
		ContractID:  req.ContractID,
		CommunityID: req.CommunityID,
		PayerID:     payerID,
		PayeeID:     req.PayeeID,
		AgentID:     req.AgentID,
		Amount:      req.Amount,
		PayMethod:   req.PayMethod,
		Status:      model.EscrowStatusPending,
	}
	e.RegionID = regionID
	if e.EscrowType == "" {
		e.EscrowType = model.EscrowTypeDeposit
	}
	if e.PayMethod == "" {
		e.PayMethod = model.EscrowPayWechat
	}
	if err := s.repo.Create(e); err != nil {
		return nil, err
	}
	return toEscrowInfo(e), nil
}

func (s *escrowService) GetByID(id uint) (*dto.EscrowResponse, error) {
	e, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEscrowNotFound
		}
		return nil, err
	}
	return toEscrowInfo(e), nil
}

func (s *escrowService) GetByNo(no string) (*dto.EscrowResponse, error) {
	e, err := s.repo.FindByEscrowNo(no)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEscrowNotFound
		}
		return nil, err
	}
	return toEscrowInfo(e), nil
}

func (s *escrowService) List(req *dto.EscrowListQuery) (*utils.Pagination, []dto.EscrowResponse, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.EscrowListOptions{
		EscrowType: req.EscrowType,
		HouseID:    req.HouseID,
		Status:     req.Status,
	}
	// payerID/payeeID/agentID 不在 EscrowListOptions 中，使用 ListByPayer/Payee 替代
	list, total, err := s.repo.List(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.EscrowResponse, 0, len(list))
	for i := range list {
		result = append(result, *toEscrowInfo(&list[i]))
	}
	return pagination, result, nil
}

func (s *escrowService) ListByPayer(payerID uint, page, pageSize int) (*utils.Pagination, []dto.EscrowResponse, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByPayer(payerID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.EscrowResponse, 0, len(list))
	for i := range list {
		result = append(result, *toEscrowInfo(&list[i]))
	}
	return pagination, result, nil
}

func (s *escrowService) ListByPayee(payeeID uint, page, pageSize int) (*utils.Pagination, []dto.EscrowResponse, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByPayee(payeeID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.EscrowResponse, 0, len(list))
	for i := range list {
		result = append(result, *toEscrowInfo(&list[i]))
	}
	return pagination, result, nil
}

func (s *escrowService) ListByHouse(houseID uint, page, pageSize int) (*utils.Pagination, []dto.EscrowResponse, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByHouse(houseID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.EscrowResponse, 0, len(list))
	for i := range list {
		result = append(result, *toEscrowInfo(&list[i]))
	}
	return pagination, result, nil
}

func (s *escrowService) ListDisputed(page, pageSize int) (*utils.Pagination, []dto.EscrowResponse, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListDisputed(pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.EscrowResponse, 0, len(list))
	for i := range list {
		result = append(result, *toEscrowInfo(&list[i]))
	}
	return pagination, result, nil
}

// MarkPaid 标记已支付（冻结中）
func (s *escrowService) MarkPaid(id uint, payMethod string, payTradeNo string) error {
	e, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrEscrowNotFound
		}
		return err
	}
	if e.Status != model.EscrowStatusPending {
		return ErrEscrowStatusInvalid
	}
	now := time.Now()
	// 自动放款时间默认 7 天后
	autoRelease := now.Add(7 * 24 * time.Hour)
	fields := map[string]interface{}{
		"status":          model.EscrowStatusPaid,
		"pay_method":      payMethod,
		"pay_trade_no":    payTradeNo,
		"paid_at":         &now,
		"frozen_at":       &now,
		"auto_release_at": &autoRelease,
	}
	return s.repo.UpdateFields(id, fields)
}

// Release 放款
func (s *escrowService) Release(id uint) error {
	e, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrEscrowNotFound
		}
		return err
	}
	if e.Status != model.EscrowStatusPaid && e.Status != model.EscrowStatusArbitrated {
		return ErrEscrowStatusInvalid
	}
	now := time.Now()
	fields := map[string]interface{}{
		"status":        model.EscrowStatusReleased,
		"release_at":    &now,
		"completed_at":  &now,
		"payee_amount":  e.Amount - e.PlatformFee - e.AgentFee,
	}
	return s.repo.UpdateFields(id, fields)
}

// Refund 退款
func (s *escrowService) Refund(id uint) error {
	e, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrEscrowNotFound
		}
		return err
	}
	if e.Status != model.EscrowStatusPaid && e.Status != model.EscrowStatusArbitrated {
		return ErrEscrowStatusInvalid
	}
	now := time.Now()
	return s.repo.UpdateFields(id, map[string]interface{}{
		"status":       model.EscrowStatusRefunded,
		"refunded_at":  &now,
		"completed_at": &now,
	})
}

// Dispute 发起争议（仅付款方/收款方可发起）
func (s *escrowService) Dispute(id uint, userID uint, req *dto.EscrowDisputeRequest) error {
	e, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrEscrowNotFound
		}
		return err
	}
	if e.PayerID != userID && e.PayeeID != userID {
		return ErrEscrowNoPermission
	}
	if e.Status != model.EscrowStatusPaid {
		return ErrEscrowStatusInvalid
	}
	return s.repo.UpdateFields(id, map[string]interface{}{
		"status":         model.EscrowStatusDisputed,
		"dispute_reason": req.Reason,
	})
}

// Arbitrate 仲裁（M 端）
func (s *escrowService) Arbitrate(id uint, handlerID uint, req *dto.EscrowArbitrateRequest) error {
	e, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrEscrowNotFound
		}
		return err
	}
	if e.Status != model.EscrowStatusDisputed {
		return ErrEscrowStatusInvalid
	}
	_ = handlerID
	now := time.Now()
	// 仲裁结果：to_payer=true 退款给付款方；to_payer=false 放款给收款方
	if req.ToPayer {
		return s.repo.UpdateFields(id, map[string]interface{}{
			"status":             model.EscrowStatusRefunded,
			"arbitration_result": req.Result,
			"refunded_at":        &now,
			"completed_at":       &now,
		})
	}
	return s.repo.UpdateFields(id, map[string]interface{}{
		"status":             model.EscrowStatusReleased,
		"arbitration_result": req.Result,
		"release_at":         &now,
		"completed_at":       &now,
		"payee_amount":       e.Amount - e.PlatformFee - e.AgentFee,
	})
}

// Cancel 取消担保（仅付款方在待支付状态可取消）
func (s *escrowService) Cancel(id uint, userID uint) error {
	e, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrEscrowNotFound
		}
		return err
	}
	if e.PayerID != userID {
		return ErrEscrowNoPermission
	}
	if e.Status != model.EscrowStatusPending {
		return ErrEscrowStatusInvalid
	}
	now := time.Now()
	return s.repo.UpdateFields(id, map[string]interface{}{
		"status":       model.EscrowStatusCanceled,
		"completed_at": &now,
	})
}

func (s *escrowService) AdminList(req *dto.EscrowListQuery) (*utils.Pagination, []dto.EscrowResponse, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.EscrowAdminListOptions{
		HouseID:    req.HouseID,
		PayerID:    req.PayerID,
		PayeeID:    req.PayeeID,
		AgentID:    req.AgentID,
		EscrowType: req.EscrowType,
		Status:     req.Status,
	}
	list, total, err := s.repo.AdminList(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.EscrowResponse, 0, len(list))
	for i := range list {
		result = append(result, *toEscrowInfo(&list[i]))
	}
	return pagination, result, nil
}

func (s *escrowService) BatchUpdateStatus(ids []uint, status int) (*dto.BatchResultResponse, error) {
	affected, err := s.repo.BatchUpdateStatus(ids, status)
	if err != nil {
		return &dto.BatchResultResponse{
			Total: len(ids), Success: 0, Failed: len(ids), FailedIDs: ids,
		}, err
	}
	return &dto.BatchResultResponse{
		Total:   len(ids),
		Success: int(affected),
		Failed:  len(ids) - int(affected),
	}, nil
}

// ListToAutoRelease 定时任务：自动放款列表
func (s *escrowService) ListToAutoRelease() ([]dto.EscrowResponse, error) {
	list, err := s.repo.ListToAutoRelease()
	if err != nil {
		return nil, err
	}
	result := make([]dto.EscrowResponse, 0, len(list))
	for i := range list {
		result = append(result, *toEscrowInfo(&list[i]))
	}
	return result, nil
}

// ===== 成交记录 =====

// DealService 成交记录业务接口
type DealService interface {
	Create(regionID uint, userID uint, req *dto.DealCreateRequest) (*dto.DealResponse, error)
	GetByID(id uint) (*dto.DealResponse, error)
	GetByNo(no string) (*dto.DealResponse, error)
	List(req *dto.DealListQuery) (*utils.Pagination, []dto.DealResponse, error)
	ListByHouse(houseID uint, page, pageSize int) (*utils.Pagination, []dto.DealResponse, error)
	ListBySeller(sellerID uint, page, pageSize int) (*utils.Pagination, []dto.DealResponse, error)
	ListByBuyer(buyerID uint, page, pageSize int) (*utils.Pagination, []dto.DealResponse, error)
	ListByAgent(agentID uint, page, pageSize int) (*utils.Pagination, []dto.DealResponse, error)
	ListByCommunity(communityID uint, page, pageSize int) (*utils.Pagination, []dto.DealResponse, error)
	Confirm(id uint, userID uint, req *dto.DealConfirmRequest) error
	Cancel(id uint, userID uint, req *dto.DealCancelRequest) error
	Complete(id uint) error

	// M 端管理
	AdminList(req *dto.DealListQuery) (*utils.Pagination, []dto.DealResponse, error)
	BatchUpdateStatus(ids []uint, status int) (*dto.BatchResultResponse, error)
}

type dealService struct {
	repo repository.DealRepository
}

// NewDealService 创建成交 service 实例
func NewDealService(repo repository.DealRepository) DealService {
	return &dealService{repo: repo}
}

// toDealInfo model -> dto
func toDealInfo(d *model.HouseDeal) *dto.DealResponse {
	return &dto.DealResponse{
		ID:             d.ID,
		DealNo:         d.DealNo,
		HouseID:        d.HouseID,
		ListingID:      d.ListingID,
		CommunityID:    d.CommunityID,
		ContractID:     d.ContractID,
		EscrowID:       d.EscrowID,
		DealType:       d.DealType,
		SellerID:       d.SellerID,
		SellerName:     d.SellerName,
		BuyerID:        d.BuyerID,
		BuyerName:      d.BuyerName,
		AgentID:        d.AgentID,
		AgentName:      d.AgentName,
		DealPrice:      d.DealPrice,
		AveragePrice:   d.AveragePrice,
		Commission:     d.Commission,
		DealDate:       d.DealDate,
		ListedAt:       d.ListedAt,
		DealDays:       d.DealDays,
		PaymentMethod:  d.PaymentMethod,
		LoanAmount:     d.LoanAmount,
		LoanPeriods:    d.LoanPeriods,
		Status:         d.Status,
		StatusText:     dealStatusText(d.Status),
		CompletedAt:    d.CompletedAt,
		CanceledAt:     d.CanceledAt,
		CanceledReason: d.CanceledReason,
		Remark:         d.Remark,
		RegionID:       d.RegionID,
		CreatedAt:      d.CreatedAt,
		UpdatedAt:      d.UpdatedAt,
	}
}

func dealStatusText(s int) string {
	switch s {
	case model.DealStatusPending:
		return "待确认"
	case model.DealStatusConfirmed:
		return "已确认"
	case model.DealStatusCompleted:
		return "已完成"
	case model.DealStatusCanceled:
		return "已取消"
	case model.DealStatusDisputed:
		return "争议中"
	}
	return ""
}

// generateDealNo 成交单号：DL yyyyMMddHHmmss . 000
func generateDealNo() string {
	return fmt.Sprintf("DL%s.%03d", time.Now().Format("20060102150405"), time.Now().Nanosecond()%1000)
}

func (s *dealService) Create(regionID uint, userID uint, req *dto.DealCreateRequest) (*dto.DealResponse, error) {
	d := &model.HouseDeal{
		DealNo:        generateDealNo(),
		HouseID:       req.HouseID,
		ListingID:     req.ListingID,
		CommunityID:   req.CommunityID,
		ContractID:    req.ContractID,
		EscrowID:      req.EscrowID,
		DealType:      req.DealType,
		SellerID:      req.SellerID,
		SellerName:    req.SellerName,
		BuyerID:       req.BuyerID,
		BuyerName:     req.BuyerName,
		AgentID:       req.AgentID,
		AgentName:     req.AgentName,
		DealPrice:     req.DealPrice,
		DealDate:      req.DealDate,
		PaymentMethod: req.PaymentMethod,
		LoanAmount:    req.LoanAmount,
		LoanPeriods:   req.LoanPeriods,
		Remark:        req.Remark,
		Status:        model.DealStatusPending,
	}
	d.RegionID = regionID
	if d.DealType == "" {
		d.DealType = model.DealTypeSale
	}
	if d.DealDate == nil {
		now := time.Now()
		d.DealDate = &now
	}
	// 计算成交周期：根据 listed_at 与 deal_date
	_ = userID
	if err := s.repo.Create(d); err != nil {
		return nil, err
	}
	return toDealInfo(d), nil
}

func (s *dealService) GetByID(id uint) (*dto.DealResponse, error) {
	d, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDealNotFound
		}
		return nil, err
	}
	return toDealInfo(d), nil
}

func (s *dealService) GetByNo(no string) (*dto.DealResponse, error) {
	d, err := s.repo.FindByDealNo(no)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDealNotFound
		}
		return nil, err
	}
	return toDealInfo(d), nil
}

func (s *dealService) List(req *dto.DealListQuery) (*utils.Pagination, []dto.DealResponse, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.DealListOptions{
		DealType: req.DealType,
		HouseID:  req.HouseID,
		Status:   req.Status,
	}
	list, total, err := s.repo.List(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.DealResponse, 0, len(list))
	for i := range list {
		result = append(result, *toDealInfo(&list[i]))
	}
	return pagination, result, nil
}

func (s *dealService) ListByHouse(houseID uint, page, pageSize int) (*utils.Pagination, []dto.DealResponse, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByHouse(houseID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.DealResponse, 0, len(list))
	for i := range list {
		result = append(result, *toDealInfo(&list[i]))
	}
	return pagination, result, nil
}

func (s *dealService) ListBySeller(sellerID uint, page, pageSize int) (*utils.Pagination, []dto.DealResponse, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListBySeller(sellerID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.DealResponse, 0, len(list))
	for i := range list {
		result = append(result, *toDealInfo(&list[i]))
	}
	return pagination, result, nil
}

func (s *dealService) ListByBuyer(buyerID uint, page, pageSize int) (*utils.Pagination, []dto.DealResponse, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByBuyer(buyerID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.DealResponse, 0, len(list))
	for i := range list {
		result = append(result, *toDealInfo(&list[i]))
	}
	return pagination, result, nil
}

func (s *dealService) ListByAgent(agentID uint, page, pageSize int) (*utils.Pagination, []dto.DealResponse, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByAgent(agentID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.DealResponse, 0, len(list))
	for i := range list {
		result = append(result, *toDealInfo(&list[i]))
	}
	return pagination, result, nil
}

func (s *dealService) ListByCommunity(communityID uint, page, pageSize int) (*utils.Pagination, []dto.DealResponse, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByCommunity(communityID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.DealResponse, 0, len(list))
	for i := range list {
		result = append(result, *toDealInfo(&list[i]))
	}
	return pagination, result, nil
}

// Confirm 确认成交：买卖双方均可确认
func (s *dealService) Confirm(id uint, userID uint, req *dto.DealConfirmRequest) error {
	d, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrDealNotFound
		}
		return err
	}
	if d.Status != model.DealStatusPending {
		return ErrDealStatusInvalid
	}
	_ = userID
	fields := map[string]interface{}{
		"status": model.DealStatusConfirmed,
	}
	if req.Remark != "" {
		fields["remark"] = req.Remark
	}
	return s.repo.UpdateFields(id, fields)
}

// Cancel 取消成交
func (s *dealService) Cancel(id uint, userID uint, req *dto.DealCancelRequest) error {
	d, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrDealNotFound
		}
		return err
	}
	if d.Status == model.DealStatusCompleted || d.Status == model.DealStatusCanceled {
		return ErrDealStatusInvalid
	}
	_ = userID
	now := time.Now()
	return s.repo.UpdateFields(id, map[string]interface{}{
		"status":          model.DealStatusCanceled,
		"canceled_at":     &now,
		"canceled_reason": req.Reason,
	})
}

// Complete 完成成交
func (s *dealService) Complete(id uint) error {
	d, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrDealNotFound
		}
		return err
	}
	if d.Status != model.DealStatusConfirmed {
		return ErrDealStatusInvalid
	}
	now := time.Now()
	return s.repo.UpdateFields(id, map[string]interface{}{
		"status":       model.DealStatusCompleted,
		"completed_at": &now,
	})
}

func (s *dealService) AdminList(req *dto.DealListQuery) (*utils.Pagination, []dto.DealResponse, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.DealAdminListOptions{
		HouseID:     req.HouseID,
		SellerID:    req.SellerID,
		BuyerID:     req.BuyerID,
		AgentID:     req.AgentID,
		CommunityID: req.CommunityID,
		DealType:    req.DealType,
		Status:      req.Status,
	}
	list, total, err := s.repo.AdminList(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.DealResponse, 0, len(list))
	for i := range list {
		result = append(result, *toDealInfo(&list[i]))
	}
	return pagination, result, nil
}

func (s *dealService) BatchUpdateStatus(ids []uint, status int) (*dto.BatchResultResponse, error) {
	affected, err := s.repo.BatchUpdateStatus(ids, status)
	if err != nil {
		return &dto.BatchResultResponse{
			Total: len(ids), Success: 0, Failed: len(ids), FailedIDs: ids,
		}, err
	}
	return &dto.BatchResultResponse{
		Total:   len(ids),
		Success: int(affected),
		Failed:  len(ids) - int(affected),
	}, nil
}

// ===== 推荐 =====

// RecommendationService 推荐业务接口
type RecommendationService interface {
	GetByID(id uint) (*dto.RecommendationResponse, error)
	List(req *dto.RecommendationListQuery) (*utils.Pagination, []dto.RecommendationResponse, error)
	ListByUser(userID uint, page, pageSize int) (*utils.Pagination, []dto.RecommendationResponse, error)
	MarkClicked(id uint, userID uint) error
	MarkContacted(id uint, userID uint) error
	MarkViewed(id uint, userID uint) error
	MarkDismissed(id uint, userID uint) error

	// M 端管理
	Create(rec *model.HouseRecommendation) error
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error
}

type recommendationService struct {
	repo repository.InteractionRepository
}

// NewRecommendationService 创建推荐 service 实例
func NewRecommendationService(repo repository.InteractionRepository) RecommendationService {
	return &recommendationService{repo: repo}
}

// toRecInfo model -> dto
func toRecInfo(r *model.HouseRecommendation) *dto.RecommendationResponse {
	return &dto.RecommendationResponse{
		ID:            r.ID,
		UserID:        r.UserID,
		HouseID:       r.HouseID,
		RecType:       r.RecType,
		Source:        r.Source,
		Score:         r.Score,
		Reason:        r.Reason,
		PriceMatch:    r.PriceMatch,
		LocationMatch: r.LocationMatch,
		LayoutMatch:   r.LayoutMatch,
		FacilityMatch: r.FacilityMatch,
		Status:        r.Status,
		StatusText:    recStatusText(r.Status),
		ClickedAt:     r.ClickedAt,
		ContactedAt:   r.ContactedAt,
		ViewedAt:      r.ViewedAt,
		DismissedAt:   r.DismissedAt,
		ExpiredAt:     r.ExpiredAt,
		CreatedAt:     r.CreatedAt,
	}
}

func recStatusText(s int) string {
	switch s {
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

func (s *recommendationService) GetByID(id uint) (*dto.RecommendationResponse, error) {
	r, err := s.repo.FindRecByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRecommendationNotFound
		}
		return nil, err
	}
	return toRecInfo(r), nil
}

func (s *recommendationService) List(req *dto.RecommendationListQuery) (*utils.Pagination, []dto.RecommendationResponse, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.RecListOptions{
		UserID:  req.UserID,
		RecType: req.RecType,
		Source:  req.Source,
		Status:  req.Status,
	}
	list, total, err := s.repo.ListRec(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.RecommendationResponse, 0, len(list))
	for i := range list {
		result = append(result, *toRecInfo(&list[i]))
	}
	return pagination, result, nil
}

func (s *recommendationService) ListByUser(userID uint, page, pageSize int) (*utils.Pagination, []dto.RecommendationResponse, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListRecByUser(userID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.RecommendationResponse, 0, len(list))
	for i := range list {
		result = append(result, *toRecInfo(&list[i]))
	}
	return pagination, result, nil
}

func (s *recommendationService) MarkClicked(id uint, userID uint) error {
	r, err := s.repo.FindRecByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRecommendationNotFound
		}
		return err
	}
	if r.UserID != userID {
		return ErrVRNoPermission
	}
	now := time.Now()
	return s.repo.UpdateRecFields(id, map[string]interface{}{
		"status":     model.RecStatusClicked,
		"clicked_at": &now,
	})
}

func (s *recommendationService) MarkContacted(id uint, userID uint) error {
	r, err := s.repo.FindRecByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRecommendationNotFound
		}
		return err
	}
	if r.UserID != userID {
		return ErrVRNoPermission
	}
	now := time.Now()
	return s.repo.UpdateRecFields(id, map[string]interface{}{
		"status":       model.RecStatusContacted,
		"contacted_at": &now,
	})
}

func (s *recommendationService) MarkViewed(id uint, userID uint) error {
	r, err := s.repo.FindRecByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRecommendationNotFound
		}
		return err
	}
	if r.UserID != userID {
		return ErrVRNoPermission
	}
	now := time.Now()
	return s.repo.UpdateRecFields(id, map[string]interface{}{
		"status":    model.RecStatusShown,
		"viewed_at": &now,
	})
}

func (s *recommendationService) MarkDismissed(id uint, userID uint) error {
	r, err := s.repo.FindRecByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRecommendationNotFound
		}
		return err
	}
	if r.UserID != userID {
		return ErrVRNoPermission
	}
	now := time.Now()
	return s.repo.UpdateRecFields(id, map[string]interface{}{
		"status":       model.RecStatusDismissed,
		"dismissed_at": &now,
	})
}

func (s *recommendationService) Create(rec *model.HouseRecommendation) error {
	return s.repo.CreateRec(rec)
}

func (s *recommendationService) Update(id uint, fields map[string]interface{}) error {
	if _, err := s.repo.FindRecByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRecommendationNotFound
		}
		return err
	}
	return s.repo.UpdateRecFields(id, fields)
}

func (s *recommendationService) Delete(id uint) error {
	if _, err := s.repo.FindRecByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRecommendationNotFound
		}
		return err
	}
	return s.repo.DeleteRec(id)
}

// ===== VR 看房 =====

// VRTourService VR 看房业务接口
type VRTourService interface {
	Create(creatorID uint, creatorName string, req *dto.VRTourCreateRequest) (*dto.VRTourResponse, error)
	Update(id uint, userID uint, req *dto.VRTourCreateRequest) error
	Delete(id uint, userID uint) error
	GetByID(id uint) (*dto.VRTourResponse, error)
	GetByNo(no string) (*dto.VRTourResponse, error)
	List(req *dto.VRTourListQuery) (*utils.Pagination, []dto.VRTourResponse, error)
	Publish(id uint, userID uint) error
	Offline(id uint, userID uint) error
	IncrViewCount(id uint) error
	IncrShareCount(id uint) error
}

type vrTourService struct {
	repo repository.InteractionRepository
}

// NewVRTourService 创建 VR service 实例
func NewVRTourService(repo repository.InteractionRepository) VRTourService {
	return &vrTourService{repo: repo}
}

// toVRInfo model -> dto
func toVRInfo(v *model.HouseVRTour) *dto.VRTourResponse {
	resp := &dto.VRTourResponse{
		ID:              v.ID,
		HouseID:         v.HouseID,
		ListingID:       v.ListingID,
		CommunityID:     v.CommunityID,
		VRNo:            v.VRNo,
		Title:           v.Title,
		Description:     v.Description,
		VRType:          v.VRType,
		VRTypeText:      vrTypeText(v.VRType),
		VRURL:           v.VRURL,
		CoverImage:      v.CoverImage,
		DurationSeconds: v.DurationSeconds,
		ViewCount:       v.ViewCount,
		ShareCount:      v.ShareCount,
		Status:          v.Status,
		StatusText:      vrStatusText(v.Status),
		RecorderID:      v.RecorderID,
		RecorderName:    v.RecorderName,
		RecordedAt:      v.RecordedAt,
		PublishedAt:     v.PublishedAt,
		OfflineAt:       v.OfflineAt,
		Equipment:       v.Equipment,
		Resolution:      v.Resolution,
		FileSize:        v.FileSize,
		CreatedAt:       v.CreatedAt,
		UpdatedAt:       v.UpdatedAt,
	}
	// 反序列化场景 JSON
	if len(v.Scenes) > 0 {
		var scenes []model.VRScene
		_ = v.Scenes.Parse(&scenes)
		resp.Scenes = scenes
	}
	return resp
}

func vrTypeText(t string) string {
	switch t {
	case model.VRTypePanorama:
		return "360°全景"
	case model.VRTypeVR:
		return "VR 看房"
	case model.VRTypeVideo:
		return "视频看房"
	case model.VRType3D:
		return "3D 模型"
	}
	return ""
}

func vrStatusText(s int) string {
	switch s {
	case model.VRStatusDraft:
		return "草稿"
	case model.VRStatusPublished:
		return "已发布"
	case model.VRStatusOffline:
		return "已下架"
	case model.VRStatusRejected:
		return "已拒绝"
	}
	return ""
}

// generateVRNo VR 单号：VR yyyyMMddHHmmss . 000
func generateVRNo() string {
	return fmt.Sprintf("VR%s.%03d", time.Now().Format("20060102150405"), time.Now().Nanosecond()%1000)
}

func (s *vrTourService) Create(creatorID uint, creatorName string, req *dto.VRTourCreateRequest) (*dto.VRTourResponse, error) {
	v := &model.HouseVRTour{
		HouseID:         req.HouseID,
		ListingID:       req.ListingID,
		CommunityID:     req.CommunityID,
		VRNo:            generateVRNo(),
		Title:           req.Title,
		Description:     req.Description,
		VRType:          req.VRType,
		VRURL:           req.VRURL,
		CoverImage:      req.CoverImage,
		DurationSeconds: req.DurationSeconds,
		RecorderID:      req.RecorderID,
		RecorderName:    req.RecorderName,
		Equipment:       req.Equipment,
		Resolution:      req.Resolution,
		FileSize:        req.FileSize,
		Status:          req.Status,
	}
	if v.VRType == "" {
		v.VRType = model.VRTypePanorama
	}
	if v.RecorderID == 0 {
		v.RecorderID = creatorID
	}
	if v.RecorderName == "" {
		v.RecorderName = creatorName
	}
	// 场景 JSON
	if len(req.Scenes) > 0 {
		b, err := model.FromJSON(req.Scenes)
		if err == nil {
			v.Scenes = b
		}
	}
	if v.Status == 0 {
		v.Status = model.VRStatusDraft
	}
	if v.Status == model.VRStatusPublished {
		now := time.Now()
		v.PublishedAt = &now
	}
	if err := s.repo.CreateVR(v); err != nil {
		return nil, err
	}
	return toVRInfo(v), nil
}

func (s *vrTourService) Update(id uint, userID uint, req *dto.VRTourCreateRequest) error {
	v, err := s.repo.FindVRByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrVRTourNotFound
		}
		return err
	}
	// 仅录制人本人可修改（MVP 简化）
	if v.RecorderID != 0 && v.RecorderID != userID {
		return ErrVRNoPermission
	}
	fields := map[string]interface{}{
		"house_id":         req.HouseID,
		"listing_id":       req.ListingID,
		"community_id":     req.CommunityID,
		"title":            req.Title,
		"description":      req.Description,
		"vr_type":          req.VRType,
		"vr_url":           req.VRURL,
		"cover_image":      req.CoverImage,
		"duration_seconds": req.DurationSeconds,
		"recorder_id":      req.RecorderID,
		"recorder_name":    req.RecorderName,
		"equipment":        req.Equipment,
		"resolution":       req.Resolution,
		"file_size":        req.FileSize,
	}
	if len(req.Scenes) > 0 {
		b, err := model.FromJSON(req.Scenes)
		if err == nil {
			fields["scenes"] = b
		}
	}
	return s.repo.UpdateVRFields(id, fields)
}

func (s *vrTourService) Delete(id uint, userID uint) error {
	v, err := s.repo.FindVRByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrVRTourNotFound
		}
		return err
	}
	if v.RecorderID != 0 && v.RecorderID != userID {
		return ErrVRNoPermission
	}
	return s.repo.DeleteVR(id)
}

func (s *vrTourService) GetByID(id uint) (*dto.VRTourResponse, error) {
	v, err := s.repo.FindVRByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrVRTourNotFound
		}
		return nil, err
	}
	return toVRInfo(v), nil
}

func (s *vrTourService) GetByNo(no string) (*dto.VRTourResponse, error) {
	v, err := s.repo.FindVRByVRNo(no)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrVRTourNotFound
		}
		return nil, err
	}
	return toVRInfo(v), nil
}

func (s *vrTourService) List(req *dto.VRTourListQuery) (*utils.Pagination, []dto.VRTourResponse, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.VRListOptions{
		HouseID:     req.HouseID,
		ListingID:   req.ListingID,
		CommunityID: req.CommunityID,
		VRType:      req.VRType,
		Status:      req.Status,
	}
	list, total, err := s.repo.ListVR(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.VRTourResponse, 0, len(list))
	for i := range list {
		result = append(result, *toVRInfo(&list[i]))
	}
	return pagination, result, nil
}

// Publish 发布 VR
func (s *vrTourService) Publish(id uint, userID uint) error {
	v, err := s.repo.FindVRByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrVRTourNotFound
		}
		return err
	}
	if v.RecorderID != 0 && v.RecorderID != userID {
		return ErrVRNoPermission
	}
	now := time.Now()
	return s.repo.UpdateVRFields(id, map[string]interface{}{
		"status":       model.VRStatusPublished,
		"published_at": &now,
	})
}

// Offline 下架 VR
func (s *vrTourService) Offline(id uint, userID uint) error {
	v, err := s.repo.FindVRByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrVRTourNotFound
		}
		return err
	}
	if v.RecorderID != 0 && v.RecorderID != userID {
		return ErrVRNoPermission
	}
	now := time.Now()
	return s.repo.UpdateVRFields(id, map[string]interface{}{
		"status":     model.VRStatusOffline,
		"offline_at": &now,
	})
}

func (s *vrTourService) IncrViewCount(id uint) error {
	return s.repo.IncrVRViewCount(id)
}

func (s *vrTourService) IncrShareCount(id uint) error {
	return s.repo.IncrVRShareCount(id)
}

// 兼容未使用的 mortgageStatusText，避免编译告警
var _ = mortgageStatusText
