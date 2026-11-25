package models

import (
	"time"

	"github.com/google/uuid"
)

type QuizAttempt struct {
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`

	QuizID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_quiz_student_attempt" json:"quiz_id"`

	StudentID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_quiz_student_attempt" json:"student_id"`

	AttemptNumber int `gorm:"not null;uniqueIndex:idx_quiz_student_attempt" json:"attempt_number"`

	StartedAt time.Time `gorm:"not null" json:"started_at"`

	DeadlineAt time.Time `gorm:"not null" json:"deadline_at"`

	SubmittedAt *time.Time `json:"submitted_at"`

	Status string `gorm:"not null;default:'in_progress'" json:"status"`

	Score *float64 `json:"score"`

	Quiz    Quiz    `gorm:"foreignKey:QuizID" json:"quiz,omitempty"`
	Student Student `gorm:"foreignKey:StudentID" json:"student,omitempty"`

	Answers []QuizAttemptAnswer `gorm:"foreignKey:AttemptID" json:"answers,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
