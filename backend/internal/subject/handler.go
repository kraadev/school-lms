package subject

import (
	"errors"
	"net/http"

	"school-lms/internal/models"

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

type CreateSubjectRequest struct {
	Code        string   `json:"code" binding:"required"`
	Name        string   `json:"name" binding:"required"`
	Description *string  `json:"description"`
	TeacherIDs  []string `json:"teacher_ids"`
}

type UpdateSubjectRequest = CreateSubjectRequest

func (h *Handler) GetAll(c *gin.Context) {
	subjects, err := h.Service.GetAll()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Gagal mengambil data mata pelajaran",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    subjects,
	})
}

func (h *Handler) GetByID(c *gin.Context) {
	subject, err := h.Service.GetByID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Mata pelajaran tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": subject})
}

func (h *Handler) Update(c *gin.Context) {
	var request UpdateSubjectRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Data mata pelajaran tidak valid"})
		return
	}
	subject, err := h.Service.Update(c.Param("id"), request.Code, request.Name, request.Description, request.TeacherIDs)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Mata pelajaran berhasil diperbarui", "data": subject})
}

func (h *Handler) Delete(c *gin.Context) {
	if err := h.Service.Delete(c.Param("id")); err != nil {
		var conflict *DeleteConflictError
		if errors.As(err, &conflict) {
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"message": err.Error(),
				"details": gin.H{
					"reason":     "Mata pelajaran masih direferensikan oleh relasi class_subjects.",
					"dependency": conflict.ClassCount,
					"next_steps": []string{
						"Buka daftar kelas yang menggunakan mata pelajaran ini.",
						"Hapus mata pelajaran dari semua kelas terlebih dahulu.",
						"Coba hapus mata pelajaran kembali setelah tidak digunakan oleh kelas.",
					},
				},
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Mata pelajaran berhasil dihapus"})
}

func (h *Handler) Search(c *gin.Context) {
	query := c.Query("q")
	if len(query) < 2 {
		c.JSON(http.StatusOK, gin.H{"success": true, "data": []models.Subject{}})
		return
	}
	subjects, err := h.Service.Search(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Gagal mencari mata pelajaran"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": subjects})
}

func (h *Handler) Create(c *gin.Context) {
	var request CreateSubjectRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Data mata pelajaran tidak valid",
			"error":   err.Error(),
		})
		return
	}

	subject, err := h.Service.Create(
		request.Code,
		request.Name,
		request.Description,
		request.TeacherIDs,
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
		"message": "Mata pelajaran berhasil dibuat",
		"data":    subject,
	})
}
