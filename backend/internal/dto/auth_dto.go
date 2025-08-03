package dto

// AuthDTO represents the authentication response data
type AuthDTO struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}
