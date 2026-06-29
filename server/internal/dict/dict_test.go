package dict

import "testing"

func TestLoadAndLookup(t *testing.T) {
	d, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	// 咖啡 (coffee) is a common HSK word; verify pinyin tone marks and a gloss.
	e, ok := d.Lookup("咖啡")
	if !ok {
		t.Fatal("expected 咖啡 in dictionary")
	}
	if e.Pinyin != "kā fēi" {
		t.Errorf("pinyin = %q, want %q", e.Pinyin, "kā fēi")
	}
	if len(e.Definitions) == 0 {
		t.Error("expected at least one definition")
	}
}

func TestLookupMissing(t *testing.T) {
	d, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if _, ok := d.Lookup("zzz-not-a-word"); ok {
		t.Error("expected miss for non-word")
	}
}

func TestBreakdown(t *testing.T) {
	d, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	parts := d.Breakdown("咖啡")
	if len(parts) != 2 {
		t.Fatalf("breakdown of 咖啡 = %d parts, want 2", len(parts))
	}
	if parts[0].Word != "咖" || parts[1].Word != "啡" {
		t.Errorf("breakdown chars = %q,%q, want 咖,啡", parts[0].Word, parts[1].Word)
	}

	// Polyphonic chars must resolve to the everyday reading, not an archaic one
	// (the generator keeps the reading with the most senses).
	listen := d.Breakdown("听说")
	if len(listen) != 2 {
		t.Fatalf("breakdown of 听说 = %d parts, want 2", len(listen))
	}
	if listen[0].Pinyin != "tīng" {
		t.Errorf("听 pinyin = %q, want tīng", listen[0].Pinyin)
	}
	if listen[1].Pinyin != "shuō" {
		t.Errorf("说 pinyin = %q, want shuō", listen[1].Pinyin)
	}

	// A single character has nothing to break down.
	if got := d.Breakdown("好"); got != nil {
		t.Errorf("breakdown of single char = %v, want nil", got)
	}
}
