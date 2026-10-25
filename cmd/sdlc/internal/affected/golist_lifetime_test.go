package affected

import (
	"context"
	"errors"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"os"
	"path/filepath"
	"testing"

	"github.com/xianxu/ariadne/pkg/testgate"
)

func TestGoListCancellationReclaimsDescendants(t *testing.T) {
	dir := t.TempDir()
	marker := filepath.Join(dir, "descendant-survived")
	pidFile := filepath.Join(dir, "descendant.pid")
	// The child inherits output pipes. Killing only its parent leaves goList
	// blocked until the child finishes, and lets it write after cancellation.
	// The child outlives the hang ceiling, so only reclaiming it lets goList
	// return, and its exit is observed rather than inferred from time (#622).
	script := "#!/bin/sh\nsh -c 'echo $$ > \"$GO_LIST_PID\"; sleep 600; printf survived > \"$GO_LIST_MARKER\"' &\nwait\n"
	if err := os.WriteFile(filepath.Join(dir, "go"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("GO_LIST_MARKER", marker)
	t.Setenv("GO_LIST_PID", pidFile)
	// Cancel on the descendant's cue, not on a timer: a fixed timeout could
	// fire before the descendant exists, under load, and prove nothing.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	returned := make(chan struct{})
	pids := make(chan int, 1)
	go func() {
		// Stop watching if goList returns first: no descendant to cue on.
		pid := testfix.AwaitPIDFile(pidFile, testgate.HangCeiling, func() bool {
			select {
			case <-returned:
				return true
			default:
				return false
			}
		})
		pids <- pid
		cancel()
	}()
	var err error
	testgate.Returns(t, "the cancelled go list", func() {
		_, err = goList(ctx, dir, "-m")
		close(returned)
	})
	if !errors.Is(err, context.Canceled) {
		t.Errorf("error = %v, want cancellation", err)
	}
	pid := <-pids
	if pid <= 0 {
		t.Fatal("the descendant never reported its pid")
	}
	if err := testfix.AwaitExit(pid, testgate.HangCeiling); err != nil {
		t.Errorf("descendant survived cancellation: %v", err)
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Errorf("descendant wrote after cancellation: marker error %v", err)
	}
}
