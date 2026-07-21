// Package service 同城商城业务逻辑层 - 收货地址
package service

import (
	"errors"

	"wuchang-tongcheng/internal/modules/mall/dto"
	"wuchang-tongcheng/internal/modules/mall/model"
	"wuchang-tongcheng/internal/modules/mall/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrAddressNotFound      = errors.New("收货地址不存在")
	ErrAddressLimitExceeded = errors.New("收货地址数量超出限制")
	ErrAddressNotOwner      = errors.New("无权操作他人收货地址")
)

// 地址数量上限（按用户）
const addressMaxPerUser = 20

// AddressService 收货地址业务接口
type AddressService interface {
	Create(regionID, userID uint, req *dto.CreateAddressRequest) (*dto.AddressInfo, error)
	Update(id, userID uint, req *dto.UpdateAddressRequest) error
	Delete(id, userID uint) error
	GetByID(id uint) (*dto.AddressInfo, error)
	GetDefault(userID uint) (*dto.AddressInfo, error)
	ListByUser(userID uint) ([]dto.AddressInfo, error)
	List(req *dto.AddressListRequest) (*utils.Pagination, []dto.AddressInfo, error)
	SetDefault(userID, id uint) error
}

type addressService struct {
	repo repository.AddressRepository
}

// NewAddressService 创建收货地址 service 实例
func NewAddressService(repo repository.AddressRepository) AddressService {
	return &addressService{repo: repo}
}

// toAddressInfo model -> dto
func toAddressInfo(a *model.Address) *dto.AddressInfo {
	return &dto.AddressInfo{
		ID:           a.ID,
		UserID:       a.UserID,
		Name:         a.Name,
		Phone:        a.Phone,
		ZipCode:      a.ZipCode,
		Province:     a.Province,
		City:         a.City,
		District:     a.District,
		ProvinceCode: a.ProvinceCode,
		CityCode:     a.CityCode,
		DistrictCode: a.DistrictCode,
		Detail:       a.Detail,
		Latitude:     a.Latitude,
		Longitude:    a.Longitude,
		Tag:          a.Tag,
		IsDefault:    a.IsDefault == model.AddressIsDefault,
		Status:       a.Status,
		RegionID:     a.RegionID,
	}
}

// Create 创建收货地址
func (s *addressService) Create(regionID, userID uint, req *dto.CreateAddressRequest) (*dto.AddressInfo, error) {
	// 校验用户地址数量
	count, err := s.repo.CountByUser(userID)
	if err == nil && count >= addressMaxPerUser {
		return nil, ErrAddressLimitExceeded
	}

	a := &model.Address{
		UserID:       userID,
		Name:         req.Name,
		Phone:        req.Phone,
		ZipCode:      req.ZipCode,
		Province:     req.Province,
		City:         req.City,
		District:     req.District,
		ProvinceCode: req.ProvinceCode,
		CityCode:     req.CityCode,
		DistrictCode: req.DistrictCode,
		Detail:       req.Detail,
		Latitude:     req.Latitude,
		Longitude:    req.Longitude,
		Tag:          req.Tag,
		Status:       1,
	}
	a.RegionID = regionID
	if req.IsDefault {
		a.IsDefault = model.AddressIsDefault
	}

	// 如果设为默认，先清除其他默认
	if a.IsDefault == model.AddressIsDefault {
		_ = s.repo.ClearDefault(userID)
	}

	if err := s.repo.Create(a); err != nil {
		return nil, err
	}
	return toAddressInfo(a), nil
}

// Update 更新收货地址
func (s *addressService) Update(id, userID uint, req *dto.UpdateAddressRequest) error {
	a, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrAddressNotFound
		}
		return err
	}
	if a.UserID != userID {
		return ErrAddressNotOwner
	}

	fields := make(map[string]interface{})
	if req.Name != nil {
		fields["name"] = *req.Name
	}
	if req.Phone != nil {
		fields["phone"] = *req.Phone
	}
	if req.ZipCode != nil {
		fields["zip_code"] = *req.ZipCode
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
	if req.ProvinceCode != nil {
		fields["province_code"] = *req.ProvinceCode
	}
	if req.CityCode != nil {
		fields["city_code"] = *req.CityCode
	}
	if req.DistrictCode != nil {
		fields["district_code"] = *req.DistrictCode
	}
	if req.Detail != nil {
		fields["detail"] = *req.Detail
	}
	if req.Latitude != nil {
		fields["latitude"] = *req.Latitude
	}
	if req.Longitude != nil {
		fields["longitude"] = *req.Longitude
	}
	if req.Tag != nil {
		fields["tag"] = *req.Tag
	}
	if req.Status != nil {
		fields["status"] = *req.Status
	}
	if req.IsDefault != nil {
		if *req.IsDefault {
			// 先清除其他默认
			if err := s.repo.ClearDefault(userID); err != nil {
				return err
			}
			fields["is_default"] = model.AddressIsDefault
		} else {
			fields["is_default"] = model.AddressNotDefault
		}
	}

	if len(fields) == 0 {
		return nil
	}
	return s.repo.UpdateFields(id, fields)
}

// Delete 删除收货地址
func (s *addressService) Delete(id, userID uint) error {
	a, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrAddressNotFound
		}
		return err
	}
	if a.UserID != userID {
		return ErrAddressNotOwner
	}
	return s.repo.Delete(id)
}

// GetByID 获取收货地址详情
func (s *addressService) GetByID(id uint) (*dto.AddressInfo, error) {
	a, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAddressNotFound
		}
		return nil, err
	}
	return toAddressInfo(a), nil
}

// GetDefault 获取用户默认地址
func (s *addressService) GetDefault(userID uint) (*dto.AddressInfo, error) {
	a, err := s.repo.FindDefault(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAddressNotFound
		}
		return nil, err
	}
	return toAddressInfo(a), nil
}

// ListByUser 按用户列出收货地址
func (s *addressService) ListByUser(userID uint) ([]dto.AddressInfo, error) {
	list, err := s.repo.ListByUser(userID)
	if err != nil {
		return nil, err
	}
	infos := make([]dto.AddressInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toAddressInfo(&list[i]))
	}
	return infos, nil
}

// List 收货地址列表（管理后台）
func (s *addressService) List(req *dto.AddressListRequest) (*utils.Pagination, []dto.AddressInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.AddressListOptions{
		UserID:   req.UserID,
		Keyword:  req.Keyword,
		RegionID: req.RegionID,
	}
	list, total, err := s.repo.List(opts, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.AddressInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toAddressInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// SetDefault 设默认地址
func (s *addressService) SetDefault(userID, id uint) error {
	a, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrAddressNotFound
		}
		return err
	}
	if a.UserID != userID {
		return ErrAddressNotOwner
	}
	return s.repo.SetDefault(userID, id)
}
