package srs

import (
	"testing"
	"time"

	fsrs "github.com/open-spaced-repetition/go-fsrs/v4"

	"github.com/google/uuid"
	"github.com/renchieyang/polyglot/server/internal/db/gen"
)

func TestRoundTrip(t *testing.T) {
	userID, wordID := uuid.New(), uuid.New()
	reviewed := time.Date(2026, 6, 15, 9, 0, 0, 0, time.UTC)

	in := fsrs.Card{
		Due:            time.Date(2026, 6, 20, 0, 0, 0, 0, time.UTC),
		Stability:      12.5,
		Difficulty:     5.1,
		ScheduledDays:  5,
		Reps:           3,
		Lapses:         1,
		State:          fsrs.Review,
		LastReview:     reviewed,
		RemainingSteps: 2,
	}

	p := ToUpsertParams(userID, wordID, in)
	if p.UserID != userID || p.WordID != wordID {
		t.Fatalf("ids not carried through: %v %v", p.UserID, p.WordID)
	}

	// Simulate the row the DB returns from the upsert.
	row := gen.Card{
		ID: uuid.New(), UserID: p.UserID, WordID: p.WordID,
		Due: p.Due, Stability: p.Stability, Difficulty: p.Difficulty,
		ScheduledDays: p.ScheduledDays, Reps: p.Reps, Lapses: p.Lapses,
		State: p.State, LastReview: p.LastReview, RemainingSteps: p.RemainingSteps,
	}

	out := FromRow(row)
	if out != in {
		t.Fatalf("round trip changed card:\n in=%+v\nout=%+v", in, out)
	}
}

func TestLastReviewNullHandling(t *testing.T) {
	// A brand-new card has a zero LastReview, which must become SQL NULL.
	newCard := fsrs.NewCard()
	if !newCard.LastReview.IsZero() {
		t.Skip("go-fsrs NewCard no longer zeroes LastReview")
	}

	p := ToUpsertParams(uuid.New(), uuid.New(), newCard)
	if p.LastReview != nil {
		t.Fatalf("zero LastReview should map to NULL, got %v", *p.LastReview)
	}

	// And NULL must come back as a zero time, not panic.
	got := FromRow(gen.Card{LastReview: nil})
	if !got.LastReview.IsZero() {
		t.Fatalf("NULL last_review should map to zero time, got %v", got.LastReview)
	}
}
