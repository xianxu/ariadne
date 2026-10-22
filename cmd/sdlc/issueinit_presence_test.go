package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// Initialization has a presence gate, not an alternate migration importer.
// Every supported issue-history root is evidence even when the issue survives
// only outside main, and the command must retain that evidence unchanged.
func TestIssueInitPresenceAcrossRootsAndScopes(t *testing.T) {
	roots := []string{"workshop/issues", "workshop/history", "workshop/history/issues"}
	scopes := []string{"pinned-main", "local-branch", "remote-only-branch", "clean-detached-worktree", "dirty-tracked-other-worktree", "untracked-other-worktree"}
	for _, root := range roots {
		for _, scope := range scopes {
			t.Run(root+"/"+scope, func(t *testing.T) {
				r := emptyInitRepo(t)
				issuePath := root + "/000001-preserved-history.md"
				fields := "status: open\n"
				if root != "workshop/issues" {
					fields = "status: done\nactual_hours: 1\n"
				}
				content := "---\nid: 000001\n" + fields + "created: 2026-10-21\n---\n\n# Preserved issue evidence\n"
				evidenceRoot := r.root
				commitEvidence := func(checkout string) {
					writeRepoFile(t, checkout, issuePath, content)
					testfix.Git(t, checkout, "add", issuePath)
					testfix.Git(t, checkout, "commit", "-qm", "preserve legacy issue history")
				}
				switch scope {
				case "pinned-main":
					commitEvidence(r.root)
					r.git("push", "-q")
				case "local-branch":
					r.git("switch", "-qc", "presence-evidence")
					commitEvidence(r.root)
					r.git("switch", "-q", "main")
				case "remote-only-branch":
					evidenceRoot = filepath.Join(t.TempDir(), "remote-writer")
					testfix.Git(t, "", "clone", "-q", r.origin, evidenceRoot)
					testfix.Git(t, evidenceRoot, "config", "core.hooksPath", t.TempDir())
					testfix.Git(t, evidenceRoot, "switch", "-qc", "remote-presence-evidence")
					commitEvidence(evidenceRoot)
					testfix.Git(t, evidenceRoot, "push", "-qu", "origin", "remote-presence-evidence")
				case "clean-detached-worktree", "dirty-tracked-other-worktree":
					r.git("switch", "-qc", "temporary-presence-seed")
					commitEvidence(r.root)
					tip := r.git("rev-parse", "HEAD")
					r.git("switch", "-q", "main")
					evidenceRoot = filepath.Join(t.TempDir(), "detached-evidence")
					r.git("worktree", "add", "-q", "--detach", evidenceRoot, tip)
					// Only the detached checkout owns this committed evidence now. Keeping a
					// branch would allow the branch scan to mask a missing worktree scan.
					r.git("branch", "-D", "temporary-presence-seed")
					if scope == "dirty-tracked-other-worktree" {
						content += "\nUncommitted operator edit retained.\n"
						writeRepoFile(t, evidenceRoot, issuePath, content)
					}
				case "untracked-other-worktree":
					evidenceRoot = filepath.Join(t.TempDir(), "untracked-evidence")
					r.git("worktree", "add", "-q", "--detach", evidenceRoot, "HEAD")
					writeRepoFile(t, evidenceRoot, issuePath, content)
				}
				// Main remains clean even in the dirty/untracked-other-worktree cases.
				// Snapshot both checkouts' HEAD, index and working state, plus actual issue
				// bytes where they exist, so a refusal cannot silently tidy up evidence.
				beforeMain := r.originMain()
				head := r.git("rev-parse", "HEAD")
				status := r.git("status", "--porcelain", "--untracked-files=all")
				if status != "" {
					t.Fatalf("fixture dirtied calling main: %s", status)
				}
				snapshot := func(checkout string) string {
					return testfix.Capture(t, checkout, "rev-parse", "HEAD") + "\n" +
						testfix.Capture(t, checkout, "write-tree") + "\n" +
						testfix.Capture(t, checkout, "ls-files", "-s") + "\n" +
						testfix.Capture(t, checkout, "status", "--porcelain", "--untracked-files=all")
				}
				callingSnapshot := snapshot(r.root)
				evidenceSnapshot := snapshot(evidenceRoot)
				originalBytes, readErr := os.ReadFile(filepath.Join(evidenceRoot, filepath.FromSlash(issuePath)))
				if readErr != nil && !os.IsNotExist(readErr) {
					t.Fatal(readErr)
				}
				_, stderr, err := executeSDLCTestCommand("issue", "init")
				if err == nil || !strings.Contains(err.Error()+stderr, "issue migrate") {
					t.Fatalf("recognized issue evidence was not refused with migration advice: %v\n%s", err, stderr)
				}
				if r.originMain() != beforeMain {
					t.Fatal("presence refusal changed published main")
				}
				if got := strings.TrimSpace(testfix.Capture(t, r.origin, "branch", "--list", "issue-tracker")); got != "" {
					t.Fatalf("presence refusal created remote tracker: %s", got)
				}
				if got := r.git("for-each-ref", "--format=%(refname)", "refs/heads/issue-tracker", "refs/remotes/origin/issue-tracker"); got != "" {
					t.Fatalf("presence refusal created local tracker: %s", got)
				}
				assertCheckoutUntouched(t, r, head, status)
				if snapshot(r.root) != callingSnapshot || snapshot(evidenceRoot) != evidenceSnapshot {
					t.Fatal("presence refusal changed checkout or evidence index")
				}
				afterBytes, afterErr := os.ReadFile(filepath.Join(evidenceRoot, filepath.FromSlash(issuePath)))
				if (readErr == nil) != (afterErr == nil) || string(afterBytes) != string(originalBytes) {
					t.Fatalf("presence refusal changed issue evidence: %v -> %v", readErr, afterErr)
				}
			})
		}
	}
}
