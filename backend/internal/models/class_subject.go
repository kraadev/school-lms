package models

import (
	"time"

	"github.com/google/uuid"
)

type ClassSubject struct {
	ID        uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	ClassID   uuid.UUID  `gorm:"type:uuid;not null"`
	SubjectID uuid.UUID  `gorm:"type:uuid;not null"`
	TeacherID *uuid.UUID `gorm:"type:uuid"`

	Class       Class        `gorm:"foreignKey:ClassID"`
	Subject     Subject      `gorm:"foreignKey:SubjectID"`
	Teacher     *Teacher     `gorm:"foreignKey:TeacherID"`
	Materials   []Material   `gorm:"foreignKey:ClassSubjectID"`
	Assignments []Assignment `gorm:"foreignKey:ClassSubjectID"`

	AttendanceSessions []AttendanceSession `gorm:"foreignKey:ClassSubjectID"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
