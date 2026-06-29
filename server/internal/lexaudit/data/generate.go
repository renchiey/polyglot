//go:build ignore

// Command generate downloads the new-HSK 3.0 wordlists and compiles them into
// hsk_new.json (a word -> level map) for embedding. Run from the lexaudit
// package directory via `go generate ./internal/lexaudit/...`.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// Each level file lists the words first introduced at that level (exclusive
// lists), so the union over all levels yields a word -> own-level map.
const baseURL = "https://raw.githubusercontent.com/drkameleon/complete-hsk-vocabulary/master/wordlists/exclusive/new"

const maxLevel = 7

const outFile = "data/hsk_new.json"

type entry struct {
	Simplified string `json:"simplified"`
	Forms      []struct {
		Traditional string `json:"traditional"`
	} `json:"forms"`
}

func main() {
	levels := make(map[string]int)
	for level := 1; level <= maxLevel; level++ {
		entries, err := fetchLevel(level)
		if err != nil {
			log.Fatalf("level %d: %v", level, err)
		}
		for _, e := range entries {
			index(levels, e.Simplified, level)
			for _, f := range e.Forms {
				index(levels, f.Traditional, level)
			}
		}
	}
	if len(levels) == 0 {
		log.Fatal("no words compiled")
	}

	data, err := json.MarshalIndent(levels, "", "")
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(outFile, data, 0o644); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("wrote %d words to %s\n", len(levels), outFile)
}

func index(levels map[string]int, word string, level int) {
	if word == "" {
		return
	}
	if cur, ok := levels[word]; !ok || level < cur {
		levels[word] = level
	}
}

func fetchLevel(level int) ([]entry, error) {
	url := fmt.Sprintf("%s/%d.json", baseURL, level)
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var entries []entry
	if err := json.Unmarshal(body, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}
