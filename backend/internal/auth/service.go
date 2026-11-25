package auth

import (
	"errors"

	"school-lms/internal/models"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Service struct {
	DB        *gorm.DB
	JWTSecret string
}

type LoginResult struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

func (s *Service) Login(username string, password string) (*LoginResult, error) {
	var user models.User

	err := s.DB.
		Preload("Role").
		Where("username = ?", username).
		First(&user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("username atau password salah")
	}

	if err != nil {
		return nil, err
	}

	if !user.IsActive {
		return nil, errors.New("akun tidak aktif")
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(password),
	)

	if err != nil {
		return nil, errors.New("username atau password salah")
	}

	token, err := GenerateToken(
		user.ID,
		user.Role.Name,
		s.JWTSecret,
	)

	if err != nil {
		return nil, err
	}

	return &LoginResult{
		Token: token,
		User:  user,
	}, nil
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)

	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func ParseUUID(id string) (uuid.UUID, error) {
	return uuid.Parse(id)
}
