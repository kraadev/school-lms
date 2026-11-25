package materialfile

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Service *Service
}

func NewHandler(
	service *Service,
) *Handler {
	return &Handler{
		Service: service,
	}
}

func (h *Handler) Upload(
	c *gin.Context,
) {

	materialID := c.Param("id")

	file, err := c.FormFile("file")

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "file wajib diupload",
		})
		return
	}

	userID := c.GetString("user_id")

	materialFile, err := h.Service.Upload(
		userID,
		materialID,
		file,
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
		"message": "file berhasil diupload",
		"data":    materialFile,
	})
}

func (h *Handler) GetFiles(c *gin.Context) {
	userID := c.GetString("user_id")
	materialID := c.Param("id")

	files, err := h.Service.GetFiles(userID, materialID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    files,
	})
}

func (h *Handler) DeleteFile(c *gin.Context) {
	userID := c.GetString("user_id")
	fileID := c.Param("file_id")

	err := h.Service.DeleteFile(userID, fileID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "file berhasil dihapus",
	})
}

func (h *Handler) DownloadStudentFile(c *gin.Context) {
	userID := c.GetString("user_id")
	fileID := c.Param("file_id")

	file, err := h.Service.GetStudentFile(userID, fileID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.Header(
		"Content-Disposition",
		`attachment; filename="`+file.OriginalName+`"`,
	)
	c.Header("Content-Type", file.MimeType)
	c.File(file.FilePath)
}

func (h *Handler) DownloadTeacherFile(c *gin.Context) {
	file, err := h.Service.GetTeacherFile(c.GetString("user_id"), c.Param("file_id"))
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "message": err.Error()})
		return
	}
	if _, err := os.Stat(file.FilePath); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "file fisik tidak ditemukan"})
		return
	}
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, file.OriginalName))
	c.Header("Content-Type", file.MimeType)
	c.File(file.FilePath)
}
