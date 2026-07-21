// Package handler 营销活动中台错误码常量定义
// 由于 error_code.go 不允许修改，本模块错误码在 plugin.go 的 init() 中注册到 utils
// 错误码区间：5801-5830（marketing 模块独占）
package handler

// ===== marketing 错误码常量（5801-5830） =====
const (
	// ad 子域 5801-5805
	CodeMarketingAdError          = 5801 // 广告位通用错误
	CodeMarketingAdNotFound       = 5802 // 广告位不存在
	CodeMarketingAdStatusInvalid  = 5803 // 广告位状态不允许此操作

	// coupon 子域 5806-5815
	CodeMarketingCouponError          = 5806 // 优惠券通用错误
	CodeMarketingCouponNotFound       = 5807 // 优惠券不存在
	CodeMarketingCouponStatusInvalid  = 5808 // 优惠券状态不允许此操作
	CodeMarketingCouponSoldOut        = 5809 // 优惠券已抢完
	CodeMarketingCouponExpired        = 5810 // 优惠券已过期
	CodeMarketingCouponNotStarted     = 5811 // 优惠券尚未开始领取
	CodeMarketingCouponAlreadyRecv    = 5812 // 已领取过该优惠券
	CodeMarketingUserCouponNotFound   = 5813 // 用户优惠券不存在
	CodeMarketingUserCouponUsed       = 5814 // 用户优惠券已使用
	CodeMarketingUserCouponExpired    = 5815 // 用户优惠券已过期

	// sign 子域 5816-5820
	CodeMarketingSignError         = 5816 // 签到通用错误
	CodeMarketingSignRuleError     = 5817 // 签到规则错误
	CodeMarketingSignRuleNotFound  = 5818 // 签到规则不存在

	// activity 子域 5819-5823
	CodeMarketingActivityError         = 5819 // 营销活动通用错误
	CodeMarketingActivityNotFound      = 5820 // 营销活动不存在
	CodeMarketingActivityStatusInvalid = 5821 // 营销活动状态不允许此操作
	CodeMarketingActivityNotOngoing    = 5822 // 活动未在进行中
)
