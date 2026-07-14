// Package plugin 插件系统核心单元测试
// 覆盖 Manager 单例、Register/Get/List/InitAll/RegisterAllRoutes/CloseAll/Unregister 全方法
// 与 PluginInitError/PluginCloseError 错误类型（Error/Unwrap）。
// 使用内存 mock Plugin + mock RouterGroup，无外部依赖（无 DB/Redis/Docker）。
package plugin

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockPlugin 内存模拟插件，记录生命周期调用顺序与可注入错误
type mockPlugin struct {
	name    string
	version string

	// 调用记录
	initCount   int
	closeCount  int
	routesCount int

	// 注入错误
	initErr  error
	closeErr error

	// 记录 RegisterRoutes 收到的 RouterGroup（用于断言分组路径）
	receivedGroup RouterGroup

	// 闭包记录调用顺序（共享切片指针）
	callLog *[]string
}

func (p *mockPlugin) Name() string    { return p.name }
func (p *mockPlugin) Version() string { return p.version }

func (p *mockPlugin) Init(ctx context.Context) error {
	p.initCount++
	if p.callLog != nil {
		*p.callLog = append(*p.callLog, "init:"+p.name)
	}
	return p.initErr
}

func (p *mockPlugin) RegisterRoutes(router RouterGroup) {
	p.routesCount++
	p.receivedGroup = router
	if p.callLog != nil {
		*p.callLog = append(*p.callLog, "routes:"+p.name)
	}
}

func (p *mockPlugin) Close() error {
	p.closeCount++
	if p.callLog != nil {
		*p.callLog = append(*p.callLog, "close:"+p.name)
	}
	return p.closeErr
}

// mockRouterGroup 内存模拟路由组，记录 Group 创建的子组与各方法注册的路径
type mockRouterGroup struct {
	groupPath   string
	groups      []*mockRouterGroup // 创建的子组
	getPaths    []string
	postPaths   []string
	putPaths    []string
	deletePaths []string
	patchPaths  []string
}

func newMockRouterGroup() *mockRouterGroup {
	return &mockRouterGroup{}
}

func (g *mockRouterGroup) Group(relativePath string, handlers ...HandlerFunc) RouterGroup {
	sub := &mockRouterGroup{groupPath: relativePath}
	g.groups = append(g.groups, sub)
	return sub
}

func (g *mockRouterGroup) GET(relativePath string, handlers ...HandlerFunc) {
	g.getPaths = append(g.getPaths, relativePath)
}
func (g *mockRouterGroup) POST(relativePath string, handlers ...HandlerFunc) {
	g.postPaths = append(g.postPaths, relativePath)
}
func (g *mockRouterGroup) PUT(relativePath string, handlers ...HandlerFunc) {
	g.putPaths = append(g.putPaths, relativePath)
}
func (g *mockRouterGroup) DELETE(relativePath string, handlers ...HandlerFunc) {
	g.deletePaths = append(g.deletePaths, relativePath)
}
func (g *mockRouterGroup) PATCH(relativePath string, handlers ...HandlerFunc) {
	g.patchPaths = append(g.patchPaths, relativePath)
}

// newManager 直接构造新 Manager（绕过单例），保证用例间状态隔离
func newManager() *Manager {
	return &Manager{
		plugins: make(map[string]Plugin),
		order:   make([]string, 0),
	}
}

// --- GetManager 单例 ---

// TestGetManager_Singleton 多次调用 GetManager 返回同一实例（sync.Once）
func TestGetManager_Singleton(t *testing.T) {
	m1 := GetManager()
	m2 := GetManager()
	require.Same(t, m1, m2, "GetManager 应返回同一单例实例")
}

// TestGetManager_NotNil 单例非空且可安全调用 List
func TestGetManager_NotNil(t *testing.T) {
	m := GetManager()
	require.NotNil(t, m)
	// 单例上调用 List 不应 panic（即便为空）
	list := m.List()
	assert.NotNil(t, list, "空 List 返回非 nil 空切片")
}

// --- Register ---

// TestRegister_Success 注册成功：可在 plugins 与 order 中查到
func TestRegister_Success(t *testing.T) {
	m := newManager()
	p := &mockPlugin{name: "user", version: "1.0.0"}

	err := m.Register(p)
	require.NoError(t, err)

	// 内部状态
	require.Len(t, m.order, 1)
	assert.Equal(t, "user", m.order[0])
	assert.Contains(t, m.plugins, "user")
}

// TestRegister_Duplicate 同名插件重复注册返回 ErrPluginAlreadyExists，状态不变
func TestRegister_Duplicate(t *testing.T) {
	m := newManager()
	p1 := &mockPlugin{name: "news", version: "1.0.0"}
	p2 := &mockPlugin{name: "news", version: "2.0.0"}

	require.NoError(t, m.Register(p1))
	err := m.Register(p2)
	assert.ErrorIs(t, err, ErrPluginAlreadyExists)

	// 仍是首个插件
	require.Len(t, m.order, 1)
	got, ok := m.Get("news")
	require.True(t, ok)
	assert.Same(t, p1, got, "重复注册不应覆盖既有插件")
}

// TestRegister_PreservesOrder 多插件按注册顺序保留在 order 中
func TestRegister_PreservesOrder(t *testing.T) {
	m := newManager()
	for _, name := range []string{"user", "region", "category", "news"} {
		require.NoError(t, m.Register(&mockPlugin{name: name}))
	}
	assert.Equal(t, []string{"user", "region", "category", "news"}, m.order)
}

// --- Get ---

// TestGet_Found 已注册插件可命中
func TestGet_Found(t *testing.T) {
	m := newManager()
	p := &mockPlugin{name: "file"}
	require.NoError(t, m.Register(p))

	got, ok := m.Get("file")
	require.True(t, ok)
	assert.Same(t, p, got)
}

// TestGet_NotFound 未注册插件返回 ok=false（不返回错误，便于零值兜底）
func TestGet_NotFound(t *testing.T) {
	m := newManager()
	got, ok := m.Get("not-exist")
	assert.False(t, ok)
	assert.Nil(t, got)
}

// --- List ---

// TestList_Empty 空管理器 List 返回非 nil 空切片
func TestList_Empty(t *testing.T) {
	m := newManager()
	list := m.List()
	require.NotNil(t, list)
	assert.Len(t, list, 0)
}

// TestList_OrderPreserved List 按注册顺序返回插件（与 order 一致）
func TestList_OrderPreserved(t *testing.T) {
	m := newManager()
	names := []string{"alpha", "beta", "gamma"}
	for _, n := range names {
		require.NoError(t, m.Register(&mockPlugin{name: n}))
	}

	list := m.List()
	require.Len(t, list, 3)
	for i, p := range list {
		assert.Equal(t, names[i], p.Name(), "List 顺序应与注册顺序一致")
	}
}

// TestList_ReturnsCopy List 返回新切片，外部修改不影响内部 order
// （实现上 result 是 make 新切片，外部 append 不会回写 m.order）
func TestList_ReturnsCopy(t *testing.T) {
	m := newManager()
	require.NoError(t, m.Register(&mockPlugin{name: "a"}))

	list := m.List()
	list = append(list, &mockPlugin{name: "injected"})

	// 内部 List 仍只含 1 个
	require.Len(t, m.List(), 1, "外部 append 不应影响内部 order")
}

// --- InitAll ---

// TestInitAll_Success 全部插件 Init 成功，调用次数均为 1
func TestInitAll_Success(t *testing.T) {
	m := newManager()
	plugins := []*mockPlugin{
		{name: "user"},
		{name: "region"},
		{name: "news"},
	}
	for _, p := range plugins {
		require.NoError(t, m.Register(p))
	}

	ctx := context.Background()
	require.NoError(t, m.InitAll(ctx))
	for _, p := range plugins {
		assert.Equal(t, 1, p.initCount, "%s Init 应被调用一次", p.name)
	}
}

// TestInitAll_OneFails 第一个失败立即返回 PluginInitError，后续不再调用
func TestInitAll_OneFails(t *testing.T) {
	m := newManager()
	sentinel := errors.New("db connection refused")
	failing := &mockPlugin{name: "region", initErr: sentinel}
	after := &mockPlugin{name: "news"}

	require.NoError(t, m.Register(&mockPlugin{name: "user"}))
	require.NoError(t, m.Register(failing))
	require.NoError(t, m.Register(after))

	err := m.InitAll(context.Background())
	require.Error(t, err)

	// 应为 PluginInitError 且包裹原始错误
	var pie *PluginInitError
	require.ErrorAs(t, err, &pie, "应返回 *PluginInitError")
	assert.Equal(t, "region", pie.PluginName)
	assert.ErrorIs(t, err, sentinel, "Unwrap 应返回原始错误")
	assert.Contains(t, err.Error(), "region")
	assert.Contains(t, err.Error(), "db connection refused")
}

// TestInitAll_StopsAtFirstFailure 失败后后续插件 Init 不被调用
func TestInitAll_StopsAtFirstFailure(t *testing.T) {
	m := newManager()
	failing := &mockPlugin{name: "a", initErr: errors.New("boom")}
	after := &mockPlugin{name: "b"}

	require.NoError(t, m.Register(failing))
	require.NoError(t, m.Register(after))

	_ = m.InitAll(context.Background())
	assert.Equal(t, 1, failing.initCount, "失败插件自身应被调用一次")
	assert.Equal(t, 0, after.initCount, "失败后的插件不应被调用")
}

// TestInitAll_Empty 空管理器 InitAll 返回 nil
func TestInitAll_Empty(t *testing.T) {
	m := newManager()
	require.NoError(t, m.InitAll(context.Background()))
}

// --- RegisterAllRoutes ---

// TestRegisterAllRoutes_GroupPath 每个插件按 /api/v1/{name} 创建子组并调用 RegisterRoutes
func TestRegisterAllRoutes_GroupPath(t *testing.T) {
	m := newManager()
	pUser := &mockPlugin{name: "user"}
	pNews := &mockPlugin{name: "news"}
	require.NoError(t, m.Register(pUser))
	require.NoError(t, m.Register(pNews))

	root := newMockRouterGroup()
	m.RegisterAllRoutes(root)

	require.Len(t, root.groups, 2)
	// 子组路径 = /api/v1/{name}
	assert.Equal(t, "/api/v1/user", root.groups[0].groupPath)
	assert.Equal(t, "/api/v1/news", root.groups[1].groupPath)
	// RegisterRoutes 被调用且收到对应子组
	assert.Equal(t, 1, pUser.routesCount)
	assert.Equal(t, 1, pNews.routesCount)
	require.NotNil(t, pUser.receivedGroup)
	sub, ok := pUser.receivedGroup.(*mockRouterGroup)
	require.True(t, ok)
	assert.Equal(t, "/api/v1/user", sub.groupPath)
}

// TestRegisterAllRoutes_Empty 空管理器不创建任何子组
func TestRegisterAllRoutes_Empty(t *testing.T) {
	m := newManager()
	root := newMockRouterGroup()
	m.RegisterAllRoutes(root)
	assert.Len(t, root.groups, 0)
}

// TestRegisterAllRoutes_Order 子组创建顺序与注册顺序一致
func TestRegisterAllRoutes_Order(t *testing.T) {
	m := newManager()
	for _, n := range []string{"zeta", "alpha", "mid"} {
		require.NoError(t, m.Register(&mockPlugin{name: n}))
	}
	root := newMockRouterGroup()
	m.RegisterAllRoutes(root)

	require.Len(t, root.groups, 3)
	assert.Equal(t, "/api/v1/zeta", root.groups[0].groupPath)
	assert.Equal(t, "/api/v1/alpha", root.groups[1].groupPath)
	assert.Equal(t, "/api/v1/mid", root.groups[2].groupPath)
}

// --- CloseAll ---

// TestCloseAll_ReverseOrder 关闭顺序与注册顺序相反（LIFO）
func TestCloseAll_ReverseOrder(t *testing.T) {
	m := newManager()
	var log []string
	for _, n := range []string{"a", "b", "c"} {
		require.NoError(t, m.Register(&mockPlugin{name: n, callLog: &log}))
	}

	require.NoError(t, m.CloseAll())
	assert.Equal(t, []string{"close:c", "close:b", "close:a"}, log, "应逆序关闭")
}

// TestCloseAll_FirstErrorCaptured 首个错误被捕获返回，但全部插件仍被尝试关闭
func TestCloseAll_FirstErrorCaptured(t *testing.T) {
	m := newManager()
	var log []string
	// 用命名变量保存错误实例，errors.Is 需要同一指针（errors.New 每次返回新值）
	errB := errors.New("b close failed")
	errC := errors.New("c close failed")
	pA := &mockPlugin{name: "a", callLog: &log}
	pB := &mockPlugin{name: "b", closeErr: errB, callLog: &log}
	pC := &mockPlugin{name: "c", closeErr: errC, callLog: &log}

	require.NoError(t, m.Register(pA))
	require.NoError(t, m.Register(pB))
	require.NoError(t, m.Register(pC))

	err := m.CloseAll()
	// 逆序：c 先失败 → 首个错误是 c 的
	require.Error(t, err)
	assert.ErrorIs(t, err, errC, "首个（逆序最早的）错误应被返回")
	assert.NotErrorIs(t, err, errB, "后续 b 的错误不应覆盖首个错误")
	// 全部均被调用一次
	assert.Equal(t, 1, pA.closeCount)
	assert.Equal(t, 1, pB.closeCount)
	assert.Equal(t, 1, pC.closeCount)
	// 顺序仍为逆序
	assert.Equal(t, []string{"close:c", "close:b", "close:a"}, log)
}

// TestCloseAll_NoError 全部成功返回 nil
func TestCloseAll_NoError(t *testing.T) {
	m := newManager()
	require.NoError(t, m.Register(&mockPlugin{name: "a"}))
	require.NoError(t, m.Register(&mockPlugin{name: "b"}))
	require.NoError(t, m.CloseAll())
}

// TestCloseAll_Empty 空管理器返回 nil
func TestCloseAll_Empty(t *testing.T) {
	m := newManager()
	require.NoError(t, m.CloseAll())
}

// --- Unregister ---

// TestUnregister_Success 注销后 Get 不可命中且 order 移除
func TestUnregister_Success(t *testing.T) {
	m := newManager()
	require.NoError(t, m.Register(&mockPlugin{name: "a"}))
	require.NoError(t, m.Register(&mockPlugin{name: "b"}))
	require.NoError(t, m.Register(&mockPlugin{name: "c"}))

	require.NoError(t, m.Unregister("b"))
	_, ok := m.Get("b")
	assert.False(t, ok)
	// order 中 b 被移除，其余顺序保留
	assert.Equal(t, []string{"a", "c"}, m.order)
}

// TestUnregister_NotFound 注销不存在的插件返回 ErrPluginNotFound，状态不变
func TestUnregister_NotFound(t *testing.T) {
	m := newManager()
	require.NoError(t, m.Register(&mockPlugin{name: "a"}))

	err := m.Unregister("not-exist")
	assert.ErrorIs(t, err, ErrPluginNotFound)
	require.Len(t, m.order, 1, "未找到时状态不变")
}

// TestUnregister_Head 注销 order 首元素
func TestUnregister_Head(t *testing.T) {
	m := newManager()
	require.NoError(t, m.Register(&mockPlugin{name: "a"}))
	require.NoError(t, m.Register(&mockPlugin{name: "b"}))

	require.NoError(t, m.Unregister("a"))
	assert.Equal(t, []string{"b"}, m.order)
}

// TestUnregister_Tail 注销 order 末尾元素
func TestUnregister_Tail(t *testing.T) {
	m := newManager()
	require.NoError(t, m.Register(&mockPlugin{name: "a"}))
	require.NoError(t, m.Register(&mockPlugin{name: "b"}))

	require.NoError(t, m.Unregister("b"))
	assert.Equal(t, []string{"a"}, m.order)
}

// TestUnregister_OnlyOne 注销唯一插件后 List 为空
func TestUnregister_OnlyOne(t *testing.T) {
	m := newManager()
	require.NoError(t, m.Register(&mockPlugin{name: "solo"}))

	require.NoError(t, m.Unregister("solo"))
	assert.Len(t, m.List(), 0)
}

// --- 错误类型 ---

// TestPluginInitError_Error Error 字符串包含插件名与原始错误
func TestPluginInitError_Error(t *testing.T) {
	inner := errors.New("timeout")
	e := &PluginInitError{PluginName: "news", Err: inner}
	s := e.Error()
	assert.Contains(t, s, "news")
	assert.Contains(t, s, "timeout")
	assert.Contains(t, s, "init failed")
}

// TestPluginInitError_Unwrap Unwrap 返回原始错误，支持 errors.Is/As
func TestPluginInitError_Unwrap(t *testing.T) {
	inner := errors.New("dial tcp: refused")
	e := &PluginInitError{PluginName: "region", Err: inner}
	assert.ErrorIs(t, e, inner)
}

// TestPluginCloseError_Error Error 字符串包含插件名与原始错误
func TestPluginCloseError_Error(t *testing.T) {
	inner := errors.New("flush failed")
	e := &PluginCloseError{PluginName: "file", Err: inner}
	s := e.Error()
	assert.Contains(t, s, "file")
	assert.Contains(t, s, "flush failed")
	assert.Contains(t, s, "close failed")
}

// TestPluginCloseError_Unwrap Unwrap 返回原始错误
func TestPluginCloseError_Unwrap(t *testing.T) {
	inner := errors.New("conn closed")
	e := &PluginCloseError{PluginName: "ws", Err: inner}
	assert.ErrorIs(t, e, inner)
}

// TestErrSentinelValues 两个哨兵错误非空且语义稳定（防止被误改为 nil）
func TestErrSentinelValues(t *testing.T) {
	require.NotNil(t, ErrPluginAlreadyExists)
	require.NotNil(t, ErrPluginNotFound)
	assert.NotEqual(t, ErrPluginAlreadyExists, ErrPluginNotFound)
	assert.Contains(t, ErrPluginAlreadyExists.Error(), "already exists")
	assert.Contains(t, ErrPluginNotFound.Error(), "not found")
}

// --- 并发安全 ---

// TestManager_ConcurrentReadWrite 并发 Register/Get/List/Unregister 不触发 race
// 用 -race 运行验证锁正确性（mu sync.RWMutex 保护 plugins + order）。
func TestManager_ConcurrentReadWrite(t *testing.T) {
	m := newManager()
	var wg sync.WaitGroup
	names := []string{"p0", "p1", "p2", "p3", "p4"}

	// 并发注册
	for _, n := range names {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			_ = m.Register(&mockPlugin{name: name})
		}(n)
	}
	// 并发读
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = m.List()
			_, _ = m.Get("p2")
		}()
	}
	// 并发注销（可能未命中）
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = m.Unregister("p3")
	}()

	wg.Wait()
	// 不做严格数量断言（并发下 Register/Unregister 交错结果不确定），
	// 仅验证不 panic 且最终状态自洽：List 长度 == len(order)
	assert.Len(t, m.List(), len(m.order))
}
