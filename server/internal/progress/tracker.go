// Package progress turns graded learner actions into Linguistic Elo updates and
// derives content difficulty from the resulting ratings. It is the single place
// that decides which vector an action nudges, so handlers stay thin.
package progress

import (
	"context"

	"github.com/google/uuid"

	"github.com/renchieyang/polyglot/server/internal/db/gen"
	"github.com/renchieyang/polyglot/server/internal/elo"
	"github.com/renchieyang/polyglot/server/internal/lexaudit"
)

// Tracker reads and writes the linguistic_elo table and looks up word
// difficulty via the Mandarin lexicon.
type Tracker struct {
	Queries  *gen.Queries
	Registry *lexaudit.Registry
}

func NewTracker(q *gen.Queries, r *lexaudit.Registry) *Tracker {
	return &Tracker{Queries: q, Registry: r}
}

// RecordVocabReview nudges the user's Vocabulary Elo from a graded SRS review.
// Difficulty comes from the word's HSK level; an out-of-lexicon word is treated
// as being at the learner's current level (a neutral, expected-0.5 outcome).
func (t *Tracker) RecordVocabReview(ctx context.Context, userID uuid.UUID, term string, grade int) (gen.LinguisticElo, error) {
	row, err := t.Queries.EnsureElo(ctx, userID)
	if err != nil {
		return gen.LinguisticElo{}, err
	}

	difficulty := row.Vocabulary
	if aud, err := t.Registry.Get("zh"); err == nil {
		if level, known := aud.LevelOf(term); known {
			difficulty = elo.DifficultyForLevel(level)
		}
	}

	row.Vocabulary = elo.Update(row.Vocabulary, difficulty, elo.ScoreForGrade(grade))
	return t.Queries.UpdateElo(ctx, gen.UpdateEloParams{
		UserID:     userID,
		Vocabulary: row.Vocabulary,
		Syntax:     row.Syntax,
		Listening:  row.Listening,
		Speaking:   row.Speaking,
	})
}

// RecordReading nudges Vocabulary Elo from a reading-comprehension outcome
// (e.g. a graded translation): score in [0,1] against an item at the given HSK
// level. An out-of-range level is treated as level 1.
func (t *Tracker) RecordReading(ctx context.Context, userID uuid.UUID, level int, score float64) (gen.LinguisticElo, error) {
	if level < 1 {
		level = 1
	}
	row, err := t.Queries.EnsureElo(ctx, userID)
	if err != nil {
		return gen.LinguisticElo{}, err
	}
	row.Vocabulary = elo.Update(row.Vocabulary, elo.DifficultyForLevel(level), score)
	return t.Queries.UpdateElo(ctx, gen.UpdateEloParams{
		UserID:     userID,
		Vocabulary: row.Vocabulary,
		Syntax:     row.Syntax,
		Listening:  row.Listening,
		Speaking:   row.Speaking,
	})
}

// Ratings returns the user's current rating row, creating it on first access.
func (t *Tracker) Ratings(ctx context.Context, userID uuid.UUID) (gen.LinguisticElo, error) {
	return t.Queries.EnsureElo(ctx, userID)
}

// RecommendedLevel maps the user's Vocabulary Elo to an HSK content level in
// [1, maxLevel] — used to pick difficulty for generation and recall.
func (t *Tracker) RecommendedLevel(ctx context.Context, userID uuid.UUID, maxLevel int) (int, error) {
	row, err := t.Queries.EnsureElo(ctx, userID)
	if err != nil {
		return 0, err
	}
	return elo.LevelForRating(row.Vocabulary, maxLevel), nil
}
