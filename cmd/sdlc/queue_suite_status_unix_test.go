//go:build unix

package main

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func TestSuiteQueueObservationReapsDescendants(t *testing.T) {
	for _, deadline := range []bool{false, true} {
		t.Run(fmt.Sprint(deadline), func(t *testing.T) {
			pidfile := filepath.Join(t.TempDir(), "child.pid")
			script := "import subprocess, time\np = subprocess.Popen(['sleep', '60'])\nopen(" + quoteHelperShell(pidfile) + ", 'w').write(str(p.pid))\ntime.sleep(60)\n"
			root := suiteQueueHelperFixture(t, script)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			expire := cancel
			if deadline {
				ctx, expire = testfix.DeadlineOnCue(ctx)
			}
			done := make(chan suiteQueueObservation, 1)
			finished := make(chan struct{})
			go func() { defer close(finished); done <- observeSuiteQueue(ctx, root) }()
			t.Cleanup(func() {
				cancel()
				select {
				case <-finished:
				case <-time.After(3 * time.Second):
					t.Error("inspection fixture did not join after cancellation")
				}
			})
			child := awaitLandingGitChild(t, pidfile)
			expire()
			select {
			case got := <-done:
				expected := "canceled"
				if deadline {
					expected = "deadline exceeded"
				}
				if got.Available || !strings.Contains(got.Reason, expected) {
					t.Fatalf("cancel observation: %+v", got)
				}
			case <-time.After(3 * time.Second):
				t.Fatal("suite observation exceeded cancellation/drain bound")
			}
			if err := testfix.AwaitExit(child, time.Second); err != nil {
				t.Fatalf("inspection descendant survived: %v", err)
			}
		})
	}
}

func TestSuiteQueueObservationReapsSuccessfulParentsChild(t *testing.T) {
	pidfile := filepath.Join(t.TempDir(), "child.pid")
	script := "import subprocess\np = subprocess.Popen(['sleep', '60'], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)\nopen(" + quoteHelperShell(pidfile) + ", 'w').write(str(p.pid))\nprint(" + quoteHelperShell(suiteQueueFixture) + ")\n"
	root := suiteQueueHelperFixture(t, script)
	got := observeSuiteQueue(context.Background(), root)
	if !got.Available {
		t.Fatalf("successful helper: %+v", got)
	}
	child := awaitLandingGitChild(t, pidfile)
	if err := testfix.AwaitExit(child, time.Second); err != nil {
		t.Fatalf("successful helper left descendant: %v", err)
	}
}
