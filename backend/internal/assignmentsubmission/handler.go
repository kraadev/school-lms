package assignmentsubmission

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	Service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		Service: service,
	}
}

// DTO untuk merapikan response JSON
type SubmissionFileResponse struct {
	ID           uuid.UUID `json:"id"`
	OriginalName string    `json:"original_name"`
	MimeType     string    `json:"mime_type"`
	FileSize     int64     `json:"file_size"`
}

type AssignmentSummaryResponse struct {
	ID       uuid.UUID `json:"id"`
	Title    string    `json:"title"`
	Deadline time.Time `json:"deadline"`
	MaxScore int       `json:"max_score"`
	Status   string    `json:"status"`
}

type StudentSubmissionResponse struct {
	ID           uuid.UUID  `json:"id"`
	AssignmentID uuid.UUID  `json:"assignment_id"`
	Status       string     `json:"status"`
	SubmittedAt  *time.Time `json:"submitted_at"`
	Score        *float64   `json:"score"`
	Feedback     *string    `json:"feedback"`

	Files []SubmissionFileResponse `json:"files"`

	Assignment AssignmentSummaryResponse `json:"assignment"`
}

type TeacherSubmissionListResponse struct {
	ID          uuid.UUID  `json:"id"`
	StudentID   uuid.UUID  `json:"student_id"`
	NIS         string     `json:"nis"`
	StudentName string     `json:"student_name"`
	Status      string     `json:"status"`
	SubmittedAt *time.Time `json:"submitted_at"`
	Score       *float64   `json:"score"`
	FileCount   int        `json:"file_count"`
}

type TeacherSubmissionDetailResponse struct {
	ID          uuid.UUID `json:"id"`
	StudentID   uuid.UUID `json:"student_id"`
	NIS         string    `json:"nis"`
	StudentName string    `json:"student_name"`

	Status      string     `json:"status"`
	SubmittedAt *time.Time `json:"submitted_at"`

	Score    *float64 `json:"score"`
	Feedback *string  `json:"feedback"`

	Files      []SubmissionFileResponse  `json:"files"`
	Assignment AssignmentSummaryResponse `json:"assignment"`
}

type GradeSubmissionRequest struct {
	Score    *float64 `json:"score" binding:"required"`
	Feedback *string  `json:"feedback"`
}

// POST /api/student/assignments/:id/submit
func (h *Handler) Submit(c *gin.Context) {

	userID := c.GetString("user_id")
	assignmentID := c.Param("id")

	file, err := c.FormFile("file")

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "file wajib diupload",
		})
		return
	}

	submission, err := h.Service.Submit(
		userID,
		assignmentID,
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
		"message": "tugas berhasil dikumpulkan",
		"data":    submission,
	})
}

// GET /api/student/assignments/:id/submission
func (h *Handler) GetStudentSubmission(c *gin.Context) {

	userID := c.GetString("user_id")
	assignmentID := c.Param("id")

	submission, err := h.Service.GetStudentSubmission(
		userID,
		assignmentID,
	)

	if err != nil {

		if err.Error() == "belum ada pengumpulan" {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    nil,
			})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})

		return
	}

	// Map data ke format DTO
	files := make([]SubmissionFileResponse, 0)

	for _, file := range submission.Files {
		files = append(files, SubmissionFileResponse{
			ID:           file.ID,
			OriginalName: file.OriginalName,
			MimeType:     file.MimeType,
			FileSize:     file.FileSize,
		})
	}

	response := StudentSubmissionResponse{
		ID:           submission.ID,
		AssignmentID: submission.AssignmentID,
		Status:       submission.Status,
		SubmittedAt:  submission.SubmittedAt,
		Score:        submission.Score,
		Feedback:     submission.Feedback,
		Files:        files,

		Assignment: AssignmentSummaryResponse{
			ID:       submission.Assignment.ID,
			Title:    submission.Assignment.Title,
			Deadline: submission.Assignment.Deadline,
			MaxScore: submission.Assignment.MaxScore,
			Status:   submission.Assignment.Status,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// GET /api/teacher/assignments/:id/submissions
func (h *Handler) GetTeacherSubmissions(c *gin.Context) {

	userID := c.GetString("user_id")
	assignmentID := c.Param("id")

	submissions, err := h.Service.GetTeacherSubmissions(
		userID,
		assignmentID,
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	data := make([]TeacherSubmissionListResponse, 0)

	for _, submission := range submissions {

		data = append(data, TeacherSubmissionListResponse{
			ID:          submission.ID,
			StudentID:   submission.StudentID,
			NIS:         submission.Student.NIS,
			StudentName: submission.Student.FullName,
			Status:      submission.Status,
			SubmittedAt: submission.SubmittedAt,
			Score:       submission.Score,
			FileCount:   len(submission.Files),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

func (h *Handler) GetTeacherSubmission(c *gin.Context) {
	userID := c.GetString("user_id")
	submissionID := c.Param("id")

	submission, err := h.Service.GetTeacherSubmission(userID, submissionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	files := make([]SubmissionFileResponse, 0, len(submission.Files))
	for _, file := range submission.Files {
		files = append(files, SubmissionFileResponse{
			ID:           file.ID,
			OriginalName: file.OriginalName,
			MimeType:     file.MimeType,
			FileSize:     file.FileSize,
		})
	}

	response := TeacherSubmissionDetailResponse{
		ID:          submission.ID,
		StudentID:   submission.StudentID,
		NIS:         submission.Student.NIS,
		StudentName: submission.Student.FullName,
		Status:      submission.Status,
		SubmittedAt: submission.SubmittedAt,
		Score:       submission.Score,
		Feedback:    submission.Feedback,
		Files:       files,
		Assignment: AssignmentSummaryResponse{
			ID:       submission.Assignment.ID,
			Title:    submission.Assignment.Title,
			Deadline: submission.Assignment.Deadline,
			MaxScore: submission.Assignment.MaxScore,
			Status:   submission.Assignment.Status,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

func (h *Handler) GradeSubmission(c *gin.Context) {
	userID := c.GetString("user_id")
	submissionID := c.Param("id")

	var request GradeSubmissionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "format request tidak valid",
		})
		return
	}

	submission, err := h.Service.GradeSubmission(
		userID,
		submissionID,
		*request.Score,
		request.Feedback,
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
		"message": "submission berhasil dinilai",
		"data": gin.H{
			"id":       submission.ID,
			"status":   submission.Status,
			"score":    submission.Score,
			"feedback": submission.Feedback,
		},
	})
}

func (h *Handler) DeleteStudentSubmissionFile(c *gin.Context) {
	userID := c.GetString("user_id")
	fileID := c.Param("file_id")

	err := h.Service.DeleteStudentSubmissionFile(userID, fileID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "file submission berhasil dihapus",
	})
}
