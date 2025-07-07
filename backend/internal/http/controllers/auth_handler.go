package controllers

import (
	"net/http"
	"os"
	"time"
	"wordpress-go-next/backend/internal/http/requests"
	"wordpress-go-next/backend/internal/http/responses"
	"wordpress-go-next/backend/internal/models"
	"wordpress-go-next/backend/internal/services"
	"wordpress-go-next/backend/pkg/config"
	"wordpress-go-next/backend/pkg/email"
	"wordpress-go-next/backend/pkg/whatsapp"

	"fmt"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	AuthService services.AuthService
}

func NewAuthHandler(authService services.AuthService) *AuthHandler {
	return &AuthHandler{AuthService: authService}
}

// Register user (real implementation)
func (h *AuthHandler) Register(c *gin.Context) {
	var req requests.AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, responses.CommonResponse{
			ResponseCode:    http.StatusBadRequest,
			ResponseMessage: "Invalid request",
		})
		return
	}
	// Check if user already exists
	if user, _ := services.UserSvc.GetUserByUsername(c.Request.Context(), req.Username); user != nil {
		c.JSON(http.StatusConflict, responses.CommonResponse{
			ResponseCode:    http.StatusConflict,
			ResponseMessage: "Username already exists",
		})
		return
	}
	if user, _ := services.UserSvc.GetUserByEmail(c.Request.Context(), req.Email); user != nil {
		c.JSON(http.StatusConflict, responses.CommonResponse{
			ResponseCode:    http.StatusConflict,
			ResponseMessage: "Email already exists",
		})
		return
	}
	// Hash password
	hash, err := h.AuthService.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.CommonResponse{
			ResponseCode:    http.StatusInternalServerError,
			ResponseMessage: "Failed to hash password",
		})
		return
	}
	user := &models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hash,
		IsActive:     true,
	}
	if err := services.UserSvc.CreateUser(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, responses.CommonResponse{
			ResponseCode:    http.StatusInternalServerError,
			ResponseMessage: "Failed to create user",
		})
		return
	}
	// Assign role if provided
	if req.Role != "" {
		role, err := services.RoleSvc.GetRoleByName(c.Request.Context(), req.Role)
		if err == nil && role != nil {
			_ = services.UserRoleSvc.AssignRoleToUser(c.Request.Context(), user, role)
		}
	}
	// Send email verification
	verificationToken, err := h.AuthService.CreateVerificationToken(c.Request.Context(), user.ID, "email")
	if err == nil && verificationToken != "" {
		cfg := config.GetConfig()
		emailSvc := email.NewEmailService(cfg.SMTP)
		verifyBaseURL := os.Getenv("EMAIL_BASE_URL")
		if verifyBaseURL == "" {
			verifyBaseURL = "http://localhost:8080" // fallback for dev
		}
		verifyURL := fmt.Sprintf("%s/verify-email?token=%s", verifyBaseURL, verificationToken)
		body := email.EmailVerificationTemplate(user.Username, verifyURL)
		_ = emailSvc.SendEmail(user.Email, "Verify your email", body)
	}
	// Send phone verification if phone is provided
	phoneVerificationSent := false
	if user.Phone != nil && *user.Phone != "" {
		phoneToken, err := h.AuthService.CreateVerificationToken(c.Request.Context(), user.ID, "phone")
		if err == nil && phoneToken != "" {
			chatId := *user.Phone + "@c.us"
			waCfg := config.GetConfig().WhatsApp
			wa := whatsapp.NewWhatsAppService(waCfg.BaseURL, waCfg.Session)
			_ = wa.StartTyping(chatId)
			err = wa.SendText(chatId, fmt.Sprintf("Your verification code is: %s", phoneToken), nil, false, false)
			_ = wa.StopTyping(chatId)

			if err == nil {
				phoneVerificationSent = true
			}
		}
	}
	// Generate JWT and refresh token
	accessToken, refreshToken, err := h.AuthService.GenerateTokens(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.CommonResponse{
			ResponseCode:    http.StatusInternalServerError,
			ResponseMessage: "Failed to generate tokens",
		})
		return
	}
	c.JSON(http.StatusCreated, responses.CommonResponse{
		ResponseCode:    http.StatusCreated,
		ResponseMessage: "Registration successful. Please verify your email and phone.",
		Data: map[string]interface{}{
			"access_token":            accessToken,
			"refresh_token":           refreshToken,
			"expires_in":              time.Now().Add(15 * time.Minute).Unix(),
			"email_verification_sent": true,
			"phone_verification_sent": phoneVerificationSent,
		},
	})
}

// Login and issue JWT + refresh token
func (h *AuthHandler) Login(c *gin.Context) {
	var req requests.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	user, err := h.AuthService.AuthenticateUser(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}
	// Generate JWT and refresh token
	accessToken, refreshToken, err := h.AuthService.GenerateTokens(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_in":    time.Now().Add(15 * time.Minute).Unix(),
	})
}

// Request password reset (stub)
func (h *AuthHandler) RequestPasswordReset(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "RequestPasswordReset not implemented"})
}

// Reset password (stub)
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "ResetPassword not implemented"})
}

// Refresh JWT using refresh token
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	accessToken, refreshToken, err := h.AuthService.RefreshTokens(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_in":    time.Now().Add(15 * time.Minute).Unix(),
	})
}

// Email/phone verification stubs
func (h *AuthHandler) RequestEmailVerification(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, responses.CommonResponse{
			ResponseCode:    http.StatusUnauthorized,
			ResponseMessage: "Unauthorized",
		})
		return
	}
	user, err := services.UserSvc.GetUserByID(c.Request.Context(), fmt.Sprintf("%v", userID))
	if err != nil || user == nil || user.Email == "" {
		c.JSON(http.StatusNotFound, responses.CommonResponse{
			ResponseCode:    http.StatusNotFound,
			ResponseMessage: "User or email not found",
		})
		return
	}
	// Rate limiting: 5 per minute per user
	rlKey := fmt.Sprintf("email_verify_rl:%v", userID)
	rateLimited, _ := h.AuthService.IsRateLimited(c.Request.Context(), rlKey, 5, time.Minute)
	if rateLimited {
		c.JSON(http.StatusTooManyRequests, responses.CommonResponse{
			ResponseCode:    http.StatusTooManyRequests,
			ResponseMessage: "Too many requests. Please wait before requesting another code.",
		})
		return
	}
	// Create verification token
	token, err := h.AuthService.CreateVerificationToken(c.Request.Context(), user.ID, "email")
	if err != nil || token == "" {
		c.JSON(http.StatusInternalServerError, responses.CommonResponse{
			ResponseCode:    http.StatusInternalServerError,
			ResponseMessage: "Failed to create email verification token",
		})
		return
	}
	cfg := config.GetConfig()
	emailSvc := email.NewEmailService(cfg.SMTP)
	verifyBaseURL := os.Getenv("EMAIL_BASE_URL")
	if verifyBaseURL == "" {
		verifyBaseURL = "http://localhost:8080" // fallback for dev
	}
	verifyURL := fmt.Sprintf("%s/verify-email?token=%s", verifyBaseURL, token)
	body := email.EmailVerificationTemplate(user.Username, verifyURL)
	err = emailSvc.SendEmail(user.Email, "Verify your email", body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.CommonResponse{
			ResponseCode:    http.StatusInternalServerError,
			ResponseMessage: "Failed to send verification email",
		})
		return
	}
	c.JSON(http.StatusOK, responses.CommonResponse{
		ResponseCode:    http.StatusOK,
		ResponseMessage: "Verification email sent",
	})
}

func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req struct {
		Token string `json:"token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Token == "" {
		c.JSON(http.StatusBadRequest, responses.CommonResponse{
			ResponseCode:    http.StatusBadRequest,
			ResponseMessage: "Token required",
		})
		return
	}
	vt, err := h.AuthService.ValidateVerificationToken(c.Request.Context(), req.Token, "email")
	if err != nil || vt == nil {
		c.JSON(http.StatusBadRequest, responses.CommonResponse{
			ResponseCode:    http.StatusBadRequest,
			ResponseMessage: "Invalid or expired token",
		})
		return
	}
	user, err := services.UserSvc.GetUserByID(c.Request.Context(), fmt.Sprintf("%d", vt.UserID))
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, responses.CommonResponse{
			ResponseCode:    http.StatusNotFound,
			ResponseMessage: "User not found",
		})
		return
	}
	err = h.AuthService.MarkEmailVerified(c.Request.Context(), user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.CommonResponse{
			ResponseCode:    http.StatusInternalServerError,
			ResponseMessage: "Failed to mark email verified",
		})
		return
	}
	c.JSON(http.StatusOK, responses.CommonResponse{
		ResponseCode:    http.StatusOK,
		ResponseMessage: "Email verified successfully",
	})
}

func (h *AuthHandler) RequestPhoneVerification(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, responses.CommonResponse{
			ResponseCode:    http.StatusUnauthorized,
			ResponseMessage: "Unauthorized",
		})
		return
	}
	user, err := services.UserSvc.GetUserByID(c.Request.Context(), fmt.Sprintf("%v", userID))
	if err != nil || user == nil || user.Phone == nil || *user.Phone == "" {
		c.JSON(http.StatusNotFound, responses.CommonResponse{
			ResponseCode:    http.StatusNotFound,
			ResponseMessage: "User or phone not found",
		})
		return
	}
	// Rate limiting: 5 per minute per user
	rlKey := fmt.Sprintf("phone_verify_rl:%v", userID)
	rateLimited, _ := h.AuthService.IsRateLimited(c.Request.Context(), rlKey, 5, time.Minute)
	if rateLimited {
		c.JSON(http.StatusTooManyRequests, responses.CommonResponse{
			ResponseCode:    http.StatusTooManyRequests,
			ResponseMessage: "Too many requests. Please wait before requesting another code.",
		})
		return
	}
	phoneToken, err := h.AuthService.CreateVerificationToken(c.Request.Context(), user.ID, "phone")
	if err != nil || phoneToken == "" {
		c.JSON(http.StatusInternalServerError, responses.CommonResponse{
			ResponseCode:    http.StatusInternalServerError,
			ResponseMessage: "Failed to create phone verification token",
		})
		return
	}
	chatId := *user.Phone + "@c.us"
	waCfg := config.GetConfig().WhatsApp
	wa := whatsapp.NewWhatsAppService(waCfg.BaseURL, waCfg.Session)
	_ = wa.StartTyping(chatId)
	err = wa.SendText(chatId, fmt.Sprintf("Your verification code is: %s", phoneToken), nil, false, false)
	_ = wa.StopTyping(chatId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.CommonResponse{
			ResponseCode:    http.StatusInternalServerError,
			ResponseMessage: "Failed to send phone verification code",
		})
		return
	}
	c.JSON(http.StatusOK, responses.CommonResponse{
		ResponseCode:    http.StatusOK,
		ResponseMessage: "Phone verification code sent",
	})
}

func (h *AuthHandler) VerifyPhone(c *gin.Context) {
	var req struct {
		Token string `json:"token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Token == "" {
		c.JSON(http.StatusBadRequest, responses.CommonResponse{
			ResponseCode:    http.StatusBadRequest,
			ResponseMessage: "Token required",
		})
		return
	}
	vt, err := h.AuthService.ValidateVerificationToken(c.Request.Context(), req.Token, "phone")
	if err != nil || vt == nil {
		c.JSON(http.StatusBadRequest, responses.CommonResponse{
			ResponseCode:    http.StatusBadRequest,
			ResponseMessage: "Invalid or expired token",
		})
		return
	}
	user, err := services.UserSvc.GetUserByID(c.Request.Context(), fmt.Sprintf("%d", vt.UserID))
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, responses.CommonResponse{
			ResponseCode:    http.StatusNotFound,
			ResponseMessage: "User not found",
		})
		return
	}
	err = h.AuthService.MarkPhoneVerified(c.Request.Context(), user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.CommonResponse{
			ResponseCode:    http.StatusInternalServerError,
			ResponseMessage: "Failed to mark phone verified",
		})
		return
	}
	c.JSON(http.StatusOK, responses.CommonResponse{
		ResponseCode:    http.StatusOK,
		ResponseMessage: "Phone verified successfully",
	})
}
