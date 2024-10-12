package utils

import (
	"os/exec"
	"strings"
	"time"
)

const ALPHANUMERICS = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GetHostname() string {
	cmd := exec.Command("hostname")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// GenerateDatetimeString creates a string with the current datetime, format YYYY-MM-DD-HH-MM-SS
func GenerateDatetimeString() string {
	now := time.Now()
	return now.Format("2006-01-02-15-04-05")
}
