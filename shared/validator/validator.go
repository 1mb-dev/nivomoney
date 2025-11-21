// Package validator provides input validation utilities for Nivo services.
package validator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/vnykmshr/nivo/shared/errors"
	"github.com/vnykmshr/nivo/shared/models"
)

// Validator wraps go-playground/validator with custom validation rules.
type Validator struct {
	validate *validator.Validate
}

// New creates a new Validator with custom validation rules.
func New() *Validator {
	validate := validator.New()

	// Use JSON tag names for validation errors
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	// Register custom validators
	v := &Validator{validate: validate}
	v.registerCustomValidators()

	return v
}

// Validate validates a struct and returns validation errors.
func (v *Validator) Validate(s interface{}) error {
	err := v.validate.Struct(s)
	if err == nil {
		return nil
	}

	// Convert validator errors to our error format
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return errors.Validation("validation failed")
	}

	return v.formatValidationErrors(validationErrors)
}

// ValidateVar validates a single variable.
func (v *Validator) ValidateVar(field interface{}, tag string) error {
	err := v.validate.Var(field, tag)
	if err == nil {
		return nil
	}

	return errors.Validation(fmt.Sprintf("validation failed: %v", err))
}

// formatValidationErrors converts validator.ValidationErrors to our error format.
func (v *Validator) formatValidationErrors(validationErrors validator.ValidationErrors) *errors.Error {
	details := make(map[string]interface{})

	for _, err := range validationErrors {
		fieldName := err.Field()
		if fieldName == "" {
			fieldName = err.StructField()
		}

		details[fieldName] = v.getErrorMessage(err)
	}

	return errors.Validation("validation failed").WithDetails(details)
}

// getErrorMessage returns a human-readable error message for a validation error.
func (v *Validator) getErrorMessage(err validator.FieldError) string {
	field := err.Field()
	tag := err.Tag()
	param := err.Param()

	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return "must be a valid email address"
	case "min":
		if err.Type().Kind() == reflect.String {
			return fmt.Sprintf("must be at least %s characters", param)
		}
		return fmt.Sprintf("must be at least %s", param)
	case "max":
		if err.Type().Kind() == reflect.String {
			return fmt.Sprintf("must be at most %s characters", param)
		}
		return fmt.Sprintf("must be at most %s", param)
	case "len":
		return fmt.Sprintf("must be exactly %s characters", param)
	case "gt":
		return fmt.Sprintf("must be greater than %s", param)
	case "gte":
		return fmt.Sprintf("must be greater than or equal to %s", param)
	case "lt":
		return fmt.Sprintf("must be less than %s", param)
	case "lte":
		return fmt.Sprintf("must be less than or equal to %s", param)
	case "alpha":
		return "must contain only letters"
	case "alphanum":
		return "must contain only letters and numbers"
	case "numeric":
		return "must be a valid number"
	case "uuid":
		return "must be a valid UUID"
	case "url":
		return "must be a valid URL"
	case "oneof":
		return fmt.Sprintf("must be one of: %s", param)
	case "currency":
		return "must be a valid ISO 4217 currency code"
	case "money_amount":
		return "must be a valid monetary amount (positive integer in cents)"
	default:
		return fmt.Sprintf("validation failed on '%s' tag", tag)
	}
}

// registerCustomValidators registers custom validation rules.
func (v *Validator) registerCustomValidators() {
	// Currency validator - checks if the currency is supported
	v.validate.RegisterValidation("currency", func(fl validator.FieldLevel) bool {
		currency := fl.Field().String()
		return models.Currency(currency).IsSupported()
	})

	// Money amount validator - checks if amount is positive
	v.validate.RegisterValidation("money_amount", func(fl validator.FieldLevel) bool {
		amount := fl.Field().Int()
		return amount > 0
	})

	// Account number validator - basic format check
	v.validate.RegisterValidation("account_number", func(fl validator.FieldLevel) bool {
		accountNumber := fl.Field().String()
		// Account numbers should be 10-20 alphanumeric characters
		if len(accountNumber) < 10 || len(accountNumber) > 20 {
			return false
		}
		for _, r := range accountNumber {
			if !((r >= '0' && r <= '9') || (r >= 'A' && r <= 'Z')) {
				return false
			}
		}
		return true
	})

	// IBAN validator - basic format check
	v.validate.RegisterValidation("iban", func(fl validator.FieldLevel) bool {
		iban := fl.Field().String()
		// Basic IBAN format: 2 letters, 2 digits, then up to 30 alphanumeric
		if len(iban) < 15 || len(iban) > 34 {
			return false
		}
		// First 2 should be letters
		if !isLetter(rune(iban[0])) || !isLetter(rune(iban[1])) {
			return false
		}
		// Next 2 should be digits
		if !isDigit(rune(iban[2])) || !isDigit(rune(iban[3])) {
			return false
		}
		return true
	})

	// Sort code validator (UK bank sort codes: 6 digits)
	v.validate.RegisterValidation("sort_code", func(fl validator.FieldLevel) bool {
		sortCode := fl.Field().String()
		// Remove hyphens
		sortCode = strings.ReplaceAll(sortCode, "-", "")
		if len(sortCode) != 6 {
			return false
		}
		for _, r := range sortCode {
			if !isDigit(r) {
				return false
			}
		}
		return true
	})

	// Routing number validator (US bank routing numbers: 9 digits)
	v.validate.RegisterValidation("routing_number", func(fl validator.FieldLevel) bool {
		routingNumber := fl.Field().String()
		if len(routingNumber) != 9 {
			return false
		}
		for _, r := range routingNumber {
			if !isDigit(r) {
				return false
			}
		}
		return true
	})
}

// Helper functions

func isLetter(r rune) bool {
	return (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z')
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

// Global validator instance
var defaultValidator *Validator

// Init initializes the global validator.
func Init() {
	defaultValidator = New()
}

// Default returns the default global validator.
func Default() *Validator {
	if defaultValidator == nil {
		Init()
	}
	return defaultValidator
}

// Validate validates using the global validator.
func Validate(s interface{}) error {
	return Default().Validate(s)
}

// ValidateVar validates a variable using the global validator.
func ValidateVar(field interface{}, tag string) error {
	return Default().ValidateVar(field, tag)
}
