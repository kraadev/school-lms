package material

import (
	"errors"
	"os"
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

func (s *Service) Create(
	userID string,
	classSubjectID string,
	title string,
	description *string,
	videoURL *string,
	externalURL *string,
	status string,
) (*models.Material, error) {

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("user ID tidak valid")
	}

	classSubjectUUID, err := uuid.Parse(classSubjectID)
	if err != nil {
		return nil, errors.New("class subject ID tidak valid")
	}

	// Cari profile guru berdasarkan user yang sedang login
	var teacher models.Teacher

	err = s.DB.
		Where("user_id = ?", userUUID).
		First(&teacher).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("profile guru tidak ditemukan")
		}

		return nil, err
	}

	// Cari class subject
	var classSubject models.ClassSubject

	err = s.DB.
		First(&classSubject, "id = ?", classSubjectUUID).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("kelas dan mata pelajaran tidak ditemukan")
		}

		return nil, err
	}

	// Authorization:
	// Guru hanya boleh membuat materi
	// untuk ClassSubject miliknya.
	if classSubject.TeacherID == nil {
		return nil, errors.New(
			"mata pelajaran ini belum memiliki guru pengampu",
		)
	}

	if *classSubject.TeacherID != teacher.ID {
		return nil, errors.New(
			"Anda bukan guru pengampu untuk kelas dan mata pelajaran ini",
		)
	}

	// Normalisasi status
	status = strings.ToLower(strings.TrimSpace(status))

	if status == "" {
		status = "draft"
	}

	if status != "draft" && status != "published" {
		return nil, errors.New(
			"status hanya boleh draft atau published",
		)
	}

	material := models.Material{
		ClassSubjectID: classSubject.ID,
		TeacherID:      teacher.ID,
		Title:          title,
		Description:    description,
		VideoURL:       videoURL,
		ExternalURL:    externalURL,
		Status:         status,
	}

	if status == "published" {
		now := time.Now()
		material.PublishedAt = &now
	}

	if err := s.DB.Create(&material).Error; err != nil {
		return nil, err
	}

	return &material, nil
}

func (s *Service) GetMyMaterials(
	userID string,
) ([]models.Material, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("user ID tidak valid")
	}

	var teacher models.Teacher

	err = s.DB.
		Where("user_id = ?", userUUID).
		First(&teacher).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("profile guru tidak ditemukan")
		}

		return nil, err
	}

	var materials []models.Material

	err = s.DB.
		Preload("ClassSubject.Class").
		Preload("ClassSubject.Subject").
		Preload("ClassSubject.Teacher").
		Preload("Files").
		Where("teacher_id = ?", teacher.ID).
		Order("created_at DESC").
		Find(&materials).Error
	if err != nil {
		return nil, err
	}

	return materials, nil
}

func (s *Service) GetMyMaterialByID(
	userID string,
	materialID string,
) (*models.Material, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("user ID tidak valid")
	}

	materialUUID, err := uuid.Parse(materialID)
	if err != nil {
		return nil, errors.New("ID materi tidak valid")
	}

	var teacher models.Teacher

	err = s.DB.
		Where("user_id = ?", userUUID).
		First(&teacher).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("profile guru tidak ditemukan")
		}

		return nil, err
	}

	var material models.Material

	err = s.DB.
		Preload("ClassSubject.Class").
		Preload("ClassSubject.Subject").
		Preload("ClassSubject.Teacher").
		Preload("Files").
		First(
			&material,
			"id = ? AND teacher_id = ?",
			materialUUID,
			teacher.ID,
		).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("materi tidak ditemukan")
		}

		return nil, err
	}

	return &material, nil
}

func (s *Service) Update(
	userID string,
	materialID string,
	title string,
	description *string,
	videoURL *string,
	externalURL *string,
	status string,
) (*models.Material, error) {
	material, err := s.GetMyMaterialByID(userID, materialID)
	if err != nil {
		return nil, err
	}
	if material.Status == "published" {
		return nil, errors.New("materi yang sudah dipublikasikan tidak dapat diedit")
	}

	status = strings.ToLower(strings.TrimSpace(status))
	if status != "draft" && status != "published" {
		return nil, errors.New("status hanya boleh draft atau published")
	}

	material.Title = title
	material.Description = description
	material.VideoURL = videoURL
	material.ExternalURL = externalURL

	if material.Status == "draft" && status == "published" {
		now := time.Now()
		material.PublishedAt = &now
	}

	if status == "draft" {
		material.PublishedAt = nil
	}

	material.Status = status

	if err := s.DB.Save(material).Error; err != nil {
		return nil, err
	}

	return material, nil
}

func (s *Service) Publish(
	userID string,
	materialID string,
) (*models.Material, error) {
	material, err := s.GetMyMaterialByID(userID, materialID)
	if err != nil {
		return nil, err
	}

	if material.Status == "published" {
		return material, nil
	}

	now := time.Now()
	material.Status = "published"
	material.PublishedAt = &now

	if err := s.DB.Save(material).Error; err != nil {
		return nil, err
	}

	return material, nil
}

func (s *Service) Delete(
	userID string,
	materialID string,
) error {
	material, err := s.GetMyMaterialByID(userID, materialID)
	if err != nil {
		return err
	}

	return s.DB.Transaction(func(tx *gorm.DB) error {
		// Hapus relasi material_files terlebih dahulu
		if err := tx.Where("material_id = ?", material.ID).Delete(&models.MaterialFile{}).Error; err != nil {
			return err
		}

		// Hapus record material
		if err := tx.Delete(material).Error; err != nil {
			return err
		}

		// Hapus file fisik dari penyimpanan disk
		for _, f := range material.Files {
			if f.FilePath != "" {
				_ = os.Remove(f.FilePath)
			}
		}

		return nil
	})
}

func (s *Service) GetStudentMaterials(
	userID string,
) ([]models.Material, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("user ID tidak valid")
	}

	var student models.Student

	err = s.DB.
		Where("user_id = ?", userUUID).
		First(&student).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("profile siswa tidak ditemukan")
		}

		return nil, err
	}

	var materials []models.Material

	err = s.DB.
		Preload("Teacher.User").
		Preload("ClassSubject.Class").
		Preload("ClassSubject.Subject").
		Preload("ClassSubject.Teacher").
		Preload("Files").
		Where(`EXISTS (
			SELECT 1
			FROM class_subjects
			JOIN class_students ON class_students.class_id = class_subjects.class_id
			WHERE class_subjects.id = materials.class_subject_id
			AND class_students.student_id = ?
		)`, student.ID).
		Where("materials.status = ?", "published").
		Order("materials.published_at DESC").
		Find(&materials).Error

	if err != nil {
		return nil, err
	}

	return materials, nil
}

func (s *Service) GetStudentMaterialByID(
	userID string,
	materialID string,
) (*models.Material, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("user ID tidak valid")
	}

	materialUUID, err := uuid.Parse(materialID)
	if err != nil {
		return nil, errors.New("ID materi tidak valid")
	}

	var student models.Student

	err = s.DB.
		Where("user_id = ?", userUUID).
		First(&student).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("profile siswa tidak ditemukan")
		}

		return nil, err
	}

	var material models.Material

	err = s.DB.
		Preload("Teacher.User").
		Preload("ClassSubject.Class").
		Preload("ClassSubject.Subject").
		Preload("ClassSubject.Teacher").
		Preload("Files").
		Where("materials.id = ?", materialUUID).
		Where(`EXISTS (
			SELECT 1
			FROM class_subjects
			JOIN class_students ON class_students.class_id = class_subjects.class_id
			WHERE class_subjects.id = materials.class_subject_id
			AND class_students.student_id = ?
		)`, student.ID).
		Where("materials.status = ?", "published").
		First(&material).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("materi tidak ditemukan atau Anda tidak memiliki akses")
		}

		return nil, err
	}

	return &material, nil
}
