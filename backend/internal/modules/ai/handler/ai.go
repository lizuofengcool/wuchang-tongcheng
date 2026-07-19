// Package handler AI 智能中台精简版HTTP处理层
package handler

import (
	"net/http"
	"strconv"

	"wuchang-tongcheng/internal/core/middleware"
	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/ai/dto"
	"wuchang-tongcheng/internal/modules/ai/service"
)

// Handler AI 中台 HTTP 处理器
type Handler struct {
	svc service.AIService
}

// NewHandler 创建 Handler 实例
func NewHandler(svc service.AIService) *Handler {
	return &Handler{svc: svc}
}

// getUserID 从上下文获取登录用户ID
func getUserID(ctx plugin.Context) uint {
	if v, ok := ctx.Get(middleware.ContextUserID); ok {
		if id, ok := v.(uint); ok {
			return id
		}
	}
	return 0
}

// getRegionID 从上下文获取地区ID
func getRegionID(ctx plugin.Context) uint {
	if v, ok := ctx.Get(middleware.RegionIDKey); ok {
		if id, ok := v.(uint); ok {
			return id
		}
	}
	return middleware.DefaultRegionID
}

// parsePagination 解析分页
func parsePagination(ctx plugin.Context) (page, pageSize int) {
	page, _ = strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ = strconv.Atoi(ctx.DefaultQuery("page_size", "20"))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	return
}

// CreateTask 创建 AI 任务
// POST /api/v1/ai/tasks
func (h *Handler) CreateTask(ctx plugin.Context) {
	userID := getUserID(ctx)
	var req dto.CreateTaskRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if req.UserID == 0 {
		req.UserID = userID
	}
	info, err := h.svc.CreateTask(getRegionID(ctx), &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(3101, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("任务创建成功", info))
}

// RunTask 执行任务
// POST /api/v1/ai/tasks/:task_id/run
func (h *Handler) RunTask(ctx plugin.Context) {
	taskID := ctx.Param("task_id")
	if taskID == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("task_id 不能为空"))
		return
	}
	info, err := h.svc.RunTask(taskID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(3101, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// GetTask 查询任务
// GET /api/v1/ai/tasks/:task_id
func (h *Handler) GetTask(ctx plugin.Context) {
	taskID := ctx.Param("task_id")
	if taskID == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("task_id 不能为空"))
		return
	}
	info, err := h.svc.GetTask(taskID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(3102, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// ListTasks 任务列表
// GET /api/v1/ai/tasks
func (h *Handler) ListTasks(ctx plugin.Context) {
	userID := getUserID(ctx)
	page, pageSize := parsePagination(ctx)
	status, _ := strconv.Atoi(ctx.DefaultQuery("status", "-1"))
	req := &dto.TaskListRequest{
		UserID:   userID,
		TaskType: ctx.Query("task_type"),
		Status:   status,
		Page:     page,
		PageSize: pageSize,
	}
	list, total, err := h.svc.ListTasks(req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(3101, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(response.NewPageResult(list, total, page, pageSize)))
}

// AddModel 添加模型（M 端）
// POST /api/v1/ai/models
func (h *Handler) AddModel(ctx plugin.Context) {
	var req dto.ModelRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.AddModel(&req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(3101, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("添加成功", nil))
}

// ListModels 模型列表
// GET /api/v1/ai/models
func (h *Handler) ListModels(ctx plugin.Context) {
	page, pageSize := parsePagination(ctx)
	provider := ctx.Query("provider")
	modelType := ctx.Query("model_type")
	list, total, err := h.svc.ListModels(provider, modelType, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(3101, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(response.NewPageResult(list, total, page, pageSize)))
}

// UpdateModelStatus 更新模型状态（M 端）
// POST /api/v1/ai/models/:id/status
func (h *Handler) UpdateModelStatus(ctx plugin.Context) {
	idStr := ctx.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)
	statusStr := ctx.Query("status")
	status, _ := strconv.Atoi(statusStr)
	if id == 0 {
		ctx.JSON(http.StatusOK, response.BadRequest("id 不能为空"))
		return
	}
	if err := h.svc.UpdateModelStatus(uint(id), status); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(3101, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("更新成功", nil))
}

// AddPrompt 添加提示词模板（M 端）
// POST /api/v1/ai/prompts
func (h *Handler) AddPrompt(ctx plugin.Context) {
	var req dto.PromptRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.AddPrompt(&req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(3101, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("添加成功", nil))
}

// ListPrompts 模板列表
// GET /api/v1/ai/prompts
func (h *Handler) ListPrompts(ctx plugin.Context) {
	page, pageSize := parsePagination(ctx)
	templateType := ctx.Query("template_type")
	list, total, err := h.svc.ListPrompts(templateType, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(3101, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(response.NewPageResult(list, total, page, pageSize)))
}

// RenderPrompt 渲染提示词
// POST /api/v1/ai/prompts/render
func (h *Handler) RenderPrompt(ctx plugin.Context) {
	var req dto.RenderPromptRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	content, err := h.svc.RenderPrompt(req.TemplateName, req.Variables)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(3101, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(dto.RenderPromptResponse{Content: content}))
}

// OptimizeTitle 标题优化
// POST /api/v1/ai/optimize-title
func (h *Handler) OptimizeTitle(ctx plugin.Context) {
	userID := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.OptimizeTitleRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	resp, err := h.svc.OptimizeTitle(getRegionID(ctx), userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(3101, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}

// GenerateDescription 描述生成
// POST /api/v1/ai/generate-description
func (h *Handler) GenerateDescription(ctx plugin.Context) {
	userID := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.GenerateDescriptionRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	resp, err := h.svc.GenerateDescription(getRegionID(ctx), userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(3101, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}

// SuggestPrice 价格建议
// POST /api/v1/ai/suggest-price
func (h *Handler) SuggestPrice(ctx plugin.Context) {
	userID := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.SuggestPriceRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	resp, err := h.svc.SuggestPrice(getRegionID(ctx), userID, &req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(3101, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}

// ListMyGenerations 我的生成记录
// GET /api/v1/ai/generations
func (h *Handler) ListMyGenerations(ctx plugin.Context) {
	userID := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	page, pageSize := parsePagination(ctx)
	list, total, err := h.svc.ListMyGenerations(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(3101, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(response.NewPageResult(list, total, page, pageSize)))
}

// RateGeneration 评分
// POST /api/v1/ai/generations/rate
func (h *Handler) RateGeneration(ctx plugin.Context) {
	userID := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.RateGenerationRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.svc.RateGeneration(userID, &req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(3101, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("评分成功", nil))
}
