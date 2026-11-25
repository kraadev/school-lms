package attendance

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

func (h *Handler) CreateSession(c *gin.Context) {
	var req CreateAttendanceSessionRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "request tidak valid",
			"error":   err.Error(),
		})
		return
	}

	userID := c.GetString("user_id")

	session, err := h.Service.CreateSession(userID, req)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "attendance session berhasil dibuat",
		"data":    session,
	})
}

func (h *Handler) GetSessions(c *gin.Context) {
	userID := c.GetString("user_id")

	sessions, err := h.Service.GetTeacherSessions(userID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    sessions,
	})
}

func (h *Handler) GetSession(c *gin.Context) {
	userID := c.GetString("user_id")
	sessionID := c.Param("id")

	session, err := h.Service.GetTeacherSession(
		userID,
		sessionID,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    session,
	})
}

func (h *Handler) CloseSession(c *gin.Context) {
	userID := c.GetString("user_id")
	sessionID := c.Param("id")

	session, err := h.Service.CloseSession(
		userID,
		sessionID,
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
		"message": "attendance session berhasil ditutup",
		"data":    session,
	})
}

func (h *Handler) OpenSession(c *gin.Context) {
	userID := c.GetString("user_id")
	sessionID := c.Param("id")

	session, err := h.Service.OpenSession(
		userID,
		sessionID,
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
		"message": "attendance session berhasil dibuka",
		"data":    session,
	})
}

func (h *Handler) DeleteSession(c *gin.Context) {
	if err := h.Service.DeleteSession(c.GetString("user_id"), c.Param("id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "attendance session berhasil dihapus"})
}

func (h *Handler) GetStudentSessions(c *gin.Context) {
	userID := c.GetString("user_id")

	sessions, err := h.Service.GetStudentSessions(userID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    sessions,
	})
}

func (h *Handler) GetStudentSession(c *gin.Context) {
	userID := c.GetString("user_id")
	sessionID := c.Param("id")

	session, err := h.Service.GetStudentSession(
		userID,
		sessionID,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    session,
	})
}

func (h *Handler) CheckIn(c *gin.Context) {
	userID := c.GetString("user_id")
	sessionID := c.Param("id")

	record, err := h.Service.CheckIn(
		userID,
		sessionID,
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
		"message": "check-in berhasil",
		"data":    record,
	})
}

func (h *Handler) GetAttendanceRecords(c *gin.Context) {
	userID := c.GetString("user_id")
	sessionID := c.Param("id")

	records, err := h.Service.GetTeacherAttendanceRecords(
		userID,
		sessionID,
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
		"data":    records,
	})
}

func (h *Handler) UpdateAttendanceRecord(c *gin.Context) {
	var req UpdateAttendanceRecordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "request tidak valid",
			"error":   err.Error(),
		})
		return
	}

	userID := c.GetString("user_id")
	recordID := c.Param("record_id")

	record, err := h.Service.UpdateAttendanceRecord(
		userID,
		recordID,
		req,
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
		"message": "attendance record berhasil diperbarui",
		"data":    record,
	})
}

func (h *Handler) GetAttendanceRecap(c *gin.Context) {
	userID := c.GetString("user_id")
	classID := c.Query("class_id")

	if classID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "class_id wajib diisi",
		})
		return
	}

	recap, err := h.Service.GetAttendanceRecap(
		userID,
		classID,
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
		"data":    recap,
	})
}

func (h *Handler) GetAttendanceStatistics(c *gin.Context) {
	userID := c.GetString("user_id")
	classID := c.Query("class_id")

	if classID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "class_id wajib diisi",
		})
		return
	}

	statistics, err := h.Service.GetAttendanceStatistics(
		userID,
		classID,
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
		"data":    statistics,
	})
}

func (h *Handler) GetStudentAttendanceStatistics(c *gin.Context) {
	userID := c.GetString("user_id")

	statistics, err := h.Service.GetStudentAttendanceStatistics(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    statistics,
	})
}
