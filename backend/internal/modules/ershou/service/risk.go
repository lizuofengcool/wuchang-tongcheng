// Package service 举报/评价/审核规则/用户信用业务逻辑层
// 依据 v3.2.1 架构方案：对标转转
package service

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"wuchang-tongcheng/internal/modules/ershou/dto"
	"wuchang-tongcheng/internal/modules/ershou/model"
	"wuchang-tongcheng/internal/modules/ershou/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrReportNotFound     = errors.New("举报记录不存在")
	ErrReportExists       = errors.New("已举报过此物品")
	ErrReportStatus       = errors.New("举报状态不允许此操作")
	ErrReviewNotFound     = errors.New("评价不存在")
	ErrReviewExists       = errors.New("已评价过此订单")
	ErrReviewNoPermission = errors.New("无权操作此评价")
	ErrAuditRuleNotFound  = errors.New("审核规则不存在")
	ErrUserCreditNotFound = errors.New("用户信用记录不存在")
)

// ===== ReportService =====

// ReportService 举报业务接口
type ReportService interface {
	Create(reporterID uint, reporterName string, req *dto.ReportCreateRequest) (*dto.ReportResponse, error)
	GetByID(reportID uint) (*dto.ReportResponse, error)
	List(query dto.ReportListQuery) (*utils.Pagination, []dto.ReportResponse, error)
	ListByErshouID(ershouID uint) ([]dto.ReportResponse, error)
	Process(reportID, handlerID uint, handlerName string, req *dto.ReportProcessRequest) (*dto.ReportResponse, error)
}

type reportService struct {
	repo        repository.ReportRepository
	ershouRepo  repository.ErshouRepository
	creditRepo  repository.UserCreditRepository
}

// NewReportService 创建举报 service 实例
func NewReportService(
	repo repository.ReportRepository,
	ershouRepo repository.ErshouRepository,
	creditRepo repository.UserCreditRepository,
) ReportService {
	return &reportService{repo: repo, ershouRepo: ershouRepo, creditRepo: creditRepo}
}

func reportStatusText(status int) string {
	switch status {
	case model.ReportStatusPending:
		return "待处理"
	case model.ReportStatusProcessing:
		return "处理中"
	case model.ReportStatusResolved:
		return "已处理（成立）"
	case model.ReportStatusRejected:
		return "已处理（不成立）"
	case model.ReportStatusAppealed:
		return "申诉中"
	case model.ReportStatusClosed:
		return "已关闭"
	}
	return "未知"
}

// genReportNo 生成举报单号：RP + yyyyMMddHHmmss + 6位随机数
func genReportNo() string {
	return fmt.Sprintf("RP%s%06d", time.Now().Format("20060102150405"), rand.Intn(1000000))
}

// Create 创建举报
func (s *reportService) Create(reporterID uint, reporterName string, req *dto.ReportCreateRequest) (*dto.ReportResponse, error) {
	e, err := s.ershouRepo.FindByID(req.ErshouID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrErshouNotFound
		}
		return nil, err
	}
	if e.UserID == reporterID {
		return nil, errors.New("不能举报自己的物品")
	}

	now := time.Now()
	slaDeadline := now.Add(72 * time.Hour) // SLA 72h 处理

	report := &model.ErshouReport{
		ReportNo:         genReportNo(),
		ErshouID:         req.ErshouID,
		ReporterID:       reporterID,
		ReporterName:     reporterName,
		ReportedUserID:   e.UserID,
		ReportedUserName: e.UserName,
		ReportType:       req.ReportType,
		Reason:           req.Reason,
		Description:      req.Description,
		Status:           model.ReportStatusPending,
		SLADeadline:      &slaDeadline,
	}
	if len(req.EvidenceImages) > 0 {
		if jb, err := model.FromJSON(req.EvidenceImages); err == nil {
			report.EvidenceImages = jb
		}
	}
	if err := s.repo.Create(report); err != nil {
		return nil, err
	}

	// 累计被举报人信用
	_ = s.creditRepo.IncrReports(e.UserID, 1)

	return s.toReportResponse(report), nil
}

func (s *reportService) GetByID(reportID uint) (*dto.ReportResponse, error) {
	r, err := s.repo.FindByID(reportID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrReportNotFound
		}
		return nil, err
	}
	return s.toReportResponse(r), nil
}

func (s *reportService) List(query dto.ReportListQuery) (*utils.Pagination, []dto.ReportResponse, error) {
	pagination := utils.NewPagination(query.Page, query.PageSize)
	list, total, err := s.repo.List(repository.ReportListQuery{
		Status:     query.Status,
		ReportType: query.ReportType,
		ErshouID:   query.ErshouID,
		ReporterID: query.ReporterID,
	}, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.ReportResponse, 0, len(list))
	for i := range list {
		result = append(result, *s.toReportResponse(&list[i]))
	}
	return pagination, result, nil
}

func (s *reportService) ListByErshouID(ershouID uint) ([]dto.ReportResponse, error) {
	list, err := s.repo.ListByErshouID(ershouID)
	if err != nil {
		return nil, err
	}
	result := make([]dto.ReportResponse, 0, len(list))
	for i := range list {
		result = append(result, *s.toReportResponse(&list[i]))
	}
	return result, nil
}

// Process 处理举报（M端审核员操作）
// Status: 1处理中 / 2已处理（成立）/ 3已处理（不成立）
func (s *reportService) Process(reportID, handlerID uint, handlerName string, req *dto.ReportProcessRequest) (*dto.ReportResponse, error) {
	r, err := s.repo.FindByID(reportID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrReportNotFound
		}
		return nil, err
	}
	if r.Status != model.ReportStatusPending && r.Status != model.ReportStatusProcessing {
		return nil, ErrReportStatus
	}

	now := time.Now()
	fields := map[string]interface{}{
		"status":        req.Status,
		"handler_id":    handlerID,
		"handler_name":  handlerName,
		"handle_result": req.HandleResult,
		"handled_at":    &now,
	}
	if req.PenaltyType != "" {
		fields["penalty_type"] = req.PenaltyType
		fields["penalty_user_id"] = r.ReportedUserID
	}

	if err := s.repo.Update(reportID, fields); err != nil {
		return nil, err
	}

	// 举报成立 → 累计被举报人处罚次数（扣 20 信用分）
	if req.Status == model.ReportStatusResolved && req.PenaltyType != "" {
		scoreDelta := -20
		switch req.PenaltyType {
		case "ban1d":
			scoreDelta = -30
		case "ban7d":
			scoreDelta = -50
		case "banForever":
			scoreDelta = -100
		case "limit":
			scoreDelta = -15
		case "warning":
			scoreDelta = -5
		}
		_ = s.creditRepo.IncrPenalties(r.ReportedUserID, 1, scoreDelta)
	}

	updated, _ := s.repo.FindByID(reportID)
	return s.toReportResponse(updated), nil
}

func (s *reportService) toReportResponse(r *model.ErshouReport) *dto.ReportResponse {
	resp := &dto.ReportResponse{
		ID:               r.ID,
		ReportNo:         r.ReportNo,
		ErshouID:         r.ErshouID,
		ReporterID:       r.ReporterID,
		ReporterName:     r.ReporterName,
		ReportedUserID:   r.ReportedUserID,
		ReportedUserName: r.ReportedUserName,
		ReportType:       r.ReportType,
		Reason:           r.Reason,
		Description:      r.Description,
		EvidenceImages:   []string{},
		Status:           r.Status,
		StatusText:       reportStatusText(r.Status),
		HandlerID:        r.HandlerID,
		HandlerName:      r.HandlerName,
		HandleResult:     r.HandleResult,
		PenaltyType:      r.PenaltyType,
		PenaltyUserID:    r.PenaltyUserID,
		SLADeadline:      r.SLADeadline,
		HandledAt:        r.HandledAt,
		CreatedAt:        r.CreatedAt,
	}
	if r.EvidenceImages != nil {
		var imgs []string
		_ = r.EvidenceImages.Parse(&imgs)
		if imgs != nil {
			resp.EvidenceImages = imgs
		}
	}
	return resp
}

// ===== ReviewService =====

// ReviewService 评价业务接口
type ReviewService interface {
	Create(orderID, reviewerID uint, reviewerName, reviewerAvatar string, req *dto.ReviewCreateRequest) (*dto.ReviewResponse, error)
	GetByID(reviewID uint) (*dto.ReviewResponse, error)
	List(query dto.ReviewListQuery) (*utils.Pagination, []dto.ReviewResponse, error)
	ListByErshouID(ershouID uint, pagination *utils.Pagination) (*utils.Pagination, []dto.ReviewResponse, error)
	Reply(reviewID, revieweeID uint, req *dto.ReviewReplyRequest) (*dto.ReviewResponse, error)
	StatsByErshouID(ershouID uint) (*dto.ReviewStats, error)
}

type reviewService struct {
	repo       repository.ReviewRepository
	orderRepo  repository.OrderRepository
	creditRepo repository.UserCreditRepository
}

// NewReviewService 创建评价 service 实例
func NewReviewService(
	repo repository.ReviewRepository,
	orderRepo repository.OrderRepository,
	creditRepo repository.UserCreditRepository,
) ReviewService {
	return &reviewService{repo: repo, orderRepo: orderRepo, creditRepo: creditRepo}
}

// Create 买家评价订单
func (s *reviewService) Create(orderID, reviewerID uint, reviewerName, reviewerAvatar string, req *dto.ReviewCreateRequest) (*dto.ReviewResponse, error) {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}
	if order.BuyerID != reviewerID {
		return nil, ErrReviewNoPermission
	}
	// 订单必须已完成
	if order.Status != model.OrderStatusCompleted {
		return nil, ErrOrderStatusInvalid
	}
	// 同一订单只能评价一次
	if existing, _ := s.repo.FindByOrderAndReviewer(orderID, reviewerID); existing != nil {
		return nil, ErrReviewExists
	}

	review := &model.ErshouReview{
		OrderID:        orderID,
		ErshouID:       0,
		ReviewerID:     reviewerID,
		ReviewerName:   reviewerName,
		ReviewerAvatar: reviewerAvatar,
		RevieweeID:     order.SellerID,
		ReviewType:     "buyer_to_seller",
		Rating:         req.Rating,
		Content:        req.Content,
		VideoURL:       req.VideoURL,
		IsAnonymous:    req.IsAnonymous,
		IsRecommended:  req.IsRecommended,
	}
	// 取第一个商品 ID 作为关联
	if items, _ := s.orderRepo.ListItems(orderID); len(items) > 0 {
		review.ErshouID = items[0].ErshouID
	}
	if len(req.Images) > 0 {
		if jb, err := model.FromJSON(req.Images); err == nil {
			review.Images = jb
		}
	}
	if len(req.Tags) > 0 {
		if jb, err := model.FromJSON(req.Tags); err == nil {
			review.Tags = jb
		}
	}
	if err := s.repo.Create(review); err != nil {
		return nil, err
	}

	// 累计卖家信用（好评 +1，中评 0，差评 -1；信用分好评+2，中评 0，差评-3）
	good, medium, bad := 0, 0, 0
	scoreDelta := 0
	switch {
	case req.Rating >= 4:
		good = 1
		scoreDelta = 2
	case req.Rating == 3:
		medium = 1
	default:
		bad = 1
		scoreDelta = -3
	}
	_ = s.creditRepo.IncrReviews(order.SellerID, good, medium, bad)
	_ = s.creditRepo.IncrTransactions(order.SellerID, 1, 0, scoreDelta)

	return s.toReviewResponse(review), nil
}

func (s *reviewService) GetByID(reviewID uint) (*dto.ReviewResponse, error) {
	r, err := s.repo.FindByID(reviewID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrReviewNotFound
		}
		return nil, err
	}
	return s.toReviewResponse(r), nil
}

func (s *reviewService) List(query dto.ReviewListQuery) (*utils.Pagination, []dto.ReviewResponse, error) {
	pagination := utils.NewPagination(query.Page, query.PageSize)
	list, total, err := s.repo.List(repository.ReviewListQuery{
		ErshouID:   query.ErshouID,
		ReviewerID: query.ReviewerID,
		Rating:     query.Rating,
	}, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.ReviewResponse, 0, len(list))
	for i := range list {
		result = append(result, *s.toReviewResponse(&list[i]))
	}
	return pagination, result, nil
}

func (s *reviewService) ListByErshouID(ershouID uint, pagination *utils.Pagination) (*utils.Pagination, []dto.ReviewResponse, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	list, total, err := s.repo.ListByErshouID(ershouID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.ReviewResponse, 0, len(list))
	for i := range list {
		result = append(result, *s.toReviewResponse(&list[i]))
	}
	return pagination, result, nil
}

// Reply 卖家回复评价
func (s *reviewService) Reply(reviewID, revieweeID uint, req *dto.ReviewReplyRequest) (*dto.ReviewResponse, error) {
	r, err := s.repo.FindByID(reviewID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrReviewNotFound
		}
		return nil, err
	}
	if r.RevieweeID != revieweeID {
		return nil, ErrReviewNoPermission
	}
	now := time.Now()
	if err := s.repo.Update(reviewID, map[string]interface{}{
		"reply":    req.Reply,
		"reply_at": &now,
	}); err != nil {
		return nil, err
	}
	updated, _ := s.repo.FindByID(reviewID)
	return s.toReviewResponse(updated), nil
}

// StatsByErshouID 评价统计
func (s *reviewService) StatsByErshouID(ershouID uint) (*dto.ReviewStats, error) {
	total, avg, good, medium, bad, err := s.repo.StatsByErshouID(ershouID)
	if err != nil {
		return nil, err
	}
	stats := &dto.ReviewStats{
		TotalReviews: int(total),
		AvgRating:    avg,
	}
	if total > 0 {
		stats.GoodRate = float64(good) / float64(total) * 100
		stats.MediumRate = float64(medium) / float64(total) * 100
		stats.BadRate = float64(bad) / float64(total) * 100
	}
	return stats, nil
}

func (s *reviewService) toReviewResponse(r *model.ErshouReview) *dto.ReviewResponse {
	resp := &dto.ReviewResponse{
		ID:             r.ID,
		OrderID:        r.OrderID,
		ErshouID:       r.ErshouID,
		ReviewerID:     r.ReviewerID,
		ReviewerName:   r.ReviewerName,
		ReviewerAvatar: r.ReviewerAvatar,
		RevieweeID:     r.RevieweeID,
		ReviewType:     r.ReviewType,
		Rating:         r.Rating,
		Content:        r.Content,
		VideoURL:       r.VideoURL,
		IsAnonymous:    r.IsAnonymous,
		IsRecommended:  r.IsRecommended,
		Reply:          r.Reply,
		ReplyAt:        r.ReplyAt,
		AppendContent:  r.AppendContent,
		AppendAt:       r.AppendAt,
		LikeCount:      r.LikeCount,
		CreatedAt:      r.CreatedAt,
		Images:         []string{},
		Tags:           []string{},
		AppendImages:   []string{},
	}
	if r.Images != nil {
		var imgs []string
		_ = r.Images.Parse(&imgs)
		if imgs != nil {
			resp.Images = imgs
		}
	}
	if r.Tags != nil {
		var tags []string
		_ = r.Tags.Parse(&tags)
		if tags != nil {
			resp.Tags = tags
		}
	}
	if r.AppendImages != nil {
		var imgs []string
		_ = r.AppendImages.Parse(&imgs)
		if imgs != nil {
			resp.AppendImages = imgs
		}
	}
	return resp
}

// ===== AuditRuleService =====

// AuditRuleService 审核规则业务接口
type AuditRuleService interface {
	Create(operatorID uint, req *dto.AuditRuleCreateRequest) (*dto.AuditRuleResponse, error)
	GetByID(id uint) (*dto.AuditRuleResponse, error)
	Update(id uint, req *dto.AuditRuleCreateRequest) (*dto.AuditRuleResponse, error)
	Delete(id uint) error
	List(ruleType string, status *int, page, pageSize int) (*utils.Pagination, []dto.AuditRuleResponse, error)
	ListEnabled() ([]dto.AuditRuleResponse, error)
}

type auditRuleService struct {
	repo repository.AuditRuleRepository
}

// NewAuditRuleService 创建审核规则 service 实例
func NewAuditRuleService(repo repository.AuditRuleRepository) AuditRuleService {
	return &auditRuleService{repo: repo}
}

func (s *auditRuleService) Create(operatorID uint, req *dto.AuditRuleCreateRequest) (*dto.AuditRuleResponse, error) {
	rule := &model.ErshouAuditRule{
		RuleName:    req.RuleName,
		RuleType:    req.RuleType,
		RuleKey:     req.RuleKey,
		Pattern:     req.Pattern,
		Action:      req.Action,
		PenaltyType: req.PenaltyType,
		Severity:    req.Severity,
		Status:      req.Status,
		Description: req.Description,
		Sort:        req.Sort,
	}
	if rule.Action == "" {
		rule.Action = "reject"
	}
	if rule.Severity == 0 {
		rule.Severity = 1
	}
	if rule.Status == 0 {
		rule.Status = model.AuditRuleStatusEnabled
	}
	if req.Threshold != nil {
		if jb, err := model.FromJSON(req.Threshold); err == nil {
			rule.Threshold = jb
		}
	}
	if err := s.repo.Create(rule); err != nil {
		return nil, err
	}
	return s.toAuditRuleResponse(rule), nil
}

func (s *auditRuleService) GetByID(id uint) (*dto.AuditRuleResponse, error) {
	r, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAuditRuleNotFound
		}
		return nil, err
	}
	return s.toAuditRuleResponse(r), nil
}

func (s *auditRuleService) Update(id uint, req *dto.AuditRuleCreateRequest) (*dto.AuditRuleResponse, error) {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAuditRuleNotFound
		}
		return nil, err
	}
	fields := map[string]interface{}{
		"rule_name":    req.RuleName,
		"rule_type":    req.RuleType,
		"rule_key":     req.RuleKey,
		"pattern":      req.Pattern,
		"action":       req.Action,
		"penalty_type": req.PenaltyType,
		"severity":     req.Severity,
		"status":       req.Status,
		"description":  req.Description,
		"sort":         req.Sort,
	}
	if req.Threshold != nil {
		if jb, err := model.FromJSON(req.Threshold); err == nil {
			fields["threshold"] = jb
		}
	}
	if err := s.repo.Update(id, fields); err != nil {
		return nil, err
	}
	updated, _ := s.repo.FindByID(id)
	return s.toAuditRuleResponse(updated), nil
}

func (s *auditRuleService) Delete(id uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrAuditRuleNotFound
		}
		return err
	}
	return s.repo.Delete(id)
}

func (s *auditRuleService) List(ruleType string, status *int, page, pageSize int) (*utils.Pagination, []dto.AuditRuleResponse, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.List(repository.AuditRuleListQuery{
		RuleType: ruleType,
		Status:   status,
	}, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.AuditRuleResponse, 0, len(list))
	for i := range list {
		result = append(result, *s.toAuditRuleResponse(&list[i]))
	}
	return pagination, result, nil
}

func (s *auditRuleService) ListEnabled() ([]dto.AuditRuleResponse, error) {
	list, err := s.repo.ListEnabled()
	if err != nil {
		return nil, err
	}
	result := make([]dto.AuditRuleResponse, 0, len(list))
	for i := range list {
		result = append(result, *s.toAuditRuleResponse(&list[i]))
	}
	return result, nil
}

func (s *auditRuleService) toAuditRuleResponse(r *model.ErshouAuditRule) *dto.AuditRuleResponse {
	resp := &dto.AuditRuleResponse{
		ID:          r.ID,
		RuleName:    r.RuleName,
		RuleType:    r.RuleType,
		RuleKey:     r.RuleKey,
		Pattern:     r.Pattern,
		Action:      r.Action,
		PenaltyType: r.PenaltyType,
		Severity:    r.Severity,
		Status:      r.Status,
		Description: r.Description,
		Sort:        r.Sort,
		CreatedAt:   r.CreatedAt,
		Threshold:   map[string]interface{}{},
	}
	if r.Threshold != nil {
		var m map[string]interface{}
		_ = r.Threshold.Parse(&m)
		if m != nil {
			resp.Threshold = m
		}
	}
	return resp
}

// ===== UserCreditService =====

// UserCreditService 用户信用业务接口
type UserCreditService interface {
	GetByUserID(userID uint) (*dto.UserCreditResponse, error)
	Update(userID uint, req *dto.UserCreditUpdateRequest) (*dto.UserCreditResponse, error)
	Ensure(userID uint) error
}

type userCreditService struct {
	repo repository.UserCreditRepository
}

// NewUserCreditService 创建用户信用 service 实例
func NewUserCreditService(repo repository.UserCreditRepository) UserCreditService {
	return &userCreditService{repo: repo}
}

func creditLevelText(level int) string {
	switch level {
	case model.CreditLevelNewbie:
		return "新手"
	case model.CreditLevelBronze:
		return "青铜"
	case model.CreditLevelSilver:
		return "白银"
	case model.CreditLevelGold:
		return "黄金"
	case model.CreditLevelPlatinum:
		return "铂金"
	case model.CreditLevelDiamond:
		return "钻石"
	}
	return "未知"
}

// Ensure 确保用户信用记录存在（首次登录/发布时调用）
func (s *userCreditService) Ensure(userID uint) error {
	if _, err := s.repo.FindByUserID(userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return s.repo.Create(&model.ErshouUserCredit{
				UserID:      userID,
				CreditScore: 100,
				CreditLevel: model.CreditLevelNewbie,
				GoodRate:    100.00,
			})
		}
		return err
	}
	return nil
}

func (s *userCreditService) GetByUserID(userID uint) (*dto.UserCreditResponse, error) {
	c, err := s.repo.FindByUserID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 不存在则初始化
			if err := s.Ensure(userID); err != nil {
				return nil, err
			}
			c, err = s.repo.FindByUserID(userID)
			if err != nil {
				return nil, ErrUserCreditNotFound
			}
		} else {
			return nil, err
		}
	}
	return s.toUserCreditResponse(c), nil
}

func (s *userCreditService) Update(userID uint, req *dto.UserCreditUpdateRequest) (*dto.UserCreditResponse, error) {
	if _, err := s.repo.FindByUserID(userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserCreditNotFound
		}
		return nil, err
	}
	fields := map[string]interface{}{
		"credit_score":  req.CreditScore,
		"credit_level":  req.CreditLevel,
		"frozen_reason": req.FrozenReason,
	}
	if req.FrozenUntil != nil {
		fields["frozen_until"] = req.FrozenUntil
	}
	if err := s.repo.Update(userID, fields); err != nil {
		return nil, err
	}
	updated, _ := s.repo.FindByUserID(userID)
	return s.toUserCreditResponse(updated), nil
}

func (s *userCreditService) toUserCreditResponse(c *model.ErshouUserCredit) *dto.UserCreditResponse {
	resp := &dto.UserCreditResponse{
		ID:                  c.ID,
		UserID:              c.UserID,
		CreditScore:         c.CreditScore,
		CreditLevel:         c.CreditLevel,
		CreditLevelText:     creditLevelText(c.CreditLevel),
		TotalTransactions:   c.TotalTransactions,
		SuccessTransactions: c.SuccessTransactions,
		CancelTransactions:  c.CancelTransactions,
		GoodReviews:         c.GoodReviews,
		MediumReviews:       c.MediumReviews,
		BadReviews:          c.BadReviews,
		GoodRate:            c.GoodRate,
		Disputes:            c.Disputes,
		Reports:             c.Reports,
		Penalties:           c.Penalties,
		LastTransactionAt:   c.LastTransactionAt,
		FrozenReason:        c.FrozenReason,
		FrozenUntil:         c.FrozenUntil,
		CreatedAt:           c.CreatedAt,
		UpdatedAt:           c.UpdatedAt,
	}
	if c.FrozenUntil != nil && c.FrozenUntil.After(time.Now()) {
		resp.IsFrozen = true
	}
	if c.CreditScore < 30 {
		resp.IsFrozen = true
	}
	return resp
}
