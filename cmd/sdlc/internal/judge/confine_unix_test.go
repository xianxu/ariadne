//go:build unix

package judge

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// confinedReviewer sets up the #339 incident's world: HOME is a temp dir, the
// working directory is a repository under it, and `claude` on PATH is script.
// It returns home and the marker the script touches when it runs.
func confinedReviewer(t *testing.T, script string) (home, marker string) {
	t.Helper()
	home = t.TempDir()
	repo := filepath.Join(home, "w", "repo")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("git", "-C", repo, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v %s", err, out)
	}
	bin := t.TempDir()
	marker = filepath.Join(bin, "ran")
	body := "#!/bin/sh\n: > " + "'" + marker + "'\n" + script +
		"\nprintf 'VERDICT: CLEAN\\n\\n\\140\\140\\140findings\\nfindings: []\\n\\140\\140\\140\\n'\n"
	if err := os.WriteFile(filepath.Join(bin, "claude"), []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Chdir(repo)
	saved := Run
	Run = runProcess
	t.Cleanup(func() { Run = saved })
	return home, marker
}

// incidentScript is what the #339 reviewer ran: its temp dir is unusable (an
// inherited CLAUDE_CODE_TMPDIR the nested sandbox could not write), mktemp
// fails, `cd` with no argument lands in $HOME, and it builds a scratch repo.
const incidentScript = `TMPDIR=/nonexistent/sdlc-339; export TMPDIR
cd $(mktemp -d 2>/dev/null)
git init -q . 2>&1
mkdir -p scripts/merge-checks.d && : > scripts/merge-checks.d/50-x.sh
git add -A 2>&1; git commit -qm a 2>&1`

func TestConfinedReviewerCannotBuildARepositoryInHome(t *testing.T) {
	home, marker := confinedReviewer(t, incidentScript)
	out, err := Dispatch(context.Background(), DispatchOptions{Agent: AgentClaude, Prompt: "review"})
	if _, serr := os.Stat(marker); serr != nil {
		t.Fatalf("the fake reviewer never ran: %v", serr)
	}
	if _, serr := os.Stat(filepath.Join(home, ".git")); serr == nil {
		t.Fatal("the reviewer created a repository in HOME")
	}
	if err == nil || !strings.Contains(err.Error(), filepath.Join(home, "scripts")) {
		t.Fatalf("Dispatch err = %v, want the outside write to %s named\noutput: %s", err, filepath.Join(home, "scripts"), out)
	}
}

func TestConfinementFailureAbortsBeforeTheReviewerRuns(t *testing.T) {
	_, marker := confinedReviewer(t, "")
	parent := t.TempDir()
	if err := os.Chmod(parent, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(parent, 0o700) })
	t.Setenv("TMPDIR", parent)
	_, err := Dispatch(context.Background(), DispatchOptions{Agent: AgentClaude, Prompt: "review"})
	if err == nil || !strings.Contains(err.Error(), "did not run") || !strings.Contains(err.Error(), "temp dir") {
		t.Fatalf("Dispatch err = %v, want an aborted review naming the temp dir", err)
	}
	if _, serr := os.Stat(marker); serr == nil {
		t.Fatal("the reviewer ran without a confinement")
	}
}

func TestConfinedReviewerGetsItsOwnTempDir(t *testing.T) {
	// The reviewer's TMPDIR works, scratch repos in it work, and the base is
	// gone afterward. The script fails the review if TMPDIR is the inherited one.
	inherited := os.Getenv("TMPDIR")
	report := filepath.Join(t.TempDir(), "tmpdir")
	_, _ = confinedReviewer(t, `echo "$TMPDIR" > '`+report+`'
[ "$TMPDIR" = "$CLAUDE_CODE_TMPDIR" ] || { echo tmpdiff >&2; exit 1; }
d=$(mktemp -d) || { echo mktemp >&2; exit 1; }
cd "$d" && git init -q . && git rev-parse --show-toplevel >/dev/null || { echo "git in $d" >&2; exit 1; }`)
	var diag strings.Builder
	out, err := Dispatch(context.Background(), DispatchOptions{Agent: AgentClaude, Prompt: "review", Stderr: &diag})
	if err != nil || !HasVerdict(out) {
		t.Fatalf("Dispatch = %q, %v\nreviewer stderr: %s", out, err, diag.String())
	}
	data, rerr := os.ReadFile(report)
	if rerr != nil {
		t.Fatal(rerr)
	}
	tmp := strings.TrimSpace(string(data))
	if tmp == "" || tmp == inherited || !strings.Contains(tmp, "sdlc-review-") {
		t.Fatalf("reviewer TMPDIR = %q (inherited %q)", tmp, inherited)
	}
	if _, serr := os.Stat(filepath.Dir(tmp)); serr == nil {
		t.Fatalf("the confinement base %s outlived the dispatch", filepath.Dir(tmp))
	}
}
