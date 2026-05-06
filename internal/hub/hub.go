package hub

import (
	"sync"
)

type MessageType string

const (
	TypeConnection MessageType = "connection" // Response from the WebSocket server on successful connection
	TypeMessage    MessageType = "message"    // General message the WebSocket client should handle
	TypeKeepAlive  MessageType = "keepAlive"  // Optional: for WebSocket clients to keep the connection alive
)

type Message struct {
	Task MessageType `json:"task"`
	Key  string      `json:"key"`
	UUID string      `json:"uuid"`
	Data any         `json:"data"`
}

type ResponseMessage struct {
	Type    MessageType `json:"type"`
	Message string      `json:"message"`
	Data    any         `json:"data,omitempty"`
}

type ClientInfo struct {
	UUID   string
	UserID string
}

type Hub struct {
	clients      map[*Connection]ClientInfo
	clientsMutex sync.RWMutex
	register     chan *Connection
	unregister   chan *Connection
	broadcast    chan ResponseMessage
	shutdown     chan struct{}
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Connection]ClientInfo),
		register:   make(chan *Connection),
		unregister: make(chan *Connection),
		broadcast:  make(chan ResponseMessage, 256),
		shutdown:   make(chan struct{}),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case conn := <-h.register:
			h.clientsMutex.Lock()
			h.clients[conn] = conn.Info
			h.clientsMutex.Unlock()
		case conn := <-h.unregister:
			h.clientsMutex.Lock()
			if _, ok := h.clients[conn]; ok {
				delete(h.clients, conn)
				close(conn.Send)
			}
			h.clientsMutex.Unlock()
		case msg := <-h.broadcast:
			h.clientsMutex.RLock()
			for c := range h.clients {
				select {
				case c.Send <- msg:
				default:
					// Drop message if client is slow
				}
			}
			h.clientsMutex.RUnlock()
		case <-h.shutdown:
			h.clientsMutex.Lock()
			for c := range h.clients {
				close(c.Send)
			}
			h.clients = make(map[*Connection]ClientInfo)
			h.clientsMutex.Unlock()
			return
		}
	}
}

func (h *Hub) Shutdown() {
	close(h.shutdown)
}

func (h *Hub) Register(conn *Connection) {
	h.register <- conn
}

func (h *Hub) Unregister(conn *Connection) {
	h.unregister <- conn
}

func (h *Hub) Broadcast(msg ResponseMessage) {
	h.broadcast <- msg
}

// Send targeted message
func (h *Hub) SendToUser(userID string, msg ResponseMessage) {
	h.clientsMutex.RLock()
	for c, info := range h.clients {
		if info.UserID == userID {
			select {
			case c.Send <- msg:
			default:
			}
		}
	}
	h.clientsMutex.RUnlock()
}
