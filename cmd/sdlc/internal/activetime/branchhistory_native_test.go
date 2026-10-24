package activetime

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseHeadHistoryNativeWorktreeTransitions(t *testing.T) {
	repo := gitInit(t)
	gitAt(t, repo, "2026-10-13T23:00:00Z", "commit", "--allow-empty", "-qm", "base")
	gitAt(t, repo, "", "branch", "-M", "main")
	worker := filepath.Join(t.TempDir(), "worker")
	other := filepath.Join(t.TempDir(), "other")
	gitAt(t, repo, "2026-10-14T00:00:00Z", "worktree", "add", "-q", "-b", "000336-a", worker)
	gitAt(t, repo, "2026-10-14T00:05:00Z", "worktree", "add", "-q", "-b", "000337-other", other)
	gitAt(t, worker, "2026-10-14T00:10:00Z", "commit", "--allow-empty", "-qm", "#336 close")
	gitAt(t, worker, "2026-10-14T00:15:00Z", "branch", "-m", "000336-renamed")
	gitAt(t, worker, "2026-10-14T00:20:00Z", "switch", "-q", "--detach")
	gitAt(t, worker, "2026-10-14T00:25:00Z", "switch", "-q", "000336-renamed")
	gitAt(t, worker, "2026-10-14T00:25:00Z", "reset", "--hard", "HEAD")
	gitAt(t, worker, "2026-10-14T00:30:00Z", "switch", "-q", "-c", "000335-b", "main")
	logPath := gitAt(t, worker, "", "rev-parse", "--path-format=absolute", "--git-path", "logs/HEAD")
	otherLog := gitAt(t, other, "", "rev-parse", "--path-format=absolute", "--git-path", "logs/HEAD")
	if logPath == otherLog {
		t.Fatal("two worktrees share HEAD history")
	}
	raw, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "000337-other") {
		t.Fatal("another slot entered the selected HEAD history")
	}
	got, err := ParseHeadHistory(raw, "000335-b", gitAt(t, worker, "", "rev-parse", "HEAD"), branchTime(2), branchTime(35))
	if err != nil {
		t.Fatalf("%v\n%s", err, raw)
	}
	if len(got) != 5 || got[0].Branch != "000336-a" || got[1].Branch != "000336-renamed" || got[2].Known || got[3].Branch != "000336-renamed" || got[4].Branch != "000335-b" {
		t.Fatalf("native transitions = %+v\n%s", got, raw)
	}
	if !got[2].Start.Equal(branchTime(20)) || !got[2].End.Equal(branchTime(25)) || !got[4].Start.Equal(branchTime(30)) {
		t.Fatalf("native transition boundaries = %+v", got)
	}
}

func TestParseHeadHistoryNativeMergeRebaseAndRestingBoundary(t *testing.T) {
	repo := gitInit(t)
	gitAt(t, repo, "2026-10-13T23:00:00Z", "commit", "--allow-empty", "-qm", "base")
	gitAt(t, repo, "", "branch", "-M", "main")
	worker := filepath.Join(t.TempDir(), "worker")
	gitAt(t, repo, "2026-10-14T00:00:00Z", "worktree", "add", "-q", "-b", "000336-a", worker)
	commitFile := func(root, file, at string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(root, file), []byte(file), 0600); err != nil {
			t.Fatal(err)
		}
		gitAt(t, root, at, "add", file)
		gitAt(t, root, at, "commit", "-qm", "#336: "+file)
	}
	commitFile(worker, "a", "2026-10-14T00:10:00Z")
	commitFile(repo, "b", "2026-10-14T00:20:00Z")
	gitAt(t, worker, "2026-10-14T00:30:00Z", "merge", "--no-edit", "main")
	read := func(until int) []BranchInterval {
		t.Helper()
		path := gitAt(t, worker, "", "rev-parse", "--path-format=absolute", "--git-path", "logs/HEAD")
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		branch := gitAt(t, worker, "", "branch", "--show-current")
		got, err := ParseHeadHistory(raw, branch, gitAt(t, worker, "", "rev-parse", "HEAD"), branchTime(2), branchTime(until))
		if err != nil {
			t.Fatalf("native history: %v\n%s", err, raw)
		}
		return got
	}
	merged := read(35)
	if len(merged) != 1 || !merged[0].Known || merged[0].Branch != "000336-a" {
		t.Fatalf("merge changed active branch: %+v", merged)
	}
	commitFile(worker, "c", "2026-10-14T00:35:00Z")
	commitFile(repo, "d", "2026-10-14T00:40:00Z")
	gitAt(t, worker, "2026-10-14T00:50:00Z", "rebase", "main")
	gitAt(t, worker, "2026-10-14T01:00:00Z", "switch", "-q", "-c", "main-slot1", "main")
	gitAt(t, worker, "2026-10-14T01:10:00Z", "switch", "-q", "-c", "000335-b", "main")
	got := read(80)
	last := got[len(got)-1]
	if !last.Known || last.Branch != "000335-b" || !last.Start.Equal(branchTime(70)) {
		t.Fatalf("fresh main-based branch missing: %+v", got)
	}
	foundA, foundResting := false, false
	for _, w := range got {
		if w.Known && w.Branch == "000336-a" && w.Start.Equal(branchTime(50)) {
			foundA = true
		}
		if w.Known && w.Branch == "main-slot1" && w.Start.Equal(branchTime(60)) && w.End.Equal(branchTime(70)) {
			foundResting = true
		}
	}
	if !foundA || !foundResting {
		t.Fatalf("rebase return or resting boundary missing: %+v", got)
	}
}

func TestParseHeadHistoryCurrentSymbolicRefIsNotHistoricalBranchProof(t *testing.T) {
	repo := gitInit(t)
	gitAt(t, repo, "2026-10-13T23:00:00Z", "commit", "--allow-empty", "-qm", "base")
	worker := filepath.Join(t.TempDir(), "worker")
	gitAt(t, repo, "2026-10-14T00:00:00Z", "worktree", "add", "-q", "-b", "000336-a", worker)
	gitAt(t, worker, "2026-10-14T00:10:00Z", "commit", "--allow-empty", "-qm", "#336: actual work")
	gitAt(t, worker, "2026-10-14T00:11:00Z", "branch", "000335-b", "HEAD")
	path := gitAt(t, worker, "", "rev-parse", "--path-format=absolute", "--git-path", "logs/HEAD")
	// Both branches point to the same commit. Native symbolic-ref records a
	// branch-affecting action even without -m; an unsupported action must fence
	// the historical inference instead of backfilling today's branch.
	gitAt(t, worker, "2026-10-14T00:20:00Z", "symbolic-ref", "HEAD", "refs/heads/000335-b")
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	got, err := ParseHeadHistory(after, "000335-b", gitAt(t, worker, "", "rev-parse", "HEAD"), branchTime(2), branchTime(15))
	if err != nil {
		return
	} // Refusal is also an honest unknown result.
	for _, w := range got {
		if w.Known && w.Branch != "000336-a" {
			t.Fatalf("today's symbolic ref manufactured historical proof: %+v", got)
		}
	}
}
