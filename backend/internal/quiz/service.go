package quiz

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"school-lms/internal/models"
	"school-lms/internal/notification"
)

type Service struct {
	DB                  *gorm.DB
	NotificationService *notification.Service
}

func NewService(db *gorm.DB) *Service {
	return &Service{
		DB: db,
	}
}

func (s *Service) validateTeacherClassSubject(
	teacherID uuid.UUID,
	classSubjectID uuid.UUID,
) (*models.ClassSubject, error) {

	var classSubject models.ClassSubject

	err := s.DB.
		Where("id = ?", classSubjectID).
		Where("teacher_id = ?", teacherID).
		First(&classSubject).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(
				"kelas mata pelajaran tidak ditemukan atau bukan milik guru",
			)
		}

		return nil, err
	}

	return &classSubject, nil
}

func (s *Service) CreateQuiz(
	userID string,
	classSubjectID uuid.UUID,
	title string,
	description *string,
	duration int,
	startAt *time.Time,
	endAt *time.Time,
	maxAttempts int,
) (*models.Quiz, error) {

	teacherUserID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("user id tidak valid")
	}

	if title == "" {
		return nil, errors.New("judul quiz wajib diisi")
	}

	if duration <= 0 {
		return nil, errors.New("durasi quiz harus lebih dari 0")
	}

	if maxAttempts <= 0 {
		return nil, errors.New("maksimal percobaan harus lebih dari 0")
	}

	if startAt != nil && endAt != nil {
		if !endAt.After(*startAt) {
			return nil, errors.New(
				"waktu selesai harus setelah waktu mulai",
			)
		}
	}

	var teacher models.Teacher

	err = s.DB.
		Where("user_id = ?", teacherUserID).
		First(&teacher).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("profile guru tidak ditemukan")
		}

		return nil, err
	}

	_, err = s.validateTeacherClassSubject(
		teacher.ID,
		classSubjectID,
	)

	if err != nil {
		return nil, err
	}

	quiz := models.Quiz{
		ClassSubjectID: classSubjectID,
		TeacherID:      teacher.ID,

		Title:       title,
		Description: description,

		Duration: duration,

		StartAt: startAt,
		EndAt:   endAt,

		MaxAttempts: maxAttempts,

		Status: "draft",
	}

	if err := s.DB.Create(&quiz).Error; err != nil {
		return nil, err
	}

	return &quiz, nil
}

func (s *Service) GetTeacherQuizzes(
	userID string,
) ([]models.Quiz, error) {

	teacherUserID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("user id tidak valid")
	}

	var teacher models.Teacher

	err = s.DB.
		Where("user_id = ?", teacherUserID).
		First(&teacher).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("profile guru tidak ditemukan")
		}

		return nil, err
	}

	var quizzes []models.Quiz

	err = s.DB.
		Where("teacher_id = ?", teacher.ID).
		Preload("ClassSubject.Class").
		Preload("ClassSubject.Subject").
		Order("created_at DESC").
		Find(&quizzes).Error

	if err != nil {
		return nil, err
	}

	return quizzes, nil
}

func (s *Service) GetTeacherQuiz(
	userID string,
	quizID string,
) (*models.Quiz, error) {

	teacherUserID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("user id tidak valid")
	}

	quizUUID, err := uuid.Parse(quizID)
	if err != nil {
		return nil, errors.New("quiz id tidak valid")
	}

	var teacher models.Teacher

	err = s.DB.
		Where("user_id = ?", teacherUserID).
		First(&teacher).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("profile guru tidak ditemukan")
		}

		return nil, err
	}

	var quiz models.Quiz

	err = s.DB.
		Where("id = ?", quizUUID).
		Where("teacher_id = ?", teacher.ID).
		Preload("ClassSubject.Class").
		Preload("ClassSubject.Subject").
		First(&quiz).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(
				"quiz tidak ditemukan atau bukan milik guru",
			)
		}

		return nil, err
	}

	return &quiz, nil
}

func (s *Service) UpdateQuiz(
	userID string,
	quizID string,
	title string,
	description *string,
	duration int,
	startAt *time.Time,
	endAt *time.Time,
	maxAttempts int,
) (*models.Quiz, error) {

	quiz, err := s.GetTeacherQuiz(
		userID,
		quizID,
	)

	if err != nil {
		return nil, err
	}

	if quiz.Status != "draft" {
		return nil, errors.New(
			"quiz hanya dapat diubah ketika masih draft",
		)
	}

	if duration <= 0 {
		return nil, errors.New(
			"durasi quiz harus lebih dari 0",
		)
	}

	if maxAttempts <= 0 {
		return nil, errors.New(
			"maksimal percobaan harus lebih dari 0",
		)
	}

	if startAt != nil && endAt != nil {
		if !endAt.After(*startAt) {
			return nil, errors.New(
				"waktu selesai harus setelah waktu mulai",
			)
		}
	}

	quiz.Title = title
	quiz.Description = description
	quiz.Duration = duration
	quiz.StartAt = startAt
	quiz.EndAt = endAt
	quiz.MaxAttempts = maxAttempts

	if err := s.DB.Save(quiz).Error; err != nil {
		return nil, err
	}

	return quiz, nil
}

func canDeleteQuiz(status string) bool {
	switch status {
	case "draft", "published", "closed":
		return true
	default:
		return false
	}
}

func (s *Service) DeleteQuiz(
	userID string,
	quizID string,
) error {

	quiz, err := s.GetTeacherQuiz(
		userID,
		quizID,
	)

	if err != nil {
		return err
	}

	if !canDeleteQuiz(quiz.Status) {
		return errors.New(
			"quiz hanya dapat dihapus dari status draft, published, atau closed",
		)
	}

	return s.DB.Transaction(func(tx *gorm.DB) error {
		var questionIDs []uuid.UUID
		tx.Model(&models.QuizQuestion{}).Where("quiz_id = ?", quiz.ID).Pluck("id", &questionIDs)
		if len(questionIDs) > 0 {
			var optionIDs []uuid.UUID
			if err := tx.Model(&models.QuizQuestionOption{}).Where("question_id IN ?", questionIDs).Pluck("id", &optionIDs).Error; err != nil {
				return err
			}
			if len(optionIDs) > 0 {
				if err := tx.Exec("DELETE FROM quiz_attempt_answer_options WHERE quiz_question_option_id IN ?", optionIDs).Error; err != nil {
					return err
				}
			}
			if err := tx.Where("question_id IN ?", questionIDs).Delete(&models.QuizQuestionOption{}).Error; err != nil {
				return err
			}
			if err := tx.Where("question_id IN ?", questionIDs).Delete(&models.QuizAttemptAnswer{}).Error; err != nil {
				return err
			}
			if err := tx.Where("quiz_id = ?", quiz.ID).Delete(&models.QuizQuestion{}).Error; err != nil {
				return err
			}
		}

		if err := tx.Where("quiz_id = ?", quiz.ID).Delete(&models.QuizAttempt{}).Error; err != nil {
			return err
		}

		return tx.Delete(quiz).Error
	})
}

func isValidQuestionType(questionType string) bool {
	switch questionType {
	case "multiple_choice", "single_choice", "true_false", "essay":
		return true
	default:
		return false
	}
}

func (s *Service) CreateQuestion(
	userID string,
	quizID string,
	question string,
	questionType string,
	points float64,
	order int,
) (*models.QuizQuestion, error) {

	quiz, err := s.GetTeacherQuiz(
		userID,
		quizID,
	)

	if err != nil {
		return nil, err
	}

	if quiz.Status != "draft" {
		return nil, errors.New(
			"question hanya dapat ditambahkan ketika quiz masih draft",
		)
	}

	if question == "" {
		return nil, errors.New(
			"pertanyaan wajib diisi",
		)
	}

	if !isValidQuestionType(questionType) {
		return nil, errors.New(
			"tipe question tidak valid",
		)
	}

	if points <= 0 {
		return nil, errors.New(
			"poin harus lebih dari 0",
		)
	}

	if order <= 0 {
		var maxOrder int
		s.DB.Model(&models.QuizQuestion{}).
			Where("quiz_id = ?", quiz.ID).
			Select("COALESCE(MAX(\"order\"), 0)").
			Scan(&maxOrder)
		order = maxOrder + 1
	} else {
		var existingQuestion models.QuizQuestion
		err = s.DB.
			Where("quiz_id = ?", quiz.ID).
			Where(`"order" = ?`, order).
			First(&existingQuestion).Error

		if err == nil {
			var maxOrder int
			s.DB.Model(&models.QuizQuestion{}).
				Where("quiz_id = ?", quiz.ID).
				Select("COALESCE(MAX(\"order\"), 0)").
				Scan(&maxOrder)
			order = maxOrder + 1
		}
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	newQuestion := models.QuizQuestion{
		QuizID:   quiz.ID,
		Question: question,
		Type:     questionType,
		Points:   points,
		Order:    order,
	}

	if err := s.DB.Create(&newQuestion).Error; err != nil {
		return nil, err
	}

	return &newQuestion, nil
}

func (s *Service) GetTeacherQuizQuestions(
	userID string,
	quizID string,
) ([]models.QuizQuestion, error) {

	quiz, err := s.GetTeacherQuiz(
		userID,
		quizID,
	)

	if err != nil {
		return nil, err
	}

	var questions []models.QuizQuestion

	err = s.DB.
		Where("quiz_id = ?", quiz.ID).
		Order(`"order" ASC`).
		Find(&questions).Error

	if err != nil {
		return nil, err
	}

	return questions, nil
}

func (s *Service) GetTeacherQuestion(
	userID string,
	questionID string,
) (*models.QuizQuestion, error) {

	teacherUserID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("user id tidak valid")
	}

	questionUUID, err := uuid.Parse(questionID)
	if err != nil {
		return nil, errors.New("question id tidak valid")
	}

	var teacher models.Teacher

	err = s.DB.
		Where("user_id = ?", teacherUserID).
		First(&teacher).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("profile guru tidak ditemukan")
		}

		return nil, err
	}

	var question models.QuizQuestion

	err = s.DB.
		Joins("JOIN quizzes ON quizzes.id = quiz_questions.quiz_id").
		Where("quiz_questions.id = ?", questionUUID).
		Where("quizzes.teacher_id = ?", teacher.ID).
		Preload("Options").
		First(&question).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(
				"question tidak ditemukan atau bukan milik guru",
			)
		}

		return nil, err
	}

	return &question, nil
}

func (s *Service) UpdateQuestion(
	userID string,
	questionID string,
	questionText string,
	questionType string,
	points float64,
	order int,
) (*models.QuizQuestion, error) {

	question, err := s.GetTeacherQuestion(
		userID,
		questionID,
	)

	if err != nil {
		return nil, err
	}

	var quiz models.Quiz

	err = s.DB.
		Where("id = ?", question.QuizID).
		First(&quiz).Error

	if err != nil {
		return nil, err
	}

	if quiz.Status != "draft" {
		return nil, errors.New(
			"question hanya dapat diubah ketika quiz masih draft",
		)
	}

	if questionText == "" {
		return nil, errors.New(
			"pertanyaan wajib diisi",
		)
	}

	if !isValidQuestionType(questionType) {
		return nil, errors.New(
			"tipe question tidak valid",
		)
	}

	if points <= 0 {
		return nil, errors.New(
			"poin harus lebih dari 0",
		)
	}

	if order <= 0 {
		return nil, errors.New(
			"urutan question harus lebih dari 0",
		)
	}

	var existingQuestion models.QuizQuestion

	err = s.DB.
		Where("quiz_id = ?", question.QuizID).
		Where(`"order" = ?`, order).
		Where("id != ?", question.ID).
		First(&existingQuestion).Error

	if err == nil {
		return nil, errors.New(
			"urutan question sudah digunakan",
		)
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	question.Question = questionText
	question.Type = questionType
	question.Points = points
	question.Order = order

	if err := s.DB.Save(question).Error; err != nil {
		return nil, err
	}

	return question, nil
}

func (s *Service) DeleteQuestion(
	userID string,
	questionID string,
) error {

	question, err := s.GetTeacherQuestion(
		userID,
		questionID,
	)

	if err != nil {
		return err
	}

	var quiz models.Quiz

	if err := s.DB.
		Where("id = ?", question.QuizID).
		First(&quiz).Error; err != nil {

		return err
	}

	if quiz.Status != "draft" {
		return errors.New(
			"question hanya dapat dihapus ketika quiz masih draft",
		)
	}

	return s.DB.Transaction(func(tx *gorm.DB) error {
		// Hapus pilihan jawaban (options) terlebih dahulu
		if err := tx.Where("question_id = ?", question.ID).Delete(&models.QuizQuestionOption{}).Error; err != nil {
			return err
		}

		// Hapus jawaban attempt jika ada
		if err := tx.Where("question_id = ?", question.ID).Delete(&models.QuizAttemptAnswer{}).Error; err != nil {
			return err
		}

		return tx.Delete(question).Error
	})
}

func (s *Service) validateQuestionCanManageOptions(
	userID string,
	questionID string,
) (*models.QuizQuestion, *models.Quiz, error) {

	question, err := s.GetTeacherQuestion(
		userID,
		questionID,
	)

	if err != nil {
		return nil, nil, err
	}

	var quiz models.Quiz

	if err := s.DB.
		Where("id = ?", question.QuizID).
		First(&quiz).Error; err != nil {

		return nil, nil, err
	}

	if quiz.Status != "draft" {
		return nil, nil, errors.New(
			"option hanya dapat dikelola ketika quiz masih draft",
		)
	}

	if question.Type == "essay" {
		return nil, nil, errors.New(
			"question essay tidak memiliki pilihan jawaban",
		)
	}

	return question, &quiz, nil
}

func (s *Service) CreateQuestionOption(
	userID string,
	questionID string,
	text string,
	isCorrect bool,
	order int,
) (*models.QuizQuestionOption, error) {

	question, _, err := s.validateQuestionCanManageOptions(
		userID,
		questionID,
	)

	if err != nil {
		return nil, err
	}

	if text == "" {
		return nil, errors.New(
			"text option wajib diisi",
		)
	}

	if order <= 0 {
		var maxOrder int
		s.DB.Model(&models.QuizQuestionOption{}).
			Where("question_id = ?", question.ID).
			Select("COALESCE(MAX(\"order\"), 0)").
			Scan(&maxOrder)
		order = maxOrder + 1
	} else {
		var existingOption models.QuizQuestionOption
		err = s.DB.
			Where("question_id = ?", question.ID).
			Where(`"order" = ?`, order).
			First(&existingOption).Error

		if err == nil {
			var maxOrder int
			s.DB.Model(&models.QuizQuestionOption{}).
				Where("question_id = ?", question.ID).
				Select("COALESCE(MAX(\"order\"), 0)").
				Scan(&maxOrder)
			order = maxOrder + 1
		}
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Khusus true_false maksimal hanya 2 option.
	if question.Type == "true_false" {

		var optionCount int64

		if err := s.DB.
			Model(&models.QuizQuestionOption{}).
			Where("question_id = ?", question.ID).
			Count(&optionCount).Error; err != nil {

			return nil, err
		}

		if optionCount >= 2 {
			return nil, errors.New(
				"question true_false hanya dapat memiliki 2 option",
			)
		}
	}

	if isCorrect && question.Type != "multiple_choice" {
		var correctOptionCount int64

		if err := s.DB.
			Model(&models.QuizQuestionOption{}).
			Where("question_id = ?", question.ID).
			Where("is_correct = ?", true).
			Count(&correctOptionCount).Error; err != nil {

			return nil, err
		}

		if correctOptionCount > 0 {
			return nil, errors.New(
				"question hanya dapat memiliki satu jawaban benar",
			)
		}
	}

	option := models.QuizQuestionOption{
		QuestionID: question.ID,
		Text:       text,
		IsCorrect:  isCorrect,
		Order:      order,
	}

	if err := s.DB.Create(&option).Error; err != nil {
		return nil, err
	}

	return &option, nil
}

func (s *Service) GetTeacherQuestionOptions(
	userID string,
	questionID string,
) ([]models.QuizQuestionOption, error) {

	question, err := s.GetTeacherQuestion(
		userID,
		questionID,
	)

	if err != nil {
		return nil, err
	}

	var options []models.QuizQuestionOption

	err = s.DB.
		Where("question_id = ?", question.ID).
		Order(`"order" ASC`).
		Find(&options).Error

	if err != nil {
		return nil, err
	}

	return options, nil
}

// Menambahkan function GetTeacherQuestionOption (singular) untuk digunakan oleh Update/Delete.
func (s *Service) GetTeacherQuestionOption(
	userID string,
	optionID string,
) (*models.QuizQuestionOption, error) {

	teacherUserID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("user id tidak valid")
	}

	optionUUID, err := uuid.Parse(optionID)
	if err != nil {
		return nil, errors.New("option id tidak valid")
	}

	var teacher models.Teacher

	err = s.DB.
		Where("user_id = ?", teacherUserID).
		First(&teacher).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("profile guru tidak ditemukan")
		}

		return nil, err
	}

	var option models.QuizQuestionOption

	err = s.DB.
		Joins("JOIN quiz_questions ON quiz_questions.id = quiz_question_options.question_id").
		Joins("JOIN quizzes ON quizzes.id = quiz_questions.quiz_id").
		Where("quiz_question_options.id = ?", optionUUID).
		Where("quizzes.teacher_id = ?", teacher.ID).
		First(&option).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(
				"option tidak ditemukan atau bukan milik guru",
			)
		}

		return nil, err
	}

	return &option, nil
}

func (s *Service) UpdateQuestionOption(
	userID string,
	optionID string,
	text string,
	isCorrect bool,
	order int,
) (*models.QuizQuestionOption, error) {

	option, err := s.GetTeacherQuestionOption(
		userID,
		optionID,
	)

	if err != nil {
		return nil, err
	}

	if text == "" {
		return nil, errors.New(
			"text option wajib diisi",
		)
	}

	if order <= 0 {
		return nil, errors.New(
			"urutan option harus lebih dari 0",
		)
	}

	var existingOption models.QuizQuestionOption

	err = s.DB.
		Where("question_id = ?", option.QuestionID).
		Where(`"order" = ?`, order).
		Where("id != ?", option.ID).
		First(&existingOption).Error

	if err == nil {
		return nil, errors.New(
			"urutan option sudah digunakan",
		)
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Jika option akan dijadikan jawaban benar,
	// pastikan tidak ada option lain yang benar (kecuali multiple_choice).
	if isCorrect {
		var question models.QuizQuestion
		if err := s.DB.Where("id = ?", option.QuestionID).First(&question).Error; err == nil {
			if question.Type != "multiple_choice" {
				var correctOptionCount int64

				err = s.DB.
					Model(&models.QuizQuestionOption{}).
					Where("question_id = ?", option.QuestionID).
					Where("is_correct = ?", true).
					Where("id != ?", option.ID).
					Count(&correctOptionCount).Error

				if err != nil {
					return nil, err
				}

				if correctOptionCount > 0 {
					return nil, errors.New(
						"question hanya dapat memiliki satu jawaban benar",
					)
				}
			}
		}
	}

	option.Text = text
	option.IsCorrect = isCorrect
	option.Order = order

	if err := s.DB.Save(option).Error; err != nil {
		return nil, err
	}

	return option, nil
}

func (s *Service) DeleteQuestionOption(
	userID string,
	optionID string,
) error {

	option, err := s.GetTeacherQuestionOption(
		userID,
		optionID,
	)

	if err != nil {
		return err
	}

	return s.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.QuizAttemptAnswer{}).Where("option_id = ?", option.ID).Update("option_id", nil).Error; err != nil {
			return err
		}
		return tx.Delete(option).Error
	})
}

// -----------------------------------------------------------------------------
// PENAMBAHAN KODE BARU
// -----------------------------------------------------------------------------

func (s *Service) ValidateQuizForPublish(
	quiz *models.Quiz,
) error {

	var questions []models.QuizQuestion

	if err := s.DB.
		Where("quiz_id = ?", quiz.ID).
		Preload("Options").
		Order(`"order" ASC`).
		Find(&questions).Error; err != nil {

		return err
	}

	// Quiz harus memiliki minimal 1 question.
	if len(questions) == 0 {
		return errors.New(
			"quiz harus memiliki minimal satu question",
		)
	}

	for _, question := range questions {

		switch question.Type {

		case "single_choice":

			if len(question.Options) < 2 {
				return errors.New(
					"pilihan ganda harus memiliki minimal 2 opsi",
				)
			}

			correctCount := 0

			for _, option := range question.Options {
				if option.IsCorrect {
					correctCount++
				}
			}

			if correctCount != 1 {
				return errors.New(
					"pilihan ganda harus memiliki tepat satu jawaban benar",
				)
			}

		case "multiple_choice":

			if len(question.Options) < 2 {
				return errors.New(
					"jawaban jamak harus memiliki minimal 2 opsi",
				)
			}

			correctCount := 0

			for _, option := range question.Options {
				if option.IsCorrect {
					correctCount++
				}
			}

			if correctCount < 1 {
				return errors.New(
					"jawaban jamak harus memiliki minimal satu jawaban benar",
				)
			}

		case "true_false":

			if len(question.Options) != 2 {
				return errors.New(
					"true false harus memiliki tepat 2 option",
				)
			}

			correctCount := 0

			for _, option := range question.Options {
				if option.IsCorrect {
					correctCount++
				}
			}

			if correctCount != 1 {
				return errors.New(
					"true false harus memiliki tepat satu jawaban benar",
				)
			}

		case "essay":

			if len(question.Options) > 0 {
				return errors.New(
					"essay tidak boleh memiliki option",
				)
			}

		default:
			return errors.New(
				"tipe question tidak valid",
			)
		}
	}

	// Validasi waktu quiz.
	if quiz.StartAt != nil &&
		quiz.EndAt != nil &&
		!quiz.EndAt.After(*quiz.StartAt) {

		return errors.New(
			"waktu selesai harus setelah waktu mulai",
		)
	}

	return nil
}

func (s *Service) PublishQuiz(
	userID string,
	quizID string,
) (*models.Quiz, error) {

	quiz, err := s.GetTeacherQuiz(
		userID,
		quizID,
	)

	if err != nil {
		return nil, err
	}

	if quiz.Status != "draft" {
		return nil, errors.New(
			"quiz hanya dapat dipublish dari status draft",
		)
	}

	if err := s.ValidateQuizForPublish(quiz); err != nil {
		return nil, err
	}

	quiz.Status = "published"

	if err := s.DB.Save(quiz).Error; err != nil {
		return nil, err
	}

	if s.NotificationService != nil {
		refType := "quiz"
		_ = s.NotificationService.NotifyClass(quiz.ClassSubject.ClassID, "quiz", "Kuis Baru: "+quiz.Title, "Kuis baru telah dipublikasikan.", &refType, &quiz.ID)
	}

	return quiz, nil
}

func (s *Service) CloseQuiz(
	userID string,
	quizID string,
) (*models.Quiz, error) {

	quiz, err := s.GetTeacherQuiz(
		userID,
		quizID,
	)

	if err != nil {
		return nil, err
	}

	if quiz.Status != "published" {
		return nil, errors.New(
			"hanya quiz dengan status published yang dapat ditutup",
		)
	}

	quiz.Status = "closed"

	if err := s.DB.Save(quiz).Error; err != nil {
		return nil, err
	}

	return quiz, nil
}

func IsQuizAvailable(quiz *models.Quiz) bool {

	now := time.Now()

	if quiz.Status != "published" {
		return false
	}

	if quiz.StartAt != nil && now.Before(*quiz.StartAt) {
		return false
	}

	if quiz.EndAt != nil && now.After(*quiz.EndAt) {
		return false
	}

	return true
}

func (s *Service) GetStudentQuizzes(
	userID string,
) ([]models.Quiz, error) {

	studentUserID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("user id tidak valid")
	}

	var student models.Student

	err = s.DB.
		Where("user_id = ?", studentUserID).
		First(&student).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("profile siswa tidak ditemukan")
		}

		return nil, err
	}

	var quizzes []models.Quiz

	err = s.DB.
		Joins(`
			JOIN class_subjects
			ON class_subjects.id = quizzes.class_subject_id
		`).
		Joins(`
			JOIN class_students
			ON class_students.class_id = class_subjects.class_id
		`).
		Where(
			"class_students.student_id = ?",
			student.ID,
		).
		Where(
			"quizzes.status = ?",
			"published",
		).
		Preload("ClassSubject").
		Order("quizzes.start_at ASC").
		Find(&quizzes).Error

	if err != nil {
		return nil, err
	}

	return quizzes, nil
}

func (s *Service) CountStudentQuizAttempts(
	userID string,
	quizID uuid.UUID,
) (int64, error) {
	studentUserID, err := uuid.Parse(userID)
	if err != nil {
		return 0, errors.New("user id tidak valid")
	}

	var student models.Student
	if err := s.DB.Where("user_id = ?", studentUserID).First(&student).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, errors.New("profile siswa tidak ditemukan")
		}
		return 0, err
	}

	var attemptsCount int64
	if err := s.DB.Model(&models.QuizAttempt{}).
		Where("quiz_id = ? AND student_id = ?", quizID, student.ID).
		Count(&attemptsCount).Error; err != nil {
		return 0, err
	}

	return attemptsCount, nil
}

func (s *Service) GetStudentQuiz(
	userID string,
	quizID string,
) (*models.Quiz, error) {

	studentUserID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("user id tidak valid")
	}

	quizUUID, err := uuid.Parse(quizID)
	if err != nil {
		return nil, errors.New("quiz id tidak valid")
	}

	var student models.Student

	err = s.DB.
		Where("user_id = ?", studentUserID).
		First(&student).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("profile siswa tidak ditemukan")
		}

		return nil, err
	}

	var quiz models.Quiz

	err = s.DB.
		Joins(`
			JOIN class_subjects
			ON class_subjects.id = quizzes.class_subject_id
		`).
		Joins(`
			JOIN class_students
			ON class_students.class_id = class_subjects.class_id
		`).
		Where("quizzes.id = ?", quizUUID).
		Where("quizzes.status = ?", "published").
		Where(
			"class_students.student_id = ?",
			student.ID,
		).
		Preload("Questions", func(db *gorm.DB) *gorm.DB {
			return db.Order(`"order" ASC`)
		}).
		Preload("Questions.Options", func(db *gorm.DB) *gorm.DB {
			return db.Order(`"order" ASC`)
		}).
		First(&quiz).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(
				"quiz tidak ditemukan atau tidak dapat diakses",
			)
		}

		return nil, err
	}

	if !IsQuizAvailable(&quiz) {
		return nil, errors.New(
			"quiz sedang tidak tersedia",
		)
	}

	return &quiz, nil
}

type BulkOptionInput struct {
	Text      string `json:"text"`
	IsCorrect bool   `json:"is_correct"`
	Order     int    `json:"order"`
}

type BulkQuestionInput struct {
	Question string            `json:"question"`
	Type     string            `json:"type"`
	Points   float64           `json:"points"`
	Order    int               `json:"order"`
	Options  []BulkOptionInput `json:"options"`
}

func (s *Service) BulkCreateQuestions(
	userID string,
	quizID string,
	inputs []BulkQuestionInput,
) ([]models.QuizQuestion, error) {
	quiz, err := s.GetTeacherQuiz(userID, quizID)
	if err != nil {
		return nil, err
	}

	if quiz.Status != "draft" {
		return nil, errors.New("question hanya dapat ditambahkan ketika quiz masih draft")
	}

	var createdQuestions []models.QuizQuestion

	err = s.DB.Transaction(func(tx *gorm.DB) error {
		for index, q := range inputs {
			qOrder := q.Order
			if qOrder <= 0 {
				qOrder = index + 1
			}
			qType := q.Type
			if qType == "" {
				qType = "single_choice"
			}
			if !isValidQuestionType(qType) {
				return errors.New("tipe question tidak valid: " + qType)
			}
			qPoints := q.Points
			if qPoints <= 0 {
				qPoints = 10
			}

			newQ := models.QuizQuestion{
				QuizID:   quiz.ID,
				Question: q.Question,
				Type:     qType,
				Points:   qPoints,
				Order:    qOrder,
			}
			if err := tx.Create(&newQ).Error; err != nil {
				return err
			}

			if qType == "multiple_choice" || qType == "single_choice" || qType == "true_false" {
				for optIndex, opt := range q.Options {
					optOrder := opt.Order
					if optOrder <= 0 {
						optOrder = optIndex + 1
					}
					newOpt := models.QuizQuestionOption{
						QuestionID: newQ.ID,
						Text:       opt.Text,
						IsCorrect:  opt.IsCorrect,
						Order:      optOrder,
					}
					if err := tx.Create(&newOpt).Error; err != nil {
						return err
					}
					newQ.Options = append(newQ.Options, newOpt)
				}
			}
			createdQuestions = append(createdQuestions, newQ)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return createdQuestions, nil
}
