package models

import (
	"time"

	"github.com/google/uuid"
)

type Quiz struct {
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	ClassSubjectID uuid.UUID `gorm:"type:uuid;not null"`
	TeacherID      uuid.UUID `gorm:"type:uuid;not null"`

	Title       string `gorm:"not null"`
	Description *string

	Duration int `gorm:"not null"` // dalam menit

	StartAt *time.Time
	EndAt   *time.Time

	MaxAttempts int `gorm:"not null;default:1"`

	Status string `gorm:"not null;default:'draft'"`

	ClassSubject ClassSubject `gorm:"foreignKey:ClassSubjectID"`
	Teacher      Teacher      `gorm:"foreignKey:TeacherID"`

	Questions []QuizQuestion `gorm:"foreignKey:QuizID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Attempts  []QuizAttempt  `gorm:"foreignKey:QuizID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
