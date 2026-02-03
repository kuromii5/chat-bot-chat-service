package validator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func Init() {
	validate = validator.New()

	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

type ValidationError struct {
	Fields map[string]string `json:"fields"`
}

func (e ValidationError) Error() string {
	return "validation failed"
}

func Validate(s any) error {
	if err := validate.Struct(s); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			fields := make(map[string]string)
			for _, fieldErr := range validationErrors {
				fields[fieldErr.Field()] = formatMessage(fieldErr)
			}
			return ValidationError{Fields: fields}
		}
		return err
	}
	return nil
}

func formatMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "oneof":
		return fmt.Sprintf("Value must be one of the following: %s", err.Param())
	case "min":
		return fmt.Sprintf("Minimum length or value: %s", err.Param())
	case "max":
		return fmt.Sprintf("Maximum length or value: %s", err.Param())
	case "uuid":
		return "Must be a valid UUID"
	case "tags":
		return "Must be a valid tags"
	default:
		return fmt.Sprintf("Invalid value (error code: %s)", err.Tag())
	}
}
