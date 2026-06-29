package generate

import (
	"fmt"
	"sort"
	"strings"

	"github.com/renchieyang/polyglot/server/internal/lexaudit"
	"github.com/renchieyang/polyglot/server/internal/llm"
)

// These prompts implement steps 1 (Generation) and 3 (Correction) of the
// PRD's Lexical Auditor pipeline. The deterministic audit (step 2) lives in
// internal/lexaudit and runs between them.

const generationSystem = `You are a Mandarin Chinese writing engine for a comprehensible-input language tutor.

Hard rules:
- Use ONLY vocabulary at or below HSK level %d (new HSK 3.0 scale, levels 1-7).
- Lean on the learner's KNOWN words: ~90-95%% of the content must come from them.
- Introduce at most a small amount (~5%%) of NEW words, and only ones within the HSK level above.
- Never use a word above the target level, even if it would read more naturally.
- Keep it short: at most %d Chinese sentences.
- Write natural, coherent Simplified Chinese. Prefer common everyday phrasing.

Output rules:
- Output ONLY the Chinese text. No pinyin, no translation, no English, no titles, no quotation marks, no commentary.`

const correctionSystem = `You are a Mandarin Chinese text corrector for a comprehensible-input language tutor.

A lexical audit flagged words in the draft that are above the learner's level or unknown.
Rewrite the draft so that EVERY flagged word is replaced with simpler vocabulary at or below
HSK level %d (new HSK 3.0 scale). Preserve the original meaning, tone, and roughly the same
length. Do not introduce any new out-of-level words.

Output rules:
- Output ONLY the corrected Chinese text. No pinyin, no translation, no English, no commentary.`

// buildGenerationPrompt assembles step 1: the initial generation request.
func buildGenerationPrompt(req GenRequest) llm.Request {
	maxSentences := 6
	system := fmt.Sprintf(generationSystem, req.TargetLevel, maxSentences)

	var b strings.Builder
	if topic := strings.TrimSpace(req.Topic); topic != "" {
		fmt.Fprintf(&b, "Topic: %s\n", topic)
	}
	kind := strings.TrimSpace(req.Kind)
	if kind == "" {
		kind = "short story"
	}
	fmt.Fprintf(&b, "Write a %s.\n", kind)
	fmt.Fprintf(&b, "Target level: HSK %d.\n", req.TargetLevel)
	if len(req.MustInclude) > 0 {
		fmt.Fprintf(&b, "The text MUST naturally use these words: %s\n", strings.Join(req.MustInclude, "、"))
	}
	b.WriteString(knownWordsBlock(req.KnownWords))

	return llm.Request{
		System:      system,
		Messages:    []llm.Message{{Role: llm.RoleUser, Content: b.String()}},
		Temperature: 0.7,
		MaxTokens:   600,
	}
}

// buildCorrectionPrompt assembles step 3: a critique pass seeded with the
// specific words the audit rejected.
func buildCorrectionPrompt(req GenRequest, draft string, rep lexaudit.Report) llm.Request {
	system := fmt.Sprintf(correctionSystem, req.TargetLevel)

	var b strings.Builder
	b.WriteString("Draft:\n")
	b.WriteString(draft)
	b.WriteString("\n\n")

	if len(rep.OutOfBounds) > 0 {
		b.WriteString("Words above the target level (word -> its HSK level), replace each:\n")
		for _, w := range sortedKeys(rep.OutOfBounds) {
			fmt.Fprintf(&b, "  %s (HSK %d)\n", w, rep.OutOfBounds[w])
		}
	}
	if len(rep.Unknown) > 0 {
		fmt.Fprintf(&b, "Unknown / out-of-scale words, replace each: %s\n", strings.Join(rep.Unknown, ", "))
	}
	fmt.Fprintf(&b, "\nTarget level: HSK %d.\n", req.TargetLevel)
	if len(req.MustInclude) > 0 {
		fmt.Fprintf(&b, "Keep these words in the text: %s; only replace the flagged words.\n", strings.Join(req.MustInclude, "、"))
	}
	b.WriteString(knownWordsBlock(req.KnownWords))

	return llm.Request{
		System:      system,
		Messages:    []llm.Message{{Role: llm.RoleUser, Content: b.String()}},
		Temperature: 0.3,
		MaxTokens:   600,
	}
}

// knownWordsBlock renders the learner's known vocabulary for the prompt. The
// list is capped so the prompt stays bounded for users with large vaults; the
// HSK level ceiling already constrains the model when the list is truncated.
func knownWordsBlock(known []string) string {
	const max = 300
	if len(known) == 0 {
		return "The learner has no saved words yet; stay well within the target level.\n"
	}
	if len(known) > max {
		known = known[:max]
	}
	return "Known words to prefer: " + strings.Join(known, " ") + "\n"
}

func sortedKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
