package models

import (
	"time"

	"github.com/google/uuid"
)

type FAQ struct {
	ID        uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Question  string     `gorm:"not null" json:"question"`
	Answer    string     `gorm:"type:text;not null" json:"answer"`
	Category  string     `gorm:"type:varchar(80);not null;default:'Umum'" json:"category"`
	Published bool       `gorm:"not null;default:false;index" json:"published"`
	CreatedBy *uuid.UUID `gorm:"type:uuid;index" json:"created_by,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	Creator   *User      `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}
