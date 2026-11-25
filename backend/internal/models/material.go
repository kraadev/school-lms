package models

import (
	"time"

	"github.com/google/uuid"
)

type Material struct {
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`

	ClassSubjectID uuid.UUID `gorm:"type:uuid;not null"`
	TeacherID      uuid.UUID `gorm:"type:uuid;not null"`

	Title       string  `gorm:"not null"`
	Description *string `gorm:"type:text"`

	VideoURL    *string
	ExternalURL *string

	Status string `gorm:"not null;default:'draft'"`

	PublishedAt *time.Time

	ClassSubject ClassSubject   `gorm:"foreignKey:ClassSubjectID"`
	Teacher      Teacher        `gorm:"foreignKey:TeacherID"`
	Files        []MaterialFile `gorm:"foreignKey:MaterialID;constraint:OnDelete:CASCADE"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
