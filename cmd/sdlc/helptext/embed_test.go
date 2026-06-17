package helptext

import (
	"strings"
	"testing"
)

func TestRootEmbedded(t *testing.T) {
	s, ok := Get("root")
	if !ok {
		t.Fatal("root.md not found in embed FS")
	}
	// root.md folds in the start-of-work runbook: `sdlc --help` is the
	// single workflow contract (the old `--agents.md` flag was merged
	// in), so these load-bearing anchors must survive edits.
	for _, want := range []string{"BEFORE WORK", "WHEN A VERB ERRORS", "checkpoint guard"} {
		if !strings.Contains(s, want) {
			t.Errorf("root.md missing %q", want)
		}
	}
	if !strings.HasSuffix(s, "\n") {
		t.Errorf("root.md content should end with a newline")
	}
}

func TestActiveTimeEmbedded(t *testing.T) {
	s, ok := Get("active-time")
	if !ok {
		t.Fatal("active-time.md not found in embed FS")
	}
	// Load-bearing anchors: the exit-code contract + the actual cross-link.
	for _, want := range []string{"EXIT CODES", "TELEMETRY UNAVAILABLE", "sdlc actual"} {
		if !strings.Contains(s, want) {
			t.Errorf("active-time.md missing %q", want)
		}
	}
}

func TestCloseEmbedded(t *testing.T) {
	s, ok := Get("close")
	if !ok {
		t.Fatal("close.md not found in embed FS")
	}
	if !strings.Contains(s, "--actual") || !strings.Contains(s, "--verified") {
		t.Errorf("close.md missing required-flag documentation")
	}
}

func TestPushEmbedded(t *testing.T) {
	s, ok := Get("push")
	if !ok {
		t.Fatal("push.md not found in embed FS")
	}
	// Regression guard for #54: push.md must frame the in-place branch
	// flow (change-code → pr → merge) as the default close path, not
	// present `sdlc push` as a co-equal "lighter path".
	if !strings.Contains(s, "change-code") {
		t.Errorf("push.md missing reference to the default change-code (in-place branch) flow")
	}
}

func TestMustGetPanicsOnMissing(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustGet on missing entry should panic")
		}
	}()
	MustGet("definitely-not-a-real-entry")
}
