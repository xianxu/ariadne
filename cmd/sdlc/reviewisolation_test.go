package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/reviewbox"
)

// unwrapReviewArgv strips the #204 isolation prefix a boundary review's argv
// carries — `env -C <dir>`, optionally under `sandbox-exec -p <profile>` — and
// returns the agent, its own args, and the checkout it runs in.
func unwrapReviewArgv(t *testing.T, name string, args []string) (agent string, agentArgs []string, dir string) {
	t.Helper()
	if name == "sandbox-exec" {
		if len(args) < 3 || args[0] != "-p" {
			t.Fatalf("sandbox-exec argv without a profile: %q", args)
		}
		name, args = args[2], args[3:]
	}
	if name != "env" || len(args) < 3 || args[0] != "-C" {
		t.Fatalf("boundary review argv is not run in a checkout: %s %q", name, args)
	}
	return args[2], args[3:], args[1]
}

// boundaryWindow commits a change under review in a closeRepo fixture and
// returns the params of its milestone boundary.
func boundaryWindow(t *testing.T) boundaryReviewParams {
	t.Helper()
	issuesDir := closeRepo(t, 69)
	base := strings.TrimSpace(captureGit(t, "rev-parse", "HEAD"))
	if err := os.WriteFile("m1.go", []byte("package main // work\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, "", "add", "m1.go")
	git(t, "", "commit", "-q", "-m", "#69 M1: the work under review")
	head := strings.TrimSpace(captureGit(t, "rev-parse", "HEAD"))
	return boundaryReviewParams{Label: "#69 M1", IssuesDir: issuesDir, IssueNum: 69, Milestone: "M1",
		Base: shortSHA(base), BaseLong: base, Head: head}
}

// sourceState is what a reviewer must not change in the source checkout.
func sourceState(t *testing.T) string {
	t.Helper()
	config, err := os.ReadFile(filepath.Join(".git", "config"))
	if err != nil {
		t.Fatal(err)
	}
	return captureGit(t, "status", "--porcelain", "--untracked-files=all") + captureGit(t, "for-each-ref") + string(config)
}

func reviewRootEntries(t *testing.T) []os.DirEntry {
	t.Helper()
	entries, err := os.ReadDir(os.Getenv("WF_REVIEW_CHECKOUT_ROOT"))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		t.Fatal(err)
	}
	return entries
}

const shipReview = "VERDICT: SHIP (confidence: high)\n\nLooks good.\n" + emptyBoundaryFindings

func TestBoundaryReviewRunsInDisposableCheckout(t *testing.T) {
	t.Setenv("WF_REVIEW_CHECKOUT_ROOT", t.TempDir())
	p := boundaryWindow(t)
	issueFile := filepath.Join(p.IssuesDir, "000069-x.md")
	live := "---\nid: 000069\n---\n\n# x\n\n## Log\n\nuncommitted log line\n"
	if err := os.WriteFile(issueFile, []byte(live), 0o644); err != nil {
		t.Fatal(err)
	}
	before := sourceState(t)

	var dir, prompt string
	orig := judge.Run
	t.Cleanup(func() { judge.Run = orig })
	judge.Run = func(ctx context.Context, onStart func(pid int), name string, args ...string) (judge.ProcessOutput, error) {
		var agentArgs []string
		_, agentArgs, dir = unwrapReviewArgv(t, name, args)
		prompt = agentArgs[len(agentArgs)-1]
		if got, err := os.ReadFile(filepath.Join(dir, issueFile)); err != nil || string(got) != live {
			t.Errorf("checkout issue file = %q (%v), want the live text", got, err)
		}
		// The reviewer mutates freely inside its checkout.
		os.WriteFile(filepath.Join(dir, "m1.go"), []byte("package main // reverted fix\n"), 0o644)
		os.WriteFile(filepath.Join(dir, "residue.bin"), []byte("built"), 0o644)
		for _, args := range [][]string{{"config", "x.y", "z"}, {"branch", "evil"}} {
			if out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput(); err != nil {
				t.Errorf("reviewer git %v: %v %s", args, err, out)
			}
		}
		return judge.ProcessOutput{Stdout: []byte(shipReview)}, nil
	}

	var stdout, stderr bytes.Buffer
	res := dispatchBoundaryReview(&stdout, &stderr, p)
	if res.Verdict != judge.VerdictShip {
		t.Fatalf("verdict = %s (%s)\n%s", res.Verdict, res.Reason, stderr.String())
	}
	if root, _ := filepath.EvalSymlinks(os.Getenv("WF_REVIEW_CHECKOUT_ROOT")); filepath.Dir(dir) != root {
		t.Fatalf("reviewer ran in %q, not a checkout under the review root", dir)
	}
	if !strings.Contains(prompt, "repository root: "+dir) || !strings.Contains(prompt, "git -C "+dir+" diff") {
		t.Errorf("the prompt does not point the reviewer at its checkout %s", dir)
	}
	if !strings.Contains(prompt, "you run in a disposable checkout") {
		t.Error("an isolated dispatch must tell the reviewer it may revert in place")
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Errorf("the checkout outlived the review: %v", err)
	}
	if entries := reviewRootEntries(t); len(entries) != 0 {
		t.Errorf("the review left %d entries under the review root", len(entries))
	}
	if after := sourceState(t); after != before {
		t.Errorf("the reviewer's activity reached the source checkout:\nbefore %s\nafter %s", before, after)
	}
	if res.Isolation != reviewbox.LevelCheckout || !strings.Contains(stderr.String(), "isolation: checkout") {
		t.Errorf("isolation = %q; stderr lacks the checkout-only warning:\n%s", res.Isolation, stderr.String())
	}
}

func TestBoundaryReviewCheckoutIsRemovedOnFailureAndCancellation(t *testing.T) {
	t.Setenv("WF_REVIEW_CHECKOUT_ROOT", t.TempDir())
	p := boundaryWindow(t)
	orig := judge.Run
	t.Cleanup(func() { judge.Run = orig })

	cases := map[string]func(context.CancelFunc) (judge.ProcessOutput, error){
		"reviewer failure": func(context.CancelFunc) (judge.ProcessOutput, error) {
			return judge.ProcessOutput{}, errors.New("exec: claude: not found")
		},
		"interruption": func(cancel context.CancelFunc) (judge.ProcessOutput, error) {
			cancel() // SIGINT/SIGTERM reach dispatch as this cancellation (main.go)
			return judge.ProcessOutput{}, context.Canceled
		},
	}
	for name, outcome := range cases {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			var dir string
			judge.Run = func(_ context.Context, _ func(int), n string, args ...string) (judge.ProcessOutput, error) {
				_, _, dir = unwrapReviewArgv(t, n, args)
				os.WriteFile(filepath.Join(dir, "residue"), []byte("x"), 0o644)
				return outcome(cancel)
			}
			q := p
			q.Context = ctx
			res := dispatchBoundaryReview(&bytes.Buffer{}, &bytes.Buffer{}, q)
			if res.DispatchError == nil || res.Verdict != judge.VerdictNotRun {
				t.Fatalf("verdict = %s, err %v; want a not-run review", res.Verdict, res.DispatchError)
			}
			if dir == "" {
				t.Fatal("the reviewer never ran")
			}
			if entries := reviewRootEntries(t); len(entries) != 0 {
				t.Errorf("%s left %d entries under the review root", name, len(entries))
			}
		})
	}
}

func TestBoundaryReviewRoundsGetFreshCheckouts(t *testing.T) {
	t.Setenv("WF_REVIEW_CHECKOUT_ROOT", t.TempDir())
	p := boundaryWindow(t)
	orig := judge.Run
	t.Cleanup(func() { judge.Run = orig })
	var dirs []string
	judge.Run = func(_ context.Context, _ func(int), n string, args ...string) (judge.ProcessOutput, error) {
		_, _, dir := unwrapReviewArgv(t, n, args)
		if _, err := os.Stat(filepath.Join(dir, "residue")); !os.IsNotExist(err) {
			t.Errorf("round %d found residue from an earlier round", len(dirs)+1)
		}
		os.WriteFile(filepath.Join(dir, "residue"), []byte("x"), 0o644)
		dirs = append(dirs, dir)
		return judge.ProcessOutput{Stdout: []byte(shipReview)}, nil
	}
	dispatchBoundaryReview(&bytes.Buffer{}, &bytes.Buffer{}, p)
	dispatchBoundaryReview(&bytes.Buffer{}, &bytes.Buffer{}, p)
	if len(dirs) != 2 || dirs[0] == dirs[1] {
		t.Fatalf("rounds ran in %q; want two distinct checkouts", dirs)
	}
}

// Under a fence the argv runs the agent inside sandbox-exec with the source
// root and git common dir write-denied, and the agent's own sandbox off.
func TestBoundaryReviewFencedArgvAndStamp(t *testing.T) {
	pinReviewAgent(t, judge.AgentClaude) // asserts Claude's argv
	t.Setenv("WF_REVIEW_CHECKOUT_ROOT", t.TempDir())
	p := boundaryWindow(t)
	p.PlansDir = t.TempDir()
	orig, origProbe := judge.Run, reviewbox.Probe
	t.Cleanup(func() { judge.Run, reviewbox.Probe = orig, origProbe })
	reviewbox.Probe = func() (reviewbox.Fence, string) { return reviewbox.FenceSeatbelt, "" }
	root, _ := filepath.EvalSymlinks(strings.TrimSpace(captureGit(t, "rev-parse", "--show-toplevel")))
	// A sibling worktree (another slot, or the operator's :0) is protected too.
	sibling := filepath.Join(t.TempDir(), "sibling")
	git(t, "", "worktree", "add", "-q", "--detach", sibling)
	sibling, _ = filepath.EvalSymlinks(sibling)

	var name string
	var args []string
	(&fakeReviewer{}).install(t)
	inner := judge.Run
	judge.Run = func(ctx context.Context, onStart func(int), n string, a ...string) (judge.ProcessOutput, error) {
		name, args = n, a
		return inner(ctx, onStart, n, a...)
	}
	var errb bytes.Buffer
	res := dispatchBoundaryReview(&bytes.Buffer{}, &errb, p)
	if name != "sandbox-exec" {
		t.Fatalf("fenced review argv starts with %q, want sandbox-exec", name)
	}
	profile := args[1]
	for _, protected := range []string{root, filepath.Join(root, ".git"), sibling} {
		if !strings.Contains(profile, `(subpath "`+protected+`")`) {
			t.Errorf("profile does not write-deny %s: %s", protected, profile)
		}
	}
	agent, agentArgs, _ := unwrapReviewArgv(t, name, args)
	if agent != "claude" || !strings.Contains(strings.Join(agentArgs, " "), `--settings {"sandbox":{"enabled":false}}`) {
		t.Errorf("fenced claude keeps its own sandbox: %s %q", agent, agentArgs)
	}
	if res.Isolation != reviewbox.LevelFenced || strings.Contains(errb.String(), "isolation: checkout") {
		t.Errorf("isolation = %q, stderr:\n%s", res.Isolation, errb.String())
	}

	persistBoundaryRound(&errb, p, res, "2026-10-11T12:00:00-07:00")
	l, err := readBoundaryGateLedger(p.PlansDir, "000069-x.md", 69)
	if err != nil || len(l.Rounds) != 1 || l.Rounds[0].Isolation != reviewbox.LevelFenced {
		t.Fatalf("ledger round isolation: %v %+v", err, l.Rounds)
	}
	sidecar, err := os.ReadFile(res.SidecarPath)
	if err != nil || !strings.Contains(string(sidecar), "| isolation | fenced |") {
		t.Errorf("sidecar lacks the isolation row (%v):\n%s", err, sidecar)
	}
}

// --dry-run shows the isolated command and creates nothing.
func TestBoundaryReviewDryRunShowsIsolation(t *testing.T) {
	t.Setenv("WF_REVIEW_CHECKOUT_ROOT", t.TempDir())
	p := boundaryWindow(t)
	var out bytes.Buffer
	if err := printBoundaryReviewDryRun(&out, &bytes.Buffer{}, p); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "env -C ") || !strings.Contains(out.String(), "<review-checkout>") {
		t.Errorf("dry-run line does not show the checkout:\n%s", out.String())
	}
	if entries := reviewRootEntries(t); len(entries) != 0 {
		t.Errorf("dry-run created %d entries", len(entries))
	}
}

// #204 close review BR-2: the no-verdict retry's second attempt runs in a reset
// checkout, not on the tree attempt 1 left mutated.
func TestBoundaryReviewRetryStartsFromAPristineCheckout(t *testing.T) {
	t.Setenv("WF_REVIEW_CHECKOUT_ROOT", t.TempDir())
	p := boundaryWindow(t)
	orig := judge.Run
	t.Cleanup(func() { judge.Run = orig })
	attempt := 0
	judge.Run = func(_ context.Context, _ func(int), n string, args ...string) (judge.ProcessOutput, error) {
		attempt++
		_, _, dir := unwrapReviewArgv(t, n, args)
		if attempt == 1 {
			os.WriteFile(filepath.Join(dir, "m1.go"), []byte("package main // reverted fix\n"), 0o644)
			os.WriteFile(filepath.Join(dir, "residue"), []byte("x"), 0o644)
			return judge.ProcessOutput{Stdout: []byte("Reverting the fix; I'll wait for the suite.\n")}, nil
		}
		if _, err := os.Stat(filepath.Join(dir, "residue")); !os.IsNotExist(err) {
			t.Error("attempt 2 found attempt 1's residue")
		}
		if got, _ := os.ReadFile(filepath.Join(dir, "m1.go")); string(got) != "package main // work\n" {
			t.Errorf("attempt 2 read attempt 1's mutation: %q", got)
		}
		return judge.ProcessOutput{Stdout: []byte(shipReview)}, nil
	}
	res := dispatchBoundaryReview(&bytes.Buffer{}, &bytes.Buffer{}, p)
	if attempt != 2 || res.Verdict != judge.VerdictShip {
		t.Fatalf("%d attempts, verdict %s", attempt, res.Verdict)
	}
}

// #204 close review BR-3: debris the sweep cannot remove is a warning, not a
// review that refuses to run.
func TestBoundaryReviewRunsDespiteSweepProblem(t *testing.T) {
	root := t.TempDir()
	t.Setenv("WF_REVIEW_CHECKOUT_ROOT", root)
	if err := os.MkdirAll(filepath.Join(root, "stuck.lock"), 0o755); err != nil {
		t.Fatal(err)
	}
	p := boundaryWindow(t)
	orig := judge.Run
	t.Cleanup(func() { judge.Run = orig })
	judge.Run = func(context.Context, func(int), string, ...string) (judge.ProcessOutput, error) {
		return judge.ProcessOutput{Stdout: []byte(shipReview)}, nil
	}
	var stderr bytes.Buffer
	res := dispatchBoundaryReview(&bytes.Buffer{}, &stderr, p)
	if res.Verdict != judge.VerdictShip {
		t.Fatalf("verdict %s (%s); a sweep problem blocked the review", res.Verdict, res.Reason)
	}
	if !strings.Contains(stderr.String(), "stuck.lock") {
		t.Errorf("the sweep problem was not reported:\n%s", stderr.String())
	}
}
