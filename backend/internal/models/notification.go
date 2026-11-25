package models

import (
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`

	UserID uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`

	Type    string `gorm:"type:varchar(50);not null" json:"type"`
	Title   string `gorm:"type:varchar(255);not null" json:"title"`
	Message string `gorm:"not null" json:"message"`

	ReferenceType *string    `json:"reference_type"`
	ReferenceID   *uuid.UUID `gorm:"type:uuid" json:"reference_id"`

	IsRead bool       `gorm:"not null;default:false" json:"is_read"`
	ReadAt *time.Time `json:"read_at,omitempty"`

	User User `gorm:"foreignKey:UserID" json:"-"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
