package project

import (
	"os"
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

func TestListActiveProjectFiles_NoParent(t *testing.T) {
	_, err := ListActiveProjectFiles(filepath.Join(t.TempDir(), "does-not-exist"), "")
	if err == nil {
		t.Error("a missing parent dir should error")
	}
}

// TestListActiveProjectFiles_ExcludeResolvesSymlink pins the EvalSymlinks-based
// subject exclusion: a subject reached through a symlinked path is still
// recognized as the subject and excluded.
func TestListActiveProjectFiles_ExcludeResolvesSymlink(t *testing.T) {
	parent := t.TempDir()
	writeProject(t, parent, "ariadne", "workshop/projects", "subject", "[ariadne#182]")
	writeProject(t, parent, "metis", "workshop/projects", "m", "[metis#2]")
	real := filepath.Join(parent, "ariadne", "workshop", "projects", "subject.md")

	// A symlink pointing at the real subject file; excluding via the symlink
	// path must still drop the real file.
	link := filepath.Join(parent, "subject-link.md")
	if err := os.Symlink(real, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	got, err := ListActiveProjectFiles(parent, link)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range got {
		if filepath.Base(f.Path) == "subject.md" {
			t.Errorf("subject reached via symlink should be excluded: %+v", f)
		}
	}
}
