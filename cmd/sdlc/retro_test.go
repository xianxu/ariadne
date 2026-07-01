package main

import (
	"bytes"
	"strings"
	"testing"
)

// TestRetroCmd_WiredInRoot exercises the full cobra tree against the real repo:
// `sdlc retro` must render the manual spanning judge prompts + skills.
func TestRetroCmd_WiredInRoot(t *testing.T) {
	root := buildRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"retro"})
	if err := root.Execute(); err != nil {
		t.Fatalf("`sdlc retro` failed: %v", err)
	}
	for _, want := range []string{"# Process manual", "milestone-review", "## Skills"} {
		if !strings.Contains(buf.String(), want) {
			t.Errorf("`sdlc retro` output missing %q:\n%s", want, buf.String())
		}
	}
}
