// Package service love 相亲交友业务逻辑层 - 认证
package service

import (
	"errors"
	"fmt"
	"time"

	"wuchang-tongcheng/internal/modules/love/dto"
	"wuchang-tongcheng/internal/modules/love/model"
	"wuchang-tongcheng/internal/modules/love/repository"
	"wuchang-tongcheng/internal/pkg/utils"
)

var (
	ErrLoveVerificationNotFound = errors.New("认证记录不存在")
	ErrLoveVerificationExists   = errors.New("已提交过此类型认证")
)

// LoveVerificationService 认证业务接口
type LoveVerificationService interface {
	Submit(loveID, userID uint, req *dto.CreateLoveVerificationRequest) (*dto.LoveVerificationInfo, error)
	GetByID(id uint) (*dto.LoveVerificationInfo, error)
	GetByLoveID(loveID uint) ([]dto.LoveVerificationInfo, error)
	GetByUserID(userID uint) ([]dto.LoveVerificationInfo, error)
	List(req *dto.LoveVerificationListRequest) (*utils.Pagination, []dto.LoveVerificationInfo, error)
	Audit(id uint, handledBy uint, handledName string, req *dto.LoveVerificationAuditRequest) error
	CountPending() (int64, error)
}

type loveVerificationService struct {
	repo repository.LoveVerificationRepository
}

// NewLoveVerificationService 创建认证 service
func NewLoveVerificationService(repo repository.LoveVerificationRepository) LoveVerificationService {
	return &loveVerificationService{repo: repo}
}

func verifyStatusText(s int) string {
	switch s {
	case model.VerifyStatusPending:
		return "待审"
	case model.VerifyStatusApproved:
		return "通过"
	case model.VerifyStatusRejected:
		return "拒绝"
	}
	return ""
}

func verifyTypeText(t string) string {
	switch t {
	case model.VerifyTypeRealName:
		return "实名认证"
	case model.VerifyTypePhoto:
		return "照片认证"
	case model.VerifyTypeVideo:
		return "视频认证"
	case model.VerifyTypeEducation:
		return "学历认证"
	case model.VerifyTypeProperty:
		return "房产认证"
	case model.VerifyTypeCar:
		return "车辆认证"
	}
	return ""
}

// maskIDCard 身份证号脱敏
func maskIDCard(idCard string) string {
	if len(idCard) < 10 {
		return ""
	}
	return idCard[:4] + "********" + idCard[len(idCard)-4:]
}

func toLoveVerificationInfo(v *model.LoveVerification) dto.LoveVerificationInfo {
	return dto.LoveVerificationInfo{
		ID:              v.ID,
		VerifyNo:        v.VerifyNo,
		LoveID:          v.LoveID,
		UserID:          v.UserID,
		Type:            v.Type,
		TypeText:        verifyTypeText(v.Type),
		RealName:        v.RealName,
		IdCardMasked:    maskIDCard(v.IDCardNo),
		IdCardFront:     v.IDCardFront,
		IdCardBack:      v.IDCardBack,
		IdCardHold:      v.IDCardHold,
		FaceImage:       v.FaceImage,
		VideoURL:        v.VideoURL,
		VideoCover:      v.VideoCover,
		VideoDuration:   v.VideoDuration,
		SchoolName:      v.SchoolName,
		DiplomaImage:    v.DiplomaImage,
		DiplomaNo:       v.DiplomaNo,
		Education:       v.Education,
		GraduationYear:  v.GraduationYear,
		PropertyImage:   v.PropertyImage,
		PropertyNo:      v.PropertyNo,
		ThirdPartyScore: v.ThirdPartyScore,
		Status:          v.Status,
		StatusText:      verifyStatusText(v.Status),
		RejectReason:    v.RejectReason,
		VerifiedAt:      v.VerifiedAt,
		VerifiedName:    v.VerifiedName,
		ExpiredAt:       v.ExpiredAt,
		CreatedAt:       v.CreatedAt,
		UpdatedAt:       v.UpdatedAt,
	}
}

func (s *loveVerificationService) Submit(loveID, userID uint, req *dto.CreateLoveVerificationRequest) (*dto.LoveVerificationInfo, error) {
	// 检查是否已存在该类型认证
	if existing, err := s.repo.FindByLoveIDAndType(loveID, req.Type); err == nil && existing != nil {
		if existing.Status == model.VerifyStatusApproved {
			return nil, ErrLoveVerificationExists
		}
		// 待审/拒绝可重新提交，先删除旧记录
		_ = s.repo.Delete(existing.ID)
	}

	verifyNo := fmt.Sprintf("VRF%s%08d", time.Now().Format("20060102150405"), loveID%100000000)
	v := &model.LoveVerification{
		VerifyNo:       verifyNo,
		LoveID:         loveID,
		UserID:         userID,
		Type:           req.Type,
		RealName:       req.RealName,
		IDCardNo:       req.IdCardNo,
		IDCardFront:    req.IdCardFront,
		IDCardBack:     req.IdCardBack,
		IDCardHold:    req.IdCardHold,
		FaceImage:      req.FaceImage,
		VideoURL:       req.VideoURL,
		VideoCover:     req.VideoCover,
		VideoDuration:  req.VideoDuration,
		SchoolName:     req.SchoolName,
		DiplomaImage:   req.DiplomaImage,
		DiplomaNo:      req.DiplomaNo,
		Education:      req.Education,
		GraduationYear: req.GraduationYear,
		PropertyImage:  req.PropertyImage,
		PropertyNo:     req.PropertyNo,
		Status:         model.VerifyStatusPending,
	}
	if err := s.repo.Create(v); err != nil {
		return nil, err
	}
	info := toLoveVerificationInfo(v)
	return &info, nil
}

func (s *loveVerificationService) GetByID(id uint) (*dto.LoveVerificationInfo, error) {
	v, err := s.repo.FindByID(id)
	if err != nil {
		return nil, ErrLoveVerificationNotFound
	}
	info := toLoveVerificationInfo(v)
	return &info, nil
}

func (s *loveVerificationService) GetByLoveID(loveID uint) ([]dto.LoveVerificationInfo, error) {
	list, err := s.repo.ListByLoveID(loveID)
	if err != nil {
		return nil, err
	}
	infos := make([]dto.LoveVerificationInfo, 0, len(list))
	for i := range list {
		infos = append(infos, toLoveVerificationInfo(&list[i]))
	}
	return infos, nil
}

func (s *loveVerificationService) GetByUserID(userID uint) ([]dto.LoveVerificationInfo, error) {
	list, err := s.repo.ListByUserID(userID)
	if err != nil {
		return nil, err
	}
	infos := make([]dto.LoveVerificationInfo, 0, len(list))
	for i := range list {
		infos = append(infos, toLoveVerificationInfo(&list[i]))
	}
	return infos, nil
}

func (s *loveVerificationService) List(req *dto.LoveVerificationListRequest) (*utils.Pagination, []dto.LoveVerificationInfo, error) {
	opts := repository.LoveVerificationListOptions{
		UserID: req.UserID,
		LoveID: req.LoveID,
		Type:   req.Type,
	}
	if req.Status != nil {
		opts.Status = req.Status
	}
	list, total, err := s.repo.List(&req.Pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.LoveVerificationInfo, 0, len(list))
	for i := range list {
		infos = append(infos, toLoveVerificationInfo(&list[i]))
	}
	req.Pagination.Total = total
	return &req.Pagination, infos, nil
}

func (s *loveVerificationService) Audit(id uint, handledBy uint, handledName string, req *dto.LoveVerificationAuditRequest) error {
	v, err := s.repo.FindByID(id)
	if err != nil {
		return ErrLoveVerificationNotFound
	}
	if v.Status != model.VerifyStatusPending {
		return errors.New("该认证已审核")
	}
	return s.repo.UpdateStatus(id, req.Status, req.RejectReason, handledBy, handledName)
}

func (s *loveVerificationService) CountPending() (int64, error) {
	return s.repo.CountPending()
}
