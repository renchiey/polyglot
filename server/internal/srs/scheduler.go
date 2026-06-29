package srs

import (
	"time"

	fsrs "github.com/open-spaced-repetition/go-fsrs/v4"
)

// Scheduler applies review grades to cards using FSRS default parameters.
// It is stateless and safe for concurrent use; build one and share it.
type Scheduler struct {
	f *fsrs.FSRS
}

// NewScheduler builds a scheduler with FSRS default parameters.
func NewScheduler() *Scheduler {
	return &Scheduler{f: fsrs.NewFSRS(fsrs.DefaultParam())}
}

// NewCard returns a fresh, never-reviewed card due at now (state New).
func NewCard(now time.Time) fsrs.Card {
	return fsrs.NewCard(now)
}

// Review applies grade to card at now and returns the rescheduled card.
func (s *Scheduler) Review(card fsrs.Card, now time.Time, grade fsrs.Rating) (fsrs.Card, error) {
	info, err := s.f.Next(card, now, grade)
	if err != nil {
		return fsrs.Card{}, err
	}
	return info.Card, nil
}

// ReviewState is the FSRS state for an acquired card. ListKnownTerms is called
// with this to gather the learner's mastered vocabulary for i+1 biasing.
const ReviewState = int16(fsrs.Review)

// ParseRating validates a 1-4 grade (1=Again, 2=Hard, 3=Good, 4=Easy) and maps
// it to an fsrs.Rating. ok is false for any other value.
func ParseRating(n int) (rating fsrs.Rating, ok bool) {
	if n < int(fsrs.Again) || n > int(fsrs.Easy) {
		return 0, false
	}
	return fsrs.Rating(n), true
}
