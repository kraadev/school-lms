package models

import (
	"time"

	"github.com/google/uuid"
)

type Semester struct {
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	AcademicYearID uuid.UUID `gorm:"type:uuid;not null;index"`
	Number         int       `gorm:"not null"`
	StartDate      time.Time `gorm:"type:date;not null"`
	EndDate        time.Time `gorm:"type:date;not null"`
	Status         string    `gorm:"-" json:"status"`
	StatusLabel    string    `gorm:"-" json:"status_label"`

	AcademicYear AcademicYear `gorm:"foreignKey:AcademicYearID"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

const (
	SemesterStatusOngoing = "ongoing"
	SemesterStatusEnded   = "ended"
	SemesterStatusComing  = "coming"
)

func (s *Semester) SetStatus(now time.Time) {
	currentDate := dateOnly(now)
	startDate := dateOnly(s.StartDate)
	endDate := dateOnly(s.EndDate)

	switch {
	case currentDate.Before(startDate):
		s.Status = SemesterStatusComing
		s.StatusLabel = "Mendatang"
	case currentDate.After(endDate):
		s.Status = SemesterStatusEnded
		s.StatusLabel = "Berlalu"
	default:
		s.Status = SemesterStatusOngoing
		s.StatusLabel = "Berjalan"
	}
}

func dateOnly(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}
