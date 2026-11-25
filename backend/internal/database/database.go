package database

import (
	"fmt"
	"log"

	"school-lms/internal/config"
	"school-lms/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(cfg *config.Config) *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
		cfg.DBSSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("Failed to connect database:", err)
	}

	if err := db.AutoMigrate(
		&models.Role{},
		&models.User{},
		&models.Admin{},
		&models.Teacher{},
		&models.Student{},
		&models.AcademicYear{},
		&models.Semester{},
		&models.AcademicAgenda{},
		&models.Conversation{},
		&models.FAQ{},
		&models.ConversationParticipant{},
		&models.Major{},
		&models.Class{},
		&models.Subject{},
		&models.Teacher{},
		&models.ClassSubject{},
		&models.Material{},
		&models.MaterialFile{},
		&models.Assignment{},
		&models.AssignmentFile{},
		&models.AssignmentSubmission{},
		&models.AssignmentSubmissionFile{},
		&models.Quiz{},
		&models.QuizQuestion{},
		&models.QuizQuestionOption{},
		&models.QuizAttempt{},
		&models.QuizAnswer{},
		&models.QuizAttemptAnswer{},
		&models.Message{},
		&models.MessageRead{},
		&models.Meeting{},
		&models.MeetingParticipant{},
		&models.MeetingActivityLog{},
	); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	if err := db.Exec(`
		ALTER TABLE conversations DROP CONSTRAINT IF EXISTS chk_conversations_type;
		ALTER TABLE conversations ADD CONSTRAINT chk_conversations_type
			CHECK (type IN ('private', 'group', 'class', 'customer_service'));
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint WHERE conname = 'chk_customer_service_status'
			) THEN
				ALTER TABLE conversations ADD CONSTRAINT chk_customer_service_status
					CHECK (status IN ('waiting', 'active', 'closed'));
			END IF;
		END $$;
		CREATE INDEX IF NOT EXISTS idx_conversations_cs_queue
			ON conversations(type, status, updated_at DESC);
	`).Error; err != nil {
		log.Fatal("Failed to enforce customer service conversation schema:", err)
	}

	if err := db.Exec(`
		ALTER TABLE meeting_participants
			DROP CONSTRAINT IF EXISTS fk_meetings_participants;
		ALTER TABLE meeting_participants
			ADD CONSTRAINT fk_meetings_participants
			FOREIGN KEY (meeting_id) REFERENCES meetings(id) ON DELETE CASCADE;
	`).Error; err != nil {
		log.Fatal("Failed to enforce meeting participant cascade delete:", err)
	}

	if err := db.Exec(`
		ALTER TABLE material_files
			DROP CONSTRAINT IF EXISTS fk_materials_files;
		ALTER TABLE material_files
			ADD CONSTRAINT fk_materials_files
			FOREIGN KEY (material_id) REFERENCES materials(id) ON DELETE CASCADE;
	`).Error; err != nil {
		log.Fatal("Failed to enforce material file cascade delete:", err)
	}

	if err := db.Exec(`
		INSERT INTO subject_teachers (subject_id, teacher_id)
		SELECT DISTINCT subject_id, teacher_id
		FROM class_subjects
		WHERE teacher_id IS NOT NULL
		ON CONFLICT (subject_id, teacher_id) DO NOTHING;

		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint
				WHERE conname = 'class_subjects_subject_id_teacher_id_fkey'
			) THEN
				ALTER TABLE class_subjects
				ADD CONSTRAINT class_subjects_subject_id_teacher_id_fkey
				FOREIGN KEY (subject_id, teacher_id)
				REFERENCES subject_teachers (subject_id, teacher_id);
			END IF;
		END $$;
	`).Error; err != nil {
		log.Fatal("Failed to enforce class subject teacher relation:", err)
	}

	if err := db.Exec(`
		DELETE FROM assignment_submissions older
		USING assignment_submissions newer
		WHERE older.assignment_id = newer.assignment_id
		  AND older.student_id = newer.student_id
		  AND older.created_at > newer.created_at;
		CREATE UNIQUE INDEX IF NOT EXISTS uq_assignment_submissions_assignment_student
		ON assignment_submissions (assignment_id, student_id);
	`).Error; err != nil {
		log.Fatal("Failed to enforce one assignment submission per student:", err)
	}

	_ = db.Exec(`
		DO $$
		BEGIN
			IF EXISTS (
				SELECT 1 FROM pg_constraint
				WHERE conname = 'fk_quiz_questions_options'
			) THEN
				ALTER TABLE quiz_question_options DROP CONSTRAINT fk_quiz_questions_options;
			END IF;

			ALTER TABLE quiz_question_options
			ADD CONSTRAINT fk_quiz_questions_options
			FOREIGN KEY (question_id) REFERENCES quiz_questions(id) ON DELETE CASCADE;

			IF EXISTS (
				SELECT 1 FROM pg_constraint
				WHERE conname = 'fk_quizzes_questions'
			) THEN
				ALTER TABLE quiz_questions DROP CONSTRAINT fk_quizzes_questions;
			END IF;

			ALTER TABLE quiz_questions
			ADD CONSTRAINT fk_quizzes_questions
			FOREIGN KEY (quiz_id) REFERENCES quizzes(id) ON DELETE CASCADE;
		END $$;
	`)

	log.Println("Database connected successfully")

	return db
}
