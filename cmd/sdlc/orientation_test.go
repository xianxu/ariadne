package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestBoundaryOrientation pins the #137 contract: orientation is derived from the
// live git repo (a fixture named "pair"), so the issue ref is repo-correct
// (pair#72) and NEVER falls back to a hardcoded ariadne#N for a non-ariadne root.
func TestBoundaryOrientation(t *testing.T) {
	parent := t.TempDir()
	repo := filepath.Join(parent, "pair") // git-root basename controls the repo name
	if err := os.MkdirAll(filepath.Join(repo, "workshop", "issues"), 0o755); err != nil {
		t.Fatal(err)
	}
	gitInit(t, repo, "https://example.com/pair.git") // remote unused; RepoTopLevel just needs the git dir
	if err := os.WriteFile(filepath.Join(repo, "workshop", "issues", "000072-x.md"),
		[]byte("---\nstatus: working\n---\n# x\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	if err := os.Chdir(repo); err != nil {
		t.Fatal(err)
	}

	// Downstream repo (no construct/base.manifest), milestone boundary.
	o := boundaryOrientation("workshop/issues", 72, "M1")
	if o.Repo != "pair" {
		t.Errorf("Repo = %q, want pair", o.Repo)
	}
	if o.IssueRef != "pair#72 M1" {
		t.Errorf("IssueRef = %q, want pair#72 M1", o.IssueRef)
	}
	if strings.HasPrefix(o.IssueRef, "ariadne#") {
		t.Errorf("must NOT fall back to ariadne#N for a non-ariadne repo: %q", o.IssueRef)
	}
	if !strings.Contains(o.Boundary, "milestone M1") {
		t.Errorf("Boundary = %q, want milestone M1", o.Boundary)
	}
	if !strings.HasSuffix(o.IssueFile, "000072-x.md") {
		t.Errorf("IssueFile = %q, want …/000072-x.md", o.IssueFile)
	}
	if !strings.Contains(o.RepoNote, "downstream") {
		t.Errorf("RepoNote should mark a downstream repo: %q", o.RepoNote)
	}

	// Add construct/base.manifest → base repo, whole-issue boundary.
	if err := os.MkdirAll(filepath.Join(repo, "construct"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "construct", "base.manifest"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	o2 := boundaryOrientation("workshop/issues", 72, "")
	if o2.IssueRef != "pair#72" {
		t.Errorf("whole-issue IssueRef = %q, want pair#72", o2.IssueRef)
	}
	if !strings.Contains(o2.Boundary, "whole-issue") {
		t.Errorf("Boundary = %q, want whole-issue close", o2.Boundary)
	}
	if strings.Contains(o2.RepoNote, "downstream") {
		t.Errorf("a base repo's note must not say downstream: %q", o2.RepoNote)
	}
}
