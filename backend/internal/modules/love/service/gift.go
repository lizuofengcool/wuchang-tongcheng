// Package service love 相亲交友业务逻辑层 - 礼物
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
// 礼物定义表为全局表（M 端配置），送礼记录复用主表（MVP 简化）
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
	ErrLoveGiftNotFound      = errors.New("礼物不存在")
	ErrLoveGiftCodeExists    = errors.New("礼物编码已存在")
	ErrLoveGiftUnavailable   = errors.New("礼物已下架")
	ErrLoveGiftLevelLimit    = errors.New("会员等级不足，无法赠送此礼物")
	ErrLoveGiftDailyLimit    = errors.New("今日赠送次数已达上限")
)

// LoveGiftService 礼物业务接口
type LoveGiftService interface {
	// M 端管理
	Create(req *dto.CreateLoveGiftRequest) (*dto.LoveGiftInfo, error)
	Update(id uint, req *dto.UpdateLoveGiftRequest) error
	Delete(id uint) error
	GetByID(id uint) (*dto.LoveGiftInfo, error)
	List(req *dto.LoveGiftListRequest) (*utils.Pagination, []dto.LoveGiftInfo, error)
	UpdateStatus(id uint, status int) error
	BatchUpdateStatus(ids []uint, status int) error

	// C 端
	ListAvailable(memberLevel int) ([]dto.LoveGiftInfo, error)
	// Send 送礼（MVP：仅记录到主表的冗余字段；完整送礼记录应落 love_gift_records 表）
	Send(userID uint, loveID uint, req *dto.SendLoveGiftRequest) (*dto.LoveGiftRecordInfo, error)
}

type loveGiftService struct {
	repo        repository.LoveGiftRepository
	recordRepo  repository.LoveGiftRecordRepository
}

// NewLoveGiftService 创建礼物 service
func NewLoveGiftService(repo repository.LoveGiftRepository, recordRepo repository.LoveGiftRecordRepository) LoveGiftService {
	return &loveGiftService{repo: repo, recordRepo: recordRepo}
}

// giftStatusText 状态文本
func giftStatusText(s int) string {
	switch s {
	case 0:
		return "下架"
	case 1:
		return "上架"
	}
	return ""
}

// toLoveGiftInfo model -> dto
func toLoveGiftInfo(g *model.LoveGift) dto.LoveGiftInfo {
	return dto.LoveGiftInfo{
		ID:                g.ID,
		GiftCode:          g.GiftCode,
		GiftName:          g.GiftName,
		Category:          g.Category,
		Description:       g.Description,
		Icon:              g.Icon,
		AnimationURL:      g.AnimationURL,
		AnimationType:     g.AnimationType,
		AnimationDuration: g.AnimationDuration,
		Price:             g.Price,
		OriginalPrice:     g.OriginalPrice,
		DiscountPrice:     g.DiscountPrice,
		MemberLevel:       g.MemberLevel,
		CharmValue:        g.CharmValue,
		IsLimited:         g.IsLimited,
		IsAnimated:        g.IsAnimated,
		IsCombo:           g.IsCombo,
		ComboMin:          g.ComboMin,
		ComboMax:          g.ComboMax,
		DailyLimit:        g.DailyLimit,
		Sort:              g.Sort,
		Status:            g.Status,
		StatusText:        giftStatusText(g.Status),
		StartAt:           g.StartAt,
		EndAt:             g.EndAt,
		CreatedAt:         g.CreatedAt,
		UpdatedAt:         g.UpdatedAt,
	}
}

// ===== M 端 =====

func (s *loveGiftService) Create(req *dto.CreateLoveGiftRequest) (*dto.LoveGiftInfo, error) {
	// 检查 code 唯一
	if req.GiftCode != "" {
		if existing, err := s.repo.FindByCode(req.GiftCode); err == nil && existing != nil {
			return nil, ErrLoveGiftCodeExists
		}
	}
	g := &model.LoveGift{
		GiftCode:          req.GiftCode,
		GiftName:          req.GiftName,
		Category:          req.Category,
		Description:       req.Description,
		Icon:              req.Icon,
		AnimationURL:      req.AnimationURL,
		AnimationType:     req.AnimationType,
		AnimationDuration: req.AnimationDuration,
		Price:             req.Price,
		OriginalPrice:     req.OriginalPrice,
		DiscountPrice:     req.DiscountPrice,
		MemberLevel:       req.MemberLevel,
		CharmValue:        req.CharmValue,
		IsLimited:         req.IsLimited,
		IsAnimated:        req.IsAnimated,
		IsCombo:           req.IsCombo,
		ComboMin:          req.ComboMin,
		ComboMax:          req.ComboMax,
		DailyLimit:        req.DailyLimit,
		Sort:              req.Sort,
		Status:            1,
	}
	if req.Status == 0 {
		g.Status = 0
	}
	if g.Category == "" {
		g.Category = "common"
	}
	if err := s.repo.Create(g); err != nil {
		return nil, err
	}
	info := toLoveGiftInfo(g)
	return &info, nil
}

func (s *loveGiftService) Update(id uint, req *dto.UpdateLoveGiftRequest) error {
	if _, err := s.repo.FindByID(id); err != nil {
		return ErrLoveGiftNotFound
	}
	fields := map[string]interface{}{}
	if req.GiftName != nil {
		fields["gift_name"] = *req.GiftName
	}
	if req.Category != nil {
		fields["category"] = *req.Category
	}
	if req.Description != nil {
		fields["description"] = *req.Description
	}
	if req.Icon != nil {
		fields["icon"] = *req.Icon
	}
	if req.AnimationURL != nil {
		fields["animation_url"] = *req.AnimationURL
	}
	if req.AnimationType != nil {
		fields["animation_type"] = *req.AnimationType
	}
	if req.AnimationDuration != nil {
		fields["animation_duration"] = *req.AnimationDuration
	}
	if req.Price != nil {
		fields["price"] = *req.Price
	}
	if req.OriginalPrice != nil {
		fields["original_price"] = *req.OriginalPrice
	}
	if req.DiscountPrice != nil {
		fields["discount_price"] = *req.DiscountPrice
	}
	if req.MemberLevel != nil {
		fields["member_level"] = *req.MemberLevel
	}
	if req.CharmValue != nil {
		fields["charm_value"] = *req.CharmValue
	}
	if req.IsLimited != nil {
		fields["is_limited"] = *req.IsLimited
	}
	if req.IsAnimated != nil {
		fields["is_animated"] = *req.IsAnimated
	}
	if req.IsCombo != nil {
		fields["is_combo"] = *req.IsCombo
	}
	if req.ComboMin != nil {
		fields["combo_min"] = *req.ComboMin
	}
	if req.ComboMax != nil {
		fields["combo_max"] = *req.ComboMax
	}
	if req.DailyLimit != nil {
		fields["daily_limit"] = *req.DailyLimit
	}
	if req.Sort != nil {
		fields["sort"] = *req.Sort
	}
	if req.Status != nil {
		fields["status"] = *req.Status
	}
	if len(fields) == 0 {
		return nil
	}
	return s.repo.UpdateFields(id, fields)
}

func (s *loveGiftService) Delete(id uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		return ErrLoveGiftNotFound
	}
	return s.repo.Delete(id)
}

func (s *loveGiftService) GetByID(id uint) (*dto.LoveGiftInfo, error) {
	g, err := s.repo.FindByID(id)
	if err != nil {
		return nil, ErrLoveGiftNotFound
	}
	info := toLoveGiftInfo(g)
	return &info, nil
}

func (s *loveGiftService) List(req *dto.LoveGiftListRequest) (*utils.Pagination, []dto.LoveGiftInfo, error) {
	opts := repository.LoveGiftListOptions{
		Category: req.Category,
	}
	if req.MemberLevel != nil {
		opts.MemberLevel = req.MemberLevel
	}
	if req.Status != nil {
		opts.Status = req.Status
	}
	list, total, err := s.repo.List(&req.Pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.LoveGiftInfo, 0, len(list))
	for i := range list {
		infos = append(infos, toLoveGiftInfo(&list[i]))
	}
	req.Pagination.Total = total
	return &req.Pagination, infos, nil
}

func (s *loveGiftService) UpdateStatus(id uint, status int) error {
	if _, err := s.repo.FindByID(id); err != nil {
		return ErrLoveGiftNotFound
	}
	return s.repo.UpdateFields(id, map[string]interface{}{"status": status})
}

func (s *loveGiftService) BatchUpdateStatus(ids []uint, status int) error {
	return s.repo.BatchUpdateStatus(ids, status)
}

// ===== C 端 =====

// ListAvailable 列出当前会员等级可用的礼物（仅上架且 member_level <= 入参）
func (s *loveGiftService) ListAvailable(memberLevel int) ([]dto.LoveGiftInfo, error) {
	list, err := s.repo.ListAvailable(memberLevel)
	if err != nil {
		return nil, err
	}
	infos := make([]dto.LoveGiftInfo, 0, len(list))
	for i := range list {
		infos = append(infos, toLoveGiftInfo(&list[i]))
	}
	return infos, nil
}

// Send 送礼（MVP 简化：仅校验 + 计算总价 + 返回记录信息）
// 完整实现应落 love_gift_records 表 + 扣金币 + 通知接收方
func (s *loveGiftService) Send(userID uint, loveID uint, req *dto.SendLoveGiftRequest) (*dto.LoveGiftRecordInfo, error) {
	g, err := s.repo.FindByID(req.GiftID)
	if err != nil {
		return nil, ErrLoveGiftNotFound
	}
	if g.Status != 1 {
		return nil, ErrLoveGiftUnavailable
	}
	// 时间限制（限时礼物）
	now := time.Now()
	if g.StartAt != nil && now.Before(*g.StartAt) {
		return nil, ErrLoveGiftUnavailable
	}
	if g.EndAt != nil && now.After(*g.EndAt) {
		return nil, ErrLoveGiftUnavailable
	}

	// 数量默认 1
	count := req.Count
	if count <= 0 {
		count = 1
	}
	if g.IsCombo && g.ComboMax > 0 && count > g.ComboMax {
		count = g.ComboMax
	}
	if g.IsCombo && g.ComboMin > 0 && count < g.ComboMin {
		count = g.ComboMin
	}

	totalPrice := g.Price * float64(count)
	// 折扣价优先
	if g.DiscountPrice > 0 {
		totalPrice = g.DiscountPrice * float64(count)
	}

	record := &dto.LoveGiftRecordInfo{
		GiftID:      g.ID,
		GiftName:    g.GiftName,
		GiftIcon:    g.Icon,
		GiftPrice:   g.Price,
		Count:       count,
		TotalPrice:  totalPrice,
		CharmValue:  g.CharmValue * count,
		FromUserID:  userID,
		FromLoveID:  loveID,
		ToUserID:    req.ToUserID,
		ToLoveID:    req.ToLoveID,
		Message:     req.Message,
		IsCombo:     req.IsCombo,
		MatchID:     req.MatchID,
		SessionID:   req.SessionID,
	}
	return record, nil
}
