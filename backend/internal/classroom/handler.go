package classroom

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		Service: service,
	}
}

type CreateClassRequest struct {
	Name           string  `json:"name" binding:"required"`
	Grade          int16   `json:"grade" binding:"required"`
	MajorID        *string `json:"major_id"`
	AcademicYearID string  `json:"academic_year_id" binding:"required"`
}

type UpdateClassRequest = CreateClassRequest

type AddStudentRequest struct {
	StudentID string `json:"student_id" binding:"required"`
}

type AddSubjectRequest struct {
	SubjectID string `json:"subject_id" binding:"required"`
	TeacherID string `json:"teacher_id" binding:"required"`
}

type AssignTeacherRequest struct {
	TeacherID string `json:"teacher_id" binding:"required"`
}

func (h *Handler) GetAll(c *gin.Context) {
	classes, err := h.Service.GetAll()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Gagal mengambil data kelas",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    classes,
	})
}

func (h *Handler) Create(c *gin.Context) {
	var request CreateClassRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Data kelas tidak valid",
			"error":   err.Error(),
		})
		return
	}

	class, err := h.Service.Create(
		request.Name,
		request.Grade,
		request.MajorID,
		request.AcademicYearID,
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Kelas berhasil dibuat",
		"data":    class,
	})
}

func (h *Handler) GetByID(c *gin.Context) {
	class, err := h.Service.GetByID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Kelas tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": class})
}

func (h *Handler) Update(c *gin.Context) {
	var request UpdateClassRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Data kelas tidak valid"})
		return
	}
	class, err := h.Service.Update(c.Param("id"), request.Name, request.Grade, request.MajorID, request.AcademicYearID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Kelas berhasil diperbarui", "data": class})
}

func (h *Handler) Delete(c *gin.Context) {
	if err := h.Service.Delete(c.Param("id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Kelas berhasil dihapus"})
}

func (h *Handler) GetStudents(c *gin.Context) {
	classID := c.Param("class_id")

	students, err := h.Service.GetStudents(classID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    students,
	})
}

func (h *Handler) AddStudent(c *gin.Context) {
	classID := c.Param("class_id")

	var request AddStudentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Data siswa tidak valid",
			"error":   err.Error(),
		})
		return
	}

	err := h.Service.AddStudent(classID, request.StudentID)
	if err != nil {
		var conflict *StudentClassConflictError
		if errors.As(err, &conflict) {
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"message": err.Error(),
				"details": gin.H{
					"reason":     "Satu siswa hanya boleh terdaftar pada satu kelas.",
					"class_name": conflict.ClassName,
					"next_steps": []string{
						"Gunakan siswa tersebut pada kelas yang sudah tertera.",
						"Keluarkan siswa dari kelas lama jika ingin memindahkannya.",
					},
				},
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Siswa berhasil ditambahkan ke kelas",
	})
}

func (h *Handler) RemoveStudent(c *gin.Context) {
	classID := c.Param("class_id")
	studentID := c.Param("student_id")

	err := h.Service.RemoveStudent(classID, studentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Siswa berhasil dikeluarkan dari kelas",
	})
}

func (h *Handler) GetSubjects(c *gin.Context) {
	classID := c.Param("class_id")

	subjects, err := h.Service.GetSubjects(classID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    subjects,
	})
}

func (h *Handler) GetTeacherClassSubjects(c *gin.Context) {
	items, err := h.Service.GetTeacherClassSubjects(c.GetString("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": items})
}

func (h *Handler) AddSubject(c *gin.Context) {
	classID := c.Param("class_id")

	var request AddSubjectRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Data mata pelajaran tidak valid",
			"error":   err.Error(),
		})
		return
	}

	err := h.Service.AddSubject(classID, request.SubjectID, request.TeacherID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Mata pelajaran berhasil ditambahkan ke kelas",
	})
}

func (h *Handler) RemoveSubject(c *gin.Context) {
	classID := c.Param("class_id")
	subjectID := c.Param("subject_id")

	err := h.Service.RemoveSubject(classID, subjectID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Mata pelajaran berhasil dikeluarkan dari kelas",
	})
}

func (h *Handler) AssignTeacher(c *gin.Context) {
	classID := c.Param("class_id")
	subjectID := c.Param("subject_id")

	var request AssignTeacherRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Data guru tidak valid",
			"error":   err.Error(),
		})
		return
	}

	err := h.Service.AssignTeacher(classID, subjectID, request.TeacherID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Guru berhasil ditugaskan",
	})
}

func (h *Handler) GetSubjectDetails(c *gin.Context) {
	classID := c.Param("class_id")

	classSubjects, err := h.Service.GetSubjectDetails(classID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    classSubjects,
	})
}
