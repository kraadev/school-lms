package major

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct{ Service *Service }

func NewHandler(service *Service) *Handler { return &Handler{Service: service} }

type SaveRequest struct {
	Code        string  `json:"code" binding:"required"`
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
}

func (h *Handler) GetAll(c *gin.Context) {
	items, err := h.Service.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Gagal mengambil jurusan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": items})
}

func (h *Handler) GetByID(c *gin.Context) {
	item, err := h.Service.GetByID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Jurusan tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": item})
}

func (h *Handler) Create(c *gin.Context) {
	var request SaveRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Data jurusan tidak valid"})
		return
	}
	item, err := h.Service.Create(request.Code, request.Name, request.Description)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "message": "Jurusan berhasil dibuat", "data": item})
}

func (h *Handler) Update(c *gin.Context) {
	var request SaveRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Data jurusan tidak valid"})
		return
	}
	item, err := h.Service.Update(c.Param("id"), request.Code, request.Name, request.Description)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Jurusan berhasil diperbarui", "data": item})
}

func (h *Handler) Delete(c *gin.Context) {
	if err := h.Service.Delete(c.Param("id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Jurusan berhasil dihapus"})
}
