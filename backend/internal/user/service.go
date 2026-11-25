package user

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"school-lms/internal/auth"
	"school-lms/internal/models"
)

type Service struct {
	DB *gorm.DB
}

type DeleteConflictError struct {
	UserID        uuid.UUID
	MaterialCount int64
}

func (e *DeleteConflictError) Error() string {
	return "guru tidak dapat dihapus karena masih memiliki materi yang tersimpan"
}

func NewService(db *gorm.DB) *Service {
	return &Service{
		DB: db,
	}
}

func (s *Service) GetAll() ([]models.User, error) {
	var users []models.User

	err := s.DB.Preload("Role").Preload("Teacher").Preload("Teacher.ClassSubjects.Subject").Preload("Teacher.ClassSubjects.Class").Preload("Student.Classes").Find(&users).Error
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (s *Service) GetByID(id string) (*models.User, error) {
	userID, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("id user tidak valid")
	}
	var user models.User
	err = s.DB.Preload("Role").Preload("Teacher").Preload("Teacher.ClassSubjects.Subject").Preload("Teacher.ClassSubjects.Class").Preload("Student.Classes").First(&user, "id = ?", userID).Error
	return &user, err
}

func (s *Service) GetProfile(userIDStr string) (*models.User, error) {
	user, err := s.GetByID(userIDStr)
	if err != nil {
		return nil, err
	}
	if user.Student != nil && user.Student.FullName != "" {
		user.DisplayName = user.Student.FullName
	} else if user.Teacher != nil && user.Teacher.FullName != "" {
		user.DisplayName = user.Teacher.FullName
	} else {
		user.DisplayName = user.Username
	}
	return user, nil
}

type UpdateProfileRequest struct {
	Email           string  `json:"email"`
	Phone           *string `json:"phone"`
	NewPassword     string  `json:"new_password"`
	ConfirmPassword string  `json:"confirm_password"`
}

func (s *Service) UpdateProfile(userIDStr string, request UpdateProfileRequest) (*models.User, error) {
	user, err := s.GetByID(userIDStr)
	if err != nil {
		return nil, err
	}
	if request.NewPassword != "" {
		if len(request.NewPassword) < 6 {
			return nil, errors.New("password baru minimal 6 karakter")
		}
		if request.NewPassword != request.ConfirmPassword {
			return nil, errors.New("konfirmasi password tidak cocok")
		}
	}

	tx := s.DB.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	emailStr := strings.TrimSpace(request.Email)
	if emailStr != "" {
		var duplicate models.User
		if err := tx.Where("email = ? AND id <> ?", emailStr, user.ID).First(&duplicate).Error; err == nil {
			tx.Rollback()
			return nil, errors.New("email sudah digunakan oleh akun lain")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			return nil, err
		}
		user.Email = &emailStr
	} else {
		user.Email = nil
	}

	if err := tx.Save(user).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if request.NewPassword != "" {
		hashed, hashErr := auth.HashPassword(request.NewPassword)
		if hashErr != nil {
			tx.Rollback()
			return nil, errors.New("gagal mengenkripsi password")
		}
		user.PasswordHash = hashed
		if err := tx.Save(user).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if user.Teacher != nil {
		user.Teacher.Phone = request.Phone
		if err := tx.Save(user.Teacher).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if user.Student != nil {
		user.Student.Phone = request.Phone
		if err := tx.Save(user.Student).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return s.GetProfile(userIDStr)
}

type UpdateRequest struct {
	Username string  `json:"username" binding:"required"`
	Password string  `json:"password"`
	Email    string  `json:"email"`
	IsActive *bool   `json:"is_active"`
	FullName string  `json:"full_name" binding:"required"`
	NIP      *string `json:"nip"`
	Phone    *string `json:"phone"`
	NIS      string  `json:"nis"`
	NISN     *string `json:"nisn"`
	Gender   *string `json:"gender"`
	Address  *string `json:"address"`
}

func (s *Service) Update(id string, request UpdateRequest) (*models.User, error) {
	user, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}
	tx := s.DB.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	var duplicate models.User
	if err := tx.Where("username = ? AND id <> ?", request.Username, user.ID).First(&duplicate).Error; err == nil {
		tx.Rollback()
		return nil, errors.New("username sudah digunakan")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		return nil, err
	}
	user.Username, user.Email = request.Username, nullableString(request.Email)
	if request.IsActive != nil {
		user.IsActive = *request.IsActive
	}
	if request.Password != "" {
		hashed, hashErr := auth.HashPassword(request.Password)
		if hashErr != nil {
			tx.Rollback()
			return nil, errors.New("gagal mengenkripsi password")
		}
		user.PasswordHash = hashed
	}
	if err := tx.Save(user).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if user.Teacher != nil {
		user.Teacher.FullName, user.Teacher.NIP, user.Teacher.Phone = request.FullName, request.NIP, request.Phone
		if err := tx.Save(user.Teacher).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}
	if user.Student != nil {
		user.Student.FullName, user.Student.NIS, user.Student.NISN = request.FullName, request.NIS, request.NISN
		user.Student.Gender, user.Student.Phone, user.Student.Address = request.Gender, request.Phone, request.Address
		if err := tx.Save(user.Student).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	return s.GetByID(id)
}

func nullableString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func (s *Service) Delete(id string) error {
	userID, err := uuid.Parse(id)
	if err != nil {
		return errors.New("id user tidak valid")
	}
	var teacher models.Teacher
	if err := s.DB.Where("user_id = ?", userID).First(&teacher).Error; err == nil {
		var materialCount int64
		if err := s.DB.Model(&models.Material{}).Where("teacher_id = ?", teacher.ID).Count(&materialCount).Error; err != nil {
			return err
		}
		if materialCount > 0 {
			return &DeleteConflictError{UserID: userID, MaterialCount: materialCount}
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	result := s.DB.Delete(&models.User{}, "id = ?", userID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("user tidak ditemukan")
	}
	return nil
}

func (s *Service) GetStudents() ([]models.Student, error) {
	var students []models.Student
	if err := s.DB.Order("full_name ASC").Find(&students).Error; err != nil {
		return nil, err
	}
	return students, nil
}

func (s *Service) SearchStudents(query string) ([]models.Student, error) {
	var students []models.Student
	term := "%" + strings.TrimSpace(query) + "%"
	err := s.DB.
		Preload("Classes").
		Where(`students.nis ILIKE ? OR students.full_name ILIKE ? OR EXISTS (
			SELECT 1
			FROM class_students
			JOIN classes ON classes.id = class_students.class_id
			WHERE class_students.student_id = students.id AND classes.name ILIKE ?
		)`, term, term, term).
		Order("students.full_name ASC").
		Limit(8).
		Find(&students).Error
	return students, err
}

func (s *Service) Create(request CreateUserRequest) (*models.User, error) {
	tx := s.DB.Begin()

	if tx.Error != nil {
		return nil, tx.Error
	}

	// =========================
	// CEK USERNAME
	// =========================

	var existingUser models.User

	err := tx.
		Where("username = ?", request.Username).
		First(&existingUser).Error

	if err == nil {
		tx.Rollback()
		return nil, errors.New("username sudah digunakan")
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		return nil, err
	}

	// =========================
	// VALIDASI ROLE
	// =========================

	if request.Role != "guru" && request.Role != "siswa" {
		tx.Rollback()
		return nil, errors.New("role hanya boleh guru atau siswa")
	}

	var role models.Role

	err = tx.
		Where("name = ?", request.Role).
		First(&role).Error

	if err != nil {
		tx.Rollback()
		return nil, errors.New("role tidak ditemukan")
	}

	// =========================
	// HASH PASSWORD
	// =========================

	hashedPassword, err := auth.HashPassword(request.Password)
	if err != nil {
		tx.Rollback()
		return nil, errors.New("gagal mengenkripsi password")
	}

	// =========================
	// BUAT USER
	// =========================

	var emailPtr *string
	if request.Email != "" {
		emailPtr = &request.Email
	}

	user := models.User{
		Username:     request.Username,
		PasswordHash: hashedPassword,
		Email:        emailPtr,
		IsActive:     true,
		RoleID:       role.ID,
	}

	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// =========================
	// BUAT PROFILE
	// =========================

	if request.Role == "guru" {
		teacher := models.Teacher{
			UserID:   user.ID,
			FullName: request.FullName,
			NIP:      request.NIP,
			Phone:    request.Phone,
		}

		if err := tx.Create(&teacher).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if request.Role == "siswa" {
		if request.NIS == "" {
			tx.Rollback()
			return nil, errors.New("NIS wajib diisi untuk siswa")
		}

		student := models.Student{
			UserID:   user.ID,
			NIS:      request.NIS,
			NISN:     request.NISN,
			FullName: request.FullName,
			Gender:   request.Gender,
			Phone:    request.Phone,
			Address:  request.Address,
		}

		if err := tx.Create(&student).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	// =========================
	// COMMIT
	// =========================

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	user.Role = role

	return &user, nil
}
