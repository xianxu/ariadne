package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

// initBarrierEffects exercises the production Git adapter, inserting one
// deterministic boundary without replacing Git or its publication protocol.
type initBarrierEffects struct {
	real          gitInitializationEffects
	before, after func(string) error
	calls         []string
}

func (b *initBarrierEffects) enter(name string) error {
	b.calls = append(b.calls, name)
	if b.before != nil {
		return b.before(name)
	}
	return nil
}
func (b *initBarrierEffects) leave(name string, err error) error {
	if err != nil {
		return err
	}
	if b.after != nil {
		return b.after(name)
	}
	return nil
}
func (b *initBarrierEffects) Inspect(ctx context.Context) (tracker.InitInspected, error) {
	if err := b.enter("inspect"); err != nil {
		return tracker.InitInspected{}, err
	}
	observed, err := b.real.Inspect(ctx)
	return observed, b.leave("inspect", err)
}
func (b *initBarrierEffects) Bootstrap(ctx context.Context) (string, error) {
	if err := b.enter("bootstrap"); err != nil {
		return "", err
	}
	root, err := b.real.Bootstrap(ctx)
	return root, b.leave("bootstrap", err)
}
func (b *initBarrierEffects) Publish(ctx context.Context, root string) (string, error) {
	if err := b.enter("publish"); err != nil {
		return "", err
	}
	commit, err := b.real.Publish(ctx, root)
	return commit, b.leave("publish", err)
}
func (b *initBarrierEffects) FastForward(ctx context.Context, commit string) error {
	if err := b.enter("checkout"); err != nil {
		return err
	}
	return b.leave("checkout", b.real.FastForward(ctx, commit))
}

func assertInitConverged(t *testing.T, r *trackerRepo, retainedRoot string) {
	t.Helper()
	root, present, err := tracker.ReadCutoverMarker(r.root)
	if err != nil || !present {
		t.Fatalf("checkout has no confirmed cutover: present=%v error=%v", present, err)
	}
	if retainedRoot != "" && root != retainedRoot {
		t.Fatalf("recreated tracker root: got %s want %s", root, retainedRoot)
	}
	if actual := strings.TrimSpace(testfix.Capture(t, r.origin, "rev-list", "--max-parents=0", "issue-tracker")); actual != root {
		t.Fatalf("remote root %s differs from checkout %s", actual, root)
	}
	if r.git("rev-parse", "HEAD") != r.originMain() {
		t.Fatal("checkout did not reach published cutover")
	}
	if got := r.git("status", "--porcelain", "--untracked-files=all"); got != "" {
		t.Fatalf("dirty completed checkout: %s", got)
	}
	if _, err := os.Stat(filepath.Join(r.root, "workshop/issues/.gitkeep")); err != nil {
		t.Fatalf("initial scaffold lost: %v", err)
	}
}

func TestIssueInitRecoveryEveryEffectBoundary(t *testing.T) {
	sequence := []string{"inspect", "bootstrap", "publish", "checkout"}
	for index, effect := range sequence {
		for _, after := range []bool{false, true} {
			timing := "before"
			if after {
				timing = "after"
			}
			t.Run(effect+"/"+timing, func(t *testing.T) {
				r := emptyInitRepo(t)
				originalMain := r.originMain()
				barrier := &initBarrierEffects{}
				injected := errors.New("simulated crash or lost acknowledgment")
				fail := func(stage string) error {
					if stage == effect {
						return injected
					}
					return nil
				}
				if after {
					barrier.after = fail
				} else {
					barrier.before = fail
				}
				state, err := runInitialization(context.Background(), barrier)
				if !errors.Is(err, injected) || !state.Terminal() {
					t.Fatalf("injection failed: state %#v error %v", state, err)
				}
				if !reflect.DeepEqual(barrier.calls, sequence[:index+1]) {
					t.Fatalf("effects continued after uncertainty: %v", barrier.calls)
				}
				retainedRoot := ""
				trackerExists := index > 1 || (index == 1 && after)
				if trackerExists {
					retainedRoot = strings.TrimSpace(testfix.Capture(t, r.origin, "rev-list", "--max-parents=0", "issue-tracker"))
				} else if got := testfix.Capture(t, r.origin, "branch", "--list", "issue-tracker"); got != "" {
					t.Fatalf("unexpected tracker publication: %s", got)
				}
				mainPublished := index > 2 || (index == 2 && after)
				if (r.originMain() != originalMain) != mainPublished {
					t.Fatal("main publication crossed wrong boundary")
				}
				if _, err = runInitialization(context.Background(), &gitInitializationEffects{}); err != nil {
					t.Fatalf("fresh invocation retry: %v", err)
				}
				assertInitConverged(t, r, retainedRoot)
				completedMain := r.originMain()
				if _, err = runInitialization(context.Background(), &gitInitializationEffects{}); err != nil {
					t.Fatalf("completed retry: %v", err)
				}
				if r.originMain() != completedMain {
					t.Fatal("completed retry republished main")
				}
			})
		}
	}
}

func TestIssueInitRecoveryRetainsDirtyCheckout(t *testing.T) {
	r := emptyInitRepo(t)
	originalHead := r.git("rev-parse", "HEAD")
	dirty := filepath.Join(r.root, "operator-notes.txt")
	barrier := &initBarrierEffects{before: func(stage string) error {
		if stage == "checkout" {
			return os.WriteFile(dirty, []byte("preserve operator work\n"), 0600)
		}
		return nil
	}}
	state, err := runInitialization(context.Background(), barrier)
	if err == nil || !strings.Contains(err.Error(), "clean") || state.Root() == "" || state.Commit() == "" {
		t.Fatalf("dirty boundary: %#v %v", state, err)
	}
	if r.git("rev-parse", "HEAD") != originalHead {
		t.Fatal("dirty checkout advanced")
	}
	if got, err := os.ReadFile(dirty); err != nil || string(got) != "preserve operator work\n" {
		t.Fatalf("operator work changed: %q %v", got, err)
	}
	if _, err = runInitialization(context.Background(), &gitInitializationEffects{}); err == nil {
		t.Fatal("retry ignored dirty work")
	}
	if err = os.Remove(dirty); err != nil {
		t.Fatal(err)
	} // operator explicitly resolves the prerequisite
	if _, err = runInitialization(context.Background(), &gitInitializationEffects{}); err != nil {
		t.Fatal(err)
	}
	assertInitConverged(t, r, state.Root())
}

func TestIssueInitRecoveryCancellationStopsNextEffect(t *testing.T) {
	for _, effect := range []string{"inspect", "bootstrap", "publish"} {
		t.Run(effect, func(t *testing.T) {
			r := emptyInitRepo(t)
			beforeMain := r.originMain()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			barrier := &initBarrierEffects{after: func(stage string) error {
				if stage == effect {
					cancel()
				}
				return nil
			}}
			state, err := runInitialization(ctx, barrier)
			if !errors.Is(err, context.Canceled) || !state.Terminal() {
				t.Fatalf("cancellation: %#v %v", state, err)
			}
			wantCalls := map[string][]string{"inspect": {"inspect"}, "bootstrap": {"inspect", "bootstrap"}, "publish": {"inspect", "bootstrap", "publish"}}[effect]
			if !reflect.DeepEqual(barrier.calls, wantCalls) {
				t.Fatalf("effects continued after cancellation: %v", barrier.calls)
			}
			if effect != "publish" && r.originMain() != beforeMain {
				t.Fatal("cancellation allowed main publication")
			}
			if effect == "bootstrap" && state.Root() == "" {
				t.Fatal("confirmed bootstrap identity lost on cancellation")
			}
			if effect == "publish" && state.Commit() == "" {
				t.Fatal("confirmed cutover identity lost on cancellation")
			}
			if _, err = runInitialization(context.Background(), &gitInitializationEffects{}); err != nil {
				t.Fatal(err)
			}
			assertInitConverged(t, r, state.Root())
		})
	}
}

func TestIssueInitRecoveryRefusesHistoryAndBranchInventory(t *testing.T) {
	for _, branchOnly := range []bool{false, true} {
		t.Run(map[bool]string{false: "history", true: "branch-only"}[branchOnly], func(t *testing.T) {
			r := emptyInitRepo(t)
			issuePath := "workshop/history/issues/000001-old.md"
			if branchOnly {
				r.git("switch", "-qc", "legacy-work")
				issuePath = "workshop/issues/000001-old.md"
			}
			writeRepoFile(t, r.root, issuePath, "---\nid: 000001\nstatus: open\ncreated: 2026-10-21\n---\n\n# Existing work\n")
			r.git("add", ".")
			r.git("commit", "-qm", "legacy evidence")
			if branchOnly {
				r.git("switch", "-q", "main")
			} else {
				r.git("push", "-q")
			}
			before := r.originMain()
			head := r.git("rev-parse", "HEAD")
			_, err := runInitialization(context.Background(), &gitInitializationEffects{})
			if err == nil || !strings.Contains(err.Error(), "issue migrate") {
				t.Fatalf("legacy inventory not refused: %v", err)
			}
			if r.originMain() != before || r.git("rev-parse", "HEAD") != head || testfix.Capture(t, r.origin, "branch", "--list", "issue-tracker") != "" {
				t.Fatal("refused inventory mutated refs")
			}
		})
	}
}

func TestIssueInitRecoveryRefusesMalformedOrMismatchedMarker(t *testing.T) {
	for _, mismatch := range []bool{false, true} {
		t.Run(map[bool]string{false: "malformed", true: "mismatched"}[mismatch], func(t *testing.T) {
			r := emptyInitRepo(t)
			if _, err := runInitialization(context.Background(), &gitInitializationEffects{}); err != nil {
				t.Fatal(err)
			}
			trackerTip := strings.TrimSpace(testfix.Capture(t, r.origin, "rev-parse", "issue-tracker"))
			raw := []byte("not a tracker marker\n")
			if mismatch {
				raw = tracker.CutoverMarkerBytes(strings.Repeat("a", 40))
			}
			writeRepoFile(t, r.root, tracker.CutoverMarkerPath, string(raw))
			r.git("add", ".")
			r.git("commit", "-qm", "inconsistent cutover evidence")
			r.git("push", "-q")
			before := r.originMain()
			head := r.git("rev-parse", "HEAD")
			if _, err := runInitialization(context.Background(), &gitInitializationEffects{}); err == nil {
				t.Fatal("inconsistent marker accepted")
			}
			if r.originMain() != before || r.git("rev-parse", "HEAD") != head || strings.TrimSpace(testfix.Capture(t, r.origin, "rev-parse", "issue-tracker")) != trackerTip {
				t.Fatal("inconsistent evidence was rewritten")
			}
		})
	}
}

func TestIssueInitRecoveryConcurrentClonePublication(t *testing.T) {
	r := emptyInitRepo(t)
	clone := filepath.Join(t.TempDir(), "second-clone")
	testfix.Git(t, "", "clone", "-q", r.origin, clone)
	testfix.Git(t, clone, "config", "core.hooksPath", t.TempDir())
	var secondRoot, secondMain string
	barrier := &initBarrierEffects{after: func(stage string) error {
		if stage != "inspect" {
			return nil
		}
		if err := os.Chdir(clone); err != nil {
			return err
		}
		state, err := runInitialization(context.Background(), &gitInitializationEffects{})
		restoreErr := os.Chdir(r.root)
		if err != nil {
			return err
		}
		if restoreErr != nil {
			return restoreErr
		}
		secondRoot = state.Root()
		secondMain = r.originMain()
		return nil
	}}
	// Both clones observed the original empty main before the first clone's
	// publication effects. A losing invocation may stop, but cannot overwrite
	// the winner or manufacture a different tracker root.
	_, _ = runInitialization(context.Background(), barrier)
	if secondRoot == "" {
		t.Fatal("competing initializer never completed")
	}
	if got := strings.TrimSpace(testfix.Capture(t, r.origin, "rev-list", "--max-parents=0", "issue-tracker")); got != secondRoot {
		t.Fatalf("winner tracker root overwritten: %s", got)
	}
	if r.originMain() != secondMain {
		t.Fatal("winner main publication overwritten")
	}
	if _, err := runInitialization(context.Background(), &gitInitializationEffects{}); err != nil {
		t.Fatalf("losing clone retry: %v", err)
	}
	assertInitConverged(t, r, secondRoot)
}
