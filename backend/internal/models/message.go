package models

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`

	ConversationID uuid.UUID `gorm:"type:uuid;not null"`
	SenderID       uuid.UUID `gorm:"type:uuid;not null"`

	MessageType string `gorm:"type:varchar(20);not null;default:'text'"`
	Content     string `gorm:"not null"`

	Conversation Conversation `gorm:"foreignKey:ConversationID"`
	Sender       User         `gorm:"foreignKey:SenderID"`

	Reads []MessageRead `gorm:"foreignKey:MessageID"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
