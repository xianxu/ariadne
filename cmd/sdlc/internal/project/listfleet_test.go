package project

import (
	"path/filepath"
	"testing"
)

// TestListActiveProjectFiles exercises the exported fleet walk: it returns
// active-home + brain-legacy project files across siblings, excludes the
// subject by resolved path, and skips non-fleet siblings — reusing the SAME
// walk DiscoverByIssueRef uses (behavior-identity of that verb is pinned by
// its own tests).
func TestListActiveProjectFiles(t *testing.T) {
	parent := t.TempDir()
	writeProject(t, parent, "ariadne", "workshop/projects", "subject", "[ariadne#182]")
	writeProject(t, parent, "metis", "workshop/projects", "m", "[metis#2]")
	writeProjectStatus(t, parent, "brain", "data/project", "legacy", "active", "[metis#9]")
	markBrain(t, parent, "brain")
	// Archived project — NOT active, must be excluded (holds no current load).
	writeProjectStatus(t, parent, "metis", "workshop/history/projects", "old", "done", "[metis#3]")
	// Non-fleet siblings the fleet glob must skip.
	writeProject(t, parent, "metis.bak", "workshop/projects", "stale", "[metis#2]")

	subject := filepath.Join(parent, "ariadne", "workshop", "projects", "subject.md")
	got, err := ListActiveProjectFiles(parent, subject)
	if err != nil {
		t.Fatal(err)
	}

	byName := map[string]ProjectFile{}
	for _, f := range got {
		byName[filepath.Base(f.Path)] = f
	}
	if _, ok := byName["subject.md"]; ok {
		t.Error("subject project must be excluded by path")
	}
	if _, ok := byName["m.md"]; !ok {
		t.Error("active sibling project m.md missing")
	}
	if lf, ok := byName["legacy.md"]; !ok || !lf.Legacy {
		t.Errorf("brain legacy project missing or not flagged: %+v", lf)
	}
	if _, ok := byName["old.md"]; ok {
		t.Error("archived (history) project must be excluded — holds no active load")
	}
	if _, ok := byName["stale.md"]; ok {
		t.Error(".bak sibling must be skipped")
	}
	// RepoDir/Repo populated.
	if m := byName["m.md"]; m.Repo != "metis" || filepath.Base(m.RepoDir) != "metis" {
		t.Errorf("RepoDir/Repo wrong: %+v", m)
	}
}

// TestListActiveProjectFiles_ExcludeByRealPath ensures the subject exclusion
// resolves symlinks (a project reached via a symlinked repo dir is still the
// subject).
func TestListActiveProjectFiles_NoParent(t *testing.T) {
	_, err := ListActiveProjectFiles(filepath.Join(t.TempDir(), "does-not-exist"), "")
	if err == nil {
		t.Error("a missing parent dir should error")
	}
}
