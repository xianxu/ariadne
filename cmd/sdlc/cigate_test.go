package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

// Every bucket gh reports, plus one it does not, and the empty cases. Each
// refusal row asserts its own reason, so a shared failure cannot pass a row.
func TestDecideCIGate(t *testing.T) {
	pass := prCheck{Name: "merge-check", Bucket: "pass"}
	skip := prCheck{Name: "release", Bucket: "skipping"}
	fail := prCheck{Name: "guard", Bucket: "fail"}
	cancel := prCheck{Name: "guard", Bucket: "cancel"}
	pending := prCheck{Name: "conformance", Bucket: "pending"}
	odd := prCheck{Name: "new", Bucket: "neutral"}
	for _, tc := range []struct {
		name      string
		checks    []prCheck
		workflows bool
		want      ciReason
		offenders []string
	}{
		{"all pass", []prCheck{pass}, true, ciGreen, nil},
		{"pass and skipping", []prCheck{pass, skip}, true, ciGreen, nil},
		{"one fail", []prCheck{pass, fail}, true, ciRed, []string{"guard"}},
		{"cancel is red", []prCheck{cancel}, true, ciRed, []string{"guard"}},
		{"pending", []prCheck{pass, pending}, true, ciPending, []string{"conformance"}},
		{"unknown bucket fails closed", []prCheck{pass, odd}, true, ciUnknown, []string{"new"}},
		{"red outranks pending and unknown", []prCheck{pending, odd, fail}, true, ciRed, []string{"conformance", "new", "guard"}},
		{"unknown outranks pending", []prCheck{pending, odd}, true, ciUnknown, []string{"conformance", "new"}},
		{"none yet with workflows", nil, true, ciNotRegistered, nil},
		{"none and no workflows", nil, false, ciNoWorkflows, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			v := decideCIGate(tc.checks, tc.workflows)
			if v.Reason != tc.want {
				t.Fatalf("reason %d, want %d", v.Reason, tc.want)
			}
			var names []string
			for _, c := range v.Offender {
				names = append(names, c.Name)
			}
			if !reflect.DeepEqual(names, tc.offenders) {
				t.Fatalf("offenders %v, want %v", names, tc.offenders)
			}
			if err := v.refusal(7); (err == nil) != tc.want.passes() {
				t.Fatalf("refusal %v for reason %d", err, tc.want)
			}
			if err := v.refusal(7); err != nil {
				for _, n := range tc.offenders {
					if !strings.Contains(err.Error(), n) {
						t.Fatalf("refusal does not name %s: %v", n, err)
					}
				}
			}
		})
	}
}

// exitErr makes a real *exec.ExitError with the given status, the type gh's
// non-zero exit produces.
func exitErr(t *testing.T, code int) error {
	t.Helper()
	err := exec.Command("sh", "-c", "exit "+strconv.Itoa(code)).Run()
	if err == nil {
		t.Fatal("no exit error")
	}
	return err
}

func TestParsePRChecks(t *testing.T) {
	// Captured 10-12, gh 2.93, `gh pr checks 201 -R xianxu/ariadne --json name,bucket,workflow,link`.
	realPass := `[{"bucket":"pass","link":"https://github.com/xianxu/ariadne/actions/runs/38288481117/job/114942483325","name":"merge-check","workflow":"merge-check"}]` + "\n"
	// Captured 10-12, gh 2.93, PR 2 (no checks): stderr, exit 1, empty stdout.
	realNone := "no checks reported on the '000006-a-test-issue' branch\n"
	// Constructed in the captured shape; the live check (#309) replaces these
	// with real failing and pending output.
	failing := `[{"bucket":"fail","link":"https://x/1","name":"local-gates-guard","workflow":"local-gates-guard"}]`
	pending := `[{"bucket":"pending","link":"","name":"merge-check","workflow":"merge-check"}]`

	got, err := parsePRChecks([]byte(realPass), nil, nil)
	if err != nil || len(got) != 1 || got[0].Bucket != "pass" || got[0].Name != "merge-check" {
		t.Fatalf("pass: %v %v", got, err)
	}
	got, err = parsePRChecks(nil, []byte(realNone), exitErr(t, 1))
	if err != nil || len(got) != 0 {
		t.Fatalf("no checks: %v %v", got, err)
	}
	got, err = parsePRChecks([]byte(failing), nil, exitErr(t, 1))
	if err != nil || len(got) != 1 || got[0].Bucket != "fail" {
		t.Fatalf("failing JSON with exit 1 must parse: %v %v", got, err)
	}
	got, err = parsePRChecks([]byte(pending), nil, exitErr(t, 8))
	if err != nil || len(got) != 1 || got[0].Bucket != "pending" {
		t.Fatalf("pending JSON with exit 8 must parse: %v %v", got, err)
	}
	if _, err = parsePRChecks(nil, []byte("HTTP 401: Bad credentials"), exitErr(t, 1)); err == nil {
		t.Fatal("auth failure read as no checks")
	}
	if _, err = parsePRChecks([]byte(""), nil, nil); err == nil {
		t.Fatal("empty success read as no checks")
	}
	if _, err = parsePRChecks([]byte(`[{"name":"x"}]`), nil, nil); err == nil {
		t.Fatal("a check without a bucket accepted")
	}
	if _, err = parsePRChecks([]byte(`[{`), nil, nil); err == nil {
		t.Fatal("truncated JSON accepted")
	}
}

func TestPRMergeArgsPinTheHead(t *testing.T) {
	got := prMergeArgs("o/r", "feature", "abc123")
	want := []string{"pr", "merge", "--repo", "o/r", "--merge", "--delete-branch", "--match-head-commit", "abc123", "feature"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%v", got)
	}
}

type checksGH struct {
	stubGH
	checks []prCheck
	err    error
	calls  int
	heads  []string // successive PRHeadOID answers; exhausted: "h"
}

func (g *checksGH) PRHeadOID(context.Context, string, int) (string, error) {
	if len(g.heads) == 0 {
		return "h", nil
	}
	h := g.heads[0]
	g.heads = g.heads[1:]
	return h, nil
}

func (g *checksGH) PRChecks(context.Context, string, int) ([]prCheck, error) {
	g.calls++
	return g.checks, g.err
}

func TestRunCIGate(t *testing.T) {
	withWF := func(args ...string) (string, error) { return ".github/workflows\n", nil }
	noWF := func(args ...string) (string, error) { return "", nil }
	brokenGit := func(args ...string) (string, error) { return "", errors.New("bad object") }
	ctx := context.Background()

	g := &checksGH{err: errors.New("HTTP 502")}
	if err := runCIGate(ctx, g, "o/r", 3, "h", withWF, false, io.Discard); err == nil || !strings.Contains(err.Error(), "could not read") {
		t.Fatalf("query failure: %v", err)
	}
	g = &checksGH{err: errors.New("HTTP 502")}
	var warn bytes.Buffer
	if err := runCIGate(ctx, g, "o/r", 3, "h", withWF, true, &warn); err != nil || g.calls != 0 || !strings.Contains(warn.String(), "--no-ci") {
		t.Fatalf("--no-ci: err %v calls %d warn %q", err, g.calls, warn.String())
	}
	g = &checksGH{}
	if err := runCIGate(ctx, g, "o/r", 3, "h", withWF, false, io.Discard); err == nil || !strings.Contains(err.Error(), "no checks yet") {
		t.Fatalf("unregistered: %v", err)
	}
	if err := runCIGate(ctx, g, "o/r", 3, "h", noWF, false, io.Discard); err != nil {
		t.Fatalf("no workflows: %v", err)
	}
	if err := runCIGate(ctx, g, "o/r", 3, "h", brokenGit, false, io.Discard); err == nil {
		t.Fatal("unreadable head tree passed")
	}
	// BR-1: GitHub's PR head lags the push, or moves during the read — the
	// checks belong to another commit, so green is not evidence.
	green := []prCheck{{Name: "merge-check", Bucket: "pass"}}
	for _, heads := range [][]string{{"old", "old"}, {"h", "new"}, {"old", "h"}} {
		g = &checksGH{checks: green, heads: heads}
		if err := runCIGate(ctx, g, "o/r", 3, "h", withWF, false, io.Discard); err == nil || !strings.Contains(err.Error(), "another commit") {
			t.Fatalf("heads %v: %v", heads, err)
		}
	}
	g = &checksGH{checks: green, heads: []string{"h", "h"}}
	if err := runCIGate(ctx, g, "o/r", 3, "h", withWF, false, io.Discard); err != nil {
		t.Fatalf("bound green: %v", err)
	}
	g = &checksGH{checks: []prCheck{{Name: "local-gates-guard", Bucket: "fail", Link: "https://x/9"}}}
	if err := runCIGate(ctx, g, "o/r", 3, "h", withWF, false, io.Discard); err == nil || !strings.Contains(err.Error(), "local-gates-guard") || !strings.Contains(err.Error(), "https://x/9") {
		t.Fatalf("red: %v", err)
	}
}

// The legacy path refuses red CI before PRMerge, and pins the merge to the
// pushed head. Fails if merge.go's runCIGate call alone is removed.
func TestRunMerge_RedCIRefusesPreMerge(t *testing.T) {
	tempRepo(t)
	gh := &e2eGH{openPR: "42", checks: []prCheck{{Name: "merge-check", Bucket: "fail"}}}
	swapMergeDeps(t, gh, nil)
	f := &mergeFlags{Yes: true, NoJudge: true, IssuesDir: "workshop/issues", HistoryDir: "workshop/history"}
	msg, died := expectDie(t, func() { runMerge(io.Discard, io.Discard, f) })
	if !died || !strings.Contains(msg, "CI gate") || !strings.Contains(msg, "merge-check") {
		t.Fatalf("died=%v msg=%q", died, msg)
	}
	if gh.prMergeCalls != 0 {
		t.Fatalf("PRMerge called %d times over red CI", gh.prMergeCalls)
	}
}

// BR-1 at the legacy site: GitHub still reports the pre-push head, whose
// checks are green; merge must not read them as the pushed head's.
func TestRunMerge_LaggingPRHeadRefuses(t *testing.T) {
	tempRepo(t)
	gh := &e2eGH{openPR: "42", prHeads: []string{"0123456789012345678901234567890123456789"}}
	swapMergeDeps(t, gh, nil)
	f := &mergeFlags{Yes: true, NoJudge: true, IssuesDir: "workshop/issues", HistoryDir: "workshop/history"}
	msg, died := expectDie(t, func() { runMerge(io.Discard, io.Discard, f) })
	if !died || !strings.Contains(msg, "another commit") || gh.prMergeCalls != 0 {
		t.Fatalf("died=%v calls=%d msg=%q", died, gh.prMergeCalls, msg)
	}
}

func TestRunMerge_GreenCIPinsPushedHead(t *testing.T) {
	dir := tempRepo(t)
	gh := &e2eGH{openPR: "42"}
	swapMergeDeps(t, gh, nil)
	head := git(t, dir, "rev-parse", "origin/feature")
	f := &mergeFlags{Yes: true, NoJudge: true, IssuesDir: "workshop/issues", HistoryDir: "workshop/history"}
	if msg, died := expectDie(t, func() { runMerge(io.Discard, io.Discard, f) }); died {
		t.Fatal(msg)
	}
	if gh.prMergeCalls != 1 || gh.mergedHead != head {
		t.Fatalf("calls %d, pinned %q, want %q", gh.prMergeCalls, gh.mergedHead, head)
	}
}

// The durable landing path refuses red CI before LandingMerge. Fails if
// landing.go's runCIGate call alone is removed.
func TestLandingRedCIRefusesBeforeMerge(t *testing.T) {
	_, _, gh := landingFixture(t, 1)
	gh.checks = []prCheck{{Name: "local-gates-guard", Bucket: "pending"}}
	err := runMerge(io.Discard, io.Discard, landingFlags())
	if err == nil || !strings.Contains(err.Error(), "local-gates-guard") || !strings.Contains(err.Error(), "still running") {
		t.Fatalf("err %v", err)
	}
	if gh.merges != 0 {
		t.Fatalf("merged %d times over pending CI", gh.merges)
	}
	f := landingFlags()
	f.NoCI = true
	if err := runMerge(io.Discard, io.Discard, f); err != nil || gh.merges != 1 {
		t.Fatalf("--no-ci: err %v merges %d", err, gh.merges)
	}
}
