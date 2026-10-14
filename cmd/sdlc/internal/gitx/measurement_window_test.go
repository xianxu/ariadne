package gitx

import (
	"context"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func TestCommitWindowAtIgnoresCallerAndLaterHead(t *testing.T) {
	a := testfix.Repo(t, testfix.InitialCommit(), testfix.Chdir())
	testfix.Git(t, a, "commit", "--allow-empty", "-qm", "#335 unrelated caller")
	b := testfix.Repo(t, testfix.InitialCommit())
	testfix.Git(t, b, "commit", "--allow-empty", "-qm", "#335 pinned work")
	pin := strings.TrimSpace(testfix.Git(t, b, "rev-parse", "HEAD"))
	testfix.Git(t, b, "commit", "--allow-empty", "-qm", "#335 later bookkeeping")
	first, _, _, err := CommitWindowAt(context.Background(), b, pin, "335")
	if err != nil || first != pin {
		t.Fatalf("window followed caller/later HEAD: %q %v, want %s", first, err, pin)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, _, err := CommitWindowAt(ctx, b, pin, "335"); err == nil {
		t.Fatal("cancelled pinned read treated as no window")
	}
}
