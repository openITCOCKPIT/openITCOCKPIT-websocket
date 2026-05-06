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

func (m *MessagingService) Broadcast(msg hub.Message) {
	m.hub.Broadcast(msg)
}

func (m *MessagingService) SendToUser(userID string, msg hub.Message) {
	m.hub.SendToUser(userID, msg)
}
