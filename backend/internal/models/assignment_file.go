package models

import (
	"time"

	"github.com/google/uuid"
)

type AssignmentFile struct {
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`

	AssignmentID uuid.UUID `gorm:"type:uuid;not null"`

	FileName     string `gorm:"not null"`
	OriginalName string `gorm:"not null"`
	MimeType     string `gorm:"not null"`
	FileSize     int64  `gorm:"not null"`
	FilePath     string `gorm:"not null"`

	Assignment Assignment `gorm:"foreignKey:AssignmentID"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
