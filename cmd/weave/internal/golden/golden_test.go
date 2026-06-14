package golden

import (
	"testing"

	"github.com/xianxu/ariadne/cmd/weave/internal/intent"
	"github.com/xianxu/ariadne/cmd/weave/internal/plan"
)

// The classifier is PURE: it takes weave's planned actions + the deferred-verb
// intents + a snapshot of the observed live FS state, and classifies every
// probe as MATCH / EXPECTED / UNEXPECTED. No real filesystem is touched here —
// Observed is synthetic. This is the purity proof (ARCH-PURE): the IO gatherer
// fills Observed from the live repos; the classification logic is testable
// mock-free.

func TestSymlinkMatch(t *testing.T) {
	// weave would create a relative symlink dst -> rel(dir(dst), src). The live
	// tree already has that exact relative link → MATCH.
	in := Input{
		RepoRoot: "/ws/nous",
		Actions: []plan.Action{
			plan.Symlink{Src: "/ws/ariadne/Makefile", Dst: "Makefile"},
		},
		Observed: map[string]Observed{
			"/ws/nous/Makefile": {Exists: true, IsSymlink: true, LinkTarget: "../ariadne/Makefile"},
		},
	}
	divs := Classify(in)
	if len(divs) != 1 {
		t.Fatalf("got %d divergences, want 1", len(divs))
	}
	if divs[0].Class != Match {
		t.Fatalf("class = %v, want MATCH (detail=%s)", divs[0].Class, divs[0].Detail)
	}
}

func TestSymlinkPointsElsewhereUnexpected(t *testing.T) {
	// The live symlink exists but points somewhere weave would NOT link → UNEXPECTED.
	in := Input{
		RepoRoot: "/ws/nous",
		Actions: []plan.Action{
			plan.Symlink{Src: "/ws/ariadne/Makefile", Dst: "Makefile"},
		},
		Observed: map[string]Observed{
			"/ws/nous/Makefile": {Exists: true, IsSymlink: true, LinkTarget: "../other/Makefile"},
		},
	}
	divs := Classify(in)
	if divs[0].Class != Unexpected {
		t.Fatalf("class = %v, want UNEXPECTED", divs[0].Class)
	}
}

func TestSymlinkMissingUnexpected(t *testing.T) {
	// weave would link, but nothing is there in live → UNEXPECTED (setup.sh
	// should have produced it; its absence is a real divergence).
	in := Input{
		RepoRoot: "/ws/nous",
		Actions: []plan.Action{
			plan.Symlink{Src: "/ws/ariadne/Makefile", Dst: "Makefile"},
		},
		Observed: map[string]Observed{
			"/ws/nous/Makefile": {Exists: false},
		},
	}
	divs := Classify(in)
	if divs[0].Class != Unexpected {
		t.Fatalf("class = %v, want UNEXPECTED", divs[0].Class)
	}
}

func TestSymlinkOccupiedByRegularFileUnexpected(t *testing.T) {
	// A regular file occupies the symlink slot (not a symlink) → UNEXPECTED.
	in := Input{
		RepoRoot: "/ws/nous",
		Actions: []plan.Action{
			plan.Symlink{Src: "/ws/ariadne/Makefile", Dst: "Makefile"},
		},
		Observed: map[string]Observed{
			"/ws/nous/Makefile": {Exists: true, IsSymlink: false},
		},
	}
	divs := Classify(in)
	if divs[0].Class != Unexpected {
		t.Fatalf("class = %v, want UNEXPECTED", divs[0].Class)
	}
}

func TestMkdirMatch(t *testing.T) {
	in := Input{
		RepoRoot: "/ws/nous",
		Actions: []plan.Action{
			plan.Mkdir{Path: ".claude/skills"},
		},
		Observed: map[string]Observed{
			"/ws/nous/.claude/skills": {Exists: true, IsDir: true},
		},
	}
	divs := Classify(in)
	if divs[0].Class != Match {
		t.Fatalf("class = %v, want MATCH", divs[0].Class)
	}
}

func TestMkdirMissingUnexpected(t *testing.T) {
	in := Input{
		RepoRoot: "/ws/nous",
		Actions: []plan.Action{
			plan.Mkdir{Path: ".claude/skills"},
		},
		Observed: map[string]Observed{
			"/ws/nous/.claude/skills": {Exists: false},
		},
	}
	divs := Classify(in)
	if divs[0].Class != Unexpected {
		t.Fatalf("class = %v, want UNEXPECTED", divs[0].Class)
	}
}

func TestWriteFileMatchAndDrift(t *testing.T) {
	// A WriteFile (e.g. a future composed AGENTS.md) whose live content matches
	// → MATCH; whose content drifts → UNEXPECTED.
	match := Input{
		RepoRoot: "/ws/nous",
		Actions:  []plan.Action{plan.WriteFile{Path: "AGENTS.md", Content: "BODY"}},
		Observed: map[string]Observed{
			"/ws/nous/AGENTS.md": {Exists: true, Content: "BODY"},
		},
	}
	if divs := Classify(match); divs[0].Class != Match {
		t.Fatalf("matching content: class = %v, want MATCH", divs[0].Class)
	}
	drift := Input{
		RepoRoot: "/ws/nous",
		Actions:  []plan.Action{plan.WriteFile{Path: "AGENTS.md", Content: "BODY"}},
		Observed: map[string]Observed{
			"/ws/nous/AGENTS.md": {Exists: true, Content: "OTHER"},
		},
	}
	if divs := Classify(drift); divs[0].Class != Unexpected {
		t.Fatalf("drifting content: class = %v, want UNEXPECTED", divs[0].Class)
	}
}

func TestTouchMatchOnExistence(t *testing.T) {
	// A Touch (create-if-missing) MATCHES whenever the target merely exists —
	// regardless of content (the file accumulates over time). Absent → UNEXPECTED.
	present := Input{
		RepoRoot: "/ws/nous",
		Actions:  []plan.Action{plan.Touch{Path: "workshop/lessons.md"}},
		Observed: map[string]Observed{
			"/ws/nous/workshop/lessons.md": {Exists: true, Content: "lots of accumulated lessons"},
		},
	}
	if divs := Classify(present); divs[0].Class != Match {
		t.Fatalf("existing touch target: class = %v, want MATCH (detail=%s)", divs[0].Class, divs[0].Detail)
	}
	absent := Input{
		RepoRoot: "/ws/nous",
		Actions:  []plan.Action{plan.Touch{Path: "workshop/lessons.md"}},
		Observed: map[string]Observed{
			"/ws/nous/workshop/lessons.md": {Exists: false},
		},
	}
	if divs := Classify(absent); divs[0].Class != Unexpected {
		t.Fatalf("absent touch target: class = %v, want UNEXPECTED", divs[0].Class)
	}
}

func TestDeferredVerbsExpected(t *testing.T) {
	// The deferred verbs (seed/merge/tool) produce NO weave action today, but
	// setup.sh DID produce their output. They are pre-registered EXPECTED-missing:
	// present in live, weave defers. The classifier ledgers them when the
	// deferred-verb intent is supplied with its target present in live.
	in := Input{
		RepoRoot: "/ws/nous",
		Deferred: []intent.Intent{
			{Kind: intent.Seed, Source: "bootstrap.sh", Target: "bootstrap.sh"},
			{Kind: intent.Merge, Source: ".claude/settings.ariadne.json", Target: ".claude/settings.json"},
			{Kind: intent.Tool, Source: "cmd/sdlc", Target: "cmd/sdlc"},
		},
		Observed: map[string]Observed{
			"/ws/nous/bootstrap.sh":          {Exists: true},
			"/ws/nous/.claude/settings.json": {Exists: true},
		},
	}
	divs := Classify(in)
	if len(divs) != 3 {
		t.Fatalf("got %d divergences, want 3 (one per deferred intent)", len(divs))
	}
	for _, d := range divs {
		if d.Class != Expected {
			t.Fatalf("deferred verb %s @ %s: class = %v, want EXPECTED", d.Verb, d.Path, d.Class)
		}
	}
}

func TestHasUnexpected(t *testing.T) {
	clean := []Divergence{{Class: Match}, {Class: Expected}}
	if HasUnexpected(clean) {
		t.Fatalf("HasUnexpected(clean) = true, want false")
	}
	dirty := []Divergence{{Class: Match}, {Class: Unexpected}}
	if !HasUnexpected(dirty) {
		t.Fatalf("HasUnexpected(dirty) = false, want true")
	}
}
