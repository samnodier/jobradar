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
	return strings.TrimSpace(input)
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
