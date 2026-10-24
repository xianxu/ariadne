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
