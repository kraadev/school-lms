package models

import (
	"time"

	"github.com/google/uuid"
)

type MessageRead struct {
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`

	MessageID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_message_user_read"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_message_user_read"`

	Message Message `gorm:"foreignKey:MessageID"`
	User    User    `gorm:"foreignKey:UserID"`

	ReadAt time.Time `gorm:"not null"`

	CreatedAt time.Time
}
