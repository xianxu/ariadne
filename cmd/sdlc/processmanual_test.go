package main

import (
	"bytes"
	"strings"
	"testing"
)

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
