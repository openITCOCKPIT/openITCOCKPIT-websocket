package hub

import (
	"sync"
)

type IncomingMessageType string

const (
	IncomingKeepAlive                       IncomingMessageType = "KeepAlive"                       //  WebSocket clients to keep the connection alive
	IncomingRegisterBrowserPushNotification IncomingMessageType = "RegisterBrowserPushNotification" //  WebSocket clients wants to receive browser push notifications
	IncomingTestPushNotification            IncomingMessageType = "TestPushNotification"            //  WebSocket clients can send this to trigger a test alert (for testing Push Notifications)
)

type IncomingMessage struct {
	Type    IncomingMessageType `json:"type"`
	Message string              `json:"message"`
	Payload any                 `json:"payload"`
}

type IncomingRegisterBrowserPushNotificationPayload struct {
	UserID      int64  `json:"userId"`
	ClientUUID  string `json:"clientUuid"`
	BrowserUUID string `json:"browserUuid"`
}

type IncomingTestPushNotificationPayload struct {
	Title string `json:"title"`
	Icon  string `json:"icon"`
}

type ResponseMessageType string

const (
	ResponseConnectionEstablished   ResponseMessageType = "ConnectionEstablished"   // Sent to client on successful connection with assigned UUID
	ResponseExportStatus            ResponseMessageType = "ExportStatus"            // Sent to clients when export status changes (payload is boolean: true=running, false=stopped)
	ResponseKeepAlive               ResponseMessageType = "KeepAlive"               // Sent to client in response to KeepAlive message (payload can be "Pong" or empty)
	ResponseError                   ResponseMessageType = "Error"                   // Sent to client when an error occurs (payload is error message)
	ResponseProcessPushNotification ResponseMessageType = "ProcessPushNotification" // Sent to client to trigger a push notification (payload contains notification details
)

type ResponseMessage struct {
	Type    ResponseMessageType `json:"type"`
	Message string              `json:"message"`
	Payload any                 `json:"payload"`
}

type ClientInfo struct {
	ClientUUID           string
	BrowserUUID          string
	UserID               int64
	SendPushNotification bool
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

// Run starts the main event loop for the Hub.
// It listens for register, unregister, broadcast, and shutdown events on their respective channels.
// - When a new connection is registered, it is added to the clients map.
// - When a connection is unregistered, it is removed and its send channel is closed.
// - Broadcast messages are sent to all connected clients (dropped if a client is slow).
// - On shutdown, all connections are closed and the clients map is cleared.
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
// It is IMPORTANT to know that this method will send the message to ALL connected clients of the user
// So a message could be send to a browser AND the desktop app.
// if dedub is false, also all browser tabs of the user will receive the message.
// if dedub is true, only one browser tab will receive the message (the first one that is found with the matching userID). This is useful for messages that should not be de-duplicated across multiple browser tabs, such as certain push notifications.
func (h *Hub) SendToUser(userID int64, msg ResponseMessage, dedub bool) {
	h.clientsMutex.RLock()
	defer h.clientsMutex.RUnlock()

	if dedub {
		// De-duplicated message by the browserUuid. Send to all clients of the user, but only one per browserUuid
		sentToBrowserUUIDs := make(map[string]bool)
		for c, info := range h.clients {
			if info.UserID == userID {
				if _, alreadySent := sentToBrowserUUIDs[info.BrowserUUID]; !alreadySent {
					select {
					case c.Send <- msg:
						sentToBrowserUUIDs[info.BrowserUUID] = true
					default:
						// Drop message if client is slow
					}
				}
			}
		}
	} else {
		// Normale message - send to all connected clients of the user
		for c, info := range h.clients {
			if info.UserID == userID {
				select {
				case c.Send <- msg:
				default:
				}
			}
		}
	}
}

// HasClients returns true if at least one client is connected to the hub.
func (h *Hub) HasClients() bool {
	h.clientsMutex.RLock()
	defer h.clientsMutex.RUnlock()
	return len(h.clients) > 0
}

// ClientWantPushNotifications returns true if at least one client of the specified user ID wants to receive push notifications.
func (h *Hub) ClientWantPushNotifications(userID int64) bool {
	h.clientsMutex.RLock()
	defer h.clientsMutex.RUnlock()
	for _, info := range h.clients {
		if info.UserID == userID && info.SendPushNotification {
			return true
		}
	}
	return false
}

func (h *Hub) SetUserPushNotificationPreference(userID int64, enabled bool) {
	h.clientsMutex.Lock()
	defer h.clientsMutex.Unlock()

	for c, info := range h.clients {
		if info.UserID == userID {
			info.SendPushNotification = enabled
			h.clients[c] = info
		}
	}
}

func (h *Hub) UpdateClientInfo(conn *Connection, newInfo ClientInfo) {
	h.clientsMutex.Lock()
	defer h.clientsMutex.Unlock()

	if _, ok := h.clients[conn]; ok {
		h.clients[conn] = newInfo
	}
}
