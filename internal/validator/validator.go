package validator

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/CDavidSV/Iris-Chat-App-Backend/internal"
	"github.com/gofiber/fiber/v2"
)

type ValidationError struct {
	FieldName string
	Field     string
	Error     string
}

type ValidationResult struct {
	IsValid bool
	Errors  []ValidationError
}

func (v *ValidationResult) SendValidationError(c *fiber.Ctx) error {
	return c.Status(http.StatusBadRequest).JSON(map[string]any{
		"status":           http.StatusBadRequest,
		"validationErrors": v.Errors,
	})
}

func Validate(obj interface{}) (ValidationResult, error) {
	v := reflect.ValueOf(obj)

	result := ValidationResult{
		IsValid: true,
		Errors:  []ValidationError{},
	}

	// Iterate of the the structs fields and proceed to validate
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		tag := v.Type().Field(i).Tag.Get("validate")

		if tag == "" {
			continue
		}

		validationRules := strings.Split(tag, ",")
		for _, rule := range validationRules {
			err := applyValidationRule(rule, &result, field, v.Type().Field(i).Name)
			if err != nil {
				return result, err
			}
		}
	}

	return result, nil
}

func applyValidationRule(rule string, result *ValidationResult, field reflect.Value, fieldName string) error {
	switch {
	case strings.HasPrefix(rule, "max="):
		max, err := strconv.Atoi(strings.Split(rule, "=")[1])
		if err != nil {
			return err
		}

		if field.Type().Kind() != reflect.String {
			return fmt.Errorf("fields using rules \"max\" or \"min\" must be of type string")
		}

		if !maxChars(field.String(), max) {
			result.IsValid = false
			result.Errors = append(result.Errors, newValidationError(fieldName, field.String(), fmt.Sprintf("%s must be less than %d characters long", fieldName, max)))
		}
	case strings.HasPrefix(rule, "min="):
		min, err := strconv.Atoi(strings.Split(rule, "=")[1])
		if err != nil {
			return err
		}

		if field.Type().Kind() != reflect.String {
			return fmt.Errorf("fields using rules \"max\" or \"min\" must be of type string")
		}

		if !minChars(field.String(), min) {
			result.IsValid = false
			result.Errors = append(result.Errors, newValidationError(fieldName, field.String(), fmt.Sprintf("%s must be at least %d characters long", fieldName, min)))
		}
	case strings.HasPrefix(rule, "email"):
		if field.Type().Kind() != reflect.String {
			return fmt.Errorf("fields using rules \"email\" must be of type string")
		}

		if !isEmail(field.String()) {
			result.IsValid = false
			result.Errors = append(result.Errors, newValidationError(fieldName, field.String(), fmt.Sprintf("%s must be a valid email address", field.String())))
		}
	case strings.HasPrefix(rule, "req"):
		if field.Type().Kind() != reflect.String {
			return fmt.Errorf("fields using rules \"req\" must be of type string")
		}

		if field.String() == "" {
			result.IsValid = false
			result.Errors = append(result.Errors, newValidationError(fieldName, field.String(), fmt.Sprintf("%s is required", fieldName)))
		}
	}

	return nil
}

func newValidationError(fieldName, field, errorString string) ValidationError {
	return ValidationError{
		FieldName: fieldName,
		Field:     field,
		Error:     errorString,
	}
}

func maxChars(text string, maxCount int) bool {
	return len(text) <= maxCount
}

func minChars(text string, minCount int) bool {
	return len(text) >= minCount
}

func isEmail(email string) bool {
	return internal.EmailRX.MatchString(email)
}
