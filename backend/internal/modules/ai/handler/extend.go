// Package handler AI 中台扩展 HTTP 处理层
// 依据 016_ai_full.sql：审核/推荐/对话/模型配置/训练数据/统计
package handler

import (
	"net/http"
	"strconv"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/modules/ai/dto"
	"wuchang-tongcheng/internal/modules/ai/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// ExtendHandler AI 中台扩展处理器
type ExtendHandler struct {
	extSvc service.AIExtendService
}

// NewExtendHandler 创建扩展 Handler 实例
func NewExtendHandler(extSvc service.AIExtendService) *ExtendHandler {
	return &ExtendHandler{extSvc: extSvc}
}

// ===== 审核 =====

// SubmitAudit 提交审核
// POST /api/v1/ai/audit/submit
func (h *ExtendHandler) SubmitAudit(ctx plugin.Context) {
	userID := getUserID(ctx)
	// DTO 中 AuditContentRequest 缺少 Content 字段，这里使用内联结构体补充
	var req struct {
		Content     string `json:"content"`
		BizModule   string `json:"biz_module" binding:"required"`
		BizID       string `json:"biz_id" binding:"required"`
		UserID      uint   `json:"user_id"`
		ContentType string `json:"content_type" binding:"required,oneof=text image video"`
		ContentHash string `json:"content_hash"`
		Algorithm   string `json:"algorithm" binding:"omitempty,oneof=local dfa aliyun tencent"`
	}
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if req.UserID == 0 {
		req.UserID = userID
	}
	bizID, _ := strconv.ParseUint(req.BizID, 10, 64)
	info, err := h.extSvc.SubmitAudit(req.Content, req.ContentType, req.BizModule, uint(bizID))
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeAIError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("审核已提交", info))
}

// GetAuditResult 查询审核结果
// GET /api/v1/ai/audit/:task_id
func (h *ExtendHandler) GetAuditResult(ctx plugin.Context) {
	taskID := ctx.Param("task_id")
	if taskID == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("task_id 不能为空"))
		return
	}
	info, err := h.extSvc.GetAuditResult(taskID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeAIError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(info))
}

// ListAuditResults 审核结果列表
// GET /api/v1/ai/audit
func (h *ExtendHandler) ListAuditResults(ctx plugin.Context) {
	level := ctx.Query("level")
	page, pageSize := parsePagination(ctx)
	list, total, err := h.extSvc.ListAuditResults(level, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeAIError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(response.NewPageResult(list, total, page, pageSize)))
}

// ReviewAudit 人工复核
// PUT /api/v1/ai/audit/:task_id/review
func (h *ExtendHandler) ReviewAudit(ctx plugin.Context) {
	taskID := ctx.Param("task_id")
	if taskID == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("task_id 不能为空"))
		return
	}
	var req struct {
		Passed  bool   `json:"passed"`
		Comment string `json:"comment"`
	}
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	reviewer := getUserID(ctx)
	if err := h.extSvc.ReviewAudit(taskID, req.Passed, strconv.FormatUint(uint64(reviewer), 10), req.Comment); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeAIError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("复核完成", nil))
}

// ===== 推荐 =====

// GetRecommendations 推荐列表
// GET /api/v1/ai/recommendations
func (h *ExtendHandler) GetRecommendations(ctx plugin.Context) {
	userID := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	recType := ctx.Query("rec_type")
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	list, err := h.extSvc.GetRecommendations(userID, recType, limit)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeAIError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(list))
}

// TrackClick 点击追踪
// POST /api/v1/ai/recommendations/:id/click
func (h *ExtendHandler) TrackClick(ctx plugin.Context) {
	recID, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if recID == 0 {
		ctx.JSON(http.StatusOK, response.BadRequest("推荐ID无效"))
		return
	}
	var req dto.TrackClickRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.extSvc.TrackClick(uint(recID), req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeAIError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("已记录点击", nil))
}

// TrackDwell 停留时长追踪
// POST /api/v1/ai/recommendations/:id/dwell
func (h *ExtendHandler) TrackDwell(ctx plugin.Context) {
	recID, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if recID == 0 {
		ctx.JSON(http.StatusOK, response.BadRequest("推荐ID无效"))
		return
	}
	var req dto.TrackDwellRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.extSvc.TrackDwell(uint(recID), req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeAIError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("已记录停留时长", nil))
}

// FeedbackRecommendation 推荐反馈
// POST /api/v1/ai/recommendations/:id/feedback
func (h *ExtendHandler) FeedbackRecommendation(ctx plugin.Context) {
	recID, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if recID == 0 {
		ctx.JSON(http.StatusOK, response.BadRequest("推荐ID无效"))
		return
	}
	var req dto.FeedbackRecommendationRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.extSvc.FeedbackRecommendation(uint(recID), req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeAIError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("反馈成功", nil))
}

// ===== 模型配置 =====

// ListModelConfigs 模型配置列表
// GET /api/v1/ai/models
func (h *ExtendHandler) ListModelConfigs(ctx plugin.Context) {
	modelType := ctx.Query("model_type")
	page, pageSize := parsePagination(ctx)
	list, total, err := h.extSvc.ListModelConfigs(modelType, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeAIError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(response.NewPageResult(list, total, page, pageSize)))
}

// UpsertModelConfig 创建/更新模型配置
// POST /api/v1/ai/models
func (h *ExtendHandler) UpsertModelConfig(ctx plugin.Context) {
	var req dto.UpsertModelConfigRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	id, err := h.extSvc.UpsertModelConfig(req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeAIError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("保存成功", map[string]uint{"id": id}))
}

// DeleteModelConfig 删除模型配置
// DELETE /api/v1/ai/models/:id
func (h *ExtendHandler) DeleteModelConfig(ctx plugin.Context) {
	id, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if id == 0 {
		ctx.JSON(http.StatusOK, response.BadRequest("配置ID无效"))
		return
	}
	if err := h.extSvc.DeleteModelConfig(uint(id)); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeAIModelConfigNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("删除成功", nil))
}

// ===== 对话 =====

// CreateChatSession 创建对话会话
// POST /api/v1/ai/chat/sessions
func (h *ExtendHandler) CreateChatSession(ctx plugin.Context) {
	userID := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.CreateChatSessionRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	info, err := h.extSvc.CreateChatSession(userID, req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeAIError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("会话创建成功", info))
}

// ListChatSessions 我的对话会话列表
// GET /api/v1/ai/chat/sessions
func (h *ExtendHandler) ListChatSessions(ctx plugin.Context) {
	userID := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	page, pageSize := parsePagination(ctx)
	list, total, err := h.extSvc.ListChatSessions(userID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeAIError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(response.NewPageResult(list, total, page, pageSize)))
}

// Chat 对话
// POST /api/v1/ai/chat/messages
func (h *ExtendHandler) Chat(ctx plugin.Context) {
	userID := getUserID(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.Unauthorized("请先登录"))
		return
	}
	var req dto.ChatRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	resp, err := h.extSvc.Chat(userID, req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeAIError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}

// ListChatMessages 对话消息列表
// GET /api/v1/ai/chat/sessions/:session_id/messages
func (h *ExtendHandler) ListChatMessages(ctx plugin.Context) {
	sessionID := ctx.Param("session_id")
	if sessionID == "" {
		ctx.JSON(http.StatusOK, response.BadRequest("session_id 不能为空"))
		return
	}
	page, pageSize := parsePagination(ctx)
	list, total, err := h.extSvc.ListChatMessages(sessionID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeAIError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(response.NewPageResult(list, total, page, pageSize)))
}

// MessageFeedback 消息反馈
// POST /api/v1/ai/chat/messages/:id/feedback
func (h *ExtendHandler) MessageFeedback(ctx plugin.Context) {
	messageID, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if messageID == 0 {
		ctx.JSON(http.StatusOK, response.BadRequest("消息ID无效"))
		return
	}
	var req dto.MessageFeedbackRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	if err := h.extSvc.MessageFeedback(uint(messageID), req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeAIChatMessageNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("反馈成功", nil))
}

// ===== 训练数据 =====

// ListTrainingData 训练数据列表
// GET /api/v1/ai/training-data
func (h *ExtendHandler) ListTrainingData(ctx plugin.Context) {
	dataType := ctx.Query("data_type")
	page, pageSize := parsePagination(ctx)
	list, total, err := h.extSvc.ListTrainingData(dataType, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeAIError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(response.NewPageResult(list, total, page, pageSize)))
}

// CreateTrainingData 创建训练数据
// POST /api/v1/ai/training-data
func (h *ExtendHandler) CreateTrainingData(ctx plugin.Context) {
	var req dto.CreateTrainingDataRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, response.BadRequest("参数错误: "+err.Error()))
		return
	}
	id, err := h.extSvc.CreateTrainingData(req)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeAIError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("创建成功", map[string]uint{"id": id}))
}

// MarkTrainingDataUsed 标记训练数据已使用
// PUT /api/v1/ai/training-data/:id/used
func (h *ExtendHandler) MarkTrainingDataUsed(ctx plugin.Context) {
	id, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if id == 0 {
		ctx.JSON(http.StatusOK, response.BadRequest("训练数据ID无效"))
		return
	}
	if err := h.extSvc.MarkTrainingDataUsed(uint(id)); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeAITrainingNotFound, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.SuccessWithMessage("标记成功", nil))
}

// ===== 统计 =====

// GetStatistics AI 统计（M 端）
// GET /api/v1/ai/admin/statistics
func (h *ExtendHandler) GetStatistics(ctx plugin.Context) {
	resp, err := h.extSvc.GetStatistics()
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(utils.CodeAIError, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, response.Success(resp))
}
