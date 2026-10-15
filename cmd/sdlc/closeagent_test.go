package main

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// #367: the close report names the window's own commits without an Agent
// trailer, and not the stamped ones, merges, or main's commits merged in.
func TestCloseReportsCommitsWithoutAgent(t *testing.T) {
	dir := testfix.Repo(t, testfix.Chdir(), testfix.InitialCommit())
	git := func(args ...string) string { return strings.TrimSpace(testfix.Capture(t, dir, args...)) }
	base := git("rev-parse", "HEAD")
	git("switch", "-q", "-c", "000009-nine")
	git("commit", "-q", "--allow-empty", "-m", "#9: stamped", "--trailer", "Agent: claude/claude-opus-5-5/high")
	git("commit", "-q", "--allow-empty", "-m", "#9: bare")
	git("switch", "-q", "main")
	git("commit", "-q", "--allow-empty", "-m", "main: someone else's work")
	git("switch", "-q", "000009-nine")
	git("merge", "-q", "--no-ff", "--no-edit", "main")
	var stderr bytes.Buffer
	reportMissingAgent(context.Background(), &stderr, base, git("rev-parse", "HEAD"))
	out := stderr.String()
	if !strings.Contains(out, "1 commit(s) in the window lack an Agent: trailer") || !strings.Contains(out, "#9: bare") {
		t.Fatalf("report did not name the bare commit:\n%s", out)
	}
	for _, unwanted := range []string{"#9: stamped", "someone else", "Merge"} {
		if strings.Contains(out, unwanted) {
			t.Errorf("report named %q:\n%s", unwanted, out)
		}
	}
}
