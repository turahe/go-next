package controllers

import (
	"wordpress-go-next/backend/internal/http/requests"
	"wordpress-go-next/backend/internal/http/responses"
	"wordpress-go-next/backend/internal/models"
	"wordpress-go-next/backend/internal/rules"
	"wordpress-go-next/backend/pkg/database"
	"wordpress-go-next/backend/pkg/utils"

	"time"

	"net/http"

	"github.com/gin-gonic/gin"
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
	RequestPasswordReset(c *gin.Context)
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

func (h *authHandler) Register(c *gin.Context) {
	var req requests.AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}
	if err := rules.Validate.Struct(req); err != nil {
		c.JSON(400, requests.FormatValidationError(err))
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
			UserID:          user.ID,
			ClientKey:       clientKey,
			SecretKey:       secretKey,
			TokenExpiration: 3600, // default 1 hour
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

func (h *authHandler) Login(c *gin.Context) {
	var req requests.AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}
	if err := rules.Validate.StructPartial(req, "Email", "Password"); err != nil {
		c.JSON(400, requests.FormatValidationError(err))
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
	tokenModel := models.Token{
		Token:        token,
		UserID:       user.ID,
		ClientSecret: "", // Not used for refresh, can be set if needed
		RefreshToken: refreshToken,
		ExpiredAt:    time.Now().Add(7 * 24 * time.Hour), // Refresh token valid for 7 days
	}
	if err := database.DB.Create(&tokenModel).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to store refresh token"})
		return
	}
	c.JSON(http.StatusOK, AuthResponse{Token: token, RefreshToken: refreshToken})
}

// RefreshToken endpoint
func (h *authHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}
	var tokenModel models.Token
	if err := database.DB.Where("refresh_token = ? AND expired_at > ?", req.RefreshToken, time.Now()).First(&tokenModel).Error; err != nil {
		c.JSON(401, gin.H{"error": "Invalid or expired refresh token"})
		return
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
	tokenModel.RefreshToken = newRefreshToken
	tokenModel.ExpiredAt = time.Now().Add(7 * 24 * time.Hour)
	if err := database.DB.Save(&tokenModel).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to update refresh token"})
		return
	}
	c.JSON(200, AuthResponse{Token: token, RefreshToken: newRefreshToken})
}

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
	database.DB.Create(&t)
	c.JSON(http.StatusOK, gin.H{"message": "Verification email sent", "token": token})
}

func (h *authHandler) VerifyEmail(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		Token string `json:"token"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	var vt models.VerificationToken
	if err := database.DB.Where("user_id = ? AND token = ? AND type = ? AND used = ? AND expires_at > ?", id, input.Token, "email_verification", false, time.Now()).First(&vt).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired token"})
		return
	}
	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	now := time.Now()
	user.EmailVerified = &now
	vt.Used = true
	database.DB.Save(&user)
	database.DB.Save(&vt)
	c.JSON(http.StatusOK, gin.H{"message": "Email verified"})
}

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
	database.DB.Create(&t)
	c.JSON(http.StatusOK, gin.H{"message": "Verification SMS sent", "token": token})
}

func (h *authHandler) VerifyPhone(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		Token string `json:"token"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	var vt models.VerificationToken
	if err := database.DB.Where("user_id = ? AND token = ? AND type = ? AND used = ? AND expires_at > ?", id, input.Token, "phone_verification", false, time.Now()).First(&vt).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired token"})
		return
	}
	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	now := time.Now()
	user.PhoneVerified = &now
	vt.Used = true
	database.DB.Save(&user)
	database.DB.Save(&vt)
	c.JSON(http.StatusOK, gin.H{"message": "Phone verified"})
}

func (h *authHandler) RequestPasswordReset(c *gin.Context) {
	var input struct {
		Email string `json:"email"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	var user models.User
	if err := database.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
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
		Type:      "password_reset",
		ExpiresAt: time.Now().Add(30 * time.Minute),
	}
	database.DB.Create(&t)
	c.JSON(http.StatusOK, gin.H{"message": "Password reset link sent", "token": token})
}

func (h *authHandler) ResetPassword(c *gin.Context) {
	var input struct {
		UserID      uint   `json:"user_id"`
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	var vt models.VerificationToken
	if err := database.DB.Where("user_id = ? AND token = ? AND type = ? AND used = ? AND expires_at > ?", input.UserID, input.Token, "password_reset", false, time.Now()).First(&vt).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired token"})
		return
	}
	var user models.User
	if err := database.DB.First(&user, input.UserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	user.PasswordHash = string(hash)
	vt.Used = true
	database.DB.Save(&user)
	database.DB.Save(&vt)
	c.JSON(http.StatusOK, gin.H{"message": "Password reset"})
}
