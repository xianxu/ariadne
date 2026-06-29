package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/estimate"
)

func TestEstimateSourceCmd_Registered(t *testing.T) {
	cmd := NewEstimateSourceCmd()
	for _, flag := range []string{"model", "brain-dir"} {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("estimate-source missing flag: --%s", flag)
		}
	}
	if got := cmd.Flags().Lookup("model").DefValue; got != estimate.CurrentModel() {
		t.Errorf("estimate-source default model = %q, want %q", got, estimate.CurrentModel())
	}
}

// estimateSourceStatus over a temp brain: present doc → Exists; a ledger newer
// than the doc → Stale; absent doc → not Exists.
func TestEstimateSourceStatus(t *testing.T) {
	brain := t.TempDir()
	velocity := filepath.Join(brain, "data", "life", "42shots", "velocity")
	if err := os.MkdirAll(velocity, 0o755); err != nil {
		t.Fatal(err)
	}
	doc := filepath.Join(velocity, "estimate-logic-v2.md")
	ledger := filepath.Join(velocity, "calibration-ledger.tsv")

	// Missing doc → not Exists.
	if st := estimateSourceStatus(brain, "estimate-logic-v2", ""); st.Exists {
		t.Fatalf("missing doc should not Exist, got %+v", st)
	}

	// Present doc, no ledger → Exists, not Stale.
	write(t, doc, "# estimate-logic-v2\n")
	if st := estimateSourceStatus(brain, "estimate-logic-v2", ""); !st.Exists || st.Stale {
		t.Fatalf("present doc should Exist + not be Stale, got %+v", st)
	}

	// Ledger newer than the doc → Stale.
	write(t, ledger, "row\n")
	older := time.Now().Add(-2 * time.Hour)
	newer := time.Now().Add(-1 * time.Hour)
	if err := os.Chtimes(doc, older, older); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(ledger, newer, newer); err != nil {
		t.Fatal(err)
	}
	if st := estimateSourceStatus(brain, "estimate-logic-v2", ""); !st.Exists || !st.Stale {
		t.Fatalf("ledger-newer-than-doc should be Stale, got %+v", st)
	}

	// An override path that exists is Exists + never Stale (no implied ledger).
	override := filepath.Join(t.TempDir(), "custom.md")
	write(t, override, "x\n")
	if st := estimateSourceStatus(brain, "estimate-logic-v2", override); !st.Exists || st.Stale {
		t.Fatalf("existing override should Exist + not be Stale, got %+v", st)
	}
	if st := estimateSourceStatus(brain, "estimate-logic-v2", override); st.Path != override {
		t.Fatalf("override should set Path, got %q", st.Path)
	}
}

// reportEstimateSource prints the guidance to stdout AND returns a non-nil error
// (→ non-zero exit) only when the source is missing, so the pull fails loud.
func TestReportEstimateSourceFailsLoudWhenMissing(t *testing.T) {
	var out strings.Builder
	var errw discard
	if err := reportEstimateSource(&out, &errw, estimate.SourceStatus{Path: "/p", Exists: false}); err == nil {
		t.Error("missing source should return an error (non-zero exit)")
	}
	if !strings.Contains(out.String(), "/p") {
		t.Errorf("should print the guidance (with the path) even on the missing path:\n%s", out.String())
	}

	out.Reset()
	if err := reportEstimateSource(&out, &errw, estimate.SourceStatus{Path: "/p", Exists: true}); err != nil {
		t.Errorf("present source should not error, got %v", err)
	}
	if !strings.Contains(out.String(), "ESTIMATE DERIVATION") {
		t.Errorf("should print the full guidance block:\n%s", out.String())
	}
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// discard is a throwaway io.Writer for tests that only check the return value.
type discard struct{}

func (*discard) Write(p []byte) (int, error) { return len(p), nil }
