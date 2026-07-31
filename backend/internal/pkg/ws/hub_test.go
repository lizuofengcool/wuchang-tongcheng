package ws

import (
	"encoding/json"
	"testing"
	"time"
)

// TestHubNoOpWhenEmpty 验证空 Hub 的安全行为：
// 无连接时 SendToUser/Broadcast 不 panic，OnlineCount 返回 0。
func TestHubNoOpWhenEmpty(t *testing.T) {
	// 保存原始状态，测试后恢复，避免污染其他测试
	origHub := hub
	t.Cleanup(func() { hub = origHub })

	Init()
	defer Close()

	if !IsAvailable() {
		t.Fatal("Init 后 IsAvailable 应为 true")
	}
	h := GetHub()

	// 空连接时向任意用户推送不应 panic
	h.SendToUser(999, &Message{Type: TypeLike, Data: LikeNotification{NewsID: 1}})

	// 空连接时广播不应 panic
	h.Broadcast(&Message{Type: TypeSystem, Data: "hello"})

	if got := h.OnlineCount(); got != 0 {
		t.Errorf("空 Hub OnlineCount 应为 0，got %d", got)
	}
}

// waitFor 轮询等待条件成立（最多 300ms），避免对事件循环做硬 sleep。
func waitFor(t *testing.T, cond func() bool, msg string) {
	t.Helper()
	deadline := time.Now().Add(300 * time.Millisecond)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("条件未满足: %s", msg)
}

// newTestClient 构造不含真实 websocket 连接的 Client。
// Hub 的注册/注销/推送/计数路径均不触碰 conn（仅 Close 会调用 conn.Close），
// 因此可在不依赖网络的情况下测试 Hub 的核心并发逻辑。
func newTestClient(userID uint, username string) *Client {
	return &Client{
		UserID:   userID,
		Username: username,
		send:     make(chan []byte, 8),
	}
}

// newTestHub 构造独立 Hub 并启动事件循环，测试间互不影响。
func newTestHub(t *testing.T) *Hub {
	t.Helper()
	h := &Hub{
		clients:    make(map[uint]map[*Client]struct{}),
		register:   make(chan *Client, 64),
		unregister: make(chan *Client, 64),
		broadcast:  make(chan []byte, 256),
	}
	go h.Run()
	return h
}

// TestHubRegisterMultiConnection 验证单用户多连接注册后 OnlineCount 正确累计。
func TestHubRegisterMultiConnection(t *testing.T) {
	h := newTestHub(t)
	c1 := newTestClient(1, "alice")
	c2 := newTestClient(1, "alice") // 同一用户的第二个连接（多端登录）
	c3 := newTestClient(2, "bob")

	h.Register(c1)
	h.Register(c2)
	h.Register(c3)

	waitFor(t, func() bool { return h.OnlineCount() == 3 }, "3 个连接应在线")
}

// TestHubRegisterIdempotentSamePointer 验证相同 Client 指针重复注册不会重复计数（集合语义）。
func TestHubRegisterIdempotentSamePointer(t *testing.T) {
	h := newTestHub(t)
	c := newTestClient(1, "alice")
	h.Register(c)
	h.Register(c) // 相同指针

	waitFor(t, func() bool { return h.OnlineCount() == 1 }, "相同指针应只计 1 个连接")
}

// TestHubSendToUserDeliversToAllConnections 验证定向推送到某用户时，
// 该用户的全部连接都能收到消息。
func TestHubSendToUserDeliversToAllConnections(t *testing.T) {
	h := newTestHub(t)
	c1 := newTestClient(1, "alice")
	c2 := newTestClient(1, "alice")
	h.Register(c1)
	h.Register(c2)
	waitFor(t, func() bool { return h.OnlineCount() == 2 }, "2 个连接在线")

	h.SendToUser(1, &Message{Type: TypeLike, Data: LikeNotification{
		NewsID: 7, NewsTitle: "T", FromUser: "bob", Liked: true, LikeCount: 5,
	}})

	for i, c := range []*Client{c1, c2} {
		select {
		case raw := <-c.send:
			var got Message
			if err := json.Unmarshal(raw, &got); err != nil {
				t.Fatalf("client %d 反序列化失败: %v", i, err)
			}
			if got.Type != TypeLike {
				t.Fatalf("client %d type=%s want %s", i, got.Type, TypeLike)
			}
		case <-time.After(200 * time.Millisecond):
			t.Fatalf("client %d 未收到推送", i)
		}
	}
}

// TestHubSendToUserUnknownUserNoOp 验证向无在线连接的用户推送是安全 no-op，
// 不会误投给其他用户。
func TestHubSendToUserUnknownUserNoOp(t *testing.T) {
	h := newTestHub(t)
	c1 := newTestClient(1, "alice")
	h.Register(c1)
	waitFor(t, func() bool { return h.OnlineCount() == 1 }, "1 个连接在线")

	h.SendToUser(999, &Message{Type: TypeSystem, Data: "hi"})

	select {
	case raw := <-c1.send:
		t.Fatalf("不应收到消息，got %s", raw)
	case <-time.After(50 * time.Millisecond):
		// 预期无消息
	}
}

// TestHubUnregisterRemovesConnection 验证注销连接后计数下降，
// 且被注销连接的 send 通道被关闭。
func TestHubUnregisterRemovesConnection(t *testing.T) {
	h := newTestHub(t)
	c1 := newTestClient(1, "alice")
	c2 := newTestClient(1, "alice")
	h.Register(c1)
	h.Register(c2)
	waitFor(t, func() bool { return h.OnlineCount() == 2 }, "2 个连接在线")

	h.Unregister(c1)
	waitFor(t, func() bool { return h.OnlineCount() == 1 }, "注销后剩 1 个连接")

	// Run 注销时会 close(c.send)，接收应返回零值且 ok=false
	if _, ok := <-c1.send; ok {
		t.Fatalf("c1.send 应已被关闭")
	}
}

// TestHubUnregisterLastRemovesUser 验证注销某用户最后一个连接后，
// 该用户条目被整体移除，后续推送为 no-op。
func TestHubUnregisterLastRemovesUser(t *testing.T) {
	h := newTestHub(t)
	c1 := newTestClient(1, "alice")
	h.Register(c1)
	waitFor(t, func() bool { return h.OnlineCount() == 1 }, "1 个连接在线")

	h.Unregister(c1)
	waitFor(t, func() bool { return h.OnlineCount() == 0 }, "注销后 0 个连接")

	// 用户已不存在，推送不应 panic
	h.SendToUser(1, &Message{Type: TypeSystem, Data: "x"})
}

// TestHubUnregisterUnknownClientSafe 验证注销一个从未注册的 Client 是安全的，不 panic。
func TestHubUnregisterUnknownClientSafe(t *testing.T) {
	h := newTestHub(t)
	c := newTestClient(1, "alice")
	h.Unregister(c) // 从未注册

	waitFor(t, func() bool { return h.OnlineCount() == 0 }, "仍为 0 连接")
}

// TestHubBroadcastDeliversToAll 验证全局广播投递给所有在线连接。
func TestHubBroadcastDeliversToAll(t *testing.T) {
	h := newTestHub(t)
	c1 := newTestClient(1, "alice")
	c2 := newTestClient(2, "bob")
	h.Register(c1)
	h.Register(c2)
	waitFor(t, func() bool { return h.OnlineCount() == 2 }, "2 个连接在线")

	h.Broadcast(&Message{Type: TypeNews, Data: "breaking"})

	for i, c := range []*Client{c1, c2} {
		select {
		case raw := <-c.send:
			var got Message
			if err := json.Unmarshal(raw, &got); err != nil {
				t.Fatalf("client %d 反序列化失败: %v", i, err)
			}
			if got.Type != TypeNews {
				t.Fatalf("client %d type=%s want %s", i, got.Type, TypeNews)
			}
		case <-time.After(200 * time.Millisecond):
			t.Fatalf("client %d 未收到广播", i)
		}
	}
}

// TestNilHubSafe 验证 nil Hub 接收者调用各方法均不 panic。
func TestNilHubSafe(t *testing.T) {
	var h *Hub
	h.SendToUser(1, &Message{Type: TypeSystem})
	h.Broadcast(&Message{Type: TypeSystem})
	if got := h.OnlineCount(); got != 0 {
		t.Fatalf("nil Hub OnlineCount=%d want 0", got)
	}
}

// TestCloseWhenNotInitialized 验证 Hub 未初始化时 Close 不 panic。
func TestCloseWhenNotInitialized(t *testing.T) {
	origHub := hub
	t.Cleanup(func() { hub = origHub })
	hub = nil
	Close()
	if IsAvailable() {
		t.Fatal("未初始化时 IsAvailable 应为 false")
	}
}

// TestNewClientFields 验证 NewClient 正确设置字段且 send 通道容量为 64。
func TestNewClientFields(t *testing.T) {
	c := NewClient(7, "alice", nil)
	if c.UserID != 7 {
		t.Fatalf("UserID=%d want 7", c.UserID)
	}
	if c.Username != "alice" {
		t.Fatalf("Username=%s want alice", c.Username)
	}
	if c.Conn() != nil {
		t.Fatal("传入 nil conn 时 Conn() 应为 nil")
	}
	if c.Send() == nil {
		t.Fatal("Send 通道不应为 nil")
	}
	// send 通道容量应为 64：前 64 次写入成功，第 65 次应阻塞/失败
	for i := 0; i < 64; i++ {
		select {
		case c.send <- []byte("x"):
		default:
			t.Fatalf("第 %d 次写入应成功（容量 64）", i)
		}
	}
	select {
	case c.send <- []byte("overflow"):
		t.Fatal("第 65 次写入应失败")
	default:
		// 预期：缓冲已满，写入被丢弃
	}
}

// TestMessageMarshal 验证 Message / LikeNotification 的 JSON 标签正确，
// 确保 C 端能按约定字段解析。
func TestMessageMarshal(t *testing.T) {
	m := &Message{Type: TypeLike, Data: LikeNotification{
		NewsID: 42, NewsTitle: "标题", FromUser: "bob", Liked: true, LikeCount: 3,
	}}
	raw, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("marshal 失败: %v", err)
	}
	var got map[string]interface{}
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal 失败: %v", err)
	}
	if got["type"] != TypeLike {
		t.Fatalf("type=%v want %s", got["type"], TypeLike)
	}
	data, ok := got["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("data 应为对象，got %T", got["data"])
	}
	if data["news_id"] != float64(42) {
		t.Fatalf("news_id=%v want 42", data["news_id"])
	}
	if data["news_title"] != "标题" {
		t.Fatalf("news_title=%v", data["news_title"])
	}
	if data["from_user"] != "bob" {
		t.Fatalf("from_user=%v", data["from_user"])
	}
	if data["liked"] != true {
		t.Fatalf("liked=%v want true", data["liked"])
	}
	if data["like_count"] != float64(3) {
		t.Fatalf("like_count=%v want 3", data["like_count"])
	}
}
