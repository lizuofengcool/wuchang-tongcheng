// Package service 拍卖业务逻辑层
// 依据 v3.2.1 架构方案：状态机 pending → live → ended → deal / failed
// 自动延拍：截拍前 5 分钟内出价，自动延 5 分钟
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
	ErrAuctionNotFound     = errors.New("拍卖不存在")
	ErrAuctionNotActive    = errors.New("拍卖未在进行中")
	ErrAuctionEnded        = errors.New("拍卖已截拍")
	ErrBidPriceTooLow      = errors.New("出价必须高于当前最高价 + 加价幅度")
	ErrBidSelfItem         = errors.New("不能竞拍自己的商品")
	ErrAuctionHasBidders   = errors.New("已有出价记录，无法取消")
)

// AuctionService 拍卖业务接口
type AuctionService interface {
	Create(ershouID, userID uint, req *dto.AuctionCreateRequest) (*dto.AuctionResponse, error)
	GetByErshouID(ershouID uint) (*dto.AuctionResponse, error)
	Bid(ershouID, userID uint, req *dto.AuctionBidRequest, ip, ua string) (*dto.AuctionResponse, error)
	EndManually(ershouID, userID uint) (*dto.AuctionResponse, error)
	List(regionID uint, status *int, pagination *utils.Pagination) (*utils.Pagination, []dto.AuctionResponse, error)
}

type auctionService struct {
	repo      repository.AuctionRepository
	ershouRepo repository.ErshouRepository
}

// NewAuctionService 创建拍卖 service 实例
func NewAuctionService(repo repository.AuctionRepository, ershouRepo repository.ErshouRepository) AuctionService {
	return &auctionService{repo: repo, ershouRepo: ershouRepo}
}

func auctionStatusText(status int) string {
	switch status {
	case model.AuctionStatusPending:
		return "待开拍"
	case model.AuctionStatusActive:
		return "进行中"
	case model.AuctionStatusEnded:
		return "已截拍"
	case model.AuctionStatusCanceled:
		return "已取消"
	case model.AuctionStatusSold:
		return "已成交"
	case model.AuctionStatusFailed:
		return "流拍"
	}
	return "未知"
}

func (s *auctionService) Create(ershouID, userID uint, req *dto.AuctionCreateRequest) (*dto.AuctionResponse, error) {
	e, err := s.ershouRepo.FindByID(ershouID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrErshouNotFound
		}
		return nil, err
	}
	if e.UserID != userID {
		return nil, ErrErshouNoPermission
	}

	// 检查是否已存在拍卖
	if existing, err := s.repo.FindByErshouID(ershouID); err == nil && existing != nil {
		return nil, errors.New("该商品已存在拍卖")
	}

	startTime := req.StartTime
	if startTime == nil {
		now := time.Now()
		startTime = &now
	}
	status := model.AuctionStatusPending
	if startTime.Before(time.Now()) {
		status = model.AuctionStatusActive
	}

	auction := &model.ErshouAuction{
		ErshouID:          ershouID,
		StartPrice:        req.StartPrice,
		StepPrice:         req.StepPrice,
		ReservePrice:      req.ReservePrice,
		BondAmount:        req.BondAmount,
		CurrentBidPrice:   req.StartPrice,
		Status:            status,
		StartTime:         startTime,
		EndTime:           req.EndTime,
		AutoExtendEnabled: req.AutoExtendEnabled,
	}
	auction.RegionID = e.RegionID

	if err := s.repo.Create(auction); err != nil {
		return nil, err
	}

	// 主表标记为拍卖商品
	_ = s.ershouRepo.UpdateFields(ershouID, map[string]interface{}{
		"is_auction":          true,
		"auction_start_time":  startTime,
		"auction_end_time":    req.EndTime,
	})

	return s.toAuctionResponse(auction), nil
}

func (s *auctionService) GetByErshouID(ershouID uint) (*dto.AuctionResponse, error) {
	auction, err := s.repo.FindByErshouID(ershouID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAuctionNotFound
		}
		return nil, err
	}
	return s.toAuctionResponse(auction), nil
}

// Bid 出价（含自动延拍逻辑）
func (s *auctionService) Bid(ershouID, userID uint, req *dto.AuctionBidRequest, ip, ua string) (*dto.AuctionResponse, error) {
	e, err := s.ershouRepo.FindByID(ershouID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrErshouNotFound
		}
		return nil, err
	}
	if e.UserID == userID {
		return nil, ErrBidSelfItem
	}

	auction, err := s.repo.FindByErshouID(ershouID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAuctionNotFound
		}
		return nil, err
	}
	if auction.Status != model.AuctionStatusActive {
		return nil, ErrAuctionNotActive
	}
	if auction.EndTime != nil && auction.EndTime.Before(time.Now()) {
		return nil, ErrAuctionEnded
	}

	// 校验出价 ≥ 当前最高价 + 加价幅度
	minBid := auction.CurrentBidPrice + auction.StepPrice
	if auction.StepPrice == 0 {
		minBid = auction.CurrentBidPrice
	}
	if req.BidPrice < minBid {
		return nil, ErrBidPriceTooLow
	}

	// 写入出价记录
	bid := &model.ErshouAuctionBid{
		AuctionID: auction.ID,
		UserID:    userID,
		BidPrice:  req.BidPrice,
		BidTime:   time.Now(),
		IP:        ip,
		UserAgent: ua,
	}
	if err := s.repo.CreateBid(bid); err != nil {
		return nil, err
	}

	// 更新当前最高价
	fields := map[string]interface{}{
		"current_bid_price":   req.BidPrice,
		"current_bid_user_id": userID,
		"bid_count":           auction.BidCount + 1,
	}

	// 自动延拍：截拍前 5 分钟内出价，自动延 5 分钟
	if auction.AutoExtendEnabled && auction.EndTime != nil {
		fiveMinBefore := auction.EndTime.Add(-5 * time.Minute)
		if time.Now().After(fiveMinBefore) {
			newEnd := auction.EndTime.Add(5 * time.Minute)
			fields["end_time"] = newEnd
			auction.EndTime = &newEnd
		}
	}

	if err := s.repo.Update(auction.ID, fields); err != nil {
		return nil, err
	}

	auction.CurrentBidPrice = req.BidPrice
	auction.CurrentBidUserID = userID
	auction.BidCount++
	return s.toAuctionResponse(auction), nil
}

// EndManually 手动截拍（仅卖家）
func (s *auctionService) EndManually(ershouID, userID uint) (*dto.AuctionResponse, error) {
	e, err := s.ershouRepo.FindByID(ershouID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrErshouNotFound
		}
		return nil, err
	}
	if e.UserID != userID {
		return nil, ErrErshouNoPermission
	}
	auction, err := s.repo.FindByErshouID(ershouID)
	if err != nil {
		return nil, ErrAuctionNotFound
	}
	if auction.Status != model.AuctionStatusActive {
		return nil, ErrAuctionNotActive
	}

	now := time.Now()
	fields := map[string]interface{}{
		"status":    model.AuctionStatusEnded,
		"end_time":  &now,
		"closed_at": &now,
	}

	// 判断是否成交
	if auction.BidCount > 0 && auction.CurrentBidPrice >= auction.ReservePrice {
		fields["status"] = model.AuctionStatusSold
		fields["winner_id"] = auction.CurrentBidUserID
		fields["winner_price"] = auction.CurrentBidPrice
	} else if auction.BidCount == 0 {
		fields["status"] = model.AuctionStatusFailed
	} else {
		// 有出价但未达保留价 → 流拍
		fields["status"] = model.AuctionStatusFailed
	}

	if err := s.repo.Update(auction.ID, fields); err != nil {
		return nil, err
	}

	updated, _ := s.repo.FindByID(auction.ID)
	return s.toAuctionResponse(updated), nil
}

func (s *auctionService) List(regionID uint, status *int, pagination *utils.Pagination) (*utils.Pagination, []dto.AuctionResponse, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	list, total, err := s.repo.List(repository.AuctionListQuery{
		RegionID: regionID,
		Status:   status,
	}, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.AuctionResponse, 0, len(list))
	for i := range list {
		result = append(result, *s.toAuctionResponse(&list[i]))
	}
	return pagination, result, nil
}

func (s *auctionService) toAuctionResponse(a *model.ErshouAuction) *dto.AuctionResponse {
	return &dto.AuctionResponse{
		ID:                a.ID,
		ErshouID:          a.ErshouID,
		StartPrice:        a.StartPrice,
		StepPrice:         a.StepPrice,
		ReservePrice:      a.ReservePrice,
		BondAmount:        a.BondAmount,
		CurrentBidPrice:   a.CurrentBidPrice,
		CurrentBidUserID:  a.CurrentBidUserID,
		BidCount:          a.BidCount,
		WatcherCount:      a.WatcherCount,
		Status:            a.Status,
		StatusText:        auctionStatusText(a.Status),
		StartTime:         a.StartTime,
		EndTime:           a.EndTime,
		AutoExtendEnabled: a.AutoExtendEnabled,
		WinnerID:          a.WinnerID,
		WinnerPrice:       a.WinnerPrice,
		ClosedAt:          a.ClosedAt,
		CreatedAt:         a.CreatedAt,
		UpdatedAt:         a.UpdatedAt,
	}
}
