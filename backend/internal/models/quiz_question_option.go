package models

import (
	"time"

	"github.com/google/uuid"
)

type QuizQuestionOption struct {
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`

	QuestionID uuid.UUID `gorm:"type:uuid;not null" json:"question_id"`

	Text string `gorm:"type:text;not null" json:"text"`

	IsCorrect bool `gorm:"not null;default:false" json:"is_correct"`

	Order int `gorm:"not null" json:"order"`

	Question QuizQuestion `gorm:"foreignKey:QuestionID" json:"-"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
