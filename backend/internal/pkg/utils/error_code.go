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

	// ===== 同城二手物品模块错误 2700-2799 =====
	CodeErshouError        = 2701
	CodeErshouNotFound     = 2702
	CodeErshouPublishError = 2703
	CodeErshouAuditError   = 2704

	// ===== 同城招聘求职模块错误 2800-2899 =====
	CodeJobError            = 2801 // 招聘通用错误
	CodeJobNotFound         = 2802 // 职位不存在
	CodeJobPublishError     = 2803 // 职位发布失败
	CodeJobAuditError       = 2804 // 职位审核错误
	CodeJobNoPermission     = 2805 // 无权操作职位
	CodeResumeError         = 2806 // 简历错误
	CodeResumeNotFound      = 2807 // 简历不存在
	CodeApplicationError    = 2808 // 投递错误
	CodeApplicationNotFound = 2809 // 投递不存在
	CodeInterviewError      = 2810 // 面试错误
	CodeInterviewNotFound   = 2811 // 面试不存在
	CodeMessageError        = 2812 // 消息错误
	CodeReportError         = 2813 // 举报错误
	CodeReviewError         = 2814 // 评价错误
	CodeCompanyError        = 2815 // 公司错误
	CodeCompanyNotFound     = 2816 // 公司不存在
	CodeAuditRuleError      = 2817 // 审核规则错误

	// ===== 房产模块错误 2900-2999 =====
	CodeHouseError         = 2901
	CodeHouseNotFound      = 2902
	CodeHousePublishError  = 2903
	CodeHouseAuditError    = 2904
	CodeHouseNoPermission  = 2905
	CodeHouseReportError   = 2906
	CodeHouseReviewError   = 2907
	CodeHouseContractError = 2908
	CodeHouseContractNotFound = 2909
	CodeHouseViewingError      = 2910
	CodeHouseViewingNotFound   = 2911
	CodeHouseCommunityError    = 2912
	CodeHouseCommunityNotFound = 2913
	CodeHouseAgentError        = 2914
	CodeHouseAgentNotFound     = 2915
	CodeHouseAuditRuleError    = 2916

	// ===== 同城车辆买卖模块错误 3000-3099 =====
	CodeCarError            = 3001
	CodeCarNotFound         = 3002
	CodeCarPublishError     = 3003
	CodeCarAuditError       = 3004
	CodeCarNoPermission     = 3005
	CodeInspectionError     = 3006
	CodeInspectionNotFound  = 3007
	CodeEvaluationError     = 3008
	CodeEvaluationNotFound  = 3009
	CodeFinancingError      = 3010
	CodeFinancingNotFound   = 3011
	CodeInsuranceError      = 3012
	CodeInsuranceNotFound   = 3013
	CodeTestDriveError      = 3014
	CodeTestDriveNotFound   = 3015
	CodeTransferError       = 3016
	CodeTransferNotFound    = 3017
	CodeReportErrorCar      = 3018
	CodeReviewErrorCar      = 3019
	CodeAuditRuleErrorCar   = 3020

	// ===== 第三方服务错误 4000-4999 =====
	CodeThirdPartyError = 4001
	CodeMapAPIError     = 4002
	CodeStorageError    = 4003
	CodeSMSError        = 4004
	CodeWeChatError     = 4005
	CodeOAuthError      = 4006 // 第三方登录错误（OAuth 流程：code 换取身份/未配置/网络失败）

	// ===== pay 支付中台错误 4007-4020 =====
	// 注：4001-4006 已被第三方服务错误占用，pay 从 4007 开始
	CodePayError             = 4007 // 支付通用错误
	CodePayOrderNotFound     = 4008 // 支付订单不存在
	CodePayOrderClosed       = 4009 // 订单已关闭
	CodePayOrderPaid         = 4010 // 订单已支付
	CodePayRefundNotFound    = 4011 // 退款单不存在
	CodePayRefundExceed      = 4012 // 退款金额超过订单
	CodePayEscrowNotFound    = 4013 // 担保交易不存在
	CodePayEscrowNotFrozen   = 4014 // 担保交易非冻结状态
	CodePayInsufficientBalance = 4015 // 余额不足
	CodePayWithdrawNotFound  = 4016 // 提现单不存在
	CodePayWithdrawNotPending = 4017 // 提现单非待审核
	CodePaySettlementNotFound = 4018 // 结算单不存在
	CodePayAccountNotFound   = 4019 // 资金账户不存在
	CodePayChannelNotFound   = 4020 // 支付渠道不存在

	// ===== im 即时通讯中台错误 4101-4120 =====
	CodeIMError              = 4101 // IM 通用错误
	CodeIMSessionNotFound    = 4102 // 会话不存在
	CodeIMNotParticipant     = 4103 // 非会话参与者
	CodeIMMessageNotFound    = 4104 // 消息不存在
	CodeIMMessageRecalled    = 4105 // 消息已撤回
	CodeIMGroupNotFound      = 4106 // 群组不存在
	CodeIMGroupMemberExists  = 4107 // 已是群成员
	CodeIMGroupMemberNotFound = 4108 // 群成员不存在
	CodeIMNotGroupOwner       = 4109 // 非群主
	CodeIMPrivacyNotFound    = 4110 // 隐私号码不存在
	CodeIMPrivacyUnbound     = 4111 // 隐私号码已解绑
	CodeIMNotificationNotFound = 4112 // 通知不存在
	CodeIMUserSettingsNotFound = 4113 // 用户设置不存在
	CodeIMSessionExists      = 4114 // 会话已存在

	// ===== material 物料中台错误 4201-4220 =====
	CodeMaterialError           = 4201 // 物料通用错误
	CodeMaterialFileNotFound    = 4202 // 文件不存在
	CodeMaterialImageNotFound   = 4203 // 图片不存在
	CodeMaterialVideoNotFound   = 4204 // 视频不存在
	CodeMaterialFeatureNotFound = 4205 // 图片特征不存在
	CodeMaterialUploadFailed    = 4206 // 文件上传失败
	CodeMaterialUnsupportedType = 4207 // 不支持的文件类型
	CodeMaterialCategoryNotFound = 4208 // 分类不存在
	CodeMaterialTagNotFound      = 4209 // 标签不存在
	CodeMaterialOCRFail          = 4210 // OCR 识别失败
	CodeMaterialSearchFail       = 4211 // 搜索失败
	CodeMaterialHashExists       = 4212 // 文件哈希已存在

	// ===== risk 风控中台错误 4301-4320 =====
	CodeRiskError              = 4301 // 风控通用错误
	CodeRiskReportNotFound    = 4302 // 举报不存在
	CodeRiskWordNotFound      = 4303 // 敏感词不存在
	CodeRiskRuleNotFound      = 4304 // 审核规则不存在
	CodeRiskBlacklistNotFound = 4305 // 黑名单记录不存在
	CodeRiskScoreNotFound     = 4306 // 用户风险分不存在
	CodeRiskViolationNotFound = 4307 // 违规处罚不存在
	CodeRiskAlreadyBlacklist  = 4308 // 已在黑名单中
	CodeRiskUserBanned        = 4309 // 用户已被封禁
	CodeRiskContentRejected   = 4310 // 内容审核未通过
	CodeRiskAppealNotFound    = 4311 // 申诉记录不存在
	CodeRiskEvidenceNotFound  = 4312 // 证据不存在
	CodeRiskAuditLogNotFound  = 4313 // 审核日志不存在

	// ===== ai 智能中台错误 4401-4420 =====
	CodeAIError               = 4401 // AI 通用错误
	CodeAITaskNotFound        = 4402 // AI 任务不存在
	CodeAIModelNotFound      = 4403 // AI 模型不存在
	CodeAIModelDisabled      = 4404 // AI 模型已禁用
	CodeAIPromptNotFound    = 4405 // 提示词模板不存在
	CodeAIGenerationNotFound = 4406 // 生成记录不存在
	CodeAIUnsupportedType    = 4407 // 不支持的 AI 任务类型
	CodeAIChatNotFound      = 4408 // 对话会话不存在
	CodeAIChatMessageNotFound = 4409 // 对话消息不存在
	CodeAIRecommendationNotFound = 4410 // 推荐记录不存在
	CodeAITrainingNotFound  = 4411 // 训练数据不存在
	CodeAIModelConfigNotFound = 4412 // 模型配置不存在

	// ===== love 相亲交友模块错误 5001-5030 =====
	CodeLoveError               = 5001 // 相亲交友通用错误
	CodeLoveNotFound            = 5002 // 用户资料不存在
	CodeLoveAlreadyLiked        = 5003 // 已经喜欢过该用户
	CodeLoveBlocked             = 5004 // 已被对方拉黑
	CodeLoveMemberExpired       = 5005 // 会员已过期
	CodeLoveVerificationFailed  = 5006 // 实名认证失败
	CodeLoveMatchFailed         = 5007 // 匹配失败
	CodeLoveStoryNotFound       = 5008 // 动态不存在
	CodeLoveGiftNotFound        = 5009 // 礼物不存在
	CodeLoveInsufficientCredits = 5010 // 金币余额不足
	CodeLoveNoPermission        = 5011 // 无权操作
	CodeLoveAlreadyMatched      = 5012 // 已经匹配过
	CodeLoveSuperLikeLimited    = 5013 // 心动信号今日已用完
	CodeLoveProfileNotComplete  = 5014 // 资料未完善
	CodeLoveNotVerified         = 5015 // 未通过实名认证
	CodeLoveReportError         = 5016 // 举报错误
	CodeLoveReportNotFound      = 5017 // 举报记录不存在
	CodeLoveAuditRuleError      = 5018 // 审核规则错误
	CodeLoveAuditRuleNotFound   = 5019 // 审核规则不存在
	CodeLoveMemberNotFound      = 5020 // 会员订阅不存在
	CodeLoveVisitNotFound       = 5021 // 访客记录不存在
	CodeLoveBlockNotFound       = 5022 // 拉黑记录不存在
	CodeLoveNotificationError   = 5023 // 通知错误
	CodeLoveNotificationNotFound = 5024 // 通知不存在
	CodeLoveRecommendationNotFound = 5025 // 推荐记录不存在
	CodeLoveImpressionError     = 5026 // 印象标签错误
	CodeLoveChatSessionNotFound = 5027 // 聊天会话不存在
	CodeLoveVerificationNotFound = 5028 // 认证记录不存在
	CodeLoveMemberLevelNotFound = 5029 // 会员等级不存在
	CodeLovePrivacyError        = 5030 // 隐私设置错误

	// ===== pinche 拼车出行模块错误 5101-5130 =====
	CodePincheError              = 5101 // 拼车通用错误
	CodePincheNotFound           = 5102 // 拼车行程不存在
	CodePincheRouteNotFound      = 5103 // 路线不存在
	CodePincheBookingConflict    = 5104 // 预订冲突
	CodePincheDriverNotVerified  = 5105 // 车主未认证
	CodePincheSeatUnavailable    = 5106 // 座位已满
	CodePincheTripOngoing        = 5107 // 行程进行中
	CodePincheInsuranceFailed    = 5108 // 保险投保失败
	CodePinchePaymentFailed      = 5109 // 支付失败
	CodePincheDriverNotFound     = 5110 // 车主不存在
	CodePincheVehicleNotFound    = 5111 // 车辆不存在
	CodePincheBookingNotFound    = 5112 // 预订不存在
	CodePincheTripNotFound       = 5113 // 行程记录不存在
	CodePincheRatingNotFound     = 5114 // 评价不存在
	CodePincheEmergencyNotFound  = 5115 // 紧急联系人不存在
	CodePinchePaymentNotFound    = 5116 // 支付记录不存在
	CodePincheRefundNotFound     = 5117 // 退款记录不存在
	CodePincheComplaintNotFound  = 5118 // 投诉不存在
	CodePincheMessageNotFound    = 5119 // 消息不存在
	CodePincheNoPermission       = 5120 // 无权操作
	CodePincheStatusInvalid      = 5121 // 状态不允许此操作
	CodePincheAuditError         = 5122 // 审核错误
	CodePincheAuditRuleError     = 5123 // 审核规则错误
	CodePincheRouteFavoriteError = 5124 // 常用路线错误
	CodePincheLocationError      = 5125 // 实时位置错误
	CodePincheCancelError        = 5126 // 取消错误
	CodePincheRefundError        = 5127 // 退款错误
	CodePincheAlreadyBooked      = 5128 // 已预订过该行程
	CodePincheDriverBusy         = 5129 // 车主行程冲突
	CodePincheTripNotStarted     = 5130 // 行程尚未开始

	// ===== linggong 零工兼职模块错误 5201-5230 =====
	CodeLinggongError                = 5201 // 零工兼职通用错误
	CodeLinggongNotFound             = 5202 // 零工兼职岗位不存在
	CodeLinggongPublishError         = 5203 // 零工兼职发布错误
	CodeLinggongAuditError           = 5204 // 零工兼职审核错误
	CodeLinggongNoPermission         = 5205 // 无权操作零工兼职
	CodeLinggongStatusInvalid        = 5206 // 状态不允许此操作
	CodeLinggongApplicationNotFound  = 5207 // 报名申请不存在
	CodeLinggongApplicationDuplicate = 5208 // 重复报名
	CodeLinggongApplicationFull      = 5209 // 报名名额已满
	CodeLinggongTaskNotFound         = 5210 // 任务不存在
	CodeLinggongTaskClaimError       = 5211 // 任务认领错误
	CodeLinggongEmployerNotFound     = 5212 // 雇主不存在
	CodeLinggongWorkerNotFound       = 5213 // 零工不存在
	CodeLinggongContractNotFound     = 5214 // 合同不存在
	CodeLinggongContractSignError    = 5215 // 合同签署错误
	CodeLinggongPaymentNotFound      = 5216 // 支付记录不存在
	CodeLinggongPaymentError         = 5217 // 支付错误
	CodeLinggongWithdrawalNotFound   = 5218 // 提现记录不存在
	CodeLinggongWithdrawalError      = 5219 // 提现错误
	CodeLinggongRatingNotFound       = 5220 // 评价不存在
	CodeLinggongRatingDuplicate      = 5221 // 重复评价
	CodeLinggongCertificationNotFound = 5222 // 资质证书不存在
	CodeLinggongCertificationError   = 5223 // 资质证书错误
	CodeLinggongCreditNotFound       = 5224 // 信用记录不存在
	CodeLinggongDisputeNotFound      = 5225 // 纠纷不存在
	CodeLinggongDisputeError         = 5226 // 纠纷处理错误
	CodeLinggongSkillNotFound        = 5227 // 技能不存在
	CodeLinggongAuditRuleNotFound    = 5228 // 审核规则不存在
	CodeLinggongAttendanceError      = 5229 // 考勤打卡错误
	CodeLinggongStatisticError       = 5230 // 统计错误

	// ===== dh114 同城114模块错误 5301-5330 =====
	CodeDh114Error              = 5301 // 同城114通用错误
	CodeDh114NotFound           = 5302 // 商户不存在
	CodeDh114PublishError       = 5303 // 商户发布错误
	CodeDh114AuditError         = 5304 // 商户审核错误
	CodeDh114NoPermission       = 5305 // 无权操作商户
	CodeDh114StatusInvalid      = 5306 // 商户状态不允许此操作
	CodeDh114CategoryError     = 5307 // 分类错误
	CodeDh114CategoryNotFound  = 5308 // 分类不存在
	CodeDh114BusinessError     = 5309 // 商户详情错误
	CodeDh114BusinessNotFound  = 5310 // 商户详情不存在
	CodeDh114GroupbuyError     = 5311 // 团购错误
	CodeDh114GroupbuyNotFound  = 5312 // 团购不存在
	CodeDh114GroupbuySoldOut   = 5313 // 团购已售罄
	CodeDh114CouponError      = 5314 // 优惠券错误
	CodeDh114CouponNotFound   = 5315 // 优惠券不存在
	CodeDh114CouponSoldOut    = 5316 // 优惠券已抢完
	CodeDh114ReviewError      = 5317 // 评价错误
	CodeDh114ReviewNotFound   = 5318 // 评价不存在
	CodeDh114FavoriteError    = 5319 // 收藏错误
	CodeDh114FavoriteExists   = 5320 // 已收藏过
	CodeDh114VerificationError = 5321 // 认证错误
	CodeDh114VerificationNotFound = 5322 // 认证记录不存在
	CodeDh114PhoneCallError  = 5323 // 电话拨打记录错误
	CodeDh114AuditRuleError  = 5324 // 审核规则错误
	CodeDh114AuditRuleNotFound = 5325 // 审核规则不存在
	CodeDh114RecommendationError = 5326 // 推荐错误
	CodeDh114RecommendationNotFound = 5327 // 推荐记录不存在
	CodeDh114StatisticError  = 5328 // 统计错误
	CodeDh114MenuError       = 5329 // 菜单错误
	CodeDh114MenuNotFound    = 5330 // 菜单不存在
	CodeDh114ReportError     = 5331 // 举报错误
	CodeDh114ReportNotFound  = 5332 // 举报记录不存在

	// ===== mall 同城商城模块错误 5401-5430 =====
	CodeMallAddressError     = 5401 // 收货地址错误
	CodeMallAddressNotFound  = 5402 // 收货地址不存在
	CodeMallAuditRuleError   = 5403 // 审核规则错误
	CodeMallAuditRuleNotFound = 5404 // 审核规则不存在
	CodeMallCartError        = 5405 // 购物车错误
	CodeMallCartNotFound     = 5406 // 购物车项不存在
	CodeMallCategoryError    = 5407 // 商城分类错误
	CodeMallCategoryNotFound = 5408 // 商城分类不存在
	CodeMallLogisticsError   = 5409 // 物流错误
	CodeMallLogisticsNotFound = 5410 // 物流记录不存在
	CodeMallOrderError       = 5411 // 订单错误
	CodeMallOrderNotFound    = 5412 // 订单不存在
	CodeMallOrderItemError   = 5413 // 订单明细错误
	CodeMallOrderItemNotFound = 5414 // 订单明细不存在
	CodeMallPaymentError     = 5415 // 支付错误
	CodeMallPaymentNotFound  = 5416 // 支付记录不存在
	CodeMallProductError     = 5417 // 商品错误
	CodeMallProductNotFound  = 5418 // 商品不存在
	CodeMallProductAuditError = 5419 // 商品审核错误
	CodeMallRefundError      = 5420 // 退款错误
	CodeMallRefundNotFound   = 5421 // 退款记录不存在
	CodeMallReportError      = 5422 // 举报错误
	CodeMallReportNotFound   = 5423 // 举报记录不存在
	CodeMallReviewError      = 5424 // 评价错误
	CodeMallReviewNotFound   = 5425 // 评价不存在
	CodeMallShopError        = 5426 // 店铺错误
	CodeMallShopNotFound     = 5427 // 店铺不存在
	CodeMallSkuError         = 5428 // SKU错误
	CodeMallSkuNotFound      = 5429 // SKU不存在
	CodeMallStatisticError   = 5430 // 统计错误

	// ===== mall 骑手扩展错误 5431-5450 =====
	CodeMallRiderError              = 5431 // 骑手通用错误
	CodeMallRiderNotFound           = 5432 // 骑手不存在
	CodeMallRiderNotVerified        = 5433 // 骑手未认证
	CodeMallRiderFrozen             = 5434 // 骑手已冻结
	CodeMallRiderOnline             = 5435 // 骑手已上线
	CodeMallRiderOffline            = 5436 // 骑手已下线
	CodeMallRiderBusy               = 5437 // 骑手配送中
	CodeMallDeliveryNotFound        = 5438 // 配送单不存在
	CodeMallDeliveryGrabbed         = 5439 // 配送单已被抢
	CodeMallDeliveryStatusInvalid   = 5440 // 配送状态不允许此操作
	CodeMallRiderSettlementNotFound = 5441 // 结算单不存在
	CodeMallRiderSettlementAudited  = 5442 // 结算单已审核
	CodeMallRiderInsufficientCredit = 5443 // 信用分不足
	CodeMallRiderSettlementError    = 5444 // 结算错误

	// ===== lbs 地图中台错误 5501-5530 =====
	CodeLbsError           = 5501 // LBS通用错误
	CodeLbsPoiNotFound     = 5502 // POI不存在
	CodeLbsRegionNotFound  = 5503 // 区域不存在
	CodeLbsGeofenceError   = 5504 // 围栏错误
	CodeLbsGeofenceNotFound = 5505 // 围栏不存在
	CodeLbsNearbyError     = 5506 // 附近搜索错误
	CodeLbsRouteError      = 5507 // 路线规划错误
	CodeLbsDistanceError   = 5508 // 距离计算错误
	CodeLbsParamInvalid    = 5509 // 参数无效
	CodeLbsNoPermission    = 5510 // 无权操作

	// ===== tenant 分站中台错误 5601-5630 =====
	CodeTenantError           = 5601 // 分站通用错误
	CodeTenantStationNotFound = 5602 // 分站不存在
	CodeTenantStaffError      = 5603 // 员工错误
	CodeTenantStaffNotFound   = 5604 // 员工不存在
	CodeTenantConfigError     = 5605 // 配置错误
	CodeTenantDomainError     = 5606 // 域名错误
	CodeTenantNoPermission    = 5607 // 无权操作

	// ===== merchant 商户中台错误 5701-5730 =====
	CodeMerchantError              = 5701 // 商户通用错误
	CodeMerchantShopNotFound       = 5702 // 店铺不存在
	CodeMerchantStaffError         = 5703 // 员工错误
	CodeMerchantStaffNotFound      = 5704 // 员工不存在
	CodeMerchantSettleError        = 5705 // 结算错误
	CodeMerchantSettleNotFound     = 5706 // 结算单不存在
	CodeMerchantCategoryError      = 5707 // 类目错误
	CodeMerchantVerifyError        = 5708 // 认证错误
	CodeMerchantVerifyNotFound     = 5709 // 认证记录不存在
	CodeMerchantNoPermission       = 5710 // 无权操作
	CodeMerchantAlreadySettled     = 5711 // 已入驻

	// ===== marketing 营销中台错误 5801-5830 =====
	CodeMarketingError             = 5801 // 营销通用错误
	CodeMarketingAdNotFound        = 5802 // 广告位不存在
	CodeMarketingCouponError       = 5803 // 优惠券错误
	CodeMarketingCouponNotFound    = 5804 // 优惠券不存在
	CodeMarketingCouponUsed        = 5805 // 优惠券已使用
	CodeMarketingCouponExpired     = 5806 // 优惠券已过期
	CodeMarketingCouponInsufficient = 5807 // 优惠券已领完
	CodeMarketingSignError         = 5808 // 签到错误
	CodeMarketingSignAlready       = 5809 // 今日已签到
	CodeMarketingActivityError     = 5810 // 活动错误
	CodeMarketingActivityNotFound  = 5811 // 活动不存在

	// ===== distribution 分销中台错误 5901-5930 =====
	CodeDistributionError              = 5901 // 分销通用错误
	CodeDistributionPartnerNotFound    = 5902 // 合伙人不存在
	CodeDistributionChannelError       = 5903 // 渠道错误
	CodeDistributionChannelNotFound    = 5904 // 渠道不存在
	CodeDistributionCommissionError    = 5905 // 佣金错误
	CodeDistributionCommissionNotFound = 5906 // 佣金记录不存在
	CodeDistributionLevelError         = 5907 // 等级错误
	CodeDistributionWithdrawalError    = 5908 // 提现错误
	CodeDistributionWithdrawalNotFound = 5909 // 提现记录不存在
	CodeDistributionInsufficientBalance = 5910 // 余额不足
	CodeDistributionAlreadyPartner     = 5911 // 已是合伙人

	// ===== diy 页面中台错误 6001-6030 =====
	CodeDiyError             = 6001 // DIY通用错误
	CodeDiyPageNotFound      = 6002 // 页面不存在
	CodeDiyComponentError    = 6003 // 组件错误
	CodeDiyComponentNotFound = 6004 // 组件不存在
	CodeDiyTemplateError     = 6005 // 模板错误
	CodeDiyTemplateNotFound  = 6006 // 模板不存在
	CodeDiyStatError         = 6007 // 统计错误
	CodeDiyNoPermission      = 6008 // 无权操作
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

	// 同城二手物品模块错误
	CodeErshouError:        "二手物品错误",
	CodeErshouNotFound:     "二手物品不存在",
	CodeErshouPublishError: "二手物品发布错误",
	CodeErshouAuditError:   "二手物品审核错误",

	// 同城招聘求职模块错误
	CodeJobError:            "招聘求职错误",
	CodeJobNotFound:         "职位不存在",
	CodeJobPublishError:     "职位发布失败",
	CodeJobAuditError:       "职位审核错误",
	CodeJobNoPermission:     "无权操作职位",
	CodeResumeError:         "简历错误",
	CodeResumeNotFound:      "简历不存在",
	CodeApplicationError:    "投递错误",
	CodeApplicationNotFound: "投递不存在",
	CodeInterviewError:      "面试错误",
	CodeInterviewNotFound:   "面试不存在",
	CodeMessageError:        "消息错误",
	CodeReportError:         "举报错误",
	CodeReviewError:         "评价错误",
	CodeCompanyError:        "公司错误",
	CodeCompanyNotFound:     "公司不存在",
	CodeAuditRuleError:      "审核规则错误",

	// 房产模块错误
	CodeHouseError:             "房产错误",
	CodeHouseNotFound:          "房产不存在",
	CodeHousePublishError:      "房产发布错误",
	CodeHouseAuditError:        "房产审核错误",
	CodeHouseNoPermission:      "无权操作房产",
	CodeHouseReportError:       "房产举报错误",
	CodeHouseReviewError:       "房产评价错误",
	CodeHouseContractError:     "合同错误",
	CodeHouseContractNotFound:  "合同不存在",
	CodeHouseViewingError:      "看房预约错误",
	CodeHouseViewingNotFound:   "看房预约不存在",
	CodeHouseCommunityError:    "小区错误",
	CodeHouseCommunityNotFound: "小区不存在",
	CodeHouseAgentError:        "经纪人错误",
	CodeHouseAgentNotFound:     "经纪人不存在",
	CodeHouseAuditRuleError:    "审核规则错误",

	// 同城车辆买卖模块错误
	CodeCarError:           "车辆错误",
	CodeCarNotFound:        "车源不存在",
	CodeCarPublishError:    "车源发布错误",
	CodeCarAuditError:      "车源审核错误",
	CodeCarNoPermission:    "无权操作车源",
	CodeInspectionError:    "车况检测错误",
	CodeInspectionNotFound: "检测记录不存在",
	CodeEvaluationError:    "车辆评估错误",
	CodeEvaluationNotFound: "评估记录不存在",
	CodeFinancingError:     "分期方案错误",
	CodeFinancingNotFound:  "分期方案不存在",
	CodeInsuranceError:     "车险方案错误",
	CodeInsuranceNotFound:  "车险方案不存在",
	CodeTestDriveError:     "试驾预约错误",
	CodeTestDriveNotFound:  "试驾预约不存在",
	CodeTransferError:      "过户办理错误",
	CodeTransferNotFound:   "过户记录不存在",
	CodeReportErrorCar:     "车源举报错误",
	CodeReviewErrorCar:     "车源评价错误",
	CodeAuditRuleErrorCar:  "审核规则错误",

	// 第三方服务错误
	CodeThirdPartyError: "第三方服务错误",
	CodeMapAPIError:     "地图API错误",
	CodeStorageError:    "存储服务错误",
	CodeSMSError:        "短信服务错误",
	CodeWeChatError:     "微信服务错误",
	CodeOAuthError:      "第三方登录错误",

	// pay 支付中台错误
	CodePayError:               "支付错误",
	CodePayOrderNotFound:       "支付订单不存在",
	CodePayOrderClosed:         "订单已关闭",
	CodePayOrderPaid:           "订单已支付",
	CodePayRefundNotFound:      "退款单不存在",
	CodePayRefundExceed:        "退款金额超过订单",
	CodePayEscrowNotFound:      "担保交易不存在",
	CodePayEscrowNotFrozen:     "担保交易非冻结状态",
	CodePayInsufficientBalance:  "余额不足",
	CodePayWithdrawNotFound:     "提现单不存在",
	CodePayWithdrawNotPending:   "提现单非待审核状态",
	CodePaySettlementNotFound:  "结算单不存在",
	CodePayAccountNotFound:     "资金账户不存在",
	CodePayChannelNotFound:     "支付渠道不存在",

	// im 即时通讯中台错误
	CodeIMError:                "IM错误",
	CodeIMSessionNotFound:      "会话不存在",
	CodeIMNotParticipant:       "非会话参与者",
	CodeIMMessageNotFound:      "消息不存在",
	CodeIMMessageRecalled:      "消息已撤回",
	CodeIMGroupNotFound:        "群组不存在",
	CodeIMGroupMemberExists:    "已是群成员",
	CodeIMGroupMemberNotFound:  "群成员不存在",
	CodeIMNotGroupOwner:        "非群主",
	CodeIMPrivacyNotFound:      "隐私号码不存在",
	CodeIMPrivacyUnbound:       "隐私号码已解绑",
	CodeIMNotificationNotFound: "通知不存在",
	CodeIMUserSettingsNotFound: "用户设置不存在",
	CodeIMSessionExists:        "会话已存在",

	// material 物料中台错误
	CodeMaterialError:            "物料错误",
	CodeMaterialFileNotFound:     "文件不存在",
	CodeMaterialImageNotFound:    "图片不存在",
	CodeMaterialVideoNotFound:    "视频不存在",
	CodeMaterialFeatureNotFound:  "图片特征不存在",
	CodeMaterialUploadFailed:     "文件上传失败",
	CodeMaterialUnsupportedType:  "不支持的文件类型",
	CodeMaterialCategoryNotFound: "分类不存在",
	CodeMaterialTagNotFound:      "标签不存在",
	CodeMaterialOCRFail:           "OCR识别失败",
	CodeMaterialSearchFail:        "搜索失败",
	CodeMaterialHashExists:       "文件哈希已存在",

	// risk 风控中台错误
	CodeRiskError:              "风控错误",
	CodeRiskReportNotFound:    "举报不存在",
	CodeRiskWordNotFound:      "敏感词不存在",
	CodeRiskRuleNotFound:      "审核规则不存在",
	CodeRiskBlacklistNotFound: "黑名单记录不存在",
	CodeRiskScoreNotFound:     "用户风险分不存在",
	CodeRiskViolationNotFound: "违规处罚不存在",
	CodeRiskAlreadyBlacklist:  "已在黑名单中",
	CodeRiskUserBanned:        "用户已被封禁",
	CodeRiskContentRejected:   "内容审核未通过",
	CodeRiskAppealNotFound:    "申诉记录不存在",
	CodeRiskEvidenceNotFound:  "证据不存在",
	CodeRiskAuditLogNotFound:  "审核日志不存在",

	// ai 智能中台错误
	CodeAIError:                  "AI错误",
	CodeAITaskNotFound:           "AI任务不存在",
	CodeAIModelNotFound:          "AI模型不存在",
	CodeAIModelDisabled:          "AI模型已禁用",
	CodeAIPromptNotFound:         "提示词模板不存在",
	CodeAIGenerationNotFound:     "生成记录不存在",
	CodeAIUnsupportedType:        "不支持的AI任务类型",
	CodeAIChatNotFound:           "对话会话不存在",
	CodeAIChatMessageNotFound:    "对话消息不存在",
	CodeAIRecommendationNotFound: "推荐记录不存在",
	CodeAITrainingNotFound:       "训练数据不存在",
	CodeAIModelConfigNotFound:    "模型配置不存在",

	// pinche 拼车出行模块错误
	CodePincheError:              "拼车错误",
	CodePincheNotFound:           "拼车行程不存在",
	CodePincheRouteNotFound:      "路线不存在",
	CodePincheBookingConflict:    "预订冲突",
	CodePincheDriverNotVerified:  "车主未认证",
	CodePincheSeatUnavailable:    "座位已满",
	CodePincheTripOngoing:        "行程进行中",
	CodePincheInsuranceFailed:    "保险投保失败",
	CodePinchePaymentFailed:      "支付失败",
	CodePincheDriverNotFound:     "车主不存在",
	CodePincheVehicleNotFound:    "车辆不存在",
	CodePincheBookingNotFound:    "预订不存在",
	CodePincheTripNotFound:       "行程记录不存在",
	CodePincheRatingNotFound:     "评价不存在",
	CodePincheEmergencyNotFound:  "紧急联系人不存在",
	CodePinchePaymentNotFound:    "支付记录不存在",
	CodePincheRefundNotFound:     "退款记录不存在",
	CodePincheComplaintNotFound:  "投诉不存在",
	CodePincheMessageNotFound:    "消息不存在",
	CodePincheNoPermission:       "无权操作",
	CodePincheStatusInvalid:      "状态不允许此操作",
	CodePincheAuditError:         "拼车审核错误",
	CodePincheAuditRuleError:     "审核规则错误",
	CodePincheRouteFavoriteError: "常用路线错误",
	CodePincheLocationError:      "实时位置错误",
	CodePincheCancelError:        "取消错误",
	CodePincheRefundError:        "退款错误",
	CodePincheAlreadyBooked:      "已预订过该行程",
        CodePincheDriverBusy:         "车主行程冲突",
        CodePincheTripNotStarted:     "行程尚未开始",

        // linggong 零工兼职模块错误
        CodeLinggongError:                "零工兼职错误",
        CodeLinggongNotFound:             "零工兼职岗位不存在",
        CodeLinggongPublishError:         "零工兼职发布错误",
        CodeLinggongAuditError:           "零工兼职审核错误",
        CodeLinggongNoPermission:         "无权操作零工兼职",
        CodeLinggongStatusInvalid:        "状态不允许此操作",
        CodeLinggongApplicationNotFound:  "报名申请不存在",
        CodeLinggongApplicationDuplicate: "重复报名",
        CodeLinggongApplicationFull:      "报名名额已满",
        CodeLinggongTaskNotFound:         "任务不存在",
        CodeLinggongTaskClaimError:       "任务认领错误",
        CodeLinggongEmployerNotFound:     "雇主不存在",
        CodeLinggongWorkerNotFound:       "零工不存在",
        CodeLinggongContractNotFound:     "合同不存在",
        CodeLinggongContractSignError:    "合同签署错误",
        CodeLinggongPaymentNotFound:      "支付记录不存在",
        CodeLinggongPaymentError:         "支付错误",
        CodeLinggongWithdrawalNotFound:   "提现记录不存在",
        CodeLinggongWithdrawalError:      "提现错误",
        CodeLinggongRatingNotFound:       "评价不存在",
        CodeLinggongRatingDuplicate:      "重复评价",
        CodeLinggongCertificationNotFound: "资质证书不存在",
        CodeLinggongCertificationError:   "资质证书错误",
        CodeLinggongCreditNotFound:       "信用记录不存在",
        CodeLinggongDisputeNotFound:      "纠纷不存在",
        CodeLinggongDisputeError:         "纠纷处理错误",
        CodeLinggongSkillNotFound:        "技能不存在",
        CodeLinggongAuditRuleNotFound:    "审核规则不存在",
        CodeLinggongAttendanceError:      "考勤打卡错误",
        CodeLinggongStatisticError:       "统计错误",

        // dh114 同城114模块错误
        CodeDh114Error:              "同城114错误",
        CodeDh114NotFound:           "商户不存在",
        CodeDh114PublishError:       "商户发布错误",
        CodeDh114AuditError:         "商户审核错误",
        CodeDh114NoPermission:       "无权操作商户",
        CodeDh114StatusInvalid:      "商户状态不允许此操作",
        CodeDh114CategoryError:      "分类错误",
        CodeDh114CategoryNotFound:   "分类不存在",
        CodeDh114BusinessError:      "商户详情错误",
        CodeDh114BusinessNotFound:   "商户详情不存在",
        CodeDh114GroupbuyError:      "团购错误",
        CodeDh114GroupbuyNotFound:   "团购不存在",
        CodeDh114GroupbuySoldOut:    "团购已售罄",
        CodeDh114CouponError:        "优惠券错误",
        CodeDh114CouponNotFound:     "优惠券不存在",
        CodeDh114CouponSoldOut:      "优惠券已抢完",
        CodeDh114ReviewError:        "评价错误",
        CodeDh114ReviewNotFound:     "评价不存在",
        CodeDh114FavoriteError:      "收藏错误",
        CodeDh114FavoriteExists:     "已收藏过",
        CodeDh114VerificationError:  "认证错误",
        CodeDh114VerificationNotFound: "认证记录不存在",
        CodeDh114PhoneCallError:     "电话拨打记录错误",
        CodeDh114AuditRuleError:     "审核规则错误",
        CodeDh114AuditRuleNotFound:  "审核规则不存在",
        CodeDh114RecommendationError: "推荐错误",
        CodeDh114RecommendationNotFound: "推荐记录不存在",
        CodeDh114StatisticError:     "统计错误",
        CodeDh114MenuError:          "菜单错误",
        CodeDh114MenuNotFound:       "菜单不存在",
        CodeDh114ReportError:        "举报错误",
        CodeDh114ReportNotFound:     "举报记录不存在",

        // love 相亲交友模块错误
        CodeLoveError:                "相亲交友错误",
        CodeLoveNotFound:             "用户资料不存在",
        CodeLoveAlreadyLiked:         "已经喜欢过该用户",
        CodeLoveBlocked:              "已被对方拉黑",
        CodeLoveMemberExpired:        "会员已过期",
        CodeLoveVerificationFailed:   "实名认证失败",
        CodeLoveMatchFailed:          "匹配失败",
        CodeLoveStoryNotFound:        "动态不存在",
        CodeLoveGiftNotFound:         "礼物不存在",
        CodeLoveInsufficientCredits:  "金币余额不足",
        CodeLoveNoPermission:         "无权操作",
        CodeLoveAlreadyMatched:       "已经匹配过",
        CodeLoveSuperLikeLimited:     "心动信号今日已用完",
        CodeLoveProfileNotComplete:   "资料未完善",
        CodeLoveNotVerified:          "未通过实名认证",
        CodeLoveReportError:          "举报错误",
        CodeLoveReportNotFound:       "举报记录不存在",
        CodeLoveAuditRuleError:       "审核规则错误",
        CodeLoveAuditRuleNotFound:    "审核规则不存在",
        CodeLoveMemberNotFound:       "会员订阅不存在",
        CodeLoveVisitNotFound:        "访客记录不存在",
        CodeLoveBlockNotFound:        "拉黑记录不存在",
        CodeLoveNotificationError:    "通知错误",
        CodeLoveNotificationNotFound: "通知不存在",
        CodeLoveRecommendationNotFound: "推荐记录不存在",
        CodeLoveImpressionError:      "印象标签错误",
        CodeLoveChatSessionNotFound:  "聊天会话不存在",
        CodeLoveVerificationNotFound: "认证记录不存在",
        CodeLoveMemberLevelNotFound:  "会员等级不存在",
        CodeLovePrivacyError:         "隐私设置错误",

	// mall 同城商城模块错误
	CodeMallAddressError:      "收货地址错误",
	CodeMallAddressNotFound:   "收货地址不存在",
	CodeMallAuditRuleError:    "审核规则错误",
	CodeMallAuditRuleNotFound: "审核规则不存在",
	CodeMallCartError:         "购物车错误",
	CodeMallCartNotFound:      "购物车项不存在",
	CodeMallCategoryError:     "商城分类错误",
	CodeMallCategoryNotFound:  "商城分类不存在",
	CodeMallLogisticsError:    "物流错误",
	CodeMallLogisticsNotFound: "物流记录不存在",
	CodeMallOrderError:        "订单错误",
	CodeMallOrderNotFound:     "订单不存在",
	CodeMallOrderItemError:    "订单明细错误",
	CodeMallOrderItemNotFound: "订单明细不存在",
	CodeMallPaymentError:      "支付错误",
	CodeMallPaymentNotFound:   "支付记录不存在",
	CodeMallProductError:      "商品错误",
	CodeMallProductNotFound:   "商品不存在",
	CodeMallProductAuditError: "商品审核错误",
	CodeMallRefundError:       "退款错误",
	CodeMallRefundNotFound:    "退款记录不存在",
	CodeMallReportError:       "举报错误",
	CodeMallReportNotFound:    "举报记录不存在",
	CodeMallReviewError:       "评价错误",
	CodeMallReviewNotFound:    "评价不存在",
	CodeMallShopError:         "店铺错误",
	CodeMallShopNotFound:      "店铺不存在",
	CodeMallSkuError:          "SKU错误",
	CodeMallSkuNotFound:       "SKU不存在",
	CodeMallStatisticError:    "统计错误",

	// mall 骑手扩展错误
	CodeMallRiderError:              "骑手错误",
	CodeMallRiderNotFound:           "骑手不存在",
	CodeMallRiderNotVerified:        "骑手未认证",
	CodeMallRiderFrozen:             "骑手已冻结",
	CodeMallRiderOnline:             "骑手已上线",
	CodeMallRiderOffline:            "骑手已下线",
	CodeMallRiderBusy:               "骑手配送中",
	CodeMallDeliveryNotFound:        "配送单不存在",
	CodeMallDeliveryGrabbed:         "配送单已被抢",
	CodeMallDeliveryStatusInvalid:   "配送状态不允许此操作",
	CodeMallRiderSettlementNotFound: "结算单不存在",
	CodeMallRiderSettlementAudited:  "结算单已审核",
	CodeMallRiderInsufficientCredit: "信用分不足",
	CodeMallRiderSettlementError:    "结算错误",
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
