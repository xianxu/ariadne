package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/affected"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

const splitBroadcastTest = "example.com/train/broadcast.TestHubRandomInterleavings"

// Replay train234's boundary: the combined suite failed before another writer
// registered the broadcast flake. The adjacent suspect changes only scribe.
// Registration is a real tracker publication while the passing prefix waits,
// rather than a registry callback returning a different handcrafted result.
func TestLandingTrainSplitRefreshesRegisteredFlake(t *testing.T) {
	testLandingTrainUnattributed(t, true)
}

func TestLandingTrainSplitDoesNotBlameUntouchedPackage(t *testing.T) {
	testLandingTrainUnattributed(t, false)
}

func testLandingTrainUnattributed(t *testing.T, register bool) {
	t.Helper()
	s := newTrainSetup(t, "train")
	barriers := t.TempDir()
	ready, release := filepath.Join(barriers, "passing-prefix"), filepath.Join(barriers, "continue")
	// Use complete Go JSON outcomes: the failed evidence names the parent
	// only. Attribution must not invent an unobserved randomized subtest.
	script := `if grep -q 'const Changed' scribe/scribe.go; then
printf '%s\n' '{"Action":"fail","Package":"example.com/train/broadcast","Test":"TestHubRandomInterleavings","Elapsed":0.01}' '{"Action":"fail","Package":"example.com/train/broadcast","Elapsed":0.01}'
exit 1
fi
if test -e shipped.txt; then
  git rev-parse HEAD > "$1"
  while ! test -e "$2"; do sleep 0.01; done
fi
printf '%s\n' '{"Action":"pass","Package":"example.com/train/broadcast","Elapsed":0.01}'
`
	checks, err := json.Marshal(map[string]any{"version": 1, "checks": []localCheck{{Name: suiteCheckName, Required: true, Command: []string{"sh", "-c", script, "split-attribution", ready, release}, TimeoutMinutes: 1}}})
	if err != nil {
		t.Fatal(err)
	}
	writeTrainAttributionGraph(t, s.server)
	procedureWrite(t, s.server, localChecksPath, string(checks))
	git(t, s.server, "add", ".")
	git(t, s.server, "commit", "-qm", "real broadcast and scribe graph with split barrier")
	git(t, s.server, "push", "-q", "origin", "main")

	const cardPath = "workshop/issue-cards/000771-broadcast-flake.md"
	card, _, err := issue.SplitCard([]byte(issue.Render(issue.ScaffoldSpec{
		ID: "000771", Title: "Broadcast flake", Today: "2026-10-21",
		ProblemBody: "Broadcast scheduling sometimes fails independently of scribe.",
	})))
	if err != nil {
		t.Fatal(err)
	}
	bootstrapTrainAttributionTracker(t, s, map[string][]byte{cardPath: card})

	const suspectBranch = "000768-scribe"
	suspect := filepath.Join(filepath.Dir(filepath.Dir(s.first)), "sample-slot2", "sample")
	git(t, suspect, "fetch", "-q", "upstream")
	git(t, suspect, "reset", "-q", "--hard", "upstream/main")
	git(t, suspect, "config", "user.name", "scribe")
	git(t, suspect, "config", "user.email", "scribe@example.com")
	git(t, suspect, "switch", "-q", "-c", suspectBranch)
	procedureWrite(t, suspect, "scribe/scribe.go", "package scribe\nconst Changed = true\n")
	git(t, suspect, "add", "scribe/scribe.go")
	git(t, suspect, "commit", "-qm", "scribe-only suspect")
	git(t, suspect, "push", "-q", "upstream", suspectBranch)
	s.addCar(t, suspectBranch, 768)
	if got := git(t, suspect, "diff", "--name-only", "upstream/main...HEAD"); got != "scribe/scribe.go" {
		t.Fatalf("suspect must change only scribe, got %q", got)
	}

	// The independent writer has its own checkout, refs, and tracker cache.
	writer := filepath.Join(t.TempDir(), "independent-tracker-writer")
	git(t, "", "clone", "-q", s.gh.remote, writer)
	git(t, writer, "config", "user.name", "flake registrar")
	git(t, writer, "config", "user.email", "registrar@example.com")
	external, err := gitx.NewTrunkFileContext(context.Background(), writer, "origin", "issue-tracker")
	if err != nil {
		t.Fatal(err)
	}
	trackerBefore := git(t, s.gh.remote, "rev-parse", "refs/heads/issue-tracker")
	ctx, cancel := context.WithCancel(withIssueRecordsScope(context.WithValue(context.Background(), trainRootKey{}, s.first)))
	// Explicitly seed the command's pre-registration snapshot. A Fresh read
	// inside this same scope is insufficient after the independent write.
	registry, err := gateFlakeRegistry(s.first)(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, covered := registry.Cover(splitBroadcastTest); covered {
		t.Fatal("the initial registry already covers the failure")
	}
	var workers sync.WaitGroup
	var stderr bytes.Buffer // inspected only after the conductor has joined
	t.Cleanup(func() {
		_ = os.WriteFile(release, nil, 0600)
		cancel()
		s.blocker.Release()
		workers.Wait()
	})
	start := func(root string) <-chan error {
		done := make(chan error, 1)
		workers.Add(1)
		go func() {
			defer workers.Done()
			f := landingFlags()
			f.NoValidate = true
			f.Context = context.WithValue(ctx, trainRootKey{}, root)
			if root != s.first {
				f.Context = withIssueRecordsScope(f.Context)
			}
			cmd := markManualLockCommand(&cobra.Command{Use: "merge"})
			cmd.SetContext(f.Context)
			f.Lock = func(run func() error) error { return withRequiredRepoTransactionLock(cmd, run) }
			var output bytes.Buffer
			err := runMerge(&output, &output, f)
			// Publish output through the result channel only after writing ends.
			if root == suspect {
				stderr.Write(output.Bytes())
			}
			done <- err
		}()
		return done
	}
	one := start(s.first)
	waitQueued(t, s.line, 1)
	two := start(suspect)
	waitVouched(t, s.line, 2)
	s.blocker.Release()
	var prefix string
	deadline := time.Now().Add(30 * time.Second)
	for prefix == "" {
		data, readErr := os.ReadFile(ready)
		if readErr == nil {
			prefix = strings.TrimSpace(string(data))
		} else if !errors.Is(readErr, os.ErrNotExist) {
			t.Fatal(readErr)
		}
		if time.Now().After(deadline) {
			t.Fatal("the passing split prefix never reached the registration barrier")
		}
		// Poll a bounded file barrier, rather than assuming a suite duration.
		time.Sleep(time.Millisecond)
	}
	failed := ""
	s.gh.ciFakeGH.mu.Lock()
	for sha, statuses := range s.gh.ciFakeGH.statuses {
		if statuses[localStatusContext(suiteCheckName)] == "failure" {
			failed = sha
		}
	}
	s.gh.ciFakeGH.mu.Unlock()
	if failed == "" || failed == prefix {
		t.Fatalf("barrier must follow a distinct failed combined head: failed=%q prefix=%q", failed, prefix)
	}
	if got := git(t, s.gh.remote, "diff", "--name-only", prefix, failed); got != "scribe/scribe.go" {
		t.Fatalf("adjacent split must isolate scribe, got %q", got)
	}
	if register {
		if err := external.UpdateMany("register broadcast while split prefix runs", func(view *gitx.TrunkView) (gitx.TrunkWrite, error) {
			current, err := view.Read(cardPath)
			if err != nil {
				return gitx.TrunkWrite{}, err
			}
			next, err := issue.SetCardField(current, "flake_tests", issue.FlowList([]string{splitBroadcastTest}))
			return gitx.TrunkWrite{Write: map[string][]byte{cardPath: next}, ExactBytes: true}, err
		}); err != nil {
			t.Fatal(err)
		}
		if after := git(t, s.gh.remote, "rev-parse", "refs/heads/issue-tracker"); after == trackerBefore {
			t.Fatal("independent writer did not publish a new tracker commit")
		}
		fresh, err := gateFlakeRegistry(s.first)(withIssueRecordsScope(ctx))
		if err != nil {
			t.Fatal(err)
		}
		if _, covered := fresh.Cover(splitBroadcastTest); !covered {
			t.Fatal("published tracker card is not a real registry entry")
		}
		cached, err := gateFlakeRegistry(s.first)(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if _, covered := cached.Cover(splitBroadcastTest); covered {
			t.Fatal("fixture lost the stale command snapshot needed for the refresh regression")
		}
	}
	if err := os.WriteFile(release, nil, 0600); err != nil {
		t.Fatal(err)
	}
	if err := finished(t, one); err != nil {
		t.Fatalf("verified passing prefix did not land: %v", err)
	}
	waiterErr := finished(t, two)
	workers.Wait()
	if waiterErr == nil {
		t.Fatal("unresolved suspect merge unexpectedly succeeded")
	}
	if len(trainEvents(t, s.line, string(landingline.PhaseLanded))) != 1 {
		t.Fatal("expected exactly one completed prefix train")
	}
	// A real waiting merge renders OutcomeRefused as its non-blaming error;
	// unlike a direct line.Wait ticket, this also exercises retry reporting.
	var settled *landingline.Settled
	if !errors.As(waiterErr, &settled) || settled.Outcome.Kind != landingline.OutcomeRefused {
		t.Fatalf("suspect outcome: %v, want non-blaming refused", waiterErr)
	}
	outcome := settled.Outcome

	wantReason := "untouched package " + splitBroadcastTest
	if register {
		wantReason = "registered flake " + splitBroadcastTest
	}
	if !strings.Contains(outcome.Reason, wantReason) {
		t.Fatalf("non-blaming refusal %q, want %q", outcome.Reason, wantReason)
	}
	if !strings.Contains(stderr.String(), "sdlc merge --branch "+suspectBranch+" --yes") {
		t.Fatalf("missing explicit retry for unresolved car:\n%s", stderr.String())
	}
	if p := s.gh.pr(suspectBranch); p.State != "OPEN" {
		t.Fatalf("failed unresolved car PR is %s", p.State)
	}
	if p := s.gh.pr(landingTestBranch); p.State != "MERGED" {
		t.Fatalf("verified passing car PR is %s", p.State)
	}
	main := git(t, s.gh.remote, "rev-parse", "refs/heads/main")
	if _, err := trainGitOut(s.gh.remote, "merge-base", "--is-ancestor", prefix, main); err != nil {
		t.Fatal("passing prefix is absent from final main")
	}
	if _, err := trainGitOut(s.gh.remote, "merge-base", "--is-ancestor", failed, main); err == nil {
		t.Fatal("registration/untouchedness promoted a failed combined head onto main")
	}
	if _, err := trainGitOut(s.gh.remote, "merge-base", "--is-ancestor", s.heads[suspectBranch], main); err == nil {
		t.Fatal("unresolved suspect landed")
	}
	events := trainEvents(t, s.line, string(landingline.PhaseLanded))
	if len(events) != 1 || events[0].Culprit != 0 || s.gh.merges != 1 {
		t.Fatalf("split must land only its verified prefix with no culprit: events=%+v merges=%d", events, s.gh.merges)
	}
	refused := false
	for _, event := range lineEvents(t, s.line)[landingline.EventCar] {
		if event.Branch == suspectBranch {
			if event.Result == string(landingline.OutcomeHeld) {
				t.Fatalf("unattributed suspect was held: %+v", event)
			}
			refused = refused || event.Result == string(landingline.OutcomeRefused)
		}
	}
	if !refused {
		t.Fatal("missing terminal refused ledger outcome for unresolved car")
	}
	for _, event := range lineEvents(t, s.line)[landingline.EventTrain] {
		if event.Culprit != 0 {
			t.Fatalf("non-blaming split assigned a culprit: %+v", event)
		}
	}
	wantNoTrainBranch(t, s.gh.remote)
	// #616: an unregistered failure in an untouched package is a missed mapping in the clone's ledger; a registered
	// flake explains its failure and records nothing.
	store, err := missedStoreFor(s.first)
	if err != nil {
		t.Fatal(err)
	}
	ledger, err := store.read()
	if err != nil {
		t.Fatal(err)
	}
	if register {
		if len(ledger.Entries) != 0 {
			t.Fatalf("a registered flake was recorded as a missed mapping: %+v", ledger.Entries)
		}
	} else if len(ledger.Entries) != 1 || ledger.Entries[0].Test != splitBroadcastTest || ledger.Entries[0].Package != "example.com/train/broadcast" ||
		ledger.Entries[0].Source != "train" || !reflect.DeepEqual(ledger.Entries[0].Paths, []string{"scribe/scribe.go"}) || ledger.Entries[0].Head != failed {
		t.Fatalf("missed mapping: %+v", ledger.Entries)
	}
}

// Only an untouched, unregistered failure is a missed mapping, and a store failure is a diagnostic line, never a
// verdict (the verdict is computed before recording).
func TestRecordSplitMisses(t *testing.T) {
	var stderr bytes.Buffer
	var got []affected.Miss
	c := &conductor{stderr: &stderr, missed: func(_ context.Context, misses []affected.Miss, ev affected.MissEvidence) (string, error) {
		got = misses
		if ev.Source != "train" || ev.Head != "hi" || ev.Main != "lo" {
			t.Fatalf("evidence %+v", ev)
		}
		return "", errors.New("disk full")
	}}
	c.recordSplitMisses(context.Background(), []landingline.SplitFailure{
		{Test: "example.com/m/a.TestUntouched", Impact: landingline.ImpactUntouched},
		{Test: "example.com/m/b.TestRegistered", Impact: landingline.ImpactUntouched, RegisteredIssue: "000771"},
		{Test: "example.com/m/c.TestAffected", Impact: landingline.ImpactAffected},
		{Test: "example.com/m/d.TestUnknown", Impact: landingline.ImpactUnknown},
	}, []string{"x"}, "lo", "hi")
	if !reflect.DeepEqual(got, []affected.Miss{{Test: "example.com/m/a.TestUntouched", Pkg: "example.com/m/a"}}) {
		t.Fatalf("misses %+v", got)
	}
	if !strings.Contains(stderr.String(), "missed mappings: not recorded (disk full)") {
		t.Fatalf("stderr %q", stderr.String())
	}
}

// writeTrainAttributionGraph installs valid endpoint graph inputs without
// committing. Callers add their own changed package and suite declaration.
func writeTrainAttributionGraph(t *testing.T, root string) {
	t.Helper()
	for path, content := range map[string]string{
		"go.mod":                      "module example.com/train\n\ngo 1.23\n",
		"broadcast/broadcast.go":      "package broadcast\n",
		"broadcast/broadcast_test.go": "package broadcast\nimport \"testing\"\nfunc TestHubRandomInterleavings(t *testing.T) {}\n",
		"scribe/scribe.go":            "package scribe\n",
		".sdlc/affected.json":         `{"version":1,"command":["true"],"full":["go.mod","go.sum",".sdlc/**","Makefile*"],"timeout_minutes":1}` + "\n",
	} {
		procedureWrite(t, root, path, content)
	}
}

// bootstrapTrainAttributionTracker makes registry reads real, including an empty
// registry for the attributable control. Car details remain legacy/unbound so
// this fixture tests split authority rather than completion-token cleanup.
func bootstrapTrainAttributionTracker(t *testing.T, s *trainSetup, cards map[string][]byte) {
	t.Helper()
	tf, err := gitx.NewTrunkFileContext(context.Background(), s.first, "upstream", "issue-tracker")
	if err != nil {
		t.Fatal(err)
	}
	seed := map[string][]byte{tracker.ManifestPath: tracker.ManifestBytes()}
	for path, card := range cards {
		if _, err := issue.ParseCard(card); err != nil {
			t.Fatalf("invalid tracker fixture %s: %v", path, err)
		}
		seed[path] = card
	}
	if _, err := tf.Bootstrap(seed, "bootstrap split attribution registry", func(gitx.BootstrapResult) error { return nil }); err != nil {
		t.Fatal(err)
	}
	markCutoverOn(t, s.server, "origin", "main")
	git(t, s.first, "fetch", "-q", "upstream")
	git(t, s.first, "branch", "-f", "main-slot1", "upstream/main")
	git(t, s.first, "rebase", "-q", "upstream/main")
	git(t, s.first, "push", "-q", "--force", "upstream", landingTestBranch)
	s.addCar(t, landingTestBranch, 1)
}
