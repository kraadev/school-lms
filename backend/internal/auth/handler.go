package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"school-lms/internal/models"
)

type Handler struct {
	Service *Service
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *Handler) Login(c *gin.Context) {
	var request LoginRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "username dan password wajib diisi",
		})
		return
	}

	result, err := h.Service.Login(
		request.Username,
		request.Password,
	)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "login berhasil",
		"data":    result,
	})
}

func (h *Handler) Me(c *gin.Context) {
	userID := c.GetString("user_id")

	id, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "Invalid user ID",
		})
		return
	}

	var user models.User

	err = h.Service.DB.
		Preload("Role").
		Preload("Student").
		Preload("Teacher").
		First(&user, "id = ?", id).Error

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "User not found",
		})
		return
	}

	displayName := user.Username
	if user.Student != nil && user.Student.FullName != "" {
		displayName = user.Student.FullName
	} else if user.Teacher != nil && user.Teacher.FullName != "" {
		displayName = user.Teacher.FullName
	}
	response := UserResponse{
		ID:          user.ID,
		Username:    user.Username,
		DisplayName: displayName,
		Email:       user.Email,
		IsActive:    user.IsActive,
		Role: RoleResponse{
			ID:   user.Role.ID,
			Name: user.Role.Name,
		},
	}
	if user.Student != nil {
		response.Student = &struct {
			FullName string `json:"full_name"`
		}{FullName: user.Student.FullName}
	}
	if user.Teacher != nil {
		response.Teacher = &struct {
			FullName string `json:"full_name"`
		}{FullName: user.Teacher.FullName}
	}

	c.JSON(http.StatusOK, gin.H{
		"user": response,
	})
}
