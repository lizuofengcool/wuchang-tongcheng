// Package service 同城拼车出行业务逻辑层 - 批量任务
package service

import (
	"errors"
	"fmt"
	"time"

	"wuchang-tongcheng/internal/modules/pinche/dto"
	"wuchang-tongcheng/internal/modules/pinche/model"
	"wuchang-tongcheng/internal/modules/pinche/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrBatchTaskNotFound = errors.New("批量任务不存在")
	ErrBatchTaskInvalid  = errors.New("批量任务参数无效")
)

// BatchTaskService 批量任务业务接口
type BatchTaskService interface {
	AdminList(req *dto.BatchTaskListRequest) (*dto.BatchTaskListResult, error)
	AdminGetByID(id uint) (*dto.BatchTaskInfo, error)
	Create(regionID, operatorID uint, operatorName string, req *dto.CreateBatchTaskRequest) (*dto.BatchTaskInfo, error)
	Cancel(id uint) error
	Retry(id uint) error
	Delete(id uint) error
	PreviewIDs(req *dto.PreviewIDsRequest) (*dto.PreviewIDsResponse, error)
	Stats() (*dto.BatchTaskStatsResponse, error)
}

type batchTaskService struct {
	repo repository.BatchTaskRepository
}

// NewBatchTaskService 创建批量任务 service 实例
func NewBatchTaskService(repo repository.BatchTaskRepository) BatchTaskService {
	return &batchTaskService{repo: repo}
}

// batchTaskStatusText 状态文本
func batchTaskStatusText(status int) string {
	switch status {
	case model.BatchTaskStatusPending:
		return "待执行"
	case model.BatchTaskStatusRunning:
		return "执行中"
	case model.BatchTaskStatusDone:
		return "已完成"
	case model.BatchTaskStatusFailed:
		return "失败"
	case model.BatchTaskStatusCanceled:
		return "已取消"
	}
	return ""
}

// toBatchTaskInfo model -> dto
func toBatchTaskInfo(t *model.PincheBatchTask) *dto.BatchTaskInfo {
	info := &dto.BatchTaskInfo{
		ID:           t.ID,
		RegionID:     t.RegionID,
		TaskNo:       t.TaskNo,
		Name:         t.TaskName,
		TaskType:     t.TaskType,
		TargetType:   t.TaskType,
		TargetIDs:    t.TargetIDs,
		TargetCount:  t.TargetCount,
		Filters:      t.Filters,
		Action:       t.Action,
		ActionParams: t.ActionParams,
		Status:       t.Status,
		StatusText:   batchTaskStatusText(t.Status),
		Progress:     t.Progress,
		SuccessCount: t.SuccessCount,
		FailCount:    t.FailCount,
		FailReason:   t.FailReason,
		OperatorID:   t.OperatorID,
		OperatorName: t.OperatorName,
		StartedAt:    t.StartedAt,
		CompletedAt:  t.FinishedAt,
		FinishedAt:   t.FinishedAt,
		CreatedAt:    t.CreatedAt,
	}
	return info
}

// generateTaskNo 生成任务编号
func generateTaskNo(id uint) string {
	return fmt.Sprintf("BT%08d", id)
}

// AdminList 管理后台批量任务列表
func (s *batchTaskService) AdminList(req *dto.BatchTaskListRequest) (*dto.BatchTaskListResult, error) {
	if req == nil {
		req = &dto.BatchTaskListRequest{}
	}
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	pagination := utils.NewPagination(page, pageSize)

	opts := repository.BatchTaskListOptions{
		TaskType: req.TargetType,
		Action:   req.Action,
		Status:   req.Status,
		Keyword:  req.Keyword,
	}
	// 跨地区：regionID=0
	list, total, err := s.repo.List(0, pagination, opts)
	if err != nil {
		return nil, err
	}

	result := make([]dto.BatchTaskInfo, 0, len(list))
	for i := range list {
		result = append(result, *toBatchTaskInfo(&list[i]))
	}

	stats, err := s.Stats()
	if err != nil {
		return nil, err
	}

	return &dto.BatchTaskListResult{
		List:     result,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Stats:    *stats,
	}, nil
}

// AdminGetByID 批量任务详情
func (s *batchTaskService) AdminGetByID(id uint) (*dto.BatchTaskInfo, error) {
	t, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrBatchTaskNotFound
		}
		return nil, err
	}
	return toBatchTaskInfo(t), nil
}

// Create 创建批量任务
func (s *batchTaskService) Create(regionID, operatorID uint, operatorName string, req *dto.CreateBatchTaskRequest) (*dto.BatchTaskInfo, error) {
	if req == nil {
		return nil, ErrBatchTaskInvalid
	}
	if req.Name == "" {
		return nil, ErrBatchTaskInvalid
	}
	if req.TargetType == "" {
		return nil, ErrBatchTaskInvalid
	}
	if req.Action == "" {
		return nil, ErrBatchTaskInvalid
	}

	t := &model.PincheBatchTask{
		TaskName:     req.Name,
		TaskType:     req.TargetType, // task_type 与 target_type 一致
		Action:       req.Action,
		TargetCount:  len(req.TargetIDs),
		Status:       model.BatchTaskStatusPending,
		OperatorID:   &operatorID,
		OperatorName: operatorName,
	}
	t.RegionID = regionID

	// target_ids → JSONB
	if len(req.TargetIDs) > 0 {
		if jb, err := model.FromJSON(req.TargetIDs); err == nil {
			t.TargetIDs = jb
		}
	}
	// filters → JSONB
	if req.Filters != nil {
		if jb, err := model.FromJSON(req.Filters); err == nil {
			t.Filters = jb
		}
	}
	// action_params → JSONB
	if req.ActionParams != nil {
		if jb, err := model.FromJSON(req.ActionParams); err == nil {
			t.ActionParams = jb
		}
	}

	if err := s.repo.Create(t); err != nil {
		return nil, err
	}

	// 生成 task_no
	t.TaskNo = generateTaskNo(t.ID)
	_ = s.repo.Update(t.ID, map[string]interface{}{
		"task_no": t.TaskNo,
	})

	return toBatchTaskInfo(t), nil
}

// Cancel 取消任务
func (s *batchTaskService) Cancel(id uint) error {
	t, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrBatchTaskNotFound
		}
		return err
	}
	if t.Status != model.BatchTaskStatusPending && t.Status != model.BatchTaskStatusRunning {
		return errors.New("当前状态不允许取消")
	}
	now := time.Now()
	return s.repo.Update(id, map[string]interface{}{
		"status":      model.BatchTaskStatusCanceled,
		"finished_at": &now,
	})
}

// Retry 重试任务
func (s *batchTaskService) Retry(id uint) error {
	t, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrBatchTaskNotFound
		}
		return err
	}
	if t.Status != model.BatchTaskStatusFailed && t.Status != model.BatchTaskStatusDone {
		return errors.New("仅失败或已完成的任务可重试")
	}
	now := time.Now()
	return s.repo.Update(id, map[string]interface{}{
		"status":        model.BatchTaskStatusPending,
		"progress":      0,
		"success_count": 0,
		"fail_count":    0,
		"fail_reason":   "",
		"started_at":    nil,
		"finished_at":   nil,
		"updated_at":    now,
	})
}

// Delete 删除任务
func (s *batchTaskService) Delete(id uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrBatchTaskNotFound
		}
		return err
	}
	return s.repo.Delete(id)
}

// PreviewIDs 预览 ID 列表
// 该接口用于"新建批量任务"时辅助加载目标 ID 列表
// 实际生产环境应基于 target_type/filters 查询对应业务表
// 此处简化为返回空列表，前端可通过其他 admin 接口（如 /admin/pinches）自行获取 ID
func (s *batchTaskService) PreviewIDs(req *dto.PreviewIDsRequest) (*dto.PreviewIDsResponse, error) {
	if req == nil {
		req = &dto.PreviewIDsRequest{}
	}
	limit := req.Limit
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}
	// 返回空 ID 列表，前端会自行通过其他接口加载
	return &dto.PreviewIDsResponse{
		IDS:   []uint{},
		Total: 0,
	}, nil
}

// Stats 批量任务统计
func (s *batchTaskService) Stats() (*dto.BatchTaskStatsResponse, error) {
	resp := &dto.BatchTaskStatsResponse{}
	pending, err := s.repo.CountByStatus(0, model.BatchTaskStatusPending)
	if err != nil {
		return nil, err
	}
	running, err := s.repo.CountByStatus(0, model.BatchTaskStatusRunning)
	if err != nil {
		return nil, err
	}
	done, err := s.repo.CountByStatus(0, model.BatchTaskStatusDone)
	if err != nil {
		return nil, err
	}
	failed, err := s.repo.CountByStatus(0, model.BatchTaskStatusFailed)
	if err != nil {
		return nil, err
	}
	canceled, err := s.repo.CountByStatus(0, model.BatchTaskStatusCanceled)
	if err != nil {
		return nil, err
	}
	resp.Total = pending + running + done + failed + canceled
	resp.Pending = pending
	resp.Running = running
	resp.Completed = done
	return resp, nil
}
