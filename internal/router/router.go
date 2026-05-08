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

// MessageInputPayload defines the expected JSON structure for incoming messages to the /message endpoint.
type PostMessage struct {
	Timestamp   int64  `json:"timestamp"`
	UserId      int64  `json:"userId"`
	Title       string `json:"title"`
	Message     string `json:"message"`
	Type        string `json:"type"`
	HostUuid    string `json:"hostUuid"`
	ServiceUuid string `json:"serviceUuid"`
	Icon        string `json:"icon"`
	State       int64  `json:"state"` // 0=OK, 1=Warning, 2=Critical, 3=Unknown
}

// NewRouter kann optional einen WebHookService und eine DB erhalten
func NewRouter(h *hub.Hub, wh *webhook.WebHookService, dbConn *db.DB) http.Handler {
	r := &Router{hub: h, webhook: wh, db: dbConn}
	mux := http.NewServeMux()

	// The root path "/" is used for WebSocket connections.
	mux.HandleFunc("/", r.handleWebSocket)

	// The "/healthz" endpoint is a simple health check that returns "ok". This can be used by Kubernetes or other monitoring tools to check if the service is running.
	// The "z" is a cloud-native Kubernetes convention for a simple health check endpoint. It can be used by Kubernetes liveness/readiness probes.
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// The "/message" endpoint accepts POST requests with a JSON body to send messages to users or broadcast to all clients
	// This is used by Naemon to trigger push notifications
	mux.HandleFunc("/message", r.handleMessageInput)
	return mux
}

// handleMessageInput handles incoming POST requests to the /message endpoint.
func (r *Router) handleMessageInput(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var data PostMessage
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&data); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if data.UserId <= 0 {
		http.Error(w, "Invalid userId", http.StatusBadRequest)
		return
	}

	if !r.hub.ClientWantPushNotifications(data.UserId) {
		http.Error(w, "User is not registered for push notifications", http.StatusBadRequest)
		return
	}

	// FIXME!!
	//msg := hub.ResponseMessage{Type: hub.MessageType(payload.Type), Message: "", Data: payload.Data}
	msg := hub.ResponseMessage{}
	//go r.webhook.Send(msg)
	r.hub.SendToUser(data.UserId, msg, false)

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("ok"))
}

// handleWebSocket handles the incoming WebSocket connection. ("/")
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
	browserUuid := req.URL.Query().Get("browserUuid")

	// Extract subprotocol (API key) from the request header
	var apiKey string
	requestedProtocols := websocket.Subprotocols(req)
	if len(requestedProtocols) == 2 && requestedProtocols[0] == "access_token" {
		apiKey = requestedProtocols[1]
	}

	if browserUuid == "" || apiKey == "" || userId <= 0 || !r.isValid(apiKey, req.Context()) {
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
		Ws:  ws,
		Hub: r.hub,
		Info: hub.ClientInfo{
			ClientUUID:           assignedUUID, // The uuid of the websocket client
			BrowserUUID:          browserUuid,  // used for de-duplication if the user has multiple browser tabs open (push notifications)
			UserID:               userId,       // the user_id of the connected client
			SendPushNotification: false,        // We only send push notifications, if the client explicitly registers for it
		},
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

// isValid checks if the provided API key is valid by comparing it against the expected key from the database.
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
