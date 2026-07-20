// Package service 同城114业务逻辑层 - 收藏
// 依据 v3.2.1 架构方案：对标大众点评/美团 收藏夹
// 支持商户/团购/优惠券收藏 + 收藏分组
package service

import (
	"errors"

	"wuchang-tongcheng/internal/modules/dh114/dto"
	"wuchang-tongcheng/internal/modules/dh114/model"
	"wuchang-tongcheng/internal/modules/dh114/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrFavoriteNotFound      = errors.New("收藏不存在")
	ErrFavoriteExists         = errors.New("已收藏过")
	ErrFavoriteNotExists      = errors.New("未收藏")
	ErrFavoriteNoPermission   = errors.New("无权操作此收藏")
)

// FavoriteService 收藏业务接口
type FavoriteService interface {
	// C 端
	Create(userID uint, req *dto.CreateFavoriteRequest) (*dto.FavoriteInfo, error)
	Delete(id uint, userID uint) error
	DeleteByTarget(userID uint, dh114ID uint, favoriteType string) error
	Update(id uint, userID uint, req *dto.UpdateFavoriteRequest) error
	GetByID(id uint) (*dto.FavoriteInfo, error)
	List(userID uint, req *dto.FavoriteListRequest) (*utils.Pagination, []dto.FavoriteInfo, error)
	ListByType(userID uint, favoriteType string, page, pageSize int) (*utils.Pagination, []dto.FavoriteInfo, error)
	ListByGroup(userID uint, groupID uint, page, pageSize int) (*utils.Pagination, []dto.FavoriteInfo, error)

	// 状态查询
	HasFaved(userID uint, dh114ID uint, favoriteType string) (bool, error)
	HasFavedBatch(userID uint, ids []uint, favoriteType string) (map[uint]bool, error)
}

type favoriteService struct {
	repo repository.FavoriteRepository
}

// NewFavoriteService 创建收藏 service 实例
func NewFavoriteService(repo repository.FavoriteRepository) FavoriteService {
	return &favoriteService{repo: repo}
}

// toFavoriteInfo model -> dto
func toFavoriteInfo(f *model.Dh114Favorite) *dto.FavoriteInfo {
	return &dto.FavoriteInfo{
		ID:           f.ID,
		UserID:       f.UserID,
		Dh114ID:      f.Dh114ID,
		BusinessID:   f.BusinessID,
		FavoriteType: f.FavoriteType,
		GroupID:      f.GroupID,
		Remark:       f.Remark,
		CreatedAt:    f.CreatedAt,
		UpdatedAt:    f.UpdatedAt,
	}
}

// Create 创建收藏
func (s *favoriteService) Create(userID uint, req *dto.CreateFavoriteRequest) (*dto.FavoriteInfo, error) {
	favType := req.FavoriteType
	if favType == "" {
		favType = model.FavoriteTypeBusiness
	}

	// 检查是否已收藏
	exists, err := s.repo.Exists(userID, req.Dh114ID, favType)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrFavoriteExists
	}

	fav := &model.Dh114Favorite{
		UserID:       userID,
		Dh114ID:      req.Dh114ID,
		FavoriteType: favType,
		GroupID:      req.GroupID,
		Remark:       req.Remark,
	}
	if err := s.repo.Create(fav); err != nil {
		return nil, err
	}
	return toFavoriteInfo(fav), nil
}

// Delete 删除收藏（按记录 ID）
func (s *favoriteService) Delete(id uint, userID uint) error {
	fav, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrFavoriteNotFound
		}
		return err
	}
	if fav.UserID != userID {
		return ErrFavoriteNoPermission
	}
	return s.repo.Delete(id)
}

// DeleteByTarget 按目标删除收藏
func (s *favoriteService) DeleteByTarget(userID uint, dh114ID uint, favoriteType string) error {
	if favoriteType == "" {
		favoriteType = model.FavoriteTypeBusiness
	}
	return s.repo.DeleteByUserAndTarget(userID, dh114ID, favoriteType)
}

// Update 更新收藏（备注/分组）
func (s *favoriteService) Update(id uint, userID uint, req *dto.UpdateFavoriteRequest) error {
	fav, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrFavoriteNotFound
		}
		return err
	}
	if fav.UserID != userID {
		return ErrFavoriteNoPermission
	}

	fields := make(map[string]interface{})
	if req.GroupID != nil {
		fields["group_id"] = *req.GroupID
	}
	if req.Remark != nil {
		fields["remark"] = *req.Remark
	}
	if len(fields) == 0 {
		return nil
	}
	return s.repo.Update(id, fields)
}

// GetByID 获取收藏详情
func (s *favoriteService) GetByID(id uint) (*dto.FavoriteInfo, error) {
	fav, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFavoriteNotFound
		}
		return nil, err
	}
	return toFavoriteInfo(fav), nil
}

// List 收藏列表
func (s *favoriteService) List(userID uint, req *dto.FavoriteListRequest) (*utils.Pagination, []dto.FavoriteInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	query := repository.FavoriteListQuery{
		UserID:       userID,
		FavoriteType: req.FavoriteType,
		GroupID:      req.GroupID,
	}
	list, total, err := s.repo.List(query, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.FavoriteInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toFavoriteInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// ListByType 按类型列出收藏
func (s *favoriteService) ListByType(userID uint, favoriteType string, page, pageSize int) (*utils.Pagination, []dto.FavoriteInfo, error) {
	if favoriteType == "" {
		favoriteType = model.FavoriteTypeBusiness
	}
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByType(userID, favoriteType, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.FavoriteInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toFavoriteInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// ListByGroup 按分组列出收藏
func (s *favoriteService) ListByGroup(userID uint, groupID uint, page, pageSize int) (*utils.Pagination, []dto.FavoriteInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByGroup(userID, groupID, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.FavoriteInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toFavoriteInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// HasFaved 检查是否已收藏
func (s *favoriteService) HasFaved(userID uint, dh114ID uint, favoriteType string) (bool, error) {
	if favoriteType == "" {
		favoriteType = model.FavoriteTypeBusiness
	}
	return s.repo.Exists(userID, dh114ID, favoriteType)
}

// HasFavedBatch 批量检查是否已收藏
func (s *favoriteService) HasFavedBatch(userID uint, ids []uint, favoriteType string) (map[uint]bool, error) {
	if favoriteType == "" {
		favoriteType = model.FavoriteTypeBusiness
	}
	return s.repo.HasFavedBatch(userID, ids, favoriteType)
}
