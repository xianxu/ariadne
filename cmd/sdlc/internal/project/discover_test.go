package project

import (
	"os"
	"path/filepath"
	"testing"
)

// writeProjectStatus writes a minimal project record with the given status and
// marker body under <parent>/<repo>/<subdir>/<name>.
func writeProjectStatus(t *testing.T, parent, repo, subdir, name, status, marker string) {
	t.Helper()
	dir := filepath.Join(parent, repo, subdir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "---\ntype: project\nname: " + name + "\ngoal: g\ndone_when: w\nstatus: " + status +
		"\n---\n\n# " + name + "\n\n- [ ] a task " + marker + "\n"
	if err := os.WriteFile(filepath.Join(dir, name+".md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// writeProject is the executing-status convenience form.
func writeProject(t *testing.T, parent, repo, subdir, name, marker string) {
	t.Helper()
	writeProjectStatus(t, parent, repo, subdir, name, "executing", marker)
}

func TestDiscoverByIssueRef_AllMatchesAcrossPeers(t *testing.T) {
	parent := t.TempDir()
	writeProject(t, parent, "metis", "workshop/projects", "p1", "[metis#18 M1]")
	writeProject(t, parent, "kbench", "workshop/projects", "p2", "[metis#18]")
	writeProject(t, parent, "nous", "workshop/projects", "p3", "[nous#9]")
	// brain legacy record — ACTIVE, so it survives ActiveOnly.
	writeProjectStatus(t, parent, "brain", "data/project", "legacy", "active", "[metis#18]")
	// an ARCHIVED match (history/projects) — only ActiveAndArchive sees it.
	writeProject(t, parent, "metis", "workshop/history/projects", "old", "[metis#18]")

	act, err := DiscoverByIssueRef(parent, "metis", "18", ActiveOnly)
	if err != nil {
		t.Fatal(err)
	}
	if len(act) != 3 {
		t.Fatalf("ActiveOnly: want 3 (2 active + 1 active-legacy), got %d: %+v", len(act), act)
	}
	legacy := 0
	for _, m := range act {
		if m.Legacy {
			legacy++
		}
	}
	if legacy != 1 {
		t.Fatalf("want 1 legacy match, got %d", legacy)
	}

	all, err := DiscoverByIssueRef(parent, "metis", "18", ActiveAndArchive)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 4 {
		t.Fatalf("ActiveAndArchive: want 4 (incl. archived), got %d: %+v", len(all), all)
	}
}

func TestDiscoverByIssueRef_DropsTerminalLegacyUnderActiveOnly(t *testing.T) {
	parent := t.TempDir()
	// a DONE brain-legacy record must NOT appear under ActiveOnly (would be
	// re-ticked by the close gate) but MUST appear under ActiveAndArchive.
	writeProjectStatus(t, parent, "brain", "data/project", "done-legacy", "done", "[metis#18]")

	act, _ := DiscoverByIssueRef(parent, "metis", "18", ActiveOnly)
	if len(act) != 0 {
		t.Fatalf("done legacy must be dropped under ActiveOnly, got %d: %+v", len(act), act)
	}
	all, _ := DiscoverByIssueRef(parent, "metis", "18", ActiveAndArchive)
	if len(all) != 1 {
		t.Fatalf("done legacy must be found under ActiveAndArchive, got %d", len(all))
	}
}

func TestDiscoverByIssueRef_SkipsStaleSiblings(t *testing.T) {
	parent := t.TempDir()
	writeProject(t, parent, "metis", "workshop/projects", "p1", "[metis#18]")
	writeProject(t, parent, "metis.bak", "workshop/projects", "p1", "[metis#18]") // backup copy
	writeProject(t, parent, "worktree", "workshop/projects", "x", "[metis#18]")    // worktree tree
	writeProject(t, parent, ".hidden", "workshop/projects", "y", "[metis#18]")     // dot-dir
	got, _ := DiscoverByIssueRef(parent, "metis", "18", ActiveOnly)
	if len(got) != 1 {
		t.Fatalf("stale siblings must be skipped; want 1, got %d: %+v", len(got), got)
	}
}

func TestDiscoverByIssueRef_ZeroMatches(t *testing.T) {
	parent := t.TempDir()
	writeProject(t, parent, "nous", "workshop/projects", "p", "[nous#1]")
	got, err := DiscoverByIssueRef(parent, "metis", "18", ActiveOnly)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("want 0 matches, got %d", len(got))
	}
}

func TestDiscoverByIssueRef_SameRepoTwoProjects(t *testing.T) {
	parent := t.TempDir()
	writeProject(t, parent, "metis", "workshop/projects", "a", "[metis#18]")
	writeProject(t, parent, "metis", "workshop/projects", "b", "[metis#18 M2]")
	got, _ := DiscoverByIssueRef(parent, "metis", "18", ActiveOnly)
	if len(got) != 2 {
		t.Fatalf("both same-repo projects should match; want 2, got %d", len(got))
	}
}

func TestSiblingRepoDirs_ReturnsAllDirs(t *testing.T) {
	parent := t.TempDir()
	for _, d := range []string{"metis", "nous", "metis.bak", "worktree"} {
		if err := os.MkdirAll(filepath.Join(parent, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	// a plain file is not a dir and must be excluded
	if err := os.WriteFile(filepath.Join(parent, "afile"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := SiblingRepoDirs(parent)
	if err != nil {
		t.Fatal(err)
	}
	// SiblingRepoDirs applies NO filtering — all four dirs returned (the skip
	// list lives in DiscoverByIssueRef so resolveRepoDir stays identical).
	if len(got) != 4 {
		t.Fatalf("want 4 dirs (no filtering), got %d: %+v", len(got), got)
	}
}
