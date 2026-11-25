package classroom

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"school-lms/internal/models"
)

type Service struct {
	DB *gorm.DB
}

type StudentClassConflictError struct {
	StudentID uuid.UUID
	ClassName string
}

func (e *StudentClassConflictError) Error() string {
	return "siswa sudah terdaftar di kelas " + e.ClassName
}

func NewService(db *gorm.DB) *Service {
	return &Service{
		DB: db,
	}
}

func (s *Service) GetAll() ([]models.Class, error) {
	var classes []models.Class

	err := s.DB.
		Preload("AcademicYear").
		Preload("Major").
		Preload("Students").
		Order("grade ASC, name ASC").
		Find(&classes).Error

	return classes, err
}

func (s *Service) GetByID(id string) (*models.Class, error) {
	classID, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("ID kelas tidak valid")
	}
	var class models.Class
	err = s.DB.Preload("AcademicYear").Preload("Major").Preload("Students").Preload("Subjects").Preload("ClassSubjects.Subject").Preload("ClassSubjects.Teacher").First(&class, "id = ?", classID).Error
	return &class, err
}

func (s *Service) Update(id, name string, grade int16, majorID *string, academicYearID string) (*models.Class, error) {
	classID, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("ID kelas tidak valid")
	}
	yearID, err := uuid.Parse(academicYearID)
	if err != nil {
		return nil, errors.New("ID tahun akademik tidak valid")
	}
	var class models.Class
	if err := s.DB.First(&class, "id = ?", classID).Error; err != nil {
		return nil, err
	}
	var year models.AcademicYear
	if err := s.DB.First(&year, "id = ?", yearID).Error; err != nil {
		return nil, errors.New("tahun ajaran tidak ditemukan")
	}
	var parsedMajorID *uuid.UUID
	if majorID != nil && *majorID != "" {
		majorUUID, parseErr := uuid.Parse(*majorID)
		if parseErr != nil {
			return nil, errors.New("ID jurusan tidak valid")
		}
		var major models.Major
		if err := s.DB.First(&major, "id = ?", majorUUID).Error; err != nil {
			return nil, errors.New("jurusan tidak ditemukan")
		}
		parsedMajorID = &majorUUID
	}
	var duplicate int64
	if err := s.DB.Model(&models.Class{}).Where("name = ? AND academic_year_id = ? AND id <> ?", name, yearID, classID).Count(&duplicate).Error; err != nil {
		return nil, err
	}
	if duplicate > 0 {
		return nil, errors.New("kelas tersebut sudah ada pada tahun ajaran ini")
	}
	class.Name, class.Grade, class.MajorID, class.AcademicYearID = name, grade, parsedMajorID, yearID
	if err := s.DB.Save(&class).Error; err != nil {
		return nil, err
	}
	return s.GetByID(id)
}

func (s *Service) Delete(id string) error {
	classID, err := uuid.Parse(id)
	if err != nil {
		return errors.New("ID kelas tidak valid")
	}
	var members, subjects int64
	if err := s.DB.Table("class_students").Where("class_id = ?", classID).Count(&members).Error; err != nil {
		return err
	}
	if err := s.DB.Table("class_subjects").Where("class_id = ?", classID).Count(&subjects).Error; err != nil {
		return err
	}
	if members > 0 || subjects > 0 {
		return errors.New("kelas masih memiliki siswa atau mata pelajaran")
	}
	result := s.DB.Delete(&models.Class{}, "id = ?", classID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("kelas tidak ditemukan")
	}
	return nil
}

func (s *Service) Create(
	name string,
	grade int16,
	majorID *string,
	academicYearID string,
) (*models.Class, error) {

	// Cek tahun ajaran
	var academicYear models.AcademicYear

	err := s.DB.
		Where("id = ?", academicYearID).
		First(&academicYear).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("tahun ajaran tidak ditemukan")
		}

		return nil, err
	}
	var parsedMajorID *uuid.UUID
	if majorID != nil && *majorID != "" {
		majorUUID, parseErr := uuid.Parse(*majorID)
		if parseErr != nil {
			return nil, errors.New("ID jurusan tidak valid")
		}
		var major models.Major
		if err := s.DB.First(&major, "id = ?", majorUUID).Error; err != nil {
			return nil, errors.New("jurusan tidak ditemukan")
		}
		parsedMajorID = &majorUUID
	}

	// Cek nama kelas pada tahun ajaran yang sama
	var existingClass models.Class

	err = s.DB.
		Where(
			"name = ? AND academic_year_id = ?",
			name,
			academicYear.ID,
		).
		First(&existingClass).Error

	if err == nil {
		return nil, errors.New("kelas tersebut sudah ada pada tahun ajaran ini")
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Buat kelas
	class := models.Class{
		Name:           name,
		Grade:          grade,
		MajorID:        parsedMajorID,
		AcademicYearID: academicYear.ID,
	}

	if err := s.DB.Create(&class).Error; err != nil {
		return nil, err
	}

	class.AcademicYear = academicYear

	return &class, nil
}

func (s *Service) GetStudents(classID string) ([]models.Student, error) {
	classUUID, err := uuid.Parse(classID)
	if err != nil {
		return nil, errors.New("ID kelas tidak valid")
	}

	var class models.Class

	err = s.DB.
		Preload("Students").
		First(&class, "id = ?", classUUID).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("kelas tidak ditemukan")
		}

		return nil, err
	}

	return class.Students, nil
}

func (s *Service) AddStudent(
	classID string,
	studentID string,
) error {
	classUUID, err := uuid.Parse(classID)
	if err != nil {
		return errors.New("ID kelas tidak valid")
	}

	var class models.Class

	err = s.DB.
		First(&class, "id = ?", classUUID).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("kelas tidak ditemukan")
		}

		return err
	}
	var student models.Student
	studentUUID, parseErr := uuid.Parse(studentID)
	if parseErr == nil {
		err = s.DB.First(&student, "id = ?", studentUUID).Error
	} else {
		err = s.DB.Where("nis = ?", studentID).First(&student).Error
	}

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("siswa tidak ditemukan")
		}

		return err
	}
	studentUUID = student.ID

	var currentMembership int64
	if err := s.DB.Table("class_students").Where("class_id = ? AND student_id = ?", classUUID, studentUUID).Count(&currentMembership).Error; err != nil {
		return err
	}
	if currentMembership > 0 {
		return errors.New("siswa sudah terdaftar di kelas tersebut")
	}

	var enrolledClass models.Class
	err = s.DB.
		Table("classes").
		Select("classes.id, classes.name").
		Joins("JOIN class_students ON class_students.class_id = classes.id").
		Where("class_students.student_id = ? AND classes.id <> ?", studentUUID, classUUID).
		First(&enrolledClass).Error
	if err == nil {
		return &StudentClassConflictError{StudentID: studentUUID, ClassName: enrolledClass.Name}
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	return s.DB.
		Table("class_students").
		Create(map[string]interface{}{
			"class_id":   classUUID,
			"student_id": studentUUID,
		}).Error
}

func (s *Service) RemoveStudent(
	classID string,
	studentID string,
) error {
	classUUID, err := uuid.Parse(classID)
	if err != nil {
		return errors.New("ID kelas tidak valid")
	}

	studentUUID, err := uuid.Parse(studentID)
	if err != nil {
		return errors.New("ID siswa tidak valid")
	}

	result := s.DB.
		Table("class_students").
		Where(
			"class_id = ? AND student_id = ?",
			classUUID,
			studentUUID,
		).
		Delete(nil)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("siswa tidak ditemukan di kelas tersebut")
	}

	return nil
}

func (s *Service) GetSubjects(classID string) ([]models.Subject, error) {
	classUUID, err := uuid.Parse(classID)
	if err != nil {
		return nil, errors.New("ID kelas tidak valid")
	}

	var class models.Class

	err = s.DB.
		Preload("Subjects.Teachers").Preload("Subjects.ClassSubjects").
		First(&class, "id = ?", classUUID).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("kelas tidak ditemukan")
		}

		return nil, err
	}

	return class.Subjects, nil
}

func (s *Service) GetTeacherClassSubjects(userID string) ([]models.ClassSubject, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("user ID tidak valid")
	}
	var teacher models.Teacher
	if err := s.DB.Where("user_id = ?", userUUID).First(&teacher).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("profil guru tidak ditemukan")
		}
		return nil, err
	}
	var items []models.ClassSubject
	err = s.DB.
		Preload("Class").
		Preload("Subject").
		Preload("Teacher").
		Where("teacher_id = ?", teacher.ID).
		Joins("JOIN classes ON classes.id = class_subjects.class_id").
		Joins("JOIN subjects ON subjects.id = class_subjects.subject_id").
		Order("classes.name ASC, subjects.name ASC").
		Find(&items).Error
	return items, err
}

func (s *Service) AddSubject(
	classID string,
	subjectID string,
	teacherID string,
) error {
	classUUID, err := uuid.Parse(classID)
	if err != nil {
		return errors.New("ID kelas tidak valid")
	}

	subjectUUID, parseErr := uuid.Parse(subjectID)

	var class models.Class

	if err := s.DB.First(&class, "id = ?", classUUID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("kelas tidak ditemukan")
		}

		return err
	}

	var subject models.Subject
	if parseErr == nil {
		err = s.DB.First(&subject, "id = ?", subjectUUID).Error
	} else {
		err = s.DB.Where("code = ?", subjectID).First(&subject).Error
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("mata pelajaran tidak ditemukan")
	}
	if err != nil {
		return err
	}
	subjectUUID = subject.ID
	teacherUUID, err := uuid.Parse(teacherID)
	if err != nil {
		return errors.New("ID guru tidak valid")
	}

	var eligible int64
	if err := s.DB.Table("subject_teachers").Where("subject_id = ? AND teacher_id = ?", subjectUUID, teacherUUID).Count(&eligible).Error; err != nil {
		return err
	}
	if eligible == 0 {
		return errors.New("guru belum ditugaskan pada mata pelajaran ini")
	}

	var count int64

	err = s.DB.
		Table("class_subjects").
		Where(
			"class_id = ? AND subject_id = ?",
			classUUID,
			subjectUUID,
		).
		Count(&count).Error

	if err != nil {
		return err
	}

	if count > 0 {
		return errors.New("mata pelajaran sudah terdaftar di kelas tersebut")
	}

	return s.DB.
		Table("class_subjects").
		Create(map[string]interface{}{
			"class_id":   classUUID,
			"subject_id": subjectUUID,
			"teacher_id": teacherUUID,
		}).Error
}

func (s *Service) RemoveSubject(
	classID string,
	subjectID string,
) error {
	classUUID, err := uuid.Parse(classID)
	if err != nil {
		return errors.New("ID kelas tidak valid")
	}

	subjectUUID, err := uuid.Parse(subjectID)
	if err != nil {
		return errors.New("ID mata pelajaran tidak valid")
	}

	result := s.DB.
		Table("class_subjects").
		Where(
			"class_id = ? AND subject_id = ?",
			classUUID,
			subjectUUID,
		).
		Delete(nil)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("mata pelajaran tidak ditemukan di kelas tersebut")
	}

	return nil
}

func (s *Service) AssignTeacher(
	classID string,
	subjectID string,
	teacherID string,
) error {
	classUUID, err := uuid.Parse(classID)
	if err != nil {
		return errors.New("ID kelas tidak valid")
	}

	subjectUUID, err := uuid.Parse(subjectID)
	if err != nil {
		return errors.New("ID mata pelajaran tidak valid")
	}

	teacherUUID, err := uuid.Parse(teacherID)
	if err != nil {
		return errors.New("ID guru tidak valid")
	}

	var classSubject models.ClassSubject

	err = s.DB.
		Where(
			"class_id = ? AND subject_id = ?",
			classUUID,
			subjectUUID,
		).
		First(&classSubject).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("mata pelajaran belum ditambahkan ke kelas")
		}

		return err
	}

	var teacher models.Teacher

	err = s.DB.
		First(&teacher, "id = ?", teacherUUID).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("guru tidak ditemukan")
		}

		return err
	}
	var eligible int64
	if err := s.DB.Table("subject_teachers").Where("subject_id = ? AND teacher_id = ?", subjectUUID, teacherUUID).Count(&eligible).Error; err != nil {
		return err
	}
	if eligible == 0 {
		return errors.New("guru belum ditugaskan pada mata pelajaran ini")
	}

	classSubject.TeacherID = &teacher.ID

	return s.DB.
		Model(&classSubject).
		Update("teacher_id", teacher.ID).Error
}

func (s *Service) GetSubjectDetails(classID string) ([]models.ClassSubject, error) {
	classUUID, err := uuid.Parse(classID)
	if err != nil {
		return nil, errors.New("ID kelas tidak valid")
	}

	var class models.Class

	err = s.DB.
		First(&class, "id = ?", classUUID).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("kelas tidak ditemukan")
		}

		return nil, err
	}

	var classSubjects []models.ClassSubject

	err = s.DB.
		Preload("Subject").
		Preload("Subject.Teachers").
		Preload("Teacher").
		Joins("JOIN subjects ON subjects.id = class_subjects.subject_id").
		Where("class_id = ?", classUUID).
		Order("subjects.name ASC").
		Find(&classSubjects).Error

	if err != nil {
		return nil, err
	}

	return classSubjects, nil
}
