package matcher

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
	userExps []string,
	titleThreshold float64,
) MatchResult {
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

	// 2. Skill Matching
	var matchedSkills []string
	skillScore := 0.0
	if len(userSkills) > 0 {
		sm := NewSkillMatcher(userSkills)
		matchedSkills = sm.FindSkills(jobDesc)
		skillScore = float64(len(matchedSkills)) / float64(len(userSkills))
	}

	// 3. Experience Matching
	expScore := ExperienceMatch(userExps, jobDesc)

	// 4. Calculate weighted score (0 to 100)
	// Weights: Title (30%), Skills (40%), Experience (30%)
	weightedScore := (bestTitleScore * 0.3) + (skillScore * 0.4) + (expScore * 0.3)
	scoreInt := max(
		// Clamp between 0 and 100
		min(
			int(weightedScore*100), 100,
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
