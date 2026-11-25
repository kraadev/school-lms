package models

import (
	"time"

	"github.com/google/uuid"
)

type Assignment struct {
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`

	ClassSubjectID uuid.UUID `gorm:"type:uuid;not null"`
	TeacherID      uuid.UUID `gorm:"type:uuid;not null"`

	Title       string `gorm:"not null"`
	Description *string
	Deadline    time.Time `gorm:"not null"`

	MaxScore int `gorm:"not null;default:100"`

	Status      string `gorm:"type:varchar(20);not null;default:'draft'"`
	PublishedAt *time.Time

	ClassSubject ClassSubject `gorm:"foreignKey:ClassSubjectID"`
	Teacher      Teacher      `gorm:"foreignKey:TeacherID"`

	Files []AssignmentFile `gorm:"foreignKey:AssignmentID"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
