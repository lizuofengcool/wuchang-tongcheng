// Package service 同城零工兼职业务逻辑层 - 任务包
// 对标斗米任务制 + 猪八戒威客
// 任务制（按件/按时/按日计费）+ 任务领取/交付/验收 + 状态机
package service

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"wuchang-tongcheng/internal/modules/linggong/dto"
	"wuchang-tongcheng/internal/modules/linggong/model"
	"wuchang-tongcheng/internal/modules/linggong/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrTaskNotFound         = errors.New("任务包不存在")
	ErrTaskNoPermission     = errors.New("无权操作此任务包")
	ErrTaskStatusInvalid    = errors.New("任务包状态不允许此操作")
	ErrTaskCountExceeded    = errors.New("任务领取数超过限制")
	ErrTaskDeadlineExceeded = errors.New("任务领取/交付截止时间已过")
)

// TaskService 任务包业务接口
type TaskService interface {
	// C 端
	Create(regionID uint, userID uint, req *dto.CreateTaskRequest) (*dto.TaskInfo, error)
	Update(id uint, operatorID uint, req *dto.UpdateTaskRequest) error
	Delete(id uint, operatorID uint) error
	GetByID(id uint) (*dto.TaskInfo, error)
	List(req *dto.TaskListRequest) (*utils.Pagination, []dto.TaskInfo, error)
	ListByLinggong(linggongID uint, page, pageSize int) (*utils.Pagination, []dto.TaskInfo, error)
	ListByEmployer(employerID uint, page, pageSize int) (*utils.Pagination, []dto.TaskInfo, error)

	// 任务领取/交付/验收
	Claim(id uint, workerID uint, workerName string, req *dto.TaskClaimRequest) error
	Submit(id uint, workerID uint, req *dto.TaskSubmitRequest) error
	Verify(id uint, employerID uint, req *dto.TaskVerifyRequest) error

	// M 端管理
	AdminList(req *dto.TaskAdminListRequest) (*utils.Pagination, []dto.TaskInfo, error)
	AdminGetByID(id uint) (*dto.TaskInfo, error)
	UpdateStatus(id uint, status int) error
}

type taskService struct {
	repo repository.TaskRepository
}

// NewTaskService 创建任务包 service 实例
func NewTaskService(repo repository.TaskRepository) TaskService {
	return &taskService{repo: repo}
}

// taskStatusText 任务状态文本
func taskStatusText(status int) string {
	switch status {
	case model.TaskStatusDraft:
		return "草稿"
	case model.TaskStatusPublished:
		return "已发布"
	case model.TaskStatusInProgress:
		return "进行中"
	case model.TaskStatusCompleted:
		return "已完成"
	case model.TaskStatusCanceled:
		return "已取消"
	case model.TaskStatusExpired:
		return "已过期"
	}
	return ""
}

// taskTypeText 任务类型文本
func taskTypeText(t string) string {
	switch t {
	case model.TaskTypeSingle:
		return "单一任务"
	case model.TaskTypeBatch:
		return "批量任务"
	case model.TaskTypeProject:
		return "项目制"
	case model.TaskTypeContest:
		return "比赛制"
	case model.TaskTypeBounty:
		return "悬赏制"
	}
	return ""
}

// taskDifficultyText 任务难度文本
func taskDifficultyText(d string) string {
	switch d {
	case model.TaskDifficultyEasy:
		return "简单"
	case model.TaskDifficultyMedium:
		return "中等"
	case model.TaskDifficultyHard:
		return "困难"
	case model.TaskDifficultyExpert:
		return "专家级"
	}
	return ""
}

// toTaskInfo model -> dto
func toTaskInfo(t *model.LinggongTask) *dto.TaskInfo {
	return &dto.TaskInfo{
		ID:              t.ID,
		TaskNo:          t.TaskNo,
		LinggongID:      t.LinggongID,
		EmployerID:      t.EmployerID,
		EmployerName:    t.EmployerName,
		Title:           t.Title,
		Description:     t.Description,
		TaskType:        t.TaskType,
		TaskTypeText:    taskTypeText(t.TaskType),
		Difficulty:      t.Difficulty,
		DifficultyText:  taskDifficultyText(t.Difficulty),
		DeliveryMethod:  t.DeliveryMethod,
		BillingType:     t.BillingType,
		UnitPrice:       t.UnitPrice,
		TotalCount:      t.TotalCount,
		ClaimedCount:    t.ClaimedCount,
		CompletedCount:  t.CompletedCount,
		VerifiedCount:   t.VerifiedCount,
		MaxClaimPerUser: t.MaxClaimPerUser,
		StartTime:       t.StartTime,
		EndTime:         t.EndTime,
		ClaimDeadline:   t.ClaimDeadline,
		SubmitDeadline:  t.SubmitDeadline,
		TotalAmount:     t.TotalAmount,
		PaidAmount:      t.PaidAmount,
		Status:          t.Status,
		StatusText:      taskStatusText(t.Status),
		AttachmentURL:   t.AttachmentURL,
		Tags:            t.Tags,
		Requirements:    t.Requirements,
		PublishedAt:     t.PublishedAt,
		CompletedAt:     t.CompletedAt,
		RegionID:        t.RegionID,
		CreatedAt:       t.CreatedAt,
		UpdatedAt:       t.UpdatedAt,
	}
}

// genTaskNo 生成任务单号：TK + yyyyMMddHHmmss + 6 位随机
func genTaskNo() string {
	return fmt.Sprintf("TK%s%06d", time.Now().Format("20060102150405"), rand.Intn(1000000))
}

// ===== C 端 =====

// Create 创建任务包
func (s *taskService) Create(regionID uint, userID uint, req *dto.CreateTaskRequest) (*dto.TaskInfo, error) {
	t := &model.LinggongTask{
		TaskNo:          genTaskNo(),
		LinggongID:      req.LinggongID,
		EmployerID:      userID,
		EmployerName:    "",
		Title:           req.Title,
		Description:     req.Description,
		TaskType:        req.TaskType,
		Difficulty:      req.Difficulty,
		DeliveryMethod:  req.DeliveryMethod,
		BillingType:     req.BillingType,
		UnitPrice:       req.UnitPrice,
		TotalCount:      req.TotalCount,
		MaxClaimPerUser: req.MaxClaimPerUser,
		StartTime:       req.StartTime,
		EndTime:         req.EndTime,
		ClaimDeadline:   req.ClaimDeadline,
		SubmitDeadline:  req.SubmitDeadline,
		TotalAmount:     req.TotalAmount,
		Status:          model.TaskStatusDraft,
		AttachmentURL:   req.AttachmentURL,
	}
	t.RegionID = regionID

	// 默认值兜底
	if t.TaskType == "" {
		t.TaskType = model.TaskTypeSingle
	}
	if t.Difficulty == "" {
		t.Difficulty = model.TaskDifficultyEasy
	}
	if t.DeliveryMethod == "" {
		t.DeliveryMethod = model.TaskDeliveryOnline
	}
	if t.BillingType == "" {
		t.BillingType = model.BillingTypeByPiece
	}
	if t.MaxClaimPerUser == 0 {
		t.MaxClaimPerUser = 1
	}
	if t.TotalCount == 0 {
		t.TotalCount = 1
	}

	// JSONB 字段
	if req.Tags != nil {
		if jb, err := model.FromJSON(req.Tags); err == nil {
			t.Tags = jb
		}
	}
	if req.Requirements != nil {
		if jb, err := model.FromJSON(req.Requirements); err == nil {
			t.Requirements = jb
		}
	}

	if err := s.repo.Create(t); err != nil {
		return nil, err
	}
	return toTaskInfo(t), nil
}

// Update 更新任务包（仅发布者本人）
func (s *taskService) Update(id uint, operatorID uint, req *dto.UpdateTaskRequest) error {
	t, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTaskNotFound
		}
		return err
	}
	if t.EmployerID != operatorID {
		return ErrTaskNoPermission
	}

	fields := map[string]interface{}{}
	if req.Title != nil {
		fields["title"] = *req.Title
	}
	if req.Description != nil {
		fields["description"] = *req.Description
	}
	if req.TaskType != nil {
		fields["task_type"] = *req.TaskType
	}
	if req.Difficulty != nil {
		fields["difficulty"] = *req.Difficulty
	}
	if req.DeliveryMethod != nil {
		fields["delivery_method"] = *req.DeliveryMethod
	}
	if req.BillingType != nil {
		fields["billing_type"] = *req.BillingType
	}
	if req.UnitPrice != nil {
		fields["unit_price"] = *req.UnitPrice
	}
	if req.TotalCount != nil {
		fields["total_count"] = *req.TotalCount
	}
	if req.MaxClaimPerUser != nil {
		fields["max_claim_per_user"] = *req.MaxClaimPerUser
	}
	if req.StartTime != nil {
		fields["start_time"] = *req.StartTime
	}
	if req.EndTime != nil {
		fields["end_time"] = *req.EndTime
	}
	if req.ClaimDeadline != nil {
		fields["claim_deadline"] = *req.ClaimDeadline
	}
	if req.SubmitDeadline != nil {
		fields["submit_deadline"] = *req.SubmitDeadline
	}
	if req.TotalAmount != nil {
		fields["total_amount"] = *req.TotalAmount
	}
	if req.AttachmentURL != nil {
		fields["attachment_url"] = *req.AttachmentURL
	}
	if req.Tags != nil {
		if jb, err := model.FromJSON(req.Tags); err == nil {
			fields["tags"] = jb
		}
	}
	if req.Requirements != nil {
		if jb, err := model.FromJSON(req.Requirements); err == nil {
			fields["requirements"] = jb
		}
	}

	if len(fields) == 0 {
		return nil
	}
	return s.repo.Update(id, fields)
}

// Delete 删除任务包（仅发布者本人）
func (s *taskService) Delete(id uint, operatorID uint) error {
	t, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTaskNotFound
		}
		return err
	}
	if t.EmployerID != operatorID {
		return ErrTaskNoPermission
	}
	return s.repo.Delete(id)
}

// GetByID 获取详情
func (s *taskService) GetByID(id uint) (*dto.TaskInfo, error) {
	t, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}
	return toTaskInfo(t), nil
}

// List 任务列表查询
func (s *taskService) List(req *dto.TaskListRequest) (*utils.Pagination, []dto.TaskInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.TaskListOptions{
		LinggongID: req.LinggongID,
		EmployerID: req.EmployerID,
		TaskType:   req.TaskType,
		Difficulty: req.Difficulty,
		Status:     req.Status,
		Keyword:    req.Keyword,
	}

	// C 端默认仅展示已发布
	if opts.Status == nil {
		published := model.TaskStatusPublished
		opts.Status = &published
	}

	list, total, err := s.repo.List(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.TaskInfo, 0, len(list))
	for i := range list {
		result = append(result, *toTaskInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListByLinggong 按岗位反查
func (s *taskService) ListByLinggong(linggongID uint, page, pageSize int) (*utils.Pagination, []dto.TaskInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByLinggong(linggongID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.TaskInfo, 0, len(list))
	for i := range list {
		result = append(result, *toTaskInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListByEmployer 按雇主反查
func (s *taskService) ListByEmployer(employerID uint, page, pageSize int) (*utils.Pagination, []dto.TaskInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByEmployer(employerID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.TaskInfo, 0, len(list))
	for i := range list {
		result = append(result, *toTaskInfo(&list[i]))
	}
	return pagination, result, nil
}

// Claim 任务领取
func (s *taskService) Claim(id uint, workerID uint, workerName string, req *dto.TaskClaimRequest) error {
	t, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTaskNotFound
		}
		return err
	}

	// 状态校验：仅"已发布"或"进行中"的任务可领取
	if t.Status != model.TaskStatusPublished && t.Status != model.TaskStatusInProgress {
		return ErrTaskStatusInvalid
	}

	// 领取截止时间校验
	if t.ClaimDeadline != nil && time.Now().After(*t.ClaimDeadline) {
		return ErrTaskDeadlineExceeded
	}

	// 库存校验
	remaining := t.TotalCount - t.ClaimedCount
	if remaining < req.Count {
		return ErrTaskCountExceeded
	}

	// 单人领取数校验
	if req.Count > t.MaxClaimPerUser {
		return ErrTaskCountExceeded
	}

	// 增加已领取数
	if err := s.repo.IncrClaimedCount(id, req.Count); err != nil {
		return err
	}

	// 若任务首次被领取，状态从"已发布"变为"进行中"
	if t.Status == model.TaskStatusPublished {
		now := time.Now()
		_ = s.repo.Update(id, map[string]interface{}{
			"status": model.TaskStatusInProgress,
		})
		_ = now // 防止 lint 报错
	}

	return nil
}

// Submit 任务交付
func (s *taskService) Submit(id uint, workerID uint, req *dto.TaskSubmitRequest) error {
	t, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTaskNotFound
		}
		return err
	}

	if t.Status != model.TaskStatusInProgress && t.Status != model.TaskStatusPublished {
		return ErrTaskStatusInvalid
	}

	// 交付截止时间校验
	if t.SubmitDeadline != nil && time.Now().After(*t.SubmitDeadline) {
		return ErrTaskDeadlineExceeded
	}

	// 增加已完成数
	if err := s.repo.IncrCompletedCount(id, req.Count); err != nil {
		return err
	}

	return nil
}

// Verify 任务验收
func (s *taskService) Verify(id uint, employerID uint, req *dto.TaskVerifyRequest) error {
	t, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTaskNotFound
		}
		return err
	}
	if t.EmployerID != employerID {
		return ErrTaskNoPermission
	}

	if t.Status != model.TaskStatusInProgress && t.Status != model.TaskStatusPublished {
		return ErrTaskStatusInvalid
	}

	// 验收通过：增加已验收数；若已验收数达到总数，标记为已完成
	if req.Pass {
		if err := s.repo.IncrVerifiedCount(id, req.Count); err != nil {
			return err
		}
		// 增加已支付金额
		paidAmount := t.UnitPrice * float64(req.Count)
		if err := s.repo.IncrPaidAmount(id, paidAmount); err != nil {
			return err
		}
		// 若已验收数达到总数，状态变为已完成
		if t.VerifiedCount+req.Count >= t.TotalCount {
			now := time.Now()
			_ = s.repo.Update(id, map[string]interface{}{
				"status":       model.TaskStatusCompleted,
				"completed_at": &now,
			})
		}
	}

	return nil
}

// ===== M 端管理 =====

func (s *taskService) AdminList(req *dto.TaskAdminListRequest) (*utils.Pagination, []dto.TaskInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.TaskAdminListOptions{
		RegionID:   req.RegionID,
		LinggongID: req.LinggongID,
		EmployerID: req.EmployerID,
		TaskType:   req.TaskType,
		Status:     req.Status,
		Keyword:    req.Keyword,
	}
	list, total, err := s.repo.AdminList(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.TaskInfo, 0, len(list))
	for i := range list {
		result = append(result, *toTaskInfo(&list[i]))
	}
	return pagination, result, nil
}

func (s *taskService) AdminGetByID(id uint) (*dto.TaskInfo, error) {
	t, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}
	return toTaskInfo(t), nil
}

// UpdateStatus M 端状态更新
func (s *taskService) UpdateStatus(id uint, status int) error {
	fields := map[string]interface{}{
		"status": status,
	}
	now := time.Now()
	switch status {
	case model.TaskStatusPublished:
		fields["published_at"] = &now
	case model.TaskStatusCompleted:
		fields["completed_at"] = &now
	}
	return s.repo.Update(id, fields)
}
