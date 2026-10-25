package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/queuegate"
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

func TestReportMergeIntent(t *testing.T) {
	in := landingline.Intent{Version: 1, Branch: "000565-x", Issue: 565, Head: strings.Repeat("c", 40), Root: "/w/slot5", PID: 77,
		StartedAt: time.Date(2026, 10, 23, 9, 0, 0, 0, time.UTC)}
	o := landingline.ObservedIntent{Intent: in, Path: "/l/intents/k"}
	orphan := reportMergeIntent(o, true, nil)
	if orphan.State != "orphaned" || orphan.Reentry != "sdlc merge --branch 000565-x --yes" || orphan.Note != "" {
		t.Fatalf("%+v", orphan)
	}
	line := formatMergeIntent(orphan)
	for _, want := range []string{`orphaned merge: #565 branch "000565-x" at cccccccc`, "pid 77, which is gone", `re-enter in "/w/slot5": sdlc merge --branch 000565-x --yes`} {
		if !strings.Contains(line, want) {
			t.Errorf("missing %q in %q", want, line)
		}
	}
	paused := reportMergeIntent(o, true, []queuegate.PauseOwner{{Owner: "ops:0", Reason: "full shutdown"}})
	if !strings.Contains(formatMergeIntent(paused), `the machine queue is paused by "ops:0" ("full shutdown"); re-enter after it resumes`) {
		t.Fatalf("no pause note: %q", formatMergeIntent(paused))
	}
	o.Locked = true
	if live := reportMergeIntent(o, true, nil); live.State != "live" || live.Reentry != "" || formatMergeIntent(live) != "" {
		t.Fatalf("a live intent: %+v", live)
	}
	o.Locked = false
	if stale := reportMergeIntent(o, false, nil); stale.State != "stale" || stale.Reentry != "" || formatMergeIntent(stale) != "" {
		t.Fatalf("a stale intent: %+v", stale)
	}
	o.Intent.Root = "/w/evil\nforged line"
	if line := formatMergeIntent(reportMergeIntent(o, true, nil)); strings.Contains(line, "\nforged") {
		t.Fatalf("an untrusted root broke the line: %q", line)
	}
	bad := reportMergeIntent(landingline.ObservedIntent{Path: "/l/intents/z", Err: errors.New("intent: not one JSON object")}, false, nil)
	if bad.State != "unreadable" || bad.Reentry != "" || !strings.Contains(formatMergeIntent(bad), `unreadable merge intent "/l/intents/z"`) {
		t.Fatalf("an unreadable intent: %+v", bad)
	}
}

// #565 Done-when 1: after an interrupted merge, landing status and the
// slot's sdlc state name the orphan and the exact re-entry command.
func TestOrphanedMergeOnStatusSurfaces(t *testing.T) {
	roots, _, _ := landingFixture(t, 1)
	t.Setenv("WF_SDLC_QUEUE_ROOT", t.TempDir())
	failFirst(t, "switch --no-overwrite-ignore", context.Canceled)
	if err := runMerge(io.Discard, io.Discard, landingFlags()); !errors.Is(err, context.Canceled) {
		t.Fatalf("want the interruption: %v", err)
	}
	want := fmt.Sprintf("re-enter in %q: sdlc merge --branch %s --yes", canonicalRoot(roots[1]), landingTestBranch)
	var status bytes.Buffer
	if err := runLandingStatus(t.Context(), &status, roots[1], defaultLandingDeps()); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(status.String(), "orphaned merge: #1 branch \""+landingTestBranch+"\"") || !strings.Contains(status.String(), want) {
		t.Fatalf("landing status:\n%s", status.String())
	}
	t.Chdir(roots[1])
	var prose, raw bytes.Buffer
	if err := runState(context.Background(), &prose, &stateFlags{IssuesDir: "workshop/issues", HistoryDir: "workshop/history"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(prose.String(), want) {
		t.Fatalf("sdlc state:\n%s", prose.String())
	}
	if err := runState(context.Background(), &raw, &stateFlags{JSON: true, IssuesDir: "workshop/issues", HistoryDir: "workshop/history"}); err != nil {
		t.Fatal(err)
	}
	var s struct {
		MergeIntents []mergeIntentReport `json:"merge_intents"`
	}
	if err := json.Unmarshal(raw.Bytes(), &s); err != nil || len(s.MergeIntents) != 1 || s.MergeIntents[0].State != "orphaned" ||
		s.MergeIntents[0].Reentry != "sdlc merge --branch "+landingTestBranch+" --yes" || s.MergeIntents[0].Root != canonicalRoot(roots[1]) {
		t.Fatalf("state --json merge_intents: %+v %v\n%s", s.MergeIntents, err, raw.String())
	}
}

// #565 Done-when 2: re-entering an interrupted merge reuses the suite that
// passed on its unchanged head (the record is a GitHub commit status, so it
// outlives the runner); a head that changed meanwhile runs it again.
func TestMergeReentryReusesThePassedSuite(t *testing.T) {
	for _, moved := range []bool{false, true} {
		t.Run(map[bool]string{false: "unchanged head", true: "main moved"}[moved], func(t *testing.T) {
			root, server, gh := integrationFixture(t)
			var tested []string
			ctx, shutdown := context.WithCancel(context.Background())
			gh.afterStatus = func(sha string) {
				tested = append(tested, sha)
				shutdown() // the shutdown lands just after the suite recorded its pass
			}
			first := landingFlags()
			first.Context = ctx
			if err := runMerge(io.Discard, io.Discard, first); !errors.Is(err, context.Canceled) || gh.merges != 0 {
				t.Fatalf("want an interruption before the merge: %v (merges %d)", err, gh.merges)
			}
			if got := landingIntents(t, root); len(got) != 1 || got[0].Locked {
				t.Fatalf("no orphan: %+v", got)
			}
			if moved {
				integrationAdvance(t, server, "moving.go", "change\n")
			}
			f := landingFlags()
			f.Branch = landingTestBranch
			if err := runMerge(io.Discard, io.Discard, f); err != nil {
				t.Fatal(err)
			}
			want := map[bool]int{false: 1, true: 2}[moved]
			if gh.merges != 1 || len(tested) != want {
				t.Fatalf("merges=%d suite runs=%d (%v), want 1 and %d", gh.merges, len(tested), tested, want)
			}
			if got := landingIntents(t, root); len(got) != 0 {
				t.Fatalf("re-entry left %+v", got)
			}
		})
	}
}
