package models

import (
	"time"

	"github.com/google/uuid"
)

type Teacher struct {
	ID       uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID   uuid.UUID `gorm:"type:uuid;unique;not null" json:"user_id"`
	NIP      *string   `gorm:"column:nip;unique" json:"nip,omitempty"`
	FullName string    `gorm:"not null" json:"full_name"`
	Phone    *string   `json:"phone,omitempty"`

	ClassSubjects []ClassSubject `gorm:"foreignKey:TeacherID" json:"class_subjects,omitempty"`
	Materials     []Material     `gorm:"foreignKey:TeacherID" json:"materials,omitempty"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`

	AttendanceSessions []AttendanceSession `gorm:"foreignKey:TeacherID" json:"attendance_sessions,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
