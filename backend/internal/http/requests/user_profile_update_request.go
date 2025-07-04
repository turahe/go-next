package requests

import "time"

type UserProfileUpdateInput struct {
	Username      string     `json:"username" validate:"required,min=3,max=32"`
	Email         string     `json:"email" validate:"required,email"`
	Phone         string     `json:"phone" validate:"omitempty,e164"`
	EmailVerified *time.Time `json:"email_verified"`
	PhoneVerified *time.Time `json:"phone_verified"`
}
