package handler

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

var (
	passwordUppercaseRegex = regexp.MustCompile(`[A-Z]`)
	passwordLowercaseRegex = regexp.MustCompile(`[a-z]`)
	passwordNumberRegex    = regexp.MustCompile(`[0-9]`)
)

type registerRequest struct {
	Nickname string `json:"nickname" binding:"required,notblank,max=20"`
	RealName string `json:"realName" binding:"required,notblank,max=20"`
	Email    string `json:"email" binding:"required,notblank,email,max=254"`
	Password string `json:"password" binding:"required,min=8,max=72,passwd"`
	Birthday int64  `json:"birthday"`
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,notblank,email,max=254"`
	Password string `json:"password" binding:"required,notblank"`
}

type createPostRequest struct {
	Content string `json:"content" binding:"required,notblank,max=2000"`
}

type validationErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		_ = v.RegisterValidation("notblank", validateNotBlank)
		_ = v.RegisterValidation("passwd", validatePasswordStrength)
	}
}

func validateNotBlank(fl validator.FieldLevel) bool {
	return strings.TrimSpace(fl.Field().String()) != ""
}

func validatePasswordStrength(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	return passwordUppercaseRegex.MatchString(password) &&
		passwordLowercaseRegex.MatchString(password) &&
		passwordNumberRegex.MatchString(password)
}

func respondValidationError(c *gin.Context, err error, payload any) {
	c.JSON(http.StatusBadRequest, gin.H{
		"error":   "invalid request parameters",
		"details": buildValidationErrorDetails(err, payload),
	})
}

func buildValidationErrorDetails(err error, payload any) []validationErrorDetail {
	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		return []validationErrorDetail{
			{Field: "body", Message: "request body is invalid"},
		}
	}

	jsonFieldNames := extractJSONFieldNames(payload)
	details := make([]validationErrorDetail, 0, len(validationErrors))
	for _, fieldErr := range validationErrors {
		field := jsonFieldNames[fieldErr.StructField()]
		if field == "" {
			field = strings.ToLower(fieldErr.Field())
		}
		details = append(details, validationErrorDetail{
			Field:   field,
			Message: validationMessage(field, fieldErr),
		})
	}

	return details
}

func extractJSONFieldNames(payload any) map[string]string {
	fieldNames := make(map[string]string)

	t := reflect.TypeOf(payload)
	if t == nil {
		return fieldNames
	}
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return fieldNames
	}

	for i := range t.NumField() {
		field := t.Field(i)
		jsonTag := strings.Split(field.Tag.Get("json"), ",")[0]
		if jsonTag == "" {
			jsonTag = strings.ToLower(field.Name)
		}
		if jsonTag == "-" {
			continue
		}
		fieldNames[field.Name] = jsonTag
	}

	return fieldNames
}

func validationMessage(field string, fieldErr validator.FieldError) string {
	switch fieldErr.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "notblank":
		return fmt.Sprintf("%s cannot be blank", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, fieldErr.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", field, fieldErr.Param())
	case "passwd":
		return "password must include uppercase, lowercase, and number"
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}
