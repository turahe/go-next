package controllers

import (
	"go-next/internal/http/requests"
	"go-next/internal/http/responses"
	"go-next/internal/models"
	"go-next/internal/services"
	"go-next/pkg/database"
	"go-next/pkg/utils"

	"time"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthHandler interface {
	Register(c *gin.Context)
	Login(c *gin.Context)
	RequestEmailVerification(c *gin.Context)
	VerifyEmail(c *gin.Context)
	RequestPhoneVerification(c *gin.Context)
	VerifyPhone(c *gin.Context)
	ResetPassword(c *gin.Context)
	RefreshToken(c *gin.Context)
}

type authHandler struct{}

func NewAuthHandler() AuthHandler {
	return &authHandler{}
}

// Use shared validator instance from a common package if available

type AuthResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

// Register creates a new user account
// @Summary Register new user
// @Description Create a new user account with the provided credentials
// @Tags auth
// @Accept json
// @Produce json
// @Param user body requests.AuthRequest true "User registration data"
// @Success 201 {object} responses.CommonResponse{data=map[string]interface{}}
// @Failure 400 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /auth/register [post]
func (h *authHandler) Register(c *gin.Context) {
	var req requests.AuthRequest
	if !requests.ValidateRequest(c, &req) {
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(500, gin.H{"error": "Error hashing password"})
		return
	}
	var count int64
	if err := database.DB.Model(&models.User{}).Where("username = ?", req.Username).Or("email = ?", req.Email).Count(&count).Error; err != nil {
		c.JSON(500, gin.H{"error": "Database error"})
		return
	}
	if count > 0 {
		var user models.User
		if err := database.DB.Where("username = ?", req.Username).First(&user).Error; err == nil {
			c.JSON(400, gin.H{"error": "Username already taken"})
			return
		}
		if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err == nil {
			c.JSON(400, gin.H{"error": "Email already registered"})
			return
		}
		c.JSON(400, gin.H{"error": "User already exists"})
		return
	}
	var clientKey, secretKey string
	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		roleName := req.Role
		if roleName == "" {
			roleName = "user"
		}
		var role models.Role
		if err := tx.Where("name = ?", roleName).First(&role).Error; err != nil {
			c.JSON(400, gin.H{"error": "Invalid or missing role"})
			return err
		}
		user := models.User{
			Username:     req.Username,
			Email:        req.Email,
			PasswordHash: string(hash),
			Roles:        []models.Role{role},
		}
		if err := tx.Create(&user).Error; err != nil {
			c.JSON(400, gin.H{"error": "User already exists or DB error"})
			return err
		}
		var err error
		clientKey, err = utils.GenerateRandomKey(16)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to generate client key"})
			return err
		}
		secretKey, err = utils.GenerateRandomKey(32)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to generate secret key"})
			return err
		}
		jwtKey := models.JWTKey{
			KeyID:     clientKey,
			Algorithm: "HS256",
			Key:       secretKey,
			IsActive:  true,
		}
		if err := tx.Create(&jwtKey).Error; err != nil {
			c.JSON(500, gin.H{"error": "Failed to store JWT key"})
			return err
		}
		return nil
	}); err != nil {
		return
	}
	c.JSON(http.StatusCreated, responses.CommonResponse{
		ResponseCode:    http.StatusCreated,
		ResponseMessage: "Registration successful",
		Data:            gin.H{"client_key": clientKey, "secret_key": secretKey},
	})
}

// Login authenticates a user and returns access tokens
// @Summary User login
// @Description Authenticate user with email and password, return access and refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body requests.AuthRequest true "Login credentials"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /auth/login [post]
func (h *authHandler) Login(c *gin.Context) {
	var req requests.AuthRequest
	if !requests.ValidateRequestPartial(c, &req, "Email", "Password") {
		return
	}
	var user models.User
	if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(401, gin.H{"error": "Invalid credentials"})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(401, gin.H{"error": "Invalid credentials"})
		return
	}
	token, err := utils.GenerateJWT(user.ID)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to generate token"})
		return
	}
	refreshToken, err := utils.GenerateRandomKey(32)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to generate refresh token"})
		return
	}
	// Store refresh token in Token table
	expiredAt := time.Now().Add(7 * 24 * time.Hour) // Refresh token valid for 7 days
	tokenModel := models.Token{
		Token:     refreshToken,
		UserID:    user.ID,
		Type:      "refresh",
		ExpiredAt: &expiredAt,
		IsActive:  true,
	}
	if err := database.DB.Create(&tokenModel).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to store refresh token"})
		return
	}
	c.JSON(http.StatusOK, AuthResponse{Token: token, RefreshToken: refreshToken})
}

// RefreshToken refreshes the access token using a refresh token
// @Summary Refresh access token
// @Description Use refresh token to get a new access token
// @Tags auth
// @Accept json
// @Produce json
// @Param refresh body object true "Refresh token data" schema(object{refresh_token=string})
// @Success 200 {object} AuthResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /auth/refresh [post]
func (h *authHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	// Try to get token from cache first
	tokenModel, err := services.TokenCacheSvc.GetTokenByValue(req.RefreshToken)
	if err != nil {
		// Cache miss, get from database
		if err := database.DB.Where("token = ? AND type = ? AND expired_at > ? AND is_active = ?", req.RefreshToken, "refresh", time.Now(), true).First(&tokenModel).Error; err != nil {
			c.JSON(401, gin.H{"error": "Invalid or expired refresh token"})
			return
		}
		// Cache the token
		services.TokenCacheSvc.CacheToken(tokenModel)
	}

	// Issue new access token
	token, err := utils.GenerateJWT(tokenModel.UserID)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to generate token"})
		return
	}
	// Optionally, rotate refresh token
	newRefreshToken, err := utils.GenerateRandomKey(32)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to generate refresh token"})
		return
	}
	newExpiredAt := time.Now().Add(7 * 24 * time.Hour)
	tokenModel.Token = newRefreshToken
	tokenModel.ExpiredAt = &newExpiredAt
	if err := database.DB.Save(&tokenModel).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to update refresh token"})
		return
	}

	// Update cache with new token
	services.TokenCacheSvc.CacheToken(tokenModel)

	c.JSON(200, AuthResponse{Token: token, RefreshToken: newRefreshToken})
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
	token, err := utils.GenerateRandomKey(32)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to generate token"})
		return
	}
	t := models.VerificationToken{
		UserID:    user.ID,
		Token:     token,
		Type:      "email_verification",
		ExpiresAt: time.Now().Add(30 * time.Minute),
	}
	if err := database.DB.Create(&t).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to create verification token"})
		return
	}

	// Cache the verification token
	services.TokenCacheSvc.CacheVerificationToken(&t)

	// Invalidate user's verification tokens cache for this type
	services.TokenCacheSvc.InvalidateUserVerificationTokens(user.ID, models.EmailVerification)

	c.JSON(http.StatusOK, gin.H{"message": "Verification email sent", "token": token})
}

// VerifyEmail verifies the user's email address using the provided token
// @Summary Verify email address
// @Description Verify user's email address using the verification token
// @Tags auth
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param verification body object true "Verification data" schema(object{token=string})
// @Success 200 {object} responses.SuccessResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /users/{id}/verify-email [post]
func (h *authHandler) VerifyEmail(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		Token string `json:"token"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Try to get verification token from cache first
	vt, err := services.TokenCacheSvc.GetVerificationTokenByValue(input.Token)
	if err != nil {
		// Cache miss, get from database
		if err := database.DB.Where("user_id = ? AND token = ? AND type = ? AND used = ? AND expires_at > ?", id, input.Token, "email_verification", false, time.Now()).First(&vt).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired token"})
			return
		}
		// Cache the token
		services.TokenCacheSvc.CacheVerificationToken(vt)
	}

	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	now := time.Now()
	user.EmailVerified = &now
	vt.Used = true
	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to update user"})
		return
	}
	if err := database.DB.Save(&vt).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to update token"})
		return
	}

	// Invalidate cached tokens
	services.TokenCacheSvc.InvalidateVerificationToken(vt.ID)
	services.TokenCacheSvc.InvalidateUserVerificationTokens(user.ID, models.EmailVerification)

	c.JSON(http.StatusOK, gin.H{"message": "Email verified"})
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
	token, err := utils.GenerateRandomKey(32)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to generate token"})
		return
	}
	t := models.VerificationToken{
		UserID:    user.ID,
		Token:     token,
		Type:      "phone_verification",
		ExpiresAt: time.Now().Add(30 * time.Minute),
	}
	if err := database.DB.Create(&t).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to create verification token"})
		return
	}

	// Cache the verification token
	services.TokenCacheSvc.CacheVerificationToken(&t)

	// Invalidate user's verification tokens cache for this type
	services.TokenCacheSvc.InvalidateUserVerificationTokens(user.ID, models.PhoneVerification)

	c.JSON(http.StatusOK, gin.H{"message": "Verification SMS sent", "token": token})
}

// VerifyPhone verifies the user's phone number using the provided token
// @Summary Verify phone number
// @Description Verify user's phone number using the verification token
// @Tags auth
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param verification body object true "Verification data" schema(object{token=string})
// @Success 200 {object} responses.SuccessResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /users/{id}/verify-phone [post]
func (h *authHandler) VerifyPhone(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		Token string `json:"token"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Try to get verification token from cache first
	vt, err := services.TokenCacheSvc.GetVerificationTokenByValue(input.Token)
	if err != nil {
		// Cache miss, get from database
		if err := database.DB.Where("user_id = ? AND token = ? AND type = ? AND used = ? AND expires_at > ?", id, input.Token, "phone_verification", false, time.Now()).First(&vt).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired token"})
			return
		}
		// Cache the token
		services.TokenCacheSvc.CacheVerificationToken(vt)
	}

	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	now := time.Now()
	user.PhoneVerified = &now
	vt.Used = true
	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to update user"})
		return
	}
	if err := database.DB.Save(&vt).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to update token"})
		return
	}

	// Invalidate cached tokens
	services.TokenCacheSvc.InvalidateVerificationToken(vt.ID)
	services.TokenCacheSvc.InvalidateUserVerificationTokens(user.ID, models.PhoneVerification)

	c.JSON(http.StatusOK, gin.H{"message": "Phone verified"})
}

// ResetPassword resets the user's password using a reset token
// @Summary Reset password
// @Description Reset user's password using the provided reset token
// @Tags auth
// @Accept json
// @Produce json
// @Param reset body object true "Password reset data" schema(object{user_id=string,token=string,new_password=string})
// @Success 200 {object} responses.SuccessResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /auth/reset-password [post]
func (h *authHandler) ResetPassword(c *gin.Context) {
	var input struct {
		UserID      string `json:"user_id"`
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Parse user ID from string to UUID
	userID, err := uuid.Parse(input.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Try to get verification token from cache first
	vt, err := services.TokenCacheSvc.GetVerificationTokenByValue(input.Token)
	if err != nil {
		// Cache miss, get from database
		if err := database.DB.Where("user_id = ? AND token = ? AND type = ? AND used = ? AND expires_at > ?", userID, input.Token, "password_reset", false, time.Now()).First(&vt).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired token"})
			return
		}
		// Cache the token
		services.TokenCacheSvc.CacheVerificationToken(vt)
	}

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to hash password"})
		return
	}
	user.PasswordHash = string(hash)
	vt.Used = true
	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to update user"})
		return
	}
	if err := database.DB.Save(&vt).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to update token"})
		return
	}

	// Invalidate cached tokens
	services.TokenCacheSvc.InvalidateVerificationToken(vt.ID)
	services.TokenCacheSvc.InvalidateUserVerificationTokens(user.ID, models.PasswordReset)

	c.JSON(http.StatusOK, gin.H{"message": "Password reset"})
}
