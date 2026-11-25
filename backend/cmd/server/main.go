package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"school-lms/internal/academicagenda"
	"school-lms/internal/academicyear"
	"school-lms/internal/assignment"
	"school-lms/internal/assignmentfile"
	"school-lms/internal/assignmentsubmission"
	"school-lms/internal/assignmentsubmissionfile"
	"school-lms/internal/attendance"
	"school-lms/internal/auth"
	"school-lms/internal/chat"
	"school-lms/internal/classroom"
	"school-lms/internal/config"
	"school-lms/internal/database"
	"school-lms/internal/helpcenter"
	"school-lms/internal/livekit"
	"school-lms/internal/major"
	"school-lms/internal/material"
	"school-lms/internal/materialfile"
	"school-lms/internal/meeting"
	"school-lms/internal/middleware"
	"school-lms/internal/notification"
	"school-lms/internal/quiz"
	"school-lms/internal/quizanswer"
	"school-lms/internal/quizattempt"
	"school-lms/internal/subject"
	"school-lms/internal/user"
)

func main() {
	cfg := config.Load()

	// 1. Inisialisasi koneksi database
	db := database.Connect(cfg)

	// 2. Jalankan seeder database
	database.SeedRoles(db)
	if err := database.SeedAdmin(db, cfg.InitialAdminPassword); err != nil {
		log.Fatal(err)
	}

	// 3. Inisialisasi router
	router := gin.Default()
	router.Use(middleware.ErrorLog())
	router.Use(func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		allowedOrigins := strings.Split(cfg.CORSAllowedOrigins, ",")
		allowed := origin == "http://localhost:3000" || origin == "http://localhost:3001"
		for _, configuredOrigin := range allowedOrigins {
			if strings.TrimSpace(configuredOrigin) == origin && origin != "" {
				allowed = true
				break
			}
		}
		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})
	router.Static("/uploads", "./uploads")

	// 4. Inisialisasi Service dan Handler
	authService := &auth.Service{
		DB:        db,
		JWTSecret: cfg.JWTSecret,
	}
	authHandler := &auth.Handler{
		Service: authService,
	}

	userService := user.NewService(db)
	userHandler := user.NewHandler(userService)

	classService := classroom.NewService(db)
	classHandler := classroom.NewHandler(classService)

	subjectService := subject.NewService(db)
	subjectHandler := subject.NewHandler(subjectService)

	majorService := major.NewService(db)
	majorHandler := major.NewHandler(majorService)

	materialService := material.NewService(db)
	materialHandler := material.NewHandler(materialService)

	materialFileService := materialfile.NewService(db)
	materialFileHandler := materialfile.NewHandler(materialFileService)

	assignmentService := assignment.NewService(db)
	assignmentHandler := assignment.NewHandler(assignmentService)

	assignmentFileService := assignmentfile.NewService(db)
	assignmentFileHandler := assignmentfile.NewHandler(assignmentFileService)

	academicYearService := academicyear.NewService(db)
	academicYearHandler := academicyear.NewHandler(academicYearService)
	agendaService := academicagenda.NewService(db)
	agendaHandler := academicagenda.NewHandler(agendaService)

	// Inisialisasi Assignment Submission
	assignmentSubmissionService := assignmentsubmission.NewService(db)
	assignmentSubmissionHandler := assignmentsubmission.NewHandler(assignmentSubmissionService)

	assignmentSubmissionFileService := assignmentsubmissionfile.NewService(db)
	assignmentSubmissionFileHandler := assignmentsubmissionfile.NewHandler(
		assignmentSubmissionFileService,
	)

	// Inisialisasi Quiz
	quizService := quiz.NewService(db)
	quizHandler := quiz.NewHandler(
		quizService,
	)

	// Inisialisasi Quiz Attempt
	quizAttemptService := quizattempt.NewService(db)
	quizAttemptService.StartAutoExpireWorker() // Start background worker untuk expire attempt
	quizAttemptHandler := quizattempt.NewHandler(quizAttemptService)

	// Inisialisasi Quiz Answer
	quizAnswerService := quizanswer.NewService(db)
	quizAnswerHandler := quizanswer.NewHandler(quizAnswerService)

	// Inisialisasi Attendance
	attendanceService := attendance.NewService(db)
	attendanceHandler := attendance.NewHandler(attendanceService)

	// Inisialisasi Notification
	notificationService := notification.NewService(db)
	notificationHandler := notification.NewHandler(notificationService)

	// Inisialisasi Chat
	chatService := chat.NewService(
		db,
		notificationService,
	)

	chatHub := chat.NewHub()

	go chatHub.Run()

	chatHandler := chat.NewHandler(
		chatService,
		chatHub,
	)
	helpCenterService := helpcenter.NewService(db)
	helpCenterHandler := helpcenter.NewHandler(helpCenterService)

	webSocketHandler := chat.NewWebSocketHandler(
		chatHub,
		chatService,
		cfg.JWTSecret,
	)
	// Inisialisasi LiveKit & Meeting Service
	liveKitService := livekit.NewService(livekit.Config{
		URL:       cfg.LiveKitURL,
		APIKey:    cfg.LiveKitAPIKey,
		APISecret: cfg.LiveKitAPISecret,
	})
	meetingService := meeting.NewService(db, liveKitService)
	meetingHandler := meeting.NewHandler(meetingService)

	// 5. Setup Route Group
	api := router.Group("/api")
	{
		// Route public
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"success": true,
				"message": "School LMS API is running",
			})
		})

		api.POST("/auth/login", authHandler.Login)

		// WebSocket route (tanpa middleware AuthMiddleware)
		api.GET(
			"/ws/chat",
			webSocketHandler.Connect,
		)

		// Route protected (semua route di dalam sini otomatis harus login / pakai JWT)
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			// Current authenticated user
			protected.GET("/auth/me", authHandler.Me)
			protected.GET("/profile", userHandler.GetProfile)
			protected.PUT("/profile", userHandler.UpdateProfile)

			// Test authentication
			protected.GET("/protected/me", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"message": "Authenticated",
					"user_id": c.GetString("user_id"),
					"role":    c.GetString("role"),
				})
			})

			// Route Admin
			admin := protected.Group("/admin")
			admin.Use(middleware.RequireRole("admin"))
			{
				admin.GET("/faqs", helpCenterHandler.List)
				admin.POST("/faqs", helpCenterHandler.Create)
				admin.PUT("/faqs/:id", helpCenterHandler.Update)
				admin.DELETE("/faqs/:id", helpCenterHandler.Delete)
				admin.GET("/customer-service", chatHandler.ListCustomerService)
				admin.POST("/customer-service/:id/claim", chatHandler.ClaimCustomerService)
				admin.POST("/customer-service/:id/close", chatHandler.CloseCustomerService)
				admin.GET("/meetings", meetingHandler.List)
				admin.DELETE("/meetings/:id", meetingHandler.Delete)
				admin.GET("/dashboard", func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{
						"message": "Welcome to Admin Dashboard",
						"role":    "admin",
					})
				})

				admin.GET("/users", userHandler.GetAll)
				admin.GET("/users/:id", userHandler.GetByID)
				admin.PUT("/users/:id", userHandler.Update)
				admin.DELETE("/users/:id", userHandler.Delete)
				admin.GET("/students", userHandler.GetStudents)
				admin.GET("/students/search", userHandler.SearchStudents)
				admin.GET("/classes", classHandler.GetAll)
				admin.POST("/classes", classHandler.Create)
				admin.GET("/classes/detail/:id", classHandler.GetByID)
				admin.PUT("/classes/update/:id", classHandler.Update)
				admin.DELETE("/classes/delete/:id", classHandler.Delete)
				admin.GET("/subjects", subjectHandler.GetAll)
				admin.GET("/subjects/:id", subjectHandler.GetByID)
				admin.PUT("/subjects/:id", subjectHandler.Update)
				admin.DELETE("/subjects/delete/:id", subjectHandler.Delete)
				admin.GET("/subjects/search", subjectHandler.Search)
				admin.GET("/academic-years", academicYearHandler.GetAll)
				admin.POST("/academic-years", academicYearHandler.Create)
				admin.PUT("/academic-years/:id", academicYearHandler.Update)
				admin.POST("/academic-years/:id/duplicate", academicYearHandler.Duplicate)
				admin.DELETE("/academic-years/:id", academicYearHandler.Delete)
				admin.POST("/academic-years/:id/semesters", academicYearHandler.CreateSemester)
				admin.PUT("/academic-years/:id/semesters/:semester_id", academicYearHandler.UpdateSemester)
				admin.DELETE("/academic-years/:id/semesters/:semester_id", academicYearHandler.DeleteSemester)
				admin.POST("/academic-years/:id/agendas", agendaHandler.Create)
				admin.PUT("/academic-years/:id/agendas/:agenda_id", agendaHandler.Update)
				admin.DELETE("/academic-years/:id/agendas/:agenda_id", agendaHandler.Delete)
				admin.POST("/subjects", subjectHandler.Create)
				admin.GET("/majors", majorHandler.GetAll)
				admin.GET("/majors/:id", majorHandler.GetByID)
				admin.POST("/majors", majorHandler.Create)
				admin.PUT("/majors/:id", majorHandler.Update)
				admin.DELETE("/majors/:id", majorHandler.Delete)
				admin.GET("/classes/:class_id/students", classHandler.GetStudents)
				admin.POST("/classes/:class_id/students", classHandler.AddStudent)
				admin.DELETE("/classes/:class_id/students/:student_id", classHandler.RemoveStudent)
				admin.GET("/classes/:class_id/subjects", classHandler.GetSubjects)
				admin.GET(
					"/classes/:class_id/subjects/detail",
					classHandler.GetSubjectDetails,
				)
				admin.POST("/classes/:class_id/subjects", classHandler.AddSubject)
				admin.DELETE("/classes/:class_id/subjects/:subject_id", classHandler.RemoveSubject)
				admin.PUT(
					"/classes/:class_id/subjects/:subject_id/teacher",
					classHandler.AssignTeacher,
				)

				admin.POST("/users", userHandler.Create)
			}
			protected.GET("/academic-years", academicYearHandler.GetAll)
			protected.GET("/academic-years/active", academicYearHandler.GetActive)
			protected.GET("/academic-years/:id/agendas", agendaHandler.List)

			// Route Guru
			guruRoutes := protected.Group("/guru")
			guruRoutes.Use(middleware.RequireRole("guru"))
			{
				guruRoutes.GET("/dashboard", func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{
						"success": true,
						"message": "Selamat datang di Guru Dashboard",
					})
				})
			}

			teacher := protected.Group("/teacher")
			teacher.Use(middleware.RequireRole("guru"))
			{
				teacher.POST("/materials", materialHandler.Create)
				teacher.GET("/materials", materialHandler.GetMyMaterials)
				teacher.GET("/material-files/:file_id/download", materialFileHandler.DownloadTeacherFile)
				teacher.GET("/class-subjects", classHandler.GetTeacherClassSubjects)
				teacher.GET(
					"/materials/:id",
					materialHandler.GetMyMaterialByID,
				)
				teacher.PUT(
					"/materials/:id",
					materialHandler.Update,
				)
				teacher.PATCH(
					"/materials/:id/publish",
					materialHandler.Publish,
				)
				teacher.DELETE(
					"/materials/:id",
					materialHandler.Delete,
				)
				teacher.POST(
					"/materials/:id/files",
					materialFileHandler.Upload,
				)
				teacher.GET(
					"/materials/:id/files",
					materialFileHandler.GetFiles,
				)
				teacher.DELETE(
					"/materials/:id/files/:file_id",
					materialFileHandler.DeleteFile,
				)
				teacher.GET(
					"/assignments",
					assignmentHandler.GetAll,
				)
				teacher.GET("/assignment-files/:file_id/download", assignmentFileHandler.DownloadTeacherFile)
				teacher.GET(
					"/assignments/:id",
					assignmentHandler.GetByID,
				)
				teacher.POST(
					"/assignments",
					assignmentHandler.Create,
				)
				teacher.PUT(
					"/assignments/:id",
					assignmentHandler.Update,
				)
				teacher.PATCH(
					"/assignments/:id/publish",
					assignmentHandler.Publish,
				)
				teacher.DELETE(
					"/assignments/:id",
					assignmentHandler.Delete,
				)
				teacher.POST(
					"/assignments/:id/files",
					assignmentFileHandler.Upload,
				)
				teacher.GET(
					"/assignments/:id/files",
					assignmentFileHandler.GetFiles,
				)
				teacher.DELETE(
					"/assignments/:id/files/:file_id",
					assignmentFileHandler.DeleteFile,
				)
				teacher.GET(
					"/assignments/:id/submissions",
					assignmentSubmissionHandler.GetTeacherSubmissions,
				)
				teacher.GET(
					"/submissions/:id",
					assignmentSubmissionHandler.GetTeacherSubmission,
				)
				teacher.PATCH(
					"/submissions/:id/grade",
					assignmentSubmissionHandler.GradeSubmission,
				)
				teacher.GET(
					"/submission-files/:file_id/download",
					assignmentSubmissionFileHandler.DownloadTeacherFile,
				)

				// Route Quiz Guru
				teacher.POST(
					"/quizzes",
					quizHandler.CreateQuiz,
				)
				teacher.GET(
					"/quizzes",
					quizHandler.GetTeacherQuizzes,
				)
				teacher.GET(
					"/quizzes/:id",
					quizHandler.GetTeacherQuiz,
				)
				teacher.PUT(
					"/quizzes/:id",
					quizHandler.UpdateQuiz,
				)
				teacher.DELETE(
					"/quizzes/:id",
					quizHandler.DeleteQuiz,
				)

				// ===== TAMBAHAN ROUTE PUBLISH & CLOSE QUIZ =====
				teacher.POST(
					"/quizzes/:id/publish",
					quizHandler.PublishQuiz,
				)
				teacher.POST(
					"/quizzes/:id/close",
					quizHandler.CloseQuiz,
				)
				// ===============================================

				// Route Quiz Question Guru
				teacher.POST(
					"/quizzes/:id/questions",
					quizHandler.CreateQuestion,
				)
				teacher.GET(
					"/quizzes/:id/questions",
					quizHandler.GetTeacherQuizQuestions,
				)
				teacher.GET(
					"/questions/:id",
					quizHandler.GetTeacherQuestion,
				)
				teacher.PUT(
					"/questions/:id",
					quizHandler.UpdateQuestion,
				)
				teacher.DELETE(
					"/questions/:id",
					quizHandler.DeleteQuestion,
				)

				// Route Quiz Question Option Guru
				teacher.POST(
					"/questions/:id/options",
					quizHandler.CreateQuestionOption,
				)
				teacher.GET(
					"/questions/:id/options",
					quizHandler.GetTeacherQuestionOptions,
				)
				teacher.PUT(
					"/options/:id",
					quizHandler.UpdateQuestionOption,
				)
				teacher.DELETE(
					"/options/:id",
					quizHandler.DeleteQuestionOption,
				)

				// ===== ROUTE UNTUK TEACHER ATTEMPTS & GRADING =====
				teacher.GET(
					"/quizzes/:id/attempts",
					quizAttemptHandler.GetTeacherAttempts,
				)
				teacher.GET(
					"/attempts/:attempt_id",
					quizAttemptHandler.GetTeacherAttempt,
				)
				teacher.PATCH(
					"/attempts/:attempt_id/answers/:answer_id/grade",
					quizAttemptHandler.GradeEssay,
				)
				// ================================================

				// ===== ROUTE ATTENDANCE UNTUK GURU =====
				teacher.POST(
					"/attendance-sessions",
					attendanceHandler.CreateSession,
				)
				teacher.GET(
					"/attendance-sessions",
					attendanceHandler.GetSessions,
				)
				teacher.GET(
					"/attendance-sessions/:id",
					attendanceHandler.GetSession,
				)
				teacher.GET(
					"/attendance-sessions/:id/records",
					attendanceHandler.GetAttendanceRecords,
				)
				teacher.DELETE(
					"/attendance-sessions/:id",
					attendanceHandler.DeleteSession,
				)
				teacher.GET(
					"/attendance-recap",
					attendanceHandler.GetAttendanceRecap,
				)
				teacher.GET(
					"/attendance-statistics",
					attendanceHandler.GetAttendanceStatistics,
				)
				teacher.POST(
					"/attendance-sessions/:id/close",
					attendanceHandler.CloseSession,
				)
				teacher.POST(
					"/attendance-sessions/:id/open",
					attendanceHandler.OpenSession,
				)
				teacher.PATCH(
					"/attendance-records/:record_id",
					attendanceHandler.UpdateAttendanceRecord,
				)
				// =======================================
			}

			// Route Siswa
			siswaRoutes := protected.Group("/siswa")
			siswaRoutes.Use(middleware.RequireRole("siswa"))
			{
				siswaRoutes.GET("/dashboard", func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{
						"success": true,
						"message": "Selamat datang di Siswa Dashboard",
					})
				})
			}

			student := protected.Group("/student")
			student.Use(middleware.RequireRole("siswa"))
			{
				student.GET(
					"/materials",
					materialHandler.GetStudentMaterials,
				)
				student.GET(
					"/materials/:id",
					materialHandler.GetStudentMaterialByID,
				)
				student.GET(
					"/material-files/:file_id/download",
					materialFileHandler.DownloadStudentFile,
				)
				student.GET(
					"/assignment-files/:file_id/download",
					assignmentFileHandler.DownloadStudentFile,
				)
				student.DELETE(
					"/submission-files/:file_id",
					assignmentSubmissionHandler.DeleteStudentSubmissionFile,
				)

				// Rute Tugas Siswa
				student.GET(
					"/assignments",
					assignmentHandler.GetStudentAssignments,
				)
				student.GET(
					"/assignments/:id",
					assignmentHandler.GetStudentAssignment,
				)

				// Rute Pengumpulan Tugas (Submission)
				student.POST(
					"/assignments/:id/submit",
					assignmentSubmissionHandler.Submit,
				)
				student.GET(
					"/assignments/:id/submission",
					assignmentSubmissionHandler.GetStudentSubmission,
				)

				// ===== TAMBAHAN ROUTE QUIZ SISWA =====
				student.GET(
					"/quizzes",
					quizHandler.GetStudentQuizzes,
				)
				student.GET(
					"/quizzes/:id",
					quizHandler.GetStudentQuiz,
				)
				// =====================================
			}

			// ===== ROUTE NOTIFICATIONS =====
			notificationRoutes := protected.Group("/notifications")
			{
				notificationRoutes.GET(
					"",
					notificationHandler.GetNotifications,
				)
				notificationRoutes.GET(
					"/unread-count",
					notificationHandler.GetUnreadCount,
				)
				notificationRoutes.PATCH(
					"/:id/read",
					notificationHandler.MarkAsRead,
				)
				notificationRoutes.PATCH(
					"/read-all",
					notificationHandler.MarkAllAsRead,
				)
			}
			// ================================
		}

		// ===== ROUTE CHAT =====
		chatRoutes := api.Group("/chat")
		chatRoutes.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			chatRoutes.GET("/faqs", helpCenterHandler.List)
			chatRoutes.POST("/customer-service", chatHandler.CreateCustomerService)
			chatRoutes.GET("/customer-service", chatHandler.ListCustomerService)
			chatRoutes.GET("/customer-service/:id/messages", chatHandler.GetMessages)
			chatRoutes.POST("/customer-service/:id/messages", chatHandler.SendMessage)
			chatRoutes.GET(
				"/contacts",
				chatHandler.GetChatContacts,
			)
			chatRoutes.GET("/classes", chatHandler.GetTeacherClasses)
			chatRoutes.POST(
				"/conversations",
				chatHandler.CreateConversation,
			)
			chatRoutes.GET(
				"/conversations",
				chatHandler.GetConversations,
			)
			chatRoutes.PUT("/conversations/:id", chatHandler.UpdateConversation)
			chatRoutes.DELETE("/conversations/:id", chatHandler.DeleteConversation)
			chatRoutes.POST("/conversations/:id/members", chatHandler.InviteStudent)
			chatRoutes.POST(
				"/conversations/:id/messages",
				chatHandler.SendMessage,
			)
			chatRoutes.GET(
				"/conversations/:id/messages",
				chatHandler.GetMessages,
			)
			chatRoutes.GET(
				"/presence/:user_id",
				chatHandler.GetUserPresence,
			)
			chatRoutes.GET(
				"/unread-counts",
				chatHandler.GetUnreadCounts,
			)
		}

		meetingRoutes := api.Group("/meetings")
		meetingRoutes.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			meetingRoutes.GET("", meetingHandler.List)
			meetingRoutes.GET("/:id", meetingHandler.Get)
			meetingRoutes.POST("/:id/request", middleware.RequireRole("siswa"), meetingHandler.Request)
			meetingRoutes.POST("/:id/join", middleware.RequireRole("guru", "siswa"), meetingHandler.Join)
			meetingRoutes.POST("/:id/leave", middleware.RequireRole("guru", "siswa"), meetingHandler.Leave)
			teacherMeetings := meetingRoutes.Group("")
			teacherMeetings.Use(middleware.RequireRole("guru"))
			{
				teacherMeetings.POST("", meetingHandler.Create)
				teacherMeetings.PUT("/:id", meetingHandler.Update)
				teacherMeetings.DELETE("/:id", meetingHandler.Delete)
				teacherMeetings.POST("/:id/start", meetingHandler.Start)
				teacherMeetings.POST("/:id/end", meetingHandler.End)
				teacherMeetings.GET("/:id/requests", meetingHandler.Requests)
				teacherMeetings.PATCH("/:id/requests/decide-all", meetingHandler.DecideAll)
				teacherMeetings.PATCH("/:id/participants/:user_id", meetingHandler.Decide)
				teacherMeetings.PATCH("/:id/participants/:user_id/permission", meetingHandler.Permission)
			}
		}
		// ========================
	}

	// ===== ROUTE GROUP KHUSUS STUDENT DENGAN MIDDLEWARE =====
	studentRoutes := router.Group("/api/student")
	studentRoutes.Use(
		middleware.AuthMiddleware(cfg.JWTSecret),
		middleware.RequireRole("siswa"),
	)
	{
		// START QUIZ
		studentRoutes.POST(
			"/quizzes/:id/start",
			quizAttemptHandler.Start,
		)
		studentRoutes.GET(
			"/quizzes/:id/attempts/latest",
			quizAttemptHandler.GetLatestStudentAttempt,
		)

		// SAVE ANSWER
		studentRoutes.PUT(
			"/attempts/:attempt_id/answers/:question_id",
			quizAnswerHandler.SaveAnswer,
		)

		// SUBMIT QUIZ
		studentRoutes.POST(
			"/attempts/:attempt_id/submit",
			quizAttemptHandler.Submit,
		)

		// GET QUIZ RESULT
		studentRoutes.GET(
			"/attempts/:attempt_id/result",
			quizAttemptHandler.GetStudentQuizResult,
		)

		// ===== ROUTE ATTENDANCE UNTUK SISWA =====
		studentRoutes.GET(
			"/attendance-sessions",
			attendanceHandler.GetStudentSessions,
		)
		studentRoutes.GET(
			"/attendance-sessions/:id",
			attendanceHandler.GetStudentSession,
		)
		studentRoutes.POST(
			"/attendance-sessions/:id/check-in",
			attendanceHandler.CheckIn,
		)
		// ========================================
	}
	// ========================================================

	fmt.Println("Server running on http://localhost:" + cfg.AppPort)

	err := router.Run(":" + cfg.AppPort)
	if err != nil {
		panic(err)
	}
}
