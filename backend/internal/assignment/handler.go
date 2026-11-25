package assignment

import (
	"net/http"
	"time"

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

type CreateAssignmentRequest struct {
	ClassSubjectID string  `json:"class_subject_id" binding:"required"`
	Title          string  `json:"title" binding:"required"`
	Description    *string `json:"description"`
	Deadline       string  `json:"deadline" binding:"required"`
	MaxScore       int     `json:"max_score" binding:"required"`
}

type UpdateAssignmentRequest struct {
	ClassSubjectID string  `json:"class_subject_id" binding:"required"`
	Title          string  `json:"title" binding:"required"`
	Description    *string `json:"description"`
	Deadline       string  `json:"deadline" binding:"required"`
	MaxScore       int     `json:"max_score" binding:"required"`
}

func parseDeadline(value string) (time.Time, error) {
	if deadline, err := time.Parse(time.RFC3339, value); err == nil {
		return deadline, nil
	}
	return time.ParseInLocation("2006-01-02T15:04", value, time.Local)
}

// GET /api/teacher/assignments
func (h *Handler) GetAll(c *gin.Context) {
	userID := c.GetString("user_id")

	assignments, err := h.Service.GetAllTeacherAssignments(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    assignments,
	})
}

// GET /api/teacher/assignments/:id
func (h *Handler) GetByID(c *gin.Context) {
	userID := c.GetString("user_id")
	assignmentID := c.Param("id")

	assignment, err := h.Service.GetTeacherAssignment(
		userID,
		assignmentID,
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
		"data":    assignment,
	})
}

// POST /api/teacher/assignments
func (h *Handler) Create(c *gin.Context) {
	userID := c.GetString("user_id")

	var req CreateAssignmentRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "data tugas tidak valid",
			"error":   err.Error(),
		})
		return
	}
	deadline, err := parseDeadline(req.Deadline)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "deadline harus menggunakan format waktu yang valid", "error": err.Error()})
		return
	}

	assignment, err := h.Service.Create(
		userID,
		req.ClassSubjectID,
		req.Title,
		req.Description,
		deadline,
		req.MaxScore,
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
		"message": "tugas berhasil dibuat",
		"data":    assignment,
	})
}

// PUT /api/teacher/assignments/:id
func (h *Handler) Update(c *gin.Context) {
	userID := c.GetString("user_id")
	assignmentID := c.Param("id")

	var req UpdateAssignmentRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "data tugas tidak valid",
			"error":   err.Error(),
		})
		return
	}
	deadline, err := parseDeadline(req.Deadline)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "deadline harus menggunakan format waktu yang valid", "error": err.Error()})
		return
	}

	assignment, err := h.Service.Update(
		userID,
		assignmentID,
		req.ClassSubjectID,
		req.Title,
		req.Description,
		deadline,
		req.MaxScore,
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
		"message": "tugas berhasil diperbarui",
		"data":    assignment,
	})
}

// PATCH /api/teacher/assignments/:id/publish
func (h *Handler) Publish(c *gin.Context) {
	userID := c.GetString("user_id")
	assignmentID := c.Param("id")

	assignment, err := h.Service.Publish(userID, assignmentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "tugas berhasil dipublikasikan",
		"data":    assignment,
	})
}

// DELETE /api/teacher/assignments/:id
func (h *Handler) Delete(c *gin.Context) {
	userID := c.GetString("user_id")
	assignmentID := c.Param("id")

	if err := h.Service.Delete(
		userID,
		assignmentID,
	); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "tugas berhasil dihapus",
	})
}

// GET /api/student/assignments
func (h *Handler) GetStudentAssignments(c *gin.Context) {
	userID := c.GetString("user_id")

	assignments, err := h.Service.GetAllStudentAssignments(
		userID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    assignments,
	})
}

// GET /api/student/assignments/:id
func (h *Handler) GetStudentAssignment(c *gin.Context) {
	userID := c.GetString("user_id")
	assignmentID := c.Param("id")

	assignment, err := h.Service.GetStudentAssignment(
		userID,
		assignmentID,
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
		"data":    assignment,
	})
}
