// Package service 同城房屋租售主表业务逻辑层
// 依据 v3.2.1 架构方案第五章：对标贝壳/链家
// 依据需求文档 1.5：内容审核必须做（MVP 简化为发布即通过，M 端可手动审核/下架）
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
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
	ErrHouseNotFound     = errors.New("房源不存在")
	ErrHouseNoPermission = errors.New("无权操作此房源")
	ErrHouseAudited      = errors.New("已审核的房源不能重复审核")
)

// HouseService 房源主表业务接口
type HouseService interface {
	// C 端
	Create(regionID uint, userID uint, userName string, userPhone string, userAvatar string, req *dto.CreateHouseRequest) (*dto.HouseInfo, error)
	Update(id uint, operatorID uint, req *dto.UpdateHouseRequest) error
	Delete(id uint, operatorID uint) error
	GetByID(id uint, userID uint) (*dto.HouseDetailResponse, error)
	List(regionID uint, req *dto.HouseListRequest) (*utils.Pagination, []dto.HouseInfo, error)
	ListNearby(regionID uint, req *dto.HouseNearbyRequest) (*utils.Pagination, []dto.HouseInfo, error)
	Search(regionID uint, req *dto.HouseSearchRequest) (*utils.Pagination, []dto.HouseInfo, error)
	AdvancedSearch(regionID uint, req *dto.HouseAdvancedSearchRequest) (*utils.Pagination, []dto.HouseInfo, error)
	ListMine(userID uint, page, pageSize int) (*utils.Pagination, []dto.HouseInfo, error)

	// 收藏
	Fav(userID, houseID uint) (*dto.FavResponse, error)
	FavStatus(userID, houseID uint) (*dto.FavResponse, error)
	ListFavs(userID uint, page, pageSize int) (*utils.Pagination, []dto.HouseInfo, error)

	// 互动
	IncrContactCount(id uint) error
	IncrShareCount(id uint) error

	// 相似推荐
	ListSimilar(houseID uint, limit int) ([]dto.SimilarHouseResponse, error)

	// M 端管理
	AdminList(req *dto.HouseAdminListRequest) (*utils.Pagination, []dto.HouseInfo, error)
	AdminGetByID(id uint) (*dto.HouseDetailResponse, error)
	Audit(id uint, auditStatus int, auditReason string) error
	AdminUpdateStatus(id uint, status int) error
	UpdatePromotion(id uint, req *dto.HousePromotionRequest) error

	// 批量操作
	BatchAudit(ids []uint, auditStatus int, auditReason string) (*dto.BatchResultResponse, error)
	BatchUpdateStatus(ids []uint, status int) (*dto.BatchResultResponse, error)
	BatchDelete(ids []uint) (*dto.BatchResultResponse, error)
}

type houseService struct {
	repo      repository.HouseRepository
	agentRepo repository.AgentRepository
	communityRepo repository.CommunityRepository
}

// NewHouseService 创建 service 实例
func NewHouseService(repo repository.HouseRepository, agentRepo repository.AgentRepository, communityRepo repository.CommunityRepository) HouseService {
	return &houseService{repo: repo, agentRepo: agentRepo, communityRepo: communityRepo}
}

// toHouseInfo model -> dto
func toHouseInfo(h *model.House, images []dto.HouseImageInfo) *dto.HouseInfo {
	info := &dto.HouseInfo{
		ID:          h.ID,
		Title:       h.Title,
		Content:     h.Content,
		CoverImage:  h.CoverImage,
		Images:      images,
		UserID:      h.UserID,
		UserName:    h.UserName,
		UserPhone:   h.UserPhone,
		UserAvatar:  h.UserAvatar,
		Status:      h.Status,
		AuditStatus: h.AuditStatus,
		AuditReason: h.AuditReason,
		PublishedAt: h.PublishedAt,
		RegionID:    h.RegionID,
		CreatedAt:   h.CreatedAt,
		UpdatedAt:   h.UpdatedAt,
		ListingType:  h.ListingType,
		PropertyType: h.PropertyType,
		SourceType:   h.SourceType,
		RentPrice:      h.RentPrice,
		RentUnit:       h.RentUnit,
		RentType:       h.RentType,
		DepositType:    h.DepositType,
		PaymentMethod:  h.PaymentMethod,
		RentNegotiable: h.RentNegotiable,
		RentMinMonths:  h.RentMinMonths,
		RentMaxMonths:  h.RentMaxMonths,
		SalePrice:      h.SalePrice,
		SaleNegotiable: h.SaleNegotiable,
		AveragePrice:   h.AveragePrice,
		OriginalPrice:  h.OriginalPrice,
		Rooms:     h.Rooms,
		Halls:     h.Halls,
		Bathrooms: h.Bathrooms,
		Kitchens:  h.Kitchens,
		Balconies: h.Balconies,
		Layout:    h.Layout,
		BuildingArea: h.BuildingArea,
		InnerArea:    h.InnerArea,
		PoolRatio:    h.PoolRatio,
		UsableArea:   h.UsableArea,
		Floor:       h.Floor,
		TotalFloor:  h.TotalFloor,
		FloorType:   h.FloorType,
		Orientation: h.Orientation,
		HasElevator: h.HasElevator,
		Decoration:        h.Decoration,
		PropertyOwnership: h.PropertyOwnership,
		PropertyYears:     h.PropertyYears,
		BuildingYear:      h.BuildingYear,
		BuildingAge:       h.BuildingAge,
		CommunityID: h.CommunityID,
		AgentID:     h.AgentID,
		CategoryID:  h.CategoryID,
		City:             h.City,
		District:         h.District,
		BusinessDistrict: h.BusinessDistrict,
		Address:          h.Address,
		Latitude:         h.Latitude,
		Longitude:        h.Longitude,
		ViewCount:     h.ViewCount,
		FavCount:      h.FavCount,
		ContactCount:  h.ContactCount,
		ShareCount:    h.ShareCount,
		ViewingCount:  h.ViewingCount,
		LastViewingAt: h.LastViewingAt,
		VideoURL:    h.VideoURL,
		VideoCover:  h.VideoCover,
		VRURL:       h.VRURL,
		PanoramaURL: h.PanoramaURL,
		Featured:       h.Featured,
		Picked:         h.Picked,
		Verified:       h.Verified,
		PromotionLevel: h.PromotionLevel,
		TrafficWeight:  h.TrafficWeight,
		RealHouseVerified:   h.RealHouseVerified,
		RealHouseVerifiedAt: h.RealHouseVerifiedAt,
		ContentHash: h.ContentHash,
		RiskScore:   h.RiskScore,
		SameHouseID: h.SameHouseID,
		Distance:    h.Distance,
	}
	// JSONB 字段反序列化
	if len(h.Facilities) > 0 {
		_ = h.Facilities.UnmarshalJSON(h.Facilities)
		var facs []model.HouseFacilityItem
		if err := h.Facilities.UnmarshalJSON(h.Facilities); err == nil {
			_ = facs
		}
	}
	if info.ListingType == "" {
		info.ListingType = model.ListingTypeRent
	}
	if info.PropertyType == "" {
		info.PropertyType = model.PropertyTypeResidential
	}
	if info.SourceType == "" {
		info.SourceType = model.SourceTypePersonal
	}
	if info.RentUnit == "" {
		info.RentUnit = model.RentUnitMonth
	}
	if info.RentType == "" {
		info.RentType = model.RentTypeEntire
	}
	if info.DepositType == "" {
		info.DepositType = model.DepositTypeOneMonth
	}
	if info.PaymentMethod == "" {
		info.PaymentMethod = model.PaymentMethodMonthly
	}
	if info.FloorType == "" {
		info.FloorType = model.FloorTypeMid
	}
	if info.Decoration == "" {
		info.Decoration = model.DecorationRough
	}
	if info.PropertyOwnership == "" {
		info.PropertyOwnership = model.PropertyOwnershipCommercial
	}
	if info.Images == nil {
		info.Images = []dto.HouseImageInfo{}
	}
	return info
}

// toHouseImageInfo model.HouseImage -> dto.HouseImageInfo
func toHouseImageInfo(img *model.HouseImage) *dto.HouseImageInfo {
	return &dto.HouseImageInfo{
		ID:        img.ID,
		URL:       img.URL,
		Thumbnail: img.Thumbnail,
		ImageType: img.ImageType,
		Title:     img.Title,
		Sort:      img.Sort,
		IsCover:   img.IsCover,
	}
}

// ===== C 端 =====

// Create 发布房源
func (s *houseService) Create(regionID uint, userID uint, userName string, userPhone string, userAvatar string, req *dto.CreateHouseRequest) (*dto.HouseInfo, error) {
	h := &model.House{
		Title:         req.Title,
		Content:       req.Content,
		CoverImage:    req.CoverImage,
		UserID:        userID,
		UserName:      userName,
		UserPhone:     userPhone,
		UserAvatar:    userAvatar,
		Status:        req.Status,
		ListingType:   req.ListingType,
		PropertyType:  req.PropertyType,
		SourceType:    req.SourceType,
		RentPrice:     req.RentPrice,
		RentUnit:      req.RentUnit,
		RentType:      req.RentType,
		DepositType:   req.DepositType,
		PaymentMethod: req.PaymentMethod,
		RentNegotiable: req.RentNegotiable,
		RentMinMonths:  req.RentMinMonths,
		RentMaxMonths:  req.RentMaxMonths,
		SalePrice:      req.SalePrice,
		SaleNegotiable: req.SaleNegotiable,
		OriginalPrice:  req.OriginalPrice,
		Rooms:     req.Rooms,
		Halls:     req.Halls,
		Bathrooms: req.Bathrooms,
		Kitchens:  req.Kitchens,
		Balconies: req.Balconies,
		BuildingArea: req.BuildingArea,
		InnerArea:    req.InnerArea,
		PoolRatio:    req.PoolRatio,
		UsableArea:   req.UsableArea,
		Floor:       req.Floor,
		TotalFloor:  req.TotalFloor,
		FloorType:   req.FloorType,
		Orientation: req.Orientation,
		HasElevator: req.HasElevator,
		Decoration:        req.Decoration,
		PropertyOwnership: req.PropertyOwnership,
		PropertyYears:     req.PropertyYears,
		BuildingYear:      req.BuildingYear,
		CommunityID: req.CommunityID,
		AgentID:     req.AgentID,
		CategoryID:  req.CategoryID,
		City:             req.City,
		District:         req.District,
		BusinessDistrict: req.BusinessDistrict,
		Address:          req.Address,
		Latitude:         req.Latitude,
		Longitude:        req.Longitude,
		VideoURL:    req.VideoURL,
		VideoCover:  req.VideoCover,
		VRURL:       req.VRURL,
		PanoramaURL: req.PanoramaURL,
		// MVP：发布即通过
		AuditStatus: model.AuditApproved,
	}
	h.RegionID = regionID

	// 默认值兜底
	if h.ListingType == "" {
		h.ListingType = model.ListingTypeRent
	}
	if h.PropertyType == "" {
		h.PropertyType = model.PropertyTypeResidential
	}
	if h.SourceType == "" {
		h.SourceType = model.SourceTypePersonal
	}
	if h.RentUnit == "" {
		h.RentUnit = model.RentUnitMonth
	}
	if h.RentType == "" {
		h.RentType = model.RentTypeEntire
	}
	if h.DepositType == "" {
		h.DepositType = model.DepositTypeOneMonth
	}
	if h.PaymentMethod == "" {
		h.PaymentMethod = model.PaymentMethodMonthly
	}
	if h.FloorType == "" {
		h.FloorType = model.FloorTypeMid
	}
	if h.Decoration == "" {
		h.Decoration = model.DecorationRough
	}
	if h.PropertyOwnership == "" {
		h.PropertyOwnership = model.PropertyOwnershipCommercial
	}

	// 户型文本
	if h.Layout == "" {
		h.Layout = buildLayout(h.Rooms, h.Halls, h.Bathrooms)
	}

	// 计算均价
	if h.BuildingArea > 0 && h.SalePrice > 0 {
		h.AveragePrice = h.SalePrice / h.BuildingArea
	}

	// 发布时间
	if req.Status == model.StatusPublished {
		now := time.Now()
		h.PublishedAt = &now
	}

	if err := s.repo.Create(h); err != nil {
		return nil, err
	}

	// 保存图片子表
	if len(req.Images) > 0 {
		imgs := make([]model.HouseImage, 0, len(req.Images))
		for i, img := range req.Images {
			m := model.HouseImage{
				HouseID:   h.ID,
				URL:       img.URL,
				Thumbnail: img.Thumbnail,
				ImageType: img.ImageType,
				Title:     img.Title,
				Sort:      img.Sort,
				IsCover:   img.IsCover,
				Status:    model.ImageStatusEnabled,
			}
			if img.ImageType == "" {
				m.ImageType = model.ImageTypeReal
			}
			if i == 0 && h.CoverImage == "" {
				m.IsCover = true
			}
			imgs = append(imgs, m)
		}
		_ = s.repo.ReplaceImages(h.ID, imgs)
	}

	return toHouseInfo(h, nil), nil
}

// buildLayout 拼装户型文本（3室2厅2卫）
func buildLayout(rooms, halls, bathrooms int) string {
	return fmtLayout(rooms, halls, bathrooms)
}

// fmtLayout 实际拼装
func fmtLayout(rooms, halls, bathrooms int) string {
	s := ""
	if rooms > 0 {
		s += itoa(rooms) + "室"
	}
	if halls > 0 {
		s += itoa(halls) + "厅"
	}
	if bathrooms > 0 {
		s += itoa(bathrooms) + "卫"
	}
	return s
}

// itoa 简单整数转字符串
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	buf := [20]byte{}
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}

// Update 更新房源（仅发布者本人）
func (s *houseService) Update(id uint, operatorID uint, req *dto.UpdateHouseRequest) error {
	h, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrHouseNotFound
		}
		return err
	}
	if h.UserID != operatorID {
		return ErrHouseNoPermission
	}

	fields := map[string]interface{}{}
	if req.Title != "" {
		fields["title"] = req.Title
	}
	if req.Content != "" {
		fields["content"] = req.Content
	}
	if req.CoverImage != "" {
		fields["cover_image"] = req.CoverImage
	}
	if req.ListingType != "" {
		fields["listing_type"] = req.ListingType
	}
	if req.PropertyType != "" {
		fields["property_type"] = req.PropertyType
	}
	if req.SourceType != "" {
		fields["source_type"] = req.SourceType
	}
	if req.RentPrice > 0 || req.RentPrice == 0 {
		fields["rent_price"] = req.RentPrice
	}
	if req.RentUnit != "" {
		fields["rent_unit"] = req.RentUnit
	}
	if req.RentType != "" {
		fields["rent_type"] = req.RentType
	}
	if req.DepositType != "" {
		fields["deposit_type"] = req.DepositType
	}
	if req.PaymentMethod != "" {
		fields["payment_method"] = req.PaymentMethod
	}
	if req.RentNegotiable != nil {
		fields["rent_negotiable"] = *req.RentNegotiable
	}
	fields["rent_min_months"] = req.RentMinMonths
	fields["rent_max_months"] = req.RentMaxMonths
	fields["sale_price"] = req.SalePrice
	if req.SaleNegotiable != nil {
		fields["sale_negotiable"] = *req.SaleNegotiable
	}
	fields["original_price"] = req.OriginalPrice
	fields["rooms"] = req.Rooms
	fields["halls"] = req.Halls
	fields["bathrooms"] = req.Bathrooms
	fields["kitchens"] = req.Kitchens
	fields["balconies"] = req.Balconies
	if req.Rooms > 0 || req.Halls > 0 || req.Bathrooms > 0 {
		fields["layout"] = buildLayout(req.Rooms, req.Halls, req.Bathrooms)
	}
	fields["building_area"] = req.BuildingArea
	fields["inner_area"] = req.InnerArea
	fields["pool_ratio"] = req.PoolRatio
	fields["usable_area"] = req.UsableArea
	fields["floor"] = req.Floor
	fields["total_floor"] = req.TotalFloor
	if req.FloorType != "" {
		fields["floor_type"] = req.FloorType
	}
	if req.Orientation != "" {
		fields["orientation"] = req.Orientation
	}
	if req.HasElevator != nil {
		fields["has_elevator"] = *req.HasElevator
	}
	if req.Decoration != "" {
		fields["decoration"] = req.Decoration
	}
	if req.PropertyOwnership != "" {
		fields["property_ownership"] = req.PropertyOwnership
	}
	fields["property_years"] = req.PropertyYears
	fields["building_year"] = req.BuildingYear
	if req.CommunityID != nil {
		fields["community_id"] = req.CommunityID
	}
	if req.AgentID != nil {
		fields["agent_id"] = req.AgentID
	}
	if req.CategoryID != nil {
		fields["category_id"] = req.CategoryID
	}
	if req.City != "" {
		fields["city"] = req.City
	}
	if req.District != "" {
		fields["district"] = req.District
	}
	if req.BusinessDistrict != "" {
		fields["business_district"] = req.BusinessDistrict
	}
	if req.Address != "" {
		fields["address"] = req.Address
	}
	fields["latitude"] = req.Latitude
	fields["longitude"] = req.Longitude
	if req.VideoURL != "" {
		fields["video_url"] = req.VideoURL
	}
	if req.VideoCover != "" {
		fields["video_cover"] = req.VideoCover
	}
	if req.VRURL != "" {
		fields["vr_url"] = req.VRURL
	}
	if req.PanoramaURL != "" {
		fields["panorama_url"] = req.PanoramaURL
	}

	// 重新计算均价
	if req.BuildingArea > 0 && req.SalePrice > 0 {
		fields["average_price"] = req.SalePrice / req.BuildingArea
	}

	// 状态变更
	if req.Status == model.StatusPublished && h.Status != model.StatusPublished {
		now := time.Now()
		fields["status"] = model.StatusPublished
		fields["published_at"] = &now
		fields["audit_status"] = model.AuditApproved
	} else if req.Status == model.StatusDraft || req.Status == model.StatusOffline || req.Status == model.StatusExpired {
		fields["status"] = req.Status
	}

	if len(fields) == 0 && req.Images == nil {
		return nil
	}

	if err := s.repo.UpdateFields(id, fields); err != nil {
		return err
	}

	// 更新图片子表
	if req.Images != nil {
		imgs := make([]model.HouseImage, 0, len(req.Images))
		for i, img := range req.Images {
			m := model.HouseImage{
				HouseID:   id,
				URL:       img.URL,
				Thumbnail: img.Thumbnail,
				ImageType: img.ImageType,
				Title:     img.Title,
				Sort:      img.Sort,
				IsCover:   img.IsCover,
				Status:    model.ImageStatusEnabled,
			}
			if img.ImageType == "" {
				m.ImageType = model.ImageTypeReal
			}
			if i == 0 {
				m.IsCover = true
			}
			imgs = append(imgs, m)
		}
		if err := s.repo.ReplaceImages(id, imgs); err != nil {
			return err
		}
	}

	return nil
}

// Delete 删除房源（仅发布者本人）
func (s *houseService) Delete(id uint, operatorID uint) error {
	h, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrHouseNotFound
		}
		return err
	}
	if h.UserID != operatorID {
		return ErrHouseNoPermission
	}
	// 同步删除图片
	_ = s.repo.DeleteImages(id)
	return s.repo.Delete(id)
}

// GetByID 获取详情（同时增加浏览量）
func (s *houseService) GetByID(id uint, userID uint) (*dto.HouseDetailResponse, error) {
	h, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrHouseNotFound
		}
		return nil, err
	}

	// 增加浏览量
	_ = s.repo.IncrViewCount(id)
	h.ViewCount++

	// 拼装图片
	images := []dto.HouseImageInfo{}
	if imgs, err := s.repo.ListImages(id); err == nil {
		for i := range imgs {
			images = append(images, *toHouseImageInfo(&imgs[i]))
		}
	}

	info := toHouseInfo(h, images)

	// 当前用户是否已收藏
	if userID > 0 {
		if hasFaved, err := s.repo.FavExists(userID, id); err == nil {
			info.HasFaved = hasFaved
		}
	}

	// 记录浏览历史
	if userID > 0 {
		_ = s.repo.CreateView(&model.HouseView{
			UserID:   userID,
			HouseID:  id,
			ViewType: model.FavoriteTypeHouse,
			Source:   model.ViewSourceList,
			RegionID: h.RegionID,
		})
	}

	// 拼装经纪人/小区
	resp := &dto.HouseDetailResponse{HouseInfo: *info}
	if h.AgentID != nil && *h.AgentID > 0 {
		if agent, err := s.agentRepo.FindByID(*h.AgentID); err == nil {
			resp.Agent = toAgentInfo(agent)
		}
	}
	if h.CommunityID != nil && *h.CommunityID > 0 {
		if community, err := s.communityRepo.FindByID(*h.CommunityID); err == nil {
			resp.Community = toCommunityInfo(community)
		}
	}

	return resp, nil
}

// List 列表查询（C 端）
func (s *houseService) List(regionID uint, req *dto.HouseListRequest) (*utils.Pagination, []dto.HouseInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.HouseListOptions{
		CategoryID:        req.CategoryID,
		CommunityID:       req.CommunityID,
		AgentID:           req.AgentID,
		Keyword:           req.Keyword,
		ListingType:       req.ListingType,
		PropertyType:      req.PropertyType,
		SourceType:        req.SourceType,
		RentType:          req.RentType,
		MinRentPrice:      req.MinRentPrice,
		MaxRentPrice:      req.MaxRentPrice,
		MinSalePrice:      req.MinSalePrice,
		MaxSalePrice:      req.MaxSalePrice,
		MinBuildingArea:   req.MinBuildingArea,
		MaxBuildingArea:   req.MaxBuildingArea,
		Rooms:             req.Rooms,
		FloorType:         req.FloorType,
		Orientation:       req.Orientation,
		Decoration:        req.Decoration,
		HasElevator:       req.HasElevator,
		Featured:          req.Featured,
		Verified:          req.Verified,
		RealHouseVerified: req.RealHouseVerified,
		Sort:              req.Sort,
		Status:            model.StatusPublished,
	}

	list, total, err := s.repo.List(regionID, pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total

	result := make([]dto.HouseInfo, 0, len(list))
	for i := range list {
		result = append(result, *toHouseInfo(&list[i], nil))
	}
	return pagination, result, nil
}

// ListNearby 附近房源
func (s *houseService) ListNearby(regionID uint, req *dto.HouseNearbyRequest) (*utils.Pagination, []dto.HouseInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	radiusKm := req.RadiusKm
	if radiusKm <= 0 {
		radiusKm = 5
	}

	opts := repository.HouseListOptions{
		ListingType: req.ListingType,
		Status:      model.StatusPublished,
	}

	list, total, err := s.repo.ListNearby(regionID, pagination, req.Latitude, req.Longitude, radiusKm, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total

	result := make([]dto.HouseInfo, 0, len(list))
	for i := range list {
		result = append(result, *toHouseInfo(&list[i], nil))
	}
	return pagination, result, nil
}

// Search 关键词搜索
func (s *houseService) Search(regionID uint, req *dto.HouseSearchRequest) (*utils.Pagination, []dto.HouseInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	list, total, err := s.repo.Search(regionID, pagination, req.Keyword)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total

	result := make([]dto.HouseInfo, 0, len(list))
	for i := range list {
		result = append(result, *toHouseInfo(&list[i], nil))
	}
	return pagination, result, nil
}

// AdvancedSearch 高级搜索
func (s *houseService) AdvancedSearch(regionID uint, req *dto.HouseAdvancedSearchRequest) (*utils.Pagination, []dto.HouseInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.HouseAdvancedSearchOptions{
		HouseListOptions: repository.HouseListOptions{
			CategoryID:        req.CategoryID,
			CommunityID:       req.CommunityID,
			Keyword:           req.Keyword,
			ListingType:       req.ListingType,
			PropertyType:      req.PropertyType,
			RentType:          req.RentType,
			MinRentPrice:      req.MinRentPrice,
			MaxRentPrice:      req.MaxRentPrice,
			MinSalePrice:      req.MinSalePrice,
			MaxSalePrice:      req.MaxSalePrice,
			MinBuildingArea:   req.MinBuildingArea,
			MaxBuildingArea:   req.MaxBuildingArea,
			Rooms:             req.Rooms,
			FloorType:         req.FloorType,
			Orientation:       req.Orientation,
			Decoration:        req.Decoration,
			HasElevator:       req.HasElevator,
			Featured:          req.Featured,
			Verified:          req.Verified,
			Sort:              req.Sort,
			Status:            model.StatusPublished,
		},
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		RadiusKm:  req.RadiusKm,
	}

	list, total, err := s.repo.AdvancedSearch(regionID, pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total

	result := make([]dto.HouseInfo, 0, len(list))
	for i := range list {
		result = append(result, *toHouseInfo(&list[i], nil))
	}
	return pagination, result, nil
}

// ListMine 我的发布
func (s *houseService) ListMine(userID uint, page, pageSize int) (*utils.Pagination, []dto.HouseInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByUser(userID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.HouseInfo, 0, len(list))
	for i := range list {
		result = append(result, *toHouseInfo(&list[i], nil))
	}
	return pagination, result, nil
}

// ===== 收藏 =====

func (s *houseService) Fav(userID, houseID uint) (*dto.FavResponse, error) {
	if userID == 0 {
		return nil, ErrHouseNoPermission
	}
	h, err := s.repo.FindByID(houseID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrHouseNotFound
		}
		return nil, err
	}

	exists, err := s.repo.FavExists(userID, houseID)
	if err != nil {
		return nil, err
	}
	if exists {
		if err := s.repo.DeleteFav(userID, houseID); err != nil {
			return nil, err
		}
		_ = s.repo.DecrFavCount(houseID)
		return &dto.FavResponse{HasFaved: false, FavCount: h.FavCount - 1}, nil
	}

	fav := &model.HouseFavorite{
		UserID:       userID,
		HouseID:      houseID,
		FavoriteType: model.FavoriteTypeHouse,
		Notify:       true,
	}
	if err := s.repo.CreateFav(fav); err != nil {
		return nil, err
	}
	_ = s.repo.IncrFavCount(houseID)
	return &dto.FavResponse{HasFaved: true, FavCount: h.FavCount + 1}, nil
}

func (s *houseService) FavStatus(userID, houseID uint) (*dto.FavResponse, error) {
	h, err := s.repo.FindByID(houseID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrHouseNotFound
		}
		return nil, err
	}
	if userID == 0 {
		return &dto.FavResponse{HasFaved: false, FavCount: h.FavCount}, nil
	}
	exists, err := s.repo.FavExists(userID, houseID)
	if err != nil {
		return nil, err
	}
	return &dto.FavResponse{HasFaved: exists, FavCount: h.FavCount}, nil
}

func (s *houseService) ListFavs(userID uint, page, pageSize int) (*utils.Pagination, []dto.HouseInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	favs, total, err := s.repo.ListFavs(userID, page, pageSize)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total

	ids := make([]uint, 0, len(favs))
	for _, f := range favs {
		ids = append(ids, f.HouseID)
	}

	result := make([]dto.HouseInfo, 0, len(favs))
	for _, id := range ids {
		h, err := s.repo.FindByID(id)
		if err != nil {
			continue
		}
		info := toHouseInfo(h, nil)
		info.HasFaved = true
		result = append(result, *info)
	}
	return pagination, result, nil
}

// ===== 互动 =====

func (s *houseService) IncrContactCount(id uint) error {
	return s.repo.IncrContactCount(id)
}

func (s *houseService) IncrShareCount(id uint) error {
	return s.repo.IncrShareCount(id)
}

// ===== 相似推荐 =====

func (s *houseService) ListSimilar(houseID uint, limit int) ([]dto.SimilarHouseResponse, error) {
	if limit <= 0 {
		limit = 5
	}
	h, err := s.repo.FindByID(houseID)
	if err != nil {
		return nil, ErrHouseNotFound
	}

	// 基于小区/分类/价格区间查找相似房源
	pagination := utils.NewPagination(1, limit)
	opts := repository.HouseListOptions{
		Status:       model.StatusPublished,
		ListingType:  h.ListingType,
		PropertyType: h.PropertyType,
	}
	if h.CommunityID != nil && *h.CommunityID > 0 {
		opts.CommunityID = *h.CommunityID
	} else if h.CategoryID != nil && *h.CategoryID > 0 {
		opts.CategoryID = *h.CategoryID
	}

	list, _, err := s.repo.List(h.RegionID, pagination, opts)
	if err != nil {
		return nil, err
	}

	result := make([]dto.SimilarHouseResponse, 0, len(list))
	for i := range list {
		if list[i].ID == houseID {
			continue
		}
		var price float64
		if h.ListingType == model.ListingTypeRent {
			price = list[i].RentPrice
		} else {
			price = list[i].SalePrice
		}
		// 简单相似度计算
		similarity := 0.5
		if h.CommunityID != nil && list[i].CommunityID != nil && *h.CommunityID == *list[i].CommunityID {
			similarity = 1.0
		} else if h.CategoryID != nil && list[i].CategoryID != nil && *h.CategoryID == *list[i].CategoryID {
			similarity = 0.8
		}
		result = append(result, dto.SimilarHouseResponse{
			HouseID:      list[i].ID,
			Title:        list[i].Title,
			CoverImage:   list[i].CoverImage,
			Price:        price,
			Layout:       list[i].Layout,
			BuildingArea: list[i].BuildingArea,
			Similarity:   similarity,
		})
		if len(result) >= limit {
			break
		}
	}
	return result, nil
}

// ===== M 端管理 =====

func (s *houseService) AdminList(req *dto.HouseAdminListRequest) (*utils.Pagination, []dto.HouseInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.HouseAdminListOptions{
		RegionID:    req.RegionID,
		UserID:      req.UserID,
		CategoryID:  req.CategoryID,
		CommunityID: req.CommunityID,
		ListingType: req.ListingType,
		Status:      req.Status,
		AuditStatus: req.AuditStatus,
		Keyword:     req.Keyword,
	}

	list, total, err := s.repo.AdminList(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total

	result := make([]dto.HouseInfo, 0, len(list))
	for i := range list {
		result = append(result, *toHouseInfo(&list[i], nil))
	}
	return pagination, result, nil
}

func (s *houseService) AdminGetByID(id uint) (*dto.HouseDetailResponse, error) {
	h, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrHouseNotFound
		}
		return nil, err
	}

	images := []dto.HouseImageInfo{}
	if imgs, err := s.repo.ListImages(id); err == nil {
		for i := range imgs {
			images = append(images, *toHouseImageInfo(&imgs[i]))
		}
	}
	info := toHouseInfo(h, images)

	resp := &dto.HouseDetailResponse{HouseInfo: *info}
	if h.AgentID != nil && *h.AgentID > 0 {
		if agent, err := s.agentRepo.FindByID(*h.AgentID); err == nil {
			resp.Agent = toAgentInfo(agent)
		}
	}
	if h.CommunityID != nil && *h.CommunityID > 0 {
		if community, err := s.communityRepo.FindByID(*h.CommunityID); err == nil {
			resp.Community = toCommunityInfo(community)
		}
	}
	return resp, nil
}

func (s *houseService) Audit(id uint, auditStatus int, auditReason string) error {
	h, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrHouseNotFound
		}
		return err
	}
	if h.AuditStatus == auditStatus && auditStatus != model.AuditPending {
		return ErrHouseAudited
	}
	return s.repo.UpdateFields(id, map[string]interface{}{
		"audit_status": auditStatus,
		"audit_reason": auditReason,
	})
}

func (s *houseService) AdminUpdateStatus(id uint, status int) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrHouseNotFound
		}
		return err
	}
	fields := map[string]interface{}{"status": status}
	if status == model.StatusPublished {
		now := time.Now()
		fields["published_at"] = &now
	}
	return s.repo.UpdateFields(id, fields)
}

func (s *houseService) UpdatePromotion(id uint, req *dto.HousePromotionRequest) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrHouseNotFound
		}
		return err
	}
	fields := map[string]interface{}{
		"promotion_level":    req.PromotionLevel,
		"traffic_weight":     req.TrafficWeight,
		"featured":           req.Featured,
		"picked":             req.Picked,
		"verified":           req.Verified,
		"real_house_verified": req.RealHouseVerified,
	}
	if req.RealHouseVerified {
		now := time.Now()
		fields["real_house_verified_at"] = &now
	}
	return s.repo.UpdateFields(id, fields)
}

// ===== 批量操作 =====

func (s *houseService) BatchAudit(ids []uint, auditStatus int, auditReason string) (*dto.BatchResultResponse, error) {
	affected, err := s.repo.BatchAudit(ids, auditStatus, auditReason)
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

func (s *houseService) BatchUpdateStatus(ids []uint, status int) (*dto.BatchResultResponse, error) {
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

func (s *houseService) BatchDelete(ids []uint) (*dto.BatchResultResponse, error) {
	affected, err := s.repo.BatchDelete(ids)
	if err != nil {
		return &dto.BatchResultResponse{
			Total: len(ids), Success: 0, Failed: len(ids), FailedIDs: ids,
		}, err
	}
	// 同步删除图片
	for _, id := range ids {
		_ = s.repo.DeleteImages(id)
	}
	return &dto.BatchResultResponse{
		Total:   len(ids),
		Success: int(affected),
		Failed:  len(ids) - int(affected),
	}, nil
}
