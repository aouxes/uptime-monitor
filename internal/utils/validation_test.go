package utils

import (
	"testing"
)

func TestValidateUser(t *testing.T) {
	tests := []struct {
		name     string
		username string
		email    string
		password string
		hasError bool
	}{
		{
			name:     "Valid user",
			username: "johndoe",
			email:    "john@example.com",
			password: "Secure123",
			hasError: false,
		},
		{
			name:     "Short username",
			username: "jo",
			email:    "john@example.com",
			password: "Secure123",
			hasError: true,
		},
		{
			name:     "Invalid email",
			username: "johndoe",
			email:    "invalid-email",
			password: "Secure123",
			hasError: true,
		},
		{
			name:     "Weak password",
			username: "johndoe",
			email:    "john@example.com",
			password: "123",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateUser(tt.username, tt.email, tt.password)

			if tt.hasError && len(errors) == 0 {
				t.Error("Expected validation errors, but got none")
			}

			if !tt.hasError && len(errors) > 0 {
				t.Errorf("Expected no errors, but got: %v", errors)
			}
		})
	}
}
