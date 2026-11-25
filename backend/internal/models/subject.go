package models

import (
	"time"

	"github.com/google/uuid"
)

type Subject struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Code        string    `gorm:"unique;not null"`
	Name        string    `gorm:"not null"`
	Description *string

	Classes       []Class        `gorm:"many2many:class_subjects"`
	ClassSubjects []ClassSubject `gorm:"foreignKey:SubjectID"`
	Teachers      []Teacher      `gorm:"many2many:subject_teachers"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
