// Package service 同城零工兼职业务逻辑层 - 报名记录
// 对标斗米/兼职猫报名流程：报名→审核→到岗→完成
// 4 维数据隔离（region_id + user_id）
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
	ErrApplicationNotFound     = errors.New("报名记录不存在")
	ErrApplicationNoPermission = errors.New("无权操作此报名记录")
	ErrApplicationStatusInvalid = errors.New("报名状态不允许此操作")
	ErrApplicationDuplicate    = errors.New("已报名过此岗位")
)

// ApplicationService 报名业务接口
type ApplicationService interface {
	// C 端
	Create(regionID uint, userID uint, workerName string, workerAvatar string, workerPhone string, req *dto.CreateApplicationRequest) (*dto.ApplicationInfo, error)
	Update(id uint, operatorID uint, req *dto.UpdateApplicationRequest) error
	Delete(id uint, operatorID uint) error
	GetByID(id uint) (*dto.ApplicationInfo, error)
	List(regionID uint, req *dto.ApplicationListRequest) (*utils.Pagination, []dto.ApplicationInfo, error)
	ListByLinggong(linggongID uint, page, pageSize int) (*utils.Pagination, []dto.ApplicationInfo, error)
	ListByWorker(workerID uint, page, pageSize int) (*utils.Pagination, []dto.ApplicationInfo, error)
	ListByEmployer(employerID uint, page, pageSize int) (*utils.Pagination, []dto.ApplicationInfo, error)

	// 报名审核/取消
	Audit(id uint, employerID uint, req *dto.ApplicationAuditRequest) error
	Cancel(id uint, workerID uint, req *dto.ApplicationCancelRequest) error

	// M 端管理
	AdminList(req *dto.ApplicationAdminListRequest) (*utils.Pagination, []dto.ApplicationInfo, error)
}

type applicationService struct {
	repo repository.ApplicationRepository
}

// NewApplicationService 创建报名 service 实例
func NewApplicationService(repo repository.ApplicationRepository) ApplicationService {
	return &applicationService{repo: repo}
}

// applicationStatusText 报名状态文本
func applicationStatusText(status int) string {
	switch status {
	case model.ApplicationStatusPending:
		return "待审核"
	case model.ApplicationStatusApproved:
		return "已通过"
	case model.ApplicationStatusRejected:
		return "已拒绝"
	case model.ApplicationStatusCanceled:
		return "已取消"
	case model.ApplicationStatusExpired:
		return "已过期"
	case model.ApplicationStatusConfirmed:
		return "已确认"
	case model.ApplicationStatusNoShow:
		return "未到岗"
	case model.ApplicationStatusWorking:
		return "工作中"
	case model.ApplicationStatusCompleted:
		return "已完成"
	case model.ApplicationStatusQuit:
		return "已离职"
	case model.ApplicationStatusFired:
		return "已辞退"
	}
	return ""
}

// applicationSourceText 报名来源文本
func applicationSourceText(s string) string {
	switch s {
	case model.ApplicationSourceSearch:
		return "搜索"
	case model.ApplicationSourceRecommend:
		return "推荐"
	case model.ApplicationSourceShare:
		return "分享"
	case model.ApplicationSourceDirect:
		return "直接报名"
	case model.ApplicationSourceInvite:
		return "雇主邀请"
	case model.ApplicationSourceFavorite:
		return "收藏触发"
	}
	return ""
}

// applicationMethodText 报名方式文本
func applicationMethodText(m string) string {
	switch m {
	case model.ApplicationMethodOnline:
		return "在线报名"
	case model.ApplicationMethodPhone:
		return "电话报名"
	case model.ApplicationMethodOnsite:
		return "现场报名"
	}
	return ""
}

// toApplicationInfo model -> dto
func toApplicationInfo(a *model.LinggongApplication) *dto.ApplicationInfo {
	return &dto.ApplicationInfo{
		ID:                a.ID,
		ApplicationNo:     a.ApplicationNo,
		LinggongID:        a.LinggongID,
		TaskID:            a.TaskID,
		EmployerID:        a.EmployerID,
		EmployerName:      a.EmployerName,
		WorkerID:          a.WorkerID,
		WorkerName:        a.WorkerName,
		WorkerAvatar:      a.WorkerAvatar,
		WorkerPhone:       a.WorkerPhone,
		WorkerAge:         a.WorkerAge,
		WorkerGender:      a.WorkerGender,
		WorkerCity:        a.WorkerCity,
		WorkerCreditScore: a.WorkerCreditScore,
		WorkerProfileID:   a.WorkerProfileID,
		Source:            a.Source,
		SourceText:        applicationSourceText(a.Source),
		Method:            a.Method,
		MethodText:        applicationMethodText(a.Method),
		Status:            a.Status,
		StatusText:        applicationStatusText(a.Status),
		AppliedCount:      a.AppliedCount,
		CoverLetter:       a.CoverLetter,
		EmployerRemark:    a.EmployerRemark,
		RejectReason:      a.RejectReason,
		CancelReason:      a.CancelReason,
		ReviewedAt:        a.ReviewedAt,
		ConfirmedAt:       a.ConfirmedAt,
		OnboardedAt:       a.OnboardedAt,
		CompletedAt:       a.CompletedAt,
		CanceledAt:        a.CanceledAt,
		Evaluated:         a.Evaluated,
		AttachmentURL:     a.AttachmentURL,
		RegionID:          a.RegionID,
		CreatedAt:         a.CreatedAt,
		UpdatedAt:         a.UpdatedAt,
	}
}

// genApplicationNo 生成报名单号：APP + yyyyMMddHHmmss + 6 位随机
func genApplicationNo() string {
	return fmt.Sprintf("APP%s%06d", time.Now().Format("20060102150405"), rand.Intn(1000000))
}

// ===== C 端 =====

// Create 创建报名
func (s *applicationService) Create(regionID uint, userID uint, workerName string, workerAvatar string, workerPhone string, req *dto.CreateApplicationRequest) (*dto.ApplicationInfo, error) {
	a := &model.LinggongApplication{
		ApplicationNo: genApplicationNo(),
		LinggongID:    req.LinggongID,
		TaskID:        req.TaskID,
		WorkerID:      userID,
		WorkerName:    workerName,
		WorkerAvatar:  workerAvatar,
		WorkerPhone:   workerPhone,
		Source:        req.Source,
		Method:        req.Method,
		Status:        model.ApplicationStatusPending,
		AppliedCount:  1,
		CoverLetter:   req.CoverLetter,
		AttachmentURL: req.AttachmentURL,
	}
	a.RegionID = regionID

	// 默认值兜底
	if a.Source == "" {
		a.Source = model.ApplicationSourceDirect
	}
	if a.Method == "" {
		a.Method = model.ApplicationMethodOnline
	}

	if err := s.repo.Create(a); err != nil {
		return nil, err
	}
	return toApplicationInfo(a), nil
}

// Update 更新报名（仅雇主可备注）
func (s *applicationService) Update(id uint, operatorID uint, req *dto.UpdateApplicationRequest) error {
	a, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrApplicationNotFound
		}
		return err
	}
	if a.EmployerID != operatorID {
		return ErrApplicationNoPermission
	}

	fields := map[string]interface{}{}
	if req.EmployerRemark != nil {
		fields["employer_remark"] = *req.EmployerRemark
	}
	if req.Status != nil {
		fields["status"] = *req.Status
	}

	if len(fields) == 0 {
		return nil
	}
	return s.repo.Update(id, fields)
}

// Delete 删除报名（仅求职者本人）
func (s *applicationService) Delete(id uint, operatorID uint) error {
	a, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrApplicationNotFound
		}
		return err
	}
	if a.WorkerID != operatorID {
		return ErrApplicationNoPermission
	}
	return s.repo.Delete(id)
}

// GetByID 获取报名详情
func (s *applicationService) GetByID(id uint) (*dto.ApplicationInfo, error) {
	a, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrApplicationNotFound
		}
		return nil, err
	}
	return toApplicationInfo(a), nil
}

// List C 端列表查询
func (s *applicationService) List(regionID uint, req *dto.ApplicationListRequest) (*utils.Pagination, []dto.ApplicationInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.ApplicationListOptions{
		LinggongID: req.LinggongID,
		TaskID:     req.TaskID,
		EmployerID: req.EmployerID,
		WorkerID:   req.WorkerID,
		Status:     req.Status,
		Source:     req.Source,
		Keyword:    req.Keyword,
	}
	list, total, err := s.repo.List(regionID, pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.ApplicationInfo, 0, len(list))
	for i := range list {
		result = append(result, *toApplicationInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListByLinggong 按岗位反查
func (s *applicationService) ListByLinggong(linggongID uint, page, pageSize int) (*utils.Pagination, []dto.ApplicationInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByLinggong(linggongID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.ApplicationInfo, 0, len(list))
	for i := range list {
		result = append(result, *toApplicationInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListByWorker 按求职者反查
func (s *applicationService) ListByWorker(workerID uint, page, pageSize int) (*utils.Pagination, []dto.ApplicationInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByWorker(workerID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.ApplicationInfo, 0, len(list))
	for i := range list {
		result = append(result, *toApplicationInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListByEmployer 按雇主反查
func (s *applicationService) ListByEmployer(employerID uint, page, pageSize int) (*utils.Pagination, []dto.ApplicationInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByEmployer(employerID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.ApplicationInfo, 0, len(list))
	for i := range list {
		result = append(result, *toApplicationInfo(&list[i]))
	}
	return pagination, result, nil
}

// Audit 雇主审核报名
func (s *applicationService) Audit(id uint, employerID uint, req *dto.ApplicationAuditRequest) error {
	a, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrApplicationNotFound
		}
		return err
	}
	if a.EmployerID != employerID {
		return ErrApplicationNoPermission
	}
	if a.Status != model.ApplicationStatusPending {
		return ErrApplicationStatusInvalid
	}

	fields := map[string]interface{}{
		"status":          req.Status,
		"employer_remark": req.EmployerRemark,
		"reject_reason":   req.RejectReason,
	}

	now := time.Now()
	switch req.Status {
	case model.ApplicationStatusApproved:
		fields["reviewed_at"] = &now
	case model.ApplicationStatusRejected:
		fields["reviewed_at"] = &now
	case model.ApplicationStatusConfirmed:
		fields["confirmed_at"] = &now
	case model.ApplicationStatusNoShow:
		fields["reviewed_at"] = &now
	case model.ApplicationStatusWorking:
		fields["onboarded_at"] = &now
	case model.ApplicationStatusCompleted:
		fields["completed_at"] = &now
	case model.ApplicationStatusQuit, model.ApplicationStatusFired:
		// 离职/辞退状态
	}

	return s.repo.Update(id, fields)
}

// Cancel 求职者取消报名
func (s *applicationService) Cancel(id uint, workerID uint, req *dto.ApplicationCancelRequest) error {
	a, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrApplicationNotFound
		}
		return err
	}
	if a.WorkerID != workerID {
		return ErrApplicationNoPermission
	}
	if a.Status == model.ApplicationStatusCompleted || a.Status == model.ApplicationStatusWorking {
		return ErrApplicationStatusInvalid
	}

	now := time.Now()
	fields := map[string]interface{}{
		"status":        model.ApplicationStatusCanceled,
		"cancel_reason": req.CancelReason,
		"canceled_at":   &now,
	}
	return s.repo.Update(id, fields)
}

// ===== M 端管理 =====

// AdminList M 端报名列表
func (s *applicationService) AdminList(req *dto.ApplicationAdminListRequest) (*utils.Pagination, []dto.ApplicationInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.ApplicationAdminListOptions{
		RegionID:   req.RegionID,
		LinggongID: req.LinggongID,
		EmployerID: req.EmployerID,
		WorkerID:   req.WorkerID,
		Status:     req.Status,
		Keyword:    req.Keyword,
	}
	list, total, err := s.repo.AdminList(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.ApplicationInfo, 0, len(list))
	for i := range list {
		result = append(result, *toApplicationInfo(&list[i]))
	}
	return pagination, result, nil
}
