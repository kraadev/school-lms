package assignmentsubmissionfile

import (
	"fmt"
	"net/http"
	"os"

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

func (h *Handler) DownloadTeacherFile(c *gin.Context) {

	userID := c.GetString("user_id")
	fileID := c.Param("file_id")

	file, err := h.Service.GetTeacherFile(
		userID,
		fileID,
	)

	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	if _, err := os.Stat(file.FilePath); err != nil {

		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "file fisik tidak ditemukan",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "gagal membaca file",
		})

		return
	}

	c.Header(
		"Content-Disposition",
		fmt.Sprintf(
			`attachment; filename="%s"`,
			file.OriginalName,
		),
	)

	c.Header(
		"Content-Type",
		file.MimeType,
	)

	c.File(file.FilePath)
}
