// Package stringutils
package stringutils

import (
	"fmt"
	"html"
	"regexp"
	"strings"
	"time"
)

var (
	htmlTagsRe     = regexp.MustCompile(`<[^>]*>`)
	controlCharsRe = regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]`)
	extraSpaceRe   = regexp.MustCompile(`\s+`)
	multiNewlineRe = regexp.MustCompile(`\n{3,}`)
)

func SanitizeStrict(input string) string {
	str := html.UnescapeString(input)
	str = htmlTagsRe.ReplaceAllString(str, " ")
	str = controlCharsRe.ReplaceAllString(str, "")
	str = extraSpaceRe.ReplaceAllString(str, " ")
	return strings.TrimSpace(str)
}

func SanitizeDescription(input string) string {
	str := html.UnescapeString(input)
	str = controlCharsRe.ReplaceAllString(str, "")
	str = multiNewlineRe.ReplaceAllString(str, "\n\n")
	return strings.TrimSpace(str)
}

func GenerateUsername(email string) string {
	if email != "" {
		parts := strings.Split(email, "@")
		return parts[0]
	}

	return fmt.Sprintf("user_%d", time.Now().Unix())
}

func IsValidUsername(username string) bool {
	// Check for alphanumeric and underscores only
	re := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	return len(username) >= 3 && len(username) <= 20 && re.MatchString(username)
}

func SanitizeStringSlice(slice []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, s := range slice {
		trimmed := strings.TrimSpace(s)
		if trimmed == "" {
			continue
		}
		valLower := strings.ToLower(trimmed)
		if seen[valLower] {
			continue
		}
		seen[valLower] = true
		result = append(result, trimmed)
	}
	if result == nil {
		return []string{}
	}
	return result
}
