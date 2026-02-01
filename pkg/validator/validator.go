package validator

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func Init() {
	validate = validator.New()
}

type ValidationError struct {
	Fields map[string]string
}

func (e ValidationError) Error() string {
	return "validation failed"
}

func Validate(s any) error {
	if err := validate.Struct(s); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			fields := make(map[string]string)
			for _, fieldErr := range validationErrors {
				fields[fieldErr.Field()] = fmt.Sprintf("failed validation: %s", fieldErr.Tag())
			}
			return ValidationError{Fields: fields}
		}
		return err
	}
	return nil
}
