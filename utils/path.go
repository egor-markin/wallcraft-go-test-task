package utils

import (
	"strconv"
	"strings"
)

// ExtractTrailingID extracts an integer ID from the last segment of a URL path. E.g., given "/products/123", it returns 123
func ExtractTrailingID(path string) (int, error) {
	parts := strings.Split(path, "/")
	lastPart := parts[len(parts)-1]

	number, err := strconv.Atoi(lastPart)
	if err != nil {
		return 0, err
	}

	return number, nil
}
