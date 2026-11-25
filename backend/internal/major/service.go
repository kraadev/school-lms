package major

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"school-lms/internal/models"
)

type Service struct{ DB *gorm.DB }

func NewService(db *gorm.DB) *Service { return &Service{DB: db} }

func (s *Service) GetAll() ([]models.Major, error) {
	var majors []models.Major
	err := s.DB.Order("name ASC").Find(&majors).Error
	return majors, err
}

func (s *Service) GetByID(id string) (*models.Major, error) {
	majorID, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("ID jurusan tidak valid")
	}
	var item models.Major
	err = s.DB.Preload("Classes").First(&item, "id = ?", majorID).Error
	return &item, err
}

func (s *Service) Create(code, name string, description *string) (*models.Major, error) {
	var existing models.Major
	if err := s.DB.Where("code = ?", code).First(&existing).Error; err == nil {
		return nil, errors.New("kode jurusan sudah digunakan")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	item := &models.Major{Code: code, Name: name, Description: description}
	return item, s.DB.Create(item).Error
}

func (s *Service) Update(id, code, name string, description *string) (*models.Major, error) {
	item, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}
	var existing models.Major
	if err := s.DB.Where("code = ? AND id <> ?", code, item.ID).First(&existing).Error; err == nil {
		return nil, errors.New("kode jurusan sudah digunakan")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	item.Code, item.Name, item.Description = code, name, description
	return item, s.DB.Save(item).Error
}

func (s *Service) Delete(id string) error {
	item, err := s.GetByID(id)
	if err != nil {
		return err
	}
	var count int64
	if err := s.DB.Model(&models.Class{}).Where("major_id = ?", item.ID).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("jurusan masih digunakan oleh kelas")
	}
	return s.DB.Delete(item).Error
}
