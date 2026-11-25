package quiz

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

type CreateQuizRequest struct {
	ClassSubjectID uuid.UUID `json:"class_subject_id" binding:"required"`

	Title       string  `json:"title" binding:"required"`
	Description *string `json:"description"`

	Duration int `json:"duration" binding:"required"`

	StartAt *time.Time `json:"start_at"`
	EndAt   *time.Time `json:"end_at"`

	MaxAttempts int `json:"max_attempts" binding:"required"`
}

// POST /api/teacher/quizzes
func (h *Handler) CreateQuiz(c *gin.Context) {
	userID := c.GetString("user_id")

	var req CreateQuizRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "data request tidak valid",
		})
		return
	}

	quiz, err := h.Service.CreateQuiz(
		userID,
		req.ClassSubjectID,
		req.Title,
		req.Description,
		req.Duration,
		req.StartAt,
		req.EndAt,
		req.MaxAttempts,
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
		"message": "quiz berhasil dibuat",
		"data":    quiz,
	})
}

// GET /api/teacher/quizzes
func (h *Handler) GetTeacherQuizzes(c *gin.Context) {
	userID := c.GetString("user_id")

	quizzes, err := h.Service.GetTeacherQuizzes(userID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    quizzes,
	})
}

// GET /api/teacher/quizzes/:id
func (h *Handler) GetTeacherQuiz(c *gin.Context) {
	userID := c.GetString("user_id")
	quizID := c.Param("id")

	quiz, err := h.Service.GetTeacherQuiz(
		userID,
		quizID,
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
		"data":    quiz,
	})
}

type UpdateQuizRequest struct {
	Title       string  `json:"title" binding:"required"`
	Description *string `json:"description"`

	Duration int `json:"duration" binding:"required"`

	StartAt *time.Time `json:"start_at"`
	EndAt   *time.Time `json:"end_at"`

	MaxAttempts int `json:"max_attempts" binding:"required"`
}

// PUT /api/teacher/quizzes/:id
func (h *Handler) UpdateQuiz(c *gin.Context) {
	userID := c.GetString("user_id")
	quizID := c.Param("id")

	var req UpdateQuizRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "data request tidak valid",
		})
		return
	}

	quiz, err := h.Service.UpdateQuiz(
		userID,
		quizID,
		req.Title,
		req.Description,
		req.Duration,
		req.StartAt,
		req.EndAt,
		req.MaxAttempts,
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
		"message": "quiz berhasil diperbarui",
		"data":    quiz,
	})
}

// DELETE /api/teacher/quizzes/:id
func (h *Handler) DeleteQuiz(c *gin.Context) {
	userID := c.GetString("user_id")
	quizID := c.Param("id")

	err := h.Service.DeleteQuiz(
		userID,
		quizID,
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
		"message": "quiz berhasil dihapus",
	})
}

// POST /api/teacher/quizzes/:id/publish
func (h *Handler) PublishQuiz(c *gin.Context) {
	userID := c.GetString("user_id")
	quizID := c.Param("id")

	quiz, err := h.Service.PublishQuiz(
		userID,
		quizID,
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
		"message": "quiz berhasil dipublish",
		"data":    quiz,
	})
}

// POST /api/teacher/quizzes/:id/close
func (h *Handler) CloseQuiz(c *gin.Context) {
	userID := c.GetString("user_id")
	quizID := c.Param("id")

	quiz, err := h.Service.CloseQuiz(
		userID,
		quizID,
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
		"message": "quiz berhasil ditutup",
		"data":    quiz,
	})
}

type CreateQuestionRequest struct {
	Question string  `json:"question" binding:"required"`
	Type     string  `json:"type" binding:"required"`
	Points   float64 `json:"points" binding:"required"`
	Order    int     `json:"order" binding:"required"`
}

// POST /api/teacher/quizzes/:id/questions
func (h *Handler) CreateQuestion(c *gin.Context) {
	userID := c.GetString("user_id")
	quizID := c.Param("id")

	var req CreateQuestionRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "data request tidak valid",
		})
		return
	}

	question, err := h.Service.CreateQuestion(
		userID,
		quizID,
		req.Question,
		req.Type,
		req.Points,
		req.Order,
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
		"message": "question berhasil dibuat",
		"data":    question,
	})
}

// GET /api/teacher/quizzes/:id/questions
func (h *Handler) GetTeacherQuizQuestions(c *gin.Context) {
	userID := c.GetString("user_id")
	quizID := c.Param("id")

	questions, err := h.Service.GetTeacherQuizQuestions(
		userID,
		quizID,
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
		"data":    questions,
	})
}

// GET /api/teacher/questions/:id
func (h *Handler) GetTeacherQuestion(c *gin.Context) {
	userID := c.GetString("user_id")
	questionID := c.Param("id")

	question, err := h.Service.GetTeacherQuestion(
		userID,
		questionID,
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
		"data":    question,
	})
}

type UpdateQuestionRequest struct {
	Question string  `json:"question" binding:"required"`
	Type     string  `json:"type" binding:"required"`
	Points   float64 `json:"points" binding:"required"`
	Order    int     `json:"order" binding:"required"`
}

// PUT /api/teacher/questions/:id
func (h *Handler) UpdateQuestion(c *gin.Context) {
	userID := c.GetString("user_id")
	questionID := c.Param("id")

	var req UpdateQuestionRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "data request tidak valid",
		})
		return
	}

	question, err := h.Service.UpdateQuestion(
		userID,
		questionID,
		req.Question,
		req.Type,
		req.Points,
		req.Order,
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
		"message": "question berhasil diperbarui",
		"data":    question,
	})
}

// DELETE /api/teacher/questions/:id
func (h *Handler) DeleteQuestion(c *gin.Context) {
	userID := c.GetString("user_id")
	questionID := c.Param("id")

	err := h.Service.DeleteQuestion(
		userID,
		questionID,
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
		"message": "question berhasil dihapus",
	})
}

// --- Tambahan untuk Question Option ---

type CreateQuestionOptionRequest struct {
	Text      string `json:"text" binding:"required"`
	IsCorrect bool   `json:"is_correct"`
	Order     int    `json:"order" binding:"required"`
}

// POST /api/teacher/questions/:id/options
func (h *Handler) CreateQuestionOption(c *gin.Context) {
	userID := c.GetString("user_id")
	questionID := c.Param("id")

	var req CreateQuestionOptionRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "data request tidak valid",
		})
		return
	}

	option, err := h.Service.CreateQuestionOption(
		userID,
		questionID,
		req.Text,
		req.IsCorrect,
		req.Order,
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
		"message": "option berhasil dibuat",
		"data":    option,
	})
}

// GET /api/teacher/questions/:id/options
func (h *Handler) GetTeacherQuestionOptions(c *gin.Context) {
	userID := c.GetString("user_id")
	questionID := c.Param("id")

	options, err := h.Service.GetTeacherQuestionOptions(
		userID,
		questionID,
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
		"data":    options,
	})
}

type UpdateQuestionOptionRequest struct {
	Text      string `json:"text" binding:"required"`
	IsCorrect bool   `json:"is_correct"`
	Order     int    `json:"order" binding:"required"`
}

// PUT /api/teacher/options/:id
func (h *Handler) UpdateQuestionOption(c *gin.Context) {
	userID := c.GetString("user_id")
	optionID := c.Param("id")

	var req UpdateQuestionOptionRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "data request tidak valid",
		})
		return
	}

	option, err := h.Service.UpdateQuestionOption(
		userID,
		optionID,
		req.Text,
		req.IsCorrect,
		req.Order,
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
		"message": "option berhasil diperbarui",
		"data":    option,
	})
}

// DELETE /api/teacher/options/:id
func (h *Handler) DeleteQuestionOption(c *gin.Context) {
	userID := c.GetString("user_id")
	optionID := c.Param("id")

	err := h.Service.DeleteQuestionOption(
		userID,
		optionID,
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
		"message": "option berhasil dihapus",
	})
}

// =============================================================================
// PENAMBAHAN UNTUK STUDENT
// =============================================================================

type StudentQuizResponse struct {
	ID             uuid.UUID  `json:"id"`
	Title          string     `json:"title"`
	Description    *string    `json:"description"`
	Duration       int        `json:"duration"`
	StartAt        *time.Time `json:"start_at"`
	EndAt          *time.Time `json:"end_at"`
	MaxAttempts    int        `json:"max_attempts"`
	AttemptsUsed   int64      `json:"attempts_used"`
	Status         string     `json:"status"`
	ClassSubjectID uuid.UUID  `json:"class_subject_id"`
}

// GET /api/student/quizzes
func (h *Handler) GetStudentQuizzes(c *gin.Context) {
	userID := c.GetString("user_id")

	quizzes, err := h.Service.GetStudentQuizzes(userID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	data := make([]StudentQuizResponse, 0)

	for _, quiz := range quizzes {

		if !IsQuizAvailable(&quiz) {
			continue
		}

		attemptsUsed, err := h.Service.CountStudentQuizAttempts(userID, quiz.ID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}

		data = append(data, StudentQuizResponse{
			ID:             quiz.ID,
			Title:          quiz.Title,
			Description:    quiz.Description,
			Duration:       quiz.Duration,
			StartAt:        quiz.StartAt,
			EndAt:          quiz.EndAt,
			MaxAttempts:    quiz.MaxAttempts,
			AttemptsUsed:   attemptsUsed,
			Status:         quiz.Status,
			ClassSubjectID: quiz.ClassSubjectID,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

type StudentQuizOptionResponse struct {
	ID    uuid.UUID `json:"id"`
	Text  string    `json:"text"`
	Order int       `json:"order"`
}

type StudentQuizQuestionResponse struct {
	ID       uuid.UUID                   `json:"id"`
	Question string                      `json:"question"`
	Type     string                      `json:"type"`
	Points   float64                     `json:"points"`
	Order    int                         `json:"order"`
	Options  []StudentQuizOptionResponse `json:"options"`
}

type StudentQuizDetailResponse struct {
	ID             uuid.UUID                     `json:"id"`
	Title          string                        `json:"title"`
	Description    *string                       `json:"description"`
	Duration       int                           `json:"duration"`
	StartAt        *time.Time                    `json:"start_at"`
	EndAt          *time.Time                    `json:"end_at"`
	MaxAttempts    int                           `json:"max_attempts"`
	AttemptsUsed   int64                         `json:"attempts_used"`
	Status         string                        `json:"status"`
	ClassSubjectID uuid.UUID                     `json:"class_subject_id"`
	Questions      []StudentQuizQuestionResponse `json:"questions"`
}

// GET /api/student/quizzes/:id
func (h *Handler) GetStudentQuiz(c *gin.Context) {
	userID := c.GetString("user_id")
	quizID := c.Param("id")

	quiz, err := h.Service.GetStudentQuiz(
		userID,
		quizID,
	)

	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	questions := make(
		[]StudentQuizQuestionResponse,
		0,
	)

	for _, question := range quiz.Questions {

		options := make(
			[]StudentQuizOptionResponse,
			0,
		)

		for _, option := range question.Options {

			options = append(
				options,
				StudentQuizOptionResponse{
					ID:    option.ID,
					Text:  option.Text,
					Order: option.Order,
				},
			)
		}

		questions = append(
			questions,
			StudentQuizQuestionResponse{
				ID:       question.ID,
				Question: question.Question,
				Type:     question.Type,
				Points:   question.Points,
				Order:    question.Order,
				Options:  options,
			},
		)
	}

	attemptsUsed, err := h.Service.CountStudentQuizAttempts(userID, quiz.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	response := StudentQuizDetailResponse{
		ID:             quiz.ID,
		Title:          quiz.Title,
		Description:    quiz.Description,
		Duration:       quiz.Duration,
		StartAt:        quiz.StartAt,
		EndAt:          quiz.EndAt,
		MaxAttempts:    quiz.MaxAttempts,
		AttemptsUsed:   attemptsUsed,
		Status:         quiz.Status,
		ClassSubjectID: quiz.ClassSubjectID,
		Questions:      questions,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

type BulkCreateQuestionsRequest struct {
	Questions []BulkQuestionInput `json:"questions" binding:"required"`
}

// POST /api/teacher/quizzes/:id/questions/bulk
func (h *Handler) BulkCreateQuestions(c *gin.Context) {
	userID := c.GetString("user_id")
	quizID := c.Param("id")

	var req BulkCreateQuestionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "data request tidak valid",
		})
		return
	}

	questions, err := h.Service.BulkCreateQuestions(userID, quizID, req.Questions)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "soal-soal berhasil diimpor",
		"data":    questions,
	})
}
