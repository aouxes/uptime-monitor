package utils

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "Secure123"

	hashed, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if hashed == password {
		t.Error("Hashed password should not equal original")
	}

	if len(hashed) == 0 {
		t.Error("Hashed password should not be empty")
	}
}

func TestCheckPasswordHash(t *testing.T) {
	password := "Secure123"
	wrongPassword := "Wrong123"

	hashed, _ := HashPassword(password)

	if !CheckPasswordHash(password, hashed) {
		t.Error("CheckPasswordHash should return true for correct password")
	}

	if CheckPasswordHash(wrongPassword, hashed) {
		t.Error("CheckPasswordHash should return false for wrong password")
	}
}
