// Package model 车险配置表（对标平安/太平洋/人保）
// 交强/商业/第三方/赔付额/免赔额
package model

import (
	"wuchang-tongcheng/internal/pkg/database"
)

// === 车险类型常量 ===
const (
	InsuranceTypeCompulsory = "compulsory" // 交强险
	InsuranceTypeCommercial = "commercial" // 商业险
	InsuranceTypeThirdParty = "third_party" // 第三方责任险
	InsuranceTypeTheft      = "theft"       // 盗抢险
	InsuranceTypeFire       = "fire"        // 自燃险
	InsuranceTypeGlass      = "glass"       // 玻璃险
	InsuranceTypeEngine     = "engine"      // 发动机涉水险
	InsuranceTypeNoDeduct   = "no_deduct"   // 不计免赔
	InsuranceTypePassenger  = "passenger"   // 车上人员责任险
	InsuranceTypeScratch    = "scratch"     // 划痕险
)

// === 状态常量 ===
const (
	InsurancePlanStatusDraft     = 0 // 草稿
	InsurancePlanStatusPublished = 1 // 已发布
	InsurancePlanStatusOffline   = 2 // 已下架
)

// CarInsurance 车险配置表
type CarInsurance struct {
	database.BaseModel // id/created_at/updated_at/deleted_at
	Name              string  `gorm:"size:64;not null" json:"name"`                                  // 险种名
	Code              string  `gorm:"size:64;not null;uniqueIndex" json:"code"`                      // 险种编码
	InsuranceType     string  `gorm:"size:32;not null;default:'compulsory';index" json:"insurance_type"` // compulsory/commercial/third_party 等
	Provider          string  `gorm:"size:128;not null;default:'';index" json:"provider"`            // 保险公司
	Coverage          string  `gorm:"size:64;not null;default:'';index" json:"coverage"`              // 保障范围
	CoverageAmount    float64 `gorm:"type:decimal(14,2);default:0" json:"coverage_amount"`            // 保额
	Premium           float64 `gorm:"type:decimal(14,2);default:0" json:"premium"`                   // 保费
	Deductible        int     `gorm:"not null;default:0" json:"deductible"`                          // 免赔额
	Description       string  `gorm:"size:500;not null;default:''" json:"description"`               // 险种说明
	Sort              int     `gorm:"not null;default:0;index" json:"sort"`                          // 排序
	Status            int     `gorm:"default:1;index" json:"status"`                                 // 0草稿 1已发布 2下架
	IsHot             bool    `gorm:"not null;default:false;index" json:"is_hot"`                    // 热门
	UseCount          int     `gorm:"not null;default:0" json:"use_count"`                           // 购买次数
}

// TableName 表名（car_ 前缀）
func (CarInsurance) TableName() string { return "car_insurance" }
