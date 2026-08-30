package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
)

// TestRunArchPrinciples_RendersRegistry pins that the command DERIVES from the one
// registry (#128): its output carries every ARCH-* marker and the requested lens
// header — so it is a tested consumer of architecture.md, not a restatement.
func TestRunArchPrinciples_RendersRegistry(t *testing.T) {
	var buf bytes.Buffer
	if err := runArchPrinciples(&buf, "at-plan"); err != nil {
		t.Fatalf("runArchPrinciples(at-plan): %v", err)
	}
	out := buf.String()
	for _, want := range []string{"ARCH-DRY", "ARCH-PURE", "ARCH-PURPOSE", "ARCH-MOCK", "ARCH-CONSTRAINTS", "at-plan"} {
		if !strings.Contains(out, want) {
			t.Errorf("at-plan output missing %q:\n%s", want, out)
		}
	}
	if !strings.Contains(out, judge.ArchitectureRegistry) {
		t.Errorf("at-plan output must carry the complete architecture registry, not marker sentinels:\n%s", out)
	}
}

// TestRunArchPrinciples_LensSwitch confirms --lens at-review changes the
// foregrounded lens header (the registry body still carries both lenses).
func TestRunArchPrinciples_LensSwitch(t *testing.T) {
	var buf bytes.Buffer
	if err := runArchPrinciples(&buf, "at-review"); err != nil {
		t.Fatalf("runArchPrinciples(at-review): %v", err)
	}
	if !strings.Contains(buf.String(), "at-review") {
		t.Errorf("at-review output should foreground the at-review lens:\n%s", buf.String())
	}
}

// TestRunArchPrinciples_UnknownLens rejects a typo'd lens rather than silently
// rendering a wrong header.
func TestRunArchPrinciples_UnknownLens(t *testing.T) {
	var buf bytes.Buffer
	if err := runArchPrinciples(&buf, "at-bogus"); err == nil {
		t.Error("expected an error for an unknown --lens")
	}
	if buf.Len() != 0 {
		t.Errorf("no output should be written on an unknown lens, got:\n%s", buf.String())
	}
}

// TestArchPrinciplesCmd_WiredInRoot exercises the command through the real command
// tree — which calls helptext.MustGet("arch-principles") (panics if the helptext
// file is missing) and confirms registration + the default at-plan lens.
func TestArchPrinciplesCmd_WiredInRoot(t *testing.T) {
	root := buildRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"arch-principles"})
	if err := root.Execute(); err != nil {
		t.Fatalf("`sdlc arch-principles` failed: %v", err)
	}
	if !strings.Contains(buf.String(), "ARCH-PURPOSE") {
		t.Errorf("`sdlc arch-principles` should render the registry:\n%s", buf.String())
	}
}
