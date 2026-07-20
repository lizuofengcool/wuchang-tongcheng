// Package model 双向评价表（对标美团/斗米）
// 工人评价雇主 + 雇主评价工人 + 评价回复/追评
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 评价状态常量 ===
const (
	RatingStatusPending  = 0 // 待审核
	RatingStatusApproved = 1 // 已通过
	RatingStatusRejected = 2 // 已拒绝
	RatingStatusHidden   = 3 // 已隐藏
)

// === 评价者类型常量 ===
const (
	RaterTypeEmployer = "employer" // 雇主评价工人
	RaterTypeWorker   = "worker"   // 工人评价雇主
)

// === 评价目标类型常量 ===
const (
	RatingTargetTypeEmployer = "employer" // 雇主
	RatingTargetTypeWorker   = "worker"    // 求职者
	RatingTargetTypeLinggong = "linggong" // 岗位
	RatingTargetTypeTask     = "task"      // 任务
)

// === 推荐常量 ===
const (
	RecommendYes   = "yes"   // 推荐
	RecommendNo    = "no"     // 不推荐
	RecommendMaybe = "maybe" // 一般
)

// LinggongRating 双向评价表
type LinggongRating struct {
	database.RegionBaseModel // id/region_id/created_at/updated_at/deleted_at
	RatingNo        string     `gorm:"size:64;not null;uniqueIndex" json:"rating_no"`              // 评价单号
	LinggongID     uint       `gorm:"not null;index" json:"linggong_id"`                          // 岗位 ID
	TaskID         uint       `gorm:"not null;default:0;index" json:"task_id"`                   // 任务包 ID
	ApplicationID  uint       `gorm:"not null;default:0;index" json:"application_id"`             // 报名记录 ID
	ContractID     uint       `gorm:"not null;default:0;index" json:"contract_id"`               // 合同 ID
	PaymentID      uint       `gorm:"not null;default:0;index" json:"payment_id"`                 // 支付单 ID
	RaterType      string     `gorm:"size:16;not null;default:'employer';index" json:"rater_type"` // employer/worker
	RaterID        uint       `gorm:"not null;index" json:"rater_id"`                              // 评价人 ID
	RaterName      string     `gorm:"size:50;not null;default:''" json:"rater_name"`               // 评价人姓名
	RaterAvatar    string     `gorm:"size:255;not null;default:''" json:"rater_avatar"`            // 评价人头像
	TargetType     string     `gorm:"size:32;not null;default:'worker';index" json:"target_type"` // employer/worker/linggong/task
	TargetID       uint       `gorm:"not null;index" json:"target_id"`                             // 目标 ID
	TargetName    string     `gorm:"size:128;not null;default:''" json:"target_name"`               // 目标名称
	Rating         int        `gorm:"not null;default:5;index" json:"rating"`                     // 总体评分 1-5
	Content        string     `gorm:"type:text" json:"content"`                                    // 评价内容
	Images         JSONB      `gorm:"type:jsonb" json:"images"`                                   // 评价图片
	VideoURL       string     `gorm:"size:255;not null;default:''" json:"video_url"`               // 评价视频
	IsAnonymous   bool       `gorm:"not null;default:false" json:"is_anonymous"`                  // 匿名评价
	IsRecommended  string     `gorm:"size:16;not null;default:'yes';index" json:"is_recommended"`  // yes/no/maybe
	Tags           JSONB      `gorm:"type:jsonb" json:"tags"`                                    // 标签
	DealAmount     float64    `gorm:"type:decimal(12,2);default:0" json:"deal_amount"`             // 成交金额
	// 多维度评分
	WorkQuality       int      `gorm:"not null;default:5" json:"work_quality"`                      // 工作质量
	Punctuality       int      `gorm:"not null;default:5" json:"punctuality"`                       // 守时性
	Communication     int      `gorm:"not null;default:5" json:"communication"`                    // 沟通能力
	Attitude          int      `gorm:"not null;default:5" json:"attitude"`                          // 工作态度
	Professionalism   int      `gorm:"not null;default:5" json:"professionalism"`                  // 专业能力
	PaymentTimeliness int      `gorm:"not null;default:5" json:"payment_timeliness"`               // 付款及时性
	WorkEnvironment   int      `gorm:"not null;default:5" json:"work_environment"`                 // 工作环境
	SalaryMatch       int      `gorm:"not null;default:5" json:"salary_match"`                     // 薪资匹配度
	// 回复
	Reply           string     `gorm:"type:text" json:"reply"`                                    // 回复内容
	ReplyAt         *time.Time `gorm:"index" json:"reply_at"`                                    // 回复时间
	AppendContent   string     `gorm:"type:text" json:"append_content"`                          // 追评内容
	AppendImages    JSONB      `gorm:"type:jsonb" json:"append_images"`                          // 追评图片
	AppendAt        *time.Time `gorm:"index" json:"append_at"`                                    // 追评时间
	LikeCount       int        `gorm:"not null;default:0" json:"like_count"`                      // 点赞数
	Status          int        `gorm:"default:1;index" json:"status"`                             // 0待审 1通过 2拒绝 3隐藏
	RejectedReason  string     `gorm:"size:500;not null;default:''" json:"rejected_reason"`         // 拒绝原因
	EvaluatedAt    *time.Time `gorm:"index" json:"evaluated_at"`                                 // 评价时间（业务字段）
}

// TableName 表名（linggong_ 前缀）
func (LinggongRating) TableName() string { return "linggong_ratings" }
