package models

import (
	"time"

	"github.com/google/uuid"
)

type Conversation struct {
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`

	Type            string `gorm:"type:varchar(20);not null;default:'private'"`
	Name            *string
	Status          string     `gorm:"type:varchar(20);not null;default:'active';index" json:"status"`
	OwnerID         *uuid.UUID `gorm:"type:uuid;index"`
	ClassID         *uuid.UUID `gorm:"type:uuid;index"`
	AssignedAdminID *uuid.UUID `gorm:"type:uuid;index" json:"assigned_admin_id,omitempty"`
	ClosedAt        *time.Time `json:"closed_at,omitempty"`

	Participants  []ConversationParticipant `gorm:"foreignKey:ConversationID"`
	Messages      []Message                 `gorm:"foreignKey:ConversationID"`
	Owner         *User                     `gorm:"foreignKey:OwnerID"`
	Class         *Class                    `gorm:"foreignKey:ClassID"`
	AssignedAdmin *User                     `gorm:"foreignKey:AssignedAdminID" json:"assigned_admin,omitempty"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
