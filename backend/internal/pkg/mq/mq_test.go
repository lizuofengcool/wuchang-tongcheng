package mq

import (
	"context"
	"errors"
	"sync"
	"testing"

	"wuchang-tongcheng/internal/pkg/config"
)

// resetClient 清除包级 client 全局变量，保证用例间状态隔离。
// t.Cleanup 在用例结束后再次清空，避免后续用例受残留影响。
func resetClient(t *testing.T) {
	t.Helper()
	mu.Lock()
	client = nil
	mu.Unlock()
	t.Cleanup(func() {
		mu.Lock()
		client = nil
		mu.Unlock()
	})
}

// newClientWithCache 构造一个仅含 declaredQueues/declaredExchanges 缓存、
// conn 与 channel 均为 nil 的 *Client，用于测试缓存命中短路路径
// （cache 命中分支在触碰 channel 前直接 return nil，故 nil channel 安全）。
func newClientWithCache(queues, exchanges map[string]struct{}) *Client {
	if queues == nil {
		queues = make(map[string]struct{})
	}
	if exchanges == nil {
		exchanges = make(map[string]struct{})
	}
	return &Client{
		conn:              nil,
		channel:           nil,
		url:               "amqp://test@test:5672/",
		declaredQueues:    queues,
		declaredExchanges: exchanges,
	}
}

// --- Init 未配置降级测试 ---

func TestInit_NilConfigSkipsInitialization(t *testing.T) {
	resetClient(t)
	if err := Init(nil); err != nil {
		t.Fatalf("Init(nil) 不应返回错误，got: %v", err)
	}
	if GetClient() != nil {
		t.Errorf("Init(nil) 后 GetClient() 应为 nil（不初始化）")
	}
	if IsAvailable() {
		t.Errorf("Init(nil) 后 IsAvailable() 应为 false")
	}
}

func TestInit_EmptyHostSkipsInitialization(t *testing.T) {
	resetClient(t)
	// Host 为空字符串应跳过初始化（允许无 MQ 降级运行）
	cfg := &config.RabbitMQConfig{Host: "", Port: 5672, User: "guest", Password: "guest"}
	if err := Init(cfg); err != nil {
		t.Fatalf("Init(Host=\"\") 不应返回错误，got: %v", err)
	}
	if GetClient() != nil {
		t.Errorf("Init(Host=\"\") 后 GetClient() 应为 nil")
	}
	if IsAvailable() {
		t.Errorf("Init(Host=\"\") 后 IsAvailable() 应为 false")
	}
}

// --- GetClient / IsAvailable / Close 未初始化状态测试 ---

func TestGetClient_UninitializedReturnsNil(t *testing.T) {
	resetClient(t)
	// 未初始化时 GetClient() 不 panic，返回 nil 指针
	c := GetClient()
	if c != nil {
		t.Errorf("未初始化时 GetClient() 应为 nil，got: %#v", c)
	}
}

func TestIsAvailable_UninitializedReturnsFalse(t *testing.T) {
	resetClient(t)
	if IsAvailable() {
		t.Errorf("未初始化时 IsAvailable() 应为 false")
	}
}

func TestClose_UninitializedReturnsNil(t *testing.T) {
	resetClient(t)
	// 未初始化时 Close() 不 panic 不报错，直接返回 nil
	if err := Close(); err != nil {
		t.Errorf("未初始化时 Close() 应返回 nil，got: %v", err)
	}
}

// --- DeclareQueue 缓存命中短路测试 ---

func TestDeclareQueue_CacheHitSkipsChannel(t *testing.T) {
	// 预填缓存，DeclaredQueues 命中分支在触碰 channel 前直接 return nil
	// 故 channel 为 nil 也不会触发 nil pointer dereference
	c := newClientWithCache(map[string]struct{}{"my-queue": {}}, nil)
	if err := c.DeclareQueue("my-queue"); err != nil {
		t.Errorf("DeclareQueue 命中缓存时应返回 nil，got: %v", err)
	}
	// 缓存集合不变
	if _, ok := c.declaredQueues["my-queue"]; !ok {
		t.Errorf("缓存命中后 declaredQueues 仍应包含该键")
	}
}

func TestDeclareTopicExchange_CacheHitSkipsChannel(t *testing.T) {
	// 预填缓存，DeclareTopicExchange 命中分支在触碰 channel 前直接 return nil
	c := newClientWithCache(nil, map[string]struct{}{"my-exchange": {}})
	if err := c.DeclareTopicExchange("my-exchange"); err != nil {
		t.Errorf("DeclareTopicExchange 命中缓存时应返回 nil，got: %v", err)
	}
	if _, ok := c.declaredExchanges["my-exchange"]; !ok {
		t.Errorf("缓存命中后 declaredExchanges 仍应包含该键")
	}
}

// --- DeclareQueue 缓存未命中将触碰 channel 测试 ---

func TestDeclareQueue_CacheMissTouchesChannelPanics(t *testing.T) {
	// channel 为 nil，cache miss 分支会调用 c.channel.QueueDeclare(...) 触发 panic
	// 验证缓存逻辑：未命中即会向下走真实声明路径
	c := newClientWithCache(nil, nil)
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("DeclareQueue 未命中缓存且 channel 为 nil 时应 panic，got nil")
		}
	}()
	_ = c.DeclareQueue("not-cached-queue")
}

func TestDeclareTopicExchange_CacheMissTouchesChannelPanics(t *testing.T) {
	// channel 为 nil，cache miss 分支会调用 c.channel.ExchangeDeclare(...) 触发 panic
	c := newClientWithCache(nil, nil)
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("DeclareTopicExchange 未命中缓存且 channel 为 nil 时应 panic，got nil")
		}
	}()
	_ = c.DeclareTopicExchange("not-cached-exchange")
}

// --- declaredMu 互斥锁测试 ---

func TestDeclaredMutex_ConcurrentCacheReadSafe(t *testing.T) {
	// 并发调用 cache-hit 路径（不触碰 channel），验证 declaredMu 锁正确性
	c := newClientWithCache(map[string]struct{}{"shared-q": {}}, map[string]struct{}{"shared-ex": {}})
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			_ = c.DeclareQueue("shared-q")
		}()
		go func() {
			defer wg.Done()
			_ = c.DeclareTopicExchange("shared-ex")
		}()
	}
	wg.Wait()
	// 仅断言不 panic 不死锁，缓存集合内容不变
	if _, ok := c.declaredQueues["shared-q"]; !ok {
		t.Errorf("并发后 declaredQueues 应仍包含 shared-q")
	}
	if _, ok := c.declaredExchanges["shared-ex"]; !ok {
		t.Errorf("并发后 declaredExchanges 应仍包含 shared-ex")
	}
}

// --- Handler 类型签名测试 ---

func TestHandler_TypeCallable(t *testing.T) {
	// Handler 是 func(body []byte) error 类型别名，验证可声明可调用
	var h Handler = func(body []byte) error {
		if len(body) == 0 {
			return errors.New("empty body")
		}
		return nil
	}
	if err := h(nil); err == nil {
		t.Errorf("Handler 对空 body 应返回 error")
	}
	if err := h([]byte("payload")); err != nil {
		t.Errorf("Handler 对非空 body 应返回 nil，got: %v", err)
	}
}

// --- ErrNotAvailable 哨兵错误测试 ---

func TestErrNotAvailable_SentinelStable(t *testing.T) {
	if ErrNotAvailable == nil {
		t.Fatalf("ErrNotAvailable 不应为 nil")
	}
	// 错误消息稳定
	want := "rabbitmq not available"
	if got := ErrNotAvailable.Error(); got != want {
		t.Errorf("ErrNotAvailable.Error() = %q, want %q", got, want)
	}
	// errors.Is 可识别同一指针
	if !errors.Is(ErrNotAvailable, ErrNotAvailable) {
		t.Errorf("errors.Is(ErrNotAvailable, ErrNotAvailable) 应为 true")
	}
	// 与其它错误区分
	other := errors.New("other error")
	if errors.Is(other, ErrNotAvailable) {
		t.Errorf("其它错误不应被 errors.Is 识别为 ErrNotAvailable")
	}
}

// --- BindQueue 调用顺序测试（缓存命中路径） ---

func TestBindQueue_CacheHitSkipsChannel(t *testing.T) {
	// queue 与 exchange 均已在缓存中，BindQueue 不应触碰 channel
	// 注意：BindQueue 末尾会调用 c.channel.QueueBind(...)，channel 为 nil 会 panic
	// 故本用例仅验证 DeclareQueue/DeclareTopicExchange 的缓存短路
	c := newClientWithCache(
		map[string]struct{}{"bind-q": {}},
		map[string]struct{}{"bind-ex": {}},
	)
	// 手动验证缓存命中分支不报错（真实 QueueBind 需要 channel，本测试不覆盖）
	if _, ok := c.declaredQueues["bind-q"]; !ok {
		t.Fatalf("预填的 bind-q 缓存缺失")
	}
	if _, ok := c.declaredExchanges["bind-ex"]; !ok {
		t.Fatalf("预填的 bind-ex 缓存缺失")
	}
	// 验证 DeclareQueue/DeclareTopicExchange 在 BindQueue 链路中走缓存命中分支
	if err := c.DeclareQueue("bind-q"); err != nil {
		t.Errorf("DeclareQueue(bind-q) 缓存命中应返回 nil，got: %v", err)
	}
	if err := c.DeclareTopicExchange("bind-ex"); err != nil {
		t.Errorf("DeclareTopicExchange(bind-ex) 缓存命中应返回 nil，got: %v", err)
	}
}

// --- Publish/SimplePublish 路径参数透传测试（缓存命中即返回前需要真实 channel，仅做 panic 断言） ---

func TestSimplePublish_CacheMissTouchesChannelPanics(t *testing.T) {
	// SimplePublish 会先 DeclareQueue；queue 未在缓存中 → channel.QueueDeclare panic
	c := newClientWithCache(nil, nil)
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("SimplePublish 未命中缓存且 channel 为 nil 时应 panic")
		}
	}()
	_ = c.SimplePublish(context.Background(), "q", []byte("body"))
}

func TestPublish_CacheMissTouchesChannelPanics(t *testing.T) {
	// Publish 会先 DeclareTopicExchange；exchange 未在缓存中 → channel.ExchangeDeclare panic
	c := newClientWithCache(nil, nil)
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("Publish 未命中缓存且 channel 为 nil 时应 panic")
		}
	}()
	_ = c.Publish(context.Background(), "ex", "key", []byte("body"))
}
