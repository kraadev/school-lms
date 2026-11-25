package notification

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		Service: service,
	}
}

func (h *Handler) GetNotifications(c *gin.Context) {
	userID := c.GetString("user_id")

	notifications, err := h.Service.GetUserNotifications(userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    notifications,
	})
}

func (h *Handler) GetUnreadCount(c *gin.Context) {
	userID := c.GetString("user_id")

	count, err := h.Service.GetUnreadCount(userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"unread_count": count,
		},
	})
}

func (h *Handler) MarkAsRead(c *gin.Context) {
	userID := c.GetString("user_id")
	notificationID := c.Param("id")

	err := h.Service.MarkAsRead(
		userID,
		notificationID,
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
		"message": "notifikasi ditandai sebagai sudah dibaca",
	})
}

func (h *Handler) MarkAllAsRead(c *gin.Context) {
	userID := c.GetString("user_id")

	err := h.Service.MarkAllAsRead(userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "semua notifikasi ditandai sebagai sudah dibaca",
	})
}
