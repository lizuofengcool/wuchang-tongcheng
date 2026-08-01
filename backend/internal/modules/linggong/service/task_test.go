// Package service 同城零工兼职业务逻辑层单元测试 - 任务包。
// 使用内存 mock 仓储覆盖：任务包创建与默认值兜底、JSONB 字段转换、
// 更新/删除权限校验与字段构建、领取校验（状态/截止时间/库存/单人限领/状态机迁移）、
// 交付校验（状态/截止时间）、验收校验（权限/状态/通过累加验收数与支付金额/完成态迁移）、
// 管理端状态更新时间戳、列表默认状态过滤等核心逻辑，不依赖 DB。
package service

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"wuchang-tongcheng/internal/modules/linggong/dto"
	"wuchang-tongcheng/internal/modules/linggong/model"
	"wuchang-tongcheng/internal/modules/linggong/repository"
	"wuchang-tongcheng/internal/pkg/utils"
)

// ===== mockTaskRepo =====

type updateCall struct {
	id     uint
	fields map[string]interface{}
}

type incrCountCall struct {
	id    uint
	count int
}

type incrAmountCall struct {
	id     uint
	amount float64
}

type mockTaskRepo struct {
	byID   map[uint]*model.LinggongTask
	nextID uint

	createErr error
	findErr   error
	updateErr error
	deleteErr error
	listErr   error

	updates        []updateCall
	deletes        []uint
	incrClaimed    []incrCountCall
	incrCompleted  []incrCountCall
	incrVerified   []incrCountCall
	incrPaidAmount []incrAmountCall

	listReturn      []model.LinggongTask
	listTotal       int64
	adminListReturn []model.LinggongTask
	adminListTotal  int64
	byLinggongList  []model.LinggongTask
	byLinggongTotal int64
	byEmployerList  []model.LinggongTask
	byEmployerTotal int64
}

func newMockTaskRepo() *mockTaskRepo {
	return &mockTaskRepo{
		byID:   make(map[uint]*model.LinggongTask),
		nextID: 1,
	}
}

// seed 预置一条任务并返回其副本指针（便于断言）。
func (m *mockTaskRepo) seed(t *model.LinggongTask) *model.LinggongTask {
	if t.ID == 0 {
		t.ID = m.nextID
		m.nextID++
	}
	if t.TaskNo == "" {
		t.TaskNo = "TK-SEED"
	}
	cp := *t
	m.byID[t.ID] = &cp
	return &cp
}

func (m *mockTaskRepo) Create(t *model.LinggongTask) error {
	if m.createErr != nil {
		return m.createErr
	}
	t.ID = m.nextID
	m.nextID++
	cp := *t
	m.byID[t.ID] = &cp
	return nil
}

func (m *mockTaskRepo) FindByID(id uint) (*model.LinggongTask, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	t, ok := m.byID[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *t
	return &cp, nil
}

func (m *mockTaskRepo) Update(id uint, fields map[string]interface{}) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.updates = append(m.updates, updateCall{id: id, fields: fields})
	// 同步内存状态，便于后续断言/连续操作
	if t, ok := m.byID[id]; ok {
		if v, ok := fields["status"]; ok {
			if s, ok := v.(int); ok {
				t.Status = s
			}
		}
		if v, ok := fields["completed_at"]; ok {
			if ts, ok := v.(*time.Time); ok {
				t.CompletedAt = ts
			}
		}
		if v, ok := fields["published_at"]; ok {
			if ts, ok := v.(*time.Time); ok {
				t.PublishedAt = ts
			}
		}
	}
	return nil
}

func (m *mockTaskRepo) Delete(id uint) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	m.deletes = append(m.deletes, id)
	delete(m.byID, id)
	return nil
}

func (m *mockTaskRepo) List(_ *utils.Pagination, _ repository.TaskListOptions) ([]model.LinggongTask, int64, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	return m.listReturn, m.listTotal, nil
}

func (m *mockTaskRepo) AdminList(_ *utils.Pagination, _ repository.TaskAdminListOptions) ([]model.LinggongTask, int64, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	return m.adminListReturn, m.adminListTotal, nil
}

func (m *mockTaskRepo) ListByLinggong(_ uint, _ *utils.Pagination) ([]model.LinggongTask, int64, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	return m.byLinggongList, m.byLinggongTotal, nil
}

func (m *mockTaskRepo) ListByEmployer(_ uint, _ *utils.Pagination) ([]model.LinggongTask, int64, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	return m.byEmployerList, m.byEmployerTotal, nil
}

func (m *mockTaskRepo) IncrClaimedCount(id uint, count int) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.incrClaimed = append(m.incrClaimed, incrCountCall{id: id, count: count})
	if t, ok := m.byID[id]; ok {
		t.ClaimedCount += count
	}
	return nil
}

func (m *mockTaskRepo) IncrCompletedCount(id uint, count int) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.incrCompleted = append(m.incrCompleted, incrCountCall{id: id, count: count})
	if t, ok := m.byID[id]; ok {
		t.CompletedCount += count
	}
	return nil
}

func (m *mockTaskRepo) IncrVerifiedCount(id uint, count int) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.incrVerified = append(m.incrVerified, incrCountCall{id: id, count: count})
	if t, ok := m.byID[id]; ok {
		t.VerifiedCount += count
	}
	return nil
}

func (m *mockTaskRepo) IncrPaidAmount(id uint, amount float64) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.incrPaidAmount = append(m.incrPaidAmount, incrAmountCall{id: id, amount: amount})
	if t, ok := m.byID[id]; ok {
		t.PaidAmount += amount
	}
	return nil
}

// ===== 文本映射辅助函数 =====

func TestTaskStatusText(t *testing.T) {
	cases := map[int]string{
		model.TaskStatusDraft:      "草稿",
		model.TaskStatusPublished:  "已发布",
		model.TaskStatusInProgress: "进行中",
		model.TaskStatusCompleted:  "已完成",
		model.TaskStatusCanceled:   "已取消",
		model.TaskStatusExpired:    "已过期",
		99:                         "",
		-1:                         "",
	}
	for status, want := range cases {
		assert.Equal(t, want, taskStatusText(status), "status=%d", status)
	}
}

func TestTaskTypeText(t *testing.T) {
	cases := map[string]string{
		model.TaskTypeSingle:  "单一任务",
		model.TaskTypeBatch:   "批量任务",
		model.TaskTypeProject: "项目制",
		model.TaskTypeContest: "比赛制",
		model.TaskTypeBounty:  "悬赏制",
		"unknown":             "",
		"":                    "",
	}
	for typ, want := range cases {
		assert.Equal(t, want, taskTypeText(typ), "type=%s", typ)
	}
}

func TestTaskDifficultyText(t *testing.T) {
	cases := map[string]string{
		model.TaskDifficultyEasy:   "简单",
		model.TaskDifficultyMedium: "中等",
		model.TaskDifficultyHard:   "困难",
		model.TaskDifficultyExpert: "专家级",
		"unknown":                  "",
		"":                         "",
	}
	for d, want := range cases {
		assert.Equal(t, want, taskDifficultyText(d), "difficulty=%s", d)
	}
}

func TestGenTaskNo(t *testing.T) {
	no := genTaskNo()
	assert.True(t, len(no) > 6, "任务单号长度应大于 6: %s", no)
	assert.Equal(t, "TK", no[:2], "任务单号应以 TK 开头")
	// 时间戳部分为 14 位 + 6 位随机 = 22
	assert.Equal(t, 22, len(no), "任务单号总长度应为 22 (TK+14位时间+6位随机): %s", no)
}

// ===== toTaskInfo 映射 =====

func TestToTaskInfo(t *testing.T) {
	now := time.Now()
	tags, _ := model.FromJSON([]string{"urgent", "remote"})
	src := &model.LinggongTask{
		TaskNo:          "TK001",
		LinggongID:      10,
		EmployerID:      20,
		EmployerName:    "雇主A",
		Title:           "标题",
		Description:     "描述",
		TaskType:        model.TaskTypeBatch,
		Difficulty:      model.TaskDifficultyHard,
		DeliveryMethod:  model.TaskDeliveryOffline,
		BillingType:     model.BillingTypeByDay,
		UnitPrice:       12.5,
		TotalCount:      8,
		ClaimedCount:    2,
		CompletedCount:  1,
		VerifiedCount:   1,
		MaxClaimPerUser: 3,
		StartTime:       &now,
		EndTime:         &now,
		ClaimDeadline:   &now,
		SubmitDeadline:  &now,
		TotalAmount:     100,
		PaidAmount:      12.5,
		Status:          model.TaskStatusInProgress,
		AttachmentURL:   "https://example.com/a.zip",
		Tags:            tags,
		Requirements:    model.JSONB(`{"exp":"3y"}`),
		PublishedAt:     &now,
		CompletedAt:     &now,
	}
	src.RegionID = 7
	src.CreatedAt = now
	src.UpdatedAt = now

	info := toTaskInfo(src)
	assert.Equal(t, src.TaskNo, info.TaskNo)
	assert.Equal(t, src.LinggongID, info.LinggongID)
	assert.Equal(t, src.EmployerID, info.EmployerID)
	assert.Equal(t, src.EmployerName, info.EmployerName)
	assert.Equal(t, src.Title, info.Title)
	assert.Equal(t, src.Description, info.Description)
	assert.Equal(t, src.TaskType, info.TaskType)
	assert.Equal(t, "批量任务", info.TaskTypeText)
	assert.Equal(t, src.Difficulty, info.Difficulty)
	assert.Equal(t, "困难", info.DifficultyText)
	assert.Equal(t, src.DeliveryMethod, info.DeliveryMethod)
	assert.Equal(t, src.BillingType, info.BillingType)
	assert.Equal(t, src.UnitPrice, info.UnitPrice)
	assert.Equal(t, src.TotalCount, info.TotalCount)
	assert.Equal(t, src.ClaimedCount, info.ClaimedCount)
	assert.Equal(t, src.CompletedCount, info.CompletedCount)
	assert.Equal(t, src.VerifiedCount, info.VerifiedCount)
	assert.Equal(t, src.MaxClaimPerUser, info.MaxClaimPerUser)
	assert.Equal(t, src.StartTime, info.StartTime)
	assert.Equal(t, src.EndTime, info.EndTime)
	assert.Equal(t, src.ClaimDeadline, info.ClaimDeadline)
	assert.Equal(t, src.SubmitDeadline, info.SubmitDeadline)
	assert.Equal(t, src.TotalAmount, info.TotalAmount)
	assert.Equal(t, src.PaidAmount, info.PaidAmount)
	assert.Equal(t, src.Status, info.Status)
	assert.Equal(t, "进行中", info.StatusText)
	assert.Equal(t, src.AttachmentURL, info.AttachmentURL)
	assert.Equal(t, src.RegionID, info.RegionID)
	assert.NotNil(t, info.Tags)
	assert.NotNil(t, info.Requirements)
}

// ===== Create =====

func TestTaskCreate_DefaultsApplied(t *testing.T) {
	repo := newMockTaskRepo()
	svc := NewTaskService(repo)

	info, err := svc.Create(5, 100, &dto.CreateTaskRequest{
		LinggongID: 9,
		Title:      "T",
	})
	require.NoError(t, err)
	require.NotNil(t, info)

	// 默认值兜底
	assert.Equal(t, model.TaskTypeSingle, info.TaskType)
	assert.Equal(t, model.TaskDifficultyEasy, info.Difficulty)
	assert.Equal(t, model.TaskDeliveryOnline, info.DeliveryMethod)
	assert.Equal(t, model.BillingTypeByPiece, info.BillingType)
	assert.Equal(t, 1, info.MaxClaimPerUser)
	assert.Equal(t, 1, info.TotalCount)
	assert.Equal(t, model.TaskStatusDraft, info.Status)
	assert.Equal(t, "草稿", info.StatusText)
	// 雇主 ID 来自登录用户
	assert.Equal(t, uint(100), info.EmployerID)
	assert.Equal(t, uint(5), info.RegionID)
	// 单号已生成
	assert.NotEmpty(t, info.TaskNo)
	assert.Equal(t, "TK", info.TaskNo[:2])
	// 已写入仓储
	assert.Len(t, repo.byID, 1)
}

func TestTaskCreate_ProvidedValuesPreserved(t *testing.T) {
	repo := newMockTaskRepo()
	svc := NewTaskService(repo)

	info, err := svc.Create(1, 1, &dto.CreateTaskRequest{
		LinggongID:      2,
		Title:           "标题",
		Description:     "描述",
		TaskType:        model.TaskTypeProject,
		Difficulty:      model.TaskDifficultyExpert,
		DeliveryMethod:  model.TaskDeliveryBoth,
		BillingType:     model.BillingTypeFixed,
		UnitPrice:       50.0,
		TotalCount:      10,
		MaxClaimPerUser: 4,
		TotalAmount:     500,
		AttachmentURL:   "https://e.com/x",
	})
	require.NoError(t, err)
	assert.Equal(t, model.TaskTypeProject, info.TaskType)
	assert.Equal(t, model.TaskDifficultyExpert, info.Difficulty)
	assert.Equal(t, model.TaskDeliveryBoth, info.DeliveryMethod)
	assert.Equal(t, model.BillingTypeFixed, info.BillingType)
	assert.Equal(t, 50.0, info.UnitPrice)
	assert.Equal(t, 10, info.TotalCount)
	assert.Equal(t, 4, info.MaxClaimPerUser)
	assert.Equal(t, 500.0, info.TotalAmount)
	assert.Equal(t, "https://e.com/x", info.AttachmentURL)
}

func TestTaskCreate_JSONBFieldsConverted(t *testing.T) {
	repo := newMockTaskRepo()
	svc := NewTaskService(repo)

	info, err := svc.Create(1, 1, &dto.CreateTaskRequest{
		LinggongID:   1,
		Title:        "T",
		Tags:         []string{"a", "b"},
		Requirements: map[string]interface{}{"exp": "3y"},
	})
	require.NoError(t, err)
	assert.NotNil(t, info.Tags)
	assert.NotNil(t, info.Requirements)
	// 仓储内已落 JSONB
	stored := repo.byID[info.ID]
	require.NotNil(t, stored)
	assert.True(t, len(stored.Tags) > 0)
	assert.True(t, len(stored.Requirements) > 0)
}

func TestTaskCreate_RepoError(t *testing.T) {
	repo := newMockTaskRepo()
	repo.createErr = errors.New("db down")
	svc := NewTaskService(repo)
	_, err := svc.Create(1, 1, &dto.CreateTaskRequest{LinggongID: 1, Title: "T"})
	require.Error(t, err)
	assert.Equal(t, "db down", err.Error())
}

// ===== Update =====

func TestTaskUpdate_NotFound(t *testing.T) {
	repo := newMockTaskRepo()
	svc := NewTaskService(repo)
	err := svc.Update(999, 1, &dto.UpdateTaskRequest{})
	assert.ErrorIs(t, err, ErrTaskNotFound)
	assert.Empty(t, repo.updates)
}

func TestTaskUpdate_NoPermission(t *testing.T) {
	repo := newMockTaskRepo()
	repo.seed(&model.LinggongTask{EmployerID: 1})
	svc := NewTaskService(repo)
	title := "new"
	err := svc.Update(1, 2, &dto.UpdateTaskRequest{Title: &title})
	assert.ErrorIs(t, err, ErrTaskNoPermission)
	assert.Empty(t, repo.updates)
}

func TestTaskUpdate_FieldsBuilt(t *testing.T) {
	repo := newMockTaskRepo()
	repo.seed(&model.LinggongTask{EmployerID: 1})
	svc := NewTaskService(repo)

	title := "新标题"
	price := 99.9
	count := 7
	maxClaim := 3
	err := svc.Update(1, 1, &dto.UpdateTaskRequest{
		Title:           &title,
		TaskType:        strPtr(model.TaskTypeBatch),
		Difficulty:      strPtr(model.TaskDifficultyHard),
		DeliveryMethod:  strPtr(model.TaskDeliveryOffline),
		BillingType:     strPtr(model.BillingTypeByHour),
		UnitPrice:       &price,
		TotalCount:      &count,
		MaxClaimPerUser: &maxClaim,
		TotalAmount:     floatPtr(700),
		AttachmentURL:   strPtr("https://e.com/b"),
		Tags:            []string{"x"},
		Requirements:    map[string]interface{}{"k": "v"},
	})
	require.NoError(t, err)
	require.Len(t, repo.updates, 1)
	f := repo.updates[0].fields
	assert.Equal(t, uint(1), repo.updates[0].id)
	assert.Equal(t, "新标题", f["title"])
	assert.Equal(t, model.TaskTypeBatch, f["task_type"])
	assert.Equal(t, model.TaskDifficultyHard, f["difficulty"])
	assert.Equal(t, model.TaskDeliveryOffline, f["delivery_method"])
	assert.Equal(t, model.BillingTypeByHour, f["billing_type"])
	assert.Equal(t, 99.9, f["unit_price"])
	assert.Equal(t, 7, f["total_count"])
	assert.Equal(t, 3, f["max_claim_per_user"])
	assert.Equal(t, 700.0, f["total_amount"])
	assert.Equal(t, "https://e.com/b", f["attachment_url"])
	// JSONB 字段以 model.JSONB 形式写入
	assert.NotNil(t, f["tags"])
	assert.NotNil(t, f["requirements"])
}

func TestTaskUpdate_EmptyRequestNoop(t *testing.T) {
	repo := newMockTaskRepo()
	repo.seed(&model.LinggongTask{EmployerID: 1})
	svc := NewTaskService(repo)
	err := svc.Update(1, 1, &dto.UpdateTaskRequest{})
	require.NoError(t, err)
	assert.Empty(t, repo.updates, "空请求不应触发仓储 Update")
}

func TestTaskUpdate_RepoError(t *testing.T) {
	repo := newMockTaskRepo()
	repo.seed(&model.LinggongTask{EmployerID: 1})
	repo.updateErr = errors.New("update fail")
	svc := NewTaskService(repo)
	title := "x"
	err := svc.Update(1, 1, &dto.UpdateTaskRequest{Title: &title})
	require.Error(t, err)
	assert.Equal(t, "update fail", err.Error())
}

func TestTaskUpdate_FindErrorNonNotFound(t *testing.T) {
	repo := newMockTaskRepo()
	repo.findErr = errors.New("conn lost")
	svc := NewTaskService(repo)
	title := "x"
	err := svc.Update(1, 1, &dto.UpdateTaskRequest{Title: &title})
	require.Error(t, err)
	assert.Equal(t, "conn lost", err.Error())
}

// ===== Delete =====

func TestTaskDelete_NotFound(t *testing.T) {
	repo := newMockTaskRepo()
	svc := NewTaskService(repo)
	err := svc.Delete(999, 1)
	assert.ErrorIs(t, err, ErrTaskNotFound)
}

func TestTaskDelete_NoPermission(t *testing.T) {
	repo := newMockTaskRepo()
	repo.seed(&model.LinggongTask{EmployerID: 1})
	svc := NewTaskService(repo)
	err := svc.Delete(1, 2)
	assert.ErrorIs(t, err, ErrTaskNoPermission)
	assert.Empty(t, repo.deletes)
}

func TestTaskDelete_Success(t *testing.T) {
	repo := newMockTaskRepo()
	repo.seed(&model.LinggongTask{EmployerID: 1})
	svc := NewTaskService(repo)
	err := svc.Delete(1, 1)
	require.NoError(t, err)
	assert.Equal(t, []uint{1}, repo.deletes)
	_, ok := repo.byID[1]
	assert.False(t, ok)
}

func TestTaskDelete_RepoError(t *testing.T) {
	repo := newMockTaskRepo()
	repo.seed(&model.LinggongTask{EmployerID: 1})
	repo.deleteErr = errors.New("del fail")
	svc := NewTaskService(repo)
	err := svc.Delete(1, 1)
	require.Error(t, err)
	assert.Equal(t, "del fail", err.Error())
}

// ===== GetByID =====

func TestTaskGetByID_NotFound(t *testing.T) {
	repo := newMockTaskRepo()
	svc := NewTaskService(repo)
	_, err := svc.GetByID(999)
	assert.ErrorIs(t, err, ErrTaskNotFound)
}

func TestTaskGetByID_Success(t *testing.T) {
	repo := newMockTaskRepo()
	repo.seed(&model.LinggongTask{Title: "T", Status: model.TaskStatusPublished})
	svc := NewTaskService(repo)
	info, err := svc.GetByID(1)
	require.NoError(t, err)
	require.NotNil(t, info)
	assert.Equal(t, "T", info.Title)
	assert.Equal(t, "已发布", info.StatusText)
}

func TestTaskGetByID_FindError(t *testing.T) {
	repo := newMockTaskRepo()
	repo.findErr = errors.New("boom")
	svc := NewTaskService(repo)
	_, err := svc.GetByID(1)
	require.Error(t, err)
	assert.Equal(t, "boom", err.Error())
}

// ===== Claim =====

func TestTaskClaim_NotFound(t *testing.T) {
	repo := newMockTaskRepo()
	svc := NewTaskService(repo)
	err := svc.Claim(999, 1, "w", &dto.TaskClaimRequest{Count: 1})
	assert.ErrorIs(t, err, ErrTaskNotFound)
	assert.Empty(t, repo.incrClaimed)
}

func TestTaskClaim_StatusInvalid(t *testing.T) {
	repo := newMockTaskRepo()
	repo.seed(&model.LinggongTask{Status: model.TaskStatusDraft, TotalCount: 10, MaxClaimPerUser: 5})
	svc := NewTaskService(repo)
	err := svc.Claim(1, 1, "w", &dto.TaskClaimRequest{Count: 1})
	assert.ErrorIs(t, err, ErrTaskStatusInvalid)
}

func TestTaskClaim_StatusCompletedInvalid(t *testing.T) {
	repo := newMockTaskRepo()
	repo.seed(&model.LinggongTask{Status: model.TaskStatusCompleted, TotalCount: 10, MaxClaimPerUser: 5})
	svc := NewTaskService(repo)
	err := svc.Claim(1, 1, "w", &dto.TaskClaimRequest{Count: 1})
	assert.ErrorIs(t, err, ErrTaskStatusInvalid)
}

func TestTaskClaim_DeadlineExceeded(t *testing.T) {
	repo := newMockTaskRepo()
	past := time.Now().Add(-time.Hour)
	repo.seed(&model.LinggongTask{Status: model.TaskStatusPublished, TotalCount: 10, MaxClaimPerUser: 5, ClaimDeadline: &past})
	svc := NewTaskService(repo)
	err := svc.Claim(1, 1, "w", &dto.TaskClaimRequest{Count: 1})
	assert.ErrorIs(t, err, ErrTaskDeadlineExceeded)
}

func TestTaskClaim_StockExceeded(t *testing.T) {
	repo := newMockTaskRepo()
	repo.seed(&model.LinggongTask{Status: model.TaskStatusPublished, TotalCount: 5, ClaimedCount: 4, MaxClaimPerUser: 5})
	svc := NewTaskService(repo)
	err := svc.Claim(1, 1, "w", &dto.TaskClaimRequest{Count: 2})
	assert.ErrorIs(t, err, ErrTaskCountExceeded)
}

func TestTaskClaim_PerUserExceeded(t *testing.T) {
	repo := newMockTaskRepo()
	repo.seed(&model.LinggongTask{Status: model.TaskStatusPublished, TotalCount: 100, MaxClaimPerUser: 2})
	svc := NewTaskService(repo)
	err := svc.Claim(1, 1, "w", &dto.TaskClaimRequest{Count: 3})
	assert.ErrorIs(t, err, ErrTaskCountExceeded)
}

func TestTaskClaim_Success_PublishedTransitionsToInProgress(t *testing.T) {
	repo := newMockTaskRepo()
	repo.seed(&model.LinggongTask{Status: model.TaskStatusPublished, TotalCount: 10, MaxClaimPerUser: 5})
	svc := NewTaskService(repo)

	err := svc.Claim(1, 7, "工人A", &dto.TaskClaimRequest{Count: 2})
	require.NoError(t, err)

	assert.Equal(t, []incrCountCall{{id: 1, count: 2}}, repo.incrClaimed)
	// 已发布 -> 进行中
	require.Len(t, repo.updates, 1)
	assert.Equal(t, model.TaskStatusInProgress, repo.updates[0].fields["status"])
}

func TestTaskClaim_Success_InProgressNoTransition(t *testing.T) {
	repo := newMockTaskRepo()
	repo.seed(&model.LinggongTask{Status: model.TaskStatusInProgress, TotalCount: 10, ClaimedCount: 2, MaxClaimPerUser: 5})
	svc := NewTaskService(repo)

	err := svc.Claim(1, 7, "w", &dto.TaskClaimRequest{Count: 1})
	require.NoError(t, err)
	assert.Equal(t, []incrCountCall{{id: 1, count: 1}}, repo.incrClaimed)
	assert.Empty(t, repo.updates, "进行中状态不应再次迁移")
}

func TestTaskClaim_IncrError(t *testing.T) {
	repo := newMockTaskRepo()
	repo.seed(&model.LinggongTask{Status: model.TaskStatusPublished, TotalCount: 10, MaxClaimPerUser: 5})
	repo.updateErr = errors.New("incr fail")
	svc := NewTaskService(repo)
	err := svc.Claim(1, 1, "w", &dto.TaskClaimRequest{Count: 1})
	require.Error(t, err)
	assert.Equal(t, "incr fail", err.Error())
}

// ===== Submit =====

func TestTaskSubmit_NotFound(t *testing.T) {
	repo := newMockTaskRepo()
	svc := NewTaskService(repo)
	err := svc.Submit(999, 1, &dto.TaskSubmitRequest{Count: 1})
	assert.ErrorIs(t, err, ErrTaskNotFound)
}

func TestTaskSubmit_StatusInvalid(t *testing.T) {
	repo := newMockTaskRepo()
	repo.seed(&model.LinggongTask{Status: model.TaskStatusCompleted})
	svc := NewTaskService(repo)
	err := svc.Submit(1, 1, &dto.TaskSubmitRequest{Count: 1})
	assert.ErrorIs(t, err, ErrTaskStatusInvalid)
}

func TestTaskSubmit_DeadlineExceeded(t *testing.T) {
	repo := newMockTaskRepo()
	past := time.Now().Add(-time.Hour)
	repo.seed(&model.LinggongTask{Status: model.TaskStatusInProgress, SubmitDeadline: &past})
	svc := NewTaskService(repo)
	err := svc.Submit(1, 1, &dto.TaskSubmitRequest{Count: 1})
	assert.ErrorIs(t, err, ErrTaskDeadlineExceeded)
}

func TestTaskSubmit_Success(t *testing.T) {
	repo := newMockTaskRepo()
	repo.seed(&model.LinggongTask{Status: model.TaskStatusInProgress, TotalCount: 10})
	svc := NewTaskService(repo)
	err := svc.Submit(1, 1, &dto.TaskSubmitRequest{Count: 3, AttachmentURL: "https://e.com/r.zip", Remark: "已完成"})
	require.NoError(t, err)
	assert.Equal(t, []incrCountCall{{id: 1, count: 3}}, repo.incrCompleted)
}

func TestTaskSubmit_IncrError(t *testing.T) {
	repo := newMockTaskRepo()
	repo.seed(&model.LinggongTask{Status: model.TaskStatusInProgress})
	repo.updateErr = errors.New("incr fail")
	svc := NewTaskService(repo)
	err := svc.Submit(1, 1, &dto.TaskSubmitRequest{Count: 1})
	require.Error(t, err)
	assert.Equal(t, "incr fail", err.Error())
}

// ===== Verify =====

func TestTaskVerify_NotFound(t *testing.T) {
	repo := newMockTaskRepo()
	svc := NewTaskService(repo)
	err := svc.Verify(999, 1, &dto.TaskVerifyRequest{Count: 1, Pass: true})
	assert.ErrorIs(t, err, ErrTaskNotFound)
}

func TestTaskVerify_NoPermission(t *testing.T) {
	repo := newMockTaskRepo()
	repo.seed(&model.LinggongTask{EmployerID: 1, Status: model.TaskStatusInProgress})
	svc := NewTaskService(repo)
	err := svc.Verify(1, 2, &dto.TaskVerifyRequest{Count: 1, Pass: true})
	assert.ErrorIs(t, err, ErrTaskNoPermission)
	assert.Empty(t, repo.incrVerified)
}

func TestTaskVerify_StatusInvalid(t *testing.T) {
	repo := newMockTaskRepo()
	repo.seed(&model.LinggongTask{EmployerID: 1, Status: model.TaskStatusCompleted})
	svc := NewTaskService(repo)
	err := svc.Verify(1, 1, &dto.TaskVerifyRequest{Count: 1, Pass: true})
	assert.ErrorIs(t, err, ErrTaskStatusInvalid)
}

func TestTaskVerify_Pass_IncrementsVerifiedAndPaid(t *testing.T) {
	repo := newMockTaskRepo()
	repo.seed(&model.LinggongTask{
		EmployerID:    1,
		Status:        model.TaskStatusInProgress,
		UnitPrice:     20.0,
		TotalCount:    10,
		VerifiedCount: 1,
	})
	svc := NewTaskService(repo)

	err := svc.Verify(1, 1, &dto.TaskVerifyRequest{Count: 2, Pass: true})
	require.NoError(t, err)

	assert.Equal(t, []incrCountCall{{id: 1, count: 2}}, repo.incrVerified)
	assert.Equal(t, []incrAmountCall{{id: 1, amount: 40.0}}, repo.incrPaidAmount)
	// 未达总数，不迁移完成态
	assert.Empty(t, repo.updates)
}

func TestTaskVerify_Pass_CompletesWhenReachingTotal(t *testing.T) {
	repo := newMockTaskRepo()
	repo.seed(&model.LinggongTask{
		EmployerID:    1,
		Status:        model.TaskStatusInProgress,
		UnitPrice:     10.0,
		TotalCount:    5,
		VerifiedCount: 3,
	})
	svc := NewTaskService(repo)

	err := svc.Verify(1, 1, &dto.TaskVerifyRequest{Count: 2, Pass: true})
	require.NoError(t, err)

	assert.Equal(t, []incrCountCall{{id: 1, count: 2}}, repo.incrVerified)
	assert.Equal(t, []incrAmountCall{{id: 1, amount: 20.0}}, repo.incrPaidAmount)
	// 已验收数(3)+本次(2) >= 总数(5) -> 完成态迁移
	require.Len(t, repo.updates, 1)
	assert.Equal(t, model.TaskStatusCompleted, repo.updates[0].fields["status"])
	ts, ok := repo.updates[0].fields["completed_at"].(*time.Time)
	require.True(t, ok, "completed_at 应为 *time.Time")
	assert.NotNil(t, ts)
}

func TestTaskVerify_Pass_IncrVerifiedError(t *testing.T) {
	repo := newMockTaskRepo()
	repo.seed(&model.LinggongTask{
		EmployerID: 1, Status: model.TaskStatusInProgress, UnitPrice: 10, TotalCount: 10, VerifiedCount: 0,
	})
	repo.updateErr = errors.New("incr fail")
	svc := NewTaskService(repo)
	err := svc.Verify(1, 1, &dto.TaskVerifyRequest{Count: 1, Pass: true})
	require.Error(t, err)
	assert.Equal(t, "incr fail", err.Error())
	// 验收计数失败时应短路，不触发支付金额累加
	assert.Empty(t, repo.incrPaidAmount)
}

func TestTaskVerify_RejectNoIncrements(t *testing.T) {
	repo := newMockTaskRepo()
	repo.seed(&model.LinggongTask{EmployerID: 1, Status: model.TaskStatusInProgress, UnitPrice: 10, TotalCount: 10, VerifiedCount: 0})
	svc := NewTaskService(repo)
	err := svc.Verify(1, 1, &dto.TaskVerifyRequest{Count: 2, Pass: false, RejectReason: "不合格"})
	require.NoError(t, err)
	assert.Empty(t, repo.incrVerified)
	assert.Empty(t, repo.incrPaidAmount)
	assert.Empty(t, repo.updates)
}

// ===== AdminGetByID =====

func TestTaskAdminGetByID_NotFound(t *testing.T) {
	repo := newMockTaskRepo()
	svc := NewTaskService(repo)
	_, err := svc.AdminGetByID(999)
	assert.ErrorIs(t, err, ErrTaskNotFound)
}

func TestTaskAdminGetByID_Success(t *testing.T) {
	repo := newMockTaskRepo()
	repo.seed(&model.LinggongTask{Title: "A", Status: model.TaskStatusCompleted})
	svc := NewTaskService(repo)
	info, err := svc.AdminGetByID(1)
	require.NoError(t, err)
	assert.Equal(t, "A", info.Title)
	assert.Equal(t, "已完成", info.StatusText)
}

// ===== UpdateStatus =====

func TestTaskUpdateStatus_PublishedSetsPublishedAt(t *testing.T) {
	repo := newMockTaskRepo()
	svc := NewTaskService(repo)
	err := svc.UpdateStatus(1, model.TaskStatusPublished)
	require.NoError(t, err)
	require.Len(t, repo.updates, 1)
	f := repo.updates[0].fields
	assert.Equal(t, model.TaskStatusPublished, f["status"])
	ts, ok := f["published_at"].(*time.Time)
	require.True(t, ok)
	assert.NotNil(t, ts)
	_, hasCompleted := f["completed_at"]
	assert.False(t, hasCompleted)
}

func TestTaskUpdateStatus_CompletedSetsCompletedAt(t *testing.T) {
	repo := newMockTaskRepo()
	svc := NewTaskService(repo)
	err := svc.UpdateStatus(1, model.TaskStatusCompleted)
	require.NoError(t, err)
	require.Len(t, repo.updates, 1)
	f := repo.updates[0].fields
	assert.Equal(t, model.TaskStatusCompleted, f["status"])
	ts, ok := f["completed_at"].(*time.Time)
	require.True(t, ok)
	assert.NotNil(t, ts)
	_, hasPublished := f["published_at"]
	assert.False(t, hasPublished)
}

func TestTaskUpdateStatus_OtherStatusOnlyStatus(t *testing.T) {
	repo := newMockTaskRepo()
	svc := NewTaskService(repo)
	err := svc.UpdateStatus(1, model.TaskStatusCanceled)
	require.NoError(t, err)
	require.Len(t, repo.updates, 1)
	f := repo.updates[0].fields
	assert.Equal(t, model.TaskStatusCanceled, f["status"])
	_, hasPublished := f["published_at"]
	assert.False(t, hasPublished)
	_, hasCompleted := f["completed_at"]
	assert.False(t, hasCompleted)
}

func TestTaskUpdateStatus_RepoError(t *testing.T) {
	repo := newMockTaskRepo()
	repo.updateErr = errors.New("up fail")
	svc := NewTaskService(repo)
	err := svc.UpdateStatus(1, model.TaskStatusPublished)
	require.Error(t, err)
	assert.Equal(t, "up fail", err.Error())
}

// ===== List =====

func TestTaskList_DefaultStatusFilterPublished(t *testing.T) {
	repo := newMockTaskRepo()
	repo.listReturn = []model.LinggongTask{{Title: "a"}}
	repo.listTotal = 1
	svc := NewTaskService(repo)

	pag, list, err := svc.List(&dto.TaskListRequest{})
	require.NoError(t, err)
	require.NotNil(t, pag)
	assert.Equal(t, int64(1), pag.Total)
	require.Len(t, list, 1)
	assert.Equal(t, "a", list[0].Title)
}

func TestTaskList_RepoError(t *testing.T) {
	repo := newMockTaskRepo()
	repo.listErr = errors.New("list fail")
	svc := NewTaskService(repo)
	_, _, err := svc.List(&dto.TaskListRequest{})
	require.Error(t, err)
	assert.Equal(t, "list fail", err.Error())
}

func TestTaskListByLinggong(t *testing.T) {
	repo := newMockTaskRepo()
	repo.byLinggongList = []model.LinggongTask{{Title: "a"}, {Title: "b"}}
	repo.byLinggongTotal = 2
	svc := NewTaskService(repo)
	pag, list, err := svc.ListByLinggong(5, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(2), pag.Total)
	assert.Len(t, list, 2)
}

func TestTaskListByEmployer(t *testing.T) {
	repo := newMockTaskRepo()
	repo.byEmployerList = []model.LinggongTask{{Title: "c"}}
	repo.byEmployerTotal = 1
	svc := NewTaskService(repo)
	pag, list, err := svc.ListByEmployer(9, 2, 5)
	require.NoError(t, err)
	assert.Equal(t, int64(1), pag.Total)
	assert.Len(t, list, 1)
}

func TestTaskAdminList(t *testing.T) {
	repo := newMockTaskRepo()
	repo.adminListReturn = []model.LinggongTask{{Title: "x"}}
	repo.adminListTotal = 1
	svc := NewTaskService(repo)
	pag, list, err := svc.AdminList(&dto.TaskAdminListRequest{RegionID: 1})
	require.NoError(t, err)
	assert.Equal(t, int64(1), pag.Total)
	require.Len(t, list, 1)
	assert.Equal(t, "x", list[0].Title)
}

func TestTaskAdminList_RepoError(t *testing.T) {
	repo := newMockTaskRepo()
	repo.listErr = errors.New("admin fail")
	svc := NewTaskService(repo)
	_, _, err := svc.AdminList(&dto.TaskAdminListRequest{})
	require.Error(t, err)
	assert.Equal(t, "admin fail", err.Error())
}

// ===== helpers =====

func strPtr(s string) *string { return &s }
func floatPtr(f float64) *float64 { return &f }
