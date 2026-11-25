package assignmentfile

import (
	"errors"
	"fmt"
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
	return &Service{DB: db}
}

const maxFileSize = 20 << 20 // 20 MB

var allowedExtensions = map[string]bool{
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

// Upload
func (s *Service) Upload(
	userID string,
	assignmentID string,
	fileHeader *multipart.FileHeader,
) (*models.AssignmentFile, error) {

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("user id tidak valid")
	}

	assignmentUUID, err := uuid.Parse(assignmentID)
	if err != nil {
		return nil, errors.New("assignment id tidak valid")
	}

	// Cari guru
	var teacher models.Teacher

	err = s.DB.
		Where("user_id = ?", userUUID).
		First(&teacher).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("profil guru tidak ditemukan")
	}

	if err != nil {
		return nil, err
	}
	// Pastikan assignment milik guru
	var assignment models.Assignment

	err = s.DB.
		Where(
			"id = ? AND teacher_id = ?",
			assignmentUUID,
			teacher.ID,
		).
		First(&assignment).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("tugas tidak ditemukan")
	}

	if err != nil {
		return nil, err
	}
	if assignment.Status == "published" {
		return nil, errors.New("lampiran tugas yang sudah dipublikasikan tidak dapat diubah")
	}

	if fileHeader == nil {
		return nil, errors.New("file wajib diupload")
	}

	if fileHeader.Size <= 0 {
		return nil, errors.New("file kosong")
	}

	if fileHeader.Size > maxFileSize {
		return nil, errors.New("ukuran file maksimal 20 MB")
	}

	extension := strings.ToLower(
		filepath.Ext(fileHeader.Filename),
	)

	if !allowedExtensions[extension] {
		return nil, errors.New(
			"format file tidak diperbolehkan",
		)
	}

	// Buka file
	file, err := fileHeader.Open()
	if err != nil {
		return nil, errors.New("gagal membuka file")
	}
	defer file.Close()

	// Buat direktori
	uploadDir := filepath.Join(
		"uploads",
		"assignments",
		assignment.ID.String(),
	)

	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, errors.New(
			"gagal membuat folder upload",
		)
	}

	// Nama file random
	fileName := uuid.New().String() + extension

	filePath := filepath.Join(
		uploadDir,
		fileName,
	)

	destination, err := os.Create(filePath)
	if err != nil {
		return nil, errors.New(
			"gagal membuat file tujuan",
		)
	}
	defer destination.Close()

	if _, err := destination.ReadFrom(file); err != nil {
		os.Remove(filePath)

		return nil, errors.New(
			"gagal menyimpan file",
		)
	}

	materialFile := models.AssignmentFile{
		AssignmentID: assignment.ID,
		FileName:     fileName,
		OriginalName: fileHeader.Filename,
		MimeType:     fileHeader.Header.Get("Content-Type"),
		FileSize:     fileHeader.Size,
		FilePath:     filePath,
	}

	if materialFile.MimeType == "" {
		materialFile.MimeType = "application/octet-stream"
	}

	if err := s.DB.Create(&materialFile).Error; err != nil {
		os.Remove(filePath)

		return nil, errors.New(
			"gagal menyimpan informasi file",
		)
	}

	return &materialFile, nil
}

// GetFiles
func (s *Service) GetFiles(
	userID string,
	assignmentID string,
) ([]models.AssignmentFile, error) {

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("user id tidak valid")
	}

	assignmentUUID, err := uuid.Parse(assignmentID)
	if err != nil {
		return nil, errors.New("assignment id tidak valid")
	}

	var teacher models.Teacher

	if err := s.DB.
		Where("user_id = ?", userUUID).
		First(&teacher).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("profil guru tidak ditemukan")
		}

		return nil, err
	}

	var assignment models.Assignment

	if err := s.DB.
		Where(
			"id = ? AND teacher_id = ?",
			assignmentUUID,
			teacher.ID,
		).
		First(&assignment).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("tugas tidak ditemukan")
		}

		return nil, err
	}

	var files []models.AssignmentFile

	if err := s.DB.
		Where("assignment_id = ?", assignment.ID).
		Order("created_at DESC").
		Find(&files).Error; err != nil {

		return nil, err
	}

	return files, nil
}

// DeleteFile
func (s *Service) DeleteFile(
	userID string,
	assignmentID string,
	fileID string,
) error {

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return errors.New("user id tidak valid")
	}

	assignmentUUID, err := uuid.Parse(assignmentID)
	if err != nil {
		return errors.New("assignment id tidak valid")
	}

	fileUUID, err := uuid.Parse(fileID)
	if err != nil {
		return errors.New("file id tidak valid")
	}

	var teacher models.Teacher

	if err := s.DB.
		Where("user_id = ?", userUUID).
		First(&teacher).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("profil guru tidak ditemukan")
		}

		return err
	}

	var assignmentFile models.AssignmentFile

	err = s.DB.
		Joins("JOIN assignments ON assignments.id = assignment_files.assignment_id").
		Where(
			"assignment_files.id = ? AND assignments.id = ? AND assignments.teacher_id = ?",
			fileUUID,
			assignmentUUID,
			teacher.ID,
		).
		First(&assignmentFile).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("file tidak ditemukan")
	}

	if err != nil {
		return err
	}
	var assignment models.Assignment
	if err := s.DB.First(&assignment, "id = ? AND teacher_id = ?", assignmentUUID, teacher.ID).Error; err != nil {
		return err
	}
	if assignment.Status == "published" {
		return errors.New("lampiran tugas yang sudah dipublikasikan tidak dapat dihapus")
	}

	if assignmentFile.FilePath != "" {
		if err := os.Remove(assignmentFile.FilePath); err != nil &&
			!os.IsNotExist(err) {

			return fmt.Errorf(
				"gagal menghapus file: %w",
				err,
			)
		}
	}

	if err := s.DB.Delete(
		&assignmentFile,
	).Error; err != nil {
		return errors.New(
			"gagal menghapus data file",
		)
	}

	return nil
}

func (s *Service) GetStudentFile(
	userID string,
	fileID string,
) (*models.AssignmentFile, error) {
	studentUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("user id tidak valid")
	}

	fileUUID, err := uuid.Parse(fileID)
	if err != nil {
		return nil, errors.New("file id tidak valid")
	}

	var student models.Student

	err = s.DB.
		Where("user_id = ?", studentUUID).
		First(&student).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("profile siswa tidak ditemukan")
		}

		return nil, err
	}

	var file models.AssignmentFile

	err = s.DB.
		Joins("JOIN assignments ON assignments.id = assignment_files.assignment_id").
		Joins("JOIN class_subjects ON class_subjects.id = assignments.class_subject_id").
		Joins("JOIN class_students ON class_students.class_id = class_subjects.class_id").
		Where("assignment_files.id = ?", fileUUID).
		Where("assignments.status = ?", "published").
		Where("class_students.student_id = ?", student.ID).
		First(&file).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("file tidak ditemukan atau tidak dapat diakses")
		}

		return nil, err
	}

	return &file, nil
}

func (s *Service) GetTeacherFile(userID, fileID string) (*models.AssignmentFile, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("user id tidak valid")
	}
	fileUUID, err := uuid.Parse(fileID)
	if err != nil {
		return nil, errors.New("file id tidak valid")
	}
	var teacher models.Teacher
	if err := s.DB.Where("user_id = ?", userUUID).First(&teacher).Error; err != nil {
		return nil, errors.New("profil guru tidak ditemukan")
	}
	var file models.AssignmentFile
	err = s.DB.Joins("JOIN assignments ON assignments.id = assignment_files.assignment_id").Where("assignment_files.id = ? AND assignments.teacher_id = ?", fileUUID, teacher.ID).First(&file).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("file tidak ditemukan atau Anda tidak memiliki akses")
	}
	return &file, err
}
