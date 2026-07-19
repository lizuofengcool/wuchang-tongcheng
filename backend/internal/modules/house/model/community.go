// Package model 小区信息表（对标贝壳/链家）
// 小区主页/位置/建筑年代/物业费/均价/开发商/物业
package model

import (
	"wuchang-tongcheng/internal/pkg/database"
)

// === 小区状态常量 ===
const (
	CommunityStatusDraft     = 0 // 草稿
	CommunityStatusPublished = 1 // 已发布
	CommunityStatusOffline   = 2 // 已下架
)

// === 建筑类型常量 ===
const (
	BuildingTypePlate     = "plate"     // 板楼
	BuildingTypeTower     = "tower"     // 塔楼
	BuildingTypePlateTower = "plate_tower" // 板塔结合
	BuildingTypeVilla     = "villa"     // 别墅
	BuildingTypeBungalow  = "bungalow"  // 平房
	BuildingTypeMixed     = "mixed"     // 混合
)

// HouseCommunity 小区信息表
type HouseCommunity struct {
	database.RegionBaseModel // id/region_id/created_at/updated_at/deleted_at
	Name             string     `gorm:"size:128;not null;uniqueIndex:uniq_house_communities_name_city" json:"name"`            // 小区名
	Alias            string     `gorm:"size:128;not null;default:''" json:"alias"`                                              // 别名
	City             string     `gorm:"size:64;not null;default:'';index" json:"city"`                                          // 城市
	District         string     `gorm:"size:64;not null;default:'';index" json:"district"`                                      // 行政区
	BusinessDistrict string     `gorm:"size:128;not null;default:'';index" json:"business_district"`                             // 商圈
	Address          string     `gorm:"size:500;not null;default:''" json:"address"`                                            // 详细地址
	Latitude         float64    `gorm:"type:decimal(10,7);default:0" json:"latitude"`                                           // 纬度
	Longitude        float64    `gorm:"type:decimal(10,7);default:0" json:"longitude"`                                          // 经度
	BuildingCount    int        `gorm:"not null;default:0" json:"building_count"`                                               // 楼栋数
	HouseCount       int        `gorm:"not null;default:0" json:"house_count"`                                                  // 房屋数
	BuildingYear     int        `gorm:"not null;default:0;index" json:"building_year"`                                          // 建造年代
	BuildingType     string     `gorm:"size:32;not null;default:''" json:"building_type"`                                       // 建筑类型
	Developer        string     `gorm:"size:128;not null;default:''" json:"developer"`                                          // 开发商
	PropertyCompany  string     `gorm:"size:128;not null;default:''" json:"property_company"`                                   // 物业公司
	PropertyFee      float64    `gorm:"type:decimal(8,2);default:0" json:"property_fee"`                                        // 物业费
	ParkingRatio     string     `gorm:"size:32;not null;default:''" json:"parking_ratio"`                                       // 车位配比
	GreeningRate     float64    `gorm:"type:decimal(5,2);default:0" json:"greening_rate"`                                       // 绿化率
	PlotRatio        float64    `gorm:"type:decimal(5,2);default:0" json:"plot_ratio"`                                          // 容积率
	AvgSalePrice     float64    `gorm:"type:decimal(10,2);default:0" json:"avg_sale_price"`                                     // 平均售价
	AvgRentPrice     float64    `gorm:"type:decimal(10,2);default:0" json:"avg_rent_price"`                                     // 平均租金
	Description      string     `gorm:"type:text" json:"description"`                                                            // 小区简介
	CoverImage       string     `gorm:"size:255;not null;default:''" json:"cover_image"`                                        // 封面图
	Images           JSONB      `gorm:"type:jsonb" json:"images"`                                                                // 小区图片 JSON
	NearbyPOIs       JSONB      `gorm:"type:jsonb" json:"nearby_pois"`                                                           // 附近 POI
	Status           int        `gorm:"default:1;index" json:"status"`                                                           // 0草稿 1已发布 2下架
	FollowerCount    int        `gorm:"default:0" json:"follower_count"`                                                         // 关注数
	OnSaleCount      int        `gorm:"default:0" json:"on_sale_count"`                                                          // 在售数
	OnRentCount      int        `gorm:"default:0" json:"on_rent_count"`                          // 在租数
}

// TableName 表名（house_ 前缀）
func (HouseCommunity) TableName() string { return "house_communities" }
