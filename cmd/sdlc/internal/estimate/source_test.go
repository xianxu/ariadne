package estimate

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestVelocityPath(t *testing.T) {
	got := VelocityPath("/w/brain", "calibration-ledger.tsv")
	want := filepath.Join("/w/brain", "data", "life", "42shots", "velocity", "calibration-ledger.tsv")
	if got != want {
		t.Fatalf("VelocityPath = %q, want %q", got, want)
	}
}

func TestSourcePath(t *testing.T) {
	// Override wins outright.
	if got := SourcePath("/w/brain", "estimate-logic-v2", "/custom/method.md"); got != "/custom/method.md" {
		t.Errorf("override should win, got %q", got)
	}
	// Default is VelocityPath(brain, model+".md").
	got := SourcePath("/w/brain", "estimate-logic-v2", "")
	want := VelocityPath("/w/brain", "estimate-logic-v2.md")
	if got != want {
		t.Errorf("SourcePath default = %q, want %q", got, want)
	}
}

func TestSourceGuidanceAlwaysNamesBothSources(t *testing.T) {
	for _, s := range []SourceStatus{
		{Path: "/w/brain/.../estimate-logic-v2.md", Model: "estimate-logic-v2", Exists: true},
		{Path: "/w/brain/.../estimate-logic-v2.md", Model: "estimate-logic-v2", Exists: true, Stale: true},
		{Path: "/w/brain/.../estimate-logic-v2.md", Model: "estimate-logic-v2", Exists: false},
	} {
		out := SourceGuidance(s)
		if out == "" {
			t.Fatalf("SourceGuidance returned empty for %+v", s)
		}
		// Shared method (single-sourced in sdlc): the grammar pointer + the
		// recognized models must appear.
		if !strings.Contains(out, "change-code --help") {
			t.Errorf("missing shared-grammar pointer in:\n%s", out)
		}
		for _, m := range Models() {
			if !strings.Contains(out, m) {
				t.Errorf("missing model %q in:\n%s", m, out)
			}
		}
		// Repo-local calibration: the resolved path must appear.
		if !strings.Contains(out, s.Path) {
			t.Errorf("missing resolved path %q in:\n%s", s.Path, out)
		}
	}
}

func TestSourceGuidanceStatusVerbs(t *testing.T) {
	ok := SourceGuidance(SourceStatus{Path: "/p", Model: "estimate-logic-v2", Exists: true})
	if !strings.Contains(strings.ToLower(ok), "ok") {
		t.Errorf("Exists status should read ok:\n%s", ok)
	}

	stale := SourceGuidance(SourceStatus{Path: "/p", Model: "estimate-logic-v2", Exists: true, Stale: true})
	if !strings.Contains(strings.ToLower(stale), "stale") || !strings.Contains(stale, "#127") {
		t.Errorf("Stale status should flag staleness + #127:\n%s", stale)
	}

	missing := SourceGuidance(SourceStatus{Path: "/p", Model: "estimate-logic-v2", Exists: false})
	// Must be a loud next-action, not silence: name the env override + forbid memory.
	if !strings.Contains(missing, "MISSING") || !strings.Contains(missing, "WF_ESTIMATOR_SRC") {
		t.Errorf("Missing status should be a loud next-action:\n%s", missing)
	}
	if !strings.Contains(strings.ToLower(missing), "memory") {
		t.Errorf("Missing status should forbid estimating from memory:\n%s", missing)
	}
}
