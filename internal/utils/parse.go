package utils

import (
	"strconv"
)

func ParseInt(value string, fallback int) int {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		parsed = fallback
	}

	return parsed
}

func ParseFloat(value string, fallback float64) float64 {
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil || value == "" {
		parsed = fallback
	}

	return parsed
}

func ParseBool(value string, fallback bool) bool {
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		parsed = fallback
	}
	return parsed
}
