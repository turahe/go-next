package dto

import (
	"go-next/internal/models"
	"time"

	"github.com/google/uuid"
)

type UserDTO struct {
	ID            uuid.UUID  `json:"id"`
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
	var phone *string
	if u.Phone != "" {
		phone = &u.Phone
	}
	return &UserDTO{
		ID:            u.ID,
		Username:      u.Username,
		Email:         u.Email,
		Phone:         phone,
		EmailVerified: u.EmailVerified,
		PhoneVerified: u.PhoneVerified,
		IsActive:      u.IsActive,
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
		Roles:         roles,
	}
}
