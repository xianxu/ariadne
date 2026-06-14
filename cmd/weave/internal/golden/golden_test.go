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
	// The still-deferred verbs (seed/merge) produce NO weave action today, but
	// setup.sh DID produce their output. They are pre-registered EXPECTED-missing:
	// present in live, weave defers. The classifier ledgers them when the
	// deferred-verb intent is supplied with its target present in live. (tool is
	// no longer deferred — it lowers to a ToolDep, classified below.)
	in := Input{
		RepoRoot: "/ws/nous",
		Deferred: []intent.Intent{
			{Kind: intent.Seed, Source: "bootstrap.sh", Target: "bootstrap.sh"},
			{Kind: intent.Merge, Source: ".claude/settings.ariadne.json", Target: ".claude/settings.json"},
		},
		Observed: map[string]Observed{
			"/ws/nous/bootstrap.sh":          {Exists: true},
			"/ws/nous/.claude/settings.json": {Exists: true},
		},
	}
	divs := Classify(in)
	if len(divs) != 2 {
		t.Fatalf("got %d divergences, want 2 (one per deferred intent)", len(divs))
	}
	for _, d := range divs {
		if d.Class != Expected {
			t.Fatalf("deferred verb %s @ %s: class = %v, want EXPECTED", d.Verb, d.Path, d.Class)
		}
	}
}

func TestToolDerivativeMatch(t *testing.T) {
	// Derivative (Owner != RepoRoot): weave appends `substrate ../ariadne` to
	// construct/deps. MATCH when the live construct/deps already carries that
	// substrate row (setup.sh's cross-target branch already ran).
	in := Input{
		RepoRoot: "/ws/nous",
		Actions:  []plan.Action{plan.ToolDep{Owner: "/ws/ariadne", Path: "cmd/sdlc"}},
		Observed: map[string]Observed{
			"/ws/nous/construct/deps": {Exists: true, Content: "substrate ../ariadne\n"},
		},
	}
	divs := Classify(in)
	if len(divs) != 1 || divs[0].Class != Match {
		t.Fatalf("got %+v, want one MATCH", divs)
	}
	if divs[0].Verb != "tool" {
		t.Fatalf("verb = %q, want tool", divs[0].Verb)
	}
}

func TestToolDerivativeMissingRowUnexpected(t *testing.T) {
	// construct/deps exists but lacks the substrate row weave would add → UNEXPECTED.
	in := Input{
		RepoRoot: "/ws/nous",
		Actions:  []plan.Action{plan.ToolDep{Owner: "/ws/ariadne", Path: "cmd/sdlc"}},
		Observed: map[string]Observed{
			"/ws/nous/construct/deps": {Exists: true, Content: "data ../d git@x\n"},
		},
	}
	if divs := Classify(in); divs[0].Class != Unexpected {
		t.Fatalf("class = %v, want UNEXPECTED", divs[0].Class)
	}
}

func TestToolDerivativeAbsentDepsUnexpected(t *testing.T) {
	// No construct/deps at all → the substrate row is absent → UNEXPECTED.
	in := Input{
		RepoRoot: "/ws/nous",
		Actions:  []plan.Action{plan.ToolDep{Owner: "/ws/ariadne", Path: "cmd/sdlc"}},
		Observed: map[string]Observed{
			"/ws/nous/construct/deps": {Exists: false},
		},
	}
	if divs := Classify(in); divs[0].Class != Unexpected {
		t.Fatalf("class = %v, want UNEXPECTED", divs[0].Class)
	}
}

func TestToolOwnerMatch(t *testing.T) {
	// Owner self-walk (Owner == RepoRoot): weave runs `go mod edit -tool
	// <module>/cmd/sdlc`. MATCH when the live go.mod already has that tool
	// directive (setup.sh's self-walk already ran).
	in := Input{
		RepoRoot: "/ws/ariadne",
		Actions:  []plan.Action{plan.ToolDep{Owner: "/ws/ariadne", Path: "cmd/sdlc"}},
		Observed: map[string]Observed{
			"/ws/ariadne/go.mod": {Exists: true, Content: "module github.com/xianxu/ariadne\n\ngo 1.26\n\ntool github.com/xianxu/ariadne/cmd/sdlc\n"},
		},
	}
	divs := Classify(in)
	if len(divs) != 1 || divs[0].Class != Match {
		t.Fatalf("got %+v, want one MATCH", divs)
	}
}

func TestToolOwnerMissingDirectiveUnexpected(t *testing.T) {
	// go.mod exists but lacks the tool directive weave would add → UNEXPECTED.
	in := Input{
		RepoRoot: "/ws/ariadne",
		Actions:  []plan.Action{plan.ToolDep{Owner: "/ws/ariadne", Path: "cmd/sdlc"}},
		Observed: map[string]Observed{
			"/ws/ariadne/go.mod": {Exists: true, Content: "module github.com/xianxu/ariadne\n\ngo 1.26\n"},
		},
	}
	if divs := Classify(in); divs[0].Class != Unexpected {
		t.Fatalf("class = %v, want UNEXPECTED", divs[0].Class)
	}
}

func TestToolNotDeferred(t *testing.T) {
	// Tool is no longer in the deferred ledger (it lowers now).
	if IsDeferred(intent.Tool) {
		t.Fatalf("IsDeferred(Tool) = true, want false (tool lowers to ToolDep now)")
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
