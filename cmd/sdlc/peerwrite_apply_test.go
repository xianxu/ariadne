package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// gitIn runs git in dir, failing the test on error.
func gitIn(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git -C %s %v: %v — %s", dir, args, err, out)
	}
	return string(out)
}

// initFleetRepo creates parent/name as a git repo on main with one initial commit.
func initFleetRepo(t *testing.T, parent, name string) string {
	t.Helper()
	dir := filepath.Join(parent, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	gitIn(t, dir, "init", "-q", "-b", "main")
	gitIn(t, dir, "config", "user.email", "t@t")
	gitIn(t, dir, "config", "user.name", "t")
	gitIn(t, dir, "config", "commit.gpgsign", "false")
	if err := os.WriteFile(filepath.Join(dir, "README"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitIn(t, dir, "add", "README")
	gitIn(t, dir, "commit", "-q", "-m", "init")
	return dir
}

// seedPeerProject commits a project file in the peer referencing ref, returning
// its repo-relative path.
func seedPeerProject(t *testing.T, peerDir, ref string) string {
	t.Helper()
	rel := filepath.Join("workshop", "projects", "p.md")
	abs := filepath.Join(peerDir, rel)
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatal(err)
	}
	content := "---\nstatus: executing\n---\n\n# p\n\n## tasks\n\n- [ ] [" + ref + "] the thing\n"
	if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	gitIn(t, peerDir, "add", rel)
	gitIn(t, peerDir, "commit", "-q", "-m", "project: seed")
	return rel
}

func TestReadRepoGitState(t *testing.T) {
	parent := t.TempDir()
	peer := initFleetRepo(t, parent, "peer")

	st := readRepoGitState(execGitRunner{}, peer, []string{"README"})
	if st.Branch != "main" || st.HasStagedChanges || st.TargetFilesDirty || st.IsBrain {
		t.Errorf("clean main repo: got %+v, want {main false false false}", st)
	}

	// Staged change flips HasStagedChanges.
	if err := os.WriteFile(filepath.Join(peer, "staged.txt"), []byte("s\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitIn(t, peer, "add", "staged.txt")
	if st := readRepoGitState(execGitRunner{}, peer, nil); !st.HasStagedChanges {
		t.Errorf("staged change not detected: %+v", st)
	}
	gitIn(t, peer, "reset", "-q", "HEAD", "staged.txt")
	if err := os.Remove(filepath.Join(peer, "staged.txt")); err != nil {
		t.Fatal(err)
	}

	// Unstaged edits to a TARGET file flip TargetFilesDirty (other files don't).
	if err := os.WriteFile(filepath.Join(peer, "README"), []byte("edited\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if st := readRepoGitState(execGitRunner{}, peer, []string{"README"}); !st.TargetFilesDirty {
		t.Errorf("unstaged edit to target file not detected: %+v", st)
	}
	if st := readRepoGitState(execGitRunner{}, peer, []string{"other.md"}); st.TargetFilesDirty {
		t.Errorf("dirt on a NON-target file must not flip TargetFilesDirty: %+v", st)
	}
	gitIn(t, peer, "checkout", "-q", "--", "README")

	// An UNTRACKED target file counts as dirty too (status --porcelain sees it).
	untracked := filepath.Join("workshop", "projects", "new.md")
	if err := os.MkdirAll(filepath.Join(peer, "workshop", "projects"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(peer, untracked), []byte("n\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if st := readRepoGitState(execGitRunner{}, peer, []string{untracked}); !st.TargetFilesDirty {
		t.Errorf("untracked target file not detected as dirty: %+v", st)
	}
	if err := os.Remove(filepath.Join(peer, untracked)); err != nil {
		t.Fatal(err)
	}

	// Off-main branch is reported.
	gitIn(t, peer, "checkout", "-q", "-b", "feature-x")
	if st := readRepoGitState(execGitRunner{}, peer, nil); st.Branch != "feature-x" {
		t.Errorf("branch = %q, want feature-x", st.Branch)
	}
	gitIn(t, peer, "checkout", "-q", "main")

	// A non-git dir yields Branch "" (never garbled error text).
	nonGit := filepath.Join(parent, "notarepo")
	if err := os.MkdirAll(nonGit, 0o755); err != nil {
		t.Fatal(err)
	}
	if st := readRepoGitState(execGitRunner{}, nonGit, nil); st.Branch != "" {
		t.Errorf("non-git dir should yield empty Branch, got %q", st.Branch)
	}

	// The .brain/config.md predicate marks a brain regardless of basename.
	if err := os.MkdirAll(filepath.Join(peer, ".brain"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(peer, ".brain", "config.md"), []byte("brain\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if st := readRepoGitState(execGitRunner{}, peer, nil); !st.IsBrain {
		t.Errorf("brain predicate not detected: %+v", st)
	}
}

// TestApplyPeerWrites drives the shell against real repos: a committing
// decision produces a scoped commit touching exactly the project file; a
// report-only decision leaves the file written but uncommitted and the
// pre-existing staged change untouched.
func TestApplyPeerWrites(t *testing.T) {
	parent := t.TempDir()
	cur := filepath.Join(parent, "ariadne") // not materialized — planner only compares the path

	clean := initFleetRepo(t, parent, "nous")
	cleanRel := seedPeerProject(t, clean, "ariadne#42")

	dirty := initFleetRepo(t, parent, "kbench")
	dirtyRel := seedPeerProject(t, dirty, "ariadne#42")
	if err := os.WriteFile(filepath.Join(dirty, "unrelated.txt"), []byte("u\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitIn(t, dirty, "add", "unrelated.txt")

	edits := map[string][]string{
		clean: {cleanRel},
		dirty: {dirtyRel},
	}
	// Production order (applyClose): snapshot state BEFORE writing the edits.
	states := map[string]RepoGitState{
		clean: readRepoGitState(execGitRunner{}, clean, edits[clean]),
		dirty: readRepoGitState(execGitRunner{}, dirty, edits[dirty]),
	}
	if err := os.WriteFile(filepath.Join(clean, cleanRel), []byte("ticked\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dirty, dirtyRel), []byte("ticked\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	decisions := planPeerWrites(edits, states, cur, "ariadne#42")

	var stdout, stderr bytes.Buffer
	applyPeerWrites(execGitRunner{}, decisions, &stdout, &stderr)

	// Clean peer: exactly one new commit, touching only the project file, tree clean.
	subject := gitIn(t, clean, "log", "-1", "--pretty=%s")
	if !strings.Contains(subject, "ariadne#42") {
		t.Errorf("clean peer commit subject %q should cite the closing ref", subject)
	}
	touched := strings.TrimSpace(gitIn(t, clean, "show", "--name-only", "--pretty=format:", "HEAD"))
	if touched != cleanRel {
		t.Errorf("scoped commit touched %q, want only %q", touched, cleanRel)
	}
	if status := strings.TrimSpace(gitIn(t, clean, "status", "--porcelain")); status != "" {
		t.Errorf("clean peer tree not clean after scoped commit:\n%s", status)
	}
	if !strings.Contains(stdout.String(), "committed "+cleanRel+" in nous") {
		t.Errorf("stdout missing committed report: %q", stdout.String())
	}

	// Dirty peer: no new commit, project file written but unstaged, staged change intact.
	if subject := gitIn(t, dirty, "log", "-1", "--pretty=%s"); !strings.Contains(subject, "seed") {
		t.Errorf("dirty peer got an unexpected commit: %q", subject)
	}
	status := gitIn(t, dirty, "status", "--porcelain")
	if !strings.Contains(status, "A  unrelated.txt") {
		t.Errorf("pre-existing staged change disturbed:\n%s", status)
	}
	if !strings.Contains(status, " M "+dirtyRel) {
		t.Errorf("project file should be modified-unstaged:\n%s", status)
	}
	if !strings.Contains(stderr.String(), "NOT committed in kbench") || !strings.Contains(stderr.String(), "staged") {
		t.Errorf("stderr missing report-only reason: %q", stderr.String())
	}
	if !strings.Contains(stderr.String(), "git add") || !strings.Contains(stderr.String(), dirtyRel) {
		t.Errorf("stderr missing exact next action: %q", stderr.String())
	}
}

// failingRunner errors on every git call — pins applyPeerWrites's
// warn-and-continue contract: a peer git failure never panics or aborts.
type failingRunner struct{}

func (failingRunner) Git(args ...string) ([]byte, error) {
	return []byte("boom"), fmt.Errorf("git %v failed", args)
}
func (failingRunner) GitInDir(dir string, args ...string) ([]byte, error) {
	return []byte("boom"), fmt.Errorf("git -C %s %v failed", dir, args)
}
func (failingRunner) MkdirAll(path string) error               { return nil }
func (failingRunner) WriteFile(path string, data []byte) error { return nil }

func TestApplyPeerWrites_GitFailureWarnsAndContinues(t *testing.T) {
	decisions := []PeerWriteDecision{
		{RepoDir: "/fleet/nous", Files: []string{"workshop/projects/p.md"}, Commit: true, Message: "m"},
		{RepoDir: "/fleet/metis", Files: []string{"workshop/projects/q.md"}, Reason: "off-main", NextAction: "cd …"},
	}
	var stdout, stderr bytes.Buffer
	applyPeerWrites(failingRunner{}, decisions, &stdout, &stderr) // must not panic or die
	if !strings.Contains(stderr.String(), "git add in nous failed") {
		t.Errorf("expected add-failure warning, got %q", stderr.String())
	}
	if !strings.Contains(stderr.String(), "NOT committed in metis") {
		t.Errorf("report-only decision must still be reported after a prior failure: %q", stderr.String())
	}
	if strings.Contains(stdout.String(), "committed") {
		t.Errorf("no commit succeeded, stdout must not claim one: %q", stdout.String())
	}
}

// fleetCloseFixture builds a temp fleet: parent/ariadne is the closing repo
// (chdir'd into, with issue #42 committed under a "#42" subject) and returns
// (parent, issuesDir). Peers are added by the caller before running the close.
func fleetCloseFixture(t *testing.T) (string, string) {
	t.Helper()
	parent := t.TempDir()
	repo := initFleetRepo(t, parent, "ariadne")
	issuesDir := "workshop/issues"
	if err := os.MkdirAll(filepath.Join(repo, issuesDir), 0o755); err != nil {
		t.Fatal(err)
	}
	content := "---\nid: 000042\nstatus: working\nestimate_hours: 1\n---\n\n" +
		"# x\n\n## Spec\n\nThing.\n\n## Plan\n\n- [x] do it\n\n## Log\n"
	if err := os.WriteFile(filepath.Join(repo, issuesDir, "000042-x.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	gitIn(t, repo, "add", ".")
	gitIn(t, repo, "commit", "-q", "-m", "#42: implement the thing")
	cwd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	if err := os.Chdir(repo); err != nil {
		t.Fatal(err)
	}
	return parent, issuesDir
}

func closeFlagsFor42(issuesDir string) *closeFlags {
	return &closeFlags{
		Issue:     42,
		NoActual:  true,
		Verified:  "peer-write test close",
		NoAtlas:   true,
		IssuesDir: issuesDir,
		BrainDir:  "../nonexistent-brain",
	}
}

// TestClose_PeerProjectCommitted is the M2-review-deferred test: a close whose
// matched project lives in a DIFFERENT repo than the closing one. The peer is
// on main with a clean index, so the close ticks the row AND commits it there.
func TestClose_PeerProjectCommitted(t *testing.T) {
	parent, issuesDir := fleetCloseFixture(t)
	peer := initFleetRepo(t, parent, "nous")
	rel := seedPeerProject(t, peer, "ariadne#42")

	var out bytes.Buffer
	if err := runClose(&out, &out, closeFlagsFor42(issuesDir)); err != nil {
		t.Fatalf("runClose: %v\n%s", err, out.String())
	}

	data, err := os.ReadFile(filepath.Join(peer, rel))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "- [x] [ariadne#42]") {
		t.Errorf("peer project row not ticked:\n%s", data)
	}
	subject := gitIn(t, peer, "log", "-1", "--pretty=%s")
	if !strings.Contains(subject, "ariadne#42") {
		t.Errorf("peer commit subject %q should be the scoped close-time update", subject)
	}
	if status := strings.TrimSpace(gitIn(t, peer, "status", "--porcelain")); status != "" {
		t.Errorf("peer tree not clean after scoped commit:\n%s", status)
	}
	if !strings.Contains(out.String(), "committed "+rel+" in nous") {
		t.Errorf("missing committed report in output: %q", out.String())
	}
}

// TestClose_PeerOffMainReportOnly: the peer sits on a feature branch, so the
// close updates the project file but hands the commit to the operator with the
// branch reason and the exact next action. The close itself still succeeds.
func TestClose_PeerOffMainReportOnly(t *testing.T) {
	parent, issuesDir := fleetCloseFixture(t)
	peer := initFleetRepo(t, parent, "nous")
	rel := seedPeerProject(t, peer, "ariadne#42")
	gitIn(t, peer, "checkout", "-q", "-b", "feature-y")

	var out bytes.Buffer
	if err := runClose(&out, &out, closeFlagsFor42(issuesDir)); err != nil {
		t.Fatalf("runClose: %v\n%s", err, out.String())
	}

	data, err := os.ReadFile(filepath.Join(peer, rel))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "- [x] [ariadne#42]") {
		t.Errorf("peer project row should still be ticked (write happens; commit doesn't):\n%s", data)
	}
	if subject := gitIn(t, peer, "log", "-1", "--pretty=%s"); !strings.Contains(subject, "seed") {
		t.Errorf("off-main peer must not get an auto-commit, HEAD is %q", subject)
	}
	msg := out.String()
	if !strings.Contains(msg, "feature-y") || !strings.Contains(msg, "NOT committed in nous") {
		t.Errorf("missing off-main report: %q", msg)
	}
	if !strings.Contains(msg, "git add") || !strings.Contains(msg, rel) {
		t.Errorf("missing exact next action: %q", msg)
	}
}

// TestClose_PeerDirtyProjectFileReportOnly pins the M3-review Important: a
// peer on main with a CLEAN index but uncommitted working-tree edits to the
// very project file being ticked must flip to report-only — the close writes
// the ticked content (preserving the local edits) but never publishes another
// session's work inside its scoped commit.
func TestClose_PeerDirtyProjectFileReportOnly(t *testing.T) {
	parent, issuesDir := fleetCloseFixture(t)
	peer := initFleetRepo(t, parent, "nous")
	rel := seedPeerProject(t, peer, "ariadne#42")

	// Another session's unstaged edit to the same project file.
	abs := filepath.Join(peer, rel)
	orig, err := os.ReadFile(abs)
	if err != nil {
		t.Fatal(err)
	}
	edited := strings.Replace(string(orig), "# p", "# p\n\nlocal note from another session", 1)
	if err := os.WriteFile(abs, []byte(edited), 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := runClose(&out, &out, closeFlagsFor42(issuesDir)); err != nil {
		t.Fatalf("runClose: %v\n%s", err, out.String())
	}

	data, err := os.ReadFile(abs)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "- [x] [ariadne#42]") {
		t.Errorf("row should still be ticked in the working tree:\n%s", data)
	}
	if !strings.Contains(string(data), "local note from another session") {
		t.Errorf("the other session's edit must be preserved in the file:\n%s", data)
	}
	if subject := gitIn(t, peer, "log", "-1", "--pretty=%s"); !strings.Contains(subject, "seed") {
		t.Errorf("dirty-target peer must not be auto-committed, HEAD is %q", subject)
	}
	if !strings.Contains(out.String(), "uncommitted edits") || !strings.Contains(out.String(), "NOT committed in nous") {
		t.Errorf("missing dirty-target report: %q", out.String())
	}
}

// TestClose_CurrentRepoProjectNotAutoCommitted: a project in the CLOSING repo
// rides the normal close commit — the peer-write path must not touch it.
func TestClose_CurrentRepoProjectNotAutoCommitted(t *testing.T) {
	parent, issuesDir := fleetCloseFixture(t)
	repo := filepath.Join(parent, "ariadne")
	rel := seedPeerProject(t, repo, "ariadne#42")

	var out bytes.Buffer
	if err := runClose(&out, &out, closeFlagsFor42(issuesDir)); err != nil {
		t.Fatalf("runClose: %v\n%s", err, out.String())
	}

	if subject := gitIn(t, repo, "log", "-1", "--pretty=%s"); strings.Contains(subject, "close-time update") {
		t.Errorf("current repo must not be auto-committed by the peer-write path, HEAD is %q", subject)
	}
	status := gitIn(t, repo, "status", "--porcelain")
	if !strings.Contains(status, " M "+rel) {
		t.Errorf("current repo project edit should sit uncommitted for the close commit:\n%s", status)
	}
	if strings.Contains(out.String(), "NOT committed in ariadne") {
		t.Errorf("peer-write path reported the current repo: %q", out.String())
	}
}
