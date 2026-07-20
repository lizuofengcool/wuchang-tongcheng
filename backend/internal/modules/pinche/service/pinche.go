// Package service 同城拼车出行业务逻辑层 - 拼车主表
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
// 依据 v3.2.1 架构方案：对标哈啰出行/嘀嗒出行/滴滴顺风车
package service

import (
	"errors"
	"math"
	"time"

	"wuchang-tongcheng/internal/modules/pinche/dto"
	"wuchang-tongcheng/internal/modules/pinche/model"
	"wuchang-tongcheng/internal/modules/pinche/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrPincheNotFound     = errors.New("拼车行程不存在")
	ErrPincheNoPermission = errors.New("无权操作此拼车行程")
	ErrPincheAudited      = errors.New("已审核的拼车行程不能重复审核")
	ErrPincheStatusInvalid = errors.New("拼车行程状态不允许此操作")
	ErrPincheNoSeats      = errors.New("剩余座位不足")
)

// PincheService 拼车主表业务接口
type PincheService interface {
	// C 端
	Create(regionID uint, userID uint, userName, userPhone, userAvatar string, req *dto.CreatePincheRequest) (*dto.PincheInfo, error)
	Update(id uint, operatorID uint, req *dto.UpdatePincheRequest) error
	Delete(id uint, operatorID uint) error
	GetByID(id uint, userID uint) (*dto.PincheInfo, error)
	List(regionID uint, req *dto.PincheListRequest) (*utils.Pagination, []dto.PincheInfo, error)
	ListNearby(regionID uint, req *dto.PincheNearbyRequest) (*utils.Pagination, []dto.PincheInfo, error)
	Search(regionID uint, req *dto.PincheSearchRequest) (*utils.Pagination, []dto.PincheInfo, error)
	ListMine(userID uint, page, pageSize int) (*utils.Pagination, []dto.PincheInfo, error)
	Match(regionID uint, req *dto.PincheMatchRequest) (*dto.PincheMatchResponse, error)

	// 互动
	IncrContact(id uint) error
	IncrShare(id uint) error
	RecordView(userID uint, req *dto.PincheViewRequest) error

	// M 端
	AdminList(req *dto.PincheAdminListRequest) (*utils.Pagination, []dto.PincheInfo, error)
	AdminGetByID(id uint) (*dto.PincheInfo, error)
	Audit(id uint, auditStatus int, auditReason string) error
	AdminUpdateStatus(id uint, status int) error
	BatchAudit(req *dto.BatchAuditRequest) (*dto.BatchResultResponse, error)
	BatchUpdateStatus(req *dto.BatchStatusUpdateRequest) (*dto.BatchResultResponse, error)
	BatchDelete(req *dto.BatchDeleteRequest) (*dto.BatchResultResponse, error)
}

type pincheService struct {
	repo repository.PincheRepository
}

// NewPincheService 创建拼车 service 实例
func NewPincheService(repo repository.PincheRepository) PincheService {
	return &pincheService{repo: repo}
}

// pincheStatusText 状态文本
func pincheStatusText(status int) string {
	switch status {
	case model.PincheStatusDraft:
		return "草稿"
	case model.PincheStatusPublished:
		return "已发布"
	case model.PincheStatusFinished:
		return "已结束"
	case model.PincheStatusCancelled:
		return "已取消"
	case model.PincheStatusOngoing:
		return "进行中"
	}
	return ""
}

// pincheAuditStatusText 审核状态文本
func pincheAuditStatusText(s int) string {
	switch s {
	case model.PincheAuditPending:
		return "待审"
	case model.PincheAuditApproved:
		return "通过"
	case model.PincheAuditRejected:
		return "拒绝"
	}
	return ""
}

// toPincheInfo model -> dto
func toPincheInfo(p *model.Pinche) *dto.PincheInfo {
	info := &dto.PincheInfo{
		ID:              p.ID,
		RegionID:        p.RegionID,
		UserID:          p.UserID,
		UserName:        p.UserName,
		UserPhone:       p.UserPhone,
		UserAvatar:      p.UserAvatar,
		TripType:        p.TripType,
		Role:            p.Role,
		Title:           p.Title,
		Content:         p.Content,
		CoverImage:      p.CoverImage,
		Status:          p.Status,
		StatusText:      pincheStatusText(p.Status),
		AuditStatus:     p.AuditStatus,
		AuditStatusText: pincheAuditStatusText(p.AuditStatus),
		AuditReason:     p.AuditReason,
		PublishedAt:     p.PublishedAt,
		DepartureTime:   p.DepartureTime,
		PickupLocation:  p.PickupLocation,
		PickupLat:       p.PickupLat,
		PickupLng:       p.PickupLng,
		DropoffLocation: p.DropoffLocation,
		DropoffLat:      p.DropoffLat,
		DropoffLng:      p.DropoffLng,
		DistanceKm:      p.DistanceKm,
		DurationMin:     p.DurationMin,
		TotalSeats:      p.TotalSeats,
		AvailableSeats:  p.AvailableSeats,
		BookedSeats:     p.BookedSeats,
		PricePerSeat:    p.PricePerSeat,
		TotalAmount:     p.TotalAmount,
		TollFee:         p.TollFee,
		VehicleID:       p.VehicleID,
		DriverID:        p.DriverID,
		RouteID:         p.RouteID,
		InsuranceID:     p.InsuranceID,
		TripID:          p.TripID,
		EmergencyContactID: p.EmergencyContactID,
		ShareToken:      p.ShareToken,
		PaymentMethod:   p.PaymentMethod,
		ViewCount:       p.ViewCount,
		FavCount:        p.FavCount,
		ContactCount:    p.ContactCount,
		ShareCount:      p.ShareCount,
		Featured:        p.Featured,
		Picked:          p.Picked,
		Verified:        p.Verified,
		PromotionLevel:  p.PromotionLevel,
		ContentHash:     p.ContentHash,
		RiskScore:       p.RiskScore,
		StartedAt:       p.StartedAt,
		CompletedAt:     p.CompletedAt,
		CancelledAt:     p.CancelledAt,
		Distance:        p.Distance,
		CreatedAt:       p.CreatedAt,
	}
	if p.Features != nil {
		info.Features = p.Features
	}
	if p.Tags != nil {
		info.Tags = p.Tags
	}
	return info
}

// ===== C 端 =====

// Create 发布拼车行程
func (s *pincheService) Create(regionID uint, userID uint, userName, userPhone, userAvatar string, req *dto.CreatePincheRequest) (*dto.PincheInfo, error) {
	p := &model.Pinche{
		UserID:          userID,
		UserName:        userName,
		UserPhone:       userPhone,
		UserAvatar:      userAvatar,
		TripType:        req.TripType,
		Role:            req.Role,
		Title:           req.Title,
		Content:         req.Content,
		CoverImage:      req.CoverImage,
		DepartureTime:   req.DepartureTime,
		PickupLocation:  req.PickupLocation,
		PickupLat:       req.PickupLat,
		PickupLng:       req.PickupLng,
		DropoffLocation: req.DropoffLocation,
		DropoffLat:      req.DropoffLat,
		DropoffLng:      req.DropoffLng,
		DistanceKm:      req.DistanceKm,
		DurationMin:     req.DurationMin,
		TotalSeats:      req.TotalSeats,
		AvailableSeats:  req.TotalSeats,
		PricePerSeat:    req.PricePerSeat,
		TollFee:         req.TollFee,
		VehicleID:       req.VehicleID,
		RouteID:         req.RouteID,
		PaymentMethod:   req.PaymentMethod,
	}
	p.RegionID = regionID

	// 默认值兜底
	if p.TripType == "" {
		p.TripType = model.TripTypeShunfeng
	}
	if p.Role == "" {
		p.Role = model.RoleDriver
	}
	if p.TotalSeats == 0 {
		p.TotalSeats = 4
		p.AvailableSeats = 4
	}
	if p.PaymentMethod == "" {
		p.PaymentMethod = model.PaymentMethodCash
	}
	// MVP 阶段简化为发布即通过
	p.AuditStatus = model.PincheAuditApproved
	p.Status = model.PincheStatusPublished
	now := time.Now()
	p.PublishedAt = &now

	// JSONB 字段
	if req.Features != nil {
		if jb, err := model.FromJSON(req.Features); err == nil {
			p.Features = jb
		}
	}
	if req.Tags != nil {
		if jb, err := model.FromJSON(req.Tags); err == nil {
			p.Tags = jb
		}
	}

	if err := s.repo.Create(p); err != nil {
		return nil, err
	}
	return toPincheInfo(p), nil
}

// Update 更新拼车行程（仅发布者本人）
func (s *pincheService) Update(id uint, operatorID uint, req *dto.UpdatePincheRequest) error {
	p, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPincheNotFound
		}
		return err
	}
	if p.UserID != operatorID {
		return ErrPincheNoPermission
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
	if req.DepartureTime != nil {
		fields["departure_time"] = *req.DepartureTime
	}
	if req.PickupLocation != nil {
		fields["pickup_location"] = *req.PickupLocation
	}
	if req.PickupLat != nil {
		fields["pickup_lat"] = *req.PickupLat
	}
	if req.PickupLng != nil {
		fields["pickup_lng"] = *req.PickupLng
	}
	if req.DropoffLocation != nil {
		fields["dropoff_location"] = *req.DropoffLocation
	}
	if req.DropoffLat != nil {
		fields["dropoff_lat"] = *req.DropoffLat
	}
	if req.DropoffLng != nil {
		fields["dropoff_lng"] = *req.DropoffLng
	}
	if req.DistanceKm != nil {
		fields["distance_km"] = *req.DistanceKm
	}
	if req.DurationMin != nil {
		fields["duration_min"] = *req.DurationMin
	}
	if req.TotalSeats != nil {
		fields["total_seats"] = *req.TotalSeats
	}
	if req.PricePerSeat != nil {
		fields["price_per_seat"] = *req.PricePerSeat
	}
	if req.TollFee != nil {
		fields["toll_fee"] = *req.TollFee
	}
	if req.VehicleID != nil {
		fields["vehicle_id"] = *req.VehicleID
	}
	if req.RouteID != nil {
		fields["route_id"] = *req.RouteID
	}
	if req.PaymentMethod != nil {
		fields["payment_method"] = *req.PaymentMethod
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

	if len(fields) == 0 {
		return nil
	}
	return s.repo.Update(id, fields)
}

// Delete 删除拼车行程（仅发布者本人）
func (s *pincheService) Delete(id uint, operatorID uint) error {
	p, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPincheNotFound
		}
		return err
	}
	if p.UserID != operatorID {
		return ErrPincheNoPermission
	}
	return s.repo.Delete(id)
}

// GetByID 获取详情（同时增加浏览量）
func (s *pincheService) GetByID(id uint, userID uint) (*dto.PincheInfo, error) {
	p, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPincheNotFound
		}
		return nil, err
	}
	_ = s.repo.IncrViewCount(id)
	p.ViewCount++
	return toPincheInfo(p), nil
}

// List 列表查询
func (s *pincheService) List(regionID uint, req *dto.PincheListRequest) (*utils.Pagination, []dto.PincheInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.PincheListOptions{
		TripType:      req.TripType,
		Role:          req.Role,
		MinPrice:      req.MinPrice,
		MaxPrice:      req.MaxPrice,
		MinSeats:      req.MinSeats,
		DepartureFrom: req.DepartureFrom,
		DepartureTo:   req.DepartureTo,
		PickupCity:    req.PickupCity,
		DropoffCity:   req.DropoffCity,
		Keyword:       req.Keyword,
		Sort:          req.Sort,
		Status:        intPtr(model.PincheStatusPublished),
	}
	list, total, err := s.repo.List(regionID, pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.PincheInfo, 0, len(list))
	for i := range list {
		result = append(result, *toPincheInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListNearby 附近查询
func (s *pincheService) ListNearby(regionID uint, req *dto.PincheNearbyRequest) (*utils.Pagination, []dto.PincheInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	radiusKm := req.RadiusKm
	if radiusKm <= 0 {
		radiusKm = 5
	}
	opts := repository.PincheNearbyOptions{
		Latitude: req.Latitude,
		Longitude: req.Longitude,
		RadiusKm: radiusKm,
		TripType: req.TripType,
		Role:     req.Role,
	}
	list, total, err := s.repo.ListNearby(regionID, pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.PincheInfo, 0, len(list))
	for i := range list {
		result = append(result, *toPincheInfo(&list[i]))
	}
	return pagination, result, nil
}

// Search 搜索
func (s *pincheService) Search(regionID uint, req *dto.PincheSearchRequest) (*utils.Pagination, []dto.PincheInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.PincheListOptions{
		TripType:   req.TripType,
		MinPrice:   req.MinPrice,
		MaxPrice:   req.MaxPrice,
		MinSeats:   req.MinSeats,
		Keyword:    req.Keyword,
		PickupCity: req.PickupLocation,
		DropoffCity: req.DropoffLocation,
		Sort:       "latest",
		Status:     intPtr(model.PincheStatusPublished),
	}
	if req.DepartureDate != "" {
		opts.DepartureFrom = req.DepartureDate + " 00:00:00"
		opts.DepartureTo = req.DepartureDate + " 23:59:59"
	}
	list, total, err := s.repo.List(regionID, pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.PincheInfo, 0, len(list))
	for i := range list {
		result = append(result, *toPincheInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListMine 我的发布
func (s *pincheService) ListMine(userID uint, page, pageSize int) (*utils.Pagination, []dto.PincheInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListMine(userID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.PincheInfo, 0, len(list))
	for i := range list {
		result = append(result, *toPincheInfo(&list[i]))
	}
	return pagination, result, nil
}

// Match 智能匹配
// 综合考虑：路线相似度（起点/终点距离）+ 出发时间相近度 + 价格 + 座位
func (s *pincheService) Match(regionID uint, req *dto.PincheMatchRequest) (*dto.PincheMatchResponse, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.PincheListOptions{
		Role:    model.RoleDriver,
		MinSeats: req.Seats,
		MaxPrice: req.MaxPrice,
		Status:   intPtr(model.PincheStatusPublished),
	}
	list, total, err := s.repo.ListMatch(regionID, pagination, opts)
	if err != nil {
		return nil, err
	}

	maxRadius := req.MaxRadiusKm
	if maxRadius <= 0 {
		maxRadius = 10
	}

	items := make([]dto.PincheMatchItem, 0, len(list))
	for i := range list {
		p := &list[i]
		// 计算匹配度
		score, reasons := s.calcMatchScore(p, req, maxRadius)
		if score <= 0 {
			continue
		}
		item := dto.PincheMatchItem{
			PincheInfo:  *toPincheInfo(p),
			MatchScore:  score,
			MatchReasons: reasons,
		}
		items = append(items, item)
	}

	// 按匹配度排序（从高到低）
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if items[j].MatchScore > items[i].MatchScore {
				items[i], items[j] = items[j], items[i]
			}
		}
	}

	return &dto.PincheMatchResponse{
		Total: int(total),
		List:  items,
	}, nil
}

// calcMatchScore 计算匹配度（0-100）
func (s *pincheService) calcMatchScore(p *model.Pinche, req *dto.PincheMatchRequest, maxRadiusKm float64) (float64, []string) {
	var score float64
	reasons := make([]string, 0, 4)

	// 起点距离评分（30分）
	pickupDist := haversineDistance(req.PickupLat, req.PickupLng, p.PickupLat, p.PickupLng)
	if pickupDist <= maxRadiusKm {
		score += 30 * (1 - pickupDist/maxRadiusKm)
		reasons = append(reasons, "起点相近")
	} else {
		return 0, nil
	}

	// 终点距离评分（30分）
	dropoffDist := haversineDistance(req.DropoffLat, req.DropoffLng, p.DropoffLat, p.DropoffLng)
	if dropoffDist <= maxRadiusKm*2 {
		score += 30 * (1 - dropoffDist/(maxRadiusKm*2))
		reasons = append(reasons, "终点相近")
	} else {
		return 0, nil
	}

	// 出发时间评分（20分）
	if req.DepartureTime != nil && p.DepartureTime != nil {
		diff := p.DepartureTime.Sub(*req.DepartureTime)
		hoursDiff := math.Abs(diff.Hours())
		if hoursDiff <= 24 {
			score += 20 * (1 - hoursDiff/24)
			reasons = append(reasons, "出发时间相近")
		}
	} else {
		score += 10
	}

	// 座位评分（10分）
	if p.AvailableSeats >= req.Seats {
		score += 10
		reasons = append(reasons, "座位充足")
	}

	// 价格评分（10分）
	if req.MaxPrice > 0 && p.PricePerSeat <= req.MaxPrice {
		score += 10
		reasons = append(reasons, "价格合适")
	} else if req.MaxPrice == 0 {
		score += 5
	}

	if score < 30 {
		return 0, nil
	}
	return score, reasons
}

// haversineDistance 计算两个经纬度之间的距离（公里）
func haversineDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371 // 地球半径（公里）
	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

// intPtr 辅助函数：返回 int 指针
func intPtr(v int) *int {
	return &v
}

// ===== 互动 =====

func (s *pincheService) IncrContact(id uint) error {
	return s.repo.IncrContactCount(id)
}

func (s *pincheService) IncrShare(id uint) error {
	return s.repo.IncrShareCount(id)
}

// RecordView 记录浏览
func (s *pincheService) RecordView(userID uint, req *dto.PincheViewRequest) error {
	_ = userID
	_ = s.repo.IncrViewCount(req.PincheID)
	return nil
}

// ===== M 端 =====

func (s *pincheService) AdminList(req *dto.PincheAdminListRequest) (*utils.Pagination, []dto.PincheInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.PincheAdminListOptions{
		RegionID:    req.RegionID,
		UserID:      req.UserID,
		TripType:    req.TripType,
		Role:        req.Role,
		Status:      req.Status,
		AuditStatus: req.AuditStatus,
		Keyword:     req.Keyword,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
	}
	list, total, err := s.repo.AdminList(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.PincheInfo, 0, len(list))
	for i := range list {
		result = append(result, *toPincheInfo(&list[i]))
	}
	return pagination, result, nil
}

func (s *pincheService) AdminGetByID(id uint) (*dto.PincheInfo, error) {
	p, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPincheNotFound
		}
		return nil, err
	}
	return toPincheInfo(p), nil
}

// Audit 审核
func (s *pincheService) Audit(id uint, auditStatus int, auditReason string) error {
	p, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPincheNotFound
		}
		return err
	}
	if p.AuditStatus != model.PincheAuditPending {
		return ErrPincheAudited
	}
	return s.repo.UpdateAudit(id, auditStatus, auditReason)
}

// AdminUpdateStatus 管理后台强制下架/恢复
func (s *pincheService) AdminUpdateStatus(id uint, status int) error {
	return s.repo.UpdateStatus(id, status)
}

// BatchAudit 批量审核
func (s *pincheService) BatchAudit(req *dto.BatchAuditRequest) (*dto.BatchResultResponse, error) {
	result := &dto.BatchResultResponse{Total: len(req.IDs)}
	failedIDs := make([]uint, 0)
	for _, id := range req.IDs {
		if err := s.repo.UpdateAudit(id, req.AuditStatus, req.AuditReason); err != nil {
			failedIDs = append(failedIDs, id)
		} else {
			result.Success++
		}
	}
	result.Failed = len(failedIDs)
	result.FailedIDs = failedIDs
	return result, nil
}

// BatchUpdateStatus 批量状态变更
func (s *pincheService) BatchUpdateStatus(req *dto.BatchStatusUpdateRequest) (*dto.BatchResultResponse, error) {
	result := &dto.BatchResultResponse{Total: len(req.IDs)}
	failedIDs := make([]uint, 0)
	for _, id := range req.IDs {
		if err := s.repo.UpdateStatus(id, req.Status); err != nil {
			failedIDs = append(failedIDs, id)
		} else {
			result.Success++
		}
	}
	result.Failed = len(failedIDs)
	result.FailedIDs = failedIDs
	return result, nil
}

// BatchDelete 批量删除
func (s *pincheService) BatchDelete(req *dto.BatchDeleteRequest) (*dto.BatchResultResponse, error) {
	result := &dto.BatchResultResponse{Total: len(req.IDs)}
	failedIDs := make([]uint, 0)
	for _, id := range req.IDs {
		if err := s.repo.Delete(id); err != nil {
			failedIDs = append(failedIDs, id)
		} else {
			result.Success++
		}
	}
	result.Failed = len(failedIDs)
	result.FailedIDs = failedIDs
	return result, nil
}
