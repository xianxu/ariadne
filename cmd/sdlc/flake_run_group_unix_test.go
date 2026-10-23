//go:build unix

package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/pkg/procgroup"
)

// This is the unit wrapper process under the real outer managed owner. It
// deliberately calls the production adapter, without configuring a group.
func TestFlakeRunOwnedGroupProcess(t *testing.T) {
	if os.Getenv("SDLC_UNIT_GROUP_CHILD") != "1" {
		return
	}
	if err := os.WriteFile(os.Getenv("SDLC_UNIT_WRAPPER_PID"), []byte(strconv.Itoa(os.Getpid())), 0600); err != nil {
		t.Fatal(err)
	}
	if err := runFlakeUnit(context.Background(), os.Stdout, "script:tests/owned.sh", execGitRunner{}); err != nil {
		t.Fatal(err)
	}
}

func TestFlakeRunWrapperDeathKeepsOuterOwnershipUntilJoin(t *testing.T) {
	// Both descendants close inherited diagnostic pipes before reporting ready.
	// Thus an orphan cannot keep cmd.Wait blocked instead of exercising the join.
	d, _ := typedUnitFixture(t, "script", "tests/owned.sh", `exec >/dev/null 2>&1
printf '%s' "$$" > "$SDLC_UNIT_SCRIPT_PID"
python3 -c 'import os,signal,sys,time
signal.signal(signal.SIGTERM, lambda *_: sys.exit(0))
open(sys.argv[1], "w").write(str(os.getpid()))
while True: time.sleep(.01)
' "$SDLC_UNIT_GRANDCHILD_PID" &
wait "$!"
`)
	paths := t.TempDir()
	wrapperPath, scriptPath, grandchildPath := filepath.Join(paths, "wrapper.pid"), filepath.Join(paths, "script.pid"), filepath.Join(paths, "grandchild.pid")
	t.Setenv("SDLC_UNIT_GROUP_CHILD", "1")
	t.Setenv("SDLC_UNIT_WRAPPER_PID", wrapperPath)
	t.Setenv("SDLC_UNIT_SCRIPT_PID", scriptPath)
	t.Setenv("SDLC_UNIT_GRANDCHILD_PID", grandchildPath)
	ownGroup, err := syscall.Getpgid(os.Getpid())
	if err != nil {
		t.Fatal(err)
	}
	lifetime := &runnerArtifactLifetime{retirable: true}
	d.StrictGroupCleanup, d.runnerLifetime = true, lifetime
	var output bytes.Buffer
	d.Stderr = &output
	joinEntered, releaseJoin := make(chan struct{}), make(chan struct{})
	var releaseOnce sync.Once
	release := func() { releaseOnce.Do(func() { close(releaseJoin) }) }
	d.runnerCleanup = func(cmd *exec.Cmd) func() error {
		realJoin := configureRunnerCleanup(cmd)
		return func() error {
			close(joinEntered)
			<-releaseJoin
			return realJoin()
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), testfix.HangCeiling)
	type result struct{ run, cleanup error }
	done := make(chan result, 1)
	go func() {
		run, cleanup := runManagedCommandBatch(ctx, d, d.Root, suiteCheckName, [][]string{{os.Args[0], "-test.run=^TestFlakeRunOwnedGroupProcess$"}}, os.Environ(), &output, "")
		done <- result{run, cleanup}
	}()
	finished := false
	wrapper, script, grandchild := 0, 0, 0
	t.Cleanup(func() {
		release()
		cancel()
		// Every PID came from this private fixture. A failed same-group assertion
		// must still clean an accidentally detached child, never the test's group.
		for _, pid := range []int{grandchild, script, wrapper} {
			if pid > 0 {
				if group, err := syscall.Getpgid(pid); err == nil && group != ownGroup {
					_ = syscall.Kill(pid, syscall.SIGKILL)
				}
			}
		}
		if !finished {
			select {
			case <-done:
			case <-time.After(testfix.HangCeiling):
				t.Error("managed owner did not finish fixture cleanup")
			}
		}
	})
	stop := func() bool {
		select {
		case r := <-done:
			finished = true
			t.Fatalf("managed wrapper exited before descendant readiness: run %v; cleanup %v\n%s", r.run, r.cleanup, output.String())
			return true
		default:
			return false
		}
	}
	for _, target := range []struct {
		path string
		pid  *int
	}{{wrapperPath, &wrapper}, {scriptPath, &script}, {grandchildPath, &grandchild}} {
		*target.pid = testfix.AwaitPIDFile(target.path, testfix.HangCeiling, stop)
		if *target.pid == 0 {
			t.Fatalf("fixture did not report %s", filepath.Base(target.path))
		}
	}
	if wrapper == ownGroup {
		t.Fatal("fixture wrapper would own the test process group")
	}
	for _, pid := range []int{wrapper, script, grandchild} {
		group, err := syscall.Getpgid(pid)
		if err != nil || group != wrapper {
			t.Fatalf("pid %d group = %d, %v; want one outer group %d", pid, group, err, wrapper)
		}
	}
	// Kill precisely the adapter, leaving its script and grandchild for the
	// enclosing owner. A nested managed group would have failed the check above.
	if err := syscall.Kill(wrapper, syscall.SIGKILL); err != nil {
		t.Fatal(err)
	}
	select {
	case <-joinEntered:
	case r := <-done:
		finished = true
		t.Fatalf("wrapper death skipped group join: %+v\n%s", r, output.String())
	case <-time.After(testfix.HangCeiling):
		t.Fatal("wrapper death never reached the group-join barrier")
	}
	// Closing joinEntered synchronizes the owner's lifetime write; the join is
	// held at this exact boundary, so this does not infer ownership from timing.
	if lifetime.mayRetire() {
		t.Fatal("wrapper death released inputs before descendant join")
	}
	if err := procgroup.Kill(wrapper, 0); err != nil {
		t.Fatalf("descendants vanished before controlled join: %v", err)
	}
	release()
	select {
	case r := <-done:
		finished = true
		if r.run == nil || r.cleanup != nil {
			t.Fatalf("wrapper death result: run %v; cleanup %v\n%s", r.run, r.cleanup, output.String())
		}
	case <-time.After(testfix.HangCeiling):
		t.Fatal("outer owner did not finish joining descendants")
	}
	if !lifetime.mayRetire() {
		t.Fatal("confirmed group completion did not release inputs")
	}
	if err := procgroup.Kill(wrapper, 0); !errors.Is(err, os.ErrProcessDone) {
		t.Fatalf("owned group remains runnable after join: %v", err)
	}
	for _, pid := range []int{wrapper, script, grandchild} {
		if err := testfix.AwaitExit(pid, time.Second); err != nil {
			t.Error(fmt.Errorf("owned descendant survived join: %w", err))
		}
	}
}
