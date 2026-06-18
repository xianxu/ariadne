package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestFormatList_NameAndDescriptionPerLine: formatList renders one line per
// prototype, "<name>\t<description>" — the matching surface an agent reads to
// pick a type. Pure (no IO); order is the input order (mergeTypes already sorts
// by name).
func TestFormatList_NameAndDescriptionPerLine(t *testing.T) {
	protos := []TypeProto{
		{Name: "continuation", Description: "Use when handing off a session.", BodyPath: "/x/continuation.md"},
		{Name: "event", Description: "Use when capturing an event.", BodyPath: "/x/event.md"},
	}
	got := formatList(protos)
	want := "continuation\tUse when handing off a session.\nevent\tUse when capturing an event.\n"
	if got != want {
		t.Fatalf("formatList =\n%q\nwant\n%q", got, want)
	}
}

// TestFormatList_Empty: an empty set yields an empty string (no trailing junk).
func TestFormatList_Empty(t *testing.T) {
	if got := formatList(nil); got != "" {
		t.Fatalf("formatList(nil) = %q, want empty", got)
	}
}

// TestFormatList_MissingDescriptionTolerated: a prototype with no description
// still emits its name (trailing tab + empty desc), so the noun is discoverable.
func TestFormatList_MissingDescriptionTolerated(t *testing.T) {
	got := formatList([]TypeProto{{Name: "bare", Description: ""}})
	if !strings.HasPrefix(got, "bare\t") {
		t.Fatalf("formatList missing-desc = %q, want it to start with the name", got)
	}
}

// TestRunShow_KnownPrintsBody / unknown lists names + non-zero: runShow resolves
// the DAG-merged set for the cwd's repo and prints the leaf-winning body; an
// unknown name lists the available names to stderr and returns errUnknownType.
// We chdir into a fixture repo so resolveTypes (findRepoRoot + Walk + mergeTypes)
// finds it.
func TestRunShow_KnownAndUnknown(t *testing.T) {
	repo := t.TempDir()
	// A single-layer repo: construct/datatype with one prototype + a base.manifest
	// so layergraph.Walk treats it as a layer root, plus construct/ for findRepoRoot.
	writeProto(t, filepath.Join(repo, "construct", "datatype", "event.md"), "an event")
	if err := os.WriteFile(filepath.Join(repo, "construct", "base.manifest"), []byte("prose AGENTS.local.md\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	prev, _ := os.Getwd()
	if err := os.Chdir(repo); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(prev)

	// Known name → prints the body.
	var out, errOut bytes.Buffer
	if err := runShow("event", &out, &errOut); err != nil {
		t.Fatalf("runShow(event): %v", err)
	}
	if !strings.Contains(out.String(), "body of event.md") {
		t.Fatalf("runShow(event) body = %q, want it to contain the prototype body", out.String())
	}

	// Unknown name → lists available names to stderr + errUnknownType.
	out.Reset()
	errOut.Reset()
	err := runShow("nope", &out, &errOut)
	if err != errUnknownType {
		t.Fatalf("runShow(nope) err = %v, want errUnknownType", err)
	}
	if out.Len() != 0 {
		t.Fatalf("runShow(unknown) wrote to stdout: %q", out.String())
	}
	if !strings.Contains(errOut.String(), "event") {
		t.Fatalf("runShow(unknown) stderr = %q, want it to list available names (event)", errOut.String())
	}
}
