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

func TestSeedNotDeferred(t *testing.T) {
	// Seed is no longer in the deferred ledger as of M5 — it lowers to a
	// plan.Seed action, classified by classifyAction (content-tracking copy).
	// Nothing is deferred now.
	if IsDeferred(intent.Seed) {
		t.Fatalf("IsDeferred(Seed) = true, want false (seed lowers to a Seed action now)")
	}
}

func TestSeedContentMatch(t *testing.T) {
	// A Seed MATCHES iff the live target exists AND its content equals the
	// upstream SOURCE's content (what applySeed would write). The probe is two
	// observed files: the target (Dst, root-relative) and the source (Src, an
	// absolute upstream path keyed by its own abs path).
	in := Input{
		RepoRoot: "/ws/nous",
		Actions: []plan.Action{
			plan.Seed{Src: "/ws/ariadne/bootstrap.sh", Dst: "bootstrap.sh"},
		},
		Observed: map[string]Observed{
			"/ws/nous/bootstrap.sh":    {Exists: true, Content: "#!/bin/sh\nboot\n"},
			"/ws/ariadne/bootstrap.sh": {Exists: true, Content: "#!/bin/sh\nboot\n"},
		},
	}
	divs := Classify(in)
	if len(divs) != 1 || divs[0].Class != Match || divs[0].Verb != "seed" {
		t.Fatalf("seed content match: got %+v, want one MATCH seed", divs)
	}
}

func TestSeedContentDriftUnexpected(t *testing.T) {
	// Live target present but its content drifted from the upstream source →
	// UNEXPECTED (weave would refresh it to the source bytes; live differs).
	in := Input{
		RepoRoot: "/ws/nous",
		Actions: []plan.Action{
			plan.Seed{Src: "/ws/ariadne/bootstrap.sh", Dst: "bootstrap.sh"},
		},
		Observed: map[string]Observed{
			"/ws/nous/bootstrap.sh":    {Exists: true, Content: "STALE v1\n"},
			"/ws/ariadne/bootstrap.sh": {Exists: true, Content: "FRESH v2\n"},
		},
	}
	divs := Classify(in)
	if len(divs) != 1 || divs[0].Class != Unexpected {
		t.Fatalf("seed drift: got %+v, want one UNEXPECTED", divs)
	}
}

func TestSeedTargetAbsentUnexpected(t *testing.T) {
	// Source present, target absent → UNEXPECTED (weave would seed it, setup.sh's
	// output isn't there).
	in := Input{
		RepoRoot: "/ws/nous",
		Actions: []plan.Action{
			plan.Seed{Src: "/ws/ariadne/bootstrap.sh", Dst: "bootstrap.sh"},
		},
		Observed: map[string]Observed{
			"/ws/ariadne/bootstrap.sh": {Exists: true, Content: "boot\n"},
			// target absent
		},
	}
	divs := Classify(in)
	if len(divs) != 1 || divs[0].Class != Unexpected {
		t.Fatalf("seed target absent: got %+v, want one UNEXPECTED", divs)
	}
}

func TestSeedSourceAbsentSkips(t *testing.T) {
	// Upstream source absent → weave's applySeed would do nothing (non-fatal
	// skip), so a present-or-absent target is not a divergence — MATCH (we can't
	// fault a target weave wouldn't touch).
	in := Input{
		RepoRoot: "/ws/nous",
		Actions: []plan.Action{
			plan.Seed{Src: "/ws/ariadne/bootstrap.sh", Dst: "bootstrap.sh"},
		},
		Observed: map[string]Observed{
			// neither source nor target present
		},
	}
	divs := Classify(in)
	if len(divs) != 1 || divs[0].Class != Match {
		t.Fatalf("seed source absent: got %+v, want one MATCH (skip)", divs)
	}
}

func TestMergeNotDeferred(t *testing.T) {
	// Merge is no longer in the deferred ledger (it lowers to a MergeSettings now).
	if IsDeferred(intent.Merge) {
		t.Fatalf("IsDeferred(Merge) = true, want false (merge lowers to MergeSettings now)")
	}
}

func TestMergeSettingsSemanticMatch(t *testing.T) {
	// A MergeSettings MATCHES iff the live settings.json (Target) SEMANTICALLY
	// equals weave's merge output (parse both JSON + deep-equal — NOT a byte
	// compare, since merge-settings.sh's jq/python key ordering need not match
	// weave's). Here the local is absent ⇒ weave's output is base-with-meta-
	// stripped; the live target equals that semantically (different key order),
	// so MATCH.
	base := `{"$comment":"x","$merge_keys":["permissions.allow"],"permissions":{"allow":["A","B"]}}`
	// Same content, but keys serialized in a different order than weave emits.
	liveTarget := `{
		"permissions": {"allow": ["A", "B"]}
	}`
	in := Input{
		RepoRoot: "/ws/ariadne",
		Actions: []plan.Action{
			plan.MergeSettings{Source: ".claude/settings.ariadne.json", Target: ".claude/settings.json"},
		},
		Observed: map[string]Observed{
			"/ws/ariadne/.claude/settings.ariadne.json": {Exists: true, Content: base},
			"/ws/ariadne/.claude/settings.local.json":   {Exists: false},
			"/ws/ariadne/.claude/settings.json":         {Exists: true, Content: liveTarget},
		},
	}
	divs := Classify(in)
	if len(divs) != 1 {
		t.Fatalf("got %d divergences, want 1", len(divs))
	}
	if divs[0].Class != Match {
		t.Fatalf("class = %v, want MATCH (detail=%s)", divs[0].Class, divs[0].Detail)
	}
	if divs[0].Verb != "merge" {
		t.Fatalf("verb = %q, want merge", divs[0].Verb)
	}
}

func TestMergeSettingsWithLocalMatch(t *testing.T) {
	// Local present: weave's output is the deep-merge; MATCH when live equals it.
	base := `{"$merge_keys":["permissions.allow"],"permissions":{"allow":["A","B"]}}`
	local := `{"$remove":{"permissions.allow":["A"]},"permissions":{"allow":["C"]}}`
	liveTarget := `{"permissions":{"allow":["B","C"]}}`
	in := Input{
		RepoRoot: "/ws/ariadne",
		Actions: []plan.Action{
			plan.MergeSettings{Source: ".claude/settings.ariadne.json", Target: ".claude/settings.json"},
		},
		Observed: map[string]Observed{
			"/ws/ariadne/.claude/settings.ariadne.json": {Exists: true, Content: base},
			"/ws/ariadne/.claude/settings.local.json":   {Exists: true, Content: local},
			"/ws/ariadne/.claude/settings.json":         {Exists: true, Content: liveTarget},
		},
	}
	divs := Classify(in)
	if len(divs) != 1 || divs[0].Class != Match {
		t.Fatalf("got %+v, want one MATCH", divs)
	}
}

func TestMergeSettingsContentDriftUnexpected(t *testing.T) {
	// Live settings.json is NOT semantically equal to weave's merge output → UNEXPECTED.
	base := `{"$merge_keys":["permissions.allow"],"permissions":{"allow":["A","B"]}}`
	liveTarget := `{"permissions":{"allow":["A","WRONG"]}}`
	in := Input{
		RepoRoot: "/ws/ariadne",
		Actions: []plan.Action{
			plan.MergeSettings{Source: ".claude/settings.ariadne.json", Target: ".claude/settings.json"},
		},
		Observed: map[string]Observed{
			"/ws/ariadne/.claude/settings.ariadne.json": {Exists: true, Content: base},
			"/ws/ariadne/.claude/settings.local.json":   {Exists: false},
			"/ws/ariadne/.claude/settings.json":         {Exists: true, Content: liveTarget},
		},
	}
	if divs := Classify(in); divs[0].Class != Unexpected {
		t.Fatalf("class = %v, want UNEXPECTED", divs[0].Class)
	}
}

func TestMergeSettingsTargetAbsentUnexpected(t *testing.T) {
	// weave would write settings.json but live has none → UNEXPECTED.
	base := `{"permissions":{"allow":["A"]}}`
	in := Input{
		RepoRoot: "/ws/ariadne",
		Actions: []plan.Action{
			plan.MergeSettings{Source: ".claude/settings.ariadne.json", Target: ".claude/settings.json"},
		},
		Observed: map[string]Observed{
			"/ws/ariadne/.claude/settings.ariadne.json": {Exists: true, Content: base},
			"/ws/ariadne/.claude/settings.local.json":   {Exists: false},
			"/ws/ariadne/.claude/settings.json":         {Exists: false},
		},
	}
	if divs := Classify(in); divs[0].Class != Unexpected {
		t.Fatalf("class = %v, want UNEXPECTED", divs[0].Class)
	}
}

func TestMergeSettingsBaseAbsentUnexpected(t *testing.T) {
	// The base settings.ariadne.json is absent (a port/setup error) → UNEXPECTED,
	// surfaced rather than silently passed.
	in := Input{
		RepoRoot: "/ws/ariadne",
		Actions: []plan.Action{
			plan.MergeSettings{Source: ".claude/settings.ariadne.json", Target: ".claude/settings.json"},
		},
		Observed: map[string]Observed{
			"/ws/ariadne/.claude/settings.ariadne.json": {Exists: false},
			"/ws/ariadne/.claude/settings.json":         {Exists: true, Content: `{}`},
		},
	}
	if divs := Classify(in); divs[0].Class != Unexpected {
		t.Fatalf("class = %v, want UNEXPECTED", divs[0].Class)
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
