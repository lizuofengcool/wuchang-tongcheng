// Package utils 工具函数集合
// 提供通用的工具函数和错误码定义
package utils

// 错误码定义
// 规则：
// - 0 表示成功
// - 1000-1999 系统级错误
// - 2000-2999 用户相关错误
// - 3000-3999 业务模块错误（按模块分配）
// - 4000-4999 第三方服务错误
const (
	// 成功
	CodeSuccess = 0

	// ===== 系统级错误 1000-1999 =====
	// 系统错误
	CodeSystemError   = 1001
	CodeParamInvalid  = 1002
	CodeParamMissing  = 1003
	CodeUnauthorized  = 1004
	CodeForbidden     = 1005
	CodeNotFound      = 1006
	CodeAlreadyExists = 1007
	CodeTimeout       = 1008
	CodeTooManyRequests = 1009

	// 数据库错误
	CodeDBError        = 1101
	CodeDBQueryError   = 1102
	CodeDBInsertError  = 1103
	CodeDBUpdateError  = 1104
	CodeDBDeleteError  = 1105
	CodeDBRecordNotFound = 1106

	// 缓存错误
	CodeRedisError = 1201

	// 文件错误
	CodeFileError      = 1301
	CodeFileUploadError = 1302
	CodeFileNotFound   = 1303
	CodeFileTooLarge   = 1304
	CodeFileTypeInvalid = 1305
	CodeFilePresignError = 1306 // 当前存储后端不支持预签名直传
	CodeFileSTSError    = 1307 // STS 临时凭据不可用（未配置或第三方错误）

	// ===== 用户相关错误 2000-2999 =====
	CodeUserError          = 2001
	CodeUserNotFound       = 2002
	CodeUserAlreadyExists  = 2003
	CodeUserPasswordError  = 2004
	CodeUserDisabled       = 2005
	CodeUserNotLoggedIn    = 2006
	CodeUserTokenExpired   = 2007
	CodeUserTokenInvalid   = 2008
	CodeUserPermissionDenied = 2009

	// ===== 地区模块错误 2100-2199 =====
	CodeRegionError       = 2101
	CodeRegionNotFound    = 2102
	CodeRegionInvalid     = 2103

	// ===== 权限模块错误 2200-2299 =====
	CodePermissionError        = 2201
	CodePermissionDenied       = 2202
	CodeRoleNotFound           = 2203
	CodeRoleAlreadyExists      = 2204
	CodePermissionNotFound     = 2205
	CodePermissionAlreadyExists = 2206

	// ===== 分类信息模块错误 2300-2399 =====
	CodeCategoryError       = 2301
	CodeCategoryNotFound    = 2302
	CodeCategoryAlreadyExists = 2303

	// ===== 同城头条模块错误 2400-2499 =====
	CodeNewsError       = 2401
	CodeNewsNotFound    = 2402
	CodeNewsPublishError = 2403

	// ===== 商家模块错误 2500-2599 =====
	CodeShopError          = 2501
	CodeShopNotFound       = 2502
	CodeShopAlreadyExists  = 2503
	CodeShopAuditError     = 2504
	CodeShopNotApproved    = 2505
	CodeShopReviewError    = 2506
	CodeShopApplyError     = 2507

	// ===== 团购优惠券模块错误 2600-2699 =====
	CodeGroupBuyError         = 2601
	CodeGroupBuyNotFound      = 2602
	CodeGroupBuyExpired       = 2603
	CodeGroupBuySoldOut       = 2604
	CodeCouponError           = 2605
	CodeCouponNotFound        = 2606
	CodeCouponAlreadyUsed     = 2607
	CodeCouponExpired         = 2608
	CodeOrderError            = 2609
	CodeOrderNotFound         = 2610
	CodeOrderAlreadyCancelled = 2611

	// ===== 第三方服务错误 4000-4999 =====
	CodeThirdPartyError = 4001
	CodeMapAPIError     = 4002
	CodeStorageError    = 4003
	CodeSMSError        = 4004
	CodeWeChatError     = 4005
	CodeOAuthError      = 4006 // 第三方登录错误（OAuth 流程：code 换取身份/未配置/网络失败）
)

// 错误消息映射
var codeMessages = map[int]string{
	CodeSuccess: "success",

	// 系统错误
	CodeSystemError:    "系统错误",
	CodeParamInvalid:   "参数无效",
	CodeParamMissing:   "参数缺失",
	CodeUnauthorized:   "未授权访问",
	CodeForbidden:      "禁止访问",
	CodeNotFound:       "资源不存在",
	CodeAlreadyExists:  "资源已存在",
	CodeTimeout:        "请求超时",
	CodeTooManyRequests: "请求过于频繁",

	// 数据库错误
	CodeDBError:          "数据库错误",
	CodeDBQueryError:     "数据库查询错误",
	CodeDBInsertError:    "数据库插入错误",
	CodeDBUpdateError:    "数据库更新错误",
	CodeDBDeleteError:    "数据库删除错误",
	CodeDBRecordNotFound: "记录不存在",

	// 缓存错误
	CodeRedisError: "Redis缓存错误",

	// 文件错误
	CodeFileError:       "文件错误",
	CodeFileUploadError: "文件上传错误",
	CodeFileNotFound:    "文件不存在",
	CodeFileTooLarge:    "文件过大",
	CodeFileTypeInvalid: "文件类型不支持",
	CodeFilePresignError: "当前存储不支持预签名直传",
	CodeFileSTSError:    "STS 临时凭据不可用，请配置 sts 或使用普通上传",

	// 用户相关错误
	CodeUserError:            "用户错误",
	CodeUserNotFound:         "用户不存在",
	CodeUserAlreadyExists:    "用户已存在",
	CodeUserPasswordError:    "密码错误",
	CodeUserDisabled:         "用户已禁用",
	CodeUserNotLoggedIn:      "用户未登录",
	CodeUserTokenExpired:     "Token已过期",
	CodeUserTokenInvalid:     "Token无效",
	CodeUserPermissionDenied: "用户权限不足",

	// 地区模块错误
	CodeRegionError:    "地区错误",
	CodeRegionNotFound: "地区不存在",
	CodeRegionInvalid:  "地区无效",

	// 权限模块错误
	CodePermissionError:        "权限错误",
	CodePermissionDenied:       "权限不足",
	CodeRoleNotFound:           "角色不存在",
	CodeRoleAlreadyExists:      "角色已存在",
	CodePermissionNotFound:     "权限不存在",
	CodePermissionAlreadyExists: "权限已存在",

	// 分类信息模块错误
	CodeCategoryError:       "分类错误",
	CodeCategoryNotFound:    "分类不存在",
	CodeCategoryAlreadyExists: "分类已存在",

	// 同城头条模块错误
	CodeNewsError:       "头条错误",
	CodeNewsNotFound:    "头条不存在",
	CodeNewsPublishError: "头条发布错误",

	// 商家模块错误
	CodeShopError:         "商家错误",
	CodeShopNotFound:      "商家不存在",
	CodeShopAlreadyExists: "商家已存在",
	CodeShopAuditError:    "商家审核错误",
	CodeShopNotApproved:   "商家未通过审核",
	CodeShopReviewError:   "商家评论错误",
	CodeShopApplyError:    "商家申请错误",

	// 团购优惠券模块错误
	CodeGroupBuyError:         "团购错误",
	CodeGroupBuyNotFound:      "团购不存在",
	CodeGroupBuyExpired:       "团购已过期",
	CodeGroupBuySoldOut:       "团购已售罄",
	CodeCouponError:           "优惠券错误",
	CodeCouponNotFound:        "优惠券不存在",
	CodeCouponAlreadyUsed:     "优惠券已使用",
	CodeCouponExpired:         "优惠券已过期",
	CodeOrderError:            "订单错误",
	CodeOrderNotFound:         "订单不存在",
	CodeOrderAlreadyCancelled: "订单已取消",

	// 第三方服务错误
	CodeThirdPartyError: "第三方服务错误",
	CodeMapAPIError:     "地图API错误",
	CodeStorageError:    "存储服务错误",
	CodeSMSError:        "短信服务错误",
	CodeWeChatError:     "微信服务错误",
	CodeOAuthError:      "第三方登录错误",
}

// GetMessage 获取错误码对应的消息
func GetMessage(code int) string {
	if msg, ok := codeMessages[code]; ok {
		return msg
	}
	return "未知错误"
}

// RegisterCode 注册自定义错误码
func RegisterCode(code int, message string) {
	codeMessages[code] = message
}
