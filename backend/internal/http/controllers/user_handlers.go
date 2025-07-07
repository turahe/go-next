package controllers

import (
	"net/http"
	"strconv"

	"wordpress-go-next/backend/internal/http/requests"
	"wordpress-go-next/backend/internal/models"
	"wordpress-go-next/backend/internal/rules"

	"wordpress-go-next/backend/internal/services"
	"wordpress-go-next/backend/pkg/database"

	"wordpress-go-next/backend/internal/http/dto"
	"wordpress-go-next/backend/internal/http/responses"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserHandler interface {
	GetUserProfile(c *gin.Context)
	UpdateUserProfile(c *gin.Context)
	UpdateUserRole(c *gin.Context)
	UserCreate(c *gin.Context)
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
// @Success      200  {object}  dto.UserDTO
// @Failure      404  {object}  map[string]string
// @Router       /users/{id} [get]
func (h *userHandler) GetUserProfile(c *gin.Context) {
	id := c.Param("id")
	user, err := h.UserService.GetUserByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, responses.CommonResponse{
			ResponseCode:    http.StatusNotFound,
			ResponseMessage: "User not found",
		})
		return
	}
	c.JSON(http.StatusOK, responses.CommonResponse{
		ResponseCode:    http.StatusOK,
		ResponseMessage: "User profile fetched successfully",
		Data:            dto.ToUserDTO(user),
	})
}

// UpdateUserProfile godoc
// @Summary      Update user profile
// @Description  Update your own user profile
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id    path      int         true  "User ID"
// @Param        user  body      models.User true  "User profile update"
// @Success      200   {object}  dto.UserDTO
// @Failure      400   {object}  map[string]string
// @Failure      403   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /users/{id} [put]
func (h *userHandler) UpdateUserProfile(c *gin.Context) {
	id := c.Param("id")
	userID, exists := c.Get("user_id")
	if !exists || strconv.Itoa(int(userID.(uint))) != id {
		c.JSON(http.StatusForbidden, responses.CommonResponse{
			ResponseCode:    http.StatusForbidden,
			ResponseMessage: "You can only update your own profile",
		})
		return
	}
	user, err := h.UserService.GetUserByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, responses.CommonResponse{
			ResponseCode:    http.StatusNotFound,
			ResponseMessage: "User not found",
		})
		return
	}
	var input requests.UserProfileUpdateRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, responses.CommonResponse{
			ResponseCode:    http.StatusBadRequest,
			ResponseMessage: "Invalid request",
		})
		return
	}
	if err := rules.Validate.Struct(input); err != nil {
		c.JSON(http.StatusBadRequest, responses.CommonResponse{
			ResponseCode:    http.StatusBadRequest,
			ResponseMessage: "Validation error",
			Data:            requests.FormatValidationError(err),
		})
		return
	}
	if err := h.UserService.UpdateUserProfile(c.Request.Context(), user, input.Username, input.Email, input.Phone, input.EmailVerified, input.PhoneVerified); err != nil {
		c.JSON(http.StatusInternalServerError, responses.CommonResponse{
			ResponseCode:    http.StatusInternalServerError,
			ResponseMessage: "Failed to update user",
		})
		return
	}
	c.JSON(http.StatusOK, responses.CommonResponse{
		ResponseCode:    http.StatusOK,
		ResponseMessage: "User profile updated successfully",
		Data:            dto.ToUserDTO(user),
	})
}

// UpdateUserRole godoc
// @Summary      Update user role
// @Description  Update a user's role
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id    path      int    true  "User ID"
// @Param        role  body      object true  "Role update"
// @Success      200   {object}  dto.UserDTO
// @Failure      400   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Router       /users/{id}/role [put]
func (h *userHandler) UpdateUserRole(c *gin.Context) {
	id := c.Param("id")
	user, err := h.UserService.GetUserByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, responses.CommonResponse{
			ResponseCode:    http.StatusNotFound,
			ResponseMessage: "User not found",
		})
		return
	}
	var input struct {
		Role string `json:"role" validate:"required,oneof=admin editor moderator user guest"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, responses.CommonResponse{
			ResponseCode:    http.StatusBadRequest,
			ResponseMessage: "Invalid request",
		})
		return
	}
	if err := rules.Validate.Struct(input); err != nil {
		c.JSON(http.StatusBadRequest, responses.CommonResponse{
			ResponseCode:    http.StatusBadRequest,
			ResponseMessage: "Validation error",
			Data:            requests.FormatValidationError(err),
		})
		return
	}
	roleName := input.Role
	if roleName == "" {
		roleName = "user"
	}
	var role models.Role
	if err := database.DB.Where("name = ?", roleName).First(&role).Error; err != nil {
		c.JSON(http.StatusBadRequest, responses.CommonResponse{
			ResponseCode:    http.StatusBadRequest,
			ResponseMessage: "Invalid or missing role",
		})
		return
	}
	user.Roles = []models.Role{role}
	c.JSON(http.StatusOK, responses.CommonResponse{
		ResponseCode:    http.StatusOK,
		ResponseMessage: "User role updated successfully",
		Data:            dto.ToUserDTO(user),
	})
}

// UserCreate godoc
// @Summary      Create a new user
// @Description  Create a new user with the provided details
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        user  body      requests.AuthRequest true  "User creation details"
// @Success      201   {object}  dto.UserDTO
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /users [post]
func (h *userHandler) UserCreate(c *gin.Context) {
	var input requests.AuthRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, responses.CommonResponse{
			ResponseCode:    http.StatusBadRequest,
			ResponseMessage: "Invalid request",
		})
		return
	}
	if err := rules.Validate.Struct(input); err != nil {
		c.JSON(http.StatusBadRequest, responses.CommonResponse{
			ResponseCode:    http.StatusBadRequest,
			ResponseMessage: "Validation error",
			Data:            requests.FormatValidationError(err),
		})
		return
	}
	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		roleName := input.Role
		if roleName == "" {
			roleName = "user"
		}
		var role models.Role
		if err := tx.Where("name = ?", roleName).First(&role).Error; err != nil {
			c.JSON(http.StatusBadRequest, responses.CommonResponse{
				ResponseCode:    http.StatusBadRequest,
				ResponseMessage: "Invalid or missing role",
			})
			return err
		}
		user := &models.User{
			Username: input.Username,
			Email:    input.Email,
			Roles:    []models.Role{role},
		}
		if err := user.HashPassword(input.Password); err != nil {
			c.JSON(http.StatusInternalServerError, responses.CommonResponse{
				ResponseCode:    http.StatusInternalServerError,
				ResponseMessage: "Failed to hash password",
			})
			return err
		}
		if err := tx.Create(user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, responses.CommonResponse{
				ResponseCode:    http.StatusInternalServerError,
				ResponseMessage: "Failed to create user",
			})
			return err
		}
		c.JSON(http.StatusCreated, responses.CommonResponse{
			ResponseCode:    http.StatusCreated,
			ResponseMessage: "User created successfully",
			Data:            dto.ToUserDTO(user),
		})
		return nil
	}); err != nil {
		return
	}
}
