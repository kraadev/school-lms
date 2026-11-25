package models

import (
	"time"

	"github.com/google/uuid"
)

type AttendanceRecord struct {
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`

	AttendanceSessionID uuid.UUID `gorm:"type:uuid;not null"`
	StudentID           uuid.UUID `gorm:"type:uuid;not null"`

	Status string `gorm:"type:varchar(20);not null;default:'present'"`

	CheckedAt *time.Time

	Note *string

	AttendanceSession AttendanceSession `gorm:"foreignKey:AttendanceSessionID"`
	Student           Student           `gorm:"foreignKey:StudentID"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
