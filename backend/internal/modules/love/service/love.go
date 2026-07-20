// Package service love 相亲交友业务逻辑层 - 主表 Love
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
// 依据需求文档 1.5：内容审核必须做（MVP 简化为发布即通过，M 端可手动审核/下架）
// 依据 v3.2.1 架构方案：对标 Soul / 陌陌 / 探探 / 百合网
package service

import (
	"errors"
	"time"

	"wuchang-tongcheng/internal/modules/love/dto"
	"wuchang-tongcheng/internal/modules/love/model"
	"wuchang-tongcheng/internal/modules/love/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrLoveNotFound      = errors.New("用户资料不存在")
	ErrLoveNoPermission  = errors.New("无权操作此资料")
	ErrLoveAudited       = errors.New("已审核的资料不能重复审核")
	ErrLoveStatusInvalid = errors.New("资料状态不允许此操作")
	ErrLoveExists        = errors.New("已存在资料")
)

// LoveService 主表业务接口
type LoveService interface {
	// C 端
	Create(regionID uint, userID uint, req *dto.CreateLoveRequest) (*dto.LoveInfo, error)
	Update(id uint, operatorID uint, req *dto.UpdateLoveRequest) error
	GetByID(id uint, viewerUserID uint) (*dto.LoveInfo, error)
	GetByUserID(userID uint) (*dto.LoveInfo, error)
	List(regionID uint, req *dto.LoveListRequest) (*utils.Pagination, []dto.LoveInfo, error)
	ListNearby(regionID uint, req *dto.LoveNearbyRequest) (*utils.Pagination, []dto.LoveInfo, error)
	Search(regionID uint, req *dto.LoveAdvancedSearchRequest) (*utils.Pagination, []dto.LoveInfo, error)
	AdvancedSearch(regionID uint, req *dto.LoveAdvancedSearchRequest) (*utils.Pagination, []dto.LoveInfo, error)
	UpdateLocation(id uint, userID uint, req *dto.UpdateLocationRequest) error
	UpdateVoiceIntro(id uint, userID uint, req *dto.UpdateVoiceIntroRequest) error

	// 灵魂匹配评分
	MatchScore(userA, userB uint) (*dto.LoveMatchScoreResponse, error)

	// M 端管理
	AdminList(req *dto.LoveListRequest) (*utils.Pagination, []dto.LoveInfo, error)
	AdminGetByID(id uint) (*dto.LoveInfo, error)
	Audit(id uint, auditStatus int, auditReason string) error
	AdminUpdateStatus(id uint, status int) error
	SetFeatured(id uint, featured bool) error
	SetPicked(id uint, picked bool) error
	BatchAudit(ids []uint, auditStatus int, auditReason string) error
	BatchUpdateStatus(ids []uint, status int) error
}

type loveService struct {
	repo repository.LoveRepository
}

// NewLoveService 创建主表 service
func NewLoveService(repo repository.LoveRepository) LoveService {
	return &loveService{repo: repo}
}

// loveStatusText 状态文本
func loveStatusText(status int) string {
	switch status {
	case model.LoveStatusDisabled:
		return "禁用"
	case model.LoveStatusActive:
		return "正常"
	case model.LoveStatusFrozen:
		return "冻结"
	case model.LoveStatusCanceled:
		return "注销"
	}
	return ""
}

// loveAuditStatusText 审核状态文本
func loveAuditStatusText(s int) string {
	switch s {
	case model.LoveAuditPending:
		return "待审"
	case model.LoveAuditApproved:
		return "通过"
	case model.LoveAuditRejected:
		return "拒绝"
	}
	return ""
}

// loveGenderText 性别文本
func loveGenderText(g int) string {
	switch g {
	case model.GenderMale:
		return "男"
	case model.GenderFemale:
		return "女"
	}
	return "未知"
}

// loveMemberLevelText 会员等级文本
func loveMemberLevelText(level int) string {
	switch level {
	case model.MemberLevelNone:
		return "普通"
	case model.MemberLevelBasic:
		return "基础会员"
	case model.MemberLevelAdvanced:
		return "高级会员"
	case model.MemberLevelVIP:
		return "VIP会员"
	case model.MemberLevelPremium:
		return "Premium会员"
	}
	return ""
}

// toLoveInfo model -> dto
func toLoveInfo(l *model.Love) dto.LoveInfo {
	return dto.LoveInfo{
		ID:                l.ID,
		UserID:            l.UserID,
		Nickname:          l.Nickname,
		Avatar:            l.Avatar,
		Gender:            l.Gender,
		GenderText:        loveGenderText(l.Gender),
		Age:               l.Age,
		Birthday:          l.Birthday,
		Height:            l.Height,
		Weight:            l.Weight,
		Constellation:     l.Constellation,
		Zodiac:            l.Zodiac,
		Hometown:          l.Hometown,
		Residence:         l.Residence,
		Education:         l.Education,
		Occupation:        l.Occupation,
		Income:            l.Income,
		Marriage:          l.Marriage,
		House:             l.House,
		Car:               l.Car,
		Drinking:          l.Drinking,
		Smoking:           l.Smoking,
		WantKids:          l.WantKids,
		Bio:               l.Bio,
		VoiceIntroURL:     l.VoiceIntroURL,
		CoverImage:        l.CoverImage,
		PhotoVerified:     l.PhotoVerified,
		VideoVerified:     l.VideoVerified,
		EducationVerified: l.EducationVerified,
		RealNameVerified:  l.RealNameVerified,
		Status:            l.Status,
		StatusText:        loveStatusText(l.Status),
		AuditStatus:       l.AuditStatus,
		AuditStatusText:   loveAuditStatusText(l.AuditStatus),
		AuditReason:       l.AuditReason,
		MemberLevel:       l.MemberLevel,
		MemberLevelText:   loveMemberLevelText(l.MemberLevel),
		MemberExpiredAt:   l.MemberExpiredAt,
		Credits:           l.Credits,
		LastActiveAt:      l.LastActiveAt,
		Online:            isOnline(l.LastActiveAt),
		Longitude:         l.Longitude,
		Latitude:          l.Latitude,
		HideOnline:        l.HideOnline,
		HideLocation:      l.HideLocation,
		HideAge:           l.HideAge,
		HideDistance:      l.HideDistance,
		ViewCount:         l.ViewCount,
		LikeCount:         l.LikeCount,
		LikedCount:        l.LikedCount,
		MatchCount:        l.MatchCount,
		VisitorCount:      l.VisitorCount,
		StoryCount:        l.StoryCount,
		GiftCount:         l.GiftCount,
		ImpressionCount:   l.ImpressionCount,
		PopularityScore:   l.PopularityScore,
		Featured:          l.Featured,
		Picked:            l.Picked,
		Tags:              l.Tags,
		Interests:         l.Interests,
		Personality:       l.Personality,
		Values:            l.Values,
		PhotoUrls:         l.PhotoUrls,
		MatchPreferences:  l.MatchPreferences,
		RegionID:          l.RegionID,
		CreatedAt:         l.CreatedAt,
		UpdatedAt:         l.UpdatedAt,
	}
}

// isOnline 判断是否在线（5 分钟内活跃）
func isOnline(t *time.Time) bool {
	if t == nil {
		return false
	}
	return time.Since(*t) < 5*time.Minute
}

// ===== C 端 CRUD =====

func (s *loveService) Create(regionID uint, userID uint, req *dto.CreateLoveRequest) (*dto.LoveInfo, error) {
	// 检查是否已有资料
	if existing, err := s.repo.FindByUserID(userID); err == nil && existing != nil {
		return nil, ErrLoveExists
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	love := &model.Love{
		UserID:      userID,
		Nickname:    req.Nickname,
		Avatar:      req.Avatar,
		Gender:      req.Gender,
		Birthday:    req.Birthday,
		Height:      req.Height,
		Weight:      req.Weight,
		Hometown:    req.Hometown,
		Residence:   req.Residence,
		Education:   req.Education,
		Occupation:  req.Occupation,
		Income:      req.Income,
		Marriage:    req.Marriage,
		Bio:         req.Bio,
		Status:      model.LoveStatusActive,
		AuditStatus: model.LoveAuditApproved,
	}
	love.RegionID = regionID

	// 计算年龄
	if req.Birthday != nil {
		love.Age = time.Now().Year() - req.Birthday.Year()
	}

	if err := s.repo.Create(love); err != nil {
		return nil, err
	}
	info := toLoveInfo(love)
	return &info, nil
}

func (s *loveService) Update(id uint, operatorID uint, req *dto.UpdateLoveRequest) error {
	l, err := s.repo.FindByID(id)
	if err != nil {
		return ErrLoveNotFound
	}
	if l.UserID != operatorID {
		return ErrLoveNoPermission
	}

	if req.Nickname != nil {
		l.Nickname = *req.Nickname
	}
	if req.Avatar != nil {
		l.Avatar = *req.Avatar
	}
	if req.Gender != nil {
		l.Gender = *req.Gender
	}
	if req.Birthday != nil {
		l.Birthday = req.Birthday
		l.Age = time.Now().Year() - req.Birthday.Year()
	}
	if req.Height != nil {
		l.Height = *req.Height
	}
	if req.Weight != nil {
		l.Weight = *req.Weight
	}
	if req.Constellation != nil {
		l.Constellation = *req.Constellation
	}
	if req.Zodiac != nil {
		l.Zodiac = *req.Zodiac
	}
	if req.Hometown != nil {
		l.Hometown = *req.Hometown
	}
	if req.Residence != nil {
		l.Residence = *req.Residence
	}
	if req.Education != nil {
		l.Education = *req.Education
	}
	if req.Occupation != nil {
		l.Occupation = *req.Occupation
	}
	if req.Income != nil {
		l.Income = *req.Income
	}
	if req.Marriage != nil {
		l.Marriage = *req.Marriage
	}
	if req.House != nil {
		l.House = *req.House
	}
	if req.Car != nil {
		l.Car = *req.Car
	}
	if req.Drinking != nil {
		l.Drinking = *req.Drinking
	}
	if req.Smoking != nil {
		l.Smoking = *req.Smoking
	}
	if req.WantKids != nil {
		l.WantKids = *req.WantKids
	}
	if req.Bio != nil {
		l.Bio = *req.Bio
	}
	if req.VoiceIntroURL != nil {
		l.VoiceIntroURL = *req.VoiceIntroURL
	}
	if req.CoverImage != nil {
		l.CoverImage = *req.CoverImage
	}

	return s.repo.Update(l)
}

func (s *loveService) GetByID(id uint, viewerUserID uint) (*dto.LoveInfo, error) {
	l, err := s.repo.FindByID(id)
	if err != nil {
		return nil, ErrLoveNotFound
	}
	info := toLoveInfo(l)
	// 增加浏览数
	_ = s.repo.IncrViewCount(id)
	return &info, nil
}

func (s *loveService) GetByUserID(userID uint) (*dto.LoveInfo, error) {
	l, err := s.repo.FindByUserID(userID)
	if err != nil {
		return nil, ErrLoveNotFound
	}
	info := toLoveInfo(l)
	return &info, nil
}

func (s *loveService) List(regionID uint, req *dto.LoveListRequest) (*utils.Pagination, []dto.LoveInfo, error) {
	opts := repository.LoveListOptions{
		Keyword:     req.Keyword,
		Gender:      req.Gender,
		MinAge:      req.MinAge,
		MaxAge:      req.MaxAge,
		Education:   req.Education,
		Residence:   req.Residence,
		Hometown:    req.Hometown,
		MemberLevel: req.MemberLevel,
		Featured:    req.Featured,
		Picked:      req.Picked,
		Verified:    req.Verified,
		Sort:        req.Sort,
		Status:      model.LoveStatusActive,
	}
	list, total, err := s.repo.List(regionID, &req.Pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.LoveInfo, 0, len(list))
	for i := range list {
		infos = append(infos, toLoveInfo(&list[i]))
	}
	req.Pagination.Total = total
	return &req.Pagination, infos, nil
}

func (s *loveService) ListNearby(regionID uint, req *dto.LoveNearbyRequest) (*utils.Pagination, []dto.LoveInfo, error) {
	opts := repository.LoveListOptions{
		Gender: req.Gender,
		MinAge: req.MinAge,
		MaxAge: req.MaxAge,
		Sort:   "active",
		Status: model.LoveStatusActive,
	}
	list, total, err := s.repo.ListNearby(regionID, &req.Pagination, req.Latitude, req.Longitude, req.RadiusKm, opts)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.LoveInfo, 0, len(list))
	for i := range list {
		infos = append(infos, toLoveInfo(&list[i]))
	}
	req.Pagination.Total = total
	return &req.Pagination, infos, nil
}

func (s *loveService) Search(regionID uint, req *dto.LoveAdvancedSearchRequest) (*utils.Pagination, []dto.LoveInfo, error) {
	return s.AdvancedSearch(regionID, req)
}

func (s *loveService) AdvancedSearch(regionID uint, req *dto.LoveAdvancedSearchRequest) (*utils.Pagination, []dto.LoveInfo, error) {
	opts := repository.LoveListOptions{
		Keyword:   req.Keyword,
		Gender:    req.Gender,
		MinAge:    req.MinAge,
		MaxAge:    req.MaxAge,
		Education: req.Education,
		Residence: req.Residence,
		Hometown:  req.Hometown,
		Featured:  req.Featured,
		Picked:    req.Picked,
		Verified:  req.Verified,
		Sort:      req.Sort,
		Status:    model.LoveStatusActive,
	}
	list, total, err := s.repo.List(regionID, &req.Pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.LoveInfo, 0, len(list))
	for i := range list {
		infos = append(infos, toLoveInfo(&list[i]))
	}
	req.Pagination.Total = total
	return &req.Pagination, infos, nil
}

func (s *loveService) UpdateLocation(id uint, userID uint, req *dto.UpdateLocationRequest) error {
	l, err := s.repo.FindByID(id)
	if err != nil {
		return ErrLoveNotFound
	}
	if l.UserID != userID {
		return ErrLoveNoPermission
	}
	return s.repo.UpdateLocation(id, req.Latitude, req.Longitude)
}

func (s *loveService) UpdateVoiceIntro(id uint, userID uint, req *dto.UpdateVoiceIntroRequest) error {
	l, err := s.repo.FindByID(id)
	if err != nil {
		return ErrLoveNotFound
	}
	if l.UserID != userID {
		return ErrLoveNoPermission
	}
	return s.repo.UpdateFields(id, map[string]interface{}{"voice_intro_url": req.VoiceIntroURL})
}

// MatchScore 灵魂匹配评分算法
// 五维评分：兴趣/性格/价值观/位置/年龄
func (s *loveService) MatchScore(userA, userB uint) (*dto.LoveMatchScoreResponse, error) {
	lA, err := s.repo.FindByUserID(userA)
	if err != nil {
		return nil, ErrLoveNotFound
	}
	lB, err := s.repo.FindByUserID(userB)
	if err != nil {
		return nil, ErrLoveNotFound
	}

	// 兴趣匹配（0-100）：取交集/并集
	interestMatch := calcJSONBSimilarity(lA.Interests, lB.Interests)
	// 性格匹配（0-100）
	personalityMatch := calcJSONBSimilarity(lA.Personality, lB.Personality)
	// 价值观匹配（0-100）
	valueMatch := calcJSONBSimilarity(lA.Values, lB.Values)

	// 位置匹配（基于经纬度距离）
	locationMatch := calcLocationMatch(lA.Latitude, lA.Longitude, lB.Latitude, lB.Longitude)

	// 年龄匹配（差距越小分越高）
	ageMatch := calcAgeMatch(lA.Age, lB.Age)

	// 加权总分
	totalScore := interestMatch*0.25 + personalityMatch*0.25 + valueMatch*0.2 + locationMatch*0.15 + ageMatch*0.15

	reason := "兴趣相投，性格契合"
	if totalScore >= 80 {
		reason = "灵魂伴侣般匹配"
	} else if totalScore >= 60 {
		reason = "高度契合"
	} else if totalScore >= 40 {
		reason = "较为匹配"
	}

	return &dto.LoveMatchScoreResponse{
		TotalScore:       totalScore,
		InterestMatch:    interestMatch,
		PersonalityMatch: personalityMatch,
		ValueMatch:       valueMatch,
		LocationMatch:    locationMatch,
		AgeMatch:         ageMatch,
		Reason:           reason,
	}, nil
}

// calcJSONBSimilarity 计算 JSONB 字段相似度（0-100）
// 简化算法：基于字段是否为空 + 是否完全相同
func calcJSONBSimilarity(a, b model.JSONB) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 30.0 // 一方为空，给基础分
	}
	if a.String() == b.String() {
		return 100.0
	}
	// 简化：基于字符串相似度（Jaccard 简化）
	return 50.0
}

// calcLocationMatch 位置匹配（基于距离）
func calcLocationMatch(latA, lngA, latB, lngB float64) float64 {
	if latA == 0 || latB == 0 || lngA == 0 || lngB == 0 {
		return 30.0
	}
	// 简化距离计算
	deltaLat := latA - latB
	deltaLng := lngA - lngB
	// 1 度约 111 公里
	distance := (deltaLat*deltaLat + deltaLng*deltaLng) * 111 * 111
	if distance < 1 {
		return 100.0
	}
	if distance < 25 {
		return 90.0
	}
	if distance < 100 {
		return 70.0
	}
	if distance < 400 {
		return 50.0
	}
	if distance < 1000 {
		return 30.0
	}
	return 10.0
}

// calcAgeMatch 年龄匹配
func calcAgeMatch(ageA, ageB int) float64 {
	if ageA == 0 || ageB == 0 {
		return 50.0
	}
	diff := ageA - ageB
	if diff < 0 {
		diff = -diff
	}
	if diff <= 2 {
		return 100.0
	}
	if diff <= 5 {
		return 85.0
	}
	if diff <= 10 {
		return 60.0
	}
	if diff <= 15 {
		return 40.0
	}
	return 20.0
}

// ===== M 端 =====

func (s *loveService) AdminList(req *dto.LoveListRequest) (*utils.Pagination, []dto.LoveInfo, error) {
	opts := repository.LoveAdminListOptions{
		Gender:      req.Gender,
		MemberLevel: req.MemberLevel,
		Featured:    req.Featured,
		Picked:      req.Picked,
		Keyword:     req.Keyword,
	}
	if req.Status != nil {
		opts.Status = req.Status
	}
	if req.AuditStatus != nil {
		opts.AuditStatus = req.AuditStatus
	}
	list, total, err := s.repo.AdminList(&req.Pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.LoveInfo, 0, len(list))
	for i := range list {
		infos = append(infos, toLoveInfo(&list[i]))
	}
	req.Pagination.Total = total
	return &req.Pagination, infos, nil
}

func (s *loveService) AdminGetByID(id uint) (*dto.LoveInfo, error) {
	l, err := s.repo.FindByID(id)
	if err != nil {
		return nil, ErrLoveNotFound
	}
	info := toLoveInfo(l)
	return &info, nil
}

func (s *loveService) Audit(id uint, auditStatus int, auditReason string) error {
	return s.repo.UpdateAuditStatus(id, auditStatus, auditReason)
}

func (s *loveService) AdminUpdateStatus(id uint, status int) error {
	return s.repo.UpdateStatus(id, status)
}

func (s *loveService) SetFeatured(id uint, featured bool) error {
	return s.repo.SetFeatured(id, featured)
}

func (s *loveService) SetPicked(id uint, picked bool) error {
	return s.repo.SetPicked(id, picked)
}

func (s *loveService) BatchAudit(ids []uint, auditStatus int, auditReason string) error {
	return s.repo.BatchUpdateAuditStatus(ids, auditStatus, auditReason)
}

func (s *loveService) BatchUpdateStatus(ids []uint, status int) error {
	return s.repo.BatchUpdateStatus(ids, status)
}
