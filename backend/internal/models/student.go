package models

import (
	"time"

	"github.com/google/uuid"
)

type Student struct {
	ID       uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID   uuid.UUID `gorm:"type:uuid;unique;not null" json:"user_id"`
	NIS      string    `gorm:"column:nis;unique;not null" json:"nis"`
	NISN     *string   `gorm:"column:nisn;unique" json:"nisn,omitempty"`
	FullName string    `gorm:"not null" json:"full_name"`
	Gender   *string   `json:"gender,omitempty"`
	Phone    *string   `json:"phone,omitempty"`
	Address  *string   `json:"address,omitempty"`

	User    User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Classes []Class `gorm:"many2many:class_students" json:"classes,omitempty"`

	// Relasi ke QuizAttempt
	QuizAttempts      []QuizAttempt      `gorm:"foreignKey:StudentID" json:"quiz_attempts,omitempty"`
	AttendanceRecords []AttendanceRecord `gorm:"foreignKey:StudentID" json:"attendance_records,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
