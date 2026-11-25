package models

import (
	"time"

	"github.com/google/uuid"
)

type ConversationParticipant struct {
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`

	ConversationID uuid.UUID `gorm:"type:uuid;not null"`
	UserID         uuid.UUID `gorm:"type:uuid;not null"`

	JoinedAt time.Time `gorm:"not null"`

	Conversation Conversation `gorm:"foreignKey:ConversationID"`
	User         User         `gorm:"foreignKey:UserID"`
}
