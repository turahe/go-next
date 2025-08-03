package validation

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/nyaruka/phonenumbers"
)

// Validator represents a Laravel-style validator for structs
type Validator struct {
	validate *validator.Validate
	errors   map[string][]string
	messages map[string]string
}

// NewValidator creates a new Laravel-style validator
func NewValidator() *Validator {
	v := validator.New()

	// Register custom validations
	registerCustomValidations(v)

	// Register unique validation functions
	uniqueValidator := NewUniqueValidator()
	uniqueValidator.RegisterUniqueValidations(v)

	return &Validator{
		validate: v,
		errors:   make(map[string][]string),
		messages: make(map[string]string),
	}
}

// ValidationResult represents the result of validation
type ValidationResult struct {
	IsValid bool
	Errors  map[string][]string
}

// Validate validates a struct with Laravel-style rules
func (lv *Validator) Validate(data interface{}) *ValidationResult {
	lv.errors = make(map[string][]string)

	// Get validation rules from struct tags
	rules := lv.extractRulesFromStruct(data)

	// Apply Laravel-style rules
	lv.ApplyValidation(data, rules)

	// Validate with validator
	err := lv.validate.Struct(data)
	if err != nil {
		lv.processValidationErrors(err)
	}

	return &ValidationResult{
		IsValid: len(lv.errors) == 0,
		Errors:  lv.errors,
	}
}

// ValidateWithRules validates with custom rules
func (lv *Validator) ValidateWithRules(data interface{}, rules map[string]string) *ValidationResult {
	lv.errors = make(map[string][]string)

	// Apply custom rules
	lv.ApplyValidation(data, rules)

	// Validate with validator
	err := lv.validate.Struct(data)
	if err != nil {
		lv.processValidationErrors(err)
	}

	return &ValidationResult{
		IsValid: len(lv.errors) == 0,
		Errors:  lv.errors,
	}
}

// AddCustomMessage adds a custom error message
func (lv *Validator) AddCustomMessage(field, rule, message string) {
	key := fmt.Sprintf("%s.%s", field, rule)
	lv.messages[key] = message
}

// AddCustomMessages adds multiple custom error messages
func (lv *Validator) AddCustomMessages(messages map[string]string) {
	for key, message := range messages {
		lv.messages[key] = message
	}
}

// extractRulesFromStruct extracts validation rules from struct tags
func (lv *Validator) extractRulesFromStruct(data interface{}) map[string]string {
	rules := make(map[string]string)

	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")
		validateTag := field.Tag.Get("validate")

		if jsonTag != "" && validateTag != "" {
			// Extract field name from json tag
			fieldName := strings.Split(jsonTag, ",")[0]
			if fieldName == "" {
				fieldName = field.Name
			}

			rules[fieldName] = validateTag
		}
	}

	return rules
}

// ApplyValidation applies Laravel-style rules to the struct
func (lv *Validator) ApplyValidation(data interface{}, rules map[string]string) {
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")

		if jsonTag != "" {
			fieldName := strings.Split(jsonTag, ",")[0]
			if fieldName == "" {
				fieldName = field.Name
			}

			if rule, exists := rules[fieldName]; exists {
				lv.validateField(v.Field(i), fieldName, rule)
			}
		}
	}
}

// validateField validates a single field
func (lv *Validator) validateField(field reflect.Value, fieldName, ruleString string) {
	rules := strings.Split(ruleString, "|")

	for _, rule := range rules {
		rule = strings.TrimSpace(rule)
		if rule == "" {
			continue
		}

		ruleName, params := lv.parseRule(rule)
		if !lv.validateFieldRule(field, fieldName, ruleName, params) {
			message := lv.getErrorMessage(fieldName, ruleName, params)
			lv.AddError(fieldName, message)
		}
	}
}

// validateFieldRule validates a single rule for a field
func (lv *Validator) validateFieldRule(field reflect.Value, fieldName, ruleName string, params []string) bool {
	switch ruleName {
	case "required":
		return lv.validateRequired(field)
	case "email":
		return lv.validateEmail(field)
	case "url":
		return lv.validateURL(field)
	case "numeric":
		return lv.validateNumeric(field)
	case "integer":
		return lv.validateInteger(field)
	case "string":
		return lv.validateString(field)
	case "min":
		return lv.validateMin(field, params)
	case "max":
		return lv.validateMax(field, params)
	case "between":
		return lv.validateBetween(field, params)
	case "size":
		return lv.validateSize(field, params)
	case "alpha":
		return lv.validateAlpha(field)
	case "alpha_num":
		return lv.validateAlphaNum(field)
	case "alpha_dash":
		return lv.validateAlphaDash(field)
	case "date":
		return lv.validateDate(field)
	case "json":
		return lv.validateJSON(field)
	case "ip":
		return lv.validateIP(field)
	case "uuid":
		return lv.validateUUID(field)
	}

	return true
}

// parseRule parses a single rule with parameters
func (lv *Validator) parseRule(rule string) (string, []string) {
	parts := strings.SplitN(rule, ":", 2)
	ruleName := parts[0]

	var params []string
	if len(parts) > 1 {
		params = strings.Split(parts[1], ",")
		for i, param := range params {
			params[i] = strings.TrimSpace(param)
		}
	}

	return ruleName, params
}

// validateRequired validates required field
func (lv *Validator) validateRequired(field reflect.Value) bool {
	if !field.IsValid() {
		return false
	}

	switch field.Kind() {
	case reflect.String:
		return field.String() != ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return field.Int() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return field.Uint() != 0
	case reflect.Float32, reflect.Float64:
		return field.Float() != 0
	case reflect.Bool:
		return true
	case reflect.Slice, reflect.Array:
		return field.Len() > 0
	case reflect.Map:
		return field.Len() > 0
	case reflect.Ptr:
		return !field.IsNil()
	}

	return true
}

// validateEmail validates email field
func (lv *Validator) validateEmail(field reflect.Value) bool {
	if field.Kind() != reflect.String {
		return false
	}

	email := field.String()
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

// validateURL validates URL field
func (lv *Validator) validateURL(field reflect.Value) bool {
	if field.Kind() != reflect.String {
		return false
	}

	url := field.String()
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
}

// validateNumeric validates numeric field
func (lv *Validator) validateNumeric(field reflect.Value) bool {
	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	case reflect.String:
		// Try to parse as number
		_, err := fmt.Sscanf(field.String(), "%f", new(float64))
		return err == nil
	}

	return false
}

// validateInteger validates integer field
func (lv *Validator) validateInteger(field reflect.Value) bool {
	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	case reflect.String:
		// Try to parse as integer
		_, err := fmt.Sscanf(field.String(), "%d", new(int))
		return err == nil
	}

	return false
}

// validateString validates string field
func (lv *Validator) validateString(field reflect.Value) bool {
	return field.Kind() == reflect.String
}

// validateMin validates minimum value/length
func (lv *Validator) validateMin(field reflect.Value, params []string) bool {
	if len(params) == 0 {
		return true
	}

	var min float64
	_, err := fmt.Sscanf(params[0], "%f", &min)
	if err != nil {
		return true
	}

	switch field.Kind() {
	case reflect.String:
		return float64(len(field.String())) >= min
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(field.Int()) >= min
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(field.Uint()) >= min
	case reflect.Float32, reflect.Float64:
		return field.Float() >= min
	case reflect.Slice, reflect.Array:
		return float64(field.Len()) >= min
	}

	return true
}

// validateMax validates maximum value/length
func (lv *Validator) validateMax(field reflect.Value, params []string) bool {
	if len(params) == 0 {
		return true
	}

	var max float64
	_, err := fmt.Sscanf(params[0], "%f", &max)
	if err != nil {
		return true
	}

	switch field.Kind() {
	case reflect.String:
		return float64(len(field.String())) <= max
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(field.Int()) <= max
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(field.Uint()) <= max
	case reflect.Float32, reflect.Float64:
		return field.Float() <= max
	case reflect.Slice, reflect.Array:
		return float64(field.Len()) <= max
	}

	return true
}

// validateBetween validates between range
func (lv *Validator) validateBetween(field reflect.Value, params []string) bool {
	if len(params) < 2 {
		return true
	}

	var min, max float64
	_, err1 := fmt.Sscanf(params[0], "%f", &min)
	_, err2 := fmt.Sscanf(params[1], "%f", &max)
	if err1 != nil || err2 != nil {
		return true
	}

	switch field.Kind() {
	case reflect.String:
		length := float64(len(field.String()))
		return length >= min && length <= max
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value := float64(field.Int())
		return value >= min && value <= max
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		value := float64(field.Uint())
		return value >= min && value <= max
	case reflect.Float32, reflect.Float64:
		value := field.Float()
		return value >= min && value <= max
	case reflect.Slice, reflect.Array:
		length := float64(field.Len())
		return length >= min && length <= max
	}

	return true
}

// validateSize validates exact size
func (lv *Validator) validateSize(field reflect.Value, params []string) bool {
	if len(params) == 0 {
		return true
	}

	var size float64
	_, err := fmt.Sscanf(params[0], "%f", &size)
	if err != nil {
		return true
	}

	switch field.Kind() {
	case reflect.String:
		return float64(len(field.String())) == size
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(field.Int()) == size
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(field.Uint()) == size
	case reflect.Float32, reflect.Float64:
		return field.Float() == size
	case reflect.Slice, reflect.Array:
		return float64(field.Len()) == size
	}

	return true
}

// validateAlpha validates alpha characters
func (lv *Validator) validateAlpha(field reflect.Value) bool {
	if field.Kind() != reflect.String {
		return false
	}

	str := field.String()
	for _, char := range str {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z')) {
			return false
		}
	}

	return true
}

// validateAlphaNum validates alphanumeric characters
func (lv *Validator) validateAlphaNum(field reflect.Value) bool {
	if field.Kind() != reflect.String {
		return false
	}

	str := field.String()
	for _, char := range str {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9')) {
			return false
		}
	}

	return true
}

// validateUsername validates username format (alphanumeric and underscore)
func (lv *Validator) validateUsername(field reflect.Value) bool {
	if field.Kind() != reflect.String {
		return false
	}

	str := field.String()
	for _, char := range str {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '_') {
			return false
		}
	}

	return true
}

// validateAlphaDash validates alphanumeric and dash characters
func (lv *Validator) validateAlphaDash(field reflect.Value) bool {
	if field.Kind() != reflect.String {
		return false
	}

	str := field.String()
	for _, char := range str {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '-' || char == '_') {
			return false
		}
	}

	return true
}

// validateDate validates date field
func (lv *Validator) validateDate(field reflect.Value) bool {
	if field.Kind() != reflect.String {
		return false
	}

	// Simple date validation - you might want to use a proper date parsing library
	dateStr := field.String()
	return len(dateStr) > 0
}

// validateJSON validates JSON field
func (lv *Validator) validateJSON(field reflect.Value) bool {
	if field.Kind() != reflect.String {
		return false
	}

	jsonStr := field.String()
	return strings.HasPrefix(jsonStr, "{") || strings.HasPrefix(jsonStr, "[")
}

// validateIP validates IP address field
func (lv *Validator) validateIP(field reflect.Value) bool {
	if field.Kind() != reflect.String {
		return false
	}

	ip := field.String()
	return strings.Contains(ip, ".") || strings.Contains(ip, ":")
}

// validateUUID validates UUID field
func (lv *Validator) validateUUID(field reflect.Value) bool {
	if field.Kind() != reflect.String {
		return false
	}

	uuid := field.String()
	return len(uuid) == 36 && strings.Count(uuid, "-") == 4
}

// AddError adds a validation error
func (lv *Validator) AddError(field, message string) {
	if lv.errors[field] == nil {
		lv.errors[field] = []string{}
	}
	lv.errors[field] = append(lv.errors[field], message)
}

// getErrorMessage gets the error message for a field and rule
func (lv *Validator) getErrorMessage(field, rule string, params []string) string {
	// Check for custom message
	key := fmt.Sprintf("%s.%s", field, rule)
	if message, exists := lv.messages[key]; exists {
		return message
	}

	// Default messages
	messages := map[string]string{
		"required":   fmt.Sprintf("The %s field is required.", field),
		"email":      fmt.Sprintf("The %s field must be a valid email address.", field),
		"url":        fmt.Sprintf("The %s field must be a valid URL.", field),
		"numeric":    fmt.Sprintf("The %s field must be a number.", field),
		"integer":    fmt.Sprintf("The %s field must be an integer.", field),
		"string":     fmt.Sprintf("The %s field must be a string.", field),
		"alpha":      fmt.Sprintf("The %s field must only contain letters.", field),
		"alpha_num":  fmt.Sprintf("The %s field must only contain letters and numbers.", field),
		"alpha_dash": fmt.Sprintf("The %s field must only contain letters, numbers, dashes and underscores.", field),
		"date":       fmt.Sprintf("The %s field must be a valid date.", field),
		"json":       fmt.Sprintf("The %s field must be a valid JSON string.", field),
		"ip":         fmt.Sprintf("The %s field must be a valid IP address.", field),
		"uuid":       fmt.Sprintf("The %s field must be a valid UUID.", field),
	}

	if message, exists := messages[rule]; exists {
		if len(params) > 0 {
			message = strings.Replace(message, ":min", params[0], -1)
			if len(params) > 1 {
				message = strings.Replace(message, ":max", params[1], -1)
			}
		}
		return message
	}

	return fmt.Sprintf("The %s field is invalid.", field)
}

// processValidationErrors processes validation errors from validator
func (lv *Validator) processValidationErrors(err error) {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			field := e.Field()
			tag := e.Tag()
			param := e.Param()

			params := []string{}
			if param != "" {
				params = []string{param}
			}

			message := lv.getErrorMessage(field, tag, params)
			lv.AddError(field, message)
		}
	}
}

// validateAlpha validates alphabetic characters
func validateAlpha(fl validator.FieldLevel) bool {
	if fl.Field().Kind() != reflect.String {
		return false
	}

	str := fl.Field().String()
	for _, char := range str {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z')) {
			return false
		}
	}

	return true
}

// validateAlphaNum validates alphanumeric characters
func validateAlphaNum(fl validator.FieldLevel) bool {
	if fl.Field().Kind() != reflect.String {
		return false
	}

	str := fl.Field().String()
	for _, char := range str {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9')) {
			return false
		}
	}

	return true
}

// validateAlphaDash validates alphanumeric and dash characters
func validateAlphaDash(fl validator.FieldLevel) bool {
	if fl.Field().Kind() != reflect.String {
		return false
	}

	str := fl.Field().String()
	for _, char := range str {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '-' || char == '_') {
			return false
		}
	}

	return true
}

// validateUsername validates username format (alphanumeric and underscore)
func validateUsername(fl validator.FieldLevel) bool {
	if fl.Field().Kind() != reflect.String {
		return false
	}

	str := fl.Field().String()
	for _, char := range str {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '_') {
			return false
		}
	}

	return true
}

// validateCountryCode validates country code using phonenumbers library
func validateCountryCode(fl validator.FieldLevel) bool {
	if fl.Field().Kind() != reflect.String {
		return false
	}

	countryCode := fl.Field().String()
	if countryCode == "" {
		return true // Empty country codes are handled by required validation
	}

	// Normalize country code to uppercase
	countryCode = strings.ToUpper(countryCode)

	// Check if the country code is valid using phonenumbers library
	// We'll try to parse a dummy number with this country code
	dummyNumber := "1234567890"
	_, err := phonenumbers.Parse(dummyNumber, countryCode)

	// If parsing succeeds, the country code is valid
	return err == nil
}

// registerCustomValidations registers custom validation functions
func registerCustomValidations(v *validator.Validate) {
	// Register custom validation functions
	v.RegisterValidation("alpha", validateAlpha)
	v.RegisterValidation("alpha_num", validateAlphaNum)
	v.RegisterValidation("alpha_dash", validateAlphaDash)
	v.RegisterValidation("username", validateUsername)
	v.RegisterValidation("country_code", validateCountryCode)
}
