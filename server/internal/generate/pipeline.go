// Package generate implements the PRD's Lexical Auditor pipeline: an LLM
// generates comprehensible-input (i+1) text, a deterministic audit checks it
// stays within the learner's level, and a correction pass rewrites any
// out-of-bounds words. Generation and correction loop until the audit passes
// or a round budget is exhausted.
package generate

import (
	"context"
	"strings"

	"github.com/renchieyang/polyglot/server/internal/lexaudit"
	"github.com/renchieyang/polyglot/server/internal/llm"
)

// Auditor grades text against a target level. *lexaudit.Auditor satisfies it;
// the indirection keeps the pipeline testable with a fake.
type Auditor interface {
	Audit(text string, targetLevel int, opts ...lexaudit.AuditOption) lexaudit.Report
}

// GenRequest describes the content to produce. KnownWords is the learner's
// vault, used to bias generation toward already-acquired vocabulary.
type GenRequest struct {
	Language    string
	TargetLevel int
	Topic       string
	Kind        string
	KnownWords  []string
	// MustInclude, when non-empty, forces the output to contain these words and
	// exempts them from the audit. Used for generative testing (quiz a due word)
	// and Daily Journey (embed several due words in one story).
	MustInclude []string
}

// GenResult is the outcome of a pipeline run. Text is the best candidate (the
// last one produced, even if it still failed after MaxRounds); Report is its
// audit; Rounds counts the LLM calls (1 generation + N corrections).
type GenResult struct {
	Text   string          `json:"text"`
	Passed bool            `json:"passed"`
	Rounds int             `json:"rounds"`
	Report lexaudit.Report `json:"report"`
}

// Pipeline runs generation → audit → correction against an LLM.
type Pipeline struct {
	LLM       llm.Client
	MaxRounds int
}

// DefaultMaxRounds bounds the total number of LLM calls per request.
const DefaultMaxRounds = 3

// NewPipeline builds a Pipeline. A non-positive maxRounds falls back to
// DefaultMaxRounds.
func NewPipeline(client llm.Client, maxRounds int) *Pipeline {
	if maxRounds < 1 {
		maxRounds = DefaultMaxRounds
	}
	return &Pipeline{LLM: client, MaxRounds: maxRounds}
}

// Run executes the pipeline: one generation call, then correction calls until
// the audit passes or the round budget runs out. The last candidate is always
// returned, so callers can surface near-misses with their report.
func (p *Pipeline) Run(ctx context.Context, aud Auditor, req GenRequest) (GenResult, error) {
	// Acquired vocabulary is exempt from the audit: reusing known words is the
	// "i" in i+1, even when those words sit above the target tier. A generative-
	// testing target is exempt too — it is the point of the exercise.
	allow := req.KnownWords
	if len(req.MustInclude) > 0 {
		allow = append(append([]string(nil), req.KnownWords...), req.MustInclude...)
	}
	known := lexaudit.WithKnownWords(allow)

	resp, err := p.LLM.Complete(ctx, buildGenerationPrompt(req))
	if err != nil {
		return GenResult{}, err
	}
	text := strings.TrimSpace(resp.Text)
	report := aud.Audit(text, req.TargetLevel, known)
	rounds := 1

	for !report.Passed && rounds < p.MaxRounds {
		resp, err = p.LLM.Complete(ctx, buildCorrectionPrompt(req, text, report))
		if err != nil {
			return GenResult{}, err
		}
		text = strings.TrimSpace(resp.Text)
		report = aud.Audit(text, req.TargetLevel, known)
		rounds++
	}

	return GenResult{
		Text:   text,
		Passed: report.Passed,
		Rounds: rounds,
		Report: report,
	}, nil
}
