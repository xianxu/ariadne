package activetime

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/pkg/workspace"
)

// #252: the tracker's card commits are read beside HEAD, and a commit
// reachable from both is one event.
func TestLoadWindowCommitsReadsExtraRefsOnce(t *testing.T) {
	repo := t.TempDir()
	run := func(args ...string) string {
		cmd := exec.Command("git", append([]string{"-C", repo}, args...)...)
		cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	run("init", "-q", "-b", "main")
	run("commit", "-q", "--allow-empty", "-m", "#9 code")
	run("switch", "-q", "--orphan", "tracker")
	run("commit", "-q", "--allow-empty", "-m", "#9: tracker: update card")
	run("switch", "-q", "main")
	orig := resolveCommitWorkspace
	resolveCommitWorkspace = func(r string) (workspace.Identity, error) { return workspace.Identity{Repo: filepath.Base(r)}, nil }
	t.Cleanup(func() { resolveCommitWorkspace = orig })

	head, err := loadWindowCommits(repo, "", "")
	if err != nil || len(head) != 1 {
		t.Fatalf("HEAD only: %v %v", head, err)
	}
	both, err := loadWindowCommits(repo, "", "", "refs/heads/tracker", "refs/heads/main")
	if err != nil || len(both) != 2 {
		t.Fatalf("with the tracker ref (and HEAD twice): %+v %v", both, err)
	}
	for _, c := range both {
		if len(c.Issues) != 1 || c.Issues[0] != "9" {
			t.Errorf("commit %q not attributed to #9: %v", c.Subject, c.Issues)
		}
	}
}
