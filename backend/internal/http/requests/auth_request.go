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
