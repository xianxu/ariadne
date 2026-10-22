package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func TestIssueInitPresenceBeyondChangedActiveFiles(t *testing.T) {
	roots := []string{"workshop/issues", "workshop/history", "workshop/history/issues"}
	for _, root := range roots {
		for _, kind := range []string{"ignored", "index-only", "inherited", "unrelated", "symlink-root", "file-root", "unreadable-root"} {
			t.Run(root+"/"+kind, func(t *testing.T) {
				r := emptyInitRepo(t)
				const body = "---\nid: 000001\nstatus: done\n---\n\n# Earlier work\n"
				issuePath := root + "/000001-earlier.md"
				switch kind {
				case "inherited":
					writeRepoFile(t, r.root, issuePath, body)
					r.git("add", ".")
					r.git("commit", "-qm", "old issue")
					r.git("branch", "old-evidence")
					r.git("rm", issuePath)
					r.git("commit", "-qm", "main no longer has issue")
					r.git("push", "-q")
				case "unrelated":
					r.git("switch", "--orphan", "unrelated-evidence")
					writeRepoFile(t, r.root, issuePath, body)
					r.git("add", ".")
					r.git("commit", "-qm", "unrelated issue")
					r.git("switch", "-q", "main")
				default:
					wt := filepath.Join(t.TempDir(), "other")
					r.git("worktree", "add", "-q", "--detach", wt, "main")
					abs := filepath.Join(wt, filepath.FromSlash(root))
					switch kind {
					case "ignored":
						writeRepoFile(t, r.root, ".git/info/exclude", "/"+root+"/\n")
						writeRepoFile(t, wt, issuePath, body)
					case "index-only":
						writeRepoFile(t, wt, issuePath, body)
						testfix.Git(t, wt, "add", issuePath)
						if err := os.Remove(filepath.Join(wt, filepath.FromSlash(issuePath))); err != nil {
							t.Fatal(err)
						}
					case "symlink-root", "file-root":
						if err := os.MkdirAll(filepath.Dir(abs), 0755); err != nil {
							t.Fatal(err)
						}
						if kind == "symlink-root" {
							// Active exists in the scaffold; remove only the disposable empty root.
							if err := os.RemoveAll(abs); err != nil {
								t.Fatal(err)
							}
							if err := os.Symlink(t.TempDir(), abs); err != nil {
								t.Fatal(err)
							}
						} else {
							if err := os.RemoveAll(abs); err != nil {
								t.Fatal(err)
							}
							if err := os.WriteFile(abs, []byte("uncertain root\n"), 0600); err != nil {
								t.Fatal(err)
							}
						}
					case "unreadable-root":
						if err := os.MkdirAll(abs, 0755); err != nil {
							t.Fatal(err)
						}
						if err := os.Chmod(abs, 0); err != nil {
							t.Fatal(err)
						}
						t.Cleanup(func() { _ = os.Chmod(abs, 0755) })
					}
				}
				main, head, status := r.originMain(), r.git("rev-parse", "HEAD"), r.git("status", "--porcelain", "--untracked-files=all")
				_, stderr, err := executeSDLCTestCommand("issue", "init")
				if err == nil {
					t.Fatal("issue/uncertain evidence was called empty")
				}
				if kind != "unreadable-root" && !strings.Contains(err.Error()+stderr, "issue migrate") {
					t.Fatalf("missing migration guidance: %v %s", err, stderr)
				}
				if r.originMain() != main || testfix.Capture(t, r.origin, "branch", "--list", "issue-tracker") != "" {
					t.Fatal("refusal published state")
				}
				assertCheckoutUntouched(t, r, head, status)
			})
		}
	}
}

func TestIssueInitPresenceExcludesNonIssueArtifacts(t *testing.T) {
	r := emptyInitRepo(t)
	for _, p := range []string{"workshop/history/plans/000001-plan.md", "workshop/history/assets/000001-evidence.md", "workshop/issues/000001-work.md.side/000001-nested.md", "workshop/history/issues/000001-work.side/000001-nested.md"} {
		writeRepoFile(t, r.root, p, "artifact, not issue details\n")
	}
	r.git("add", ".")
	r.git("commit", "-qm", "non-issue artifacts")
	r.git("push", "-q")
	if _, stderr, err := executeSDLCTestCommand("issue", "init"); err != nil {
		t.Fatalf("non-issue artifacts prevented init: %v %s", err, stderr)
	}
}
