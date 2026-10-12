package matcher

import "strings"

type FeedbackExample struct {
	Title     string
	Skills    []string
	Sentiment int
}

// Explicit votes provide a small content-based signal. Shrink toward zero with
// two neutral pseudo-votes so a single click cannot dominate professional fit.
// No cross-user tracking, protected traits, or hiring-probability claims.
func FeedbackAdjustment(title string, skills []string, examples []FeedbackExample) int {
	evidence, weight := 0.0, 2.0
	for _, example := range examples {
		if example.Sentiment != -1 && example.Sentiment != 1 {
			continue
		}
		titleSimilarity := TokenSortJaroWinkler(title, example.Title)
		if titleSimilarity < 0.75 {
			continue
		}
		similarity := titleSimilarity
		if len(skills) > 0 && len(example.Skills) > 0 {
			known := map[string]bool{}
			for _, skill := range example.Skills {
				known[strings.ToLower(skill)] = true
			}
			overlap := 0
			for _, skill := range skills {
				if known[strings.ToLower(skill)] {
					overlap++
				}
			}
			similarity = (similarity + float64(overlap)/float64(len(skills))) / 2
		}
		evidence += similarity * float64(example.Sentiment)
		weight += similarity
	}
	return max(-10, min(10, int(10*evidence/weight)))
}
