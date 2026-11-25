package academicagenda

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct{ Service *Service }

func NewHandler(service *Service) *Handler { return &Handler{Service: service} }

func (h *Handler) List(c *gin.Context) {
	items, err := h.Service.List(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": items})
}
func (h *Handler) Create(c *gin.Context) {
	var req SaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Data agenda tidak valid"})
		return
	}
	var createdByID *uuid.UUID
	if userID := c.GetString("user_id"); userID != "" {
		if parsed, parseErr := uuid.Parse(userID); parseErr == nil {
			createdByID = &parsed
		}
	}
	item, err := h.Service.Create(c.Param("id"), req, createdByID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "message": "Agenda berhasil dibuat", "data": item})
}
func (h *Handler) Update(c *gin.Context) {
	var req SaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Data agenda tidak valid"})
		return
	}
	item, err := h.Service.Update(c.Param("id"), c.Param("agenda_id"), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Agenda berhasil diperbarui", "data": item})
}
func (h *Handler) Delete(c *gin.Context) {
	if err := h.Service.Delete(c.Param("id"), c.Param("agenda_id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Agenda dihapus"})
}
