package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	mrand "math/rand"
	"regexp"
	"strings"
)

// GenerateOTP creates a cryptographically secure 6-digit numeric code
func GenerateOTP() (string, error) {
	const length = 6
	const digits = "0123456789"
	otp := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		otp[i] = digits[num.Int64()]
	}
	return string(otp), nil
}

// GenerateUsername creates a unique username based on the user's full name
// It converts the name to lowercase, removes spaces, and appends a random 4-digit number
func GenerateUsername(fullName string) string {

	// Clean the name: lowercase and remove non-alphanumeric characters
	reg, _ := regexp.Compile("[^a-zA-Z0-9]+")
	cleanName := strings.ToLower(fullName)
	cleanName = reg.ReplaceAllString(cleanName, "")

	// If name is too short, provide a default
	if len(cleanName) < 3 {
		cleanName = "user"
	}

	// Truncate if too long (optional, e.g., max 15 chars before suffix)
	if len(cleanName) > 15 {
		cleanName = cleanName[:15]
	}

	// Generate a 4-digit random number
	randomSuffix := mrand.Intn(9000) + 1000

	return fmt.Sprintf("%s%d", cleanName, randomSuffix)
}

// GenerateAccountNumber creates a 10-digit account number from a phone number
// It takes the last 10 digits of the phone number
func GenerateAccountNumber(phoneNumber string) string {
	// Remove all non-numeric characters
	reg, _ := regexp.Compile("[^0-9]+")
	cleanPhone := reg.ReplaceAllString(phoneNumber, "")

	// Take the last 10 digits if available
	if len(cleanPhone) >= 10 {
		return cleanPhone[len(cleanPhone)-10:]
	}

	// If phone number is too short, pad with zeros or return as is (should not happen with valid phone)
	return fmt.Sprintf("%010s", cleanPhone)
}
