package main

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// Both callers must observe the publication remote's current branch heads,
// even when the calling clone's cached refs and fetch configuration omit the
// only issue evidence. The second clone models a quiescent legacy writer that
// finished before the command starts; this does not claim a cross-clone lock.
func TestIssueInitAndMigrateObserveCurrentRemoteInventory(t *testing.T) {
	for _, scenario := range []string{"new-remote-branch", "advanced-cached-branch", "narrowed-fetch"} {
		for _, command := range []string{"init", "migrate"} {
			t.Run(scenario+"/"+command, func(t *testing.T) {
				r := emptyInitRepo(t)
				other := filepath.Join(t.TempDir(), "legacy-writer")
				testfix.Git(t, "", "clone", "-q", r.origin, other)
				testfix.Git(t, other, "config", "core.hooksPath", t.TempDir())
				const branch = "remote-only-legacy-work"
				const issuePath = "workshop/issues/000001-remote-only.md"
				testfix.Git(t, other, "switch", "-qc", branch)
				if scenario == "advanced-cached-branch" {
					testfix.Git(t, other, "push", "-qu", "origin", branch)
					r.git("fetch", "-q", "origin")
					if got := r.git("rev-parse", "refs/remotes/origin/"+branch); got != r.git("rev-parse", "HEAD") {
						t.Fatal("cached branch is not the known empty baseline")
					}
				} else {
					// Establish the caller's last fetch before the remote branch exists.
					r.git("fetch", "-q", "origin")
				}
				if scenario == "narrowed-fetch" {
					r.git("config", "--replace-all", "remote.origin.fetch", "+refs/heads/main:refs/remotes/origin/main")
				}
				writeRepoFile(t, other, issuePath, "---\nid: 000001\nstatus: open\ncreated: 2026-10-21\n---\n\n# Unsynced remote issue\n")
				testfix.Git(t, other, "add", issuePath)
				testfix.Git(t, other, "commit", "-qm", "legacy work not published to main")
				testfix.Git(t, other, "push", "-qu", "origin", branch)
				remoteTip := strings.TrimSpace(testfix.Capture(t, r.origin, "rev-parse", "refs/heads/"+branch))
				cached := r.git("for-each-ref", "--format=%(objectname)", "refs/remotes/origin/"+branch)
				if cached == remoteTip {
					t.Fatal("fixture accidentally refreshed the caller's cached issue evidence")
				}
				beforeMain := r.originMain()
				beforeHead := r.git("rev-parse", "HEAD")
				beforeStatus := r.git("status", "--porcelain", "--untracked-files=all")
				beforeIndex := r.git("write-tree")
				beforeFiles := r.git("ls-files", "-s")
				stdout, stderr, err := executeSDLCTestCommand("issue", command)
				diagnostic := stdout + stderr
				if err != nil {
					diagnostic += "\n" + err.Error()
				}
				if err == nil {
					t.Fatalf("%s ignored %s remote issue inventory:\n%s", command, scenario, diagnostic)
				}
				if command == "init" {
					if !strings.Contains(diagnostic, "issue migrate") {
						t.Fatalf("init did not name reviewed migration:\n%s", diagnostic)
					}
				} else {
					for _, want := range []string{issuePath, branch, "issue exists only on a branch"} {
						if !strings.Contains(diagnostic, want) {
							t.Errorf("migration report omits %q:\n%s", want, diagnostic)
						}
					}
				}
				if got := r.originMain(); got != beforeMain {
					t.Fatalf("refusal changed remote main: %s -> %s", beforeMain, got)
				}
				if got := strings.TrimSpace(testfix.Capture(t, r.origin, "rev-parse", "refs/heads/"+branch)); got != remoteTip {
					t.Fatalf("refusal changed legacy branch: %s -> %s", remoteTip, got)
				}
				if got := strings.TrimSpace(testfix.Capture(t, r.origin, "branch", "--list", "issue-tracker")); got != "" {
					t.Fatalf("refusal created tracker: %s", got)
				}
				assertCheckoutUntouched(t, r, beforeHead, beforeStatus)
				if r.git("write-tree") != beforeIndex || r.git("ls-files", "-s") != beforeFiles {
					t.Fatal("refusal rewrote checkout index")
				}
			})
		}
	}
}
