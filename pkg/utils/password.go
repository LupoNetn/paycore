package utils

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes a plain text password using bcrypt with a default cost.
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// CheckPassword compares a bcrypt hashed password with its possible plaintext equivalent.
// Returns nil on success, or an error on failure.
func CheckPassword(password, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
