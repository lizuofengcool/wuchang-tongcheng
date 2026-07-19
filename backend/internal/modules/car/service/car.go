// Package service 同城车辆买卖业务逻辑层 - 车源主表
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
// 依据需求文档 1.5：内容审核必须做（MVP 简化为发布即通过，M 端可手动审核/下架）
// 依据 v3.2.1 架构方案：对标瓜子/人人车/懂车帝
package service

import (
	"errors"
	"time"

	"wuchang-tongcheng/internal/modules/car/dto"
	"wuchang-tongcheng/internal/modules/car/model"
	"wuchang-tongcheng/internal/modules/car/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrCarNotFound     = errors.New("车源不存在")
	ErrCarNoPermission = errors.New("无权操作此车源")
	ErrCarAudited      = errors.New("已审核的车源不能重复审核")
	ErrCarStatusInvalid = errors.New("车源状态不允许此操作")
)

// CarService 车源主表业务接口
type CarService interface {
	// C 端
	Create(regionID uint, userID uint, userName string, userPhone string, userAvatar string, req *dto.CreateCarRequest) (*dto.CarInfo, error)
	Update(id uint, operatorID uint, req *dto.UpdateCarRequest) error
	Delete(id uint, operatorID uint) error
	GetByID(id uint, userID uint) (*dto.CarInfo, error)
	List(regionID uint, req *dto.CarListRequest) (*utils.Pagination, []dto.CarInfo, error)
	ListNearby(regionID uint, req *dto.CarNearbyRequest) (*utils.Pagination, []dto.CarInfo, error)
	Search(regionID uint, req *dto.CarSearchRequest) (*utils.Pagination, []dto.CarInfo, error)
	ListMine(userID uint, page, pageSize int) (*utils.Pagination, []dto.CarInfo, error)
	AdvancedSearch(regionID uint, req *dto.AdvancedSearchRequest) (*utils.Pagination, []dto.CarInfo, error)

	// 收藏
	Fav(userID, carID uint) (*dto.FavResponse, error)
	FavStatus(userID, carID uint) (*dto.FavResponse, error)
	ListFavs(userID uint, page, pageSize int) (*utils.Pagination, []dto.CarInfo, error)

	// 互动
	IncrContact(id uint) error
	IncrShare(id uint) error
	RecordView(userID uint, ip string, req *dto.CarViewRequest) error

	// M 端管理
	AdminList(req *dto.CarAdminListRequest) (*utils.Pagination, []dto.CarInfo, error)
	AdminGetByID(id uint) (*dto.CarInfo, error)
	Audit(id uint, auditStatus int, auditReason string) error
	AdminUpdateStatus(id uint, status int) error
	RealCarVerify(id uint, verified bool, reason string) error
	UpdatePromotion(id uint, req *dto.PromotionRequest) error
}

type carService struct {
	repo      repository.CarRepository
	imageRepo repository.ImageRepository
}

// NewCarService 创建车源 service 实例
func NewCarService(repo repository.CarRepository, imageRepo repository.ImageRepository) CarService {
	return &carService{repo: repo, imageRepo: imageRepo}
}

// carStatusText 状态文本
func carStatusText(status int) string {
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
	case model.StatusSold:
		return "已售出"
	}
	return ""
}

// carAuditStatusText 审核状态文本
func carAuditStatusText(s int) string {
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

// toCarInfo model -> dto
func toCarInfo(c *model.Car, images []model.CarImage) *dto.CarInfo {
	info := &dto.CarInfo{
		ID:                    c.ID,
		Title:                 c.Title,
		Content:               c.Content,
		CoverImage:            c.CoverImage,
		UserID:                c.UserID,
		UserName:              c.UserName,
		UserPhone:             c.UserPhone,
		UserAvatar:            c.UserAvatar,
		Status:                c.Status,
		StatusText:            carStatusText(c.Status),
		AuditStatus:           c.AuditStatus,
		AuditStatusText:       carAuditStatusText(c.AuditStatus),
		AuditReason:           c.AuditReason,
		PublishedAt:           c.PublishedAt,
		ListingType:           c.ListingType,
		SourceType:            c.SourceType,
		CarType:               c.CarType,
		BrandID:               c.BrandID,
		BrandName:             c.BrandName,
		ModelID:               c.ModelID,
		ModelName:             c.ModelName,
		Series:                c.Series,
		CategoryID:            c.CategoryID,
		Price:                 c.Price,
		OriginalPrice:         c.OriginalPrice,
		AveragePrice:          c.AveragePrice,
		PriceNegotiable:       c.PriceNegotiable,
		DealerPrice:           c.DealerPrice,
		RegistrationYear:      c.RegistrationYear,
		RegistrationMonth:     c.RegistrationMonth,
		FirstRegistrationDate: c.FirstRegistrationDate,
		Mileage:               c.Mileage,
		MileageUnit:           c.MileageUnit,
		Displacement:          c.Displacement,
		Transmission:          c.Transmission,
		FuelType:              c.FuelType,
		EmissionStandard:      c.EmissionStandard,
		EngineType:            c.EngineType,
		Horsepower:            c.Horsepower,
		ExteriorColor:         c.ExteriorColor,
		InteriorColor:         c.InteriorColor,
		SeatCount:             c.SeatCount,
		DoorCount:             c.DoorCount,
		ConditionLevel:        c.ConditionLevel,
		ConditionScore:        c.ConditionScore,
		AccidentCount:         c.AccidentCount,
		TransferCount:         c.TransferCount,
		LastTransferDate:      c.LastTransferDate,
		AnnualInspectionDue:   c.AnnualInspectionDue,
		AnnualInspectionStatus: c.AnnualInspectionStatus,
		InsuranceDue:          c.InsuranceDue,
		InsuranceStatus:       c.InsuranceStatus,
		CommercialInsuranceDue: c.CommercialInsuranceDue,
		VIN:                   c.VIN,
		LicensePlate:          c.LicensePlate,
		LicenseLocation:       c.LicenseLocation,
		EngineNo:              c.EngineNo,
		UseType:               c.UseType,
		MaintenanceCount:      c.MaintenanceCount,
		LastMaintenanceMileage: c.LastMaintenanceMileage,
		City:                  c.City,
		District:              c.District,
		BusinessDistrict:      c.BusinessDistrict,
		Address:               c.Address,
		Latitude:              c.Latitude,
		Longitude:             c.Longitude,
		Distance:              c.Distance,
		ViewCount:             c.ViewCount,
		FavCount:              c.FavCount,
		ContactCount:          c.ContactCount,
		ShareCount:            c.ShareCount,
		TestDriveCount:        c.TestDriveCount,
		LastTestDriveAt:       c.LastTestDriveAt,
		ContentHash:           c.ContentHash,
		RiskScore:             c.RiskScore,
		SameCarID:             c.SameCarID,
		VideoURL:              c.VideoURL,
		VideoCover:            c.VideoCover,
		Panorama360URL:        c.Panorama360URL,
		Featured:              c.Featured,
		Picked:                c.Picked,
		Verified:              c.Verified,
		PromotionLevel:        c.PromotionLevel,
		TrafficWeight:         c.TrafficWeight,
		RealCarVerified:       c.RealCarVerified,
		RealCarVerifiedAt:     c.RealCarVerifiedAt,
		RegionID:              c.RegionID,
		CreatedAt:             c.CreatedAt,
		UpdatedAt:             c.UpdatedAt,
	}
	// JSONB 字段透传（前端解析）
	if c.Features != nil {
		info.Features = c.Features
	}
	if c.Tags != nil {
		info.Tags = c.Tags
	}
	if c.InspectionItems != nil {
		info.InspectionItems = c.InspectionItems
	}
	if c.AccidentHistory != nil {
		info.AccidentHistory = c.AccidentHistory
	}
	// 图片
	if images != nil {
		imgInfos := make([]dto.CarImageInfo, 0, len(images))
		for _, img := range images {
			imgInfos = append(imgInfos, dto.CarImageInfo{
				ID:          img.ID,
				ImageType:   img.ImageType,
				URL:         img.URL,
				Thumbnail:   img.Thumbnail,
				Title:       img.Title,
				Description: img.Description,
				Sort:        img.Sort,
				IsCover:     img.IsCover,
				Width:       img.Width,
				Height:      img.Height,
				Size:        img.Size,
				Tag:         img.Tag,
			})
		}
		info.Images = imgInfos
	}
	return info
}

// ===== C 端 =====

// Create 发布车源
func (s *carService) Create(regionID uint, userID uint, userName string, userPhone string, userAvatar string, req *dto.CreateCarRequest) (*dto.CarInfo, error) {
	c := &model.Car{
		Title:               req.Title,
		Content:             req.Content,
		CoverImage:          req.CoverImage,
		UserID:              userID,
		UserName:            userName,
		UserPhone:           userPhone,
		UserAvatar:          userAvatar,
		Status:              req.Status,
		ListingType:         req.ListingType,
		SourceType:          req.SourceType,
		CarType:             req.CarType,
		BrandID:             req.BrandID,
		BrandName:           req.BrandName,
		ModelID:             req.ModelID,
		ModelName:           req.ModelName,
		Series:              req.Series,
		CategoryID:          req.CategoryID,
		Price:               req.Price,
		OriginalPrice:       req.OriginalPrice,
		PriceNegotiable:     req.PriceNegotiable,
		DealerPrice:         req.DealerPrice,
		RegistrationYear:    req.RegistrationYear,
		RegistrationMonth:   req.RegistrationMonth,
		FirstRegistrationDate: req.FirstRegistrationDate,
		Mileage:             req.Mileage,
		MileageUnit:         req.MileageUnit,
		Displacement:        req.Displacement,
		Transmission:        req.Transmission,
		FuelType:            req.FuelType,
		EmissionStandard:    req.EmissionStandard,
		EngineType:          req.EngineType,
		Horsepower:          req.Horsepower,
		ExteriorColor:       req.ExteriorColor,
		InteriorColor:       req.InteriorColor,
		SeatCount:           req.SeatCount,
		DoorCount:           req.DoorCount,
		ConditionLevel:      req.ConditionLevel,
		TransferCount:       req.TransferCount,
		LastTransferDate:    req.LastTransferDate,
		AnnualInspectionDue: req.AnnualInspectionDue,
		AnnualInspectionStatus: req.AnnualInspectionStatus,
		InsuranceDue:        req.InsuranceDue,
		InsuranceStatus:     req.InsuranceStatus,
		CommercialInsuranceDue: req.CommercialInsuranceDue,
		VIN:                 req.VIN,
		LicensePlate:        req.LicensePlate,
		LicenseLocation:     req.LicenseLocation,
		EngineNo:            req.EngineNo,
		UseType:             req.UseType,
		MaintenanceCount:    req.MaintenanceCount,
		LastMaintenanceMileage: req.LastMaintenanceMileage,
		City:                req.City,
		District:            req.District,
		BusinessDistrict:    req.BusinessDistrict,
		Address:             req.Address,
		Latitude:            req.Latitude,
		Longitude:           req.Longitude,
		VideoURL:            req.VideoURL,
		VideoCover:          req.VideoCover,
		Panorama360URL:      req.Panorama360URL,
	}
	c.RegionID = regionID

	// 默认值兜底
	if c.ListingType == "" {
		c.ListingType = model.ListingTypeUsed
	}
	if c.SourceType == "" {
		c.SourceType = model.SourceTypePersonal
	}
	if c.CarType == "" {
		c.CarType = model.CarTypeSedan
	}
	if c.MileageUnit == "" {
		c.MileageUnit = model.MileageUnitKM
	}
	if c.ConditionLevel == "" {
		c.ConditionLevel = model.ConditionLevelA
	}
	if c.UseType == "" {
		c.UseType = model.UseTypeNonOperational
	}
	// MVP 阶段简化为发布即通过
	c.AuditStatus = model.AuditApproved

	// JSONB 字段
	if req.Features != nil {
		if jb, err := model.FromJSON(req.Features); err == nil {
			c.Features = jb
		}
	}
	if req.Tags != nil {
		if jb, err := model.FromJSON(req.Tags); err == nil {
			c.Tags = jb
		}
	}
	if req.InspectionItems != nil {
		if jb, err := model.FromJSON(req.InspectionItems); err == nil {
			c.InspectionItems = jb
		}
	}
	if req.AccidentHistory != nil {
		if jb, err := model.FromJSON(req.AccidentHistory); err == nil {
			c.AccidentHistory = jb
		}
	}

	// 发布时间
	if req.Status == model.StatusPublished {
		now := time.Now()
		c.PublishedAt = &now
	}

	if err := s.repo.Create(c); err != nil {
		return nil, err
	}

	// 保存图片子表
	if len(req.Images) > 0 {
		imgs := make([]model.CarImage, 0, len(req.Images))
		for _, in := range req.Images {
			imgs = append(imgs, model.CarImage{
				ImageType:   in.ImageType,
				URL:         in.URL,
				Thumbnail:   in.Thumbnail,
				Title:       in.Title,
				Description: in.Description,
				Sort:        in.Sort,
				IsCover:     in.IsCover,
				Width:       in.Width,
				Height:      in.Height,
				Size:        in.Size,
				Tag:         in.Tag,
			})
		}
		_ = s.repo.ReplaceImages(c.ID, regionID, imgs)
	}

	return toCarInfo(c, nil), nil
}

// Update 更新车源（仅发布者本人）
func (s *carService) Update(id uint, operatorID uint, req *dto.UpdateCarRequest) error {
	c, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCarNotFound
		}
		return err
	}
	if c.UserID != operatorID {
		return ErrCarNoPermission
	}

	fields := map[string]interface{}{}
	if req.Title != nil {
		fields["title"] = *req.Title
	}
	if req.Content != nil {
		fields["content"] = *req.Content
	}
	if req.CoverImage != nil {
		fields["cover_image"] = *req.CoverImage
	}
	if req.ListingType != nil {
		fields["listing_type"] = *req.ListingType
	}
	if req.SourceType != nil {
		fields["source_type"] = *req.SourceType
	}
	if req.CarType != nil {
		fields["car_type"] = *req.CarType
	}
	if req.BrandID != nil {
		fields["brand_id"] = *req.BrandID
	}
	if req.BrandName != nil {
		fields["brand_name"] = *req.BrandName
	}
	if req.ModelID != nil {
		fields["model_id"] = *req.ModelID
	}
	if req.ModelName != nil {
		fields["model_name"] = *req.ModelName
	}
	if req.Series != nil {
		fields["series"] = *req.Series
	}
	if req.CategoryID != nil {
		fields["category_id"] = *req.CategoryID
	}
	if req.Price != nil {
		fields["price"] = *req.Price
	}
	if req.OriginalPrice != nil {
		fields["original_price"] = *req.OriginalPrice
	}
	if req.PriceNegotiable != nil {
		fields["price_negotiable"] = *req.PriceNegotiable
	}
	if req.DealerPrice != nil {
		fields["dealer_price"] = *req.DealerPrice
	}
	if req.RegistrationYear != nil {
		fields["registration_year"] = *req.RegistrationYear
	}
	if req.RegistrationMonth != nil {
		fields["registration_month"] = *req.RegistrationMonth
	}
	if req.FirstRegistrationDate != nil {
		fields["first_registration_date"] = *req.FirstRegistrationDate
	}
	if req.Mileage != nil {
		fields["mileage"] = *req.Mileage
	}
	if req.MileageUnit != nil {
		fields["mileage_unit"] = *req.MileageUnit
	}
	if req.Displacement != nil {
		fields["displacement"] = *req.Displacement
	}
	if req.Transmission != nil {
		fields["transmission"] = *req.Transmission
	}
	if req.FuelType != nil {
		fields["fuel_type"] = *req.FuelType
	}
	if req.EmissionStandard != nil {
		fields["emission_standard"] = *req.EmissionStandard
	}
	if req.EngineType != nil {
		fields["engine_type"] = *req.EngineType
	}
	if req.Horsepower != nil {
		fields["horsepower"] = *req.Horsepower
	}
	if req.ExteriorColor != nil {
		fields["exterior_color"] = *req.ExteriorColor
	}
	if req.InteriorColor != nil {
		fields["interior_color"] = *req.InteriorColor
	}
	if req.SeatCount != nil {
		fields["seat_count"] = *req.SeatCount
	}
	if req.DoorCount != nil {
		fields["door_count"] = *req.DoorCount
	}
	if req.ConditionLevel != nil {
		fields["condition_level"] = *req.ConditionLevel
	}
	if req.ConditionScore != nil {
		fields["condition_score"] = *req.ConditionScore
	}
	if req.AccidentCount != nil {
		fields["accident_count"] = *req.AccidentCount
	}
	if req.TransferCount != nil {
		fields["transfer_count"] = *req.TransferCount
	}
	if req.LastTransferDate != nil {
		fields["last_transfer_date"] = *req.LastTransferDate
	}
	if req.AnnualInspectionDue != nil {
		fields["annual_inspection_due"] = *req.AnnualInspectionDue
	}
	if req.AnnualInspectionStatus != nil {
		fields["annual_inspection_status"] = *req.AnnualInspectionStatus
	}
	if req.InsuranceDue != nil {
		fields["insurance_due"] = *req.InsuranceDue
	}
	if req.InsuranceStatus != nil {
		fields["insurance_status"] = *req.InsuranceStatus
	}
	if req.CommercialInsuranceDue != nil {
		fields["commercial_insurance_due"] = *req.CommercialInsuranceDue
	}
	if req.VIN != nil {
		fields["vin"] = *req.VIN
	}
	if req.LicensePlate != nil {
		fields["license_plate"] = *req.LicensePlate
	}
	if req.LicenseLocation != nil {
		fields["license_location"] = *req.LicenseLocation
	}
	if req.EngineNo != nil {
		fields["engine_no"] = *req.EngineNo
	}
	if req.UseType != nil {
		fields["use_type"] = *req.UseType
	}
	if req.MaintenanceCount != nil {
		fields["maintenance_count"] = *req.MaintenanceCount
	}
	if req.LastMaintenanceMileage != nil {
		fields["last_maintenance_mileage"] = *req.LastMaintenanceMileage
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
	if req.Panorama360URL != nil {
		fields["panorama_360_url"] = *req.Panorama360URL
	}
	if req.Features != nil {
		if jb, err := model.FromJSON(req.Features); err == nil {
			fields["features"] = jb
		}
	}
	if req.Tags != nil {
		if jb, err := model.FromJSON(req.Tags); err == nil {
			fields["tags"] = jb
		}
	}
	if req.InspectionItems != nil {
		if jb, err := model.FromJSON(req.InspectionItems); err == nil {
			fields["inspection_items"] = jb
		}
	}
	if req.AccidentHistory != nil {
		if jb, err := model.FromJSON(req.AccidentHistory); err == nil {
			fields["accident_history"] = jb
		}
	}

	// 状态变更
	if req.Status != nil {
		if *req.Status == model.StatusPublished && c.Status != model.StatusPublished {
			now := time.Now()
			fields["status"] = model.StatusPublished
			fields["published_at"] = &now
			fields["audit_status"] = model.AuditApproved
		} else {
			fields["status"] = *req.Status
		}
	}

	if len(fields) == 0 && req.Images == nil {
		return nil
	}

	if err := s.repo.UpdateFields(id, fields); err != nil {
		return err
	}

	// 更新图片子表
	if req.Images != nil {
		imgs := make([]model.CarImage, 0, len(*req.Images))
		for _, in := range *req.Images {
			imgs = append(imgs, model.CarImage{
				ImageType:   in.ImageType,
				URL:         in.URL,
				Thumbnail:   in.Thumbnail,
				Title:       in.Title,
				Description: in.Description,
				Sort:        in.Sort,
				IsCover:     in.IsCover,
				Width:       in.Width,
				Height:      in.Height,
				Size:        in.Size,
				Tag:         in.Tag,
			})
		}
		if err := s.repo.ReplaceImages(id, c.RegionID, imgs); err != nil {
			return err
		}
	}
	return nil
}

// Delete 删除车源（仅发布者本人）
func (s *carService) Delete(id uint, operatorID uint) error {
	c, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCarNotFound
		}
		return err
	}
	if c.UserID != operatorID {
		return ErrCarNoPermission
	}
	return s.repo.Delete(id)
}

// GetByID 获取详情（同时增加浏览量）
func (s *carService) GetByID(id uint, userID uint) (*dto.CarInfo, error) {
	c, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCarNotFound
		}
		return nil, err
	}

	// 增加浏览量
	_ = s.repo.IncrViewCount(id)
	c.ViewCount++

	// 拼装图片
	images, _ := s.repo.ListImages(id)
	if images == nil {
		images = []model.CarImage{}
	}

	info := toCarInfo(c, images)

	// 当前用户是否已收藏
	if userID > 0 {
		if hasFaved, err := s.repo.FavExists(userID, id); err == nil {
			info.HasFaved = hasFaved
		}
	}
	return info, nil
}

// List 列表查询（C 端）
func (s *carService) List(regionID uint, req *dto.CarListRequest) (*utils.Pagination, []dto.CarInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.CarListOptions{
		CategoryID:       req.CategoryID,
		BrandID:          req.BrandID,
		ModelID:          req.ModelID,
		Keyword:          req.Keyword,
		CarType:          req.CarType,
		ListingType:      req.ListingType,
		SourceType:       req.SourceType,
		FuelType:         req.FuelType,
		Transmission:     req.Transmission,
		MinPrice:         req.MinPrice,
		MaxPrice:         req.MaxPrice,
		MinMileage:       req.MinMileage,
		MaxMileage:       req.MaxMileage,
		MinYear:          req.MinYear,
		MaxYear:          req.MaxYear,
		ConditionLevel:   req.ConditionLevel,
		City:             req.City,
		Featured:         req.Featured,
		Picked:           req.Picked,
		Verified:         req.Verified,
		RealCarVerified:  req.RealCarVerified,
		PriceNegotiable:  req.PriceNegotiable,
		Sort:             req.Sort,
		Status:           model.StatusPublished,
	}

	list, total, err := s.repo.List(regionID, pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total

	result := make([]dto.CarInfo, 0, len(list))
	for i := range list {
		result = append(result, *toCarInfo(&list[i], nil))
	}
	return pagination, result, nil
}

// ListNearby 附近查询
func (s *carService) ListNearby(regionID uint, req *dto.CarNearbyRequest) (*utils.Pagination, []dto.CarInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	radiusKm := req.RadiusKm
	if radiusKm <= 0 {
		radiusKm = 5
	}
	opts := repository.CarListOptions{
		CategoryID: req.CategoryID,
		BrandID:    req.BrandID,
		CarType:    req.CarType,
		MinPrice:   req.MinPrice,
		MaxPrice:   req.MaxPrice,
		Sort:       req.Sort,
		Status:     model.StatusPublished,
	}
	list, total, err := s.repo.ListNearby(regionID, pagination, req.Latitude, req.Longitude, radiusKm, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.CarInfo, 0, len(list))
	for i := range list {
		result = append(result, *toCarInfo(&list[i], nil))
	}
	return pagination, result, nil
}

// Search 搜索
func (s *carService) Search(regionID uint, req *dto.CarSearchRequest) (*utils.Pagination, []dto.CarInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	list, total, err := s.repo.Search(regionID, pagination, req.Keyword)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.CarInfo, 0, len(list))
	for i := range list {
		result = append(result, *toCarInfo(&list[i], nil))
	}
	return pagination, result, nil
}

// ListMine 我的发布
func (s *carService) ListMine(userID uint, page, pageSize int) (*utils.Pagination, []dto.CarInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByUser(userID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.CarInfo, 0, len(list))
	for i := range list {
		result = append(result, *toCarInfo(&list[i], nil))
	}
	return pagination, result, nil
}

// AdvancedSearch 高级搜索（C 端）
func (s *carService) AdvancedSearch(regionID uint, req *dto.AdvancedSearchRequest) (*utils.Pagination, []dto.CarInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.CarListOptions{
		CategoryID:      req.CategoryID,
		BrandID:         req.BrandID,
		ModelID:         req.ModelID,
		Keyword:         req.Keyword,
		CarType:         req.CarType,
		ListingType:     req.ListingType,
		SourceType:      req.SourceType,
		FuelType:        req.FuelType,
		Transmission:    req.Transmission,
		MinPrice:        req.MinPrice,
		MaxPrice:        req.MaxPrice,
		MinMileage:      req.MinMileage,
		MaxMileage:      req.MaxMileage,
		MinYear:         req.MinYear,
		MaxYear:         req.MaxYear,
		ConditionLevel:  req.ConditionLevel,
		City:            req.City,
		Featured:        req.Featured,
		Picked:          req.Picked,
		Verified:        req.Verified,
		RealCarVerified: req.RealCarVerified,
		PriceNegotiable: req.PriceNegotiable,
		Sort:            req.Sort,
		Status:          model.StatusPublished,
	}

	// 如果同时提供经纬度+半径，则走附近查询
	if req.Latitude != 0 && req.Longitude != 0 && req.RadiusKm > 0 {
		list, total, err := s.repo.ListNearby(regionID, pagination, req.Latitude, req.Longitude, req.RadiusKm, opts)
		if err != nil {
			return nil, nil, err
		}
		pagination.Total = total
		result := make([]dto.CarInfo, 0, len(list))
		for i := range list {
			result = append(result, *toCarInfo(&list[i], nil))
		}
		return pagination, result, nil
	}

	list, total, err := s.repo.List(regionID, pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.CarInfo, 0, len(list))
	for i := range list {
		result = append(result, *toCarInfo(&list[i], nil))
	}
	return pagination, result, nil
}

// ===== 收藏 =====

func (s *carService) Fav(userID, carID uint) (*dto.FavResponse, error) {
	if userID == 0 {
		return nil, ErrCarNoPermission
	}
	c, err := s.repo.FindByID(carID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCarNotFound
		}
		return nil, err
	}

	exists, err := s.repo.FavExists(userID, carID)
	if err != nil {
		return nil, err
	}
	if exists {
		if err := s.repo.DeleteFav(userID, carID); err != nil {
			return nil, err
		}
		_ = s.repo.DecrFavCount(carID)
		return &dto.FavResponse{HasFaved: false, FavCount: c.FavCount - 1}, nil
	}

	fav := &model.CarFavorite{
		UserID: userID,
		CarID:  carID,
	}
	if err := s.repo.CreateFav(fav); err != nil {
		return nil, err
	}
	_ = s.repo.IncrFavCount(carID)
	return &dto.FavResponse{HasFaved: true, FavCount: c.FavCount + 1}, nil
}

func (s *carService) FavStatus(userID, carID uint) (*dto.FavResponse, error) {
	c, err := s.repo.FindByID(carID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCarNotFound
		}
		return nil, err
	}
	if userID == 0 {
		return &dto.FavResponse{HasFaved: false, FavCount: c.FavCount}, nil
	}
	exists, err := s.repo.FavExists(userID, carID)
	if err != nil {
		return nil, err
	}
	return &dto.FavResponse{HasFaved: exists, FavCount: c.FavCount}, nil
}

func (s *carService) ListFavs(userID uint, page, pageSize int) (*utils.Pagination, []dto.CarInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	favs, total, err := s.repo.ListFavs(userID, page, pageSize)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total

	ids := make([]uint, 0, len(favs))
	for _, f := range favs {
		ids = append(ids, f.CarID)
	}
	if len(ids) == 0 {
		return pagination, []dto.CarInfo{}, nil
	}

	list := make([]dto.CarInfo, 0, len(ids))
	for _, id := range ids {
		c, err := s.repo.FindByID(id)
		if err != nil {
			continue
		}
		list = append(list, *toCarInfo(c, nil))
	}
	return pagination, list, nil
}

// ===== 互动 =====

func (s *carService) IncrContact(id uint) error {
	return s.repo.IncrContactCount(id)
}

func (s *carService) IncrShare(id uint) error {
	return s.repo.IncrShareCount(id)
}

// RecordView 记录浏览
func (s *carService) RecordView(userID uint, ip string, req *dto.CarViewRequest) error {
	v := &model.CarView{
		CarID:     req.CarID,
		ListingID: req.ListingID,
		UserID:    userID,
		IP:        ip,
		Device:    req.Device,
		Source:    req.Source,
		Duration:  req.Duration,
		Longitude: req.Longitude,
		Latitude:  req.Latitude,
	}
	if v.Device == "" {
		v.Device = "pc"
	}
	if v.Source == "" {
		v.Source = "direct"
	}
	_ = s.repo.IncrViewCount(req.CarID)
	return s.repo.CreateView(v)
}

// ===== M 端管理 =====

func (s *carService) AdminList(req *dto.CarAdminListRequest) (*utils.Pagination, []dto.CarInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.CarAdminListOptions{
		RegionID:    req.RegionID,
		UserID:      req.UserID,
		CategoryID:  req.CategoryID,
		BrandID:     req.BrandID,
		Status:      req.Status,
		AuditStatus: req.AuditStatus,
		ListingType: req.ListingType,
		SourceType:  req.SourceType,
		CarType:     req.CarType,
		Keyword:     req.Keyword,
	}
	list, total, err := s.repo.AdminList(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.CarInfo, 0, len(list))
	for i := range list {
		result = append(result, *toCarInfo(&list[i], nil))
	}
	return pagination, result, nil
}

func (s *carService) AdminGetByID(id uint) (*dto.CarInfo, error) {
	c, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCarNotFound
		}
		return nil, err
	}
	images, _ := s.repo.ListImages(id)
	if images == nil {
		images = []model.CarImage{}
	}
	return toCarInfo(c, images), nil
}

// Audit 审核
func (s *carService) Audit(id uint, auditStatus int, auditReason string) error {
	c, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCarNotFound
		}
		return err
	}

	fields := map[string]interface{}{
		"audit_status": auditStatus,
		"audit_reason": auditReason,
	}

	// 审核通过且物品状态为草稿：自动发布
	if auditStatus == model.AuditApproved && c.Status == model.StatusDraft {
		now := time.Now()
		fields["status"] = model.StatusPublished
		fields["published_at"] = &now
	}
	// 审核拒绝：强制下架
	if auditStatus == model.AuditRejected && c.Status == model.StatusPublished {
		fields["status"] = model.StatusOffline
	}

	return s.repo.UpdateFields(id, fields)
}

// AdminUpdateStatus 管理后台强制下架/恢复
func (s *carService) AdminUpdateStatus(id uint, status int) error {
	fields := map[string]interface{}{
		"status": status,
	}
	if status == model.StatusPublished {
		now := time.Now()
		fields["published_at"] = &now
		fields["audit_status"] = model.AuditApproved
	}
	return s.repo.UpdateFields(id, fields)
}

// RealCarVerify 真车认证（M 端）
func (s *carService) RealCarVerify(id uint, verified bool, reason string) error {
	fields := map[string]interface{}{
		"real_car_verified": verified,
	}
	if verified {
		now := time.Now()
		fields["real_car_verified_at"] = &now
	} else {
		fields["real_car_verified_at"] = nil
	}
	_ = reason
	return s.repo.UpdateFields(id, fields)
}

// UpdatePromotion 更新推广配置（M 端）
func (s *carService) UpdatePromotion(id uint, req *dto.PromotionRequest) error {
	fields := map[string]interface{}{}
	if req.Featured != nil {
		fields["featured"] = *req.Featured
	}
	if req.Picked != nil {
		fields["picked"] = *req.Picked
	}
	if req.Verified != nil {
		fields["verified"] = *req.Verified
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
