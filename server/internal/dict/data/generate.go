//go:build ignore

// Command generate downloads CC-CEDICT, filters it to the entries this app can
// surface (HSK 3.0 words plus every single CJK character, for sub-character
// breakdowns), converts numbered pinyin to tone marks, and writes a gzipped
// lookup table for embedding. Run from the dict package directory via
// `go generate ./internal/dict/...`.
//
// CC-CEDICT is licensed CC BY-SA 4.0 (https://cc-cedict.org/wiki/) — see the
// NOTICE emitted alongside the data.
package main

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const (
	cedictURL = "https://www.mdbg.net/chinese/export/cedict/cedict_1_0_ts_utf-8_mdbg.txt.gz"
	hskFile   = "../lexaudit/data/hsk_new.json"
	outFile   = "data/cedict.json.gz"
	maxDefs   = 6
)

// rawEntry is the compact on-disk shape: pinyin + definitions.
type rawEntry struct {
	P string   `json:"p"`
	D []string `json:"d"`
}

// line: TRAD SIMP [pin1 yin1] /def/def/
var lineRE = regexp.MustCompile(`^\S+\s+(\S+)\s+\[([^\]]*)\]\s+/(.*)/\s*$`)

func main() {
	hsk, err := loadHSKSet()
	if err != nil {
		log.Fatalf("hsk set: %v", err)
	}
	log.Printf("loaded %d HSK words", len(hsk))

	lines, err := fetchCedict()
	if err != nil {
		log.Fatalf("fetch cedict: %v", err)
	}

	// A character or word can have several CC-CEDICT readings (e.g. 说 shuo1
	// vs the archaic shui4). The everyday reading reliably carries the most
	// senses, so keep, per simplified form, the reading with the most defs.
	type candidate struct {
		entry rawEntry
		full  int // full sense count, before capping, used to compare readings
	}
	best := make(map[string]candidate)
	for _, ln := range lines {
		if ln == "" || strings.HasPrefix(ln, "#") {
			continue
		}
		m := lineRE.FindStringSubmatch(ln)
		if m == nil {
			continue
		}
		simp, pinyin, defs := m[1], m[2], m[3]

		// Keep HSK vocabulary and every single CJK character (the latter powers
		// sub-character breakdowns of any HSK compound).
		if !hsk[simp] && !isSingleCJK(simp) {
			continue
		}

		senses := splitDefs(defs)
		if len(senses) == 0 {
			continue
		}
		if cur, ok := best[simp]; ok && cur.full >= len(senses) {
			continue // keep the richer reading already stored
		}
		capped := senses
		if len(capped) > maxDefs {
			capped = capped[:maxDefs]
		}
		best[simp] = candidate{entry: rawEntry{P: numberedToMarks(pinyin), D: capped}, full: len(senses)}
	}
	if len(best) == 0 {
		log.Fatal("no entries compiled")
	}

	out := make(map[string]rawEntry, len(best))
	for simp, c := range best {
		out[simp] = c.entry
	}

	if err := writeGzipJSON(outFile, out); err != nil {
		log.Fatalf("write: %v", err)
	}
	fmt.Printf("wrote %d entries to %s\n", len(out), outFile)
}

func loadHSKSet() (map[string]bool, error) {
	data, err := os.ReadFile(hskFile)
	if err != nil {
		return nil, err
	}
	var levels map[string]int
	if err := json.Unmarshal(data, &levels); err != nil {
		return nil, err
	}
	set := make(map[string]bool, len(levels))
	for w := range levels {
		set[w] = true
	}
	return set, nil
}

func fetchCedict() ([]string, error) {
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(cedictURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %s", resp.Status)
	}
	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	var lines []string
	sc := bufio.NewScanner(gz)
	sc.Buffer(make([]byte, 1024*1024), 1024*1024)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	return lines, sc.Err()
}

func splitDefs(defs string) []string {
	parts := strings.Split(defs, "/")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func isSingleCJK(s string) bool {
	if utf8.RuneCountInString(s) != 1 {
		return false
	}
	r, _ := utf8.DecodeRuneInString(s)
	return r >= '一' && r <= '鿿'
}

func writeGzipJSON(path string, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	gz := gzip.NewWriter(f)
	if _, err := gz.Write(data); err != nil {
		return err
	}
	return gz.Close()
}

// --- numbered pinyin -> tone marks ---

var toneMarks = map[rune][4]rune{
	'a': {'ā', 'á', 'ǎ', 'à'},
	'e': {'ē', 'é', 'ě', 'è'},
	'i': {'ī', 'í', 'ǐ', 'ì'},
	'o': {'ō', 'ó', 'ǒ', 'ò'},
	'u': {'ū', 'ú', 'ǔ', 'ù'},
	'ü': {'ǖ', 'ǘ', 'ǚ', 'ǜ'},
}

func numberedToMarks(s string) string {
	fields := strings.Fields(s)
	for i, f := range fields {
		fields[i] = convertSyllable(f)
	}
	return strings.Join(fields, " ")
}

func convertSyllable(syl string) string {
	syl = strings.ReplaceAll(syl, "u:", "ü")
	syl = strings.ReplaceAll(syl, "U:", "Ü")
	if syl == "" {
		return syl
	}
	last := syl[len(syl)-1]
	if last < '1' || last > '5' {
		return syl // no tone digit (punctuation, etc.)
	}
	base := syl[:len(syl)-1]
	tone := int(last - '0')
	if tone >= 5 { // neutral tone: no mark
		return base
	}
	runes := []rune(base)
	idx := targetVowel(runes)
	if idx < 0 {
		return base
	}
	if marks, ok := toneMarks[unicode.ToLower(runes[idx])]; ok {
		accented := marks[tone-1]
		if unicode.IsUpper(runes[idx]) {
			accented = unicode.ToUpper(accented)
		}
		runes[idx] = accented
	}
	return string(runes)
}

func isVowel(r rune) bool {
	switch unicode.ToLower(r) {
	case 'a', 'e', 'i', 'o', 'u', 'ü':
		return true
	}
	return false
}

// targetVowel applies the standard tone-placement rule: a/e takes the mark;
// otherwise in "ou" the o takes it; otherwise the last vowel.
func targetVowel(runes []rune) int {
	for i, r := range runes {
		if l := unicode.ToLower(r); l == 'a' || l == 'e' {
			return i
		}
	}
	for i := 0; i+1 < len(runes); i++ {
		if unicode.ToLower(runes[i]) == 'o' && unicode.ToLower(runes[i+1]) == 'u' {
			return i
		}
	}
	for i := len(runes) - 1; i >= 0; i-- {
		if isVowel(runes[i]) {
			return i
		}
	}
	return -1
}
