// Package service 同城零工兼职业务逻辑层 - 双向评价
// 对标美团/斗米：工人评价雇主 + 雇主评价工人 + 评价回复/追评 + 点赞 + 多维度评分
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
	ErrRatingNotFound      = errors.New("评价不存在")
	ErrRatingNoPermission  = errors.New("无权操作此评价")
	ErrRatingStatusInvalid = errors.New("评价状态不允许此操作")
)

// RatingService 双向评价业务接口
type RatingService interface {
	// C 端
	Create(regionID uint, userID uint, req *dto.CreateRatingRequest) (*dto.RatingInfo, error)
	Update(id uint, operatorID uint, req *dto.UpdateRatingRequest) error
	Delete(id uint, operatorID uint) error
	GetByID(id uint) (*dto.RatingInfo, error)
	GetByRatingNo(no string) (*dto.RatingInfo, error)
	List(regionID uint, req *dto.RatingListRequest) (*utils.Pagination, []dto.RatingInfo, error)
	ListByLinggong(linggongID uint, page, pageSize int) (*utils.Pagination, []dto.RatingInfo, error)
	ListByRater(raterID uint, page, pageSize int) (*utils.Pagination, []dto.RatingInfo, error)
	ListByTarget(targetType string, targetID uint, page, pageSize int) (*utils.Pagination, []dto.RatingInfo, error)

	// 回复/追评/点赞
	Reply(id uint, req *dto.RatingReplyRequest) error
	Append(id uint, req *dto.RatingAppendRequest) error
	Like(id uint) error

	// 评价统计
	GetStats(targetType string, targetID uint) (*repository.RatingStatsResult, error)

	// M 端管理
	AdminList(req *dto.RatingAdminListRequest) (*utils.Pagination, []dto.RatingInfo, error)
	Audit(id uint, req *dto.RatingAuditRequest) error
}

type ratingService struct {
	repo repository.RatingRepository
}

// NewRatingService 创建双向评价 service 实例
func NewRatingService(repo repository.RatingRepository) RatingService {
	return &ratingService{repo: repo}
}

// ratingStatusText 评价状态文本
func ratingStatusText(s int) string {
	switch s {
	case model.RatingStatusPending:
		return "待审核"
	case model.RatingStatusApproved:
		return "已通过"
	case model.RatingStatusRejected:
		return "已拒绝"
	case model.RatingStatusHidden:
		return "已隐藏"
	}
	return ""
}

// raterTypeText 评价者类型文本
func raterTypeText(t string) string {
	switch t {
	case model.RaterTypeEmployer:
		return "雇主"
	case model.RaterTypeWorker:
		return "求职者"
	}
	return ""
}

// ratingTargetTypeText 评价目标类型文本
func ratingTargetTypeText(t string) string {
	switch t {
	case model.RatingTargetTypeEmployer:
		return "雇主"
	case model.RatingTargetTypeWorker:
		return "求职者"
	case model.RatingTargetTypeLinggong:
		return "岗位"
	case model.RatingTargetTypeTask:
		return "任务"
	}
	return ""
}

// recommendText 推荐文本
func recommendText(r string) string {
	switch r {
	case model.RecommendYes:
		return "推荐"
	case model.RecommendNo:
		return "不推荐"
	case model.RecommendMaybe:
		return "一般"
	}
	return ""
}

// toRatingInfo model -> dto
func toRatingInfo(r *model.LinggongRating) *dto.RatingInfo {
	return &dto.RatingInfo{
		ID:                r.ID,
		RatingNo:          r.RatingNo,
		LinggongID:        r.LinggongID,
		TaskID:            r.TaskID,
		ApplicationID:     r.ApplicationID,
		ContractID:        r.ContractID,
		PaymentID:         r.PaymentID,
		RaterType:         r.RaterType,
		RaterTypeText:     raterTypeText(r.RaterType),
		RaterID:           r.RaterID,
		RaterName:         r.RaterName,
		RaterAvatar:       r.RaterAvatar,
		TargetType:        r.TargetType,
		TargetTypeText:    ratingTargetTypeText(r.TargetType),
		TargetID:          r.TargetID,
		TargetName:        r.TargetName,
		Rating:            r.Rating,
		Content:           r.Content,
		Images:            r.Images,
		VideoURL:          r.VideoURL,
		IsAnonymous:       r.IsAnonymous,
		IsRecommended:     r.IsRecommended,
		IsRecommendedText: recommendText(r.IsRecommended),
		Tags:              r.Tags,
		DealAmount:        r.DealAmount,
		WorkQuality:       r.WorkQuality,
		Punctuality:       r.Punctuality,
		Communication:     r.Communication,
		Attitude:          r.Attitude,
		Professionalism:   r.Professionalism,
		PaymentTimeliness: r.PaymentTimeliness,
		WorkEnvironment:   r.WorkEnvironment,
		SalaryMatch:       r.SalaryMatch,
		Reply:             r.Reply,
		ReplyAt:           r.ReplyAt,
		AppendContent:     r.AppendContent,
		AppendImages:      r.AppendImages,
		AppendAt:          r.AppendAt,
		LikeCount:         r.LikeCount,
		Status:            r.Status,
		StatusText:        ratingStatusText(r.Status),
		RejectedReason:    r.RejectedReason,
		EvaluatedAt:       r.EvaluatedAt,
		RegionID:          r.RegionID,
		CreatedAt:         r.CreatedAt,
		UpdatedAt:         r.UpdatedAt,
	}
}

// genRatingNo 生成评价单号：LR + yyyyMMddHHmmss + 6 位随机
func genRatingNo() string {
	return fmt.Sprintf("LR%s%06d", time.Now().Format("20060102150405"), rand.Intn(1000000))
}

// ===== C 端 =====

// Create 创建评价
func (s *ratingService) Create(regionID uint, userID uint, req *dto.CreateRatingRequest) (*dto.RatingInfo, error) {
	now := time.Now()
	r := &model.LinggongRating{
		RatingNo:         genRatingNo(),
		LinggongID:       req.LinggongID,
		TaskID:           req.TaskID,
		ApplicationID:    req.ApplicationID,
		ContractID:       req.ContractID,
		PaymentID:        req.PaymentID,
		RaterType:        req.RaterType,
		RaterID:          userID,
		TargetType:       req.TargetType,
		TargetID:         req.TargetID,
		TargetName:       req.TargetName,
		Rating:           req.Rating,
		Content:          req.Content,
		VideoURL:         req.VideoURL,
		IsAnonymous:      req.IsAnonymous,
		IsRecommended:    req.IsRecommended,
		DealAmount:       req.DealAmount,
		WorkQuality:      req.WorkQuality,
		Punctuality:      req.Punctuality,
		Communication:    req.Communication,
		Attitude:         req.Attitude,
		Professionalism:  req.Professionalism,
		PaymentTimeliness: req.PaymentTimeliness,
		WorkEnvironment:  req.WorkEnvironment,
		SalaryMatch:      req.SalaryMatch,
		Status:           model.RatingStatusApproved, // MVP：发布即通过
		EvaluatedAt:      &now,
	}
	r.RegionID = regionID

	// 默认值兜底
	if r.RaterType == "" {
		r.RaterType = model.RaterTypeWorker
	}
	if r.TargetType == "" {
		r.TargetType = model.RatingTargetTypeWorker
	}
	if r.IsRecommended == "" {
		r.IsRecommended = model.RecommendYes
	}
	if r.Rating == 0 {
		r.Rating = 5
	}

	// JSONB 字段
	if req.Images != nil {
		if jb, err := model.FromJSON(req.Images); err == nil {
			r.Images = jb
		}
	}
	if req.Tags != nil {
		if jb, err := model.FromJSON(req.Tags); err == nil {
			r.Tags = jb
		}
	}

	if err := s.repo.Create(r); err != nil {
		return nil, err
	}
	return toRatingInfo(r), nil
}

// Update 更新评价（仅本人，且状态为待审核/已通过）
func (s *ratingService) Update(id uint, operatorID uint, req *dto.UpdateRatingRequest) error {
	r, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRatingNotFound
		}
		return err
	}
	if r.RaterID != operatorID {
		return ErrRatingNoPermission
	}
	if r.Status == model.RatingStatusRejected || r.Status == model.RatingStatusHidden {
		return ErrRatingStatusInvalid
	}

	fields := map[string]interface{}{}
	if req.Rating != nil {
		fields["rating"] = *req.Rating
	}
	if req.Content != nil {
		fields["content"] = *req.Content
	}
	if req.VideoURL != nil {
		fields["video_url"] = *req.VideoURL
	}
	if req.IsAnonymous != nil {
		fields["is_anonymous"] = *req.IsAnonymous
	}
	if req.IsRecommended != nil {
		fields["is_recommended"] = *req.IsRecommended
	}
	if req.DealAmount != nil {
		fields["deal_amount"] = *req.DealAmount
	}
	if req.WorkQuality != nil {
		fields["work_quality"] = *req.WorkQuality
	}
	if req.Punctuality != nil {
		fields["punctuality"] = *req.Punctuality
	}
	if req.Communication != nil {
		fields["communication"] = *req.Communication
	}
	if req.Attitude != nil {
		fields["attitude"] = *req.Attitude
	}
	if req.Professionalism != nil {
		fields["professionalism"] = *req.Professionalism
	}
	if req.PaymentTimeliness != nil {
		fields["payment_timeliness"] = *req.PaymentTimeliness
	}
	if req.WorkEnvironment != nil {
		fields["work_environment"] = *req.WorkEnvironment
	}
	if req.SalaryMatch != nil {
		fields["salary_match"] = *req.SalaryMatch
	}
	if req.Images != nil {
		if jb, err := model.FromJSON(req.Images); err == nil {
			fields["images"] = jb
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

// Delete 删除评价（仅本人）
func (s *ratingService) Delete(id uint, operatorID uint) error {
	r, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRatingNotFound
		}
		return err
	}
	if r.RaterID != operatorID {
		return ErrRatingNoPermission
	}
	return s.repo.Delete(id)
}

// GetByID 获取评价详情
func (s *ratingService) GetByID(id uint) (*dto.RatingInfo, error) {
	r, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRatingNotFound
		}
		return nil, err
	}
	return toRatingInfo(r), nil
}

// GetByRatingNo 按评价单号查询
func (s *ratingService) GetByRatingNo(no string) (*dto.RatingInfo, error) {
	r, err := s.repo.FindByRatingNo(no)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRatingNotFound
		}
		return nil, err
	}
	return toRatingInfo(r), nil
}

// List C 端评价列表（默认仅展示已通过）
func (s *ratingService) List(regionID uint, req *dto.RatingListRequest) (*utils.Pagination, []dto.RatingInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.RatingListOptions{
		LinggongID:    req.LinggongID,
		TaskID:        req.TaskID,
		ApplicationID: req.ApplicationID,
		RaterType:     req.RaterType,
		TargetType:    req.TargetType,
		TargetID:      req.TargetID,
		Rating:        req.Rating,
		IsRecommended: req.IsRecommended,
		Status:        req.Status,
		Keyword:       req.Keyword,
	}
	// C 端默认仅展示已通过
	if opts.Status == nil {
		approved := model.RatingStatusApproved
		opts.Status = &approved
	}
	list, total, err := s.repo.List(regionID, pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.RatingInfo, 0, len(list))
	for i := range list {
		result = append(result, *toRatingInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListByLinggong 按岗位反查
func (s *ratingService) ListByLinggong(linggongID uint, page, pageSize int) (*utils.Pagination, []dto.RatingInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByLinggong(linggongID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.RatingInfo, 0, len(list))
	for i := range list {
		result = append(result, *toRatingInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListByRater 按评价人反查
func (s *ratingService) ListByRater(raterID uint, page, pageSize int) (*utils.Pagination, []dto.RatingInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByRater(raterID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.RatingInfo, 0, len(list))
	for i := range list {
		result = append(result, *toRatingInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListByTarget 按目标反查
func (s *ratingService) ListByTarget(targetType string, targetID uint, page, pageSize int) (*utils.Pagination, []dto.RatingInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByTarget(targetType, targetID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.RatingInfo, 0, len(list))
	for i := range list {
		result = append(result, *toRatingInfo(&list[i]))
	}
	return pagination, result, nil
}

// Reply 回复评价
func (s *ratingService) Reply(id uint, req *dto.RatingReplyRequest) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRatingNotFound
		}
		return err
	}
	now := time.Now()
	fields := map[string]interface{}{
		"reply":    req.Reply,
		"reply_at": &now,
	}
	return s.repo.Update(id, fields)
}

// Append 追评
func (s *ratingService) Append(id uint, req *dto.RatingAppendRequest) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRatingNotFound
		}
		return err
	}
	now := time.Now()
	fields := map[string]interface{}{
		"append_content": req.AppendContent,
		"append_at":      &now,
	}
	if req.AppendImages != nil {
		if jb, err := model.FromJSON(req.AppendImages); err == nil {
			fields["append_images"] = jb
		}
	}
	return s.repo.Update(id, fields)
}

// Like 点赞
func (s *ratingService) Like(id uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRatingNotFound
		}
		return err
	}
	return s.repo.IncrLikeCount(id)
}

// GetStats 获取评价统计
func (s *ratingService) GetStats(targetType string, targetID uint) (*repository.RatingStatsResult, error) {
	return s.repo.GetRatingStats(targetType, targetID)
}

// ===== M 端管理 =====

// AdminList M 端评价列表（跨地区）
func (s *ratingService) AdminList(req *dto.RatingAdminListRequest) (*utils.Pagination, []dto.RatingInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.RatingAdminListOptions{
		RegionID:   req.RegionID,
		LinggongID: req.LinggongID,
		RaterID:    req.RaterID,
		TargetType: req.TargetType,
		TargetID:   req.TargetID,
		Rating:     req.Rating,
		Status:     req.Status,
		Keyword:    req.Keyword,
	}
	list, total, err := s.repo.AdminList(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.RatingInfo, 0, len(list))
	for i := range list {
		result = append(result, *toRatingInfo(&list[i]))
	}
	return pagination, result, nil
}

// Audit 评价审核
func (s *ratingService) Audit(id uint, req *dto.RatingAuditRequest) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRatingNotFound
		}
		return err
	}
	fields := map[string]interface{}{
		"status":          req.Status,
		"rejected_reason": req.RejectedReason,
	}
	return s.repo.Update(id, fields)
}
