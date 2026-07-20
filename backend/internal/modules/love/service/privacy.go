// Package service love 相亲交友业务逻辑层 - 隐私设置
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
package service

import (
	"errors"

	"wuchang-tongcheng/internal/modules/love/dto"
	"wuchang-tongcheng/internal/modules/love/model"
	"wuchang-tongcheng/internal/modules/love/repository"
)

var (
	ErrLovePrivacyNotFound = errors.New("隐私设置不存在")
	ErrLovePrivacyNoPermission = errors.New("无权操作此隐私设置")
)

// LovePrivacyService 隐私设置业务接口
type LovePrivacyService interface {
	// C 端
	Get(userID uint) (*dto.LovePrivacySettingInfo, error)
	GetByLoveID(loveID uint) (*dto.LovePrivacySettingInfo, error)
	Update(userID, loveID uint, req *dto.UpdateLovePrivacySettingRequest) (*dto.LovePrivacySettingInfo, error)
	Reset(userID uint) error
	// 内部调用：其他模块判断当前用户的可见性配置
	CanBeSeenBy(viewerUserID uint, targetUserID uint) (bool, error)
	IsVisible(field string, targetUserID uint) (bool, error)
}

type lovePrivacyService struct {
	repo repository.LovePrivacySettingRepository
}

// NewLovePrivacyService 创建隐私设置 service
func NewLovePrivacyService(repo repository.LovePrivacySettingRepository) LovePrivacyService {
	return &lovePrivacyService{repo: repo}
}

// toLovePrivacySettingInfo model -> dto
func toLovePrivacySettingInfo(p *model.LovePrivacySetting) dto.LovePrivacySettingInfo {
	return dto.LovePrivacySettingInfo{
		ID:                   p.ID,
		UserID:               p.UserID,
		LoveID:               p.LoveID,
		HideOnline:           p.HideOnline,
		HideLocation:         p.HideLocation,
		HideAge:              p.HideAge,
		HideDistance:         p.HideDistance,
		HideConstellation:   p.HideConstellation,
		HideHometown:        p.HideHometown,
		HideOccupation:      p.HideOccupation,
		HideIncome:          p.HideIncome,
		HideLastActive:      p.HideLastActive,
		HideVisitors:        p.HideVisitors,
		OnlyVerifiedCanSee:   p.OnlyVerifiedCanSee,
		OnlyVerifiedCanMatch: p.OnlyVerifiedCanMatch,
		OnlyMemberCanChat:    p.OnlyMemberCanChat,
		BlockStrangers:      p.BlockStrangers,
		BlockSameCity:       p.BlockSameCity,
		AllowPhoneLookup:    p.AllowPhoneLookup,
		AllowContactImport:  p.AllowContactImport,
		AllowRecommendation: p.AllowRecommendation,
		AllowStory:          p.AllowStory,
		AllowImpression:     p.AllowImpression,
		DistanceVisibility:  p.DistanceVisibility,
		AgeVisibility:       p.AgeVisibility,
		Status:              p.Status,
	}
}

// defaultPrivacySetting 默认隐私设置
func defaultPrivacySetting(userID, loveID uint) *model.LovePrivacySetting {
	return &model.LovePrivacySetting{
		UserID:               userID,
		LoveID:               loveID,
		HideLastActive:       true,
		AllowRecommendation: true,
		AllowStory:           true,
		AllowImpression:      true,
		Status:               1,
	}
}

// Get 获取用户隐私设置（不存在则返回默认值）
func (s *lovePrivacyService) Get(userID uint) (*dto.LovePrivacySettingInfo, error) {
	p, err := s.repo.FindByUserID(userID)
	if err != nil {
		// 不存在返回默认值
		info := toLovePrivacySettingInfo(defaultPrivacySetting(userID, 0))
		return &info, nil
	}
	info := toLovePrivacySettingInfo(p)
	return &info, nil
}

func (s *lovePrivacyService) GetByLoveID(loveID uint) (*dto.LovePrivacySettingInfo, error) {
	p, err := s.repo.FindByLoveID(loveID)
	if err != nil {
		info := toLovePrivacySettingInfo(defaultPrivacySetting(0, loveID))
		return &info, nil
	}
	info := toLovePrivacySettingInfo(p)
	return &info, nil
}

// Update 更新隐私设置（不存在则创建）
func (s *lovePrivacyService) Update(userID, loveID uint, req *dto.UpdateLovePrivacySettingRequest) (*dto.LovePrivacySettingInfo, error) {
	p, err := s.repo.FindByUserID(userID)
	if err != nil || p == nil {
		// 不存在则创建
		p = defaultPrivacySetting(userID, loveID)
	}
	if req.HideOnline != nil {
		p.HideOnline = *req.HideOnline
	}
	if req.HideLocation != nil {
		p.HideLocation = *req.HideLocation
	}
	if req.HideAge != nil {
		p.HideAge = *req.HideAge
	}
	if req.HideDistance != nil {
		p.HideDistance = *req.HideDistance
	}
	if req.HideConstellation != nil {
		p.HideConstellation = *req.HideConstellation
	}
	if req.HideHometown != nil {
		p.HideHometown = *req.HideHometown
	}
	if req.HideOccupation != nil {
		p.HideOccupation = *req.HideOccupation
	}
	if req.HideIncome != nil {
		p.HideIncome = *req.HideIncome
	}
	if req.HideLastActive != nil {
		p.HideLastActive = *req.HideLastActive
	}
	if req.HideVisitors != nil {
		p.HideVisitors = *req.HideVisitors
	}
	if req.OnlyVerifiedCanSee != nil {
		p.OnlyVerifiedCanSee = *req.OnlyVerifiedCanSee
	}
	if req.OnlyVerifiedCanMatch != nil {
		p.OnlyVerifiedCanMatch = *req.OnlyVerifiedCanMatch
	}
	if req.OnlyMemberCanChat != nil {
		p.OnlyMemberCanChat = *req.OnlyMemberCanChat
	}
	if req.BlockStrangers != nil {
		p.BlockStrangers = *req.BlockStrangers
	}
	if req.BlockSameCity != nil {
		p.BlockSameCity = *req.BlockSameCity
	}
	if req.AllowPhoneLookup != nil {
		p.AllowPhoneLookup = *req.AllowPhoneLookup
	}
	if req.AllowContactImport != nil {
		p.AllowContactImport = *req.AllowContactImport
	}
	if req.AllowRecommendation != nil {
		p.AllowRecommendation = *req.AllowRecommendation
	}
	if req.AllowStory != nil {
		p.AllowStory = *req.AllowStory
	}
	if req.AllowImpression != nil {
		p.AllowImpression = *req.AllowImpression
	}
	if req.DistanceVisibility != nil {
		p.DistanceVisibility = *req.DistanceVisibility
	}
	if req.AgeVisibility != nil {
		p.AgeVisibility = *req.AgeVisibility
	}
	if err := s.repo.Upsert(p); err != nil {
		return nil, err
	}
	info := toLovePrivacySettingInfo(p)
	return &info, nil
}

// Reset 重置为默认
func (s *lovePrivacyService) Reset(userID uint) error {
	p := defaultPrivacySetting(userID, 0)
	return s.repo.Upsert(p)
}

// CanBeSeenBy 判断 viewer 是否能看到 target 用户
// 简化判断：仅根据 target 的 OnlyVerifiedCanSee 配置（MVP 不接入认证体系）
func (s *lovePrivacyService) CanBeSeenBy(viewerUserID uint, targetUserID uint) (bool, error) {
	p, err := s.repo.FindByUserID(targetUserID)
	if err != nil || p == nil {
		return true, nil
	}
	if p.OnlyVerifiedCanSee {
		// 实际应查询 viewer 是否已认证，MVP 简化返回 true
		return true, nil
	}
	return true, nil
}

// IsVisible 判断 target 用户的某个字段是否可见
// field 取值：online/location/age/distance/constellation/hometown/occupation/income/last_active/visitors
func (s *lovePrivacyService) IsVisible(field string, targetUserID uint) (bool, error) {
	p, err := s.repo.FindByUserID(targetUserID)
	if err != nil || p == nil {
		return true, nil
	}
	switch field {
	case "online":
		return !p.HideOnline, nil
	case "location":
		return !p.HideLocation, nil
	case "age":
		return !p.HideAge, nil
	case "distance":
		return !p.HideDistance, nil
	case "constellation":
		return !p.HideConstellation, nil
	case "hometown":
		return !p.HideHometown, nil
	case "occupation":
		return !p.HideOccupation, nil
	case "income":
		return !p.HideIncome, nil
	case "last_active":
		return !p.HideLastActive, nil
	case "visitors":
		return !p.HideVisitors, nil
	}
	return true, nil
}
