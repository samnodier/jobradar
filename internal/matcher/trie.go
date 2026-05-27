package matcher

import (
	"sort"
	"strings"
	"unicode"
)

// Pattern holds the matching skills name and it's rune length for boundary calculation
type Pattern struct {
	Name       string
	RuneLength int
}

// TrieNode represents a single character node in the state machine
type TrieNode struct {
	Children map[rune]*TrieNode
	Fail     *TrieNode
	Output   []Pattern // All skills that match at this state
}

type SkillMatcher struct {
	Root *TrieNode
}

// NewSkillMatcher builds the Aho-Corasick state machine for a list of skills
func NewSkillMatcher(skills []string) *SkillMatcher {
	root := &TrieNode{
		Children: make(map[rune]*TrieNode),
	}

	seen := make(map[string]struct{})

	// 1. Build the Trie
	for _, skill := range skills {
		normKey := strings.ToLower(skill)
		if _, isSeen := seen[normKey]; isSeen || normKey == "" {
			continue
		}
		seen[normKey] = struct{}{}

		runes := []rune(normKey)
		curr := root
		for _, r := range runes {
			if _, ok := curr.Children[r]; !ok {
				curr.Children[r] = &TrieNode{
					Children: make(map[rune]*TrieNode),
				}
			}
			curr = curr.Children[r]
		}
		// Store the original (display) name but deduplicated
		curr.Output = append(curr.Output, Pattern{
			Name:       skill,
			RuneLength: len(runes),
		})
	}

	// 2. Build Failure Links using BFS
	var queue []*TrieNode
	// Root's immediate children fail back to root
	for _, child := range root.Children {
		child.Fail = root
		queue = append(queue, child)
	}
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		for r, child := range curr.Children {
			failNode := curr.Fail
			// Follow fail links until we find a node with child character 'r'
			for failNode != nil {
				if nextNode, ok := failNode.Children[r]; ok {
					child.Fail = nextNode
					break
				}
				failNode = failNode.Fail
			}
			if failNode == nil {
				child.Fail = root
			}
			// Propagate outputs from the fail node to the current child node
			existing := child.Output
			child.Output = make([]Pattern, len(existing), len(existing)+len(child.Fail.Output))
			copy(child.Output, existing)
			child.Output = append(child.Output, child.Fail.Output...)
			queue = append(queue, child)
		}
	}
	return &SkillMatcher{Root: root}
}

// isWordBoundary checks if a character is a divider (non-alphanumeric).
func isWordBoundary(r rune) bool {
	return !unicode.IsLetter(r) && !unicode.IsDigit(r)
}

// FindSkills scans the text and returns all matched skills.
func (sm *SkillMatcher) FindSkills(text string) []string {
	textLower := strings.ToLower(text)
	runes := []rune(textLower)
	n := len(runes)
	matchedMap := make(map[string]struct{})

	curr := sm.Root
	for i := range n {
		r := runes[i]

		// If no matching child, follow fail links
		for curr != nil {
			if next, ok := curr.Children[r]; ok {
				curr = next
				break
			}
			curr = curr.Fail
		}
		if curr == nil {
			curr = sm.Root
			continue

		}

		// Check if we matched any skill patterns at this node
		for _, pattern := range curr.Output {
			start := i - pattern.RuneLength + 1
			end := i

			// FIX (Gap 3 — punctuation override trap): boundary is purely about
			// what sits immediately outside the match — never about what the
			// skill itself starts or ends with.
			//
			// isWordBoundary already handles symbols correctly:
			//   "C++ Developer" → runes[end+1] is ' ', isWordBoundary(' ') = true
			//   "ASP.NET"       → runes[start-1] is 'P', isWordBoundary('P') = false
			//                     so ".NET" inside "ASP.NET" is correctly rejected.
			leftOk := start == 0 || isWordBoundary(runes[start-1])
			rightOk := end == n-1 || isWordBoundary(runes[end+1])

			if leftOk && rightOk {
				matchedMap[pattern.Name] = struct{}{}
			}
		}
	}

	// Extract and sort results
	var results []string
	for skill := range matchedMap {
		results = append(results, skill)
	}
	sort.Strings(results)
	return results
}
