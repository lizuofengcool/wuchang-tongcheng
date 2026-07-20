// Package service 同城零工兼职业务逻辑层 - 收藏
// 求职者收藏岗位 + 雇主收藏求职者
// 注意：BaseModel 无 region_id，使用 user_id 维度隔离
package service

import (
	"errors"

	"wuchang-tongcheng/internal/modules/linggong/dto"
	"wuchang-tongcheng/internal/modules/linggong/model"
	"wuchang-tongcheng/internal/modules/linggong/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrFavoriteNotFound     = errors.New("收藏不存在")
	ErrFavoriteNoPermission = errors.New("无权操作此收藏")
	ErrFavoriteDuplicate    = errors.New("已收藏过该内容")
)

// FavoriteService 收藏业务接口
type FavoriteService interface {
	// C 端
	Create(userID uint, req *dto.CreateFavoriteRequest) (*dto.FavoriteInfo, error)
	Delete(id uint, operatorID uint) error
	DeleteByTarget(userID uint, targetID uint, favoriteType string) error
	GetByID(id uint) (*dto.FavoriteInfo, error)
	List(userID uint, req *dto.FavoriteListRequest) (*utils.Pagination, []dto.FavoriteInfo, error)
	ListByUser(userID uint, page, pageSize int) (*utils.Pagination, []dto.FavoriteInfo, error)
	ListByType(userID uint, favoriteType string, page, pageSize int) (*utils.Pagination, []dto.FavoriteInfo, error)
	Exists(userID uint, targetID uint, favoriteType string) (bool, error)
	CountByTarget(targetID uint, favoriteType string) (int64, error)
}

type favoriteService struct {
	repo repository.FavoriteRepository
}

// NewFavoriteService 创建收藏 service 实例
func NewFavoriteService(repo repository.FavoriteRepository) FavoriteService {
	return &favoriteService{repo: repo}
}

// favoriteTypeText 收藏类型文本
func favoriteTypeText(t string) string {
	switch t {
	case model.FavoriteTypeLinggong:
		return "岗位"
	case model.FavoriteTypeWorker:
		return "求职者"
	case model.FavoriteTypeEmployer:
		return "雇主"
	case model.FavoriteTypeTask:
		return "任务"
	case model.FavoriteTypeSearch:
		return "搜索条件"
	}
	return ""
}

// toFavoriteInfo model -> dto
func toFavoriteInfo(f *model.LinggongFavorite) *dto.FavoriteInfo {
	return &dto.FavoriteInfo{
		ID:               f.ID,
		UserID:           f.UserID,
		TargetID:         f.TargetID,
		FavoriteType:     f.FavoriteType,
		FavoriteTypeText: favoriteTypeText(f.FavoriteType),
		Remark:           f.Remark,
		NotifyOnUpdate:   f.NotifyOnUpdate,
		CreatedAt:        f.CreatedAt,
		UpdatedAt:        f.UpdatedAt,
	}
}

// ===== C 端 =====

// Create 创建收藏
func (s *favoriteService) Create(userID uint, req *dto.CreateFavoriteRequest) (*dto.FavoriteInfo, error) {
	if req.TargetID == 0 {
		return nil, ErrFavoriteNotFound
	}
	favoriteType := req.FavoriteType
	if favoriteType == "" {
		favoriteType = model.FavoriteTypeLinggong
	}

	// 幂等：若已收藏则返回已有记录
	if existing, err := s.repo.FindByUserAndTarget(userID, req.TargetID, favoriteType); err == nil {
		return toFavoriteInfo(existing), nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	f := &model.LinggongFavorite{
		UserID:         userID,
		TargetID:       req.TargetID,
		FavoriteType:   favoriteType,
		Remark:         req.Remark,
		NotifyOnUpdate: req.NotifyOnUpdate,
	}
	if err := s.repo.Create(f); err != nil {
		return nil, err
	}
	return toFavoriteInfo(f), nil
}

// Delete 删除收藏（仅持有人本人）
func (s *favoriteService) Delete(id uint, operatorID uint) error {
	f, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrFavoriteNotFound
		}
		return err
	}
	if f.UserID != operatorID {
		return ErrFavoriteNoPermission
	}
	return s.repo.Delete(id)
}

// DeleteByTarget 按目标取消收藏（仅持有人本人）
func (s *favoriteService) DeleteByTarget(userID uint, targetID uint, favoriteType string) error {
	f, err := s.repo.FindByUserAndTarget(userID, targetID, favoriteType)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrFavoriteNotFound
		}
		return err
	}
	return s.repo.Delete(f.ID)
}

// GetByID 获取收藏详情
func (s *favoriteService) GetByID(id uint) (*dto.FavoriteInfo, error) {
	f, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFavoriteNotFound
		}
		return nil, err
	}
	return toFavoriteInfo(f), nil
}

// List 收藏列表
func (s *favoriteService) List(userID uint, req *dto.FavoriteListRequest) (*utils.Pagination, []dto.FavoriteInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.FavoriteListOptions{
		FavoriteType: req.FavoriteType,
	}
	list, total, err := s.repo.List(userID, pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.FavoriteInfo, 0, len(list))
	for i := range list {
		result = append(result, *toFavoriteInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListByUser 按用户反查所有收藏
func (s *favoriteService) ListByUser(userID uint, page, pageSize int) (*utils.Pagination, []dto.FavoriteInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByUser(userID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.FavoriteInfo, 0, len(list))
	for i := range list {
		result = append(result, *toFavoriteInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListByType 按用户和类型反查收藏
func (s *favoriteService) ListByType(userID uint, favoriteType string, page, pageSize int) (*utils.Pagination, []dto.FavoriteInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByType(userID, favoriteType, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.FavoriteInfo, 0, len(list))
	for i := range list {
		result = append(result, *toFavoriteInfo(&list[i]))
	}
	return pagination, result, nil
}

// Exists 检查是否已收藏
func (s *favoriteService) Exists(userID uint, targetID uint, favoriteType string) (bool, error) {
	return s.repo.Exists(userID, targetID, favoriteType)
}

// CountByTarget 统计目标的收藏数
func (s *favoriteService) CountByTarget(targetID uint, favoriteType string) (int64, error) {
	return s.repo.CountByTarget(targetID, favoriteType)
}
