package fleet

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

// #290: a tracked repository whose remote never answers degrades to unknown
// with the timeout as its reason, within the read deadline — it never stalls
// the inventory.
func TestHangingRemoteDegradesWithinTheDeadline(t *testing.T) {
	root := testfix.Repo(t, testfix.InitialCommit())
	hang := filepath.Join(t.TempDir(), "hang.sh")
	if err := os.WriteFile(hang, []byte("#!/bin/sh\nexec sleep 30\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	testfix.Git(t, root, "remote", "add", "origin", "ssh://tracker.invalid/repo.git")
	t.Setenv("GIT_SSH_COMMAND", hang) // the environment wins over core.sshCommand
	testfix.Git(t, root, "config", "branch.main.remote", "origin")
	testfix.Git(t, root, "config", "branch.main.merge", "refs/heads/main")
	// A migrated checkout: the cutover marker names the tracker root, so the
	// read goes to the remote (tracker.CutOver).
	manifestFile := filepath.Join(t.TempDir(), "manifest")
	if err := os.WriteFile(manifestFile, tracker.ManifestBytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	blob := strings.TrimSpace(testfix.Capture(t, root, "hash-object", "-w", manifestFile))
	mktree := exec.Command("git", "-C", root, "mktree")
	mktree.Stdin = strings.NewReader("100644 blob " + blob + "\t" + tracker.ManifestPath + "\n")
	out, err := mktree.Output()
	if err != nil {
		t.Fatal(err)
	}
	tree := strings.TrimSpace(string(out))
	head := strings.TrimSpace(testfix.Capture(t, root, "-c", "user.name=t", "-c", "user.email=t@t", "commit-tree", tree, "-m", "tracker"))
	testfix.Git(t, root, "update-ref", "refs/remotes/origin/issue-tracker", head)
	marker := filepath.Join(root, filepath.FromSlash(tracker.CutoverMarkerPath))
	if err := os.MkdirAll(filepath.Dir(marker), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(marker, tracker.CutoverMarkerBytes(head), 0o644); err != nil {
		t.Fatal(err)
	}
	testfix.Git(t, root, "add", "--", tracker.CutoverMarkerPath)
	testfix.Git(t, root, "commit", "-qm", "migrate: cutover marker")
	prev := recordsReadDeadline
	recordsReadDeadline = 500 * time.Millisecond
	t.Cleanup(func() { recordsReadDeadline = prev })

	start := time.Now()
	got := LookupRepoClaims(context.Background(), root)
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("the read took %v; the deadline is %v", elapsed, recordsReadDeadline)
	}
	if got.State != ClaimsUnknown || !strings.Contains(got.Error, "deadline") {
		t.Fatalf("hanging remote: %+v, want unknown naming the deadline", got)
	}
}
