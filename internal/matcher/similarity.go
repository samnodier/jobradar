// Package matcher
package matcher

import (
	"sort"
	"strings"
	"unicode"
)

// Jaro calculates the Jaro similarity between two strings
// Uses runes for Unicode safety
func Jaro(s1, s2 string) float64 {
	// Calculate the window
	r1 := []rune(s1)
	r2 := []rune(s2)
	len1 := len(s1)
	len2 := len(s2)
	if len1 == 0 && len2 == 0 {
		return 1.0
	}
	if len1 == 0 || len2 == 0 {
		return 0.0
	}
	window := max(len1, len2)/2 - 1
	window = max(0, window)

	r1Matches := make([]bool, len1)
	r2Matches := make([]bool, len2)
	matches := 0

	// Find the matching characters within the match window
	for i := range len1 {
		start := max(0, i-window)
		end := min(len2-1, i+window)
		for j := start; j <= end; j++ {
			if !r2Matches[j] && r1[i] == r2[j] {
				r1Matches[i] = true
				r2Matches[j] = true
				matches++
				break
			}
		}
	}
	if matches == 0 {
		return 0.0
	}

	// Count transpositions (mismatched ordering of matching characters)
	mismatches := 0
	k := 0
	for i := range len1 {
		if !r1Matches[i] {
			continue
		}
		// Find the next matched character in s2
		for k < len2 && !r2Matches[k] {
			k++
		}
		if k < len2 && s1[i] != s2[k] {
			mismatches++
		}
		k++
	}
	t := float64(mismatches) / 2.0
	m := float64(matches)

	// Jaro formula 1/3 * (m/|s1| + m/|s2| + (m-t)/m)
	return (m/float64(len1) + m/float64(len2) + (m-t)/m) / 3.0
}

// JaroWinkler calculates the Jaro-Winkler similarity between two strings
// Giving higher score for matching prefixes
func JaroWinkler(s1, s2 string) float64 {
	j := Jaro(s1, s2)
	if j < 0.7 { // Boost only applies if the base Jaro score is high enough (threshold of 0.7)
		return j
	}
	r1 := []rune(s1)
	r2 := []rune(s2)
	// Find length of common prefix (up to 4 characters)
	prefixLen := 0
	maxPrefix := min(4, min(len(r1), len(r2)))
	for prefixLen < maxPrefix && r1[prefixLen] == r2[prefixLen] {
		prefixLen++
	}
	p := 0.1 // Constant scaling factor
	return j + float64(prefixLen)*p*(1.0-j)
}

// TokenSortJaroWinkler normalizes, tokenizes, sorts, and compares two strings
func TokenSortJaroWinkler(s1, s2 string) float64 {
	clean1 := cleanAndSortString(s1)
	clean2 := cleanAndSortString(s2)
	return JaroWinkler(clean1, clean2)
}

// Domain-specific stop words to ignore in job titles
var stopWords = map[string]struct{}{
	"in":        {},
	"of":        {},
	"for":       {},
	"at":        {},
	"and":       {},
	"with":      {},
	"a":         {},
	"an":        {},
	"the":       {},
	"to":        {},
	"by":        {},
	"role":      {},
	"position":  {},
	"level":     {},
	"ii":        {},
	"iii":       {},
	"i":         {},
	"status":    {},
	"associate": {},
	"staff":     {},
}

// Abbreviation/synonym expansion map for job titles
var normalizations = map[string][]string{
	"sr":  {"senior"},
	"jr":  {"junior"},
	"vp":  {"vice", "president"},
	"eng": {"engineer"},
	"mgr": {"manager"},
	"dev": {"developer"},
	"swe": {"software", "engineer"},
	"cto": {"chief", "technology", "officer"},
	"ceo": {"chief", "executive", "officer"},
}

// cleanAndSortString lowercases, extracts alphanumeric tokens, filters stop words, sorts them, and joins them with spaces.
func cleanAndSortString(s string) string {
	s = strings.ToLower(s)

	var tokens []string
	var currentToken strings.Builder

	addToken := func(word string) {
		if _, isStop := stopWords[word]; isStop {
			return
		}
		// Expand abbreviations if they exist in the map
		if expanded, ok := normalizations[word]; ok {
			for _, w := range expanded {
				if _, isStop := stopWords[w]; !isStop {
					tokens = append(tokens, w)
				}
			}
		} else {
			tokens = append(tokens, word)
		}
	}

	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			currentToken.WriteRune(r)
		} else {
			if currentToken.Len() > 0 {
				addToken(currentToken.String())
				currentToken.Reset()
			}
		}
	}
	if currentToken.Len() > 0 {
		addToken(currentToken.String())
	}

	sort.Strings(tokens)
	return strings.Join(tokens, " ")
}
