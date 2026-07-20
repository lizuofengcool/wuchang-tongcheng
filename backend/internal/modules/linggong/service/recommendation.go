// Package service 同城零工兼职业务逻辑层 - 推荐岗位
// 对标斗米智能推荐 + 猪八戒威客匹配：岗位推荐 + 求职者推荐 + 热门推荐
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
	ErrRecommendationNotFound = errors.New("推荐记录不存在")
	ErrRecommendationExpired  = errors.New("推荐已过期")
)

// RecommendationService 推荐业务接口
type RecommendationService interface {
	// C 端
	GetByID(id uint) (*dto.RecommendationInfo, error)
	List(userID uint, req *dto.RecommendationListRequest) (*utils.Pagination, []dto.RecommendationInfo, error)
	ListByUser(userID uint, page, pageSize int) (*utils.Pagination, []dto.RecommendationInfo, error)
	ListByLinggong(linggongID uint, page, pageSize int) (*utils.Pagination, []dto.RecommendationInfo, error)
	// 推荐反馈
	Click(id uint, userID uint) error
	Apply(id uint, userID uint) error
	View(id uint, userID uint) error
	Dismiss(id uint, userID uint) error

	// M 端 / 内部调用
	Create(regionID uint, req *model.LinggongRecommendation) (*dto.RecommendationInfo, error)
	Delete(id uint) error
}

type recommendationService struct {
	repo repository.RecommendationRepository
}

// NewRecommendationService 创建推荐 service 实例
func NewRecommendationService(repo repository.RecommendationRepository) RecommendationService {
	return &recommendationService{repo: repo}
}

// recTypeText 推荐类型文本
func recTypeText(t string) string {
	switch t {
	case model.RecTypeLinggongToWorker:
		return "岗位推人"
	case model.RecTypeWorkerToLinggong:
		return "人推岗位"
	case model.RecTypeSimilar:
		return "相似岗位"
	case model.RecTypeNearby:
		return "附近岗位"
	case model.RecTypeRecentlyViewed:
		return "看过又推荐"
	case model.RecTypeHot:
		return "热门推荐"
	case model.RecTypeNew:
		return "最新推荐"
	case model.RecTypeBySkill:
		return "技能匹配"
	case model.RecTypeByCategory:
		return "分类匹配"
	}
	return ""
}

// recStatusText 推荐状态文本
func recStatusText(s int) string {
	switch s {
	case model.RecStatusPending:
		return "待展示"
	case model.RecStatusShown:
		return "已展示"
	case model.RecStatusClicked:
		return "已点击"
	case model.RecStatusApplied:
		return "已报名"
	case model.RecStatusDismissed:
		return "已忽略"
	case model.RecStatusExpired:
		return "已过期"
	}
	return ""
}

// toRecommendationInfo model -> dto
func toRecommendationInfo(r *model.LinggongRecommendation) *dto.RecommendationInfo {
	return &dto.RecommendationInfo{
		ID:            r.ID,
		UserID:        r.UserID,
		LinggongID:    r.LinggongID,
		RecType:       r.RecType,
		RecTypeText:   recTypeText(r.RecType),
		Source:        r.Source,
		Score:         r.Score,
		Reason:        r.Reason,
		SalaryMatch:   r.SalaryMatch,
		SkillMatch:    r.SkillMatch,
		LocationMatch: r.LocationMatch,
		TimeMatch:     r.TimeMatch,
		CreditMatch:   r.CreditMatch,
		Status:        r.Status,
		StatusText:    recStatusText(r.Status),
		ClickedAt:     r.ClickedAt,
		AppliedAt:     r.AppliedAt,
		ViewedAt:      r.ViewedAt,
		DismissedAt:   r.DismissedAt,
		ExpiredAt:     r.ExpiredAt,
		RegionID:      r.RegionID,
		CreatedAt:     r.CreatedAt,
	}
}

// genRecNo 生成推荐编号（仅用于日志/调试）：REC + yyyyMMddHHmmss + 6 位随机
func genRecNo() string {
	return fmt.Sprintf("REC%s%06d", time.Now().Format("20060102150405"), rand.Intn(1000000))
}

// ===== C 端 =====

// GetByID 获取推荐详情
func (s *recommendationService) GetByID(id uint) (*dto.RecommendationInfo, error) {
	r, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRecommendationNotFound
		}
		return nil, err
	}
	return toRecommendationInfo(r), nil
}

// List 推荐列表
func (s *recommendationService) List(userID uint, req *dto.RecommendationListRequest) (*utils.Pagination, []dto.RecommendationInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.RecommendationListOptions{
		RecType: req.RecType,
		Status:  req.Status,
	}
	list, total, err := s.repo.List(userID, pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.RecommendationInfo, 0, len(list))
	for i := range list {
		result = append(result, *toRecommendationInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListByUser 按用户反查推荐
func (s *recommendationService) ListByUser(userID uint, page, pageSize int) (*utils.Pagination, []dto.RecommendationInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByUser(userID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.RecommendationInfo, 0, len(list))
	for i := range list {
		result = append(result, *toRecommendationInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListByLinggong 按岗位反查推荐
func (s *recommendationService) ListByLinggong(linggongID uint, page, pageSize int) (*utils.Pagination, []dto.RecommendationInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByLinggong(linggongID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.RecommendationInfo, 0, len(list))
	for i := range list {
		result = append(result, *toRecommendationInfo(&list[i]))
	}
	return pagination, result, nil
}

// Click 推荐点击反馈
func (s *recommendationService) Click(id uint, userID uint) error {
	r, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRecommendationNotFound
		}
		return err
	}
	if r.UserID != userID {
		return ErrRecommendationNotFound
	}
	now := time.Now()
	fields := map[string]interface{}{
		"status":     model.RecStatusClicked,
		"clicked_at": &now,
	}
	return s.repo.Update(id, fields)
}

// Apply 推荐报名反馈
func (s *recommendationService) Apply(id uint, userID uint) error {
	r, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRecommendationNotFound
		}
		return err
	}
	if r.UserID != userID {
		return ErrRecommendationNotFound
	}
	now := time.Now()
	fields := map[string]interface{}{
		"status":     model.RecStatusApplied,
		"applied_at": &now,
	}
	return s.repo.Update(id, fields)
}

// View 推荐查看反馈
func (s *recommendationService) View(id uint, userID uint) error {
	r, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRecommendationNotFound
		}
		return err
	}
	if r.UserID != userID {
		return ErrRecommendationNotFound
	}
	now := time.Now()
	fields := map[string]interface{}{
		"viewed_at": &now,
	}
	if r.Status == model.RecStatusPending {
		fields["status"] = model.RecStatusShown
	}
	return s.repo.Update(id, fields)
}

// Dismiss 推荐忽略反馈
func (s *recommendationService) Dismiss(id uint, userID uint) error {
	r, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRecommendationNotFound
		}
		return err
	}
	if r.UserID != userID {
		return ErrRecommendationNotFound
	}
	now := time.Now()
	fields := map[string]interface{}{
		"status":       model.RecStatusDismissed,
		"dismissed_at": &now,
	}
	return s.repo.Update(id, fields)
}

// ===== M 端 / 内部调用 =====

// Create 创建推荐（由推荐引擎或 M 端手动触发）
func (s *recommendationService) Create(regionID uint, req *model.LinggongRecommendation) (*dto.RecommendationInfo, error) {
	if req.UserID == 0 || req.LinggongID == 0 {
		return nil, errors.New("用户 ID 与岗位 ID 不能为空")
	}
	if req.RecType == "" {
		req.RecType = model.RecTypeLinggongToWorker
	}
	if req.Source == "" {
		req.Source = model.RecSourceAI
	}
	if req.Status == 0 {
		req.Status = model.RecStatusPending
	}
	req.RegionID = regionID
	_ = genRecNo()
	if err := s.repo.Create(req); err != nil {
		return nil, err
	}
	return toRecommendationInfo(req), nil
}

// Delete 删除推荐（M 端管理）
func (s *recommendationService) Delete(id uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRecommendationNotFound
		}
		return err
	}
	return s.repo.Delete(id)
}
