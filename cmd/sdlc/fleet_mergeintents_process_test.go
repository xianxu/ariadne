//go:build unix

package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
)

func TestFleetMergeIntentsCancelsActiveGitProbe(t *testing.T) {
	isolateFleetIntegrationGit(t)
	base := canonicalRoot(t.TempDir())
	root := newFleetE2ERepo(t, base, "sample", "")
	line, err := landingLineDir(execGitRunner{}, root)
	if err != nil {
		t.Fatal(err)
	}
	for _, branch := range []string{"000624-first", "000624-second"} {
		runFleetIntegrationGit(t, root, "branch", branch)
		hold, err := landingline.OpenIntents(line).Hold(landingline.Intent{Branch: branch, Head: git(t, root, "rev-parse", "HEAD"), Root: root, PID: 4242, StartedAt: time.Now().UTC()})
		if err != nil {
			t.Fatal(err)
		}
		if err := hold.Release(true); err != nil {
			t.Fatal(err)
		}
	}
	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	bin := t.TempDir()
	ready, probes := filepath.Join(bin, "ready"), filepath.Join(bin, "probes")
	script := "#!/bin/sh\ncase \"$*\" in *refs/heads/*) printf 'probe\\n' >> " + quoteHelperShell(probes) + "; printf '%s' \"$$\" > " + quoteHelperShell(ready) + "; exec sleep 60;; esac\nexec " + quoteHelperShell(realGit) + " \"$@\"\n"
	if err := os.WriteFile(filepath.Join(bin, "git"), []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	done := make(chan error, 1)
	go func() { _, err := observeFleetMergeIntents(ctx, root, ""); done <- err }()
	child := awaitLandingGitChild(t, ready)
	defer func() { _ = syscall.Kill(child, syscall.SIGKILL) }()
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("active cancellation lost: %v", err)
		}
	case <-time.After(3 * time.Second):
		_ = syscall.Kill(child, syscall.SIGKILL)
		// Release any later unsafe probe too, so the RED control cannot leak.
		_ = os.WriteFile(filepath.Join(bin, "git"), []byte("#!/bin/sh\nexit 1\n"), 0755)
		<-done
		t.Fatal("caller cancellation did not interrupt the active production Git probe")
	}
	raw, err := os.ReadFile(probes)
	if err != nil || strings.Count(string(raw), "probe\n") != 1 {
		t.Fatalf("cancelled observer started further probes: %q %v", raw, err)
	}
}
