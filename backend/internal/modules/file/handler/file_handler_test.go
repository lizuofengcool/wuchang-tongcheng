// Package handler_test 文件模块 HTTP 处理层单元测试。
//
// 使用 gin + httptest + mock FileService，覆盖 handler 全部分支：
//   - 未登录拦截（写操作 userID=0 → 401：Upload/Presign/Commit/STS）
//   - multipart 文件上传缺失（FormFile 失败 → 1303 "请上传文件"）
//   - URL :id 参数解析失败（非数字 → 400 "无效的文件ID"）
//   - 请求体 Bind 失败（非法 JSON → 400）
//   - file_name/object_name 缺失或纯空格（binding + TrimSpace 双层校验）
//   - service 错误码路由（ErrPresignNotSupported → 1306 / ErrFileTypeInvalid → 1305 /
//     ErrFileTooLarge → 1304 / ErrFileEmpty → 1301 / sts.ErrNotConfigured → 1307 / 其他 → 1302）
//   - regionID 注入（Upload/Presign/Commit/STS/List 从上下文读取 regionID）
//   - 公开读取无需登录（List/Delete 在 handler 层不校验 userID，
//     鉴权由 plugin.go 的 RequirePermission 中间件负责，handler 层仅写操作做防御性登录校验）
//
// 不依赖 DB/Redis/Docker/真实对象存储，纯内存 mock service 验证 handler 装配层逻辑。
// 与 setting/region/category handler 测试同风格，区别在于 file 涉及 multipart 上传 +
// STS context 参数 + 多种 sentinel 错误码路由（Presign/Commit/STS 三接口各有专属错误码分支）。
package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wuchang-tongcheng/internal/core/middleware"
	coreRouter "wuchang-tongcheng/internal/core/router"
	"wuchang-tongcheng/internal/modules/file/dto"
	fileHandler "wuchang-tongcheng/internal/modules/file/handler"
	"wuchang-tongcheng/internal/modules/file/model"
	"wuchang-tongcheng/internal/modules/file/service"
	"wuchang-tongcheng/internal/pkg/storage"
	"wuchang-tongcheng/internal/pkg/sts"
	"wuchang-tongcheng/internal/pkg/utils"
)

// apiResponse 解析统一响应体 {code, message, data}
type apiResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// mockFileService 内存 mock，实现 service.FileService 接口。
// 记录最近一次调用入参，返回预设 result/err，便于断言 handler 装配层逻辑。
type mockFileService struct {
	// 调用记录
	lastUploadRegionID uint
	lastUploadUserID   uint
	lastUploadFilename string
	lastUploadMIME     string
	lastUploadSize     int64
	lastListRegionID   uint
	lastListReq        *dto.ListFilesRequest
	lastDeleteID       uint
	lastPresignRegionID uint
	lastPresignUserID   uint
	lastPresignFilename string
	lastCommitRegionID  uint
	lastCommitUserID    uint
	lastCommitFilename  string
	lastCommitObject    string
	lastCommitMIME      string
	lastCommitSize      int64
	lastSTSRegionID     uint
	lastSTSUserID       uint

	// 返回值预设
	uploadResult   *model.FileUpload
	uploadErr      error
	listPagination *utils.Pagination
	listResult     []model.FileUpload
	listErr        error
	deleteErr      error
	presignResult  *dto.PresignUploadResponse
	presignErr     error
	commitResult   *model.FileUpload
	commitErr      error
	stsResult      *dto.STSCredentialsResponse
	stsErr         error
}

func (m *mockFileService) Upload(regionID uint, userID uint, filename string, mimeType string, size int64, reader io.Reader) (*model.FileUpload, error) {
	m.lastUploadRegionID = regionID
	m.lastUploadUserID = userID
	m.lastUploadFilename = filename
	m.lastUploadMIME = mimeType
	m.lastUploadSize = size
	// 排空 reader 避免 multipart 上传场景下后续读取阻塞
	if reader != nil {
		_, _ = io.Copy(io.Discard, reader)
	}
	return m.uploadResult, m.uploadErr
}

func (m *mockFileService) List(regionID uint, req *dto.ListFilesRequest) (*utils.Pagination, []model.FileUpload, error) {
	m.lastListRegionID = regionID
	m.lastListReq = req
	return m.listPagination, m.listResult, m.listErr
}

func (m *mockFileService) Delete(id uint) error {
	m.lastDeleteID = id
	return m.deleteErr
}

func (m *mockFileService) PresignUpload(regionID uint, userID uint, filename string) (*dto.PresignUploadResponse, error) {
	m.lastPresignRegionID = regionID
	m.lastPresignUserID = userID
	m.lastPresignFilename = filename
	return m.presignResult, m.presignErr
}

func (m *mockFileService) CommitUpload(regionID uint, userID uint, filename, objectName, mimeType string, size int64) (*model.FileUpload, error) {
	m.lastCommitRegionID = regionID
	m.lastCommitUserID = userID
	m.lastCommitFilename = filename
	m.lastCommitObject = objectName
	m.lastCommitMIME = mimeType
	m.lastCommitSize = size
	return m.commitResult, m.commitErr
}

func (m *mockFileService) GetSTSCredentials(ctx context.Context, regionID uint, userID uint) (*dto.STSCredentialsResponse, error) {
	_ = ctx // handler 内部用 context.WithTimeout 创建新 ctx，此处仅占位不校验
	m.lastSTSRegionID = regionID
	m.lastSTSUserID = userID
	return m.stsResult, m.stsErr
}

// 确保 mockFileService 实现 service.FileService 接口
var _ service.FileService = (*mockFileService)(nil)

// handlerEnv handler 测试环境
type handlerEnv struct {
	engine *gin.Engine
	mock   *mockFileService
}

// newHandlerEnv 构造 gin 引擎并注册 file 路由（与 file/plugin.go RegisterRoutes 路径一致）。
// ctxUserID 用于模拟 Auth 中间件注入的 user_id（0 表示未登录）。
// ctxRegionID 用于模拟 Region 中间件注入的 region_id（0 表示未注入，handler 兜底 0）。
// 注意：file handler 的读操作（List/Delete）不校验 userID，
// 故 ctxUserID 仅影响写操作（Upload/Presign/Commit/STS）的登录拦截分支。
func newHandlerEnv(t *testing.T, ctxUserID uint, ctxRegionID uint) *handlerEnv {
	t.Helper()
	gin.SetMode(gin.TestMode)

	// 构造预设上传记录（嵌入字段 ID/RegionID 需单独设置）
	uploadRec := &model.FileUpload{
		FileName: "test.jpg",
		FileURL:  "https://cdn.example.com/uploads/2026/07/test.jpg",
		FileSize: 1024,
		FileType: "image",
		MimeType: "image/jpeg",
	}
	uploadRec.ID = 1
	uploadRec.RegionID = 5

	commitRec := &model.FileUpload{
		FileName: "photo.png",
		FileURL:  "https://cdn.example.com/uploads/2026/07/abc.png",
		FileSize: 2048,
		FileType: "image",
		MimeType: "image/png",
	}
	commitRec.ID = 2
	commitRec.RegionID = 5

	mock := &mockFileService{
		uploadResult: uploadRec,
		commitResult: commitRec,
		listPagination: &utils.Pagination{Page: 1, PageSize: 10, Total: 1},
		listResult: []model.FileUpload{*uploadRec},
		presignResult: &dto.PresignUploadResponse{
			UploadURL:  "https://minio.example.com/uploads/2026/07/abc.jpg?X-Amz-Signature=xxx",
			AccessURL:  "https://cdn.example.com/uploads/2026/07/abc.jpg",
			ObjectName: "uploads/2026/07/abc.jpg",
			ExpiresIn:  900,
			FileName:   "test.jpg",
			FileType:   "image",
		},
		stsResult: &dto.STSCredentialsResponse{
			AccessKeyID:     "STS.mockAccessKeyID",
			AccessKeySecret: "STS.mockAccessKeySecret",
			SecurityToken:   "STS.mockSecurityToken",
			Expiration:      "2026-07-14T12:00:00Z",
			Bucket:          "wuchang-files",
			Region:          "oss-cn-hangzhou",
			Endpoint:        "https://oss-cn-hangzhou.aliyuncs.com",
			ObjectPrefix:    "uploads/2026/07/",
			ExpiresIn:       3600,
		},
	}

	r := coreRouter.NewRouter()
	// 模拟 Auth + Region 中间件：注入 user_id 与 region_id
	r.Use(func(c *gin.Context) {
		c.Set(middleware.ContextUserID, ctxUserID)
		if ctxRegionID > 0 {
			c.Set(middleware.RegionIDKey, ctxRegionID)
		}
		c.Next()
	})

	h := fileHandler.NewHandler(mock)
	// 注册路由，路径与 file/plugin.go RegisterRoutes 保持一致（去掉权限中间件，纯测 handler）
	root := r.Group("/api/v1/file")
	root.POST("/upload", h.Upload)
	root.POST("/presign", h.Presign)
	root.POST("/commit", h.Commit)
	root.POST("/sts", h.STS)
	root.GET("", h.List)
	root.DELETE("/:id", h.Delete)

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

// doMultipart 构造 multipart/form-data 上传请求并返回响应。
// filename 为空时表示不附加 file 字段（用于测试 FormFile 失败分支）。
// file 字段名固定为 "file"（与 handler.ctx.FormFile() 内部 gin.FormFile("file") 一致）。
func (e *handlerEnv) doMultipart(t *testing.T, path, filename string, content []byte) *apiResponse {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if filename != "" {
		part, err := writer.CreateFormFile("file", filename)
		require.NoError(t, err)
		_, err = part.Write(content)
		require.NoError(t, err)
	}
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

// ---------- Upload ----------

func TestHandler_Upload_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 5) // 登录用户 + regionID=5
	resp := env.doMultipart(t, "/api/v1/file/upload", "test.jpg", []byte("fake-image-bytes"))

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "上传成功", resp.Message)
	// 透传 regionID/userID/filename/mimeType/size
	assert.Equal(t, uint(5), env.mock.lastUploadRegionID)
	assert.Equal(t, uint(1), env.mock.lastUploadUserID)
	assert.Equal(t, "test.jpg", env.mock.lastUploadFilename)
	assert.Equal(t, "image/jpeg", env.mock.lastUploadMIME) // handler.guessMIME 按 .jpg 推断
	assert.Equal(t, int64(len("fake-image-bytes")), env.mock.lastUploadSize)
	// data 透传 service 返回值
	var info struct {
		ID        uint   `json:"id"`
		FileName  string `json:"file_name"`
		FileURL   string `json:"file_url"`
		FileType  string `json:"file_type"`
	}
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, uint(1), info.ID)
	assert.Equal(t, "test.jpg", info.FileName)
	assert.Equal(t, "image", info.FileType)
}

func TestHandler_Upload_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 5) // 未登录
	resp := env.doMultipart(t, "/api/v1/file/upload", "test.jpg", []byte("x"))

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	// 未登录不应调用 service
	assert.Equal(t, uint(0), env.mock.lastUploadUserID)
}

func TestHandler_Upload_NoFile(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	// 不附加 file 字段 → ctx.FormFile() 失败
	resp := env.doMultipart(t, "/api/v1/file/upload", "", nil)

	// 业务码 CodeFileNotFound=1303 + "请上传文件"
	assert.Equal(t, 1303, resp.Code)
	assert.Equal(t, "请上传文件", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastUploadUserID)
}

func TestHandler_Upload_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.uploadErr = errors.New("文件过大")
	env.mock.uploadResult = nil
	resp := env.doMultipart(t, "/api/v1/file/upload", "big.mp4", []byte("x"))

	// 业务码 CodeFileUploadError=1302 + err.Error() 透传
	assert.Equal(t, 1302, resp.Code)
	assert.Equal(t, "文件过大", resp.Message)
	// 仍透传了 filename/size 到 service
	assert.Equal(t, "big.mp4", env.mock.lastUploadFilename)
}

func TestHandler_Upload_RegionIDInjection(t *testing.T) {
	env := newHandlerEnv(t, 7, 99) // 用户 7 + regionID=99
	resp := env.doMultipart(t, "/api/v1/file/upload", "a.png", []byte("x"))

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(99), env.mock.lastUploadRegionID)
	assert.Equal(t, uint(7), env.mock.lastUploadUserID)
}

func TestHandler_Upload_RegionIDMissing_FallbackZero(t *testing.T) {
	env := newHandlerEnv(t, 1, 0) // 登录但不注入 regionID
	resp := env.doMultipart(t, "/api/v1/file/upload", "a.png", []byte("x"))

	assert.Equal(t, 0, resp.Code)
	// file handler 无 DefaultRegionID 兜底（不同于 setting），regionID=0 透传
	assert.Equal(t, uint(0), env.mock.lastUploadRegionID)
}

// ---------- List ----------

func TestHandler_List_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/file", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "success", resp.Message)
	assert.Equal(t, uint(5), env.mock.lastListRegionID)
	require.NotNil(t, env.mock.lastListReq)
	// data 是 PageResult {list, total, page, pageSize}
	var page struct {
		List     []model.FileUpload `json:"list"`
		Total    int64              `json:"total"`
		Page     int                `json:"page"`
		PageSize int                `json:"pageSize"`
	}
	require.NoError(t, json.Unmarshal(resp.Data, &page))
	assert.Equal(t, int64(1), page.Total)
	assert.Equal(t, 1, page.Page)
	assert.Equal(t, 10, page.PageSize)
	require.Len(t, page.List, 1)
	assert.Equal(t, "test.jpg", page.List[0].FileName)
}

func TestHandler_List_WithParams(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	// query 参数透传：file_type/keyword/page/page_size
	resp := env.doJSON(t, http.MethodGet, "/api/v1/file?file_type=image&keyword=test&page=2&page_size=20", nil)

	assert.Equal(t, 0, resp.Code)
	require.NotNil(t, env.mock.lastListReq)
	assert.Equal(t, "image", env.mock.lastListReq.FileType)
	assert.Equal(t, "test", env.mock.lastListReq.Keyword)
	assert.Equal(t, 2, env.mock.lastListReq.Page)
	assert.Equal(t, 20, env.mock.lastListReq.PageSize)
}

func TestHandler_List_Empty(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.listResult = nil
	env.mock.listPagination = &utils.Pagination{Page: 1, PageSize: 10, Total: 0}
	resp := env.doJSON(t, http.MethodGet, "/api/v1/file", nil)

	assert.Equal(t, 0, resp.Code)
	var page struct {
		List  []model.FileUpload `json:"list"`
		Total int64              `json:"total"`
	}
	require.NoError(t, json.Unmarshal(resp.Data, &page))
	assert.Equal(t, int64(0), page.Total)
	assert.Empty(t, page.List)
}

func TestHandler_List_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.listErr = errors.New("db connection lost")
	env.mock.listResult = nil
	env.mock.listPagination = nil
	resp := env.doJSON(t, http.MethodGet, "/api/v1/file", nil)

	// 业务码 CodeFileError=1301 + err.Error() 透传
	assert.Equal(t, 1301, resp.Code)
	assert.Equal(t, "db connection lost", resp.Message)
}

func TestHandler_List_NoLoginRequired(t *testing.T) {
	env := newHandlerEnv(t, 0, 5) // 未登录
	resp := env.doJSON(t, http.MethodGet, "/api/v1/file", nil)

	// List 在 handler 层不校验 userID，鉴权由 plugin.go RequirePermission 负责
	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(5), env.mock.lastListRegionID)
}

// ---------- Delete ----------

func TestHandler_Delete_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/file/7", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "删除成功", resp.Message)
	assert.Equal(t, uint(7), env.mock.lastDeleteID)
}

func TestHandler_Delete_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/file/notnum", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的文件ID", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastDeleteID)
}

func TestHandler_Delete_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.deleteErr = errors.New("文件不存在")
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/file/9", nil)

	// 业务码 CodeFileError=1301 + err.Error() 透传
	assert.Equal(t, 1301, resp.Code)
	assert.Equal(t, "文件不存在", resp.Message)
	assert.Equal(t, uint(9), env.mock.lastDeleteID)
}

func TestHandler_Delete_NoLoginRequired(t *testing.T) {
	env := newHandlerEnv(t, 0, 5) // 未登录
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/file/3", nil)

	// Delete 在 handler 层不校验 userID（鉴权由 plugin.go RequirePermission 负责）
	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(3), env.mock.lastDeleteID)
}

// ---------- Presign ----------

func TestHandler_Presign_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	body := dto.PresignUploadRequest{FileName: "test.jpg"}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/file/presign", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "预签名 URL 生成成功", resp.Message)
	// 透传 regionID/userID/filename
	assert.Equal(t, uint(5), env.mock.lastPresignRegionID)
	assert.Equal(t, uint(1), env.mock.lastPresignUserID)
	assert.Equal(t, "test.jpg", env.mock.lastPresignFilename)
	// data 透传 service 返回值
	var info dto.PresignUploadResponse
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, "test.jpg", info.FileName)
	assert.Equal(t, "image", info.FileType)
	assert.Equal(t, "uploads/2026/07/abc.jpg", info.ObjectName)
	assert.Equal(t, 900, info.ExpiresIn)
}

func TestHandler_Presign_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/file/presign", dto.PresignUploadRequest{FileName: "test.jpg"})

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastPresignUserID)
}

func TestHandler_Presign_BindError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	// 非法 JSON 触发 ShouldBind 失败
	resp := env.doRaw(t, http.MethodPost, "/api/v1/file/presign", "{not json", "application/json")

	// Bind 失败或 file_name 为空统一返回 400 "请提供文件名 file_name"
	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "请提供文件名 file_name", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastPresignUserID)
}

func TestHandler_Presign_EmptyFileName(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	// file_name 缺失 → binding:"required" 失败 → 400
	resp := env.doJSON(t, http.MethodPost, "/api/v1/file/presign", map[string]string{})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "请提供文件名 file_name", resp.Message)
}

func TestHandler_Presign_WhitespaceFileName(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	// file_name 为纯空格 → binding:"required" 通过（非空字符串）→ TrimSpace 失败 → 400
	resp := env.doJSON(t, http.MethodPost, "/api/v1/file/presign", dto.PresignUploadRequest{FileName: "   "})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "请提供文件名 file_name", resp.Message)
}

func TestHandler_Presign_ErrPresignNotSupported(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.presignErr = storage.ErrPresignNotSupported
	env.mock.presignResult = nil
	resp := env.doJSON(t, http.MethodPost, "/api/v1/file/presign", dto.PresignUploadRequest{FileName: "test.jpg"})

	// 业务码 CodeFilePresignError=1306 + 降级提示
	assert.Equal(t, 1306, resp.Code)
	assert.Contains(t, resp.Message, "当前存储不支持预签名直传")
}

func TestHandler_Presign_ErrFileTypeInvalid(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.presignErr = service.ErrFileTypeInvalid
	env.mock.presignResult = nil
	resp := env.doJSON(t, http.MethodPost, "/api/v1/file/presign", dto.PresignUploadRequest{FileName: "malware.exe"})

	// 业务码 CodeFileTypeInvalid=1305 + err.Error() 透传
	assert.Equal(t, 1305, resp.Code)
	assert.Equal(t, service.ErrFileTypeInvalid.Error(), resp.Message)
}

func TestHandler_Presign_OtherError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.presignErr = errors.New("签名服务不可用")
	env.mock.presignResult = nil
	resp := env.doJSON(t, http.MethodPost, "/api/v1/file/presign", dto.PresignUploadRequest{FileName: "test.jpg"})

	// 业务码 CodeFileUploadError=1302 + err.Error() 透传
	assert.Equal(t, 1302, resp.Code)
	assert.Equal(t, "签名服务不可用", resp.Message)
}

// ---------- Commit ----------

func TestHandler_Commit_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	body := dto.CommitUploadRequest{
		FileName:   "photo.png",
		ObjectName: "uploads/2026/07/abc.png",
		MimeType:   "image/png",
		FileSize:   2048,
	}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/file/commit", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "文件记录已保存", resp.Message)
	// 透传所有参数
	assert.Equal(t, uint(5), env.mock.lastCommitRegionID)
	assert.Equal(t, uint(1), env.mock.lastCommitUserID)
	assert.Equal(t, "photo.png", env.mock.lastCommitFilename)
	assert.Equal(t, "uploads/2026/07/abc.png", env.mock.lastCommitObject)
	assert.Equal(t, "image/png", env.mock.lastCommitMIME)
	assert.Equal(t, int64(2048), env.mock.lastCommitSize)
	// data 透传 service 返回值
	var info struct {
		ID       uint   `json:"id"`
		FileName string `json:"file_name"`
		FileURL  string `json:"file_url"`
	}
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, uint(2), info.ID)
	assert.Equal(t, "photo.png", info.FileName)
}

func TestHandler_Commit_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	body := dto.CommitUploadRequest{FileName: "x.png", ObjectName: "uploads/x.png", FileSize: 1}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/file/commit", body)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, uint(0), env.mock.lastCommitUserID)
}

func TestHandler_Commit_BindError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/file/commit", "{bad json", "application/json")

	// Bind 失败 → 400 "请求参数无效"
	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "请求参数无效", resp.Message)
}

func TestHandler_Commit_MissingRequiredFields(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	// 缺 object_name → binding:"required" 失败 → 400 "请求参数无效"
	resp := env.doJSON(t, http.MethodPost, "/api/v1/file/commit", dto.CommitUploadRequest{
		FileName: "x.png",
		FileSize: 1,
	})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "请求参数无效", resp.Message)
}

func TestHandler_Commit_WhitespaceObjectName(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	// object_name 为纯空格 → binding 通过 → TrimSpace 失败 → 400 "object_name 不能为空"
	resp := env.doJSON(t, http.MethodPost, "/api/v1/file/commit", dto.CommitUploadRequest{
		FileName:   "x.png",
		ObjectName: "   ",
		FileSize:   1,
	})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "object_name 不能为空", resp.Message)
}

func TestHandler_Commit_ErrPresignNotSupported(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.commitErr = storage.ErrPresignNotSupported
	env.mock.commitResult = nil
	body := dto.CommitUploadRequest{FileName: "x.png", ObjectName: "uploads/x.png", FileSize: 1}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/file/commit", body)

	// 业务码 CodeFilePresignError=1306 + 降级提示
	assert.Equal(t, 1306, resp.Code)
	assert.Contains(t, resp.Message, "当前存储不支持预签名直传")
}

func TestHandler_Commit_ErrFileTypeInvalid(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.commitErr = service.ErrFileTypeInvalid
	env.mock.commitResult = nil
	body := dto.CommitUploadRequest{FileName: "malware.exe", ObjectName: "uploads/x.exe", FileSize: 1}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/file/commit", body)

	// 业务码 CodeFileTypeInvalid=1305
	assert.Equal(t, 1305, resp.Code)
	assert.Equal(t, service.ErrFileTypeInvalid.Error(), resp.Message)
}

func TestHandler_Commit_ErrFileTooLarge(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.commitErr = service.ErrFileTooLarge
	env.mock.commitResult = nil
	body := dto.CommitUploadRequest{FileName: "big.mp4", ObjectName: "uploads/big.mp4", FileSize: 100*1024*1024 + 1}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/file/commit", body)

	// 业务码 CodeFileTooLarge=1304
	assert.Equal(t, 1304, resp.Code)
	assert.Equal(t, service.ErrFileTooLarge.Error(), resp.Message)
}

func TestHandler_Commit_ErrFileEmpty(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.commitErr = service.ErrFileEmpty
	env.mock.commitResult = nil
	// FileSize 必须非 0 以通过 binding:"required"（service 内部 size<=0 才返回 ErrFileEmpty，
	// 此处 mock 直接返回 ErrFileEmpty 验证 handler 错误码路由，与 service 实际校验逻辑解耦）
	body := dto.CommitUploadRequest{FileName: "empty.png", ObjectName: "uploads/empty.png", FileSize: 1}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/file/commit", body)

	// 业务码 CodeFileError=1301（CommitUpload 的 ErrFileEmpty 分支映射到 CodeFileError）
	assert.Equal(t, 1301, resp.Code)
	assert.Equal(t, service.ErrFileEmpty.Error(), resp.Message)
}

func TestHandler_Commit_OtherError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.commitErr = errors.New("数据库写入失败")
	env.mock.commitResult = nil
	body := dto.CommitUploadRequest{FileName: "x.png", ObjectName: "uploads/x.png", FileSize: 1}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/file/commit", body)

	// 业务码 CodeFileUploadError=1302 + err.Error() 透传
	assert.Equal(t, 1302, resp.Code)
	assert.Equal(t, "数据库写入失败", resp.Message)
}

// ---------- STS ----------

func TestHandler_STS_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/file/sts", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "STS 临时凭据获取成功", resp.Message)
	// 透传 regionID/userID
	assert.Equal(t, uint(5), env.mock.lastSTSRegionID)
	assert.Equal(t, uint(1), env.mock.lastSTSUserID)
	// data 透传 service 返回值
	var info dto.STSCredentialsResponse
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, "STS.mockAccessKeyID", info.AccessKeyID)
	assert.Equal(t, "wuchang-files", info.Bucket)
	assert.Equal(t, "oss-cn-hangzhou", info.Region)
	assert.Equal(t, 3600, info.ExpiresIn)
}

func TestHandler_STS_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/file/sts", nil)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastSTSUserID)
}

func TestHandler_STS_ErrNotConfigured(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.stsErr = sts.ErrNotConfigured
	env.mock.stsResult = nil
	resp := env.doJSON(t, http.MethodPost, "/api/v1/file/sts", nil)

	// 业务码 CodeFileSTSError=1307 + 降级提示
	assert.Equal(t, 1307, resp.Code)
	assert.Contains(t, resp.Message, "STS 未配置")
}

func TestHandler_STS_OtherError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.stsErr = errors.New("AssumeRole 调用超时")
	env.mock.stsResult = nil
	resp := env.doJSON(t, http.MethodPost, "/api/v1/file/sts", nil)

	// 业务码 CodeFileSTSError=1307 + err.Error() 透传
	assert.Equal(t, 1307, resp.Code)
	assert.Equal(t, "AssumeRole 调用超时", resp.Message)
}

// ---------- regionID 注入与公开读取 ----------

// TestHandler_RegionIDInjection_AllWriteOps 验证所有写操作（Upload/Presign/Commit/STS）
// 均从上下文读取 regionID 并透传给 service。
func TestHandler_RegionIDInjection_AllWriteOps(t *testing.T) {
	env := newHandlerEnv(t, 8, 42) // 用户 8 + regionID=42

	// Upload
	resp := env.doMultipart(t, "/api/v1/file/upload", "a.png", []byte("x"))
	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(42), env.mock.lastUploadRegionID)

	// Presign
	resp = env.doJSON(t, http.MethodPost, "/api/v1/file/presign", dto.PresignUploadRequest{FileName: "a.png"})
	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(42), env.mock.lastPresignRegionID)

	// Commit
	resp = env.doJSON(t, http.MethodPost, "/api/v1/file/commit", dto.CommitUploadRequest{
		FileName: "a.png", ObjectName: "uploads/a.png", FileSize: 1,
	})
	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(42), env.mock.lastCommitRegionID)

	// STS
	resp = env.doJSON(t, http.MethodPost, "/api/v1/file/sts", nil)
	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(42), env.mock.lastSTSRegionID)

	// List
	resp = env.doJSON(t, http.MethodGet, "/api/v1/file", nil)
	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(42), env.mock.lastListRegionID)
}

// TestHandler_PublicRead_NoAuthRequired 验证 file 的读操作（List/Delete）在 handler 层不校验 userID，
// 即使未登录（userID=0）也能正常调用 service。鉴权由 plugin.go 的 RequirePermission 中间件负责，
// handler 层仅写操作做防御性登录校验。
func TestHandler_PublicRead_NoAuthRequired(t *testing.T) {
	env := newHandlerEnv(t, 0, 5) // 未登录
	// List 不应被 401 拦截
	resp := env.doJSON(t, http.MethodGet, "/api/v1/file", nil)
	assert.Equal(t, 0, resp.Code)

	// Delete 不应被 401 拦截（handler 无登录校验）
	resp = env.doJSON(t, http.MethodDelete, "/api/v1/file/1", nil)
	assert.Equal(t, 0, resp.Code)
}
