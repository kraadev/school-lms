package chat

import (
	"encoding/json"
	"sync"

	"github.com/google/uuid"

	"school-lms/internal/models"
)

type Client struct {
	UserID uuid.UUID
	Conn   *WebSocketConnection
	Send   chan []byte

	Hub     *Hub
	Service *Service
}

type Hub struct {
	clients map[uuid.UUID]map[*Client]bool

	register   chan *Client
	unregister chan *Client

	mu sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[uuid.UUID]map[*Client]bool),

		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {

		case client := <-h.register:

			h.mu.Lock()

			wasOffline := false

			if len(h.clients[client.UserID]) == 0 {
				wasOffline = true
			}

			if h.clients[client.UserID] == nil {
				h.clients[client.UserID] = make(
					map[*Client]bool,
				)
			}

			h.clients[client.UserID][client] = true

			h.mu.Unlock()

			if wasOffline {
				go h.BroadcastPresence(
					client.UserID,
					true,
				)
			}

		case client := <-h.unregister:

			h.mu.Lock()

			becameOffline := false

			if userClients, exists := h.clients[client.UserID]; exists {

				if _, ok := userClients[client]; ok {

					delete(
						userClients,
						client,
					)

					close(client.Send)
				}

				if len(userClients) == 0 {

					delete(
						h.clients,
						client.UserID,
					)

					becameOffline = true
				}
			}

			h.mu.Unlock()

			if becameOffline {
				go h.BroadcastPresence(
					client.UserID,
					false,
				)
			}
		}
	}
}

func (h *Hub) BroadcastToConversation(
	conversationID uuid.UUID,
	message []byte,
	service *Service,
) error {

	var participants []models.ConversationParticipant

	err := service.DB.
		Where(
			"conversation_id = ?",
			conversationID,
		).
		Find(&participants).
		Error

	if err != nil {
		return err
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, participant := range participants {

		userClients, exists := h.clients[participant.UserID]

		if !exists {
			continue
		}

		for client := range userClients {

			select {
			case client.Send <- message:

			default:
				// Jika buffer penuh,
				// koneksi dianggap bermasalah.
				go func(c *Client) {
					h.unregister <- c
				}(client)
			}
		}
	}

	return nil
}

func (h *Hub) BroadcastToConversationExcept(
	conversationID uuid.UUID,
	message []byte,
	service *Service,
	exceptUserID uuid.UUID,
) error {

	var participants []models.ConversationParticipant

	err := service.DB.
		Where(
			"conversation_id = ?",
			conversationID,
		).
		Find(&participants).
		Error

	if err != nil {
		return err
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, participant := range participants {

		// Jangan kirim kembali ke pengirim
		if participant.UserID == exceptUserID {
			continue
		}

		userClients, exists := h.clients[participant.UserID]

		if !exists {
			continue
		}

		for client := range userClients {

			select {

			case client.Send <- message:

			default:

				go func(c *Client) {
					h.unregister <- c
				}(client)
			}
		}
	}

	return nil
}

func (h *Hub) IsOnline(
	userID uuid.UUID,
) bool {

	h.mu.RLock()
	defer h.mu.RUnlock()

	userClients, exists := h.clients[userID]

	if !exists {
		return false
	}

	return len(userClients) > 0
}

func (h *Hub) GetOnlineUserIDs() []uuid.UUID {

	h.mu.RLock()
	defer h.mu.RUnlock()

	users := make(
		[]uuid.UUID,
		0,
		len(h.clients),
	)

	for userID, clients := range h.clients {

		if len(clients) > 0 {
			users = append(
				users,
				userID,
			)
		}
	}

	return users
}

func (h *Hub) BroadcastPresence(
	userID uuid.UUID,
	isOnline bool,
) {

	payload := map[string]interface{}{
		"type":      "presence",
		"user_id":   userID.String(),
		"is_online": isOnline,
	}

	data, err := json.Marshal(payload)

	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, userClients := range h.clients {

		for client := range userClients {

			select {
			case client.Send <- data:

			default:
			}
		}
	}
}
