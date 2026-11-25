package quizattempt

import (
	"errors"
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

type AttemptResponse struct {
	ID            uuid.UUID  `json:"id"`
	QuizID        uuid.UUID  `json:"quiz_id"`
	AttemptNumber int        `json:"attempt_number"`
	Status        string     `json:"status"`
	StartedAt     time.Time  `json:"started_at"`
	DeadlineAt    time.Time  `json:"deadline_at"`
	SubmittedAt   *time.Time `json:"submitted_at"`
	Score         *float64   `json:"score"`
}

type TeacherAttemptResponse struct {
	ID            uuid.UUID  `json:"id"`
	StudentID     uuid.UUID  `json:"student_id"`
	StudentName   string     `json:"student_name"`
	AttemptNumber int        `json:"attempt_number"`
	Status        string     `json:"status"`
	StartedAt     time.Time  `json:"started_at"`
	DeadlineAt    time.Time  `json:"deadline_at"`
	SubmittedAt   *time.Time `json:"submitted_at"`
	Score         *float64   `json:"score"`
}

type QuizResultAnswerResponse struct {
	ID         uuid.UUID `json:"id"`
	QuestionID uuid.UUID `json:"question_id"`
	Question   string    `json:"question"`
	Type       string    `json:"type"`
	Points     float64   `json:"points"`

	OptionID    *uuid.UUID `json:"option_id"`
	OptionText  *string    `json:"option_text"`
	OptionTexts []string   `json:"option_texts,omitempty"`

	EssayAnswer *string  `json:"essay_answer"`
	Score       *float64 `json:"score"`
	Feedback    *string  `json:"feedback"`

	IsCorrect *bool `json:"is_correct"`
}

type QuizResultResponse struct {
	ID        uuid.UUID `json:"id"`
	QuizID    uuid.UUID `json:"quiz_id"`
	QuizTitle string    `json:"quiz_title"`

	StudentID     uuid.UUID `json:"student_id"`
	AttemptNumber int       `json:"attempt_number"`

	Status      string     `json:"status"`
	StartedAt   time.Time  `json:"started_at"`
	DeadlineAt  time.Time  `json:"deadline_at"`
	SubmittedAt *time.Time `json:"submitted_at"`

	Score *float64 `json:"score"`

	TotalQuestions int `json:"total_questions"`
	Answered       int `json:"answered"`
	Correct        int `json:"correct"`

	Answers []QuizResultAnswerResponse `json:"answers"`
}

// TeacherAnswerDetailResponse is the per-answer DTO returned to the teacher.
type TeacherAnswerDetailResponse struct {
	ID         uuid.UUID `json:"id"`
	QuestionID uuid.UUID `json:"question_id"`
	Question   string    `json:"question"`
	Type       string    `json:"type"`
	Points     float64   `json:"points"`

	OptionID    *uuid.UUID `json:"option_id"`
	OptionText  *string    `json:"option_text"`
	OptionTexts []string   `json:"option_texts,omitempty"`

	EssayAnswer *string  `json:"essay_answer"`
	Score       *float64 `json:"score"`
	Feedback    *string  `json:"feedback"`
	IsCorrect   *bool    `json:"is_correct"`
}

// TeacherAttemptDetailResponse is the shaped response for the teacher grading modal.
type TeacherAttemptDetailResponse struct {
	ID            uuid.UUID                     `json:"id"`
	StudentID     uuid.UUID                     `json:"student_id"`
	StudentName   string                        `json:"student_name"`
	AttemptNumber int                           `json:"attempt_number"`
	Status        string                        `json:"status"`
	StartedAt     time.Time                     `json:"started_at"`
	SubmittedAt   *time.Time                    `json:"submitted_at"`
	Score         *float64                      `json:"score"`
	Answers       []TeacherAnswerDetailResponse `json:"answers"`
}

type GradeEssayRequest struct {
	Score    *float64 `json:"score" binding:"required"`
	Feedback *string  `json:"feedback"`
}

func canStudentSeeQuizResult(status string) bool {
	return status == "graded"
}

func isObjectiveAnswerCorrect(
	question models.QuizQuestion,
	answer *models.QuizAttemptAnswer,
) bool {
	if answer == nil {
		return false
	}

	if question.Type == "single_choice" || question.Type == "true_false" {
		if answer.OptionID == nil {
			return false
		}
		for _, option := range question.Options {
			if option.ID == *answer.OptionID {
				return option.IsCorrect
			}
		}
		return false
	}

	if question.Type != "multiple_choice" {
		return false
	}

	correctIDs := make(map[uuid.UUID]struct{})
	for _, option := range question.Options {
		if option.IsCorrect {
			correctIDs[option.ID] = struct{}{}
		}
	}

	selectedIDs := make(map[uuid.UUID]struct{})
	for _, option := range answer.Options {
		selectedIDs[option.ID] = struct{}{}
	}

	if len(selectedIDs) != len(correctIDs) {
		return false
	}
	for id := range correctIDs {
		if _, selected := selectedIDs[id]; !selected {
			return false
		}
	}

	return true
}

// Helper function untuk konversi model ke response
func toAttemptResponse(
	attempt *models.QuizAttempt,
) *AttemptResponse {

	return &AttemptResponse{
		ID:            attempt.ID,
		QuizID:        attempt.QuizID,
		AttemptNumber: attempt.AttemptNumber,
		Status:        attempt.Status,
		StartedAt:     attempt.StartedAt,
		DeadlineAt:    attempt.DeadlineAt,
		SubmittedAt:   attempt.SubmittedAt,
		Score:         attempt.Score,
	}
}

func (s *Service) getTeacherByUserID(userID string) (*models.Teacher, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("user id tidak valid")
	}

	var teacher models.Teacher

	err = s.DB.
		Where("user_id = ?", id).
		First(&teacher).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("profil guru tidak ditemukan")
		}

		return nil, err
	}

	return &teacher, nil
}

func (s *Service) Start(
	userID string,
	quizID string,
) (*AttemptResponse, error) {

	studentUserID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("user id tidak valid")
	}

	quizUUID, err := uuid.Parse(quizID)
	if err != nil {
		return nil, errors.New("quiz id tidak valid")
	}

	// Cari student berdasarkan user_id
	var student models.Student

	err = s.DB.
		Where("user_id = ?", studentUserID).
		First(&student).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("profil siswa tidak ditemukan")
		}

		return nil, err
	}

	// Ambil quiz
	var quiz models.Quiz

	err = s.DB.
		Preload("Questions").
		First(&quiz, "id = ?", quizUUID).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("quiz tidak ditemukan")
		}

		return nil, err
	}

	// Quiz harus published
	if quiz.Status != "published" {
		return nil, errors.New("quiz belum tersedia")
	}

	now := time.Now()

	// Belum masuk waktu mulai
	if quiz.StartAt != nil && now.Before(*quiz.StartAt) {
		return nil, errors.New("quiz belum dimulai")
	}

	// Sudah melewati batas quiz
	if quiz.EndAt != nil && now.After(*quiz.EndAt) {
		return nil, errors.New("waktu quiz telah berakhir")
	}

	// Duration harus valid
	if quiz.Duration <= 0 {
		return nil, errors.New("durasi quiz tidak valid")
	}

	// Pastikan siswa memang berada di kelas yang memiliki quiz
	var count int64

	err = s.DB.
		Table("class_students cs").
		Joins("JOIN class_subjects csub ON csub.class_id = cs.class_id").
		Where("cs.student_id = ?", student.ID).
		Where("csub.id = ?", quiz.ClassSubjectID).
		Count(&count).Error

	if err != nil {
		return nil, err
	}

	if count == 0 {
		return nil, errors.New("quiz tidak tersedia untuk siswa ini")
	}

	// Cek apakah masih ada attempt yang sedang berjalan (in_progress)
	var activeAttempt models.QuizAttempt

	err = s.DB.
		Where("quiz_id = ?", quiz.ID).
		Where("student_id = ?", student.ID).
		Where("status = ?", "in_progress").
		Order("attempt_number DESC").
		First(&activeAttempt).Error

	if err == nil {
		// Ada attempt yang sedang berjalan, cek apakah masih dalam batas waktu
		deadline := activeAttempt.StartedAt.Add(
			time.Duration(quiz.Duration) * time.Minute,
		)

		// Jika ada EndAt quiz, gunakan yang lebih cepat
		if quiz.EndAt != nil && quiz.EndAt.Before(deadline) {
			deadline = *quiz.EndAt
		}

		// Attempt masih aktif (belum melewati deadline)
		if now.Before(deadline) {
			return toAttemptResponse(&activeAttempt), nil
		}

		// Kalau deadline sudah lewat, otomatis tandai expired/complete
		activeAttempt.Status = "expired"
		if err := s.DB.Save(&activeAttempt).Error; err != nil {
			return nil, err
		}
	}

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Hitung jumlah attempt yang sudah dilakukan (termasuk yang expired)
	var attemptsCount int64

	err = s.DB.
		Model(&models.QuizAttempt{}).
		Where("quiz_id = ?", quiz.ID).
		Where("student_id = ?", student.ID).
		Count(&attemptsCount).Error

	if err != nil {
		return nil, err
	}

	// Cek maksimal attempt
	if quiz.MaxAttempts > 0 &&
		int(attemptsCount) >= quiz.MaxAttempts {

		return nil, errors.New("jumlah percobaan quiz sudah mencapai batas")
	}

	// Attempt berikutnya
	attemptNumber := int(attemptsCount) + 1

	// Hitung deadline
	deadline := now.Add(
		time.Duration(quiz.Duration) * time.Minute,
	)

	if quiz.EndAt != nil && quiz.EndAt.Before(deadline) {
		deadline = *quiz.EndAt
	}

	attempt := models.QuizAttempt{
		QuizID:        quiz.ID,
		StudentID:     student.ID,
		AttemptNumber: attemptNumber,
		Status:        "in_progress",
		StartedAt:     now,
		DeadlineAt:    deadline,
		SubmittedAt:   nil,
		Score:         nil,
	}

	if err := s.DB.Create(&attempt).Error; err != nil {
		return nil, err
	}

	return toAttemptResponse(&attempt), nil
}

// Validasi deadline untuk SaveAnswer (akan dipanggil dari quizanswer package)
func (s *Service) ValidateAttemptForAnswer(attemptID string, studentID uuid.UUID) (*models.QuizAttempt, error) {
	var attempt models.QuizAttempt

	err := s.DB.
		Where("id = ? AND student_id = ?", attemptID, studentID).
		First(&attempt).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("attempt tidak ditemukan")
		}
		return nil, err
	}

	if attempt.Status != "in_progress" {
		return nil, errors.New("quiz sudah tidak dapat dikerjakan")
	}

	if !time.Now().Before(attempt.DeadlineAt) {
		attempt.Status = "expired"
		if err := s.DB.Save(&attempt).Error; err != nil {
			return nil, err
		}
		return nil, errors.New("waktu quiz telah habis")
	}

	return &attempt, nil
}

func (s *Service) Submit(
	userID string,
	attemptID string,
) (*AttemptResponse, error) {

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("user id tidak valid")
	}

	attemptUUID, err := uuid.Parse(attemptID)
	if err != nil {
		return nil, errors.New("attempt id tidak valid")
	}

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

	if attempt.Status == "expired" {
		return nil, errors.New("waktu quiz telah habis")
	}

	if attempt.Status == "graded" {
		return nil, errors.New("quiz sudah selesai")
	}

	if attempt.Status != "in_progress" {
		return nil, errors.New("quiz sudah tidak dapat dikumpulkan")
	}

	if !time.Now().Before(attempt.DeadlineAt) {
		attempt.Status = "expired"

		if err := s.DB.Save(&attempt).Error; err != nil {
			return nil, err
		}

		return nil, errors.New("waktu quiz telah habis")
	}

	now := time.Now()

	err = s.DB.Transaction(func(tx *gorm.DB) error {

		attempt.Status = "submitted"
		attempt.SubmittedAt = &now

		if err := tx.Save(&attempt).Error; err != nil {
			return err
		}

		return s.autoGradeWithDB(tx, &attempt)
	})

	if err != nil {
		return nil, err
	}

	return toAttemptResponse(&attempt), nil
}

func (s *Service) GetLatestStudentAttempt(
	userID string,
	quizID string,
) (*AttemptResponse, error) {
	studentUserID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("user id tidak valid")
	}

	quizUUID, err := uuid.Parse(quizID)
	if err != nil {
		return nil, errors.New("quiz id tidak valid")
	}

	var student models.Student
	if err := s.DB.Where("user_id = ?", studentUserID).First(&student).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("profil siswa tidak ditemukan")
		}
		return nil, err
	}

	var attempt models.QuizAttempt
	if err := s.DB.
		Where("quiz_id = ? AND student_id = ?", quizUUID, student.ID).
		Order("attempt_number DESC").
		First(&attempt).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("attempt tidak ditemukan")
		}
		return nil, err
	}

	return toAttemptResponse(&attempt), nil
}

// AutoGrade wrapper untuk penggunaan dengan DB default
func (s *Service) AutoGrade(
	attempt *models.QuizAttempt,
) error {
	return s.autoGradeWithDB(s.DB, attempt)
}

// autoGradeWithDB fungsi internal untuk grading dengan transaction support
func (s *Service) autoGradeWithDB(
	db *gorm.DB,
	attempt *models.QuizAttempt,
) error {

	var quiz models.Quiz

	err := db.
		Preload("Questions.Options").
		First(&quiz, "id = ?", attempt.QuizID).Error

	if err != nil {
		return err
	}

	var answers []models.QuizAttemptAnswer

	err = db.
		Where("attempt_id = ?", attempt.ID).
		Preload("Options").
		Find(&answers).Error

	if err != nil {
		return err
	}

	answerMap := make(
		map[uuid.UUID]*models.QuizAttemptAnswer,
	)

	for i := range answers {
		answerMap[answers[i].QuestionID] = &answers[i]
	}

	var totalScore float64
	hasEssay := false

	for _, question := range quiz.Questions {

		answer := answerMap[question.ID]

		switch question.Type {

		case "single_choice", "true_false":

			if answer == nil || answer.OptionID == nil {
				continue
			}

			for _, option := range question.Options {
				if option.ID == *answer.OptionID && option.IsCorrect {
					totalScore += question.Points
					break
				}
			}

		case "multiple_choice":

			if answer == nil || len(answer.Options) == 0 {
				continue
			}

			// Build set of correct option IDs for the question
			correctIDs := make(map[uuid.UUID]struct{})
			for _, opt := range question.Options {
				if opt.IsCorrect {
					correctIDs[opt.ID] = struct{}{}
				}
			}

			// Build set of selected option IDs
			selectedIDs := make(map[uuid.UUID]struct{})
			for _, opt := range answer.Options {
				selectedIDs[opt.ID] = struct{}{}
			}

			// Full-match scoring: all correct selected and none wrong selected
			if len(selectedIDs) == len(correctIDs) {
				allMatch := true
				for id := range correctIDs {
					if _, ok := selectedIDs[id]; !ok {
						allMatch = false
						break
					}
				}
				if allMatch {
					totalScore += question.Points
				}
			}

		case "essay":

			hasEssay = true
		}
	}

	attempt.Score = &totalScore

	if !hasEssay {
		attempt.Status = "graded"
	}

	return db.Save(attempt).Error
}

func (s *Service) GetTeacherAttempts(
	userID string,
	quizID string,
) ([]TeacherAttemptResponse, error) {

	teacher, err := s.getTeacherByUserID(userID)
	if err != nil {
		return nil, err
	}

	quizUUID, err := uuid.Parse(quizID)
	if err != nil {
		return nil, errors.New("quiz id tidak valid")
	}

	var quiz models.Quiz

	err = s.DB.
		Where("id = ? AND teacher_id = ?", quizUUID, teacher.ID).
		First(&quiz).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("quiz tidak ditemukan")
		}

		return nil, err
	}

	var attempts []models.QuizAttempt

	err = s.DB.
		Preload("Student").
		Where("quiz_id = ?", quiz.ID).
		Order("attempt_number ASC").
		Find(&attempts).Error

	if err != nil {
		return nil, err
	}

	result := make([]TeacherAttemptResponse, 0, len(attempts))

	for _, attempt := range attempts {
		result = append(result, TeacherAttemptResponse{
			ID:            attempt.ID,
			StudentID:     attempt.StudentID,
			StudentName:   attempt.Student.FullName,
			AttemptNumber: attempt.AttemptNumber,
			Status:        attempt.Status,
			StartedAt:     attempt.StartedAt,
			DeadlineAt:    attempt.DeadlineAt,
			SubmittedAt:   attempt.SubmittedAt,
			Score:         attempt.Score,
		})
	}

	return result, nil
}

func (s *Service) GetTeacherAttempt(
	userID string,
	attemptID string,
) (*TeacherAttemptDetailResponse, error) {

	teacher, err := s.getTeacherByUserID(userID)
	if err != nil {
		return nil, err
	}

	id, err := uuid.Parse(attemptID)
	if err != nil {
		return nil, errors.New("attempt id tidak valid")
	}

	var attempt models.QuizAttempt

	err = s.DB.
		Preload("Student").
		Preload("Quiz").
		Preload("Quiz.Questions").
		Preload("Quiz.Questions.Options").
		Preload("Answers").
		Preload("Answers.Question").
		Preload("Answers.Option").
		Preload("Answers.Options").
		Joins("JOIN quizzes ON quizzes.id = quiz_attempts.quiz_id").
		Where(
			"quiz_attempts.id = ? AND quizzes.teacher_id = ?",
			id,
			teacher.ID,
		).
		First(&attempt).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("attempt tidak ditemukan")
		}

		return nil, err
	}

	// Build answer map keyed by question ID for ordered output
	answerByQuestion := make(map[uuid.UUID]*models.QuizAttemptAnswer)
	for i := range attempt.Answers {
		answerByQuestion[attempt.Answers[i].QuestionID] = &attempt.Answers[i]
	}

	answerDTOs := make([]TeacherAnswerDetailResponse, 0, len(attempt.Quiz.Questions))

	for _, question := range attempt.Quiz.Questions {
		ans := answerByQuestion[question.ID]

		dto := TeacherAnswerDetailResponse{
			QuestionID: question.ID,
			Question:   question.Question,
			Type:       question.Type,
			Points:     question.Points,
		}

		if ans != nil {
			dto.ID = ans.ID
			dto.OptionID = ans.OptionID
			dto.EssayAnswer = ans.EssayAnswer
			dto.Score = ans.Score
			dto.Feedback = ans.Feedback

			// Single option text (legacy / single_choice)
			if ans.Option != nil {
				t := ans.Option.Text
				dto.OptionText = &t
			}

			// Multiple option texts (multiple_choice via join table)
			if len(ans.Options) > 0 {
				texts := make([]string, 0, len(ans.Options))
				for _, o := range ans.Options {
					texts = append(texts, o.Text)
				}
				dto.OptionTexts = texts
			}

			// Compute correctness for objective questions
			if question.Type == "single_choice" || question.Type == "true_false" {
				if ans.OptionID != nil {
					for _, opt := range question.Options {
						if opt.ID == *ans.OptionID {
							isCorrect := opt.IsCorrect
							dto.IsCorrect = &isCorrect
							break
						}
					}
				}
			} else if question.Type == "multiple_choice" && len(ans.Options) > 0 {
				correctIDs := make(map[uuid.UUID]struct{})
				for _, opt := range question.Options {
					if opt.IsCorrect {
						correctIDs[opt.ID] = struct{}{}
					}
				}
				selectedIDs := make(map[uuid.UUID]struct{})
				for _, opt := range ans.Options {
					selectedIDs[opt.ID] = struct{}{}
				}
				allMatch := len(selectedIDs) == len(correctIDs)
				if allMatch {
					for id := range correctIDs {
						if _, ok := selectedIDs[id]; !ok {
							allMatch = false
							break
						}
					}
				}
				dto.IsCorrect = &allMatch
			}
		}

		answerDTOs = append(answerDTOs, dto)
	}

	return &TeacherAttemptDetailResponse{
		ID:            attempt.ID,
		StudentID:     attempt.StudentID,
		StudentName:   attempt.Student.FullName,
		AttemptNumber: attempt.AttemptNumber,
		Status:        attempt.Status,
		StartedAt:     attempt.StartedAt,
		SubmittedAt:   attempt.SubmittedAt,
		Score:         attempt.Score,
		Answers:       answerDTOs,
	}, nil
}

func (s *Service) GradeEssay(
	userID string,
	attemptID string,
	answerID string,
	score float64,
	feedback *string,
) (*TeacherAttemptDetailResponse, error) {

	teacher, err := s.getTeacherByUserID(userID)
	if err != nil {
		return nil, err
	}

	attemptUUID, err := uuid.Parse(attemptID)
	if err != nil {
		return nil, errors.New("attempt id tidak valid")
	}

	answerUUID, err := uuid.Parse(answerID)
	if err != nil {
		return nil, errors.New("answer id tidak valid")
	}

	var attempt models.QuizAttempt

	err = s.DB.
		Preload("Quiz").
		Preload("Answers").
		Where(
			"quiz_attempts.id = ?",
			attemptUUID,
		).
		Joins("JOIN quizzes ON quizzes.id = quiz_attempts.quiz_id").
		Where("quizzes.teacher_id = ?", teacher.ID).
		First(&attempt).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("attempt tidak ditemukan")
		}

		return nil, err
	}

	if attempt.Status == "in_progress" {
		return nil, errors.New("quiz belum dikumpulkan oleh siswa")
	}

	if attempt.Status == "expired" {
		return nil, errors.New("attempt sudah expired")
	}

	var answer models.QuizAttemptAnswer

	err = s.DB.
		Preload("Question").
		Where(
			"(id = ? OR question_id = ?) AND attempt_id = ?",
			answerUUID,
			answerUUID,
			attempt.ID,
		).
		First(&answer).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Jika siswa tidak menjawab essay ini, cari question dan buat answer baru
			var question models.QuizQuestion
			if qErr := s.DB.Where("id = ? AND quiz_id = ?", answerUUID, attempt.QuizID).First(&question).Error; qErr == nil && question.Type == "essay" {
				now := time.Now()
				emptyAns := ""
				answer = models.QuizAttemptAnswer{
					AttemptID:   attempt.ID,
					QuestionID:  question.ID,
					EssayAnswer: &emptyAns,
					AnsweredAt:  &now,
					Question:    question,
				}
				if createErr := s.DB.Create(&answer).Error; createErr != nil {
					return nil, createErr
				}
			} else {
				return nil, errors.New("jawaban tidak ditemukan")
			}
		} else {
			return nil, err
		}
	}

	if answer.Question.Type != "essay" {
		return nil, errors.New("jawaban ini bukan soal essay")
	}

	if score < 0 {
		return nil, errors.New("nilai tidak boleh kurang dari 0")
	}

	if score > answer.Question.Points {
		return nil, errors.New("nilai melebihi poin soal")
	}

	err = s.DB.Transaction(func(tx *gorm.DB) error {

		answer.Score = &score
		answer.Feedback = feedback

		if err := tx.Save(&answer).Error; err != nil {
			return err
		}

		return s.recalculateAttemptScore(tx, &attempt)
	})

	if err != nil {
		return nil, err
	}

	return s.GetTeacherAttempt(userID, attemptID)
}

func (s *Service) recalculateAttemptScore(
	db *gorm.DB,
	attempt *models.QuizAttempt,
) error {

	var quiz models.Quiz

	err := db.
		Preload("Questions.Options").
		First(&quiz, "id = ?", attempt.QuizID).Error

	if err != nil {
		return err
	}

	var answers []models.QuizAttemptAnswer

	err = db.
		Where("attempt_id = ?", attempt.ID).
		Preload("Question").
		Preload("Options").
		Find(&answers).Error

	if err != nil {
		return err
	}

	answerMap := make(
		map[uuid.UUID]*models.QuizAttemptAnswer,
	)

	for i := range answers {
		answerMap[answers[i].QuestionID] = &answers[i]
	}

	var totalScore float64
	allEssayGraded := true
	hasEssay := false

	for _, question := range quiz.Questions {

		answer := answerMap[question.ID]

		switch question.Type {

		case "single_choice", "true_false":

			if answer == nil || answer.OptionID == nil {
				continue
			}

			for _, option := range question.Options {
				if option.ID == *answer.OptionID && option.IsCorrect {
					totalScore += question.Points
					break
				}
			}

		case "multiple_choice":

			if answer == nil || len(answer.Options) == 0 {
				continue
			}

			correctIDs := make(map[uuid.UUID]struct{})
			for _, opt := range question.Options {
				if opt.IsCorrect {
					correctIDs[opt.ID] = struct{}{}
				}
			}
			selectedIDs := make(map[uuid.UUID]struct{})
			for _, opt := range answer.Options {
				selectedIDs[opt.ID] = struct{}{}
			}
			if len(selectedIDs) == len(correctIDs) {
				allMatch := true
				for id := range correctIDs {
					if _, ok := selectedIDs[id]; !ok {
						allMatch = false
						break
					}
				}
				if allMatch {
					totalScore += question.Points
				}
			}

		case "essay":

			hasEssay = true

			// Essay tidak dijawab = nilai 0.
			if answer == nil {
				continue
			}

			// Essay sudah dijawab tetapi belum dinilai.
			if answer.Score == nil {
				allEssayGraded = false
				continue
			}

			totalScore += *answer.Score
		}
	}

	attempt.Score = &totalScore

	if !hasEssay || allEssayGraded {
		attempt.Status = "graded"
	} else {
		attempt.Status = "submitted"
	}

	return db.Save(attempt).Error
}

func (s *Service) GetStudentQuizResult(
	userID string,
	attemptID string,
) (*QuizResultResponse, error) {

	studentUserID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("user id tidak valid")
	}

	attemptUUID, err := uuid.Parse(attemptID)
	if err != nil {
		return nil, errors.New("attempt id tidak valid")
	}

	var student models.Student

	err = s.DB.
		Where("user_id = ?", studentUserID).
		First(&student).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("profil siswa tidak ditemukan")
		}

		return nil, err
	}

	var attempt models.QuizAttempt

	err = s.DB.
		Preload("Quiz").
		Preload("Quiz.Questions").
		Preload("Quiz.Questions.Options").
		Preload("Answers").
		Preload("Answers.Options").
		Preload("Answers.Question").
		Preload("Answers.Option").
		Where(
			"id = ? AND student_id = ?",
			attemptUUID,
			student.ID,
		).
		First(&attempt).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("hasil quiz tidak ditemukan")
		}

		return nil, err
	}

	// Hasil kuis hanya bisa dilihat setelah guru menyelesaikan penilaian,
	// termasuk essay yang masih menunggu nilai.
	if !canStudentSeeQuizResult(attempt.Status) {
		return nil, errors.New("hasil kuis belum dapat ditampilkan karena penilaian belum selesai")
	}

	answerMap := make(
		map[uuid.UUID]*models.QuizAttemptAnswer,
	)

	for i := range attempt.Answers {
		answerMap[attempt.Answers[i].QuestionID] = &attempt.Answers[i]
	}

	result := &QuizResultResponse{
		ID:             attempt.ID,
		QuizID:         attempt.QuizID,
		QuizTitle:      attempt.Quiz.Title,
		StudentID:      attempt.StudentID,
		AttemptNumber:  attempt.AttemptNumber,
		Status:         attempt.Status,
		StartedAt:      attempt.StartedAt,
		DeadlineAt:     attempt.DeadlineAt,
		SubmittedAt:    attempt.SubmittedAt,
		Score:          attempt.Score,
		TotalQuestions: len(attempt.Quiz.Questions),
		Answers:        make([]QuizResultAnswerResponse, 0),
	}

	for _, question := range attempt.Quiz.Questions {

		answer := answerMap[question.ID]

		answerResponse := QuizResultAnswerResponse{
			QuestionID: question.ID,
			Question:   question.Question,
			Type:       question.Type,
			Points:     question.Points,
		}

		if answer != nil {

			result.Answered++

			answerResponse.ID = answer.ID
			answerResponse.OptionID = answer.OptionID
			answerResponse.EssayAnswer = answer.EssayAnswer
			answerResponse.Score = answer.Score
			answerResponse.Feedback = answer.Feedback

			if len(answer.Options) > 0 {
				texts := make([]string, 0, len(answer.Options))
				for _, opt := range answer.Options {
					texts = append(texts, opt.Text)
				}
				answerResponse.OptionTexts = texts
			}

			if answer.Option != nil {
				optionText := answer.Option.Text
				answerResponse.OptionText = &optionText
			}

			// Objective correctness uses the same full-match rule as grading.
			if question.Type == "multiple_choice" ||
				question.Type == "single_choice" ||
				question.Type == "true_false" {
				isCorrect := isObjectiveAnswerCorrect(question, answer)
				answerResponse.IsCorrect = &isCorrect
				if isCorrect {
					result.Correct++
				}
			} else if question.Type == "essay" && answer.Score != nil {
				isCorrect := *answer.Score >= question.Points
				answerResponse.IsCorrect = &isCorrect
				if isCorrect {
					result.Correct++
				}
			}
		}

		result.Answers = append(
			result.Answers,
			answerResponse,
		)
	}

	return result, nil
}

// ExpireOverdueAttempts menandai attempt yang melewati deadline sebagai expired
func (s *Service) ExpireOverdueAttempts() error {

	now := time.Now()

	result := s.DB.
		Model(&models.QuizAttempt{}).
		Where(
			"status = ? AND deadline_at <= ?",
			"in_progress",
			now,
		).
		Updates(map[string]interface{}{
			"status": "expired",
		})

	return result.Error
}

// StartAutoExpireWorker menjalankan worker untuk mengecek expired attempt secara periodik
func (s *Service) StartAutoExpireWorker() {

	ticker := time.NewTicker(5 * time.Second)

	go func() {

		defer ticker.Stop()

		for range ticker.C {

			if err := s.ExpireOverdueAttempts(); err != nil {
				// Jangan matikan worker hanya karena satu query gagal.
				// Logging bisa ditambahkan nanti.
				continue
			}
		}
	}()
}
