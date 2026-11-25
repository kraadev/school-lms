package academicyear

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct{ Service *Service }

func NewHandler(service *Service) *Handler { return &Handler{Service: service} }

func (h *Handler) GetAll(c *gin.Context) {
	years, err := h.Service.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Gagal mengambil tahun akademik"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": years})
}

func (h *Handler) GetActive(c *gin.Context) {
	year, err := h.Service.GetActive()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Belum ada tahun akademik aktif"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": year})
}

func (h *Handler) Create(c *gin.Context) {
	var req SaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Data tahun akademik tidak valid"})
		return
	}
	year, err := h.Service.Create(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "message": "Tahun akademik berhasil dibuat", "data": year})
}

func (h *Handler) Update(c *gin.Context) {
	var req SaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Data tahun akademik tidak valid"})
		return
	}
	year, err := h.Service.Update(c.Param("id"), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Tahun akademik berhasil diperbarui", "data": year})
}

func (h *Handler) Duplicate(c *gin.Context) {
	var req DuplicateRequest
	if err := c.ShouldBindJSON(&req); err != nil && err.Error() != "EOF" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Data duplikasi tahun akademik tidak valid"})
		return
	}
	year, err := h.Service.Duplicate(c.Param("id"), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "message": "Tahun akademik berhasil diduplikasi", "data": year})
}

func (h *Handler) Delete(c *gin.Context) {
	if err := h.Service.Delete(c.Param("id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Tahun akademik berhasil dihapus"})
}

func (h *Handler) CreateSemester(c *gin.Context) {
	var req SemesterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Data semester tidak valid"})
		return
	}
	semester, err := h.Service.CreateSemester(c.Param("id"), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "message": "Semester berhasil dibuat", "data": semester})
}

func (h *Handler) UpdateSemester(c *gin.Context) {
	var req SemesterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Data semester tidak valid"})
		return
	}
	semester, err := h.Service.UpdateSemester(c.Param("id"), c.Param("semester_id"), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Semester berhasil diperbarui", "data": semester})
}

func (h *Handler) DeleteSemester(c *gin.Context) {
	if err := h.Service.DeleteSemester(c.Param("id"), c.Param("semester_id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Semester berhasil dihapus"})
}
