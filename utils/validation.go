package utils

import (
	"regexp"
)

func ValidatePhone(phone string) bool {
	// Phone should begin with + and international calling number
	phoneRegex := regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
	return phoneRegex.MatchString(phone)
}

func ValidateEmail(email string) bool {
	// Simple email regex pattern
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}
