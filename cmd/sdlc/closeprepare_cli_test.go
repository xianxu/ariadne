package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/repolock"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

// Both barriers are real executable boundaries. In particular, the first stops
// the expensive churn read, not a test-only callback which could bypass the lock.
// An observed lock-wait diagnostic is the regression signal; elapsed time is
// only the enclosing test's safety bound.
func TestClosePreparationCLIAllowsSiblingBookkeeping(t *testing.T) {
	binary := buildFleetE2EBinary(t)
	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	for _, verb := range []string{"close", "milestone-close"} {
		t.Run(verb, func(t *testing.T) {
			plan := "- [x] implement\n"
			if verb == "milestone-close" {
				plan = "- [x] M1 — implement\n- [ ] M2 — finish\n"
			}
			r, cardPath, _ := closeReadyWithPlan(t, 617, plan)
			if _, stderr, err := executeSDLCTestCommand("issue", "new", "Unrelated", "--slug", "unrelated", "--no-claim"); err != nil {
				t.Fatalf("seed unrelated card: %v\n%s", err, stderr)
			}
			sibling := filepath.Join(t.TempDir(), "sibling")
			r.git("worktree", "add", "-q", "-b", "sibling-bookkeeping", sibling, "origin/main")
			barriers := t.TempDir()
			agents := t.TempDir()
			write := func(name, body string) {
				t.Helper()
				if err := os.WriteFile(filepath.Join(agents, name), []byte(body), 0755); err != nil {
					t.Fatal(err)
				}
			}
			// A private fence stub makes this an explicit checkout-only fixture; no
			// installed model binary or live broker can be reached.
			for _, name := range []string{"sandbox-exec", "bwrap"} {
				write(name, "#!/bin/sh\nexit 1\n")
			}
			write("git", `#!/bin/sh
if [ "$1" = log ] && [ "$2" = --numstat ] && [ "$3" = --format= ] && [ ! -f "$CLOSE_PREP_BARRIER/prepared" ]; then
 echo preparing > "$CLOSE_PREP_BARRIER/preparing"
 n=0
 while [ ! -f "$CLOSE_PREP_BARRIER/prepare-release" ]; do
  n=$((n+1)); [ "$n" -lt "$CLOSE_PREP_SPINS" ] || exit 98
  sleep 0.01
 done
 echo prepared > "$CLOSE_PREP_BARRIER/prepared"
fi
exec "$CLOSE_PREP_GIT" "$@"
`)
			write("claude", `#!/bin/sh
: "${CLOSE_PREP_BARRIER:?}" "${CLOSE_PREP_SPINS:?}"
echo reviewing > "$CLOSE_PREP_BARRIER/reviewing"
n=0
while [ ! -f "$CLOSE_PREP_BARRIER/review-release" ]; do
 n=$((n+1)); [ "$n" -lt "$CLOSE_PREP_SPINS" ] || exit 98
 sleep 0.01
done
printf 'VERDICT: SHIP (confidence: high)\n\nPrivate executable review completed.\n\140\140\140findings\nfindings: []\n\140\140\140\n'
`)
			ctx, cancel := context.WithTimeout(context.Background(), 2*testfix.HangCeiling)
			if deadline, ok := t.Deadline(); ok {
				var stop context.CancelFunc
				ctx, stop = context.WithDeadline(ctx, deadline)
				t.Cleanup(stop)
			}
			t.Cleanup(cancel)
			env := append(os.Environ(), "COUCH_TREE=", "PATH="+agents+":"+os.Getenv("PATH"), "CLOSE_PREP_GIT="+realGit, "CLOSE_PREP_BARRIER="+barriers,
				fmt.Sprintf("CLOSE_PREP_SPINS=%d", 2*testfix.HangCeiling/(10*time.Millisecond)), "WF_REVIEW_CHECKOUT_ROOT="+t.TempDir())
			args := []string{verb, "--issue", "617", "--no-actual", "--verified", "private executable lock conformance; measurement not applicable", "--no-atlas", "--no-project", "--agent", "claude"}
			if verb == "milestone-close" {
				args = append(args, "--milestone", "M1")
			}
			cmd := exec.CommandContext(ctx, binary, args...)
			cmd.Dir, cmd.Env, cmd.WaitDelay = r.root, env, 2*time.Second
			output := testfix.NewWatch(repolock.WaitMessagePrefix)
			cmd.Stdout, cmd.Stderr = output, output
			if err := cmd.Start(); err != nil {
				t.Fatal(err)
			}
			done := make(chan error, 1)
			go func() { done <- cmd.Wait() }()
			joined := false
			release := func(name string) {
				t.Helper()
				if err := os.WriteFile(filepath.Join(barriers, name), nil, 0600); err != nil {
					t.Fatal(err)
				}
			}
			t.Cleanup(func() {
				// Release descendants before canceling/reaping their owned sdlc parent.
				_ = os.WriteFile(filepath.Join(barriers, "prepare-release"), nil, 0600)
				_ = os.WriteFile(filepath.Join(barriers, "review-release"), nil, 0600)
				cancel()
				if !joined {
					<-done
				}
			})
			barrier := func(name string) {
				t.Helper()
				ticker := time.NewTicker(10 * time.Millisecond)
				defer ticker.Stop()
				for {
					if _, err := os.Stat(filepath.Join(barriers, name)); err == nil {
						return
					}
					select {
					case err := <-done:
						joined = true
						t.Fatalf("%s ended before %s: %v\n%s", verb, name, err, output)
					case <-ctx.Done():
						t.Fatalf("%s barrier: %v\n%s", name, ctx.Err(), output)
					case <-ticker.C:
					}
				}
			}
			bookkeeping := func(title string) {
				t.Helper()
				childCtx, childCancel := context.WithCancel(ctx)
				defer childCancel()
				other := exec.CommandContext(childCtx, binary, "issue", "set-title", title, "--issue", "618")
				other.Dir, other.Env, other.WaitDelay = sibling, env, 2*time.Second
				waits := testfix.NewWatch(repolock.WaitMessagePrefix)
				other.Stdout, other.Stderr = waits, waits
				if err := other.Start(); err != nil {
					t.Fatal(err)
				}
				ended := make(chan error, 1)
				go func() { ended <- other.Wait() }()
				select {
				case err := <-ended:
					if err != nil {
						t.Fatalf("bookkeeping during %s: %v\n%s", title, err, waits)
					}
				case <-waits.Seen():
					childCancel()
					<-ended
					t.Fatalf("bookkeeping blocked during %s:\n%s", title, waits)
				case <-ctx.Done():
					childCancel()
					<-ended
					t.Fatalf("bookkeeping safety bound: %v\n%s", ctx.Err(), waits)
				}
			}
			barrier("preparing")
			bookkeeping("During preparation")
			release("prepare-release")
			barrier("reviewing")
			bookkeeping("During review")
			release("review-release")
			select {
			case err := <-done:
				joined = true
				if err != nil {
					t.Fatalf("%s finalization: %v\n%s", verb, err, output)
				}
			case <-ctx.Done():
				t.Fatalf("%s finalization safety bound: %v\n%s", verb, ctx.Err(), output)
			}
			if !strings.Contains(r.card(tracker.CardPath("000618", "unrelated")), "# During review") {
				t.Fatal("close overwrote independent bookkeeping")
			}
			if verb == "close" {
				if !strings.Contains(r.card(cardPath), "status: codecomplete") {
					t.Fatalf("close did not finalize:\n%s", r.card(cardPath))
				}
			} else if msg := r.git("log", "-2", "--format=%B"); !strings.Contains(msg, "Review-Verdict: SHIP") || !strings.Contains(msg, "M1") {
				t.Fatalf("milestone did not finalize:\n%s", msg)
			}
		})
	}
}
