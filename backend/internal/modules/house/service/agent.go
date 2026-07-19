// Package service 经纪人业务逻辑层
// 依据 v3.2.1 架构方案第五章：对标贝壳/链家
package service

import (
	"errors"
	"time"

	"wuchang-tongcheng/internal/modules/house/dto"
	"wuchang-tongcheng/internal/modules/house/model"
	"wuchang-tongcheng/internal/modules/house/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrAgentNotFound     = errors.New("经纪人不存在")
	ErrAgentNoPermission = errors.New("无权操作此经纪人")
	ErrAgentExists       = errors.New("经纪人已存在")
)

// AgentService 经纪人业务接口
type AgentService interface {
	// C 端
	Create(regionID uint, userID uint, req *dto.AgentCreateRequest) (*dto.AgentResponse, error)
	Update(id uint, operatorID uint, req *dto.AgentCreateRequest) error
	GetByID(id uint, userID uint) (*dto.AgentResponse, error)
	List(regionID uint, req *dto.AgentListQuery) (*utils.Pagination, []dto.AgentResponse, error)
	GetMine(userID uint) (*dto.AgentResponse, error)

	// 关注
	Follow(userID, agentID uint, notify bool) (*dto.FavResponse, error)
	FollowStatus(userID, agentID uint) (*dto.FavResponse, error)

	// M 端管理
	AdminList(req *dto.AgentAdminListQuery) (*utils.Pagination, []dto.AgentResponse, error)
	Audit(id uint, status int, reason string) error
	UpdateOnlineStatus(userID uint, status int) error
}

type agentService struct {
	repo repository.AgentRepository
}

// NewAgentService 创建 service 实例
func NewAgentService(repo repository.AgentRepository) AgentService {
	return &agentService{repo: repo}
}

// toAgentInfo model -> dto
func toAgentInfo(a *model.HouseAgent) *dto.AgentResponse {
	info := &dto.AgentResponse{
		ID:             a.ID,
		UserID:         a.UserID,
		Name:           a.Name,
		Phone:          a.Phone,
		Avatar:         a.Avatar,
		Gender:         a.Gender,
		StoreID:        a.StoreID,
		StoreName:      a.StoreName,
		Company:        a.Company,
		Title:          a.Title,
		Level:          a.Level,
		LevelText:      agentLevelText(a.Level),
		LicenseNo:      a.LicenseNo,
		LicenseImage:   a.LicenseImage,
		IDCardFront:    a.IDCardFront,
		IDCardBack:     a.IDCardBack,
		BusinessCard:   a.BusinessCard,
		Description:    a.Description,
		Rating:         a.Rating,
		RatingCount:    a.RatingCount,
		ListingCount:   a.ListingCount,
		DealCount:      a.DealCount,
		TotalAmount:    a.TotalAmount,
		ResponseTime:   a.ResponseTime,
		ResponseRate:   a.ResponseRate,
		OnlineStatus:   a.OnlineStatus,
		OnlineText:     agentOnlineText(a.OnlineStatus),
		LastActiveAt:   a.LastActiveAt,
		VerifiedAt:     a.VerifiedAt,
		ApprovedAt:     a.ApprovedAt,
		RejectedReason: a.RejectedReason,
		Status:         a.Status,
		StatusText:     agentStatusText(a.Status),
		FollowerCount:  a.FollowerCount,
		RegionID:       a.RegionID,
		CreatedAt:      a.CreatedAt,
		UpdatedAt:      a.UpdatedAt,
	}
	return info
}

func agentLevelText(level int) string {
	switch level {
	case model.AgentLevelTrainee:
		return "实习"
	case model.AgentLevelJunior:
		return "初级"
	case model.AgentLevelSenior:
		return "高级"
	case model.AgentLevelMaster:
		return "资深"
	case model.AgentLevelPartner:
		return "合伙人"
	}
	return "初级"
}

func agentOnlineText(s int) string {
	switch s {
	case model.AgentOffline:
		return "离线"
	case model.AgentOnline:
		return "在线"
	case model.AgentBusy:
		return "忙碌"
	case model.AgentAway:
		return "离开"
	}
	return "离线"
}

func agentStatusText(s int) string {
	switch s {
	case model.AgentStatusPending:
		return "待审核"
	case model.AgentStatusApproved:
		return "已通过"
	case model.AgentStatusRejected:
		return "已拒绝"
	case model.AgentStatusFrozen:
		return "已冻结"
	case model.AgentStatusRevoked:
		return "已撤销"
	}
	return "待审核"
}

// ===== C 端 =====

func (s *agentService) Create(regionID uint, userID uint, req *dto.AgentCreateRequest) (*dto.AgentResponse, error) {
	// 校验是否已申请
	if existing, err := s.repo.FindByUserID(userID); err == nil && existing != nil {
		return nil, ErrAgentExists
	}

	a := &model.HouseAgent{
		UserID:       userID,
		Name:         req.Name,
		Phone:        req.Phone,
		Avatar:       req.Avatar,
		Gender:       req.Gender,
		StoreID:      req.StoreID,
		StoreName:    req.StoreName,
		Company:      req.Company,
		Title:        req.Title,
		Level:        req.Level,
		LicenseNo:    req.LicenseNo,
		LicenseImage: req.LicenseImage,
		IDCardFront:  req.IDCardFront,
		IDCardBack:   req.IDCardBack,
		BusinessCard: req.BusinessCard,
		Description:  req.Description,
		Rating:       5.0,
		Status:       model.AgentStatusPending,
	}
	a.RegionID = regionID

	if err := s.repo.Create(a); err != nil {
		return nil, err
	}
	return toAgentInfo(a), nil
}

func (s *agentService) Update(id uint, operatorID uint, req *dto.AgentCreateRequest) error {
	a, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrAgentNotFound
		}
		return err
	}
	if a.UserID != operatorID {
		return ErrAgentNoPermission
	}

	fields := map[string]interface{}{
		"name":          req.Name,
		"phone":         req.Phone,
		"avatar":        req.Avatar,
		"gender":        req.Gender,
		"store_id":      req.StoreID,
		"store_name":    req.StoreName,
		"company":       req.Company,
		"title":         req.Title,
		"level":         req.Level,
		"license_no":    req.LicenseNo,
		"license_image": req.LicenseImage,
		"id_card_front": req.IDCardFront,
		"id_card_back":  req.IDCardBack,
		"business_card": req.BusinessCard,
		"description":   req.Description,
	}
	return s.repo.UpdateFields(id, fields)
}

func (s *agentService) GetByID(id uint, userID uint) (*dto.AgentResponse, error) {
	a, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAgentNotFound
		}
		return nil, err
	}
	info := toAgentInfo(a)
	if userID > 0 {
		if followed, err := s.repo.FollowExists(userID, id); err == nil {
			info.HasFollowed = followed
		}
	}
	return info, nil
}

func (s *agentService) List(regionID uint, req *dto.AgentListQuery) (*utils.Pagination, []dto.AgentResponse, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.AgentListOptions{
		City:         req.City,
		Company:      req.Company,
		StoreID:      req.StoreID,
		Level:        req.Level,
		Status:       req.Status,
		OnlineStatus: req.OnlineStatus,
		Keyword:      req.Keyword,
		Sort:         req.Sort,
	}
	if opts.Status == nil {
		approved := model.AgentStatusApproved
		opts.Status = &approved
	}

	list, total, err := s.repo.List(regionID, pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total

	result := make([]dto.AgentResponse, 0, len(list))
	for i := range list {
		result = append(result, *toAgentInfo(&list[i]))
	}
	return pagination, result, nil
}

func (s *agentService) GetMine(userID uint) (*dto.AgentResponse, error) {
	a, err := s.repo.FindByUserID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAgentNotFound
		}
		return nil, err
	}
	return toAgentInfo(a), nil
}

// ===== 关注 =====

func (s *agentService) Follow(userID, agentID uint, notify bool) (*dto.FavResponse, error) {
	if userID == 0 {
		return nil, ErrAgentNoPermission
	}
	a, err := s.repo.FindByID(agentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAgentNotFound
		}
		return nil, err
	}

	exists, err := s.repo.FollowExists(userID, agentID)
	if err != nil {
		return nil, err
	}
	if exists {
		if err := s.repo.DeleteFollow(userID, agentID); err != nil {
			return nil, err
		}
		_ = s.repo.DecrFollowerCount(agentID)
		return &dto.FavResponse{HasFaved: false, FavCount: a.FollowerCount - 1}, nil
	}

	fav := &model.HouseFavorite{
		UserID:       userID,
		AgentID:      agentID,
		FavoriteType: model.FavoriteTypeAgent,
		Notify:       notify,
	}
	if err := s.repo.CreateFollow(fav); err != nil {
		return nil, err
	}
	_ = s.repo.IncrFollowerCount(agentID)
	return &dto.FavResponse{HasFaved: true, FavCount: a.FollowerCount + 1}, nil
}

func (s *agentService) FollowStatus(userID, agentID uint) (*dto.FavResponse, error) {
	a, err := s.repo.FindByID(agentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAgentNotFound
		}
		return nil, err
	}
	if userID == 0 {
		return &dto.FavResponse{HasFaved: false, FavCount: a.FollowerCount}, nil
	}
	exists, err := s.repo.FollowExists(userID, agentID)
	if err != nil {
		return nil, err
	}
	return &dto.FavResponse{HasFaved: exists, FavCount: a.FollowerCount}, nil
}

// ===== M 端 =====

func (s *agentService) AdminList(req *dto.AgentAdminListQuery) (*utils.Pagination, []dto.AgentResponse, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.AgentAdminListOptions{
		RegionID: req.RegionID,
		UserID:   req.UserID,
		Status:   req.Status,
		Level:    req.Level,
		Keyword:  req.Keyword,
	}
	list, total, err := s.repo.AdminList(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total

	result := make([]dto.AgentResponse, 0, len(list))
	for i := range list {
		result = append(result, *toAgentInfo(&list[i]))
	}
	return pagination, result, nil
}

func (s *agentService) Audit(id uint, status int, reason string) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrAgentNotFound
		}
		return err
	}
	fields := map[string]interface{}{
		"status":          status,
		"rejected_reason": reason,
	}
	if status == model.AgentStatusApproved {
		now := time.Now()
		fields["approved_at"] = &now
	}
	return s.repo.UpdateFields(id, fields)
}

func (s *agentService) UpdateOnlineStatus(userID uint, status int) error {
	a, err := s.repo.FindByUserID(userID)
	if err != nil {
		return ErrAgentNotFound
	}
	now := time.Now()
	return s.repo.UpdateFields(a.ID, map[string]interface{}{
		"online_status": status,
		"last_active_at": &now,
	})
}
