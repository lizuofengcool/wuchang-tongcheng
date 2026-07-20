// Package service love 相亲交友业务逻辑层 - 详细资料
package service

import (
	"errors"

	"wuchang-tongcheng/internal/modules/love/dto"
	"wuchang-tongcheng/internal/modules/love/model"
	"wuchang-tongcheng/internal/modules/love/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrLoveProfileNotFound = errors.New("详细资料不存在")
)

// LoveProfileService 详细资料业务接口
type LoveProfileService interface {
	Create(loveID, userID uint, req *dto.CreateLoveProfileRequest) (*dto.LoveProfileInfo, error)
	Update(id uint, operatorID uint, req *dto.UpdateLoveProfileRequest) error
	UpdateByLoveID(loveID, userID uint, req *dto.UpdateLoveProfileRequest) error
	GetByID(id uint) (*dto.LoveProfileInfo, error)
	GetByLoveID(loveID uint) (*dto.LoveProfileInfo, error)
	GetByUserID(userID uint) (*dto.LoveProfileInfo, error)
	List(req *dto.LoveProfileListRequest) (*utils.Pagination, []dto.LoveProfileInfo, error)
	UpdateStep(loveID uint, step int) error
}

type loveProfileService struct {
	repo repository.LoveProfileRepository
}

// NewLoveProfileService 创建详细资料 service
func NewLoveProfileService(repo repository.LoveProfileRepository) LoveProfileService {
	return &loveProfileService{repo: repo}
}

func toLoveProfileInfo(p *model.LoveProfile) dto.LoveProfileInfo {
	return dto.LoveProfileInfo{
		ID:                 p.ID,
		LoveID:             p.LoveID,
		UserID:             p.UserID,
		Nickname:           p.Nickname,
		Avatar:             p.Avatar,
		Gender:             p.Gender,
		Age:                p.Age,
		Height:             p.Height,
		Weight:             p.Weight,
		City:               p.City,
		District:           p.District,
		Occupation:         p.Occupation,
		Company:            p.Company,
		Industry:           p.Industry,
		Education:          p.Education,
		School:             p.School,
		Income:             p.Income,
		Marriage:           p.Marriage,
		ChildrenStatus:     p.ChildrenStatus,
		HouseStatus:        p.HouseStatus,
		CarStatus:          p.CarStatus,
		Drinking:           p.Drinking,
		Smoking:            p.Smoking,
		Exercise:           p.Exercise,
		Diet:               p.Diet,
		Sleep:              p.Sleep,
		Pets:               p.Pets,
		Languages:          p.Languages,
		Interests:          p.Interests,
		Skills:             p.Skills,
		SelfIntro:          p.SelfIntro,
		IdealPartner:       p.IdealPartner,
		IdealAgeMin:        p.IdealAgeMin,
		IdealAgeMax:        p.IdealAgeMax,
		IdealHeightMin:     p.IdealHeightMin,
		IdealHeightMax:     p.IdealHeightMax,
		IdealCities:        p.IdealCities,
		IdealEducation:     p.IdealEducation,
		IdealIncome:        p.IdealIncome,
		IdealMarriage:      p.IdealMarriage,
		IdealHouse:         p.IdealHouse,
		IdealCar:           p.IdealCar,
		IdealSmoking:       p.IdealSmoking,
		IdealDrinking:      p.IdealDrinking,
		VoiceIntroURL:      p.VoiceIntroURL,
		VoiceIntroDuration: p.VoiceIntroDuration,
		VideoIntroURL:      p.VideoIntroURL,
		VideoCover:         p.VideoCover,
		PhotoUrls:          p.PhotoUrls,
		PhotoCount:         p.PhotoCount,
		ProfileScore:       p.ProfileScore,
		CompletedStep:      p.CompletedStep,
		CompletedAt:        p.CompletedAt,
		Status:             p.Status,
		CreatedAt:          p.CreatedAt,
		UpdatedAt:          p.UpdatedAt,
	}
}

func (s *loveProfileService) Create(loveID, userID uint, req *dto.CreateLoveProfileRequest) (*dto.LoveProfileInfo, error) {
	p := &model.LoveProfile{
		LoveID:           loveID,
		UserID:           userID,
		City:             req.City,
		District:         req.District,
		Occupation:       req.Occupation,
		Company:          req.Company,
		Industry:         req.Industry,
		Education:        req.Education,
		School:           req.School,
		Income:           req.Income,
		Marriage:         req.Marriage,
		ChildrenStatus:   req.ChildrenStatus,
		HouseStatus:      req.HouseStatus,
		CarStatus:        req.CarStatus,
		Drinking:         req.Drinking,
		Smoking:           req.Smoking,
		Exercise:         req.Exercise,
		Diet:             req.Diet,
		Sleep:            req.Sleep,
		Pets:             req.Pets,
		SelfIntro:        req.SelfIntro,
		IdealPartner:     req.IdealPartner,
		IdealAgeMin:      req.IdealAgeMin,
		IdealAgeMax:      req.IdealAgeMax,
		IdealHeightMin:   req.IdealHeightMin,
		IdealHeightMax:   req.IdealHeightMax,
		IdealIncome:      req.IdealIncome,
		IdealMarriage:    req.IdealMarriage,
		IdealHouse:       req.IdealHouse,
		IdealCar:         req.IdealCar,
		IdealSmoking:     req.IdealSmoking,
		IdealDrinking:    req.IdealDrinking,
		VoiceIntroURL:    req.VoiceIntroURL,
		VoiceIntroDuration: req.VoiceIntroDuration,
		VideoIntroURL:    req.VideoIntroURL,
		VideoCover:       req.VideoCover,
		Status:           1,
	}
	if err := s.repo.Create(p); err != nil {
		return nil, err
	}
	info := toLoveProfileInfo(p)
	return &info, nil
}

func (s *loveProfileService) Update(id uint, operatorID uint, req *dto.UpdateLoveProfileRequest) error {
	p, err := s.repo.FindByID(id)
	if err != nil {
		return ErrLoveProfileNotFound
	}
	if p.UserID != operatorID {
		return ErrLoveNoPermission
	}
	applyProfileUpdate(p, req)
	return s.repo.Update(p)
}

func (s *loveProfileService) UpdateByLoveID(loveID, userID uint, req *dto.UpdateLoveProfileRequest) error {
	p, err := s.repo.FindByLoveID(loveID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 不存在则创建
			createReq := &dto.CreateLoveProfileRequest{}
			if req.City != nil {
				createReq.City = *req.City
			}
			if req.Occupation != nil {
				createReq.Occupation = *req.Occupation
			}
			_, err := s.Create(loveID, userID, createReq)
			return err
		}
		return err
	}
	if p.UserID != userID {
		return ErrLoveNoPermission
	}
	applyProfileUpdate(p, req)
	return s.repo.Update(p)
}

func applyProfileUpdate(p *model.LoveProfile, req *dto.UpdateLoveProfileRequest) {
	if req.City != nil {
		p.City = *req.City
	}
	if req.District != nil {
		p.District = *req.District
	}
	if req.Occupation != nil {
		p.Occupation = *req.Occupation
	}
	if req.Company != nil {
		p.Company = *req.Company
	}
	if req.Industry != nil {
		p.Industry = *req.Industry
	}
	if req.Education != nil {
		p.Education = *req.Education
	}
	if req.School != nil {
		p.School = *req.School
	}
	if req.Income != nil {
		p.Income = *req.Income
	}
	if req.Marriage != nil {
		p.Marriage = *req.Marriage
	}
	if req.ChildrenStatus != nil {
		p.ChildrenStatus = *req.ChildrenStatus
	}
	if req.HouseStatus != nil {
		p.HouseStatus = *req.HouseStatus
	}
	if req.CarStatus != nil {
		p.CarStatus = *req.CarStatus
	}
	if req.Drinking != nil {
		p.Drinking = *req.Drinking
	}
	if req.Smoking != nil {
		p.Smoking = *req.Smoking
	}
	if req.Exercise != nil {
		p.Exercise = *req.Exercise
	}
	if req.Diet != nil {
		p.Diet = *req.Diet
	}
	if req.Sleep != nil {
		p.Sleep = *req.Sleep
	}
	if req.Pets != nil {
		p.Pets = *req.Pets
	}
	if req.SelfIntro != nil {
		p.SelfIntro = *req.SelfIntro
	}
	if req.IdealPartner != nil {
		p.IdealPartner = *req.IdealPartner
	}
	if req.IdealAgeMin != nil {
		p.IdealAgeMin = *req.IdealAgeMin
	}
	if req.IdealAgeMax != nil {
		p.IdealAgeMax = *req.IdealAgeMax
	}
	if req.IdealHeightMin != nil {
		p.IdealHeightMin = *req.IdealHeightMin
	}
	if req.IdealHeightMax != nil {
		p.IdealHeightMax = *req.IdealHeightMax
	}
	if req.IdealIncome != nil {
		p.IdealIncome = *req.IdealIncome
	}
	if req.IdealMarriage != nil {
		p.IdealMarriage = *req.IdealMarriage
	}
	if req.IdealHouse != nil {
		p.IdealHouse = *req.IdealHouse
	}
	if req.IdealCar != nil {
		p.IdealCar = *req.IdealCar
	}
	if req.IdealSmoking != nil {
		p.IdealSmoking = *req.IdealSmoking
	}
	if req.IdealDrinking != nil {
		p.IdealDrinking = *req.IdealDrinking
	}
	if req.VoiceIntroURL != nil {
		p.VoiceIntroURL = *req.VoiceIntroURL
	}
	if req.VoiceIntroDuration != nil {
		p.VoiceIntroDuration = *req.VoiceIntroDuration
	}
	if req.VideoIntroURL != nil {
		p.VideoIntroURL = *req.VideoIntroURL
	}
	if req.VideoCover != nil {
		p.VideoCover = *req.VideoCover
	}
}

func (s *loveProfileService) GetByID(id uint) (*dto.LoveProfileInfo, error) {
	p, err := s.repo.FindByID(id)
	if err != nil {
		return nil, ErrLoveProfileNotFound
	}
	info := toLoveProfileInfo(p)
	return &info, nil
}

func (s *loveProfileService) GetByLoveID(loveID uint) (*dto.LoveProfileInfo, error) {
	p, err := s.repo.FindByLoveID(loveID)
	if err != nil {
		return nil, ErrLoveProfileNotFound
	}
	info := toLoveProfileInfo(p)
	return &info, nil
}

func (s *loveProfileService) GetByUserID(userID uint) (*dto.LoveProfileInfo, error) {
	p, err := s.repo.FindByUserID(userID)
	if err != nil {
		return nil, ErrLoveProfileNotFound
	}
	info := toLoveProfileInfo(p)
	return &info, nil
}

func (s *loveProfileService) List(req *dto.LoveProfileListRequest) (*utils.Pagination, []dto.LoveProfileInfo, error) {
	opts := repository.LoveProfileListOptions{
		Gender:  req.Gender,
		City:    req.City,
		Keyword: req.Keyword,
	}
	list, total, err := s.repo.List(&req.Pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.LoveProfileInfo, 0, len(list))
	for i := range list {
		infos = append(infos, toLoveProfileInfo(&list[i]))
	}
	req.Pagination.Total = total
	return &req.Pagination, infos, nil
}

func (s *loveProfileService) UpdateStep(loveID uint, step int) error {
	return s.repo.UpdateCompletedStep(loveID, step)
}
