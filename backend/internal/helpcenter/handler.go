package helpcenter

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct{ Service *Service }

func NewHandler(service *Service) *Handler { return &Handler{Service: service} }

func (h *Handler) List(c *gin.Context) {
	items, err := h.Service.List(c.GetString("role"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": items})
}
func (h *Handler) Create(c *gin.Context) {
	var input FAQInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "pertanyaan dan jawaban wajib diisi"})
		return
	}
	item, err := h.Service.Create(c.GetString("user_id"), input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": item})
}
func (h *Handler) Update(c *gin.Context) {
	var input FAQInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "request FAQ tidak valid"})
		return
	}
	item, err := h.Service.Update(c.Param("id"), input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": item})
}
func (h *Handler) Delete(c *gin.Context) {
	if err := h.Service.Delete(c.Param("id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}
