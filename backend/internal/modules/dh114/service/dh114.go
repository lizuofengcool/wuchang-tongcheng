// Package service 同城114业务逻辑层 - 商户主表
// 依据 v3.2.1 架构方案：对标大众点评/美团/58同城
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
// 依据需求文档 1.5：内容审核必须做（MVP 简化为发布即通过，M 端可手动审核/下架）
package service

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"wuchang-tongcheng/internal/modules/dh114/dto"
	"wuchang-tongcheng/internal/modules/dh114/model"
	"wuchang-tongcheng/internal/modules/dh114/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrDh114NotFound      = errors.New("商户不存在")
	ErrDh114NoPermission  = errors.New("无权操作此商户")
	ErrDh114Audited       = errors.New("已审核的商户不能重复审核")
	ErrDh114StatusInvalid = errors.New("商户状态不允许此操作")
	ErrDh114HasFaved      = errors.New("已收藏过此商户")
	ErrDh114NotFaved      = errors.New("未收藏此商户")
)

// Dh114Service 商户主表业务接口
type Dh114Service interface {
	// C 端
	Create(regionID uint, userID uint, userName string, userPhone string, userAvatar string, req *dto.CreateDh114Request) (*dto.Dh114Info, error)
	Update(id uint, operatorID uint, req *dto.UpdateDh114Request) error
	Delete(id uint, operatorID uint) error
	GetByID(id uint, userID uint) (*dto.Dh114Info, error)
	List(regionID uint, req *dto.Dh114ListRequest) (*utils.Pagination, []dto.Dh114Info, error)
	ListNearby(regionID uint, req *dto.Dh114NearbyRequest) (*utils.Pagination, []dto.Dh114Info, error)
	Search(regionID uint, req *dto.Dh114SearchRequest) (*utils.Pagination, []dto.Dh114Info, error)
	ListMine(userID uint, page, pageSize int) (*utils.Pagination, []dto.Dh114Info, error)
	AdvancedSearch(regionID uint, req *dto.AdvancedSearchRequest) (*utils.Pagination, []dto.Dh114Info, error)

	// 收藏
	Fav(userID, dh114ID uint) (*dto.FavResponse, error)
	Unfav(userID, dh114ID uint) (*dto.FavResponse, error)
	FavStatus(userID, dh114ID uint) (*dto.FavResponse, error)
	ListFavs(userID uint, req *dto.FavoriteListRequest) (*utils.Pagination, []dto.Dh114Info, error)

	// 互动
	IncrContact(id uint) error
	IncrShare(id uint) error
	RecordCall(userID uint, dh114ID uint, phone string, req *dto.PhoneCallRequest, ip, userAgent string) (*dto.CallResponse, error)
	RecordView(userID uint, ip string, req *dto.Dh114ViewRequest) error

	// M 端管理
	AdminList(req *dto.Dh114AdminListRequest) (*utils.Pagination, []dto.Dh114Info, error)
	AdminGetByID(id uint) (*dto.Dh114Info, error)
	Audit(id uint, auditStatus int, auditReason string) error
	BatchAudit(req *dto.BatchAuditRequest) (*dto.BatchResultResponse, error)
	AdminUpdateStatus(id uint, status int) error
	UpdatePromotion(id uint, req *dto.PromotionRequest) error
}

type dh114Service struct {
	repo          repository.Dh114Repository
	imageRepo     repository.ImageRepository
	visitRepo     repository.VisitRepository
	favoriteRepo  repository.FavoriteRepository
	phoneCallRepo repository.PhoneCallRepository
}

// NewDh114Service 创建商户 service 实例
func NewDh114Service(
	repo repository.Dh114Repository,
	imageRepo repository.ImageRepository,
	visitRepo repository.VisitRepository,
	favoriteRepo repository.FavoriteRepository,
	phoneCallRepo repository.PhoneCallRepository,
) Dh114Service {
	return &dh114Service{
		repo:          repo,
		imageRepo:     imageRepo,
		visitRepo:     visitRepo,
		favoriteRepo:  favoriteRepo,
		phoneCallRepo: phoneCallRepo,
	}
}

// dh114StatusText 状态文本
func dh114StatusText(status int) string {
	switch status {
	case model.StatusDraft:
		return "草稿"
	case model.StatusPublished:
		return "已发布"
	case model.StatusOffline:
		return "已下架"
	case model.StatusExpired:
		return "已过期"
	case model.StatusDeleted:
		return "已删除"
	}
	return ""
}

// dh114AuditStatusText 审核状态文本
func dh114AuditStatusText(s int) string {
	switch s {
	case model.AuditPending:
		return "待审"
	case model.AuditApproved:
		return "通过"
	case model.AuditRejected:
		return "拒绝"
	}
	return ""
}

// toDh114Info model -> dto
func toDh114Info(d *model.Dh114) *dto.Dh114Info {
	info := &dto.Dh114Info{
		ID:              d.ID,
		Title:           d.Title,
		Content:         d.Content,
		CoverImage:      d.CoverImage,
		UserID:          d.UserID,
		UserName:        d.UserName,
		UserPhone:       d.UserPhone,
		UserAvatar:      d.UserAvatar,
		Status:          d.Status,
		StatusText:      dh114StatusText(d.Status),
		AuditStatus:     d.AuditStatus,
		AuditStatusText: dh114AuditStatusText(d.AuditStatus),
		AuditReason:     d.AuditReason,
		PublishedAt:     d.PublishedAt,

		CategoryID:   d.CategoryID,
		CategoryName: d.CategoryName,
		BusinessType: d.BusinessType,
		SourceType:   d.SourceType,

		Phone:    d.Phone,
		AltPhone: d.AltPhone,
		Website:  d.Website,
		Wechat:   d.Wechat,

		City:             d.City,
		District:         d.District,
		BusinessDistrict: d.BusinessDistrict,
		Address:          d.Address,
		Latitude:         d.Latitude,
		Longitude:        d.Longitude,
		Distance:         d.Distance,

		Rating:      d.Rating,
		ReviewCount: d.ReviewCount,
		PriceAvg:    d.PriceAvg,

		ViewCount:    d.ViewCount,
		FavCount:     d.FavCount,
		ContactCount: d.ContactCount,
		ShareCount:   d.ShareCount,
		CallCount:    d.CallCount,
		LastCallAt:   d.LastCallAt,

		ContentHash: d.ContentHash,
		RiskScore:   d.RiskScore,

		VideoURL:   d.VideoURL,
		VideoCover: d.VideoCover,
		VRURL:      d.VRURL,

		Featured:       d.Featured,
		Picked:         d.Picked,
		Verified:       d.Verified,
		PromotionLevel: d.PromotionLevel,
		TrafficWeight:  d.TrafficWeight,
		VerifiedAt:     d.VerifiedAt,

		RegionID:  d.RegionID,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}

	// JSONB 字段透传
	if d.Images != nil {
		info.Images = d.Images
	}
	if d.Tags != nil {
		info.Tags = d.Tags
	}
	if d.BusinessHours != nil {
		info.BusinessHours = d.BusinessHours
	}
	if d.Features != nil {
		info.Features = d.Features
	}
	return info
}

// toDh114InfoWithFav 带收藏状态的转换
func toDh114InfoWithFav(d *model.Dh114, hasFaved bool) *dto.Dh114Info {
	info := toDh114Info(d)
	info.HasFaved = hasFaved
	return info
}

// generateDh114No 生成商户相关业务单号
func generateDh114No(prefix string) string {
	return fmt.Sprintf("%s%s%06d", prefix, time.Now().Format("20060102150405"), rand.Intn(1000000))
}

// Create 发布商户
func (s *dh114Service) Create(regionID uint, userID uint, userName string, userPhone string, userAvatar string, req *dto.CreateDh114Request) (*dto.Dh114Info, error) {
	d := &model.Dh114{
		Title:           req.Title,
		Content:         req.Content,
		CoverImage:      req.CoverImage,
		UserID:          userID,
		UserName:        userName,
		UserPhone:       userPhone,
		UserAvatar:      userAvatar,
		CategoryID:      req.CategoryID,
		BusinessType:    req.BusinessType,
		SourceType:      req.SourceType,
		Phone:           req.Phone,
		AltPhone:        req.AltPhone,
		Website:         req.Website,
		Wechat:          req.Wechat,
		City:            req.City,
		District:        req.District,
		BusinessDistrict: req.BusinessDistrict,
		Address:         req.Address,
		Latitude:        req.Latitude,
		Longitude:       req.Longitude,
		VideoURL:        req.VideoURL,
		VideoCover:      req.VideoCover,
		VRURL:           req.VRURL,
		Status:          req.Status,
		AuditStatus:     model.AuditApproved, // MVP：发布即通过
	}
	d.RegionID = regionID

	// JSONB 字段处理
	if req.Images != nil {
		if b, err := model.FromJSON(req.Images); err == nil {
			d.Images = b
		}
	}
	if req.Tags != nil {
		if b, err := model.FromJSON(req.Tags); err == nil {
			d.Tags = b
		}
	}
	if req.BusinessHours != nil {
		if b, err := model.FromJSON(req.BusinessHours); err == nil {
			d.BusinessHours = b
		}
	}
	if req.Features != nil {
		if b, err := model.FromJSON(req.Features); err == nil {
			d.Features = b
		}
	}

	// 默认值
	if d.BusinessType == "" {
		d.BusinessType = model.BusinessTypeOther
	}
	if d.SourceType == "" {
		d.SourceType = model.SourceTypePersonal
	}
	if d.Status == 1 {
		now := time.Now()
		d.PublishedAt = &now
	}

	if err := s.repo.Create(d); err != nil {
		return nil, err
	}
	return toDh114Info(d), nil
}

// Update 更新商户（仅发布者本人）
func (s *dh114Service) Update(id uint, operatorID uint, req *dto.UpdateDh114Request) error {
	d, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrDh114NotFound
		}
		return err
	}
	if d.UserID != operatorID {
		return ErrDh114NoPermission
	}

	fields := make(map[string]interface{})
	if req.Title != nil {
		fields["title"] = *req.Title
	}
	if req.Content != nil {
		fields["content"] = *req.Content
	}
	if req.CoverImage != nil {
		fields["cover_image"] = *req.CoverImage
	}
	if req.CategoryID != nil {
		fields["category_id"] = *req.CategoryID
	}
	if req.CategoryName != nil {
		fields["category_name"] = *req.CategoryName
	}
	if req.BusinessType != nil {
		fields["business_type"] = *req.BusinessType
	}
	if req.SourceType != nil {
		fields["source_type"] = *req.SourceType
	}
	if req.Phone != nil {
		fields["phone"] = *req.Phone
	}
	if req.AltPhone != nil {
		fields["alt_phone"] = *req.AltPhone
	}
	if req.Website != nil {
		fields["website"] = *req.Website
	}
	if req.Wechat != nil {
		fields["wechat"] = *req.Wechat
	}
	if req.City != nil {
		fields["city"] = *req.City
	}
	if req.District != nil {
		fields["district"] = *req.District
	}
	if req.BusinessDistrict != nil {
		fields["business_district"] = *req.BusinessDistrict
	}
	if req.Address != nil {
		fields["address"] = *req.Address
	}
	if req.Latitude != nil {
		fields["latitude"] = *req.Latitude
	}
	if req.Longitude != nil {
		fields["longitude"] = *req.Longitude
	}
	if req.VideoURL != nil {
		fields["video_url"] = *req.VideoURL
	}
	if req.VideoCover != nil {
		fields["video_cover"] = *req.VideoCover
	}
	if req.VRURL != nil {
		fields["vr_url"] = *req.VRURL
	}
	if req.Status != nil {
		fields["status"] = *req.Status
		// 状态变更为已发布时记录发布时间
		if *req.Status == model.StatusPublished && d.PublishedAt == nil {
			now := time.Now()
			fields["published_at"] = &now
		}
	}

	// JSONB 字段处理
	if req.Images != nil {
		if b, err := model.FromJSON(req.Images); err == nil {
			fields["images"] = b
		}
	}
	if req.Tags != nil {
		if b, err := model.FromJSON(req.Tags); err == nil {
			fields["tags"] = b
		}
	}
	if req.BusinessHours != nil {
		if b, err := model.FromJSON(req.BusinessHours); err == nil {
			fields["business_hours"] = b
		}
	}
	if req.Features != nil {
		if b, err := model.FromJSON(req.Features); err == nil {
			fields["features"] = b
		}
	}

	if len(fields) == 0 {
		return nil
	}
	return s.repo.UpdateFields(id, fields)
}

// Delete 删除商户（仅发布者本人）
func (s *dh114Service) Delete(id uint, operatorID uint) error {
	d, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrDh114NotFound
		}
		return err
	}
	if d.UserID != operatorID {
		return ErrDh114NoPermission
	}
	return s.repo.Delete(id)
}

// GetByID 获取商户详情（同时增加浏览量）
func (s *dh114Service) GetByID(id uint, userID uint) (*dto.Dh114Info, error) {
	d, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDh114NotFound
		}
		return nil, err
	}

	// 异步增加浏览量（这里同步执行，简单可靠）
	_ = s.repo.IncrViewCount(id)

	// 检查是否已收藏
	hasFaved := false
	if userID > 0 {
		hasFaved, _ = s.favoriteRepo.Exists(userID, id, model.FavoriteTypeBusiness)
	}
	return toDh114InfoWithFav(d, hasFaved), nil
}

// List 商户列表
func (s *dh114Service) List(regionID uint, req *dto.Dh114ListRequest) (*utils.Pagination, []dto.Dh114Info, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.Dh114ListOptions{
		CategoryID:   req.CategoryID,
		BusinessType: req.BusinessType,
		SourceType:   req.SourceType,
		City:         req.City,
		District:     req.District,
		MinPrice:     req.MinPrice,
		MaxPrice:     req.MaxPrice,
		MinRating:    req.MinRating,
		Featured:     req.Featured,
		Picked:       req.Picked,
		Verified:     req.Verified,
		Keyword:      req.Keyword,
		Sort:         req.Sort,
		Status:       model.StatusPublished,
	}
	list, total, err := s.repo.List(regionID, pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.Dh114Info, 0, len(list))
	for i := range list {
		infos = append(infos, *toDh114Info(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// ListNearby 附近商户
func (s *dh114Service) ListNearby(regionID uint, req *dto.Dh114NearbyRequest) (*utils.Pagination, []dto.Dh114Info, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.Dh114ListOptions{
		CategoryID:   req.CategoryID,
		BusinessType: req.BusinessType,
		Keyword:      req.Keyword,
		Sort:         req.Sort,
		Status:       model.StatusPublished,
	}
	list, total, err := s.repo.ListNearby(regionID, pagination, req.Latitude, req.Longitude, req.RadiusKm, opts)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.Dh114Info, 0, len(list))
	for i := range list {
		infos = append(infos, *toDh114Info(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// Search 搜索商户
func (s *dh114Service) Search(regionID uint, req *dto.Dh114SearchRequest) (*utils.Pagination, []dto.Dh114Info, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	list, total, err := s.repo.Search(regionID, pagination, req.Keyword)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.Dh114Info, 0, len(list))
	for i := range list {
		infos = append(infos, *toDh114Info(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// ListMine 我的商户
func (s *dh114Service) ListMine(userID uint, page, pageSize int) (*utils.Pagination, []dto.Dh114Info, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByUser(userID, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.Dh114Info, 0, len(list))
	for i := range list {
		infos = append(infos, *toDh114Info(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// AdvancedSearch 高级搜索
func (s *dh114Service) AdvancedSearch(regionID uint, req *dto.AdvancedSearchRequest) (*utils.Pagination, []dto.Dh114Info, error) {
	// 如果有经纬度和半径，使用附近搜索
	if req.Latitude != 0 && req.Longitude != 0 && req.RadiusKm > 0 {
		nearbyReq := &dto.Dh114NearbyRequest{
			Latitude:    req.Latitude,
			Longitude:   req.Longitude,
			RadiusKm:    req.RadiusKm,
			CategoryID:  req.CategoryID,
			BusinessType: req.BusinessType,
			Keyword:     req.Keyword,
			Sort:        req.Sort,
		}
		nearbyReq.Page = req.Page
		nearbyReq.PageSize = req.PageSize
		return s.ListNearby(regionID, nearbyReq)
	}

	// 否则使用普通列表
	listReq := &dto.Dh114ListRequest{
		Keyword:      req.Keyword,
		CategoryID:   req.CategoryID,
		BusinessType: req.BusinessType,
		SourceType:   req.SourceType,
		City:         req.City,
		District:     req.District,
		MinPrice:     req.MinPrice,
		MaxPrice:     req.MaxPrice,
		MinRating:    req.MinRating,
		Featured:     req.Featured,
		Picked:       req.Picked,
		Verified:     req.Verified,
		Sort:         req.Sort,
	}
	listReq.Page = req.Page
	listReq.PageSize = req.PageSize
	return s.List(regionID, listReq)
}

// Fav 收藏商户
func (s *dh114Service) Fav(userID, dh114ID uint) (*dto.FavResponse, error) {
	d, err := s.repo.FindByID(dh114ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDh114NotFound
		}
		return nil, err
	}

	// 检查是否已收藏
	exists, err := s.favoriteRepo.Exists(userID, dh114ID, model.FavoriteTypeBusiness)
	if err != nil {
		return nil, err
	}
	if exists {
		return &dto.FavResponse{HasFaved: true, FavCount: d.FavCount}, nil
	}

	// 创建收藏
	fav := &model.Dh114Favorite{
		UserID:       userID,
		Dh114ID:      dh114ID,
		BusinessID:   0,
		FavoriteType: model.FavoriteTypeBusiness,
	}
	if err := s.favoriteRepo.Create(fav); err != nil {
		return nil, err
	}

	// 增加商户收藏计数
	_ = s.repo.IncrFavCount(dh114ID)

	return &dto.FavResponse{HasFaved: true, FavCount: d.FavCount + 1}, nil
}

// Unfav 取消收藏
func (s *dh114Service) Unfav(userID, dh114ID uint) (*dto.FavResponse, error) {
	d, err := s.repo.FindByID(dh114ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDh114NotFound
		}
		return nil, err
	}

	if err := s.favoriteRepo.DeleteByUserAndTarget(userID, dh114ID, model.FavoriteTypeBusiness); err != nil {
		return nil, err
	}

	// 减少商户收藏计数
	_ = s.repo.DecrFavCount(dh114ID)

	newCount := d.FavCount
	if newCount > 0 {
		newCount--
	}
	return &dto.FavResponse{HasFaved: false, FavCount: newCount}, nil
}

// FavStatus 收藏状态
func (s *dh114Service) FavStatus(userID, dh114ID uint) (*dto.FavResponse, error) {
	d, err := s.repo.FindByID(dh114ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDh114NotFound
		}
		return nil, err
	}
	hasFaved := false
	if userID > 0 {
		hasFaved, _ = s.favoriteRepo.Exists(userID, dh114ID, model.FavoriteTypeBusiness)
	}
	return &dto.FavResponse{HasFaved: hasFaved, FavCount: d.FavCount}, nil
}

// ListFavs 我的收藏
func (s *dh114Service) ListFavs(userID uint, req *dto.FavoriteListRequest) (*utils.Pagination, []dto.Dh114Info, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	favType := req.FavoriteType
	if favType == "" {
		favType = model.FavoriteTypeBusiness
	}
	favs, total, err := s.favoriteRepo.ListByType(userID, favType, pagination)
	if err != nil {
		return nil, nil, err
	}
	// 收集 dh114_id 并查询商户信息
	infos := make([]dto.Dh114Info, 0, len(favs))
	for _, fav := range favs {
		d, err := s.repo.FindByID(fav.Dh114ID)
		if err != nil {
			continue
		}
		info := toDh114InfoWithFav(d, true)
		info.ID = fav.ID // 返回收藏记录 ID，便于取消收藏
		infos = append(infos, *info)
	}
	pagination.Total = total
	return pagination, infos, nil
}

// IncrContact 增加联系数
func (s *dh114Service) IncrContact(id uint) error {
	return s.repo.IncrContactCount(id)
}

// IncrShare 增加分享数
func (s *dh114Service) IncrShare(id uint) error {
	return s.repo.IncrShareCount(id)
}

// RecordCall 记录电话拨打（一键拨号核心）
func (s *dh114Service) RecordCall(userID uint, dh114ID uint, phone string, req *dto.PhoneCallRequest, ip, userAgent string) (*dto.CallResponse, error) {
	d, err := s.repo.FindByID(dh114ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDh114NotFound
		}
		return nil, err
	}

	callType := req.CallType
	if callType == "" {
		callType = model.CallTypeClick
	}

	callerPhone := ""
	callerName := ""
	if userID > 0 {
		callerName = d.UserName
		callerPhone = d.UserPhone
	}

	call := &model.Dh114PhoneCall{
		CallNo:      generateDh114No("DH114CALL"),
		Dh114ID:     dh114ID,
		BusinessID:  0,
		Phone:       phone,
		CallerID:    userID,
		CallerPhone: callerPhone,
		CallerName:  callerName,
		CallType:    callType,
		Device:      req.Device,
		IP:          ip,
		UserAgent:   userAgent,
		Status:      model.CallStatusSuccess,
		CalledAt:    time.Now(),
	}
	call.RegionID = d.RegionID

	if err := s.phoneCallRepo.Create(call); err != nil {
		return nil, err
	}

	// 增加商户拨打计数
	_ = s.repo.IncrCallCount(dh114ID)

	// 更新最近拨打时间
	now := time.Now()
	_ = s.repo.UpdateFields(dh114ID, map[string]interface{}{
		"last_call_at": &now,
	})

	return &dto.CallResponse{
		CallNo:    call.CallNo,
		Phone:     phone,
		CallCount: d.CallCount + 1,
	}, nil
}

// RecordView 记录浏览
func (s *dh114Service) RecordView(userID uint, ip string, req *dto.Dh114ViewRequest) error {
	visit := &model.Dh114Visit{
		UserID:     userID,
		Dh114ID:    req.Dh114ID,
		VisitType:  req.VisitType,
		IP:         ip,
		Device:     req.Device,
		Source:     req.Source,
		Duration:   req.Duration,
		Longitude:  req.Longitude,
		Latitude:   req.Latitude,
	}
	visit.RegionID = 1 // 默认地区 ID，由 Region 中间件注入会覆盖
	if visit.VisitType == "" {
		visit.VisitType = "business"
	}

	if err := s.visitRepo.Create(visit); err != nil {
		return err
	}

	// 增加商户浏览计数
	return s.repo.IncrViewCount(req.Dh114ID)
}

// AdminList 管理后台列表
func (s *dh114Service) AdminList(req *dto.Dh114AdminListRequest) (*utils.Pagination, []dto.Dh114Info, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.Dh114AdminListOptions{
		RegionID:     req.RegionID,
		UserID:       req.UserID,
		CategoryID:   req.CategoryID,
		Status:       req.Status,
		AuditStatus:  req.AuditStatus,
		BusinessType: req.BusinessType,
		SourceType:   req.SourceType,
		Keyword:      req.Keyword,
	}
	list, total, err := s.repo.AdminList(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.Dh114Info, 0, len(list))
	for i := range list {
		infos = append(infos, *toDh114Info(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// AdminGetByID 管理后台获取详情
func (s *dh114Service) AdminGetByID(id uint) (*dto.Dh114Info, error) {
	d, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDh114NotFound
		}
		return nil, err
	}
	return toDh114Info(d), nil
}

// Audit 审核
func (s *dh114Service) Audit(id uint, auditStatus int, auditReason string) error {
	d, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrDh114NotFound
		}
		return err
	}
	// 已审核的不能重复审核
	if d.AuditStatus != model.AuditPending && auditStatus != d.AuditStatus {
		// 允许重新审核，不报错
	}
	fields := map[string]interface{}{
		"audit_status": auditStatus,
		"audit_reason": auditReason,
	}
	// 审核通过且未发布过，自动发布
	if auditStatus == model.AuditApproved && d.PublishedAt == nil {
		now := time.Now()
		fields["published_at"] = &now
		if d.Status == model.StatusDraft {
			fields["status"] = model.StatusPublished
		}
	}
	return s.repo.UpdateFields(id, fields)
}

// BatchAudit 批量审核
func (s *dh114Service) BatchAudit(req *dto.BatchAuditRequest) (*dto.BatchResultResponse, error) {
	result := &dto.BatchResultResponse{Total: len(req.IDs)}
	failedIDs := make([]uint, 0)
	for _, id := range req.IDs {
		if err := s.Audit(id, req.AuditStatus, req.AuditReason); err != nil {
			failedIDs = append(failedIDs, id)
		} else {
			result.Success++
		}
	}
	result.Failed = len(failedIDs)
	result.FailedIDs = failedIDs
	return result, nil
}

// AdminUpdateStatus 管理后台更新状态
func (s *dh114Service) AdminUpdateStatus(id uint, status int) error {
	fields := map[string]interface{}{"status": status}
	// 状态变更为已发布时记录发布时间
	if status == model.StatusPublished {
		d, err := s.repo.FindByID(id)
		if err == nil && d.PublishedAt == nil {
			now := time.Now()
			fields["published_at"] = &now
		}
	}
	return s.repo.UpdateFields(id, fields)
}

// UpdatePromotion 更新推广配置
func (s *dh114Service) UpdatePromotion(id uint, req *dto.PromotionRequest) error {
	fields := make(map[string]interface{})
	if req.Featured != nil {
		fields["featured"] = *req.Featured
	}
	if req.Picked != nil {
		fields["picked"] = *req.Picked
	}
	if req.Verified != nil {
		fields["verified"] = *req.Verified
		// 认证时间
		if *req.Verified {
			now := time.Now()
			fields["verified_at"] = &now
		}
	}
	if req.PromotionLevel != nil {
		fields["promotion_level"] = *req.PromotionLevel
	}
	if req.TrafficWeight != nil {
		fields["traffic_weight"] = *req.TrafficWeight
	}
	if len(fields) == 0 {
		return nil
	}
	return s.repo.UpdateFields(id, fields)
}
