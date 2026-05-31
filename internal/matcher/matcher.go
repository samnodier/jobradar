package matcher

import "math"

// MatchResult holds the combined matching scores and outcomes.
type MatchResult struct {
	Score         int      // Final score (0 to 100)
	TitleScore    float64  // Jaro-Winkler score of the title
	SkillScore    float64  // Skill coverage ratio (0. 0 to 1.0)
	ExpScore      float64  // Experience overlap ratio (0.0 to 1.0)
	MatchedSkills []string // List of skills matched in the job description
	Skipped       bool     // True if skipped due to low title match
}

// MatchJob runs the matching logic and calculates the final matching score.
func MatchJob(
	jobTitle string,
	jobDesc string,
	desiredRoles []string,
	userSkills []string,
	jobSkills []string,
	userExps []string,
	titleThreshold float64,
) MatchResult {
	weightedSum := 0.0
	activeWeight := 0.0

	// 1. Title Matching (Early Exit)
	bestTitleScore := 0.0
	for _, role := range desiredRoles {
		score := TokenSortJaroWinkler(jobTitle, role)
		if score > bestTitleScore {
			bestTitleScore = score
		}
	}

	// Early exit check
	if bestTitleScore < titleThreshold {
		return MatchResult{
			Score:   0,
			Skipped: true,
		}
	}
	weightedSum += 0.45 * bestTitleScore
	activeWeight += 0.45

	// 2. Skill Matching
	var matchedSkills []string
	skillScore := 0.0
	if len(userSkills) > 0 {
		sm := NewSkillMatcher(userSkills)
		matchedSkills = sm.FindSkills(jobDesc)
		if len(jobSkills) > 0 {
			skillScore = float64(len(matchedSkills)) / float64(len(jobSkills))
		}
		skillScore = math.Min(skillScore, 1.0)
	}

	if len(jobSkills) > 0 && len(userSkills) > 0 {
		weightedSum += 0.3 * skillScore
		activeWeight += 0.3
	}

	// 3. Experience Matching
	expScore := 0.0
	if len(userExps) != 0 {
		expScore = ExperienceMatch(userExps, jobDesc)
		weightedSum += 0.25 * expScore
		activeWeight += 0.25
	}

	// 4. Calculate final weighted score (0 to 100)
	// Weights: Title (45%), Skills (30%), Experience (25%)
	// activeWeight >= 0.45 always; title is unconditional past the gate
	finalWeightedScore := weightedSum / activeWeight
	scoreInt := max(
		// Clamp between 0 and 100
		min(
			int(finalWeightedScore*100), 100,
		), 0,
	)

	return MatchResult{
		Score:         scoreInt,
		TitleScore:    bestTitleScore,
		SkillScore:    skillScore,
		ExpScore:      expScore,
		MatchedSkills: matchedSkills,
		Skipped:       false,
	}
}
