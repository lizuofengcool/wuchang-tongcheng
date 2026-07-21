// Package service 营销活动中台业务错误定义
package service

import "errors"

// ===== ad 子域 =====
var (
	ErrAdNotFound      = errors.New("广告位不存在")
	ErrAdStatusInvalid = errors.New("广告位状态不允许此操作")
)

// ===== coupon 子域 =====
var (
	ErrCouponNotFound      = errors.New("优惠券不存在")
	ErrCouponStatusInvalid = errors.New("优惠券状态不允许此操作")
	ErrCouponSoldOut       = errors.New("优惠券已抢完")
	ErrCouponExpired       = errors.New("优惠券已过期")
	ErrCouponNotStarted    = errors.New("优惠券尚未开始领取")
	ErrCouponAlreadyRecv   = errors.New("已领取过该优惠券")
	ErrUserCouponNotFound  = errors.New("用户优惠券不存在")
	ErrUserCouponUsed      = errors.New("用户优惠券已使用")
	ErrUserCouponExpired   = errors.New("用户优惠券已过期")
)

// ===== sign 子域 =====
var (
	ErrSignAlreadyToday = errors.New("今日已签到")
	ErrSignRuleNotFound = errors.New("签到规则不存在")
	ErrSignRuleExists   = errors.New("该天数的签到规则已存在")
)

// ===== activity 子域 =====
var (
	ErrActivityNotFound      = errors.New("营销活动不存在")
	ErrActivityStatusInvalid = errors.New("营销活动状态不允许此操作")
	ErrActivityNotOngoing    = errors.New("活动未在进行中")
)
