package models

// PostMessage represents the structure of the message that Naemon (send_push_notification.go) will send to the /message endpoint of the push_notification service.
type PostMessage struct {
	Timestamp        int64  `json:"timestamp"`
	UserId           int64  `json:"userId"`
	Type             string `json:"type"`
	HostUuid         string `json:"hostUuid"`
	ServiceUuid      string `json:"serviceUuid"`
	Icon             string `json:"icon"`
	State            int    `json:"state"`
	NotificationType string `json:"notificationtype"`
	Output           string `json:"output"`
	AckAuthor        string `json:"ackauthor"`
	AckComment       string `json:"ackcomment"`
}
