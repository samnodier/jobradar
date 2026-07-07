package matcher

import (
	"maps"
	"math"
	"slices"
)

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
	userSkills map[string]float64,
	jobSkills []string,
	userExperiences []string,
	titleThreshold float64,
) MatchResult {
	weightedSum := 0.0
	activeWeight := 0.0

	// 1. Title Matching (Early Exit)
	// The title dimension only participates when the user has desired roles.
	// No roles = a missing signal, not a zero — title drops out and the other
	// dimensions renormalize, exactly like skills/experience below. The early
	// exit therefore also only applies when roles exist; callers can disable
	// it entirely with titleThreshold 0 (used for user-imported jobs, where
	// explicit intent overrides the firehose filter).
	bestTitleScore := 0.0
	if len(desiredRoles) > 0 {
		for _, role := range desiredRoles {
			score := TokenSortJaroWinkler(jobTitle, role)
			if score > bestTitleScore {
				bestTitleScore = score
			}
		}

		if bestTitleScore < titleThreshold {
			return MatchResult{
				Score:   0,
				Skipped: true,
			}
		}
		weightedSum += 0.45 * bestTitleScore
		activeWeight += 0.45
	}

	// 2. Skill Matching
	// NOTE: matchedSkills is computed for DISPLAY whenever the user has skills,
	// even if jobSkills is empty (so the skill dimension contributes nothing to
	// the score). This means MatchedSkills can be non-empty while SkillScore is
	// 0 — the two are separate concerns. Do not infer one from the other.
	var matchedSkills []string
	skillScore := 0.0
	if len(userSkills) > 0 {
		sm := NewSkillMatcher(slices.Collect(maps.Keys(userSkills)))
		matchedSkills = sm.FindSkills(jobDesc)
		var totalMatchedSkillsWeight float64
		for _, skill := range matchedSkills {
			totalMatchedSkillsWeight += userSkills[skill]
		}
		if len(jobSkills) > 0 {
			skillScore = totalMatchedSkillsWeight / float64(len(jobSkills))
		}
		skillScore = math.Min(skillScore, 1.0)
	}

	if len(jobSkills) > 0 && len(userSkills) > 0 {
		weightedSum += 0.3 * skillScore
		activeWeight += 0.3
	}

	// 3. Experience Matching
	expScore := 0.0
	if len(userExperiences) != 0 {
		expScore = ExperienceMatch(userExperiences, jobDesc)
		weightedSum += 0.25 * expScore
		activeWeight += 0.25
	}

	// 4. Calculate final weighted score (0 to 100)
	// Weights: Title (45%), Skills (30%), Experience (25%), renormalized over
	// the dimensions actually present.
	if activeWeight == 0 {
		// No roles, no skills, no experiences — there is nothing to score.
		// Skip rather than fabricate a number from zero signal.
		return MatchResult{
			Score:   0,
			Skipped: true,
		}
	}
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
