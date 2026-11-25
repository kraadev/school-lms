package models

import (
	"time"

	"github.com/google/uuid"
)

type AcademicAgenda struct {
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	AcademicYearID uuid.UUID `gorm:"type:uuid;not null;index"`
	Title          string    `gorm:"not null"`
	Description    *string   `gorm:"type:text"`
	StartDate      time.Time `gorm:"type:date;not null"`
	EndDate        time.Time `gorm:"type:date;not null"`
	StartTime      *string
	EndTime        *string
	Category       string `gorm:"not null;default:'Kegiatan sekolah'"`
	Audience       string `gorm:"not null;default:'Semua'"`
	Location       *string
	Status         string     `gorm:"not null;default:'Terjadwal'"`
	CreatedByID    *uuid.UUID `gorm:"type:uuid;index"`

	AcademicYear AcademicYear `gorm:"foreignKey:AcademicYearID"`
	CreatedBy    *User        `gorm:"foreignKey:CreatedByID"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
