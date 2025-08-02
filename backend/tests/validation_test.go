package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	internal "go-next/internal"
)

type ValidationErrorResponse struct {
	Message string              `json:"message"`
	Errors  map[string][]string `json:"errors"`
}

func setupValidationTestDB() *gorm.DB {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&internal.User{}, &internal.Post{}, &internal.Comment{}, &internal.Category{})
	return db
}

func setupValidationRouter(db *gorm.DB) *gin.Engine {
	internal.DB = db
	r := gin.Default()
	r.POST("/api/register", internal.RegisterHandlerGin)
	r.POST("/api/posts", internal.CreatePost)
	return r
}

func TestRegisterValidationErrors(t *testing.T) {
	db := setupValidationTestDB()
	r := setupValidationRouter(db)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/register", bytes.NewBuffer([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
	var resp ValidationErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Validation failed", resp.Message)
	assert.Contains(t, resp.Errors, "Username")
	assert.Contains(t, resp.Errors, "Email")
	assert.Contains(t, resp.Errors, "Password")
}

func TestPostCreateValidationErrors(t *testing.T) {
	db := setupValidationTestDB()
	r := setupValidationRouter(db)

	w := httptest.NewRecorder()
	invalid := map[string]interface{}{"title": "", "content": "", "user_id": 0, "category_id": 0}
	body, _ := json.Marshal(invalid)
	req, _ := http.NewRequest("POST", "/api/posts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
	var resp ValidationErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Validation failed", resp.Message)
	assert.Contains(t, resp.Errors, "Title")
	assert.Contains(t, resp.Errors, "Content")
	assert.Contains(t, resp.Errors, "UserID")
	assert.Contains(t, resp.Errors, "CategoryID")
}
