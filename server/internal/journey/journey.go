// Package journey generates the story behind the PRD's Daily Journey. The
// learner reads it (Input), marks the words they don't know, learns those, and
// recalls them — but those later phases are driven client-side over the words,
// lookup, recall and review endpoints, so this package only produces the story.
package journey

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/renchieyang/polyglot/server/internal/dict"
	"github.com/renchieyang/polyglot/server/internal/generate"
	"github.com/renchieyang/polyglot/server/internal/lexaudit"
	"github.com/renchieyang/polyglot/server/internal/llm"
)

// Token is one segmented piece of the story. Pinyin is supplied so the client
// can reveal it on demand (it stays hidden by default).
type Token struct {
	Text   string `json:"text"`
	Pinyin string `json:"pinyin,omitempty"`
	CJK    bool   `json:"cjk"`
}

// Interaction is the handoff to the conversation phase. Voice is Phase 6, so for
// now it is a labelled stub.
type Interaction struct {
	Persona string `json:"persona"`
	Opening string `json:"opening"`
	Voice   bool   `json:"voice"`
}

// Suggestion is an extra topic-related word offered in the Learn phase, so the
// learner picks up something new even if they understood the whole story.
type Suggestion struct {
	Term   string `json:"term"`
	Pinyin string `json:"pinyin,omitempty"`
	Gloss  string `json:"gloss,omitempty"`
}

// Content is the persisted session payload (stored as JSON).
type Content struct {
	Story       []Token      `json:"story"`
	Suggestions []Suggestion `json:"suggestions,omitempty"`
	Interaction Interaction  `json:"interaction"`
	Passed      bool         `json:"passed"`
}

// Builder assembles journey stories from the generation pipeline and dictionary.
type Builder struct {
	Pipeline *generate.Pipeline
	Dict     *dict.Dictionary
}

func NewBuilder(p *generate.Pipeline, d *dict.Dictionary) *Builder {
	return &Builder{Pipeline: p, Dict: d}
}

// Build generates an i+1 story at the learner's level and returns it segmented,
// with per-word pinyin. No words are pre-selected — the learner marks their own
// unknowns while reading.
func (b *Builder) Build(ctx context.Context, aud *lexaudit.Auditor, level int, topic string, known []string) (Content, error) {
	res, err := b.Pipeline.Run(ctx, aud, generate.GenRequest{
		Language:    "zh",
		TargetLevel: level,
		Topic:       topic,
		Kind:        "short story (4-6 sentences)",
		KnownWords:  known,
	})
	if err != nil {
		return Content{}, err
	}

	story := make([]Token, 0)
	for _, tok := range aud.Segment(res.Text) {
		t := Token{Text: tok, CJK: lexaudit.IsCJK(tok)}
		if t.CJK {
			t.Pinyin = b.pinyin(tok)
		}
		story = append(story, t)
	}

	return Content{
		Story:       story,
		Suggestions: b.suggest(ctx, aud, level, topic, known),
		Passed:      res.Passed,
		Interaction: Interaction{
			Persona: "A character from the story",
			Opening: "Voice chat is coming soon — you'll talk this story through here.",
			Voice:   false,
		},
	}, nil
}

const suggestSystem = `You suggest useful Mandarin vocabulary for a learner. Output ONLY a JSON array like [{"word":"图书馆","gloss":"library"}]. Use Simplified Chinese words at or below HSK level %d. No pinyin, no commentary.`

// suggest asks the LLM for a few topic-related words at or below the level,
// keeping only those in the HSK lexicon and not already known. Best effort: any
// failure yields no suggestions and the journey proceeds.
func (b *Builder) suggest(ctx context.Context, aud *lexaudit.Auditor, level int, topic string, known []string) []Suggestion {
	resp, err := b.Pipeline.LLM.Complete(ctx, llm.Request{
		System:      fmt.Sprintf(suggestSystem, level),
		Messages:    []llm.Message{{Role: llm.RoleUser, Content: "Suggest 6 words about: " + topic}},
		Temperature: 0.6,
		MaxTokens:   300,
	})
	if err != nil {
		return nil
	}

	var items []struct {
		Word  string `json:"word"`
		Gloss string `json:"gloss"`
	}
	a, z := strings.Index(resp.Text, "["), strings.LastIndex(resp.Text, "]")
	if a < 0 || z <= a || json.Unmarshal([]byte(resp.Text[a:z+1]), &items) != nil {
		return nil
	}

	knownSet := make(map[string]bool, len(known))
	for _, k := range known {
		knownSet[k] = true
	}
	seen := make(map[string]bool)

	var out []Suggestion
	for _, it := range items {
		w := strings.TrimSpace(it.Word)
		if w == "" || seen[w] || knownSet[w] || !lexaudit.IsCJK(w) {
			continue
		}
		if lvl, ok := aud.LevelOf(w); !ok || lvl > level {
			continue
		}
		seen[w] = true
		out = append(out, Suggestion{Term: w, Pinyin: b.pinyin(w), Gloss: b.gloss(it.Gloss, w)})
		if len(out) >= 5 {
			break
		}
	}
	return out
}

// gloss prefers the model's gloss, falling back to the dictionary definition.
func (b *Builder) gloss(modelGloss, word string) string {
	if g := strings.TrimSpace(modelGloss); g != "" {
		return g
	}
	if b.Dict != nil {
		if e, ok := b.Dict.Lookup(word); ok && len(e.Definitions) > 0 {
			return strings.Join(e.Definitions[:min(len(e.Definitions), 2)], "; ")
		}
	}
	return ""
}

func (b *Builder) pinyin(word string) string {
	if b.Dict == nil {
		return ""
	}
	if e, ok := b.Dict.Lookup(word); ok {
		return e.Pinyin
	}
	return ""
}
