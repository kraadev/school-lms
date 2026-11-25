package database

import (
	"log"

	"school-lms/internal/models"

	"gorm.io/gorm"
)

func SeedRoles(db *gorm.DB) {
	roles := []models.Role{
		{Name: "admin"},
		{Name: "guru"},
		{Name: "siswa"},
	}

	for _, role := range roles {
		var existing models.Role

		result := db.Where("name = ?", role.Name).First(&existing)

		if result.Error == gorm.ErrRecordNotFound {
			if err := db.Create(&role).Error; err != nil {
				log.Printf("failed to create role %s: %v", role.Name, err)
			}
		}
	}
}
