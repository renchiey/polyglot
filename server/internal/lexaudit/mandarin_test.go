package lexaudit

import "testing"

// TestMandarinAuditor exercises the real Mandarin auditor end to end against
// the embedded HSK lexicon.
func TestMandarinAuditor(t *testing.T) {
	a, err := newMandarinAuditor()
	if err != nil {
		t.Fatal(err)
	}

	// 咖啡 (coffee) is HSK 3 in the new-HSK dataset, so this sentence is within
	// level 3 but not level 2.
	if r := a.Audit("我喜欢喝咖啡", 3); !r.Passed {
		t.Errorf("expected pass at HSK 3, got %+v", r)
	}
	if r := a.Audit("我喜欢喝咖啡", 2); r.Passed {
		t.Errorf("expected fail at HSK 2, got %+v", r)
	}

	// 气氛 (atmosphere, HSK 6) is above an HSK 2 target.
	r := a.Audit("这家咖啡馆的气氛很好。", 2)
	if r.Passed {
		t.Fatalf("expected fail, got %+v", r)
	}
	if r.OutOfBounds["气氛"] == 0 {
		t.Errorf("expected 气氛 flagged above level, got %+v", r)
	}
}
