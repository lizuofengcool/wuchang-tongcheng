// Package geo 地理空间计算与 PostGIS 能力探测。
// 提供 Haversine 距离、经纬度边界框计算（用于 SQL 预筛），
// 以及 PostGIS 扩展可用性检测（不可用时业务层降级走纯 SQL Haversine）。
package geo

import (
	"math"
	"sync"

	"gorm.io/gorm"
)

// EarthRadiusKm 地球平均半径（km），与 PostGIS geography 默认 SPHEROID 一致量级
const EarthRadiusKm = 6371.0

// 地球赤道周长约 40075 km，1 度纬度 ≈ 111.0 km
const kmPerDegreeLat = 111.0

var (
	postgisOnce    sync.Once
	postgisChecked bool
	postgisEnabled bool
)

// HaversineKm 使用 Haversine 公式计算两个经纬度坐标之间的球面距离（公里）。
// 与 PostGIS ST_Distance(geography) 结果在米级精度内一致，用于降级路径。
func HaversineKm(lat1, lng1, lat2, lng2 float64) float64 {
	la1 := toRadians(lat1)
	la2 := toRadians(lat2)
	dlat := toRadians(lat2 - lat1)
	dlng := toRadians(lng2 - lng1)

	a := math.Sin(dlat/2)*math.Sin(dlat/2) +
		math.Cos(la1)*math.Cos(la2)*math.Sin(dlng/2)*math.Sin(dlng/2)
	c := 2 * math.Asin(math.Sqrt(a))
	return EarthRadiusKm * c
}

// BoundingBox 根据中心点与半径（km）计算经纬度边界框，
// 用于在 SQL 中先用 BETWEEN 预筛候选行（走普通 btree 索引），
// 再由 Haversine/ST_DWithin 精确过滤。半径较大时边界框为粗略近似。
func BoundingBox(lat, lng, radiusKm float64) (minLat, maxLat, minLng, maxLng float64) {
	if radiusKm < 0 {
		radiusKm = 0
	}
	latDelta := radiusKm / kmPerDegreeLat
	minLat = lat - latDelta
	maxLat = lat + latDelta

	// 经度 1 度距离随纬度变化：cos(lat)
	cosLat := math.Cos(toRadians(lat))
	if cosLat < 1e-6 {
		// 极点附近退化为全经度范围
		minLng = -180
		maxLng = 180
	} else {
		lngDelta := radiusKm / (kmPerDegreeLat * cosLat)
		minLng = lng - lngDelta
		maxLng = lng + lngDelta
	}
	return
}

// PostGISAvailable 探测当前数据库连接是否启用了 PostGIS 扩展。
// 结果在进程生命周期内缓存（扩展安装是 DB 级别一次性操作）。
// 探测失败或扩展缺失时返回 false，调用方应降级走 Haversine SQL。
func PostGISAvailable(db *gorm.DB) bool {
	if db == nil {
		return false
	}
	postgisOnce.Do(func() {
		postgisChecked = true
		var v interface{}
		// postgis_version() 仅在安装了 postgis 扩展的库中存在
		err := db.Raw("SELECT postgis_version()").Row().Scan(&v)
		if err == nil && v != nil {
			postgisEnabled = true
		}
	})
	return postgisEnabled
}

// ResetForTest 重置 PostGIS 探测缓存，仅用于测试。
func ResetForTest() {
	postgisOnce = sync.Once{}
	postgisChecked = false
	postgisEnabled = false
}

func toRadians(deg float64) float64 {
	return deg * math.Pi / 180.0
}
