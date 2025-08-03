package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken  = errors.New("invalid token")
	ErrExpiredToken  = errors.New("token expired")
	ErrInvalidClaims = errors.New("invalid token claims")
)

// Claims represents the JWT claims structure
type Claims struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	jwt.RegisteredClaims
}

// Config holds JWT configuration
type Config struct {
	SecretKey     string        `json:"secret_key"`
	AccessExpiry  time.Duration `json:"access_expiry"`
	RefreshExpiry time.Duration `json:"refresh_expiry"`
}

// DefaultConfig returns default JWT configuration
func DefaultConfig() *Config {
	return &Config{
		SecretKey:     "your-secret-key-here-change-in-production",
		AccessExpiry:  15 * time.Minute,   // 15 minutes
		RefreshExpiry: 7 * 24 * time.Hour, // 7 days
	}
}

// Service provides JWT token operations
type Service struct {
	config *Config
}

// NewService creates a new JWT service
func NewService(config *Config) *Service {
	if config == nil {
		config = DefaultConfig()
	}
	return &Service{config: config}
}

// GenerateAccessToken generates an access token for the given user
func (s *Service) GenerateAccessToken(userID uuid.UUID, username, email string) (string, error) {
	claims := &Claims{
		UserID:   userID,
		Username: username,
		Email:    email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.config.AccessExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "go-next-api",
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.SecretKey))
}

// GenerateRefreshToken generates a refresh token for the given user
func (s *Service) GenerateRefreshToken(userID uuid.UUID) (string, error) {
	claims := &Claims{
		UserID:   userID,
		Username: "",
		Email:    "",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.config.RefreshExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "go-next-api",
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.SecretKey))
}

// ValidateToken validates and parses a JWT token
func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.SecretKey), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidClaims
}

// ExtractUserID extracts user ID from a valid token
func (s *Service) ExtractUserID(tokenString string) (uuid.UUID, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return uuid.Nil, err
	}
	return claims.UserID, nil
}

// IsTokenExpired checks if a token is expired
func (s *Service) IsTokenExpired(tokenString string) bool {
	_, err := s.ValidateToken(tokenString)
	return errors.Is(err, ErrExpiredToken)
}
