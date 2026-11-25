package meeting

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"school-lms/internal/livekit"
	"school-lms/internal/models"
)

const (
	StatusScheduled      = "scheduled"
	StatusActive         = "active"
	StatusLive           = "live" // Alias for active
	StatusEnded          = "ended"
	StatusCancelled      = "cancelled"
	ParticipantRequested = "requested"
	ParticipantApproved  = "approved"
	ParticipantRejected  = "rejected"
	ParticipantKicked    = "kicked"
)

type Service struct {
	DB      *gorm.DB
	LiveKit *livekit.Service
}

func NewService(db *gorm.DB, liveKitService *livekit.Service) *Service {
	return &Service{
		DB:      db,
		LiveKit: liveKitService,
	}
}

func (s *Service) UserDisplayName(userID uuid.UUID) string {
	var user models.User
	if err := s.DB.Preload("Teacher").Preload("Student").First(&user, "id = ?", userID).Error; err != nil {
		return "Peserta"
	}
	if user.Teacher != nil && user.Teacher.FullName != "" {
		return user.Teacher.FullName
	}
	if user.Student != nil && user.Student.FullName != "" {
		return user.Student.FullName
	}
	return user.Username
}

type CreateRequest struct {
	Title            string      `json:"title" binding:"required"`
	ClassSubjectID   uuid.UUID   `json:"class_subject_id" binding:"required"`
	Description      string      `json:"description"`
	Type             string      `json:"type"`
	ApprovalRequired *bool       `json:"approval_required"`
	MaxParticipants  int         `json:"max_participants"`
	ScheduledAt      *time.Time  `json:"scheduled_at"`
	ParticipantIDs   []uuid.UUID `json:"participant_ids"`
}

type UpdateRequest struct {
	Title            *string    `json:"title"`
	Description      *string    `json:"description"`
	MaxParticipants  *int       `json:"max_participants"`
	ApprovalRequired *bool      `json:"approval_required"`
	ScheduledAt      *time.Time `json:"scheduled_at"`
}

type JoinResponse struct {
	Token    string `json:"token"`
	URL      string `json:"url"`
	RoomName string `json:"room_name"`
	Identity string `json:"identity"`
}

func (s *Service) teacherClassSubject(userID, classSubjectID uuid.UUID) (*models.ClassSubject, error) {
	var subject models.ClassSubject
	if err := s.DB.Preload("Class").Preload("Subject").
		Joins("JOIN teachers ON teachers.id = class_subjects.teacher_id").
		Where("class_subjects.id = ? AND teachers.user_id = ?", classSubjectID, userID).
		First(&subject).Error; err != nil {
		return nil, errors.New("kelas dan mata pelajaran tidak ditugaskan kepada guru ini")
	}
	return &subject, nil
}

func (s *Service) studentHasAccess(userID uuid.UUID, classSubjectID *uuid.UUID) bool {
	if classSubjectID == nil {
		return false
	}
	var count int64
	s.DB.Table("class_students").
		Joins("JOIN students ON students.id = class_students.student_id").
		Joins("JOIN class_subjects ON class_subjects.class_id = class_students.class_id").
		Where("students.user_id = ? AND class_subjects.id = ?", userID, classSubjectID).
		Count(&count)
	return count > 0
}

func (s *Service) LogActivity(meetingID, userID uuid.UUID, action, details string) {
	_ = s.DB.Create(&models.MeetingActivityLog{
		MeetingID: meetingID,
		UserID:    userID,
		Action:    action,
		Details:   details,
		CreatedAt: time.Now(),
	})
}

func (s *Service) Create(userID uuid.UUID, req CreateRequest) (*models.Meeting, error) {
	if req.Type == "" {
		req.Type = "meeting"
	}
	if req.Type != "voice" && req.Type != "video" && req.Type != "meeting" {
		return nil, errors.New("tipe meeting tidak valid")
	}
	max := req.MaxParticipants
	if max <= 0 {
		max = 50
	}
	if max > 500 {
		return nil, errors.New("maksimal peserta adalah 500")
	}
	approval := true
	if req.ApprovalRequired != nil {
		approval = *req.ApprovalRequired
	}

	if _, err := s.teacherClassSubject(userID, req.ClassSubjectID); err != nil {
		return nil, err
	}

	meetingID := uuid.New()
	roomName := fmt.Sprintf("meeting-%s", meetingID.String())

	classSubjectID := req.ClassSubjectID
	meeting := &models.Meeting{
		ID:               meetingID,
		HostID:           userID,
		ClassSubjectID:   &classSubjectID,
		Title:            strings.TrimSpace(req.Title),
		Description:      req.Description,
		Type:             req.Type,
		RoomName:         roomName,
		Status:           StatusScheduled,
		ApprovalRequired: approval,
		MaxParticipants:  max,
		ScheduledAt:      req.ScheduledAt,
	}
	if meeting.Title == "" {
		return nil, errors.New("judul meeting wajib diisi")
	}

	if err := s.DB.Create(meeting).Error; err != nil {
		return nil, err
	}

	// Add host as approved participant with full permissions
	if err := s.DB.Create(&models.MeetingParticipant{
		MeetingID:          meeting.ID,
		UserID:             userID,
		Status:             ParticipantApproved,
		MicEnabled:         true,
		CameraEnabled:      true,
		ScreenShareEnabled: true,
	}).Error; err != nil {
		return nil, err
	}

	for _, participantID := range req.ParticipantIDs {
		if participantID == userID {
			continue
		}
		s.DB.FirstOrCreate(&models.MeetingParticipant{
			MeetingID: meeting.ID,
			UserID:    participantID,
			Status:    ParticipantRequested,
		})
	}

	s.LogActivity(meeting.ID, userID, "meeting_created", fmt.Sprintf("Meeting '%s' dibuat", meeting.Title))
	return s.Get(userID, meeting.ID, "guru", true)
}

func (s *Service) Get(userID, meetingID uuid.UUID, role string, includePrivate bool) (*models.Meeting, error) {
	var meeting models.Meeting
	if err := s.DB.Preload("Host.Role").
		Preload("ClassSubject.Class").
		Preload("ClassSubject.Subject").
		Preload("Participants.User.Role").
		Preload("Participants.User.Student").
		Preload("Participants.User.Teacher").
		First(&meeting, "id = ?", meetingID).Error; err != nil {
		return nil, err
	}

	// Format room name if empty
	if meeting.RoomName == "" {
		meeting.RoomName = fmt.Sprintf("meeting-%s", meeting.ID.String())
	}

	if includePrivate && role != "admin" {
		if meeting.HostID != userID {
			if !s.studentHasAccess(userID, meeting.ClassSubjectID) {
				return nil, errors.New("meeting ini bukan untuk kelas siswa")
			}
		}
	}
	return &meeting, nil
}

func (s *Service) List(userID uuid.UUID, role string) ([]models.Meeting, error) {
	query := s.DB.Preload("Host.Role").
		Preload("ClassSubject.Class").
		Preload("ClassSubject.Subject").
		Preload("Participants.User.Role").
		Preload("Participants.User.Student").
		Preload("Participants.User.Teacher")

	var meetings []models.Meeting
	if role == "guru" {
		query = query.Where("host_id = ?", userID)
	} else if role == "siswa" {
		query = query.Joins("JOIN class_subjects cs ON cs.id = meetings.class_subject_id").
			Joins("JOIN class_students cst ON cst.class_id = cs.class_id").
			Joins("JOIN students st ON st.id = cst.student_id").
			Where("st.user_id = ?", userID)
	} else {
		// Admin
		query = query.Order("created_at DESC")
	}

	err := query.Order("created_at DESC").Find(&meetings).Error
	if err == nil {
		for i := range meetings {
			if meetings[i].RoomName == "" {
				meetings[i].RoomName = fmt.Sprintf("meeting-%s", meetings[i].ID.String())
			}
		}
	}
	return meetings, err
}

func (s *Service) Update(userID, meetingID uuid.UUID, req UpdateRequest) (*models.Meeting, error) {
	var meeting models.Meeting
	if err := s.DB.First(&meeting, "id = ? AND host_id = ?", meetingID, userID).Error; err != nil {
		return nil, errors.New("meeting bukan milik guru ini")
	}
	if meeting.Status == StatusEnded {
		return nil, errors.New("meeting sudah berakhir")
	}
	updates := map[string]interface{}{}
	if req.Title != nil {
		updates["title"] = strings.TrimSpace(*req.Title)
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.MaxParticipants != nil {
		if *req.MaxParticipants < 1 || *req.MaxParticipants > 500 {
			return nil, errors.New("batas peserta harus 1 sampai 500")
		}
		updates["max_participants"] = *req.MaxParticipants
	}
	if req.ApprovalRequired != nil {
		updates["approval_required"] = *req.ApprovalRequired
	}
	if req.ScheduledAt != nil {
		updates["scheduled_at"] = req.ScheduledAt
	}
	if err := s.DB.Model(&meeting).Updates(updates).Error; err != nil {
		return nil, err
	}

	s.LogActivity(meetingID, userID, "meeting_updated", "Pengaturan meeting diperbarui")
	return s.Get(userID, meetingID, "guru", false)
}

func (s *Service) Delete(userID, meetingID uuid.UUID, role string) error {
	var meeting models.Meeting
	if role == "admin" {
		if err := s.DB.First(&meeting, "id = ?", meetingID).Error; err != nil {
			return err
		}
	} else {
		if err := s.DB.First(&meeting, "id = ? AND host_id = ?", meetingID, userID).Error; err != nil {
			return errors.New("meeting bukan milik guru ini")
		}
	}

	roomName := meeting.RoomName
	if roomName == "" {
		roomName = fmt.Sprintf("meeting-%s", meeting.ID.String())
	}

	if s.LiveKit != nil {
		_ = s.LiveKit.DeleteRoom(context.Background(), roomName)
	}

	s.LogActivity(meetingID, userID, "meeting_deleted", "Meeting dihapus")
	return s.DB.Delete(&meeting).Error
}

func (s *Service) Start(userID, meetingID uuid.UUID) (*models.Meeting, error) {
	return s.changeStatus(userID, meetingID, StatusActive)
}

func (s *Service) End(userID, meetingID uuid.UUID) (*models.Meeting, error) {
	m, err := s.changeStatus(userID, meetingID, StatusEnded)
	if err != nil {
		return nil, err
	}

	roomName := m.RoomName
	if roomName == "" {
		roomName = fmt.Sprintf("meeting-%s", m.ID.String())
	}

	if s.LiveKit != nil {
		_ = s.LiveKit.DeleteRoom(context.Background(), roomName)
	}

	s.LogActivity(meetingID, userID, "meeting_ended", "Meeting diakhiri oleh guru")
	return m, nil
}

func (s *Service) changeStatus(userID, meetingID uuid.UUID, status string) (*models.Meeting, error) {
	var m models.Meeting
	if err := s.DB.First(&m, "id = ? AND host_id = ?", meetingID, userID).Error; err != nil {
		return nil, errors.New("meeting bukan milik guru ini")
	}
	if (status == StatusActive || status == StatusLive) && m.Status == StatusEnded {
		return nil, errors.New("meeting sudah berakhir")
	}
	now := time.Now()
	updates := map[string]interface{}{"status": status}
	if status == StatusActive || status == StatusLive {
		updates["started_at"] = &now
		s.LogActivity(meetingID, userID, "meeting_started", "Meeting dimulai oleh guru")
	}
	if status == StatusEnded {
		updates["ended_at"] = &now
	}
	if err := s.DB.Model(&m).Updates(updates).Error; err != nil {
		return nil, err
	}
	return s.Get(userID, meetingID, "guru", false)
}

func (s *Service) Request(userID, meetingID uuid.UUID) error {
	var m models.Meeting
	if err := s.DB.First(&m, "id = ?", meetingID).Error; err != nil {
		return err
	}
	if m.HostID == userID {
		return errors.New("guru tidak perlu mengajukan akses")
	}
	if !s.studentHasAccess(userID, m.ClassSubjectID) {
		return errors.New("meeting ini bukan untuk kelas siswa")
	}
	status := ParticipantApproved
	if m.ApprovalRequired {
		status = ParticipantRequested
	}
	var p models.MeetingParticipant
	err := s.DB.Where("meeting_id = ? AND user_id = ?", meetingID, userID).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = s.DB.Create(&models.MeetingParticipant{
			MeetingID:          meetingID,
			UserID:             userID,
			Status:             status,
			MicEnabled:         true,
			CameraEnabled:      true,
			ScreenShareEnabled: false,
		}).Error
		if err == nil {
			s.LogActivity(meetingID, userID, "access_requested", "Siswa mengajukan akses masuk meeting")
		}
		return err
	}
	if err != nil {
		return err
	}
	s.LogActivity(meetingID, userID, "access_requested", "Siswa mengajukan ulang akses masuk meeting")
	return s.DB.Model(&p).Updates(map[string]interface{}{"status": status, "left_at": nil}).Error
}

func (s *Service) Requests(hostID, meetingID uuid.UUID) ([]models.MeetingParticipant, error) {
	var m models.Meeting
	if err := s.DB.First(&m, "id = ? AND host_id = ?", meetingID, hostID).Error; err != nil {
		return nil, errors.New("meeting bukan milik guru ini")
	}
	var rows []models.MeetingParticipant
	err := s.DB.Preload("User").
		Preload("User.Role").
		Preload("User.Student").
		Preload("User.Student.Classes").
		Preload("User.Teacher").
		Where("meeting_id = ? AND status = ?", meetingID, ParticipantRequested).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	for i := range rows {
		if rows[i].User.Student != nil && rows[i].User.Student.FullName != "" {
			rows[i].User.DisplayName = rows[i].User.Student.FullName
		} else if rows[i].User.Teacher != nil && rows[i].User.Teacher.FullName != "" {
			rows[i].User.DisplayName = rows[i].User.Teacher.FullName
		} else {
			rows[i].User.DisplayName = rows[i].User.Username
		}
	}
	return rows, nil
}

func (s *Service) Decide(hostID, meetingID, targetUserID uuid.UUID, status string) error {
	if status != ParticipantApproved && status != ParticipantRejected {
		return errors.New("keputusan tidak valid")
	}
	var m models.Meeting
	if err := s.DB.First(&m, "id = ? AND host_id = ?", meetingID, hostID).Error; err != nil {
		return errors.New("meeting bukan milik guru ini")
	}
	err := s.DB.Model(&models.MeetingParticipant{}).
		Where("meeting_id = ? AND user_id = ?", meetingID, targetUserID).
		Update("status", status).Error
	if err == nil {
		s.LogActivity(meetingID, hostID, "participant_decided", fmt.Sprintf("Pengajuan peserta %s diubah menjadi %s", targetUserID, status))
	}
	return err
}

func (s *Service) DecideAll(hostID, meetingID uuid.UUID, status string) error {
	if status != ParticipantApproved && status != ParticipantRejected {
		return errors.New("keputusan tidak valid")
	}
	var m models.Meeting
	if err := s.DB.First(&m, "id = ? AND host_id = ?", meetingID, hostID).Error; err != nil {
		return errors.New("meeting bukan milik guru ini")
	}
	err := s.DB.Model(&models.MeetingParticipant{}).
		Where("meeting_id = ? AND status = ?", meetingID, ParticipantRequested).
		Update("status", status).Error
	if err == nil {
		s.LogActivity(meetingID, hostID, "all_participants_decided", fmt.Sprintf("Semua pengajuan peserta diubah menjadi %s", status))
	}
	return err
}

func (s *Service) SetPermission(hostID, meetingID, targetUserID uuid.UUID, mic, camera, screenShare *bool) error {
	var m models.Meeting
	if err := s.DB.First(&m, "id = ? AND host_id = ?", meetingID, hostID).Error; err != nil {
		return errors.New("meeting bukan milik guru ini")
	}

	var p models.MeetingParticipant
	if err := s.DB.Where("meeting_id = ? AND user_id = ?", meetingID, targetUserID).First(&p).Error; err != nil {
		return errors.New("peserta tidak ditemukan di meeting ini")
	}

	updates := map[string]interface{}{}
	if mic != nil {
		updates["mic_enabled"] = *mic
		p.MicEnabled = *mic
	}
	if camera != nil {
		updates["camera_enabled"] = *camera
		p.CameraEnabled = *camera
	}
	if screenShare != nil {
		updates["screen_share_enabled"] = *screenShare
		p.ScreenShareEnabled = *screenShare
	}

	if err := s.DB.Model(&p).Updates(updates).Error; err != nil {
		return err
	}

	// Update live LiveKit participant permissions
	if s.LiveKit != nil {
		roomName := m.RoomName
		if roomName == "" {
			roomName = fmt.Sprintf("meeting-%s", m.ID.String())
		}
		sources := make([]string, 0)
		if p.MicEnabled {
			sources = append(sources, "microphone")
		}
		if p.CameraEnabled {
			sources = append(sources, "camera")
		}
		if p.ScreenShareEnabled {
			sources = append(sources, "screen_share", "screen_share_audio")
		}
		canPublish := p.MicEnabled || p.CameraEnabled || p.ScreenShareEnabled
		_ = s.LiveKit.UpdateParticipantPermissions(
			context.Background(),
			roomName,
			targetUserID.String(),
			canPublish,
			true,
			true,
			sources,
		)
	}

	s.LogActivity(meetingID, hostID, "permission_updated", fmt.Sprintf("Izin peserta %s diperbarui: mic=%v, cam=%v, screen=%v", targetUserID, p.MicEnabled, p.CameraEnabled, p.ScreenShareEnabled))
	return nil
}

func (s *Service) KickParticipant(hostID, meetingID, targetUserID uuid.UUID) error {
	var m models.Meeting
	if err := s.DB.First(&m, "id = ? AND host_id = ?", meetingID, hostID).Error; err != nil {
		return errors.New("meeting bukan milik guru ini")
	}

	now := time.Now()
	err := s.DB.Model(&models.MeetingParticipant{}).
		Where("meeting_id = ? AND user_id = ?", meetingID, targetUserID).
		Updates(map[string]interface{}{
			"status":  ParticipantKicked,
			"left_at": &now,
		}).Error
	if err != nil {
		return err
	}

	if s.LiveKit != nil {
		roomName := m.RoomName
		if roomName == "" {
			roomName = fmt.Sprintf("meeting-%s", m.ID.String())
		}
		_ = s.LiveKit.RemoveParticipant(context.Background(), roomName, targetUserID.String())
	}

	s.LogActivity(meetingID, hostID, "participant_kicked", fmt.Sprintf("Peserta %s dikeluarkan dari meeting", targetUserID))
	return nil
}

func (s *Service) MuteParticipant(hostID, meetingID, targetUserID uuid.UUID, trackSid string) error {
	var m models.Meeting
	if err := s.DB.First(&m, "id = ? AND host_id = ?", meetingID, hostID).Error; err != nil {
		return errors.New("meeting bukan milik guru ini")
	}

	if s.LiveKit != nil {
		roomName := m.RoomName
		if roomName == "" {
			roomName = fmt.Sprintf("meeting-%s", m.ID.String())
		}
		_ = s.LiveKit.MuteParticipantTrack(context.Background(), roomName, targetUserID.String(), trackSid, true)
	}

	s.LogActivity(meetingID, hostID, "participant_muted", fmt.Sprintf("Peserta %s dimute oleh guru", targetUserID))
	return nil
}

func (s *Service) Join(userID uuid.UUID, meetingID uuid.UUID, role string) (*JoinResponse, error) {
	var m models.Meeting
	if err := s.DB.First(&m, "id = ?", meetingID).Error; err != nil {
		return nil, errors.New("meeting tidak ditemukan")
	}

	isHost := m.HostID == userID
	var p models.MeetingParticipant

	if isHost {
		p = models.MeetingParticipant{
			MeetingID:          m.ID,
			UserID:             userID,
			Status:             ParticipantApproved,
			MicEnabled:         true,
			CameraEnabled:      true,
			ScreenShareEnabled: true,
		}
	} else {
		// Verify student role & class access
		if !s.studentHasAccess(userID, m.ClassSubjectID) {
			return nil, errors.New("meeting ini bukan untuk kelas siswa")
		}

		err := s.DB.Where("meeting_id = ? AND user_id = ?", m.ID, userID).First(&p).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if m.ApprovalRequired {
				return nil, errors.New("akses meeting belum diajukan atau disetujui oleh guru")
			}
			p = models.MeetingParticipant{
				MeetingID:          m.ID,
				UserID:             userID,
				Status:             ParticipantApproved,
				MicEnabled:         true,
				CameraEnabled:      true,
				ScreenShareEnabled: false,
			}
			if err := s.DB.Create(&p).Error; err != nil {
				return nil, err
			}
		} else if err != nil {
			return nil, err
		}

		if p.Status == ParticipantKicked {
			return nil, errors.New("anda telah dikeluarkan dari meeting ini")
		}
		if p.Status == ParticipantRejected {
			return nil, errors.New("pengajuan akses meeting Anda ditolak oleh guru")
		}
		if m.ApprovalRequired && p.Status != ParticipantApproved {
			return nil, errors.New("akses meeting sedang menunggu persetujuan guru")
		}
		if m.Status != StatusActive && m.Status != StatusLive {
			return nil, errors.New("meeting belum dimulai oleh guru")
		}
	}

	// Verify participant limit
	var activeCount int64
	s.DB.Model(&models.MeetingParticipant{}).
		Where("meeting_id = ? AND joined_at IS NOT NULL AND left_at IS NULL AND status = ?", meetingID, ParticipantApproved).
		Count(&activeCount)
	if !isHost && activeCount >= int64(m.MaxParticipants) {
		return nil, errors.New("meeting telah mencapai batas maksimal peserta")
	}

	roomName := m.RoomName
	if roomName == "" {
		roomName = fmt.Sprintf("meeting-%s", m.ID.String())
	}

	displayName := s.UserDisplayName(userID)

	canPublishAudio := p.MicEnabled || isHost
	canPublishVideo := (p.CameraEnabled && m.Type != "voice") || isHost
	canScreenShare := p.ScreenShareEnabled || isHost

	tokenOpts := livekit.TokenOptions{
		RoomName:        roomName,
		Identity:        userID.String(),
		DisplayName:     displayName,
		Role:            role,
		CanPublishAudio: canPublishAudio,
		CanPublishVideo: canPublishVideo,
		CanPublishData:  true,
		CanScreenShare:  canScreenShare,
		IsRoomAdmin:     isHost,
		ValidFor:        6 * time.Hour,
	}

	if s.LiveKit == nil {
		return nil, errors.New("livekit service tidak terkonfigurasi")
	}

	token, err := s.LiveKit.GenerateToken(tokenOpts)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat token livekit: %w", err)
	}

	now := time.Now()
	_ = s.DB.Model(&models.MeetingParticipant{}).
		Where("meeting_id = ? AND user_id = ?", meetingID, userID).
		Updates(map[string]interface{}{
			"joined_at": &now,
			"left_at":   nil,
		})

	s.LogActivity(meetingID, userID, "participant_joined", fmt.Sprintf("%s bergabung ke meeting", displayName))

	return &JoinResponse{
		Token:    token,
		URL:      s.LiveKit.GetURL(),
		RoomName: roomName,
		Identity: userID.String(),
	}, nil
}

func (s *Service) Leave(userID, meetingID uuid.UUID) error {
	now := time.Now()
	displayName := s.UserDisplayName(userID)
	s.LogActivity(meetingID, userID, "participant_left", fmt.Sprintf("%s meninggalkan meeting", displayName))
	return s.DB.Model(&models.MeetingParticipant{}).
		Where("meeting_id = ? AND user_id = ?", meetingID, userID).
		Updates(map[string]interface{}{"left_at": &now}).Error
}
