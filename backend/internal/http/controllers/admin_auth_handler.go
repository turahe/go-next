package controllers

import (
	"go-next/internal/http/requests"
	"go-next/internal/http/responses"
	"go-next/internal/models"
	"go-next/internal/rules"
	"go-next/pkg/database"
	"go-next/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AdminAuthHandler interface {
	Register(c *gin.Context)
}

type adminAuthHandler struct{}

func NewAdminAuthHandler() AdminAuthHandler {
	return &adminAuthHandler{}
}

func (h *adminAuthHandler) Register(c *gin.Context) {
	var req requests.AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	if err := rules.Validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, requests.FormatValidationError(err))
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error hashing password"})
		return
	}
	var count int64
	if err := database.DB.Model(&models.User{}).Where("username = ?", req.Username).Or("email = ?", req.Email).Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if count > 0 {
		var user models.User
		if err := database.DB.Where("username = ?", req.Username).First(&user).Error; err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Username already taken"})
			return
		}
		if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email already registered"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "User already exists"})
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or missing role"})
			return err
		}
		user := models.User{
			Username:     req.Username,
			Email:        req.Email,
			PasswordHash: string(hash),
			Roles:        []models.Role{role},
		}
		if err := tx.Create(&user).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User already exists or DB error"})
			return err
		}
		var err error
		clientKey, err = utils.GenerateRandomKey(16)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate client key"})
			return err
		}
		secretKey, err = utils.GenerateRandomKey(32)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate secret key"})
			return err
		}
		jwtKey := models.JWTKey{
			KeyID:     clientKey,
			Algorithm: "HS256",
			Key:       secretKey,
			IsActive:  true,
		}
		if err := tx.Create(&jwtKey).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store JWT key"})
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
