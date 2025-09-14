package utils

import (
	"net/mail"
	"unicode"
)

func ValidateUser(username, email, password string) map[string]string {
	errors := make(map[string]string)

	if len(username) < 3 {
		errors["username"] = "Username must be at least 3 characters"
	}

	if _, err := mail.ParseAddress(email); err != nil {
		errors["email"] = "Invalid email address"
	}

	if len(password) < 6 {
		errors["password"] = "Password must be at least 6 characters"
	}

	var hasUpper, hasLower, hasNumber bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		}
	}

	if !hasUpper || !hasLower || !hasNumber {
		errors["password"] = "Password must contain uppercase, lowercase letters and numbers"
	}

	return errors
}
