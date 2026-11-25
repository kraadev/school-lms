package assignment

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
	return &Service{DB: db}
}

// GetAllTeacherAssignments
func (s *Service) GetAllTeacherAssignments(userID string) ([]models.Assignment, error) {
	teacherUserID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("user id tidak valid")
	}

	var teacher models.Teacher

	if err := s.DB.
		Where("user_id = ?", teacherUserID).
		First(&teacher).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("profil guru tidak ditemukan")
		}

		return nil, err
	}

	var assignments []models.Assignment

	err = s.DB.
		Preload("ClassSubject").
		Preload("ClassSubject.Class").
		Preload("ClassSubject.Subject").
		Preload("Teacher.User").
		Preload("Files").
		Where("teacher_id = ?", teacher.ID).
		Order("created_at DESC").
		Find(&assignments).Error

	return assignments, err
}

// GetTeacherAssignment
func (s *Service) GetTeacherAssignment(
	userID string,
	assignmentID string,
) (*models.Assignment, error) {

	teacherUserID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("user id tidak valid")
	}

	id, err := uuid.Parse(assignmentID)
	if err != nil {
		return nil, errors.New("assignment id tidak valid")
	}

	var teacher models.Teacher

	if err := s.DB.
		Where("user_id = ?", teacherUserID).
		First(&teacher).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("profil guru tidak ditemukan")
		}

		return nil, err
	}

	var assignment models.Assignment

	err = s.DB.
		Preload("ClassSubject").
		Preload("ClassSubject.Class").
		Preload("ClassSubject.Subject").
		Preload("Files").
		Where("id = ? AND teacher_id = ?", id, teacher.ID).
		First(&assignment).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("tugas tidak ditemukan")
	}

	if err != nil {
		return nil, err
	}

	return &assignment, nil
}

// Create
func (s *Service) Create(
	userID string,
	classSubjectID string,
	title string,
	description *string,
	deadline time.Time,
	maxScore int,
) (*models.Assignment, error) {

	teacherUserID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("user id tidak valid")
	}

	classSubjectUUID, err := uuid.Parse(classSubjectID)
	if err != nil {
		return nil, errors.New("class subject id tidak valid")
	}

	var teacher models.Teacher

	if err := s.DB.
		Where("user_id = ?", teacherUserID).
		First(&teacher).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("profil guru tidak ditemukan")
		}

		return nil, err
	}

	// Pastikan ClassSubject memang diajar oleh guru tersebut.
	var classSubject models.ClassSubject

	err = s.DB.
		Where(
			"id = ? AND teacher_id = ?",
			classSubjectUUID,
			teacher.ID,
		).
		First(&classSubject).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New(
			"mata pelajaran tersebut bukan tanggung jawab guru ini",
		)
	}

	if err != nil {
		return nil, err
	}

	if title == "" {
		return nil, errors.New("judul tugas wajib diisi")
	}

	if maxScore <= 0 {
		return nil, errors.New("maksimal nilai harus lebih dari 0")
	}

	if deadline.IsZero() {
		return nil, errors.New("deadline wajib diisi")
	}

	assignment := models.Assignment{
		ClassSubjectID: classSubject.ID,
		TeacherID:      teacher.ID,
		Title:          title,
		Description:    description,
		Deadline:       deadline,
		MaxScore:       maxScore,
		Status:         "draft",
	}

	if err := s.DB.Create(&assignment).Error; err != nil {
		return nil, err
	}

	return s.GetTeacherAssignment(
		userID,
		assignment.ID.String(),
	)
}

// Update
func (s *Service) Update(
	userID string,
	assignmentID string,
	classSubjectID string,
	title string,
	description *string,
	deadline time.Time,
	maxScore int,
) (*models.Assignment, error) {

	assignment, err := s.GetTeacherAssignment(
		userID,
		assignmentID,
	)

	if err != nil {
		return nil, err
	}
	if assignment.Status == "published" {
		return nil, errors.New("tugas yang sudah dipublikasikan tidak dapat diedit")
	}

	classSubjectUUID, err := uuid.Parse(classSubjectID)
	if err != nil {
		return nil, errors.New("class subject id tidak valid")
	}

	if title == "" {
		return nil, errors.New("judul tugas wajib diisi")
	}

	if deadline.IsZero() {
		return nil, errors.New("deadline wajib diisi")
	}

	if maxScore <= 0 {
		return nil, errors.New("maksimal nilai harus lebih dari 0")
	}

	var teacher models.Teacher

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("user id tidak valid")
	}

	if err := s.DB.
		Where("user_id = ?", userUUID).
		First(&teacher).Error; err != nil {
		return nil, err
	}

	var classSubject models.ClassSubject

	err = s.DB.
		Where(
			"id = ? AND teacher_id = ?",
			classSubjectUUID,
			teacher.ID,
		).
		First(&classSubject).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New(
			"mata pelajaran tersebut bukan tanggung jawab guru ini",
		)
	}

	if err != nil {
		return nil, err
	}

	if assignment.Status == "published" &&
		assignment.ClassSubjectID != classSubject.ID {
		return nil, errors.New(
			"tugas yang sudah dipublikasikan tidak dapat dipindahkan ke mata pelajaran lain",
		)
	}

	assignment.ClassSubjectID = classSubject.ID
	assignment.Title = title
	assignment.Description = description
	assignment.Deadline = deadline
	assignment.MaxScore = maxScore

	if err := s.DB.Save(assignment).Error; err != nil {
		return nil, err
	}

	return s.GetTeacherAssignment(
		userID,
		assignment.ID.String(),
	)
}

func (s *Service) Publish(
	userID string,
	assignmentID string,
) (*models.Assignment, error) {
	assignment, err := s.GetTeacherAssignment(userID, assignmentID)
	if err != nil {
		return nil, err
	}

	if assignment.Status == "published" {
		return nil, errors.New("tugas sudah dipublikasikan")
	}

	if assignment.Deadline.IsZero() {
		return nil, errors.New("deadline tugas belum ditentukan")
	}

	now := time.Now()
	assignment.Status = "published"
	assignment.PublishedAt = &now

	if err := s.DB.Save(assignment).Error; err != nil {
		return nil, err
	}

	if s.NotificationService != nil {
		refType := "assignment"
		_ = s.NotificationService.NotifyClass(assignment.ClassSubject.ClassID, "assignment", "Tugas Baru: "+assignment.Title, "Tugas baru telah diterbitkan dengan batas waktu "+assignment.Deadline.Format("02 Jan 2006 15:04"), &refType, &assignment.ID)
	}

	return s.GetTeacherAssignment(userID, assignment.ID.String())
}

// Delete
func (s *Service) Delete(
	userID string,
	assignmentID string,
) error {

	assignment, err := s.GetTeacherAssignment(
		userID,
		assignmentID,
	)

	if err != nil {
		return err
	}

	if err := s.DB.Delete(&assignment).Error; err != nil {
		return err
	}

	return nil
}

// GetAllStudentAssignments
func (s *Service) GetAllStudentAssignments(
	userID string,
) ([]models.Assignment, error) {

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("user id tidak valid")
	}

	// Cari profil siswa
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

	var assignments []models.Assignment

	err = s.DB.
		Joins("JOIN class_subjects ON class_subjects.id = assignments.class_subject_id").
		Joins("JOIN class_students ON class_students.class_id = class_subjects.class_id").
		Where(
			"class_students.student_id = ? AND assignments.status = ?",
			student.ID,
			"published",
		).
		Preload("ClassSubject").
		Preload("ClassSubject.Class").
		Preload("ClassSubject.Subject").
		Preload("Files").
		Order("assignments.deadline ASC").
		Find(&assignments).Error

	return assignments, err
}

// GetStudentAssignment
func (s *Service) GetStudentAssignment(
	userID string,
	assignmentID string,
) (*models.Assignment, error) {

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
		Preload("ClassSubject").
		Preload("ClassSubject.Class").
		Preload("ClassSubject.Subject").
		Preload("Files").
		First(&assignment).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("tugas tidak ditemukan")
	}

	if err != nil {
		return nil, err
	}

	return &assignment, nil
}
