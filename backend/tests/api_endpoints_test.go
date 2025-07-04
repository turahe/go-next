package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	internal "wordpress-go-next/backend/internal"
	controllers "wordpress-go-next/backend/internal/http/controllers"

	"github.com/stretchr/testify/assert"
)

func setupTestDB() *gorm.DB {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	internal.DB = db
	internal.AutoMigrate()
	return db
}

func setupRouter() *gin.Engine {
	r := gin.Default()
	r.POST("/api/register", controllers.RegisterHandlerGin)
	r.POST("/api/login", controllers.LoginHandlerGin)
	// Add more routes as needed for testing
	return r
}

func TestRegisterAndLogin(t *testing.T) {
	db := setupTestDB()
	_ = db
	r := setupRouter()

	// Register
	regBody := map[string]string{"username": "testuser", "email": "test@example.com", "password": "password123"}
	regJSON, _ := json.Marshal(regBody)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/register", bytes.NewBuffer(regJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, 201, w.Code)

	// Login
	loginBody := map[string]string{"email": "test@example.com", "password": "password123"}
	loginJSON, _ := json.Marshal(loginBody)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/login", bytes.NewBuffer(loginJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	var loginResp map[string]string
	json.Unmarshal(w.Body.Bytes(), &loginResp)
	assert.NotEmpty(t, loginResp["token"])
}

func TestPostEndpoints(t *testing.T) {
	// TODO: Implement happy-path and error-path tests for post CRUD
}

func TestCommentEndpoints(t *testing.T) {
	// TODO: Implement happy-path and error-path tests for comment CRUD
}

func TestCategoryEndpoints(t *testing.T) {
	// TODO: Implement happy-path and error-path tests for category CRUD
}

func TestUserEndpoints(t *testing.T) {
	// TODO: Implement happy-path and error-path tests for user profile, update, etc.
}

func TestRoleEndpoints(t *testing.T) {
	// TODO: Implement happy-path and error-path tests for role CRUD
}

func TestUserRoleAssignmentEndpoints(t *testing.T) {
	// TODO: Implement happy-path and error-path tests for assigning, removing, and listing user roles
}
