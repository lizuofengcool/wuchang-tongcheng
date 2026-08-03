// Package handler_test 素材存储中台主 Handler HTTP 处理层单元测试。
//
// 使用 gin + httptest + mock Service，覆盖 handler 装配层全部分支：
//   - 文件上传（Upload）：未登录 401 "请先登录"、缺文件 400 "文件不能为空"、regionID+userID 注入、service 成功/错误透传（业务码 2901）
//   - 文件查询（GetFile）：file_id 参数、service 成功/错误透传（业务码 2902）
//   - 文件列表（ListFiles）：分页默认值与回退、userID 注入、service 错误透传（业务码 2901）、PageResult 结构断言
//   - 删除文件（DeleteFile）：file_id 参数、service 成功/错误透传（业务码 2901）
//   - 以图搜图（SearchByImage）：Bind 校验、regionID 注入、service 错误透传（业务码 2901）
//   - 添加水印（AddWatermark）：Bind 校验（required/oneof/max）、service 错误透传（业务码 2901）
//   - 生成缩略图（GenerateThumbnail）：Bind 校验、service 错误透传（业务码 2901）
//
// 鉴权由 AuthRequired / RequirePermission 中间件负责（测试中去掉，纯测 handler 装配层）。
// 不依赖 DB/Redis/Docker，纯内存 mock service 验证 handler 装配层逻辑。
// 与 ai/shop/category/region/news/file/setting/permission handler 测试同风格。
package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wuchang-tongcheng/internal/core/middleware"
	coreRouter "wuchang-tongcheng/internal/core/router"
	"wuchang-tongcheng/internal/modules/material/dto"
	materialHandler "wuchang-tongcheng/internal/modules/material/handler"
	"wuchang-tongcheng/internal/modules/material/service"
)

// apiResponse 解析统一响应体 {code, message, data}
type apiResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// pageData 解析 PageResult {list, total, page, pageSize}
type pageData struct {
	List     json.RawMessage `json:"list"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"pageSize"`
}

// mockMaterialService 内存 mock，实现 service.MaterialService 接口。
// 记录最近一次调用入参，返回预设 result/err，便于断言 handler 装配层逻辑。
type mockMaterialService struct {
	// Upload
	lastUploadRegionID uint
	lastUploadUserID   uint
	lastUploadReq      *dto.UploadRequest
	lastUploadFilename string
	lastUploadSize     int64
	lastUploadMime     string
	uploadResult       *dto.UploadResponse
	uploadErr          error

	// GetFile
	lastGetFileID string
	getResult     *dto.FileInfo
	getErr        error

	// ListFiles
	lastListReq  *dto.FileInfoListRequest
	listResult   []dto.FileInfo
	listErr      error

	// DeleteFile
	lastDeleteFileID string
	deleteErr        error

	// GenerateThumbnail
	lastThumbnailReq *dto.ThumbnailRequest
	thumbnailResult  string
	thumbnailErr     error

	// AddWatermark
	lastWatermarkReq *dto.WatermarkRequest
	watermarkErr     error

	// SearchByImage
	lastSearchReq  *dto.SearchByImageRequest
	searchResult   []dto.SimilarImage
	searchErr      error

	// UpdateTranscodeStatus（接口要求，handler 未直接调用）
	lastTranscodeFileID string
	lastTranscodeStatus int
	lastTranscodeJobs   string
	transcodeErr        error
}

// ===== 文件上传 =====

func (m *mockMaterialService) Upload(regionID uint, userID uint, req *dto.UploadRequest, filename string, size int64, mimeType string, reader io.Reader) (*dto.UploadResponse, error) {
	m.lastUploadRegionID = regionID
	m.lastUploadUserID = userID
	m.lastUploadReq = req
	m.lastUploadFilename = filename
	m.lastUploadSize = size
	m.lastUploadMime = mimeType
	// 读取 reader 以模拟消费（不实际存储）
	if reader != nil {
		_, _ = io.Copy(io.Discard, reader)
	}
	return m.uploadResult, m.uploadErr
}

// ===== 文件查询 =====

func (m *mockMaterialService) GetFile(fileID string) (*dto.FileInfo, error) {
	m.lastGetFileID = fileID
	return m.getResult, m.getErr
}

func (m *mockMaterialService) ListFiles(req *dto.FileInfoListRequest) ([]dto.FileInfo, int64, error) {
	m.lastListReq = req
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	return m.listResult, int64(len(m.listResult)), nil
}

func (m *mockMaterialService) DeleteFile(fileID string) error {
	m.lastDeleteFileID = fileID
	return m.deleteErr
}

// ===== 图片处理 =====

func (m *mockMaterialService) GenerateThumbnail(req *dto.ThumbnailRequest) (string, error) {
	m.lastThumbnailReq = req
	return m.thumbnailResult, m.thumbnailErr
}

func (m *mockMaterialService) AddWatermark(req *dto.WatermarkRequest) error {
	m.lastWatermarkReq = req
	return m.watermarkErr
}

// ===== 以图搜图 =====

func (m *mockMaterialService) SearchByImage(req *dto.SearchByImageRequest) ([]dto.SimilarImage, error) {
	m.lastSearchReq = req
	return m.searchResult, m.searchErr
}

// ===== 视频转码 =====

func (m *mockMaterialService) UpdateTranscodeStatus(fileID string, status int, jobs string) error {
	m.lastTranscodeFileID = fileID
	m.lastTranscodeStatus = status
	m.lastTranscodeJobs = jobs
	return m.transcodeErr
}

// 确保 mockMaterialService 实现 service.MaterialService 接口
var _ service.MaterialService = (*mockMaterialService)(nil)

// handlerEnv handler 测试环境
type handlerEnv struct {
	engine *gin.Engine
	mock   *mockMaterialService
}

// newHandlerEnv 构造 gin 引擎并注册 material 主 Handler 路由（路径与 material/plugin.go RegisterRoutes 一致）。
// ctxUserID 模拟 Auth 中间件注入的 user_id（0 表示未登录）。
// regionID 模拟 Region 中间件注入的 region_id。
// 路由注册去掉 AuthRequired / RequirePermission 中间件，纯测 handler 装配层逻辑。
func newHandlerEnv(t *testing.T, ctxUserID uint, regionID uint) *handlerEnv {
	t.Helper()
	gin.SetMode(gin.TestMode)

	mock := &mockMaterialService{
		uploadResult: &dto.UploadResponse{
			FileID:       "F20260803120000001",
			FileURL:      "/uploads/F20260803120000001.jpg",
			OriginalName: "avatar.jpg",
			FileSize:     1024,
			MimeType:     "image/jpeg",
			FileHash:     "abc123",
		},
		getResult: &dto.FileInfo{
			ID: 1, FileID: "F20260803120000001", UserID: ctxUserID, FileType: "image",
			FileURL: "/uploads/F20260803120000001.jpg", FileSize: 1024, MimeType: "image/jpeg",
			FileHash: "abc123", OriginalName: "avatar.jpg", Category: "user", RegionID: regionID,
		},
		listResult: []dto.FileInfo{
			{ID: 1, FileID: "F20260803120000001", FileType: "image", FileURL: "/uploads/a.jpg", RegionID: regionID},
			{ID: 2, FileID: "F20260803120000002", FileType: "video", FileURL: "/uploads/b.mp4", RegionID: regionID},
		},
		thumbnailResult: `{"100x100":"/uploads/F20260803120000001_100x100.jpg"}`,
		searchResult: []dto.SimilarImage{
			{FileID: "F20260803120000002", FileURL: "/uploads/b.jpg", Similarity: 0.875},
		},
	}

	r := coreRouter.NewRouter()
	// 模拟 Region + Auth 中间件：注入 region_id 和 user_id
	r.Use(func(c *gin.Context) {
		c.Set(middleware.RegionIDKey, regionID)
		c.Set(middleware.ContextUserID, ctxUserID)
		c.Next()
	})

	h := materialHandler.NewHandler(mock)
	// 注册路由，路径与 material/plugin.go RegisterRoutes 保持一致（去掉权限中间件，纯测 handler）
	root := r.Group("/api/v1/material")
	root.POST("/files", h.Upload)
	root.GET("/files", h.ListFiles)
	root.GET("/files/:file_id", h.GetFile)
	root.DELETE("/files/:file_id", h.DeleteFile)
	root.POST("/search-by-image", h.SearchByImage)
	root.POST("/watermark", h.AddWatermark)
	root.POST("/thumbnails", h.GenerateThumbnail)

	return &handlerEnv{engine: r.Engine(), mock: mock}
}

// doJSON 发起 JSON 请求，返回解析后的响应体。
func (e *handlerEnv) doJSON(t *testing.T, method, path string, body interface{}) *apiResponse {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		b, err := json.Marshal(body)
		require.NoError(t, err)
		buf.Write(b)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	e.engine.ServeHTTP(w, req)

	var resp apiResponse
	if w.Body.Len() > 0 {
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	}
	return &resp
}

// doRaw 发起原始请求（用于测试 Bind 失败：非法 JSON body）。
func (e *handlerEnv) doRaw(t *testing.T, method, path string, rawBody string, contentType string) *apiResponse {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(rawBody))
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()
	e.engine.ServeHTTP(w, req)

	var resp apiResponse
	if w.Body.Len() > 0 {
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	}
	return &resp
}

// doMultipart 发起 multipart/form-data 上传请求，field 字段名为 file，附带额外表单字段。
func (e *handlerEnv) doMultipart(t *testing.T, path, filename string, content []byte, extraFields map[string]string) *apiResponse {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for k, v := range extraFields {
		require.NoError(t, writer.WriteField(k, v))
	}
	part, err := writer.CreateFormFile("file", filename)
	require.NoError(t, err)
	_, err = part.Write(content)
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPost, path, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	e.engine.ServeHTTP(w, req)

	var resp apiResponse
	if w.Body.Len() > 0 {
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	}
	return &resp
}

// parsePage 解析响应 data 为 pageData
func parsePage(t *testing.T, resp *apiResponse) *pageData {
	t.Helper()
	var p pageData
	require.NoError(t, json.Unmarshal(resp.Data, &p))
	return &p
}

// assertParamError 断言 Bind 失败响应（消息以 "参数错误" 开头）
func assertParamError(t *testing.T, resp *apiResponse) {
	t.Helper()
	assert.Equal(t, 400, resp.Code)
	assert.True(t, strings.HasPrefix(resp.Message, "参数错误"), "expected message start with 参数错误, got: %s", resp.Message)
}

// ==================== 文件上传 ====================

// ---------- Upload ----------

func TestHandler_Upload_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 2)
	content := []byte("fake-image-bytes")
	resp := env.doMultipart(t, "/api/v1/material/files", "avatar.jpg", content, map[string]string{
		"category":  "user",
		"file_type": "image",
	})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "上传成功", resp.Message)
	assert.Equal(t, uint(2), env.mock.lastUploadRegionID)
	assert.Equal(t, uint(7), env.mock.lastUploadUserID)
	require.NotNil(t, env.mock.lastUploadReq)
	assert.Equal(t, "user", env.mock.lastUploadReq.Category)
	assert.Equal(t, "image", env.mock.lastUploadReq.FileType)
	assert.Equal(t, "avatar.jpg", env.mock.lastUploadFilename)
	assert.Equal(t, int64(len(content)), env.mock.lastUploadSize)
	assert.NotEmpty(t, env.mock.lastUploadMime)
	var out dto.UploadResponse
	require.NoError(t, json.Unmarshal(resp.Data, &out))
	assert.Equal(t, "F20260803120000001", out.FileID)
	assert.Equal(t, "/uploads/F20260803120000001.jpg", out.FileURL)
}

func TestHandler_Upload_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 2)
	resp := env.doMultipart(t, "/api/v1/material/files", "avatar.jpg", []byte("x"), nil)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastUploadUserID)
	assert.Nil(t, env.mock.lastUploadReq)
}

func TestHandler_Upload_NoFile(t *testing.T) {
	env := newHandlerEnv(t, 7, 2)
	// 不带 file 字段的空 multipart 请求
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	require.NoError(t, writer.Close())
	req := httptest.NewRequest(http.MethodPost, "/api/v1/material/files", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	env.engine.ServeHTTP(w, req)

	var resp apiResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, 400, resp.Code)
	assert.True(t, strings.HasPrefix(resp.Message, "文件不能为空"), "expected message start with 文件不能为空, got: %s", resp.Message)
	assert.Nil(t, env.mock.lastUploadReq)
}

func TestHandler_Upload_EmptyFormFields(t *testing.T) {
	// 不传 category/file_type 表单字段时，req 字段为零值，仍能上传
	env := newHandlerEnv(t, 7, 9)
	resp := env.doMultipart(t, "/api/v1/material/files", "pic.png", []byte("data"), nil)

	assert.Equal(t, 0, resp.Code)
	require.NotNil(t, env.mock.lastUploadReq)
	assert.Equal(t, "", env.mock.lastUploadReq.Category)
	assert.Equal(t, "", env.mock.lastUploadReq.FileType)
	assert.Equal(t, uint(9), env.mock.lastUploadRegionID)
}

func TestHandler_Upload_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 2)
	env.mock.uploadResult = nil
	env.mock.uploadErr = errors.New("不支持的文件类型")
	resp := env.doMultipart(t, "/api/v1/material/files", "x.bin", []byte("x"), nil)

	// 2901
	assert.Equal(t, 2901, resp.Code)
	assert.Equal(t, "不支持的文件类型", resp.Message)
	assert.Equal(t, uint(2), env.mock.lastUploadRegionID)
}

// ==================== 文件查询 ====================

// ---------- GetFile ----------

func TestHandler_GetFile_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 2)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/material/files/F20260803120000001", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "success", resp.Message)
	assert.Equal(t, "F20260803120000001", env.mock.lastGetFileID)
	var info dto.FileInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, "image", info.FileType)
	assert.Equal(t, "/uploads/F20260803120000001.jpg", info.FileURL)
}

func TestHandler_GetFile_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 2)
	env.mock.getResult = nil
	env.mock.getErr = errors.New("文件不存在")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/material/files/F999", nil)

	// 2902
	assert.Equal(t, 2902, resp.Code)
	assert.Equal(t, "文件不存在", resp.Message)
	assert.Equal(t, "F999", env.mock.lastGetFileID)
}

// ---------- ListFiles ----------

func TestHandler_ListFiles_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 2)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/material/files?page=2&page_size=15&file_type=image&category=user", nil)

	assert.Equal(t, 0, resp.Code)
	require.NotNil(t, env.mock.lastListReq)
	assert.Equal(t, uint(7), env.mock.lastListReq.UserID)
	assert.Equal(t, "image", env.mock.lastListReq.FileType)
	assert.Equal(t, "user", env.mock.lastListReq.Category)
	assert.Equal(t, 2, env.mock.lastListReq.Page)
	assert.Equal(t, 15, env.mock.lastListReq.PageSize)
	p := parsePage(t, resp)
	assert.Equal(t, int64(2), p.Total)
	var list []dto.FileInfo
	require.NoError(t, json.Unmarshal(p.List, &list))
	require.Len(t, list, 2)
	assert.Equal(t, "image", list[0].FileType)
}

func TestHandler_ListFiles_DefaultPagination(t *testing.T) {
	env := newHandlerEnv(t, 7, 2)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/material/files", nil)

	assert.Equal(t, 0, resp.Code)
	require.NotNil(t, env.mock.lastListReq)
	// 默认 page=1, page_size=20
	assert.Equal(t, 1, env.mock.lastListReq.Page)
	assert.Equal(t, 20, env.mock.lastListReq.PageSize)
}

func TestHandler_ListFiles_InvalidPaginationFallback(t *testing.T) {
	env := newHandlerEnv(t, 7, 2)
	// page=0, page_size=0 → 回退默认值
	resp := env.doJSON(t, http.MethodGet, "/api/v1/material/files?page=0&page_size=0", nil)

	assert.Equal(t, 0, resp.Code)
	require.NotNil(t, env.mock.lastListReq)
	assert.Equal(t, 1, env.mock.lastListReq.Page)
	assert.Equal(t, 20, env.mock.lastListReq.PageSize)
}

func TestHandler_ListFiles_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 2)
	env.mock.listResult = nil
	env.mock.listErr = errors.New("db down")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/material/files", nil)

	// 2901
	assert.Equal(t, 2901, resp.Code)
	assert.Equal(t, "db down", resp.Message)
}

// ==================== 删除文件 ====================

// ---------- DeleteFile ----------

func TestHandler_DeleteFile_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 2)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/material/files/F20260803120000001", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "删除成功", resp.Message)
	assert.Equal(t, "F20260803120000001", env.mock.lastDeleteFileID)
}

func TestHandler_DeleteFile_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 2)
	env.mock.deleteErr = errors.New("文件不存在")
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/material/files/F999", nil)

	// 2901
	assert.Equal(t, 2901, resp.Code)
	assert.Equal(t, "文件不存在", resp.Message)
	assert.Equal(t, "F999", env.mock.lastDeleteFileID)
}

// ==================== 以图搜图 ====================

// ---------- SearchByImage ----------

func TestHandler_SearchByImage_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 9)
	body := dto.SearchByImageRequest{FileID: "F20260803120000001", Limit: 5}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/material/search-by-image", body)

	assert.Equal(t, 0, resp.Code)
	require.NotNil(t, env.mock.lastSearchReq)
	assert.Equal(t, "F20260803120000001", env.mock.lastSearchReq.FileID)
	assert.Equal(t, 5, env.mock.lastSearchReq.Limit)
	// regionID 未在请求体指定时由 handler 注入
	assert.Equal(t, uint(9), env.mock.lastSearchReq.RegionID)
	var list []dto.SimilarImage
	require.NoError(t, json.Unmarshal(resp.Data, &list))
	require.Len(t, list, 1)
	assert.Equal(t, "F20260803120000002", list[0].FileID)
	assert.InDelta(t, 0.875, list[0].Similarity, 1e-9)
}

func TestHandler_SearchByImage_RegionIDPreserved(t *testing.T) {
	// 请求体显式带 region_id 时不被覆盖
	env := newHandlerEnv(t, 7, 9)
	body := dto.SearchByImageRequest{FileID: "F1", RegionID: 100}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/material/search-by-image", body)

	assert.Equal(t, 0, resp.Code)
	require.NotNil(t, env.mock.lastSearchReq)
	assert.Equal(t, uint(100), env.mock.lastSearchReq.RegionID)
}

func TestHandler_SearchByImage_BindError_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 7, 9)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/material/search-by-image", "{not json", "application/json")

	assertParamError(t, resp)
	assert.Nil(t, env.mock.lastSearchReq)
}

func TestHandler_SearchByImage_BindError_MissingFileID(t *testing.T) {
	env := newHandlerEnv(t, 7, 9)
	// 缺少 file_id（required）
	resp := env.doRaw(t, http.MethodPost, "/api/v1/material/search-by-image", `{"limit":5}`, "application/json")

	assertParamError(t, resp)
	assert.Nil(t, env.mock.lastSearchReq)
}

func TestHandler_SearchByImage_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 9)
	env.mock.searchResult = nil
	env.mock.searchErr = errors.New("图片不存在")
	body := dto.SearchByImageRequest{FileID: "F999"}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/material/search-by-image", body)

	// 2901
	assert.Equal(t, 2901, resp.Code)
	assert.Equal(t, "图片不存在", resp.Message)
}

// ==================== 添加水印 ====================

// ---------- AddWatermark ----------

func TestHandler_AddWatermark_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 2)
	body := dto.WatermarkRequest{FileID: "F20260803120000001", Text: "武昌同城", Position: "bottom-right"}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/material/watermark", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "水印添加成功", resp.Message)
	require.NotNil(t, env.mock.lastWatermarkReq)
	assert.Equal(t, "F20260803120000001", env.mock.lastWatermarkReq.FileID)
	assert.Equal(t, "武昌同城", env.mock.lastWatermarkReq.Text)
	assert.Equal(t, "bottom-right", env.mock.lastWatermarkReq.Position)
}

func TestHandler_AddWatermark_BindError_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 7, 2)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/material/watermark", "{bad", "application/json")

	assertParamError(t, resp)
	assert.Nil(t, env.mock.lastWatermarkReq)
}

func TestHandler_AddWatermark_BindError_MissingFileID(t *testing.T) {
	env := newHandlerEnv(t, 7, 2)
	// 缺少 file_id（required）
	resp := env.doRaw(t, http.MethodPost, "/api/v1/material/watermark", `{"text":"x"}`, "application/json")

	assertParamError(t, resp)
	assert.Nil(t, env.mock.lastWatermarkReq)
}

func TestHandler_AddWatermark_BindError_MissingText(t *testing.T) {
	env := newHandlerEnv(t, 7, 2)
	// 缺少 text（required）
	resp := env.doRaw(t, http.MethodPost, "/api/v1/material/watermark", `{"file_id":"F1"}`, "application/json")

	assertParamError(t, resp)
	assert.Nil(t, env.mock.lastWatermarkReq)
}

func TestHandler_AddWatermark_BindError_InvalidPosition(t *testing.T) {
	env := newHandlerEnv(t, 7, 2)
	// position=foo 不满足 oneof
	resp := env.doRaw(t, http.MethodPost, "/api/v1/material/watermark", `{"file_id":"F1","text":"x","position":"foo"}`, "application/json")

	assertParamError(t, resp)
	assert.Nil(t, env.mock.lastWatermarkReq)
}

func TestHandler_AddWatermark_BindError_TextTooLong(t *testing.T) {
	env := newHandlerEnv(t, 7, 2)
	// text 超过 max=64
	longText := strings.Repeat("字", 65)
	body, err := json.Marshal(dto.WatermarkRequest{FileID: "F1", Text: longText})
	require.NoError(t, err)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/material/watermark", string(body), "application/json")

	assertParamError(t, resp)
	assert.Nil(t, env.mock.lastWatermarkReq)
}

func TestHandler_AddWatermark_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 2)
	env.mock.watermarkErr = errors.New("图片不存在")
	body := dto.WatermarkRequest{FileID: "F999", Text: "x"}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/material/watermark", body)

	// 2901
	assert.Equal(t, 2901, resp.Code)
	assert.Equal(t, "图片不存在", resp.Message)
}

// ==================== 生成缩略图 ====================

// ---------- GenerateThumbnail ----------

func TestHandler_GenerateThumbnail_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 2)
	body := dto.ThumbnailRequest{FileID: "F20260803120000001", Sizes: []string{"100x100", "300x300"}}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/material/thumbnails", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "缩略图生成成功", resp.Message)
	require.NotNil(t, env.mock.lastThumbnailReq)
	assert.Equal(t, "F20260803120000001", env.mock.lastThumbnailReq.FileID)
	require.Len(t, env.mock.lastThumbnailReq.Sizes, 2)
	var out string
	require.NoError(t, json.Unmarshal(resp.Data, &out))
	assert.Contains(t, out, "100x100")
}

func TestHandler_GenerateThumbnail_Success_EmptySizes(t *testing.T) {
	// sizes 为空时由 service 补默认值，handler 仅透传
	env := newHandlerEnv(t, 7, 2)
	body := dto.ThumbnailRequest{FileID: "F1"}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/material/thumbnails", body)

	assert.Equal(t, 0, resp.Code)
	require.NotNil(t, env.mock.lastThumbnailReq)
	assert.Empty(t, env.mock.lastThumbnailReq.Sizes)
}

func TestHandler_GenerateThumbnail_BindError_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 7, 2)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/material/thumbnails", "{bad", "application/json")

	assertParamError(t, resp)
	assert.Nil(t, env.mock.lastThumbnailReq)
}

func TestHandler_GenerateThumbnail_BindError_MissingFileID(t *testing.T) {
	env := newHandlerEnv(t, 7, 2)
	// 缺少 file_id（required）
	resp := env.doRaw(t, http.MethodPost, "/api/v1/material/thumbnails", `{"sizes":["100x100"]}`, "application/json")

	assertParamError(t, resp)
	assert.Nil(t, env.mock.lastThumbnailReq)
}

func TestHandler_GenerateThumbnail_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 2)
	env.mock.thumbnailResult = ""
	env.mock.thumbnailErr = errors.New("文件不存在")
	body := dto.ThumbnailRequest{FileID: "F999"}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/material/thumbnails", body)

	// 2901
	assert.Equal(t, 2901, resp.Code)
	assert.Equal(t, "文件不存在", resp.Message)
}

// ==================== regionID 注入聚合 ====================

func TestHandler_RegionIDInjection_Aggregate(t *testing.T) {
	// 验证所有接收 regionID 的接口均透传 context 中的 regionID
	env := newHandlerEnv(t, 1, 9)

	// Upload
	env.doMultipart(t, "/api/v1/material/files", "a.jpg", []byte("x"), nil)
	assert.Equal(t, uint(9), env.mock.lastUploadRegionID)

	// SearchByImage（未显式带 region_id 时注入）
	env.doJSON(t, http.MethodPost, "/api/v1/material/search-by-image", dto.SearchByImageRequest{FileID: "F1"})
	require.NotNil(t, env.mock.lastSearchReq)
	assert.Equal(t, uint(9), env.mock.lastSearchReq.RegionID)
}

// ==================== userID 注入聚合 ====================

func TestHandler_UserIDInjection_Aggregate(t *testing.T) {
	// 验证所有接收 userID 的接口均透传 context 中的 userID
	env := newHandlerEnv(t, 8, 2)

	// Upload 用登录用户
	env.doMultipart(t, "/api/v1/material/files", "a.jpg", []byte("x"), nil)
	assert.Equal(t, uint(8), env.mock.lastUploadUserID)

	// ListFiles 用登录用户作为过滤条件
	env.doJSON(t, http.MethodGet, "/api/v1/material/files", nil)
	require.NotNil(t, env.mock.lastListReq)
	assert.Equal(t, uint(8), env.mock.lastListReq.UserID)
}
