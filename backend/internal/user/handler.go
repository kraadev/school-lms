package user

import (
	"errors"
	"log"
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

func (h *Handler) GetAll(c *gin.Context) {
	users, err := h.Service.GetAll()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Gagal mengambil data user",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    users,
	})
}

func (h *Handler) GetProfile(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Autentikasi diperlukan"})
		return
	}
	user, err := h.Service.GetProfile(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Profil pengguna tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": user})
}

func (h *Handler) UpdateProfile(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Autentikasi diperlukan"})
		return
	}
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Data profil tidak valid: " + err.Error()})
		return
	}
	user, err := h.Service.UpdateProfile(userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Profil berhasil diperbarui", "data": user})
}

func (h *Handler) GetByID(c *gin.Context) {
	user, err := h.Service.GetByID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "User tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": user})
}

func (h *Handler) Update(c *gin.Context) {
	var request UpdateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Data user tidak valid"})
		return
	}
	user, err := h.Service.Update(c.Param("id"), request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "User berhasil diperbarui", "data": user})
}

func (h *Handler) Delete(c *gin.Context) {
	if err := h.Service.Delete(c.Param("id")); err != nil {
		var conflict *DeleteConflictError
		if errors.As(err, &conflict) {
			log.Printf("PENGHAPUSAN DITOLAK: user %s adalah guru dengan %d materi. Pindahkan atau hapus materi terlebih dahulu sebelum menghapus guru.", conflict.UserID, conflict.MaterialCount)
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"message": err.Error(),
				"details": gin.H{
					"reason":     "Materi masih mereferensikan guru ini melalui foreign key fk_materials_teacher.",
					"dependency": conflict.MaterialCount,
					"next_steps": []string{
						"Buka daftar materi milik guru tersebut.",
						"Pindahkan materi ke guru lain atau hapus materi yang tidak diperlukan.",
						"Coba hapus akun guru kembali setelah semua materi tidak lagi mereferensikannya.",
					},
				},
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "User berhasil dihapus"})
}

func (h *Handler) GetStudents(c *gin.Context) {
	students, err := h.Service.GetStudents()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Gagal mengambil data siswa",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    students,
	})
}

func (h *Handler) SearchStudents(c *gin.Context) {
	query := c.Query("q")
	if len(query) < 2 {
		c.JSON(http.StatusOK, gin.H{"success": true, "data": []models.Student{}})
		return
	}
	students, err := h.Service.SearchStudents(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Gagal mencari siswa"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": students})
}

type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	Email    string `json:"email"`
	Role     string `json:"role" binding:"required"`

	FullName string `json:"full_name" binding:"required"`

	NIP   *string `json:"nip"`
	Phone *string `json:"phone"`

	NIS     string  `json:"nis"`
	NISN    *string `json:"nisn"`
	Gender  *string `json:"gender"`
	Address *string `json:"address"`
}

func (h *Handler) Create(c *gin.Context) {
	var request CreateUserRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Data user tidak valid",
			"error":   err.Error(),
		})
		return
	}

	user, err := h.Service.Create(request)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "User berhasil dibuat",
		"data":    user,
	})
}
