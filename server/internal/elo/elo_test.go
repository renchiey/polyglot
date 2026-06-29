package elo

import "testing"

func TestUpdateRewardsSuccessPenalizesFailure(t *testing.T) {
	const d = 1000.0
	up := Update(1000, d, 1.0)   // easy recall at level
	down := Update(1000, d, 0.0) // total miss at level
	if up <= 1000 {
		t.Errorf("success should raise rating, got %.1f", up)
	}
	if down >= 1000 {
		t.Errorf("failure should lower rating, got %.1f", down)
	}
	// Symmetric around the expected 0.5 outcome at equal rating/difficulty.
	if diffUp, diffDown := up-1000, 1000-down; abs(diffUp-diffDown) > 0.001 {
		t.Errorf("expected symmetric move, got +%.3f / -%.3f", diffUp, diffDown)
	}
}

func TestHarderItemRewardsMore(t *testing.T) {
	atLevel := Update(1000, DifficultyForLevel(2), 1.0)
	harder := Update(1000, DifficultyForLevel(6), 1.0)
	if harder <= atLevel {
		t.Errorf("beating a harder item should gain more: hard=%.2f vs at-level=%.2f", harder, atLevel)
	}
}

func TestLevelRoundTrip(t *testing.T) {
	for lvl := 1; lvl <= 7; lvl++ {
		if got := LevelForRating(DifficultyForLevel(lvl), 7); got != lvl {
			t.Errorf("level %d round-trips to %d", lvl, got)
		}
	}
	// Clamping at the bounds.
	if got := LevelForRating(0, 7); got != 1 {
		t.Errorf("low rating clamps to 1, got %d", got)
	}
	if got := LevelForRating(99999, 7); got != 7 {
		t.Errorf("high rating clamps to maxLevel, got %d", got)
	}
}

func TestScoreForGrade(t *testing.T) {
	if ScoreForGrade(1) != 0.0 || ScoreForGrade(4) != 1.0 {
		t.Error("Again should score 0 and Easy 1")
	}
	if !(ScoreForGrade(2) < ScoreForGrade(3)) {
		t.Error("Hard should score below Good")
	}
}

func abs(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}
