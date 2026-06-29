// Package srs bridges the go-fsrs scheduler types and the sqlc-generated
// database types. go-fsrs uses uint64 counters and a non-nullable LastReview
// (zero value when unreviewed); Postgres has no unsigned integers and models an
// unreviewed card as NULL, so the conversions live here in one place.
package srs

import (
	"time"

	fsrs "github.com/open-spaced-repetition/go-fsrs/v4"

	"github.com/google/uuid"
	"github.com/renchieyang/polyglot/server/internal/db/gen"
)

// ToUpsertParams converts a scheduler Card into params for UpsertCard.
func ToUpsertParams(userID, wordID uuid.UUID, c fsrs.Card) gen.UpsertCardParams {
	return gen.UpsertCardParams{
		UserID:         userID,
		WordID:         wordID,
		Due:            c.Due,
		Stability:      c.Stability,
		Difficulty:     c.Difficulty,
		ScheduledDays:  int64(c.ScheduledDays),
		Reps:           int32(c.Reps),
		Lapses:         int32(c.Lapses),
		State:          int16(c.State),
		LastReview:     nullableTime(c.LastReview),
		RemainingSteps: int32(c.RemainingSteps),
	}
}

// FromRow rebuilds a scheduler Card from a stored row.
func FromRow(row gen.Card) fsrs.Card {
	return fsrs.Card{
		Due:            row.Due,
		Stability:      row.Stability,
		Difficulty:     row.Difficulty,
		ScheduledDays:  uint64(row.ScheduledDays),
		Reps:           uint64(row.Reps),
		Lapses:         uint64(row.Lapses),
		State:          fsrs.State(row.State),
		LastReview:     timeOrZero(row.LastReview),
		RemainingSteps: int(row.RemainingSteps),
	}
}

// nullableTime maps the scheduler's zero LastReview to SQL NULL.
func nullableTime(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

// timeOrZero maps SQL NULL back to the scheduler's zero LastReview.
func timeOrZero(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}
