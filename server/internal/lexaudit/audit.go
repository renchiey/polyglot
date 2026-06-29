// Package lexaudit checks whether a sentence stays within a learner's
// vocabulary level. It is language-agnostic: a Segmenter splits text into
// words and a Lexicon maps each word to its difficulty level. Concrete
// languages (e.g. Mandarin/HSK) are assembled in their own files and exposed
// through the Registry.
package lexaudit

// Segmenter splits a piece of text into individual words.
type Segmenter interface {
	Segment(text string) []string
}

// Lexicon maps a word to its difficulty level for some grading scale.
type Lexicon interface {
	// MaxLevel is the highest level on the scale (e.g. 7 for new HSK 3.0).
	MaxLevel() int
	// LevelOf returns the word's level and whether it is known to the lexicon.
	LevelOf(word string) (level int, known bool)
}

// Report is the result of auditing one sentence.
type Report struct {
	Passed      bool   `json:"passed"`
	Language    string `json:"language"`
	TargetLevel int    `json:"target_level"`
	// SentenceLevel is the highest level among known words (0 if none),
	// useful as feedback even when the audit passes.
	SentenceLevel int `json:"sentence_level"`
	// OutOfBounds maps each known word above the target to its level.
	OutOfBounds map[string]int `json:"out_of_bounds"`
	// Unknown lists auditable words absent from the lexicon.
	Unknown []string `json:"unknown"`
}

// Auditor combines a segmenter and a lexicon for one language.
type Auditor struct {
	language  string
	seg       Segmenter
	lex       Lexicon
	auditable func(word string) bool
}

// AuditOption configures a single Audit call.
type AuditOption func(*auditConfig)

type auditConfig struct {
	known map[string]bool
}

// WithKnownWords marks words the learner has already acquired so they are not
// flagged as out-of-bounds or unknown, even when above the target level. This
// models "i" (known input) in the i+1 principle: reusing already-acquired
// vocabulary is desirable, not a violation, regardless of its tier.
func WithKnownWords(words []string) AuditOption {
	return func(c *auditConfig) {
		if len(words) == 0 {
			return
		}
		if c.known == nil {
			c.known = make(map[string]bool, len(words))
		}
		for _, w := range words {
			c.known[w] = true
		}
	}
}

// NewAuditor builds an Auditor. auditable filters out tokens that should not
// be graded (punctuation, whitespace, foreign tokens); if nil, every
// non-empty token is audited.
func NewAuditor(language string, seg Segmenter, lex Lexicon, auditable func(string) bool) *Auditor {
	if auditable == nil {
		auditable = func(string) bool { return true }
	}
	return &Auditor{language: language, seg: seg, lex: lex, auditable: auditable}
}

// MaxLevel reports the highest level on this auditor's scale.
func (a *Auditor) MaxLevel() int { return a.lex.MaxLevel() }

// Segment splits text into tokens using this auditor's segmenter. The tokens
// concatenate back to the original text (punctuation and spaces included), so
// callers can map a character offset to the word that covers it.
func (a *Auditor) Segment(text string) []string { return a.seg.Segment(text) }

// LevelOf reports a word's level on this auditor's scale and whether the word
// is known to the lexicon (e.g. its HSK band, for difficulty estimation).
func (a *Auditor) LevelOf(word string) (level int, known bool) { return a.lex.LevelOf(word) }

// Audit segments text and grades each auditable word against target. A word
// above target is out of bounds; a word missing from the lexicon is unknown.
// The sentence passes only when both buckets are empty. Words passed via
// WithKnownWords are exempt from both checks.
func (a *Auditor) Audit(text string, target int, opts ...AuditOption) Report {
	var cfg auditConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	report := Report{
		Language:    a.language,
		TargetLevel: target,
		OutOfBounds: map[string]int{},
		Unknown:     []string{},
	}

	seen := map[string]bool{}
	for _, word := range a.seg.Segment(text) {
		if seen[word] || !a.auditable(word) {
			continue
		}
		seen[word] = true

		level, inLexicon := a.lex.LevelOf(word)

		// Already-acquired words are never penalized, even above target or
		// outside the lexicon: they are part of "i" (known input).
		if cfg.known[word] {
			if inLexicon && level > report.SentenceLevel {
				report.SentenceLevel = level
			}
			continue
		}

		switch {
		case !inLexicon:
			report.Unknown = append(report.Unknown, word)
		case level > target:
			report.OutOfBounds[word] = level
		}
		if inLexicon && level > report.SentenceLevel {
			report.SentenceLevel = level
		}
	}

	report.Passed = len(report.OutOfBounds) == 0 && len(report.Unknown) == 0
	return report
}
