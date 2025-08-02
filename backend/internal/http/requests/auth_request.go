package requests

// EmailVerificationRequest represents the request body for verifying email
// swagger:model
type EmailVerificationRequest struct {
	Token string `json:"token"`
}

// PhoneVerificationRequest represents the request body for verifying phone
// swagger:model
type PhoneVerificationRequest struct {
	Token string `json:"token"`
}

// RefreshTokenRequest represents the request body for refreshing JWT
// swagger:model
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type AuthRequest struct {
	Username string `json:"username" validate:"required,min=3,max=32"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Role     string `json:"role,omitempty" validate:"omitempty,oneof=admin editor moderator user guest"`
}

type LoginRequest struct {
	Username string `json:"username" validate:"required,min=3,max=32"`
	Password string `json:"password" validate:"required,min=6"`
}
