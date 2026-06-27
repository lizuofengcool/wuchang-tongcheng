package middleware

import (
	"net/http"
	"time"

	"wuchang-tongcheng/internal/pkg/jwt"
	"wuchang-tongcheng/internal/pkg/ws"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// wsUpgrader WebSocket 升级器（开发环境允许所有 Origin，生产应配置白名单）
var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 生产环境应校验 Origin 白名单
	},
}

// WebSocketHandler WebSocket 升级处理器
// 鉴权方式：query 参数 ?token=<JWT>（浏览器 WS 握手无法设置自定义头）
// 握手成功后保持长连接，服务端可向该用户推送实时通知（点赞/新闻/系统）。
func WebSocketHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 鉴权：从 query 取 token（WS 无法用 Authorization 头）
		token := c.Query("token")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "缺少 token"})
			return
		}
		claims, err := jwt.ParseToken(token)
		if err != nil || claims.UserID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "token 无效或已过期"})
			return
		}

		// 2. Hub 不可用则拒绝
		hub := ws.GetHub()
		if hub == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "message": "实时服务不可用"})
			return
		}

		// 3. 升级为 WebSocket
		conn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return // Upgrade 已写入错误响应
		}

		// 4. 注册客户端并启动读写泵
		client := ws.NewClient(claims.UserID, claims.Username, conn)
		hub.Register(client)

		go writePump(client)
		readPump(client) // 阻塞直到连接关闭
	}
}

// readPump 读泵：持续读取客户端消息（用于探测连接存活），
// 收到消息暂不处理（后续可扩展为客户端指令），出错即关闭连接。
func readPump(c *ws.Client) {
	defer func() {
		ws.GetHub().Unregister(c)
		_ = c.Conn().Close()
	}()
	c.Conn().SetReadLimit(4096)
	_ = c.Conn().SetReadDeadline(time.Now().Add(60 * time.Second))
	// 客户端可定期发 ping 保活，服务端收到任意消息重置截止时间
	for {
		_, _, err := c.Conn().ReadMessage()
		if err != nil {
			break // 连接关闭/出错
		}
		_ = c.Conn().SetReadDeadline(time.Now().Add(60 * time.Second))
	}
}

// writePump 写泵：从 send 通道取消息写入连接，并定期发 ping 保活。
func writePump(c *ws.Client) {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		_ = c.Conn().Close()
	}()
	for {
		select {
		case msg, ok := <-c.Send():
			_ = c.Conn().SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				_ = c.Conn().WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.Conn().WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.Conn().SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn().WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
