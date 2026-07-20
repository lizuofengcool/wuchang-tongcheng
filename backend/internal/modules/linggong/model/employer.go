// Package model 雇主认证表（对标斗米企业认证 + 猪八戒雇主）
// 企业/个人雇主 + 营业执照 + 法人认证 + 信用等级
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 雇主认证状态常量 ===
const (
	EmployerStatusPending  = 0 // 待审核
	EmployerStatusApproved = 1 // 已通过
	EmployerStatusRejected = 2 // 已拒绝
	EmployerStatusFrozen   = 3 // 已冻结
	EmployerStatusBanned   = 4 // 已封禁
)

// === 雇主类型常量 ===
const (
	EmployerTypePersonal = "personal" // 个人雇主
	EmployerTypeCompany  = "company"  // 企业雇主
	EmployerTypeAgent    = "agent"    // 中介
)

// === 雇主等级常量（信用等级） ===
const (
	EmployerLevelBronze   = 1 // 青铜
	EmployerLevelSilver   = 2 // 白银
	EmployerLevelGold     = 3 // 黄金
	EmployerLevelPlatinum = 4 // 铂金
	EmployerLevelDiamond  = 5 // 钻石
)

// === 认证类型常量 ===
const (
	EmployerCertBusinessLicense = "business_license" // 营业执照
	EmployerCertLegalPerson      = "legal_person"      // 法人认证
	EmployerCertIDCard           = "id_card"           // 身份证
	EmployerCertBankAccount      = "bank_account"      // 银行账户
	EmployerCertBrandAuth        = "brand_auth"        // 品牌授权
)

// LinggongEmployer 雇主认证表
type LinggongEmployer struct {
	database.RegionBaseModel // id/region_id/created_at/updated_at/deleted_at
	UserID           uint       `gorm:"not null;index;uniqueIndex:uniq_linggong_employers_user" json:"user_id"` // 用户 ID
	EmployerType     string     `gorm:"size:32;not null;default:'personal';index" json:"employer_type"`           // personal/company/agent
	CompanyName      string     `gorm:"size:128;not null;default:'';index" json:"company_name"`                   // 公司名称
	CompanyShortName string     `gorm:"size:64;not null;default:''" json:"company_short_name"`                    // 公司简称
	ContactName      string     `gorm:"size:50;not null;default:''" json:"contact_name"`                          // 联系人
	ContactPhone     string     `gorm:"size:20;not null;default:'';index" json:"contact_phone"`                  // 联系电话
	ContactEmail     string     `gorm:"size:128;not null;default:''" json:"contact_email"`                        // 联系邮箱
	ContactWechat    string     `gorm:"size:64;not null;default:''" json:"contact_wechat"`                        // 微信
	LicenseNo        string     `gorm:"size:64;not null;default:''" json:"license_no"`                            // 营业执照号
	LicenseURL       string     `gorm:"size:255;not null;default:''" json:"license_url"`                          // 营业执照图片
	LegalPerson      string     `gorm:"size:50;not null;default:''" json:"legal_person"`                          // 法人姓名
	LegalPersonIDCard string    `gorm:"size:32;not null;default:''" json:"legal_person_id_card"`                  // 法人身份证
	LegalPersonIDCardURL string  `gorm:"size:255;not null;default:''" json:"legal_person_id_card_url"`             // 法人身份证图片
	BankAccount      string     `gorm:"size:64;not null;default:''" json:"bank_account"`                          // 银行账号
	BankName         string     `gorm:"size:64;not null;default:''" json:"bank_name"`                            // 开户行
	BrandAuthURL     string     `gorm:"size:255;not null;default:''" json:"brand_auth_url"`                       // 品牌授权书
	CompanyAddress   string     `gorm:"size:500;not null;default:''" json:"company_address"`                      // 公司地址
	CompanyLatitude  float64    `gorm:"type:decimal(10,7);default:0" json:"company_latitude"`                       // 公司纬度
	CompanyLongitude float64    `gorm:"type:decimal(10,7);default:0" json:"company_longitude"`                     // 公司经度
	CompanyDescription string   `gorm:"type:text" json:"company_description"`                                    // 公司简介
	CompanyLogo      string     `gorm:"size:255;not null;default:''" json:"company_logo"`                         // 公司 Logo
	CompanyCover     string     `gorm:"size:255;not null;default:''" json:"company_cover"`                        // 公司封面图
	Industry         string     `gorm:"size:64;not null;default:''" json:"industry"`                              // 行业
	CompanySize      string     `gorm:"size:32;not null;default:''" json:"company_size"`                          // 公司规模
	Level            int        `gorm:"not null;default:1;index" json:"level"`                                    // 信用等级 1-5
	CreditScore      int        `gorm:"not null;default:100;index" json:"credit_score"`                            // 信用分
	Status           int        `gorm:"default:0;index" json:"status"`                                          // 0待审 1通过 2拒绝 3冻结 4封禁
	RejectReason     string     `gorm:"size:500;not null;default:''" json:"reject_reason"`                        // 拒绝原因
	VerifiedAt       *time.Time `gorm:"index" json:"verified_at"`                                                 // 认证时间
	VerifiedBy       uint       `gorm:"not null;default:0" json:"verified_by"`                                    // 审核人 ID
	VerifiedByName   string     `gorm:"size:50;not null;default:''" json:"verified_by_name"`                       // 审核人姓名
	PublishedCount   int        `gorm:"not null;default:0" json:"published_count"`                                // 已发布岗位数
	OngoingCount    int        `gorm:"not null;default:0" json:"ongoing_count"`                                  // 进行中岗位数
	CompletedCount  int        `gorm:"not null;default:0" json:"completed_count"`                                // 已完成岗位数
	TotalWorkers     int        `gorm:"not null;default:0" json:"total_workers"`                                  // 累计招工数
	TotalPaid        float64    `gorm:"type:decimal(12,2);default:0" json:"total_paid"`                            // 累计支付薪资
	AvgRating        float64    `gorm:"type:decimal(3,2);default:0" json:"avg_rating"`                            // 平均评分
	RatingCount      int        `gorm:"not null;default:0" json:"rating_count"`                                  // 评分次数
}

// TableName 表名（linggong_ 前缀）
func (LinggongEmployer) TableName() string { return "linggong_employers" }
