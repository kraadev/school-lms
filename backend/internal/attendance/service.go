package attendance

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"school-lms/internal/models"
	"school-lms/internal/notification"
)

type Service struct {
	DB                  *gorm.DB
	NotificationService *notification.Service
}

func NewService(db *gorm.DB) *Service {
	return &Service{
		DB: db,
	}
}

type CreateAttendanceSessionRequest struct {
	ClassSubjectID string  `json:"class_subject_id" binding:"required"`
	Title          string  `json:"title" binding:"required"`
	Description    *string `json:"description"`
	AttendanceDate string  `json:"attendance_date" binding:"required"`
	StartTime      string  `json:"start_time" binding:"required"`
	EndTime        string  `json:"end_time" binding:"required"`
}

type AttendanceSessionResponse struct {
	ID             uuid.UUID `json:"id"`
	ClassSubjectID uuid.UUID `json:"class_subject_id"`
	TeacherID      uuid.UUID `json:"teacher_id"`
	Title          string    `json:"title"`
	Description    *string   `json:"description"`
	AttendanceDate string    `json:"attendance_date"`
	StartTime      string    `json:"start_time"`
	EndTime        string    `json:"end_time"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type StudentAttendanceSessionResponse struct {
	ID               uuid.UUID `json:"id"`
	ClassSubjectID   uuid.UUID `json:"class_subject_id"`
	TeacherID        uuid.UUID `json:"teacher_id"`
	Title            string    `json:"title"`
	Description      *string   `json:"description"`
	AttendanceDate   string    `json:"attendance_date"`
	StartTime        string    `json:"start_time"`
	EndTime          string    `json:"end_time"`
	Status           string    `json:"status"`
	CheckedIn        bool      `json:"checked_in"`
	AttendanceStatus *string   `json:"attendance_status"`
}

type CheckInResponse struct {
	AttendanceRecordID  uuid.UUID `json:"attendance_record_id"`
	AttendanceSessionID uuid.UUID `json:"attendance_session_id"`
	StudentID           uuid.UUID `json:"student_id"`
	Status              string    `json:"status"`
	CheckedAt           time.Time `json:"checked_at"`
}

type AttendanceRecordResponse struct {
	ID          uuid.UUID  `json:"id"`
	StudentID   uuid.UUID  `json:"student_id"`
	NIS         string     `json:"nis"`
	StudentName string     `json:"student_name"`
	Status      string     `json:"status"`
	CheckedAt   *time.Time `json:"checked_at"`
	Note        *string    `json:"note"`
}

type UpdateAttendanceRecordRequest struct {
	SessionID *string `json:"session_id"`
	StudentID *string `json:"student_id"`
	Status    string  `json:"status" binding:"required"`
	Note      *string `json:"note"`
}

type StudentAttendanceStatsResponse struct {
	TotalSessions        int     `json:"total_sessions"`
	AttendedSessions     int     `json:"attended_sessions"`
	ExcusedSessions      int     `json:"excused_sessions"`
	AbsentSessions       int     `json:"absent_sessions"`
	AttendancePercentage float64 `json:"attendance_percentage"`
}

type AttendanceRecapStudentResponse struct {
	StudentID   uuid.UUID `json:"student_id"`
	NIS         string    `json:"nis"`
	StudentName string    `json:"student_name"`

	Present int `json:"present"`
	Late    int `json:"late"`
	Excused int `json:"excused"`
	Absent  int `json:"absent"`
	Total   int `json:"total"`
}

type AttendanceRecapResponse struct {
	ClassID       uuid.UUID                        `json:"class_id"`
	ClassName     string                           `json:"class_name"`
	TotalSessions int                              `json:"total_sessions"`
	Students      []AttendanceRecapStudentResponse `json:"students"`
}

type AttendanceStatisticsStudentResponse struct {
	StudentID   uuid.UUID `json:"student_id"`
	NIS         string    `json:"nis"`
	StudentName string    `json:"student_name"`

	Present int `json:"present"`
	Late    int `json:"late"`
	Excused int `json:"excused"`
	Absent  int `json:"absent"`
	Total   int `json:"total"`

	AttendancePercentage float64 `json:"attendance_percentage"`
}

type AttendanceStatisticsResponse struct {
	ClassID   uuid.UUID `json:"class_id"`
	ClassName string    `json:"class_name"`

	TotalStudents int `json:"total_students"`
	TotalSessions int `json:"total_sessions"`

	TotalPresent int `json:"total_present"`
	TotalLate    int `json:"total_late"`
	TotalExcused int `json:"total_excused"`
	TotalAbsent  int `json:"total_absent"`

	AttendancePercentage float64 `json:"attendance_percentage"`

	Students []AttendanceStatisticsStudentResponse `json:"students"`
}

func (s *Service) getTeacherByUserID(userID string) (*models.Teacher, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("user id tidak valid")
	}

	var teacher models.Teacher

	err = s.DB.
		Where("user_id = ?", id).
		First(&teacher).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("profile guru tidak ditemukan")
		}

		return nil, err
	}

	return &teacher, nil
}

func (s *Service) getStudentByUserID(userID string) (*models.Student, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("user id tidak valid")
	}

	var student models.Student

	err = s.DB.
		Where("user_id = ?", id).
		First(&student).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("profile siswa tidak ditemukan")
		}

		return nil, err
	}

	return &student, nil
}

func (s *Service) CreateSession(
	userID string,
	req CreateAttendanceSessionRequest,
) (*AttendanceSessionResponse, error) {

	teacher, err := s.getTeacherByUserID(userID)
	if err != nil {
		return nil, err
	}

	classSubjectID, err := uuid.Parse(req.ClassSubjectID)
	if err != nil {
		return nil, errors.New("class_subject_id tidak valid")
	}

	attendanceDate, err := time.Parse("2006-01-02", req.AttendanceDate)
	if err != nil {
		return nil, errors.New(
			"attendance_date harus menggunakan format YYYY-MM-DD",
		)
	}

	startTime, err := time.ParseInLocation(
		"2006-01-02 15:04:05",
		req.StartTime,
		time.Local,
	)
	if err != nil {
		startTime, err = time.Parse("2006-01-02 15:04:05", req.StartTime)
		if err != nil {
			return nil, errors.New(
				"start_time harus menggunakan format YYYY-MM-DD HH:MM:SS",
			)
		}
	}

	endTime, err := time.ParseInLocation(
		"2006-01-02 15:04:05",
		req.EndTime,
		time.Local,
	)
	if err != nil {
		endTime, err = time.Parse("2006-01-02 15:04:05", req.EndTime)
		if err != nil {
			return nil, errors.New(
				"end_time harus menggunakan format YYYY-MM-DD HH:MM:SS",
			)
		}
	}

	if !endTime.After(startTime) {
		return nil, errors.New("end_time harus lebih besar dari start_time")
	}

	// Pastikan class_subject memang milik guru yang sedang login.
	var classSubject models.ClassSubject

	err = s.DB.
		Where(
			"id = ? AND teacher_id = ?",
			classSubjectID,
			teacher.ID,
		).
		First(&classSubject).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(
				"class subject tidak ditemukan atau bukan milik guru ini",
			)
		}

		return nil, err
	}

	// Cegah attendance session yang waktunya bertabrakan
	// untuk class subject yang sama.
	var overlappingSession models.AttendanceSession

	err = s.DB.
		Where(
			"class_subject_id = ? AND attendance_date = ? AND start_time < ? AND end_time > ?",
			classSubjectID,
			attendanceDate,
			endTime,
			startTime,
		).
		First(&overlappingSession).Error

	if err == nil {
		return nil, errors.New(
			"terdapat attendance session lain yang waktunya bertabrakan",
		)
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	session := models.AttendanceSession{
		ClassSubjectID: classSubjectID,
		TeacherID:      teacher.ID,
		Title:          req.Title,
		Description:    req.Description,
		AttendanceDate: attendanceDate,
		StartTime:      startTime,
		EndTime:        endTime,
		Status:         "open",
	}

	if err := s.DB.Create(&session).Error; err != nil {
		return nil, err
	}

	if s.NotificationService != nil {
		refType := "attendance"
		_ = s.NotificationService.NotifyClass(classSubject.ClassID, "attendance", "Presensi Baru: "+session.Title, "Sesi presensi baru telah dibuka untuk tanggal "+req.AttendanceDate, &refType, &session.ID)
	}

	return s.toResponse(&session), nil
}

func (s *Service) GetTeacherSessions(
	userID string,
) ([]AttendanceSessionResponse, error) {

	teacher, err := s.getTeacherByUserID(userID)
	if err != nil {
		return nil, err
	}

	var sessions []models.AttendanceSession

	err = s.DB.
		Where("teacher_id = ?", teacher.ID).
		Preload("ClassSubject.Class").
		Preload("ClassSubject.Subject").
		Order("attendance_date DESC, start_time DESC").
		Find(&sessions).Error

	if err != nil {
		return nil, err
	}

	result := make([]AttendanceSessionResponse, 0, len(sessions))

	for _, session := range sessions {
		result = append(
			result,
			*s.toResponse(&session),
		)
	}

	return result, nil
}

func (s *Service) GetTeacherSession(
	userID string,
	sessionID string,
) (*AttendanceSessionResponse, error) {

	teacher, err := s.getTeacherByUserID(userID)
	if err != nil {
		return nil, err
	}

	id, err := uuid.Parse(sessionID)
	if err != nil {
		return nil, errors.New("attendance session id tidak valid")
	}

	var session models.AttendanceSession

	err = s.DB.
		Where(
			"id = ? AND teacher_id = ?",
			id,
			teacher.ID,
		).
		First(&session).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(
				"attendance session tidak ditemukan",
			)
		}

		return nil, err
	}

	return s.toResponse(&session), nil
}

func (s *Service) toResponse(
	session *models.AttendanceSession,
) *AttendanceSessionResponse {

	return &AttendanceSessionResponse{
		ID:             session.ID,
		ClassSubjectID: session.ClassSubjectID,
		TeacherID:      session.TeacherID,
		Title:          session.Title,
		Description:    session.Description,
		AttendanceDate: session.AttendanceDate.Format("2006-01-02"),
		StartTime:      session.StartTime.Format("2006-01-02 15:04:05"),
		EndTime:        session.EndTime.Format("2006-01-02 15:04:05"),
		Status:         session.Status,
		CreatedAt:      session.CreatedAt,
		UpdatedAt:      session.UpdatedAt,
	}
}

func (s *Service) CloseSession(
	userID string,
	sessionID string,
) (*AttendanceSessionResponse, error) {

	teacher, err := s.getTeacherByUserID(userID)
	if err != nil {
		return nil, err
	}

	id, err := uuid.Parse(sessionID)
	if err != nil {
		return nil, errors.New("attendance session id tidak valid")
	}

	var session models.AttendanceSession

	err = s.DB.
		Where(
			"id = ? AND teacher_id = ?",
			id,
			teacher.ID,
		).
		First(&session).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(
				"attendance session tidak ditemukan",
			)
		}

		return nil, err
	}

	if session.Status == "closed" {
		return nil, errors.New(
			"attendance session sudah ditutup",
		)
	}

	session.Status = "closed"

	if err := s.DB.Save(&session).Error; err != nil {
		return nil, err
	}

	return s.toResponse(&session), nil
}

func (s *Service) OpenSession(
	userID string,
	sessionID string,
) (*AttendanceSessionResponse, error) {

	teacher, err := s.getTeacherByUserID(userID)
	if err != nil {
		return nil, err
	}

	id, err := uuid.Parse(sessionID)
	if err != nil {
		return nil, errors.New("attendance session id tidak valid")
	}

	var session models.AttendanceSession

	err = s.DB.
		Where(
			"id = ? AND teacher_id = ?",
			id,
			teacher.ID,
		).
		First(&session).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(
				"attendance session tidak ditemukan",
			)
		}

		return nil, err
	}

	if session.Status == "open" {
		return nil, errors.New(
			"attendance session sudah terbuka",
		)
	}

	session.Status = "open"

	if err := s.DB.Save(&session).Error; err != nil {
		return nil, err
	}

	return s.toResponse(&session), nil
}

func (s *Service) DeleteSession(userID string, sessionID string) error {
	teacher, err := s.getTeacherByUserID(userID)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(sessionID)
	if err != nil {
		return errors.New("attendance session id tidak valid")
	}
	var session models.AttendanceSession
	if err := s.DB.Where("id = ? AND teacher_id = ?", id, teacher.ID).First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("attendance session tidak ditemukan atau bukan milik guru ini")
		}
		return err
	}
	return s.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("attendance_session_id = ?", session.ID).Delete(&models.AttendanceRecord{}).Error; err != nil {
			return err
		}
		return tx.Delete(&session).Error
	})
}

func (s *Service) GetStudentSessions(
	userID string,
) ([]StudentAttendanceSessionResponse, error) {

	student, err := s.getStudentByUserID(userID)
	if err != nil {
		return nil, err
	}

	// Ambil semua class yang diikuti siswa.
	var classIDs []uuid.UUID

	err = s.DB.
		Table("class_students").
		Where("student_id = ?", student.ID).
		Pluck("class_id", &classIDs).Error

	if err != nil {
		return nil, err
	}

	if len(classIDs) == 0 {
		return []StudentAttendanceSessionResponse{}, nil
	}

	var sessions []models.AttendanceSession

	err = s.DB.
		Joins("JOIN class_subjects ON class_subjects.id = attendance_sessions.class_subject_id").
		Where("class_subjects.class_id IN ?", classIDs).
		Order("attendance_date DESC, start_time DESC").
		Find(&sessions).Error

	if err != nil {
		return nil, err
	}

	result := make(
		[]StudentAttendanceSessionResponse,
		0,
		len(sessions),
	)

	for _, session := range sessions {
		item := StudentAttendanceSessionResponse{
			ID:               session.ID,
			ClassSubjectID:   session.ClassSubjectID,
			TeacherID:        session.TeacherID,
			Title:            session.Title,
			Description:      session.Description,
			AttendanceDate:   session.AttendanceDate.Format("2006-01-02"),
			StartTime:        session.StartTime.Format("2006-01-02 15:04:05"),
			EndTime:          session.EndTime.Format("2006-01-02 15:04:05"),
			Status:           session.Status,
			CheckedIn:        false,
			AttendanceStatus: nil,
		}

		var record models.AttendanceRecord

		err := s.DB.
			Where(
				"attendance_session_id = ? AND student_id = ?",
				session.ID,
				student.ID,
			).
			First(&record).Error

		if err == nil {
			item.CheckedIn = true
			item.AttendanceStatus = &record.Status
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

		result = append(result, item)
	}

	return result, nil
}

func (s *Service) GetStudentSession(
	userID string,
	sessionID string,
) (*StudentAttendanceSessionResponse, error) {

	student, err := s.getStudentByUserID(userID)
	if err != nil {
		return nil, err
	}

	id, err := uuid.Parse(sessionID)
	if err != nil {
		return nil, errors.New("attendance session id tidak valid")
	}

	var session models.AttendanceSession

	err = s.DB.
		Joins("JOIN class_subjects ON class_subjects.id = attendance_sessions.class_subject_id").
		Joins("JOIN class_students ON class_students.class_id = class_subjects.class_id").
		Where(
			"attendance_sessions.id = ? AND class_students.student_id = ?",
			id,
			student.ID,
		).
		First(&session).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(
				"attendance session tidak ditemukan atau bukan untuk kelas siswa ini",
			)
		}

		return nil, err
	}

	result := &StudentAttendanceSessionResponse{
		ID:               session.ID,
		ClassSubjectID:   session.ClassSubjectID,
		TeacherID:        session.TeacherID,
		Title:            session.Title,
		Description:      session.Description,
		AttendanceDate:   session.AttendanceDate.Format("2006-01-02"),
		StartTime:        session.StartTime.Format("2006-01-02 15:04:05"),
		EndTime:          session.EndTime.Format("2006-01-02 15:04:05"),
		Status:           session.Status,
		CheckedIn:        false,
		AttendanceStatus: nil,
	}

	var record models.AttendanceRecord

	err = s.DB.
		Where(
			"attendance_session_id = ? AND student_id = ?",
			session.ID,
			student.ID,
		).
		First(&record).Error

	if err == nil {
		result.CheckedIn = true
		result.AttendanceStatus = &record.Status
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	return result, nil
}

func (s *Service) CheckIn(
	userID string,
	sessionID string,
) (*CheckInResponse, error) {

	student, err := s.getStudentByUserID(userID)
	if err != nil {
		return nil, err
	}

	id, err := uuid.Parse(sessionID)
	if err != nil {
		return nil, errors.New("attendance session id tidak valid")
	}

	// Pastikan session memang untuk kelas siswa.
	var session models.AttendanceSession

	err = s.DB.
		Joins("JOIN class_subjects ON class_subjects.id = attendance_sessions.class_subject_id").
		Joins("JOIN class_students ON class_students.class_id = class_subjects.class_id").
		Where(
			"attendance_sessions.id = ? AND class_students.student_id = ?",
			id,
			student.ID,
		).
		First(&session).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(
				"attendance session tidak ditemukan atau bukan untuk kelas siswa ini",
			)
		}

		return nil, err
	}

	// Session harus dalam keadaan open.
	if session.Status != "open" {
		return nil, errors.New(
			"attendance session sudah ditutup",
		)
	}

	now := time.Now()
	startTime := session.StartTime
	endTime := session.EndTime

	if !startTime.IsZero() {
		localStart := startTime.In(now.Location())
		wallStart := time.Date(startTime.Year(), startTime.Month(), startTime.Day(), startTime.Hour(), startTime.Minute(), startTime.Second(), 0, now.Location())
		if now.Before(localStart) && now.Before(wallStart) {
			return nil, errors.New(
				"attendance belum dimulai",
			)
		}
	}

	if !endTime.IsZero() {
		localEnd := endTime.In(now.Location())
		wallEnd := time.Date(endTime.Year(), endTime.Month(), endTime.Day(), endTime.Hour(), endTime.Minute(), endTime.Second(), 0, now.Location())
		if !now.Before(localEnd) && !now.Before(wallEnd) {
			return nil, errors.New(
				"waktu attendance sudah berakhir",
			)
		}
	}

	// Pastikan siswa belum pernah check-in.
	var existingRecord models.AttendanceRecord

	err = s.DB.
		Where(
			"attendance_session_id = ? AND student_id = ?",
			session.ID,
			student.ID,
		).
		First(&existingRecord).Error

	if err == nil {
		return nil, errors.New(
			"siswa sudah melakukan check-in pada attendance ini",
		)
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Tentukan status.
	status := "present"

	if now.After(session.StartTime) {
		status = "late"
	}

	record := models.AttendanceRecord{
		AttendanceSessionID: session.ID,
		StudentID:           student.ID,
		Status:              status,
		CheckedAt:           &now,
	}

	if err := s.DB.Create(&record).Error; err != nil {
		return nil, err
	}

	if s.NotificationService != nil {
		refType := "attendance"
		var teacher models.Teacher
		if err := s.DB.Where("id = ?", session.TeacherID).First(&teacher).Error; err == nil {
			_ = s.NotificationService.NotifyUser(teacher.UserID, "attendance_checkin", "Siswa Check-in Presensi", student.FullName+" telah melakukan check-in pada sesi "+session.Title, &refType, &session.ID)
		}
	}

	return &CheckInResponse{
		AttendanceRecordID:  record.ID,
		AttendanceSessionID: record.AttendanceSessionID,
		StudentID:           record.StudentID,
		Status:              record.Status,
		CheckedAt:           *record.CheckedAt,
	}, nil
}

func (s *Service) GetTeacherAttendanceRecords(
	userID string,
	sessionID string,
) ([]AttendanceRecordResponse, error) {

	teacher, err := s.getTeacherByUserID(userID)
	if err != nil {
		return nil, err
	}

	id, err := uuid.Parse(sessionID)
	if err != nil {
		return nil, errors.New("attendance session id tidak valid")
	}

	// Pastikan session memang milik guru yang sedang login.
	var session models.AttendanceSession

	err = s.DB.
		Where(
			"id = ? AND teacher_id = ?",
			id,
			teacher.ID,
		).
		First(&session).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(
				"attendance session tidak ditemukan",
			)
		}

		return nil, err
	}

	// Ambil seluruh siswa yang terdaftar di class
	// dari class_subject milik attendance session.
	var students []models.Student

	err = s.DB.
		Joins("JOIN class_students ON class_students.student_id = students.id").
		Joins("JOIN class_subjects ON class_subjects.class_id = class_students.class_id").
		Where(
			"class_subjects.id = ?",
			session.ClassSubjectID,
		).
		Order("students.full_name ASC").
		Find(&students).Error

	if err != nil {
		return nil, err
	}

	result := make(
		[]AttendanceRecordResponse,
		0,
		len(students),
	)

	for _, student := range students {

		var record models.AttendanceRecord

		err := s.DB.
			Where(
				"attendance_session_id = ? AND student_id = ?",
				session.ID,
				student.ID,
			).
			First(&record).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			result = append(result, AttendanceRecordResponse{
				ID:          uuid.Nil,
				StudentID:   student.ID,
				NIS:         student.NIS,
				StudentName: student.FullName,
				Status:      "not_checked_in",
				CheckedAt:   nil,
				Note:        nil,
			})

			continue
		}

		if err != nil {
			return nil, err
		}

		result = append(result, AttendanceRecordResponse{
			ID:          record.ID,
			StudentID:   student.ID,
			NIS:         student.NIS,
			StudentName: student.FullName,
			Status:      record.Status,
			CheckedAt:   record.CheckedAt,
			Note:        record.Note,
		})
	}

	return result, nil
}

func (s *Service) UpdateAttendanceRecord(
	userID string,
	recordID string,
	req UpdateAttendanceRecordRequest,
) (*AttendanceRecordResponse, error) {

	teacher, err := s.getTeacherByUserID(userID)
	if err != nil {
		return nil, err
	}

	id, err := uuid.Parse(recordID)
	if err != nil {
		return nil, errors.New("attendance record id tidak valid")
	}

	// Validasi status.
	switch req.Status {
	case "present", "late", "absent", "excused":
		// valid
	default:
		return nil, errors.New(
			"status harus present, late, absent, atau excused",
		)
	}

	// Ambil record sekaligus memastikan session
	// memang milik guru yang sedang login.
	var record models.AttendanceRecord

	err = s.DB.
		Joins(
			"JOIN attendance_sessions ON attendance_sessions.id = attendance_records.attendance_session_id",
		).
		Where(
			"attendance_records.id = ? AND attendance_sessions.teacher_id = ?",
			id,
			teacher.ID,
		).
		First(&record).Error

	if err != nil {
		if (errors.Is(err, gorm.ErrRecordNotFound) || id == uuid.Nil) && req.SessionID != nil && req.StudentID != nil {
			sessUUID, _ := uuid.Parse(*req.SessionID)
			studUUID, _ := uuid.Parse(*req.StudentID)
			if sessUUID != uuid.Nil && studUUID != uuid.Nil {
				record = models.AttendanceRecord{
					AttendanceSessionID: sessUUID,
					StudentID:           studUUID,
					Status:              req.Status,
					Note:                req.Note,
				}
				if req.Status == "present" || req.Status == "late" {
					now := time.Now()
					record.CheckedAt = &now
				}
				if err := s.DB.Create(&record).Error; err != nil {
					return nil, err
				}
				var student models.Student
				_ = s.DB.Where("id = ?", studUUID).First(&student).Error
				return &AttendanceRecordResponse{
					ID:          record.ID,
					StudentID:   record.StudentID,
					NIS:         student.NIS,
					StudentName: student.FullName,
					Status:      record.Status,
					CheckedAt:   record.CheckedAt,
					Note:        record.Note,
				}, nil
			}
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(
				"attendance record tidak ditemukan atau bukan milik guru ini",
			)
		}

		return nil, err
	}

	record.Status = req.Status
	record.Note = req.Note

	// Jika status menjadi absent/excused,
	// checked_at tidak lagi dianggap sebagai waktu check-in.
	if req.Status == "absent" || req.Status == "excused" {
		record.CheckedAt = nil
	}

	if err := s.DB.Save(&record).Error; err != nil {
		return nil, err
	}

	// Ambil data siswa untuk response.
	var student models.Student

	if err := s.DB.
		Where("id = ?", record.StudentID).
		First(&student).Error; err != nil {
		return nil, err
	}

	return &AttendanceRecordResponse{
		ID:          record.ID,
		StudentID:   record.StudentID,
		NIS:         student.NIS,
		StudentName: student.FullName,
		Status:      record.Status,
		CheckedAt:   record.CheckedAt,
		Note:        record.Note,
	}, nil
}

func (s *Service) GetAttendanceRecap(
	userID string,
	classID string,
) (*AttendanceRecapResponse, error) {

	// 1. Ambil profile guru
	teacher, err := s.getTeacherByUserID(userID)
	if err != nil {
		return nil, err
	}

	// 2. Parse class ID
	classUUID, err := uuid.Parse(classID)
	if err != nil {
		return nil, errors.New("class id tidak valid")
	}

	// 3. Pastikan kelas ada
	var class models.Class

	err = s.DB.
		Where("id = ?", classUUID).
		First(&class).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("kelas tidak ditemukan")
		}

		return nil, err
	}

	// 4. Ambil attendance session milik guru
	//    yang berasal dari class_subject kelas tersebut
	var sessions []models.AttendanceSession

	err = s.DB.
		Joins(
			"JOIN class_subjects ON class_subjects.id = attendance_sessions.class_subject_id",
		).
		Where(
			"class_subjects.class_id = ? AND attendance_sessions.teacher_id = ?",
			classUUID,
			teacher.ID,
		).
		Order("attendance_sessions.attendance_date ASC, attendance_sessions.start_time ASC").
		Find(&sessions).Error

	if err != nil {
		return nil, err
	}

	// 5. Ambil seluruh siswa dari kelas
	var students []models.Student

	err = s.DB.
		Joins(
			"JOIN class_students ON class_students.student_id = students.id",
		).
		Where(
			"class_students.class_id = ?",
			classUUID,
		).
		Order("students.full_name ASC").
		Find(&students).Error

	if err != nil {
		return nil, err
	}

	// 6. Siapkan response
	response := &AttendanceRecapResponse{
		ClassID:       classUUID,
		ClassName:     class.Name,
		TotalSessions: len(sessions),
		Students:      make([]AttendanceRecapStudentResponse, 0, len(students)),
	}

	// Kalau belum ada session, tetap return daftar siswa
	if len(sessions) == 0 {
		for _, student := range students {
			response.Students = append(
				response.Students,
				AttendanceRecapStudentResponse{
					StudentID:   student.ID,
					NIS:         student.NIS,
					StudentName: student.FullName,
					Total:       0,
				},
			)
		}

		return response, nil
	}

	// 7. Ambil ID semua session
	sessionIDs := make([]uuid.UUID, 0, len(sessions))

	for _, session := range sessions {
		sessionIDs = append(sessionIDs, session.ID)
	}

	// 8. Ambil seluruh attendance record sekaligus
	var records []models.AttendanceRecord

	err = s.DB.
		Where("attendance_session_id IN ?", sessionIDs).
		Find(&records).Error

	if err != nil {
		return nil, err
	}

	// 9. Buat map:
	//    session_id + student_id -> attendance record
	recordMap := make(map[string]models.AttendanceRecord)

	for _, record := range records {
		key := record.AttendanceSessionID.String() + "_" + record.StudentID.String()

		recordMap[key] = record
	}

	// 10. Hitung rekap setiap siswa
	for _, student := range students {

		recap := AttendanceRecapStudentResponse{
			StudentID:   student.ID,
			NIS:         student.NIS,
			StudentName: student.FullName,
			Total:       len(sessions),
		}

		for _, session := range sessions {

			key := session.ID.String() + "_" + student.ID.String()

			record, exists := recordMap[key]

			// Tidak memiliki record = absent
			if !exists {
				recap.Absent++
				continue
			}

			switch record.Status {
			case "present":
				recap.Present++

			case "late":
				recap.Late++

			case "excused":
				recap.Excused++

			case "absent":
				recap.Absent++

			default:
				recap.Absent++
			}
		}

		response.Students = append(
			response.Students,
			recap,
		)
	}

	return response, nil
}

func (s *Service) GetAttendanceStatistics(
	userID string,
	classID string,
) (*AttendanceStatisticsResponse, error) {

	// 1. Ambil profile guru
	teacher, err := s.getTeacherByUserID(userID)
	if err != nil {
		return nil, err
	}

	// 2. Parse class ID
	classUUID, err := uuid.Parse(classID)
	if err != nil {
		return nil, errors.New("class id tidak valid")
	}

	// 3. Ambil data kelas
	var class models.Class

	err = s.DB.
		Where("id = ?", classUUID).
		First(&class).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("kelas tidak ditemukan")
		}

		return nil, err
	}

	// 4. Ambil semua attendance session
	//    milik guru untuk kelas tersebut
	var sessions []models.AttendanceSession

	err = s.DB.
		Joins(
			"JOIN class_subjects ON class_subjects.id = attendance_sessions.class_subject_id",
		).
		Where(
			"class_subjects.class_id = ? AND attendance_sessions.teacher_id = ?",
			classUUID,
			teacher.ID,
		).
		Order(
			"attendance_sessions.attendance_date ASC, attendance_sessions.start_time ASC",
		).
		Find(&sessions).Error

	if err != nil {
		return nil, err
	}

	// 5. Ambil semua siswa kelas
	var students []models.Student

	err = s.DB.
		Joins(
			"JOIN class_students ON class_students.student_id = students.id",
		).
		Where(
			"class_students.class_id = ?",
			classUUID,
		).
		Order("students.full_name ASC").
		Find(&students).Error

	if err != nil {
		return nil, err
	}

	// 6. Siapkan response
	response := &AttendanceStatisticsResponse{
		ClassID:       classUUID,
		ClassName:     class.Name,
		TotalStudents: len(students),
		TotalSessions: len(sessions),
		Students: make(
			[]AttendanceStatisticsStudentResponse,
			0,
			len(students),
		),
	}

	// Jika belum ada attendance session
	if len(sessions) == 0 {
		return response, nil
	}

	// 7. Ambil semua ID session
	sessionIDs := make([]uuid.UUID, 0, len(sessions))

	for _, session := range sessions {
		sessionIDs = append(sessionIDs, session.ID)
	}

	// 8. Ambil semua attendance record
	var records []models.AttendanceRecord

	err = s.DB.
		Where(
			"attendance_session_id IN ?",
			sessionIDs,
		).
		Find(&records).Error

	if err != nil {
		return nil, err
	}

	// 9. Buat map attendance record
	recordMap := make(map[string]models.AttendanceRecord)

	for _, record := range records {
		key := record.AttendanceSessionID.String() +
			"_" +
			record.StudentID.String()

		recordMap[key] = record
	}

	// 10. Hitung statistik masing-masing siswa
	for _, student := range students {

		studentStats := AttendanceStatisticsStudentResponse{
			StudentID:   student.ID,
			NIS:         student.NIS,
			StudentName: student.FullName,
			Total:       len(sessions),
		}

		for _, session := range sessions {

			key := session.ID.String() +
				"_" +
				student.ID.String()

			record, exists := recordMap[key]

			// Tidak ada record berarti absent
			if !exists {
				studentStats.Absent++
				continue
			}

			switch record.Status {

			case "present":
				studentStats.Present++

			case "late":
				studentStats.Late++

			case "excused":
				studentStats.Excused++

			case "absent":
				studentStats.Absent++

			default:
				studentStats.Absent++
			}
		}

		// Persentase kehadiran
		//
		// Present + Late dianggap hadir.
		//
		// Contoh:
		// present = 10
		// late = 2
		// total = 15
		//
		// (10 + 2) / 15 * 100
		if studentStats.Total > 0 {
			studentStats.AttendancePercentage =
				float64(studentStats.Present+studentStats.Late) /
					float64(studentStats.Total) *
					100
		}

		// Tambahkan ke total kelas
		response.TotalPresent += studentStats.Present
		response.TotalLate += studentStats.Late
		response.TotalExcused += studentStats.Excused
		response.TotalAbsent += studentStats.Absent

		response.Students = append(
			response.Students,
			studentStats,
		)
	}

	// 11. Persentase kehadiran keseluruhan kelas
	totalAttendance := response.TotalPresent + response.TotalLate
	totalPossibleAttendance := response.TotalStudents * response.TotalSessions

	if totalPossibleAttendance > 0 {
		response.AttendancePercentage =
			float64(totalAttendance) /
				float64(totalPossibleAttendance) *
				100
	}

	return response, nil
}

func (s *Service) GetStudentAttendanceStatistics(userID string) (*StudentAttendanceStatsResponse, error) {
	student, err := s.getStudentByUserID(userID)
	if err != nil {
		return nil, err
	}

	var classIDs []uuid.UUID
	err = s.DB.Table("class_students").Where("student_id = ?", student.ID).Pluck("class_id", &classIDs).Error
	if err != nil {
		return nil, err
	}

	if len(classIDs) == 0 {
		return &StudentAttendanceStatsResponse{}, nil
	}

	var totalSessions int64
	err = s.DB.Model(&models.AttendanceSession{}).
		Joins("JOIN class_subjects ON class_subjects.id = attendance_sessions.class_subject_id").
		Where("class_subjects.class_id IN ?", classIDs).
		Count(&totalSessions).Error
	if err != nil {
		return nil, err
	}

	var records []models.AttendanceRecord
	err = s.DB.Model(&models.AttendanceRecord{}).
		Where("student_id = ?", student.ID).
		Find(&records).Error
	if err != nil {
		return nil, err
	}

	attended := 0
	excused := 0
	for _, r := range records {
		switch r.Status {
		case "present", "late", "hadir":
			attended++
		case "excused", "izin", "sakit":
			excused++
		}
	}

	absent := int(totalSessions) - attended - excused
	if absent < 0 {
		absent = 0
	}

	percentage := 0.0
	if totalSessions > 0 {
		percentage = (float64(attended) / float64(totalSessions)) * 100
	}

	return &StudentAttendanceStatsResponse{
		TotalSessions:        int(totalSessions),
		AttendedSessions:     attended,
		ExcusedSessions:      excused,
		AbsentSessions:       absent,
		AttendancePercentage: percentage,
	}, nil
}
