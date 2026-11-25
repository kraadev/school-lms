package models

import (
	"time"

	"github.com/google/uuid"
)

type Class struct {
	ID             uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	AcademicYearID uuid.UUID  `gorm:"type:uuid;not null" json:"academic_year_id"`
	Name           string     `gorm:"not null" json:"name"`
	Grade          int16      `gorm:"not null" json:"grade"`
	MajorID        *uuid.UUID `gorm:"type:uuid;index" json:"major_id,omitempty"`

	AcademicYear AcademicYear `gorm:"foreignKey:AcademicYearID" json:"academic_year,omitempty"`
	Major        *Major       `gorm:"foreignKey:MajorID" json:"major,omitempty"`

	Students      []Student      `gorm:"many2many:class_students" json:"students,omitempty"`
	Subjects      []Subject      `gorm:"many2many:class_subjects" json:"subjects,omitempty"`
	ClassSubjects []ClassSubject `gorm:"foreignKey:ClassID" json:"class_subjects,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
