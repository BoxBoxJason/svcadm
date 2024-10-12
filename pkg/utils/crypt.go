package utils

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
)

// GenerateRandomPassword generates a random password of the specified length
func GenerateRandomPassword(length int) (string, error) {
	if length < 1 {
		return "", errors.New("password length must be at least 1")
	}
	password := make([]byte, length)
	if _, err := rand.Read(password); err != nil {
		return "", err
	}
	password_str := base64.URLEncoding.EncodeToString(password)

	return password_str[:length], nil
}
