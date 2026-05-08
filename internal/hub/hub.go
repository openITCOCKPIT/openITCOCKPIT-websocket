package hub

import (
	"sync"
)

type IncomingMessageType string

const (
	IncomingKeepAlive IncomingMessageType = "KeepAlive" //  WebSocket clients to keep the connection alive
)

type IncomingMessage struct {
	Type    IncomingMessageType `json:"task"`
	Message string              `json:"message"`
	Payload any                 `json:"payload"`
}

type ResponseMessageType string

const (
	ResponseConnectionEstablished ResponseMessageType = "ConnectionEstablished" // Sent to client on successful connection with assigned UUID
	ResponseExportStatus          ResponseMessageType = "ExportStatus"
	ResponseKeepAlive             ResponseMessageType = "KeepAlive"
)

type ResponseMessage struct {
	Type    ResponseMessageType `json:"type"`
	Message string              `json:"message"`
	Payload any                 `json:"payload"`
}

type ClientInfo struct {
	UUID   string
	UserID int64
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
func (h *Hub) SendToUser(userID int64, msg ResponseMessage) {
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
