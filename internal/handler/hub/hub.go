package hub

import (
	"Sparrow/internal/model"
	"strconv"
	"sync"
)

// Hub 维护了 userID => set(of channels) 的映射
type Hub struct {
	mu      sync.RWMutex
	clients map[string]map[chan model.Message]struct{}
}

var GlobalHub = New()

// New 创建并返回一个 Hub
func New() *Hub {
	return &Hub{
		clients: make(map[string]map[chan model.Message]struct{}),
	}
}

// Subscribe 用户 userID 注册一个消息通道 ch
func (h *Hub) Subscribe(userID string, ch chan model.Message) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.clients[userID]; !ok {
		h.clients[userID] = make(map[chan model.Message]struct{})
	}
	h.clients[userID][ch] = struct{}{}
}

// Unsubscribe 注销 userID 的某个通道 ch
func (h *Hub) Unsubscribe(userID string, ch chan model.Message) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if conns, ok := h.clients[userID]; ok {
		delete(conns, ch)
		if len(conns) == 0 {
			delete(h.clients, userID)
		}
	}
}

// Broadcast 将一条消息发给它的目标用户
func (h *Hub) Broadcast(msg model.Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	conns, ok := h.clients[strconv.FormatInt(msg.ReceiverID, 10)]
	if !ok {
		return
	}
	for ch := range conns {
		// 尝试发送，避免阻塞
		select {
		case ch <- msg:
		default:
		}
	}
}
