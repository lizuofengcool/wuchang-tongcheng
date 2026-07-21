// Package service LBS地图中台业务逻辑层 - 地理围栏
package service

import (
	"errors"
	"math"

	"wuchang-tongcheng/internal/modules/lbs/dto"
	"wuchang-tongcheng/internal/modules/lbs/model"
	"wuchang-tongcheng/internal/modules/lbs/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrGeofenceNotFound     = errors.New("围栏不存在")
	ErrGeofenceTypeInvalid  = errors.New("围栏类型错误")
	ErrGeofencePointsInvalid = errors.New("围栏顶点格式错误")
	ErrGeofenceStatusInvalid = errors.New("围栏状态不允许此操作")
)

// GeofenceService 围栏业务接口
type GeofenceService interface {
	Create(req *dto.CreateGeofenceRequest) (*dto.GeofenceInfo, error)
	Update(id uint, req *dto.UpdateGeofenceRequest) error
	Delete(id uint) error
	GetByID(id uint) (*dto.GeofenceInfo, error)
	List(req *dto.GeofenceListRequest) (*utils.Pagination, []dto.GeofenceInfo, error)
	ListByRegion(regionID uint) ([]dto.GeofenceInfo, error)
	ListByOwner(ownerID uint, ownerType string) ([]dto.GeofenceInfo, error)

	CheckPoint(id uint, lat, lng float64) (*dto.CheckPointResponse, error)
	CheckPointInRegion(regionID uint, lat, lng float64) (*dto.CheckPointResponse, error)

	AdminUpdateStatus(id uint, status int) error
}

type geofenceService struct {
	repo repository.GeofenceRepository
}

// NewGeofenceService 创建围栏 service 实例
func NewGeofenceService(repo repository.GeofenceRepository) GeofenceService {
	return &geofenceService{repo: repo}
}

// geofenceStatusText 状态文本
func geofenceStatusText(status int) string {
	switch status {
	case model.LBSGeofenceStatusDisabled:
		return "禁用"
	case model.LBSGeofenceStatusEnabled:
		return "启用"
	}
	return ""
}

// toGeofenceInfo model → dto
func toGeofenceInfo(g *model.Geofence) *dto.GeofenceInfo {
	info := &dto.GeofenceInfo{
		ID:          g.ID,
		RegionID:    g.RegionID,
		Name:        g.Name,
		Type:        g.Type,
		Status:      g.Status,
		StatusText:  geofenceStatusText(g.Status),
		Sort:        g.Sort,
		Description: g.Description,
		CenterLat:   g.CenterLat,
		CenterLng:   g.CenterLng,
		Radius:      g.Radius,
		OwnerID:     g.OwnerID,
		OwnerType:   g.OwnerType,
		CreatedAt:   g.CreatedAt,
		UpdatedAt:   g.UpdatedAt,
	}
	if g.Points != nil {
		info.Points = g.Points
	}
	if g.Extra != nil {
		info.Extra = g.Extra
	}
	return info
}

// Create 创建围栏
func (s *geofenceService) Create(req *dto.CreateGeofenceRequest) (*dto.GeofenceInfo, error) {
	g := &model.Geofence{
		RegionID:    req.RegionID,
		Name:        req.Name,
		Type:        req.Type,
		Status:      req.Status,
		Sort:        req.Sort,
		Description: req.Description,
		CenterLat:   req.CenterLat,
		CenterLng:   req.CenterLng,
		Radius:      req.Radius,
		OwnerID:     req.OwnerID,
		OwnerType:   req.OwnerType,
	}
	if g.Status == 0 {
		g.Status = model.LBSGeofenceStatusEnabled
	}
	if req.Points != nil {
		if b, err := model.FromJSON(req.Points); err == nil {
			g.Points = b
		}
	}
	if req.Extra != nil {
		if b, err := model.FromJSON(req.Extra); err == nil {
			g.Extra = b
		}
	}
	if err := s.validateGeofence(g); err != nil {
		return nil, err
	}
	if err := s.repo.Create(g); err != nil {
		return nil, err
	}
	return toGeofenceInfo(g), nil
}

// Update 更新围栏
func (s *geofenceService) Update(id uint, req *dto.UpdateGeofenceRequest) error {
	g, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrGeofenceNotFound
		}
		return err
	}
	if req.RegionID != nil {
		g.RegionID = *req.RegionID
	}
	if req.Name != nil {
		g.Name = *req.Name
	}
	if req.Type != nil {
		g.Type = *req.Type
	}
	if req.Status != nil {
		g.Status = *req.Status
	}
	if req.Sort != nil {
		g.Sort = *req.Sort
	}
	if req.Description != nil {
		g.Description = *req.Description
	}
	if req.CenterLat != nil {
		g.CenterLat = *req.CenterLat
	}
	if req.CenterLng != nil {
		g.CenterLng = *req.CenterLng
	}
	if req.Radius != nil {
		g.Radius = *req.Radius
	}
	if req.OwnerID != nil {
		g.OwnerID = *req.OwnerID
	}
	if req.OwnerType != nil {
		g.OwnerType = *req.OwnerType
	}
	if req.Points != nil {
		if b, err := model.FromJSON(req.Points); err == nil {
			g.Points = b
		}
	}
	if req.Extra != nil {
		if b, err := model.FromJSON(req.Extra); err == nil {
			g.Extra = b
		}
	}
	if err := s.validateGeofence(g); err != nil {
		return err
	}
	return s.repo.Update(g)
}

// validateGeofence 校验围栏参数
func (s *geofenceService) validateGeofence(g *model.Geofence) error {
	switch g.Type {
	case model.LBSGeofenceTypeCircle:
		if g.Radius <= 0 {
			return ErrGeofencePointsInvalid
		}
		if g.CenterLat == 0 || g.CenterLng == 0 {
			return ErrGeofencePointsInvalid
		}
	case model.LBSGeofenceTypePolygon:
		if g.Points == nil || len(g.Points) == 0 {
			return ErrGeofencePointsInvalid
		}
		var points []model.GeofencePoint
		if err := g.Points.Parse(&points); err != nil {
			return ErrGeofencePointsInvalid
		}
		if len(points) < 3 {
			return ErrGeofencePointsInvalid
		}
	default:
		return ErrGeofenceTypeInvalid
	}
	return nil
}

// Delete 删除围栏
func (s *geofenceService) Delete(id uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrGeofenceNotFound
		}
		return err
	}
	return s.repo.Delete(id)
}

// GetByID 获取围栏详情
func (s *geofenceService) GetByID(id uint) (*dto.GeofenceInfo, error) {
	g, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGeofenceNotFound
		}
		return nil, err
	}
	return toGeofenceInfo(g), nil
}

// List 围栏列表
func (s *geofenceService) List(req *dto.GeofenceListRequest) (*utils.Pagination, []dto.GeofenceInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.GeofenceListOptions{
		RegionID:  req.RegionID,
		OwnerID:   req.OwnerID,
		OwnerType: req.OwnerType,
		Type:      req.Type,
		Status:    req.Status,
		Keyword:   req.Keyword,
	}
	list, total, err := s.repo.List(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	infos := make([]dto.GeofenceInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toGeofenceInfo(&list[i]))
	}
	return pagination, infos, nil
}

// ListByRegion 按区域列出
func (s *geofenceService) ListByRegion(regionID uint) ([]dto.GeofenceInfo, error) {
	list, err := s.repo.ListByRegion(regionID)
	if err != nil {
		return nil, err
	}
	infos := make([]dto.GeofenceInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toGeofenceInfo(&list[i]))
	}
	return infos, nil
}

// ListByOwner 按所有者列出
func (s *geofenceService) ListByOwner(ownerID uint, ownerType string) ([]dto.GeofenceInfo, error) {
	list, err := s.repo.ListByOwner(ownerID, ownerType)
	if err != nil {
		return nil, err
	}
	infos := make([]dto.GeofenceInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toGeofenceInfo(&list[i]))
	}
	return infos, nil
}

// CheckPoint 检查点是否在指定围栏内
func (s *geofenceService) CheckPoint(id uint, lat, lng float64) (*dto.CheckPointResponse, error) {
	g, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGeofenceNotFound
		}
		return nil, err
	}
	inside := s.isPointInGeofence(g, lat, lng)
	return &dto.CheckPointResponse{
		Inside:     inside,
		GeofenceID: g.ID,
		Name:       g.Name,
	}, nil
}

// CheckPointInRegion 检查点是否在区域内任一围栏内（返回第一个匹配）
func (s *geofenceService) CheckPointInRegion(regionID uint, lat, lng float64) (*dto.CheckPointResponse, error) {
	list, err := s.repo.ListEnabledByRegion(regionID)
	if err != nil {
		return nil, err
	}
	for i := range list {
		if s.isPointInGeofence(&list[i], lat, lng) {
			return &dto.CheckPointResponse{
				Inside:     true,
				GeofenceID: list[i].ID,
				Name:       list[i].Name,
			}, nil
		}
	}
	return &dto.CheckPointResponse{
		Inside: false,
	}, nil
}

// isPointInGeofence 判断点是否在围栏内
// 圆形：Haversine 距离 ≤ 半径
// 多边形：射线法 point-in-polygon
func (s *geofenceService) isPointInGeofence(g *model.Geofence, lat, lng float64) bool {
	switch g.Type {
	case model.LBSGeofenceTypeCircle:
		return s.isPointInCircle(lat, lng, g.CenterLat, g.CenterLng, g.Radius)
	case model.LBSGeofenceTypePolygon:
		return s.isPointInPolygon(lat, lng, g.Points)
	}
	return false
}

// isPointInCircle 点是否在圆形围栏内
// lat/lng: 待判断点；centerLat/CenterLng: 圆心；radiusMeter: 半径（米）
func (s *geofenceService) isPointInCircle(lat, lng, centerLat, centerLng, radiusMeter float64) bool {
	if radiusMeter <= 0 || centerLat == 0 || centerLng == 0 {
		return false
	}
	distance := haversineMeter(lat, lng, centerLat, centerLng)
	return distance <= radiusMeter
}

// isPointInPolygon 点是否在多边形内（射线法）
// points JSONB 解析为 [{lat,lng},...]
func (s *geofenceService) isPointInPolygon(lat, lng float64, pointsJSON model.JSONB) bool {
	if pointsJSON == nil || len(pointsJSON) == 0 {
		return false
	}
	var points []model.GeofencePoint
	if err := pointsJSON.Parse(&points); err != nil {
		return false
	}
	if len(points) < 3 {
		return false
	}

	// 射线法：从待判断点向右发射水平射线，统计与多边形边的交点数
	// 奇数次相交 → 在内部；偶数次 → 在外部
	n := len(points)
	inside := false
	j := n - 1
	for i := 0; i < n; i++ {
		yi := points[i].Latitude
		xi := points[i].Longitude
		yj := points[j].Latitude
		xj := points[j].Longitude

		// 判断边的两个端点是否在待判断点的上下两侧（yi>y 且 yj<=y 或反之）
		if (yi > lat) != (yj > lat) {
			// 计算射线与边的交点 x 坐标
			xIntersect := (yj-yi)*(lng-xi)/(yj-yi) + xi
			// 等价于 xIntersect := xi + (lng-yi)*(xj-xi)/(yj-yi)
			// 注：上面公式存在符号问题，正确公式为：
			//   xIntersect = (xj - xi) * (lat - yi) / (yj - yi) + xi
			// 但传统 PNPOLY 算法使用：
			//   if (yi > lat) != (yj > lat) && lng < (xj - xi) * (lat - yi) / (yj - yi) + xi
			// 这里使用 PNPOLY 标准实现：
			xIntersect = (xj-xi)*(lat-yi)/(yj-yi) + xi
			if lng < xIntersect {
				inside = !inside
			}
		}
		j = i
	}
	return inside
}

// haversineMeter Haversine 公式计算两点距离（米）
func haversineMeter(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadiusMeter = 6371000.0
	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Asin(math.Sqrt(a))
	return earthRadiusMeter * c
}

// AdminUpdateStatus 管理后台更新状态
func (s *geofenceService) AdminUpdateStatus(id uint, status int) error {
	if status != model.LBSGeofenceStatusDisabled && status != model.LBSGeofenceStatusEnabled {
		return ErrGeofenceStatusInvalid
	}
	return s.repo.UpdateFields(id, map[string]interface{}{"status": status})
}
