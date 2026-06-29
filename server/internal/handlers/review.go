package handlers

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	fsrs "github.com/open-spaced-repetition/go-fsrs/v4"

	"github.com/renchieyang/polyglot/server/internal/db/gen"
	"github.com/renchieyang/polyglot/server/internal/srs"
)

// applyReview runs an FSRS review for a (user, word) and persists the result,
// starting from a fresh card when none exists yet. Shared by the study queue
// (/cards/review) and Daily Journey recall answers so both schedule identically.
func applyReview(ctx context.Context, q *gen.Queries, sched *srs.Scheduler, userID, wordID uuid.UUID, grade fsrs.Rating) (gen.Card, error) {
	now := time.Now()
	card := srs.NewCard(now)
	if row, err := q.GetCardByWord(ctx, gen.GetCardByWordParams{UserID: userID, WordID: wordID}); err == nil {
		card = srs.FromRow(row)
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return gen.Card{}, err
	}

	updated, err := sched.Review(card, now, grade)
	if err != nil {
		return gen.Card{}, err
	}
	return q.UpsertCard(ctx, srs.ToUpsertParams(userID, wordID, updated))
}
