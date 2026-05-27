package matcher

import (
	"strings"
	"unicode"
)

// ExperienceMatch calculates the percentage of unique words in user experiences that appear in the job description.
func ExperienceMatch(expTexts []string, jobDesc string) float64 {
	// 1. Build a set of unique words from all user experiences (excluding stop words)
	expWords := make(map[string]struct{})
	for _, text := range expTexts {
		words := tokenizeAndFilter(text)

		for _, w := range words {
			expWords[w] = struct{}{}
		}
	}

	if len(expWords) == 0 {
		return 0.0
	}

	// 2. Build a set of unique words from the job description
	jobWordsList := tokenizeAndFilter(jobDesc)
	jobWords := make(map[string]struct{})
	for _, w := range jobWordsList {
		jobWords[w] = struct{}{}
	}

	// 3. Count how many of the experience words occur in the job description
	matched := 0
	for w := range expWords {
		if _, ok := jobWords[w]; ok {
			matched++
		}
	}

	// Return overlap ratio (0.0 to 1.0)
	return float64(matched) / float64(len(expWords))
}

// tokenizeAndFilter splits text into alphanumeric lowercase words, filtering out stop words.
func tokenizeAndFilter(text string) []string {
	text = strings.ToLower(text)
	var tokens []string
	var currentToken strings.Builder

	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			currentToken.WriteRune(r)
		} else {
			if currentToken.Len() > 0 {
				word := currentToken.String()
				if _, isStop := stopWords[word]; !isStop {
					tokens = append(tokens, word)
				}
				currentToken.Reset()
			}
		}
	}
	if currentToken.Len() > 0 {
		word := currentToken.String()
		if _, isStop := stopWords[word]; !isStop {
			tokens = append(tokens, word)
		}
	}
	return tokens
}
