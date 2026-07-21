// Package service 营销活动中台业务逻辑层 - 签到（sign 子域）
package service

import (
	"errors"
	"time"

	"wuchang-tongcheng/internal/modules/marketing/dto"
	"wuchang-tongcheng/internal/modules/marketing/model"
	"wuchang-tongcheng/internal/modules/marketing/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// SignService 签到业务接口
type SignService interface {
	// 签到
	CheckIn(userID uint) (*dto.SignCheckInResponse, error)
	GetCalendar(userID uint, month string) (*dto.SignCalendarResponse, error)

	// 签到规则 CRUD
	CreateRule(req *dto.CreateSignRuleRequest) (*dto.SignRuleInfo, error)
	UpdateRule(id uint, req *dto.UpdateSignRuleRequest) error
	DeleteRule(id uint) error
	GetRuleByID(id uint) (*dto.SignRuleInfo, error)
	ListRules(req *dto.SignRuleListRequest) (*utils.Pagination, []dto.SignRuleInfo, error)
	ListEnabledRules() ([]dto.SignRuleInfo, error)
}

type signService struct {
	repo repository.SignRepository
}

// NewSignService 创建签到 service 实例
func NewSignService(repo repository.SignRepository) SignService {
	return &signService{repo: repo}
}

// signRuleStatusText 签到规则状态文本
func signRuleStatusText(s int) string {
	switch s {
	case model.SignRuleStatusDisabled:
		return "禁用"
	case model.SignRuleStatusEnabled:
		return "启用"
	}
	return ""
}

// toSignRuleInfo model -> dto
func toSignRuleInfo(r *model.SignRule) *dto.SignRuleInfo {
	info := &dto.SignRuleInfo{
		ID:         r.ID,
		Day:        r.Day,
		Points:     r.Points,
		Status:     r.Status,
		StatusText: signRuleStatusText(r.Status),
		CreatedAt:  r.CreatedAt,
		UpdatedAt:  r.UpdatedAt,
	}
	if r.ExtraReward != nil && len(r.ExtraReward) > 0 {
		var extra interface{}
		if err := r.ExtraReward.Parse(&extra); err == nil {
			info.ExtraReward = extra
		} else {
			info.ExtraReward = r.ExtraReward.String()
		}
	}
	return info
}

// toSignRecordInfo model -> dto
func toSignRecordInfo(r *model.SignRecord) *dto.SignRecordInfo {
	return &dto.SignRecordInfo{
		ID:             r.ID,
		UserID:         r.UserID,
		SignDate:       r.SignDate,
		ContinuousDays: r.ContinuousDays,
		Points:         r.Points,
		CreatedAt:      r.CreatedAt,
	}
}

// todayDate 返回当日 00:00（本地时区）
func todayDate() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

// CheckIn 每日签到
func (s *signService) CheckIn(userID uint) (*dto.SignCheckInResponse, error) {
	today := todayDate()
	// 1. 检查今日是否已签到
	if existing, err := s.repo.FindTodayRecord(userID, today); err == nil && existing != nil {
		return nil, ErrSignAlreadyToday
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 2. 计算连续签到天数
	continuousDays := 1
	if latest, err := s.repo.FindLatestRecord(userID); err == nil && latest != nil {
		// 昨日有签到则连续 +1，否则重置为 1
		yesterday := today.AddDate(0, 0, -1)
		if latest.SignDate.Equal(yesterday) {
			continuousDays = latest.ContinuousDays + 1
		}
	}

	// 3. 计算奖励积分（命中签到规则取规则积分，否则取基础 1 积分）
	points := 1
	var extraReward interface{}
	if rule, err := s.repo.FindRuleByDay(continuousDays); err == nil && rule != nil && rule.Status == model.SignRuleStatusEnabled {
		points = rule.Points
		if rule.ExtraReward != nil && len(rule.ExtraReward) > 0 {
			_ = rule.ExtraReward.Parse(&extraReward)
		}
	}

	// 4. 落库签到记录
	rec := &model.SignRecord{
		UserID:         userID,
		SignDate:       today,
		ContinuousDays: continuousDays,
		Points:         points,
	}
	if err := s.repo.CreateRecord(rec); err != nil {
		return nil, err
	}

	// 5. 累计积分
	totalPoints, _ := s.repo.SumPoints(userID)

	return &dto.SignCheckInResponse{
		Record:         toSignRecordInfo(rec),
		ContinuousDays: continuousDays,
		Points:         points,
		TotalPoints:    totalPoints,
		ExtraReward:    extraReward,
	}, nil
}

// GetCalendar 签到日历
func (s *signService) GetCalendar(userID uint, month string) (*dto.SignCalendarResponse, error) {
	loc := time.Now().Location()
	var start time.Time
	if month != "" {
		t, err := time.ParseInLocation("2006-01", month, loc)
		if err != nil {
			return nil, errors.New("月份格式应为 YYYY-MM")
		}
		start = t
	} else {
		now := time.Now()
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
	}
	end := start.AddDate(0, 1, 0)

	records, err := s.repo.ListRecordsByMonth(userID, start, end)
	if err != nil {
		return nil, err
	}

	infos := make([]dto.SignRecordInfo, 0, len(records))
	var signedPoints int
	for i := range records {
		infos = append(infos, *toSignRecordInfo(&records[i]))
		signedPoints += records[i].Points
	}

	// 当前连续天数：取最新记录
	continuousDays := 0
	if latest, err := s.repo.FindLatestRecord(userID); err == nil && latest != nil {
		today := todayDate()
		// 若最新签到是今日或昨日，连续天数有效
		if latest.SignDate.Equal(today) || latest.SignDate.Equal(today.AddDate(0, 0, -1)) {
			continuousDays = latest.ContinuousDays
		}
	}

	// 当月天数
	monthDays := end.AddDate(0, 0, -1).Day()

	return &dto.SignCalendarResponse{
		Records:        infos,
		ContinuousDays: continuousDays,
		MonthDays:      monthDays,
		SignedDays:     len(infos),
		TotalPoints:    signedPoints,
	}, nil
}

// ===== 签到规则 CRUD =====

// CreateRule 创建签到规则
func (s *signService) CreateRule(req *dto.CreateSignRuleRequest) (*dto.SignRuleInfo, error) {
	// 同一天数规则唯一
	if existing, err := s.repo.FindRuleByDay(req.Day); err == nil && existing != nil {
		return nil, ErrSignRuleExists
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	status := req.Status
	if status == 0 {
		status = model.SignRuleStatusEnabled
	}
	rule := &model.SignRule{
		Day:    req.Day,
		Points: req.Points,
		Status: status,
	}
	if req.ExtraReward != nil {
		if b, err := model.FromJSON(req.ExtraReward); err == nil {
			rule.ExtraReward = b
		}
	}
	if err := s.repo.CreateRule(rule); err != nil {
		return nil, err
	}
	return toSignRuleInfo(rule), nil
}

// UpdateRule 更新签到规则
func (s *signService) UpdateRule(id uint, req *dto.UpdateSignRuleRequest) error {
	if _, err := s.repo.FindRuleByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSignRuleNotFound
		}
		return err
	}
	fields := make(map[string]interface{})
	if req.Points != nil {
		fields["points"] = *req.Points
	}
	if req.Status != nil {
		fields["status"] = *req.Status
	}
	if req.ExtraReward != nil {
		if b, err := model.FromJSON(req.ExtraReward); err == nil {
			fields["extra_reward"] = b
		}
	}
	if len(fields) == 0 {
		return nil
	}
	return s.repo.UpdateRule(id, fields)
}

// DeleteRule 删除签到规则
func (s *signService) DeleteRule(id uint) error {
	if _, err := s.repo.FindRuleByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSignRuleNotFound
		}
		return err
	}
	return s.repo.DeleteRule(id)
}

// GetRuleByID 获取签到规则详情
func (s *signService) GetRuleByID(id uint) (*dto.SignRuleInfo, error) {
	rule, err := s.repo.FindRuleByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSignRuleNotFound
		}
		return nil, err
	}
	return toSignRuleInfo(rule), nil
}

// ListRules 签到规则列表
func (s *signService) ListRules(req *dto.SignRuleListRequest) (*utils.Pagination, []dto.SignRuleInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	query := repository.SignRuleListQuery{Status: req.Status}
	list, total, err := s.repo.ListRules(query, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.SignRuleInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toSignRuleInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// ListEnabledRules 启用的签到规则列表
func (s *signService) ListEnabledRules() ([]dto.SignRuleInfo, error) {
	list, err := s.repo.ListEnabledRules()
	if err != nil {
		return nil, err
	}
	infos := make([]dto.SignRuleInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toSignRuleInfo(&list[i]))
	}
	return infos, nil
}
