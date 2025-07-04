package requests

import (
	"github.com/go-playground/validator/v10"
)

type ValidationErrorResponse struct {
	Message string              `json:"message"`
	Errors  map[string][]string `json:"errors"`
}

func FormatValidationError(err error) ValidationErrorResponse {
	res := ValidationErrorResponse{
		Message: "Validation failed",
		Errors:  map[string][]string{},
	}
	if errs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range errs {
			field := e.Field()
			var msg string
			switch e.Tag() {
			case "required":
				msg = "The " + field + " field is required."
			case "email":
				msg = "The " + field + " must be a valid email address."
			case "min":
				msg = "The " + field + " must be at least " + e.Param() + " characters."
			case "max":
				msg = "The " + field + " may not be greater than " + e.Param() + " characters."
			case "oneof":
				msg = "The " + field + " must be one of: " + e.Param() + "."
			case "gt":
				msg = "The " + field + " must be greater than " + e.Param() + "."
			default:
				msg = "The " + field + " is invalid."
			}
			res.Errors[field] = append(res.Errors[field], msg)
		}
	}
	return res
}
