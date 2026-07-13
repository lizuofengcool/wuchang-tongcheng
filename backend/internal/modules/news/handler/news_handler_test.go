// Package handler_test 同城头条模块 HTTP 处理层单元测试。
//
// 使用 gin + httptest + mock NewsService，覆盖 handler 全部分支：
//   - 写操作未登录拦截（userID=0 → 401）：Create/Update/Delete/Like/Fav/CreateComment/DeleteComment/ListFavorites/ListMessages/MarkRead
//   - 只读状态接口未登录返回默认零值（不拦截）：LikeStatus/FavStatus/UnreadCount
//   - URL :id 参数解析失败（非数字 → 400）
//   - 请求体 Bind 失败（非法 JSON → 400）
//   - 业务参数校验（Search 空关键词 → 400；Nearby 经纬度越界 → 400）
//   - service 成功/错误透传（业务码 CodeNewsPublishError=2403/CodeNewsError=2401/CodeNewsNotFound=2402 + message + data 透传）
//   - 地区ID 上下文注入（regionID 透传给 service）
//   - query 参数兜底（page/page_size 缺失默认 1/10 或 1/20）
//   - 点赞/收藏 toggle 分支（Liked/Faved true/false 不同 message）
//
// 不依赖 DB/Redis/Docker/ES/MQ，纯内存 mock service 验证 handler 装配层逻辑。
// 与 category/region handler 测试同风格，区别在于 news 模块方法更多（18 个）且包含
// 点赞/收藏/评论/消息四组业务，是项目最复杂的模块。
package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wuchang-tongcheng/internal/core/middleware"
	coreRouter "wuchang-tongcheng/internal/core/router"
	"wuchang-tongcheng/internal/modules/news/dto"
	newsHandler "wuchang-tongcheng/internal/modules/news/handler"
	"wuchang-tongcheng/internal/modules/news/service"
	"wuchang-tongcheng/internal/pkg/utils"

	"errors"
)

// apiResponse 解析统一响应体 {code, message, data}
type apiResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// mockNewsService 内存 mock，实现 service.NewsService 接口。
// 记录最近一次调用入参，返回预设 result/err，便于断言 handler 装配层逻辑。
type mockNewsService struct {
	// 调用记录
	lastCreateRegionID   uint
	lastCreateAuthorID   uint
	lastCreateAuthorName string
	lastCreateReq        *dto.CreateNewsRequest

	lastUpdateID   uint
	lastUpdateOpID uint
	lastUpdateReq  *dto.UpdateNewsRequest

	lastDeleteID   uint
	lastDeleteOpID uint

	lastGetByID uint

	lastListRegionID uint
	lastListReq      *dto.NewsListRequest

	lastNearbyRegionID uint
	lastNearbyReq      *dto.NewsNearbyRequest

	lastSearchRegionID uint
	lastSearchReq      *dto.NewsSearchRequest

	lastLikeUserID uint
	lastLikeNewsID uint

	lastLikeStatusUserID uint
	lastLikeStatusNewsID uint

	lastFavUserID uint
	lastFavNewsID uint

	lastFavStatusUserID uint
	lastFavStatusNewsID uint

	lastListFavoritesUserID   uint
	lastListFavoritesPage     int
	lastListFavoritesPageSize int

	lastCreateCommentNewsID    uint
	lastCreateCommentUserID    uint
	lastCreateCommentUserName  string
	lastCreateCommentAvatar    string
	lastCreateCommentReq       *dto.CreateCommentRequest

	lastListCommentsNewsID   uint
	lastListCommentsPage     int
	lastListCommentsPageSize int

	lastDeleteCommentID    uint
	lastDeleteCommentUserID uint

	lastListMessagesUserID   uint
	lastListMessagesPage     int
	lastListMessagesPageSize int

	lastUnreadCountUserID uint

	lastMarkReadUserID uint
	lastMarkReadIDs    []uint

	// 返回值预设
	createResult        *dto.NewsInfo
	createErr           error
	updateErr           error
	deleteErr           error
	getByIDResult       *dto.NewsInfo
	getByIDErr          error
	listPagination      *utils.Pagination
	listResult          []dto.NewsInfo
	listErr             error
	nearbyPagination    *utils.Pagination
	nearbyResult        []dto.NewsInfo
	nearbyErr           error
	searchPagination    *utils.Pagination
	searchResult        []dto.NewsInfo
	searchErr           error
	likeResult          *dto.LikeResponse
	likeErr             error
	likeStatusResult    *dto.LikeResponse
	likeStatusErr       error
	favResult           *dto.FavResponse
	favErr              error
	favStatusResult     *dto.FavResponse
	favStatusErr        error
	listFavoritesPagination *utils.Pagination
	listFavoritesResult     []dto.NewsInfo
	listFavoritesErr        error
	createCommentResult *dto.CommentInfo
	createCommentErr    error
	listCommentsResult  []dto.CommentInfo
	listCommentsTotal   int64
	listCommentsErr     error
	deleteCommentErr    error
	listMessagesResult  []dto.MessageInfo
	listMessagesTotal   int64
	listMessagesErr     error
	unreadCountResult   int64
	unreadCountErr      error
	markReadErr         error
}

func (m *mockNewsService) Create(regionID uint, authorID uint, authorName string, req *dto.CreateNewsRequest) (*dto.NewsInfo, error) {
	m.lastCreateRegionID = regionID
	m.lastCreateAuthorID = authorID
	m.lastCreateAuthorName = authorName
	m.lastCreateReq = req
	return m.createResult, m.createErr
}
func (m *mockNewsService) Update(id uint, operatorID uint, req *dto.UpdateNewsRequest) error {
	m.lastUpdateID = id
	m.lastUpdateOpID = operatorID
	m.lastUpdateReq = req
	return m.updateErr
}
func (m *mockNewsService) Delete(id uint, operatorID uint) error {
	m.lastDeleteID = id
	m.lastDeleteOpID = operatorID
	return m.deleteErr
}
func (m *mockNewsService) GetByID(id uint) (*dto.NewsInfo, error) {
	m.lastGetByID = id
	return m.getByIDResult, m.getByIDErr
}
func (m *mockNewsService) List(regionID uint, req *dto.NewsListRequest) (*utils.Pagination, []dto.NewsInfo, error) {
	m.lastListRegionID = regionID
	m.lastListReq = req
	return m.listPagination, m.listResult, m.listErr
}
func (m *mockNewsService) ListNearby(regionID uint, req *dto.NewsNearbyRequest) (*utils.Pagination, []dto.NewsInfo, error) {
	m.lastNearbyRegionID = regionID
	m.lastNearbyReq = req
	return m.nearbyPagination, m.nearbyResult, m.nearbyErr
}
func (m *mockNewsService) Search(regionID uint, req *dto.NewsSearchRequest) (*utils.Pagination, []dto.NewsInfo, error) {
	m.lastSearchRegionID = regionID
	m.lastSearchReq = req
	return m.searchPagination, m.searchResult, m.searchErr
}
func (m *mockNewsService) Like(userID, newsID uint) (*dto.LikeResponse, error) {
	m.lastLikeUserID = userID
	m.lastLikeNewsID = newsID
	return m.likeResult, m.likeErr
}
func (m *mockNewsService) LikeStatus(userID, newsID uint) (*dto.LikeResponse, error) {
	m.lastLikeStatusUserID = userID
	m.lastLikeStatusNewsID = newsID
	return m.likeStatusResult, m.likeStatusErr
}
func (m *mockNewsService) Fav(userID, newsID uint) (*dto.FavResponse, error) {
	m.lastFavUserID = userID
	m.lastFavNewsID = newsID
	return m.favResult, m.favErr
}
func (m *mockNewsService) FavStatus(userID, newsID uint) (*dto.FavResponse, error) {
	m.lastFavStatusUserID = userID
	m.lastFavStatusNewsID = newsID
	return m.favStatusResult, m.favStatusErr
}
func (m *mockNewsService) ListFavorites(userID uint, page, pageSize int) (*utils.Pagination, []dto.NewsInfo, error) {
	m.lastListFavoritesUserID = userID
	m.lastListFavoritesPage = page
	m.lastListFavoritesPageSize = pageSize
	return m.listFavoritesPagination, m.listFavoritesResult, m.listFavoritesErr
}
func (m *mockNewsService) CreateComment(newsID uint, userID uint, userName string, avatar string, req *dto.CreateCommentRequest) (*dto.CommentInfo, error) {
	m.lastCreateCommentNewsID = newsID
	m.lastCreateCommentUserID = userID
	m.lastCreateCommentUserName = userName
	m.lastCreateCommentAvatar = avatar
	m.lastCreateCommentReq = req
	return m.createCommentResult, m.createCommentErr
}
func (m *mockNewsService) ListComments(newsID uint, page, pageSize int) ([]dto.CommentInfo, int64, error) {
	m.lastListCommentsNewsID = newsID
	m.lastListCommentsPage = page
	m.lastListCommentsPageSize = pageSize
	return m.listCommentsResult, m.listCommentsTotal, m.listCommentsErr
}
func (m *mockNewsService) DeleteComment(id uint, userID uint) error {
	m.lastDeleteCommentID = id
	m.lastDeleteCommentUserID = userID
	return m.deleteCommentErr
}
func (m *mockNewsService) ListMessages(userID uint, page, pageSize int) ([]dto.MessageInfo, int64, error) {
	m.lastListMessagesUserID = userID
	m.lastListMessagesPage = page
	m.lastListMessagesPageSize = pageSize
	return m.listMessagesResult, m.listMessagesTotal, m.listMessagesErr
}
func (m *mockNewsService) UnreadCount(userID uint) (int64, error) {
	m.lastUnreadCountUserID = userID
	return m.unreadCountResult, m.unreadCountErr
}
func (m *mockNewsService) MarkRead(userID uint, ids []uint) error {
	m.lastMarkReadUserID = userID
	m.lastMarkReadIDs = ids
	return m.markReadErr
}

// 确保 mockNewsService 实现 service.NewsService 接口
var _ service.NewsService = (*mockNewsService)(nil)

// newsHandlerEnv handler 测试环境
type newsHandlerEnv struct {
	engine *gin.Engine
	mock   *mockNewsService
}

// newNewsHandlerEnv 构造 gin 引擎并注册 news 路由（与 news/plugin.go RegisterRoutes 路径一致）。
// ctxUserID 用于模拟 Auth 中间件注入的 user_id（0 表示未登录）。
// regionID 用于模拟 Region 中间件注入的 region_id。
func newNewsHandlerEnv(t *testing.T, ctxUserID uint, regionID uint) *newsHandlerEnv {
	t.Helper()
	gin.SetMode(gin.TestMode)

	mock := &mockNewsService{
		createResult:        &dto.NewsInfo{ID: 1, Title: "二手手机", AuthorID: ctxUserID, AuthorName: "tester", Status: 1, RegionID: regionID},
		getByIDResult:       &dto.NewsInfo{ID: 1, Title: "二手手机", Status: 1},
		listPagination:      &utils.Pagination{Page: 1, PageSize: 10, Total: 1},
		listResult:          []dto.NewsInfo{{ID: 1, Title: "二手手机", Status: 1}},
		nearbyPagination:    &utils.Pagination{Page: 1, PageSize: 10, Total: 1},
		nearbyResult:        []dto.NewsInfo{{ID: 1, Title: "二手手机", Distance: 1.2}},
		searchPagination:    &utils.Pagination{Page: 1, PageSize: 10, Total: 1},
		searchResult:        []dto.NewsInfo{{ID: 1, Title: "二手手机"}},
		likeResult:          &dto.LikeResponse{Liked: true, LikeCount: 1},
		likeStatusResult:    &dto.LikeResponse{Liked: false, LikeCount: 0},
		favResult:           &dto.FavResponse{Faved: true, FavCount: 1},
		favStatusResult:     &dto.FavResponse{Faved: false, FavCount: 0},
		listFavoritesPagination: &utils.Pagination{Page: 1, PageSize: 10, Total: 1},
		listFavoritesResult:     []dto.NewsInfo{{ID: 1, Title: "二手手机"}},
		createCommentResult: &dto.CommentInfo{ID: 9, NewsID: 1, UserID: ctxUserID, UserName: "tester", Content: "好货"},
		listCommentsResult:  []dto.CommentInfo{{ID: 9, NewsID: 1, UserName: "tester", Content: "好货"}},
		listCommentsTotal:   1,
		listMessagesResult:  []dto.MessageInfo{{ID: 1, Type: "like", Content: "有人赞了你的信息"}},
		listMessagesTotal:   1,
		unreadCountResult:   3,
	}

	r := coreRouter.NewRouter()
	// 模拟 Region + Auth 中间件：注入 region_id、user_id、username
	r.Use(func(c *gin.Context) {
		c.Set(middleware.RegionIDKey, regionID)
		c.Set(middleware.ContextUserID, ctxUserID)
		c.Set(middleware.ContextUsername, "tester")
		c.Next()
	})

	h := newsHandler.NewHandler(mock)
	// 注册路由，路径与 news/plugin.go RegisterRoutes 保持一致（去掉权限/限流中间件，纯测 handler）
	// 注意：/nearby、/favorites、/search、/messages 需注册在 /:id 之前避免被参数路由吞掉
	root := r.Group("/api/v1/news")
	root.GET("", h.List)
	root.GET("/search", h.Search)
	root.GET("/nearby", h.Nearby)
	root.GET("/favorites", h.ListFavorites)
	root.GET("/messages", h.ListMessages)
	root.GET("/messages/unread", h.UnreadCount)
	root.GET("/:id", h.GetByID)
	root.GET("/:id/comments", h.ListComments)
	root.GET("/:id/like", h.LikeStatus)
	root.GET("/:id/fav", h.FavStatus)
	root.POST("", h.Create)
	root.PUT("/:id", h.Update)
	root.DELETE("/:id", h.Delete)
	root.POST("/:id/like", h.Like)
	root.POST("/:id/fav", h.Fav)
	root.POST("/:id/comments", h.CreateComment)
	root.DELETE("/comments/:id", h.DeleteComment)
	root.PUT("/messages/read", h.MarkRead)

	return &newsHandlerEnv{engine: r.Engine(), mock: mock}
}

// doJSON 发起 JSON 请求，返回解析后的响应体。
func (e *newsHandlerEnv) doJSON(t *testing.T, method, path string, body interface{}) *apiResponse {
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
func (e *newsHandlerEnv) doRaw(t *testing.T, method, path string, rawBody string, contentType string) *apiResponse {
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

// pageResultData 解析 utils.PageResult 的 data 字段（{list, total, page, page_size}）
type pageResultData struct {
	List     json.RawMessage `json:"list"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
}

// ---------- Create ----------

func TestNewsHandler_Create_Success(t *testing.T) {
	env := newNewsHandlerEnv(t, 5, 2)
	// ListingType/Condition 必须满足 binding:"oneof"；Status 满足 oneof=0 1
	body := dto.CreateNewsRequest{
		Title: "二手手机", Content: "九成新", CategoryID: 1,
		ListingType: "sell", Condition: "used", Price: 999, PriceUnit: "元", Status: 1,
	}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/news", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "发布成功", resp.Message)
	// 透传 regionID/authorID/authorName
	assert.Equal(t, uint(2), env.mock.lastCreateRegionID)
	assert.Equal(t, uint(5), env.mock.lastCreateAuthorID)
	assert.Equal(t, "tester", env.mock.lastCreateAuthorName)
	// 透传请求体
	require.NotNil(t, env.mock.lastCreateReq)
	assert.Equal(t, "二手手机", env.mock.lastCreateReq.Title)
	// data 透传 service 返回值
	var info dto.NewsInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, uint(1), info.ID)
	assert.Equal(t, "二手手机", info.Title)
}

func TestNewsHandler_Create_NotLoggedIn(t *testing.T) {
	env := newNewsHandlerEnv(t, 0, 2) // 未登录
	body := dto.CreateNewsRequest{Title: "x", ListingType: "sell", Condition: "used"}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/news", body)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	// 未登录不应调用 service
	assert.Nil(t, env.mock.lastCreateReq)
}

func TestNewsHandler_Create_BindError(t *testing.T) {
	env := newNewsHandlerEnv(t, 1, 2)
	// 非法 JSON 触发 ShouldBind 失败
	resp := env.doRaw(t, http.MethodPost, "/api/v1/news", "{not json", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
	assert.Nil(t, env.mock.lastCreateReq)
}

func TestNewsHandler_Create_ServiceError(t *testing.T) {
	env := newNewsHandlerEnv(t, 1, 2)
	env.mock.createErr = errors.New("标题重复")
	env.mock.createResult = nil
	body := dto.CreateNewsRequest{Title: "x", ListingType: "sell", Condition: "used"}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/news", body)

	// 业务码 CodeNewsPublishError=2403 + err.Error() 透传
	assert.Equal(t, 2403, resp.Code)
	assert.Equal(t, "标题重复", resp.Message)
	assert.Equal(t, uint(2), env.mock.lastCreateRegionID)
}

// ---------- Update ----------

func TestNewsHandler_Update_Success(t *testing.T) {
	env := newNewsHandlerEnv(t, 5, 2)
	body := dto.UpdateNewsRequest{Title: "降价", Price: 888}
	resp := env.doJSON(t, http.MethodPut, "/api/v1/news/7", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "更新成功", resp.Message)
	assert.Equal(t, uint(7), env.mock.lastUpdateID)
	assert.Equal(t, uint(5), env.mock.lastUpdateOpID)
	require.NotNil(t, env.mock.lastUpdateReq)
	assert.Equal(t, "降价", env.mock.lastUpdateReq.Title)
}

func TestNewsHandler_Update_NotLoggedIn(t *testing.T) {
	env := newNewsHandlerEnv(t, 0, 2)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/news/7", dto.UpdateNewsRequest{Title: "x"})

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, uint(0), env.mock.lastUpdateID)
}

func TestNewsHandler_Update_InvalidID(t *testing.T) {
	env := newNewsHandlerEnv(t, 1, 2)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/news/abc", dto.UpdateNewsRequest{Title: "x"})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastUpdateID)
}

func TestNewsHandler_Update_ServiceError(t *testing.T) {
	env := newNewsHandlerEnv(t, 1, 2)
	env.mock.updateErr = errors.New("无权操作此信息")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/news/7", dto.UpdateNewsRequest{Title: "x"})

	// 业务码 CodeNewsError=2401
	assert.Equal(t, 2401, resp.Code)
	assert.Equal(t, "无权操作此信息", resp.Message)
}

// ---------- Delete ----------

func TestNewsHandler_Delete_Success(t *testing.T) {
	env := newNewsHandlerEnv(t, 5, 2)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/news/9", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "删除成功", resp.Message)
	assert.Equal(t, uint(9), env.mock.lastDeleteID)
	assert.Equal(t, uint(5), env.mock.lastDeleteOpID)
}

func TestNewsHandler_Delete_NotLoggedIn(t *testing.T) {
	env := newNewsHandlerEnv(t, 0, 2)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/news/9", nil)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, uint(0), env.mock.lastDeleteID)
}

func TestNewsHandler_Delete_InvalidID(t *testing.T) {
	env := newNewsHandlerEnv(t, 1, 2)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/news/xyz", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestNewsHandler_Delete_ServiceError(t *testing.T) {
	env := newNewsHandlerEnv(t, 1, 2)
	env.mock.deleteErr = errors.New("信息不存在")
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/news/9", nil)

	assert.Equal(t, 2401, resp.Code)
	assert.Equal(t, "信息不存在", resp.Message)
}

// ---------- GetByID ----------

func TestNewsHandler_GetByID_Success(t *testing.T) {
	env := newNewsHandlerEnv(t, 1, 2)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/news/3", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "success", resp.Message)
	assert.Equal(t, uint(3), env.mock.lastGetByID)
	var info dto.NewsInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, uint(1), info.ID)
}

func TestNewsHandler_GetByID_InvalidID(t *testing.T) {
	env := newNewsHandlerEnv(t, 1, 2)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/news/notnum", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestNewsHandler_GetByID_ServiceError(t *testing.T) {
	env := newNewsHandlerEnv(t, 1, 2)
	env.mock.getByIDErr = errors.New("信息不存在")
	env.mock.getByIDResult = nil
	resp := env.doJSON(t, http.MethodGet, "/api/v1/news/999", nil)

	// CodeNewsNotFound=2402
	assert.Equal(t, 2402, resp.Code)
	assert.Equal(t, "信息不存在", resp.Message)
}

// ---------- List ----------

func TestNewsHandler_List_Success(t *testing.T) {
	env := newNewsHandlerEnv(t, 1, 2)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/news?page=1&page_size=10", nil)

	assert.Equal(t, 0, resp.Code)
	// 透传 regionID
	assert.Equal(t, uint(2), env.mock.lastListRegionID)
	require.NotNil(t, env.mock.lastListReq)
	assert.Equal(t, 1, env.mock.lastListReq.Page)
	// data 为 PageResult
	var pr pageResultData
	require.NoError(t, json.Unmarshal(resp.Data, &pr))
	assert.Equal(t, int64(1), pr.Total)
	var list []dto.NewsInfo
	require.NoError(t, json.Unmarshal(pr.List, &list))
	require.Len(t, list, 1)
	assert.Equal(t, "二手手机", list[0].Title)
}

func TestNewsHandler_List_ServiceError(t *testing.T) {
	env := newNewsHandlerEnv(t, 1, 2)
	env.mock.listErr = errors.New("db down")
	env.mock.listResult = nil
	resp := env.doJSON(t, http.MethodGet, "/api/v1/news", nil)

	assert.Equal(t, 2401, resp.Code)
	assert.Equal(t, "db down", resp.Message)
}

// ---------- Search ----------

func TestNewsHandler_Search_EmptyKeyword(t *testing.T) {
	env := newNewsHandlerEnv(t, 1, 2)
	// 不传 keyword → handler 校验返回 400
	resp := env.doJSON(t, http.MethodGet, "/api/v1/news/search", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "关键词不能为空", resp.Message)
	assert.Nil(t, env.mock.lastSearchReq)
}

func TestNewsHandler_Search_Success(t *testing.T) {
	env := newNewsHandlerEnv(t, 1, 2)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/news/search?keyword=手机&page=1", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(2), env.mock.lastSearchRegionID)
	require.NotNil(t, env.mock.lastSearchReq)
	assert.Equal(t, "手机", env.mock.lastSearchReq.Keyword)
	var pr pageResultData
	require.NoError(t, json.Unmarshal(resp.Data, &pr))
	assert.Equal(t, int64(1), pr.Total)
}

func TestNewsHandler_Search_ServiceError(t *testing.T) {
	env := newNewsHandlerEnv(t, 1, 2)
	env.mock.searchErr = errors.New("es down")
	env.mock.searchResult = nil
	resp := env.doJSON(t, http.MethodGet, "/api/v1/news/search?keyword=手机", nil)

	assert.Equal(t, 2401, resp.Code)
	assert.Equal(t, "es down", resp.Message)
}

// ---------- Nearby ----------

func TestNewsHandler_Nearby_InvalidLatLon(t *testing.T) {
	env := newNewsHandlerEnv(t, 1, 2)
	// 经纬度越界（纬度 999）→ 400
	resp := env.doJSON(t, http.MethodGet, "/api/v1/news/nearby?latitude=999&longitude=114", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "经纬度参数无效", resp.Message)
	assert.Nil(t, env.mock.lastNearbyReq)
}

func TestNewsHandler_Nearby_Success(t *testing.T) {
	env := newNewsHandlerEnv(t, 1, 2)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/news/nearby?latitude=30.5&longitude=114.3&radius_km=5", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(2), env.mock.lastNearbyRegionID)
	require.NotNil(t, env.mock.lastNearbyReq)
	assert.Equal(t, 30.5, env.mock.lastNearbyReq.Latitude)
	assert.Equal(t, 114.3, env.mock.lastNearbyReq.Longitude)
	assert.Equal(t, 5.0, env.mock.lastNearbyReq.RadiusKm)
	// 附近查询结果携带 distance 字段
	var pr pageResultData
	require.NoError(t, json.Unmarshal(resp.Data, &pr))
	var list []dto.NewsInfo
	require.NoError(t, json.Unmarshal(pr.List, &list))
	require.Len(t, list, 1)
	assert.Equal(t, 1.2, list[0].Distance)
}

func TestNewsHandler_Nearby_ServiceError(t *testing.T) {
	env := newNewsHandlerEnv(t, 1, 2)
	env.mock.nearbyErr = errors.New("postgis unavailable")
	env.mock.nearbyResult = nil
	resp := env.doJSON(t, http.MethodGet, "/api/v1/news/nearby?latitude=30.5&longitude=114.3", nil)

	assert.Equal(t, 2401, resp.Code)
	assert.Equal(t, "postgis unavailable", resp.Message)
}

// ---------- Like ----------

func TestNewsHandler_Like_Liked(t *testing.T) {
	env := newNewsHandlerEnv(t, 5, 2)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/news/8/like", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "点赞成功", resp.Message)
	assert.Equal(t, uint(5), env.mock.lastLikeUserID)
	assert.Equal(t, uint(8), env.mock.lastLikeNewsID)
	var lr dto.LikeResponse
	require.NoError(t, json.Unmarshal(resp.Data, &lr))
	assert.True(t, lr.Liked)
	assert.Equal(t, 1, lr.LikeCount)
}

func TestNewsHandler_Like_UnLike(t *testing.T) {
	env := newNewsHandlerEnv(t, 5, 2)
	// 预设 service 返回取消点赞
	env.mock.likeResult = &dto.LikeResponse{Liked: false, LikeCount: 0}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/news/8/like", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "已取消点赞", resp.Message)
	var lr dto.LikeResponse
	require.NoError(t, json.Unmarshal(resp.Data, &lr))
	assert.False(t, lr.Liked)
}

func TestNewsHandler_Like_NotLoggedIn(t *testing.T) {
	env := newNewsHandlerEnv(t, 0, 2)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/news/8/like", nil)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, uint(0), env.mock.lastLikeNewsID)
}

func TestNewsHandler_Like_InvalidID(t *testing.T) {
	env := newNewsHandlerEnv(t, 1, 2)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/news/abc/like", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestNewsHandler_Like_ServiceError(t *testing.T) {
	env := newNewsHandlerEnv(t, 1, 2)
	env.mock.likeErr = errors.New("信息不存在")
	env.mock.likeResult = nil
	resp := env.doJSON(t, http.MethodPost, "/api/v1/news/8/like", nil)

	assert.Equal(t, 2401, resp.Code)
	assert.Equal(t, "信息不存在", resp.Message)
}

// ---------- LikeStatus ----------

func TestNewsHandler_LikeStatus_NotLoggedIn(t *testing.T) {
	env := newNewsHandlerEnv(t, 0, 2)
	// 未登录不拦截，返回默认零值 {liked:false, like_count:0}
	resp := env.doJSON(t, http.MethodGet, "/api/v1/news/8/like", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(0), env.mock.lastLikeStatusNewsID)
	var lr dto.LikeResponse
	require.NoError(t, json.Unmarshal(resp.Data, &lr))
	assert.False(t, lr.Liked)
	assert.Equal(t, 0, lr.LikeCount)
}

func TestNewsHandler_LikeStatus_Success(t *testing.T) {
	env := newNewsHandlerEnv(t, 5, 2)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/news/8/like", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(5), env.mock.lastLikeStatusUserID)
	assert.Equal(t, uint(8), env.mock.lastLikeStatusNewsID)
}

func TestNewsHandler_LikeStatus_ServiceError(t *testing.T) {
	env := newNewsHandlerEnv(t, 1, 2)
	env.mock.likeStatusErr = errors.New("信息不存在")
	env.mock.likeStatusResult = nil
	resp := env.doJSON(t, http.MethodGet, "/api/v1/news/8/like", nil)

	// CodeNewsNotFound=2402
	assert.Equal(t, 2402, resp.Code)
	assert.Equal(t, "信息不存在", resp.Message)
}

// ---------- Fav ----------

func TestNewsHandler_Fav_Faved(t *testing.T) {
	env := newNewsHandlerEnv(t, 5, 2)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/news/8/fav", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "收藏成功", resp.Message)
	assert.Equal(t, uint(5), env.mock.lastFavUserID)
	assert.Equal(t, uint(8), env.mock.lastFavNewsID)
	var fr dto.FavResponse
	require.NoError(t, json.Unmarshal(resp.Data, &fr))
	assert.True(t, fr.Faved)
	assert.Equal(t, 1, fr.FavCount)
}

func TestNewsHandler_Fav_UnFav(t *testing.T) {
	env := newNewsHandlerEnv(t, 5, 2)
	env.mock.favResult = &dto.FavResponse{Faved: false, FavCount: 0}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/news/8/fav", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "已取消收藏", resp.Message)
}

func TestNewsHandler_Fav_NotLoggedIn(t *testing.T) {
	env := newNewsHandlerEnv(t, 0, 2)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/news/8/fav", nil)

	assert.Equal(t, 401, resp.Code)
}

func TestNewsHandler_Fav_ServiceError(t *testing.T) {
	env := newNewsHandlerEnv(t, 1, 2)
	env.mock.favErr = errors.New("信息不存在")
	env.mock.favResult = nil
	resp := env.doJSON(t, http.MethodPost, "/api/v1/news/8/fav", nil)

	assert.Equal(t, 2401, resp.Code)
	assert.Equal(t, "信息不存在", resp.Message)
}

// ---------- FavStatus ----------

func TestNewsHandler_FavStatus_NotLoggedIn(t *testing.T) {
	env := newNewsHandlerEnv(t, 0, 2)
	// 未登录不拦截，返回默认零值
	resp := env.doJSON(t, http.MethodGet, "/api/v1/news/8/fav", nil)

	assert.Equal(t, 0, resp.Code)
	var fr dto.FavResponse
	require.NoError(t, json.Unmarshal(resp.Data, &fr))
	assert.False(t, fr.Faved)
	assert.Equal(t, 0, fr.FavCount)
}

func TestNewsHandler_FavStatus_Success(t *testing.T) {
	env := newNewsHandlerEnv(t, 5, 2)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/news/8/fav", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(8), env.mock.lastFavStatusNewsID)
}

// ---------- ListFavorites ----------

func TestNewsHandler_ListFavorites_NotLoggedIn(t *testing.T) {
	env := newNewsHandlerEnv(t, 0, 2)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/news/favorites", nil)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, uint(0), env.mock.lastListFavoritesUserID)
}

func TestNewsHandler_ListFavorites_Success(t *testing.T) {
	env := newNewsHandlerEnv(t, 5, 2)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/news/favorites?page=2&page_size=15", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(5), env.mock.lastListFavoritesUserID)
	assert.Equal(t, 2, env.mock.lastListFavoritesPage)
	assert.Equal(t, 15, env.mock.lastListFavoritesPageSize)
	var pr pageResultData
	require.NoError(t, json.Unmarshal(resp.Data, &pr))
	assert.Equal(t, int64(1), pr.Total)
}

func TestNewsHandler_ListFavorites_DefaultPaging(t *testing.T) {
	env := newNewsHandlerEnv(t, 5, 2)
	// 不传 page/page_size → 默认 1/10
	resp := env.doJSON(t, http.MethodGet, "/api/v1/news/favorites", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, 1, env.mock.lastListFavoritesPage)
	assert.Equal(t, 10, env.mock.lastListFavoritesPageSize)
}

func TestNewsHandler_ListFavorites_ServiceError(t *testing.T) {
	env := newNewsHandlerEnv(t, 1, 2)
	env.mock.listFavoritesErr = errors.New("db down")
	env.mock.listFavoritesResult = nil
	resp := env.doJSON(t, http.MethodGet, "/api/v1/news/favorites", nil)

	assert.Equal(t, 2401, resp.Code)
	assert.Equal(t, "db down", resp.Message)
}

// ---------- CreateComment ----------

func TestNewsHandler_CreateComment_Success(t *testing.T) {
	env := newNewsHandlerEnv(t, 5, 2)
	body := dto.CreateCommentRequest{Content: "好货"}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/news/1/comments", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "评论成功", resp.Message)
	assert.Equal(t, uint(1), env.mock.lastCreateCommentNewsID)
	assert.Equal(t, uint(5), env.mock.lastCreateCommentUserID)
	assert.Equal(t, "tester", env.mock.lastCreateCommentUserName)
	require.NotNil(t, env.mock.lastCreateCommentReq)
	assert.Equal(t, "好货", env.mock.lastCreateCommentReq.Content)
	var c dto.CommentInfo
	require.NoError(t, json.Unmarshal(resp.Data, &c))
	assert.Equal(t, "好货", c.Content)
}

func TestNewsHandler_CreateComment_NotLoggedIn(t *testing.T) {
	env := newNewsHandlerEnv(t, 0, 2)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/news/1/comments", dto.CreateCommentRequest{Content: "x"})

	assert.Equal(t, 401, resp.Code)
	assert.Nil(t, env.mock.lastCreateCommentReq)
}

func TestNewsHandler_CreateComment_InvalidID(t *testing.T) {
	env := newNewsHandlerEnv(t, 1, 2)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/news/abc/comments", dto.CreateCommentRequest{Content: "x"})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestNewsHandler_CreateComment_BindError(t *testing.T) {
	env := newNewsHandlerEnv(t, 1, 2)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/news/1/comments", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
	assert.Nil(t, env.mock.lastCreateCommentReq)
}

func TestNewsHandler_CreateComment_ServiceError(t *testing.T) {
	env := newNewsHandlerEnv(t, 1, 2)
	env.mock.createCommentErr = errors.New("信息不存在")
	env.mock.createCommentResult = nil
	resp := env.doJSON(t, http.MethodPost, "/api/v1/news/1/comments", dto.CreateCommentRequest{Content: "x"})

	assert.Equal(t, 2401, resp.Code)
	assert.Equal(t, "信息不存在", resp.Message)
}

// ---------- ListComments ----------

func TestNewsHandler_ListComments_Success(t *testing.T) {
	env := newNewsHandlerEnv(t, 1, 2)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/news/1/comments?page=2&page_size=20", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(1), env.mock.lastListCommentsNewsID)
	assert.Equal(t, 2, env.mock.lastListCommentsPage)
	assert.Equal(t, 20, env.mock.lastListCommentsPageSize)
	var pr pageResultData
	require.NoError(t, json.Unmarshal(resp.Data, &pr))
	assert.Equal(t, int64(1), pr.Total)
	var list []dto.CommentInfo
	require.NoError(t, json.Unmarshal(pr.List, &list))
	require.Len(t, list, 1)
	assert.Equal(t, "好货", list[0].Content)
}

func TestNewsHandler_ListComments_DefaultPaging(t *testing.T) {
	env := newNewsHandlerEnv(t, 1, 2)
	// 不传分页 → 默认 page=1, page_size=20
	resp := env.doJSON(t, http.MethodGet, "/api/v1/news/1/comments", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, 1, env.mock.lastListCommentsPage)
	assert.Equal(t, 20, env.mock.lastListCommentsPageSize)
}

func TestNewsHandler_ListComments_InvalidID(t *testing.T) {
	env := newNewsHandlerEnv(t, 1, 2)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/news/abc/comments", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

// ---------- DeleteComment ----------

func TestNewsHandler_DeleteComment_Success(t *testing.T) {
	env := newNewsHandlerEnv(t, 5, 2)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/news/comments/9", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "删除成功", resp.Message)
	assert.Equal(t, uint(9), env.mock.lastDeleteCommentID)
	assert.Equal(t, uint(5), env.mock.lastDeleteCommentUserID)
}

func TestNewsHandler_DeleteComment_NotLoggedIn(t *testing.T) {
	env := newNewsHandlerEnv(t, 0, 2)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/news/comments/9", nil)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, uint(0), env.mock.lastDeleteCommentID)
}

func TestNewsHandler_DeleteComment_InvalidID(t *testing.T) {
	env := newNewsHandlerEnv(t, 1, 2)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/news/comments/xyz", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的评论ID", resp.Message)
}

func TestNewsHandler_DeleteComment_ServiceError(t *testing.T) {
	env := newNewsHandlerEnv(t, 1, 2)
	env.mock.deleteCommentErr = errors.New("评论不存在")
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/news/comments/9", nil)

	assert.Equal(t, 2401, resp.Code)
	assert.Equal(t, "评论不存在", resp.Message)
}

// ---------- ListMessages ----------

func TestNewsHandler_ListMessages_NotLoggedIn(t *testing.T) {
	env := newNewsHandlerEnv(t, 0, 2)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/news/messages", nil)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, uint(0), env.mock.lastListMessagesUserID)
}

func TestNewsHandler_ListMessages_Success(t *testing.T) {
	env := newNewsHandlerEnv(t, 5, 2)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/news/messages?page=1&page_size=20", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(5), env.mock.lastListMessagesUserID)
	assert.Equal(t, 1, env.mock.lastListMessagesPage)
	assert.Equal(t, 20, env.mock.lastListMessagesPageSize)
	var pr pageResultData
	require.NoError(t, json.Unmarshal(resp.Data, &pr))
	assert.Equal(t, int64(1), pr.Total)
	var list []dto.MessageInfo
	require.NoError(t, json.Unmarshal(pr.List, &list))
	require.Len(t, list, 1)
	assert.Equal(t, "like", list[0].Type)
}

func TestNewsHandler_ListMessages_ServiceError(t *testing.T) {
	env := newNewsHandlerEnv(t, 1, 2)
	env.mock.listMessagesErr = errors.New("db down")
	env.mock.listMessagesResult = nil
	resp := env.doJSON(t, http.MethodGet, "/api/v1/news/messages", nil)

	assert.Equal(t, 2401, resp.Code)
	assert.Equal(t, "db down", resp.Message)
}

// ---------- UnreadCount ----------

func TestNewsHandler_UnreadCount_NotLoggedIn(t *testing.T) {
	env := newNewsHandlerEnv(t, 0, 2)
	// 未登录不拦截，返回默认 {count:0}
	resp := env.doJSON(t, http.MethodGet, "/api/v1/news/messages/unread", nil)

	assert.Equal(t, 0, resp.Code)
	var data map[string]int64
	require.NoError(t, json.Unmarshal(resp.Data, &data))
	assert.Equal(t, int64(0), data["count"])
	assert.Equal(t, uint(0), env.mock.lastUnreadCountUserID)
}

func TestNewsHandler_UnreadCount_Success(t *testing.T) {
	env := newNewsHandlerEnv(t, 5, 2)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/news/messages/unread", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(5), env.mock.lastUnreadCountUserID)
	var data map[string]int64
	require.NoError(t, json.Unmarshal(resp.Data, &data))
	assert.Equal(t, int64(3), data["count"])
}

func TestNewsHandler_UnreadCount_ServiceError(t *testing.T) {
	env := newNewsHandlerEnv(t, 1, 2)
	env.mock.unreadCountErr = errors.New("redis down")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/news/messages/unread", nil)

	assert.Equal(t, 2401, resp.Code)
	assert.Equal(t, "redis down", resp.Message)
}

// ---------- MarkRead ----------

func TestNewsHandler_MarkRead_Success(t *testing.T) {
	env := newNewsHandlerEnv(t, 5, 2)
	body := struct {
		IDs []uint `json:"ids"`
	}{IDs: []uint{1, 2, 3}}
	resp := env.doJSON(t, http.MethodPut, "/api/v1/news/messages/read", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "已标记已读", resp.Message)
	assert.Equal(t, uint(5), env.mock.lastMarkReadUserID)
	assert.Equal(t, []uint{1, 2, 3}, env.mock.lastMarkReadIDs)
}

func TestNewsHandler_MarkRead_NotLoggedIn(t *testing.T) {
	env := newNewsHandlerEnv(t, 0, 2)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/news/messages/read", struct {
		IDs []uint `json:"ids"`
	}{IDs: []uint{1}})

	assert.Equal(t, 401, resp.Code)
	assert.Nil(t, env.mock.lastMarkReadIDs)
}

func TestNewsHandler_MarkRead_BindError(t *testing.T) {
	env := newNewsHandlerEnv(t, 1, 2)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/news/messages/read", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
}

func TestNewsHandler_MarkRead_ServiceError(t *testing.T) {
	env := newNewsHandlerEnv(t, 1, 2)
	env.mock.markReadErr = errors.New("db down")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/news/messages/read", struct {
		IDs []uint `json:"ids"`
	}{IDs: []uint{1}})

	assert.Equal(t, 2401, resp.Code)
	assert.Equal(t, "db down", resp.Message)
}
