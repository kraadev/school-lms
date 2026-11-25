package models

import (
	"time"

	"github.com/google/uuid"
)

type Major struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Code        string    `gorm:"unique;not null"`
	Name        string    `gorm:"not null"`
	Description *string

	Classes []Class `gorm:"foreignKey:MajorID"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
