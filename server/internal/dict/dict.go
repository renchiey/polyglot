// Package dict provides Mandarin word lookups — pinyin (tone marks),
// definitions, and per-character breakdowns — backing the app's Assisted
// Noticing feature. Data is derived from CC-CEDICT (CC BY-SA 4.0), filtered to
// HSK 3.0 vocabulary plus every single CJK character and embedded in the binary.
package dict

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"unicode/utf8"
)

// cedictData is the gzipped lookup table, vendored in the repo and embedded.
// Regenerate it with:
//
//	go generate ./internal/dict/...
//
//go:embed data/cedict.json.gz
var cedictData []byte

//go:generate go run data/generate.go

// Entry is a single dictionary result.
type Entry struct {
	Word        string   `json:"word"`
	Pinyin      string   `json:"pinyin"`
	Definitions []string `json:"definitions"`
}

type rawEntry struct {
	P string   `json:"p"`
	D []string `json:"d"`
}

// Dictionary is an in-memory word -> entry table.
type Dictionary struct {
	entries map[string]rawEntry
}

// Load decompresses and parses the embedded dictionary.
func Load() (*Dictionary, error) {
	gz, err := gzip.NewReader(bytes.NewReader(cedictData))
	if err != nil {
		return nil, fmt.Errorf("open embedded dict: %w", err)
	}
	defer gz.Close()

	data, err := io.ReadAll(gz)
	if err != nil {
		return nil, fmt.Errorf("read embedded dict: %w", err)
	}

	var entries map[string]rawEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("parse embedded dict: %w", err)
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("dictionary is empty")
	}
	return &Dictionary{entries: entries}, nil
}

// Lookup returns the entry for an exact word.
func (d *Dictionary) Lookup(word string) (Entry, bool) {
	raw, ok := d.entries[word]
	if !ok {
		return Entry{}, false
	}
	return Entry{Word: word, Pinyin: raw.P, Definitions: raw.D}, true
}

// Breakdown returns the entry for each character of a multi-character word, in
// order, skipping characters absent from the dictionary. It returns nil for a
// single-character (or empty) word, since there is nothing to break down.
func (d *Dictionary) Breakdown(word string) []Entry {
	if utf8.RuneCountInString(word) < 2 {
		return nil
	}
	var out []Entry
	for _, r := range word {
		if e, ok := d.Lookup(string(r)); ok {
			out = append(out, e)
		}
	}
	return out
}
