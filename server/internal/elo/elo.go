// Package elo implements the rating math behind the PRD's "Linguistic Elo":
// the learner is a player, each reviewed item has a difficulty derived from its
// HSK level, and every graded action nudges the relevant proficiency vector.
//
// It is pure and stateless — persistence and which-vector-to-nudge decisions
// live in the handlers/DB layer.
package elo

import "math"

const (
	// DefaultRating is where every vector starts (matches the DB default):
	// DifficultyForLevel(1), i.e. a brand-new learner sits at HSK 1.
	DefaultRating = 750.0

	// KFactor caps how much a single action can move a rating.
	KFactor = 24.0

	// HSK level L maps to difficulty = levelBase + L*levelStep, so HSK1≈750
	// and HSK7≈1650 on the same scale as ratings.
	levelBase = 600.0
	levelStep = 150.0
)

// DifficultyForLevel maps an HSK level (1..N) to a rating-scale difficulty.
func DifficultyForLevel(level int) float64 {
	return levelBase + float64(level)*levelStep
}

// LevelForRating maps a rating back to the nearest HSK level, clamped to
// [1, maxLevel]. Used to pick content difficulty from Vocabulary Elo.
func LevelForRating(rating float64, maxLevel int) int {
	level := int(math.Round((rating - levelBase) / levelStep))
	if level < 1 {
		return 1
	}
	if level > maxLevel {
		return maxLevel
	}
	return level
}

// Expected is the logistic probability the learner succeeds against an item of
// the given difficulty (standard Elo, 400-point scale).
func Expected(rating, difficulty float64) float64 {
	return 1.0 / (1.0 + math.Pow(10, (difficulty-rating)/400.0))
}

// Update returns the new rating after an outcome scored in [0,1].
func Update(rating, difficulty, score float64) float64 {
	return rating + KFactor*(score-Expected(rating, difficulty))
}

// ScoreForGrade maps an FSRS-style grade (1=Again..4=Easy) to an outcome score.
// Again is a clear miss; Hard a shaky pass; Good/Easy solid recalls.
func ScoreForGrade(grade int) float64 {
	switch grade {
	case 1: // Again
		return 0.0
	case 2: // Hard
		return 0.4
	case 3: // Good
		return 0.8
	case 4: // Easy
		return 1.0
	default:
		return 0.0
	}
}
