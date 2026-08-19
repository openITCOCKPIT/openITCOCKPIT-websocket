package messaging

import (
	"push_notification/internal/common"
	"push_notification/internal/hub"
)

type MessagingService struct {
	hub *hub.Hub
}

func NewMessagingService(h *hub.Hub) *MessagingService {
	return &MessagingService{hub: h}
}

func (m *MessagingService) Broadcast(msg common.ResponseMessage) {
	m.hub.Broadcast(msg)
}

// SendToUser sends a message to a specific user by their user ID.
func (m *MessagingService) SendToUser(userID int64, msg common.ResponseMessage) {
	m.hub.SendToUser(userID, msg, false)
}

// SendToUserDeDeuplicated sends a message to a specific user, but is intended for messages that should not be de-duplicated across multiple browser tabs.
func (m *MessagingService) SendToUserDeDeuplicated(userID int64, msg common.ResponseMessage) {
	m.hub.SendToUser(userID, msg, true)
}
