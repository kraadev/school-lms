package quizanswer

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

func (h *Handler) SaveAnswer(c *gin.Context) {

	userID := c.GetString("user_id")

	attemptID := c.Param("attempt_id")
	questionID := c.Param("question_id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "unauthorized",
		})
		return
	}

	var req SaveAnswerRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "format request tidak valid",
		})
		return
	}

	answer, err := h.Service.SaveAnswer(
		userID,
		attemptID,
		questionID,
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
		"message": "jawaban berhasil disimpan",
		"data": gin.H{
			"id":           answer.ID,
			"attempt_id":   answer.AttemptID,
			"question_id":  answer.QuestionID,
			"option_id":    answer.OptionID,
			"essay_answer": answer.EssayAnswer,
			"answered_at":  answer.AnsweredAt,
		},
	})
}
