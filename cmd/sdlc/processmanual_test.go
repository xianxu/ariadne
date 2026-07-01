package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRunProcessManual_IncludeMemoryWithOutRefused pins the safety-critical guard
// (#153 M2 boundary-review Important): --include-memory writes private, machine-
// local memory content, so combining it with --out (which tends to get committed)
// must be refused and write NO file. Regression net for the memory-privacy footgun
// the issue Log records was once hit manually.
func TestRunProcessManual_IncludeMemoryWithOutRefused(t *testing.T) {
	out := filepath.Join(t.TempDir(), "manual.md")
	var stdout, stderr bytes.Buffer
	err := runProcessManual(&stdout, &stderr, out, false /*full*/, true /*includeMemory*/)
	if err == nil {
		t.Fatal("expected error: --include-memory with --out must be refused")
	}
	if _, statErr := os.Stat(out); !os.IsNotExist(statErr) {
		t.Errorf("no file must be written on refusal, but %s exists", out)
	}
}

// TestProcessManualCmd_WiredInRoot exercises the full cobra tree against the real
// repo: `sdlc process-manual` must render the manual spanning judge prompts + skills.
func TestProcessManualCmd_WiredInRoot(t *testing.T) {
	root := buildRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"process-manual"})
	if err := root.Execute(); err != nil {
		t.Fatalf("`sdlc process-manual` failed: %v", err)
	}
	for _, want := range []string{"# Process manual", "milestone-review", "## Skills"} {
		if !strings.Contains(buf.String(), want) {
			t.Errorf("`sdlc process-manual` output missing %q:\n%s", want, buf.String())
		}
	}
}
