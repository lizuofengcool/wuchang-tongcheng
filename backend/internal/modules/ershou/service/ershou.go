// Package service 同城二手物品业务逻辑层
// 依据需求文档 2.2.A.10：商品发布/分类/搜索/留言/交易
// 依据需求文档 1.5：内容审核必须做（初始版本：发布即通过，后续接第三方审核）
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
package service

import (
	"errors"
	"time"

	"wuchang-tongcheng/internal/modules/ershou/dto"
	"wuchang-tongcheng/internal/modules/ershou/model"
	"wuchang-tongcheng/internal/modules/ershou/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrErshouNotFound     = errors.New("二手物品不存在")
	ErrErshouNoPermission = errors.New("无权操作此二手物品")
	ErrErshouAudited      = errors.New("已审核的物品不能重复审核")
	ErrMessageNotFound    = errors.New("留言不存在")
)

// ErshouService 二手物品业务逻辑接口
type ErshouService interface {
	// C端
	Create(regionID uint, userID uint, userName string, userPhone string, userAvatar string, req *dto.CreateErshouRequest) (*dto.ErshouInfo, error)
	Update(id uint, operatorID uint, req *dto.UpdateErshouRequest) error
	Delete(id uint, operatorID uint) error
	GetByID(id uint, userID uint) (*dto.ErshouInfo, error)
	List(regionID uint, req *dto.ErshouListRequest) (*utils.Pagination, []dto.ErshouInfo, error)
	ListNearby(regionID uint, req *dto.ErshouNearbyRequest) (*utils.Pagination, []dto.ErshouInfo, error)
	Search(regionID uint, req *dto.ErshouSearchRequest) (*utils.Pagination, []dto.ErshouInfo, error)
	ListMine(userID uint, page, pageSize int) (*utils.Pagination, []dto.ErshouInfo, error)

	// 收藏
	Fav(userID, ershouID uint) (*dto.FavResponse, error)
	FavStatus(userID, ershouID uint) (*dto.FavResponse, error)
	ListFavs(userID uint, page, pageSize int) (*utils.Pagination, []dto.ErshouInfo, error)

	// 留言
	CreateMessage(ershouID uint, fromUserID uint, fromName string, fromAvatar string, req *dto.CreateMessageRequest) (*dto.MessageInfo, error)
	ListMessages(ershouID uint, page, pageSize int) ([]dto.MessageInfo, int64, error)

	// M端管理
	AdminList(req *dto.ErshouAdminListRequest) (*utils.Pagination, []dto.ErshouInfo, error)
	AdminGetByID(id uint) (*dto.ErshouInfo, error)
	Audit(id uint, auditStatus int, auditReason string) error
	AdminUpdateStatus(id uint, status int) error
}

type ershouService struct {
	repo repository.ErshouRepository
}

// NewErshouService 创建 service 实例
func NewErshouService(repo repository.ErshouRepository) ErshouService {
	return &ershouService{repo: repo}
}

// toErshouInfo model -> dto（含图片列表拼装，但本方法不查询图片，调用方需在合适场景补图）
func toErshouInfo(e *model.Ershou, images []string) *dto.ErshouInfo {
	info := &dto.ErshouInfo{
		ID:            e.ID,
		Title:         e.Title,
		Content:       e.Content,
		CoverImage:    e.CoverImage,
		Images:        images,
		Summary:       e.Summary,
		UserID:        e.UserID,
		UserName:      e.UserName,
		UserPhone:     e.UserPhone,
		UserAvatar:    e.UserAvatar,
		CategoryID:    e.CategoryID,
		CategoryName:  e.CategoryName,
		Price:         e.Price,
		OriginalPrice: e.OriginalPrice,
		PriceUnit:     e.PriceUnit,
		Condition:     e.Condition,
		Brand:         e.Brand,
		ContactPhone:  e.ContactPhone,
		ContactWechat: e.ContactWechat,
		Address:       e.Address,
		Latitude:      e.Latitude,
		Longitude:     e.Longitude,
		Distance:      e.Distance,
		DeliveryMethod: e.DeliveryMethod,
		IsUrgent:      e.IsUrgent,
		ExpiryTime:    e.ExpiryTime,
		ViewCount:     e.ViewCount,
		FavCount:      e.FavCount,
		MessageCount:  e.MessageCount,
		Status:        e.Status,
		AuditStatus:   e.AuditStatus,
		AuditReason:   e.AuditReason,
		PublishedAt:   e.PublishedAt,
		RegionID:      e.RegionID,
		CreatedAt:     e.CreatedAt,
		UpdatedAt:     e.UpdatedAt,
	}
	if info.PriceUnit == "" {
		info.PriceUnit = "元"
	}
	if info.Condition == "" {
		info.Condition = model.ConditionUsed
	}
	if info.DeliveryMethod == "" {
		info.DeliveryMethod = model.DeliveryFace
	}
	if info.Images == nil {
		info.Images = []string{}
	}
	return info
}

// toMessageInfo model.Message -> dto.MessageInfo
func toMessageInfo(m *model.ErshouMessage) *dto.MessageInfo {
	return &dto.MessageInfo{
		ID:         m.ID,
		ErshouID:   m.ErshouID,
		FromUserID: m.FromUserID,
		FromName:   m.FromName,
		FromAvatar: m.FromAvatar,
		Content:    m.Content,
		IsRead:     m.IsRead,
		CreatedAt:  m.CreatedAt,
	}
}

// ===== C端 =====

// Create 发布二手物品
func (s *ershouService) Create(regionID uint, userID uint, userName string, userPhone string, userAvatar string, req *dto.CreateErshouRequest) (*dto.ErshouInfo, error) {
	// 过期时间
	expireDays := req.ExpireDays
	if expireDays <= 0 {
		expireDays = 30
	}
	expiryTime := time.Now().AddDate(0, 0, expireDays)

	e := &model.Ershou{
		Title:         req.Title,
		Content:       req.Content,
		CoverImage:    req.CoverImage,
		Summary:       req.Summary,
		UserID:        userID,
		UserName:      userName,
		UserPhone:     userPhone,
		UserAvatar:    userAvatar,
		CategoryID:    req.CategoryID,
		Price:         req.Price,
		OriginalPrice: req.OriginalPrice,
		PriceUnit:     req.PriceUnit,
		Condition:     req.Condition,
		Brand:         req.Brand,
		ContactPhone:  req.ContactPhone,
		ContactWechat: req.ContactWechat,
		Address:       req.Address,
		Latitude:      req.Latitude,
		Longitude:     req.Longitude,
		DeliveryMethod: req.DeliveryMethod,
		IsUrgent:      req.IsUrgent,
		ExpiryTime:    &expiryTime,
		Status:        req.Status,
		// 初始审核状态：依据 1.5 内容审核必须做。
		// MVP 阶段简化为发布即通过（AuditApproved），
		// 后续接入第三方审核时改为 AuditPending，由审核回调更新状态。
		AuditStatus: model.AuditApproved,
	}
	e.RegionID = regionID

	// 默认值兜底
	if e.PriceUnit == "" {
		e.PriceUnit = "元"
	}
	if e.Condition == "" {
		e.Condition = model.ConditionUsed
	}
	if e.DeliveryMethod == "" {
		e.DeliveryMethod = model.DeliveryFace
	}

	// 发布时间
	if req.Status == model.StatusPublished {
		now := time.Now()
		e.PublishedAt = &now
	}

	if err := s.repo.Create(e); err != nil {
		return nil, err
	}

	// 保存图片子表
	if len(req.Images) > 0 {
		_ = s.repo.ReplaceImages(e.ID, req.Images)
	}

	return toErshouInfo(e, req.Images), nil
}

// Update 更新二手物品（仅发布者本人）
func (s *ershouService) Update(id uint, operatorID uint, req *dto.UpdateErshouRequest) error {
	e, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrErshouNotFound
		}
		return err
	}
	// 用户隔离：仅本人可改
	if e.UserID != operatorID {
		return ErrErshouNoPermission
	}

	fields := map[string]interface{}{}
	if req.Title != "" {
		fields["title"] = req.Title
	}
	if req.Content != "" {
		fields["content"] = req.Content
	}
	if req.CoverImage != "" {
		fields["cover_image"] = req.CoverImage
	}
	if req.Summary != "" {
		fields["summary"] = req.Summary
	}
	fields["category_id"] = req.CategoryID
	fields["price"] = req.Price
	fields["original_price"] = req.OriginalPrice
	if req.PriceUnit != "" {
		fields["price_unit"] = req.PriceUnit
	}
	if req.Condition != "" {
		fields["condition"] = req.Condition
	}
	if req.Brand != "" {
		fields["brand"] = req.Brand
	}
	if req.ContactPhone != "" {
		fields["contact_phone"] = req.ContactPhone
	}
	if req.ContactWechat != "" {
		fields["contact_wechat"] = req.ContactWechat
	}
	if req.Address != "" {
		fields["address"] = req.Address
	}
	fields["latitude"] = req.Latitude
	fields["longitude"] = req.Longitude
	if req.DeliveryMethod != "" {
		fields["delivery_method"] = req.DeliveryMethod
	}
	if req.IsUrgent != nil {
		fields["is_urgent"] = *req.IsUrgent
	}
	if req.ExpireDays > 0 {
		expiryTime := time.Now().AddDate(0, 0, req.ExpireDays)
		fields["expiry_time"] = &expiryTime
	}

	// 状态变更：0→1 设发布时间；其他直接改 status
	if req.Status == model.StatusPublished && e.Status != model.StatusPublished {
		now := time.Now()
		fields["status"] = model.StatusPublished
		fields["published_at"] = &now
		// 重新发布时重置审核状态
		fields["audit_status"] = model.AuditApproved
	} else if req.Status == model.StatusDraft || req.Status == model.StatusOffline {
		fields["status"] = req.Status
	}

	if len(fields) == 0 && req.Images == nil {
		return nil
	}

	if err := s.repo.UpdateFields(id, fields); err != nil {
		return err
	}

	// 更新图片子表
	if req.Images != nil {
		if err := s.repo.ReplaceImages(id, req.Images); err != nil {
			return err
		}
	}

	return nil
}

// Delete 删除二手物品（仅发布者本人）
func (s *ershouService) Delete(id uint, operatorID uint) error {
	e, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrErshouNotFound
		}
		return err
	}
	if e.UserID != operatorID {
		return ErrErshouNoPermission
	}
	return s.repo.Delete(id)
}

// GetByID 获取详情（同时增加浏览量，登录用户会标记留言为已读）
func (s *ershouService) GetByID(id uint, userID uint) (*dto.ErshouInfo, error) {
	e, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrErshouNotFound
		}
		return nil, err
	}

	// 增加浏览量（异步可优化，这里同步执行）
	_ = s.repo.IncrViewCount(id)
	e.ViewCount++

	// 拼装图片
	images := []string{}
	if imgs, err := s.repo.ListImages(id); err == nil {
		for _, img := range imgs {
			images = append(images, img.URL)
		}
	}

	info := toErshouInfo(e, images)

	// 当前用户是否已收藏
	if userID > 0 {
		if hasFaved, err := s.repo.FavExists(userID, id); err == nil {
			info.HasFaved = hasFaved
		}
	}

	// 标记留言已读（如果当前用户是发布者）
	if userID > 0 && e.UserID == userID {
		_ = s.repo.MarkMessagesRead(id, userID)
	}

	return info, nil
}

// List 列表查询（C端）
func (s *ershouService) List(regionID uint, req *dto.ErshouListRequest) (*utils.Pagination, []dto.ErshouInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.ListOptions{
		CategoryID: req.CategoryID,
		Keyword:    req.Keyword,
		MinPrice:   req.MinPrice,
		MaxPrice:   req.MaxPrice,
		Condition:  req.Condition,
		Brand:      req.Brand,
		IsUrgent:   req.IsUrgent,
		Sort:       req.Sort,
		Status:     1, // 仅已发布
	}

	list, total, err := s.repo.List(regionID, pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total

	result := make([]dto.ErshouInfo, 0, len(list))
	for i := range list {
		// 列表用 cover_image，详情才查图集，避免 N+1
		result = append(result, *toErshouInfo(&list[i], nil))
	}
	return pagination, result, nil
}

// ListNearby 附近查询
func (s *ershouService) ListNearby(regionID uint, req *dto.ErshouNearbyRequest) (*utils.Pagination, []dto.ErshouInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	radiusKm := req.RadiusKm
	if radiusKm <= 0 {
		radiusKm = 5
	}

	list, total, err := s.repo.ListNearby(regionID, pagination, req.Latitude, req.Longitude, radiusKm, repository.ListOptions{
		Status: 1,
	})
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total

	result := make([]dto.ErshouInfo, 0, len(list))
	for i := range list {
		result = append(result, *toErshouInfo(&list[i], nil))
	}
	return pagination, result, nil
}

// Search 搜索
func (s *ershouService) Search(regionID uint, req *dto.ErshouSearchRequest) (*utils.Pagination, []dto.ErshouInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	list, total, err := s.repo.Search(regionID, pagination, req.Keyword)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total

	result := make([]dto.ErshouInfo, 0, len(list))
	for i := range list {
		result = append(result, *toErshouInfo(&list[i], nil))
	}
	return pagination, result, nil
}

// ListMine 我的发布
func (s *ershouService) ListMine(userID uint, page, pageSize int) (*utils.Pagination, []dto.ErshouInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByUser(userID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.ErshouInfo, 0, len(list))
	for i := range list {
		result = append(result, *toErshouInfo(&list[i], nil))
	}
	return pagination, result, nil
}

// ===== 收藏 =====

func (s *ershouService) Fav(userID, ershouID uint) (*dto.FavResponse, error) {
	if userID == 0 {
		return nil, ErrErshouNoPermission
	}
	// 校验物品是否存在
	e, err := s.repo.FindByID(ershouID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrErshouNotFound
		}
		return nil, err
	}

	// 已收藏则取消，未收藏则创建（toggle 语义）
	exists, err := s.repo.FavExists(userID, ershouID)
	if err != nil {
		return nil, err
	}
	if exists {
		if err := s.repo.DeleteFav(userID, ershouID); err != nil {
			return nil, err
		}
		_ = s.repo.DecrFavCount(ershouID)
		return &dto.FavResponse{HasFaved: false, FavCount: e.FavCount - 1}, nil
	}

	fav := &model.ErshouFavorite{
		UserID:   userID,
		ErshouID: ershouID,
	}
	if err := s.repo.CreateFav(fav); err != nil {
		return nil, err
	}
	_ = s.repo.IncrFavCount(ershouID)
	return &dto.FavResponse{HasFaved: true, FavCount: e.FavCount + 1}, nil
}

func (s *ershouService) FavStatus(userID, ershouID uint) (*dto.FavResponse, error) {
	e, err := s.repo.FindByID(ershouID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrErshouNotFound
		}
		return nil, err
	}
	if userID == 0 {
		return &dto.FavResponse{HasFaved: false, FavCount: e.FavCount}, nil
	}
	exists, err := s.repo.FavExists(userID, ershouID)
	if err != nil {
		return nil, err
	}
	return &dto.FavResponse{HasFaved: exists, FavCount: e.FavCount}, nil
}

func (s *ershouService) ListFavs(userID uint, page, pageSize int) (*utils.Pagination, []dto.ErshouInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	favs, total, err := s.repo.ListFavs(userID, page, pageSize)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total

	// 取对应 ershouID
	ids := make([]uint, 0, len(favs))
	for _, f := range favs {
		ids = append(ids, f.ErshouID)
	}
	if len(ids) == 0 {
		return pagination, []dto.ErshouInfo{}, nil
	}

	// 查物品详情（一次性查全部，避免 N+1）
	list := make([]dto.ErshouInfo, 0, len(ids))
	for _, id := range ids {
		e, err := s.repo.FindByID(id)
		if err != nil {
			continue
		}
		list = append(list, *toErshouInfo(e, nil))
	}
	return pagination, list, nil
}

// ===== 留言 =====

func (s *ershouService) CreateMessage(ershouID uint, fromUserID uint, fromName string, fromAvatar string, req *dto.CreateMessageRequest) (*dto.MessageInfo, error) {
	// 校验物品存在
	if _, err := s.repo.FindByID(ershouID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrErshouNotFound
		}
		return nil, err
	}

	msg := &model.ErshouMessage{
		ErshouID:   ershouID,
		FromUserID: fromUserID,
		FromName:   fromName,
		FromAvatar: fromAvatar,
		Content:    req.Content,
		Status:     1,
	}
	if err := s.repo.CreateMessage(msg); err != nil {
		return nil, err
	}
	_ = s.repo.IncrMessageCount(ershouID)
	return toMessageInfo(msg), nil
}

func (s *ershouService) ListMessages(ershouID uint, page, pageSize int) ([]dto.MessageInfo, int64, error) {
	list, total, err := s.repo.ListMessages(ershouID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.MessageInfo, 0, len(list))
	for i := range list {
		result = append(result, *toMessageInfo(&list[i]))
	}
	return result, total, nil
}

// ===== M端管理 =====

func (s *ershouService) AdminList(req *dto.ErshouAdminListRequest) (*utils.Pagination, []dto.ErshouInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.AdminListOptions{
		RegionID:    req.RegionID,
		UserID:      req.UserID,
		CategoryID:  req.CategoryID,
		Status:      req.Status,
		AuditStatus: req.AuditStatus,
		Keyword:     req.Keyword,
	}
	list, total, err := s.repo.AdminList(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.ErshouInfo, 0, len(list))
	for i := range list {
		// 管理后台列表需要图片数量提示，但避免 N+1，列表不查图片
		result = append(result, *toErshouInfo(&list[i], nil))
	}
	return pagination, result, nil
}

func (s *ershouService) AdminGetByID(id uint) (*dto.ErshouInfo, error) {
	e, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrErshouNotFound
		}
		return nil, err
	}
	// 管理后台详情需要图片
	images := []string{}
	if imgs, err := s.repo.ListImages(id); err == nil {
		for _, img := range imgs {
			images = append(images, img.URL)
		}
	}
	return toErshouInfo(e, images), nil
}

// Audit 审核
func (s *ershouService) Audit(id uint, auditStatus int, auditReason string) error {
	e, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrErshouNotFound
		}
		return err
	}

	fields := map[string]interface{}{
		"audit_status": auditStatus,
		"audit_reason": auditReason,
	}

	// 审核通过且物品状态为草稿：自动发布
	if auditStatus == model.AuditApproved && e.Status == model.StatusDraft {
		now := time.Now()
		fields["status"] = model.StatusPublished
		fields["published_at"] = &now
	}
	// 审核拒绝：强制下架
	if auditStatus == model.AuditRejected && e.Status == model.StatusPublished {
		fields["status"] = model.StatusOffline
	}

	return s.repo.UpdateFields(id, fields)
}

// AdminUpdateStatus 管理后台强制下架/恢复
func (s *ershouService) AdminUpdateStatus(id uint, status int) error {
	fields := map[string]interface{}{
		"status": status,
	}
	if status == model.StatusPublished {
		now := time.Now()
		fields["published_at"] = &now
		fields["audit_status"] = model.AuditApproved
	}
	return s.repo.UpdateFields(id, fields)
}
