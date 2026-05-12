package hub

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 8192
)

type Connection struct {
	Ws   *websocket.Conn
	Hub  *Hub
	Info ClientInfo
	Send chan ResponseMessage
}

// ReadPump listens for incoming messages from the WebSocket connection.
// It runs in its own goroutine for each client.
//
// - Reads JSON messages from the client in a loop.
// - Handles WebSocket pings/pongs for keepalive.
// - On error or disconnect, unregisters the connection and closes the socket.
func (c *Connection) ReadPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Ws.Close()
	}()
	c.Ws.SetReadLimit(maxMessageSize)
	c.Ws.SetReadDeadline(time.Now().Add(pongWait))
	c.Ws.SetPongHandler(func(string) error {
		c.Ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		var msg IncomingMessage
		if err := c.Ws.ReadJSON(&msg); err != nil {
			log.Println("read error:", err)
			break
		}

		switch msg.Type {
		case IncomingKeepAlive:
			// KeepAlive message - respond with "Pong"
			response := ResponseMessage{
				Type:    ResponseKeepAlive,
				Message: "Pong",
			}
			c.Send <- response

		case IncomingRegisterBrowserPushNotification:
			var payload IncomingRegisterBrowserPushNotificationPayload
			if err := mapToStruct(msg.Payload, &payload); err != nil {
				log.Println("Invalid payload for RegisterBrowserPushNotification:", err)
				continue
			}

			//log.Println("Enable Browser Push Notifications for this Client")
			c.Hub.SetUserPushNotificationPreference(c.Info.UserID, true)

		case IncomingTestPushNotification:
			log.Println("Received TestPushNotification command from client, sending test alert")
			var payload IncomingTestPushNotificationPayload
			if err := mapToStruct(msg.Payload, &payload); err != nil {
				log.Println("Invalid payload for IncomingTestPushNotificationPayload:", err)
				continue
			}

			testMsg := ResponseMessage{
				Type:    ResponseProcessPushNotification,
				Message: "This is a test push notification triggered by the client",
				Payload: ResponsePushNotificationPayload{
					Timestamp: time.Now().Unix(),
					UserId:    c.Info.UserID,
					Title:     payload.Title,
					Message:   payload.Title,
					Type:      "host", // host or service
					Icon:      payload.Icon,
				},
			}
			c.Hub.SendToUser(c.Info.UserID, testMsg, true)

		default:
			log.Println("Unknown message type:", msg.Type)
			response := ResponseMessage{
				Type:    ResponseError,
				Message: "Unknown message type: " + string(msg.Type),
			}
			c.Send <- response
		}
	}
}

// WritePump sends outgoing messages to the WebSocket client.
// It runs in its own goroutine for each client.
//
// - Listens on the Send channel for messages to deliver to the client.
// - Periodically sends WebSocket ping frames to keep the connection alive.
// - On error or disconnect, closes the socket.
func (c *Connection) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Ws.Close()
	}()
	for {
		select {
		case msg, ok := <-c.Send:
			c.Ws.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Ws.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.Ws.WriteJSON(msg); err != nil {
				return
			}
		case <-ticker.C:
			c.Ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Ws.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// mapToStruct decodes a map[string]interface{} (or any JSON-like map) into a struct using JSON marshal/unmarshal.
func mapToStruct(input any, out any) error {
	b, err := json.Marshal(input)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, out)
}
