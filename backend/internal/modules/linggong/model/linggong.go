// Package model 同城零工兼职数据模型
// 依据 v3.2.1 架构方案第六章：对标斗米/青团兼职/兼职猫/猪八戒
// 依据需求文档 1.10：4 维数据隔离（region_id 地区隔离 + user_id 用户隔离）
// 依据需求文档 7.1：通用字段 id/region_id/created_at/updated_at/deleted_at + status + audit_status
// 依据需求文档 7.2：主表 linggongs（保持兼容已发布数据）
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 岗位状态常量 ===
const (
	LinggongStatusDraft       = 0 // 草稿
	LinggongStatusPublished   = 1 // 已发布
	LinggongStatusOffline     = 2 // 已下架
	LinggongStatusExpired     = 3 // 已过期
	LinggongStatusDeleted     = 4 // 已删除
	LinggongStatusFulfilled   = 5 // 已满员
	LinggongStatusClosed      = 6 // 已关闭
	LinggongStatusCompleted   = 7 // 已完成
)

// === 审核状态常量（依据需求文档 7.1） ===
const (
	LinggongAuditPending  = 0 // 待审
	LinggongAuditApproved = 1 // 通过
	LinggongAuditRejected = 2 // 拒绝
)

// === 岗位类型常量（长短期分类） ===
const (
	LinggongTypeShortTerm = "short_term" // 短期兼职（≤7 天）
	LinggongTypeLongTerm  = "long_term"  // 长期兼职（>7 天）
	LinggongTypeTask      = "task"       // 任务制（按件计费）
	LinggongTypeHourly    = "hourly"     // 小时工
	LinggongTypeDaily     = "daily"       // 日结工
	LinggongTypeTemp      = "temp"        // 临时工
)

// === 计费方式常量 ===
const (
	BillingTypeByPiece = "by_piece" // 按件
	BillingTypeByHour  = "by_hour"  // 按时
	BillingTypeByDay   = "by_day"   // 按日
	BillingTypeByWeek  = "by_week" // 按周
	BillingTypeByMonth = "by_month" // 按月
	BillingTypeFixed   = "fixed"   // 固定
	BillingTypeNegotiable = "negotiable" // 面议
)

// === 结算周期常量（薪资日结） ===
const (
	SettlementT0 = "T+0" // 当日结
	SettlementT1 = "T+1" // 次日结
	SettlementT3 = "T+3" // 三日结
	SettlementT7 = "T+7" // 周结
	SettlementM1 = "M+1" // 月结
	SettlementProject = "project" // 项目结
)

// === 发布者类型常量 ===
const (
	PublisherTypePersonal = "personal" // 个人雇主
	PublisherTypeCompany  = "company"  // 企业雇主
	PublisherTypeAgent    = "agent"    // 中介
	PublisherTypeHeadhunter = "headhunter" // 猎头
)

// === 工作强度常量 ===
const (
	WorkIntensityLight   = "light"   // 轻松
	WorkIntensityMedium  = "medium"  // 中等
	WorkIntensityHeavy   = "heavy"   // 繁重
	WorkIntensityExtreme = "extreme" // 极重
)

// === 结算币种常量 ===
const (
	CurrencyCNY = "CNY" // 人民币
	CurrencyUSD = "USD" // 美元
)

// Linggong 零工兼职主表（保持 linggongs 表名兼容已发布数据）
type Linggong struct {
	database.RegionBaseModel // 含 id/region_id/created_at/updated_at/deleted_at（地区隔离）

	// === 基础信息 ===
	Title       string     `gorm:"size:200;not null" json:"title"`                          // 标题
	Content     string     `gorm:"type:text" json:"content"`                                // 详情描述
	CoverImage  string     `gorm:"size:255" json:"cover_image"`                            // 封面图
	UserID      uint       `gorm:"index;not null" json:"user_id"`                           // 发布者 ID（雇主）
	UserName    string     `gorm:"size:50" json:"user_name"`                                // 发布者昵称
	UserPhone   string     `gorm:"size:20" json:"user_phone"`                               // 发布者手机
	UserAvatar  string     `gorm:"size:255" json:"user_avatar"`                             // 发布者头像
	Status      int        `gorm:"default:0;index" json:"status"`                           // 0草稿 1已发布 2下架 3过期 4删除 5满员 6关闭 7完成
	AuditStatus int        `gorm:"default:0;index" json:"audit_status"`                     // 0待审 1通过 2拒绝
	AuditReason string     `gorm:"size:500" json:"audit_reason"`                            // 审核拒绝原因
	PublishedAt *time.Time `gorm:"index" json:"published_at"`                              // 发布时间

	// === 岗位类型/发布者类型 ===
	LinggongType string `gorm:"size:32;not null;default:'short_term';index" json:"linggong_type"` // short_term/long_term/task/hourly/daily/temp
	PublisherType string `gorm:"size:32;not null;default:'personal';index" json:"publisher_type"` // personal/company/agent/headhunter

	// === 雇主关联 ===
	EmployerID   uint   `gorm:"not null;default:0;index" json:"employer_id"`                // 雇主认证 ID
	CompanyName  string `gorm:"size:128;not null;default:'';index" json:"company_name"`     // 公司名称（冗余）
	ContactName  string `gorm:"size:50;not null;default:''" json:"contact_name"`            // 联系人
	ContactPhone string `gorm:"size:20;not null;default:''" json:"contact_phone"`           // 联系电话
	ContactWechat string `gorm:"size:64;not null;default:''" json:"contact_wechat"`         // 微信

	// === 计费方式/薪资 ===
	BillingType   string  `gorm:"size:32;not null;default:'by_day';index" json:"billing_type"` // by_piece/by_hour/by_day/by_week/by_month/fixed/negotiable
	SalaryMin     float64 `gorm:"type:decimal(12,2);default:0;index" json:"salary_min"`        // 薪资范围下限
	SalaryMax     float64 `gorm:"type:decimal(12,2);default:0;index" json:"salary_max"`        // 薪资范围上限
	SalaryUnit    string  `gorm:"size:32;not null;default:''" json:"salary_unit"`               // 薪资单位（元/件、元/时、元/日）
	SalaryNegotiable bool `gorm:"not null;default:false" json:"salary_negotiable"`             // 薪资面议
	Settlement    string  `gorm:"size:16;not null;default:'T+1';index" json:"settlement"`      // T+0/T+1/T+3/T+7/M+1/project
	Currency      string  `gorm:"size:16;not null;default:'CNY'" json:"currency"`              // 结算币种

	// === 工作时间 ===
	WorkStartDate *time.Time `gorm:"type:date;index" json:"work_start_date"`                // 工作开始日期
	WorkEndDate   *time.Time `gorm:"type:date;index" json:"work_end_date"`                  // 工作结束日期
	WorkDays      int        `gorm:"not null;default:0" json:"work_days"`                   // 工作天数
	WorkHours     int        `gorm:"not null;default:0" json:"work_hours"`                  // 每日工作时长（小时）
	WorkTimeStart string     `gorm:"size:16;not null;default:''" json:"work_time_start"`     // 上班时间 HH:mm
	WorkTimeEnd   string     `gorm:"size:16;not null;default:''" json:"work_time_end"`       // 下班时间 HH:mm
	WorkWeekdays  string     `gorm:"size:32;not null;default:'1,2,3,4,5'" json:"work_weekdays"` // 工作日（1-7 逗号分隔）
	WorkIntensity string     `gorm:"size:16;not null;default:'medium'" json:"work_intensity"` // light/medium/heavy/extreme

	// === 招募要求 ===
	RecruitCount        int    `gorm:"not null;default:1" json:"recruit_count"`              // 招募人数
	AppliedCount        int    `gorm:"not null;default:0" json:"applied_count"`              // 已报名数
	ConfirmedCount      int    `gorm:"not null;default:0" json:"confirmed_count"`            // 已确认数
	NeedGender          string `gorm:"size:16;not null;default:'any'" json:"need_gender"`    // any/male/female
	MinAge              int    `gorm:"not null;default:0" json:"min_age"`                    // 最低年龄
	MaxAge              int    `gorm:"not null;default:0" json:"max_age"`                    // 最高年龄
	Education           string `gorm:"size:32;not null;default:''" json:"education"`         // 学历要求
	Experience          string `gorm:"size:64;not null;default:''" json:"experience"`         // 经验要求
	NeedHealthCert       bool   `gorm:"not null;default:false" json:"need_health_cert"`       // 需健康证
	NeedIDCard           bool   `gorm:"not null;default:true" json:"need_id_card"`            // 需身份证
	MinCreditScore      int    `gorm:"not null;default:0" json:"min_credit_score"`           // 最低信用分要求

	// === 工作地点 ===
	Province        string  `gorm:"size:64;not null;default:'';index" json:"province"`      // 省
	City            string  `gorm:"size:64;not null;default:'';index" json:"city"`           // 市
	District        string  `gorm:"size:64;not null;default:'';index" json:"district"`      // 区/县
	BusinessDistrict string  `gorm:"size:128;not null;default:''" json:"business_district"`  // 商圈
	Address         string  `gorm:"size:500;not null;default:''" json:"address"`             // 详细地址
	Latitude        float64 `gorm:"type:decimal(10,7);default:0" json:"latitude"`            // 纬度
	Longitude       float64 `gorm:"type:decimal(10,7);default:0" json:"longitude"`           // 经度
	WorkLocationType string `gorm:"size:16;not null;default:'onsite'" json:"work_location_type"` // onsite 现场 / remote 远程 / hybrid 混合

	// === 任务制相关（按件计费） ===
	TaskID         uint    `gorm:"not null;default:0;index" json:"task_id"`                 // 关联任务包 ID
	TotalTaskCount int     `gorm:"not null;default:0" json:"total_task_count"`               // 任务总数（按件）
	ClaimedCount   int     `gorm:"not null;default:0" json:"claimed_count"`                 // 已领取数
	CompletedTaskCount int `gorm:"not null;default:0" json:"completed_task_count"`           // 已完成任务数

	// === 互动统计 ===
	ViewCount      int        `gorm:"default:0" json:"view_count"`      // 浏览数
	FavCount       int        `gorm:"default:0" json:"fav_count"`        // 收藏数
	ContactCount   int        `gorm:"default:0" json:"contact_count"`    // 联系数
	ShareCount     int        `gorm:"default:0" json:"share_count"`       // 分享数
	ApplicationCount int      `gorm:"default:0" json:"application_count"` // 报名数
	LastAppliedAt  *time.Time `gorm:"index" json:"last_applied_at"`      // 最近报名时间

	// === 风控 ===
	ContentHash string `gorm:"size:64;index" json:"content_hash"`   // 图文指纹
	RiskScore   int    `gorm:"default:0;index" json:"risk_score"`   // 风险评分 0-100

	// === 视频/图片 ===
	VideoURL       string `gorm:"size:255" json:"video_url"`        // 视频 URL
	VideoCover     string `gorm:"size:255" json:"video_cover"`      // 视频封面

	// === 配置/特征（JSONB）===
	Features        JSONB `gorm:"type:jsonb" json:"features"`        // 岗位特征 JSON
	Tags            JSONB `gorm:"type:jsonb" json:"tags"`            // 标签数组（最多 5 个）
	SkillTags       JSONB `gorm:"type:jsonb" json:"skill_tags"`      // 技能标签 ID 数组
	WelfareTags     JSONB `gorm:"type:jsonb" json:"welfare_tags"`    // 福利标签（包吃/包住/五险一金/年终奖）
	Images          JSONB `gorm:"type:jsonb" json:"images"`          // 工作环境图片
	Requirements    JSONB `gorm:"type:jsonb" json:"requirements"`     // 详细要求 JSON

	// === 运营字段 ===
	Featured       bool    `gorm:"default:false;index" json:"featured"`                       // 精选推荐
	Picked         bool    `gorm:"default:false;index" json:"picked"`                         // 运营甄选
	Verified       bool    `gorm:"default:false;index" json:"verified"`                       // 官方验真
	PromotionLevel int     `gorm:"default:0" json:"promotion_level"`                          // 推广等级 0-10
	TrafficWeight float64 `gorm:"type:decimal(3,2);default:1.00" json:"traffic_weight"`      // 流量权重 0.00-9.99

	// === 雇主认证状态 ===
	EmployerVerified   bool       `gorm:"default:false;index" json:"employer_verified"`     // 雇主已认证
	EmployerVerifiedAt *time.Time `gorm:"index" json:"employer_verified_at"`                // 雇主认证时间

	// Distance 仅在"附近"查询时由 SQL 计算并回填，非持久化字段（公里）
	Distance float64 `gorm:"-" json:"-"`
}

// TableName 表名（依据需求文档 7.2：主表 {module}s）
// 注意：保持现有 linggongs 表名不变以兼容已发布数据
func (Linggong) TableName() string { return "linggongs" }
