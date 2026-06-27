// Package ws WebSocket 连接管理
// 提供基于用户的实时推送 Hub：单用户多连接、定向推送、全局广播。
// 用于点赞通知、新闻推送等实时场景。
package ws

import (
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
)

// MessageType 通知类型
const (
	TypeLike   = "like"   // 点赞通知
	TypeNews   = "news"   // 新闻/头条通知
	TypeSystem = "system" // 系统通知
)

// Message 推送给客户端的消息协议
type Message struct {
	Type string      `json:"type"` // like | news | system
	Data interface{} `json:"data"`
}

// LikeNotification 点赞通知负载
type LikeNotification struct {
	NewsID    uint   `json:"news_id"`
	NewsTitle string `json:"news_title"`
	FromUser  string `json:"from_user"` // 点赞者用户名
	Liked     bool   `json:"liked"`     // true=点赞 false=取消
	LikeCount int    `json:"like_count"`
}

// Client 单个 WebSocket 连接
type Client struct {
	UserID   uint
	Username string
	conn     *websocket.Conn
	send     chan []byte
	hub      *Hub
}

// Conn 返回底层 WebSocket 连接（供读写泵使用）
func (c *Client) Conn() *websocket.Conn { return c.conn }

// Send 返回发送通道（供写泵消费）
func (c *Client) Send() <-chan []byte { return c.send }

// Hub 连接管理器，单用户可多连接（多端登录）
type Hub struct {
	mu         sync.RWMutex
	clients    map[uint]map[*Client]struct{} // userID -> clients set
	register   chan *Client
	unregister chan *Client
	broadcast   chan []byte
}

var (
	hub *Hub
)

// Init 初始化 Hub 并启动事件循环
func Init() {
	hub = &Hub{
		clients:    make(map[uint]map[*Client]struct{}),
		register:   make(chan *Client, 64),
		unregister: make(chan *Client, 64),
		broadcast:  make(chan []byte, 256),
	}
	go hub.Run()
}

// GetHub 获取 Hub 实例
func GetHub() *Hub {
	return hub
}

// IsAvailable 检查 Hub 是否已初始化
func IsAvailable() bool {
	return hub != nil
}

// Close 关闭 Hub（清空所有连接）
func Close() {
	if hub == nil {
		return
	}
	hub.mu.Lock()
	defer hub.mu.Unlock()
	for userID, clients := range hub.clients {
		for c := range clients {
			close(c.send)
			_ = c.conn.Close()
		}
		delete(hub.clients, userID)
	}
}

// NewClient 创建客户端连接
func NewClient(userID uint, username string, conn *websocket.Conn) *Client {
	return &Client{
		UserID:   userID,
		Username: username,
		conn:     conn,
		send:     make(chan []byte, 64),
		hub:      hub,
	}
}

// Register 注册客户端到 Hub
func (h *Hub) Register(c *Client) {
	h.register <- c
}

// Unregister 注销客户端
func (h *Hub) Unregister(c *Client) {
	h.unregister <- c
}

// Run 事件循环：处理注册/注销/广播
func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			h.mu.Lock()
			if h.clients[c.UserID] == nil {
				h.clients[c.UserID] = make(map[*Client]struct{})
			}
			h.clients[c.UserID][c] = struct{}{}
			h.mu.Unlock()
		case c := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.clients[c.UserID]; ok {
				if _, exists := clients[c]; exists {
					delete(clients, c)
					close(c.send)
					if len(clients) == 0 {
						delete(h.clients, c.UserID)
					}
				}
			}
			h.mu.Unlock()
		case msg := <-h.broadcast:
			h.mu.RLock()
			for _, clients := range h.clients {
				for c := range clients {
					select {
					case c.send <- msg:
					default:
						// 发送缓冲满，丢弃该客户端（后续会被读泵剔除）
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// SendToUser 向指定用户的所有连接推送消息（fire-and-forget：连接已关闭则丢弃）
func (h *Hub) SendToUser(userID uint, msg *Message) {
	if h == nil {
		return
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	clients, ok := h.clients[userID]
	if !ok {
		return // 用户无在线连接，丢弃
	}
	for c := range clients {
		select {
		case c.send <- data:
		default:
			// 缓冲满，丢弃
		}
	}
}

// Broadcast 全局广播
func (h *Hub) Broadcast(msg *Message) {
	if h == nil {
		return
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	h.broadcast <- data
}

// OnlineCount 在线连接总数
func (h *Hub) OnlineCount() int {
	if h == nil {
		return 0
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	count := 0
	for _, clients := range h.clients {
		count += len(clients)
	}
	return count
}
