package controllers

import (
	"net/http"
	"strconv"

	"go-next/internal/http/requests"
	"go-next/internal/models"
	"go-next/internal/rules"

	"go-next/internal/http/responses"
	"go-next/internal/services"
	"go-next/pkg/database"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserHandler interface {
	GetUserProfile(c *gin.Context)
	GetUsers(c *gin.Context)
	UpdateUserProfile(c *gin.Context)
	UpdateUserRole(c *gin.Context)
	UserCreate(c *gin.Context)
	DeleteUser(c *gin.Context)
}

type userHandler struct {
	UserService services.UserService
}

func NewUserHandler(userService services.UserService) UserHandler {
	return &userHandler{UserService: userService}
}

// GetUserProfile godoc
// @Summary      Get user profile
// @Description  Get a user profile by ID
// @Tags         users
// @Produce      json
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}  models.User
// @Failure      404  {object}  map[string]string
// @Router       /users/{id} [get]
func (h *userHandler) GetUserProfile(c *gin.Context) {
	id := c.Param("id")
	user, err := h.UserService.GetUserByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	// Hide password hash
	user.PasswordHash = ""
	c.JSON(http.StatusOK, user)
}

// UpdateUserProfile godoc
// @Summary      Update user profile
// @Description  Update your own user profile
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id    path      int         true  "User ID"
// @Param        user  body      models.User true  "User profile update"
// @Success      200   {object}  models.User
// @Failure      400   {object}  map[string]string
// @Failure      403   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /users/{id} [put]
func (h *userHandler) UpdateUserProfile(c *gin.Context) {
	id := c.Param("id")
	userID, exists := c.Get("user_id")
	if !exists || strconv.Itoa(int(userID.(uint))) != id {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only update your own profile"})
		return
	}
	user, err := h.UserService.GetUserByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	var input requests.UserProfileUpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	if err := rules.Validate.Struct(input); err != nil {
		c.JSON(http.StatusBadRequest, requests.FormatValidationError(err))
		return
	}
	if err := h.UserService.UpdateUserProfile(user, input.Username, input.Email, input.Phone, input.EmailVerified, input.PhoneVerified); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}
	user.PasswordHash = ""
	c.JSON(http.StatusOK, user)
}

// UpdateUserRole godoc
// @Summary      Update user role
// @Description  Update a user's role
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id    path      int    true  "User ID"
// @Param        role  body      object true  "Role update"
// @Success      200   {object}  models.User
// @Failure      400   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Router       /users/{id}/role [put]
func (h *userHandler) UpdateUserRole(c *gin.Context) {
	id := c.Param("id")
	user, err := h.UserService.GetUserByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	var input struct {
		Role string `json:"role" validate:"required,oneof=admin editor moderator user guest"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	if err := rules.Validate.Struct(input); err != nil {
		c.JSON(http.StatusBadRequest, requests.FormatValidationError(err))
		return
	}
	roleName := input.Role
	if roleName == "" {
		roleName = "user"
	}
	var role models.Role
	if err := database.DB.Where("name = ?", roleName).First(&role).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or missing role"})
		return
	}
	user.Roles = []models.Role{role}
	user.PasswordHash = ""
	c.JSON(http.StatusOK, user)
}

// GetUsers godoc
// @Summary      List users
// @Description  Get all users with pagination
// @Tags         users
// @Produce      json
// @Param        page   query     int    false "Page number"
// @Param        limit  query     int    false "Items per page"
// @Param        search query     string false "Search term"
// @Success      200    {object}  responses.LaravelPaginationResponse
// @Failure      400    {object}  map[string]string
// @Router       /users [get]
func (h *userHandler) GetUsers(c *gin.Context) {
	params := responses.ParsePaginationParams(c)
	search := c.Query("search")

	offset := (params.Page - 1) * params.PerPage

	var users []models.User
	var total int64

	query := database.DB.Model(&models.User{}).Preload("Roles")

	if search != "" {
		query = query.Where("username ILIKE ? OR email ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to count users")
		return
	}

	// Get paginated users
	if err := query.Offset(offset).Limit(params.PerPage).Find(&users).Error; err != nil {
		responses.SendError(c, http.StatusInternalServerError, "Failed to fetch users")
		return
	}

	// Hide password hashes
	for i := range users {
		users[i].PasswordHash = ""
	}

	// Send Laravel-style pagination response
	responses.SendLaravelPaginationWithMessage(c, "Users retrieved successfully", users, total, int64(params.Page), int64(params.PerPage))
}

// UserCreate godoc
// @Summary      Create a new user
// @Description  Create a new user with the provided details
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        user  body      requests.AuthRequest true  "User creation details"
// @Success      201   {object}  models.User
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /users [post]
func (h *userHandler) UserCreate(c *gin.Context) {
	var input requests.AuthRequest
	if !requests.ValidateRequest(c, &input) {
		return
	}
	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		roleName := input.Role
		if roleName == "" {
			roleName = "user"
		}
		var role models.Role
		if err := tx.Where("name = ?", roleName).First(&role).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or missing role"})
			return err
		}
		user := &models.User{
			Username: input.Username,
			Email:    input.Email,
			Roles:    []models.Role{role},
		}
		if err := user.HashPassword(input.Password); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return err
		}
		if err := tx.Create(user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return err
		}
		user.PasswordHash = ""
		c.JSON(http.StatusCreated, user)
		return nil
	}); err != nil {
		return
	}
}

// DeleteUser godoc
// @Summary      Delete a user
// @Description  Delete a user by ID
// @Tags         users
// @Produce      json
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /users/{id} [delete]
func (h *userHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")

	// Check if user exists
	user, err := h.UserService.GetUserByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Delete user
	if err := database.DB.Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}
