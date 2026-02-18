package handler

import (
	"Sparrow/internal/handler/hub"
	"Sparrow/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/lampctl/go-sse"
	"strconv"
)

// Notifies SSE 推送 handler
// 客户端访问： GET /notify/:token
func Notifies(c *gin.Context) {
	userID := c.GetString("userID")

	//给每个连接都开一个带缓冲的 channel
	msgCh := make(chan model.Message, 16)
	// 注册到全局 Hub
	hub.GlobalHub.Subscribe(userID, msgCh)
	defer func() {
		hub.GlobalHub.Unsubscribe(userID, msgCh)
		close(msgCh)
	}()

	h := sse.NewHandler(nil)

	go func() {
		for msg := range msgCh {
			h.Send(&sse.Event{
				Type: strconv.Itoa(int(msg.Type)),
				Data: msg.Data,
				ID:   strconv.FormatInt(msg.SenderID, 10),
			})
		}
		h.Close()
	}()

	h.ServeHTTP(c.Writer, c.Request)
}
