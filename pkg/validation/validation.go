package validation

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

func ParseValidationErrors(err error) map[string]string {
	fields := map[string]string{}

	for _, e := range err.(validator.ValidationErrors) {
		field := strings.ToLower(e.Field())

		switch e.Tag() {
		case "required":
			fields[field] = field + " field is required"
		case "email":
			fields[field] = "invalid email format"
		case "min":
			fields[field] = field + " field must be atleast " + e.Param() + " characters"
		case "max":
			fields[field] = field + " field must be at most " + e.Param() + " characters"
		case "alpha":
			fields[field] = field + " field must contain letters only"
		case "alphanum":
			fields[field] = field + " field must contain letters and numbers only"
		default:
			fields[field] = field + " field is invalid"
		}

	}

	return fields
}
