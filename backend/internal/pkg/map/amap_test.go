package amap

import (
	"context"
	"errors"
	"testing"

	"wuchang-tongcheng/internal/pkg/config"
)

// TestInitDegradation 验证 key 未配置或占位值时 Init 不激活服务
func TestInitDegradation(t *testing.T) {
	// 占位 key 不应激活
	Init(&config.MapConfig{Type: "amap", Key: "your-map-api-key"})
	if IsAvailable() {
		t.Error("占位 key 不应激活地图服务")
	}
	// 空 key 不应激活
	Init(&config.MapConfig{Type: "amap", Key: ""})
	if IsAvailable() {
		t.Error("空 key 不应激活地图服务")
	}
	// 非 amap 类型不应激活
	Init(&config.MapConfig{Type: "baidu", Key: "real-key"})
	if IsAvailable() {
		t.Error("非 amap 类型不应激活")
	}
}

// TestNotAvailableReturnsError 验证未激活时调用 API 返回 ErrNotAvailable
func TestNotAvailableReturnsError(t *testing.T) {
	// 重置为未激活态
	apiKey = ""
	ready = false

	_, err := Regeocode(context.Background(), 116.39, 39.9)
	if !errors.Is(err, ErrNotAvailable) {
		t.Errorf("未激活时 Regeocode 应返回 ErrNotAvailable，got %v", err)
	}
	_, err = Geocode(context.Background(), "北京")
	if !errors.Is(err, ErrNotAvailable) {
		t.Errorf("未激活时 Geocode 应返回 ErrNotAvailable，got %v", err)
	}
	_, err = Around(context.Background(), 116.39, 39.9, 1000, "", 10)
	if !errors.Is(err, ErrNotAvailable) {
		t.Errorf("未激活时 Around 应返回 ErrNotAvailable，got %v", err)
	}
}

// TestParseLocation 验证 "lng,lat" 解析
func TestParseLocation(t *testing.T) {
	loc, err := parseLocation("116.397428,39.90923")
	if err != nil {
		t.Fatalf("parseLocation 出错: %v", err)
	}
	if loc.Lng != 116.397428 || loc.Lat != 39.90923 {
		t.Errorf("解析结果错误: %+v", loc)
	}

	// 非法格式
	if _, err := parseLocation("invalid"); err == nil {
		t.Error("非法格式应返回错误")
	}
	if _, err := parseLocation("116.39,abc"); err == nil {
		t.Error("非数字经纬度应返回错误")
	}
}
