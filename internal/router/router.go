package router

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"push_notification/internal/db"
	"push_notification/internal/hub"
	"push_notification/internal/webhook"
	"strconv"

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

	// The "z" is a cloud-native Kubernetes convention for a simple health check endpoint. It can be used by Kubernetes liveness/readiness probes.
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	mux.HandleFunc("/message", r.handleMessageInput)
	return mux
}

// MessageInputPayload ist das erwartete JSON für /message
type MessageInputPayload struct {
	UserID  int64       `json:"user_id"`
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

	// FIXME!!
	//msg := hub.ResponseMessage{Type: hub.MessageType(payload.Type), Message: "", Data: payload.Data}
	msg := hub.ResponseMessage{}
	if payload.WebHook && r.webhook != nil {
		go r.webhook.Send(msg)
	} else if payload.UserID > 0 {
		r.hub.SendToUser(payload.UserID, msg)
	} else {
		r.hub.Broadcast(msg)
	}
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("ok"))
}

// handleWebSocket handles the incoming WebSocket connection.
func (r *Router) handleWebSocket(w http.ResponseWriter, req *http.Request) {
	// For authentication, the Client send an API key as a subprotocol during the WebSocket handshake. (HTTP Header)
	// Sec-WebSocket-Protocol: access_token, <API_KEY>
	upgrader := websocket.Upgrader{
		CheckOrigin:  func(r *http.Request) bool { return true },
		Subprotocols: []string{"access_token"},
	}

	// The userId is expected as a query parameter in the WebSocket URL, e.g. ws://localhost:8083/?userId=123
	userIdStr := req.URL.Query().Get("userId")
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract subprotocol (API key) from the request header
	var apiKey string
	requestedProtocols := websocket.Subprotocols(req)
	if len(requestedProtocols) == 2 && requestedProtocols[0] == "access_token" {
		apiKey = requestedProtocols[1]
	}

	if apiKey == "" || userId <= 0 || !r.isValid(apiKey, req.Context()) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// We have to tell the browser which subprotocol we accepted (in this case "access_token")
	// This is mandatory!
	header := http.Header{}
	header.Set("Sec-WebSocket-Protocol", requestedProtocols[0]) // We accepted "access_token"

	ws, err := upgrader.Upgrade(w, req, header)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	defer ws.Close()

	assignedUUID := uuid.NewString()
	welcomeMsg := hub.ResponseMessage{
		Type:    hub.ResponseConnectionEstablished,
		Message: "Connection established successfully",
		Payload: map[string]string{
			"clientUuid": assignedUUID,
		},
	}

	if err := ws.WriteJSON(welcomeMsg); err != nil {
		log.Println("Failed to send welcome message:", err)
		return
	}

	// Register the connection with the hub so we can keep track of connected clients and send messages to them.
	conn := &hub.Connection{
		Ws:   ws,
		Hub:  r.hub,
		Info: hub.ClientInfo{UUID: assignedUUID, UserID: userId},
		Send: make(chan hub.ResponseMessage, 256),
	}
	r.hub.Register(conn)

	// Fire up the WritePump in a new goroutine. It listens for outgoing messages on the Send channel and writes them to the WebSocket,
	// as well as sending periodic ping messages to keep the connection alive.
	go conn.WritePump()

	// The ReadPump method is called in the current goroutine and continuously reads incoming messages from the WebSocket
	// handling them as needed (such as responding to keep-alive messages or processing client commands).
	// When ReadPump exits (for example, if the connection is closed or an error occurs), it ensures the connection is unregistered and cleaned up.
	conn.ReadPump()
}

func (r *Router) isValid(key string, ctx context.Context) bool {
	if r.db != nil {
		validKey, err := r.db.GetAPIKey(ctx)
		if err != nil {
			log.Println("DB error while validating API key:", err)
			return false
		}
		return key == validKey
	}

	return false
}
