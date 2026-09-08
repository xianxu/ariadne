package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/queue"
)

// fakeTrunk is a stateful in-memory stand-in for the trunk. It is NOT a
// function-call mock: it holds content across calls, so a transform that
// replays sees what a peer wrote, which is the behavior these tests are about.
type fakeTrunk struct {
	content  map[string][]byte
	warn     string
	readErr  error
	writeErr error
	// peer, when set, runs once before the first transform and simulates
	// another checkout landing an edit.
	peer  func(f *fakeTrunk)
	calls int
}

func newFakeTrunk(seed string) *fakeTrunk {
	return &fakeTrunk{content: map[string][]byte{queuePath: []byte(seed)}}
}

func (f *fakeTrunk) ReadDegraded(path string) ([]byte, string, error) {
	if f.readErr != nil {
		return nil, f.warn, f.readErr
	}
	return f.content[path], f.warn, nil
}

func (f *fakeTrunk) Update(path, _ string, transform func([]byte) ([]byte, error)) error {
	if f.writeErr != nil {
		return f.writeErr
	}
	f.calls++
	if f.calls == 1 && f.peer != nil {
		f.peer(f)
	}
	next, err := transform(f.content[path])
	if err != nil {
		return err
	}
	f.content[path] = next
	return nil
}

func TestQueueList_ReadsTrunkAndSkipsProse(t *testing.T) {
	f := newFakeTrunk("# Queue\n\nprose\n\n- a#1 — first [x]\n- b#2 — second\n")
	var out, errOut bytes.Buffer
	if err := runQueueList(&out, &errOut, f); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	if !strings.Contains(got, "a#1") || !strings.Contains(got, "b#2") {
		t.Errorf("entries missing:\n%s", got)
	}
	if strings.Contains(got, "prose") || strings.Contains(got, "# Queue") {
		t.Errorf("stdout must carry entries only, so it stays pipeable:\n%s", got)
	}
}

// A degraded read warns on STDERR, keeping stdout clean.
func TestQueueList_OfflineWarningGoesToStderr(t *testing.T) {
	f := newFakeTrunk("- a#1 — first\n")
	f.warn = "origin unreachable — read from the stale ref"
	var out, errOut bytes.Buffer
	if err := runQueueList(&out, &errOut, f); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(errOut.String(), "unreachable") {
		t.Error("the staleness warning must be shown")
	}
	if strings.Contains(out.String(), "unreachable") {
		t.Error("the warning must not pollute stdout")
	}
}

func TestQueueEdit_AddLandsOnTrunk(t *testing.T) {
	f := newFakeTrunk("- a#1 — first\n")
	var out, errOut bytes.Buffer
	err := runQueueEdit(&out, &errOut, f, queue.Intent{
		Op: queue.OpAdd, Ref: "b#2", WhyNow: "second", Tag: "sdlc"})
	if err != nil {
		t.Fatal(err)
	}
	got := string(f.content[queuePath])
	if !strings.Contains(got, "- b#2 — second [sdlc]") {
		t.Errorf("trunk = %q", got)
	}
	if !strings.Contains(got, "- a#1 — first") {
		t.Error("the existing entry must survive")
	}
}

// THE verb-level concurrency test: Intent.Apply is handed to Update as the
// transform, so a peer's edit landing between read and write is preserved rather
// than clobbered. This is the property the whole design exists for.
func TestQueueEdit_PeerEditSurvivesTheReplay(t *testing.T) {
	f := newFakeTrunk("- a#1 — first\n")
	f.peer = func(f *fakeTrunk) {
		f.content[queuePath] = []byte("- a#1 — first\n- peer#9 — landed first\n")
	}
	var out, errOut bytes.Buffer
	if err := runQueueEdit(&out, &errOut, f, queue.Intent{
		Op: queue.OpAdd, Ref: "b#2", WhyNow: "mine"}); err != nil {
		t.Fatal(err)
	}
	got := string(f.content[queuePath])
	for _, want := range []string{"a#1", "peer#9", "b#2"} {
		if !strings.Contains(got, want) {
			t.Errorf("trunk lost %q:\n%s", want, got)
		}
	}
}

// A converged no-op is announced, not silent: the operator asked for something
// and deserves to know it was already true.
func TestQueueEdit_ConvergenceIsReported(t *testing.T) {
	f := newFakeTrunk("- a#1 — first\n")
	var out, errOut bytes.Buffer
	if err := runQueueEdit(&out, &errOut, f, queue.Intent{Op: queue.OpRemove, Ref: "gone#9"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(errOut.String(), "nothing to remove") {
		t.Errorf("a converged no-op must be announced, got:\n%s", errOut.String())
	}
}

// A refusal is a HANDOFF: it must carry the trunk's current state and the intent
// that could not be applied, so the operator or an agent can re-derive the edit.
// Asserting only "it returned an error" would pass on a bare failure and prove
// nothing about the property the design promises.
func TestQueueEdit_RefusalCarriesTrunkStateAndIntent(t *testing.T) {
	f := newFakeTrunk("- a#1 — first\n- b#2 — second\n")
	var out, errOut bytes.Buffer
	err := runQueueEdit(&out, &errOut, f, queue.Intent{
		Op: queue.OpMove, Ref: "a#1", Anchor: "vanished#9"})
	if err == nil {
		t.Fatal("a missing anchor must refuse")
	}
	s := errOut.String()
	for _, want := range []string{
		"queue: move a#1 before vanished#9", // the intent that failed
		"a#1", "b#2",                        // the trunk state, so a re-derivation has something to aim at
		"anchor is gone", // what to do next
	} {
		if !strings.Contains(s, want) {
			t.Errorf("refusal missing %q:\n%s", want, s)
		}
	}
	// The returned error must not double the "queue: " prefix — caught when the
	// real seed run printed "queue: queue: add ariadne#207 refused".
	if strings.Contains(err.Error(), "queue: queue:") {
		t.Errorf("doubled prefix in %q", err.Error())
	}
	// And the trunk is untouched.
	if got := string(f.content[queuePath]); got != "- a#1 — first\n- b#2 — second\n" {
		t.Errorf("trunk modified by a refused edit: %q", got)
	}
}

func TestQueueEdit_ValidationRefusesBeforeTouchingTheTrunk(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   queue.Intent
	}{
		{"bad ref", queue.Intent{Op: queue.OpAdd, Ref: "bad ref", WhyNow: "x"}},
		{"newline in why-now", queue.Intent{Op: queue.OpAdd, Ref: "a#1", WhyNow: "a\n- forged#1 — x"}},
		{"bracket in why-now", queue.Intent{Op: queue.OpAdd, Ref: "a#1", WhyNow: "see [RFC]"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newFakeTrunk("- z#9 — untouched\n")
			var out, errOut bytes.Buffer
			if err := runQueueEdit(&out, &errOut, f, tc.in); err == nil {
				t.Fatal("expected refusal")
			}
			if got := string(f.content[queuePath]); got != "- z#9 — untouched\n" {
				t.Errorf("trunk modified: %q", got)
			}
		})
	}
}

// An offline write refuses and the cause reaches the operator.
func TestQueueEdit_OfflineWriteSurfacesCause(t *testing.T) {
	f := newFakeTrunk("- a#1 — first\n")
	f.writeErr = errors.New("origin unreachable (offline?)")
	var out, errOut bytes.Buffer
	if err := runQueueEdit(&out, &errOut, f, queue.Intent{
		Op: queue.OpAdd, Ref: "b#2", WhyNow: "x"}); err == nil {
		t.Fatal("expected refusal")
	}
	if !strings.Contains(errOut.String(), "unreachable") {
		t.Errorf("cause not surfaced:\n%s", errOut.String())
	}
}

func TestQueueCommitMessage(t *testing.T) {
	for _, tc := range []struct {
		in   queue.Intent
		want string
	}{
		{queue.Intent{Op: queue.OpAdd, Ref: "a#1"}, "queue: add a#1"},
		{queue.Intent{Op: queue.OpRemove, Ref: "a#1"}, "queue: remove a#1"},
		{queue.Intent{Op: queue.OpMove, Ref: "a#1", Anchor: "b#2"}, "queue: move a#1 before b#2"},
		{queue.Intent{Op: queue.OpMove, Ref: "a#1", Anchor: "b#2", After: true}, "queue: move a#1 after b#2"},
	} {
		if got := queueCommitMessage(tc.in); got != tc.want {
			t.Errorf("got %q, want %q", got, tc.want)
		}
	}
}
