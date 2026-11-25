package models

import (
	"time"

	"github.com/google/uuid"
)

type MaterialFile struct {
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`

	MaterialID uuid.UUID `gorm:"type:uuid;not null"`

	FileName     string `gorm:"not null"`
	OriginalName string `gorm:"not null"`
	MimeType     string `gorm:"not null"`
	FileSize     int64  `gorm:"not null"`
	FilePath     string `gorm:"not null"`

	Material Material `gorm:"foreignKey:MaterialID"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
