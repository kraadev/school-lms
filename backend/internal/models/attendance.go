package models

import (
	"time"

	"github.com/google/uuid"
)

type AttendanceSession struct {
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`

	ClassSubjectID uuid.UUID `gorm:"type:uuid;not null"`
	TeacherID      uuid.UUID `gorm:"type:uuid;not null"`

	Title       string `gorm:"not null"`
	Description *string

	AttendanceDate time.Time `gorm:"type:date;not null"`

	StartTime time.Time `gorm:"not null"`
	EndTime   time.Time `gorm:"not null"`

	Status string `gorm:"type:varchar(20);not null;default:'open'"`

	ClassSubject ClassSubject `gorm:"foreignKey:ClassSubjectID"`
	Teacher      Teacher      `gorm:"foreignKey:TeacherID"`

	Records []AttendanceRecord `gorm:"foreignKey:AttendanceSessionID"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
