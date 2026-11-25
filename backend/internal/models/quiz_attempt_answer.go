package models

import (
	"time"

	"github.com/google/uuid"
)

type QuizAttemptAnswer struct {
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`

	AttemptID  uuid.UUID  `gorm:"type:uuid;not null" json:"attempt_id"`
	QuestionID uuid.UUID  `gorm:"type:uuid;not null" json:"question_id"`
	OptionID   *uuid.UUID `gorm:"type:uuid" json:"option_id"`
	// Options for multiple selections
	Options    []QuizQuestionOption `gorm:"many2many:quiz_attempt_answer_options;" json:"options,omitempty"`

	EssayAnswer *string    `json:"essay_answer"`
	AnsweredAt  *time.Time `json:"answered_at"`
	Score       *float64   `json:"score"`
	Feedback    *string    `json:"feedback"`

	Attempt  QuizAttempt         `gorm:"foreignKey:AttemptID" json:"-"`
	Question QuizQuestion        `gorm:"foreignKey:QuestionID" json:"question,omitempty"`
	Option   *QuizQuestionOption `gorm:"foreignKey:OptionID" json:"option,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
