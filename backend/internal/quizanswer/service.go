package quizanswer

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"school-lms/internal/models"
)

type Service struct {
	DB *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{
		DB: db,
	}
}

type SaveAnswerRequest struct {
	OptionID    *uuid.UUID `json:"option_id"`
	OptionIDs   []string   `json:"option_ids"`
	EssayAnswer *string    `json:"essay_answer"`
}

func (s *Service) SaveAnswer(
	userID string,
	attemptID string,
	questionID string,
	req SaveAnswerRequest,
) (*models.QuizAttemptAnswer, error) {

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("user id tidak valid")
	}

	attemptUUID, err := uuid.Parse(attemptID)
	if err != nil {
		return nil, errors.New("attempt id tidak valid")
	}

	questionUUID, err := uuid.Parse(questionID)
	if err != nil {
		return nil, errors.New("question id tidak valid")
	}

	// Cari student
	var student models.Student

	err = s.DB.
		Where("user_id = ?", userUUID).
		First(&student).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("profil siswa tidak ditemukan")
		}

		return nil, err
	}

	// Cari attempt milik siswa
	var attempt models.QuizAttempt

	err = s.DB.
		Where("id = ?", attemptUUID).
		Where("student_id = ?", student.ID).
		First(&attempt).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("attempt tidak ditemukan")
		}

		return nil, err
	}

	// Attempt harus masih berjalan
	if attempt.Status != "in_progress" {
		return nil, errors.New("quiz sudah tidak dapat dikerjakan")
	}

	now := time.Now()

	// Cek apakah attempt masih dalam batas waktu
	// Hitung deadline dari quiz duration
	var quiz models.Quiz

	err = s.DB.
		Where("id = ?", attempt.QuizID).
		First(&quiz).Error

	if err != nil {
		return nil, err
	}

	deadline := attempt.StartedAt.Add(
		time.Duration(quiz.Duration) * time.Minute,
	)

	// Jika ada EndAt quiz, gunakan yang lebih cepat
	if quiz.EndAt != nil && quiz.EndAt.Before(deadline) {
		deadline = *quiz.EndAt
	}

	if now.After(deadline) {
		attempt.Status = "expired"

		if err := s.DB.Save(&attempt).Error; err != nil {
			return nil, err
		}

		return nil, errors.New("waktu quiz telah berakhir")
	}

	// Cari question
	var question models.QuizQuestion

	err = s.DB.
		Preload("Options").
		Where("id = ?", questionUUID).
		Where("quiz_id = ?", attempt.QuizID).
		First(&question).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("soal tidak ditemukan")
		}

		return nil, err
	}

	// Validasi jawaban berdasarkan tipe soal
	var options []models.QuizQuestionOption
	parsedOptionIDs := make([]uuid.UUID, 0, len(req.OptionIDs))
	for _, rawID := range req.OptionIDs {
		trimmed := strings.TrimSpace(rawID)
		if trimmed == "" {
			continue
		}
		parsedUUID, err := uuid.Parse(trimmed)
		if err != nil {
			return nil, errors.New("pilihan jawaban tidak valid")
		}
		parsedOptionIDs = append(parsedOptionIDs, parsedUUID)
	}

	switch question.Type {

	case "multiple_choice", "single_choice", "true_false":

		if len(parsedOptionIDs) > 0 {
			err = s.DB.
				Where("id IN ?", parsedOptionIDs).
				Where("question_id = ?", question.ID).
				Find(&options).Error

			if err != nil {
				return nil, err
			}

			if len(options) != len(parsedOptionIDs) {
				return nil, errors.New("satu atau lebih pilihan jawaban tidak valid")
			}
		} else if req.OptionID != nil {
			var option models.QuizQuestionOption
			err = s.DB.
				Where("id = ?", *req.OptionID).
				Where("question_id = ?", question.ID).
				First(&option).Error

			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil, errors.New("pilihan jawaban tidak valid")
				}
				return nil, err
			}
			options = append(options, option)
		} else {
			return nil, errors.New("pilihan jawaban wajib diisi")
		}

		req.EssayAnswer = nil

	case "essay":

		if req.OptionID != nil || len(parsedOptionIDs) > 0 {
			return nil, errors.New("soal essay tidak menggunakan pilihan jawaban")
		}

		if req.EssayAnswer == nil {
			return nil, errors.New("jawaban essay wajib diisi")
		}

		answer := strings.TrimSpace(*req.EssayAnswer)

		if answer == "" {
			return nil, errors.New("jawaban essay tidak boleh kosong")
		}

		req.EssayAnswer = &answer

	default:
		return nil, errors.New("tipe soal tidak valid")
	}

	// Cari jawaban yang sudah ada
	var answer models.QuizAttemptAnswer

	err = s.DB.
		Where("attempt_id = ?", attempt.ID).
		Where("question_id = ?", question.ID).
		First(&answer).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {

		// Belum pernah menjawab
		answer = models.QuizAttemptAnswer{
			AttemptID:   attempt.ID,
			QuestionID:  question.ID,
			OptionID:    req.OptionID,
			EssayAnswer: req.EssayAnswer,
			AnsweredAt:  &now,
		}

		err = s.DB.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(&answer).Error; err != nil {
				return err
			}
			if len(options) > 0 {
				return tx.Model(&answer).Association("Options").Replace(options)
			}
			return nil
		})

		if err != nil {
			return nil, err
		}

		return &answer, nil
	}

	if err != nil {
		return nil, err
	}

	// Sudah ada → update jawaban
	answer.OptionID = req.OptionID
	answer.EssayAnswer = req.EssayAnswer
	answer.AnsweredAt = &now

	err = s.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&answer).Error; err != nil {
			return err
		}
		return tx.Model(&answer).Association("Options").Replace(options)
	})

	if err != nil {
		return nil, err
	}

	return &answer, nil
}
