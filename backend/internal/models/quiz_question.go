package models

import (
	"time"

	"github.com/google/uuid"
)

type QuizQuestion struct {
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`

	QuizID uuid.UUID `gorm:"type:uuid;not null" json:"quiz_id"`

	Question string `gorm:"type:text;not null" json:"question"`

	Type string `gorm:"not null" json:"type"`

	Points float64 `gorm:"not null;default:1" json:"points"`

	Order int `gorm:"not null" json:"order"`

	Quiz Quiz `gorm:"foreignKey:QuizID" json:"-"`

	Options []QuizQuestionOption `gorm:"foreignKey:QuestionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"options"`

	Answers []QuizAttemptAnswer `gorm:"foreignKey:QuestionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"answers,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
