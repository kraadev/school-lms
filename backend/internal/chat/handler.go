package chat

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	Service *Service
	Hub     *Hub
}

func NewHandler(
	service *Service,
	hub *Hub,
) *Handler {

	return &Handler{
		Service: service,
		Hub:     hub,
	}
}

type CreateConversationBody struct {
	Type           string      `json:"type"`
	Name           *string     `json:"name"`
	ClassID        *uuid.UUID  `json:"class_id"`
	ParticipantIDs []uuid.UUID `json:"participant_ids"`
}

type UpdateConversationBody struct {
	Name *string `json:"name"`
}
type InviteStudentBody struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
}

type SendMessageBody struct {
	MessageType string `json:"message_type"`
	Content     string `json:"content" binding:"required"`
}

func (h *Handler) CreateConversation(c *gin.Context) {

	userID := c.GetString("user_id")

	var body CreateConversationBody

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "request tidak valid",
		})
		return
	}

	conversation, err := h.Service.CreateConversation(
		userID,
		CreateConversationRequest{
			Type:           body.Type,
			Name:           body.Name,
			ClassID:        body.ClassID,
			ParticipantIDs: body.ParticipantIDs,
		},
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    conversation,
	})
}

func (h *Handler) UpdateConversation(c *gin.Context) {
	var body UpdateConversationBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "request tidak valid"})
		return
	}
	conversation, err := h.Service.UpdateConversation(c.GetString("user_id"), c.Param("id"), UpdateConversationRequest{Name: body.Name})
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": conversation})
}

func (h *Handler) DeleteConversation(c *gin.Context) {
	if err := h.Service.DeleteConversation(c.GetString("user_id"), c.Param("id")); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Ruang chat dihapus"})
}

func (h *Handler) InviteStudent(c *gin.Context) {
	var body InviteStudentBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "user_id siswa wajib diisi"})
		return
	}
	if err := h.Service.InviteStudent(c.GetString("user_id"), c.Param("id"), body.UserID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "message": "Siswa berhasil diundang"})
}

func (h *Handler) GetChatContacts(c *gin.Context) {
	var classID *uuid.UUID
	if raw := c.Query("class_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "class_id tidak valid"})
			return
		}
		classID = &id
	}
	users, err := h.Service.GetChatContacts(c.GetString("user_id"), c.GetString("role"), classID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Daftar kontak chat belum dapat dimuat"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": users})
}

func (h *Handler) GetTeacherClasses(c *gin.Context) {
	classes, err := h.Service.GetTeacherClasses(c.GetString("user_id"))
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": classes})
}

func (h *Handler) GetConversations(c *gin.Context) {

	userID := c.GetString("user_id")

	conversations, err := h.Service.GetUserConversations(userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    conversations,
	})
}

func (h *Handler) CreateCustomerService(c *gin.Context) {
	conversation, err := h.Service.CreateCustomerService(c.GetString("user_id"))
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": conversation})
}

func (h *Handler) ListCustomerService(c *gin.Context) {
	items, err := h.Service.ListCustomerService(c.GetString("user_id"), c.GetString("role"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": items})
}

func (h *Handler) ClaimCustomerService(c *gin.Context) {
	conversation, err := h.Service.ClaimCustomerService(c.GetString("user_id"), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": conversation})
}

func (h *Handler) CloseCustomerService(c *gin.Context) {
	if err := h.Service.CloseCustomerService(c.GetString("user_id"), c.Param("id")); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Percakapan ditutup"})
}

func (h *Handler) SendMessage(c *gin.Context) {

	userID := c.GetString("user_id")
	conversationID := c.Param("id")

	var body SendMessageBody

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "request tidak valid",
		})
		return
	}

	message, err := h.Service.SendMessage(
		userID,
		conversationID,
		SendMessageRequest{
			MessageType: body.MessageType,
			Content:     body.Content,
		},
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    message,
	})
	data, marshalErr := json.Marshal(map[string]interface{}{"type": "message", "data": message})
	if marshalErr == nil {
		_ = h.Hub.BroadcastToConversation(message.ConversationID, data, h.Service)
	}
}

func (h *Handler) GetMessages(c *gin.Context) {

	userID := c.GetString("user_id")
	conversationID := c.Param("id")

	messages, err := h.Service.GetMessages(
		userID,
		conversationID,
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    messages,
	})
}

func (h *Handler) GetUserPresence(
	c *gin.Context,
) {

	userID := c.Param("user_id")

	isOnline, err := h.Service.GetUserPresence(
		userID,
		h.Hub,
	)

	if err != nil {

		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"success": false,
				"message": err.Error(),
			},
		)

		return
	}

	c.JSON(
		http.StatusOK,
		gin.H{
			"success": true,
			"data": gin.H{
				"user_id":   userID,
				"is_online": isOnline,
			},
		},
	)
}

func (h *Handler) GetUnreadCounts(
	c *gin.Context,
) {

	userID := c.GetString("user_id")

	counts, err := h.Service.GetUnreadCounts(
		userID,
	)

	if err != nil {

		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"success": false,
				"message": err.Error(),
			},
		)

		return
	}

	c.JSON(
		http.StatusOK,
		gin.H{
			"success": true,
			"data":    counts,
		},
	)
}
