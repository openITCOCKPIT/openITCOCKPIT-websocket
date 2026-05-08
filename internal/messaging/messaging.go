package messaging

import (
	"push_notification/internal/hub"
)

type MessagingService struct {
	hub *hub.Hub
}

func NewMessagingService(h *hub.Hub) *MessagingService {
	return &MessagingService{hub: h}
}

func (m *MessagingService) Broadcast(msg hub.ResponseMessage) {
	m.hub.Broadcast(msg)
}

func (m *MessagingService) SendToUser(userID string, msg hub.ResponseMessage) {
	m.hub.SendToUser(userID, msg)
}
