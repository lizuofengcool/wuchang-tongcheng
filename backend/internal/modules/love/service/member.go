// Package service love 相亲交友业务逻辑层 - 会员
package service

import (
	"errors"
	"fmt"
	"time"

	"wuchang-tongcheng/internal/modules/love/dto"
	"wuchang-tongcheng/internal/modules/love/model"
	"wuchang-tongcheng/internal/modules/love/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrLoveMemberLevelNotFound  = errors.New("会员等级不存在")
	ErrLoveMembershipNotFound   = errors.New("会员订阅不存在")
	ErrLoveMembershipNoPermission = errors.New("无权操作此订阅")
	ErrLoveMembershipInvalidOp  = errors.New("订阅状态不允许此操作")
	ErrLoveLevelCodeExists      = errors.New("等级编码已存在")
)

// LoveMemberLevelService 会员等级业务接口（M 端配置）
type LoveMemberLevelService interface {
	Create(req *dto.CreateLoveMemberLevelRequest) (*dto.LoveMemberLevelInfo, error)
	Update(id uint, req *dto.UpdateLoveMemberLevelRequest) error
	Delete(id uint) error
	GetByID(id uint) (*dto.LoveMemberLevelInfo, error)
	GetByLevelCode(code string) (*dto.LoveMemberLevelInfo, error)
	List(req *dto.LoveMembershipListRequest) (*utils.Pagination, []dto.LoveMemberLevelInfo, error)
	ListAll() ([]dto.LoveMemberLevelInfo, error)
}

// LoveMembershipService 会员订阅业务接口
type LoveMembershipService interface {
	Open(loveID, userID uint, req *dto.CreateLoveMembershipRequest) (*dto.LoveMembershipInfo, error)
	GetByID(id uint) (*dto.LoveMembershipInfo, error)
	GetMyActive(userID uint) (*dto.LoveMembershipInfo, error)
	List(req *dto.LoveMembershipListRequest) (*utils.Pagination, []dto.LoveMembershipInfo, error)
	ListByUser(userID uint, req *dto.LoveMembershipListRequest) (*utils.Pagination, []dto.LoveMembershipInfo, error)
	Cancel(id uint, userID uint, req *dto.CancelLoveMembershipRequest) error
	Refund(id uint, userID uint, req *dto.RefundLoveMembershipRequest) error
	MarkPaid(id uint, payMethod, payOrderNo string) error
	UpdateAutoRenew(id uint, userID uint, autoRenew bool) error
}

// ===== 会员等级 service =====

type loveMemberLevelService struct {
	repo repository.LoveMemberLevelRepository
}

// NewLoveMemberLevelService 创建会员等级 service
func NewLoveMemberLevelService(repo repository.LoveMemberLevelRepository) LoveMemberLevelService {
	return &loveMemberLevelService{repo: repo}
}

func memberLevelStatusText(s int) string {
	switch s {
	case 0:
		return "禁用"
	case 1:
		return "启用"
	}
	return ""
}

func toLoveMemberLevelInfo(l *model.LoveMemberLevel) dto.LoveMemberLevelInfo {
	info := dto.LoveMemberLevelInfo{
		ID:                   l.ID,
		LevelCode:            l.LevelCode,
		LevelName:            l.LevelName,
		Level:                l.Level,
		Description:          l.Description,
		Icon:                 l.Icon,
		Color:                l.Color,
		MonthlyPrice:         l.MonthlyPrice,
		QuarterlyPrice:       l.QuarterlyPrice,
		YearlyPrice:          l.YearlyPrice,
		DailySuperLikes:      l.DailySuperLikes,
		DailyLikes:           l.DailyLikes,
		DailyVisits:          l.DailyVisits,
		DailyRecommendations: l.DailyRecommendations,
		CanSeeVisitors:       l.CanSeeVisitors,
		CanSeeLikes:          l.CanSeeLikes,
		CanHideOnline:        l.CanHideOnline,
		CanHideLocation:      l.CanHideLocation,
		CanFilterVerified:    l.CanFilterVerified,
		CanAdvancedFilter:    l.CanAdvancedFilter,
		CanSuperLike:         l.CanSuperLike,
		CanUndoSwipe:         l.CanUndoSwipe,
		CanBoostProfile:      l.CanBoostProfile,
		CanSeeMatchScore:     l.CanSeeMatchScore,
		Sort:                 l.Sort,
		Status:               l.Status,
		CreatedAt:            l.CreatedAt,
		UpdatedAt:            l.UpdatedAt,
	}
	if l.Perks != nil {
		info.Perks = l.Perks
	}
	return info
}

func (s *loveMemberLevelService) Create(req *dto.CreateLoveMemberLevelRequest) (*dto.LoveMemberLevelInfo, error) {
	if existing, err := s.repo.FindByLevelCode(req.LevelCode); err == nil && existing != nil {
		return nil, ErrLoveLevelCodeExists
	}
	l := &model.LoveMemberLevel{
		LevelCode:            req.LevelCode,
		LevelName:            req.LevelName,
		Level:                req.Level,
		Description:          req.Description,
		Icon:                 req.Icon,
		Color:                req.Color,
		MonthlyPrice:         req.MonthlyPrice,
		QuarterlyPrice:       req.QuarterlyPrice,
		YearlyPrice:          req.YearlyPrice,
		DailySuperLikes:      req.DailySuperLikes,
		DailyLikes:           req.DailyLikes,
		DailyVisits:          req.DailyVisits,
		DailyRecommendations: req.DailyRecommendations,
		CanSeeVisitors:       req.CanSeeVisitors,
		CanSeeLikes:          req.CanSeeLikes,
		CanHideOnline:        req.CanHideOnline,
		CanHideLocation:      req.CanHideLocation,
		CanFilterVerified:    req.CanFilterVerified,
		CanAdvancedFilter:    req.CanAdvancedFilter,
		CanSuperLike:         req.CanSuperLike,
		CanUndoSwipe:         req.CanUndoSwipe,
		CanBoostProfile:      req.CanBoostProfile,
		CanSeeMatchScore:     req.CanSeeMatchScore,
		Sort:                 req.Sort,
		Status:               1,
	}
	if req.Status == 0 {
		l.Status = 0
	}
	if req.Perks != nil {
		if jb, err := model.FromJSON(req.Perks); err == nil {
			l.Perks = jb
		}
	}
	if err := s.repo.Create(l); err != nil {
		return nil, err
	}
	info := toLoveMemberLevelInfo(l)
	return &info, nil
}

func (s *loveMemberLevelService) Update(id uint, req *dto.UpdateLoveMemberLevelRequest) error {
	l, err := s.repo.FindByID(id)
	if err != nil {
		return ErrLoveMemberLevelNotFound
	}
	fields := map[string]interface{}{}
	if req.LevelName != nil {
		fields["level_name"] = *req.LevelName
	}
	if req.Description != nil {
		fields["description"] = *req.Description
	}
	if req.Icon != nil {
		fields["icon"] = *req.Icon
	}
	if req.Color != nil {
		fields["color"] = *req.Color
	}
	if req.MonthlyPrice != nil {
		fields["monthly_price"] = *req.MonthlyPrice
	}
	if req.QuarterlyPrice != nil {
		fields["quarterly_price"] = *req.QuarterlyPrice
	}
	if req.YearlyPrice != nil {
		fields["yearly_price"] = *req.YearlyPrice
	}
	if req.DailySuperLikes != nil {
		fields["daily_super_likes"] = *req.DailySuperLikes
	}
	if req.DailyLikes != nil {
		fields["daily_likes"] = *req.DailyLikes
	}
	if req.DailyVisits != nil {
		fields["daily_visits"] = *req.DailyVisits
	}
	if req.DailyRecommendations != nil {
		fields["daily_recommendations"] = *req.DailyRecommendations
	}
	if req.Sort != nil {
		fields["sort"] = *req.Sort
	}
	if req.Status != nil {
		fields["status"] = *req.Status
	}
	if len(fields) == 0 {
		return nil
	}
	_ = l
	return s.repo.UpdateFields(id, fields)
}

func (s *loveMemberLevelService) Delete(id uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		return ErrLoveMemberLevelNotFound
	}
	return s.repo.Delete(id)
}

func (s *loveMemberLevelService) GetByID(id uint) (*dto.LoveMemberLevelInfo, error) {
	l, err := s.repo.FindByID(id)
	if err != nil {
		return nil, ErrLoveMemberLevelNotFound
	}
	info := toLoveMemberLevelInfo(l)
	return &info, nil
}

func (s *loveMemberLevelService) GetByLevelCode(code string) (*dto.LoveMemberLevelInfo, error) {
	l, err := s.repo.FindByLevelCode(code)
	if err != nil {
		return nil, ErrLoveMemberLevelNotFound
	}
	info := toLoveMemberLevelInfo(l)
	return &info, nil
}

func (s *loveMemberLevelService) List(req *dto.LoveMembershipListRequest) (*utils.Pagination, []dto.LoveMemberLevelInfo, error) {
	opts := repository.LoveMemberLevelListOptions{}
	if req.Status != nil {
		opts.Status = req.Status
	}
	list, total, err := s.repo.List(&req.Pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.LoveMemberLevelInfo, 0, len(list))
	for i := range list {
		infos = append(infos, toLoveMemberLevelInfo(&list[i]))
	}
	req.Pagination.Total = total
	return &req.Pagination, infos, nil
}

func (s *loveMemberLevelService) ListAll() ([]dto.LoveMemberLevelInfo, error) {
	list, err := s.repo.ListAll()
	if err != nil {
		return nil, err
	}
	infos := make([]dto.LoveMemberLevelInfo, 0, len(list))
	for i := range list {
		infos = append(infos, toLoveMemberLevelInfo(&list[i]))
	}
	return infos, nil
}

// ===== 会员订阅 service =====

type loveMembershipService struct {
	repo     repository.LoveMembershipRepository
	levelRepo repository.LoveMemberLevelRepository
}

// NewLoveMembershipService 创建会员订阅 service
func NewLoveMembershipService(repo repository.LoveMembershipRepository, levelRepo repository.LoveMemberLevelRepository) LoveMembershipService {
	return &loveMembershipService{repo: repo, levelRepo: levelRepo}
}

func membershipStatusText(s int) string {
	switch s {
	case 0:
		return "未支付"
	case 1:
		return "有效"
	case 2:
		return "已取消"
	case 3:
		return "已退款"
	case 4:
		return "已过期"
	}
	return ""
}

func toLoveMembershipInfo(m *model.LoveMembership) dto.LoveMembershipInfo {
	info := dto.LoveMembershipInfo{
		ID:            m.ID,
		SubNo:         m.SubNo,
		UserID:        m.UserID,
		LoveID:        m.LoveID,
		LevelCode:     m.LevelCode,
		LevelName:     m.LevelName,
		Level:         m.Level,
		Plan:          m.Plan,
		Period:        m.Period,
		StartAt:       m.StartAt,
		EndAt:         m.EndAt,
		Price:         m.Price,
		PayAmount:     m.PayAmount,
		PayMethod:     m.PayMethod,
		PayOrderNo:    m.PayOrderNo,
		PayAt:         m.PayAt,
		AutoRenew:     m.AutoRenew,
		RenewCount:    m.RenewCount,
		Status:        m.Status,
		StatusText:    membershipStatusText(m.Status),
		CancelAt:      m.CancelAt,
		CancelReason:  m.CancelReason,
		RefundAmount:  m.RefundAmount,
		RefundAt:      m.RefundAt,
		RefundReason:  m.RefundReason,
		Source:        m.Source,
		Remark:        m.Remark,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
	if m.PerksSnapshot != nil {
		info.PerksSnapshot = m.PerksSnapshot
	}
	return info
}

func membershipPlanDuration(plan string, period int) time.Duration {
	switch plan {
	case model.MemberPlanMonthly:
		return time.Duration(period) * 30 * 24 * time.Hour
	case model.MemberPlanQuarterly:
		return time.Duration(period) * 90 * 24 * time.Hour
	case model.MemberPlanYearly:
		return time.Duration(period) * 365 * 24 * time.Hour
	}
	return time.Duration(period) * 30 * 24 * time.Hour
}

func (s *loveMembershipService) Open(loveID, userID uint, req *dto.CreateLoveMembershipRequest) (*dto.LoveMembershipInfo, error) {
	level, err := s.levelRepo.FindByLevelCode(req.LevelCode)
	if err != nil {
		return nil, ErrLoveMemberLevelNotFound
	}

	var price float64
	switch req.Plan {
	case model.MemberPlanMonthly:
		price = level.MonthlyPrice
	case model.MemberPlanQuarterly:
		price = level.QuarterlyPrice
	case model.MemberPlanYearly:
		price = level.YearlyPrice
	}
	if req.Period > 1 {
		price = price * float64(req.Period)
	}

	now := time.Now()
	duration := membershipPlanDuration(req.Plan, req.Period)
	endAt := now.Add(duration)
	subNo := fmt.Sprintf("LVM%s%08d", now.Format("20060102150405"), userID%100000000)

	m := &model.LoveMembership{
		SubNo:     subNo,
		UserID:    userID,
		LoveID:    loveID,
		LevelCode: level.LevelCode,
		LevelName: level.LevelName,
		Level:     level.Level,
		Plan:      req.Plan,
		Period:    req.Period,
		StartAt:   now,
		EndAt:     endAt,
		Price:     price,
		PayAmount: price,
		PayMethod: req.PayMethod,
		AutoRenew: req.AutoRenew,
		Status:    0, // 未支付
		Source:    "self",
	}
	if level.Perks != nil {
		m.PerksSnapshot = level.Perks
	}

	if err := s.repo.Create(m); err != nil {
		return nil, err
	}
	info := toLoveMembershipInfo(m)
	return &info, nil
}

func (s *loveMembershipService) GetByID(id uint) (*dto.LoveMembershipInfo, error) {
	m, err := s.repo.FindByID(id)
	if err != nil {
		return nil, ErrLoveMembershipNotFound
	}
	info := toLoveMembershipInfo(m)
	return &info, nil
}

func (s *loveMembershipService) GetMyActive(userID uint) (*dto.LoveMembershipInfo, error) {
	m, err := s.repo.FindByUserActive(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	info := toLoveMembershipInfo(m)
	return &info, nil
}

func (s *loveMembershipService) List(req *dto.LoveMembershipListRequest) (*utils.Pagination, []dto.LoveMembershipInfo, error) {
	opts := repository.LoveMembershipListOptions{
		UserID:    req.UserID,
		LoveID:    req.LoveID,
		LevelCode: req.LevelCode,
		Plan:      req.Plan,
	}
	if req.Status != nil {
		opts.Status = req.Status
	}
	list, total, err := s.repo.List(&req.Pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.LoveMembershipInfo, 0, len(list))
	for i := range list {
		infos = append(infos, toLoveMembershipInfo(&list[i]))
	}
	req.Pagination.Total = total
	return &req.Pagination, infos, nil
}

func (s *loveMembershipService) ListByUser(userID uint, req *dto.LoveMembershipListRequest) (*utils.Pagination, []dto.LoveMembershipInfo, error) {
	req.UserID = userID
	return s.List(req)
}

func (s *loveMembershipService) Cancel(id uint, userID uint, req *dto.CancelLoveMembershipRequest) error {
	m, err := s.repo.FindByID(id)
	if err != nil {
		return ErrLoveMembershipNotFound
	}
	if m.UserID != userID {
		return ErrLoveMembershipNoPermission
	}
	if m.Status != 1 && m.Status != 0 {
		return ErrLoveMembershipInvalidOp
	}
	return s.repo.Cancel(id, req.Reason)
}

func (s *loveMembershipService) Refund(id uint, userID uint, req *dto.RefundLoveMembershipRequest) error {
	m, err := s.repo.FindByID(id)
	if err != nil {
		return ErrLoveMembershipNotFound
	}
	if m.UserID != userID {
		return ErrLoveMembershipNoPermission
	}
	if m.Status != 1 {
		return ErrLoveMembershipInvalidOp
	}
	if req.Amount > m.PayAmount {
		return errors.New("退款金额不能超过实付金额")
	}
	return s.repo.Refund(id, req.Amount, req.Reason)
}

func (s *loveMembershipService) MarkPaid(id uint, payMethod, payOrderNo string) error {
	m, err := s.repo.FindByID(id)
	if err != nil {
		return ErrLoveMembershipNotFound
	}
	if m.Status != 0 {
		return ErrLoveMembershipInvalidOp
	}
	return s.repo.MarkPaid(id, payMethod, payOrderNo)
}

func (s *loveMembershipService) UpdateAutoRenew(id uint, userID uint, autoRenew bool) error {
	m, err := s.repo.FindByID(id)
	if err != nil {
		return ErrLoveMembershipNotFound
	}
	if m.UserID != userID {
		return ErrLoveMembershipNoPermission
	}
	return s.repo.UpdateFields(id, map[string]interface{}{"auto_renew": autoRenew})
}
