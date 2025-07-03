package middleware

import (
	"reflect"
	"strings"

	"dev.azure.com/clearpointhealth/ClearQuoteV3/_git/helpchat/internal/models"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// CustomValidator is a custom validator for Echo
type CustomValidator struct {
	validator *validator.Validate
}

// NewCustomValidator creates a new custom validator
func NewCustomValidator() *CustomValidator {
	v := validator.New()

	// Register custom validation for UserRole
	v.RegisterValidation("user_role", validateUserRole)

	return &CustomValidator{validator: v}
}

// Validate validates a struct
func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

// validateUserRole validates that a UserRole is one of the allowed values
func validateUserRole(fl validator.FieldLevel) bool {
	role := fl.Field().String()
	allowedRoles := []string{
		string(models.RoleEndUser),
		string(models.RoleSupportAgent),
		string(models.RoleAdministrator),
		string(models.RoleManager),
	}

	for _, allowedRole := range allowedRoles {
		if role == allowedRole {
			return true
		}
	}
	return false
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

// ValidationErrorResponse represents a validation error response
type ValidationErrorResponse struct {
	Status  string            `json:"status"`
	Message string            `json:"message"`
	Errors  []ValidationError `json:"errors"`
}

// ValidationMiddleware creates middleware that handles validation errors
func ValidationMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Set custom validator
			c.Echo().Validator = NewCustomValidator()

			// Continue to next handler
			return next(c)
		}
	}
}

// HandleValidationError handles validation errors and returns a proper response
func HandleValidationError(err error) *ValidationErrorResponse {
	var errors []ValidationError

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			field := strings.ToLower(e.Field())
			tag := e.Tag()
			value := e.Value()

			message := getValidationMessage(field, tag, value)

			errors = append(errors, ValidationError{
				Field:   field,
				Tag:     tag,
				Value:   toString(value),
				Message: message,
			})
		}
	}

	return &ValidationErrorResponse{
		Status:  "error",
		Message: "Validation failed",
		Errors:  errors,
	}
}

// getValidationMessage returns a user-friendly validation message
func getValidationMessage(field, tag, value interface{}) string {
	fieldStr := toString(field)
	tagStr := toString(tag)

	switch tagStr {
	case "required":
		return fieldStr + " is required"
	case "email":
		return fieldStr + " must be a valid email address"
	case "min":
		return fieldStr + " must be at least " + toString(value) + " characters"
	case "max":
		return fieldStr + " must be at most " + toString(value) + " characters"
	case "user_role":
		return fieldStr + " must be one of: END_USER, SUPPORT_AGENT, ADMINISTRATOR, MANAGER"
	default:
		return fieldStr + " failed validation: " + tagStr
	}
}

// toString converts a value to string
func toString(value interface{}) string {
	if value == nil {
		return ""
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return string(rune(v.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return string(rune(v.Uint()))
	case reflect.Float32, reflect.Float64:
		return string(rune(int(v.Float())))
	default:
		return ""
	}
}
