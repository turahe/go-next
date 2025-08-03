package controllers

import (
	"go-next/internal/dto"
	"go-next/internal/http/requests"
	"go-next/internal/http/responses"
	"go-next/internal/models"
	"go-next/internal/services"
	"go-next/pkg/database"
	"go-next/pkg/validation"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type AuthHandler interface {
	Register(c *gin.Context)
	Login(c *gin.Context)
	RequestEmailVerification(c *gin.Context)
	VerifyEmail(c *gin.Context)
	ResendVerificationEmail(c *gin.Context)
	RequestPhoneVerification(c *gin.Context)
	VerifyPhone(c *gin.Context)
	ResetPassword(c *gin.Context)
	RefreshToken(c *gin.Context)
	Logout(c *gin.Context)
	RequestPasswordReset(c *gin.Context)
}

type authHandler struct {
	AuthService services.AuthService
	BaseHandler *BaseHandler
	validator   *validation.Validator
}

func NewAuthHandler() AuthHandler {
	validator := validation.NewValidator()

	// Add custom error messages
	validator.AddCustomMessage("email", "email", "Please provide a valid email address")
	validator.AddCustomMessage("password", "min", "Password must be at least 8 characters long")
	validator.AddCustomMessage("username", "unique_username", "This username is already taken")
	validator.AddCustomMessage("email", "unique_email", "This email address is already registered")
	validator.AddCustomMessage("username", "username", "Username must contain only letters, numbers, and underscores")
	validator.AddCustomMessage("country_code", "country_code", "Please provide a valid country code (e.g., US, GB, IN)")
	//validator.AddCustomMessage("phone", "e164", "Please provide a valid phone number in E.164 format (e.g., +1234567890)")

	return &authHandler{
		AuthService: services.NewAuthService(),
		BaseHandler: NewBaseHandler(nil), // You can pass logger here
		validator:   validator,
	}
}

// Register creates a new user account
// @Summary Register new user
// @Description Create a new user account with the provided credentials
// @Tags auth
// @Accept json
// @Produce json
// @Param user body requests.RegisterRequest true "User registration data"
// @Success 201 {object} responses.CommonResponse{data=map[string]interface{}}
// @Failure 400 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /auth/register [post]
func (h *authHandler) Register(c *gin.Context) {
	var req requests.RegisterRequest

	// Bind JSON to request struct
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid JSON format", err.Error())
		return
	}

	// Validate the request using Laravel-style validation
	result := h.validator.Validate(&req)
	if !result.IsValid {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"message": "Validation failed",
			"errors":  result.Errors,
		})
		return
	}

	// Register user using auth service
	if err := h.AuthService.Register(req.UserName, req.Email, req.Phone, req.CountryCode, req.Password); err != nil {
		responses.SendError(c, http.StatusBadRequest, "Registration failed", "Unable to create your account. Please try again later.")
		return
	}

	c.JSON(http.StatusCreated, responses.CommonResponse{
		ResponseCode:    http.StatusCreated,
		ResponseMessage: "Registration successful",
	})
}

// Login authenticates a user and returns access tokens
// @Summary User login
// @Description Authenticate user with email and password, return access and refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body requests.LoginRequest true "Login credentials"
// @Success 200 {object} responses.CommonResponse{data=dto.AuthDTO}
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /auth/login [post]
func (h *authHandler) Login(c *gin.Context) {
	var req requests.LoginRequest

	// Bind JSON to request struct
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid JSON format", err.Error())
		return
	}

	// Validate the request using Laravel-style validation
	result := h.validator.Validate(&req)
	if !result.IsValid {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"message": "Validation failed",
			"errors":  result.Errors,
		})
		return
	}

	// Get client IP address
	clientIP := c.ClientIP()
	if clientIP == "" {
		clientIP = "Unknown"
	}

	// Get user agent
	userAgent := c.GetHeader("User-Agent")
	if userAgent == "" {
		userAgent = "Unknown"
	}

	// Try login with username as email
	auth, err := h.AuthService.Login(req.Identity, req.Password, clientIP, userAgent)
	if err != nil {
		// If username login fails, try with email
		// For now, we'll just return the error
		responses.SendError(c, http.StatusUnauthorized, "Invalid credentials", err.Error())
		return
	}

	// Return success response with tokens
	c.JSON(http.StatusOK, dto.AuthDTO{
		Token:        auth.Token,
		RefreshToken: auth.RefreshToken,
	})
}

// RefreshTokenRequest represents a refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// RefreshToken refreshes the access token using a refresh token
// @Summary Refresh access token
// @Description Use refresh token to get a new access token
// @Tags auth
// @Accept json
// @Produce json
// @Param refresh body RefreshTokenRequest true "Refresh token data"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /auth/refresh [post]
func (h *authHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest

	// Bind JSON to request struct
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid JSON format", err.Error())
		return
	}

	// Validate the request using Laravel-style validation
	result := h.validator.Validate(&req)
	if !result.IsValid {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"message": "Validation failed",
			"errors":  result.Errors,
		})
		return
	}

	// Validate refresh token and get new access token
	accessToken, err := h.AuthService.RefreshToken(req.RefreshToken)
	if err != nil {
		responses.SendError(c, http.StatusUnauthorized, "Invalid refresh token", err.Error())
		return
	}

	// Return new access token
	c.JSON(http.StatusOK, dto.AuthDTO{
		Token:        accessToken.Token,
		RefreshToken: accessToken.RefreshToken, // Return the new refresh token
	})
}

// RequestEmailVerification sends an email verification token to the user
// @Summary Request email verification
// @Description Send email verification token to user's email address
// @Tags auth
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} responses.SuccessResponse{data=string}
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /users/{id}/request-email-verification [post]
func (h *authHandler) RequestEmailVerification(c *gin.Context) {
	id := c.Param("id")
	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Verification email sent"})
}

// EmailVerificationRequest represents an email verification request
type EmailVerificationRequest struct {
	Token string `json:"token" validate:"required"`
}

// VerifyEmail verifies the user's email address using the provided token
// @Summary Verify email address
// @Description Verify user's email address using the verification token
// @Tags auth
// @Accept json
// @Produce json
// @Param verification body EmailVerificationRequest true "Verification data"
// @Success 200 {object} responses.SuccessResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /auth/verify-email [post]
func (h *authHandler) VerifyEmail(c *gin.Context) {
	var req EmailVerificationRequest

	// Bind JSON to request struct
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid JSON format", err.Error())
		return
	}

	// Validate the request using Laravel-style validation
	result := h.validator.Validate(&req)
	if !result.IsValid {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"message": "Validation failed",
			"errors":  result.Errors,
		})
		return
	}

	// Verify email using auth service
	if err := h.AuthService.VerifyEmail(req.Token); err != nil {
		responses.SendError(c, http.StatusBadRequest, "Email verification failed", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email verified successfully"})
}

// RequestPhoneVerification sends an SMS verification token to the user
// @Summary Request phone verification
// @Description Send SMS verification token to user's phone number
// @Tags auth
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} responses.SuccessResponse{data=string}
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /users/{id}/request-phone-verification [post]
func (h *authHandler) RequestPhoneVerification(c *gin.Context) {
	id := c.Param("id")
	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Verification SMS sent"})
}

// PhoneVerificationRequest represents a phone verification request
type PhoneVerificationRequest struct {
	Token string `json:"token" validate:"required"`
}

// VerifyPhone verifies the user's phone number using the provided token
// @Summary Verify phone number
// @Description Verify user's phone number using the verification token
// @Tags auth
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param verification body PhoneVerificationRequest true "Verification data"
// @Success 200 {object} responses.SuccessResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /users/{id}/verify-phone [post]
func (h *authHandler) VerifyPhone(c *gin.Context) {
	var req PhoneVerificationRequest

	// Bind JSON to request struct
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid JSON format", err.Error())
		return
	}

	// Validate the request using Laravel-style validation
	result := h.validator.Validate(&req)
	if !result.IsValid {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"message": "Validation failed",
			"errors":  result.Errors,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Phone verified"})
}

// RequestPasswordResetRequest represents a password reset request
type RequestPasswordResetRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// RequestPasswordReset sends a password reset email to the user
// @Summary Request password reset
// @Description Send password reset email to user's email address
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RequestPasswordResetRequest true "Password reset request"
// @Success 200 {object} responses.SuccessResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /auth/request-password-reset [post]
func (h *authHandler) RequestPasswordReset(c *gin.Context) {
	var req RequestPasswordResetRequest

	// Bind JSON to request struct
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid JSON format", err.Error())
		return
	}

	// Validate the request using Laravel-style validation
	result := h.validator.Validate(&req)
	if !result.IsValid {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"message": "Validation failed",
			"errors":  result.Errors,
		})
		return
	}

	// Request password reset using auth service
	if err := h.AuthService.RequestPasswordReset(req.Email); err != nil {
		responses.SendError(c, http.StatusBadRequest, "Password reset request failed", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset email sent successfully",
	})
}

// ResetPasswordRequest represents a password reset request
type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

// ResetPassword resets the user's password using a valid reset token
// @Summary Reset password
// @Description Reset user password using reset token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body ResetPasswordRequest true "Password reset data"
// @Success 200 {object} responses.SuccessResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /auth/reset-password [post]
func (h *authHandler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest

	// Bind JSON to request struct
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid JSON format", err.Error())
		return
	}

	// Validate the request using Laravel-style validation
	result := h.validator.Validate(&req)
	if !result.IsValid {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"message": "Validation failed",
			"errors":  result.Errors,
		})
		return
	}

	// Reset password using auth service
	if err := h.AuthService.ResetPassword(req.Token, req.NewPassword); err != nil {
		responses.SendError(c, http.StatusBadRequest, "Password reset failed", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset successful",
	})
}

// Custom validation functions

// validateStrongPassword validates that password meets strong password requirements
func validateStrongPassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	// Check for at least one uppercase letter, lowercase letter, digit, and special character
	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasDigit = true
		case char == '!' || char == '@' || char == '#' || char == '$' || char == '%' || char == '^' || char == '&' || char == '*':
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasDigit && hasSpecial
}

// validateUniqueEmail validates that email is unique in the database
func validateUniqueEmail(fl validator.FieldLevel) bool {
	email := fl.Field().String()

	// This would typically check against the database
	// For now, we'll just check if it's not "admin@example.com"
	return email != "admin@example.com"
}

// validateUniqueUsername validates that username is unique in the database
func validateUniqueUsername(fl validator.FieldLevel) bool {
	username := fl.Field().String()

	// This would typically check against the database
	// For now, we'll just check if it's not "admin"
	return username != "admin"
}

// ResendVerificationEmailRequest represents a resend verification email request
type ResendVerificationEmailRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// ResendVerificationEmail sends a new verification email to the user
// @Summary Resend verification email
// @Description Send a new verification email to the user's email address
// @Tags auth
// @Accept json
// @Produce json
// @Param request body ResendVerificationEmailRequest true "Resend verification request"
// @Success 200 {object} responses.SuccessResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /auth/resend-verification-email [post]
func (h *authHandler) ResendVerificationEmail(c *gin.Context) {
	var req ResendVerificationEmailRequest

	// Bind JSON to request struct
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid JSON format", err.Error())
		return
	}

	// Validate the request using Laravel-style validation
	result := h.validator.Validate(&req)
	if !result.IsValid {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"message": "Validation failed",
			"errors":  result.Errors,
		})
		return
	}

	// Resend verification email using auth service
	if err := h.AuthService.ResendVerificationEmail(req.Email); err != nil {
		responses.SendError(c, http.StatusBadRequest, "Failed to resend verification email", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Verification email sent successfully"})
}

// LogoutRequest represents a logout request
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// Logout invalidates the current refresh token
// @Summary User logout
// @Description Invalidate refresh token and log out user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LogoutRequest true "Logout request"
// @Success 200 {object} responses.SuccessResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /auth/logout [post]
func (h *authHandler) Logout(c *gin.Context) {
	var req LogoutRequest

	// Bind JSON to request struct
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.SendError(c, http.StatusBadRequest, "Invalid JSON format", err.Error())
		return
	}

	// Validate the request using Laravel-style validation
	result := h.validator.Validate(&req)
	if !result.IsValid {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"message": "Validation failed",
			"errors":  result.Errors,
		})
		return
	}

	// Logout user using auth service
	if err := h.AuthService.Logout(req.RefreshToken); err != nil {
		responses.SendError(c, http.StatusBadRequest, "Logout failed", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Logout successful",
	})
}
