package quizattempt

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

func (h *Handler) Start(c *gin.Context) {

	userID := c.GetString("user_id")

	quizID := c.Param("id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "unauthorized",
		})
		return
	}

	attempt, err := h.Service.Start(
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

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "quiz berhasil dimulai",
		"data":    attempt,
	})
}

func (h *Handler) Submit(c *gin.Context) {

	userID := c.GetString("user_id")
	attemptID := c.Param("attempt_id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "unauthorized",
		})
		return
	}

	attempt, err := h.Service.Submit(
		userID,
		attemptID,
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
		"message": "quiz berhasil disubmit",
		"data":    attempt,
	})
}

func (h *Handler) GetLatestStudentAttempt(c *gin.Context) {
	userID := c.GetString("user_id")
	quizID := c.Param("id")

	attempt, err := h.Service.GetLatestStudentAttempt(userID, quizID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    attempt,
	})
}

func (h *Handler) GetTeacherAttempts(c *gin.Context) {

	userID := c.GetString("user_id")
	quizID := c.Param("id")

	attempts, err := h.Service.GetTeacherAttempts(
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
		"data":    attempts,
	})
}

func (h *Handler) GetTeacherAttempt(c *gin.Context) {

	userID := c.GetString("user_id")
	attemptID := c.Param("attempt_id")

	attempt, err := h.Service.GetTeacherAttempt(
		userID,
		attemptID,
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
		"data":    attempt,
	})
}

func (h *Handler) GradeEssay(c *gin.Context) {

	userID := c.GetString("user_id")

	attemptID := c.Param("attempt_id")
	answerID := c.Param("answer_id")

	var req GradeEssayRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "data grading tidak valid",
		})
		return
	}

	attempt, err := h.Service.GradeEssay(
		userID,
		attemptID,
		answerID,
		*req.Score,
		req.Feedback,
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
		"message": "jawaban essay berhasil dinilai",
		"data":    attempt,
	})
}

func (h *Handler) GetStudentQuizResult(c *gin.Context) {

	userID := c.GetString("user_id")
	attemptID := c.Param("attempt_id")

	result, err := h.Service.GetStudentQuizResult(
		userID,
		attemptID,
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
		"data":    result,
	})
}
