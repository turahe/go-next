package middleware

import (
	"net/http"
	"reflect"
	"strings"

	"go-next/internal/http/responses"
	"go-next/internal/rules"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ValidationMiddleware handles validation errors globally
func ValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any validation errors in the context
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				if validationErr, ok := err.Err.(validator.ValidationErrors); ok {
					responses.SendValidationError(c, validationErr)
					return
				}
			}
		}
	}
}

// ValidateRequest is a helper function to validate request structs
func ValidateRequest(c *gin.Context, request interface{}) bool {
	if err := c.ShouldBindJSON(request); err != nil {
		if validationErr, ok := err.(validator.ValidationErrors); ok {
			responses.SendValidationError(c, validationErr)
			return false
		}
		// Handle other binding errors (malformed JSON, etc.)
		responses.SendError(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return false
	}

	// Additional struct validation
	if err := rules.Validate.Struct(request); err != nil {
		if validationErr, ok := err.(validator.ValidationErrors); ok {
			responses.SendValidationError(c, validationErr)
			return false
		}
		responses.SendError(c, http.StatusBadRequest, "Validation failed", err.Error())
		return false
	}

	return true
}

// ValidateRequestPartial validates only specific fields of a request struct
func ValidateRequestPartial(c *gin.Context, request interface{}, fields ...string) bool {
	if err := c.ShouldBindJSON(request); err != nil {
		if validationErr, ok := err.(validator.ValidationErrors); ok {
			responses.SendValidationError(c, validationErr)
			return false
		}
		responses.SendError(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return false
	}

	// Partial struct validation
	if err := rules.Validate.StructPartial(request, fields...); err != nil {
		if validationErr, ok := err.(validator.ValidationErrors); ok {
			responses.SendValidationError(c, validationErr)
			return false
		}
		responses.SendError(c, http.StatusBadRequest, "Validation failed", err.Error())
		return false
	}

	return true
}

// ValidateQuery validates query parameters
func ValidateQuery(c *gin.Context, request interface{}) bool {
	if err := c.ShouldBindQuery(request); err != nil {
		if validationErr, ok := err.(validator.ValidationErrors); ok {
			responses.SendValidationError(c, validationErr)
			return false
		}
		responses.SendError(c, http.StatusBadRequest, "Invalid query parameters", err.Error())
		return false
	}

	if err := rules.Validate.Struct(request); err != nil {
		if validationErr, ok := err.(validator.ValidationErrors); ok {
			responses.SendValidationError(c, validationErr)
			return false
		}
		responses.SendError(c, http.StatusBadRequest, "Validation failed", err.Error())
		return false
	}

	return true
}

// ValidateForm validates form data
func ValidateForm(c *gin.Context, request interface{}) bool {
	if err := c.ShouldBind(request); err != nil {
		if validationErr, ok := err.(validator.ValidationErrors); ok {
			responses.SendValidationError(c, validationErr)
			return false
		}
		responses.SendError(c, http.StatusBadRequest, "Invalid form data", err.Error())
		return false
	}

	if err := rules.Validate.Struct(request); err != nil {
		if validationErr, ok := err.(validator.ValidationErrors); ok {
			responses.SendValidationError(c, validationErr)
			return false
		}
		responses.SendError(c, http.StatusBadRequest, "Validation failed", err.Error())
		return false
	}

	return true
}

// ValidateUUID validates if a string is a valid UUID
func ValidateUUID(c *gin.Context, paramName string) bool {
	param := c.Param(paramName)
	if param == "" {
		responses.SendError(c, http.StatusBadRequest, "Missing required parameter: "+paramName)
		return false
	}

	// Basic UUID format validation (you can use a more robust UUID library if needed)
	if len(param) != 36 || !strings.Contains(param, "-") {
		responses.SendError(c, http.StatusBadRequest, "Invalid UUID format: "+paramName)
		return false
	}

	return true
}

// ValidateRequiredFields validates that required fields are present in the request
func ValidateRequiredFields(c *gin.Context, fields ...string) bool {
	var missingFields []string

	for _, field := range fields {
		value := c.PostForm(field)
		if value == "" {
			// Try JSON field
			if jsonValue, exists := c.Get(field); !exists || jsonValue == nil {
				missingFields = append(missingFields, field)
			}
		}
	}

	if len(missingFields) > 0 {
		responses.SendError(c, http.StatusBadRequest, "Missing required fields: "+strings.Join(missingFields, ", "))
		return false
	}

	return true
}

// ValidateStruct validates a struct and returns validation errors
func ValidateStruct(c *gin.Context, obj interface{}) bool {
	if err := rules.Validate.Struct(obj); err != nil {
		if validationErr, ok := err.(validator.ValidationErrors); ok {
			responses.SendValidationError(c, validationErr)
			return false
		}
		responses.SendError(c, http.StatusBadRequest, "Validation failed", err.Error())
		return false
	}
	return true
}

// ValidateStructPartial validates specific fields of a struct
func ValidateStructPartial(c *gin.Context, obj interface{}, fields ...string) bool {
	if err := rules.Validate.StructPartial(obj, fields...); err != nil {
		if validationErr, ok := err.(validator.ValidationErrors); ok {
			responses.SendValidationError(c, validationErr)
			return false
		}
		responses.SendError(c, http.StatusBadRequest, "Validation failed", err.Error())
		return false
	}
	return true
}

// ValidateVar validates a single variable
func ValidateVar(c *gin.Context, value interface{}, tag string) bool {
	if err := rules.Validate.Var(value, tag); err != nil {
		if validationErr, ok := err.(validator.ValidationErrors); ok {
			responses.SendValidationError(c, validationErr)
			return false
		}
		responses.SendError(c, http.StatusBadRequest, "Validation failed", err.Error())
		return false
	}
	return true
}

// ValidateSlice validates a slice of values
func ValidateSlice(c *gin.Context, slice interface{}, tag string) bool {
	if err := rules.Validate.Var(slice, tag); err != nil {
		if validationErr, ok := err.(validator.ValidationErrors); ok {
			responses.SendValidationError(c, validationErr)
			return false
		}
		responses.SendError(c, http.StatusBadRequest, "Validation failed", err.Error())
		return false
	}
	return true
}

// ValidateMap validates a map of values
func ValidateMap(c *gin.Context, m interface{}, tag string) bool {
	if err := rules.Validate.Var(m, tag); err != nil {
		if validationErr, ok := err.(validator.ValidationErrors); ok {
			responses.SendValidationError(c, validationErr)
			return false
		}
		responses.SendError(c, http.StatusBadRequest, "Validation failed", err.Error())
		return false
	}
	return true
}

// ValidateConditional validates a struct conditionally based on a condition
func ValidateConditional(c *gin.Context, obj interface{}, condition func() bool, fields ...string) bool {
	if !condition() {
		return true // Skip validation if condition is not met
	}

	if len(fields) > 0 {
		return ValidateStructPartial(c, obj, fields...)
	}
	return ValidateStruct(c, obj)
}

// ValidateNested validates nested structs
func ValidateNested(c *gin.Context, obj interface{}) bool {
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		responses.SendError(c, http.StatusBadRequest, "Object must be a struct")
		return false
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Skip unexported fields
		if !field.CanInterface() {
			continue
		}

		// Validate nested structs
		if field.Kind() == reflect.Struct && fieldType.Anonymous == false {
			if !ValidateStruct(c, field.Interface()) {
				return false
			}
		}

		// Validate slices of structs
		if field.Kind() == reflect.Slice {
			for j := 0; j < field.Len(); j++ {
				sliceElement := field.Index(j)
				if sliceElement.Kind() == reflect.Struct {
					if !ValidateStruct(c, sliceElement.Interface()) {
						return false
					}
				}
			}
		}
	}

	return ValidateStruct(c, obj)
}
