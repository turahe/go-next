package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	internal "wordpress-go-next/backend/internal"
)

func setupTestDB() *gorm.DB {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&internal.User{}, &internal.Post{}, &internal.Comment{}, &internal.Category{})
	return db
}

func setupRouter(db *gorm.DB) *gin.Engine {
	internal.DB = db
	r := gin.Default()
	r.POST("/api/register", internal.RegisterHandlerGin)
	r.POST("/api/login", internal.LoginHandlerGin)
	api := r.Group("/api")
	{
		api.GET("/posts", internal.GetPosts)
		api.GET("/posts/:id", internal.GetPost)
		api.POST("/posts", internal.AuthMiddleware(), internal.CreatePost)
		api.PUT("/posts/:id", internal.AuthMiddleware(), internal.UpdatePost)
		api.DELETE("/posts/:id", internal.AuthMiddleware(), internal.DeletePost)
	}
	return r
}

func TestPostCRUD(t *testing.T) {
	os.Setenv("JWT_SECRET", "testsecret")
	db := setupTestDB()
	r := setupRouter(db)

	// Register
	regBody := map[string]string{"username": "testuser", "email": "test@example.com", "password": "password"}
	regJSON, _ := json.Marshal(regBody)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/register", bytes.NewBuffer(regJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != 201 {
		t.Fatalf("Registration failed: %d", w.Code)
	}

	// Login
	loginBody := map[string]string{"email": "test@example.com", "password": "password"}
	loginJSON, _ := json.Marshal(loginBody)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/login", bytes.NewBuffer(loginJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("Login failed: %d", w.Code)
	}
	var loginResp map[string]string
	json.Unmarshal(w.Body.Bytes(), &loginResp)
	token := loginResp["token"]
	if token == "" {
		t.Fatal("No token returned")
	}

	// Create Post
	postBody := map[string]interface{}{"title": "Test Post", "content": "Hello World", "user_id": 1, "category_id": 1}
	postJSON, _ := json.Marshal(postBody)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/posts", bytes.NewBuffer(postJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	if w.Code != 201 {
		t.Fatalf("Create post failed: %d", w.Code)
	}

	// List Posts
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/posts", nil)
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("List posts failed: %d", w.Code)
	}
}
