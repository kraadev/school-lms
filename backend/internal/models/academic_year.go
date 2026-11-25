package models

import (
	"time"

	"github.com/google/uuid"
)

type AcademicYear struct {
	ID         uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name       string    `gorm:"not null"`
	IsActive   bool      `gorm:"default:false"`
	IsArchived bool      `gorm:"default:false"`
	StartDate  time.Time `gorm:"type:date;not null"`
	EndDate    time.Time `gorm:"type:date;not null"`

	Semesters []Semester `gorm:"foreignKey:AcademicYearID;constraint:OnDelete:CASCADE"`
	Classes   []Class    `gorm:"foreignKey:AcademicYearID"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
