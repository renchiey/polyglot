package generate

import (
	"context"
	"strings"
	"testing"

	"github.com/renchieyang/polyglot/server/internal/lexaudit"
	"github.com/renchieyang/polyglot/server/internal/llm"
)

// fakeAuditor fails any text containing a banned substring; otherwise passes.
type fakeAuditor struct {
	banned string
	level  int
}

func (f fakeAuditor) Audit(text string, target int, _ ...lexaudit.AuditOption) lexaudit.Report {
	rep := lexaudit.Report{
		Language:    "zh",
		TargetLevel: target,
		OutOfBounds: map[string]int{},
		Unknown:     []string{},
	}
	if strings.Contains(text, f.banned) {
		rep.OutOfBounds[f.banned] = f.level
	}
	rep.Passed = len(rep.OutOfBounds) == 0
	return rep
}

func TestPipelineCorrectsThenPasses(t *testing.T) {
	// First generation is out of bounds; the correction pass returns clean text.
	client := llm.NewMockClient("", "他很高兴。", "他很好。")
	p := NewPipeline(client, DefaultMaxRounds)
	aud := fakeAuditor{banned: "高兴", level: 2}

	res, err := p.Run(context.Background(), aud, GenRequest{
		Language: "zh", TargetLevel: 1, Topic: "feelings", Kind: "sentence",
		KnownWords: []string{"他", "很", "好"},
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if !res.Passed {
		t.Errorf("expected pass after correction, got report %+v", res.Report)
	}
	if res.Rounds != 2 {
		t.Errorf("expected 2 rounds (generate + 1 correction), got %d", res.Rounds)
	}
	if res.Text != "他很好。" {
		t.Errorf("expected corrected text, got %q", res.Text)
	}

	// The second LLM call must be a correction seeded with the flagged word.
	calls := client.Calls()
	if len(calls) != 2 {
		t.Fatalf("expected 2 LLM calls, got %d", len(calls))
	}
	if !strings.Contains(calls[1].Messages[0].Content, "高兴") {
		t.Errorf("correction prompt missing flagged word, got %q", calls[1].Messages[0].Content)
	}
}

func TestPipelineStopsAtMaxRounds(t *testing.T) {
	// Every reply stays out of bounds; the loop must stop at MaxRounds and
	// return the last (still failing) candidate.
	client := llm.NewMockClient("他很高兴。")
	p := NewPipeline(client, 3)
	aud := fakeAuditor{banned: "高兴", level: 2}

	res, err := p.Run(context.Background(), aud, GenRequest{Language: "zh", TargetLevel: 1})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.Passed {
		t.Error("expected failure when no candidate is clean")
	}
	if res.Rounds != 3 {
		t.Errorf("expected to stop at 3 rounds, got %d", res.Rounds)
	}
	if len(client.Calls()) != 3 {
		t.Errorf("expected 3 LLM calls, got %d", len(client.Calls()))
	}
}

// fixedSeg returns a preset token list, ignoring input — the mock controls the
// text, so we control the tokens. mapLex is a word->level lexicon. Together they
// build a real lexaudit.Auditor whose WithKnownWords exemption we exercise.
type fixedSeg struct{ tokens []string }

func (f fixedSeg) Segment(string) []string { return f.tokens }

type mapLex map[string]int

func (m mapLex) MaxLevel() int { return 7 }
func (m mapLex) LevelOf(w string) (int, bool) {
	l, ok := m[w]
	return l, ok
}

func TestPipelineExemptsMustInclude(t *testing.T) {
	// 高兴 (level 2) is above target 1 and would normally be flagged, but as the
	// MustInclude target it must be exempt, so the run passes in one round. Uses
	// a real Auditor so the exemption goes through the real audit logic.
	aud := lexaudit.NewAuditor("zh",
		fixedSeg{tokens: []string{"很", "高兴"}},
		mapLex{"很": 1, "高兴": 2},
		lexaudit.IsCJK,
	)
	client := llm.NewMockClient("", "我很高兴。")
	p := NewPipeline(client, DefaultMaxRounds)

	res, err := p.Run(context.Background(), aud, GenRequest{
		Language: "zh", TargetLevel: 1, MustInclude: []string{"高兴"},
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !res.Passed || res.Rounds != 1 {
		t.Errorf("expected pass in 1 round with MustInclude exempt, got passed=%v rounds=%d report=%+v",
			res.Passed, res.Rounds, res.Report)
	}

	// Sanity: without the exemption the same text fails the audit.
	if plain := aud.Audit("我很高兴。", 1); plain.Passed {
		t.Error("expected 高兴 to fail the audit when not exempt")
	}

	// The generation prompt must instruct the model to use the target word.
	if !strings.Contains(client.Calls()[0].Messages[0].Content, "高兴") {
		t.Error("generation prompt missing MustInclude word")
	}
}

func TestPipelinePassesFirstTry(t *testing.T) {
	client := llm.NewMockClient("", "他很好。")
	p := NewPipeline(client, DefaultMaxRounds)
	aud := fakeAuditor{banned: "高兴", level: 2}

	res, err := p.Run(context.Background(), aud, GenRequest{Language: "zh", TargetLevel: 1})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !res.Passed || res.Rounds != 1 {
		t.Errorf("expected pass in 1 round, got passed=%v rounds=%d", res.Passed, res.Rounds)
	}
}
