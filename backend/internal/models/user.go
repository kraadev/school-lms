package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	RoleID       uuid.UUID `gorm:"type:uuid;not null" json:"role_id"`
	Username     string    `gorm:"unique;not null" json:"username"`
	Email        *string   `gorm:"unique" json:"email,omitempty"`
	PasswordHash string    `gorm:"not null" json:"-"`
	IsActive     bool      `gorm:"not null;default:true" json:"is_active"`
	DisplayName  string    `gorm:"-" json:"display_name,omitempty"`

	Role    Role     `gorm:"foreignKey:RoleID" json:"role,omitempty"`
	Teacher *Teacher `gorm:"foreignKey:UserID" json:"teacher,omitempty"`
	Student *Student `gorm:"foreignKey:UserID" json:"student,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
