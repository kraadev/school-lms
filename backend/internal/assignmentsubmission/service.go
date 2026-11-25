package assignmentsubmission

import (
	"errors"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
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

// Submit
func (s *Service) Submit(
	userID string,
	assignmentID string,
	fileHeader *multipart.FileHeader,
) (*models.AssignmentSubmission, error) {

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("user id tidak valid")
	}

	assignmentUUID, err := uuid.Parse(assignmentID)
	if err != nil {
		return nil, errors.New("assignment id tidak valid")
	}

	// Cari siswa
	var student models.Student

	err = s.DB.
		Where("user_id = ?", userUUID).
		First(&student).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("profil siswa tidak ditemukan")
	}

	if err != nil {
		return nil, err
	}

	// Pastikan assignment memang dapat diakses siswa
	var assignment models.Assignment

	err = s.DB.
		Joins("JOIN class_subjects ON class_subjects.id = assignments.class_subject_id").
		Joins("JOIN class_students ON class_students.class_id = class_subjects.class_id").
		Where(
			"assignments.id = ? AND class_students.student_id = ? AND assignments.status = ?",
			assignmentUUID,
			student.ID,
			"published",
		).
		First(&assignment).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("tugas tidak ditemukan")
	}

	if err != nil {
		return nil, err
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
		return nil, errors.New("format file tidak diperbolehkan")
	}

	// Tentukan status berdasarkan deadline
	now := time.Now()

	status := "submitted"

	if now.After(assignment.Deadline) {
		status = "late"
	}

	// Satu siswa hanya boleh mengumpulkan satu kali untuk setiap tugas.
	var submission models.AssignmentSubmission

	err = s.DB.
		Where(
			"assignment_id = ? AND student_id = ?",
			assignment.ID,
			student.ID,
		).
		First(&submission).Error

	if err == nil {
		return nil, errors.New("tugas sudah dikumpulkan dan tidak dapat dikumpulkan kembali")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	submission = models.AssignmentSubmission{
		AssignmentID: assignment.ID,
		StudentID:    student.ID,
		SubmittedAt:  &now,
		Status:       status,
	}
	if err := s.DB.Create(&submission).Error; err != nil {
		return nil, err
	}

	var fileCount int64

	err = s.DB.
		Model(&models.AssignmentSubmissionFile{}).
		Where("submission_id = ?", submission.ID).
		Count(&fileCount).Error

	if err != nil {
		return nil, err
	}

	if fileCount >= 5 {
		return nil, errors.New(
			"maksimal 5 file untuk satu submission",
		)
	}

	// Buka file
	file, err := fileHeader.Open()
	if err != nil {
		return nil, errors.New("gagal membuka file")
	}
	defer file.Close()

	// Folder submission
	uploadDir := filepath.Join(
		"uploads",
		"assignment-submissions",
		submission.ID.String(),
	)

	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, errors.New(
			"gagal membuat folder upload",
		)
	}

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

	submissionFile := models.AssignmentSubmissionFile{
		SubmissionID: submission.ID,
		FileName:     fileName,
		OriginalName: fileHeader.Filename,
		MimeType:     fileHeader.Header.Get("Content-Type"),
		FileSize:     fileHeader.Size,
		FilePath:     filePath,
	}

	if submissionFile.MimeType == "" {
		submissionFile.MimeType = "application/octet-stream"
	}

	if err := s.DB.Create(&submissionFile).Error; err != nil {
		os.Remove(filePath)

		return nil, errors.New(
			"gagal menyimpan informasi file",
		)
	}

	// Ambil kembali submission beserta file
	err = s.DB.
		Preload("Files").
		Preload("Assignment").
		First(&submission, "id = ?", submission.ID).Error

	if err != nil {
		return nil, err
	}

	if s.NotificationService != nil {
		refType := "assignment"
		var teacher models.Teacher
		if err := s.DB.Where("id = ?", assignment.TeacherID).First(&teacher).Error; err == nil {
			_ = s.NotificationService.NotifyUser(teacher.UserID, "assignment_submission", "Tugas Dikumpulkan", student.FullName+" telah mengumpulkan tugas "+assignment.Title, &refType, &assignment.ID)
		}
	}

	return &submission, nil
}

// GetStudentSubmission
func (s *Service) GetStudentSubmission(
	userID string,
	assignmentID string,
) (*models.AssignmentSubmission, error) {

	studentUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("user id tidak valid")
	}

	assignmentUUID, err := uuid.Parse(assignmentID)
	if err != nil {
		return nil, errors.New("assignment id tidak valid")
	}

	// Cari profile siswa
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

	// Pastikan assignment memang bisa diakses siswa
	var assignment models.Assignment

	err = s.DB.
		Joins("JOIN class_subjects ON class_subjects.id = assignments.class_subject_id").
		Joins("JOIN class_students ON class_students.class_id = class_subjects.class_id").
		Where("assignments.id = ?", assignmentUUID).
		Where("assignments.status = ?", "published").
		Where("class_students.student_id = ?", student.ID).
		Preload("ClassSubject").
		Preload("ClassSubject.Class").
		Preload("ClassSubject.Subject").
		First(&assignment).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(
				"tugas tidak ditemukan atau tidak dapat diakses",
			)
		}

		return nil, err
	}

	// Cari submission milik siswa tersebut
	var submission models.AssignmentSubmission

	err = s.DB.
		Where("assignment_id = ?", assignment.ID).
		Where("student_id = ?", student.ID).
		Preload("Files").
		Preload("Assignment").
		Preload("Assignment.ClassSubject").
		Preload("Assignment.ClassSubject.Class").
		Preload("Assignment.ClassSubject.Subject").
		First(&submission).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("belum ada pengumpulan")
		}

		return nil, err
	}

	return &submission, nil
}

// GetTeacherSubmissions
func (s *Service) GetTeacherSubmissions(
	userID string,
	assignmentID string,
) ([]models.AssignmentSubmission, error) {

	teacherUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("user id tidak valid")
	}

	assignmentUUID, err := uuid.Parse(assignmentID)
	if err != nil {
		return nil, errors.New("assignment id tidak valid")
	}

	// Cari profile guru
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

	// Pastikan assignment milik guru
	var assignment models.Assignment

	err = s.DB.
		Where("id = ?", assignmentUUID).
		Where("teacher_id = ?", teacher.ID).
		First(&assignment).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(
				"tugas tidak ditemukan atau bukan milik guru",
			)
		}

		return nil, err
	}

	var submissions []models.AssignmentSubmission

	err = s.DB.
		Where("assignment_id = ?", assignment.ID).
		Preload("Student").
		Preload("Files").
		Order("submitted_at ASC").
		Find(&submissions).Error

	if err != nil {
		return nil, err
	}

	return submissions, nil
}

func (s *Service) GetTeacherSubmission(
	userID string,
	submissionID string,
) (*models.AssignmentSubmission, error) {

	teacherUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("user id tidak valid")
	}

	submissionUUID, err := uuid.Parse(submissionID)
	if err != nil {
		return nil, errors.New("submission id tidak valid")
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

	var submission models.AssignmentSubmission

	err = s.DB.
		Joins("JOIN assignments ON assignments.id = assignment_submissions.assignment_id").
		Where("assignment_submissions.id = ?", submissionUUID).
		Where("assignments.teacher_id = ?", teacher.ID).
		Preload("Student").
		Preload("Assignment").
		Preload("Assignment.ClassSubject").
		Preload("Assignment.ClassSubject.Class").
		Preload("Assignment.ClassSubject.Subject").
		Preload("Files").
		First(&submission).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(
				"submission tidak ditemukan atau bukan milik tugas guru",
			)
		}

		return nil, err
	}

	return &submission, nil
}

func (s *Service) GradeSubmission(
	userID string,
	submissionID string,
	score float64,
	feedback *string,
) (*models.AssignmentSubmission, error) {

	teacherUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("user id tidak valid")
	}

	submissionUUID, err := uuid.Parse(submissionID)
	if err != nil {
		return nil, errors.New("submission id tidak valid")
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

	var submission models.AssignmentSubmission

	err = s.DB.
		Joins("JOIN assignments ON assignments.id = assignment_submissions.assignment_id").
		Where("assignment_submissions.id = ?", submissionUUID).
		Where("assignments.teacher_id = ?", teacher.ID).
		Preload("Assignment").
		Preload("Student").
		Preload("Files").
		First(&submission).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(
				"submission tidak ditemukan atau bukan milik tugas guru",
			)
		}

		return nil, err
	}

	if submission.SubmittedAt == nil {
		return nil, errors.New(
			"siswa belum mengumpulkan tugas",
		)
	}

	if score < 0 {
		return nil, errors.New(
			"nilai tidak boleh kurang dari 0",
		)
	}

	if score > float64(submission.Assignment.MaxScore) {
		return nil, errors.New(
			"nilai melebihi nilai maksimal tugas",
		)
	}

	submission.Score = &score
	submission.Feedback = feedback
	submission.Status = "graded"

	if err := s.DB.Save(&submission).Error; err != nil {
		return nil, err
	}

	return &submission, nil
}

func (s *Service) DeleteStudentSubmissionFile(
	userID string,
	fileID string,
) error {
	studentUUID, err := uuid.Parse(userID)
	if err != nil {
		return errors.New("user id tidak valid")
	}

	fileUUID, err := uuid.Parse(fileID)
	if err != nil {
		return errors.New("file id tidak valid")
	}

	var student models.Student

	err = s.DB.
		Where("user_id = ?", studentUUID).
		First(&student).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("profile siswa tidak ditemukan")
		}

		return err
	}

	var file models.AssignmentSubmissionFile

	err = s.DB.
		Joins(`
			JOIN assignment_submissions
			ON assignment_submissions.id =
			assignment_submission_files.submission_id
		`).
		Where("assignment_submission_files.id = ?", fileUUID).
		Where("assignment_submissions.student_id = ?", student.ID).
		Preload("Submission").
		First(&file).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("file tidak ditemukan atau bukan milik siswa")
		}

		return err
	}

	if file.Submission.Status == "graded" {
		return errors.New("file tidak dapat diubah karena tugas sudah dinilai")
	}

	filePath := file.FilePath

	if err := s.DB.Delete(&file).Error; err != nil {
		return err
	}

	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}
