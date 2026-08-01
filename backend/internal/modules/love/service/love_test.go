// Package service 同城相亲交友业务逻辑层单元测试。
// 使用内存 mock 仓储覆盖：资料创建默认值/年龄推导/重复注册拦截、
// 用户隔离更新与全部字段透传、详情浏览量自增、列表/附近/高级搜索分页、
// 位置与语音介绍权限校验、灵魂匹配五维评分算法、
// 管理后台列表/详情/审核/状态/精选/批量操作等核心逻辑，不依赖 DB。
package service

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"wuchang-tongcheng/internal/modules/love/dto"
	"wuchang-tongcheng/internal/modules/love/model"
	"wuchang-tongcheng/internal/modules/love/repository"
	"wuchang-tongcheng/internal/pkg/utils"
)

// ===== mockLoveRepo =====

type mockLoveRepo struct {
	byID     map[uint]*model.Love
	byUser   map[uint]*model.Love
	nextID   uint

	// 可注入错误
	createErr             error
	findErr               error
	findByUserErr         error
	updateErr             error
	updateFieldsErr       error
	deleteErr             error
	listErr               error
	adminListErr          error
	nearbyErr             error
	searchErr             error
	advancedSearchErr     error
	incrViewErr           error
	updateStatusErr       error
	updateAuditStatusErr  error
	updateLocationErr     error
	setFeaturedErr        error
	setPickedErr          error
	batchStatusErr        error
	batchAuditErr         error

	// 调用计数
	incrViewCalls       []uint
	updateFieldsCalls   []struct {
		id     uint
		fields map[string]interface{}
	}
	updateStatusCalls      []struct {
		id     uint
		status int
	}
	updateAuditCalls       []struct {
		id          uint
		auditStatus int
		reason      string
	}
	updateLocationCalls    []struct {
		id      uint
		lat     float64
		lng     float64
	}
	setFeaturedCalls       []struct {
		id       uint
		featured bool
	}
	setPickedCalls         []struct {
		id     uint
		picked bool
	}
	batchStatusCalls       []struct {
		ids    []uint
		status int
	}
	batchAuditCalls        []struct {
		ids         []uint
		auditStatus int
		reason      string
	}
	listCalls              []struct {
		regionID uint
		opts     repository.LoveListOptions
	}
	nearbyCalls            []struct {
		regionID uint
		lat      float64
		lng      float64
		radius   float64
		opts     repository.LoveListOptions
	}
	adminListCalls         []repository.LoveAdminListOptions
}

func newMockLoveRepo() *mockLoveRepo {
	return &mockLoveRepo{
		byID:   make(map[uint]*model.Love),
		byUser: make(map[uint]*model.Love),
		nextID: 1,
	}
}

func (m *mockLoveRepo) Create(l *model.Love) error {
	if m.createErr != nil {
		return m.createErr
	}
	l.ID = m.nextID
	m.nextID++
	cp := *l
	m.byID[l.ID] = &cp
	m.byUser[l.UserID] = &cp
	return nil
}

func (m *mockLoveRepo) FindByID(id uint) (*model.Love, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	l, ok := m.byID[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *l
	return &cp, nil
}

func (m *mockLoveRepo) FindByUserID(userID uint) (*model.Love, error) {
	if m.findByUserErr != nil {
		return nil, m.findByUserErr
	}
	l, ok := m.byUser[userID]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *l
	return &cp, nil
}

func (m *mockLoveRepo) Update(l *model.Love) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	cp := *l
	m.byID[l.ID] = &cp
	m.byUser[l.UserID] = &cp
	return nil
}

func (m *mockLoveRepo) UpdateFields(id uint, fields map[string]interface{}) error {
	if m.updateFieldsErr != nil {
		return m.updateFieldsErr
	}
	m.updateFieldsCalls = append(m.updateFieldsCalls, struct {
		id     uint
		fields map[string]interface{}
	}{id, fields})
	l, ok := m.byID[id]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	if v, ok := fields["voice_intro_url"]; ok {
		if s, ok := v.(string); ok {
			l.VoiceIntroURL = s
		}
	}
	return nil
}

func (m *mockLoveRepo) Delete(id uint) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if l, ok := m.byID[id]; ok {
		delete(m.byID, id)
		delete(m.byUser, l.UserID)
	}
	return nil
}

func (m *mockLoveRepo) List(regionID uint, pagination *utils.Pagination, opts repository.LoveListOptions) ([]model.Love, int64, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	m.listCalls = append(m.listCalls, struct {
		regionID uint
		opts     repository.LoveListOptions
	}{regionID, opts})
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Love
	for _, l := range m.byID {
		if regionID > 0 && l.RegionID != regionID {
			continue
		}
		if opts.Status == -1 {
			// 全部
		} else if opts.Status > 0 {
			if l.Status != opts.Status {
				continue
			}
		} else {
			if l.Status != model.LoveStatusActive {
				continue
			}
		}
		if opts.Gender != nil && l.Gender != *opts.Gender {
			continue
		}
		list = append(list, *l)
	}
	return list, int64(len(list)), nil
}

func (m *mockLoveRepo) AdminList(pagination *utils.Pagination, opts repository.LoveAdminListOptions) ([]model.Love, int64, error) {
	if m.adminListErr != nil {
		return nil, 0, m.adminListErr
	}
	m.adminListCalls = append(m.adminListCalls, opts)
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Love
	for _, l := range m.byID {
		if opts.RegionID > 0 && l.RegionID != opts.RegionID {
			continue
		}
		if opts.UserID > 0 && l.UserID != opts.UserID {
			continue
		}
		if opts.Status != nil && l.Status != *opts.Status {
			continue
		}
		if opts.AuditStatus != nil && l.AuditStatus != *opts.AuditStatus {
			continue
		}
		list = append(list, *l)
	}
	return list, int64(len(list)), nil
}

func (m *mockLoveRepo) ListNearby(regionID uint, pagination *utils.Pagination, lat, lng, radiusKm float64, opts repository.LoveListOptions) ([]model.Love, int64, error) {
	if m.nearbyErr != nil {
		return nil, 0, m.nearbyErr
	}
	m.nearbyCalls = append(m.nearbyCalls, struct {
		regionID uint
		lat      float64
		lng      float64
		radius   float64
		opts     repository.LoveListOptions
	}{regionID, lat, lng, radiusKm, opts})
	var list []model.Love
	for _, l := range m.byID {
		if l.Status != model.LoveStatusActive {
			continue
		}
		if regionID > 0 && l.RegionID != regionID {
			continue
		}
		list = append(list, *l)
	}
	return list, int64(len(list)), nil
}

func (m *mockLoveRepo) Search(regionID uint, pagination *utils.Pagination, keyword string) ([]model.Love, int64, error) {
	if m.searchErr != nil {
		return nil, 0, m.searchErr
	}
	return m.List(regionID, pagination, repository.LoveListOptions{Keyword: keyword, Status: model.LoveStatusActive})
}

func (m *mockLoveRepo) AdvancedSearch(regionID uint, pagination *utils.Pagination, opts repository.LoveListOptions) ([]model.Love, int64, error) {
	if m.advancedSearchErr != nil {
		return nil, 0, m.advancedSearchErr
	}
	return m.List(regionID, pagination, opts)
}

func (m *mockLoveRepo) IncrViewCount(id uint) error {
	m.incrViewCalls = append(m.incrViewCalls, id)
	if m.incrViewErr != nil {
		return m.incrViewErr
	}
	if l, ok := m.byID[id]; ok {
		l.ViewCount++
	}
	return nil
}

func (m *mockLoveRepo) IncrLikeCount(id uint) error {
	if l, ok := m.byID[id]; ok {
		l.LikeCount++
	}
	return nil
}

func (m *mockLoveRepo) DecrLikeCount(id uint) error {
	if l, ok := m.byID[id]; ok && l.LikeCount > 0 {
		l.LikeCount--
	}
	return nil
}

func (m *mockLoveRepo) IncrLikedCount(id uint) error {
	if l, ok := m.byID[id]; ok {
		l.LikedCount++
	}
	return nil
}

func (m *mockLoveRepo) DecrLikedCount(id uint) error {
	if l, ok := m.byID[id]; ok && l.LikedCount > 0 {
		l.LikedCount--
	}
	return nil
}

func (m *mockLoveRepo) IncrMatchCount(id uint) error {
	if l, ok := m.byID[id]; ok {
		l.MatchCount++
	}
	return nil
}

func (m *mockLoveRepo) DecrMatchCount(id uint) error {
	if l, ok := m.byID[id]; ok && l.MatchCount > 0 {
		l.MatchCount--
	}
	return nil
}

func (m *mockLoveRepo) IncrVisitorCount(id uint) error {
	if l, ok := m.byID[id]; ok {
		l.VisitorCount++
	}
	return nil
}

func (m *mockLoveRepo) IncrStoryCount(id uint) error {
	if l, ok := m.byID[id]; ok {
		l.StoryCount++
	}
	return nil
}

func (m *mockLoveRepo) DecrStoryCount(id uint) error {
	if l, ok := m.byID[id]; ok && l.StoryCount > 0 {
		l.StoryCount--
	}
	return nil
}

func (m *mockLoveRepo) IncrGiftCount(id uint) error {
	if l, ok := m.byID[id]; ok {
		l.GiftCount++
	}
	return nil
}

func (m *mockLoveRepo) IncrImpressionCount(id uint) error {
	if l, ok := m.byID[id]; ok {
		l.ImpressionCount++
	}
	return nil
}

func (m *mockLoveRepo) IncrPopularityScore(id uint, score float64) error {
	if l, ok := m.byID[id]; ok {
		l.PopularityScore += score
	}
	return nil
}

func (m *mockLoveRepo) UpdateStatus(id uint, status int) error {
	if m.updateStatusErr != nil {
		return m.updateStatusErr
	}
	m.updateStatusCalls = append(m.updateStatusCalls, struct {
		id     uint
		status int
	}{id, status})
	if l, ok := m.byID[id]; ok {
		l.Status = status
	}
	return nil
}

func (m *mockLoveRepo) UpdateAuditStatus(id uint, auditStatus int, reason string) error {
	if m.updateAuditStatusErr != nil {
		return m.updateAuditStatusErr
	}
	m.updateAuditCalls = append(m.updateAuditCalls, struct {
		id          uint
		auditStatus int
		reason      string
	}{id, auditStatus, reason})
	if l, ok := m.byID[id]; ok {
		l.AuditStatus = auditStatus
		l.AuditReason = reason
	}
	return nil
}

func (m *mockLoveRepo) UpdateLocation(id uint, lat, lng float64) error {
	if m.updateLocationErr != nil {
		return m.updateLocationErr
	}
	m.updateLocationCalls = append(m.updateLocationCalls, struct {
		id  uint
		lat float64
		lng float64
	}{id, lat, lng})
	if l, ok := m.byID[id]; ok {
		l.Latitude = lat
		l.Longitude = lng
	}
	return nil
}

func (m *mockLoveRepo) UpdateLastActive(id uint, ip string) error {
	if l, ok := m.byID[id]; ok {
		now := time.Now()
		l.LastActiveAt = &now
		l.LastActiveIP = ip
	}
	return nil
}

func (m *mockLoveRepo) UpdateMemberLevel(id uint, level int, expiredAt interface{}) error {
	if l, ok := m.byID[id]; ok {
		l.MemberLevel = level
	}
	return nil
}

func (m *mockLoveRepo) UpdateCredits(id uint, credits float64) error {
	if l, ok := m.byID[id]; ok {
		l.Credits = credits
	}
	return nil
}

func (m *mockLoveRepo) IncrementCredits(id uint, delta float64) error {
	if l, ok := m.byID[id]; ok {
		l.Credits += delta
	}
	return nil
}

func (m *mockLoveRepo) SetFeatured(id uint, featured bool) error {
	if m.setFeaturedErr != nil {
		return m.setFeaturedErr
	}
	m.setFeaturedCalls = append(m.setFeaturedCalls, struct {
		id       uint
		featured bool
	}{id, featured})
	if l, ok := m.byID[id]; ok {
		l.Featured = featured
	}
	return nil
}

func (m *mockLoveRepo) SetPicked(id uint, picked bool) error {
	if m.setPickedErr != nil {
		return m.setPickedErr
	}
	m.setPickedCalls = append(m.setPickedCalls, struct {
		id     uint
		picked bool
	}{id, picked})
	if l, ok := m.byID[id]; ok {
		l.Picked = picked
	}
	return nil
}

func (m *mockLoveRepo) UpdateRiskScore(id uint, score int) error {
	if l, ok := m.byID[id]; ok {
		l.RiskScore = score
	}
	return nil
}

func (m *mockLoveRepo) BatchUpdateStatus(ids []uint, status int) error {
	if m.batchStatusErr != nil {
		return m.batchStatusErr
	}
	m.batchStatusCalls = append(m.batchStatusCalls, struct {
		ids    []uint
		status int
	}{ids, status})
	for _, id := range ids {
		if l, ok := m.byID[id]; ok {
			l.Status = status
		}
	}
	return nil
}

func (m *mockLoveRepo) BatchUpdateAuditStatus(ids []uint, auditStatus int, reason string) error {
	if m.batchAuditErr != nil {
		return m.batchAuditErr
	}
	m.batchAuditCalls = append(m.batchAuditCalls, struct {
		ids         []uint
		auditStatus int
		reason      string
	}{ids, auditStatus, reason})
	for _, id := range ids {
		if l, ok := m.byID[id]; ok {
			l.AuditStatus = auditStatus
			l.AuditReason = reason
		}
	}
	return nil
}

// ===== 测试辅助 =====

func newTestLoveService() (*loveService, *mockLoveRepo) {
	repo := newMockLoveRepo()
	svc := NewLoveService(repo).(*loveService)
	return svc, repo
}

func newLove(id uint, l model.Love) *model.Love {
	l.ID = id
	return &l
}

func strPtr(s string) *string { return &s }
func intPtr(i int) *int       { return &i }
func boolPtr(b bool) *bool    { return &b }

// ===== 文本工具函数测试 =====

func TestLoveStatusText(t *testing.T) {
	assert.Equal(t, "禁用", loveStatusText(model.LoveStatusDisabled))
	assert.Equal(t, "正常", loveStatusText(model.LoveStatusActive))
	assert.Equal(t, "冻结", loveStatusText(model.LoveStatusFrozen))
	assert.Equal(t, "注销", loveStatusText(model.LoveStatusCanceled))
	assert.Equal(t, "", loveStatusText(999))
}

func TestLoveAuditStatusText(t *testing.T) {
	assert.Equal(t, "待审", loveAuditStatusText(model.LoveAuditPending))
	assert.Equal(t, "通过", loveAuditStatusText(model.LoveAuditApproved))
	assert.Equal(t, "拒绝", loveAuditStatusText(model.LoveAuditRejected))
	assert.Equal(t, "", loveAuditStatusText(999))
}

func TestLoveGenderText(t *testing.T) {
	assert.Equal(t, "男", loveGenderText(model.GenderMale))
	assert.Equal(t, "女", loveGenderText(model.GenderFemale))
	assert.Equal(t, "未知", loveGenderText(model.GenderUnknown))
	assert.Equal(t, "未知", loveGenderText(999))
}

func TestLoveMemberLevelText(t *testing.T) {
	assert.Equal(t, "普通", loveMemberLevelText(model.MemberLevelNone))
	assert.Equal(t, "基础会员", loveMemberLevelText(model.MemberLevelBasic))
	assert.Equal(t, "高级会员", loveMemberLevelText(model.MemberLevelAdvanced))
	assert.Equal(t, "VIP会员", loveMemberLevelText(model.MemberLevelVIP))
	assert.Equal(t, "Premium会员", loveMemberLevelText(model.MemberLevelPremium))
	assert.Equal(t, "", loveMemberLevelText(999))
}

func TestIsOnline(t *testing.T) {
	assert.False(t, isOnline(nil))
	now := time.Now()
	assert.True(t, isOnline(&now))
	old := time.Now().Add(-10 * time.Minute)
	assert.False(t, isOnline(&old))
}

// ===== toLoveInfo 测试 =====

func TestToLoveInfo(t *testing.T) {
	t.Run("默认值兜底", func(t *testing.T) {
		l := newLove(1, model.Love{UserID: 10, Nickname: "Alice"})
		info := toLoveInfo(l)
		assert.Equal(t, uint(1), info.ID)
		assert.Equal(t, uint(10), info.UserID)
		assert.Equal(t, "Alice", info.Nickname)
		assert.Equal(t, "未知", info.GenderText) // Gender=0 → 未知
		assert.Equal(t, "禁用", info.StatusText) // Status=0 → 禁用
		assert.False(t, info.Online)             // LastActiveAt=nil
	})

	t.Run("全部状态文本映射", func(t *testing.T) {
		l := newLove(1, model.Love{
			Gender:        model.GenderFemale,
			Status:        model.LoveStatusFrozen,
			AuditStatus:   model.LoveAuditRejected,
			MemberLevel:   model.MemberLevelVIP,
			AuditReason:   "违规",
		})
		info := toLoveInfo(l)
		assert.Equal(t, "女", info.GenderText)
		assert.Equal(t, "冻结", info.StatusText)
		assert.Equal(t, "拒绝", info.AuditStatusText)
		assert.Equal(t, "VIP会员", info.MemberLevelText)
		assert.Equal(t, "违规", info.AuditReason)
	})

	t.Run("在线状态判定", func(t *testing.T) {
		now := time.Now()
		l := newLove(1, model.Love{LastActiveAt: &now})
		info := toLoveInfo(l)
		assert.True(t, info.Online)
	})
}

// ===== Create 创建资料测试 =====

func TestLoveCreate(t *testing.T) {
	t.Run("正常创建含默认值", func(t *testing.T) {
		svc, repo := newTestLoveService()
		req := &dto.CreateLoveRequest{
			Nickname:  "Alice",
			Avatar:    "avatar.jpg",
			Gender:    model.GenderFemale,
			Hometown:  "武汉",
			Residence: "武昌",
		}
		info, err := svc.Create(2, 10, req)
		require.NoError(t, err)
		assert.Equal(t, uint(1), info.ID)
		assert.Equal(t, uint(10), info.UserID)
		assert.Equal(t, "Alice", info.Nickname)
		assert.Equal(t, "avatar.jpg", info.Avatar)
		assert.Equal(t, model.GenderFemale, info.Gender)
		assert.Equal(t, model.LoveStatusActive, info.Status)
		assert.Equal(t, model.LoveAuditApproved, info.AuditStatus) // MVP 发布即通过
		assert.Equal(t, uint(2), info.RegionID)
		// 仓储写入
		l := repo.byID[1]
		assert.Equal(t, uint(2), l.RegionID)
		assert.Equal(t, model.LoveStatusActive, l.Status)
	})

	t.Run("根据生日计算年龄", func(t *testing.T) {
		svc, repo := newTestLoveService()
		birthday := time.Date(1995, 6, 15, 0, 0, 0, 0, time.UTC)
		req := &dto.CreateLoveRequest{Nickname: "Bob", Birthday: &birthday}
		info, err := svc.Create(1, 1, req)
		require.NoError(t, err)
		expectedAge := time.Now().Year() - 1995
		assert.Equal(t, expectedAge, info.Age)
		assert.Equal(t, expectedAge, repo.byID[info.ID].Age)
	})

	t.Run("已存在资料拒绝创建", func(t *testing.T) {
		svc, _ := newTestLoveService()
		req := &dto.CreateLoveRequest{Nickname: "A"}
		_, err := svc.Create(1, 10, req)
		require.NoError(t, err)
		_, err = svc.Create(1, 10, req)
		assert.ErrorIs(t, err, ErrLoveExists)
	})

	t.Run("查重时返回非 NotFound 错误透传", func(t *testing.T) {
		svc, repo := newTestLoveService()
		repo.findByUserErr = errors.New("db connection lost")
		_, err := svc.Create(1, 10, &dto.CreateLoveRequest{Nickname: "A"})
		assert.Equal(t, "db connection lost", err.Error())
	})

	t.Run("仓储创建失败", func(t *testing.T) {
		svc, repo := newTestLoveService()
		repo.createErr = errors.New("db error")
		_, err := svc.Create(1, 10, &dto.CreateLoveRequest{Nickname: "A"})
		assert.Error(t, err)
	})
}

// ===== Update 更新测试 =====

func TestLoveUpdate(t *testing.T) {
	t.Run("资料不存在", func(t *testing.T) {
		svc, _ := newTestLoveService()
		err := svc.Update(999, 10, &dto.UpdateLoveRequest{Nickname: strPtr("x")})
		assert.ErrorIs(t, err, ErrLoveNotFound)
	})

	t.Run("无权操作他人资料", func(t *testing.T) {
		svc, repo := newTestLoveService()
		repo.byID[1] = newLove(1, model.Love{UserID: 99})
		err := svc.Update(1, 10, &dto.UpdateLoveRequest{Nickname: strPtr("x")})
		assert.ErrorIs(t, err, ErrLoveNoPermission)
	})

	t.Run("全部字段透传", func(t *testing.T) {
		svc, repo := newTestLoveService()
		repo.byID[1] = newLove(1, model.Love{UserID: 10})
		birthday := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
		err := svc.Update(1, 10, &dto.UpdateLoveRequest{
			Nickname:      strPtr("NewName"),
			Avatar:        strPtr("new.jpg"),
			Gender:        intPtr(model.GenderMale),
			Birthday:      &birthday,
			Height:        intPtr(180),
			Weight:        intPtr(75),
			Constellation: strPtr("白羊座"),
			Zodiac:        strPtr("属鼠"),
			Hometown:      strPtr("北京"),
			Residence:     strPtr("上海"),
			Education:     strPtr("本科"),
			Occupation:    strPtr("工程师"),
			Income:        strPtr("20-30万"),
			Marriage:      strPtr("未婚"),
			House:         strPtr("有房"),
			Car:           strPtr("有车"),
			Drinking:      strPtr("不喝"),
			Smoking:       strPtr("不抽"),
			WantKids:      strPtr("想要"),
			Bio:           strPtr("个人简介"),
			VoiceIntroURL: strPtr("voice.mp3"),
			CoverImage:    strPtr("cover.jpg"),
		})
		require.NoError(t, err)
		l := repo.byID[1]
		assert.Equal(t, "NewName", l.Nickname)
		assert.Equal(t, "new.jpg", l.Avatar)
		assert.Equal(t, model.GenderMale, l.Gender)
		assert.Equal(t, 180, l.Height)
		assert.Equal(t, 75, l.Weight)
		assert.Equal(t, "白羊座", l.Constellation)
		assert.Equal(t, "属鼠", l.Zodiac)
		assert.Equal(t, "北京", l.Hometown)
		assert.Equal(t, "上海", l.Residence)
		assert.Equal(t, "本科", l.Education)
		assert.Equal(t, "工程师", l.Occupation)
		assert.Equal(t, "20-30万", l.Income)
		assert.Equal(t, "未婚", l.Marriage)
		assert.Equal(t, "有房", l.House)
		assert.Equal(t, "有车", l.Car)
		assert.Equal(t, "不喝", l.Drinking)
		assert.Equal(t, "不抽", l.Smoking)
		assert.Equal(t, "想要", l.WantKids)
		assert.Equal(t, "个人简介", l.Bio)
		assert.Equal(t, "voice.mp3", l.VoiceIntroURL)
		assert.Equal(t, "cover.jpg", l.CoverImage)
		// 生日更新会重算年龄
		assert.Equal(t, time.Now().Year()-1990, l.Age)
	})

	t.Run("仅更新部分字段不影响其他字段", func(t *testing.T) {
		svc, repo := newTestLoveService()
		repo.byID[1] = newLove(1, model.Love{UserID: 10, Nickname: "Old", Bio: "OldBio"})
		err := svc.Update(1, 10, &dto.UpdateLoveRequest{Bio: strPtr("NewBio")})
		require.NoError(t, err)
		l := repo.byID[1]
		assert.Equal(t, "Old", l.Nickname)        // 未变
		assert.Equal(t, "NewBio", l.Bio)          // 已变
	})

	t.Run("仓储 Update 失败", func(t *testing.T) {
		svc, repo := newTestLoveService()
		repo.byID[1] = newLove(1, model.Love{UserID: 10})
		repo.updateErr = errors.New("db error")
		err := svc.Update(1, 10, &dto.UpdateLoveRequest{Nickname: strPtr("x")})
		assert.Error(t, err)
	})
}

// ===== GetByID 详情测试 =====

func TestLoveGetByID(t *testing.T) {
	t.Run("资料不存在", func(t *testing.T) {
		svc, _ := newTestLoveService()
		_, err := svc.GetByID(999, 10)
		assert.ErrorIs(t, err, ErrLoveNotFound)
	})

	t.Run("正常获取并自增浏览数", func(t *testing.T) {
		svc, repo := newTestLoveService()
		repo.byID[1] = newLove(1, model.Love{UserID: 10, Nickname: "A", ViewCount: 5})
		info, err := svc.GetByID(1, 20)
		require.NoError(t, err)
		assert.Equal(t, "A", info.Nickname)
		assert.Equal(t, []uint{1}, repo.incrViewCalls)
		assert.Equal(t, 6, repo.byID[1].ViewCount)
	})

	t.Run("浏览数自增失败不影响详情返回", func(t *testing.T) {
		svc, repo := newTestLoveService()
		repo.byID[1] = newLove(1, model.Love{UserID: 10})
		repo.incrViewErr = errors.New("redis busy")
		info, err := svc.GetByID(1, 20)
		require.NoError(t, err)
		assert.NotNil(t, info)
	})
}

// ===== GetByUserID 测试 =====

func TestLoveGetByUserID(t *testing.T) {
	t.Run("资料不存在", func(t *testing.T) {
		svc, _ := newTestLoveService()
		_, err := svc.GetByUserID(999)
		assert.ErrorIs(t, err, ErrLoveNotFound)
	})

	t.Run("正常获取", func(t *testing.T) {
		svc, _ := newTestLoveService()
		svc.repo.(*mockLoveRepo).byID[1] = newLove(1, model.Love{UserID: 10, Nickname: "A"})
		svc.repo.(*mockLoveRepo).byUser[10] = svc.repo.(*mockLoveRepo).byID[1]
		info, err := svc.GetByUserID(10)
		require.NoError(t, err)
		assert.Equal(t, "A", info.Nickname)
	})
}

// ===== List 列表测试 =====

func TestLoveList(t *testing.T) {
	t.Run("正常分页列表", func(t *testing.T) {
		svc, repo := newTestLoveService()
		repo.byID[1] = newLove(1, model.Love{UserID: 10, Status: model.LoveStatusActive, Nickname: "A"})
		repo.byID[1].RegionID = 2
		repo.byID[2] = newLove(2, model.Love{UserID: 11, Status: model.LoveStatusActive, Nickname: "B"})
		repo.byID[2].RegionID = 2
		repo.byID[3] = newLove(3, model.Love{UserID: 12, Status: model.LoveStatusActive, Nickname: "C"})
		repo.byID[3].RegionID = 3
		req := &dto.LoveListRequest{Pagination: *utils.NewPagination(1, 10)}
		pg, list, err := svc.List(2, req)
		require.NoError(t, err)
		assert.Equal(t, int64(2), pg.Total)
		assert.Len(t, list, 2)
		// 默认状态过滤为 active
		assert.Equal(t, model.LoveStatusActive, repo.listCalls[0].opts.Status)
	})

	t.Run("C端列表强制过滤为 active-冻结资料被排除", func(t *testing.T) {
		svc, repo := newTestLoveService()
		repo.byID[1] = newLove(1, model.Love{UserID: 10, Status: model.LoveStatusActive})
		repo.byID[1].RegionID = 1
		repo.byID[2] = newLove(2, model.Love{UserID: 11, Status: model.LoveStatusFrozen})
		repo.byID[2].RegionID = 1
		req := &dto.LoveListRequest{
			Pagination: *utils.NewPagination(1, 10),
			Status:     intPtr(-1), // service 仍会强制覆盖为 active
		}
		_, list, err := svc.List(1, req)
		require.NoError(t, err)
		assert.Len(t, list, 1) // 仅 active 返回
		// service 硬编码 Status=active，忽略 req.Status
		assert.Equal(t, model.LoveStatusActive, repo.listCalls[0].opts.Status)
	})

	t.Run("仓储列表错误", func(t *testing.T) {
		svc, repo := newTestLoveService()
		repo.listErr = errors.New("db error")
		req := &dto.LoveListRequest{Pagination: *utils.NewPagination(1, 10)}
		_, _, err := svc.List(1, req)
		assert.Error(t, err)
	})
}

// ===== ListNearby 附近测试 =====

func TestLoveListNearby(t *testing.T) {
	t.Run("正常返回附近用户", func(t *testing.T) {
		svc, repo := newTestLoveService()
		repo.byID[1] = newLove(1, model.Love{UserID: 10, Status: model.LoveStatusActive})
		repo.byID[1].RegionID = 1
		repo.byID[2] = newLove(2, model.Love{UserID: 11, Status: model.LoveStatusFrozen}) // 被过滤
		repo.byID[2].RegionID = 1
		req := &dto.LoveNearbyRequest{
			Latitude:  30.5,
			Longitude: 114.3,
			RadiusKm:  5,
			Pagination: *utils.NewPagination(1, 10),
		}
		pg, list, err := svc.ListNearby(1, req)
		require.NoError(t, err)
		assert.Equal(t, int64(1), pg.Total)
		assert.Len(t, list, 1)
		require.Len(t, repo.nearbyCalls, 1)
		assert.Equal(t, 30.5, repo.nearbyCalls[0].lat)
		assert.Equal(t, 114.3, repo.nearbyCalls[0].lng)
		assert.Equal(t, 5.0, repo.nearbyCalls[0].radius)
		// 附近强制按 active 排序
		assert.Equal(t, "active", repo.nearbyCalls[0].opts.Sort)
		assert.Equal(t, model.LoveStatusActive, repo.nearbyCalls[0].opts.Status)
	})

	t.Run("仓储附近错误", func(t *testing.T) {
		svc, repo := newTestLoveService()
		repo.nearbyErr = errors.New("postgis unavailable")
		req := &dto.LoveNearbyRequest{Pagination: *utils.NewPagination(1, 10)}
		_, _, err := svc.ListNearby(1, req)
		assert.Error(t, err)
	})
}

// ===== Search / AdvancedSearch 测试 =====

func TestLoveSearch(t *testing.T) {
	t.Run("Search 委托 AdvancedSearch", func(t *testing.T) {
		svc, repo := newTestLoveService()
		repo.byID[1] = newLove(1, model.Love{UserID: 10, Status: model.LoveStatusActive, Nickname: "Alice"})
		repo.byID[1].RegionID = 1
		req := &dto.LoveAdvancedSearchRequest{
			Keyword: "Ali",
			Pagination: *utils.NewPagination(1, 10),
		}
		pg1, list1, err1 := svc.Search(1, req)
		require.NoError(t, err1)
		pg2, list2, err2 := svc.AdvancedSearch(1, req)
		require.NoError(t, err2)
		assert.Equal(t, pg1.Total, pg2.Total)
		assert.Len(t, list1, len(list2))
	})

	t.Run("AdvancedSearch 高级过滤", func(t *testing.T) {
		svc, _ := newTestLoveService()
		svc.repo.(*mockLoveRepo).byID[1] = newLove(1, model.Love{UserID: 10, Status: model.LoveStatusActive})
		svc.repo.(*mockLoveRepo).byID[1].RegionID = 1
		req := &dto.LoveAdvancedSearchRequest{
			Keyword: "test",
			Pagination: *utils.NewPagination(1, 20),
		}
		pg, list, err := svc.AdvancedSearch(1, req)
		require.NoError(t, err)
		assert.Equal(t, int64(1), pg.Total)
		assert.Len(t, list, 1)
	})

	t.Run("AdvancedSearch 仓储错误", func(t *testing.T) {
		svc, repo := newTestLoveService()
		repo.listErr = errors.New("db error")
		req := &dto.LoveAdvancedSearchRequest{Pagination: *utils.NewPagination(1, 10)}
		_, _, err := svc.AdvancedSearch(1, req)
		assert.Error(t, err)
	})
}

// ===== UpdateLocation 测试 =====

func TestLoveUpdateLocation(t *testing.T) {
	t.Run("资料不存在", func(t *testing.T) {
		svc, _ := newTestLoveService()
		err := svc.UpdateLocation(999, 10, &dto.UpdateLocationRequest{Latitude: 30, Longitude: 114})
		assert.ErrorIs(t, err, ErrLoveNotFound)
	})

	t.Run("无权操作", func(t *testing.T) {
		svc, repo := newTestLoveService()
		repo.byID[1] = newLove(1, model.Love{UserID: 99})
		err := svc.UpdateLocation(1, 10, &dto.UpdateLocationRequest{Latitude: 30, Longitude: 114})
		assert.ErrorIs(t, err, ErrLoveNoPermission)
	})

	t.Run("正常更新位置", func(t *testing.T) {
		svc, repo := newTestLoveService()
		repo.byID[1] = newLove(1, model.Love{UserID: 10})
		err := svc.UpdateLocation(1, 10, &dto.UpdateLocationRequest{Latitude: 30.5, Longitude: 114.3})
		require.NoError(t, err)
		require.Len(t, repo.updateLocationCalls, 1)
		assert.Equal(t, 30.5, repo.updateLocationCalls[0].lat)
		assert.Equal(t, 114.3, repo.updateLocationCalls[0].lng)
	})

	t.Run("仓储 UpdateLocation 失败", func(t *testing.T) {
		svc, repo := newTestLoveService()
		repo.byID[1] = newLove(1, model.Love{UserID: 10})
		repo.updateLocationErr = errors.New("db error")
		err := svc.UpdateLocation(1, 10, &dto.UpdateLocationRequest{Latitude: 30, Longitude: 114})
		assert.Error(t, err)
	})
}

// ===== UpdateVoiceIntro 测试 =====

func TestLoveUpdateVoiceIntro(t *testing.T) {
	t.Run("资料不存在", func(t *testing.T) {
		svc, _ := newTestLoveService()
		err := svc.UpdateVoiceIntro(999, 10, &dto.UpdateVoiceIntroRequest{VoiceIntroURL: "v.mp3"})
		assert.ErrorIs(t, err, ErrLoveNotFound)
	})

	t.Run("无权操作", func(t *testing.T) {
		svc, repo := newTestLoveService()
		repo.byID[1] = newLove(1, model.Love{UserID: 99})
		err := svc.UpdateVoiceIntro(1, 10, &dto.UpdateVoiceIntroRequest{VoiceIntroURL: "v.mp3"})
		assert.ErrorIs(t, err, ErrLoveNoPermission)
	})

	t.Run("正常更新语音介绍", func(t *testing.T) {
		svc, repo := newTestLoveService()
		repo.byID[1] = newLove(1, model.Love{UserID: 10})
		err := svc.UpdateVoiceIntro(1, 10, &dto.UpdateVoiceIntroRequest{VoiceIntroURL: "voice.mp3"})
		require.NoError(t, err)
		require.Len(t, repo.updateFieldsCalls, 1)
		assert.Equal(t, uint(1), repo.updateFieldsCalls[0].id)
		assert.Equal(t, "voice.mp3", repo.updateFieldsCalls[0].fields["voice_intro_url"])
	})

	t.Run("仓储 UpdateFields 失败", func(t *testing.T) {
		svc, repo := newTestLoveService()
		repo.byID[1] = newLove(1, model.Love{UserID: 10})
		repo.updateFieldsErr = errors.New("db error")
		err := svc.UpdateVoiceIntro(1, 10, &dto.UpdateVoiceIntroRequest{VoiceIntroURL: "v.mp3"})
		assert.Error(t, err)
	})
}

// ===== MatchScore 灵魂匹配评分测试 =====

func TestLoveMatchScore(t *testing.T) {
	t.Run("用户A资料不存在", func(t *testing.T) {
		svc, _ := newTestLoveService()
		_, err := svc.MatchScore(999, 20)
		assert.ErrorIs(t, err, ErrLoveNotFound)
	})

	t.Run("用户B资料不存在", func(t *testing.T) {
		svc, repo := newTestLoveService()
		repo.byUser[10] = newLove(1, model.Love{UserID: 10})
		_, err := svc.MatchScore(10, 999)
		assert.ErrorIs(t, err, ErrLoveNotFound)
	})

	t.Run("五维评分全匹配-灵魂伴侣", func(t *testing.T) {
		svc, repo := newTestLoveService()
		interests, _ := model.FromJSON([]string{"读书", "电影"})
		personality, _ := model.FromJSON([]string{"INTJ"})
		values, _ := model.FromJSON([]string{"独立"})
		repo.byID[1] = newLove(1, model.Love{UserID: 10, Age: 25, Latitude: 30.5, Longitude: 114.3, Interests: interests, Personality: personality, Values: values})
		repo.byID[2] = newLove(2, model.Love{UserID: 11, Age: 25, Latitude: 30.5, Longitude: 114.3, Interests: interests, Personality: personality, Values: values})
		repo.byUser[10] = repo.byID[1]
		repo.byUser[11] = repo.byID[2]
		resp, err := svc.MatchScore(10, 11)
		require.NoError(t, err)
		assert.InDelta(t, 100.0, resp.TotalScore, 0.01)
		assert.Equal(t, "灵魂伴侣般匹配", resp.Reason)
	})

	t.Run("评分区间对应不同 reason", func(t *testing.T) {
		// 60-79 高度契合
		svc, repo := newTestLoveService()
		// 兴趣/性格/价值观相同（100），位置和年龄差距大拉低总分
		interests, _ := model.FromJSON([]string{"x"})
		repo.byID[1] = newLove(1, model.Love{UserID: 10, Age: 25, Latitude: 30, Longitude: 114, Interests: interests, Personality: interests, Values: interests})
		repo.byID[2] = newLove(2, model.Love{UserID: 11, Age: 40, Latitude: 35, Longitude: 120, Interests: interests, Personality: interests, Values: interests})
		repo.byUser[10] = repo.byID[1]
		repo.byUser[11] = repo.byID[2]
		resp, err := svc.MatchScore(10, 11)
		require.NoError(t, err)
		// interest 100 + personality 100 + value 100 + location (35-30,120-114) ~625 km² → 30 + age diff 15 → 20
		// total = 100*0.25 + 100*0.25 + 100*0.2 + 30*0.15 + 20*0.15 = 25+25+20+4.5+3 = 77.5 → 高度契合
		assert.InDelta(t, 77.5, resp.TotalScore, 1.0)
		assert.Equal(t, "高度契合", resp.Reason)
	})

	t.Run("评分 40-59 较为匹配", func(t *testing.T) {
		svc, repo := newTestLoveService()
		// JSONB 空 → 30/30/30，位置相同 → 100，同龄 → 100
		// total = 30*0.25+30*0.25+30*0.2+100*0.15+100*0.15 = 7.5+7.5+6+15+15 = 51 → 较为匹配
		repo.byID[1] = newLove(1, model.Love{UserID: 10, Age: 25, Latitude: 30.5, Longitude: 114.3})
		repo.byID[2] = newLove(2, model.Love{UserID: 11, Age: 25, Latitude: 30.5, Longitude: 114.3})
		repo.byUser[10] = repo.byID[1]
		repo.byUser[11] = repo.byID[2]
		resp, err := svc.MatchScore(10, 11)
		require.NoError(t, err)
		assert.InDelta(t, 51.0, resp.TotalScore, 0.5)
		assert.Equal(t, "较为匹配", resp.Reason)
	})

	t.Run("评分低于 40 兴趣相投", func(t *testing.T) {
		svc, repo := newTestLoveService()
		// 全部 JSONB 空 → 30/30/30，location 远 → 10，age diff 25 → 20
		// total = 30*0.25+30*0.25+30*0.2+10*0.15+20*0.15 = 7.5+7.5+6+1.5+3 = 25.5 → 兴趣相投
		repo.byID[1] = newLove(1, model.Love{UserID: 10, Age: 25, Latitude: 30, Longitude: 114})
		repo.byID[2] = newLove(2, model.Love{UserID: 11, Age: 50, Latitude: 40, Longitude: 125})
		repo.byUser[10] = repo.byID[1]
		repo.byUser[11] = repo.byID[2]
		resp, err := svc.MatchScore(10, 11)
		require.NoError(t, err)
		assert.LessOrEqual(t, resp.TotalScore, 39.99)
		assert.Equal(t, "兴趣相投，性格契合", resp.Reason)
	})
}

// ===== 评分辅助函数测试 =====

func TestCalcJSONBSimilarity(t *testing.T) {
	// 一方为空
	assert.Equal(t, 30.0, calcJSONBSimilarity(nil, model.JSONB("[]")))
	assert.Equal(t, 30.0, calcJSONBSimilarity(model.JSONB("[]"), nil))
	// 完全相同
	a := model.JSONB(`["a","b"]`)
	b := model.JSONB(`["a","b"]`)
	assert.Equal(t, 100.0, calcJSONBSimilarity(a, b))
	// 不同但都非空
	c := model.JSONB(`["c"]`)
	assert.Equal(t, 50.0, calcJSONBSimilarity(a, c))
}

func TestCalcLocationMatch(t *testing.T) {
	// 任一为 0
	assert.Equal(t, 30.0, calcLocationMatch(0, 114, 30, 114))
	assert.Equal(t, 30.0, calcLocationMatch(30, 0, 30, 114))
	// 同点（distance² < 1）
	assert.Equal(t, 100.0, calcLocationMatch(30.5, 114.3, 30.5, 114.3))
	// 极近（distance² < 1）
	assert.Equal(t, 100.0, calcLocationMatch(30.5, 114.3, 30.501, 114.301))
	// distance² < 25（delta≈0.03）
	assert.Equal(t, 90.0, calcLocationMatch(30.0, 114.0, 30.03, 114.03))
	// 25 ≤ distance² < 100（delta≈0.06）
	assert.Equal(t, 70.0, calcLocationMatch(30.0, 114.0, 30.06, 114.06))
	// 100 ≤ distance² < 400（delta≈0.1）
	assert.Equal(t, 50.0, calcLocationMatch(30.0, 114.0, 30.1, 114.1))
	// 400 ≤ distance² < 1000（delta≈0.2）
	assert.Equal(t, 30.0, calcLocationMatch(30.0, 114.0, 30.2, 114.2))
	// distance² ≥ 1000（delta=1）
	assert.Equal(t, 10.0, calcLocationMatch(30.0, 114.0, 31.0, 115.0))
}

func TestCalcAgeMatch(t *testing.T) {
	// 任一为 0
	assert.Equal(t, 50.0, calcAgeMatch(0, 25))
	assert.Equal(t, 50.0, calcAgeMatch(25, 0))
	// 同龄
	assert.Equal(t, 100.0, calcAgeMatch(25, 25))
	// 差 2 岁
	assert.Equal(t, 100.0, calcAgeMatch(25, 27))
	// 差 5 岁
	assert.Equal(t, 85.0, calcAgeMatch(25, 30))
	// 差 10 岁
	assert.Equal(t, 60.0, calcAgeMatch(25, 35))
	// 差 15 岁
	assert.Equal(t, 40.0, calcAgeMatch(25, 40))
	// 差 20 岁
	assert.Equal(t, 20.0, calcAgeMatch(25, 45))
	// 负差也对称
	assert.Equal(t, 100.0, calcAgeMatch(27, 25))
}

// ===== AdminList 管理后台列表测试 =====

func TestLoveAdminList(t *testing.T) {
	t.Run("默认列表", func(t *testing.T) {
		svc, repo := newTestLoveService()
		repo.byID[1] = newLove(1, model.Love{UserID: 10, Status: model.LoveStatusActive})
		repo.byID[2] = newLove(2, model.Love{UserID: 11, Status: model.LoveStatusFrozen})
		req := &dto.LoveListRequest{Pagination: *utils.NewPagination(1, 10)}
		pg, list, err := svc.AdminList(req)
		require.NoError(t, err)
		assert.Equal(t, int64(2), pg.Total)
		assert.Len(t, list, 2)
		require.Len(t, repo.adminListCalls, 1)
		assert.Nil(t, repo.adminListCalls[0].Status)
		assert.Nil(t, repo.adminListCalls[0].AuditStatus)
	})

	t.Run("Status/AuditStatus 透传到仓储", func(t *testing.T) {
		svc, repo := newTestLoveService()
		repo.byID[1] = newLove(1, model.Love{UserID: 10, Status: model.LoveStatusFrozen, AuditStatus: model.LoveAuditPending})
		req := &dto.LoveListRequest{
			Pagination: *utils.NewPagination(1, 10),
			Status:     intPtr(model.LoveStatusFrozen),
			AuditStatus: intPtr(model.LoveAuditPending),
		}
		_, list, err := svc.AdminList(req)
		require.NoError(t, err)
		assert.Len(t, list, 1)
		require.Len(t, repo.adminListCalls, 1)
		require.NotNil(t, repo.adminListCalls[0].Status)
		assert.Equal(t, model.LoveStatusFrozen, *repo.adminListCalls[0].Status)
		require.NotNil(t, repo.adminListCalls[0].AuditStatus)
		assert.Equal(t, model.LoveAuditPending, *repo.adminListCalls[0].AuditStatus)
	})

	t.Run("仓储错误", func(t *testing.T) {
		svc, repo := newTestLoveService()
		repo.adminListErr = errors.New("db error")
		req := &dto.LoveListRequest{Pagination: *utils.NewPagination(1, 10)}
		_, _, err := svc.AdminList(req)
		assert.Error(t, err)
	})
}

// ===== AdminGetByID 测试 =====

func TestLoveAdminGetByID(t *testing.T) {
	t.Run("资料不存在", func(t *testing.T) {
		svc, _ := newTestLoveService()
		_, err := svc.AdminGetByID(999)
		assert.ErrorIs(t, err, ErrLoveNotFound)
	})

	t.Run("正常获取", func(t *testing.T) {
		svc, repo := newTestLoveService()
		repo.byID[1] = newLove(1, model.Love{UserID: 10, Nickname: "A"})
		info, err := svc.AdminGetByID(1)
		require.NoError(t, err)
		assert.Equal(t, "A", info.Nickname)
		// 管理后台不增加浏览数
		assert.Empty(t, repo.incrViewCalls)
	})
}

// ===== Audit 审核测试 =====

func TestLoveAudit(t *testing.T) {
	t.Run("正常审核通过", func(t *testing.T) {
		svc, repo := newTestLoveService()
		repo.byID[1] = newLove(1, model.Love{UserID: 10, AuditStatus: model.LoveAuditPending})
		err := svc.Audit(1, model.LoveAuditApproved, "")
		require.NoError(t, err)
		require.Len(t, repo.updateAuditCalls, 1)
		assert.Equal(t, model.LoveAuditApproved, repo.updateAuditCalls[0].auditStatus)
	})

	t.Run("审核拒绝带原因", func(t *testing.T) {
		svc, repo := newTestLoveService()
		repo.byID[1] = newLove(1, model.Love{UserID: 10, AuditStatus: model.LoveAuditPending})
		err := svc.Audit(1, model.LoveAuditRejected, "虚假资料")
		require.NoError(t, err)
		assert.Equal(t, model.LoveAuditRejected, repo.byID[1].AuditStatus)
		assert.Equal(t, "虚假资料", repo.byID[1].AuditReason)
	})

	t.Run("仓储审核失败", func(t *testing.T) {
		svc, repo := newTestLoveService()
		repo.updateAuditStatusErr = errors.New("db error")
		err := svc.Audit(1, model.LoveAuditApproved, "")
		assert.Error(t, err)
	})
}

// ===== AdminUpdateStatus 测试 =====

func TestLoveAdminUpdateStatus(t *testing.T) {
	t.Run("正常更新状态", func(t *testing.T) {
		svc, repo := newTestLoveService()
		repo.byID[1] = newLove(1, model.Love{UserID: 10, Status: model.LoveStatusActive})
		err := svc.AdminUpdateStatus(1, model.LoveStatusFrozen)
		require.NoError(t, err)
		require.Len(t, repo.updateStatusCalls, 1)
		assert.Equal(t, model.LoveStatusFrozen, repo.updateStatusCalls[0].status)
		assert.Equal(t, model.LoveStatusFrozen, repo.byID[1].Status)
	})

	t.Run("仓储失败", func(t *testing.T) {
		svc, repo := newTestLoveService()
		repo.updateStatusErr = errors.New("db error")
		err := svc.AdminUpdateStatus(1, model.LoveStatusFrozen)
		assert.Error(t, err)
	})
}

// ===== SetFeatured / SetPicked 测试 =====

func TestLoveSetFeatured(t *testing.T) {
	t.Run("设置精选", func(t *testing.T) {
		svc, repo := newTestLoveService()
		repo.byID[1] = newLove(1, model.Love{UserID: 10, Featured: false})
		err := svc.SetFeatured(1, true)
		require.NoError(t, err)
		require.Len(t, repo.setFeaturedCalls, 1)
		assert.True(t, repo.setFeaturedCalls[0].featured)
		assert.True(t, repo.byID[1].Featured)
	})

	t.Run("仓储失败", func(t *testing.T) {
		svc, repo := newTestLoveService()
		repo.setFeaturedErr = errors.New("db error")
		err := svc.SetFeatured(1, true)
		assert.Error(t, err)
	})
}

func TestLoveSetPicked(t *testing.T) {
	t.Run("设置甄选", func(t *testing.T) {
		svc, repo := newTestLoveService()
		repo.byID[1] = newLove(1, model.Love{UserID: 10, Picked: false})
		err := svc.SetPicked(1, true)
		require.NoError(t, err)
		require.Len(t, repo.setPickedCalls, 1)
		assert.True(t, repo.setPickedCalls[0].picked)
		assert.True(t, repo.byID[1].Picked)
	})

	t.Run("取消甄选", func(t *testing.T) {
		svc, repo := newTestLoveService()
		repo.byID[1] = newLove(1, model.Love{UserID: 10, Picked: true})
		err := svc.SetPicked(1, false)
		require.NoError(t, err)
		assert.False(t, repo.byID[1].Picked)
	})

	t.Run("仓储失败", func(t *testing.T) {
		svc, repo := newTestLoveService()
		repo.setPickedErr = errors.New("db error")
		err := svc.SetPicked(1, true)
		assert.Error(t, err)
	})
}

// ===== BatchAudit / BatchUpdateStatus 测试 =====

func TestLoveBatchAudit(t *testing.T) {
	t.Run("批量审核", func(t *testing.T) {
		svc, repo := newTestLoveService()
		repo.byID[1] = newLove(1, model.Love{UserID: 10, AuditStatus: model.LoveAuditPending})
		repo.byID[2] = newLove(2, model.Love{UserID: 11, AuditStatus: model.LoveAuditPending})
		err := svc.BatchAudit([]uint{1, 2}, model.LoveAuditApproved, "")
		require.NoError(t, err)
		require.Len(t, repo.batchAuditCalls, 1)
		assert.Equal(t, []uint{1, 2}, repo.batchAuditCalls[0].ids)
		assert.Equal(t, model.LoveAuditApproved, repo.byID[1].AuditStatus)
		assert.Equal(t, model.LoveAuditApproved, repo.byID[2].AuditStatus)
	})

	t.Run("批量审核拒绝带原因", func(t *testing.T) {
		svc, repo := newTestLoveService()
		repo.byID[1] = newLove(1, model.Love{UserID: 10})
		err := svc.BatchAudit([]uint{1}, model.LoveAuditRejected, "违规")
		require.NoError(t, err)
		assert.Equal(t, "违规", repo.byID[1].AuditReason)
	})

	t.Run("仓储失败", func(t *testing.T) {
		svc, repo := newTestLoveService()
		repo.batchAuditErr = errors.New("db error")
		err := svc.BatchAudit([]uint{1}, model.LoveAuditApproved, "")
		assert.Error(t, err)
	})
}

func TestLoveBatchUpdateStatus(t *testing.T) {
	t.Run("批量冻结", func(t *testing.T) {
		svc, repo := newTestLoveService()
		repo.byID[1] = newLove(1, model.Love{UserID: 10, Status: model.LoveStatusActive})
		repo.byID[2] = newLove(2, model.Love{UserID: 11, Status: model.LoveStatusActive})
		err := svc.BatchUpdateStatus([]uint{1, 2}, model.LoveStatusFrozen)
		require.NoError(t, err)
		require.Len(t, repo.batchStatusCalls, 1)
		assert.Equal(t, model.LoveStatusFrozen, repo.byID[1].Status)
		assert.Equal(t, model.LoveStatusFrozen, repo.byID[2].Status)
	})

	t.Run("空 ID 列表也透传到仓储", func(t *testing.T) {
		svc, repo := newTestLoveService()
		err := svc.BatchUpdateStatus(nil, model.LoveStatusFrozen)
		require.NoError(t, err)
		require.Len(t, repo.batchStatusCalls, 1)
		assert.Empty(t, repo.batchStatusCalls[0].ids)
	})

	t.Run("仓储失败", func(t *testing.T) {
		svc, repo := newTestLoveService()
		repo.batchStatusErr = errors.New("db error")
		err := svc.BatchUpdateStatus([]uint{1}, model.LoveStatusFrozen)
		assert.Error(t, err)
	})
}
