package lexaudit

import (
	"strings"
	"testing"
)

// fakeSegmenter splits on spaces so tests avoid loading a real dictionary.
type fakeSegmenter struct{}

func (fakeSegmenter) Segment(text string) []string {
	return strings.Fields(text)
}

// fakeLexicon is a fixed word -> level map.
type fakeLexicon map[string]int

func (l fakeLexicon) MaxLevel() int { return 7 }

func (l fakeLexicon) LevelOf(word string) (int, bool) {
	lvl, ok := l[word]
	return lvl, ok
}

func newTestAuditor() *Auditor {
	lex := fakeLexicon{"a": 1, "b": 2, "c": 5}
	return NewAuditor("test", fakeSegmenter{}, lex, nil)
}

func TestAuditPassesWithinLevel(t *testing.T) {
	r := newTestAuditor().Audit("a b a", 2)
	if !r.Passed {
		t.Fatalf("expected pass, got %+v", r)
	}
	if r.SentenceLevel != 2 {
		t.Errorf("SentenceLevel = %d, want 2", r.SentenceLevel)
	}
	if len(r.OutOfBounds) != 0 || len(r.Unknown) != 0 {
		t.Errorf("expected empty buckets, got %+v", r)
	}
}

func TestAuditFailsAboveLevel(t *testing.T) {
	r := newTestAuditor().Audit("a b c", 2)
	if r.Passed {
		t.Fatalf("expected fail, got %+v", r)
	}
	if got, ok := r.OutOfBounds["c"]; !ok || got != 5 {
		t.Errorf("OutOfBounds[c] = %d (ok=%v), want 5", got, ok)
	}
	if r.SentenceLevel != 5 {
		t.Errorf("SentenceLevel = %d, want 5", r.SentenceLevel)
	}
}

func TestAuditFailsOnUnknown(t *testing.T) {
	r := newTestAuditor().Audit("a z", 5)
	if r.Passed {
		t.Fatalf("expected fail, got %+v", r)
	}
	if len(r.Unknown) != 1 || r.Unknown[0] != "z" {
		t.Errorf("Unknown = %v, want [z]", r.Unknown)
	}
}

func TestAuditExemptsKnownWordsAboveLevel(t *testing.T) {
	// "c" is HSK 5, above target 2, but the learner already knows it.
	r := newTestAuditor().Audit("a b c", 2, WithKnownWords([]string{"c"}))
	if !r.Passed {
		t.Fatalf("expected pass with known word exempt, got %+v", r)
	}
	if len(r.OutOfBounds) != 0 {
		t.Errorf("OutOfBounds = %v, want empty", r.OutOfBounds)
	}
	// Known in-lexicon words still count toward the sentence level.
	if r.SentenceLevel != 5 {
		t.Errorf("SentenceLevel = %d, want 5", r.SentenceLevel)
	}
}

func TestAuditExemptsKnownUnknownWords(t *testing.T) {
	// "z" is absent from the lexicon but the learner knows it (e.g. a name).
	r := newTestAuditor().Audit("a z", 2, WithKnownWords([]string{"z"}))
	if !r.Passed {
		t.Fatalf("expected pass, got %+v", r)
	}
	if len(r.Unknown) != 0 {
		t.Errorf("Unknown = %v, want empty", r.Unknown)
	}
}

func TestAuditDedupesWords(t *testing.T) {
	r := newTestAuditor().Audit("c c c", 2)
	if len(r.OutOfBounds) != 1 {
		t.Errorf("expected 1 out-of-bounds word, got %v", r.OutOfBounds)
	}
}
