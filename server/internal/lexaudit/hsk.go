package lexaudit

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

// hskNewMaxLevel is the highest new-HSK 3.0 band; level 7 covers HSK 7-9.
const hskNewMaxLevel = 7

// hskNewData is the compiled new-HSK 3.0 word -> level map, vendored in the
// repo and embedded into the binary. Regenerate it with:
//
//	go generate ./internal/lexaudit/...
//
//go:embed data/hsk_new.json
var hskNewData []byte

//go:generate go run data/generate.go

// loadHSKLevelMap parses the embedded word -> level map. Both simplified and
// traditional forms are present, each mapped to the word's HSK level.
func loadHSKLevelMap() (map[string]int, error) {
	var levels map[string]int
	if err := json.Unmarshal(hskNewData, &levels); err != nil {
		return nil, fmt.Errorf("parse embedded HSK data: %w", err)
	}
	if len(levels) == 0 {
		return nil, fmt.Errorf("HSK lexicon is empty")
	}
	return levels, nil
}
