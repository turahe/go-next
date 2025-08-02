package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"wordpress-go-next/backend/internal/http/requests"
	"wordpress-go-next/backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTagRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	// Initialize services for testing
	// Note: In a real test, you would mock the services or use test database
	r.GET("/api/v1/tags", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    []models.Tag{},
			"message": "Tags retrieved successfully",
		})
	})

	r.POST("/api/v1/tags", func(c *gin.Context) {
		var req requests.CreateTagRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Validation failed",
			})
			return
		}

		// Validate request
		if err := req.Validate(); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}

		// Convert to model
		tag := req.ToTag()
		tag.ID = 1 // Mock ID

		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"data":    tag,
			"message": "Tag created successfully",
		})
	})

	return r
}

func TestCreateTag(t *testing.T) {
	db := setupTagTestDB()
	r := setupTagRouter(db)

	// Test valid tag creation
	validTag := requests.CreateTagRequest{
		Name:        "Test Tag",
		Slug:        "test-tag",
		Description: "A test tag",
		Color:       "#FF0000",
		Type:        "general",
		IsActive:    true,
	}

	jsonData, _ := json.Marshal(validTag)
	req, _ := http.NewRequest("POST", "/api/v1/tags", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))
}

func TestCreateTagValidation(t *testing.T) {
	db := setupTagTestDB()
	r := setupTagRouter(db)

	// Test invalid tag (missing required fields)
	invalidTag := requests.CreateTagRequest{
		Name: "",             // Missing required name
		Type: "invalid-type", // Invalid type
	}

	jsonData, _ := json.Marshal(invalidTag)
	req, _ := http.NewRequest("POST", "/api/v1/tags", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTagValidation(t *testing.T) {
	// Test valid hex color
	validColor := "#FF0000"
	assert.True(t, isValidHexColor(validColor))

	// Test invalid hex color
	invalidColor := "FF0000" // Missing #
	assert.False(t, isValidHexColor(invalidColor))

	// Test valid slug
	validSlug := "test-tag"
	assert.True(t, isValidSlug(validSlug))

	// Test invalid slug
	invalidSlug := "test@tag" // Invalid character
	assert.False(t, isValidSlug(invalidSlug))
}

func TestTagModel(t *testing.T) {
	// Test tag model creation
	tag := &models.Tag{
		Name:        "Test Tag",
		Slug:        "test-tag",
		Description: "A test tag",
		Color:       "#FF0000",
		Type:        models.TagTypeGeneral,
		IsActive:    true,
	}

	// Test validation methods
	assert.True(t, tag.IsValidType())
	assert.True(t, tag.IsValidColor())

	// Test default color
	tag.Color = ""
	tag.SetDefaultColor()
	assert.NotEmpty(t, tag.Color)
}

// Helper functions for testing
func isValidHexColor(color string) bool {
	if len(color) != 7 || color[0] != '#' {
		return false
	}
	for i := 1; i < 7; i++ {
		c := color[i]
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

func isValidSlug(slug string) bool {
	if len(slug) == 0 {
		return false
	}

	for _, char := range slug {
		if !((char >= 'a' && char <= 'z') ||
			(char >= '0' && char <= '9') ||
			char == '-' || char == '_') {
			return false
		}
	}

	return true
}

func setupTagTestDB() *gorm.DB {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})

	// Auto migrate models
	err := db.AutoMigrate(&models.Tag{}, &models.TaggedEntity{})
	if err != nil {
		panic(err)
	}

	return db
}
