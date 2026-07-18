// Package service 文件业务主流程单元测试。
//
// 补齐 Upload/List/Delete 三大主流程的输入校验、参数透传、错误码映射、
// 错误透传与降级路径覆盖。与 file_presign_test.go / file_sts_test.go 互补，
// 共同构成 file service 的完整单元测试矩阵。
//
// 测试策略：
//   - Upload：仅覆盖输入校验路径（size/扩展名），成功路径会真实落盘 ./uploads，
//     避免污染工作目录故不测。
//   - List/Delete：使用内存 mockFileRepo 实现 FileRepository 四方法，
//     验证 pagination / fileType / keyword 透传、total 回填、
//     NotFound → ErrFileNotFound 映射、storage.Delete 失败不阻塞、
//     repo 错误原样透传。
package service

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"wuchang-tongcheng/internal/modules/file/dto"
	"wuchang-tongcheng/internal/modules/file/model"
	"wuchang-tongcheng/internal/pkg/utils"
)

// mockFileRepo 内存 mock，实现 FileRepository 接口。
type mockFileRepo struct {
	byID     map[uint]*model.FileUpload
	nextID   uint
	// 记录最近一次调用参数（断言透传）
	lastListRegionID  uint
	lastListFileType  string
	lastListKeyword   string
	lastListPage      int
	lastListPageSize  int
	lastDeleteID      uint
	// 注入错误
	createErr error
	getErr    error
	listErr   error
	deleteErr error
}

func newMockFileRepo() *mockFileRepo {
	return &mockFileRepo{
		byID:   make(map[uint]*model.FileUpload),
		nextID: 1,
	}
}

func (m *mockFileRepo) Create(record *model.FileUpload) error {
	if m.createErr != nil {
		return m.createErr
	}
	record.ID = m.nextID
	m.nextID++
	cp := *record
	m.byID[record.ID] = &cp
	return nil
}

func (m *mockFileRepo) GetByID(id uint) (*model.FileUpload, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	r, ok := m.byID[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *r
	return &cp, nil
}

func (m *mockFileRepo) List(regionID uint, p *utils.Pagination, fileType, keyword string) ([]model.FileUpload, int64, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	m.lastListRegionID = regionID
	m.lastListFileType = fileType
	m.lastListKeyword = keyword
	if p != nil {
		m.lastListPage = p.Page
		m.lastListPageSize = p.PageSize
	}
	var out []model.FileUpload
	var total int64
	for _, r := range m.byID {
		if r.RegionID != regionID {
			continue
		}
		if fileType != "" && r.FileType != fileType {
			continue
		}
		if keyword != "" && !contains(r.FileName, keyword) {
			continue
		}
		total++
		out = append(out, *r)
	}
	return out, total, nil
}

func (m *mockFileRepo) Delete(id uint) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	m.lastDeleteID = id
	delete(m.byID, id)
	return nil
}

// contains 简单子串匹配（模拟 SQL LIKE %keyword%）
func contains(s, sub string) bool {
	if sub == "" {
		return true
	}
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// ===== Upload 输入校验路径 =====

// TestUpload_EmptySize 空文件应拒绝（size=0）
func TestUpload_EmptySize(t *testing.T) {
	svc := NewFileService(newMockFileRepo())
	_, err := svc.Upload(1, 100, "photo.jpg", "image/jpeg", 0, nil)
	assert.ErrorIs(t, err, ErrFileEmpty)
}

// TestUpload_NegativeSize 负大小视为空文件
func TestUpload_NegativeSize(t *testing.T) {
	svc := NewFileService(newMockFileRepo())
	_, err := svc.Upload(1, 100, "photo.jpg", "image/jpeg", -1, nil)
	assert.ErrorIs(t, err, ErrFileEmpty)
}

// TestUpload_TooLarge 超出 50MB 上限应拒绝
func TestUpload_TooLarge(t *testing.T) {
	svc := NewFileService(newMockFileRepo())
	_, err := svc.Upload(1, 100, "big.mp4", "video/mp4", testMaxFileSize+1, nil)
	assert.ErrorIs(t, err, ErrFileTooLarge)
}

// TestUpload_InvalidType 不支持的扩展名应拒绝
func TestUpload_InvalidType(t *testing.T) {
	svc := NewFileService(newMockFileRepo())
	_, err := svc.Upload(1, 100, "malware.exe", "application/octet-stream", 1024, nil)
	assert.ErrorIs(t, err, ErrFileTypeInvalid)
}

// ===== List 主流程 =====

// TestList_Empty 空列表应返回 nil/空切片 + total=0
func TestList_Empty(t *testing.T) {
	repo := newMockFileRepo()
	svc := NewFileService(repo)

	p, list, err := svc.List(1, &dto.ListFilesRequest{Page: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Empty(t, list) // nil 或空切片均可
	assert.Equal(t, int64(0), p.Total)
}

// TestList_ResultsAndTotal 多结果回写 + total 计数 + 分页参数透传
func TestList_ResultsAndTotal(t *testing.T) {
	repo := newMockFileRepo()
	// 注入 3 条记录（region 1）
	r1 := &model.FileUpload{FileName: "a.jpg", FileType: "image"}
	r1.RegionID = 1
	require.NoError(t, repo.Create(r1))
	r2 := &model.FileUpload{FileName: "b.jpg", FileType: "image"}
	r2.RegionID = 1
	require.NoError(t, repo.Create(r2))
	r3 := &model.FileUpload{FileName: "c.mp4", FileType: "video"}
	r3.RegionID = 1
	require.NoError(t, repo.Create(r3))
	// region 2 的记录不应被查出
	r4 := &model.FileUpload{FileName: "d.jpg", FileType: "image"}
	r4.RegionID = 2
	require.NoError(t, repo.Create(r4))

	svc := NewFileService(repo)
	p, list, err := svc.List(1, &dto.ListFilesRequest{Page: 2, PageSize: 5})
	require.NoError(t, err)
	assert.Equal(t, int64(3), p.Total)
	assert.Len(t, list, 3)
	assert.Equal(t, 2, repo.lastListPage)
	assert.Equal(t, 5, repo.lastListPageSize)
	assert.Equal(t, uint(1), repo.lastListRegionID)
}

// TestList_FileTypeFilter 文件类型筛选透传
func TestList_FileTypeFilter(t *testing.T) {
	repo := newMockFileRepo()
	r1 := &model.FileUpload{FileName: "a.jpg", FileType: "image"}
	r1.RegionID = 1
	require.NoError(t, repo.Create(r1))
	r2 := &model.FileUpload{FileName: "b.mp4", FileType: "video"}
	r2.RegionID = 1
	require.NoError(t, repo.Create(r2))

	svc := NewFileService(repo)
	_, list, err := svc.List(1, &dto.ListFilesRequest{Page: 1, PageSize: 10, FileType: "image"})
	require.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, "a.jpg", list[0].FileName)
	assert.Equal(t, "image", repo.lastListFileType)
}

// TestList_KeywordFilter 关键词筛选透传
func TestList_KeywordFilter(t *testing.T) {
	repo := newMockFileRepo()
	r1 := &model.FileUpload{FileName: "report.pdf", FileType: "doc"}
	r1.RegionID = 1
	require.NoError(t, repo.Create(r1))
	r2 := &model.FileUpload{FileName: "photo.jpg", FileType: "image"}
	r2.RegionID = 1
	require.NoError(t, repo.Create(r2))

	svc := NewFileService(repo)
	_, list, err := svc.List(1, &dto.ListFilesRequest{Page: 1, PageSize: 10, Keyword: "port"})
	require.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, "report.pdf", list[0].FileName)
	assert.Equal(t, "port", repo.lastListKeyword)
}

// TestList_DefaultPagination page/pageSize 为 0 时走默认值（1/10）
func TestList_DefaultPagination(t *testing.T) {
	repo := newMockFileRepo()
	svc := NewFileService(repo)

	_, _, err := svc.List(1, &dto.ListFilesRequest{}) // Page/PageSize 为 0
	require.NoError(t, err)
	assert.Equal(t, 1, repo.lastListPage)
	assert.Equal(t, 10, repo.lastListPageSize)
}

// TestList_RepoError repo.List 错误原样透传
func TestList_RepoError(t *testing.T) {
	repo := newMockFileRepo()
	repo.listErr = errors.New("db connection lost")
	svc := NewFileService(repo)

	_, _, err := svc.List(1, &dto.ListFilesRequest{Page: 1, PageSize: 10})
	assert.Equal(t, "db connection lost", err.Error())
}

// ===== Delete 主流程 =====

// TestDelete_NotFound 记录不存在转 ErrFileNotFound
func TestDelete_NotFound(t *testing.T) {
	repo := newMockFileRepo()
	svc := NewFileService(repo)

	err := svc.Delete(999)
	assert.ErrorIs(t, err, ErrFileNotFound)
}

// TestDelete_GetByIDError 非 NotFound 错误原样透传
func TestDelete_GetByIDError(t *testing.T) {
	repo := newMockFileRepo()
	repo.getErr = errors.New("db timeout")
	svc := NewFileService(repo)

	err := svc.Delete(1)
	assert.Equal(t, "db timeout", err.Error())
}

// TestDelete_StorageFailureNotBlocking storage.Delete 失败不阻塞记录删除
//
// FileURL 使用非 /uploads/ 前缀，LocalStorage.Delete 会返回 "invalid file url" 错误，
// 但 service.Delete 忽略此错误，仍调用 repo.Delete 并返回 nil。
func TestDelete_StorageFailureNotBlocking(t *testing.T) {
	repo := newMockFileRepo()
	require.NoError(t, repo.Create(&model.FileUpload{FileName: "a.jpg", FileURL: "http://example.com/a.jpg"}))
	svc := NewFileService(repo)

	// LocalStorage.Delete 对非 /uploads/ 前缀的 URL 返回错误，但 service 层忽略
	err := svc.Delete(1)
	require.NoError(t, err)
	assert.Equal(t, uint(1), repo.lastDeleteID)
	// 记录应已被删除
	_, err = repo.GetByID(1)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

// TestDelete_RepoDeleteError repo.Delete 错误透传
func TestDelete_RepoDeleteError(t *testing.T) {
	repo := newMockFileRepo()
	require.NoError(t, repo.Create(&model.FileUpload{FileName: "a.jpg", FileURL: "/uploads/2026/07/a.jpg"}))
	repo.deleteErr = errors.New("db locked")
	svc := NewFileService(repo)

	err := svc.Delete(1)
	assert.Equal(t, "db locked", err.Error())
}

// TestDelete_SuccessWithValidPrefix FileURL 为合法 /uploads/ 前缀时 storage.Delete 静默（文件不存在不报错）
func TestDelete_SuccessWithValidPrefix(t *testing.T) {
	repo := newMockFileRepo()
	require.NoError(t, repo.Create(&model.FileUpload{FileName: "a.jpg", FileURL: "/uploads/2026/07/nonexist.jpg"}))
	svc := NewFileService(repo)

	err := svc.Delete(1)
	require.NoError(t, err)
	assert.Equal(t, uint(1), repo.lastDeleteID)
}

// ===== 构造函数 =====

// TestNewFileService 返回值类型断言
func TestNewFileService(t *testing.T) {
	svc := NewFileService(nil)
	assert.NotNil(t, svc)
	_, ok := svc.(*fileService)
	assert.True(t, ok, "NewFileService should return *fileService")
}
