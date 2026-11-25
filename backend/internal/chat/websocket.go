package chat

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"school-lms/internal/auth"
	"school-lms/internal/models"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,

	// Browser mengirim:
	// Sec-WebSocket-Protocol:
	// access_token, <JWT>
	//
	// Server akan memilih "access_token"
	// sebagai subprotocol yang digunakan.
	Subprotocols: []string{
		"access_token",
	},

	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebSocketMessageRequest struct {
	Type           string   `json:"type"`
	ConversationID string   `json:"conversation_id"`
	MessageType    string   `json:"message_type"`
	Content        string   `json:"content"`
	IsTyping       bool     `json:"is_typing"`
	MessageIDs     []string `json:"message_ids"`
}

type WebSocketConnection struct {
	*websocket.Conn
}

func NewWebSocketConnection(
	conn *websocket.Conn,
) *WebSocketConnection {
	return &WebSocketConnection{
		Conn: conn,
	}
}

type WebSocketHandler struct {
	Hub       *Hub
	Service   *Service
	JWTSecret string
}

func NewWebSocketHandler(
	hub *Hub,
	service *Service,
	jwtSecret string,
) *WebSocketHandler {
	return &WebSocketHandler{
		Hub:       hub,
		Service:   service,
		JWTSecret: jwtSecret,
	}
}

// authenticateRequest membaca JWT dari:
//
// Sec-WebSocket-Protocol:
// access_token, <JWT>
//
// Browser WebSocket tidak dapat mengirim
// Authorization header secara langsung.
func (h *WebSocketHandler) authenticateRequest(
	c *gin.Context,
) (uuid.UUID, error) {

	protocolHeader := c.GetHeader(
		"Sec-WebSocket-Protocol",
	)

	if protocolHeader == "" {
		return uuid.Nil, jwt.ErrTokenMalformed
	}

	parts := strings.Split(
		protocolHeader,
		",",
	)

	var tokenString string

	for _, part := range parts {
		part = strings.TrimSpace(part)

		// Support juga jika suatu saat client
		// mengirim Bearer <token>.
		if strings.HasPrefix(part, "Bearer ") {
			tokenString = strings.TrimSpace(
				strings.TrimPrefix(part, "Bearer "),
			)
			break
		}
	}

	// Format utama frontend:
	//
	// access_token, JWT
	//
	if tokenString == "" && len(parts) >= 2 {
		if strings.TrimSpace(parts[0]) == "access_token" {
			tokenString = strings.TrimSpace(parts[1])
		}
	}

	if tokenString == "" {
		return uuid.Nil, jwt.ErrTokenMalformed
	}

	claims := &auth.Claims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (interface{}, error) {

			// Hanya izinkan HMAC.
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}

			return []byte(h.JWTSecret), nil
		},
	)

	if err != nil || !token.Valid {
		return uuid.Nil, jwt.ErrTokenInvalidClaims
	}

	if claims.UserID == uuid.Nil {
		return uuid.Nil, jwt.ErrTokenInvalidClaims
	}

	return claims.UserID, nil
}

func (h *WebSocketHandler) Connect(
	c *gin.Context,
) {
	// ==========================================
	// 1. AUTHENTICATION
	// ==========================================

	userID, err := h.authenticateRequest(c)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "token WebSocket tidak valid atau sudah expired",
		})
		return
	}

	// ==========================================
	// 2. UPGRADE HTTP -> WEBSOCKET
	// ==========================================

	conn, err := upgrader.Upgrade(
		c.Writer,
		c.Request,
		nil,
	)

	if err != nil {
		return
	}

	// ==========================================
	// 3. CREATE CLIENT
	// ==========================================

	wsConn := NewWebSocketConnection(conn)

	client := &Client{
		UserID:  userID,
		Conn:    wsConn,
		Send:    make(chan []byte, 256),
		Hub:     h.Hub,
		Service: h.Service,
	}

	// ==========================================
	// 4. REGISTER CLIENT
	// ==========================================

	h.Hub.register <- client

	// ==========================================
	// 5. START WRITE PUMP
	// ==========================================

	go client.writePump()

	// ==========================================
	// 6. START READ PUMP
	// ==========================================

	client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(
		1024 * 1024,
	)

	_ = c.Conn.SetReadDeadline(
		time.Now().Add(pongWait),
	)

	c.Conn.SetPongHandler(
		func(string) error {
			_ = c.Conn.SetReadDeadline(
				time.Now().Add(pongWait),
			)

			return nil
		},
	)

	for {
		_, data, err := c.Conn.ReadMessage()

		if err != nil {
			break
		}

		var request WebSocketMessageRequest

		if err := json.Unmarshal(
			data,
			&request,
		); err != nil {
			c.sendError(
				"format message tidak valid",
			)

			continue
		}

		switch request.Type {

		case "message":
			c.handleMessage(request)

		case "typing":
			c.handleTyping(request)

		case "read":
			c.handleRead(request)

		default:
			c.sendError(
				"type websocket tidak valid",
			)
		}
	}
}

func (c *Client) handleMessage(
	request WebSocketMessageRequest,
) {
	if request.ConversationID == "" {
		c.sendError(
			"conversation_id wajib diisi",
		)

		return
	}

	if request.Content == "" {
		c.sendError(
			"content tidak boleh kosong",
		)

		return
	}

	message, err := c.Service.SendMessage(
		c.UserID.String(),
		request.ConversationID,
		SendMessageRequest{
			MessageType: request.MessageType,
			Content:     request.Content,
		},
	)

	if err != nil {
		c.sendError(
			err.Error(),
		)

		return
	}

	// ==========================================
	// CHAT NOTIFICATION
	// ==========================================

	c.sendChatNotification(message)

	// ==========================================
	// REALTIME MESSAGE
	// ==========================================

	response := map[string]interface{}{
		"type": "message",
		"data": message,
	}

	data, err := json.Marshal(response)
	if err != nil {
		return
	}
	if err := c.Hub.BroadcastToConversation(message.ConversationID, data, c.Service); err != nil {
		c.sendError("gagal broadcast message")
	}
}

func (c *Client) handleTyping(
	request WebSocketMessageRequest,
) {
	if request.ConversationID == "" {
		c.sendError(
			"conversation_id wajib diisi",
		)

		return
	}

	conversationID, err := uuid.Parse(
		request.ConversationID,
	)

	if err != nil {
		c.sendError(
			"conversation id tidak valid",
		)

		return
	}

	if _, err := c.Service.authorizeConversation(c.UserID.String(), request.ConversationID, true); err != nil {
		c.sendError(err.Error())
		return
	}

	response := map[string]interface{}{
		"type":            "typing",
		"conversation_id": request.ConversationID,
		"user_id":         c.UserID.String(),
		"is_typing":       request.IsTyping,
	}

	data, err := json.Marshal(response)

	if err != nil {
		return
	}

	// Broadcast ke participant lain.
	err = c.Hub.BroadcastToConversationExcept(
		conversationID,
		data,
		c.Service,
		c.UserID,
	)

	if err != nil {
		c.sendError(
			"gagal broadcast typing",
		)
	}
}

func (c *Client) handleRead(
	request WebSocketMessageRequest,
) {
	if request.ConversationID == "" {
		c.sendError(
			"conversation_id wajib diisi",
		)

		return
	}

	if len(request.MessageIDs) == 0 {
		c.sendError(
			"message_ids wajib diisi",
		)

		return
	}

	reads, err := c.Service.MarkMessagesAsRead(
		c.UserID.String(),
		request.ConversationID,
		request.MessageIDs,
	)

	if err != nil {
		c.sendError(
			err.Error(),
		)

		return
	}

	if len(reads) == 0 {
		return
	}

	response := map[string]interface{}{
		"type":            "read",
		"conversation_id": request.ConversationID,
		"user_id":         c.UserID.String(),
		"message_ids":     request.MessageIDs,
		"read_at":         time.Now(),
	}

	data, err := json.Marshal(response)

	if err != nil {
		return
	}

	conversationID, err := uuid.Parse(
		request.ConversationID,
	)

	if err != nil {
		return
	}

	// Kirim read receipt ke participant lain.
	err = c.Hub.BroadcastToConversationExcept(
		conversationID,
		data,
		c.Service,
		c.UserID,
	)

	if err != nil {
		c.sendError(
			"gagal broadcast read receipt",
		)
	}
}

func (c *Client) sendChatNotification(
	message *models.Message,
) {
	response := map[string]interface{}{
		"type":              "notification",
		"notification_type": "chat_message",
		"conversation_id":   message.ConversationID.String(),
		"message_id":        message.ID.String(),
		"sender_id":         message.SenderID.String(),
	}

	data, err := json.Marshal(response)

	if err != nil {
		return
	}

	err = c.Hub.BroadcastToConversationExcept(
		message.ConversationID,
		data,
		c.Service,
		c.UserID,
	)

	if err != nil {
		return
	}
}

func (c *Client) sendError(
	message string,
) {
	response := map[string]interface{}{
		"type":    "error",
		"message": message,
	}

	data, err := json.Marshal(response)

	if err != nil {
		return
	}

	select {
	case c.Send <- data:
	default:
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {

		case message, ok := <-c.Send:

			_ = c.Conn.SetWriteDeadline(
				time.Now().Add(writeWait),
			)

			if !ok {
				_ = c.Conn.WriteMessage(
					websocket.CloseMessage,
					[]byte{},
				)

				return
			}

			if err := c.Conn.WriteMessage(
				websocket.TextMessage,
				message,
			); err != nil {
				return
			}

		case <-ticker.C:

			_ = c.Conn.SetWriteDeadline(
				time.Now().Add(writeWait),
			)

			if err := c.Conn.WriteMessage(
				websocket.PingMessage,
				nil,
			); err != nil {
				return
			}
		}
	}
}
