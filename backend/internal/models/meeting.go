package models

import (
	"time"

	"github.com/google/uuid"
)

type Meeting struct {
	ID               uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	HostID           uuid.UUID  `gorm:"type:uuid;not null;index" json:"host_id"`
	ClassSubjectID   *uuid.UUID `gorm:"type:uuid;index" json:"class_subject_id,omitempty"`
	Title            string     `gorm:"not null" json:"title"`
	Description      string     `json:"description"`
	Type             string     `gorm:"not null;default:'meeting'" json:"type"`
	RoomName         string     `gorm:"not null;default:''" json:"room_name"`
	Status           string     `gorm:"not null;default:'scheduled';index" json:"status"`
	ApprovalRequired bool       `gorm:"not null;default:true" json:"approval_required"`
	MaxParticipants  int        `gorm:"not null;default:50" json:"max_participants"`
	ScheduledAt      *time.Time `json:"scheduled_at,omitempty"`
	StartedAt        *time.Time `json:"started_at,omitempty"`
	EndedAt          *time.Time `json:"ended_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`

	Host         User                 `gorm:"foreignKey:HostID" json:"host,omitempty"`
	ClassSubject ClassSubject         `gorm:"foreignKey:ClassSubjectID" json:"class_subject,omitempty"`
	Participants []MeetingParticipant `gorm:"foreignKey:MeetingID" json:"participants,omitempty"`
}

type MeetingParticipant struct {
	ID                 uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	MeetingID          uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:uq_meeting_participant" json:"meeting_id"`
	UserID             uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:uq_meeting_participant" json:"user_id"`
	Status             string     `gorm:"not null;default:'requested';index" json:"status"`
	MicEnabled         bool       `gorm:"not null;default:true" json:"mic_enabled"`
	CameraEnabled      bool       `gorm:"not null;default:true" json:"camera_enabled"`
	ScreenShareEnabled bool       `gorm:"not null;default:false" json:"screen_share_enabled"`
	JoinedAt           *time.Time `json:"joined_at,omitempty"`
	LeftAt             *time.Time `json:"left_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

type MeetingActivityLog struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	MeetingID uuid.UUID `gorm:"type:uuid;not null;index" json:"meeting_id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Action    string    `gorm:"not null" json:"action"`
	Details   string    `json:"details,omitempty"`
	CreatedAt time.Time `json:"created_at"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
