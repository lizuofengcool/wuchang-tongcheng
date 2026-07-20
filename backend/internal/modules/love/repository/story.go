// Package repository love 相亲交友数据访问层 - 动态广场
package repository

import (
	"wuchang-tongcheng/internal/modules/love/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// LoveStoryRepository 动态仓储接口
type LoveStoryRepository interface {
	Create(s *model.LoveStory) error
	FindByID(id uint) (*model.LoveStory, error)
	FindByStoryNo(storyNo string) (*model.LoveStory, error)
	Update(s *model.LoveStory) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(pagination *utils.Pagination, opts LoveStoryListOptions) ([]model.LoveStory, int64, error)
	ListByLoveID(loveID uint, pagination *utils.Pagination) ([]model.LoveStory, int64, error)
	ListByUserID(userID uint, pagination *utils.Pagination) ([]model.LoveStory, int64, error)
	ListFeatured(pagination *utils.Pagination) ([]model.LoveStory, int64, error)
	ListByTopic(topic string, pagination *utils.Pagination) ([]model.LoveStory, int64, error)

	IncrViewCount(id uint) error
	IncrLikeCount(id uint) error
	DecrLikeCount(id uint) error
	IncrCommentCount(id uint) error
	IncrShareCount(id uint) error
	IncrReportCount(id uint) error

	UpdateStatus(id uint, status int) error
	UpdateAuditStatus(id uint, auditStatus int, reason string) error
	SetFeatured(id uint, featured bool) error
	UpdateHotScore(id uint, score float64) error

	BatchUpdateStatus(ids []uint, status int) error
	BatchUpdateAuditStatus(ids []uint, auditStatus int, reason string) error
}

// LoveStoryListOptions 动态列表过滤
type LoveStoryListOptions struct {
	LoveID      uint
	UserID      uint
	MediaType   string
	Topic       string
	Status      *int
	AuditStatus *int
	Featured    *bool
	Keyword     string
	Sort        string
}

type loveStoryRepository struct {
	db *gorm.DB
}

// NewLoveStoryRepository 创建动态仓储
func NewLoveStoryRepository(db *gorm.DB) LoveStoryRepository {
	return &loveStoryRepository{db: db}
}

func (r *loveStoryRepository) Create(s *model.LoveStory) error {
	return r.db.Create(s).Error
}

func (r *loveStoryRepository) FindByID(id uint) (*model.LoveStory, error) {
	var s model.LoveStory
	if err := r.db.First(&s, id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *loveStoryRepository) FindByStoryNo(storyNo string) (*model.LoveStory, error) {
	var s model.LoveStory
	if err := r.db.Where("story_no = ?", storyNo).First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *loveStoryRepository) Update(s *model.LoveStory) error {
	return r.db.Save(s).Error
}

func (r *loveStoryRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.LoveStory{}).Where("id = ?", id).Updates(fields).Error
}

func (r *loveStoryRepository) Delete(id uint) error {
	return r.db.Delete(&model.LoveStory{}, id).Error
}

func (r *loveStoryRepository) List(pagination *utils.Pagination, opts LoveStoryListOptions) ([]model.LoveStory, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LoveStory
	var total int64

	query := r.db.Model(&model.LoveStory{})
	if opts.LoveID > 0 {
		query = query.Where("love_id = ?", opts.LoveID)
	}
	if opts.UserID > 0 {
		query = query.Where("user_id = ?", opts.UserID)
	}
	if opts.MediaType != "" {
		query = query.Where("media_type = ?", opts.MediaType)
	}
	if opts.Topic != "" {
		query = query.Where("topic = ?", opts.Topic)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.AuditStatus != nil {
		query = query.Where("audit_status = ?", *opts.AuditStatus)
	}
	if opts.Featured != nil {
		query = query.Where("featured = ?", *opts.Featured)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("title ILIKE ? OR content ILIKE ?", like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderClause := "published_at DESC, id DESC"
	switch opts.Sort {
	case "hot":
		orderClause = "hot_score DESC, id DESC"
	case "likes":
		orderClause = "like_count DESC, id DESC"
	}
	orderClause = "featured DESC, " + orderClause

	if err := query.Scopes(utils.Paginate(pagination)).Order(orderClause).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *loveStoryRepository) ListByLoveID(loveID uint, pagination *utils.Pagination) ([]model.LoveStory, int64, error) {
	return r.List(pagination, LoveStoryListOptions{LoveID: loveID})
}

func (r *loveStoryRepository) ListByUserID(userID uint, pagination *utils.Pagination) ([]model.LoveStory, int64, error) {
	return r.List(pagination, LoveStoryListOptions{UserID: userID})
}

func (r *loveStoryRepository) ListFeatured(pagination *utils.Pagination) ([]model.LoveStory, int64, error) {
	featured := true
	return r.List(pagination, LoveStoryListOptions{Featured: &featured, Sort: "hot"})
}

func (r *loveStoryRepository) ListByTopic(topic string, pagination *utils.Pagination) ([]model.LoveStory, int64, error) {
	return r.List(pagination, LoveStoryListOptions{Topic: topic})
}

func (r *loveStoryRepository) IncrViewCount(id uint) error {
	return r.db.Model(&model.LoveStory{}).Where("id = ?", id).UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

func (r *loveStoryRepository) IncrLikeCount(id uint) error {
	return r.db.Model(&model.LoveStory{}).Where("id = ?", id).UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error
}

func (r *loveStoryRepository) DecrLikeCount(id uint) error {
	return r.db.Model(&model.LoveStory{}).Where("id = ?", id).UpdateColumn("like_count", gorm.Expr("GREATEST(like_count - 1, 0)")).Error
}

func (r *loveStoryRepository) IncrCommentCount(id uint) error {
	return r.db.Model(&model.LoveStory{}).Where("id = ?", id).UpdateColumn("comment_count", gorm.Expr("comment_count + 1")).Error
}

func (r *loveStoryRepository) IncrShareCount(id uint) error {
	return r.db.Model(&model.LoveStory{}).Where("id = ?", id).UpdateColumn("share_count", gorm.Expr("share_count + 1")).Error
}

func (r *loveStoryRepository) IncrReportCount(id uint) error {
	return r.db.Model(&model.LoveStory{}).Where("id = ?", id).UpdateColumn("report_count", gorm.Expr("report_count + 1")).Error
}

func (r *loveStoryRepository) UpdateStatus(id uint, status int) error {
	return r.db.Model(&model.LoveStory{}).Where("id = ?", id).Update("status", status).Error
}

func (r *loveStoryRepository) UpdateAuditStatus(id uint, auditStatus int, reason string) error {
	return r.db.Model(&model.LoveStory{}).Where("id = ?", id).Updates(map[string]interface{}{
		"audit_status": auditStatus,
		"audit_reason": reason,
	}).Error
}

func (r *loveStoryRepository) SetFeatured(id uint, featured bool) error {
	return r.db.Model(&model.LoveStory{}).Where("id = ?", id).Update("featured", featured).Error
}

func (r *loveStoryRepository) UpdateHotScore(id uint, score float64) error {
	return r.db.Model(&model.LoveStory{}).Where("id = ?", id).Update("hot_score", score).Error
}

func (r *loveStoryRepository) BatchUpdateStatus(ids []uint, status int) error {
	return r.db.Model(&model.LoveStory{}).Where("id IN ?", ids).Update("status", status).Error
}

func (r *loveStoryRepository) BatchUpdateAuditStatus(ids []uint, auditStatus int, reason string) error {
	return r.db.Model(&model.LoveStory{}).Where("id IN ?", ids).Updates(map[string]interface{}{
		"audit_status": auditStatus,
		"audit_reason": reason,
	}).Error
}
