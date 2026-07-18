// Package geo 地理空间计算单元测试（纯函数，无 DB/外部依赖）。
package geo

import (
	"math"
	"testing"
)

func approxEqual(a, b, eps float64) bool {
	return math.Abs(a-b) <= eps
}

func TestHaversineKm_SamePoint(t *testing.T) {
	if got := HaversineKm(39.9, 116.4, 39.9, 116.4); got != 0 {
		t.Errorf("同一点距离应为 0，得到 %v", got)
	}
}

func TestHaversineKm_Symmetry(t *testing.T) {
	d1 := HaversineKm(31.2304, 121.4737, 39.9042, 116.4074)
	d2 := HaversineKm(39.9042, 116.4074, 31.2304, 121.4737)
	if !approxEqual(d1, d2, 1e-6) {
		t.Errorf("距离不满足对称性: %v vs %v", d1, d2)
	}
}

func TestHaversineKm_BeijingShanghai(t *testing.T) {
	// 北京 → 上海 直线距离约 1067 km
	got := HaversineKm(39.9042, 116.4074, 31.2304, 121.4737)
	if !approxEqual(got, 1067, 5) {
		t.Errorf("北京-上海距离 %v，期望约 1067 km（±5）", got)
	}
}

func TestHaversineKm_EquatorOneDegree(t *testing.T) {
	// 赤道上经度相差 1 度 ≈ 111.19 km
	got := HaversineKm(0, 0, 0, 1)
	if !approxEqual(got, 111.19, 0.5) {
		t.Errorf("赤道 1 度距离 %v，期望约 111.19 km", got)
	}
}

func TestHaversineKm_Antipode(t *testing.T) {
	// 对跖点距离 ≈ 地球半周长 20015 km
	got := HaversineKm(0, 0, 0, 180)
	if !approxEqual(got, 20015, 50) {
		t.Errorf("对跖点距离 %v，期望约 20015 km", got)
	}
}

func TestHaversineKm_HarbinWuchang(t *testing.T) {
	// 哈尔滨 (45.8038, 126.5350) → 五常 (44.9225, 127.1500) 约 105 km
	got := HaversineKm(45.8038, 126.5350, 44.9225, 127.1500)
	if !approxEqual(got, 105, 5) {
		t.Errorf("哈尔滨-五常距离 %v，期望约 105 km", got)
	}
}

func TestBoundingBox_Equator(t *testing.T) {
	// 赤道中心，半径 111 km：纬度 ±1 度，经度 ±1 度（cos(0)=1）
	minLat, maxLat, minLng, maxLng := BoundingBox(0, 0, 111)
	if !approxEqual(minLat, -1, 1e-6) || !approxEqual(maxLat, 1, 1e-6) {
		t.Errorf("赤道纬度边界 (%v,%v)，期望 ±1", minLat, maxLat)
	}
	if !approxEqual(minLng, -1, 1e-6) || !approxEqual(maxLng, 1, 1e-6) {
		t.Errorf("赤道经度边界 (%v,%v)，期望 ±1", minLng, maxLng)
	}
}

func TestBoundingBox_CenterInside(t *testing.T) {
	lat, lng, r := 45.0, 126.0, 50.0
	minLat, maxLat, minLng, maxLng := BoundingBox(lat, lng, r)
	if lat < minLat || lat > maxLat || lng < minLng || lng > maxLng {
		t.Errorf("中心点 (%v,%v) 不在边界框 (%v~%v, %v~%v) 内", lat, lng, minLat, maxLat, minLng, maxLng)
	}
}

func TestBoundingBox_HigherLatitudeNarrowerLng(t *testing.T) {
	// 纬度 60 度，经度 1 度的距离约为赤道的一半（cos(60)=0.5）
	// 半径 111 km → 纬度 ±1，经度 ±2
	minLat, maxLat, minLng, maxLng := BoundingBox(60, 0, 111)
	if !approxEqual(maxLat-minLat, 2, 1e-6) {
		t.Errorf("纬度 60 度纬度跨度 %v，期望 2", maxLat-minLat)
	}
	if !approxEqual(maxLng-minLng, 4, 1e-6) {
		t.Errorf("纬度 60 度经度跨度 %v，期望 4（cos60=0.5）", maxLng-minLng)
	}
}

func TestBoundingBox_ZeroRadius(t *testing.T) {
	minLat, maxLat, minLng, maxLng := BoundingBox(45, 126, 0)
	if minLat != 45 || maxLat != 45 || minLng != 126 || maxLng != 126 {
		t.Errorf("零半径边界框应退化为中心点，得到 (%v~%v,%v~%v)", minLat, maxLat, minLng, maxLng)
	}
}

func TestBoundingBox_NegativeRadius(t *testing.T) {
	minLat, maxLat, _, _ := BoundingBox(45, 126, -10)
	if minLat != 45 || maxLat != 45 {
		t.Errorf("负半径应按 0 处理，得到纬度 (%v~%v)", minLat, maxLat)
	}
}

func TestBoundingBox_PoleDegenerate(t *testing.T) {
	// 极点附近 cosLat → 0，经度退化为全范围
	_, _, minLng, maxLng := BoundingBox(90, 0, 50)
	if minLng != -180 || maxLng != 180 {
		t.Errorf("极点经度应退化为 ±180，得到 (%v~%v)", minLng, maxLng)
	}
}

func TestPostGISAvailable_NilDB(t *testing.T) {
	ResetForTest()
	if PostGISAvailable(nil) {
		t.Error("nil db 应回报 PostGIS 不可用")
	}
}

func TestResetForTest(t *testing.T) {
	// 确保重置后探测缓存可重新求值（不 panic 即可）
	ResetForTest()
	PostGISAvailable(nil)
	ResetForTest()
}
