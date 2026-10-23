package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/bench"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
)

func benchAdmissionFixture(t *testing.T) (*benchStore, string, *benchPublicContext) {
	t.Helper()
	task := preparedFixtureTask(t)
	s, err := prepareFixture(t, task)
	if err != nil {
		t.Fatal(err)
	}
	work := filepath.Join(s.root, "public", "work")
	t.Chdir(work)
	var gateOutput bytes.Buffer
	var gateArgs []string
	previousExit := exitWithCode
	exitWithCode = func(code int) { t.Fatalf("sdlc %v exited %d:\n%s", gateArgs, code, gateOutput.String()) }
	t.Cleanup(func() { exitWithCode = previousExit })
	run := func(args ...string) {
		t.Helper()
		gateArgs = args
		gateOutput.Reset()
		root := buildRoot()
		root.SetArgs(args)
		root.SetOut(&gateOutput)
		root.SetErr(&gateOutput)
		if err := root.Execute(); err != nil {
			t.Fatalf("sdlc %v: %v\n%s", args, err, gateOutput.String())
		}
	}
	run("claim", "--issue", "1")
	run("start-plan", "--issue", "1")
	detail := filepath.Join(work, "workshop", "issues", "000001-bench-replay.md")
	text, err := os.ReadFile(detail)
	if err != nil {
		t.Fatal(err)
	}
	prepareTestWrite(t, work, "workshop/issues/000001-bench-replay.md", strings.Replace(string(text), "## Done when\n\n"+task.DoneWhen, "## Done when\n\n- "+task.DoneWhen, 1), 0644)
	run("change-code", "--issue", "1", "--worktree=no", "--no-judge", "--no-estimate", "--no-estimate-recon")
	prepareTestWrite(t, work, "src/main.txt", "fixed\n", 0644)
	prepareTestGit(t, work, "add", "src/main.txt")
	prepareTestGit(t, work, "commit", "-qm", "#1: implement fixture")
	candidate := prepareTestGit(t, work, "rev-parse", "HEAD")
	stubJudge(t, "VERDICT: SHIP (confidence: high)\n\nfine\n"+emptyBoundaryFindings)
	run("close", "--issue", "1", "--verified", "bench admission native fixture", "--actual", "1", "--no-atlas")
	c, err := loadBenchContext(context.Background(), work)
	if err != nil {
		t.Fatal(err)
	}
	r, _, err := s.read()
	if err != nil {
		t.Fatal(err)
	}
	identity := func(root, slot string, p bench.AgentProfile) bench.RuntimeIdentity {
		return bench.RuntimeIdentity{Root: root, Slot: slot, Session: slot + "-session", Generation: "1", Agent: p.Agent, Model: p.Model, Effort: p.Effort}
	}
	h, w := identity(c.HubRoot, "ops-bench:0", task.TL), identity(c.WorkRoot, "work:0", task.Worker)
	for _, kind := range []bench.EventKind{bench.BeginDispatch, bench.DispatchConfirmed, bench.ProposeWouldLand} {
		op := "dispatch"
		if kind == bench.ProposeWouldLand {
			op = "proof"
		}
		if _, _, err := s.apply(bench.Event{ID: string(kind), RunID: r.ID, Kind: kind, OperationID: op, Hub: h, Worker: w}); err != nil {
			t.Fatal(err)
		}
	}
	return s, candidate, c
}

func TestBenchAdmissionRealCloseMetadataAndSuiteBoundary(t *testing.T) {
	s, candidate, c := benchAdmissionFixture(t)
	calls := 0
	suite := func(_ context.Context, root, head, check string) (benchLocalCheckReceipt, error) {
		calls++
		if root != c.WorkRoot || head != candidate || check != "suite" {
			t.Fatal("suite verified wrong subject")
		}
		return benchLocalCheckReceipt{Version: 1, RunID: c.RunID, WorkRoot: root, OriginIdentity: c.OriginIdentity, Head: head, Check: check, Record: localRunRecord{Name: check, SHA: head, LogSHA256: strings.Repeat("a", 64)}}, nil
	}
	facts, err := verifyBenchWouldLandWithSuite(context.Background(), s, candidate, suite)
	if err != nil {
		t.Fatal(err)
	}
	if calls != 1 || facts.ObservedHead != candidate || facts.PushedHead != candidate || facts.CompletionEvidence == candidate || facts.ReviewedHead != candidate || facts.ReviewVerdict != "SHIP" {
		t.Fatalf("metadata/candidate confused: %+v", facts)
	}
	if err := bench.ValidateWouldLand(facts); err != nil {
		t.Fatal(err)
	}
	if _, err := verifyBenchWouldLand(context.Background(), s, candidate); err == nil {
		t.Fatal("missing native suite proof accepted")
	}
	if focused, err := loadBenchFocusedInputs(context.Background(), s, nil); err != nil || focused != nil {
		t.Fatalf("manual actual manufactured focused inputs: %+v %v", focused, err)
	}
	proof, err := readBenchCloseProof(context.Background(), s)
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, 10, 23, 0, 0, 0, 0, time.UTC)
	stamp := func(offset time.Duration) string { return start.Add(offset).Format(time.RFC3339Nano) }
	retained := gatestate.MeasurementInputs{Version: 1, Method: gatestate.MeasurementMethodEngagementV1, Repository: c.OriginIdentity, CodeHead: candidate, Since: stamp(0), Until: stamp(time.Hour), GapMinutes: 30, CommitWeight: 1, Engagements: []gatestate.MeasurementEngagement{{Start: stamp(0), End: stamp(time.Hour), Began: stamp(0), Roots: []string{c.WorkRoot}}}, Sources: []gatestate.MeasurementSource{{Harness: "codex", ID: "native-worker", CWD: c.WorkRoot, Digest: strings.Repeat("a", 64), Bytes: 3, Started: stamp(0), Windows: []gatestate.MeasurementInterval{{Start: stamp(0), End: stamp(time.Hour)}}}}}
	retained.TrackerHead = prepareTestGit(t, c.WorkRoot, "rev-parse", "refs/remotes/origin/issue-tracker")
	for i := range proof.ledger.Rounds {
		if proof.ledger.Rounds[i].Reviewed == candidate {
			proof.ledger.Rounds[i].ActualInputs = &retained
		}
	}
	source := benchCapturedSource{benchSourceInput: benchSourceInput{Runtime: bench.RuntimeIdentity{Root: c.WorkRoot, Agent: "codex"}, NativeSession: "native-worker", File: "/observer/captured.jsonl"}, OriginalFile: "/explicit/native.jsonl", Length: 10, Digest: strings.Repeat("b", 64), Complete: true}
	focused, err := benchFocusedFromProof(context.Background(), proof, []benchCapturedSource{source})
	if err != nil {
		t.Fatal(err)
	}
	if focused.Prefixes[source.OriginalFile].Bytes != 3 || focused.Prefixes[source.OriginalFile].Digest != strings.Repeat("a", 64) || focused.Files[0] != source.OriginalFile || focused.Scope.Issue != "1" {
		t.Fatalf("retained bounds replaced by full capture: %+v", focused)
	}
	for _, bad := range []benchCapturedSource{func() benchCapturedSource { x := source; x.Length = 2; return x }(), func() benchCapturedSource { x := source; x.NativeSession = "foreign"; return x }(), func() benchCapturedSource { x := source; x.Runtime.Root = "/foreign"; return x }()} {
		if _, err := benchFocusedFromProof(context.Background(), proof, []benchCapturedSource{bad}); err == nil {
			t.Fatal("unbound retained source accepted")
		}
	}
	for name, read := range map[string]benchAdmissionSuiteReader{
		"failure": func(ctx context.Context, root, head, check string) (benchLocalCheckReceipt, error) {
			r, err := suite(ctx, root, head, check)
			r.Record.Exit = 1
			return r, err
		},
		"stale": func(ctx context.Context, root, head, check string) (benchLocalCheckReceipt, error) {
			r, err := suite(ctx, root, head, check)
			r.Head = strings.Repeat("f", 40)
			return r, err
		},
		"foreign backend": func(ctx context.Context, root, head, check string) (benchLocalCheckReceipt, error) {
			r, err := suite(ctx, root, head, check)
			r.OriginIdentity = "file:/foreign"
			return r, err
		},
		"missing verifier": nil,
		"unavailable": func(context.Context, string, string, string) (benchLocalCheckReceipt, error) {
			return benchLocalCheckReceipt{}, errors.New("suite unavailable")
		},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := verifyBenchWouldLandWithSuite(context.Background(), s, candidate, read); err == nil {
				t.Fatal("invalid suite evidence accepted")
			}
		})
	}
	if _, err := verifyBenchWouldLandWithSuite(context.Background(), s, strings.Repeat("f", 40), suite); err == nil {
		t.Fatal("foreign candidate accepted")
	}
	dirty := filepath.Join(c.WorkRoot, "untracked.txt")
	if err := os.WriteFile(dirty, []byte("dirty"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := verifyBenchWouldLandWithSuite(context.Background(), s, candidate, suite); err == nil {
		t.Fatal("dirty checkout accepted")
	}
	if err := os.Remove(dirty); err != nil {
		t.Fatal(err)
	}
	moving := func(ctx context.Context, root, head, check string) (benchLocalCheckReceipt, error) {
		r, err := suite(ctx, root, head, check)
		prepareTestWrite(t, root, "atlas/late.md", "changed after graph proof\n", 0644)
		prepareTestGit(t, root, "add", "atlas/late.md")
		prepareTestGit(t, root, "commit", "-qm", "late unapproved metadata")
		return r, err
	}
	if _, err := verifyBenchWouldLandWithSuite(context.Background(), s, candidate, moving); err == nil {
		t.Fatal("HEAD movement during suite proof accepted")
	}
}

func TestBenchAdmissionRefusesIncompleteInputs(t *testing.T) {
	for _, candidate := range []string{"HEAD", "abcd", strings.Repeat("A", 40)} {
		if _, err := verifyBenchWouldLand(context.Background(), nil, candidate); err == nil {
			t.Fatal("invalid candidate accepted")
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := verifyBenchWouldLand(ctx, nil, strings.Repeat("a", 40)); err == nil {
		t.Fatal("cancelled admission accepted")
	}
}

func TestBenchAdmissionUsesNativeGateCap(t *testing.T) {
	ledger := gatestate.Ledger{Gate: "boundary-review", IDPrefix: "BR", IssueNum: 1}
	for i := 1; i <= gatestate.DefaultRoundCap+1; i++ {
		round := gatestate.Round{N: i}
		if i == 1 {
			round.New = []gatestate.Finding{{ID: "BR-1", Severity: "Important", Title: "retained important finding", Round: 1}}
		}
		ledger.Rounds = append(ledger.Rounds, round)
	}
	native := gatestate.DecideScoped(gatestate.FilterBoundary(ledger, ""), openScopeFor(ledger, ""), gatestate.DefaultRoundCap)
	if native.Block || len(native.Demoted) != 1 {
		t.Fatal("fixture does not exercise native demotion")
	}
	if err := verifyBenchCloseGate(ledger); err != nil {
		t.Fatalf("bench tightened native cap: %v", err)
	}
	beforeCap := ledger
	beforeCap.Rounds = beforeCap.Rounds[:1]
	if err := verifyBenchCloseGate(beforeCap); err == nil {
		t.Fatal("native blocking finding accepted")
	}
	ledger.Rounds[0].New[0].Severity = "Critical"
	if err := verifyBenchCloseGate(ledger); err == nil {
		t.Fatal("hard blocker demoted by bench")
	}
}
