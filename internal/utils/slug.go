package utils

import (
	"regexp"
	"strings"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
var invalidSlugChars = regexp.MustCompile(`[^a-z0-9]+`)

func NormalizeSlug(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	normalized = invalidSlugChars.ReplaceAllString(normalized, "-")
	normalized = strings.Trim(normalized, "-")
	return normalized
}

func IsValidSlug(value string) bool {
	return slugPattern.MatchString(value)
}
