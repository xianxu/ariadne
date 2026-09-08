package main

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/queue"
)

// fakeTrunk is a stateful in-memory stand-in for the trunk. It is NOT a
// function-call mock: it holds content across calls, so a transform that replays
// sees what a peer wrote, which is the behavior these tests are about.
//
// It also MODELS THE REAL ORDERING: gitx.TrunkFile.Update fetches and reads the
// trunk BEFORE invoking the transform. The first version of this fake called the
// transform immediately, and that gap let
// TestQueueEdit_ValidationRefusesBeforeTouchingTheTrunk pass while asserting
// something false — validation lived inside the transform, so a malformed ref
// actually cost a fetch. `reached` records that the trunk was touched, which is
// what makes the claim testable instead of assumed.
//
// The rule this cost us: a fake must reproduce the ORDERING its real counterpart
// guarantees, or tests written against it assert the fake's behavior rather than
// the system's.
type fakeTrunk struct {
	content  map[string][]byte
	warn     string
	readErr  error
	writeErr error
	// peer, when set, simulates another checkout landing an edit BETWEEN our read
	// and our write — which is the only interleaving that forces a retry.
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
	// Order matters: the real Update fetches and reads before the transform runs,
	// so `calls` is incremented BEFORE the transform — that is what lets
	// TestQueueEdit_ValidationRefusesBeforeTouchingTheTrunk observe a refusal
	// that never reached the trunk.
	if f.writeErr != nil {
		return f.writeErr
	}
	// MODELS THE REAL RETRY. gitx.TrunkFile reads, transforms, then pushes as a
	// compare-and-swap, and on rejection re-reads and re-runs the transform. A
	// double that calls the transform exactly once cannot produce that, so a test
	// named "the peer's edit survives the replay" would pass without any replay
	// happening — the property would be asserted by the test's name only.
	for attempt := 1; attempt <= 3; attempt++ {
		base := append([]byte{}, f.content[path]...)
		f.calls++
		next, err := transform(base)
		if err != nil {
			return err
		}
		if f.calls == 1 && f.peer != nil {
			f.peer(f) // lands AFTER our read: the CAS will reject
		}
		if !bytes.Equal(f.content[path], base) {
			continue // the trunk moved under us — re-read and re-run
		}
		if bytes.Equal(base, next) {
			return nil // unchanged: the real Update pushes nothing
		}
		f.content[path] = next
		return nil
	}
	return errors.New("trunk moved under 3 attempts")
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
	// The replay must actually have happened. Without this the test passes on a
	// double that never retried, which is what it was doing before.
	if f.calls != 2 {
		t.Errorf("transform ran %d times, want 2 — no replay occurred, so the property is untested", f.calls)
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
		{"newline in tag", queue.Intent{Op: queue.OpAdd, Ref: "a#1", WhyNow: "x", Tag: "t\n- forged#2 — x"}},
		{"bracket in tag", queue.Intent{Op: queue.OpAdd, Ref: "a#1", WhyNow: "x", Tag: "a]b"}},
		{"bad anchor on move", queue.Intent{Op: queue.OpMove, Ref: "z#9", Anchor: "bad anchor"}},
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
			// The point of the test: the trunk is not CONTACTED, not merely
			// unmodified. The refusal path re-reads to render the handoff, so
			// allow that one read and assert no Update-side contact preceded it.
			if f.calls != 0 {
				t.Errorf("Update ran %d times — validation must refuse before any git call", f.calls)
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

// Converge MERGES rather than replaces: a re-add that omits --tag means "I did
// not mention the tag", not "remove it". Wholesale replacement silently dropped
// the tag and could flip an issue line into a project line, while the note
// claimed only the why-now had moved.
func TestQueueEdit_ConvergePreservesUnmentionedFields(t *testing.T) {
	f := newFakeTrunk("- a#1 — original [sdlc]\n")
	var out, errOut bytes.Buffer
	if err := runQueueEdit(&out, &errOut, f, queue.Intent{
		Op: queue.OpAdd, Ref: "a#1", WhyNow: "sharper reason"}); err != nil {
		t.Fatal(err)
	}
	got := string(f.content[queuePath])
	if !strings.Contains(got, "[sdlc]") {
		t.Errorf("the unmentioned tag was dropped: %q", got)
	}
	if !strings.Contains(got, "sharper reason") {
		t.Errorf("the why-now did not update: %q", got)
	}
	if n := errOut.String(); !strings.Contains(n, "why-now") || strings.Contains(n, "tag") {
		t.Errorf("the note must name exactly what changed, got: %s", n)
	}
}

// A line that looks like an entry but does not parse is preserved in the file —
// and must be REPORTED, not silently missing from the listing.
func TestQueueList_ReportsUnrecognizedItems(t *testing.T) {
	// A hyphen where the separator should be an em-dash: the classic hand-edit.
	f := newFakeTrunk("- a#1 — fine\n- b#2 - hyphen not em-dash\nplain prose\n")
	var out, errOut bytes.Buffer
	if err := runQueueList(&out, &errOut, f); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), "b#2") {
		t.Error("an unparsed line must not be listed as an entry")
	}
	w := errOut.String()
	if !strings.Contains(w, "1 line") {
		t.Errorf("the unrecognized line must be reported, got: %s", w)
	}
	if !strings.Contains(w, "em-dash") {
		t.Errorf("the warning must state the format so it is actionable, got: %s", w)
	}
	// Plain prose is NOT counted — warning about it would train the reader to
	// ignore the warning.
	if strings.Contains(w, "2 line") {
		t.Error("prose must not be counted as an unrecognized entry")
	}
}

// The spine guard on the write verbs is pinned here, because it is deliberately
// absent from processmanual.WorkflowVerbs (queue's writes are not lifecycle
// stages) and so no drift test enumerates it. Without this, removing the guard
// would be silent.
func TestQueueCmd_WriteVerbsAreSpineGuarded(t *testing.T) {
	cmd := NewQueueCmd()
	guarded := map[string]bool{"add": true, "remove": true, "move": true}
	seen := map[string]bool{}
	for _, sub := range cmd.Commands() {
		name := strings.Fields(sub.Use)[0]
		seen[name] = true
		src := subcommandGuardSource(t, name)
		if guarded[name] != strings.Contains(src, "guardSpineRepo") {
			t.Errorf("%q: guardSpineRepo present=%v, want %v", name, !guarded[name], guarded[name])
		}
	}
	for name := range guarded {
		if !seen[name] {
			t.Errorf("subcommand %q disappeared — the guard claim is now untested", name)
		}
	}
	// The bare list must NOT be guarded: reads stay unguarded by construction.
	if src := subcommandGuardSource(t, "queue-root"); strings.Contains(src, "guardSpineRepo") {
		t.Error("the bare list must stay unguarded, matching the charter's read carve-out")
	}
}

// subcommandGuardSource returns the RunE body for a queue subcommand, read from
// source. Reading source is the honest way to assert "this call is present"
// without a brain-repo fixture; the alternative — invoking the verb — would exit
// the process via die().
func subcommandGuardSource(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile("queue.go")
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	markers := map[string]string{
		"add": "func newQueueAddCmd", "remove": "func newQueueRemoveCmd",
		"move": "func newQueueMoveCmd", "queue-root": "func NewQueueCmd",
	}
	i := strings.Index(s, markers[name])
	if i < 0 {
		t.Fatalf("could not find %s in queue.go", markers[name])
	}
	rest := s[i:]
	if j := strings.Index(rest[1:], "\nfunc "); j >= 0 {
		rest = rest[:j]
	}
	return rest
}

// The guard must run BEFORE flag validation, so a brain repo gets the charter
// refusal rather than a complaint about flags on a command it may not run at
// all. A source-grep test cannot observe ordering, so this asserts position.
func TestQueueMove_GuardPrecedesFlagValidation(t *testing.T) {
	src := subcommandGuardSource(t, "move")
	g := strings.Index(src, "guardSpineRepo")
	v := strings.Index(src, "move needs exactly one of")
	if g < 0 || v < 0 {
		t.Fatalf("expected both the guard and the flag check in newQueueMoveCmd (guard=%d check=%d)", g, v)
	}
	if g > v {
		t.Error("guardSpineRepo must precede flag validation — otherwise a brain repo is told about flags, not about the charter")
	}
}

// A converged no-op must report honestly and push nothing. The commit subject is
// the most durable message this system emits, and "queue: add X" for an edit
// that did not happen is a permanent false claim on the trunk.
func TestQueueEdit_ConvergedNoOpReportsNoChange(t *testing.T) {
	f := newFakeTrunk("- a#1 — original [sdlc]\n")
	var out, errOut bytes.Buffer
	if err := runQueueEdit(&out, &errOut, f, queue.Intent{
		Op: queue.OpAdd, Ref: "a#1", WhyNow: "original", Tag: "sdlc"}); err != nil {
		t.Fatal(err)
	}
	s := errOut.String()
	if !strings.Contains(s, "no change") || !strings.Contains(s, "nothing pushed") {
		t.Errorf("a no-op must say so, got: %s", s)
	}
	if strings.Contains(s, "[ok] queue: add a#1") {
		t.Error("must not report an edit it did not make")
	}
}

// --project must be distinguishable from "not mentioned" at the CLI layer. The
// type-level KindUnspecified was unreachable while the command hardcoded
// KindIssue, so the fix was dead code and a re-add still destroyed a peer's
// project marker.
func TestQueueAddCmd_UnmentionedProjectFlagIsUnspecified(t *testing.T) {
	src := subcommandGuardSource(t, "add")
	if !strings.Contains(src, "KindUnspecified") {
		t.Error("add must start from KindUnspecified so an omitted --project means 'not mentioned'")
	}
	if !strings.Contains(src, `Changed("project")`) {
		t.Error("add must consult Flags().Changed, not the bool's zero value")
	}
}
