// Package service 同城招聘求职主表业务逻辑层单元测试 - 职位包。
// 使用内存 mock 仓储覆盖：发布与默认值兜底（薪资单位/学历/性别/招聘类型等）、
// 月薪展示值按薪资单位换算（年/月/日/时）、过期天数兜底、发布即审核通过、
// 图片子表替换、更新字段构建与发布时间补齐、删除权限校验、详情浏览量自增与收藏状态、
// 列表/附近/搜索/高级搜索/我的发布/相似推荐的分页与错误传递、状态机迁移、
// 收藏切换（创建/删除 + 计数增减）、推广字段与置顶/紧急到期时间构建、
// 管理端审核去重与状态联动（通过同步发布、拒绝同步下架）、管理端状态更新等核心逻辑，不依赖 DB。
package service

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"wuchang-tongcheng/internal/modules/job/dto"
	"wuchang-tongcheng/internal/modules/job/model"
	"wuchang-tongcheng/internal/modules/job/repository"
	"wuchang-tongcheng/internal/pkg/utils"
)

// ===== mockJobRepo =====

type mockJobRepo struct {
	byID   map[uint]*model.Job
	nextID uint

	images map[uint][]model.JobImage

	createErr       error
	findErr         error
	updateFieldsErr error
	deleteErr       error
	replaceImgErr   error
	listErr         error
	nearbyErr       error
	searchErr       error
	advancedErr     error
	similarErr      error
	incrViewErr     error
	incrFavErr      error
	decrFavErr      error

	updateFieldsCalls []struct {
		id     uint
		fields map[string]interface{}
	}
	deleteCalls        []uint
	replaceImagesCalls []struct {
		jobID uint
		urls  []string
	}
	incrViewCalls []uint
	incrFavCalls  []uint
	decrFavCalls  []uint

	listReturn      []model.Job
	listTotal       int64
	adminListReturn []model.Job
	adminListTotal  int64
	nearbyReturn    []model.Job
	nearbyTotal     int64
	searchReturn    []model.Job
	searchTotal     int64
	advancedReturn  []model.Job
	advancedTotal   int64
	similarReturn   []model.Job
	byUserReturn    []model.Job
	byUserTotal     int64
}

func newMockJobRepo() *mockJobRepo {
	return &mockJobRepo{
		byID:    make(map[uint]*model.Job),
		nextID:  100,
		images:  make(map[uint][]model.JobImage),
	}
}

func (m *mockJobRepo) Create(j *model.Job) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.nextID++
	j.ID = m.nextID
	m.byID[j.ID] = j
	return nil
}

func (m *mockJobRepo) FindByID(id uint) (*model.Job, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	j, ok := m.byID[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *j
	return &cp, nil
}

func (m *mockJobRepo) Update(j *model.Job) error {
	m.byID[j.ID] = j
	return nil
}

func (m *mockJobRepo) UpdateFields(id uint, fields map[string]interface{}) error {
	if m.updateFieldsErr != nil {
		return m.updateFieldsErr
	}
	m.updateFieldsCalls = append(m.updateFieldsCalls, struct {
		id     uint
		fields map[string]interface{}
	}{id: id, fields: fields})
	j, ok := m.byID[id]
	if ok {
		if v, exists := fields["status"]; exists {
			if sv, ok := v.(int); ok {
				j.Status = sv
			}
		}
		if v, exists := fields["audit_status"]; exists {
			if av, ok := v.(int); ok {
				j.AuditStatus = av
			}
		}
		if v, exists := fields["audit_reason"]; exists {
			if rv, ok := v.(string); ok {
				j.AuditReason = rv
			}
		}
		if v, exists := fields["published_at"]; exists {
			if pv, ok := v.(*time.Time); ok {
				j.PublishedAt = pv
			}
		}
		if v, exists := fields["fav_count"]; exists {
			if fv, ok := v.(int); ok {
				j.FavCount = fv
			}
		}
		if v, exists := fields["promotion_level"]; exists {
			if pv, ok := v.(int); ok {
				j.PromotionLevel = pv
			}
		}
		if v, exists := fields["is_top"]; exists {
			if tv, ok := v.(bool); ok {
				j.IsTop = tv
			}
		}
		if v, exists := fields["is_urgent"]; exists {
			if uv, ok := v.(bool); ok {
				j.IsUrgent = uv
			}
		}
	}
	return nil
}

func (m *mockJobRepo) Delete(id uint) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	m.deleteCalls = append(m.deleteCalls, id)
	delete(m.byID, id)
	delete(m.images, id)
	return nil
}

func (m *mockJobRepo) List(regionID uint, pagination *utils.Pagination, opts repository.JobListOptions) ([]model.Job, int64, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	return m.listReturn, m.listTotal, nil
}

func (m *mockJobRepo) AdminList(pagination *utils.Pagination, opts repository.JobAdminListOptions) ([]model.Job, int64, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	return m.adminListReturn, m.adminListTotal, nil
}

func (m *mockJobRepo) ListNearby(regionID uint, pagination *utils.Pagination, lat, lng, radiusKm float64, opts repository.JobListOptions) ([]model.Job, int64, error) {
	if m.nearbyErr != nil {
		return nil, 0, m.nearbyErr
	}
	return m.nearbyReturn, m.nearbyTotal, nil
}

func (m *mockJobRepo) Search(regionID uint, pagination *utils.Pagination, keyword string) ([]model.Job, int64, error) {
	if m.searchErr != nil {
		return nil, 0, m.searchErr
	}
	return m.searchReturn, m.searchTotal, nil
}

func (m *mockJobRepo) AdvancedSearch(regionID uint, pagination *utils.Pagination, opts repository.JobAdvancedSearchOptions) ([]model.Job, int64, error) {
	if m.advancedErr != nil {
		return nil, 0, m.advancedErr
	}
	return m.advancedReturn, m.advancedTotal, nil
}

func (m *mockJobRepo) ListSimilar(jobID uint, limit int) ([]model.Job, error) {
	if m.similarErr != nil {
		return nil, m.similarErr
	}
	return m.similarReturn, nil
}

func (m *mockJobRepo) IncrViewCount(id uint) error {
	if m.incrViewErr != nil {
		return m.incrViewErr
	}
	m.incrViewCalls = append(m.incrViewCalls, id)
	return nil
}

func (m *mockJobRepo) IncrFavCount(id uint) error {
	if m.incrFavErr != nil {
		return m.incrFavErr
	}
	m.incrFavCalls = append(m.incrFavCalls, id)
	if j, ok := m.byID[id]; ok {
		j.FavCount++
	}
	return nil
}

func (m *mockJobRepo) DecrFavCount(id uint) error {
	if m.decrFavErr != nil {
		return m.decrFavErr
	}
	m.decrFavCalls = append(m.decrFavCalls, id)
	if j, ok := m.byID[id]; ok && j.FavCount > 0 {
		j.FavCount--
	}
	return nil
}

func (m *mockJobRepo) IncrDeliverCount(id uint) error    { return nil }
func (m *mockJobRepo) IncrInterviewCount(id uint) error  { return nil }
func (m *mockJobRepo) IncrOfferCount(id uint) error      { return nil }
func (m *mockJobRepo) IncrMessageCount(id uint) error    { return nil }

func (m *mockJobRepo) ListImages(jobID uint) ([]model.JobImage, error) {
	return m.images[jobID], nil
}

func (m *mockJobRepo) ReplaceImages(jobID uint, urls []string) error {
	if m.replaceImgErr != nil {
		return m.replaceImgErr
	}
	m.replaceImagesCalls = append(m.replaceImagesCalls, struct {
		jobID uint
		urls  []string
	}{jobID: jobID, urls: urls})
	imgs := make([]model.JobImage, 0, len(urls))
	for i, u := range urls {
		imgs = append(imgs, model.JobImage{JobID: jobID, URL: u, Sort: i})
	}
	m.images[jobID] = imgs
	return nil
}

func (m *mockJobRepo) DeleteImages(jobID uint) error {
	delete(m.images, jobID)
	return nil
}

func (m *mockJobRepo) ListByUser(userID uint, pagination *utils.Pagination) ([]model.Job, int64, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	return m.byUserReturn, m.byUserTotal, nil
}

func (m *mockJobRepo) ListByCompany(companyID uint, pagination *utils.Pagination) ([]model.Job, int64, error) {
	return nil, 0, nil
}

// ===== mockInteractionRepo =====

type mockInteractionRepo struct {
	favExists    bool
	favExistsErr error
	createFavErr error
	deleteFavErr error
	listFavsErr  error

	favExistsCalls []struct {
		userID   uint
		favType  string
		targetID uint
	}
	createFavCalls []*model.JobFavorite
	deleteFavCalls []struct {
		userID   uint
		favType  string
		targetID uint
	}

	listFavsReturn []model.JobFavorite
	listFavsTotal  int64
}

func (m *mockInteractionRepo) CreateFav(fav *model.JobFavorite) error {
	if m.createFavErr != nil {
		return m.createFavErr
	}
	m.createFavCalls = append(m.createFavCalls, fav)
	return nil
}

func (m *mockInteractionRepo) DeleteFav(userID uint, favType string, targetID uint) error {
	if m.deleteFavErr != nil {
		return m.deleteFavErr
	}
	m.deleteFavCalls = append(m.deleteFavCalls, struct {
		userID   uint
		favType  string
		targetID uint
	}{userID: userID, favType: favType, targetID: targetID})
	return nil
}

func (m *mockInteractionRepo) FavExists(userID uint, favType string, targetID uint) (bool, error) {
	m.favExistsCalls = append(m.favExistsCalls, struct {
		userID   uint
		favType  string
		targetID uint
	}{userID: userID, favType: favType, targetID: targetID})
	if m.favExistsErr != nil {
		return false, m.favExistsErr
	}
	return m.favExists, nil
}

func (m *mockInteractionRepo) ListFavs(userID uint, favType string, pagination *utils.Pagination) ([]model.JobFavorite, int64, error) {
	if m.listFavsErr != nil {
		return nil, 0, m.listFavsErr
	}
	return m.listFavsReturn, m.listFavsTotal, nil
}

func (m *mockInteractionRepo) HasFavedBatch(userID uint, favType string, ids []uint) (map[uint]bool, error) {
	return nil, nil
}

func (m *mockInteractionRepo) ClearFavs(userID uint, favType string) error { return nil }

func (m *mockInteractionRepo) CreateView(view *model.JobView) error { return nil }

func (m *mockInteractionRepo) ListViews(userID uint, viewType string, pagination *utils.Pagination) ([]model.JobView, int64, error) {
	return nil, 0, nil
}

func (m *mockInteractionRepo) ListViewsByTarget(targetType string, targetID uint, pagination *utils.Pagination) ([]model.JobView, int64, error) {
	return nil, 0, nil
}

func (m *mockInteractionRepo) ClearViews(userID uint, viewType string) error { return nil }

// ===== 辅助 =====

func newService(repo *mockJobRepo, favRepo *mockInteractionRepo) JobService {
	if repo == nil {
		repo = newMockJobRepo()
	}
	if favRepo == nil {
		favRepo = &mockInteractionRepo{}
	}
	return NewJobService(repo, favRepo)
}

// jobWithID 构造带 ID 的 Job 指针（ID 来自嵌入 RegionBaseModel.BaseModel，
// 无法在复合字面量中直接设置，需通过字段赋值）。
func jobWithID(id uint, j model.Job) *model.Job {
	j.ID = id
	return &j
}

// jobsWithID 构造带 ID 的 Job 切片（同 jobWithID 原因）。
func jobsWithID(js []model.Job, ids ...uint) []model.Job {
	for i := range js {
		if i < len(ids) {
			js[i].ID = ids[i]
		}
	}
	return js
}

// ===== Create 测试 =====

func TestJobCreate_DefaultValuesFallback(t *testing.T) {
	repo := newMockJobRepo()
	svc := newService(repo, nil)

	req := &dto.CreateJobRequest{
		Title:      "Golang 工程师",
		SalaryMin:  10000,
		Status:     model.StatusPublished,
	}
	info, err := svc.Create(1, 7, "张三", "13800000000", "avatar.png", req)
	require.NoError(t, err)
	require.NotNil(t, info)

	// 默认值兜底
	assert.Equal(t, model.SalaryUnitMonth, info.SalaryUnit)
	assert.Equal(t, model.EducationUnlimited, info.Education)
	assert.Equal(t, model.GenderUnlimited, info.GenderRequirement)
	assert.Equal(t, model.RecruitmentTypeFullTime, info.RecruitmentType)
	assert.Equal(t, model.EmploymentTypeRegular, info.EmploymentType)
	assert.Equal(t, model.TravelNone, info.TravelFrequency)
	assert.Equal(t, model.OvertimeUnknown, info.OvertimeStatus)

	// 发布即审核通过
	assert.Equal(t, model.AuditApproved, info.AuditStatus)
	// 发布状态 -> PublishedAt 已设置
	require.NotNil(t, info.PublishedAt)
	// 过期时间已设置（默认 30 天）
	require.NotNil(t, info.ExpiryTime)
	// 月薪按"月"换算 = SalaryMin
	assert.Equal(t, 10000.0, info.SalaryMonthly)
	// region 隔离
	assert.Equal(t, uint(1), info.RegionID)
	// 发布者信息
	assert.Equal(t, uint(7), info.UserID)
	assert.Equal(t, "张三", info.UserName)
	// Images nil -> []
	assert.Equal(t, []string{}, info.Images)
}

func TestJobCreate_ExpireDaysDefaultWhenZeroOrNegative(t *testing.T) {
	cases := []int{0, -5}
	for _, days := range cases {
		repo := newMockJobRepo()
		svc := newService(repo, nil)
		req := &dto.CreateJobRequest{Title: "测试岗位", ExpireDays: days}
		info, err := svc.Create(1, 1, "u", "p", "a", req)
		require.NoError(t, err)
		require.NotNil(t, info.ExpiryTime)
		// 默认 30 天：过期时间应在 29~31 天范围内
		diff := info.ExpiryTime.Sub(time.Now()).Hours() / 24
		assert.InDelta(t, 30, diff, 1, "days=%d 应兜底为 30 天", days)
	}
}

func TestJobCreate_CustomExpireDays(t *testing.T) {
	repo := newMockJobRepo()
	svc := newService(repo, nil)
	req := &dto.CreateJobRequest{Title: "测试", ExpireDays: 7}
	info, err := svc.Create(1, 1, "u", "p", "a", req)
	require.NoError(t, err)
	diff := info.ExpiryTime.Sub(time.Now()).Hours() / 24
	assert.InDelta(t, 7, diff, 1)
}

func TestJobCreate_SalaryMonthlyCalcByUnit(t *testing.T) {
	cases := []struct {
		name     string
		unit     string
		salary   float64
		expected float64
	}{
		{"月薪", model.SalaryUnitMonth, 10000, 10000},
		{"年薪除12", model.SalaryUnitYear, 120000, 10000},
		{"日薪乘22", model.SalaryUnitDay, 500, 11000},
		{"时薪乘8乘22", model.SalaryUnitHour, 50, 8800},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			repo := newMockJobRepo()
			svc := newService(repo, nil)
			req := &dto.CreateJobRequest{
				Title:      "测试",
				SalaryMin:  c.salary,
				SalaryUnit: c.unit,
			}
			info, err := svc.Create(1, 1, "u", "p", "a", req)
			require.NoError(t, err)
			assert.Equal(t, c.expected, info.SalaryMonthly)
		})
	}
}

func TestJobCreate_SalaryMonthlyZeroWhenSalaryMinZero(t *testing.T) {
	repo := newMockJobRepo()
	svc := newService(repo, nil)
	req := &dto.CreateJobRequest{Title: "测试", SalaryMin: 0, SalaryUnit: model.SalaryUnitMonth}
	info, err := svc.Create(1, 1, "u", "p", "a", req)
	require.NoError(t, err)
	assert.Equal(t, 0.0, info.SalaryMonthly)
}

func TestJobCreate_DraftStatusNoPublishedAt(t *testing.T) {
	repo := newMockJobRepo()
	svc := newService(repo, nil)
	req := &dto.CreateJobRequest{Title: "草稿", Status: model.StatusDraft}
	info, err := svc.Create(1, 1, "u", "p", "a", req)
	require.NoError(t, err)
	assert.Nil(t, info.PublishedAt)
	assert.Equal(t, model.StatusDraft, info.Status)
}

func TestJobCreate_ImagesReplaced(t *testing.T) {
	repo := newMockJobRepo()
	svc := newService(repo, nil)
	req := &dto.CreateJobRequest{
		Title:  "带图职位",
		Images: []string{"http://img/1.png", "http://img/2.png"},
	}
	info, err := svc.Create(1, 1, "u", "p", "a", req)
	require.NoError(t, err)
	assert.Equal(t, req.Images, info.Images)
	require.Len(t, repo.replaceImagesCalls, 1)
	assert.Equal(t, req.Images, repo.replaceImagesCalls[0].urls)
}

func TestJobCreate_JSONBFieldsParsed(t *testing.T) {
	repo := newMockJobRepo()
	svc := newService(repo, nil)
	req := &dto.CreateJobRequest{
		Title:         "测试",
		Benefits:      []uint{1, 2, 3},
		Skills:        []uint{10, 20},
		Tags:          []string{"急招", "高薪"},
		WelfareTags:   []string{"五险一金", "餐补"},
		Allowances:    map[string]interface{}{"housing": 800},
		PromotionChannels: map[string]interface{}{"path": "tech"},
	}
	info, err := svc.Create(1, 1, "u", "p", "a", req)
	require.NoError(t, err)
	assert.Equal(t, []uint{1, 2, 3}, info.Benefits)
	assert.Equal(t, []uint{10, 20}, info.Skills)
	assert.Equal(t, []string{"急招", "高薪"}, info.Tags)
	assert.Equal(t, []string{"五险一金", "餐补"}, info.WelfareTags)
	assert.Equal(t, 800.0, info.Allowances["housing"])
	assert.Equal(t, "tech", info.PromotionChannels["path"])
}

func TestJobCreate_CreateErrorPropagation(t *testing.T) {
	repo := newMockJobRepo()
	repo.createErr = errors.New("db down")
	svc := newService(repo, nil)
	_, err := svc.Create(1, 1, "u", "p", "a", &dto.CreateJobRequest{Title: "x"})
	require.Error(t, err)
	assert.Equal(t, "db down", err.Error())
}

func TestJobCreate_ReplaceImagesErrorIgnored(t *testing.T) {
	repo := newMockJobRepo()
	repo.replaceImgErr = errors.New("replace failed")
	svc := newService(repo, nil)
	// ReplaceImages 错误被忽略（_ =），Create 仍成功
	info, err := svc.Create(1, 1, "u", "p", "a", &dto.CreateJobRequest{Title: "x", Images: []string{"a"}})
	require.NoError(t, err)
	assert.NotNil(t, info)
}

// ===== toJobInfo 测试 =====

func TestToJobInfo_DefaultsAndJSONBParse(t *testing.T) {
	j := &model.Job{
		Title:     "测试",
		SalaryMin: 8000,
	}
	bens, _ := model.FromJSON([]uint{1, 2})
	j.Benefits = bens
	tags, _ := model.FromJSON([]string{"a"})
	j.Tags = tags
	info := toJobInfo(j, nil)
	assert.Equal(t, model.SalaryUnitMonth, info.SalaryUnit)
	assert.Equal(t, model.EducationUnlimited, info.Education)
	assert.Equal(t, model.GenderUnlimited, info.GenderRequirement)
	assert.Equal(t, []string{}, info.Images)
	assert.Equal(t, []uint{1, 2}, info.Benefits)
	assert.Equal(t, []string{"a"}, info.Tags)
}

func TestToJobInfo_NilJSONBFields(t *testing.T) {
	j := &model.Job{Title: "x"}
	info := toJobInfo(j, []string{"img"})
	assert.Equal(t, []string{"img"}, info.Images)
	assert.Nil(t, info.Benefits)
	assert.Nil(t, info.Skills)
	assert.Nil(t, info.Tags)
	assert.Nil(t, info.WelfareTags)
	assert.Nil(t, info.Allowances)
	assert.Nil(t, info.PromotionChannels)
}

// ===== Update 测试 =====

func TestJobUpdate_NotFound(t *testing.T) {
	repo := newMockJobRepo()
	svc := newService(repo, nil)
	err := svc.Update(999, 1, &dto.UpdateJobRequest{Title: "新标题"})
	assert.ErrorIs(t, err, ErrJobNotFound)
}

func TestJobUpdate_NoPermission(t *testing.T) {
	repo := newMockJobRepo()
	repo.byID[1] = &model.Job{Title: "原标题", UserID: 10}
	svc := newService(repo, nil)
	err := svc.Update(1, 99, &dto.UpdateJobRequest{Title: "新标题"})
	assert.ErrorIs(t, err, ErrJobNoPermission)
}

func TestJobUpdate_FieldsBuilding(t *testing.T) {
	repo := newMockJobRepo()
	repo.byID[1] = &model.Job{Title: "原标题", UserID: 10}
	svc := newService(repo, nil)
	t1 := true
	err := svc.Update(1, 10, &dto.UpdateJobRequest{
		Title:            "新标题",
		Content:          "新内容",
		SalaryMin:        15000,
		SalaryUnit:       model.SalaryUnitMonth,
		Education:        model.EducationBachelor,
		WorkCity:         "武汉",
		HiringCount:      3,
		IsUrgent:         &t1,
		HasSocialInsurance: &t1,
		ContactPhone:     "13900000000",
	})
	require.NoError(t, err)
	require.Len(t, repo.updateFieldsCalls, 1)
	fields := repo.updateFieldsCalls[0].fields
	assert.Equal(t, "新标题", fields["title"])
	assert.Equal(t, "新内容", fields["content"])
	assert.Equal(t, 15000.0, fields["salary_min"])
	assert.Equal(t, model.SalaryUnitMonth, fields["salary_unit"])
	assert.Equal(t, model.EducationBachelor, fields["education"])
	assert.Equal(t, "武汉", fields["work_city"])
	assert.Equal(t, 3, fields["hiring_count"])
	assert.Equal(t, true, fields["is_urgent"])
	assert.Equal(t, true, fields["has_social_insurance"])
	assert.Equal(t, "13900000000", fields["contact_phone"])
}

func TestJobUpdate_ZeroValuesNotIncluded(t *testing.T) {
	repo := newMockJobRepo()
	repo.byID[1] = &model.Job{Title: "原标题", UserID: 10}
	svc := newService(repo, nil)
	// 空字符串/零值字段不应进入 fields
	err := svc.Update(1, 10, &dto.UpdateJobRequest{})
	require.NoError(t, err)
	fields := repo.updateFieldsCalls[0].fields
	_, hasTitle := fields["title"]
	assert.False(t, hasTitle, "空标题不应进入 fields")
	_, hasCity := fields["work_city"]
	assert.False(t, hasCity)
}

func TestJobUpdate_PublishedAtSetOnPublish(t *testing.T) {
	repo := newMockJobRepo()
	repo.byID[1] = &model.Job{UserID: 10, Status: model.StatusDraft}
	svc := newService(repo, nil)
	err := svc.Update(1, 10, &dto.UpdateJobRequest{Status: model.StatusPublished})
	require.NoError(t, err)
	fields := repo.updateFieldsCalls[0].fields
	assert.Equal(t, model.StatusPublished, fields["status"])
	pa, ok := fields["published_at"]
	assert.True(t, ok)
	assert.NotNil(t, pa)
}

func TestJobUpdate_PublishedAtNotResetIfAlreadySet(t *testing.T) {
	repo := newMockJobRepo()
	now := time.Now()
	repo.byID[1] = &model.Job{UserID: 10, Status: model.StatusPublished, PublishedAt: &now}
	svc := newService(repo, nil)
	err := svc.Update(1, 10, &dto.UpdateJobRequest{Status: model.StatusPublished})
	require.NoError(t, err)
	fields := repo.updateFieldsCalls[0].fields
	_, hasPA := fields["published_at"]
	assert.False(t, hasPA, "已有 PublishedAt 不应重复设置")
}

func TestJobUpdate_ExpireDaysSetsExpiryTime(t *testing.T) {
	repo := newMockJobRepo()
	repo.byID[1] = &model.Job{UserID: 10}
	svc := newService(repo, nil)
	err := svc.Update(1, 10, &dto.UpdateJobRequest{ExpireDays: 15})
	require.NoError(t, err)
	fields := repo.updateFieldsCalls[0].fields
	et, ok := fields["expiry_time"]
	require.True(t, ok)
	require.NotNil(t, et)
}

func TestJobUpdate_ImagesReplacedWhenNotNil(t *testing.T) {
	repo := newMockJobRepo()
	repo.byID[1] = &model.Job{UserID: 10}
	svc := newService(repo, nil)
	err := svc.Update(1, 10, &dto.UpdateJobRequest{Images: []string{"a", "b"}})
	require.NoError(t, err)
	require.Len(t, repo.replaceImagesCalls, 1)
	assert.Equal(t, []string{"a", "b"}, repo.replaceImagesCalls[0].urls)
}

func TestJobUpdate_ImagesNotReplacedWhenNil(t *testing.T) {
	repo := newMockJobRepo()
	repo.byID[1] = &model.Job{UserID: 10}
	svc := newService(repo, nil)
	err := svc.Update(1, 10, &dto.UpdateJobRequest{Title: "x"})
	require.NoError(t, err)
	assert.Empty(t, repo.replaceImagesCalls)
}

func TestJobUpdate_UpdateFieldsError(t *testing.T) {
	repo := newMockJobRepo()
	repo.byID[1] = &model.Job{UserID: 10}
	repo.updateFieldsErr = errors.New("update failed")
	svc := newService(repo, nil)
	err := svc.Update(1, 10, &dto.UpdateJobRequest{Title: "x"})
	require.Error(t, err)
	assert.Equal(t, "update failed", err.Error())
}

func TestJobUpdate_JSONBFields(t *testing.T) {
	repo := newMockJobRepo()
	repo.byID[1] = &model.Job{UserID: 10}
	svc := newService(repo, nil)
	err := svc.Update(1, 10, &dto.UpdateJobRequest{
		Benefits:    []uint{5},
		Tags:        []string{"new"},
		Allowances:  map[string]interface{}{"car": 500},
	})
	require.NoError(t, err)
	fields := repo.updateFieldsCalls[0].fields
	assert.NotNil(t, fields["benefits"])
	assert.NotNil(t, fields["tags"])
	assert.NotNil(t, fields["allowances"])
}

// ===== Delete 测试 =====

func TestJobDelete_NotFound(t *testing.T) {
	repo := newMockJobRepo()
	svc := newService(repo, nil)
	err := svc.Delete(999, 1)
	assert.ErrorIs(t, err, ErrJobNotFound)
}

func TestJobDelete_NoPermission(t *testing.T) {
	repo := newMockJobRepo()
	repo.byID[1] = &model.Job{UserID: 10}
	svc := newService(repo, nil)
	err := svc.Delete(1, 99)
	assert.ErrorIs(t, err, ErrJobNoPermission)
}

func TestJobDelete_Success(t *testing.T) {
	repo := newMockJobRepo()
	repo.byID[1] = &model.Job{UserID: 10}
	svc := newService(repo, nil)
	err := svc.Delete(1, 10)
	require.NoError(t, err)
	require.Len(t, repo.deleteCalls, 1)
	assert.Equal(t, uint(1), repo.deleteCalls[0])
}

func TestJobDelete_DeleteError(t *testing.T) {
	repo := newMockJobRepo()
	repo.byID[1] = &model.Job{UserID: 10}
	repo.deleteErr = errors.New("delete failed")
	svc := newService(repo, nil)
	err := svc.Delete(1, 10)
	require.Error(t, err)
	assert.Equal(t, "delete failed", err.Error())
}

// ===== GetByID 测试 =====

func TestJobGetByID_NotFound(t *testing.T) {
	repo := newMockJobRepo()
	svc := newService(repo, nil)
	_, err := svc.GetByID(999, 1)
	assert.ErrorIs(t, err, ErrJobNotFound)
}

func TestJobGetByID_IncrViewAndImages(t *testing.T) {
	repo := newMockJobRepo()
	repo.byID[1] = jobWithID(1, model.Job{Title: "职位", UserID: 10})
	repo.images[1] = []model.JobImage{{URL: "a.png"}, {URL: "b.png"}}
	svc := newService(repo, nil)
	resp, err := svc.GetByID(1, 0)
	require.NoError(t, err)
	require.Len(t, repo.incrViewCalls, 1)
	assert.Equal(t, uint(1), repo.incrViewCalls[0])
	assert.Equal(t, []string{"a.png", "b.png"}, resp.Images)
	assert.False(t, resp.HasFaved, "userID=0 时不查收藏")
}

func TestJobGetByID_HasFavedWhenUserLoggedIn(t *testing.T) {
	repo := newMockJobRepo()
	repo.byID[1] = jobWithID(1, model.Job{UserID: 10})
	favRepo := &mockInteractionRepo{favExists: true}
	svc := newService(repo, favRepo)
	resp, err := svc.GetByID(1, 5)
	require.NoError(t, err)
	assert.True(t, resp.HasFaved)
	require.Len(t, favRepo.favExistsCalls, 1)
	assert.Equal(t, uint(5), favRepo.favExistsCalls[0].userID)
	assert.Equal(t, model.FavoriteTypeJob, favRepo.favExistsCalls[0].favType)
}

func TestJobGetByID_FavExistsErrorIgnored(t *testing.T) {
	repo := newMockJobRepo()
	repo.byID[1] = jobWithID(1, model.Job{UserID: 10})
	favRepo := &mockInteractionRepo{favExistsErr: errors.New("fav err")}
	svc := newService(repo, favRepo)
	resp, err := svc.GetByID(1, 5)
	require.NoError(t, err, "FavExists 错误被忽略")
	assert.False(t, resp.HasFaved)
}

// ===== List 测试 =====

func TestJobList_SuccessWithImages(t *testing.T) {
	repo := newMockJobRepo()
	repo.listReturn = jobsWithID([]model.Job{{Title: "a"}, {Title: "b"}}, 1, 2)
	repo.listTotal = 2
	repo.images[1] = []model.JobImage{{URL: "img1"}}
	svc := newService(repo, nil)
	p, list, err := svc.List(1, &dto.JobListRequest{})
	require.NoError(t, err)
	assert.Equal(t, int64(2), p.Total)
	require.Len(t, list, 2)
	assert.Equal(t, []string{"img1"}, list[0].Images)
	assert.Equal(t, []string{}, list[1].Images, "无图职位 Images 应为空切片")
}

func TestJobList_ErrorPropagation(t *testing.T) {
	repo := newMockJobRepo()
	repo.listErr = errors.New("list err")
	svc := newService(repo, nil)
	_, _, err := svc.List(1, &dto.JobListRequest{})
	require.Error(t, err)
	assert.Equal(t, "list err", err.Error())
}

func TestJobList_EmptyResult(t *testing.T) {
	repo := newMockJobRepo()
	svc := newService(repo, nil)
	p, list, err := svc.List(1, &dto.JobListRequest{})
	require.NoError(t, err)
	assert.Equal(t, int64(0), p.Total)
	assert.Empty(t, list)
}

// ===== ListNearby 测试 =====

func TestJobListNearby_Success(t *testing.T) {
	repo := newMockJobRepo()
	repo.nearbyReturn = jobsWithID([]model.Job{{Title: "附近职位", Distance: 2.5}}, 1)
	repo.nearbyTotal = 1
	svc := newService(repo, nil)
	p, list, err := svc.ListNearby(1, &dto.JobNearbyRequest{Latitude: 30.5, Longitude: 114.3})
	require.NoError(t, err)
	assert.Equal(t, int64(1), p.Total)
	require.Len(t, list, 1)
	assert.Equal(t, "附近职位", list[0].Title)
}

func TestJobListNearby_ErrorPropagation(t *testing.T) {
	repo := newMockJobRepo()
	repo.nearbyErr = errors.New("nearby err")
	svc := newService(repo, nil)
	_, _, err := svc.ListNearby(1, &dto.JobNearbyRequest{Latitude: 30, Longitude: 114})
	require.Error(t, err)
	assert.Equal(t, "nearby err", err.Error())
}

// ===== Search 测试 =====

func TestJobSearch_Success(t *testing.T) {
	repo := newMockJobRepo()
	repo.searchReturn = jobsWithID([]model.Job{{Title: "Golang"}}, 1)
	repo.searchTotal = 1
	svc := newService(repo, nil)
	p, list, err := svc.Search(1, &dto.JobSearchRequest{Keyword: "Golang"})
	require.NoError(t, err)
	assert.Equal(t, int64(1), p.Total)
	require.Len(t, list, 1)
}

func TestJobSearch_ErrorPropagation(t *testing.T) {
	repo := newMockJobRepo()
	repo.searchErr = errors.New("search err")
	svc := newService(repo, nil)
	_, _, err := svc.Search(1, &dto.JobSearchRequest{Keyword: "x"})
	require.Error(t, err)
}

// ===== AdvancedSearch 测试 =====

func TestJobAdvancedSearch_Success(t *testing.T) {
	repo := newMockJobRepo()
	repo.advancedReturn = jobsWithID([]model.Job{{}, {}}, 1, 2)
	repo.advancedTotal = 2
	svc := newService(repo, nil)
	p, list, err := svc.AdvancedSearch(1, &dto.AdvancedSearchRequest{Keyword: "go", Sort: "salary_desc"})
	require.NoError(t, err)
	assert.Equal(t, int64(2), p.Total)
	require.Len(t, list, 2)
}

func TestJobAdvancedSearch_ErrorPropagation(t *testing.T) {
	repo := newMockJobRepo()
	repo.advancedErr = errors.New("advanced err")
	svc := newService(repo, nil)
	_, _, err := svc.AdvancedSearch(1, &dto.AdvancedSearchRequest{})
	require.Error(t, err)
	assert.Equal(t, "advanced err", err.Error())
}

// ===== ListMine 测试 =====

func TestJobListMine_Success(t *testing.T) {
	repo := newMockJobRepo()
	repo.byUserReturn = jobsWithID([]model.Job{{Title: "我的职位"}}, 1)
	repo.byUserTotal = 1
	svc := newService(repo, nil)
	p, list, err := svc.ListMine(10, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), p.Total)
	require.Len(t, list, 1)
}

func TestJobListMine_ErrorPropagation(t *testing.T) {
	repo := newMockJobRepo()
	repo.listErr = errors.New("byuser err")
	svc := newService(repo, nil)
	_, _, err := svc.ListMine(10, 1, 10)
	require.Error(t, err)
}

// ===== ListSimilar 测试 =====

func TestJobListSimilar_Success(t *testing.T) {
	repo := newMockJobRepo()
	repo.similarReturn = jobsWithID([]model.Job{
		{Title: "相似A", SalaryMin: 8000, SalaryUnit: model.SalaryUnitMonth, WorkCity: "武汉"},
		{Title: "相似B", SalaryMin: 10000, SalaryUnit: model.SalaryUnitMonth, WorkCity: "北京"},
	}, 2, 3)
	svc := newService(repo, nil)
	list, err := svc.ListSimilar(1, 5)
	require.NoError(t, err)
	require.Len(t, list, 2)
	assert.Equal(t, uint(2), list[0].JobID)
	assert.Equal(t, 0.8, list[0].Similarity, "相似度固定 0.8")
	assert.Equal(t, "武汉", list[0].WorkCity)
}

func TestJobListSimilar_Empty(t *testing.T) {
	repo := newMockJobRepo()
	svc := newService(repo, nil)
	list, err := svc.ListSimilar(1, 5)
	require.NoError(t, err)
	assert.Empty(t, list)
}

func TestJobListSimilar_ErrorPropagation(t *testing.T) {
	repo := newMockJobRepo()
	repo.similarErr = errors.New("similar err")
	svc := newService(repo, nil)
	_, err := svc.ListSimilar(1, 5)
	require.Error(t, err)
}

// ===== UpdateStatus 测试 =====

func TestJobUpdateStatus_NotFound(t *testing.T) {
	repo := newMockJobRepo()
	svc := newService(repo, nil)
	err := svc.UpdateStatus(999, 1, model.StatusPublished)
	assert.ErrorIs(t, err, ErrJobNotFound)
}

func TestJobUpdateStatus_NoPermission(t *testing.T) {
	repo := newMockJobRepo()
	repo.byID[1] = &model.Job{UserID: 10}
	svc := newService(repo, nil)
	err := svc.UpdateStatus(1, 99, model.StatusPublished)
	assert.ErrorIs(t, err, ErrJobNoPermission)
}

func TestJobUpdateStatus_PublishSetsPublishedAt(t *testing.T) {
	repo := newMockJobRepo()
	repo.byID[1] = &model.Job{UserID: 10, Status: model.StatusDraft}
	svc := newService(repo, nil)
	err := svc.UpdateStatus(1, 10, model.StatusPublished)
	require.NoError(t, err)
	fields := repo.updateFieldsCalls[0].fields
	assert.Equal(t, model.StatusPublished, fields["status"])
	_, ok := fields["published_at"]
	assert.True(t, ok, "发布应设置 published_at")
}

func TestJobUpdateStatus_OfflineNoPublishedAt(t *testing.T) {
	repo := newMockJobRepo()
	repo.byID[1] = &model.Job{UserID: 10}
	svc := newService(repo, nil)
	err := svc.UpdateStatus(1, 10, model.StatusOffline)
	require.NoError(t, err)
	fields := repo.updateFieldsCalls[0].fields
	assert.Equal(t, model.StatusOffline, fields["status"])
	_, ok := fields["published_at"]
	assert.False(t, ok, "下架不应设置 published_at")
}

// ===== Fav 测试 =====

func TestJobFav_CreateNew(t *testing.T) {
	repo := newMockJobRepo()
	repo.byID[1] = jobWithID(1, model.Job{FavCount: 0})
	favRepo := &mockInteractionRepo{favExists: false}
	svc := newService(repo, favRepo)
	resp, err := svc.Fav(5, 1)
	require.NoError(t, err)
	assert.True(t, resp.HasFaved)
	require.Len(t, favRepo.createFavCalls, 1)
	assert.Equal(t, uint(5), favRepo.createFavCalls[0].UserID)
	assert.Equal(t, uint(1), favRepo.createFavCalls[0].JobID)
	assert.Equal(t, model.FavoriteTypeJob, favRepo.createFavCalls[0].FavoriteType)
	assert.True(t, favRepo.createFavCalls[0].Notify, "默认通知")
	require.Len(t, repo.incrFavCalls, 1)
	assert.Equal(t, 1, resp.FavCount)
}

func TestJobFav_AlreadyExistsNoCreate(t *testing.T) {
	repo := newMockJobRepo()
	repo.byID[1] = jobWithID(1, model.Job{FavCount: 3})
	favRepo := &mockInteractionRepo{favExists: true}
	svc := newService(repo, favRepo)
	resp, err := svc.Fav(5, 1)
	require.NoError(t, err)
	assert.True(t, resp.HasFaved)
	assert.Empty(t, favRepo.createFavCalls, "已存在不应重复创建")
	assert.Empty(t, repo.incrFavCalls, "已存在不应自增计数")
	assert.Equal(t, 3, resp.FavCount)
}

func TestJobFav_FavExistsError(t *testing.T) {
	repo := newMockJobRepo()
	favRepo := &mockInteractionRepo{favExistsErr: errors.New("exists err")}
	svc := newService(repo, favRepo)
	_, err := svc.Fav(5, 1)
	require.Error(t, err)
	assert.Equal(t, "exists err", err.Error())
}

func TestJobFav_CreateFavError(t *testing.T) {
	repo := newMockJobRepo()
	favRepo := &mockInteractionRepo{favExists: false, createFavErr: errors.New("create err")}
	svc := newService(repo, favRepo)
	_, err := svc.Fav(5, 1)
	require.Error(t, err)
	assert.Equal(t, "create err", err.Error())
}

// ===== Unfav 测试 =====

func TestJobUnfav_DeleteExisting(t *testing.T) {
	repo := newMockJobRepo()
	repo.byID[1] = jobWithID(1, model.Job{FavCount: 2})
	favRepo := &mockInteractionRepo{favExists: true}
	svc := newService(repo, favRepo)
	resp, err := svc.Unfav(5, 1)
	require.NoError(t, err)
	assert.False(t, resp.HasFaved)
	require.Len(t, favRepo.deleteFavCalls, 1)
	assert.Equal(t, uint(5), favRepo.deleteFavCalls[0].userID)
	assert.Equal(t, model.FavoriteTypeJob, favRepo.deleteFavCalls[0].favType)
	require.Len(t, repo.decrFavCalls, 1)
	assert.Equal(t, 1, resp.FavCount)
}

func TestJobUnfav_NotExistsNoDelete(t *testing.T) {
	repo := newMockJobRepo()
	repo.byID[1] = jobWithID(1, model.Job{FavCount: 0})
	favRepo := &mockInteractionRepo{favExists: false}
	svc := newService(repo, favRepo)
	resp, err := svc.Unfav(5, 1)
	require.NoError(t, err)
	assert.False(t, resp.HasFaved)
	assert.Empty(t, favRepo.deleteFavCalls)
	assert.Empty(t, repo.decrFavCalls)
}

func TestJobUnfav_DeleteFavError(t *testing.T) {
	repo := newMockJobRepo()
	favRepo := &mockInteractionRepo{favExists: true, deleteFavErr: errors.New("del err")}
	svc := newService(repo, favRepo)
	_, err := svc.Unfav(5, 1)
	require.Error(t, err)
	assert.Equal(t, "del err", err.Error())
}

// ===== FavStatus 测试 =====

func TestJobFavStatus_True(t *testing.T) {
	repo := newMockJobRepo()
	repo.byID[1] = jobWithID(1, model.Job{FavCount: 5})
	favRepo := &mockInteractionRepo{favExists: true}
	svc := newService(repo, favRepo)
	resp, err := svc.FavStatus(5, 1)
	require.NoError(t, err)
	assert.True(t, resp.HasFaved)
	assert.Equal(t, 5, resp.FavCount)
}

func TestJobFavStatus_FavExistsError(t *testing.T) {
	repo := newMockJobRepo()
	favRepo := &mockInteractionRepo{favExistsErr: errors.New("err")}
	svc := newService(repo, favRepo)
	_, err := svc.FavStatus(5, 1)
	require.Error(t, err)
}

// ===== ListFavs 测试 =====

func TestJobListFavs_Success(t *testing.T) {
	repo := newMockJobRepo()
	repo.byID[1] = jobWithID(1, model.Job{Title: "收藏职位"})
	repo.byID[2] = jobWithID(2, model.Job{Title: "另一个"})
	favRepo := &mockInteractionRepo{
		listFavsReturn: []model.JobFavorite{
			{JobID: 1},
			{JobID: 2},
		},
		listFavsTotal: 2,
	}
	svc := newService(repo, favRepo)
	p, list, err := svc.ListFavs(5, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(2), p.Total)
	require.Len(t, list, 2)
	assert.Equal(t, "收藏职位", list[0].Title)
}

func TestJobListFavs_SkipMissingJobs(t *testing.T) {
	repo := newMockJobRepo()
	repo.byID[1] = jobWithID(1, model.Job{Title: "存在"})
	// JobID=2 不存在 -> FindByID 返回 ErrRecordNotFound -> 跳过
	favRepo := &mockInteractionRepo{
		listFavsReturn: []model.JobFavorite{{JobID: 1}, {JobID: 2}},
		listFavsTotal:  2,
	}
	svc := newService(repo, favRepo)
	p, list, err := svc.ListFavs(5, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(2), p.Total, "total 来自仓储")
	require.Len(t, list, 1, "不存在的职位应被跳过")
	assert.Equal(t, "存在", list[0].Title)
}

func TestJobListFavs_ErrorPropagation(t *testing.T) {
	repo := newMockJobRepo()
	favRepo := &mockInteractionRepo{listFavsErr: errors.New("list favs err")}
	svc := newService(repo, favRepo)
	_, _, err := svc.ListFavs(5, 1, 10)
	require.Error(t, err)
}

// ===== Promotion 测试 =====

func TestJobPromotion_NotFound(t *testing.T) {
	repo := newMockJobRepo()
	svc := newService(repo, nil)
	err := svc.Promotion(999, 1, &dto.JobPromotionRequest{})
	assert.ErrorIs(t, err, ErrJobNotFound)
}

func TestJobPromotion_NoPermission(t *testing.T) {
	repo := newMockJobRepo()
	repo.byID[1] = &model.Job{UserID: 10}
	svc := newService(repo, nil)
	err := svc.Promotion(1, 99, &dto.JobPromotionRequest{})
	assert.ErrorIs(t, err, ErrJobNoPermission)
}

func TestJobPromotion_FieldsBuilding(t *testing.T) {
	repo := newMockJobRepo()
	repo.byID[1] = &model.Job{UserID: 10}
	svc := newService(repo, nil)
	err := svc.Promotion(1, 10, &dto.JobPromotionRequest{
		PromotionLevel: 5,
		TrafficWeight:  3.5,
		Featured:       true,
		Picked:         true,
		Verified:       true,
		IsTop:          true,
		TopDays:        7,
		IsUrgent:       true,
		UrgentDays:     3,
	})
	require.NoError(t, err)
	fields := repo.updateFieldsCalls[0].fields
	assert.Equal(t, 5, fields["promotion_level"])
	assert.Equal(t, 3.5, fields["traffic_weight"])
	assert.Equal(t, true, fields["featured"])
	assert.Equal(t, true, fields["picked"])
	assert.Equal(t, true, fields["verified"])
	assert.Equal(t, true, fields["is_top"])
	assert.Equal(t, true, fields["is_urgent"])
	te, ok := fields["top_expire"]
	assert.True(t, ok, "置顶应设置 top_expire")
	require.NotNil(t, te)
	ue, ok := fields["urgent_expire"]
	assert.True(t, ok, "紧急应设置 urgent_expire")
	require.NotNil(t, ue)
}

func TestJobPromotion_NoExpireWhenDaysZero(t *testing.T) {
	repo := newMockJobRepo()
	repo.byID[1] = &model.Job{UserID: 10}
	svc := newService(repo, nil)
	err := svc.Promotion(1, 10, &dto.JobPromotionRequest{IsTop: true, TopDays: 0, IsUrgent: true, UrgentDays: 0})
	require.NoError(t, err)
	fields := repo.updateFieldsCalls[0].fields
	_, hasTop := fields["top_expire"]
	assert.False(t, hasTop, "TopDays=0 不应设置 top_expire")
	_, hasUrgent := fields["urgent_expire"]
	assert.False(t, hasUrgent, "UrgentDays=0 不应设置 urgent_expire")
}

func TestJobPromotion_UpdateFieldsError(t *testing.T) {
	repo := newMockJobRepo()
	repo.byID[1] = &model.Job{UserID: 10}
	repo.updateFieldsErr = errors.New("upd err")
	svc := newService(repo, nil)
	err := svc.Promotion(1, 10, &dto.JobPromotionRequest{})
	require.Error(t, err)
}

// ===== AdminList 测试 =====

func TestJobAdminList_Success(t *testing.T) {
	repo := newMockJobRepo()
	repo.adminListReturn = jobsWithID([]model.Job{{Title: "管理端"}}, 1)
	repo.adminListTotal = 1
	svc := newService(repo, nil)
	p, list, err := svc.AdminList(&dto.JobAdminListRequest{Keyword: "管理"})
	require.NoError(t, err)
	assert.Equal(t, int64(1), p.Total)
	require.Len(t, list, 1)
}

func TestJobAdminList_ErrorPropagation(t *testing.T) {
	repo := newMockJobRepo()
	repo.listErr = errors.New("admin err")
	svc := newService(repo, nil)
	_, _, err := svc.AdminList(&dto.JobAdminListRequest{})
	require.Error(t, err)
}

// ===== AdminGetByID 测试 =====

func TestAdminGetByID_Success(t *testing.T) {
	repo := newMockJobRepo()
	repo.byID[1] = jobWithID(1, model.Job{Title: "职位"})
	repo.images[1] = []model.JobImage{{URL: "img"}}
	svc := newService(repo, nil)
	resp, err := svc.AdminGetByID(1)
	require.NoError(t, err)
	assert.Equal(t, "职位", resp.Title)
	assert.Equal(t, []string{"img"}, resp.Images)
}

func TestAdminGetByID_NotFound(t *testing.T) {
	repo := newMockJobRepo()
	svc := newService(repo, nil)
	_, err := svc.AdminGetByID(999)
	assert.ErrorIs(t, err, ErrJobNotFound)
}

// ===== Audit 测试 =====

func TestJobAudit_NotFound(t *testing.T) {
	repo := newMockJobRepo()
	svc := newService(repo, nil)
	err := svc.Audit(999, model.AuditApproved, "")
	assert.ErrorIs(t, err, ErrJobNotFound)
}

func TestJobAudit_DuplicateStatusRejected(t *testing.T) {
	repo := newMockJobRepo()
	repo.byID[1] = &model.Job{AuditStatus: model.AuditApproved}
	svc := newService(repo, nil)
	err := svc.Audit(1, model.AuditApproved, "")
	assert.ErrorIs(t, err, ErrJobAudited)
}

func TestJobAudit_ApproveDraftSyncsPublish(t *testing.T) {
	repo := newMockJobRepo()
	repo.byID[1] = &model.Job{Status: model.StatusDraft, AuditStatus: model.AuditPending}
	svc := newService(repo, nil)
	err := svc.Audit(1, model.AuditApproved, "")
	require.NoError(t, err)
	fields := repo.updateFieldsCalls[0].fields
	assert.Equal(t, model.AuditApproved, fields["audit_status"])
	assert.Equal(t, model.StatusPublished, fields["status"], "审核通过 + 草稿 -> 同步发布")
	_, ok := fields["published_at"]
	assert.True(t, ok, "应设置 published_at")
}

func TestJobAudit_ApproveAlreadyPublishedNoStatusChange(t *testing.T) {
	repo := newMockJobRepo()
	now := time.Now()
	repo.byID[1] = &model.Job{Status: model.StatusPublished, AuditStatus: model.AuditPending, PublishedAt: &now}
	svc := newService(repo, nil)
	err := svc.Audit(1, model.AuditApproved, "")
	require.NoError(t, err)
	fields := repo.updateFieldsCalls[0].fields
	_, hasStatus := fields["status"]
	assert.False(t, hasStatus, "已发布状态不应重复设置 status")
	_, hasPA := fields["published_at"]
	assert.False(t, hasPA, "已有 PublishedAt 不应重复设置")
}

func TestJobAudit_RejectSyncsOffline(t *testing.T) {
	repo := newMockJobRepo()
	repo.byID[1] = &model.Job{Status: model.StatusPublished, AuditStatus: model.AuditApproved}
	svc := newService(repo, nil)
	err := svc.Audit(1, model.AuditRejected, "内容违规")
	require.NoError(t, err)
	fields := repo.updateFieldsCalls[0].fields
	assert.Equal(t, model.AuditRejected, fields["audit_status"])
	assert.Equal(t, "内容违规", fields["audit_reason"])
	assert.Equal(t, model.StatusOffline, fields["status"], "审核拒绝 -> 同步下架")
}

func TestJobAudit_UpdateFieldsError(t *testing.T) {
	repo := newMockJobRepo()
	repo.byID[1] = &model.Job{AuditStatus: model.AuditPending}
	repo.updateFieldsErr = errors.New("upd err")
	svc := newService(repo, nil)
	err := svc.Audit(1, model.AuditApproved, "")
	require.Error(t, err)
}

// ===== AdminUpdateStatus 测试 =====

func TestAdminUpdateStatus_PublishSetsPublishedAt(t *testing.T) {
	repo := newMockJobRepo()
	svc := newService(repo, nil)
	err := svc.AdminUpdateStatus(1, model.StatusPublished)
	require.NoError(t, err)
	fields := repo.updateFieldsCalls[0].fields
	assert.Equal(t, model.StatusPublished, fields["status"])
	_, ok := fields["published_at"]
	assert.True(t, ok)
}

func TestAdminUpdateStatus_OfflineNoPublishedAt(t *testing.T) {
	repo := newMockJobRepo()
	svc := newService(repo, nil)
	err := svc.AdminUpdateStatus(1, model.StatusOffline)
	require.NoError(t, err)
	fields := repo.updateFieldsCalls[0].fields
	assert.Equal(t, model.StatusOffline, fields["status"])
	_, ok := fields["published_at"]
	assert.False(t, ok)
}

func TestAdminUpdateStatus_UpdateFieldsError(t *testing.T) {
	repo := newMockJobRepo()
	repo.updateFieldsErr = errors.New("upd err")
	svc := newService(repo, nil)
	err := svc.AdminUpdateStatus(1, model.StatusPublished)
	require.Error(t, err)
}

// ===== NewJobService 测试 =====

func TestNewJobService_NotNil(t *testing.T) {
	svc := NewJobService(newMockJobRepo(), &mockInteractionRepo{})
	assert.NotNil(t, svc)
}
