// Package dto 同城零工兼职数据传输对象 - 信用分 + 审核规则 + 统计 + 收藏 + 推荐
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// CreditInfo 信用分变更详情响应
type CreditInfo struct {
	ID            uint      `json:"id"`
	UserID        uint      `json:"user_id"`
	UserType      string    `json:"user_type"`
	UserTypeText  string    `json:"user_type_text"`
	Reason        string    `json:"reason"`
	ReasonText    string    `json:"reason_text"`
	ChangeType    string    `json:"change_type"`
	ChangeTypeText string   `json:"change_type_text"`
	ChangeScore   int       `json:"change_score"`
	BeforeScore   int       `json:"before_score"`
	AfterScore    int       `json:"after_score"`
	LinggongID    uint      `json:"linggong_id"`
	TaskID        uint      `json:"task_id"`
	ApplicationID uint      `json:"application_id"`
	RatingID      uint      `json:"rating_id"`
	OperatorID    uint      `json:"operator_id"`
	OperatorName  string    `json:"operator_name"`
	Description   string    `json:"description"`
	EvidenceURL   string    `json:"evidence_url"`
	RegionID      uint      `json:"region_id"`
	CreatedAt     time.Time `json:"created_at"`
}

// CreditAdjustRequest 信用分调整请求
type CreditAdjustRequest struct {
	UserID        uint    `json:"user_id" binding:"required"`
	UserType      string `json:"user_type" binding:"omitempty,oneof=worker employer"`
	Reason        string `json:"reason" binding:"omitempty,oneof=fulfill breach late absent no_show good_rating bad_rating verified skill_cert report appeal manual invite_friend daily_login complete_profile"`
	ChangeType    string `json:"change_type" binding:"omitempty,oneof=add deduct reset"`
	ChangeScore   int     `json:"change_score"`
	LinggongID    uint    `json:"linggong_id"`
	TaskID        uint    `json:"task_id"`
	ApplicationID uint    `json:"application_id"`
	RatingID      uint    `json:"rating_id"`
	Description   string  `json:"description"`
	EvidenceURL   string  `json:"evidence_url" binding:"max=255"`
}

// CreditListRequest 信用分变更记录列表请求
type CreditListRequest struct {
	UserID     uint   `form:"user_id" json:"user_id"`
	UserType   string `form:"user_type" json:"user_type"`
	Reason     string `form:"reason" json:"reason"`
	ChangeType string `form:"change_type" json:"change_type"`
	utils.Pagination
}

// CreditScoreResponse 信用分查询响应
type CreditScoreResponse struct {
	UserID      uint   `json:"user_id"`
	UserType    string `json:"user_type"`
	CreditScore int    `json:"credit_score"`
	Level       int    `json:"level"`
}

// AuditRuleInfo 审核规则信息
type AuditRuleInfo struct {
	ID          uint      `json:"id"`
	RuleName    string    `json:"rule_name"`
	RuleType    string    `json:"rule_type"`
	RuleTypeText string    `json:"rule_type_text"`
	RuleKey     string    `json:"rule_key"`
	Pattern     string    `json:"pattern"`
	Threshold   interface{} `json:"threshold"`
	Action      string    `json:"action"`
	ActionText  string    `json:"action_text"`
	PenaltyType string    `json:"penalty_type"`
	Severity    int       `json:"severity"`
	Status      int       `json:"status"`
	StatusText  string    `json:"status_text"`
	Description string    `json:"description"`
	Sort        int       `json:"sort"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateAuditRuleRequest 创建审核规则请求
type CreateAuditRuleRequest struct {
	RuleName    string      `json:"rule_name" binding:"required,max=128"`
	RuleType    string      `json:"rule_type" binding:"omitempty,oneof=sensitive_word salary_check frequency fake_job contact prohibited duplicate blacklist"`
	RuleKey     string      `json:"rule_key" binding:"max=64"`
	Pattern     string      `json:"pattern"`
	Threshold   interface{} `json:"threshold"`
	Action      string      `json:"action" binding:"omitempty,oneof=reject approval filter limit"`
	PenaltyType string      `json:"penalty_type" binding:"max=32"`
	Severity    int         `json:"severity" binding:"min=1,max=5"`
	Status      int         `json:"status" binding:"oneof=0 1"`
	Description string      `json:"description" binding:"max=500"`
	Sort        int         `json:"sort"`
}

// UpdateAuditRuleRequest 更新审核规则请求
type UpdateAuditRuleRequest struct {
	RuleName    *string `json:"rule_name" binding:"omitempty,max=128"`
	RuleType    *string `json:"rule_type" binding:"omitempty,oneof=sensitive_word salary_check frequency fake_job contact prohibited duplicate blacklist"`
	RuleKey     *string `json:"rule_key" binding:"omitempty,max=64"`
	Pattern     *string `json:"pattern"`
	Threshold   interface{} `json:"threshold"`
	Action      *string `json:"action" binding:"omitempty,oneof=reject approval filter limit"`
	PenaltyType *string `json:"penalty_type" binding:"omitempty,max=32"`
	Severity    *int    `json:"severity" binding:"omitempty,min=1,max=5"`
	Status      *int    `json:"status" binding:"omitempty,oneof=0 1"`
	Description *string `json:"description" binding:"omitempty,max=500"`
	Sort        *int    `json:"sort"`
}

// AuditRuleListRequest 审核规则列表请求
type AuditRuleListRequest struct {
	RuleType string `form:"rule_type" json:"rule_type"`
	Action   string `form:"action" json:"action"`
	Status   *int   `form:"status" json:"status"`
	Keyword  string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// StatisticInfo 统计信息
type StatisticInfo struct {
	ID                uint      `json:"id"`
	StatDate          time.Time `json:"stat_date"`
	StatType          string    `json:"stat_type"`
	StatTypeText      string    `json:"stat_type_text"`
	TargetID          uint      `json:"target_id"`
	TargetName        string    `json:"target_name"`
	ImpressionCount   int       `json:"impression_count"`
	ClickCount        int       `json:"click_count"`
	FavCount          int       `json:"fav_count"`
	ContactCount      int       `json:"contact_count"`
	ApplicationCount  int       `json:"application_count"`
	HiredCount        int       `json:"hired_count"`
	CompletedCount    int       `json:"completed_count"`
	DealCount         int       `json:"deal_count"`
	ConversionRate    float64   `json:"conversion_rate"`
	TotalSalary       float64   `json:"total_salary"`
	AvgSalary         float64   `json:"avg_salary"`
	AvgDealDays       int       `json:"avg_deal_days"`
	RegionID          uint      `json:"region_id"`
}

// StatListRequest 统计列表请求
type StatListRequest struct {
	StatType  string `form:"stat_type" json:"stat_type"`
	TargetID  uint   `form:"target_id" json:"target_id"`
	StartDate string `form:"start_date" json:"start_date"`
	EndDate   string `form:"end_date" json:"end_date"`
	Keyword   string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// StatisticUpsertRequest 统计 upsert 请求（M 端维护）
type StatisticUpsertRequest struct {
	StatDate         time.Time `json:"stat_date" binding:"required"`
	StatType         string    `json:"stat_type" binding:"omitempty,oneof=linggong employer worker skill region platform task category"`
	TargetID         uint      `json:"target_id"`
	TargetName       string    `json:"target_name" binding:"max=128"`
	ImpressionCount  int       `json:"impression_count"`
	ClickCount       int       `json:"click_count"`
	FavCount         int       `json:"fav_count"`
	ContactCount     int       `json:"contact_count"`
	ApplicationCount int       `json:"application_count"`
	HiredCount       int       `json:"hired_count"`
	CompletedCount   int       `json:"completed_count"`
	DealCount        int       `json:"deal_count"`
	ConversionRate   float64   `json:"conversion_rate"`
	TotalSalary      float64   `json:"total_salary"`
	AvgSalary        float64   `json:"avg_salary"`
	AvgDealDays      int       `json:"avg_deal_days"`
}

// StatOverviewResponse 统计总览响应
type StatOverviewResponse struct {
	ImpressionCount  int64   `json:"impression_count"`
	ClickCount       int64   `json:"click_count"`
	FavCount         int64   `json:"fav_count"`
	ContactCount     int64   `json:"contact_count"`
	ApplicationCount int64   `json:"application_count"`
	HiredCount       int64   `json:"hired_count"`
	CompletedCount   int64   `json:"completed_count"`
	DealCount        int64   `json:"deal_count"`
	TotalSalary      float64 `json:"total_salary"`
	AvgSalary        float64 `json:"avg_salary"`
	ConversionRate   float64 `json:"conversion_rate"`
}

// FavoriteInfo 收藏信息
type FavoriteInfo struct {
	ID            uint      `json:"id"`
	UserID        uint      `json:"user_id"`
	TargetID      uint      `json:"target_id"`
	FavoriteType  string    `json:"favorite_type"`
	FavoriteTypeText string `json:"favorite_type_text"`
	Remark        string    `json:"remark"`
	NotifyOnUpdate bool     `json:"notify_on_update"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// CreateFavoriteRequest 创建收藏请求
type CreateFavoriteRequest struct {
	TargetID      uint   `json:"target_id" binding:"required"`
	FavoriteType  string `json:"favorite_type" binding:"omitempty,oneof=linggong worker employer task search"`
	Remark        string `json:"remark" binding:"max=200"`
	NotifyOnUpdate bool  `json:"notify_on_update"`
}

// FavoriteListRequest 收藏列表请求
type FavoriteListRequest struct {
	FavoriteType string `form:"favorite_type" json:"favorite_type"`
	utils.Pagination
}

// RecommendationInfo 推荐信息
type RecommendationInfo struct {
	ID            uint      `json:"id"`
	UserID        uint      `json:"user_id"`
	LinggongID    uint      `json:"linggong_id"`
	RecType       string    `json:"rec_type"`
	RecTypeText   string    `json:"rec_type_text"`
	Source        string    `json:"source"`
	Score         float64   `json:"score"`
	Reason        string    `json:"reason"`
	SalaryMatch   float64   `json:"salary_match"`
	SkillMatch    float64   `json:"skill_match"`
	LocationMatch float64   `json:"location_match"`
	TimeMatch     float64   `json:"time_match"`
	CreditMatch   float64   `json:"credit_match"`
	Status        int       `json:"status"`
	StatusText    string    `json:"status_text"`
	ClickedAt     *time.Time `json:"clicked_at"`
	AppliedAt     *time.Time `json:"applied_at"`
	ViewedAt      *time.Time `json:"viewed_at"`
	DismissedAt   *time.Time `json:"dismissed_at"`
	ExpiredAt     *time.Time `json:"expired_at"`
	RegionID      uint      `json:"region_id"`
	CreatedAt     time.Time `json:"created_at"`
}

// RecommendationListRequest 推荐列表请求
type RecommendationListRequest struct {
	RecType string `form:"rec_type" json:"rec_type"`
	Status  *int   `form:"status" json:"status"`
	utils.Pagination
}
