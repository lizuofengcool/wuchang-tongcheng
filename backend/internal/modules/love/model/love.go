// Package model 同城相亲交友数据模型 - 主表 Loves
// 依据 v3.2.1 架构方案第六章：对标 Soul / 陌陌 / 探探 / 百合网
// 依据需求文档 1.10：4 维数据隔离（region_id 地区隔离 + user_id 用户隔离）
// 依据需求文档 7.1：通用字段 id/region_id/created_at/updated_at/deleted_at + status + audit_status
// 依据需求文档 7.2：主表 loves
//
// 关键差异化能力：
//   - 灵魂匹配算法：interests/personality/values 三个 JSONB 字段
//   - 语音匹配：voice_intro_url
//   - 视频认证：video_verified + verifications 表 type=video
//   - 会员等级：基础/高级/VIP/Premium 四级（member_level 字段）
//   - 心动信号：super_like 字段每天限量 1 次（super_likes_today/super_likes_reset_at）
//   - 隐私保护：hide_online/hide_location/hide_age/hide_distance
//   - 印象标签：impression 表（他人评价）
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 状态常量 ===
const (
	LoveStatusDisabled = 0 // 禁用
	LoveStatusActive   = 1 // 正常
	LoveStatusFrozen   = 2 // 冻结
	LoveStatusCanceled = 3 // 注销
)

// === 审核状态常量 ===
const (
	LoveAuditPending  = 0 // 待审
	LoveAuditApproved = 1 // 通过
	LoveAuditRejected = 2 // 拒绝
)

// === 性别常量 ===
const (
	GenderUnknown = 0 // 未知
	GenderMale    = 1 // 男
	GenderFemale  = 2 // 女
)

// === 会员等级常量 ===
const (
	MemberLevelNone     = 0 // 普通用户
	MemberLevelBasic    = 1 // 基础会员
	MemberLevelAdvanced = 2 // 高级会员
	MemberLevelVIP      = 3 // VIP 会员
	MemberLevelPremium  = 4 // Premium 会员
)

// === 会员等级编码 ===
const (
	MemberLevelCodeBasic    = "basic"    // 基础
	MemberLevelCodeAdvanced = "advanced" // 高级
	MemberLevelCodeVIP      = "vip"      // VIP
	MemberLevelCodePremium  = "premium"  // Premium
)

// === 订阅计划 ===
const (
	MemberPlanMonthly   = "monthly"   // 月度订阅
	MemberPlanQuarterly = "quarterly" // 季度订阅
	MemberPlanYearly    = "yearly"    // 年度订阅
)

// === 认证类型常量 ===
const (
	VerifyTypeRealName  = "real_name"  // 实名认证
	VerifyTypePhoto     = "photo"      // 照片认证
	VerifyTypeVideo     = "video"      // 视频认证
	VerifyTypeEducation = "education"  // 学历认证
	VerifyTypeProperty  = "property"   // 房产认证
	VerifyTypeCar       = "car"        // 车辆认证
)

// === 认证状态常量 ===
const (
	VerifyStatusPending  = 0 // 待审
	VerifyStatusApproved = 1 // 通过
	VerifyStatusRejected = 2 // 拒绝
)

// === 喜欢/不喜欢/心动信号动作 ===
const (
	LikeActionLike    = "like"     // 喜欢
	LikeActionDislike = "dislike"  // 不喜欢
	LikeActionSkip    = "skip"     // 跳过
	LikeActionSuper   = "super"    // 心动信号
)

// === 匹配类型 ===
const (
	MatchTypeBothLike    = "both_like"    // 双向喜欢
	MatchTypeSuperLike   = "super_like"   // 心动信号
	MatchTypeRecommend   = "recommend"    // 推荐匹配
	MatchTypeCompatibility = "compatibility" // 灵魂匹配
)

// === 匹配状态 ===
const (
	MatchStatusActive    = 1 // 活跃
	MatchStatusUnmuted   = 2 // 解除匹配
	MatchStatusDissolved = 3 // 已解除
	MatchStatusBlocked   = 4 // 已拉黑
)

// === 动态媒体类型 ===
const (
	StoryMediaTypeImage = "image" // 图文
	StoryMediaTypeVideo = "video" // 视频
	StoryMediaTypeVoice = "voice" // 语音
)

// === 推荐类型 ===
const (
	RecTypeDaily    = "daily"    // 每日推荐
	RecTypeNearby   = "nearby"   // 附近推荐
	RecTypeSameCity = "same_city" // 同城推荐
	RecTypeSameHometown = "same_hometown" // 同乡推荐
	RecTypeInterest = "interest" // 兴趣推荐
	RecTypeSoulmate = "soulmate" // 灵魂匹配
	RecTypeNewUser  = "new_user" // 新人推荐
)

// === 推荐来源 ===
const (
	RecSourceAlgorithm = "algorithm" // 算法
	RecSourceManual    = "manual"    // 人工
	RecSourceAI        = "ai"        // AI
	RecSourceBoost     = "boost"     // 加权
)

// === 推荐状态 ===
const (
	RecStatusPending  = 0 // 待处理
	RecStatusViewed   = 1 // 已查看
	RecStatusLiked    = 2 // 已喜欢
	RecStatusDisliked = 3 // 已不喜欢
	RecStatusSkipped  = 4 // 已跳过
	RecStatusDismissed = 5 // 已忽略
	RecStatusExpired  = 6 // 已过期
)

// === 通知类型 ===
const (
	NotifyTypeLike        = "like"         // 喜欢
	NotifyTypeSuperLike   = "super_like"   // 心动信号
	NotifyTypeMatch       = "match"        // 匹配
	NotifyTypeVisit       = "visit"        // 访客
	NotifyTypeGift        = "gift"         // 礼物
	NotifyTypeMessage     = "message"      // 消息
	NotifyTypeImpression  = "impression"   // 印象标签
	NotifyTypeStoryLike   = "story_like"   // 动态点赞
	NotifyTypeStoryComment = "story_comment" // 动态评论
	NotifyTypeMember      = "member"       // 会员
	NotifyTypeVerification = "verification" // 认证
	NotifyTypeSystem      = "system"       // 系统
	NotifyTypeReport      = "report"       // 举报处理
)

// === 举报目标类型 ===
const (
	ReportTargetUser      = "user"      // 用户
	ReportTargetStory     = "story"     // 动态
	ReportTargetMessage   = "message"   // 消息
	ReportTargetImpression = "impression" // 印象
	ReportTargetGift      = "gift"      // 礼物
)

// === 举报原因类型 ===
const (
	ReportReasonFakeInfo      = "fake_info"      // 虚假资料
	ReportReasonFraud         = "fraud"          // 诈骗
	ReportReasonHarassment    = "harassment"     // 骚扰
	ReportReasonPorn          = "porn"           // 色情
	ReportReasonPolitical     = "political"      // 政治
	ReportReasonInsult        = "insult"         // 辱骂
	ReportReasonSpam          = "spam"           // 垃圾信息
	ReportReasonMinor         = "minor"          // 未成年人
	ReportReasonOther         = "other"          // 其他
)

// === 举报状态 ===
const (
	ReportStatusPending  = 0 // 待处理
	ReportStatusHandling = 1 // 处理中
	ReportStatusHandled  = 2 // 已处理
	ReportStatusRejected = 3 // 已驳回
)

// === 举报申诉状态 ===
const (
	AppealStatusNone     = 0 // 未申诉
	AppealStatusPending  = 1 // 申诉中
	AppealStatusApproved = 2 // 申诉成功
	AppealStatusRejected = 3 // 申诉驳回
)

// === 访客来源 ===
const (
	VisitSourceRecommend = "recommend" // 推荐
	VisitSourceNearby    = "nearby"    // 附近
	VisitSourceSearch    = "search"    // 搜索
	VisitSourceStory     = "story"     // 动态
	VisitSourceMatch     = "match"     // 匹配
	VisitSourceProfile   = "profile"   // 资料
	VisitSourceOther     = "other"     // 其他
)

// === 访客类型 ===
const (
	VisitTypeProfile = "profile" // 主页
	VisitTypeStory   = "story"   // 动态
	VisitTypePhoto   = "photo"   // 相册
)

// === 礼物分类 ===
const (
	GiftCategoryCommon   = "common"   // 普通
	GiftCategoryLuxury   = "luxury"   // 奢华
	GiftCategoryAnimated = "animated" // 动画
	GiftCategoryFestival = "festival" // 节日
	GiftCategoryLimited  = "limited"  // 限量
)

// === 审核规则类型 ===
const (
	AuditRuleTypeSensitiveWord  = "sensitive_word"  // 敏感词
	AuditRuleTypeProhibited     = "prohibited"      // 违禁内容
	AuditRuleTypeContact        = "contact"         // 联系方式
	AuditRuleTypeFakeInfo       = "fake_info"       // 虚假资料
	AuditRuleTypeFrequency      = "frequency"       // 频率限制
	AuditRuleTypePorn           = "porn"            // 色情
	AuditRuleTypePolitical      = "political"       // 政治
	AuditRuleTypeAgeCheck       = "age_check"       // 年龄校验
	AuditRuleTypePhotoCheck     = "photo_check"     // 照片校验
)

// === 审核动作 ===
const (
	AuditActionReject  = "reject"  // 拒绝
	AuditActionReview  = "review"  // 人工审核
	AuditActionBlock   = "block"   // 封禁
	AuditActionShadow  = "shadow"  // 影子封禁
	AuditActionWarning = "warning" // 警告
)

// === 频率限制类型 ===
const (
	LimitTypeLike        = "like"         // 每日喜欢
	LimitTypeSuperLike   = "super_like"   // 每日心动信号
	LimitTypeVisit       = "visit"        // 每日访客查看
	LimitTypeRecommend   = "recommend"    // 每日推荐
	LimitTypeStory       = "story"        // 每日动态
	LimitTypeGift        = "gift"         // 每日礼物
	LimitTypeMatch       = "match"        // 每日匹配
	LimitTypeMessage     = "message"      // 每日消息
)

// Love 相亲交友主表（保持 loves 表名）
type Love struct {
	database.RegionBaseModel // 含 id/region_id/created_at/updated_at/deleted_at（地区隔离）

	// === 用户基础信息 ===
	UserID      uint   `gorm:"index;not null" json:"user_id"`         // 用户 ID
	Nickname    string `gorm:"size:64;not null;default:''" json:"nickname"`         // 昵称
	Avatar      string `gorm:"size:255;not null;default:''" json:"avatar"`           // 头像
	Gender      int    `gorm:"not null;default:0;index" json:"gender"`              // 性别：0未知 1男 2女
	Age         int    `gorm:"not null;default:0;index" json:"age"`                 // 年龄
	Birthday    *time.Time `gorm:"type:date" json:"birthday"`                        // 生日
	Height      int    `gorm:"not null;default:0" json:"height"`                    // 身高（cm）
	Weight      int    `gorm:"not null;default:0" json:"weight"`                    // 体重（kg）
	Constellation string `gorm:"size:16;not null;default:''" json:"constellation"`  // 星座
	Zodiac      string `gorm:"size:16;not null;default:''" json:"zodiac"`           // 生肖
	Hometown    string `gorm:"size:128;not null;default:'';index" json:"hometown"`  // 家乡
	Residence   string `gorm:"size:128;not null;default:'';index" json:"residence"` // 居住地
	Education   string `gorm:"size:32;not null;default:''" json:"education"`        // 学历
	Occupation  string `gorm:"size:64;not null;default:''" json:"occupation"`       // 职业
	Income      string `gorm:"size:32;not null;default:''" json:"income"`           // 收入
	Marriage    string `gorm:"size:32;not null;default:''" json:"marriage"`         // 婚姻状况
	House       string `gorm:"size:32;not null;default:''" json:"house"`            // 房产
	Car         string `gorm:"size:32;not null;default:''" json:"car"`              // 车辆
	Drinking    string `gorm:"size:32;not null;default:''" json:"drinking"`         // 饮酒
	Smoking     string `gorm:"size:32;not null;default:''" json:"smoking"`          // 吸烟
	WantKids    string `gorm:"size:32;not null;default:''" json:"want_kids"`        // 想要孩子
	Bio         string `gorm:"type:text" json:"bio"`                                 // 个人简介
	VoiceIntroURL string `gorm:"size:255;not null;default:''" json:"voice_intro_url"` // 语音自我介绍 URL
	CoverImage  string `gorm:"size:255;not null;default:''" json:"cover_image"`     // 封面图

	// === 认证状态 ===
	PhotoVerified     bool `gorm:"not null;default:false;index" json:"photo_verified"`         // 照片认证
	VideoVerified     bool `gorm:"not null;default:false;index" json:"video_verified"`         // 视频认证
	EducationVerified bool `gorm:"not null;default:false" json:"education_verified"`           // 学历认证
	RealNameVerified  bool `gorm:"not null;default:false;index" json:"real_name_verified"`     // 实名认证

	// === 状态/审核 ===
	Status      int    `gorm:"not null;default:1;index" json:"status"`        // 状态：0禁用 1正常 2冻结 3注销
	AuditStatus int    `gorm:"not null;default:1;index" json:"audit_status"` // 审核：0待审 1通过 2拒绝
	AuditReason string `gorm:"size:500;not null;default:''" json:"audit_reason"` // 审核拒绝原因

	// === 会员 ===
	MemberLevel     int        `gorm:"not null;default:0;index" json:"member_level"` // 会员等级：0普通 1基础 2高级 3VIP 4Premium
	MemberExpiredAt *time.Time `gorm:"index" json:"member_expired_at"`               // 会员过期时间
	Credits         float64    `gorm:"type:decimal(12,2);not null;default:0" json:"credits"` // 金币余额

	// === 心动信号 ===
	SuperLikesToday   int        `gorm:"not null;default:0" json:"super_likes_today"`     // 今日心动信号已使用次数
	SuperLikesResetAt *time.Time `gorm:"index" json:"super_likes_reset_at"`               // 心动信号重置时间

	// === 活跃/位置 ===
	LastActiveAt     *time.Time `gorm:"index" json:"last_active_at"`             // 最近活跃时间
	LastActiveIP     string     `gorm:"size:64;not null;default:''" json:"last_active_ip"` // 最近活跃 IP
	Longitude        float64    `gorm:"type:decimal(10,7);not null;default:0" json:"longitude"` // 经度
	Latitude         float64    `gorm:"type:decimal(10,7);not null;default:0" json:"latitude"`  // 纬度
	LocationUpdatedAt *time.Time `gorm:"index" json:"location_updated_at"`        // 位置更新时间

	// === 隐私设置 ===
	HideOnline        bool `gorm:"not null;default:false" json:"hide_online"`         // 隐藏在线状态
	HideLocation      bool `gorm:"not null;default:false" json:"hide_location"`       // 隐藏位置
	HideAge           bool `gorm:"not null;default:false" json:"hide_age"`            // 隐藏年龄
	HideDistance      bool `gorm:"not null;default:false" json:"hide_distance"`       // 隐藏距离
	OnlyVerifiedMatch bool `gorm:"not null;default:false" json:"only_verified_match"` // 仅认证用户可匹配
	ContactPrice      float64 `gorm:"type:decimal(12,2);not null;default:0" json:"contact_price"` // 联系价格（金币）

	// === 互动统计 ===
	ViewCount       int     `gorm:"not null;default:0" json:"view_count"`        // 浏览数
	LikeCount       int     `gorm:"not null;default:0" json:"like_count"`        // 喜欢数（主动）
	LikedCount      int     `gorm:"not null;default:0" json:"liked_count"`       // 被喜欢数
	MatchCount      int     `gorm:"not null;default:0" json:"match_count"`       // 匹配数
	VisitorCount    int     `gorm:"not null;default:0" json:"visitor_count"`     // 访客数
	StoryCount      int     `gorm:"not null;default:0" json:"story_count"`       // 动态数
	GiftCount       int     `gorm:"not null;default:0" json:"gift_count"`        // 礼物数
	ImpressionCount int     `gorm:"not null;default:0" json:"impression_count"`  // 印象数
	PopularityScore float64 `gorm:"type:decimal(5,2);not null;default:0;index" json:"popularity_score"` // 人气分

	// === 运营 ===
	Featured bool `gorm:"not null;default:false;index" json:"featured"` // 精选
	Picked   bool `gorm:"not null;default:false;index" json:"picked"`   // 运营甄选

	// === 风控 ===
	ContentHash string `gorm:"size:64;not null;default:''" json:"content_hash"` // 内容指纹
	RiskScore   int    `gorm:"not null;default:0" json:"risk_score"`           // 风险评分

	// === JSONB 字段 ===
	Tags              JSONB `gorm:"type:jsonb" json:"tags"`                // 个人标签数组
	Interests         JSONB `gorm:"type:jsonb" json:"interests"`           // 兴趣标签
	Personality       JSONB `gorm:"type:jsonb" json:"personality"`         // 性格测试
	Values            JSONB `gorm:"type:jsonb" json:"values"`              // 价值观
	PhotoUrls         JSONB `gorm:"type:jsonb" json:"photo_urls"`          // 相册
	MatchPreferences  JSONB `gorm:"type:jsonb" json:"match_preferences"`   // 匹配偏好

	// Distance 仅在"附近"查询时由 SQL 计算并回填，非持久化字段（公里）
	Distance float64 `gorm:"-" json:"-"`
}

// TableName 表名（依据需求文档 7.2：主表 {module}s）
func (Love) TableName() string { return "loves" }
