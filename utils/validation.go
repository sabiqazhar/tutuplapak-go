package utils

import (
	"regexp"
)

func ValidatePhone(phone string) bool {
	// Phone should begin with + and international calling number
	phoneRegex := regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
	return phoneRegex.MatchString(phone)
}
