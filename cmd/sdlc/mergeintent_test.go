package main

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
)

func landingIntents(t *testing.T, root string) []landingline.ObservedIntent {
	t.Helper()
	dir, err := landingLineDir(mergeRunner, root)
	if err != nil {
		t.Fatal(err)
	}
	got, err := landingline.OpenIntents(dir).List()
	if err != nil {
		t.Fatal(err)
	}
	return got
}

// failFirst makes the first git call whose argv contains arg fail with cause.
func failFirst(t *testing.T, arg string, cause error) *bool {
	t.Helper()
	original := mergeRunner
	t.Cleanup(func() { mergeRunner = original })
	hit := false
	mergeRunner = landingHookRunner{gitRunner: original, before: func(_ string, args []string) error {
		if !hit && strings.Contains(strings.Join(args, " "), arg) {
			hit = true
			return cause
		}
		return nil
	}}
	return &hit
}

// #565: a merge holds its intent while it runs. A terminal outcome removes
// it; an interruption (a shutdown's cancel) leaves an orphan whose printed
// command re-enters the landing once.
func TestMergeIntentLifecycle(t *testing.T) {
	t.Run("a landed merge leaves none", func(t *testing.T) {
		roots, _, _ := landingFixture(t, 1)
		if err := runMerge(io.Discard, io.Discard, landingFlags()); err != nil {
			t.Fatal(err)
		}
		if got := landingIntents(t, roots[1]); len(got) != 0 {
			t.Fatalf("left %+v", got)
		}
	})
	t.Run("a refused merge leaves none", func(t *testing.T) {
		roots, _, _ := landingFixture(t, 1)
		hit := failFirst(t, "switch --no-overwrite-ignore", errors.New("injected refusal"))
		if err := runMerge(io.Discard, io.Discard, landingFlags()); err == nil || !*hit {
			t.Fatalf("want the injected refusal: %v", err)
		}
		if got := landingIntents(t, roots[1]); len(got) != 0 {
			t.Fatalf("left %+v", got)
		}
	})
	t.Run("an interrupted merge leaves an orphan that re-entry lands", func(t *testing.T) {
		roots, _, gh := landingFixture(t, 1)
		head := git(t, roots[1], "rev-parse", "refs/heads/"+landingTestBranch)
		hit := failFirst(t, "switch --no-overwrite-ignore", context.Canceled)
		if err := runMerge(io.Discard, io.Discard, landingFlags()); !errors.Is(err, context.Canceled) || !*hit {
			t.Fatalf("want the interruption: %v", err)
		}
		got := landingIntents(t, roots[1])
		if len(got) != 1 || got[0].Locked || got[0].Err != nil || got[0].Intent.Branch != landingTestBranch || got[0].Intent.Head != head ||
			got[0].Intent.Root != canonicalRoot(roots[1]) || got[0].Intent.PID != os.Getpid() {
			t.Fatalf("orphan: %+v", got)
		}
		f := landingFlags()
		f.Branch = landingTestBranch
		if err := runMerge(io.Discard, io.Discard, f); err != nil {
			t.Fatal(err)
		}
		if gh.merges != 1 { // merged before the interruption; re-entry only finishes cleanup
			t.Fatalf("re-entry merged %d times", gh.merges)
		}
		if got := landingIntents(t, roots[1]); len(got) != 0 {
			t.Fatalf("re-entry left %+v", got)
		}
	})
	t.Run("a second merge of a live intent refuses, naming its pid", func(t *testing.T) {
		roots, _, gh := landingFixture(t, 1)
		dir, err := landingLineDir(mergeRunner, roots[1])
		if err != nil {
			t.Fatal(err)
		}
		live := landingline.Intent{Branch: landingTestBranch, Head: git(t, roots[1], "rev-parse", "refs/heads/"+landingTestBranch),
			Root: canonicalRoot(roots[1]), PID: 4242, StartedAt: time.Now().UTC()}
		hold, err := landingline.OpenIntents(dir).Hold(live)
		if err != nil {
			t.Fatal(err)
		}
		defer hold.Release(false)
		err = runMerge(io.Discard, io.Discard, landingFlags())
		var held *landingline.IntentHeldError
		if !errors.As(err, &held) || !strings.Contains(err.Error(), "pid 4242") || gh.merges != 0 {
			t.Fatalf("want a refusal naming pid 4242, got %v (merges %d)", err, gh.merges)
		}
	})
}
