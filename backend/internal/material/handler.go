package material

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

type CreateMaterialRequest struct {
	ClassSubjectID string  `json:"class_subject_id" binding:"required"`
	Title          string  `json:"title" binding:"required"`
	Description    *string `json:"description"`

	VideoURL    *string `json:"video_url"`
	ExternalURL *string `json:"external_url"`

	Status string `json:"status"`
}

type UpdateMaterialRequest struct {
	Title       string  `json:"title" binding:"required"`
	Description *string `json:"description"`

	VideoURL    *string `json:"video_url"`
	ExternalURL *string `json:"external_url"`

	Status string `json:"status" binding:"required"`
}

func (h *Handler) Create(c *gin.Context) {

	var request CreateMaterialRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Data materi tidak valid",
			"error":   err.Error(),
		})
		return
	}

	userID := c.GetString("user_id")

	material, err := h.Service.Create(
		userID,
		request.ClassSubjectID,
		request.Title,
		request.Description,
		request.VideoURL,
		request.ExternalURL,
		request.Status,
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
		"message": "Materi berhasil dibuat",
		"data":    material,
	})
}

func (h *Handler) GetMyMaterials(c *gin.Context) {
	userID := c.GetString("user_id")

	materials, err := h.Service.GetMyMaterials(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    materials,
	})
}

func (h *Handler) GetMyMaterialByID(c *gin.Context) {
	userID := c.GetString("user_id")
	materialID := c.Param("id")

	material, err := h.Service.GetMyMaterialByID(userID, materialID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    material,
	})
}

func (h *Handler) Update(c *gin.Context) {
	userID := c.GetString("user_id")
	materialID := c.Param("id")

	var request UpdateMaterialRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Data materi tidak valid",
			"error":   err.Error(),
		})
		return
	}

	material, err := h.Service.Update(
		userID,
		materialID,
		request.Title,
		request.Description,
		request.VideoURL,
		request.ExternalURL,
		request.Status,
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
		"message": "Materi berhasil diperbarui",
		"data":    material,
	})
}

func (h *Handler) Publish(c *gin.Context) {
	userID := c.GetString("user_id")
	materialID := c.Param("id")

	material, err := h.Service.Publish(userID, materialID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Materi berhasil dipublikasikan",
		"data":    material,
	})
}

func (h *Handler) Delete(c *gin.Context) {
	userID := c.GetString("user_id")
	materialID := c.Param("id")

	err := h.Service.Delete(userID, materialID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Materi berhasil dihapus",
	})
}

func (h *Handler) GetStudentMaterials(c *gin.Context) {
	userID := c.GetString("user_id")

	materials, err := h.Service.GetStudentMaterials(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    materials,
	})
}

func (h *Handler) GetStudentMaterialByID(c *gin.Context) {
	userID := c.GetString("user_id")
	materialID := c.Param("id")

	material, err := h.Service.GetStudentMaterialByID(userID, materialID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    material,
	})
}
