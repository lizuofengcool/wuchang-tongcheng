// Package amap 高德地图 API 封装
// 提供地理编码、逆地理编码、周边 POI 搜索能力。
// key 未配置时 IsAvailable()=false，调用返回 ErrNotAvailable，业务可降级。
package amap

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"wuchang-tongcheng/internal/pkg/config"
)

// ErrNotAvailable 地图服务不可用（key 未配置或类型非 amap）
var ErrNotAvailable = errors.New("map service not available: key not configured")

const (
	amapBaseURL = "https://restapi.amap.com/v3"
	httpTimeout = 5 * time.Second
)

var (
	apiKey string
	ready  bool
)

// Init 初始化高德地图客户端
func Init(cfg *config.MapConfig) {
	if cfg == nil {
		return
	}
	// 仅支持 amap，baidu 待补齐
	if strings.ToLower(cfg.Type) == "amap" && cfg.Key != "" && cfg.Key != "your-map-api-key" {
		apiKey = cfg.Key
		ready = true
	}
}

// IsAvailable 检查地图服务是否可用
func IsAvailable() bool {
	return ready
}

// GetKey 暴露 key（供构造请求 URL）
func GetKey() string {
	return apiKey
}

// Location 经纬度
type Location struct {
	Lng float64 `json:"lng"` // 经度
	Lat float64 `json:"lat"` // 纬度
}

// AddressComponent 逆地理编码返回的行政区划
type AddressComponent struct {
	Province string `json:"province"` // 省
	City     string `json:"city"`     // 市（直辖市可能为空）
	District string `json:"district"` // 区县
	Street   string `json:"street"`   // 街道
	Number   string `json:"number"`   // 门牌号
}

// RegeocodeResult 逆地理编码结果
type RegeocodeResult struct {
	FormattedAddress string           `json:"formatted_address"` // 完整地址
	Province         string           `json:"province"`
	City             string           `json:"city"`
	District         string           `json:"district"`
	Location         Location         `json:"location"`
	Components       AddressComponent `json:"components"`
}

// GeocodeResult 地理编码结果
type GeocodeResult struct {
	Location Location `json:"location"`
	Level    string   `json:"level"` // 精度级别
}

// POI 周边兴趣点
type POI struct {
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	Address  string   `json:"address"`
	Location Location `json:"location"`
	Tel      string   `json:"tel"`
	Distance string   `json:"distance"` // 距离（米）
}

// amapResponse 高德 API 通用响应壳
type amapResponse struct {
	Status   string `json:"status"`   // "1" 成功
	Info     string `json:"info"`     // 状态说明
	InfoCode string `json:"infocode"` // 状态码
}

// Regeocode 逆地理编码：经纬度 → 地址
func Regeocode(ctx context.Context, lng, lat float64) (*RegeocodeResult, error) {
	if !ready {
		return nil, ErrNotAvailable
	}
	loc := fmt.Sprintf("%f,%f", lng, lat)
	params := url.Values{
		"key":      {apiKey},
		"location": {loc},
		"output":   {"json"},
		"extensions": {"base"},
	}
	var resp struct {
		amapResponse
		Regeocode struct {
			FormattedAddress string           `json:"formatted_address"`
			AddressComponent AddressComponent `json:"addressComponent"`
		} `json:"regeocode"`
	}
	if err := doGet(ctx, "/geocode/regeo", params, &resp); err != nil {
		return nil, err
	}
	if resp.Status != "1" {
		return nil, fmt.Errorf("amap regeo failed: %s (code %s)", resp.Info, resp.InfoCode)
	}
	return &RegeocodeResult{
		FormattedAddress: resp.Regeocode.FormattedAddress,
		Province:         resp.Regeocode.AddressComponent.Province,
		City:             resp.Regeocode.AddressComponent.City,
		District:         resp.Regeocode.AddressComponent.District,
		Location:         Location{Lng: lng, Lat: lat},
		Components:        resp.Regeocode.AddressComponent,
	}, nil
}

// Geocode 地理编码：地址 → 经纬度
func Geocode(ctx context.Context, address string) (*GeocodeResult, error) {
	if !ready {
		return nil, ErrNotAvailable
	}
	params := url.Values{
		"key":     {apiKey},
		"address": {address},
		"output":  {"json"},
	}
	var resp struct {
		amapResponse
		Geocodes []struct {
			Location string `json:"location"` // "lng,lat"
			Level    string `json:"level"`
		} `json:"geocodes"`
	}
	if err := doGet(ctx, "/geocode/geo", params, &resp); err != nil {
		return nil, err
	}
	if resp.Status != "1" || len(resp.Geocodes) == 0 {
		return nil, fmt.Errorf("amap geo failed: %s (code %s)", resp.Info, resp.InfoCode)
	}
	loc, err := parseLocation(resp.Geocodes[0].Location)
	if err != nil {
		return nil, err
	}
	return &GeocodeResult{Location: loc, Level: resp.Geocodes[0].Level}, nil
}

// Around 周边搜索：经纬度 + 半径 + 类型 → POI 列表
func Around(ctx context.Context, lng, lat float64, radius int, types string, limit int) ([]POI, error) {
	if !ready {
		return nil, ErrNotAvailable
	}
	if limit <= 0 || limit > 25 {
		limit = 10
	}
	params := url.Values{
		"key":      {apiKey},
		"location": {fmt.Sprintf("%f,%f", lng, lat)},
		"radius":   {strconv.Itoa(radius)},
		"output":   {"json"},
		"offset":   {strconv.Itoa(limit)},
	}
	if types != "" {
		params.Set("types", types)
	}
	var resp struct {
		amapResponse
		Pois []struct {
			Name     string `json:"name"`
			Type     string `json:"type"`
			Address  string `json:"address"`
			Location string `json:"location"`
			Tel      string `json:"tel"`
			Distance string `json:"distance"`
		} `json:"pois"`
	}
	if err := doGet(ctx, "/place/around", params, &resp); err != nil {
		return nil, err
	}
	if resp.Status != "1" {
		return nil, fmt.Errorf("amap around failed: %s (code %s)", resp.Info, resp.InfoCode)
	}
	pois := make([]POI, 0, len(resp.Pois))
	for _, p := range resp.Pois {
		loc, _ := parseLocation(p.Location)
		pois = append(pois, POI{
			Name: p.Name, Type: p.Type, Address: p.Address,
			Location: loc, Tel: p.Tel, Distance: p.Distance,
		})
	}
	return pois, nil
}

// doGet 发起 GET 请求并解析 JSON
func doGet(ctx context.Context, path string, params url.Values, dst interface{}) error {
	reqURL := amapBaseURL + path + "?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: httpTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("amap request failed: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, dst)
}

// parseLocation 解析 "lng,lat" 字符串
func parseLocation(s string) (Location, error) {
	parts := strings.Split(s, ",")
	if len(parts) != 2 {
		return Location{}, fmt.Errorf("invalid location format: %s", s)
	}
	lng, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return Location{}, err
	}
	lat, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return Location{}, err
	}
	return Location{Lng: lng, Lat: lat}, nil
}
