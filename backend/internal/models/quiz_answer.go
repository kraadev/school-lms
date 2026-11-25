package models

import (
	"time"

	"github.com/google/uuid"
)

type QuizAnswer struct {
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`

	AttemptID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_attempt_question"`

	QuestionID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_attempt_question"`

	OptionID *uuid.UUID `gorm:"type:uuid"`

	AnswerText *string `gorm:"type:text"`

	IsCorrect *bool

	Score *float64

	Feedback *string `gorm:"type:text"`

	Attempt  QuizAttempt         `gorm:"foreignKey:AttemptID"`
	Question QuizQuestion        `gorm:"foreignKey:QuestionID"`
	Option   *QuizQuestionOption `gorm:"foreignKey:OptionID"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
