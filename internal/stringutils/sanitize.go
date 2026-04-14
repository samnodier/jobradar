package stringutils

import (
	"html"
	"regexp"
	"strings"
)

var (
	htmlTags     = regexp.MustCompile(`<[^>]*>`)
	whitespace   = regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]`)
	controlChars = regexp.MustCompile(`\s+`)
)

func Sanitize(input string) string {
	// Unescape HTML entities
	str := html.UnescapeString(input)
	str = htmlTags.ReplaceAllString(str, " ")
	str = controlChars.ReplaceAllString(str, " ")
	str = whitespace.ReplaceAllString(str, " ")
	return strings.TrimSpace(str)
}
