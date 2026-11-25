package assignmentsubmissionfile

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"school-lms/internal/models"
)

type Service struct {
	DB *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{DB: db}
}

func (s *Service) GetTeacherFile(
	userID string,
	fileID string,
) (*models.AssignmentSubmissionFile, error) {

	teacherUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("user id tidak valid")
	}

	fileUUID, err := uuid.Parse(fileID)
	if err != nil {
		return nil, errors.New("file id tidak valid")
	}

	var teacher models.Teacher

	err = s.DB.
		Where("user_id = ?", teacherUUID).
		First(&teacher).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("profile guru tidak ditemukan")
		}

		return nil, err
	}

	var file models.AssignmentSubmissionFile

	err = s.DB.
		Joins(`
			JOIN assignment_submissions
			ON assignment_submissions.id =
			assignment_submission_files.submission_id
		`).
		Joins(`
			JOIN assignments
			ON assignments.id =
			assignment_submissions.assignment_id
		`).
		Where(
			"assignment_submission_files.id = ?",
			fileUUID,
		).
		Where(
			"assignments.teacher_id = ?",
			teacher.ID,
		).
		First(&file).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(
				"file tidak ditemukan atau bukan milik tugas guru",
			)
		}

		return nil, err
	}

	return &file, nil
}
