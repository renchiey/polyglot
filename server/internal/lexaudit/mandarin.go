package lexaudit

import (
	"fmt"

	"github.com/go-ego/gse"
)

// gseSegmenter adapts gse to the Segmenter interface.
type gseSegmenter struct {
	seg gse.Segmenter
}

// Segment splits Chinese text into words. HMM is disabled on purpose: the HMM
// invents out-of-dictionary words (e.g. merging "喝咖啡" into one token), which
// would mislabel valid in-level vocabulary as unknown. Dictionary-DAG
// segmentation keeps boundaries aligned with real words, including the HSK
// vocabulary seeded into the dictionary at build time.
func (g *gseSegmenter) Segment(text string) []string {
	return g.seg.CutDAGNoHMM(text)
}

// hskLexicon is a word -> HSK level map.
type hskLexicon struct {
	levels   map[string]int
	maxLevel int
}

func (l *hskLexicon) MaxLevel() int { return l.maxLevel }

func (l *hskLexicon) LevelOf(word string) (int, bool) {
	level, ok := l.levels[word]
	return level, ok
}

// IsCJK reports whether a token contains at least one CJK ideograph, which is
// how we skip punctuation, whitespace, Latin letters and digits when grading.
func IsCJK(token string) bool {
	for _, r := range token {
		if r >= '一' && r <= '鿿' {
			return true
		}
	}
	return false
}

// newMandarinAuditor builds the Mandarin (new HSK 3.0) auditor. The segmenter's
// dictionary is built from the HSK vocabulary alone (not gse's general
// dictionary), so every recognised token is an HSK word and boundaries align
// with HSK entries by construction; characters outside the HSK vocabulary fall
// out as single tokens and are reported as unknown. Every word is given the
// same frequency, so gse's route selection minimises the token count and yields
// the longest HSK-word match at each position (e.g. "服务员" over "服务" + "员").
// The HSK level map is embedded in the binary.
func newMandarinAuditor() (*Auditor, error) {
	levels, err := loadHSKLevelMap()
	if err != nil {
		return nil, err
	}

	// A uniform frequency makes gse prefer the longest HSK-word match; the value
	// only needs to exceed the freq of 2 that gse forces on single characters,
	// so that multi-character HSK words win over a char-by-char split.
	dict := make([]map[string]string, 0, len(levels))
	for word := range levels {
		dict = append(dict, map[string]string{"text": word, "freq": "100"})
	}

	var seg gse.Segmenter
	if err := seg.LoadDictMap(dict); err != nil {
		return nil, fmt.Errorf("load HSK dictionary: %w", err)
	}

	lex := &hskLexicon{levels: levels, maxLevel: hskNewMaxLevel}
	return NewAuditor("zh", &gseSegmenter{seg: seg}, lex, IsCJK), nil
}
