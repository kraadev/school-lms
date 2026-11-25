package quizattempt

import (
	"testing"

	"school-lms/internal/models"

	"github.com/google/uuid"
)

func TestCanStudentSeeQuizResult(t *testing.T) {
	if !canStudentSeeQuizResult("graded") {
		t.Fatal("status graded seharusnya dapat melihat hasil")
	}

	if canStudentSeeQuizResult("submitted") {
		t.Fatal("status submitted belum boleh melihat hasil sebelum guru selesai menilai")
	}

	if canStudentSeeQuizResult("in_progress") {
		t.Fatal("status in_progress tidak boleh melihat hasil")
	}
}

func TestIsObjectiveAnswerCorrectForMultipleChoice(t *testing.T) {
	correctOptionID := uuid.New()
	otherOptionID := uuid.New()
	question := models.QuizQuestion{
		Type: "multiple_choice",
		Options: []models.QuizQuestionOption{
			{ID: correctOptionID, IsCorrect: true},
			{ID: otherOptionID, IsCorrect: false},
		},
	}

	correctAnswer := &models.QuizAttemptAnswer{
		Options: []models.QuizQuestionOption{{ID: correctOptionID}},
	}
	if !isObjectiveAnswerCorrect(question, correctAnswer) {
		t.Fatal("matching multiple-choice options should be correct")
	}

	wrongAnswer := &models.QuizAttemptAnswer{
		Options: []models.QuizQuestionOption{{ID: otherOptionID}},
	}
	if isObjectiveAnswerCorrect(question, wrongAnswer) {
		t.Fatal("non-matching multiple-choice options should be incorrect")
	}
}

func TestEssayWithFullScoreCountsAsCorrect(t *testing.T) {
	question := models.QuizQuestion{Type: "essay", Points: 20}
	fullScore := 20.0
	answer := &models.QuizAttemptAnswer{Score: &fullScore}

	if answer.Score == nil || *answer.Score < question.Points {
		t.Fatal("essay dengan nilai penuh seharusnya dianggap benar")
	}
}
