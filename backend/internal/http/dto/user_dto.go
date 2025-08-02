package dto

import (
	"time"
	"wordpress-go-next/backend/internal/models"
)

type UserDTO struct {
	ID            uint64     `json:"id"`
	Username      string     `json:"username"`
	Email         string     `json:"email"`
	Phone         *string    `json:"phone,omitempty"`
	EmailVerified *time.Time `json:"email_verified_at,omitempty"`
	PhoneVerified *time.Time `json:"phone_verified_at,omitempty"`
	IsActive      bool       `json:"is_active"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	Roles         []string   `json:"roles"`
}

func ToUserDTO(u *models.User) *UserDTO {
	roles := make([]string, len(u.Roles))
	for i, r := range u.Roles {
		roles[i] = r.Name
	}
	return &UserDTO{
		ID:            u.ID,
		Username:      u.Username,
		Email:         u.Email,
		Phone:         u.Phone,
		EmailVerified: u.EmailVerified,
		PhoneVerified: u.PhoneVerified,
		IsActive:      u.IsActive,
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
		Roles:         roles,
	}
}
