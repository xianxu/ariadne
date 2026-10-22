package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

// Real local and pushed refs provide the identity proof; the stateful GH fake
// owns independent server checks and exact-commit local-check records.
func carProductionCIFixture(t *testing.T, affected, guard bool) (landingTarget, *landingFakeGH, string) {
	t.Helper()
	root, server, gh := integrationFixture(t)
	procedureWrite(t, server, localChecksPath, `{"version":2,"checks":[{"name":"suite","command":["sh","-c","touch unexpected-local-run"],"required":true},{"name":"vm","command":["true"],"scope":"landing","resource":"runner","required":true}]}`+"\n")
	if affected {
		procedureWrite(t, server, affectedConfigPath, `{"version":1,"command":["true"]}`+"\n")
	}
	if guard {
		procedureWrite(t, server, localGatesWorkflow, "name: local-gates-guard\n")
	}
	git(t, server, "add", ".")
	git(t, server, "commit", "-qm", "declare car checks")
	git(t, server, "push", "-q", "origin", "main")
	git(t, root, "fetch", "-q", "upstream", "main")
	// This code commit also prevents an earlier record from being inherited
	// across the changed code, while creating workflows on the actual head.
	procedureWrite(t, root, "feature.go", "package feature\n")
	procedureWrite(t, root, ".github/workflows/ci.yml", "name: build\n")
	git(t, root, "add", ".")
	git(t, root, "commit", "-qm", "head requires CI")
	git(t, root, "push", "-q", "upstream", landingTestBranch)
	head := procedureHead(t, root)
	gh.pr.HeadOID = head
	gh.ci.setHead(gh.pr.Number, head)
	return landingTarget{Root: root, Remote: "upstream", Repo: "test/repo"}, gh, head
}

func assertCarCIReadOnly(t *testing.T, target landingTarget, gh *landingFakeGH, head string) {
	t.Helper()
	if procedureHead(t, target.Root) != head || git(t, target.Root, "ls-remote", target.Remote, "refs/heads/"+landingTestBranch) != head+"\trefs/heads/"+landingTestBranch {
		t.Fatal("readiness changed local or pushed head")
	}
	if got := git(t, target.Root, "status", "--porcelain"); got != "" {
		t.Fatalf("readiness changed worktree: %s", got)
	}
	if gh.merges != 0 || len(gh.ci.comments) != 0 {
		t.Fatalf("readiness effects: merges=%d comments=%v", gh.merges, gh.ci.comments)
	}
}

func TestCarProductionCIMissingCheckVerdicts(t *testing.T) {
	for _, tc := range []struct {
		name    string
		bucket  string
		guard   bool
		reason  ciReason
		context string
	}{
		{"green", "pass", false, ciLocalNeeded, "needs recorded local checks"},
		{"pending", "pending", false, ciPending, "checks still running"},
		{"red", "fail", false, ciRed, "failing checks"},
		{"unknown", "new-state", false, ciUnknown, "state this gate does not know"},
		{"not registered", "", false, ciNotRegistered, "reports no checks yet"},
		{"guard missing", "pass", true, ciMissing, "has not reported " + localGatesVerdict},
	} {
		t.Run(tc.name, func(t *testing.T) {
			target, gh, head := carProductionCIFixture(t, false, tc.guard)
			if tc.bucket != "" {
				gh.ci.setHead(gh.pr.Number, head, prCheck{Name: "build", Bucket: tc.bucket})
			}
			gates := productionCarGates(execGitRunner{}, gh, target, &mergeFlags{NoJudge: true})
			ready, merged, err := carReadiness(context.Background(), execGitRunner{}, gh, target, landingTestBranch, gates, time.Now())
			if ready.Ready || merged || err == nil {
				t.Fatalf("ready=%+v merged=%v error=%v", ready, merged, err)
			}
			var verdict *ciGateError
			if !errors.As(err, &verdict) || verdict.Reason != tc.reason || !reflect.DeepEqual(verdict.LocalNeeded, []string{suiteCheckName}) {
				t.Fatalf("typed error=%+v error=%v", verdict, err)
			}
			for _, want := range []string{tc.context, "merge stopped: run sdlc local-check run --name suite, then re-enter sdlc merge", head, "already queued or running"} {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("missing %q: %v", want, err)
				}
			}
			if len(gh.ci.statuses) != 0 {
				t.Fatalf("readiness posted local records: %v", gh.ci.statuses)
			}
			assertCarCIReadOnly(t, target, gh, head)
		})
	}
}

func TestCarProductionCISatisfiedRecords(t *testing.T) {
	for _, tc := range []struct {
		name     string
		affected bool
		check    string
		ancestor bool
	}{
		{"suite at head", false, suiteCheckName, false},
		{"affected at head", true, affectedCheckName, false},
		{"full-suite fallback", true, suiteCheckName, false},
		{"suite before docs only", false, suiteCheckName, true},
		{"affected before docs only", true, affectedCheckName, true},
		{"fallback before docs only", true, suiteCheckName, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			target, gh, head := carProductionCIFixture(t, tc.affected, false)
			recorded := head
			postStatus(t, gh.ci, recorded, tc.check)
			if tc.ancestor {
				procedureWrite(t, target.Root, "atlas/settled.md", "# Docs only\n")
				git(t, target.Root, "add", ".")
				git(t, target.Root, "commit", "-qm", "docs after recorded code")
				git(t, target.Root, "push", "-q", target.Remote, landingTestBranch)
				head = procedureHead(t, target.Root)
				gh.pr.HeadOID = head
			}
			gh.ci.setHead(gh.pr.Number, head, prCheck{Name: "build", Bucket: "pass"})
			gates := productionCarGates(execGitRunner{}, gh, target, &mergeFlags{NoJudge: true})
			ready, merged, err := carReadiness(context.Background(), execGitRunner{}, gh, target, landingTestBranch, gates, time.Now())
			if !ready.Ready || ready.Head != head || merged || err != nil {
				t.Fatalf("ready=%+v merged=%v error=%v", ready, merged, err)
			}
			if len(gh.ci.statuses) != 1 || gh.ci.statuses[recorded][localStatusContext(tc.check)] != "success" {
				t.Fatalf("readiness mutated records: %v", gh.ci.statuses)
			}
			// Every main declaration also requires vm at landing scope; no
			// branch vm record exists, so success defends that separation.
			assertCarCIReadOnly(t, target, gh, head)
		})
	}
}

func TestCarProductionCINoCI(t *testing.T) {
	target, gh, head := carProductionCIFixture(t, true, true)
	gh.ci.setHead(gh.pr.Number, head, prCheck{Name: "build", Bucket: "fail"})
	gates := productionCarGates(execGitRunner{}, gh, target, &mergeFlags{NoJudge: true, NoCI: true})
	ready, merged, err := carReadiness(context.Background(), execGitRunner{}, gh, target, landingTestBranch, gates, time.Now())
	if !ready.Ready || merged || err != nil {
		t.Fatalf("ready=%+v merged=%v error=%v", ready, merged, err)
	}
	if len(gh.ci.statuses) != 0 {
		t.Fatal("bypass posted records")
	}
	assertCarCIReadOnly(t, target, gh, head)
}

func TestCarProductionCIPair779Capture(t *testing.T) {
	// Immutable expectations transcribed from the retained tool capture. The
	// capture is an oracle for the historical failure, never a decision input.
	const capturedPending = "landing train: not ready: CI gate: CI gate: PR #535 has checks still running on its head"
	const capturedMissing = "landing train: not ready: CI gate: CI gate: PR #535 needs recorded local checks on its head: affected or its full-suite fallback"
	capture, err := os.ReadFile(filepath.Join("testdata", "landing", "pair779-missing-check.txt"))
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range []string{capturedPending, capturedMissing} {
		if !strings.Contains(string(capture), line) {
			t.Fatalf("historical capture lacks %q", line)
		}
	}
	target, gh, head := carProductionCIFixture(t, true, false)
	gh.pr.Number = 535
	gates := productionCarGates(execGitRunner{}, gh, target, &mergeFlags{NoJudge: true})
	for i, bucket := range []string{"pending", "pass"} {
		gh.ci.setHead(535, head, prCheck{Name: "build", Bucket: bucket})
		ready, merged, terminal := carReadiness(context.Background(), execGitRunner{}, gh, target, landingTestBranch, gates, time.Now())
		want := []string{capturedPending, capturedMissing}[i]
		if "landing train: not ready: "+ready.Reason != want || ready.Ready || merged || terminal == nil {
			t.Fatalf("historical step %d: ready=%+v merged=%v error=%v", i, ready, merged, terminal)
		}
		for _, action := range []string{"merge stopped: run sdlc local-check run --name affected, then re-enter sdlc merge", "If affected fell back", "passing suite on this head", head} {
			if !strings.Contains(terminal.Error(), action) {
				t.Errorf("step %d lacks %q: %v", i, action, terminal)
			}
		}
	}
	if len(gh.ci.statuses) != 0 {
		t.Fatal("readiness posted records")
	}
	assertCarCIReadOnly(t, target, gh, head)
}

// Exercise the public merge entry, its real queue enrollment, and the held
// landing ticket as one flow. A second CI observation cancels the negative
// control, so the former polling behavior fails without a timing assumption.
func TestCarProductionCIMergeReturnsMissingAction(t *testing.T) {
	target, gh, head := carProductionCIFixture(t, true, false)
	server := filepath.Join(t.TempDir(), "train-config")
	git(t, "", "clone", "-q", gh.remote, server)
	git(t, server, "config", "user.name", "fixture")
	git(t, server, "config", "user.email", "fixture@example.com")
	procedureWrite(t, server, landingConfigPath, "{\"train_interval_minutes\":60}\n")
	git(t, server, "add", ".")
	git(t, server, "commit", "-qm", "enable trains")
	git(t, server, "push", "-q", "origin", "main")
	git(t, target.Root, "fetch", "-q", target.Remote, "main")
	line, held := heldLine(t, target.Root)

	ctx, cancel := context.WithCancel(context.Background())
	if deadline, ok := t.Deadline(); ok {
		cancel()
		ctx, cancel = context.WithDeadline(context.Background(), deadline)
	}
	defer cancel()
	bounded := &carMissingMergeGH{landingFakeGH: gh, stop: cancel}
	ghClient = bounded // landingFixture's cleanup restores the original adapter
	oldEvery := carVouchEvery
	carVouchEvery = 0 // a repeated poll reaches the explicit negative control
	t.Cleanup(func() { carVouchEvery = oldEvery })
	gh.ci.setHead(gh.pr.Number, head, prCheck{Name: "build", Bucket: "pass"})
	f := landingFlags()
	f.Context, f.NoJudge, f.NoCI = ctx, true, false
	output := &watchWriter{want: "unused", seen: make(chan struct{})}
	err := runMerge(output, output, f)
	if err == nil || !strings.Contains(err.Error(), "merge stopped: run sdlc local-check run --name affected, then re-enter sdlc merge") || bounded.reads != 1 {
		t.Fatalf("reads=%d error=%v\n%s", bounded.reads, err, output.String())
	}
	if current, waiting, statusErr := line.Status(); statusErr != nil || current == nil || current.Seq != held.Ticket().Seq || len(waiting) != 0 {
		t.Fatalf("merge did not retire only its own ticket: head=%v waiting=%v error=%v", current, waiting, statusErr)
	}
	if branch := git(t, target.Root, "branch", "--show-current"); branch != landingTestBranch {
		t.Fatalf("merge changed branch: %s", branch)
	}
	if len(gh.ci.statuses) != 0 {
		t.Fatalf("merge ran a local check: %v", gh.ci.statuses)
	}
	assertCarCIReadOnly(t, target, gh, head)
}

type carMissingMergeGH struct {
	*landingFakeGH
	stop  context.CancelFunc
	reads int
}

func (g *carMissingMergeGH) PRChecks(ctx context.Context, repo string, number int) ([]prCheck, error) {
	g.reads++
	if g.reads > 1 {
		g.stop()
		return nil, context.Canceled
	}
	return g.landingFakeGH.PRChecks(ctx, repo, number)
}
