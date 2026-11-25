package subject

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"school-lms/internal/models"
)

type Service struct {
	DB *gorm.DB
}

type DeleteConflictError struct {
	SubjectID  uuid.UUID
	ClassCount int64
}

func (e *DeleteConflictError) Error() string {
	return "mata pelajaran masih digunakan oleh kelas"
}

func (s *Service) GetByID(id string) (*models.Subject, error) {
	subjectID, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("ID mata pelajaran tidak valid")
	}
	var subject models.Subject
	err = s.DB.Preload("Classes").Preload("Teachers.User.Role").Preload("ClassSubjects.Class").Preload("ClassSubjects.Teacher").First(&subject, "id = ?", subjectID).Error
	return &subject, err
}

func (s *Service) Update(id, code, name string, description *string, teacherIDs []string) (*models.Subject, error) {
	subject, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}
	var duplicate models.Subject
	if err := s.DB.Where("code = ? AND id <> ?", code, subject.ID).First(&duplicate).Error; err == nil {
		return nil, errors.New("kode mata pelajaran sudah digunakan")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	subject.Code, subject.Name, subject.Description = code, name, description
	teachers, err := s.parseTeachers(teacherIDs)
	if err != nil {
		return nil, err
	}
	if err := s.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(subject).Error; err != nil {
			return err
		}
		return tx.Model(subject).Association("Teachers").Replace(teachers)
	}); err != nil {
		return nil, err
	}
	return s.GetByID(id)
}

func (s *Service) Delete(id string) error {
	subject, err := s.GetByID(id)
	if err != nil {
		return err
	}
	var used int64
	if err := s.DB.Table("class_subjects").Where("subject_id = ?", subject.ID).Count(&used).Error; err != nil {
		return err
	}
	if used > 0 {
		return &DeleteConflictError{SubjectID: subject.ID, ClassCount: used}
	}
	return s.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("subject_teachers").Where("subject_id = ?", subject.ID).Delete(nil).Error; err != nil {
			return err
		}
		return tx.Delete(subject).Error
	})
}

func NewService(db *gorm.DB) *Service {
	return &Service{
		DB: db,
	}
}

func (s *Service) GetAll() ([]models.Subject, error) {
	var subjects []models.Subject

	err := s.DB.
		Preload("Teachers.User.Role").
		Order("name ASC").
		Find(&subjects).Error

	return subjects, err
}

func (s *Service) Search(query string) ([]models.Subject, error) {
	var subjects []models.Subject
	term := "%" + strings.TrimSpace(query) + "%"
	err := s.DB.Preload("Teachers.User.Role").Where("code ILIKE ? OR name ILIKE ?", term, term).Order("name ASC").Limit(8).Find(&subjects).Error
	return subjects, err
}

func (s *Service) Create(
	code string,
	name string,
	description *string,
	teacherIDs []string,
) (*models.Subject, error) {

	var existing models.Subject

	err := s.DB.
		Where("code = ?", code).
		First(&existing).Error

	if err == nil {
		return nil, errors.New("kode mata pelajaran sudah digunakan")
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	subject := models.Subject{
		Code:        code,
		Name:        name,
		Description: description,
	}
	teachers, err := s.parseTeachers(teacherIDs)
	if err != nil {
		return nil, err
	}
	if err := s.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&subject).Error; err != nil {
			return err
		}
		return tx.Model(&subject).Association("Teachers").Replace(teachers)
	}); err != nil {
		return nil, err
	}

	return &subject, nil
}

func (s *Service) parseTeachers(ids []string) ([]models.Teacher, error) {
	if len(ids) == 0 {
		return []models.Teacher{}, nil
	}
	uuidIDs := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		parsed, err := uuid.Parse(id)
		if err != nil {
			return nil, errors.New("ID guru tidak valid")
		}
		uuidIDs = append(uuidIDs, parsed)
	}
	var teachers []models.Teacher
	if err := s.DB.Preload("User").Where("id IN ?", uuidIDs).Find(&teachers).Error; err != nil {
		return nil, err
	}
	if len(teachers) != len(uuidIDs) {
		return nil, errors.New("salah satu guru tidak ditemukan")
	}
	return teachers, nil
}
