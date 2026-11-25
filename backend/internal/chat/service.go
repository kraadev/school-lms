package chat

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

func NewService(
	db *gorm.DB,
	notificationService *notification.Service,
) *Service {

	return &Service{
		DB:                  db,
		NotificationService: notificationService,
	}
}

type CreateConversationRequest struct {
	Type           string
	Name           *string
	ClassID        *uuid.UUID
	ParticipantIDs []uuid.UUID
}

type UpdateConversationRequest struct {
	Name *string
}

type SendMessageRequest struct {
	MessageType string
	Content     string
}

type ConversationUnreadCount struct {
	ConversationID uuid.UUID `json:"conversation_id"`
	UnreadCount    int64     `json:"unread_count"`
}

func (s *Service) GetChatContacts(userID, role string, classID *uuid.UUID) ([]models.User, error) {
	currentUserID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("user id tidak valid")
	}
	var users []models.User
	query := s.DB.Preload("Role").Preload("Teacher").Preload("Student").Where("users.id <> ? AND users.is_active = ?", currentUserID, true)
	if role == "siswa" {
		query = query.Joins("JOIN roles ON roles.id = users.role_id").Where("roles.name IN ?", []string{"admin", "guru"})
	} else if role == "guru" {
		if classID == nil {
			return []models.User{}, nil
		}
		owns, err := s.teacherOwnsClass(currentUserID, *classID)
		if err != nil {
			return nil, err
		}
		if !owns {
			return nil, errors.New("guru hanya dapat memilih kelas yang dia ajar")
		}
		query = query.Joins("JOIN roles ON roles.id = users.role_id").Joins("JOIN students ON students.user_id = users.id").Joins("JOIN class_students ON class_students.student_id = students.id").Where("roles.name = ? AND class_students.class_id = ?", "siswa", *classID)
	}
	err = query.Order("users.username ASC").Find(&users).Error
	setChatDisplayNames(users)
	return users, err
}

func setChatDisplayName(user *models.User) {
	if user == nil {
		return
	}
	if user.Role.Name == "admin" {
		user.DisplayName = "Admin"
	} else if user.Student != nil && user.Student.FullName != "" {
		user.DisplayName = user.Student.FullName
	} else if user.Teacher != nil && user.Teacher.FullName != "" {
		user.DisplayName = user.Teacher.FullName
	} else {
		user.DisplayName = user.Username
	}
}

func setChatDisplayNames(users []models.User) {
	for index := range users {
		setChatDisplayName(&users[index])
	}
}

func (s *Service) GetTeacherClasses(userID string) ([]models.Class, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("user id tidak valid")
	}
	teacherID, err := s.teacherIDForUser(uid)
	if err != nil {
		return nil, err
	}
	var classes []models.Class
	err = s.DB.Joins("JOIN class_subjects ON class_subjects.class_id = classes.id").Where("class_subjects.teacher_id = ?", teacherID).Preload("Major").Distinct("classes.*").Order("classes.name ASC").Find(&classes).Error
	return classes, err
}

func (s *Service) userRole(userID uuid.UUID) (string, error) {
	var user models.User
	if err := s.DB.Preload("Role").First(&user, "id = ?", userID).Error; err != nil {
		return "", err
	}
	return user.Role.Name, nil
}

func (s *Service) teacherIDForUser(userID uuid.UUID) (uuid.UUID, error) {
	var teacher models.Teacher
	if err := s.DB.Where("user_id = ?", userID).First(&teacher).Error; err != nil {
		return uuid.Nil, errors.New("akun guru tidak ditemukan")
	}
	return teacher.ID, nil
}

func (s *Service) teacherOwnsClass(userID, classID uuid.UUID) (bool, error) {
	teacherID, err := s.teacherIDForUser(userID)
	if err != nil {
		return false, err
	}
	var count int64
	err = s.DB.Model(&models.ClassSubject{}).Where("class_id = ? AND teacher_id = ?", classID, teacherID).Count(&count).Error
	return count > 0, err
}

func (s *Service) studentInClass(userID, classID uuid.UUID) (bool, error) {
	var count int64
	err := s.DB.Table("class_students").Joins("JOIN students ON students.id = class_students.student_id").Where("class_students.class_id = ? AND students.user_id = ?", classID, userID).Count(&count).Error
	return count > 0, err
}

func (s *Service) authorizeConversation(userID, conversationID string, write bool) (*models.Conversation, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("user id tidak valid")
	}
	cid, err := uuid.Parse(conversationID)
	if err != nil {
		return nil, errors.New("conversation id tidak valid")
	}
	var conversation models.Conversation
	if err := s.DB.First(&conversation, "id = ?", cid).Error; err != nil {
		return nil, errors.New("ruang chat tidak ditemukan")
	}
	role, err := s.userRole(uid)
	if err != nil {
		return nil, err
	}
	if conversation.Type == "customer_service" {
		if role == "admin" {
			if write && (conversation.AssignedAdminID == nil || *conversation.AssignedAdminID != uid) {
				return nil, errors.New("admin harus menerima percakapan ini terlebih dahulu")
			}
			return &conversation, nil
		}
		if role != "guru" && role != "siswa" {
			return nil, errors.New("hanya siswa dan guru yang dapat menggunakan customer service")
		}
		var membership models.ConversationParticipant
		if err := s.DB.Where("conversation_id = ? AND user_id = ?", cid, uid).First(&membership).Error; err != nil {
			return nil, errors.New("anda bukan pemilik percakapan customer service ini")
		}
		if conversation.Status == "closed" && write {
			return nil, errors.New("percakapan customer service sudah ditutup")
		}
		return &conversation, nil
	}
	if role == "admin" {
		if conversation.Type != "private" {
			return nil, errors.New("admin hanya dapat mengakses private chat")
		}
		return &conversation, nil
	}
	if role == "guru" {
		if conversation.Type == "private" {
			if write {
				return nil, errors.New("guru hanya dapat menerima private chat dari admin")
			}
			var membership models.ConversationParticipant
			if err := s.DB.Where("conversation_id = ? AND user_id = ?", cid, uid).First(&membership).Error; err != nil {
				return nil, errors.New("anda bukan peserta private chat ini")
			}
			return &conversation, nil
		}
		if conversation.Type != "class" {
			return nil, errors.New("guru hanya dapat mengakses ruang chat kelas")
		}
		if conversation.OwnerID == nil || *conversation.OwnerID != uid {
			return nil, errors.New("guru hanya dapat mengakses ruang chat miliknya sendiri")
		}
		return &conversation, nil
	}
	if role != "siswa" {
		return nil, errors.New("role tidak diizinkan untuk chat")
	}
	if conversation.Type == "private" {
		if write {
			return nil, errors.New("siswa hanya dapat menerima private chat dari admin")
		}
		var membership models.ConversationParticipant
		if err := s.DB.Where("conversation_id = ? AND user_id = ?", cid, uid).First(&membership).Error; err != nil {
			return nil, errors.New("anda bukan peserta private chat ini")
		}
		return &conversation, nil
	}
	if conversation.Type != "class" {
		return nil, errors.New("siswa hanya dapat mengakses ruang chat kelas")
	}
	var membership models.ConversationParticipant
	if err := s.DB.Where("conversation_id = ? AND user_id = ?", cid, uid).First(&membership).Error; err != nil {
		return nil, errors.New("anda belum menjadi member ruang chat ini")
	}
	if conversation.ClassID == nil {
		return nil, errors.New("ruang chat siswa tidak memiliki kelas")
	}
	inClass, err := s.studentInClass(uid, *conversation.ClassID)
	if err != nil {
		return nil, err
	}
	if !inClass {
		return nil, errors.New("anda tidak terdaftar di kelas ruang chat ini")
	}
	return &conversation, nil
}

func (s *Service) CreateCustomerService(userID string) (*models.Conversation, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("user id tidak valid")
	}
	role, err := s.userRole(uid)
	if err != nil {
		return nil, err
	}
	if role != "siswa" && role != "guru" {
		return nil, errors.New("customer service hanya tersedia untuk siswa dan guru")
	}
	var existing models.Conversation
	if err := s.DB.Where("type = ? AND status IN ?", "customer_service", []string{"waiting", "active"}).
		Joins("JOIN conversation_participants cp ON cp.conversation_id = conversations.id").
		Where("cp.user_id = ?", uid).Order("conversations.updated_at DESC").First(&existing).Error; err == nil {
		return &existing, nil
	}
	name := "Customer Service"
	conversation := models.Conversation{Type: "customer_service", Status: "waiting", Name: &name, OwnerID: &uid}
	if err := s.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&conversation).Error; err != nil {
			return err
		}
		return tx.Create(&models.ConversationParticipant{ConversationID: conversation.ID, UserID: uid}).Error
	}); err != nil {
		return nil, err
	}
	return &conversation, nil
}

func (s *Service) ListCustomerService(userID, role string) ([]models.Conversation, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("user id tidak valid")
	}
	query := s.DB.Preload("Participants.User").Preload("Participants.User.Role").Preload("Participants.User.Student").Preload("Participants.User.Teacher").
		Preload("AssignedAdmin").Where("conversations.type = ?", "customer_service").Order("conversations.updated_at DESC")
	if role != "admin" {
		query = query.Joins("JOIN conversation_participants cp ON cp.conversation_id = conversations.id").Where("cp.user_id = ?", uid)
	}
	var conversations []models.Conversation
	if err := query.Find(&conversations).Error; err != nil {
		return nil, err
	}
	for i := range conversations {
		for j := range conversations[i].Participants {
			setChatDisplayName(&conversations[i].Participants[j].User)
		}
	}
	return conversations, nil
}

func (s *Service) ClaimCustomerService(userID, conversationID string) (*models.Conversation, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("user id tidak valid")
	}
	if role, err := s.userRole(uid); err != nil || role != "admin" {
		return nil, errors.New("hanya admin yang dapat menerima percakapan")
	}
	var conversation models.Conversation
	if err := s.DB.Where("id = ? AND type = ?", conversationID, "customer_service").First(&conversation).Error; err != nil {
		return nil, errors.New("percakapan customer service tidak ditemukan")
	}
	if conversation.Status == "closed" {
		return nil, errors.New("percakapan sudah ditutup")
	}
	nowStatus := "active"
	conversation.Status = nowStatus
	conversation.AssignedAdminID = &uid
	if err := s.DB.Save(&conversation).Error; err != nil {
		return nil, err
	}
	member := models.ConversationParticipant{ConversationID: conversation.ID, UserID: uid}
	if err := s.DB.Where("conversation_id = ? AND user_id = ?", conversation.ID, uid).FirstOrCreate(&member).Error; err != nil {
		return nil, err
	}
	return &conversation, nil
}

func (s *Service) CloseCustomerService(userID, conversationID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return errors.New("user id tidak valid")
	}
	if role, err := s.userRole(uid); err != nil || role != "admin" {
		return errors.New("hanya admin yang dapat menutup percakapan")
	}
	now := time.Now()
	return s.DB.Model(&models.Conversation{}).Where("id = ? AND type = ?", conversationID, "customer_service").Updates(map[string]any{"status": "closed", "closed_at": &now}).Error
}

func (s *Service) CreateConversation(
	userID string,
	req CreateConversationRequest,
) (*models.Conversation, error) {

	currentUserID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("user id tidak valid")
	}

	role, err := s.userRole(currentUserID)
	if err != nil {
		return nil, err
	}
	if req.Type == "" {
		req.Type = "private"
	}
	if role == "siswa" {
		return nil, errors.New("siswa tidak dapat membuat ruang chat")
	}
	if role == "guru" && req.Type != "class" {
		return nil, errors.New("guru hanya dapat membuat ruang chat kelas")
	}
	if role == "admin" && req.Type != "private" {
		return nil, errors.New("admin hanya dapat membuat private chat dari menu ini")
	}
	if req.Type != "private" && req.Type != "class" {
		return nil, errors.New("tipe ruang chat tidak valid")
	}
	if req.Type == "class" && req.ClassID == nil {
		return nil, errors.New("kelas wajib dipilih untuk ruang chat")
	}
	if req.Type == "class" {
		owns, err := s.teacherOwnsClass(currentUserID, *req.ClassID)
		if err != nil {
			return nil, err
		}
		if !owns {
			return nil, errors.New("guru hanya dapat memilih kelas yang dia ajar")
		}
	}

	// Pastikan pembuat conversation ikut menjadi participant.
	participantMap := make(map[uuid.UUID]bool)

	participantMap[currentUserID] = true

	for _, participantID := range req.ParticipantIDs {
		participantMap[participantID] = true
	}

	participants := make([]uuid.UUID, 0, len(participantMap))

	for participantID := range participantMap {
		participants = append(participants, participantID)
	}

	if req.Type == "private" {
		if role != "admin" {
			return nil, errors.New("private chat hanya dapat dibuat oleh admin")
		}
		for _, participantID := range participants {
			if participantID == currentUserID {
				continue
			}
			targetRole, err := s.userRole(participantID)
			if err != nil {
				return nil, err
			}
			if targetRole != "guru" && targetRole != "siswa" {
				return nil, errors.New("private chat admin hanya dapat dikirim kepada guru atau siswa")
			}
		}
	}
	// Private chat harus tepat 2 orang.
	if req.Type == "private" && len(participants) != 2 {
		return nil, errors.New(
			"private conversation harus memiliki 2 peserta",
		)
	}

	// Pastikan semua user memang ada.
	var userCount int64

	err = s.DB.
		Model(&models.User{}).
		Where("id IN ?", participants).
		Count(&userCount).Error

	if err != nil {
		return nil, err
	}

	if userCount != int64(len(participants)) {
		return nil, errors.New(
			"salah satu participant tidak ditemukan",
		)
	}
	if req.Type == "class" {
		var studentCount int64
		studentCountQuery := s.DB.Table("students").Where("students.user_id IN ?", participants)
		studentCountQuery = studentCountQuery.Joins("JOIN class_students ON class_students.student_id = students.id").Where("class_students.class_id = ?", *req.ClassID)
		if err := studentCountQuery.Count(&studentCount).Error; err != nil {
			return nil, err
		}
		if studentCount != int64(len(participants)-1) {
			return nil, errors.New("semua peserta ruang chat harus siswa dari kelas yang dipilih")
		}
	}

	conversation := models.Conversation{
		Type:    req.Type,
		Name:    req.Name,
		OwnerID: &currentUserID,
		ClassID: req.ClassID,
	}

	err = s.DB.Transaction(func(tx *gorm.DB) error {

		if err := tx.Create(&conversation).Error; err != nil {
			return err
		}

		for _, participantID := range participants {

			participant := models.ConversationParticipant{
				ConversationID: conversation.ID,
				UserID:         participantID,
			}

			if err := tx.Create(&participant).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	err = s.DB.
		Preload("Participants").
		First(&conversation, "id = ?", conversation.ID).
		Error

	if err != nil {
		return nil, err
	}

	return &conversation, nil
}

func (s *Service) GetUserConversations(
	userID string,
) ([]models.Conversation, error) {

	currentUserID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("user id tidak valid")
	}

	var conversations []models.Conversation

	role, err := s.userRole(currentUserID)
	if err != nil {
		return nil, err
	}
	query := s.DB.
		Preload("Participants.User").
		Preload("Participants.User.Role").
		Preload("Participants.User.Teacher").
		Preload("Participants.User.Student").
		Order("conversations.updated_at DESC")
	if role == "admin" {
		err = query.Where("type = ?", "private").Find(&conversations).Error
	} else if role == "guru" {
		err = query.Where("owner_id = ? AND type = ?", currentUserID, "class").Find(&conversations).Error
	} else {
		err = query.Joins("JOIN conversation_participants cp ON cp.conversation_id = conversations.id").Where("cp.user_id = ? AND conversations.type = ?", currentUserID, "class").Find(&conversations).Error
	}

	if err != nil {
		return nil, err
	}
	for index := range conversations {
		for participantIndex := range conversations[index].Participants {
			setChatDisplayName(&conversations[index].Participants[participantIndex].User)
		}
	}

	return conversations, nil
}

func (s *Service) IsParticipant(
	userID string,
	conversationID string,
) (bool, error) {

	currentUserID, err := uuid.Parse(userID)
	if err != nil {
		return false, errors.New("user id tidak valid")
	}

	currentConversationID, err := uuid.Parse(conversationID)
	if err != nil {
		return false, errors.New("conversation id tidak valid")
	}

	var count int64

	err = s.DB.
		Model(&models.ConversationParticipant{}).
		Where(
			"conversation_id = ? AND user_id = ?",
			currentConversationID,
			currentUserID,
		).
		Count(&count).
		Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (s *Service) UpdateConversation(userID, conversationID string, req UpdateConversationRequest) (*models.Conversation, error) {
	conversation, err := s.authorizeConversation(userID, conversationID, false)
	if err != nil {
		return nil, err
	}
	role, err := s.userRole(uuid.MustParse(userID))
	if err != nil {
		return nil, err
	}
	if role != "guru" || conversation.Type != "class" || conversation.OwnerID == nil || conversation.OwnerID.String() != userID {
		return nil, errors.New("hanya pemilik ruang chat yang dapat mengeditnya")
	}
	conversation.Name = req.Name
	return conversation, s.DB.Save(conversation).Error
}

func (s *Service) DeleteConversation(userID, conversationID string) error {
	conversation, err := s.authorizeConversation(userID, conversationID, false)
	if err != nil {
		return err
	}
	role, err := s.userRole(uuid.MustParse(userID))
	if err != nil {
		return err
	}
	if role != "admin" && (conversation.OwnerID == nil || conversation.OwnerID.String() != userID) {
		return errors.New("hanya pemilik ruang chat yang dapat menghapusnya")
	}
	return s.DB.Delete(&models.Conversation{}, "id = ?", conversation.ID).Error
}

func (s *Service) InviteStudent(userID, conversationID string, studentUserID uuid.UUID) error {
	conversation, err := s.authorizeConversation(userID, conversationID, false)
	if err != nil {
		return err
	}
	uid := uuid.MustParse(userID)
	role, err := s.userRole(uid)
	if err != nil {
		return err
	}
	if role != "admin" && (conversation.OwnerID == nil || *conversation.OwnerID != uid) {
		return errors.New("hanya pemilik ruang chat yang dapat mengundang siswa")
	}
	if conversation.ClassID == nil {
		return errors.New("ruang chat ini tidak terikat kelas")
	}
	inClass, err := s.studentInClass(studentUserID, *conversation.ClassID)
	if err != nil {
		return err
	}
	if !inClass {
		return errors.New("siswa bukan anggota kelas ruang chat")
	}
	member := models.ConversationParticipant{ConversationID: conversation.ID, UserID: studentUserID}
	return s.DB.Where("conversation_id = ? AND user_id = ?", conversation.ID, studentUserID).FirstOrCreate(&member).Error
}

func (s *Service) SendMessage(
	userID string,
	conversationID string,
	req SendMessageRequest,
) (*models.Message, error) {
	currentUserID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("user id tidak valid")
	}

	currentConversationID, err := uuid.Parse(conversationID)
	if err != nil {
		return nil, errors.New("conversation id tidak valid")
	}

	if _, err := s.authorizeConversation(userID, conversationID, true); err != nil {
		return nil, err
	}

	// Default message type
	if req.MessageType == "" {
		req.MessageType = "text"
	}

	// Validasi message type
	switch req.MessageType {
	case "text", "file", "image", "system":
	default:
		return nil, errors.New("tipe message tidak valid")
	}

	// Untuk sementara content wajib ada.
	if req.Content == "" {
		return nil, errors.New("message tidak boleh kosong")
	}

	message := models.Message{
		ConversationID: currentConversationID,
		SenderID:       currentUserID,
		MessageType:    req.MessageType,
		Content:        req.Content,
	}

	err = s.DB.Transaction(func(tx *gorm.DB) error {

		if err := tx.Create(&message).Error; err != nil {
			return err
		}

		// Update waktu conversation.
		if err := tx.
			Model(&models.Conversation{}).
			Where("id = ?", currentConversationID).
			Update("updated_at", message.CreatedAt).
			Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Ambil kembali dengan data sender
	err = s.DB.
		Preload("Sender").
		Preload("Sender.Role").
		Preload("Sender.Teacher").
		Preload("Sender.Student").
		First(&message, "id = ?", message.ID).
		Error

	if err != nil {
		return nil, err
	}
	setChatDisplayName(&message.Sender)

	// Kirim notifikasi ke participant lain
	go s.notifyMessageParticipants(
		&message,
	)

	return &message, nil
}

func (s *Service) GetMessages(
	userID string,
	conversationID string,
) ([]models.Message, error) {

	currentConversationID, err := uuid.Parse(conversationID)
	if err != nil {
		return nil, errors.New("conversation id tidak valid")
	}

	if _, err := s.authorizeConversation(userID, conversationID, false); err != nil {
		return nil, err
	}

	var messages []models.Message

	err = s.DB.
		Preload("Sender").
		Preload("Sender.Role").
		Preload("Sender.Teacher").
		Preload("Sender.Student").
		Where(
			"conversation_id = ?",
			currentConversationID,
		).
		Order("created_at ASC").
		Find(&messages).
		Error

	if err != nil {
		return nil, err
	}
	for index := range messages {
		setChatDisplayName(&messages[index].Sender)
	}

	return messages, nil
}

func (s *Service) GetUserPresence(
	userID string,
	hub *Hub,
) (bool, error) {

	id, err := uuid.Parse(userID)

	if err != nil {
		return false, errors.New(
			"user id tidak valid",
		)
	}

	// Pastikan user ada.
	var user models.User

	err = s.DB.
		First(&user, "id = ?", id).
		Error

	if err != nil {

		if errors.Is(
			err,
			gorm.ErrRecordNotFound,
		) {
			return false, errors.New(
				"user tidak ditemukan",
			)
		}

		return false, err
	}

	return hub.IsOnline(id), nil
}

func (s *Service) MarkMessagesAsRead(
	userID string,
	conversationID string,
	messageIDs []string,
) ([]models.MessageRead, error) {

	currentUserID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("user id tidak valid")
	}

	currentConversationID, err := uuid.Parse(conversationID)
	if err != nil {
		return nil, errors.New("conversation id tidak valid")
	}
	if _, err := s.authorizeConversation(userID, conversationID, false); err != nil {
		return nil, err
	}

	if len(messageIDs) == 0 {
		return nil, errors.New(
			"message_ids wajib diisi",
		)
	}

	var reads []models.MessageRead

	err = s.DB.Transaction(
		func(tx *gorm.DB) error {

			for _, messageIDString := range messageIDs {

				messageID, err := uuid.Parse(
					messageIDString,
				)

				if err != nil {
					return errors.New(
						"message id tidak valid",
					)
				}

				// Pastikan message berada
				// di conversation tersebut
				var message models.Message

				err = tx.
					Where(
						"id = ? AND conversation_id = ?",
						messageID,
						currentConversationID,
					).
					First(&message).
					Error

				if err != nil {

					if errors.Is(
						err,
						gorm.ErrRecordNotFound,
					) {
						return errors.New(
							"message tidak ditemukan",
						)
					}

					return err
				}

				// Jangan tandai pesan sendiri
				if message.SenderID == currentUserID {
					continue
				}

				read := models.MessageRead{
					MessageID: messageID,
					UserID:    currentUserID,
					ReadAt:    time.Now(),
				}

				// Hindari duplicate read
				err = tx.
					Where(
						"message_id = ? AND user_id = ?",
						messageID,
						currentUserID,
					).
					FirstOrCreate(&read).
					Error

				if err != nil {
					return err
				}

				reads = append(reads, read)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return reads, nil
}

func (s *Service) GetUnreadCounts(
	userID string,
) ([]ConversationUnreadCount, error) {

	currentUserID, err := uuid.Parse(userID)

	if err != nil {
		return nil, errors.New(
			"user id tidak valid",
		)
	}

	var results []ConversationUnreadCount

	err = s.DB.
		Table("conversation_participants AS cp").
		Select(`
			cp.conversation_id,
			COUNT(m.id) AS unread_count
		`).
		Joins(`
			LEFT JOIN messages m
			ON m.conversation_id = cp.conversation_id
			AND m.sender_id != ?
		`, currentUserID).
		Joins(`
			LEFT JOIN message_reads mr
			ON mr.message_id = m.id
			AND mr.user_id = ?
		`, currentUserID).
		Where(
			"cp.user_id = ? AND mr.id IS NULL",
			currentUserID,
		).
		Group(
			"cp.conversation_id",
		).
		Find(&results).
		Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

func (s *Service) notifyMessageParticipants(
	message *models.Message,
) {

	if s.NotificationService == nil {
		return
	}

	var participants []models.ConversationParticipant

	err := s.DB.
		Where(
			"conversation_id = ?",
			message.ConversationID,
		).
		Find(&participants).
		Error

	if err != nil {
		return
	}

	messageType := "Pesan baru"

	if message.MessageType == "image" {
		messageType = "Gambar baru"
	}

	if message.MessageType == "file" {
		messageType = "File baru"
	}

	for _, participant := range participants {

		// Jangan kirim notifikasi
		// kepada pengirim sendiri.
		if participant.UserID == message.SenderID {
			continue
		}

		referenceType := "conversation"

		_, _ = s.NotificationService.Create(
			notification.CreateNotificationRequest{
				UserID:        participant.UserID,
				Type:          "chat_message",
				Title:         messageType,
				Message:       message.Content,
				ReferenceType: &referenceType,
				ReferenceID:   &message.ConversationID,
			},
		)
	}
}
