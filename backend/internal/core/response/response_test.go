// Package response 统一响应封装单元测试
// 覆盖 Success/SuccessWithMessage/Fail/FailWithData/BadRequest/
// Unauthorized/Forbidden/NotFound/ServerError 构造器（含空消息默认值兜底）、
// HTTPStatus 业务码→HTTP 状态码映射（含 default 未知码降级 200）、
// NewPageResult 字段透传。纯函数无外部依赖。
package response

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// --- 常量稳定性（防止意外修改业务码，破坏前端协议） ---

func TestCodeConstants(t *testing.T) {
	assert.Equal(t, 0, CodeSuccess, "CodeSuccess 应为 0")
	assert.Equal(t, 400, CodeBadRequest)
	assert.Equal(t, 401, CodeUnauthorized)
	assert.Equal(t, 403, CodeForbidden)
	assert.Equal(t, 404, CodeNotFound)
	assert.Equal(t, 500, CodeServerError)
}

// --- Success ---

func TestSuccess_DefaultMessage(t *testing.T) {
	r := Success("hello")
	assert.Equal(t, CodeSuccess, r.Code)
	assert.Equal(t, "success", r.Message, "Success 默认消息应为 success")
	assert.Equal(t, "hello", r.Data)
}

func TestSuccess_NilData(t *testing.T) {
	r := Success(nil)
	assert.Equal(t, CodeSuccess, r.Code)
	assert.Nil(t, r.Data)
}

func TestSuccess_StructData(t *testing.T) {
	type item struct{ ID int }
	r := Success(item{ID: 7})
	assert.Equal(t, CodeSuccess, r.Code)
	if it, ok := r.Data.(item); !ok {
		t.Fatalf("Data 类型应为 item，实际 %T", r.Data)
	} else {
		assert.Equal(t, 7, it.ID)
	}
}

// --- SuccessWithMessage ---

func TestSuccessWithMessage_CustomMessageAndData(t *testing.T) {
	r := SuccessWithMessage("创建成功", map[string]int{"id": 42})
	assert.Equal(t, CodeSuccess, r.Code)
	assert.Equal(t, "创建成功", r.Message)
	assert.Equal(t, map[string]int{"id": 42}, r.Data)
}

func TestSuccessWithMessage_EmptyMessage(t *testing.T) {
	// 空消息不做兜底，原样保留（区别于 Unauthorized 等的默认值兜底）
	r := SuccessWithMessage("", nil)
	assert.Equal(t, CodeSuccess, r.Code)
	assert.Equal(t, "", r.Message)
	assert.Nil(t, r.Data)
}

// --- Fail ---

func TestFail_CodeAndMessage(t *testing.T) {
	r := Fail(1002, "参数无效")
	assert.Equal(t, 1002, r.Code)
	assert.Equal(t, "参数无效", r.Message)
	assert.Nil(t, r.Data, "Fail 默认 Data 为 nil")
}

func TestFail_ZeroCode(t *testing.T) {
	// 业务上不推荐，但代码不应阻止
	r := Fail(0, "ok")
	assert.Equal(t, 0, r.Code)
	assert.Equal(t, "ok", r.Message)
}

// --- FailWithData ---

func TestFailWithData_DataPassedThrough(t *testing.T) {
	r := FailWithData(2004, "密码错误", map[string]any{"attempts": 3})
	assert.Equal(t, 2004, r.Code)
	assert.Equal(t, "密码错误", r.Message)
	assert.Equal(t, map[string]any{"attempts": 3}, r.Data)
}

func TestFailWithData_NilDataExplicit(t *testing.T) {
	r := FailWithData(1001, "系统错误", nil)
	assert.Nil(t, r.Data)
}

// --- BadRequest ---

func TestBadRequest_CodeAndMessage(t *testing.T) {
	r := BadRequest("字段缺失")
	assert.Equal(t, CodeBadRequest, r.Code)
	assert.Equal(t, "字段缺失", r.Message)
	assert.Nil(t, r.Data)
}

func TestBadRequest_EmptyMessageNotDefaulted(t *testing.T) {
	// BadRequest 不做空消息兜底（区别于 Unauthorized 等）
	r := BadRequest("")
	assert.Equal(t, CodeBadRequest, r.Code)
	assert.Equal(t, "", r.Message)
}

// --- Unauthorized ---

func TestUnauthorized_CustomMessage(t *testing.T) {
	r := Unauthorized("token 已失效")
	assert.Equal(t, CodeUnauthorized, r.Code)
	assert.Equal(t, "token 已失效", r.Message)
	assert.Nil(t, r.Data)
}

func TestUnauthorized_EmptyMessageDefaultsTo未授权访问(t *testing.T) {
	r := Unauthorized("")
	assert.Equal(t, CodeUnauthorized, r.Code)
	assert.Equal(t, "未授权访问", r.Message, "空消息应兜底为 未授权访问")
}

// --- Forbidden ---

func TestForbidden_CustomMessage(t *testing.T) {
	r := Forbidden("无此权限")
	assert.Equal(t, CodeForbidden, r.Code)
	assert.Equal(t, "无此权限", r.Message)
}

func TestForbidden_EmptyMessageDefaultsTo禁止访问(t *testing.T) {
	r := Forbidden("")
	assert.Equal(t, CodeForbidden, r.Code)
	assert.Equal(t, "禁止访问", r.Message)
}

// --- NotFound ---

func TestNotFound_CustomMessage(t *testing.T) {
	r := NotFound("用户不存在")
	assert.Equal(t, CodeNotFound, r.Code)
	assert.Equal(t, "用户不存在", r.Message)
}

func TestNotFound_EmptyMessageDefaultsTo资源不存在(t *testing.T) {
	r := NotFound("")
	assert.Equal(t, CodeNotFound, r.Code)
	assert.Equal(t, "资源不存在", r.Message)
}

// --- ServerError ---

func TestServerError_CustomMessage(t *testing.T) {
	r := ServerError("数据库断开")
	assert.Equal(t, CodeServerError, r.Code)
	assert.Equal(t, "数据库断开", r.Message)
}

func TestServerError_EmptyMessageDefaultsTo服务器内部错误(t *testing.T) {
	r := ServerError("")
	assert.Equal(t, CodeServerError, r.Code)
	assert.Equal(t, "服务器内部错误", r.Message)
}

// --- HTTPStatus ---

func TestHTTPStatus_KnownCodes(t *testing.T) {
	cases := []struct {
		code int
		want int
	}{
		{CodeSuccess, http.StatusOK},
		{CodeBadRequest, http.StatusBadRequest},
		{CodeUnauthorized, http.StatusUnauthorized},
		{CodeForbidden, http.StatusForbidden},
		{CodeNotFound, http.StatusNotFound},
		{CodeServerError, http.StatusInternalServerError},
	}
	for _, c := range cases {
		if got := HTTPStatus(c.code); got != c.want {
			t.Errorf("HTTPStatus(%d) = %d, want %d", c.code, got, c.want)
		}
	}
}

func TestHTTPStatus_UnknownBusinessCodeFallsBackTo200(t *testing.T) {
	// 业务自定义错误码（如 1002/2004/2402...）统一返回 200，
	// 由前端按 body.code 判定（避免 HTTP 状态码与业务码语义重叠）
	codes := []int{1002, 1009, 2004, 2102, 2402, 4001, 9999}
	for _, c := range codes {
		if got := HTTPStatus(c); got != http.StatusOK {
			t.Errorf("HTTPStatus(未知业务码 %d) = %d, want 200", c, got)
		}
	}
}

func TestHTTPStatus_NegativeAndZeroNonSuccess(t *testing.T) {
	// 仅 0 命中 CodeSuccess 分支，其他负数/零区段未定义也走 default
	assert.Equal(t, http.StatusOK, HTTPStatus(0))
	// 注意：0 即 CodeSuccess，命中 OK 分支，与 default 路径结果一致
	// 用 -1 / 99 显式走 default 分支
	assert.Equal(t, http.StatusOK, HTTPStatus(-1))
	assert.Equal(t, http.StatusOK, HTTPStatus(99))
}

// --- NewPageResult ---

func TestNewPageResult_AllFieldsPassedThrough(t *testing.T) {
	list := []string{"a", "b", "c"}
	r := NewPageResult(list, int64(100), 2, 20)
	assert.Equal(t, list, r.List)
	assert.Equal(t, int64(100), r.Total)
	assert.Equal(t, 2, r.Page)
	assert.Equal(t, 20, r.PageSize)
}

func TestNewPageResult_ZeroValues(t *testing.T) {
	r := NewPageResult(nil, 0, 0, 0)
	assert.Nil(t, r.List)
	assert.Equal(t, int64(0), r.Total)
	assert.Equal(t, 0, r.Page)
	assert.Equal(t, 0, r.PageSize)
}

func TestNewPageResult_ReturnsPointer(t *testing.T) {
	r := NewPageResult(nil, 0, 0, 0)
	if r == nil {
		t.Fatal("NewPageResult 不应返回 nil 指针")
	}
}

// --- JSON 序列化协议（保证字段名/类型对外稳定） ---

func TestResponse_JSONShape(t *testing.T) {
	r := Success(map[string]int{"id": 1})
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("Marshal 失败: %v", err)
	}
	// 期望 {"code":0,"message":"success","data":{"id":1}}
	want := `{"code":0,"message":"success","data":{"id":1}}`
	assert.JSONEq(t, want, string(b))
}

func TestPageResult_JSONShape(t *testing.T) {
	r := NewPageResult([]int{1, 2}, int64(2), 1, 10)
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("Marshal 失败: %v", err)
	}
	want := `{"list":[1,2],"total":2,"page":1,"pageSize":10}`
	assert.JSONEq(t, want, string(b))
}

func TestResponse_FailJSONShapeDataNull(t *testing.T) {
	// Fail 的 data 应序列化为 null（不是省略字段）
	r := Fail(1002, "参数无效")
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("Marshal 失败: %v", err)
	}
	want := `{"code":1002,"message":"参数无效","data":null}`
	assert.JSONEq(t, want, string(b))
}

// --- 组合行为：构造器返回的 *Response 可被复用 ---

func TestBuilders_ReturnNonNilPointer(t *testing.T) {
	// 所有构造器返回的应是非空 *Response，避免调用方误判空指针
	assert.NotNil(t, Success(nil))
	assert.NotNil(t, SuccessWithMessage("ok", nil))
	assert.NotNil(t, Fail(0, ""))
	assert.NotNil(t, FailWithData(0, "", nil))
	assert.NotNil(t, BadRequest(""))
	assert.NotNil(t, Unauthorized(""))
	assert.NotNil(t, Forbidden(""))
	assert.NotNil(t, NotFound(""))
	assert.NotNil(t, ServerError(""))
}

func TestUnauthorized_Forbidden_NotFound_ServerError_AllInheritFailShape(t *testing.T) {
	// Unauthorized/Forbidden/NotFound/ServerError 本质是 Fail(code, msg) 的语义糖
	// 应保持 Code + Message + Data(nil) 三字段契约
	for _, r := range []*Response{
		Unauthorized(""),
		Forbidden(""),
		NotFound(""),
		ServerError(""),
	} {
		if r.Data != nil {
			t.Errorf("code=%d 的 Data 应为 nil，实际 %v", r.Code, r.Data)
		}
	}
}
