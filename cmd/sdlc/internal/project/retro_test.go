package project

import "testing"

func TestLatestRetroDate(t *testing.T) {
	d, err := ParseDoc("---\ntype: project\n---\n## Log\n### 2026-07-01 — retro\nold\n### 2026-07-12 — retro\nnew\n")
	if err != nil {
		t.Fatal(err)
	}
	if got := LatestRetroDate(d); got != "2026-07-12" {
		t.Fatalf("LatestRetroDate = %q", got)
	}
}

func TestRetroStale(t *testing.T) {
	parse := func(body string) *Doc {
		d, err := ParseDoc(body)
		if err != nil {
			t.Fatal(err)
		}
		return d
	}
	legacy := parse("---\ntype: project\n---\n# legacy\n")
	if RetroStale(legacy, "2026-07-20", 7) {
		t.Fatal("legacy document without Log was nudged")
	}
	empty := parse("---\ntype: project\n---\n## Log\n")
	if !RetroStale(empty, "2026-07-20", 7) {
		t.Fatal("empty modeled Log was not stale")
	}
	fresh := parse("---\ntype: project\n---\n## Log\n### 2026-07-13 — retro\n")
	if RetroStale(fresh, "2026-07-20", 7) {
		t.Fatal("seven-day retro should be fresh")
	}
	old := parse("---\ntype: project\n---\n## Log\n### 2026-07-12 — retro\n")
	if !RetroStale(old, "2026-07-20", 7) {
		t.Fatal("eight-day retro should be stale")
	}
}
