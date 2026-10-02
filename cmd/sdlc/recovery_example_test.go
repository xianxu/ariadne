package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/recovery"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

// exampleHarnessFlags are the only additions to the documented commands:
// switches for what a disposable fixture cannot do or does not have — the model
// judges and the estimate (--no-judge, --no-estimate, --no-estimate-recon),
// a measured actual (--actual), a separate worktree (--worktree=no) and an
// atlas (--no-atlas).
var exampleHarnessFlags = map[string][]string{
	"change-code": {"--worktree=no", "--no-judge", "--no-estimate", "--no-estimate-recon"},
	"close":       {"--actual", "1", "--no-atlas"},
}

// #280: the scheduling example in `sdlc help recovery` is executed, step by
// step, from the same data that renders it, once per legal close verdict — so
// an expectation that holds for only one variant fails (an Expect is guidance
// for every successful run, never a fixture's value). The recipient runs each
// command; the coordinator, in another checkout, checks each expectation
// against `sdlc issue show N --json`. Every convergent-retry step is delivered
// twice — a duplicated message — and must still succeed.
func TestSchedulingExampleRuns(t *testing.T) {
	for i, verdict := range []string{"SHIP", "FIX-THEN-SHIP"} {
		t.Run(verdict, func(t *testing.T) { runSchedulingExample(t, 420+i, verdict) })
	}
}

func runSchedulingExample(t *testing.T, n int, verdict string) {
	pid := fmt.Sprintf("%06d", n)
	full := fmt.Sprintf("---\nid: %s\nstatus: open\ndeps: []\ncreated: 2026-09-01\nupdated: 2026-09-01\n---\n\n# example\n\n"+
		"## Problem\n\nA gap.\n\n## Spec\n\nA thing.\n\n## Done when\n\n- it works\n\n## Plan\n\n- [x] do it\n\n## Log\n", pid)
	card, detail, err := issue.SplitCardWithFormat([]byte(full), "sha1")
	if err != nil {
		t.Fatal(err)
	}
	r := newTrackerRepo(t, map[string]string{tracker.CardPath(pid, "example"): string(card)},
		map[string]string{syncIssuesDir + "/" + pid + "-example.md": string(detail)})
	coordinator := filepath.Join(t.TempDir(), "coordinator")
	testfix.Git(t, r.root, "worktree", "add", "-q", "--detach", coordinator, "origin/main")
	stubJudge(t, "VERDICT: "+verdict+" (confidence: high)\n\nfine\n")

	observeJSON := func() map[string]any {
		t.Helper()
		t.Chdir(coordinator)
		defer t.Chdir(r.root)
		invalidateIssueRecords(context.Background())
		var out, errs bytes.Buffer
		if err := runIssueShow(context.Background(), &out, &errs, &issueShowFlags{IssuesDir: "workshop/issues", JSON: true}, itoa(n)); err != nil {
			t.Fatalf("issue show --json: %v\n%s", err, errs.String())
		}
		var doc map[string]any
		if err := json.Unmarshal(out.Bytes(), &doc); err != nil {
			t.Fatal(err)
		}
		return doc
	}

	// Before the recipient acts — the message lost or unread — the coordinator
	// sees an open card: the "otherwise" of the claim check.
	if got, _ := recovery.Lookup(observeJSON(), "card.status"); got != "open" {
		t.Fatalf("before the claim: card.status = %q", got)
	}

	// The fixture's own facts, checked where the guidance observes the
	// surrounding state — not guidance themselves: its one-box plan is the
	// quick flow, and its stubbed judge returns this variant's verdict.
	fixtureFacts := map[string][2]string{
		"checkpoints.state": {"checkpoints.flow.kind", "quick"},
		"completion.state":  {"checkpoints.reviews[close].verdict", verdict},
	}
	checked := map[string]bool{}
	commits := 0
	for i, step := range recovery.Example {
		label := fmt.Sprintf("step %d (%s: %s)", i+1, step.Actor, step.Does)
		if step.IfVerdict != "" && step.IfVerdict != verdict {
			continue
		}
		switch {
		case step.Actor == recovery.Coordinator && step.Command == "":
			// The request itself: a Couch message, outside sdlc.
		case step.Actor == recovery.Recipient && step.Command == "":
			commits++
			writeRepoFile(t, r.root, "cmd/a.go", fmt.Sprintf("package a // %d\n", commits))
			r.git("add", "cmd/a.go")
			r.git("commit", "-qm", fmt.Sprintf("#%d: %s", n, step.Does))
		case step.Actor == recovery.Recipient:
			args := exampleArgs(t, step.Command, n)
			deliveries := 1
			if c, ok := contractFor(args); ok && c.Class == recovery.ConvergentRetry {
				deliveries = 2
			}
			for d := 0; d < deliveries; d++ {
				invalidateIssueRecords(context.Background())
				if _, stderr, err := executeSDLCTestCommand(args...); err != nil {
					t.Fatalf("%s, delivery %d: %v\n%s", label, d+1, err, stderr)
				}
			}
		case step.Actor == recovery.Coordinator:
			if step.Command != "sdlc issue show N --json" {
				t.Fatalf("%s: a coordinator only observes, got %q", label, step.Command)
			}
			if step.TrackerUnreachable {
				if err := os.Rename(r.origin, r.origin+".gone"); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = os.Rename(r.origin+".gone", r.origin) })
			}
			doc := observeJSON()
			if step.TrackerUnreachable { // restored at once: a later step has a tracker
				if err := os.Rename(r.origin+".gone", r.origin); err != nil {
					t.Fatal(err)
				}
			}
			for _, e := range step.Expect {
				if f, ok := fixtureFacts[e.Path]; ok {
					checked[e.Path] = true
					if got, _ := recovery.Lookup(doc, f[0]); got != f[1] {
						t.Errorf("%s: the fixture's %s = %q, want %q", label, f[0], got, f[1])
					}
				}
				if got, ok := recovery.Lookup(doc, e.Path); !ok || got != e.Equals {
					t.Errorf("%s: %s = %q (resolved %v), want %q", label, e.Path, got, ok, e.Equals)
				}
			}
		default:
			t.Fatalf("%s: unknown actor", label)
		}
	}
	for path := range fixtureFacts {
		if !checked[path] {
			t.Errorf("no step observed %s; the fixture's facts went unchecked", path)
		}
	}
}

// exampleArgs turns a documented command into the in-process arguments: N
// becomes the issue, a placeholder becomes evidence, and the harness flags
// that stand in for model calls are appended.
func exampleArgs(t *testing.T, command string, n int) []string {
	t.Helper()
	fields := strings.Fields(strings.ReplaceAll(command, "'<evidence>'", "example-e2e"))
	if len(fields) < 2 || fields[0] != "sdlc" {
		t.Fatalf("not an sdlc command: %q", command)
	}
	args := fields[1:]
	for i, a := range args {
		if a == "N" {
			args[i] = itoa(n)
		}
	}
	return append(args, exampleHarnessFlags[args[0]]...)
}

// contractFor resolves a command's contract by its longest verb prefix
// ("issue sync --issue 4" → "issue sync"), the identity the registry uses.
func contractFor(args []string) (recovery.Contract, bool) {
	for n := len(args); n > 0; n-- {
		if c, ok := recovery.For(strings.Join(args[:n], " ")); ok {
			return c, true
		}
	}
	return recovery.Contract{}, false
}
