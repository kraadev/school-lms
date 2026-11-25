package database

import (
	"errors"
	"fmt"

	"school-lms/internal/auth"
	"school-lms/internal/models"

	"gorm.io/gorm"
)

func SeedAdmin(db *gorm.DB, initialPassword string) error {
	var role models.Role

	if err := db.Where("name = ?", "admin").First(&role).Error; err != nil {
		return fmt.Errorf("find admin role: %w", err)
	}

	var user models.User

	result := db.Where("username = ?", "admin").First(&user)

	if result.Error == nil {
		return nil
	}
	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return fmt.Errorf("find initial admin user: %w", result.Error)
	}
	if initialPassword == "" {
		return errors.New("ADMIN_INITIAL_PASSWORD must be set to create the initial admin account")
	}

	password, err := auth.HashPassword(initialPassword)

	if err != nil {
		return fmt.Errorf("hash initial admin password: %w", err)
	}

	user = models.User{
		RoleID:       role.ID,
		Username:     "admin",
		PasswordHash: password,
		IsActive:     true,
	}

	if err := db.Create(&user).Error; err != nil {
		return fmt.Errorf("create initial admin user: %w", err)
	}

	admin := models.Admin{
		UserID:   user.ID,
		FullName: "Administrator",
	}

	if err := db.Create(&admin).Error; err != nil {
		return fmt.Errorf("create initial admin profile: %w", err)
	}

	return nil
}
