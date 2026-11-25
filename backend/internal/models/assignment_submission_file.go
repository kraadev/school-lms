package models

import (
	"time"

	"github.com/google/uuid"
)

type AssignmentSubmissionFile struct {
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`

	SubmissionID uuid.UUID `gorm:"type:uuid;not null"`

	FileName     string `gorm:"not null"`
	OriginalName string `gorm:"not null"`
	MimeType     string `gorm:"not null"`
	FileSize     int64  `gorm:"not null"`
	FilePath     string `gorm:"not null"`

	Submission AssignmentSubmission `gorm:"foreignKey:SubmissionID"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
