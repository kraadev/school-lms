package models

import (
	"time"

	"github.com/google/uuid"
)

type AssignmentSubmission struct {
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`

	AssignmentID uuid.UUID `gorm:"type:uuid;not null"`
	StudentID    uuid.UUID `gorm:"type:uuid;not null"`

	SubmittedAt *time.Time

	Status string `gorm:"type:varchar(20);not null;default:'draft'"`

	Score    *float64
	Feedback *string

	Assignment Assignment `gorm:"foreignKey:AssignmentID"`
	Student    Student    `gorm:"foreignKey:StudentID"`

	Files []AssignmentSubmissionFile `gorm:"foreignKey:SubmissionID"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
