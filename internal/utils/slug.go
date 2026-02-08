package utils

import (
	"regexp"
	"strings"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func NormalizeSlug(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func IsValidSlug(value string) bool {
	return slugPattern.MatchString(value)
}
