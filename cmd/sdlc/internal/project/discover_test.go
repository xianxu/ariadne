package project

import (
	"os"
	"path/filepath"
	"strings"
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

// markBrain writes <parent>/<repo>/.brain/config.md so gitx.IsBrainRepo treats
// the dir as a brain (the canonical predicate — a basename is not enough).
func markBrain(t *testing.T, parent, repo string) {
	t.Helper()
	dir := filepath.Join(parent, repo, ".brain")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.md"), []byte("brain\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestDiscoverByIssueRef_AllMatchesAcrossPeers(t *testing.T) {
	parent := t.TempDir()
	writeProject(t, parent, "metis", "workshop/projects", "p1", "[metis#18 M1]")
	writeProject(t, parent, "kbench", "workshop/projects", "p2", "[metis#18]")
	writeProject(t, parent, "nous", "workshop/projects", "p3", "[nous#9]")
	// brain legacy record — ACTIVE, so it survives ActiveOnly. Requires the
	// .brain/config.md marker for gitx.IsBrainRepo to treat it as brain.
	writeProjectStatus(t, parent, "brain", "data/project", "legacy", "active", "[metis#18]")
	markBrain(t, parent, "brain")
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
	byRepo := map[string]ProjectMatch{}
	for _, m := range act {
		byRepo[m.Repo] = m
		if m.Legacy {
			legacy++
		}
		// RepoDir must be the owning repo root (M3's peer-commit scoping keys
		// on this) and Repo its basename.
		if filepath.Base(m.RepoDir) != m.Repo {
			t.Errorf("Repo %q != base(RepoDir %q)", m.Repo, m.RepoDir)
		}
		if !strings.HasPrefix(m.Path, m.RepoDir+string(filepath.Separator)) {
			t.Errorf("Path %q not under RepoDir %q", m.Path, m.RepoDir)
		}
	}
	if legacy != 1 {
		t.Fatalf("want 1 legacy match, got %d", legacy)
	}
	if byRepo["metis"].RepoDir != filepath.Join(parent, "metis") {
		t.Errorf("metis RepoDir = %q, want %q", byRepo["metis"].RepoDir, filepath.Join(parent, "metis"))
	}
	if byRepo["brain"].RepoDir != filepath.Join(parent, "brain") || !byRepo["brain"].Legacy {
		t.Errorf("brain match RepoDir/Legacy wrong: %+v", byRepo["brain"])
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
	markBrain(t, parent, "brain")

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
	writeProject(t, parent, "worktree", "workshop/projects", "x", "[metis#18]")   // worktree tree
	writeProject(t, parent, ".hidden", "workshop/projects", "y", "[metis#18]")    // dot-dir
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

// A repo is brain by the .brain/config.md predicate, NOT its basename: a repo
// named "capture" with the marker is scanned as brain (data/project, legacy),
// while a repo literally named "brain" WITHOUT the marker is a normal fleet home
// (workshop/projects). This pins finding 3.1's fix.
func TestDiscoverByIssueRef_BrainIsPredicateNotBasename(t *testing.T) {
	parent := t.TempDir()
	// "capture" is a brain (has the marker); its project lives in data/project.
	writeProjectStatus(t, parent, "capture", "data/project", "p", "active", "[metis#18]")
	markBrain(t, parent, "capture")
	// a dir literally named "brain" but WITHOUT the marker is a normal repo;
	// its project lives in workshop/projects and data/project is NOT scanned.
	writeProject(t, parent, "brain", "workshop/projects", "q", "[metis#18]")
	writeProject(t, parent, "brain", "data/project", "ignored", "[metis#18]")

	got, _ := DiscoverByIssueRef(parent, "metis", "18", ActiveOnly)
	if len(got) != 2 {
		t.Fatalf("want 2 (capture legacy + brain-named fleet home), got %d: %+v", len(got), got)
	}
	for _, m := range got {
		switch m.Repo {
		case "capture":
			if !m.Legacy || filepath.Base(filepath.Dir(m.Path)) != "project" {
				t.Errorf("capture should be a legacy data/project match: %+v", m)
			}
		case "brain":
			if m.Legacy || filepath.Base(filepath.Dir(m.Path)) != "projects" {
				t.Errorf("brain-named repo (no marker) should be a fleet workshop/projects match: %+v", m)
			}
		}
	}
}

func TestDiscoverByIssueRef_DedupsSymlinkAndSkipsUnreadable(t *testing.T) {
	parent := t.TempDir()
	writeProject(t, parent, "metis", "workshop/projects", "p", "[metis#18]")
	// A second .md in the same scanned dir that symlinks to the real file: both
	// glob-match but resolve (EvalSymlinks) to one real path → counted once.
	projects := filepath.Join(parent, "metis", "workshop", "projects")
	if err := os.Symlink(filepath.Join(projects, "p.md"), filepath.Join(projects, "alias.md")); err != nil {
		t.Skipf("symlinks unsupported: %v", err)
	}
	got, err := DiscoverByIssueRef(parent, "metis", "18", ActiveOnly)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("symlinked duplicate must be deduped to 1, got %d: %+v", len(got), got)
	}

	// An unreadable project file is skipped best-effort, not fatal.
	bad := filepath.Join(parent, "kbench", "workshop", "projects")
	if err := os.MkdirAll(bad, 0o755); err != nil {
		t.Fatal(err)
	}
	badFile := filepath.Join(bad, "x.md")
	if err := os.WriteFile(badFile, []byte("[metis#18]"), 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(badFile, 0o644) })
	if _, err := DiscoverByIssueRef(parent, "metis", "18", ActiveOnly); err != nil {
		t.Fatalf("unreadable file must be skipped, not error: %v", err)
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

// TestDiscoverByIssueRef_IDBoundary pins the marker boundary (#171 M4 review):
// searching #18 must NOT match a record referencing #180 (the open-bracket
// prefix alone would), while both documented forms — "[repo#18]" and
// "[repo#18 Mx]" — still match. The close gate shares this match, so without
// the boundary a close of #18 would falsely tick #180's project.
func TestDiscoverByIssueRef_IDBoundary(t *testing.T) {
	parent := t.TempDir()
	writeProject(t, parent, "metis", "workshop/projects", "longer", "[metis#180]")
	writeProject(t, parent, "kbench", "workshop/projects", "bare", "[metis#18]")
	writeProject(t, parent, "nous", "workshop/projects", "tagged", "[metis#18 M2]")

	got, err := DiscoverByIssueRef(parent, "metis", "18", ActiveAndArchive)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 matches (bare + milestone-tagged), got %d: %+v", len(got), got)
	}
	for _, m := range got {
		if strings.Contains(m.Path, "longer") {
			t.Errorf("#18 matched the #180 record: %+v", m)
		}
	}

	longer, err := DiscoverByIssueRef(parent, "metis", "180", ActiveAndArchive)
	if err != nil {
		t.Fatal(err)
	}
	if len(longer) != 1 || !strings.Contains(longer[0].Path, "longer") {
		t.Fatalf("#180 should match exactly its own record: %+v", longer)
	}
}

func TestContainsIssueMarker(t *testing.T) {
	cases := []struct {
		text string
		want bool
	}{
		{"- [ ] [metis#18] t", true},
		{"- [ ] [metis#18 M2] t", true},
		{"- [ ] [metis#180] t", false},
		{"tail is the marker [metis#18", true}, // EOF is a non-digit boundary
		{"[metis#180] then [metis#18]", true},  // later true occurrence still found
		{"no ref here", false},
	}
	for _, c := range cases {
		if got := containsIssueMarker(c.text, "[metis#18"); got != c.want {
			t.Errorf("containsIssueMarker(%q) = %v, want %v", c.text, got, c.want)
		}
	}
}
