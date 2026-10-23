//go:build unix

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
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

// Exercise the actual nested helper/provider boundary, not a same-group sleep
// substituted for the inspector. A detached provider defeats parent-only tests.
func TestSuiteQueueRealInspectorOwnsProviderTree(t *testing.T) {
	for _, mode := range []string{"cancel", "deadline", "completion", "provider-timeout"} {
		t.Run(mode, func(t *testing.T) {
			source, err := os.ReadFile("../../scripts/test-lock.py")
			if err != nil {
				t.Fatal(err)
			}
			root := suiteQueueHelperFixture(t, string(source))
			private := t.TempDir()
			base := filepath.Join(private, "lock")
			directory := base + ".tickets"
			if err := os.Mkdir(directory, 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(directory, "alloc"), nil, 0644); err != nil {
				t.Fatal(err)
			}
			common := filepath.Join(private, "repo.git")
			raw, err := json.Marshal(map[string]any{"repo": "fixture", "worktree": root, "pid": os.Getpid(), "started": "fixture", "subject": map[string]any{"common_dir": common, "issue": 1}})
			if err != nil {
				t.Fatal(err)
			}
			ticket, err := os.OpenFile(filepath.Join(directory, "000000000001"), os.O_CREATE|os.O_RDWR, 0644)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = ticket.Close() })
			if _, err := ticket.Write(raw); err != nil {
				t.Fatal(err)
			}
			if err := syscall.Flock(int(ticket.Fd()), syscall.LOCK_EX); err != nil {
				t.Fatal(err)
			}
			parentFile, childFile := filepath.Join(private, "provider.pid"), filepath.Join(private, "child.pid")
			provider := filepath.Join(private, "provider")
			finish := "time.sleep(60)"
			if mode == "completion" {
				finish = "print(json.dumps({'version':1,'sources':[{'common_dir':" + quoteHelperShell(common) + ",'deadlines':{},'refresh_at':0}]}))"
			}
			script := "#!/usr/bin/env python3\nimport os, subprocess, time, json\np = subprocess.Popen(['sleep', '60'], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)\nopen(" + quoteHelperShell(childFile) + ", 'w').write(str(p.pid))\nopen(" + quoteHelperShell(parentFile) + ", 'w').write(str(os.getpid()))\n" + finish + "\n"
			if err := os.WriteFile(provider, []byte(script), 0755); err != nil {
				t.Fatal(err)
			}
			t.Setenv("WF_TEST_LOCK_FILE", base)
			t.Setenv("WF_TEST_PRIORITY_PROVIDER", provider)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			expire := cancel
			if mode == "deadline" {
				ctx, expire = testfix.DeadlineOnCue(ctx)
			}
			result := make(chan suiteQueueObservation, 1)
			finished := make(chan struct{})
			go func() { defer close(finished); result <- observeSuiteQueue(ctx, root) }()
			t.Cleanup(func() {
				cancel()
				select {
				case <-finished:
				case <-time.After(3 * time.Second):
					t.Error("real inspector did not join")
				}
			})
			parent, child := awaitLandingGitChild(t, parentFile), awaitLandingGitChild(t, childFile)
			if mode == "cancel" || mode == "deadline" {
				expire()
			}
			var got suiteQueueObservation
			select {
			case got = <-result:
			case <-time.After(3 * time.Second):
				t.Fatal("real inspector exceeded cancellation/provider drain bound")
			}
			if mode == "completion" && !got.Available {
				t.Errorf("successful inspector: %+v", got)
			}
			if mode != "completion" && got.Available {
				t.Errorf("interrupted inspector advertised order: %+v", got)
			}
			for _, pid := range []int{parent, child} {
				if err := testfix.AwaitExit(pid, time.Second); err != nil {
					t.Errorf("inspector left provider tree running: %v", err)
				}
			}
		})
	}
}
