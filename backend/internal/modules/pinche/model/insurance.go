// Package model 拼车顺风车保险数据模型
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// PincheInsurance 顺风车保险
type PincheInsurance struct {
	database.RegionBaseModel

	PincheID  uint  `gorm:"index;not null" json:"pinche_id"`
	BookingID *uint `gorm:"index" json:"booking_id"`

	PolicyNo          string  `gorm:"size:64;index" json:"policy_no"`
	InsuranceCompany  string  `gorm:"size:64" json:"insurance_company"`
	InsuranceType     string  `gorm:"size:32;not null;default:'passenger'" json:"insurance_type"` // passenger/driver/both
	CoverageAmount    float64 `gorm:"type:decimal(12,2);default:0" json:"coverage_amount"`
	Premium           float64 `gorm:"type:decimal(12,2);default:0" json:"premium"`

	InsuredName   string `gorm:"size:50" json:"insured_name"`
	InsuredIDCard string `gorm:"size:32" json:"insured_id_card"`

	StartTime *time.Time `gorm:"index" json:"start_time"`
	EndTime   *time.Time `gorm:"index" json:"end_time"`

	Status       int        `gorm:"default:0;index" json:"status"` // 0待生效 1生效中 2已结束 3已理赔
	ClaimAmount  float64    `gorm:"type:decimal(12,2);default:0" json:"claim_amount"`
	ClaimReason  string     `gorm:"size:500" json:"claim_reason"`
	ClaimedAt    *time.Time `gorm:"index" json:"claimed_at"`

	Beneficiaries JSONB `gorm:"type:jsonb" json:"beneficiaries"`
}

// TableName 表名
func (PincheInsurance) TableName() string { return "pinche_insurances" }
