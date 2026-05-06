package router

import (
	"encoding/json"
	"log"
	"net/http"
	"push_notification/internal/db"
	"push_notification/internal/hub"
	"push_notification/internal/webhook"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Router struct {
	hub     *hub.Hub
	webhook *webhook.WebHookService
	db      *db.DB
}

// NewRouter kann optional einen WebHookService und eine DB erhalten
func NewRouter(h *hub.Hub, wh *webhook.WebHookService, dbConn *db.DB) http.Handler {
	r := &Router{hub: h, webhook: wh, db: dbConn}
	mux := http.NewServeMux()
	mux.HandleFunc("/", r.handleWebSocket)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	mux.HandleFunc("/message", r.handleMessageInput)
	return mux
}

// MessageInputPayload ist das erwartete JSON für /message
type MessageInputPayload struct {
	UserID  string      `json:"user_id"`
	Type    string      `json:"type"`
	Data    interface{} `json:"data"`
	WebHook bool        `json:"webhook,omitempty"` // Optional: explizit an WebHook senden
}

// handleMessageInput nimmt POST-JSON entgegen und verteilt die Nachricht
func (r *Router) handleMessageInput(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var payload MessageInputPayload
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	msg := hub.ResponseMessage{Type: hub.MessageType(payload.Type), Message: "", Data: payload.Data}
	if payload.WebHook && r.webhook != nil {
		go r.webhook.Send(msg)
	} else if payload.UserID != "" {
		r.hub.SendToUser(payload.UserID, msg)
	} else {
		r.hub.Broadcast(msg)
	}
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("ok"))
}

type handshakePayload struct {
	Task string `json:"task"`
	Key  string `json:"key"`
	UUID string `json:"uuid"`
	Data struct {
		UserID      int64  `json:"userId"`
		BrowserUUID string `json:"browserUuid"`
	} `json:"data"`
}

func (r *Router) handleWebSocket(w http.ResponseWriter, req *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	ws, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	// onOpen: Begrüßungsnachricht sofort senden, UUID generieren
	assignedUUID := uuid.NewString()
	openMsg := map[string]interface{}{
		"type":    hub.TypeConnection,
		"message": "Connection established successfully",
		"data": map[string]string{
			"uuid": assignedUUID,
		},
	}
	if err := ws.WriteJSON(openMsg); err != nil {
		log.Println("Failed to send onOpen message:", err)
		ws.Close()
		return
	}

	// Jetzt Auth/Handshake vom Client erwarten
	var payload handshakePayload
	_, msg, err := ws.ReadMessage()
	if err != nil {
		log.Println("Handshake read error:", err)
		ws.Close()
		return
	}
	if err := json.Unmarshal(msg, &payload); err != nil {
		log.Println("Handshake JSON error:", err)
		ws.Close()
		return
	}
	// API-Key validieren
	if r.db != nil {
		ctx := req.Context()
		validKey, err := r.db.GetAPIKey(ctx)
		if err != nil || payload.Key != validKey {
			log.Println("Invalid API key")
			ws.WriteMessage(websocket.CloseMessage, []byte("Invalid API key"))
			ws.Close()
			return
		}
	}
	// Jetzt Connection registrieren
	conn := &hub.Connection{
		Ws:   ws,
		Hub:  r.hub,
		Info: hub.ClientInfo{UUID: assignedUUID, UserID: string(payload.Data.UserID)},
		Send: make(chan hub.ResponseMessage, 256),
	}
	r.hub.Register(conn)
	go conn.WritePump()
	conn.ReadPump() // onMessage and onClose is handled inside ReadPump
}
