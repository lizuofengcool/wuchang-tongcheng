// Package service 同城零工兼职业务逻辑层 - 岗位主表
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
// 依据需求文档 7.1：通用字段 status + audit_status
// MVP 简化：发布即通过审核（M 端可手动审核/下架）
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
	ErrLinggongNotFound       = errors.New("岗位不存在")
	ErrLinggongNoPermission   = errors.New("无权操作此岗位")
	ErrLinggongAudited        = errors.New("已审核的岗位不能重复审核")
	ErrLinggongStatusInvalid  = errors.New("岗位状态不允许此操作")
	ErrLinggongRecruitFull    = errors.New("招募人数已满")
)

// LinggongService 岗位业务接口
type LinggongService interface {
	// C 端
	Create(regionID uint, userID uint, userName string, userAvatar string, userPhone string, req *dto.CreateLinggongRequest) (*dto.LinggongInfo, error)
	Update(id uint, operatorID uint, req *dto.UpdateLinggongRequest) error
	Delete(id uint, operatorID uint) error
	GetByID(id uint) (*dto.LinggongInfo, error)
	List(regionID uint, req *dto.LinggongListRequest) (*utils.Pagination, []dto.LinggongInfo, error)
	ListMine(userID uint, page, pageSize int) (*utils.Pagination, []dto.LinggongInfo, error)
	ListByEmployer(employerID uint, page, pageSize int) (*utils.Pagination, []dto.LinggongInfo, error)
	Search(regionID uint, keyword string, page, pageSize int) (*utils.Pagination, []dto.LinggongInfo, error)
	Nearby(regionID uint, lat, lng, radiusKm float64, page, pageSize int) (*utils.Pagination, []dto.LinggongInfo, error)
	IncrViewCount(id uint) error
	IncrContactCount(id uint) error
	IncrShareCount(id uint) error

	// M 端管理
	AdminList(req *dto.LinggongAdminListRequest) (*utils.Pagination, []dto.LinggongInfo, error)
	AdminGetByID(id uint) (*dto.LinggongInfo, error)
	Audit(id uint, req *dto.LinggongAuditRequest) error
	UpdateStatus(id uint, status int) error
	BatchUpdateStatus(ids []uint, status int) error
}

type linggongService struct {
	repo repository.LinggongRepository
}

// NewLinggongService 创建岗位 service 实例
func NewLinggongService(repo repository.LinggongRepository) LinggongService {
	return &linggongService{repo: repo}
}

// linggongStatusText 状态文本
func linggongStatusText(status int) string {
	switch status {
	case model.LinggongStatusDraft:
		return "草稿"
	case model.LinggongStatusPublished:
		return "已发布"
	case model.LinggongStatusOffline:
		return "已下架"
	case model.LinggongStatusExpired:
		return "已过期"
	case model.LinggongStatusDeleted:
		return "已删除"
	case model.LinggongStatusFulfilled:
		return "已满员"
	case model.LinggongStatusClosed:
		return "已关闭"
	case model.LinggongStatusCompleted:
		return "已完成"
	}
	return ""
}

// linggongAuditStatusText 审核状态文本
func linggongAuditStatusText(s int) string {
	switch s {
	case model.LinggongAuditPending:
		return "待审"
	case model.LinggongAuditApproved:
		return "通过"
	case model.LinggongAuditRejected:
		return "拒绝"
	}
	return ""
}

// linggongTypeText 岗位类型文本
func linggongTypeText(t string) string {
	switch t {
	case model.LinggongTypeShortTerm:
		return "短期兼职"
	case model.LinggongTypeLongTerm:
		return "长期兼职"
	case model.LinggongTypeTask:
		return "任务制"
	case model.LinggongTypeHourly:
		return "小时工"
	case model.LinggongTypeDaily:
		return "日结工"
	case model.LinggongTypeTemp:
		return "临时工"
	}
	return ""
}

// publisherTypeText 发布者类型文本
func publisherTypeText(t string) string {
	switch t {
	case model.PublisherTypePersonal:
		return "个人雇主"
	case model.PublisherTypeCompany:
		return "企业雇主"
	case model.PublisherTypeAgent:
		return "中介"
	case model.PublisherTypeHeadhunter:
		return "猎头"
	}
	return ""
}

// billingTypeText 计费方式文本
func billingTypeText(t string) string {
	switch t {
	case model.BillingTypeByPiece:
		return "按件"
	case model.BillingTypeByHour:
		return "按时"
	case model.BillingTypeByDay:
		return "按日"
	case model.BillingTypeByWeek:
		return "按周"
	case model.BillingTypeByMonth:
		return "按月"
	case model.BillingTypeFixed:
		return "固定"
	case model.BillingTypeNegotiable:
		return "面议"
	}
	return ""
}

// settlementText 结算周期文本
func settlementText(s string) string {
	switch s {
	case model.SettlementT0:
		return "当日结"
	case model.SettlementT1:
		return "次日结"
	case model.SettlementT3:
		return "三日结"
	case model.SettlementT7:
		return "周结"
	case model.SettlementM1:
		return "月结"
	case model.SettlementProject:
		return "项目结"
	}
	return ""
}

// toLinggongInfo model -> dto
func toLinggongInfo(l *model.Linggong) *dto.LinggongInfo {
	info := &dto.LinggongInfo{
		ID:                 l.ID,
		Title:              l.Title,
		Content:            l.Content,
		CoverImage:         l.CoverImage,
		UserID:             l.UserID,
		UserName:           l.UserName,
		UserPhone:          l.UserPhone,
		UserAvatar:         l.UserAvatar,
		Status:             l.Status,
		StatusText:         linggongStatusText(l.Status),
		AuditStatus:        l.AuditStatus,
		AuditStatusText:    linggongAuditStatusText(l.AuditStatus),
		AuditReason:        l.AuditReason,
		PublishedAt:        l.PublishedAt,
		LinggongType:       l.LinggongType,
		LinggongTypeText:   linggongTypeText(l.LinggongType),
		PublisherType:      l.PublisherType,
		PublisherTypeText:  publisherTypeText(l.PublisherType),
		EmployerID:         l.EmployerID,
		CompanyName:        l.CompanyName,
		ContactName:        l.ContactName,
		ContactPhone:       l.ContactPhone,
		ContactWechat:      l.ContactWechat,
		BillingType:        l.BillingType,
		BillingTypeText:    billingTypeText(l.BillingType),
		SalaryMin:          l.SalaryMin,
		SalaryMax:          l.SalaryMax,
		SalaryUnit:         l.SalaryUnit,
		SalaryNegotiable:   l.SalaryNegotiable,
		Settlement:         l.Settlement,
		SettlementText:     settlementText(l.Settlement),
		Currency:           l.Currency,
		WorkStartDate:      l.WorkStartDate,
		WorkEndDate:        l.WorkEndDate,
		WorkDays:           l.WorkDays,
		WorkHours:          l.WorkHours,
		WorkTimeStart:      l.WorkTimeStart,
		WorkTimeEnd:        l.WorkTimeEnd,
		WorkWeekdays:        l.WorkWeekdays,
		WorkIntensity:      l.WorkIntensity,
		RecruitCount:       l.RecruitCount,
		AppliedCount:       l.AppliedCount,
		ConfirmedCount:     l.ConfirmedCount,
		NeedGender:         l.NeedGender,
		MinAge:             l.MinAge,
		MaxAge:             l.MaxAge,
		Education:          l.Education,
		Experience:         l.Experience,
		NeedHealthCert:     l.NeedHealthCert,
		NeedIDCard:         l.NeedIDCard,
		MinCreditScore:    l.MinCreditScore,
		Province:          l.Province,
		City:              l.City,
		District:          l.District,
		BusinessDistrict:   l.BusinessDistrict,
		Address:           l.Address,
		Latitude:          l.Latitude,
		Longitude:         l.Longitude,
		WorkLocationType:  l.WorkLocationType,
		TaskID:            l.TaskID,
		TotalTaskCount:    l.TotalTaskCount,
		ClaimedCount:      l.ClaimedCount,
		CompletedTaskCount: l.CompletedTaskCount,
		ViewCount:         l.ViewCount,
		FavCount:          l.FavCount,
		ContactCount:      l.ContactCount,
		ShareCount:        l.ShareCount,
		ApplicationCount:  l.ApplicationCount,
		LastAppliedAt:     l.LastAppliedAt,
		ContentHash:       l.ContentHash,
		RiskScore:         l.RiskScore,
		VideoURL:          l.VideoURL,
		VideoCover:        l.VideoCover,
		Features:          l.Features,
		Tags:              l.Tags,
		SkillTags:         l.SkillTags,
		WelfareTags:       l.WelfareTags,
		Images:            l.Images,
		Requirements:      l.Requirements,
		Featured:          l.Featured,
		Picked:            l.Picked,
		Verified:          l.Verified,
		PromotionLevel:    l.PromotionLevel,
		TrafficWeight:    l.TrafficWeight,
		EmployerVerified:   l.EmployerVerified,
		EmployerVerifiedAt: l.EmployerVerifiedAt,
		RegionID:          l.RegionID,
		CreatedAt:         l.CreatedAt,
		UpdatedAt:         l.UpdatedAt,
	}
	return info
}

// genLinggongNo 生成岗位编号：LG + yyyyMMddHHmmss + 6 位随机
func genLinggongNo() string {
	return fmt.Sprintf("LG%s%06d", time.Now().Format("20060102150405"), rand.Intn(1000000))
}

// ===== C 端 =====

// Create 创建岗位
func (s *linggongService) Create(regionID uint, userID uint, userName string, userAvatar string, userPhone string, req *dto.CreateLinggongRequest) (*dto.LinggongInfo, error) {
	l := &model.Linggong{
		Title:           req.Title,
		Content:         req.Content,
		CoverImage:      req.CoverImage,
		UserID:          userID,
		UserName:        userName,
		UserPhone:       userPhone,
		UserAvatar:      userAvatar,
		Status:          model.LinggongStatusDraft,
		AuditStatus:     model.LinggongAuditApproved, // MVP 简化：默认通过
		LinggongType:    req.LinggongType,
		PublisherType:   req.PublisherType,
		EmployerID:      req.EmployerID,
		CompanyName:     req.CompanyName,
		ContactName:     req.ContactName,
		ContactPhone:    req.ContactPhone,
		ContactWechat:   req.ContactWechat,
		BillingType:     req.BillingType,
		SalaryMin:       req.SalaryMin,
		SalaryMax:       req.SalaryMax,
		SalaryUnit:      req.SalaryUnit,
		SalaryNegotiable: req.SalaryNegotiable,
		Settlement:      req.Settlement,
		Currency:        model.CurrencyCNY,
		WorkStartDate:   req.WorkStartDate,
		WorkEndDate:     req.WorkEndDate,
		WorkDays:        req.WorkDays,
		WorkHours:       req.WorkHours,
		WorkTimeStart:   req.WorkTimeStart,
		WorkTimeEnd:     req.WorkTimeEnd,
		WorkWeekdays:     req.WorkWeekdays,
		WorkIntensity:   req.WorkIntensity,
		RecruitCount:    req.RecruitCount,
		NeedGender:      req.NeedGender,
		MinAge:          req.MinAge,
		MaxAge:          req.MaxAge,
		Education:       req.Education,
		Experience:      req.Experience,
		NeedHealthCert:  req.NeedHealthCert,
		NeedIDCard:      req.NeedIDCard,
		MinCreditScore: req.MinCreditScore,
		Province:        req.Province,
		City:            req.City,
		District:        req.District,
		BusinessDistrict: req.BusinessDistrict,
		Address:         req.Address,
		Latitude:        req.Latitude,
		Longitude:       req.Longitude,
		WorkLocationType: req.WorkLocationType,
		TaskID:          req.TaskID,
		TotalTaskCount:  req.TotalTaskCount,
		VideoURL:        req.VideoURL,
		VideoCover:      req.VideoCover,
	}
	l.RegionID = regionID

	// 默认值兜底
	if l.LinggongType == "" {
		l.LinggongType = model.LinggongTypeShortTerm
	}
	if l.PublisherType == "" {
		l.PublisherType = model.PublisherTypePersonal
	}
	if l.BillingType == "" {
		l.BillingType = model.BillingTypeByDay
	}
	if l.Settlement == "" {
		l.Settlement = model.SettlementT1
	}
	if l.WorkIntensity == "" {
		l.WorkIntensity = model.WorkIntensityMedium
	}
	if l.NeedGender == "" {
		l.NeedGender = "any"
	}
	if l.WorkLocationType == "" {
		l.WorkLocationType = "onsite"
	}

	// JSONB 字段转换
	if req.Features != nil {
		if jb, err := model.FromJSON(req.Features); err == nil {
			l.Features = jb
		}
	}
	if req.Tags != nil {
		if jb, err := model.FromJSON(req.Tags); err == nil {
			l.Tags = jb
		}
	}
	if req.SkillTags != nil {
		if jb, err := model.FromJSON(req.SkillTags); err == nil {
			l.SkillTags = jb
		}
	}
	if req.WelfareTags != nil {
		if jb, err := model.FromJSON(req.WelfareTags); err == nil {
			l.WelfareTags = jb
		}
	}
	if req.Images != nil {
		if jb, err := model.FromJSON(req.Images); err == nil {
			l.Images = jb
		}
	}
	if req.Requirements != nil {
		if jb, err := model.FromJSON(req.Requirements); err == nil {
			l.Requirements = jb
		}
	}

	if err := s.repo.Create(l); err != nil {
		return nil, err
	}
	return toLinggongInfo(l), nil
}

// Update 更新岗位（仅发布者本人）
func (s *linggongService) Update(id uint, operatorID uint, req *dto.UpdateLinggongRequest) error {
	l, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrLinggongNotFound
		}
		return err
	}
	if l.UserID != operatorID {
		return ErrLinggongNoPermission
	}

	fields := map[string]interface{}{}
	if req.Title != nil {
		fields["title"] = *req.Title
	}
	if req.Content != nil {
		fields["content"] = *req.Content
	}
	if req.CoverImage != nil {
		fields["cover_image"] = *req.CoverImage
	}
	if req.CompanyName != nil {
		fields["company_name"] = *req.CompanyName
	}
	if req.ContactName != nil {
		fields["contact_name"] = *req.ContactName
	}
	if req.ContactPhone != nil {
		fields["contact_phone"] = *req.ContactPhone
	}
	if req.ContactWechat != nil {
		fields["contact_wechat"] = *req.ContactWechat
	}
	if req.BillingType != nil {
		fields["billing_type"] = *req.BillingType
	}
	if req.SalaryMin != nil {
		fields["salary_min"] = *req.SalaryMin
	}
	if req.SalaryMax != nil {
		fields["salary_max"] = *req.SalaryMax
	}
	if req.SalaryUnit != nil {
		fields["salary_unit"] = *req.SalaryUnit
	}
	if req.SalaryNegotiable != nil {
		fields["salary_negotiable"] = *req.SalaryNegotiable
	}
	if req.Settlement != nil {
		fields["settlement"] = *req.Settlement
	}
	if req.WorkStartDate != nil {
		fields["work_start_date"] = *req.WorkStartDate
	}
	if req.WorkEndDate != nil {
		fields["work_end_date"] = *req.WorkEndDate
	}
	if req.RecruitCount != nil {
		fields["recruit_count"] = *req.RecruitCount
	}
	if req.Province != nil {
		fields["province"] = *req.Province
	}
	if req.City != nil {
		fields["city"] = *req.City
	}
	if req.District != nil {
		fields["district"] = *req.District
	}
	if req.Address != nil {
		fields["address"] = *req.Address
	}
	if req.VideoURL != nil {
		fields["video_url"] = *req.VideoURL
	}

	// 状态变更
	if req.Status != nil {
		now := time.Now()
		switch *req.Status {
		case model.LinggongStatusPublished:
			if l.Status != model.LinggongStatusPublished {
				fields["status"] = model.LinggongStatusPublished
				fields["published_at"] = &now
				fields["audit_status"] = model.LinggongAuditApproved
			}
		case model.LinggongStatusOffline:
			fields["status"] = model.LinggongStatusOffline
		case model.LinggongStatusClosed:
			fields["status"] = model.LinggongStatusClosed
		case model.LinggongStatusCompleted:
			fields["status"] = model.LinggongStatusCompleted
		default:
			fields["status"] = *req.Status
		}
	}

	// JSONB 字段更新
	if req.Features != nil {
		if jb, err := model.FromJSON(req.Features); err == nil {
			fields["features"] = jb
		}
	}
	if req.Tags != nil {
		if jb, err := model.FromJSON(req.Tags); err == nil {
			fields["tags"] = jb
		}
	}

	if len(fields) == 0 {
		return nil
	}
	return s.repo.Update(id, fields)
}

// Delete 删除岗位（仅发布者本人）
func (s *linggongService) Delete(id uint, operatorID uint) error {
	l, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrLinggongNotFound
		}
		return err
	}
	if l.UserID != operatorID {
		return ErrLinggongNoPermission
	}
	return s.repo.Delete(id)
}

// GetByID 获取详情（同时增加浏览量）
func (s *linggongService) GetByID(id uint) (*dto.LinggongInfo, error) {
	l, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrLinggongNotFound
		}
		return nil, err
	}
	_ = s.repo.IncrViewCount(id)
	l.ViewCount++
	return toLinggongInfo(l), nil
}

// List C 端列表查询（地区隔离）
func (s *linggongService) List(regionID uint, req *dto.LinggongListRequest) (*utils.Pagination, []dto.LinggongInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.LinggongListOptions{
		LinggongType:     req.LinggongType,
		PublisherType:    req.PublisherType,
		BillingType:      req.BillingType,
		Settlement:       req.Settlement,
		MinSalary:        req.MinSalary,
		MaxSalary:        req.MaxSalary,
		Status:           req.Status,
		AuditStatus:      req.AuditStatus,
		EmployerID:       req.EmployerID,
		Province:         req.Province,
		City:             req.City,
		District:         req.District,
		WorkLocationType: req.WorkLocationType,
		Featured:         req.Featured,
		Picked:           req.Picked,
		Verified:         req.Verified,
		EmployerVerified: req.EmployerVerified,
		NeedGender:       req.NeedGender,
		Education:        req.Education,
		Keyword:          req.Keyword,
		Sort:             req.Sort,
	}

	// C 端默认仅展示已发布+审核通过
	if opts.Status == nil {
		published := model.LinggongStatusPublished
		opts.Status = &published
	}
	if opts.AuditStatus == nil {
		approved := model.LinggongAuditApproved
		opts.AuditStatus = &approved
	}

	list, total, err := s.repo.List(regionID, pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.LinggongInfo, 0, len(list))
	for i := range list {
		result = append(result, *toLinggongInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListMine 我的发布
func (s *linggongService) ListMine(userID uint, page, pageSize int) (*utils.Pagination, []dto.LinggongInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByUser(userID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.LinggongInfo, 0, len(list))
	for i := range list {
		result = append(result, *toLinggongInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListByEmployer 按雇主反查
func (s *linggongService) ListByEmployer(employerID uint, page, pageSize int) (*utils.Pagination, []dto.LinggongInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByEmployer(employerID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.LinggongInfo, 0, len(list))
	for i := range list {
		result = append(result, *toLinggongInfo(&list[i]))
	}
	return pagination, result, nil
}

// Search 关键词搜索
func (s *linggongService) Search(regionID uint, keyword string, page, pageSize int) (*utils.Pagination, []dto.LinggongInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.Search(regionID, keyword, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.LinggongInfo, 0, len(list))
	for i := range list {
		result = append(result, *toLinggongInfo(&list[i]))
	}
	return pagination, result, nil
}

// Nearby 附近岗位
func (s *linggongService) Nearby(regionID uint, lat, lng, radiusKm float64, page, pageSize int) (*utils.Pagination, []dto.LinggongInfo, error) {
	if radiusKm <= 0 {
		radiusKm = 5 // 默认 5 公里
	}
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.Nearby(regionID, lat, lng, radiusKm, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.LinggongInfo, 0, len(list))
	for i := range list {
		result = append(result, *toLinggongInfo(&list[i]))
	}
	return pagination, result, nil
}

// IncrViewCount 增加浏览量
func (s *linggongService) IncrViewCount(id uint) error {
	return s.repo.IncrViewCount(id)
}

// IncrContactCount 增加联系数
func (s *linggongService) IncrContactCount(id uint) error {
	return s.repo.IncrContactCount(id)
}

// IncrShareCount 增加分享数
func (s *linggongService) IncrShareCount(id uint) error {
	return s.repo.IncrShareCount(id)
}

// ===== M 端管理 =====

func (s *linggongService) AdminList(req *dto.LinggongAdminListRequest) (*utils.Pagination, []dto.LinggongInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.LinggongAdminListOptions{
		RegionID:      req.RegionID,
		UserID:        req.UserID,
		EmployerID:    req.EmployerID,
		LinggongType:  req.LinggongType,
		PublisherType: req.PublisherType,
		BillingType:   req.BillingType,
		Status:        req.Status,
		AuditStatus:   req.AuditStatus,
		Keyword:       req.Keyword,
	}
	list, total, err := s.repo.AdminList(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.LinggongInfo, 0, len(list))
	for i := range list {
		result = append(result, *toLinggongInfo(&list[i]))
	}
	return pagination, result, nil
}

func (s *linggongService) AdminGetByID(id uint) (*dto.LinggongInfo, error) {
	l, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrLinggongNotFound
		}
		return nil, err
	}
	return toLinggongInfo(l), nil
}

// Audit M 端审核
func (s *linggongService) Audit(id uint, req *dto.LinggongAuditRequest) error {
	l, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrLinggongNotFound
		}
		return err
	}

	fields := map[string]interface{}{
		"audit_status": req.AuditStatus,
		"audit_reason": req.AuditReason,
	}

	now := time.Now()
	// 审核通过且当前为草稿：自动发布
	if req.AuditStatus == model.LinggongAuditApproved && l.Status == model.LinggongStatusDraft {
		fields["status"] = model.LinggongStatusPublished
		fields["published_at"] = &now
	}
	// 审核拒绝：强制下架
	if req.AuditStatus == model.LinggongAuditRejected && l.Status == model.LinggongStatusPublished {
		fields["status"] = model.LinggongStatusOffline
	}

	return s.repo.Update(id, fields)
}

// UpdateStatus M 端强制下架/恢复
func (s *linggongService) UpdateStatus(id uint, status int) error {
	now := time.Now()
	fields := map[string]interface{}{
		"status": status,
	}
	switch status {
	case model.LinggongStatusPublished:
		fields["published_at"] = &now
		fields["audit_status"] = model.LinggongAuditApproved
	}
	return s.repo.Update(id, fields)
}

// BatchUpdateStatus 批量更新状态
func (s *linggongService) BatchUpdateStatus(ids []uint, status int) error {
	if len(ids) == 0 {
		return nil
	}
	return s.repo.BatchUpdateStatus(ids, status)
}
