// Package service love 相亲交友业务逻辑层 - 动态广场
package service

import (
	"errors"
	"time"

	"wuchang-tongcheng/internal/modules/love/dto"
	"wuchang-tongcheng/internal/modules/love/model"
	"wuchang-tongcheng/internal/modules/love/repository"
	"wuchang-tongcheng/internal/pkg/utils"
)

var (
	ErrLoveStoryNotFound      = errors.New("动态不存在")
	ErrLoveStoryNoPermission  = errors.New("无权操作此动态")
	ErrLoveStoryStatusInvalid = errors.New("动态状态不允许此操作")
)

// LoveStoryService 动态业务接口
type LoveStoryService interface {
	Create(loveID, userID uint, req *dto.CreateLoveStoryRequest) (*dto.LoveStoryInfo, error)
	Update(id uint, operatorID uint, req *dto.UpdateLoveStoryRequest) error
	Delete(id uint, operatorID uint) error
	GetByID(id uint, viewerUserID uint) (*dto.LoveStoryInfo, error)
	List(req *dto.LoveStoryListRequest) (*utils.Pagination, []dto.LoveStoryInfo, error)
	ListByLoveID(loveID uint, page, pageSize int) (*utils.Pagination, []dto.LoveStoryInfo, error)
	ListByUser(userID uint, page, pageSize int) (*utils.Pagination, []dto.LoveStoryInfo, error)
	ListFeatured(page, pageSize int) (*utils.Pagination, []dto.LoveStoryInfo, error)
	ListByTopic(topic string, page, pageSize int) (*utils.Pagination, []dto.LoveStoryInfo, error)

	IncrView(id uint) error
	IncrLike(id uint) error
	DecrLike(id uint) error
	IncrComment(id uint) error
	IncrShare(id uint) error

	AdminList(req *dto.LoveStoryListRequest) (*utils.Pagination, []dto.LoveStoryInfo, error)
	Audit(id uint, auditStatus int, auditReason string) error
	UpdateStatus(id uint, status int) error
	SetFeatured(id uint, featured bool) error
	BatchAudit(ids []uint, auditStatus int, auditReason string) error
}

type loveStoryService struct {
	repo repository.LoveStoryRepository
}

// NewLoveStoryService 创建动态 service
func NewLoveStoryService(repo repository.LoveStoryRepository) LoveStoryService {
	return &loveStoryService{repo: repo}
}

func storyStatusText(s int) string {
	switch s {
	case 0:
		return "删除"
	case 1:
		return "正常"
	case 2:
		return "下架"
	}
	return ""
}

func storyAuditStatusText(s int) string {
	return loveAuditStatusText(s)
}

func toLoveStoryInfo(s *model.LoveStory) dto.LoveStoryInfo {
	return dto.LoveStoryInfo{
		ID:              s.ID,
		StoryNo:         s.StoryNo,
		LoveID:          s.LoveID,
		UserID:          s.UserID,
		UserNickname:    s.UserNickname,
		UserAvatar:      s.UserAvatar,
		UserGender:      s.UserGender,
		Title:           s.Title,
		Content:         s.Content,
		MediaType:       s.MediaType,
		ImageUrls:       s.ImageUrls,
		VideoURL:        s.VideoURL,
		VideoCover:      s.VideoCover,
		VideoDuration:   s.VideoDuration,
		VoiceURL:        s.VoiceURL,
		VoiceDuration:   s.VoiceDuration,
		Location:        s.Location,
		Latitude:        s.Latitude,
		Longitude:       s.Longitude,
		Tags:            s.Tags,
		Topic:           s.Topic,
		ViewCount:       s.ViewCount,
		LikeCount:       s.LikeCount,
		CommentCount:    s.CommentCount,
		ShareCount:      s.ShareCount,
		ReportCount:     s.ReportCount,
		Featured:        s.Featured,
		Status:          s.Status,
		StatusText:      storyStatusText(s.Status),
		AuditStatus:     s.AuditStatus,
		AuditStatusText: storyAuditStatusText(s.AuditStatus),
		AuditReason:     s.AuditReason,
		PublishedAt:     s.PublishedAt,
		HotScore:        s.HotScore,
		CreatedAt:       s.CreatedAt,
		UpdatedAt:       s.UpdatedAt,
	}
}

func (s *loveStoryService) Create(loveID, userID uint, req *dto.CreateLoveStoryRequest) (*dto.LoveStoryInfo, error) {
	mediaType := req.MediaType
	if mediaType == "" {
		mediaType = model.StoryMediaTypeImage
	}
	story := &model.LoveStory{
		StoryNo:      generateMatchNo("ST", loveID),
		LoveID:       loveID,
		UserID:       userID,
		Title:        req.Title,
		Content:      req.Content,
		MediaType:    mediaType,
		VideoURL:     req.VideoURL,
		VideoCover:   req.VideoCover,
		VideoDuration: req.VideoDuration,
		VoiceURL:     req.VoiceURL,
		VoiceDuration: req.VoiceDuration,
		Location:     req.Location,
		Latitude:     req.Latitude,
		Longitude:    req.Longitude,
		Topic:        req.Topic,
		Status:       1,
		AuditStatus:  model.LoveAuditApproved,
		PublishedAt:  nil,
	}
	publishedAt := time.Now()
	story.PublishedAt = &publishedAt
	if err := s.repo.Create(story); err != nil {
		return nil, err
	}
	info := toLoveStoryInfo(story)
	return &info, nil
}

func (s *loveStoryService) Update(id uint, operatorID uint, req *dto.UpdateLoveStoryRequest) error {
	story, err := s.repo.FindByID(id)
	if err != nil {
		return ErrLoveStoryNotFound
	}
	if story.UserID != operatorID {
		return ErrLoveStoryNoPermission
	}
	if req.Title != nil {
		story.Title = *req.Title
	}
	if req.Content != nil {
		story.Content = *req.Content
	}
	if req.VideoURL != nil {
		story.VideoURL = *req.VideoURL
	}
	if req.VideoCover != nil {
		story.VideoCover = *req.VideoCover
	}
	if req.VideoDuration != nil {
		story.VideoDuration = *req.VideoDuration
	}
	if req.VoiceURL != nil {
		story.VoiceURL = *req.VoiceURL
	}
	if req.VoiceDuration != nil {
		story.VoiceDuration = *req.VoiceDuration
	}
	if req.Location != nil {
		story.Location = *req.Location
	}
	if req.Latitude != nil {
		story.Latitude = *req.Latitude
	}
	if req.Longitude != nil {
		story.Longitude = *req.Longitude
	}
	if req.Topic != nil {
		story.Topic = *req.Topic
	}
	return s.repo.Update(story)
}

func (s *loveStoryService) Delete(id uint, operatorID uint) error {
	story, err := s.repo.FindByID(id)
	if err != nil {
		return ErrLoveStoryNotFound
	}
	if story.UserID != operatorID {
		return ErrLoveStoryNoPermission
	}
	return s.repo.Delete(id)
}

func (s *loveStoryService) GetByID(id uint, viewerUserID uint) (*dto.LoveStoryInfo, error) {
	story, err := s.repo.FindByID(id)
	if err != nil {
		return nil, ErrLoveStoryNotFound
	}
	_ = s.repo.IncrViewCount(id)
	info := toLoveStoryInfo(story)
	return &info, nil
}

func (s *loveStoryService) List(req *dto.LoveStoryListRequest) (*utils.Pagination, []dto.LoveStoryInfo, error) {
	opts := repository.LoveStoryListOptions{
		LoveID:      req.LoveID,
		UserID:      req.UserID,
		MediaType:   req.MediaType,
		Topic:       req.Topic,
		Featured:    req.Featured,
		Keyword:     req.Keyword,
		Sort:        req.Sort,
	}
	activeStatus := 1
	opts.Status = &activeStatus
	approvedStatus := model.LoveAuditApproved
	opts.AuditStatus = &approvedStatus

	list, total, err := s.repo.List(&req.Pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.LoveStoryInfo, 0, len(list))
	for i := range list {
		infos = append(infos, toLoveStoryInfo(&list[i]))
	}
	req.Pagination.Total = total
	return &req.Pagination, infos, nil
}

func (s *loveStoryService) ListByLoveID(loveID uint, page, pageSize int) (*utils.Pagination, []dto.LoveStoryInfo, error) {
	p := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByLoveID(loveID, p)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.LoveStoryInfo, 0, len(list))
	for i := range list {
		infos = append(infos, toLoveStoryInfo(&list[i]))
	}
	p.Total = total
	return p, infos, nil
}

func (s *loveStoryService) ListByUser(userID uint, page, pageSize int) (*utils.Pagination, []dto.LoveStoryInfo, error) {
	p := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByUserID(userID, p)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.LoveStoryInfo, 0, len(list))
	for i := range list {
		infos = append(infos, toLoveStoryInfo(&list[i]))
	}
	p.Total = total
	return p, infos, nil
}

func (s *loveStoryService) ListFeatured(page, pageSize int) (*utils.Pagination, []dto.LoveStoryInfo, error) {
	p := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListFeatured(p)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.LoveStoryInfo, 0, len(list))
	for i := range list {
		infos = append(infos, toLoveStoryInfo(&list[i]))
	}
	p.Total = total
	return p, infos, nil
}

func (s *loveStoryService) ListByTopic(topic string, page, pageSize int) (*utils.Pagination, []dto.LoveStoryInfo, error) {
	p := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByTopic(topic, p)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.LoveStoryInfo, 0, len(list))
	for i := range list {
		infos = append(infos, toLoveStoryInfo(&list[i]))
	}
	p.Total = total
	return p, infos, nil
}

func (s *loveStoryService) IncrView(id uint) error {
	return s.repo.IncrViewCount(id)
}

func (s *loveStoryService) IncrLike(id uint) error {
	return s.repo.IncrLikeCount(id)
}

func (s *loveStoryService) DecrLike(id uint) error {
	return s.repo.DecrLikeCount(id)
}

func (s *loveStoryService) IncrComment(id uint) error {
	return s.repo.IncrCommentCount(id)
}

func (s *loveStoryService) IncrShare(id uint) error {
	return s.repo.IncrShareCount(id)
}

func (s *loveStoryService) AdminList(req *dto.LoveStoryListRequest) (*utils.Pagination, []dto.LoveStoryInfo, error) {
	opts := repository.LoveStoryListOptions{
		LoveID:    req.LoveID,
		UserID:    req.UserID,
		MediaType: req.MediaType,
		Topic:     req.Topic,
		Featured:  req.Featured,
		Keyword:   req.Keyword,
	}
	if req.Status != nil {
		opts.Status = req.Status
	}
	if req.AuditStatus != nil {
		opts.AuditStatus = req.AuditStatus
	}
	list, total, err := s.repo.List(&req.Pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.LoveStoryInfo, 0, len(list))
	for i := range list {
		infos = append(infos, toLoveStoryInfo(&list[i]))
	}
	req.Pagination.Total = total
	return &req.Pagination, infos, nil
}

func (s *loveStoryService) Audit(id uint, auditStatus int, auditReason string) error {
	return s.repo.UpdateAuditStatus(id, auditStatus, auditReason)
}

func (s *loveStoryService) UpdateStatus(id uint, status int) error {
	return s.repo.UpdateStatus(id, status)
}

func (s *loveStoryService) SetFeatured(id uint, featured bool) error {
	return s.repo.SetFeatured(id, featured)
}

func (s *loveStoryService) BatchAudit(ids []uint, auditStatus int, auditReason string) error {
	return s.repo.BatchUpdateAuditStatus(ids, auditStatus, auditReason)
}
