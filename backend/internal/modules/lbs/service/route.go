// Package service LBS地图中台业务逻辑层 - 路线规划
// 依据 v3.2.1 架构方案第 4.8 节：调用高德 API 计算距离和路线，预留接口
package service

import (
	"errors"
	"math"

	"wuchang-tongcheng/internal/modules/lbs/dto"
	"wuchang-tongcheng/internal/pkg/geo"
)

var (
	ErrRouteInvalid    = errors.New("路线规划参数错误")
	ErrRouteNotConfigured = errors.New("路线规划服务未配置高德 API Key，仅返回直线距离")
)

// RouteService 路线规划业务接口
type RouteService interface {
	// CalculateDistance 计算两点间直线距离（Haversine 公式）
	CalculateDistance(req *dto.DistanceRequest) (*dto.DistanceResponse, error)
	// PlanRoute 路线规划（高德 API，未配置时降级为直线距离）
	PlanRoute(req *dto.RouteRequest) (*dto.RouteResponse, error)
}

type routeService struct {
	amapKey string
}

// NewRouteService 创建路线规划 service 实例
// amapKey 为高德开放平台 API Key，为空时降级为本地直线距离计算
func NewRouteService(amapKey string) RouteService {
	return &routeService{amapKey: amapKey}
}

// CalculateDistance 计算两点间直线距离
func (s *routeService) CalculateDistance(req *dto.DistanceRequest) (*dto.DistanceResponse, error) {
	if req.FromLat == 0 || req.FromLng == 0 || req.ToLat == 0 || req.ToLng == 0 {
		return nil, ErrRouteInvalid
	}
	km := geo.HaversineKm(req.FromLat, req.FromLng, req.ToLat, req.ToLng)
	return &dto.DistanceResponse{
		StraightKm:    km,
		StraightMeter: km * 1000,
	}, nil
}

// PlanRoute 路线规划
// MVP 实现：未配置 amapKey 时，仅返回直线距离，duration/distance 估算
// 后续可对接高德 /v3/direction 路线规划 API
func (s *routeService) PlanRoute(req *dto.RouteRequest) (*dto.RouteResponse, error) {
	if req.FromLat == 0 || req.FromLng == 0 || req.ToLat == 0 || req.ToLng == 0 {
		return nil, ErrRouteInvalid
	}

	mode := req.Mode
	if mode == "" {
		mode = "driving"
	}

	straightKm := geo.HaversineKm(req.FromLat, req.FromLng, req.ToLat, req.ToLng)

	// MVP 阶段未配置高德 Key，降级为直线距离估算
	if s.amapKey == "" {
		// 各模式估算系数（直线距离 vs 实际路径）
		var distanceKm float64
		var speedKmH float64
		switch mode {
		case "walking":
			distanceKm = straightKm * 1.2
			speedKmH = 5
		case "riding":
			distanceKm = straightKm * 1.3
			speedKmH = 15
		case "transit":
			distanceKm = straightKm * 1.4
			speedKmH = 30
		case "driving":
			fallthrough
		default:
			mode = "driving"
			distanceKm = straightKm * 1.3
			speedKmH = 40
		}
		durationMin := int(math.Ceil(distanceKm / speedKmH * 60))
		if durationMin < 1 {
			durationMin = 1
		}
		return &dto.RouteResponse{
			Mode:        mode,
			DistanceKm:  distanceKm,
			DurationMin: durationMin,
			StraightKm:  straightKm,
			Source:      "local_estimate",
		}, nil
	}

	// TODO: 对接高德 /v3/direction API
	// 此处保留接口签名，后续实现时根据 mode 调用：
	//   driving:  /v3/direction/driving
	//   walking:  /v3/direction/walking
	//   transit:  /v3/direction/transit/integrated
	//   riding:   /v3/direction/riding
	return &dto.RouteResponse{
		Mode:        mode,
		DistanceKm:  straightKm,
		DurationMin: int(math.Ceil(straightKm / 40 * 60)),
		StraightKm:  straightKm,
		Source:      "amap_placeholder",
	}, nil
}
