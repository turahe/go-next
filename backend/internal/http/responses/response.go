package responses

import (
	"fmt"
	"net/http"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type DataUnwrapper interface {
	UnwrapData(interface{}) error
}

// ValidationError represents a single validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

// ValidationErrorResponse represents the Laravel-style validation error response
type ValidationErrorResponse struct {
	Message   string            `json:"message"`
	Errors    []ValidationError `json:"errors"`
	Status    int               `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Path      string            `json:"path,omitempty"`
	RequestID string            `json:"request_id,omitempty"`
}

// ErrorResponse represents the structured error response for general errors
type ErrorResponse struct {
	Message   string    `json:"message"`
	Status    int       `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Path      string    `json:"path,omitempty"`
	RequestID string    `json:"request_id,omitempty"`
	Details   string    `json:"details,omitempty"`
}

// SuccessResponse represents a successful response
type SuccessResponse struct {
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Status    int         `json:"status"`
	Timestamp time.Time   `json:"timestamp"`
	Path      string      `json:"path,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
}

// Standard Response (keeping for backward compatibility)
type CommonResponse struct {
	ResponseCode    int    `json:"code"`
	ResponseMessage string `json:"message"`
	Data            any    `json:"data,omitempty"`
	RequestID       string `json:"request_id,omitempty"`
	Path            string `json:"path,omitempty"`
}

func (resp *CommonResponse) UnwrapData(target interface{}) error {
	bs, err := sonic.Marshal(resp.Data)
	if err != nil {
		return err
	}

	if err := sonic.Unmarshal(bs, target); err != nil {
		return err
	}

	return nil
}

func FormatValidationError(err error, c *gin.Context) ValidationErrorResponse {
	var validationErrors []ValidationError

	if errs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range errs {
			field := e.Field()
			value := e.Value()
			valueStr := ""
			if value != nil {
				valueStr = toString(value)
			}

			message := getValidationMessage(e.Tag(), field, e.Param())

			validationErrors = append(validationErrors, ValidationError{
				Field:   field,
				Message: message,
				Value:   valueStr,
			})
		}
	}

	// Get request context
	var requestID, path string
	if c != nil {
		if id, exists := c.Get("request_id"); exists {
			if idStr, ok := id.(string); ok {
				requestID = idStr
			}
		}
		path = c.Request.URL.Path
	}

	return ValidationErrorResponse{
		Message:   "The given data was invalid.",
		Errors:    validationErrors,
		Status:    http.StatusUnprocessableEntity, // 422 like Laravel
		Timestamp: time.Now(),
		Path:      path,
		RequestID: requestID,
	}
}

// SendValidationError sends a validation error response
func SendValidationError(c *gin.Context, err error) {
	response := FormatValidationError(err, c)
	c.JSON(response.Status, response)
}

// SendError sends a general error response
func SendError(c *gin.Context, status int, message string, details ...string) {
	var requestID, path string
	if id, exists := c.Get("request_id"); exists {
		if idStr, ok := id.(string); ok {
			requestID = idStr
		}
	}
	path = c.Request.URL.Path

	response := ErrorResponse{
		Message:   message,
		Status:    status,
		Timestamp: time.Now(),
		Path:      path,
		RequestID: requestID,
	}

	if len(details) > 0 {
		response.Details = details[0]
	}

	c.JSON(status, response)
}

// SendSuccess sends a success response
func SendSuccess(c *gin.Context, status int, message string, data interface{}) {
	var requestID, path string
	if id, exists := c.Get("request_id"); exists {
		if idStr, ok := id.(string); ok {
			requestID = idStr
		}
	}
	path = c.Request.URL.Path

	response := SuccessResponse{
		Message:   message,
		Data:      data,
		Status:    status,
		Timestamp: time.Now(),
		Path:      path,
		RequestID: requestID,
	}

	c.JSON(status, response)
}

// Helper function to convert value to string
func toString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%f", v)
	case bool:
		return fmt.Sprintf("%t", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// getValidationMessage returns Laravel-style validation messages
func getValidationMessage(tag, field, param string) string {
	switch tag {
	case "required":
		return "The " + field + " field is required."
	case "email":
		return "The " + field + " must be a valid email address."
	case "min":
		return "The " + field + " must be at least " + param + " characters."
	case "max":
		return "The " + field + " may not be greater than " + param + " characters."
	case "oneof":
		return "The selected " + field + " is invalid."
	case "gt":
		return "The " + field + " must be greater than " + param + "."
	case "gte":
		return "The " + field + " must be greater than or equal to " + param + "."
	case "lt":
		return "The " + field + " must be less than " + param + "."
	case "lte":
		return "The " + field + " must be less than or equal to " + param + "."
	case "numeric":
		return "The " + field + " must be a number."
	case "alpha":
		return "The " + field + " may only contain letters."
	case "alphanum":
		return "The " + field + " may only contain letters and numbers."
	case "url":
		return "The " + field + " format is invalid."
	case "uuid":
		return "The " + field + " must be a valid UUID."
	case "unique":
		return "The " + field + " has already been taken."
	case "exists":
		return "The selected " + field + " is invalid."
	case "confirmed":
		return "The " + field + " confirmation does not match."
	case "different":
		return "The " + field + " and " + param + " must be different."
	case "same":
		return "The " + field + " and " + param + " must match."
	case "after":
		return "The " + field + " must be a date after " + param + "."
	case "before":
		return "The " + field + " must be a date before " + param + "."
	case "date":
		return "The " + field + " is not a valid date."
	case "json":
		return "The " + field + " must be a valid JSON string."
	case "file":
		return "The " + field + " must be a file."
	case "image":
		return "The " + field + " must be an image."
	case "mimes":
		return "The " + field + " must be a file of type: " + param + "."
	case "size":
		return "The " + field + " may not be greater than " + param + " kilobytes."
	default:
		return "The " + field + " field is invalid."
	}
}
