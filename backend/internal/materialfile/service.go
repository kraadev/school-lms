package materialfile

import (
	"errors"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

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

func (s *Service) Upload(
	userID string,
	materialID string,
	file *multipart.FileHeader,
) (*models.MaterialFile, error) {

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("user ID tidak valid")
	}

	materialUUID, err := uuid.Parse(materialID)
	if err != nil {
		return nil, errors.New("ID materi tidak valid")
	}

	// Cari guru berdasarkan user login
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
	// Pastikan materi milik guru tersebut
	var material models.Material

	err = s.DB.
		First(
			&material,
			"id = ? AND teacher_id = ?",
			materialUUID,
			teacher.ID,
		).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(
				"materi tidak ditemukan atau Anda tidak memiliki akses",
			)
		}

		return nil, err
	}
	if material.Status == "published" {
		return nil, errors.New("lampiran materi yang sudah dipublikasikan tidak dapat diubah")
	}

	// Batasi ukuran file: 20 MB
	const maxFileSize = 20 * 1024 * 1024

	if file.Size > maxFileSize {
		return nil, errors.New(
			"ukuran file maksimal 20 MB",
		)
	}

	// Extension
	extension := strings.ToLower(
		filepath.Ext(file.Filename),
	)

	// Format file yang diizinkan
	allowedExtensions := map[string]bool{
		".pdf":  true,
		".doc":  true,
		".docx": true,
		".ppt":  true,
		".pptx": true,
		".xls":  true,
		".xlsx": true,
		".zip":  true,
		".jpg":  true,
		".jpeg": true,
		".png":  true,
	}

	if !allowedExtensions[extension] {
		return nil, errors.New(
			"format file tidak diizinkan",
		)
	}

	// Buat nama file unik
	fileID := uuid.New().String()

	fileName := fileID + extension

	uploadDir := filepath.Join(
		"uploads",
		"materials",
	)

	if err := os.MkdirAll(
		uploadDir,
		os.ModePerm,
	); err != nil {
		return nil, err
	}

	destination := filepath.Join(
		uploadDir,
		fileName,
	)

	// Buka source file
	source, err := file.Open()
	if err != nil {
		return nil, err
	}

	defer source.Close()

	// Buat destination file
	target, err := os.Create(destination)
	if err != nil {
		return nil, err
	}

	defer target.Close()

	// Copy file
	if _, err := io.Copy(target, source); err != nil {
		return nil, err
	}

	materialFile := models.MaterialFile{
		MaterialID:   material.ID,
		FileName:     fileName,
		OriginalName: file.Filename,
		MimeType:     file.Header.Get("Content-Type"),
		FileSize:     file.Size,
		FilePath:     destination,
	}

	if err := s.DB.
		Create(&materialFile).
		Error; err != nil {

		// Jika database gagal,
		// hapus file yang sudah tersimpan
		_ = os.Remove(destination)

		return nil, err
	}

	return &materialFile, nil
}

func (s *Service) GetFiles(
	userID string,
	materialID string,
) ([]models.MaterialFile, error) {
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
		Where(
			"id = ? AND teacher_id = ?",
			materialUUID,
			teacher.ID,
		).
		First(&material).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("materi tidak ditemukan atau Anda tidak memiliki akses")
		}

		return nil, err
	}

	var files []models.MaterialFile

	err = s.DB.
		Where("material_id = ?", material.ID).
		Order("created_at DESC").
		Find(&files).Error

	if err != nil {
		return nil, err
	}

	return files, nil
}

func (s *Service) DeleteFile(
	userID string,
	fileID string,
) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return errors.New("user ID tidak valid")
	}

	fileUUID, err := uuid.Parse(fileID)
	if err != nil {
		return errors.New("ID file tidak valid")
	}

	var teacher models.Teacher

	err = s.DB.
		Where("user_id = ?", userUUID).
		First(&teacher).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("profile guru tidak ditemukan")
		}

		return err
	}
	var materialFile models.MaterialFile

	err = s.DB.
		Joins("JOIN materials ON materials.id = material_files.material_id").
		Where(
			"material_files.id = ? AND materials.teacher_id = ?",
			fileUUID,
			teacher.ID,
		).
		First(&materialFile).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("file tidak ditemukan atau Anda tidak memiliki akses")
		}

		return err
	}
	var material models.Material
	if err := s.DB.First(&material, "id = ?", materialFile.MaterialID).Error; err != nil {
		return err
	}
	if material.Status == "published" {
		return errors.New("lampiran materi yang sudah dipublikasikan tidak dapat dihapus")
	}

	if err := os.Remove(materialFile.FilePath); err != nil && !os.IsNotExist(err) {
		return err
	}

	return s.DB.Delete(&materialFile).Error
}

func (s *Service) GetStudentFile(
	userID string,
	fileID string,
) (*models.MaterialFile, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("user ID tidak valid")
	}

	fileUUID, err := uuid.Parse(fileID)
	if err != nil {
		return nil, errors.New("ID file tidak valid")
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

	var materialFile models.MaterialFile

	err = s.DB.
		Joins("JOIN materials ON materials.id = material_files.material_id").
		Joins("JOIN class_subjects ON class_subjects.id = materials.class_subject_id").
		Joins("JOIN class_students ON class_students.class_id = class_subjects.class_id").
		Where("material_files.id = ?", fileUUID).
		Where("class_students.student_id = ?", student.ID).
		Where("materials.status = ?", "published").
		First(&materialFile).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("file tidak ditemukan atau Anda tidak memiliki akses")
		}

		return nil, err
	}

	return &materialFile, nil
}

func (s *Service) GetTeacherFile(userID, fileID string) (*models.MaterialFile, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("user ID tidak valid")
	}
	fileUUID, err := uuid.Parse(fileID)
	if err != nil {
		return nil, errors.New("ID file tidak valid")
	}
	var teacher models.Teacher
	if err := s.DB.Where("user_id = ?", userUUID).First(&teacher).Error; err != nil {
		return nil, errors.New("profil guru tidak ditemukan")
	}
	var file models.MaterialFile
	err = s.DB.Joins("JOIN materials ON materials.id = material_files.material_id").Where("material_files.id = ? AND materials.teacher_id = ?", fileUUID, teacher.ID).First(&file).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("file tidak ditemukan atau Anda tidak memiliki akses")
	}
	return &file, err
}
