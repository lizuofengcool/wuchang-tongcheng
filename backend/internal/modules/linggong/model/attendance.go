// Package model 考勤打卡表（对标钉钉/企业微信考勤）
// GPS 定位 + WiFi + 人脸 + 工时统计
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 考勤类型常量 ===
const (
	AttendanceTypeClockIn  = "clock_in"  // 上班打卡
	AttendanceTypeClockOut = "clock_out" // 下班打卡
	AttendanceTypeBreak    = "break"     // 休息开始
	AttendanceTypeResume   = "resume"    // 休息结束
	AttendanceTypeOvertime = "overtime" // 加班
)

// === 考勤状态常量 ===
const (
	AttendanceStatusNormal    = 0 // 正常
	AttendanceStatusLate      = 1 // 迟到
	AttendanceStatusEarlyLeave = 2 // 早退
	AttendanceStatusAbsent    = 3 // 缺勤
	AttendanceStatusLeave     = 4 // 请假
	AttendanceStatusBusinessTrip = 5 // 出差
	AttendanceStatusRemote   = 6 // 远程
	AttendanceStatusOut      = 7 // 外勤
)

// === 打卡方式常量 ===
const (
	ClockMethodGPS    = "gps"     // GPS 定位
	ClockMethodWiFi   = "wifi"    // WiFi
	ClockMethodFace   = "face"    // 人脸识别
	ClockMethodManual = "manual" // 人工补卡
	ClockMethodQRCode = "qr_code" // 二维码
)

// LinggongAttendance 考勤打卡表
type LinggongAttendance struct {
	database.RegionBaseModel // id/region_id/created_at/updated_at/deleted_at
	AttendanceNo  string     `gorm:"size:64;not null;uniqueIndex" json:"attendance_no"`         // 考勤单号
	LinggongID   uint       `gorm:"not null;index" json:"linggong_id"`                          // 岗位 ID
	TaskID       uint       `gorm:"not null;default:0;index" json:"task_id"`                   // 任务包 ID
	ApplicationID uint      `gorm:"not null;default:0;index" json:"application_id"`             // 报名记录 ID
	ContractID   uint       `gorm:"not null;default:0;index" json:"contract_id"`               // 合同 ID
	EmployerID   uint       `gorm:"not null;index" json:"employer_id"`                          // 雇主 ID
	WorkerID     uint       `gorm:"not null;index" json:"worker_id"`                            // 求职者 ID
	WorkerName   string     `gorm:"size:50;not null;default:''" json:"worker_name"`              // 求职者姓名
	AttendanceType string    `gorm:"size:16;not null;default:'clock_in';index" json:"attendance_type"` // clock_in/clock_out/break/resume/overtime
	ClockMethod  string     `gorm:"size:16;not null;default:'gps'" json:"clock_method"`          // gps/wifi/face/manual/qr_code
	ClockTime    time.Time  `gorm:"not null;index" json:"clock_time"`                            // 打卡时间
	ClockDate    *time.Time `gorm:"type:date;index" json:"clock_date"`                            // 打卡日期
	Latitude     float64    `gorm:"type:decimal(10,7);default:0" json:"latitude"`                 // 纬度
	Longitude    float64    `gorm:"type:decimal(10,7);default:0" json:"longitude"`               // 经度
	Address      string     `gorm:"size:500;not null;default:''" json:"address"`                // 打卡地址
	WifiName    string     `gorm:"size:128;not null;default:''" json:"wifi_name"`               // WiFi 名称
	WifiMAC      string     `gorm:"size:64;not null;default:''" json:"wifi_mac"`                 // WiFi MAC
	FaceImageURL string    `gorm:"size:255;not null;default:''" json:"face_image_url"`           // 人脸照片
	QRCodeContent string    `gorm:"size:255;not null;default:''" json:"qr_code_content"`         // 二维码内容
	Status       int        `gorm:"default:0;index" json:"status"`                              // 0正常 1迟到 2早退 3缺勤 4请假 5出差 6远程 7外勤
	LateMinutes  int        `gorm:"not null;default:0" json:"late_minutes"`                       // 迟到分钟数
	EarlyMinutes int        `gorm:"not null;default:0" json:"early_minutes"`                     // 早退分钟数
	WorkHours    float64    `gorm:"type:decimal(8,2);default:0" json:"work_hours"`               // 工作时长
	OvertimeHours float64   `gorm:"type:decimal(8,2);default:0" json:"overtime_hours"`           // 加班时长
	BreakHours   float64    `gorm:"type:decimal(8,2);default:0" json:"break_hours"`              // 休息时长
	TaskCount    int        `gorm:"not null;default:0" json:"task_count"`                        // 完成任务数
	Remark       string     `gorm:"type:text" json:"remark"`                                     // 备注
	Approved     bool       `gorm:"not null;default:false;index" json:"approved"`                // 是否已审核
	ApprovedBy   uint       `gorm:"not null;default:0" json:"approved_by"`                        // 审核人 ID
	ApprovedAt   *time.Time `gorm:"index" json:"approved_at"`                                    // 审核时间
}

// TableName 表名（linggong_ 前缀）
func (LinggongAttendance) TableName() string { return "linggong_attendances" }
