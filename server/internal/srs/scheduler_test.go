package srs

import (
	"testing"
	"time"

	fsrs "github.com/open-spaced-repetition/go-fsrs/v4"
)

func TestNewCardIsDueAndNew(t *testing.T) {
	now := time.Now()
	c := NewCard(now)
	if c.State != fsrs.New {
		t.Errorf("state = %v, want New", c.State)
	}
	if c.Due.After(now) {
		t.Errorf("new card due %v is after now %v", c.Due, now)
	}
	if c.Reps != 0 {
		t.Errorf("reps = %d, want 0", c.Reps)
	}
}

func TestReviewAdvancesSchedule(t *testing.T) {
	s := NewScheduler()
	now := time.Now()

	good, err := s.Review(NewCard(now), now, fsrs.Good)
	if err != nil {
		t.Fatalf("review: %v", err)
	}
	if good.Reps != 1 {
		t.Errorf("reps = %d, want 1", good.Reps)
	}
	if !good.Due.After(now) {
		t.Errorf("due %v not pushed past now %v", good.Due, now)
	}
	if good.LastReview.IsZero() {
		t.Error("LastReview should be set after a review")
	}
}

func TestReviewGradesOrderByInterval(t *testing.T) {
	s := NewScheduler()
	now := time.Now()
	card := NewCard(now)

	again, _ := s.Review(card, now, fsrs.Again)
	good, _ := s.Review(card, now, fsrs.Good)
	easy, _ := s.Review(card, now, fsrs.Easy)

	// Harder grades must not schedule further out than easier ones.
	if again.Due.After(good.Due) {
		t.Errorf("Again due %v after Good due %v", again.Due, good.Due)
	}
	if good.Due.After(easy.Due) {
		t.Errorf("Good due %v after Easy due %v", good.Due, easy.Due)
	}
}

func TestParseRating(t *testing.T) {
	cases := map[int]struct {
		want fsrs.Rating
		ok   bool
	}{
		1: {fsrs.Again, true},
		2: {fsrs.Hard, true},
		3: {fsrs.Good, true},
		4: {fsrs.Easy, true},
		0: {0, false},
		5: {0, false},
	}
	for in, exp := range cases {
		got, ok := ParseRating(in)
		if ok != exp.ok || got != exp.want {
			t.Errorf("ParseRating(%d) = (%v, %v), want (%v, %v)", in, got, ok, exp.want, exp.ok)
		}
	}
}
