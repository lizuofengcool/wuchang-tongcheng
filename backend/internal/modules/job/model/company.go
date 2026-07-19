// Package model 公司信息表（对标 BOSS直聘）
// 公司主页/Logo/Banner/规模/行业/营业执照/法人
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 公司状态常量 ===
const (
	CompanyStatusPending  = 0 // 待审核
	CompanyStatusApproved = 1 // 已通过
	CompanyStatusRejected = 2 // 已拒绝
	CompanyStatusFrozen   = 3 // 已冻结
	CompanyStatusClosed   = 4 // 已关闭
)

// === 公司等级常量 ===
const (
	CompanyLevelNormal  = 0 // 普通企业
	CompanyLevelVerified = 1 // 认证企业
	CompanyLevelGold    = 2 // 金牌企业
	CompanyLevelDiamond = 3 // 钻石企业
)

// === 公司规模常量 ===
const (
	CompanyScaleLess20       = "0-20"          // 0-20人
	CompanyScale20To100      = "20-100"        // 20-100人
	CompanyScale100To500     = "100-500"       // 100-500人
	CompanyScale500To1000    = "500-1000"      // 500-1000人
	CompanyScale1000To10000  = "1000-10000"    // 1000-10000人
	CompanyScaleMore10000    = "10000+"        // 10000人以上
)

// JobCompany 公司信息表（对标 BOSS直聘）
type JobCompany struct {
	database.RegionBaseModel // id/region_id/created_at/updated_at/deleted_at
	UserID              uint       `gorm:"not null;uniqueIndex" json:"user_id"`                       // 店主用户 ID（一对一）
	Name                string     `gorm:"size:128;not null;index" json:"name"`                      // 公司全称
	ShortName           string     `gorm:"size:64" json:"short_name"`                                 // 公司简称
	Logo                string     `gorm:"size:255" json:"logo"`                                       // Logo URL
	Banner              string     `gorm:"size:255" json:"banner"`                                     // Banner URL
	Description         string     `gorm:"type:text" json:"description"`                              // 公司简介
	Industry            string     `gorm:"size:64;index" json:"industry"`                            // 所属行业
	Scale               string     `gorm:"size:32;index" json:"scale"`                                // 公司规模
	Level               int        `gorm:"default:0;index" json:"level"`                              // 0普通 1认证 2金牌 3钻石
	Status              int        `gorm:"default:0;index" json:"status"`                              // 0待审 1通过 2拒绝 3冻结 4关闭
	ContactName         string     `gorm:"size:50" json:"contact_name"`                              // 联系人
	ContactPhone        string     `gorm:"size:20" json:"contact_phone"`                            // 联系电话
	ContactEmail        string     `gorm:"size:128" json:"contact_email"`                            // 联系邮箱
	ContactWechat       string     `gorm:"size:50" json:"contact_wechat"`                            // 联系微信
	Address             string     `gorm:"size:500" json:"address"`                                    // 公司地址
	Latitude            float64    `gorm:"type:decimal(10,7)" json:"latitude"`                       // 纬度
	Longitude           float64    `gorm:"type:decimal(10,7)" json:"longitude"`                      // 经度
	BusinessLicense     string     `gorm:"size:255" json:"business_license"`                          // 营业执照 URL
	LicenseNo           string     `gorm:"size:64;index" json:"license_no"`                          // 执照编号
	IDCardFront         string     `gorm:"size:255" json:"id_card_front"`                              // 身份证正面 URL
	IDCardBack          string     `gorm:"size:255" json:"id_card_back"`                               // 身份证背面 URL
	LegalPerson         string     `gorm:"size:50" json:"legal_person"`                                // 法人代表
	LegalPersonIDCard   string     `gorm:"size:32" json:"legal_person_id_card"`                       // 法人身份证号
	RegisteredCapital   float64    `gorm:"type:decimal(14,2);default:0" json:"registered_capital"`     // 注册资本
	FoundedAt           *time.Time `gorm:"type:date" json:"founded_at"`                                // 成立日期
	Website             string     `gorm:"size:255" json:"website"`                                    // 公司网站
	VerifiedAt          *time.Time `gorm:"index" json:"verified_at"`                                   // 认证通过时间
	ApprovedAt          *time.Time `gorm:"index" json:"approved_at"`                                    // 审核通过时间
	RejectedReason      string     `gorm:"size:500" json:"rejected_reason"`                            // 拒绝原因
	ClosedAt            *time.Time `gorm:"index" json:"closed_at"`                                      // 关闭时间
	FollowerCount       int        `gorm:"default:0" json:"follower_count"`                          // 关注数
	JobCount            int        `gorm:"default:0" json:"job_count"`                                // 在招职位数
	EmployeeCount       int        `gorm:"default:0" json:"employee_count"`                          // 员工数
	ActiveJobCount      int        `gorm:"default:0" json:"active_job_count"`                        // 活跃职位数
	TotalHiredCount     int        `gorm:"default:0" json:"total_hired_count"`                       // 累计招聘数
	GoodRate            float64    `gorm:"type:decimal(5,2);default:100.00" json:"good_rate"`        // 好评率
	Deposit             float64    `gorm:"type:decimal(12,2);default:0" json:"deposit"`              // 保证金
	Tags                JSONB      `gorm:"type:jsonb" json:"tags"`                                      // 公司标签
}

// TableName 表名（job_ 前缀）
func (JobCompany) TableName() string { return "job_companies" }
