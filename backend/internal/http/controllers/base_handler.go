package controllers

import (
	"fmt"
	"net/http"
	"strconv"

	"go-next/internal/http/responses"
	"go-next/pkg/database"
	"go-next/pkg/validation"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// BaseHandler provides common functionality for all handlers
type BaseHandler struct {
	logger    *zap.Logger
	validator *validation.Validator
}

// NewBaseHandler creates a new base handler
func NewBaseHandler(logger *zap.Logger) *BaseHandler {
	return &BaseHandler{
		logger:    logger,
		validator: validation.NewValidator(),
	}
}

// GetUserIDFromContext safely extracts user ID from gin context
func (h *BaseHandler) GetUserIDFromContext(c *gin.Context) (uuid.UUID, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, responses.NewError("User not authenticated", http.StatusUnauthorized)
	}

	switch v := userID.(type) {
	case uuid.UUID:
		return v, nil
	case string:
		return uuid.Parse(v)
	default:
		return uuid.Nil, responses.NewError("Invalid user ID format", http.StatusBadRequest)
	}
}

// GetParamAsUUID safely extracts UUID parameter from URL
func (h *BaseHandler) GetParamAsUUID(c *gin.Context, paramName string) (uuid.UUID, error) {
	param := c.Param(paramName)
	if param == "" {
		return uuid.Nil, responses.NewError("Missing "+paramName+" parameter", http.StatusBadRequest)
	}

	id, err := uuid.Parse(param)
	if err != nil {
		return uuid.Nil, responses.NewError("Invalid "+paramName+" format", http.StatusBadRequest)
	}

	return id, nil
}

// GetParamAsInt safely extracts integer parameter from URL
func (h *BaseHandler) GetParamAsInt(c *gin.Context, paramName string) (int, error) {
	param := c.Param(paramName)
	if param == "" {
		return 0, responses.NewError("Missing "+paramName+" parameter", http.StatusBadRequest)
	}

	val, err := strconv.Atoi(param)
	if err != nil {
		return 0, responses.NewError("Invalid "+paramName+" format", http.StatusBadRequest)
	}

	return val, nil
}

// GetQueryAsInt safely extracts integer query parameter
func (h *BaseHandler) GetQueryAsInt(c *gin.Context, queryName string, defaultValue int) int {
	query := c.Query(queryName)
	if query == "" {
		return defaultValue
	}

	val, err := strconv.Atoi(query)
	if err != nil {
		return defaultValue
	}

	return val
}

// HandleServiceError standardizes error handling for service layer errors
func (h *BaseHandler) HandleServiceError(c *gin.Context, err error, operation string) {
	if err == nil {
		return
	}

	h.logger.Error("Service error",
		zap.String("operation", operation),
		zap.Error(err),
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
	)

	// Check for specific error types
	switch {
	case responses.IsNotFoundError(err):
		responses.SendError(c, http.StatusNotFound, err.Error())
	case responses.IsValidationError(err):
		responses.SendError(c, http.StatusBadRequest, err.Error())
	case responses.IsUnauthorizedError(err):
		responses.SendError(c, http.StatusUnauthorized, err.Error())
	case responses.IsForbiddenError(err):
		responses.SendError(c, http.StatusForbidden, err.Error())
	default:
		responses.SendError(c, http.StatusInternalServerError, "Internal server error")
	}
}

// HandleDatabaseError handles database-specific errors
func (h *BaseHandler) HandleDatabaseError(c *gin.Context, err error, operation string) {
	if err == nil {
		return
	}

	h.logger.Error("Database error",
		zap.String("operation", operation),
		zap.Error(err),
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
	)

	// Check for specific database errors
	switch {
	case database.IsNotFoundError(err):
		responses.SendError(c, http.StatusNotFound, "Resource not found")
	case database.IsDuplicateError(err):
		responses.SendError(c, http.StatusConflict, "Resource already exists")
	case database.IsConstraintError(err):
		responses.SendError(c, http.StatusBadRequest, "Invalid data provided")
	default:
		responses.SendError(c, http.StatusInternalServerError, "Database error")
	}
}

// LogRequest logs incoming requests for debugging
func (h *BaseHandler) LogRequest(c *gin.Context, operation string) {
	h.logger.Info("Request received",
		zap.String("operation", operation),
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
		zap.String("user_agent", c.Request.UserAgent()),
		zap.String("remote_addr", c.ClientIP()),
	)
}

// LogResponse logs response information
func (h *BaseHandler) LogResponse(c *gin.Context, operation string, status int, duration int64) {
	h.logger.Info("Response sent",
		zap.String("operation", operation),
		zap.Int("status", status),
		zap.Int64("duration_ms", duration),
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
	)
}

// ValidateOwnership checks if the current user owns the resource
func (h *BaseHandler) ValidateOwnership(c *gin.Context, resourceUserID uuid.UUID) error {
	currentUserID, err := h.GetUserIDFromContext(c)
	if err != nil {
		return err
	}

	if currentUserID != resourceUserID {
		return responses.NewError("Access denied", http.StatusForbidden)
	}

	return nil
}

// ValidatePermission checks if the current user has the required permission
func (h *BaseHandler) ValidatePermission(c *gin.Context, permission string) error {
	// This would integrate with your permission system (Casbin, etc.)
	// For now, we'll just check if user is authenticated
	_, err := h.GetUserIDFromContext(c)
	if err != nil {
		return err
	}

	// TODO: Implement actual permission checking
	// This is a placeholder for the permission system
	return nil
}

// ValidateRequestParams validates request parameters using Laravel-style validation
func (h *BaseHandler) ValidateRequestParams(c *gin.Context, request interface{}) error {
	// Bind JSON to request struct
	if err := c.ShouldBindJSON(&request); err != nil {
		return err
	}

	// Validate the request
	result := h.validator.Validate(request)
	if !result.IsValid {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"message": "Validation failed",
			"errors":  result.Errors,
		})
		return fmt.Errorf("validation failed")
	}

	return nil
}

// ValidateRequestWithRules validates request with custom rules
func (h *BaseHandler) ValidateRequestWithRules(c *gin.Context, request interface{}, rules map[string]string) error {
	// Bind JSON to request struct
	if err := c.ShouldBindJSON(&request); err != nil {
		return err
	}

	// Validate with custom rules
	result := h.validator.ValidateWithRules(request, rules)
	if !result.IsValid {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"message": "Validation failed",
			"errors":  result.Errors,
		})
		return fmt.Errorf("validation failed")
	}

	return nil
}

// AddCustomValidationMessage adds a custom validation message
func (h *BaseHandler) AddCustomValidationMessage(field, rule, message string) {
	h.validator.AddCustomMessage(field, rule, message)
}

// AddCustomValidationMessages adds multiple custom validation messages
func (h *BaseHandler) AddCustomValidationMessages(messages map[string]string) {
	h.validator.AddCustomMessages(messages)
}
