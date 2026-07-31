package utils

import (
	"sort"
	"testing"
)

// codeSpec 描述一个错误码常量的预期契约：名称、数值、消息
type codeSpec struct {
	name    string
	code    int
	message string
}

// allErrorCodes 汇总全部已定义错误码常量的契约，用于一致性校验。
// 每个 const 必须在 codeMessages 中有对应消息，否则 GetMessage 会静默回退 "未知错误"。
// 通过切片（而非 map）保留所有条目，可检测常量值重复碰撞。
func allErrorCodes() []codeSpec {
	return []codeSpec{
		// 成功
		{"CodeSuccess", CodeSuccess, "success"},
		// 系统级 1000-1999
		{"CodeSystemError", CodeSystemError, "系统错误"},
		{"CodeParamInvalid", CodeParamInvalid, "参数无效"},
		{"CodeParamMissing", CodeParamMissing, "参数缺失"},
		{"CodeUnauthorized", CodeUnauthorized, "未授权访问"},
		{"CodeForbidden", CodeForbidden, "禁止访问"},
		{"CodeNotFound", CodeNotFound, "资源不存在"},
		{"CodeAlreadyExists", CodeAlreadyExists, "资源已存在"},
		{"CodeTimeout", CodeTimeout, "请求超时"},
		{"CodeTooManyRequests", CodeTooManyRequests, "请求过于频繁"},
		{"CodeDBError", CodeDBError, "数据库错误"},
		{"CodeDBQueryError", CodeDBQueryError, "数据库查询错误"},
		{"CodeDBInsertError", CodeDBInsertError, "数据库插入错误"},
		{"CodeDBUpdateError", CodeDBUpdateError, "数据库更新错误"},
		{"CodeDBDeleteError", CodeDBDeleteError, "数据库删除错误"},
		{"CodeDBRecordNotFound", CodeDBRecordNotFound, "记录不存在"},
		{"CodeRedisError", CodeRedisError, "Redis缓存错误"},
		{"CodeFileError", CodeFileError, "文件错误"},
		{"CodeFileUploadError", CodeFileUploadError, "文件上传错误"},
		{"CodeFileNotFound", CodeFileNotFound, "文件不存在"},
		{"CodeFileTooLarge", CodeFileTooLarge, "文件过大"},
		{"CodeFileTypeInvalid", CodeFileTypeInvalid, "文件类型不支持"},
		{"CodeFilePresignError", CodeFilePresignError, "当前存储不支持预签名直传"},
		{"CodeFileSTSError", CodeFileSTSError, "STS 临时凭据不可用，请配置 sts 或使用普通上传"},
		// 用户 2000-2999
		{"CodeUserError", CodeUserError, "用户错误"},
		{"CodeUserNotFound", CodeUserNotFound, "用户不存在"},
		{"CodeUserAlreadyExists", CodeUserAlreadyExists, "用户已存在"},
		{"CodeUserPasswordError", CodeUserPasswordError, "密码错误"},
		{"CodeUserDisabled", CodeUserDisabled, "用户已禁用"},
		{"CodeUserNotLoggedIn", CodeUserNotLoggedIn, "用户未登录"},
		{"CodeUserTokenExpired", CodeUserTokenExpired, "Token已过期"},
		{"CodeUserTokenInvalid", CodeUserTokenInvalid, "Token无效"},
		{"CodeUserPermissionDenied", CodeUserPermissionDenied, "用户权限不足"},
		// 地区 2100-2199
		{"CodeRegionError", CodeRegionError, "地区错误"},
		{"CodeRegionNotFound", CodeRegionNotFound, "地区不存在"},
		{"CodeRegionInvalid", CodeRegionInvalid, "地区无效"},
		// 权限 2200-2299
		{"CodePermissionError", CodePermissionError, "权限错误"},
		{"CodePermissionDenied", CodePermissionDenied, "权限不足"},
		{"CodeRoleNotFound", CodeRoleNotFound, "角色不存在"},
		{"CodeRoleAlreadyExists", CodeRoleAlreadyExists, "角色已存在"},
		{"CodePermissionNotFound", CodePermissionNotFound, "权限不存在"},
		{"CodePermissionAlreadyExists", CodePermissionAlreadyExists, "权限已存在"},
		// 分类 2300-2399
		{"CodeCategoryError", CodeCategoryError, "分类错误"},
		{"CodeCategoryNotFound", CodeCategoryNotFound, "分类不存在"},
		{"CodeCategoryAlreadyExists", CodeCategoryAlreadyExists, "分类已存在"},
		// 头条 2400-2499
		{"CodeNewsError", CodeNewsError, "头条错误"},
		{"CodeNewsNotFound", CodeNewsNotFound, "头条不存在"},
		{"CodeNewsPublishError", CodeNewsPublishError, "头条发布错误"},
		// 第三方 4000-4999
		{"CodeThirdPartyError", CodeThirdPartyError, "第三方服务错误"},
		{"CodeMapAPIError", CodeMapAPIError, "地图API错误"},
		{"CodeStorageError", CodeStorageError, "存储服务错误"},
		{"CodeSMSError", CodeSMSError, "短信服务错误"},
		{"CodeWeChatError", CodeWeChatError, "微信服务错误"},
		{"CodeOAuthError", CodeOAuthError, "第三方登录错误"},
	}
}

func TestGetMessage_Success(t *testing.T) {
	if got := GetMessage(CodeSuccess); got != "success" {
		t.Errorf("GetMessage(CodeSuccess) = %q, want %q", got, "success")
	}
}

func TestGetMessage_KnownCodes(t *testing.T) {
	cases := []struct {
		code int
		want string
	}{
		{CodeSystemError, "系统错误"},
		{CodeUnauthorized, "未授权访问"},
		{CodeNotFound, "资源不存在"},
		{CodeAlreadyExists, "资源已存在"},
		{CodeUserNotFound, "用户不存在"},
		{CodeUserAlreadyExists, "用户已存在"},
		{CodeUserTokenExpired, "Token已过期"},
		{CodeUserTokenInvalid, "Token无效"},
		{CodeRegionNotFound, "地区不存在"},
		{CodeRoleNotFound, "角色不存在"},
		{CodePermissionDenied, "权限不足"},
		{CodeCategoryNotFound, "分类不存在"},
		{CodeNewsNotFound, "头条不存在"},
		{CodeNewsPublishError, "头条发布错误"},
		{CodeFileTooLarge, "文件过大"},
		{CodeFilePresignError, "当前存储不支持预签名直传"},
		{CodeFileSTSError, "STS 临时凭据不可用，请配置 sts 或使用普通上传"},
		{CodeSMSError, "短信服务错误"},
		{CodeOAuthError, "第三方登录错误"},
	}
	for _, c := range cases {
		if got := GetMessage(c.code); got != c.want {
			t.Errorf("GetMessage(%d) = %q, want %q", c.code, got, c.want)
		}
	}
}

func TestGetMessage_UnknownCode(t *testing.T) {
	// 选取一个未在常量列表中定义的码值
	unknown := 99999
	if got := GetMessage(unknown); got != "未知错误" {
		t.Errorf("GetMessage(%d) = %q, want %q", unknown, got, "未知错误")
	}
	// 0 之外的小正数也视为未知
	if got := GetMessage(1); got != "未知错误" {
		t.Errorf("GetMessage(1) = %q, want %q", got, "未知错误")
	}
}

// TestAllDefinedCodesHaveMessage 防御性校验：每个错误码常量必须在 codeMessages 中注册，
// 否则 GetMessage 会静默回退 "未知错误"，导致前端无法展示正确提示。
func TestAllDefinedCodesHaveMessage(t *testing.T) {
	for _, spec := range allErrorCodes() {
		got := GetMessage(spec.code)
		if got == "未知错误" {
			t.Errorf("错误码 %s(%d) 未注册消息，GetMessage 回退为 \"未知错误\"", spec.name, spec.code)
		}
		if got != spec.message {
			t.Errorf("GetMessage(%s=%d) = %q, want %q", spec.name, spec.code, got, spec.message)
		}
	}
}

// TestErrorCodeConstants_NoDuplicateValues 校验所有错误码常量值两两不重复，
// 防止新增常量时误用既有码值导致语义冲突（map 构造会无声合并同值键，故用切片+显式去重检测）。
func TestErrorCodeConstants_NoDuplicateValues(t *testing.T) {
	specs := allErrorCodes()
	seen := make(map[int]string, len(specs))
	for _, spec := range specs {
		if prev, ok := seen[spec.code]; ok {
			t.Errorf("错误码 %d 重复定义：%s 与 %s", spec.code, prev, spec.name)
			continue
		}
		seen[spec.code] = spec.name
	}
}

// TestErrorCodeConstantValues_Stable 锁定被 handler 业务码路由依赖的关键哨兵码值，
// 防止误改常量数值导致 handler 错误码透传断言失效。
func TestErrorCodeConstantValues_Stable(t *testing.T) {
	// 直接断言常量等于文档化数值（map 形式会以值为键退化成恒真断言，故用显式 == 校验）
	assertEq := func(name string, got, want int) {
		t.Helper()
		if got != want {
			t.Errorf("错误码常量值漂移：%s 实际=%d, 期望=%d", name, got, want)
		}
	}
	assertEq("CodeSuccess", CodeSuccess, 0)
	assertEq("CodeSystemError", CodeSystemError, 1001)
	assertEq("CodeParamInvalid", CodeParamInvalid, 1002)
	assertEq("CodeParamMissing", CodeParamMissing, 1003)
	assertEq("CodeUnauthorized", CodeUnauthorized, 1004)
	assertEq("CodeForbidden", CodeForbidden, 1005)
	assertEq("CodeNotFound", CodeNotFound, 1006)
	assertEq("CodeAlreadyExists", CodeAlreadyExists, 1007)
	assertEq("CodeTimeout", CodeTimeout, 1008)
	assertEq("CodeTooManyRequests", CodeTooManyRequests, 1009)
	assertEq("CodeDBError", CodeDBError, 1101)
	assertEq("CodeRedisError", CodeRedisError, 1201)
	assertEq("CodeFileError", CodeFileError, 1301)
	assertEq("CodeFileUploadError", CodeFileUploadError, 1302)
	assertEq("CodeFileNotFound", CodeFileNotFound, 1303)
	assertEq("CodeFileTooLarge", CodeFileTooLarge, 1304)
	assertEq("CodeFileTypeInvalid", CodeFileTypeInvalid, 1305)
	assertEq("CodeFilePresignError", CodeFilePresignError, 1306)
	assertEq("CodeFileSTSError", CodeFileSTSError, 1307)
	assertEq("CodeUserError", CodeUserError, 2001)
	assertEq("CodeUserNotFound", CodeUserNotFound, 2002)
	assertEq("CodeUserAlreadyExists", CodeUserAlreadyExists, 2003)
	assertEq("CodeUserPasswordError", CodeUserPasswordError, 2004)
	assertEq("CodeUserDisabled", CodeUserDisabled, 2005)
	assertEq("CodeUserNotLoggedIn", CodeUserNotLoggedIn, 2006)
	assertEq("CodeUserTokenExpired", CodeUserTokenExpired, 2007)
	assertEq("CodeUserTokenInvalid", CodeUserTokenInvalid, 2008)
	assertEq("CodeUserPermissionDenied", CodeUserPermissionDenied, 2009)
	assertEq("CodeRegionError", CodeRegionError, 2101)
	assertEq("CodeRegionNotFound", CodeRegionNotFound, 2102)
	assertEq("CodeRegionInvalid", CodeRegionInvalid, 2103)
	assertEq("CodePermissionError", CodePermissionError, 2201)
	assertEq("CodePermissionDenied", CodePermissionDenied, 2202)
	assertEq("CodeRoleNotFound", CodeRoleNotFound, 2203)
	assertEq("CodeRoleAlreadyExists", CodeRoleAlreadyExists, 2204)
	assertEq("CodePermissionNotFound", CodePermissionNotFound, 2205)
	assertEq("CodePermissionAlreadyExists", CodePermissionAlreadyExists, 2206)
	assertEq("CodeCategoryError", CodeCategoryError, 2301)
	assertEq("CodeCategoryNotFound", CodeCategoryNotFound, 2302)
	assertEq("CodeCategoryAlreadyExists", CodeCategoryAlreadyExists, 2303)
	assertEq("CodeNewsError", CodeNewsError, 2401)
	assertEq("CodeNewsNotFound", CodeNewsNotFound, 2402)
	assertEq("CodeNewsPublishError", CodeNewsPublishError, 2403)
	assertEq("CodeThirdPartyError", CodeThirdPartyError, 4001)
	assertEq("CodeMapAPIError", CodeMapAPIError, 4002)
	assertEq("CodeStorageError", CodeStorageError, 4003)
	assertEq("CodeSMSError", CodeSMSError, 4004)
	assertEq("CodeWeChatError", CodeWeChatError, 4005)
	assertEq("CodeOAuthError", CodeOAuthError, 4006)
}

// TestRegisterCode_NewCode 注册全新码值后 GetMessage 可读取
func TestRegisterCode_NewCode(t *testing.T) {
	const customCode = 99001
	const customMsg = "测试自定义错误"
	// 确保测试前未注册
	if got := GetMessage(customCode); got != "未知错误" {
		t.Fatalf("前置条件失败：码 %d 已存在消息 %q", customCode, got)
	}

	RegisterCode(customCode, customMsg)
	t.Cleanup(func() {
		delete(codeMessages, customCode)
	})

	if got := GetMessage(customCode); got != customMsg {
		t.Errorf("RegisterCode 后 GetMessage(%d) = %q, want %q", customCode, got, customMsg)
	}
}

// TestRegisterCode_OverrideExisting 注册已存在码值会覆盖原消息
func TestRegisterCode_OverrideExisting(t *testing.T) {
	const code = CodeNotFound
	const original = "资源不存在"
	const override = "测试覆盖消息"

	if got := GetMessage(code); got != original {
		t.Fatalf("前置条件失败：GetMessage(%d) = %q, want %q", code, got, original)
	}

	RegisterCode(code, override)
	t.Cleanup(func() {
		// 恢复原始消息
		RegisterCode(code, original)
	})

	if got := GetMessage(code); got != override {
		t.Errorf("覆盖后 GetMessage(%d) = %q, want %q", code, got, override)
	}
}

// TestAllErrorCodes_RangeSanity 排序后校验码值范围，确保码值分布符合分段规则
// （0 成功；1000-1999 系统/DB/缓存/文件；2000-2999 用户/地区/权限/分类/头条；4000-4999 第三方）。
func TestAllErrorCodes_RangeSanity(t *testing.T) {
	specs := allErrorCodes()
	list := make([]int, 0, len(specs))
	for _, spec := range specs {
		list = append(list, spec.code)
	}
	sort.Ints(list)

	// 成功码 0
	if list[0] != CodeSuccess {
		t.Errorf("最小码值应为 CodeSuccess(0)，实际 %d", list[0])
	}
	// 业务码下限 1000，上限 < 5000（第三方 4000-4999）
	for _, c := range list {
		if c != 0 && (c < 1000 || c >= 5000) {
			t.Errorf("错误码 %d 超出预期范围 [1000, 5000)", c)
		}
	}
}
