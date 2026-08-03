// Package handler_test AI 智能中台主 Handler HTTP 处理层单元测试。
//
// 使用 gin + httptest + mock Service，覆盖 handler 装配层全部分支：
//   - 任务接口（CreateTask/RunTask/GetTask/ListTasks）：Bind 校验、task_id 参数、regionID 注入、service 成功/错误透传（业务码 3101/3102）
//   - 模型管理（AddModel/ListModels/UpdateModelStatus）：oneof 校验、:id 解析失败 400、service 错误透传
//   - 提示词模板（AddPrompt/ListPrompts/RenderPrompt）：Bind 校验、渲染结果透传
//   - 高级 AI 接口（OptimizeTitle/GenerateDescription/SuggestPrice）：未登录 401 "请先登录"、Bind 校验、regionID+userID 注入、service 错误透传
//   - 生成记录（ListMyGenerations/RateGeneration）：未登录 401、rating min/max 校验、service 错误透传
//   - 分页结果 PageResult 结构断言（list/total/page/pageSize）
//
// 鉴权由 AuthRequired / RequirePermission 中间件负责（测试中去掉，纯测 handler 装配层）。
// 不依赖 DB/Redis/Docker，纯内存 mock service 验证 handler 装配层逻辑。
// 与 shop/category/region/news/file/setting/permission handler 测试同风格。
package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wuchang-tongcheng/internal/core/middleware"
	coreRouter "wuchang-tongcheng/internal/core/router"
	"wuchang-tongcheng/internal/modules/ai/dto"
	aiHandler "wuchang-tongcheng/internal/modules/ai/handler"
	"wuchang-tongcheng/internal/modules/ai/model"
	"wuchang-tongcheng/internal/modules/ai/service"
)

// apiResponse 解析统一响应体 {code, message, data}
type apiResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// pageData 解析 PageResult {list, total, page, pageSize}
type pageData struct {
	List     json.RawMessage `json:"list"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"pageSize"`
}

// mockAIService 内存 mock，实现 service.AIService 接口。
// 记录最近一次调用入参，返回预设 result/err，便于断言 handler 装配层逻辑。
type mockAIService struct {
	// CreateTask
	lastCreateTaskRegionID uint
	lastCreateTaskReq      *dto.CreateTaskRequest
	createTaskResult       *dto.TaskInfo
	createTaskErr          error

	// GetTask
	lastGetTaskID  string
	getTaskResult  *dto.TaskInfo
	getTaskErr     error

	// ListTasks
	lastListTasksReq  *dto.TaskListRequest
	listTasksResult   []dto.TaskInfo
	listTasksErr      error

	// RunTask
	lastRunTaskID  string
	runTaskResult  *dto.TaskInfo
	runTaskErr     error

	// AddModel
	lastAddModelReq *dto.ModelRequest
	addModelErr     error

	// ListModels
	lastListModelsProvider  string
	lastListModelsModelType string
	lastListModelsPage      int
	lastListModelsPageSize  int
	listModelsResult        []model.Model
	listModelsErr           error

	// UpdateModelStatus
	lastUpdateModelStatusID     uint
	lastUpdateModelStatusStatus int
	updateModelStatusErr        error

	// AddPrompt
	lastAddPromptReq *dto.PromptRequest
	addPromptErr     error

	// ListPrompts
	lastListPromptsType     string
	lastListPromptsPage     int
	lastListPromptsPageSize int
	listPromptsResult       []model.Prompt
	listPromptsErr          error

	// RenderPrompt
	lastRenderTemplateName string
	lastRenderVariables    map[string]interface{}
	renderResult           string
	renderErr              error

	// ListMyGenerations
	lastListGensUserID   uint
	lastListGensPage     int
	lastListGensPageSize int
	listGensResult       []dto.GenerationInfo
	listGensErr          error

	// RateGeneration
	lastRateUserID  uint
	lastRateReq     *dto.RateGenerationRequest
	rateErr         error

	// OptimizeTitle
	lastOptimizeRegionID uint
	lastOptimizeUserID   uint
	lastOptimizeReq      *dto.OptimizeTitleRequest
	optimizeResult       *dto.OptimizeTitleResponse
	optimizeErr          error

	// GenerateDescription
	lastGenDescRegionID uint
	lastGenDescUserID   uint
	lastGenDescReq      *dto.GenerateDescriptionRequest
	genDescResult       *dto.GenerateDescriptionResponse
	genDescErr          error

	// SuggestPrice
	lastSuggestRegionID uint
	lastSuggestUserID   uint
	lastSuggestReq      *dto.SuggestPriceRequest
	suggestResult       *dto.SuggestPriceResponse
	suggestErr          error
}

// ===== 任务 =====

func (m *mockAIService) CreateTask(regionID uint, req *dto.CreateTaskRequest) (*dto.TaskInfo, error) {
	m.lastCreateTaskRegionID = regionID
	m.lastCreateTaskReq = req
	return m.createTaskResult, m.createTaskErr
}

func (m *mockAIService) GetTask(taskID string) (*dto.TaskInfo, error) {
	m.lastGetTaskID = taskID
	return m.getTaskResult, m.getTaskErr
}

func (m *mockAIService) ListTasks(req *dto.TaskListRequest) ([]dto.TaskInfo, int64, error) {
	m.lastListTasksReq = req
	if m.listTasksErr != nil {
		return nil, 0, m.listTasksErr
	}
	return m.listTasksResult, int64(len(m.listTasksResult)), nil
}

func (m *mockAIService) RunTask(taskID string) (*dto.TaskInfo, error) {
	m.lastRunTaskID = taskID
	return m.runTaskResult, m.runTaskErr
}

// ===== 模型管理 =====

func (m *mockAIService) AddModel(req *dto.ModelRequest) error {
	m.lastAddModelReq = req
	return m.addModelErr
}

func (m *mockAIService) ListModels(provider, modelType string, page, pageSize int) ([]model.Model, int64, error) {
	m.lastListModelsProvider = provider
	m.lastListModelsModelType = modelType
	m.lastListModelsPage = page
	m.lastListModelsPageSize = pageSize
	if m.listModelsErr != nil {
		return nil, 0, m.listModelsErr
	}
	return m.listModelsResult, int64(len(m.listModelsResult)), nil
}

func (m *mockAIService) UpdateModelStatus(id uint, status int) error {
	m.lastUpdateModelStatusID = id
	m.lastUpdateModelStatusStatus = status
	return m.updateModelStatusErr
}

// ===== 提示词模板 =====

func (m *mockAIService) AddPrompt(req *dto.PromptRequest) error {
	m.lastAddPromptReq = req
	return m.addPromptErr
}

func (m *mockAIService) ListPrompts(templateType string, page, pageSize int) ([]model.Prompt, int64, error) {
	m.lastListPromptsType = templateType
	m.lastListPromptsPage = page
	m.lastListPromptsPageSize = pageSize
	if m.listPromptsErr != nil {
		return nil, 0, m.listPromptsErr
	}
	return m.listPromptsResult, int64(len(m.listPromptsResult)), nil
}

func (m *mockAIService) RenderPrompt(templateName string, variables map[string]interface{}) (string, error) {
	m.lastRenderTemplateName = templateName
	m.lastRenderVariables = variables
	return m.renderResult, m.renderErr
}

// ===== 生成记录 =====

func (m *mockAIService) ListMyGenerations(userID uint, page, pageSize int) ([]dto.GenerationInfo, int64, error) {
	m.lastListGensUserID = userID
	m.lastListGensPage = page
	m.lastListGensPageSize = pageSize
	if m.listGensErr != nil {
		return nil, 0, m.listGensErr
	}
	return m.listGensResult, int64(len(m.listGensResult)), nil
}

func (m *mockAIService) RateGeneration(userID uint, req *dto.RateGenerationRequest) error {
	m.lastRateUserID = userID
	m.lastRateReq = req
	return m.rateErr
}

// ===== 高级接口 =====

func (m *mockAIService) OptimizeTitle(regionID, userID uint, req *dto.OptimizeTitleRequest) (*dto.OptimizeTitleResponse, error) {
	m.lastOptimizeRegionID = regionID
	m.lastOptimizeUserID = userID
	m.lastOptimizeReq = req
	return m.optimizeResult, m.optimizeErr
}

func (m *mockAIService) GenerateDescription(regionID, userID uint, req *dto.GenerateDescriptionRequest) (*dto.GenerateDescriptionResponse, error) {
	m.lastGenDescRegionID = regionID
	m.lastGenDescUserID = userID
	m.lastGenDescReq = req
	return m.genDescResult, m.genDescErr
}

func (m *mockAIService) SuggestPrice(regionID, userID uint, req *dto.SuggestPriceRequest) (*dto.SuggestPriceResponse, error) {
	m.lastSuggestRegionID = regionID
	m.lastSuggestUserID = userID
	m.lastSuggestReq = req
	return m.suggestResult, m.suggestErr
}

// 确保 mockAIService 实现 service.AIService 接口
var _ service.AIService = (*mockAIService)(nil)

// handlerEnv handler 测试环境
type handlerEnv struct {
	engine *gin.Engine
	mock   *mockAIService
}

// newHandlerEnv 构造 gin 引擎并注册 ai 主 Handler 路由（路径与 ai/plugin.go RegisterRoutes 一致）。
// ctxUserID 模拟 Auth 中间件注入的 user_id（0 表示未登录）。
// regionID 模拟 Region 中间件注入的 region_id。
// 路由注册去掉 AuthRequired / RequirePermission 中间件，纯测 handler 装配层逻辑。
func newHandlerEnv(t *testing.T, ctxUserID uint, regionID uint) *handlerEnv {
	t.Helper()
	gin.SetMode(gin.TestMode)

	mock := &mockAIService{
		createTaskResult: &dto.TaskInfo{ID: 1, TaskID: "AI202608030000001", TaskType: "optimize_title", UserID: ctxUserID, Status: 0},
		getTaskResult:    &dto.TaskInfo{ID: 1, TaskID: "AI202608030000001", TaskType: "optimize_title", Status: 2},
		runTaskResult:    &dto.TaskInfo{ID: 1, TaskID: "AI202608030000001", TaskType: "optimize_title", Status: 2},
		listTasksResult: []dto.TaskInfo{
			{ID: 1, TaskID: "AI202608030000001", TaskType: "optimize_title", Status: 2},
			{ID: 2, TaskID: "AI202608030000002", TaskType: "suggest_price", Status: 2},
		},
		listModelsResult: []model.Model{
			{ID: 1, ModelName: "qwen-max", Provider: "qwen", ModelType: "llm", Status: 1},
		},
		listPromptsResult: []model.Prompt{
			{ID: 1, TemplateName: "optimize_title_default", TemplateType: "optimize_title", Content: "优化标题：{{title}}", Status: 1},
		},
		renderResult:     "优化标题：iPhone 13",
		listGensResult:   []dto.GenerationInfo{{ID: 1, TaskID: "AI202608030000001", UserID: ctxUserID, GenerationType: "title", Rating: 5}},
		optimizeResult:   &dto.OptimizeTitleResponse{OriginalTitle: "iPhone 13", Optimized: "【Apple】iPhone 13", Alternatives: []string{"iPhone 13（急转）"}, TaskID: "AI202608030000001"},
		genDescResult:    &dto.GenerateDescriptionResponse{Description: "【Apple】iPhone 13，9成新", Alternatives: []string{"私聊详询"}, TaskID: "AI202608030000002"},
		suggestResult:    &dto.SuggestPriceResponse{SuggestedPrice: 700, MinPrice: 595, MaxPrice: 805, Reason: "基于原价 1000 元", TaskID: "AI202608030000003"},
	}

	r := coreRouter.NewRouter()
	// 模拟 Region + Auth 中间件：注入 region_id 和 user_id
	r.Use(func(c *gin.Context) {
		c.Set(middleware.RegionIDKey, regionID)
		c.Set(middleware.ContextUserID, ctxUserID)
		c.Next()
	})

	h := aiHandler.NewHandler(mock)
	// 注册路由，路径与 ai/plugin.go RegisterRoutes 保持一致（去掉权限中间件，纯测 handler）
	root := r.Group("/api/v1/ai")
	// 任务
	root.POST("/tasks", h.CreateTask)
	root.POST("/tasks/:task_id/run", h.RunTask)
	root.GET("/tasks/:task_id", h.GetTask)
	root.GET("/tasks", h.ListTasks)
	// 模型管理
	root.POST("/models", h.AddModel)
	root.GET("/models", h.ListModels)
	root.POST("/models/:id/status", h.UpdateModelStatus)
	// 提示词模板
	root.POST("/prompts", h.AddPrompt)
	root.GET("/prompts", h.ListPrompts)
	root.POST("/prompts/render", h.RenderPrompt)
	// 高级 AI 接口
	root.POST("/optimize-title", h.OptimizeTitle)
	root.POST("/generate-description", h.GenerateDescription)
	root.POST("/suggest-price", h.SuggestPrice)
	// 生成记录
	root.GET("/generations", h.ListMyGenerations)
	root.POST("/generations/rate", h.RateGeneration)

	return &handlerEnv{engine: r.Engine(), mock: mock}
}

// doJSON 发起 JSON 请求，返回解析后的响应体。
func (e *handlerEnv) doJSON(t *testing.T, method, path string, body interface{}) *apiResponse {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		b, err := json.Marshal(body)
		require.NoError(t, err)
		buf.Write(b)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	e.engine.ServeHTTP(w, req)

	var resp apiResponse
	if w.Body.Len() > 0 {
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	}
	return &resp
}

// doRaw 发起原始请求（用于测试 Bind 失败：非法 JSON body）。
func (e *handlerEnv) doRaw(t *testing.T, method, path string, rawBody string, contentType string) *apiResponse {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(rawBody))
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()
	e.engine.ServeHTTP(w, req)

	var resp apiResponse
	if w.Body.Len() > 0 {
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	}
	return &resp
}

// parsePage 解析响应 data 为 pageData
func parsePage(t *testing.T, resp *apiResponse) *pageData {
	t.Helper()
	var p pageData
	require.NoError(t, json.Unmarshal(resp.Data, &p))
	return &p
}

// assertParamError 断言 Bind 失败响应（消息以 "参数错误" 开头）
func assertParamError(t *testing.T, resp *apiResponse) {
	t.Helper()
	assert.Equal(t, 400, resp.Code)
	assert.True(t, strings.HasPrefix(resp.Message, "参数错误"), "expected message start with 参数错误, got: %s", resp.Message)
}

// ==================== 任务接口 ====================

// ---------- CreateTask ----------

func TestHandler_CreateTask_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 2)
	body := dto.CreateTaskRequest{
		TaskType:  "optimize_title",
		Input:     map[string]interface{}{"title": "iPhone 13"},
		ModelName: "qwen-max",
	}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/ai/tasks", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "任务创建成功", resp.Message)
	assert.Equal(t, uint(2), env.mock.lastCreateTaskRegionID)
	require.NotNil(t, env.mock.lastCreateTaskReq)
	assert.Equal(t, "optimize_title", env.mock.lastCreateTaskReq.TaskType)
	assert.Equal(t, "qwen-max", env.mock.lastCreateTaskReq.ModelName)
	// UserID 为空时由 handler 用登录用户填充
	assert.Equal(t, uint(7), env.mock.lastCreateTaskReq.UserID)
	var info dto.TaskInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, "AI202608030000001", info.TaskID)
}

func TestHandler_CreateTask_UserIDPreserved(t *testing.T) {
	// 请求体显式带 user_id 时不被覆盖
	env := newHandlerEnv(t, 7, 2)
	body := dto.CreateTaskRequest{
		TaskType: "suggest_price",
		UserID:   99,
		Input:    map[string]interface{}{"title": "x"},
	}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/ai/tasks", body)

	assert.Equal(t, 0, resp.Code)
	require.NotNil(t, env.mock.lastCreateTaskReq)
	assert.Equal(t, uint(99), env.mock.lastCreateTaskReq.UserID)
}

func TestHandler_CreateTask_BindError_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 7, 2)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/ai/tasks", "{not json", "application/json")

	assertParamError(t, resp)
	assert.Nil(t, env.mock.lastCreateTaskReq)
}

func TestHandler_CreateTask_BindError_MissingTaskType(t *testing.T) {
	env := newHandlerEnv(t, 7, 2)
	// 缺少 task_type（required）+ input（required）
	resp := env.doRaw(t, http.MethodPost, "/api/v1/ai/tasks", `{"model_name":"x"}`, "application/json")

	assertParamError(t, resp)
	assert.Nil(t, env.mock.lastCreateTaskReq)
}

func TestHandler_CreateTask_BindError_InvalidTaskType(t *testing.T) {
	env := newHandlerEnv(t, 7, 2)
	// task_type=foo 不满足 oneof
	resp := env.doRaw(t, http.MethodPost, "/api/v1/ai/tasks", `{"task_type":"foo","input":{"x":1}}`, "application/json")

	assertParamError(t, resp)
	assert.Nil(t, env.mock.lastCreateTaskReq)
}

func TestHandler_CreateTask_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 2)
	env.mock.createTaskResult = nil
	env.mock.createTaskErr = errors.New("模型已禁用")
	body := dto.CreateTaskRequest{TaskType: "optimize_title", Input: map[string]interface{}{"title": "x"}}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/ai/tasks", body)

	// 3101
	assert.Equal(t, 3101, resp.Code)
	assert.Equal(t, "模型已禁用", resp.Message)
	assert.Equal(t, uint(2), env.mock.lastCreateTaskRegionID)
}

// ---------- RunTask ----------

func TestHandler_RunTask_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 2)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/ai/tasks/AI202608030000001/run", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "AI202608030000001", env.mock.lastRunTaskID)
	var info dto.TaskInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, 2, info.Status)
}

func TestHandler_RunTask_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 2)
	env.mock.runTaskResult = nil
	env.mock.runTaskErr = errors.New("AI任务不存在")
	resp := env.doJSON(t, http.MethodPost, "/api/v1/ai/tasks/AI999/run", nil)

	// 3101
	assert.Equal(t, 3101, resp.Code)
	assert.Equal(t, "AI任务不存在", resp.Message)
	assert.Equal(t, "AI999", env.mock.lastRunTaskID)
}

// ---------- GetTask ----------

func TestHandler_GetTask_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 2)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/ai/tasks/AI202608030000001", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "success", resp.Message)
	assert.Equal(t, "AI202608030000001", env.mock.lastGetTaskID)
	var info dto.TaskInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, "optimize_title", info.TaskType)
}

func TestHandler_GetTask_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 2)
	env.mock.getTaskResult = nil
	env.mock.getTaskErr = errors.New("AI任务不存在")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/ai/tasks/AI999", nil)

	// 3102
	assert.Equal(t, 3102, resp.Code)
	assert.Equal(t, "AI任务不存在", resp.Message)
}

// ---------- ListTasks ----------

func TestHandler_ListTasks_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 2)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/ai/tasks?page=2&page_size=15&task_type=optimize_title&status=2", nil)

	assert.Equal(t, 0, resp.Code)
	require.NotNil(t, env.mock.lastListTasksReq)
	assert.Equal(t, uint(7), env.mock.lastListTasksReq.UserID)
	assert.Equal(t, "optimize_title", env.mock.lastListTasksReq.TaskType)
	assert.Equal(t, 2, env.mock.lastListTasksReq.Status)
	assert.Equal(t, 2, env.mock.lastListTasksReq.Page)
	assert.Equal(t, 15, env.mock.lastListTasksReq.PageSize)
	p := parsePage(t, resp)
	assert.Equal(t, int64(2), p.Total)
	var list []dto.TaskInfo
	require.NoError(t, json.Unmarshal(p.List, &list))
	require.Len(t, list, 2)
	assert.Equal(t, "optimize_title", list[0].TaskType)
}

func TestHandler_ListTasks_DefaultPagination(t *testing.T) {
	env := newHandlerEnv(t, 7, 2)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/ai/tasks", nil)

	assert.Equal(t, 0, resp.Code)
	require.NotNil(t, env.mock.lastListTasksReq)
	// 默认 page=1, page_size=20, status=-1
	assert.Equal(t, 1, env.mock.lastListTasksReq.Page)
	assert.Equal(t, 20, env.mock.lastListTasksReq.PageSize)
	assert.Equal(t, -1, env.mock.lastListTasksReq.Status)
}

func TestHandler_ListTasks_InvalidPaginationFallback(t *testing.T) {
	env := newHandlerEnv(t, 7, 2)
	// page=0, page_size=0 → 回退默认值
	resp := env.doJSON(t, http.MethodGet, "/api/v1/ai/tasks?page=0&page_size=0", nil)

	assert.Equal(t, 0, resp.Code)
	require.NotNil(t, env.mock.lastListTasksReq)
	assert.Equal(t, 1, env.mock.lastListTasksReq.Page)
	assert.Equal(t, 20, env.mock.lastListTasksReq.PageSize)
}

func TestHandler_ListTasks_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 2)
	env.mock.listTasksResult = nil
	env.mock.listTasksErr = errors.New("db down")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/ai/tasks", nil)

	// 3101
	assert.Equal(t, 3101, resp.Code)
	assert.Equal(t, "db down", resp.Message)
}

// ==================== 模型管理 ====================

// ---------- AddModel ----------

func TestHandler_AddModel_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	body := dto.ModelRequest{
		ModelName: "qwen-max",
		Provider:  "qwen",
		ModelType: "llm",
		APIKey:    "sk-xxx",
		Endpoint:  "https://dashscope.aliyuncs.com",
	}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/ai/models", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "添加成功", resp.Message)
	require.NotNil(t, env.mock.lastAddModelReq)
	assert.Equal(t, "qwen-max", env.mock.lastAddModelReq.ModelName)
	assert.Equal(t, "qwen", env.mock.lastAddModelReq.Provider)
	assert.Equal(t, "llm", env.mock.lastAddModelReq.ModelType)
}

func TestHandler_AddModel_BindError_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/ai/models", "{bad", "application/json")

	assertParamError(t, resp)
	assert.Nil(t, env.mock.lastAddModelReq)
}

func TestHandler_AddModel_BindError_MissingRequired(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	// 缺少 model_name/provider/model_type
	resp := env.doRaw(t, http.MethodPost, "/api/v1/ai/models", `{"api_key":"x"}`, "application/json")

	assertParamError(t, resp)
	assert.Nil(t, env.mock.lastAddModelReq)
}

func TestHandler_AddModel_BindError_InvalidProvider(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	// provider=foo 不满足 oneof
	resp := env.doRaw(t, http.MethodPost, "/api/v1/ai/models", `{"model_name":"x","provider":"foo","model_type":"llm"}`, "application/json")

	assertParamError(t, resp)
	assert.Nil(t, env.mock.lastAddModelReq)
}

func TestHandler_AddModel_BindError_InvalidModelType(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	// model_type=foo 不满足 oneof
	resp := env.doRaw(t, http.MethodPost, "/api/v1/ai/models", `{"model_name":"x","provider":"qwen","model_type":"foo"}`, "application/json")

	assertParamError(t, resp)
	assert.Nil(t, env.mock.lastAddModelReq)
}

func TestHandler_AddModel_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	env.mock.addModelErr = errors.New("模型已存在")
	body := dto.ModelRequest{ModelName: "qwen-max", Provider: "qwen", ModelType: "llm"}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/ai/models", body)

	// 3101
	assert.Equal(t, 3101, resp.Code)
	assert.Equal(t, "模型已存在", resp.Message)
}

// ---------- ListModels ----------

func TestHandler_ListModels_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/ai/models?provider=qwen&model_type=llm&page=1&page_size=10", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "qwen", env.mock.lastListModelsProvider)
	assert.Equal(t, "llm", env.mock.lastListModelsModelType)
	assert.Equal(t, 1, env.mock.lastListModelsPage)
	assert.Equal(t, 10, env.mock.lastListModelsPageSize)
	p := parsePage(t, resp)
	assert.Equal(t, int64(1), p.Total)
	var list []model.Model
	require.NoError(t, json.Unmarshal(p.List, &list))
	require.Len(t, list, 1)
	assert.Equal(t, "qwen-max", list[0].ModelName)
}

func TestHandler_ListModels_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	env.mock.listModelsResult = nil
	env.mock.listModelsErr = errors.New("db error")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/ai/models", nil)

	// 3101
	assert.Equal(t, 3101, resp.Code)
	assert.Equal(t, "db error", resp.Message)
}

// ---------- UpdateModelStatus ----------

func TestHandler_UpdateModelStatus_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/ai/models/5/status?status=0", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "更新成功", resp.Message)
	assert.Equal(t, uint(5), env.mock.lastUpdateModelStatusID)
	assert.Equal(t, 0, env.mock.lastUpdateModelStatusStatus)
}

func TestHandler_UpdateModelStatus_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/ai/models/abc/status?status=0", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "id 不能为空", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastUpdateModelStatusID)
}

func TestHandler_UpdateModelStatus_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	env.mock.updateModelStatusErr = errors.New("AI模型不存在")
	resp := env.doJSON(t, http.MethodPost, "/api/v1/ai/models/5/status?status=1", nil)

	// 3101
	assert.Equal(t, 3101, resp.Code)
	assert.Equal(t, "AI模型不存在", resp.Message)
}

// ==================== 提示词模板 ====================

// ---------- AddPrompt ----------

func TestHandler_AddPrompt_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	body := dto.PromptRequest{
		TemplateName: "optimize_title_default",
		TemplateType: "optimize_title",
		Content:      "优化标题：{{title}}",
		Variables:    []string{"title"},
	}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/ai/prompts", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "添加成功", resp.Message)
	require.NotNil(t, env.mock.lastAddPromptReq)
	assert.Equal(t, "optimize_title_default", env.mock.lastAddPromptReq.TemplateName)
	assert.Equal(t, "optimize_title", env.mock.lastAddPromptReq.TemplateType)
}

func TestHandler_AddPrompt_BindError_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/ai/prompts", "{bad", "application/json")

	assertParamError(t, resp)
	assert.Nil(t, env.mock.lastAddPromptReq)
}

func TestHandler_AddPrompt_BindError_MissingContent(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	// 缺少 content（required）+ template_name + template_type
	resp := env.doRaw(t, http.MethodPost, "/api/v1/ai/prompts", `{"description":"x"}`, "application/json")

	assertParamError(t, resp)
	assert.Nil(t, env.mock.lastAddPromptReq)
}

func TestHandler_AddPrompt_BindError_InvalidTemplateType(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	// template_type=foo 不满足 oneof
	resp := env.doRaw(t, http.MethodPost, "/api/v1/ai/prompts", `{"template_name":"x","template_type":"foo","content":"c"}`, "application/json")

	assertParamError(t, resp)
	assert.Nil(t, env.mock.lastAddPromptReq)
}

func TestHandler_AddPrompt_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	env.mock.addPromptErr = errors.New("模板已存在")
	body := dto.PromptRequest{TemplateName: "x", TemplateType: "optimize_title", Content: "c"}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/ai/prompts", body)

	// 3101
	assert.Equal(t, 3101, resp.Code)
	assert.Equal(t, "模板已存在", resp.Message)
}

// ---------- ListPrompts ----------

func TestHandler_ListPrompts_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/ai/prompts?template_type=optimize_title&page=1&page_size=10", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "optimize_title", env.mock.lastListPromptsType)
	assert.Equal(t, 1, env.mock.lastListPromptsPage)
	assert.Equal(t, 10, env.mock.lastListPromptsPageSize)
	p := parsePage(t, resp)
	assert.Equal(t, int64(1), p.Total)
	var list []model.Prompt
	require.NoError(t, json.Unmarshal(p.List, &list))
	require.Len(t, list, 1)
	assert.Equal(t, "optimize_title_default", list[0].TemplateName)
}

func TestHandler_ListPrompts_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	env.mock.listPromptsResult = nil
	env.mock.listPromptsErr = errors.New("db error")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/ai/prompts", nil)

	// 3101
	assert.Equal(t, 3101, resp.Code)
	assert.Equal(t, "db error", resp.Message)
}

// ---------- RenderPrompt ----------

func TestHandler_RenderPrompt_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	body := dto.RenderPromptRequest{
		TemplateName: "optimize_title_default",
		Variables:    map[string]interface{}{"title": "iPhone 13"},
	}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/ai/prompts/render", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "optimize_title_default", env.mock.lastRenderTemplateName)
	require.NotNil(t, env.mock.lastRenderVariables)
	assert.Equal(t, "iPhone 13", env.mock.lastRenderVariables["title"])
	var out dto.RenderPromptResponse
	require.NoError(t, json.Unmarshal(resp.Data, &out))
	assert.Equal(t, "优化标题：iPhone 13", out.Content)
}

func TestHandler_RenderPrompt_BindError_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/ai/prompts/render", "{bad", "application/json")

	assertParamError(t, resp)
	assert.Equal(t, "", env.mock.lastRenderTemplateName)
}

func TestHandler_RenderPrompt_BindError_MissingRequired(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	// 缺少 template_name（required）+ variables（required）
	resp := env.doRaw(t, http.MethodPost, "/api/v1/ai/prompts/render", `{}`, "application/json")

	assertParamError(t, resp)
	assert.Equal(t, "", env.mock.lastRenderTemplateName)
}

func TestHandler_RenderPrompt_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	env.mock.renderErr = errors.New("提示词模板不存在")
	body := dto.RenderPromptRequest{TemplateName: "missing", Variables: map[string]interface{}{"x": 1}}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/ai/prompts/render", body)

	// 3101
	assert.Equal(t, 3101, resp.Code)
	assert.Equal(t, "提示词模板不存在", resp.Message)
}

// ==================== 高级 AI 接口 ====================

// ---------- OptimizeTitle ----------

func TestHandler_OptimizeTitle_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 9)
	body := dto.OptimizeTitleRequest{Title: "iPhone 13", Category: "手机", Brand: "Apple"}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/ai/optimize-title", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(9), env.mock.lastOptimizeRegionID)
	assert.Equal(t, uint(7), env.mock.lastOptimizeUserID)
	require.NotNil(t, env.mock.lastOptimizeReq)
	assert.Equal(t, "iPhone 13", env.mock.lastOptimizeReq.Title)
	var out dto.OptimizeTitleResponse
	require.NoError(t, json.Unmarshal(resp.Data, &out))
	assert.Equal(t, "【Apple】iPhone 13", out.Optimized)
	require.Len(t, out.Alternatives, 1)
}

func TestHandler_OptimizeTitle_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 9)
	body := dto.OptimizeTitleRequest{Title: "iPhone 13"}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/ai/optimize-title", body)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Nil(t, env.mock.lastOptimizeReq)
}

func TestHandler_OptimizeTitle_BindError_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 7, 9)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/ai/optimize-title", "{bad", "application/json")

	assertParamError(t, resp)
	assert.Nil(t, env.mock.lastOptimizeReq)
}

func TestHandler_OptimizeTitle_BindError_MissingTitle(t *testing.T) {
	env := newHandlerEnv(t, 7, 9)
	// 缺少 title（required）
	resp := env.doRaw(t, http.MethodPost, "/api/v1/ai/optimize-title", `{"brand":"Apple"}`, "application/json")

	assertParamError(t, resp)
	assert.Nil(t, env.mock.lastOptimizeReq)
}

func TestHandler_OptimizeTitle_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 9)
	env.mock.optimizeResult = nil
	env.mock.optimizeErr = errors.New("模型已禁用")
	body := dto.OptimizeTitleRequest{Title: "iPhone 13"}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/ai/optimize-title", body)

	// 3101
	assert.Equal(t, 3101, resp.Code)
	assert.Equal(t, "模型已禁用", resp.Message)
	assert.Equal(t, uint(9), env.mock.lastOptimizeRegionID)
}

// ---------- GenerateDescription ----------

func TestHandler_GenerateDescription_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 9)
	body := dto.GenerateDescriptionRequest{Title: "iPhone 13", Brand: "Apple", Condition: "9成新", Keywords: []string{"便宜"}}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/ai/generate-description", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(9), env.mock.lastGenDescRegionID)
	assert.Equal(t, uint(7), env.mock.lastGenDescUserID)
	require.NotNil(t, env.mock.lastGenDescReq)
	assert.Equal(t, "iPhone 13", env.mock.lastGenDescReq.Title)
	var out dto.GenerateDescriptionResponse
	require.NoError(t, json.Unmarshal(resp.Data, &out))
	assert.Contains(t, out.Description, "iPhone 13")
}

func TestHandler_GenerateDescription_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 9)
	body := dto.GenerateDescriptionRequest{Title: "iPhone 13"}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/ai/generate-description", body)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Nil(t, env.mock.lastGenDescReq)
}

func TestHandler_GenerateDescription_BindError_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 7, 9)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/ai/generate-description", "{bad", "application/json")

	assertParamError(t, resp)
	assert.Nil(t, env.mock.lastGenDescReq)
}

func TestHandler_GenerateDescription_BindError_MissingTitle(t *testing.T) {
	env := newHandlerEnv(t, 7, 9)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/ai/generate-description", `{"brand":"Apple"}`, "application/json")

	assertParamError(t, resp)
	assert.Nil(t, env.mock.lastGenDescReq)
}

func TestHandler_GenerateDescription_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 9)
	env.mock.genDescResult = nil
	env.mock.genDescErr = errors.New("模型已禁用")
	body := dto.GenerateDescriptionRequest{Title: "iPhone 13"}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/ai/generate-description", body)

	// 3101
	assert.Equal(t, 3101, resp.Code)
	assert.Equal(t, "模型已禁用", resp.Message)
}

// ---------- SuggestPrice ----------

func TestHandler_SuggestPrice_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 9)
	body := dto.SuggestPriceRequest{Title: "iPhone 13", Brand: "Apple", Condition: "9成新", OriginalPrice: 1000}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/ai/suggest-price", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(9), env.mock.lastSuggestRegionID)
	assert.Equal(t, uint(7), env.mock.lastSuggestUserID)
	require.NotNil(t, env.mock.lastSuggestReq)
	assert.Equal(t, "iPhone 13", env.mock.lastSuggestReq.Title)
	assert.Equal(t, 1000.0, env.mock.lastSuggestReq.OriginalPrice)
	var out dto.SuggestPriceResponse
	require.NoError(t, json.Unmarshal(resp.Data, &out))
	assert.Equal(t, 700.0, out.SuggestedPrice)
	assert.Equal(t, "AI202608030000003", out.TaskID)
}

func TestHandler_SuggestPrice_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 9)
	body := dto.SuggestPriceRequest{Title: "iPhone 13"}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/ai/suggest-price", body)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Nil(t, env.mock.lastSuggestReq)
}

func TestHandler_SuggestPrice_BindError_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 7, 9)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/ai/suggest-price", "{bad", "application/json")

	assertParamError(t, resp)
	assert.Nil(t, env.mock.lastSuggestReq)
}

func TestHandler_SuggestPrice_BindError_MissingTitle(t *testing.T) {
	env := newHandlerEnv(t, 7, 9)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/ai/suggest-price", `{"brand":"Apple"}`, "application/json")

	assertParamError(t, resp)
	assert.Nil(t, env.mock.lastSuggestReq)
}

func TestHandler_SuggestPrice_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 9)
	env.mock.suggestResult = nil
	env.mock.suggestErr = errors.New("模型已禁用")
	body := dto.SuggestPriceRequest{Title: "iPhone 13"}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/ai/suggest-price", body)

	// 3101
	assert.Equal(t, 3101, resp.Code)
	assert.Equal(t, "模型已禁用", resp.Message)
}

// ==================== 生成记录 ====================

// ---------- ListMyGenerations ----------

func TestHandler_ListMyGenerations_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 2)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/ai/generations?page=1&page_size=10", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(7), env.mock.lastListGensUserID)
	assert.Equal(t, 1, env.mock.lastListGensPage)
	assert.Equal(t, 10, env.mock.lastListGensPageSize)
	p := parsePage(t, resp)
	assert.Equal(t, int64(1), p.Total)
	var list []dto.GenerationInfo
	require.NoError(t, json.Unmarshal(p.List, &list))
	require.Len(t, list, 1)
	assert.Equal(t, "title", list[0].GenerationType)
}

func TestHandler_ListMyGenerations_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 2)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/ai/generations", nil)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastListGensUserID)
}

func TestHandler_ListMyGenerations_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 2)
	env.mock.listGensResult = nil
	env.mock.listGensErr = errors.New("db error")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/ai/generations", nil)

	// 3101
	assert.Equal(t, 3101, resp.Code)
	assert.Equal(t, "db error", resp.Message)
}

// ---------- RateGeneration ----------

func TestHandler_RateGeneration_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 2)
	body := dto.RateGenerationRequest{GenerationID: 10, Rating: 5, Feedback: "很满意"}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/ai/generations/rate", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "评分成功", resp.Message)
	assert.Equal(t, uint(7), env.mock.lastRateUserID)
	require.NotNil(t, env.mock.lastRateReq)
	assert.Equal(t, uint(10), env.mock.lastRateReq.GenerationID)
	assert.Equal(t, 5, env.mock.lastRateReq.Rating)
	assert.Equal(t, "很满意", env.mock.lastRateReq.Feedback)
}

func TestHandler_RateGeneration_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 2)
	body := dto.RateGenerationRequest{GenerationID: 10, Rating: 5}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/ai/generations/rate", body)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Nil(t, env.mock.lastRateReq)
}

func TestHandler_RateGeneration_BindError_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 7, 2)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/ai/generations/rate", "{bad", "application/json")

	assertParamError(t, resp)
	assert.Nil(t, env.mock.lastRateReq)
}

func TestHandler_RateGeneration_BindError_MissingGenerationID(t *testing.T) {
	env := newHandlerEnv(t, 7, 2)
	// 缺少 generation_id（required）
	resp := env.doRaw(t, http.MethodPost, "/api/v1/ai/generations/rate", `{"rating":5}`, "application/json")

	assertParamError(t, resp)
	assert.Nil(t, env.mock.lastRateReq)
}

func TestHandler_RateGeneration_BindError_RatingOutOfRange(t *testing.T) {
	env := newHandlerEnv(t, 7, 2)
	// rating=9 不满足 min=1,max=5
	resp := env.doRaw(t, http.MethodPost, "/api/v1/ai/generations/rate", `{"generation_id":10,"rating":9}`, "application/json")

	assertParamError(t, resp)
	assert.Nil(t, env.mock.lastRateReq)
}

func TestHandler_RateGeneration_BindError_RatingZero(t *testing.T) {
	env := newHandlerEnv(t, 7, 2)
	// rating=0 不满足 required,min=1
	resp := env.doRaw(t, http.MethodPost, "/api/v1/ai/generations/rate", `{"generation_id":10,"rating":0}`, "application/json")

	assertParamError(t, resp)
	assert.Nil(t, env.mock.lastRateReq)
}

func TestHandler_RateGeneration_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 2)
	env.mock.rateErr = errors.New("无权评价他人生成记录")
	body := dto.RateGenerationRequest{GenerationID: 10, Rating: 4}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/ai/generations/rate", body)

	// 3101
	assert.Equal(t, 3101, resp.Code)
	assert.Equal(t, "无权评价他人生成记录", resp.Message)
}

// ==================== regionID 注入聚合 ====================

func TestHandler_RegionIDInjection_Aggregate(t *testing.T) {
	// 验证所有接收 regionID 的接口均透传 context 中的 regionID
	env := newHandlerEnv(t, 1, 9)

	// CreateTask
	env.doJSON(t, http.MethodPost, "/api/v1/ai/tasks", dto.CreateTaskRequest{TaskType: "optimize_title", Input: map[string]interface{}{"title": "x"}})
	assert.Equal(t, uint(9), env.mock.lastCreateTaskRegionID)

	// OptimizeTitle
	env.doJSON(t, http.MethodPost, "/api/v1/ai/optimize-title", dto.OptimizeTitleRequest{Title: "x"})
	assert.Equal(t, uint(9), env.mock.lastOptimizeRegionID)

	// GenerateDescription
	env.doJSON(t, http.MethodPost, "/api/v1/ai/generate-description", dto.GenerateDescriptionRequest{Title: "x"})
	assert.Equal(t, uint(9), env.mock.lastGenDescRegionID)

	// SuggestPrice
	env.doJSON(t, http.MethodPost, "/api/v1/ai/suggest-price", dto.SuggestPriceRequest{Title: "x"})
	assert.Equal(t, uint(9), env.mock.lastSuggestRegionID)
}

// ==================== userID 注入聚合 ====================

func TestHandler_UserIDInjection_Aggregate(t *testing.T) {
	// 验证所有接收 userID 的接口均透传 context 中的 userID
	env := newHandlerEnv(t, 8, 2)

	// ListTasks 用登录用户作为过滤条件
	env.doJSON(t, http.MethodGet, "/api/v1/ai/tasks", nil)
	require.NotNil(t, env.mock.lastListTasksReq)
	assert.Equal(t, uint(8), env.mock.lastListTasksReq.UserID)

	// OptimizeTitle / GenerateDescription / SuggestPrice / ListMyGenerations / RateGeneration
	env.doJSON(t, http.MethodPost, "/api/v1/ai/optimize-title", dto.OptimizeTitleRequest{Title: "x"})
	assert.Equal(t, uint(8), env.mock.lastOptimizeUserID)

	env.doJSON(t, http.MethodPost, "/api/v1/ai/generate-description", dto.GenerateDescriptionRequest{Title: "x"})
	assert.Equal(t, uint(8), env.mock.lastGenDescUserID)

	env.doJSON(t, http.MethodPost, "/api/v1/ai/suggest-price", dto.SuggestPriceRequest{Title: "x"})
	assert.Equal(t, uint(8), env.mock.lastSuggestUserID)

	env.doJSON(t, http.MethodGet, "/api/v1/ai/generations", nil)
	assert.Equal(t, uint(8), env.mock.lastListGensUserID)

	env.doJSON(t, http.MethodPost, "/api/v1/ai/generations/rate", dto.RateGenerationRequest{GenerationID: 1, Rating: 5})
	assert.Equal(t, uint(8), env.mock.lastRateUserID)
}
