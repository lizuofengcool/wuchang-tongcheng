// Package service 同城零工兼职业务逻辑层 - 信用分
// 对标芝麻信用/猪八戒：履约 +10 / 违约 -20 / 影响接单 + 历史变更记录
// 4 维数据隔离（region_id + user_id）
package service

import (
	"errors"

	"wuchang-tongcheng/internal/modules/linggong/dto"
	"wuchang-tongcheng/internal/modules/linggong/model"
	"wuchang-tongcheng/internal/modules/linggong/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrCreditNotFound     = errors.New("信用记录不存在")
	ErrCreditUserNotFound = errors.New("用户不存在")
	ErrCreditInvalid      = errors.New("信用调整参数无效")
)

// CreditService 信用分业务接口
type CreditService interface {
	// C 端
	GetByID(id uint) (*dto.CreditInfo, error)
	List(regionID uint, req *dto.CreditListRequest) (*utils.Pagination, []dto.CreditInfo, error)
	ListByUser(userID uint, userType string, page, pageSize int) (*utils.Pagination, []dto.CreditInfo, error)
	GetScore(userID uint, userType string) (*dto.CreditScoreResponse, error)

	// M 端管理
	Adjust(regionID uint, operatorID uint, operatorName string, req *dto.CreditAdjustRequest) (*dto.CreditInfo, error)
	Delete(id uint) error
}

type creditService struct {
	repo repository.CreditRepository
}

// NewCreditService 创建信用分 service 实例
func NewCreditService(repo repository.CreditRepository) CreditService {
	return &creditService{repo: repo}
}

// creditReasonText 信用变更原因文本
func creditReasonText(r string) string {
	switch r {
	case model.CreditReasonFulfill:
		return "履约完成"
	case model.CreditReasonBreach:
		return "违约"
	case model.CreditReasonLate:
		return "迟到"
	case model.CreditReasonAbsent:
		return "缺勤"
	case model.CreditReasonNoShow:
		return "放鸽子"
	case model.CreditReasonGoodRating:
		return "好评"
	case model.CreditReasonBadRating:
		return "差评"
	case model.CreditReasonVerified:
		return "实名认证"
	case model.CreditReasonSkillCert:
		return "技能认证"
	case model.CreditReasonReport:
		return "被举报"
	case model.CreditReasonAppeal:
		return "申诉成功"
	case model.CreditReasonManual:
		return "人工调整"
	case model.CreditReasonInviteFriend:
		return "邀请好友"
	case model.CreditReasonDailyLogin:
		return "每日登录"
	case model.CreditReasonCompleteProfile:
		return "完善资料"
	}
	return ""
}

// creditChangeTypeText 变更类型文本
func creditChangeTypeText(t string) string {
	switch t {
	case model.CreditTypeAdd:
		return "加分"
	case model.CreditTypeDeduct:
		return "扣分"
	case model.CreditTypeReset:
		return "重置"
	}
	return ""
}

// creditUserTypeText 用户类型文本
func creditUserTypeText(t string) string {
	switch t {
	case model.CreditUserTypeWorker:
		return "求职者"
	case model.CreditUserTypeEmployer:
		return "雇主"
	}
	return ""
}

// creditLevel 信用等级（按分数段映射）
// 0-59 青铜 / 60-79 白银 / 80-99 黄金 / 100-119 铂金 / 120+ 钻石
func creditLevel(score int) int {
	switch {
	case score >= 120:
		return 5
	case score >= 100:
		return 4
	case score >= 80:
		return 3
	case score >= 60:
		return 2
	default:
		return 1
	}
}

// toCreditInfo model -> dto
func toCreditInfo(c *model.LinggongCredit) *dto.CreditInfo {
	return &dto.CreditInfo{
		ID:             c.ID,
		UserID:         c.UserID,
		UserType:       c.UserType,
		UserTypeText:   creditUserTypeText(c.UserType),
		Reason:         c.Reason,
		ReasonText:     creditReasonText(c.Reason),
		ChangeType:     c.ChangeType,
		ChangeTypeText: creditChangeTypeText(c.ChangeType),
		ChangeScore:    c.ChangeScore,
		BeforeScore:    c.BeforeScore,
		AfterScore:     c.AfterScore,
		LinggongID:     c.LinggongID,
		TaskID:         c.TaskID,
		ApplicationID:  c.ApplicationID,
		RatingID:       c.RatingID,
		OperatorID:     c.OperatorID,
		OperatorName:   c.OperatorName,
		Description:    c.Description,
		EvidenceURL:    c.EvidenceURL,
		RegionID:       c.RegionID,
		CreatedAt:      c.CreatedAt,
	}
}

// ===== C 端 =====

// GetByID 获取信用变更详情
func (s *creditService) GetByID(id uint) (*dto.CreditInfo, error) {
	c, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCreditNotFound
		}
		return nil, err
	}
	return toCreditInfo(c), nil
}

// List 信用变更记录列表
func (s *creditService) List(regionID uint, req *dto.CreditListRequest) (*utils.Pagination, []dto.CreditInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.CreditListOptions{
		UserID:     req.UserID,
		UserType:   req.UserType,
		Reason:     req.Reason,
		ChangeType: req.ChangeType,
	}
	list, total, err := s.repo.List(regionID, pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.CreditInfo, 0, len(list))
	for i := range list {
		result = append(result, *toCreditInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListByUser 按用户反查信用变更记录
func (s *creditService) ListByUser(userID uint, userType string, page, pageSize int) (*utils.Pagination, []dto.CreditInfo, error) {
	if userType == "" {
		userType = model.CreditUserTypeWorker
	}
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByUser(userID, userType, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.CreditInfo, 0, len(list))
	for i := range list {
		result = append(result, *toCreditInfo(&list[i]))
	}
	return pagination, result, nil
}

// GetScore 查询用户当前信用分
func (s *creditService) GetScore(userID uint, userType string) (*dto.CreditScoreResponse, error) {
	if userType == "" {
		userType = model.CreditUserTypeWorker
	}
	score, err := s.repo.GetLatestScore(userID, userType)
	if err != nil {
		return nil, err
	}
	return &dto.CreditScoreResponse{
		UserID:      userID,
		UserType:    userType,
		CreditScore: score,
		Level:       creditLevel(score),
	}, nil
}

// ===== M 端管理 =====

// Adjust 信用分调整（M 端调用，自动计算 before/after 分数）
func (s *creditService) Adjust(regionID uint, operatorID uint, operatorName string, req *dto.CreditAdjustRequest) (*dto.CreditInfo, error) {
	if req.UserID == 0 {
		return nil, ErrCreditInvalid
	}
	userType := req.UserType
	if userType == "" {
		userType = model.CreditUserTypeWorker
	}
	changeType := req.ChangeType
	if changeType == "" {
		changeType = model.CreditTypeAdd
	}
	reason := req.Reason
	if reason == "" {
		reason = model.CreditReasonManual
	}

	// 取变更前分数
	before, err := s.repo.GetLatestScore(req.UserID, userType)
	if err != nil {
		return nil, err
	}

	// 计算变更后分数
	after := before
	switch changeType {
	case model.CreditTypeAdd:
		after = before + req.ChangeScore
	case model.CreditTypeDeduct:
		after = before - req.ChangeScore
		if after < 0 {
			after = 0
		}
	case model.CreditTypeReset:
		after = req.ChangeScore // reset 时 ChangeScore 表示目标分数
	}

	c := &model.LinggongCredit{
		UserID:        req.UserID,
		UserType:      userType,
		Reason:        reason,
		ChangeType:    changeType,
		ChangeScore:   req.ChangeScore,
		BeforeScore:   before,
		AfterScore:    after,
		LinggongID:    req.LinggongID,
		TaskID:        req.TaskID,
		ApplicationID: req.ApplicationID,
		RatingID:      req.RatingID,
		OperatorID:    operatorID,
		OperatorName:  operatorName,
		Description:   req.Description,
		EvidenceURL:   req.EvidenceURL,
	}
	c.RegionID = regionID

	if err := s.repo.Create(c); err != nil {
		return nil, err
	}
	return toCreditInfo(c), nil
}

// Delete 删除信用变更记录（仅 M 端，谨慎操作）
func (s *creditService) Delete(id uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCreditNotFound
		}
		return err
	}
	return s.repo.Delete(id)
}
