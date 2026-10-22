package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
)

// A code-main reintegration rebuilds only the final head. A later failure
// cannot make an old-base smaller prefix evidence against a car, or permission
// to land an untested integration of that smaller prefix with current main.
func TestLandingTrainCodeReintegrationDoesNotSplitOnStalePrefixes(t *testing.T) {
	s := newTrainSetup(t, "train")
	started, release := blockLandingSuite(t, s.server)
	t.Cleanup(func() { _ = os.WriteFile(release, nil, 0o600) })
	data, err := os.ReadFile(filepath.Join(s.server, localChecksPath))
	if err != nil {
		t.Fatal(err)
	}
	decl, err := parseLocalChecks(data)
	if err != nil {
		t.Fatal(err)
	}
	suite, _ := decl.check(suiteCheckName)
	fail := suiteFailing("example.com/train.TestMainInteraction").Command[2]
	s.declareSuite(t, "if test -e code-main.go && test -e second.go; then "+fail+"; fi; "+suite.Command[2])
	s.pushCar(t, "000002-second", "second", 2, 2)
	one := s.merge(s.first)
	waitQueued(t, s.line, 1)
	two := s.waitingCar(t, "000002-second", 2, 2)
	waitVouched(t, s.line, 2)
	s.blocker.Release()
	deadline := time.NewTimer(30 * time.Second)
	defer deadline.Stop()
	tick := time.NewTicker(5 * time.Millisecond)
	defer tick.Stop()
	for {
		if b, err := os.ReadFile(started); err == nil && len(strings.TrimSpace(string(b))) == 40 {
			break
		}
		select {
		case <-deadline.C:
			t.Fatal("initial full suite did not reach its main-advance barrier")
		case <-tick.C:
		}
	}
	main := integrationAdvance(t, s.server, "code-main.go", "package fixture\n")
	if err := os.WriteFile(release, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	firstErr := finished(t, one)
	other := wantSettled(t, two, landingline.OutcomeRefused)
	if firstErr == nil || !strings.Contains(firstErr.Error(), "unverified:") || !strings.Contains(other.Reason, "unverified:") {
		t.Fatalf("stale prefix authorized continuation: first=%v other=%+v", firstErr, other)
	}
	if s.gh.merges != 0 || git(t, s.gh.remote, "rev-parse", "main") != main {
		t.Fatal("a stale prefix landed an untested code-main integration")
	}
	for _, event := range lineEvents(t, s.line)[landingline.EventCar] {
		if event.Result == string(landingline.OutcomeHeld) || event.Result == string(landingline.OutcomeReleased) {
			t.Fatalf("unverified evidence caused blame or automatic retraining: %+v", event)
		}
	}
	wantNoTrainBranch(t, s.gh.remote)
}
