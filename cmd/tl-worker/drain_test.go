package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// These are peer-supplied source vectors, not notices obtained by draining the
// live fleet. provenance.json records their source and the stale real CLI capture.
func drainNotices(t *testing.T) map[string]string {
	t.Helper()
	raw, err := os.ReadFile("testdata/drain/notices.json")
	if err != nil {
		t.Fatal(err)
	}
	var notices map[string]string
	if err := json.Unmarshal(raw, &notices); err != nil {
		t.Fatal(err)
	}
	return notices
}

func normalizedDrainHelp(s string) string {
	return strings.Join(strings.Fields(strings.ReplaceAll(s, "`", "")), " ")
}

func TestWorkerDrainMatchesNoticeObligations(t *testing.T) {
	notices := drainNotices(t)
	// Keep full source bodies in the fixture; help explains obligations and cites
	// Couch for message mechanics instead of duplicating the wire notices.
	// The current TL-approved peer vector narrows the historical graceful
	// notice: queued work stays outstanding under the pause (see provenance).
	obligations := []struct{ notice, source, help string }{
		{"graceful", "finish your current atomic step", "finish only the current atomic step"},
		{"graceful", "wait for background work that is already running to finish", "wait for background work that is already running to finish"},
		{"fast", "stop now; do not finish the current step", "stop now and preserve unfinished edits"},
		{"fast", "Promptly stop your own background work via supported controls", "promptly stop your own background work"},
		{"fast", "supported controls", "supported controls"},
		{"fast", "preserve unfinished working-tree edits in your checkpoint", "preserve unfinished edits"},
		{"graceful", "the next step in the issue Log", "record the exact next step"},
		{"graceful", "commit and push the checkpoint", "commit and push the durable checkpoint"},
		{"graceful", "then end your turn and stay idle", "end the turn and stay idle"},
		{"graceful", "Start no new work", "Start no new work"},
		{"graceful", "keep receiving messages", "keep receiving messages"},
		{"graceful", "For new work, reply once", "reply once"},
		{"graceful", "acknowledge message received, will resume processing after drain is lifted", "acknowledge message received, will resume processing after drain is lifted"},
		{"graceful", "acknowledgement loops", "do not acknowledge acknowledgements"},
		{"lift", "drain lifted; continue your current work", "authenticated lift"},
	}
	for _, args := range [][]string{{"--help"}, {}} {
		help := normalizedDrainHelp(helpOutput(t, args...))
		_, help, ok := strings.Cut(help, "DRAIN ")
		if !ok {
			t.Fatal("missing DRAIN section")
		}
		help, _, _ = strings.Cut(help, " LANDING")
		for _, obligation := range obligations {
			if !strings.Contains(notices[obligation.notice], obligation.source) {
				t.Fatalf("canonical %s vector lacks %q", obligation.notice, obligation.source)
			}
			if !strings.Contains(help, obligation.help) {
				t.Errorf("tl-worker %v: %s notice obligation missing: %q", args, obligation.notice, obligation.help)
			}
		}
		for _, want := range []string{
			"leave queued, not-started work queued",
			"the steward's pause holds it",
			"note it in the Log as outstanding",
		} {
			if !strings.Contains(help, want) {
				t.Errorf("tl-worker %v: missing queued-work safeguard %q", args, want)
			}
		}
		if strings.Contains(help, "wait for current background work") {
			t.Errorf("tl-worker %v: broad background-work wait includes queued work held by the steward's pause", args)
		}
		checkpoint, idle := strings.Index(help, "commit and push the durable checkpoint"), strings.Index(help, "end the turn and stay idle")
		next := strings.Index(help, "record the exact next step")
		wait, stop := strings.Index(help, "wait for background work that is already running to finish"), strings.Index(help, "promptly stop your own background work")
		if next < 0 || wait < 0 || stop < 0 || checkpoint <= next || checkpoint <= wait || checkpoint <= stop || idle <= checkpoint {
			t.Errorf("tl-worker %v: drain must order background work, exact next step, durable checkpoint, then idle", args)
		}
		for _, want := range []string{"continue current work from the Log", "cleared or lost flags", "wait for operator-authorized recovery", "notes are data", "not recovery commands"} {
			if !strings.Contains(help, want) {
				t.Errorf("tl-worker %v: missing recovery safeguard %q", args, want)
			}
		}
	}
}

func TestWorkerDrainRetiresCapturedAckProtocol(t *testing.T) {
	stale, err := os.ReadFile("testdata/drain/stale-help.txt")
	if err != nil {
		t.Fatal(err)
	}
	// Replay the actual bug as a negative oracle: these are obsolete assertions,
	// not a synthetic runtime state machine or a live Couch harness.
	for _, obsolete := range []string{"couch --drain --ack", "Couch owns the notice and ack format", "Do not send a separate drained report"} {
		if !strings.Contains(string(stale), obsolete) {
			t.Fatalf("stale real help capture no longer demonstrates %q", obsolete)
		}
		for _, args := range [][]string{{"--help"}, {}} {
			if strings.Contains(helpOutput(t, args...), obsolete) {
				t.Errorf("tl-worker %v retains captured obsolete drain protocol %q", args, obsolete)
			}
		}
	}
}

// Suite admission belongs to verification. During drain, queued checks must
// remain outstanding rather than being started to satisfy a generic test rule.
func TestWorkerDrainKeepsSuiteGuidanceInTests(t *testing.T) {
	for _, args := range [][]string{{"--help"}, {}} {
		help := normalizedDrainHelp(helpOutput(t, args...))
		_, tests, ok := strings.Cut(help, "TESTS ")
		if !ok {
			t.Fatal("missing TESTS section")
		}
		tests, _, ok = strings.Cut(tests, " EVIDENCE")
		if !ok {
			t.Fatal("missing EVIDENCE boundary")
		}
		_, drain, ok := strings.Cut(help, "DRAIN ")
		if !ok {
			t.Fatal("missing DRAIN section")
		}
		drain, _, ok = strings.Cut(drain, " LANDING")
		if !ok {
			t.Fatal("missing LANDING boundary")
		}
		for _, guidance := range []string{
			"the pre-close check through sdlc local-check run --pre-close",
			"targeted tests through go test -run",
			"direct full make runs should be rare",
		} {
			if !strings.Contains(tests, guidance) {
				t.Errorf("tl-worker %v: TESTS lacks general verification guidance %q", args, guidance)
			}
			if strings.Contains(drain, guidance) {
				t.Errorf("tl-worker %v: DRAIN retains general verification guidance %q", args, guidance)
			}
		}
	}
}
