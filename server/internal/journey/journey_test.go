package journey

import (
	"context"
	"strings"
	"testing"

	"github.com/renchieyang/polyglot/server/internal/generate"
	"github.com/renchieyang/polyglot/server/internal/lexaudit"
	"github.com/renchieyang/polyglot/server/internal/llm"
)

// fixedSeg returns a preset token list (the mock controls the text); mapLex is a
// word->level lexicon. Together they make a real Auditor without loading HSK.
type fixedSeg struct{ tokens []string }

func (f fixedSeg) Segment(string) []string { return f.tokens }

type mapLex map[string]int

func (m mapLex) MaxLevel() int { return 7 }
func (m mapLex) LevelOf(w string) (int, bool) {
	l, ok := m[w]
	return l, ok
}

func TestBuildSegmentsStory(t *testing.T) {
	aud := lexaudit.NewAuditor("zh",
		fixedSeg{tokens: []string{"我", "喜欢", "喝", "咖啡", "。"}},
		mapLex{"我": 1, "喜欢": 1, "喝": 1, "咖啡": 1},
		lexaudit.IsCJK,
	)
	client := llm.NewMockClient("我喜欢喝咖啡。")
	b := NewBuilder(generate.NewPipeline(client, generate.DefaultMaxRounds), nil)

	content, err := b.Build(context.Background(), aud, 1, "coffee", nil)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	// The full story comes back segmented; reassembling the tokens yields the
	// original text, and punctuation is flagged non-CJK.
	if got := join(content.Story); got != "我喜欢喝咖啡。" {
		t.Errorf("reassembled story = %q", got)
	}
	if len(content.Story) != 5 {
		t.Fatalf("expected 5 tokens, got %d", len(content.Story))
	}
	last := content.Story[len(content.Story)-1]
	if last.Text != "。" || last.CJK {
		t.Errorf("trailing punctuation token = %+v, want non-CJK 。", last)
	}
	if !content.Story[0].CJK {
		t.Error("我 should be flagged CJK")
	}
	if content.Interaction.Voice {
		t.Error("interaction voice should be a stub (false) for now")
	}
}

func join(toks []Token) string {
	var b strings.Builder
	for _, t := range toks {
		b.WriteString(t.Text)
	}
	return b.String()
}
