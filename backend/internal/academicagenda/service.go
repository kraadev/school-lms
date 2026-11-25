package academicagenda

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"school-lms/internal/models"
)

type Service struct{ DB *gorm.DB }

type SaveRequest struct {
	Title       string    `json:"title" binding:"required"`
	Description *string   `json:"description"`
	StartDate   time.Time `json:"start_date" binding:"required"`
	EndDate     time.Time `json:"end_date" binding:"required"`
	StartTime   *string   `json:"start_time"`
	EndTime     *string   `json:"end_time"`
	Category    string    `json:"category"`
	Audience    string    `json:"audience"`
	Location    *string   `json:"location"`
	Status      string    `json:"status"`
}

func NewService(db *gorm.DB) *Service { return &Service{DB: db} }

func (s *Service) List(yearID string) ([]models.AcademicAgenda, error) {
	id, err := uuid.Parse(yearID)
	if err != nil {
		return nil, errors.New("id tahun akademik tidak valid")
	}
	var items []models.AcademicAgenda
	err = s.DB.Preload("CreatedBy").Preload("CreatedBy.Teacher").Where("academic_year_id = ?", id).Order("start_date ASC, title ASC").Find(&items).Error
	return items, err
}

func (s *Service) validateYear(yearID string, req SaveRequest) (uuid.UUID, error) {
	id, err := uuid.Parse(yearID)
	if err != nil {
		return uuid.Nil, errors.New("id tahun akademik tidak valid")
	}
	if !req.EndDate.After(req.StartDate) && !req.EndDate.Equal(req.StartDate) {
		return uuid.Nil, errors.New("tanggal selesai agenda harus sama atau setelah tanggal mulai")
	}
	var year models.AcademicYear
	if err := s.DB.First(&year, "id = ?", id).Error; err != nil {
		return uuid.Nil, err
	}
	if year.IsArchived {
		return uuid.Nil, errors.New("tahun akademik sudah diarsipkan dan agendanya tidak dapat diubah")
	}
	if req.StartDate.Before(year.StartDate) || req.EndDate.After(year.EndDate) {
		return uuid.Nil, errors.New("rentang agenda harus berada di dalam rentang tahun akademik")
	}
	return id, nil
}

func (s *Service) Create(yearID string, req SaveRequest, createdByID *uuid.UUID) (*models.AcademicAgenda, error) {
	id, err := s.validateYear(yearID, req)
	if err != nil {
		return nil, err
	}
	item := &models.AcademicAgenda{AcademicYearID: id, Title: req.Title, Description: req.Description, StartDate: req.StartDate, EndDate: req.EndDate, StartTime: req.StartTime, EndTime: req.EndTime, Category: req.Category, Audience: req.Audience, Location: req.Location, Status: req.Status, CreatedByID: createdByID}
	if item.Category == "" {
		item.Category = "Kegiatan sekolah"
	}
	if item.Audience == "" {
		item.Audience = "Semua"
	}
	if item.Status == "" {
		item.Status = "Terjadwal"
	}
	return item, s.DB.Create(item).Error
}

func (s *Service) Update(yearID, agendaID string, req SaveRequest) (*models.AcademicAgenda, error) {
	yearUUID, err := s.validateYear(yearID, req)
	if err != nil {
		return nil, err
	}
	id, err := uuid.Parse(agendaID)
	if err != nil {
		return nil, errors.New("id agenda tidak valid")
	}
	var item models.AcademicAgenda
	if err := s.DB.Where("id = ? AND academic_year_id = ?", id, yearUUID).First(&item).Error; err != nil {
		return nil, err
	}
	item.Title, item.Description, item.StartDate, item.EndDate = req.Title, req.Description, req.StartDate, req.EndDate
	item.StartTime, item.EndTime = req.StartTime, req.EndTime
	item.Category, item.Audience, item.Location, item.Status = req.Category, req.Audience, req.Location, req.Status
	if item.Status == "" {
		item.Status = "Terjadwal"
	}
	return &item, s.DB.Save(&item).Error
}

func (s *Service) Delete(yearID, agendaID string) error {
	yearUUID, err := uuid.Parse(yearID)
	if err != nil {
		return errors.New("id tahun akademik tidak valid")
	}
	var year models.AcademicYear
	if err := s.DB.First(&year, "id = ?", yearUUID).Error; err != nil {
		return err
	}
	if year.IsArchived {
		return errors.New("tahun akademik sudah diarsipkan dan agendanya tidak dapat diubah")
	}
	id, err := uuid.Parse(agendaID)
	if err != nil {
		return errors.New("id agenda tidak valid")
	}
	return s.DB.Delete(&models.AcademicAgenda{}, "id = ? AND academic_year_id = ?", id, yearUUID).Error
}
