package ws

import "testing"

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
