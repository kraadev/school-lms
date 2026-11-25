package academicyear

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

type SaveRequest struct {
	Name      string    `json:"name" binding:"required"`
	IsActive  bool      `json:"is_active"`
	StartDate time.Time `json:"start_date" binding:"required"`
	EndDate   time.Time `json:"end_date" binding:"required"`
}

type SemesterRequest struct {
	Number    int       `json:"number" binding:"required,min=1"`
	StartDate time.Time `json:"start_date" binding:"required"`
	EndDate   time.Time `json:"end_date" binding:"required"`
}

type DuplicateRequest struct {
	Name string `json:"name"`
}

func NewService(db *gorm.DB) *Service { return &Service{DB: db} }

func setSemesterStatuses(year *models.AcademicYear, now time.Time) {
	for index := range year.Semesters {
		year.Semesters[index].SetStatus(now)
	}
}

func (s *Service) GetAll() ([]models.AcademicYear, error) {
	var years []models.AcademicYear
	err := s.DB.Preload("Semesters", func(db *gorm.DB) *gorm.DB {
		return db.Order("number ASC")
	}).Order("name ASC").Find(&years).Error
	if err == nil {
		now := time.Now()
		for index := range years {
			setSemesterStatuses(&years[index], now)
		}
	}
	return years, err
}

func (s *Service) GetActive() (*models.AcademicYear, error) {
	var year models.AcademicYear
	err := s.DB.Preload("Semesters", func(db *gorm.DB) *gorm.DB {
		return db.Order("number ASC")
	}).Where("is_active = ?", true).First(&year).Error
	if err == nil {
		setSemesterStatuses(&year, time.Now())
	}
	return &year, err
}

func (s *Service) Create(req SaveRequest) (*models.AcademicYear, error) {
	if !req.EndDate.After(req.StartDate) {
		return nil, errors.New("tanggal selesai tahun akademik harus setelah tanggal mulai")
	}
	year := &models.AcademicYear{
		Name: req.Name, IsActive: req.IsActive, StartDate: req.StartDate, EndDate: req.EndDate,
	}
	return year, s.DB.Transaction(func(tx *gorm.DB) error {
		if req.IsActive {
			if err := tx.Model(&models.AcademicYear{}).Where("is_active = ?", true).Updates(map[string]interface{}{"is_active": false, "is_archived": true}).Error; err != nil {
				return errors.New("gagal memperbarui tahun akademik aktif")
			}
		}
		return tx.Create(year).Error
	})
}

func (s *Service) Update(id string, req SaveRequest) (*models.AcademicYear, error) {
	yearID, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("id tahun akademik tidak valid")
	}
	if !req.EndDate.After(req.StartDate) {
		return nil, errors.New("tanggal selesai tahun akademik harus setelah tanggal mulai")
	}
	var year models.AcademicYear
	if err := s.DB.First(&year, "id = ?", yearID).Error; err != nil {
		return nil, err
	}
	if year.IsArchived {
		return nil, errors.New("tahun akademik sudah diarsipkan dan tidak dapat diedit")
	}
	return &year, s.DB.Transaction(func(tx *gorm.DB) error {
		if req.IsActive {
			if err := tx.Model(&models.AcademicYear{}).Where("id <> ? AND is_active = ?", yearID, true).Updates(map[string]interface{}{"is_active": false, "is_archived": true}).Error; err != nil {
				return errors.New("gagal memperbarui tahun akademik aktif")
			}
		}
		year.Name, year.IsActive = req.Name, req.IsActive
		year.StartDate, year.EndDate = req.StartDate, req.EndDate
		return tx.Save(&year).Error
	})
}

func (s *Service) CreateSemester(yearID string, req SemesterRequest) (*models.Semester, error) {
	academicYearID, err := uuid.Parse(yearID)
	if err != nil {
		return nil, errors.New("id tahun akademik tidak valid")
	}
	if !req.EndDate.After(req.StartDate) {
		return nil, errors.New("tanggal selesai harus setelah tanggal mulai")
	}
	var year models.AcademicYear
	if err := s.DB.First(&year, "id = ?", academicYearID).Error; err != nil {
		return nil, err
	}
	if year.IsArchived {
		return nil, errors.New("tahun akademik sudah diarsipkan dan tidak dapat diubah")
	}
	if req.StartDate.Before(year.StartDate) || req.EndDate.After(year.EndDate) {
		return nil, errors.New("rentang semester harus berada di dalam rentang tahun akademik")
	}
	var overlap int64
	if err := s.DB.Model(&models.Semester{}).
		Where("academic_year_id = ? AND start_date <= ? AND end_date >= ?", academicYearID, req.EndDate, req.StartDate).
		Count(&overlap).Error; err != nil {
		return nil, err
	}
	if overlap > 0 {
		return nil, errors.New("rentang tanggal semester bertumpang tindih dengan semester lain")
	}
	var existing int64
	if err := s.DB.Model(&models.Semester{}).Where("academic_year_id = ? AND number = ?", academicYearID, req.Number).Count(&existing).Error; err != nil {
		return nil, err
	}
	if existing > 0 {
		return nil, errors.New("nomor semester tersebut sudah ada pada tahun akademik ini")
	}
	semester := &models.Semester{AcademicYearID: academicYearID, Number: req.Number, StartDate: req.StartDate, EndDate: req.EndDate}
	err = s.DB.Create(semester).Error
	if err == nil {
		semester.SetStatus(time.Now())
	}
	return semester, err
}

func (s *Service) UpdateSemester(yearID, semesterID string, req SemesterRequest) (*models.Semester, error) {
	academicYearID, err := uuid.Parse(yearID)
	if err != nil {
		return nil, errors.New("id tahun akademik tidak valid")
	}
	semesterUUID, err := uuid.Parse(semesterID)
	if err != nil {
		return nil, errors.New("id semester tidak valid")
	}
	if !req.EndDate.After(req.StartDate) {
		return nil, errors.New("tanggal selesai harus setelah tanggal mulai")
	}
	var semester models.Semester
	if err := s.DB.Where("id = ? AND academic_year_id = ?", semesterUUID, academicYearID).First(&semester).Error; err != nil {
		return nil, err
	}
	var year models.AcademicYear
	if err := s.DB.First(&year, "id = ?", academicYearID).Error; err != nil {
		return nil, err
	}
	if year.IsArchived {
		return nil, errors.New("tahun akademik sudah diarsipkan dan tidak dapat diubah")
	}
	if req.StartDate.Before(year.StartDate) || req.EndDate.After(year.EndDate) {
		return nil, errors.New("rentang semester harus berada di dalam rentang tahun akademik")
	}
	var overlap int64
	if err := s.DB.Model(&models.Semester{}).
		Where("academic_year_id = ? AND id <> ? AND start_date <= ? AND end_date >= ?", academicYearID, semesterUUID, req.EndDate, req.StartDate).
		Count(&overlap).Error; err != nil {
		return nil, err
	}
	if overlap > 0 {
		return nil, errors.New("rentang tanggal semester bertumpang tindih dengan semester lain")
	}
	var existing int64
	if err := s.DB.Model(&models.Semester{}).Where("academic_year_id = ? AND number = ? AND id <> ?", academicYearID, req.Number, semesterUUID).Count(&existing).Error; err != nil {
		return nil, err
	}
	if existing > 0 {
		return nil, errors.New("nomor semester tersebut sudah ada pada tahun akademik ini")
	}
	semester.Number, semester.StartDate, semester.EndDate = req.Number, req.StartDate, req.EndDate
	err = s.DB.Save(&semester).Error
	if err == nil {
		semester.SetStatus(time.Now())
	}
	return &semester, err
}

func (s *Service) DeleteSemester(yearID, semesterID string) error {
	academicYearID, err := uuid.Parse(yearID)
	if err != nil {
		return errors.New("id tahun akademik tidak valid")
	}
	semesterUUID, err := uuid.Parse(semesterID)
	if err != nil {
		return errors.New("id semester tidak valid")
	}
	var year models.AcademicYear
	if err := s.DB.First(&year, "id = ?", academicYearID).Error; err != nil {
		return err
	}
	if year.IsArchived {
		return errors.New("tahun akademik sudah diarsipkan dan semesternya tidak dapat diubah")
	}
	return s.DB.Delete(&models.Semester{}, "id = ? AND academic_year_id = ?", semesterUUID, academicYearID).Error
}

func (s *Service) Delete(id string) error {
	yearID, err := uuid.Parse(id)
	if err != nil {
		return errors.New("id tahun akademik tidak valid")
	}
	var year models.AcademicYear
	if err := s.DB.First(&year, "id = ?", yearID).Error; err != nil {
		return err
	}
	if year.IsActive {
		return errors.New("tahun akademik aktif tidak dapat dihapus; aktifkan tahun lain terlebih dahulu")
	}
	var classCount int64
	if err := s.DB.Model(&models.Class{}).Where("academic_year_id = ?", yearID).Count(&classCount).Error; err != nil {
		return err
	}
	if classCount > 0 {
		return errors.New("tahun akademik masih digunakan oleh kelas")
	}
	return s.DB.Delete(&models.AcademicYear{}, "id = ?", yearID).Error
}

func (s *Service) Duplicate(id string, req DuplicateRequest) (*models.AcademicYear, error) {
	sourceID, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("id tahun akademik tidak valid")
	}
	var source models.AcademicYear
	if err := s.DB.Preload("Semesters").First(&source, "id = ?", sourceID).Error; err != nil {
		return nil, err
	}
	var sourceAgendas []models.AcademicAgenda
	if err := s.DB.Where("academic_year_id = ?", sourceID).Find(&sourceAgendas).Error; err != nil {
		return nil, err
	}
	name := req.Name
	if name == "" {
		name = source.Name + " (salinan)"
	}
	year := &models.AcademicYear{Name: name, StartDate: source.StartDate.AddDate(1, 0, 0), EndDate: source.EndDate.AddDate(1, 0, 0)}
	err = s.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(year).Error; err != nil {
			return err
		}
		for _, sourceSemester := range source.Semesters {
			semester := models.Semester{
				AcademicYearID: year.ID,
				Number:         sourceSemester.Number,
				StartDate:      sourceSemester.StartDate.AddDate(1, 0, 0),
				EndDate:        sourceSemester.EndDate.AddDate(1, 0, 0),
			}
			if err := tx.Create(&semester).Error; err != nil {
				return err
			}
		}
		for _, sourceAgenda := range sourceAgendas {
			agenda := models.AcademicAgenda{
				AcademicYearID: year.ID,
				Title:          sourceAgenda.Title,
				Description:    sourceAgenda.Description,
				StartDate:      sourceAgenda.StartDate.AddDate(1, 0, 0),
				EndDate:        sourceAgenda.EndDate.AddDate(1, 0, 0),
				StartTime:      sourceAgenda.StartTime,
				EndTime:        sourceAgenda.EndTime,
				Category:       sourceAgenda.Category,
				Audience:       sourceAgenda.Audience,
				Location:       sourceAgenda.Location,
				Status:         sourceAgenda.Status,
			}
			if err := tx.Create(&agenda).Error; err != nil {
				return err
			}
		}
		return nil
	})
	return year, err
}
