package models

import (
	"time"

	"github.com/google/uuid"
)

type Admin struct {
	ID       uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID   uuid.UUID `gorm:"type:uuid;unique;not null"`
	FullName string    `gorm:"not null"`

	User User `gorm:"foreignKey:UserID"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
