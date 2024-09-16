package utils

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
)

// GenerateRandomPassword generates a random password of the specified length
func GenerateRandomPassword(length int) (string, error) {
	if length < 4 {
		return "", errors.New("password length must be at least 4")
	}
	password := make([]byte, length/4)
	if _, err := rand.Read(password); err != nil {
		return "", err
	}
	return string(base64.StdEncoding.EncodeToString(password)), nil
}
