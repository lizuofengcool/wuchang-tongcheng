// Package service 同城零工兼职业务逻辑层 - 考勤打卡
// 对标钉钉/企业微信考勤：GPS/WiFi/人脸/二维码 + 工时统计
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
	ErrAttendanceNotFound = errors.New("考勤记录不存在")
)

// AttendanceService 考勤业务接口
type AttendanceService interface {
	// C 端
	ClockIn(regionID uint, userID uint, workerName string, req *dto.ClockInRequest) (*dto.AttendanceInfo, error)
	GetByID(id uint) (*dto.AttendanceInfo, error)
	List(regionID uint, req *dto.AttendanceListRequest) (*utils.Pagination, []dto.AttendanceInfo, error)
	ListByApplication(applicationID uint, page, pageSize int) (*utils.Pagination, []dto.AttendanceInfo, error)
	ListByLinggong(linggongID uint, page, pageSize int) (*utils.Pagination, []dto.AttendanceInfo, error)
	ListByWorker(workerID uint, startDate, endDate string, page, pageSize int) (*utils.Pagination, []dto.AttendanceInfo, error)

	// M 端管理
	UpdateStatus(id uint, approved bool, approvedBy uint, approvedByName string) error
}

type attendanceService struct {
	repo repository.AttendanceRepository
}

// NewAttendanceService 创建考勤 service 实例
func NewAttendanceService(repo repository.AttendanceRepository) AttendanceService {
	return &attendanceService{repo: repo}
}

// attendanceTypeText 考勤类型文本
func attendanceTypeText(t string) string {
	switch t {
	case model.AttendanceTypeClockIn:
		return "上班打卡"
	case model.AttendanceTypeClockOut:
		return "下班打卡"
	case model.AttendanceTypeBreak:
		return "休息开始"
	case model.AttendanceTypeResume:
		return "休息结束"
	case model.AttendanceTypeOvertime:
		return "加班"
	}
	return ""
}

// attendanceStatusText 考勤状态文本
func attendanceStatusText(s int) string {
	switch s {
	case model.AttendanceStatusNormal:
		return "正常"
	case model.AttendanceStatusLate:
		return "迟到"
	case model.AttendanceStatusEarlyLeave:
		return "早退"
	case model.AttendanceStatusAbsent:
		return "缺勤"
	case model.AttendanceStatusLeave:
		return "请假"
	case model.AttendanceStatusBusinessTrip:
		return "出差"
	case model.AttendanceStatusRemote:
		return "远程"
	case model.AttendanceStatusOut:
		return "外勤"
	}
	return ""
}

// clockMethodText 打卡方式文本
func clockMethodText(m string) string {
	switch m {
	case model.ClockMethodGPS:
		return "GPS定位"
	case model.ClockMethodWiFi:
		return "WiFi"
	case model.ClockMethodFace:
		return "人脸识别"
	case model.ClockMethodManual:
		return "人工补卡"
	case model.ClockMethodQRCode:
		return "二维码"
	}
	return ""
}

// toAttendanceInfo model -> dto
func toAttendanceInfo(a *model.LinggongAttendance) *dto.AttendanceInfo {
	return &dto.AttendanceInfo{
		ID:                 a.ID,
		AttendanceNo:       a.AttendanceNo,
		ApplicationID:      a.ApplicationID,
		LinggongID:         a.LinggongID,
		WorkerID:           a.WorkerID,
		WorkerName:         a.WorkerName,
		EmployerID:         a.EmployerID,
		AttendanceType:     a.AttendanceType,
		AttendanceTypeText: attendanceTypeText(a.AttendanceType),
		ClockTime:          &a.ClockTime,
		ClockMethod:        a.ClockMethod,
		ClockMethodText:    clockMethodText(a.ClockMethod),
		Latitude:           a.Latitude,
		Longitude:          a.Longitude,
		Address:            a.Address,
		WifiSSID:           a.WifiName,
		FaceConfidence:     0,
		PhotoURL:           a.FaceImageURL,
		Status:             fmt.Sprintf("%d", a.Status),
		StatusText:         attendanceStatusText(a.Status),
		WorkHours:          a.WorkHours,
		Remark:             a.Remark,
		RegionID:           a.RegionID,
		CreatedAt:          a.CreatedAt,
	}
}

// genAttendanceNo 生成考勤单号：ATT + yyyyMMddHHmmss + 6 位随机
func genAttendanceNo() string {
	return fmt.Sprintf("ATT%s%06d", time.Now().Format("20060102150405"), rand.Intn(1000000))
}

// ===== C 端 =====

// ClockIn 打卡
func (s *attendanceService) ClockIn(regionID uint, userID uint, workerName string, req *dto.ClockInRequest) (*dto.AttendanceInfo, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	a := &model.LinggongAttendance{
		AttendanceNo:   genAttendanceNo(),
		LinggongID:     0, // 由调用方通过 application 反查后填充（此处简化直接 0）
		ApplicationID:  req.ApplicationID,
		WorkerID:       userID,
		WorkerName:     workerName,
		AttendanceType: req.AttendanceType,
		ClockMethod:    req.ClockMethod,
		ClockTime:      now,
		ClockDate:      &today,
		Latitude:       req.Latitude,
		Longitude:      req.Longitude,
		Address:        req.Address,
		WifiName:       req.WifiSSID,
		FaceImageURL:   req.PhotoURL,
		Status:         model.AttendanceStatusNormal,
		Remark:         req.Remark,
	}
	a.RegionID = regionID

	// 默认值兜底
	if a.AttendanceType == "" {
		a.AttendanceType = model.AttendanceTypeClockIn
	}
	if a.ClockMethod == "" {
		a.ClockMethod = model.ClockMethodGPS
	}

	if err := s.repo.Create(a); err != nil {
		return nil, err
	}
	return toAttendanceInfo(a), nil
}

// GetByID 获取考勤详情
func (s *attendanceService) GetByID(id uint) (*dto.AttendanceInfo, error) {
	a, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAttendanceNotFound
		}
		return nil, err
	}
	return toAttendanceInfo(a), nil
}

// List C 端列表查询
func (s *attendanceService) List(regionID uint, req *dto.AttendanceListRequest) (*utils.Pagination, []dto.AttendanceInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.AttendanceListOptions{
		ApplicationID:  req.ApplicationID,
		LinggongID:    req.LinggongID,
		WorkerID:      req.WorkerID,
		EmployerID:    req.EmployerID,
		AttendanceType: req.AttendanceType,
		Status:        req.Status,
	}
	list, total, err := s.repo.List(regionID, pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.AttendanceInfo, 0, len(list))
	for i := range list {
		result = append(result, *toAttendanceInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListByApplication 按报名反查
func (s *attendanceService) ListByApplication(applicationID uint, page, pageSize int) (*utils.Pagination, []dto.AttendanceInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByApplication(applicationID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.AttendanceInfo, 0, len(list))
	for i := range list {
		result = append(result, *toAttendanceInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListByLinggong 按岗位反查
func (s *attendanceService) ListByLinggong(linggongID uint, page, pageSize int) (*utils.Pagination, []dto.AttendanceInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByLinggong(linggongID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.AttendanceInfo, 0, len(list))
	for i := range list {
		result = append(result, *toAttendanceInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListByWorker 按求职者反查（带日期范围）
func (s *attendanceService) ListByWorker(workerID uint, startDate, endDate string, page, pageSize int) (*utils.Pagination, []dto.AttendanceInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	var start, end time.Time
	if startDate != "" {
		if t, err := time.ParseInLocation("2006-01-02", startDate, time.Local); err == nil {
			start = t
		}
	}
	if endDate != "" {
		if t, err := time.ParseInLocation("2006-01-02", endDate, time.Local); err == nil {
			end = t.Add(24 * time.Hour)
		}
	}
	list, total, err := s.repo.ListByWorker(workerID, start, end, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.AttendanceInfo, 0, len(list))
	for i := range list {
		result = append(result, *toAttendanceInfo(&list[i]))
	}
	return pagination, result, nil
}

// ===== M 端管理 =====

// UpdateStatus M 端审核考勤
func (s *attendanceService) UpdateStatus(id uint, approved bool, approvedBy uint, approvedByName string) error {
	now := time.Now()
	fields := map[string]interface{}{
		"approved":         approved,
		"approved_by":      approvedBy,
		"approved_at":      &now,
	}
	return s.repo.Update(id, fields)
}
