package common

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

type ResponsePushNotificationPayload struct {
	Timestamp   int64  `json:"timestamp"`
	UserId      int64  `json:"userId"`
	Title       string `json:"title"`
	Message     string `json:"message"`
	Type        string `json:"type"` // host or service
	HostUuid    string `json:"hostUuid,omitempty"`
	ServiceUuid string `json:"serviceUuid,omitempty"`
	Icon        string `json:"icon,omitempty"`
}
