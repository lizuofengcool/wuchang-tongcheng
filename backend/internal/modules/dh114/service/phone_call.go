// Package service 同城114业务逻辑层 - 电话拨打记录
// 依据 v3.2.1 架构方案：一键拨号核心
// 记录用户点击拨号/直接拨打的次数与设备信息
package service

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"wuchang-tongcheng/internal/modules/dh114/dto"
	"wuchang-tongcheng/internal/modules/dh114/model"
	"wuchang-tongcheng/internal/modules/dh114/repository"
	"wuchang-tongcheng/internal/pkg/utils"
)

var (
	ErrPhoneCallNotFound = errors.New("电话拨打记录不存在")
)

// PhoneCallService 电话拨打业务接口
type PhoneCallService interface {
	// 记录拨打
	Create(regionID uint, userID uint, req *dto.PhoneCallRequest, ip, userAgent string) (*dto.PhoneCallInfo, error)
	GetByID(id uint) (*dto.PhoneCallInfo, error)
	List(req *dto.PhoneCallListRequest) (*utils.Pagination, []dto.PhoneCallInfo, error)
	ListByDh114(dh114ID uint, page, pageSize int) (*utils.Pagination, []dto.PhoneCallInfo, error)
	ListByCaller(callerID uint, page, pageSize int) (*utils.Pagination, []dto.PhoneCallInfo, error)

	// 统计
	CountByDh114(dh114ID uint) (int64, error)
	CountTodayByDh114(dh114ID uint) (int64, error)

	// M 端管理
	AdminList(req *dto.PhoneCallAdminListRequest) (*utils.Pagination, []dto.PhoneCallInfo, error)
}

type phoneCallService struct {
	repo repository.PhoneCallRepository
}

// NewPhoneCallService 创建电话拨打 service 实例
func NewPhoneCallService(repo repository.PhoneCallRepository) PhoneCallService {
	return &phoneCallService{repo: repo}
}

// toPhoneCallInfo model -> dto
func toPhoneCallInfo(c *model.Dh114PhoneCall) *dto.PhoneCallInfo {
	return &dto.PhoneCallInfo{
		ID:          c.ID,
		CallNo:      c.CallNo,
		Dh114ID:     c.Dh114ID,
		BusinessID:  c.BusinessID,
		Phone:       c.Phone,
		CallerID:    c.CallerID,
		CallerPhone: c.CallerPhone,
		CallerName:  c.CallerName,
		CallType:    c.CallType,
		Device:      c.Device,
		IP:          c.IP,
		Status:      c.Status,
		Duration:    c.Duration,
		CalledAt:    c.CalledAt,
		RegionID:    c.RegionID,
		CreatedAt:   c.CreatedAt,
	}
}

// generateCallNo 生成拨打单号
func generateCallNo() string {
	return fmt.Sprintf("DH114CL%s%06d", time.Now().Format("20060102150405"), rand.Intn(1000000))
}

// Create 记录电话拨打
func (s *phoneCallService) Create(regionID uint, userID uint, req *dto.PhoneCallRequest, ip, userAgent string) (*dto.PhoneCallInfo, error) {
	callType := req.CallType
	if callType == "" {
		callType = model.CallTypeClick
	}

	c := &model.Dh114PhoneCall{
		CallNo:     generateCallNo(),
		Dh114ID:    req.Dh114ID,
		Phone:      "", // 由 handler 层根据 dh114_id 查询商户后填充
		CallerID:   userID,
		Device:     req.Device,
		IP:         ip,
		UserAgent:  userAgent,
		Status:     model.CallStatusSuccess,
		CalledAt:   time.Now(),
		CallType:   callType,
	}
	c.RegionID = regionID

	if err := s.repo.Create(c); err != nil {
		return nil, err
	}
	return toPhoneCallInfo(c), nil
}

// GetByID 获取拨打记录详情
func (s *phoneCallService) GetByID(id uint) (*dto.PhoneCallInfo, error) {
	c, err := s.repo.FindByID(id)
	if err != nil {
		return nil, ErrPhoneCallNotFound
	}
	return toPhoneCallInfo(c), nil
}

// List 拨打记录列表
func (s *phoneCallService) List(req *dto.PhoneCallListRequest) (*utils.Pagination, []dto.PhoneCallInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	query := repository.PhoneCallListQuery{
		Dh114ID:  req.Dh114ID,
		CallerID: req.CallerID,
		CallType: req.CallType,
		Status:   req.Status,
	}
	list, total, err := s.repo.List(query, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.PhoneCallInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toPhoneCallInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// ListByDh114 按商户列出拨打记录
func (s *phoneCallService) ListByDh114(dh114ID uint, page, pageSize int) (*utils.Pagination, []dto.PhoneCallInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByDh114(dh114ID, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.PhoneCallInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toPhoneCallInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// ListByCaller 按主叫列出拨打记录
func (s *phoneCallService) ListByCaller(callerID uint, page, pageSize int) (*utils.Pagination, []dto.PhoneCallInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByCaller(callerID, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.PhoneCallInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toPhoneCallInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// CountByDh114 商户总拨打次数
func (s *phoneCallService) CountByDh114(dh114ID uint) (int64, error) {
	return s.repo.CountByDh114(dh114ID)
}

// CountTodayByDh114 商户今日拨打次数
func (s *phoneCallService) CountTodayByDh114(dh114ID uint) (int64, error) {
	return s.repo.CountTodayByDh114(dh114ID)
}

// AdminList 管理后台列表
func (s *phoneCallService) AdminList(req *dto.PhoneCallAdminListRequest) (*utils.Pagination, []dto.PhoneCallInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	query := repository.PhoneCallListQuery{
		Dh114ID:  req.Dh114ID,
		CallerID: req.CallerID,
		CallType: req.CallType,
		Status:   req.Status,
	}
	list, total, err := s.repo.List(query, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.PhoneCallInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toPhoneCallInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}
