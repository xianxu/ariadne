//go:build manual

// planreplay_test.go — the regression test for "did the tuning weaken review?" (#187
// Done-when). Manual-tagged: it spends real agent latency, so CI does not run it.
//
//	# one round, fresh ledger:
//	REPLAY_DIR=$TMPDIR/replay go test ./cmd/sdlc -tags manual -run TestReplayPair127 -v -timeout 30m
//	# next round, same ledger, edited plan:
//	REPLAY_DIR=$TMPDIR/replay REPLAY_ISSUE=/path/to/edited.md go test ... -run TestReplayPair127 -v
//
// It drives runPlanQualityJudge DIRECTLY, not runChangeCode. runChangeCode iterates
// changeCodeGates and calls exitWithCode(1) on any gate failure (changecode.go → term.go →
// os.Exit), so it never RETURNS a gate failure — an `err := runChangeCode(…)` loop would
// os.Exit(1) the test process on round 1, precisely the round whose blocking is the point.
// (expectDie doesn't help: it swaps the `die` var, which this path bypasses.)
// runPlanQualityJudge owns the entire ledger path — read → dispatch → parse → assign ids →
// Decide → write — and returns an error, so rounds stay countable in-process.
//
// Because it drives the plan gate directly, the estimate gates are not in the loop at all
// and need no --no-estimate. Do NOT "seed a reconciling ## Estimate block" instead:
// runEstimateQualityJudge skips silently only when the block is ABSENT, so seeding one
// would dispatch a second real agent per round whose failure would end the run for a
// reason unrelated to the plan gate.
//
// The harness deliberately does NOT simulate a plan author. Rounds 2+ require a real plan
// edit responding to round 1's findings; that is the session's work, and faking it would
// prove nothing about convergence. The harness's job is to make each round reproducible
// and the ledger inspectable.
package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
)

const replayIssueFile = "000900-replay.md"

// replayDir is the ledger directory. It comes from the environment rather than t.TempDir()
// because rounds must ACCUMULATE across invocations — a fresh temp dir per round would
// reset the ledger and turn every round into round 1, measuring nothing.
func replayDir(t *testing.T) string {
	t.Helper()
	dir := os.Getenv("REPLAY_DIR")
	if dir == "" {
		t.Skip("set REPLAY_DIR to a persistent directory (the ledger must survive between rounds)")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

func readReplayIssue(t *testing.T) string {
	t.Helper()
	path := os.Getenv("REPLAY_ISSUE")
	if path == "" {
		path = filepath.Join("testdata", "pair127-round1-issue.md")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read replay issue %s: %v", path, err)
	}
	t.Logf("replay input: %s (%d bytes)", path, len(b))
	return string(b)
}

// logLedger dumps the ledger state after a round: every round's findings and dispositions,
// plus the open set. This output IS the evidence — the four Done-when checks are answered
// by reading it, so it prints verbatim titles and details rather than counts alone.
func logLedger(t *testing.T, dir string) gatestate.Ledger {
	t.Helper()
	l, err := readPlanGateLedger(dir, replayIssueFile, 900)
	if err != nil {
		t.Fatalf("ledger: %v", err)
	}
	t.Logf("── ledger: %d round(s), content_hash=%q", len(l.Rounds), l.ContentHash)
	for _, r := range l.Rounds {
		t.Logf("── round %d (agent %s) blocked=%v forced=%q protocol_error=%q",
			r.N, r.Agent, r.Blocked, r.Forced, r.ProtocolError)
		for _, d := range r.Dispositions {
			t.Logf("   dispose %s → %s  %s", d.ID, d.State, d.Note)
		}
		for _, f := range r.New {
			t.Logf("   [%s] %s: %s", f.Severity, f.ID, f.Title)
			for _, line := range strings.Split(strings.TrimRight(f.Detail, "\n"), "\n") {
				t.Logf("       | %s", line)
			}
		}
	}
	open := gatestate.OpenFindings(l)
	t.Logf("── open findings: %d", len(open))
	for _, f := range open {
		t.Logf("   OPEN [%s] %s: %s", f.Severity, f.ID, f.Title)
	}
	counts, openN := gatestate.DispositionCounts(l)
	t.Logf("── dispositions: %v, open %d", counts, openN)
	return l
}

// TestReplayPair127 runs ONE round of the tuned plan gate against pair#127's plan as it
// stood at its first change-code invocation. Re-run it with REPLAY_ISSUE pointing at an
// edited plan to drive the next round against the same ledger.
//
// Baseline to beat: 6 invocations / 5 rejections. It is not an assertion here — the round
// count is what the run MEASURES, and a hard bound would either pass vacuously or fail the
// build for a judge that had a fair point.
func TestReplayPair127(t *testing.T) {
	dir := replayDir(t)
	f := &changeCodeFlags{Issue: 900, PlansDir: dir}

	err := runPlanQualityJudge(os.Stdout, os.Stderr, f, "000900-replay", replayIssueFile,
		readReplayIssue(t), "")
	t.Logf("round outcome: err=%v", err)

	l := logLedger(t, dir)
	if len(l.Rounds) == 0 {
		t.Fatal("no round was persisted — the gate cannot converge on a round it did not record")
	}
	last := l.Rounds[len(l.Rounds)-1]
	if last.ProtocolError != "" {
		t.Errorf("round %d was a PROTOCOL MISS (%s) — the judge emitted no valid findings block. "+
			"Per Risk 5 this replay is the live conformance check for the ```findings fence, so "+
			"this is a real failure, not a flake.", last.N, last.ProtocolError)
	}
}

// TestC1TestSurfaceShape is C1's semantic check, and it needs BOTH halves. Done-when says a
// plan naming test *functions* + strategy PASSES the gate, and a plan enumerating ~15 prose
// test cases DRAWS a finding. A prompt that rejected every shape of test description would
// satisfy the negative half alone and ship green, so the positive control is what makes the
// check mean anything.
//
// The two variants are the same plan differing ONLY in how the test work is described —
// that is what isolates shape from substance. Each runs on its own fresh ledger dir.
//
// Why synthetic rather than pair#127's own text: its ~15 prose test cases lived in an
// INTERMEDIATE plan state that was never committed (git holds the 4-case pre-gate revision
// and the final strategy-form one, nothing between). Recorded in the evidence file as a
// deviation; C1 is a claim about shape, so a controlled pair of shapes tests it more
// sharply than an uncontrolled historical artifact would.
func TestC1TestSurfaceShape(t *testing.T) {
	root := replayDir(t)
	for _, tc := range []struct {
		name        string
		plan        string
		wantFinding bool
	}{
		{"prose-enumeration", c1ProsePlan, true},
		{"strategy-lines", c1StrategyPlan, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := filepath.Join(root, "c1-"+tc.name)
			if err := os.MkdirAll(dir, 0o755); err != nil {
				t.Fatal(err)
			}
			f := &changeCodeFlags{Issue: 900, PlansDir: dir}
			err := runPlanQualityJudge(os.Stdout, os.Stderr, f, "000900-replay", replayIssueFile,
				c1IssueShell+tc.plan, "")
			t.Logf("%s: err=%v", tc.name, err)
			l := logLedger(t, dir)

			hit, which := testSurfaceFinding(l)
			t.Logf("%s: test-surface finding present=%v (%s)", tc.name, hit, which)
			if hit != tc.wantFinding {
				t.Errorf("%s: test-surface finding present=%v, want %v — %s", tc.name, hit, tc.wantFinding,
					map[bool]string{
						true:  "the prompt no longer tells an enumerated plan to compress (C1 did not land)",
						false: "the prompt rejects the strategy-line shape it is supposed to ACCEPT, so the negative half proves nothing",
					}[tc.wantFinding])
			}
		})
	}
}

// testSurfaceFinding reports whether any finding is about how the plan describes its TESTS.
// Keyword matching is a heuristic and the run logs every finding verbatim so the evidence
// record rests on reading, not on this function — but it has to be mechanical to be a gate.
func testSurfaceFinding(l gatestate.Ledger) (bool, string) {
	keys := []string{"enumerat", "compress", "test case", "test-case", "prose test", "one test per",
		"list of tests", "test list", "strategy"}
	for _, r := range l.Rounds {
		for _, f := range r.New {
			hay := strings.ToLower(f.Title + " " + f.Detail)
			if !strings.Contains(hay, "test") {
				continue
			}
			for _, k := range keys {
				if strings.Contains(hay, k) {
					return true, f.ID + " " + f.Title
				}
			}
		}
	}
	return false, "none"
}

// c1IssueShell is a deliberately SOUND issue — a real defect, a real root cause, a real
// done-when — so that any finding the judge raises is about the plan's test description
// rather than about the problem being underspecified. Derived from pair#127's actual M1.
const c1IssueShell = `---
id: 000900
status: working
---

# SGR mouse release kills the pane's keyboard

## Problem

An SGR (1006) mouse event is ` + "`\\x1b[<button;col;row`" + ` plus a terminator: ` + "`M`" + ` = press,
` + "`m`" + ` = RELEASE. Both ` + "`parseSGRMousePressPrefix`" + ` and ` + "`isSGRMousePrefix`" + ` search only
for ` + "`'M'`" + `, so a release matches "sequence not finished yet" and is parked in
` + "`pumpStdin`" + `'s ` + "`held`" + ` buffer. ` + "`held`" + ` is prepended to the next read, which re-matches
the same way — so the release and every keystroke typed after it accumulate and never
reach the child.

## Spec

Recognize both SGR terminators in the prefix parser and the partial-sequence test, carry
the press/release distinction on the event, and forward releases to the child. A release
is never a wheel tick (the wheel reports press-only), so it must not fall into the scroll
branch.

## Done when

- A click-drag-release leaves the pane's keyboard live.
- ` + "`go test ./cmd/internal/termcmd/`" + ` green.

`

// c1ProsePlan enumerates its test work case by case — the shape C1 says must draw a
// finding telling it to compress.
const c1ProsePlan = `## Plan

- [ ] Recognize ` + "`m`" + ` as a terminator in ` + "`parseSGRMousePressPrefix`" + `
- [ ] Recognize ` + "`m`" + ` as a terminator in ` + "`isSGRMousePrefix`" + `
- [ ] Carry ` + "`Release bool`" + ` on the mouse event; skip the scroll branch when set
- [ ] Tests:
      - a lone press is forwarded
      - a lone release is forwarded
      - a press then a release in one read are both forwarded
      - a press then a release in two reads are both forwarded
      - a keystroke typed after a release is not swallowed
      - two keystrokes after a release are both delivered
      - a release with a payload following it in the same read
      - a wheel-up tick still routes to the scroll branch
      - a wheel-down tick still routes to the scroll branch
      - a release does NOT route to the scroll branch
      - a truncated sequence with no terminator is still held
      - a sequence split across three reads is reassembled
      - a malformed button field is not treated as a mouse event
      - an empty read does not clear ` + "`held`" + `
      - a non-mouse escape sequence passes through untouched
`

// c1StrategyPlan describes the same test work by naming the functions under test plus one
// adversarial strategy line each — the shape Done-when says must PASS.
const c1StrategyPlan = `## Plan

- [ ] Recognize ` + "`m`" + ` as a terminator in ` + "`parseSGRMousePressPrefix`" + ` and
      ` + "`isSGRMousePrefix`" + `; carry ` + "`Release bool`" + ` on the event and keep releases out of
      the scroll branch.
- [ ] Tests on the existing ` + "`fakeMux`" + ` / ` + "`splitReader`" + ` harness:
      - ` + "`parseSGRMousePressPrefix`" + ` / ` + "`isSGRMousePrefix`" + `: one table over the
        press/release × complete/truncated × split-read cross product, driven from a
        generated case list rather than hand-written cases, so a new terminator cannot be
        added without a row.
      - ` + "`pumpStdin`" + `: the invariant is "nothing typed after a recognized event is ever
        held" — assert it by feeding a release followed by N random keystrokes and checking
        the mux received every byte, which fails for any terminator the parser misses
        rather than only for the ones someone thought to enumerate.
      - Scroll routing: a derived negative suite asserting the scroll branch sees exactly
        the wheel forms and nothing else, so a release silently reclassified as a wheel
        tick fails here rather than in a live session.
`
