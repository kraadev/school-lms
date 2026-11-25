package auth

import "github.com/google/uuid"

type RoleResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type UserResponse struct {
	ID          uuid.UUID    `json:"id"`
	Username    string       `json:"username"`
	DisplayName string       `json:"display_name,omitempty"`
	Email       *string      `json:"email,omitempty"`
	IsActive    bool         `json:"is_active"`
	Role        RoleResponse `json:"role"`
	Student     *struct {
		FullName string `json:"full_name"`
	} `json:"student,omitempty"`
	Teacher *struct {
		FullName string `json:"full_name"`
	} `json:"teacher,omitempty"`
}
