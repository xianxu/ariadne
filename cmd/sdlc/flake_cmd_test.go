package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flakes"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

const (
	flakeRepo = "xianxu/ariadne"
	flakeT1   = "github.com/xianxu/ariadne/scripts/test.TestConstructScripts"
	flakeT2   = "github.com/xianxu/ariadne/cmd/sdlc.TestOther"
)

// flakeWorld is a stateful fake of the verbs' seams: a store of attempts,
// tracker cards with flake_tests, and filed issues.
type flakeWorld struct {
	attempts []SuiteAttempt
	cards    []flakes.Card
	filed    []string // bodies of issues filed
	regErr   error
	// raceAdd is a registration another slot lands on a card between this
	// verb's registry read and its write.
	raceAdd map[string]string
	// logSubtests are the failing subtests each attempt's saved log names.
	logSubtests map[string][]string
}

func (w *flakeWorld) deps() flakeDeps {
	now := time.Date(2026, 10, 19, 12, 0, 0, 0, time.UTC)
	return flakeDeps{
		now:      func() time.Time { return now },
		attempts: func() ([]SuiteAttempt, []string, error) { return w.attempts, nil, nil },
		repo:     func() (string, error) { return flakeRepo, nil },
		logPath:  func(a SuiteAttempt) string { return "/store/failed/" + a.ID + ".log" },
		subtests: func(a SuiteAttempt) ([]string, error) {
			subs, ok := w.logSubtests[a.ID]
			if !ok {
				return nil, errors.New("no log kept")
			}
			return subs, nil
		},
		registry: func(context.Context) (flakes.Registry, []flakes.Card, error) {
			if w.regErr != nil {
				return flakes.Registry{}, nil, w.regErr
			}
			snapshot := append([]flakes.Card(nil), w.cards...)
			for i := range w.cards {
				if t, ok := w.raceAdd[w.cards[i].ID]; ok {
					w.cards[i].Tests = append(append([]string(nil), w.cards[i].Tests...), t)
				}
			}
			r, err := flakes.BuildRegistry(snapshot, flakeIssueOpen)
			return r, snapshot, err
		},
		addTests: func(_ context.Context, _, _ io.Writer, id int, add []string) error {
			for i := range w.cards {
				if w.cards[i].ID != sixDigits(id) {
					continue
				}
				tests, err := mergeFlakeTests(w.cards[i].ID, w.cards[i].Status, w.cards[i].Tests, add) // the production decision
				if err != nil {
					return err
				}
				w.cards[i].Tests = tests
				return nil
			}
			return errors.New("no such card")
		},
		editTests: func(_ context.Context, _, _ io.Writer, id int, _ string, edit func(id, status string, have []string) ([]string, error)) error {
			for i := range w.cards {
				if w.cards[i].ID == sixDigits(id) {
					tests, err := edit(w.cards[i].ID, w.cards[i].Status, w.cards[i].Tests)
					if err == nil {
						w.cards[i].Tests = tests
					}
					return err
				}
			}
			return errors.New("no such card")
		},
		newIssue: func(_ context.Context, _, _ io.Writer, _, body string) (int, error) {
			w.filed = append(w.filed, body)
			id := 600 + len(w.filed)
			w.cards = append(w.cards, flakes.Card{ID: sixDigits(id), Status: "open"})
			return id, nil
		},
	}
}

func sixDigits(n int) string { return fmt.Sprintf("%06d", n) }

func attemptAt(id, sha string, minute, exit int, failures ...string) SuiteAttempt {
	at := time.Date(2026, 10, 18, 10, minute, 0, 0, time.UTC)
	return SuiteAttempt{ID: id, Repo: flakeRepo, SHA: sha, Check: suiteCheckName, StartedAt: at, FinishedAt: at, Exit: exit, Failures: failures}
}

func TestFlakeRegisterNeedsSameHeadEvidence(t *testing.T) {
	w := &flakeWorld{cards: []flakes.Card{{ID: "000448", Status: "open"}}}
	w.attempts = []SuiteAttempt{attemptAt("a1", "s1", 1, 1, flakeT1)}
	var out, errb bytes.Buffer
	err := runFlakeRegister(context.Background(), &out, &errb, flakeRegisterFlags{Issue: 448, Tests: []string{flakeT1}}, w.deps())
	if err == nil || !strings.Contains(err.Error(), "no same-head fail→pass evidence") {
		t.Fatalf("registered without evidence: %v", err)
	}
	w.attempts = append(w.attempts, attemptAt("a2", "s1", 2, 0))
	if err := runFlakeRegister(context.Background(), &out, &errb, flakeRegisterFlags{Issue: 448, Tests: []string{flakeT1}}, w.deps()); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(w.cards[0].Tests, []string{flakeT1}) || !strings.Contains(out.String(), "failed in attempt a1, passed on the same head in attempt a2") {
		t.Fatalf("card %v\n%s", w.cards[0].Tests, out.String())
	}
}

func TestFlakeRegisterRules(t *testing.T) {
	w := &flakeWorld{cards: []flakes.Card{{ID: "000448", Status: "open", Tests: []string{flakeT1}}, {ID: "000449", Status: "open"}}}
	run := func(f flakeRegisterFlags) error {
		return runFlakeRegister(context.Background(), io.Discard, io.Discard, f, w.deps())
	}
	if err := run(flakeRegisterFlags{Issue: 449, Tests: []string{flakeT1}, Evidence: "ci run 1"}); err == nil || !strings.Contains(err.Error(), "already registered on issue #000448") {
		t.Errorf("one owner per test: %v", err)
	}
	if err := run(flakeRegisterFlags{Issue: 448, Tests: []string{flakeT2}, Evidence: "ci run 2"}); err != nil || !slices.Equal(w.cards[0].Tests, []string{flakeT1, flakeT2}) {
		t.Errorf("adding to an issue keeps its tests: %v %v", err, w.cards[0].Tests)
	}
	if err := run(flakeRegisterFlags{Tests: []string{"TestBare"}, Evidence: "x"}); err == nil {
		t.Error("a bare test name registered")
	}
	if err := run(flakeRegisterFlags{Tests: []string{"github.com/xianxu/ariadne/pkg/x.TestNew"}, Evidence: "ci run 3"}); err != nil {
		t.Fatal(err)
	}
	if len(w.filed) != 1 || !strings.Contains(w.filed[0], "pkg/x.TestNew: ci run 3") || !slices.Equal(w.cards[2].Tests, []string{"github.com/xianxu/ariadne/pkg/x.TestNew"}) {
		t.Errorf("filed %q, cards %+v", w.filed, w.cards)
	}
	w.regErr = errors.New("tracker unreachable")
	if err := run(flakeRegisterFlags{Issue: 449, Tests: []string{flakeT2}, Evidence: "x"}); err == nil {
		t.Error("registered while the registry is unreadable")
	}
}

// Another slot registers on the same card between this verb's read and its
// write: both registrations survive (#394 M1 review, BR-2).
func TestFlakeRegisterKeepsARacingRegistration(t *testing.T) {
	racer := "github.com/xianxu/ariadne/pkg/r.TestRacer"
	w := &flakeWorld{cards: []flakes.Card{{ID: "000448", Status: "open", Tests: []string{flakeT1}}}, raceAdd: map[string]string{"000448": racer}}
	if err := runFlakeRegister(context.Background(), io.Discard, io.Discard, flakeRegisterFlags{Issue: 448, Tests: []string{flakeT2}, Evidence: "ci"}, w.deps()); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(w.cards[0].Tests, []string{flakeT1, racer, flakeT2}) {
		t.Fatalf("a racing registration was lost: %v", w.cards[0].Tests)
	}
	w.cards = append(w.cards, flakes.Card{ID: "000450", Status: "done"})
	if err := runFlakeRegister(context.Background(), io.Discard, io.Discard, flakeRegisterFlags{Issue: 450, Tests: []string{"github.com/xianxu/ariadne/pkg/z.TestZ"}, Evidence: "ci"}, w.deps()); err == nil {
		t.Error("registered on a closed issue")
	}
}

func TestFlakeList(t *testing.T) {
	w := &flakeWorld{cards: []flakes.Card{{ID: "000448", Status: "working", Tests: []string{flakeT1}, Owner: "ariadne:3", Created: "2026-10-16"}}}
	w.attempts = []SuiteAttempt{
		attemptAt("a1", "s1", 1, 1, flakeT2), attemptAt("a2", "s1", 2, 0), // a same-head flake nobody registered
		attemptAt("a3", "s2", 3, 1, "github.com/xianxu/ariadne/pkg/y.TestBroken"), // unexplained, over a day old
	}
	var out bytes.Buffer
	if err := runFlakeList(context.Background(), &out, w.deps()); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		flakeT1 + "  #000448 (working, owner ariadne:3, open 3d)",
		flakeT2 + "  attempt a1 on s1, 2026-10-18T10:01:00Z, log /store/failed/a1.log",
		"attempt a3 on s2, 2026-10-18T10:03:00Z: github.com/xianxu/ariadne/pkg/y.TestBroken, log /store/failed/a3.log",
		"CI failures: not covered (#451)",
	} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("missing %q:\n%s", want, out.String())
		}
	}
	if strings.Contains(out.String(), "attempt a1 on s1, 2026-10-18T10:01:00Z: ") {
		t.Errorf("a confirmed flake listed as unexplained:\n%s", out.String())
	}
	w.regErr = errors.New("tracker unreachable")
	out.Reset()
	if err := runFlakeList(context.Background(), &out, w.deps()); err != nil || !strings.Contains(out.String(), "unavailable: tracker unreachable") {
		t.Errorf("an unreadable registry: %v\n%s", err, out.String())
	}
}

// The registry comes from the real tracker cards (#394 D4): an open card's
// flake_tests register, a closed card's don't, and the setter writes the field
// on any card with no owner check.
func TestFlakeRegistryFromTrackerCards(t *testing.T) {
	openPath, openCard, openDetailPath, openDetail := seededIssue(t, "000051", "flaky")
	donePath, doneCard, doneDetailPath, doneDetail := seededIssue(t, "000052", "fixed")
	openCard = strings.Replace(openCard, "status: open", "status: working\nflake_tests: ["+flakeT1+"]", 1)
	doneCard = strings.Replace(doneCard, "status: open", "status: done\nactual_hours: 1\nflake_tests: ["+flakeT2+"]", 1)
	newTrackerRepo(t, map[string]string{openPath: openCard, donePath: doneCard},
		map[string]string{openDetailPath: openDetail, doneDetailPath: doneDetail})
	ctx := withIssueRecordsScope(context.Background())
	reg, _, err := loadFlakeRegistry(ctx, "workshop/issues")
	if err != nil {
		t.Fatal(err)
	}
	if id, ok := reg.Lookup(flakeT1); !ok || id != "000051" {
		t.Errorf("an open card's test: %q %v", id, ok)
	}
	if _, ok := reg.Lookup(flakeT2); ok {
		t.Error("a closed card's test is registered")
	}
	d := defaultFlakeDeps("workshop/issues")
	if err := d.addTests(ctx, io.Discard, io.Discard, 51, []string{flakeT2}); err != nil {
		t.Fatal(err)
	}
	if err := d.addTests(ctx, io.Discard, io.Discard, 52, []string{flakeT1}); err == nil || !strings.Contains(err.Error(), "a closed issue registers nothing") {
		t.Errorf("the setter on a closed card: %v", err)
	}
	invalidateIssueRecords(ctx)
	if reg, _, err = loadFlakeRegistry(ctx, "workshop/issues"); err != nil {
		t.Fatal(err)
	}
	if id, ok := reg.Lookup(flakeT2); !ok || id != "000051" {
		t.Errorf("after the setter: %q %v", id, ok)
	}
	if id, ok := reg.Lookup(flakeT1); !ok || id != "000051" {
		t.Errorf("the setter dropped the card's existing test: %q %v (BR-2)", id, ok)
	}
	// A repeat of the same registration writes nothing (its recovery contract).
	var out strings.Builder
	if err := d.addTests(ctx, &out, io.Discard, 51, []string{flakeT2}); err != nil {
		t.Errorf("a repeated registration: %v", err)
	}
}

// The setter's decision on a real card: the tests already there survive, the
// cap holds against them, and a closed card refuses (#394 M1 review, BR-2).
func TestAddFlakeTestsMergesOntoTheCard(t *testing.T) {
	path, card, _, _ := seededIssue(t, "000061", "flaky")
	parse := func(raw string) tracker.Record {
		t.Helper()
		c, err := issue.ParseCard([]byte(raw))
		if err != nil {
			t.Fatal(err)
		}
		return tracker.Record{ID: "000061", Path: path, Card: c, Raw: []byte(raw)}
	}
	withTests := strings.Replace(card, "status: open", "status: working\nflake_tests: ["+flakeT1+"]", 1)
	out, err := addFlakeTests(parse(withTests), []string{flakeT2})
	if err != nil {
		t.Fatal(err)
	}
	got, _, err := issue.FieldList(parse(string(out)).Card.Frontmatter, flakeTestsField)
	if err != nil || !slices.Equal(got, []string{flakeT1, flakeT2}) {
		t.Fatalf("merged %v %v, want the existing test kept", got, err)
	}
	five := "github.com/x/p.TestA, github.com/x/p.TestB, github.com/x/p.TestC, github.com/x/p.TestD, github.com/x/p.TestE"
	full := strings.Replace(card, "status: open", "status: working\nflake_tests: ["+five+"]", 1)
	if _, err := addFlakeTests(parse(full), []string{flakeT2}); err == nil || !strings.Contains(err.Error(), "at most 5") {
		t.Errorf("the cap counts the card's tests: %v", err)
	}
	closed := strings.Replace(card, "status: open", "status: done\nactual_hours: 1", 1)
	if _, err := addFlakeTests(parse(closed), []string{flakeT2}); err == nil || !strings.Contains(err.Error(), "a closed issue registers nothing") {
		t.Errorf("a closed card: %v", err)
	}
}

const flakeSub = flakeT1 + "/plain/dev-aliases.test.sh"

func TestRemoveAndNarrowFlakeTestsDecisions(t *testing.T) {
	have := []string{flakeT1, flakeT2}
	got, err := removeFlakeTests("000465", "working", have, flakeT1)
	if err != nil || !slices.Equal(got, []string{flakeT2}) {
		t.Fatalf("remove: %v %v", got, err)
	}
	if again, err := removeFlakeTests("000465", "working", got, flakeT1); err != nil || !slices.Equal(again, got) {
		t.Fatalf("a repeated remove changed the list: %v %v", again, err)
	}
	got, err = narrowFlakeTests("000465", "working", have, flakeT1, flakeSub)
	if err != nil || !slices.Equal(got, []string{flakeSub, flakeT2}) {
		t.Fatalf("narrow: %v %v", got, err)
	}
	if again, err := narrowFlakeTests("000465", "working", got, flakeT1, flakeSub); err != nil || !slices.Equal(again, got) {
		t.Fatalf("a repeated narrow changed the list: %v %v", again, err)
	}
	for name, c := range map[string]struct {
		have         []string
		old, to, why string
	}{
		"not a subtest":     {have, flakeT1, flakeT2, "not a subtest"},
		"the same name":     {have, flakeT1, flakeT1, "not a subtest"},
		"a widening":        {[]string{flakeSub}, flakeSub, flakeT1, "not a subtest"},
		"nothing to narrow": {[]string{flakeT2}, flakeT1, flakeSub, "nothing to narrow"},
	} {
		if _, err := narrowFlakeTests("000465", "working", c.have, c.old, c.to); err == nil || !strings.Contains(err.Error(), c.why) {
			t.Errorf("%s: %v", name, err)
		}
	}
	if _, err := removeFlakeTests("000465", "done", have, flakeT1); err == nil {
		t.Error("a closed card was edited")
	}
}

// A subtest registers with its top-level test's same-head fail→pass plus the
// failed attempt's log naming it; without that log line it does not.
func TestFlakeRegisterASubtest(t *testing.T) {
	w := &flakeWorld{
		attempts:    []SuiteAttempt{attemptAt("f1", "abc", 1, 1, flakeT1), attemptAt("p1", "abc", 2, 0)},
		cards:       []flakes.Card{{ID: "000465", Status: "working"}},
		logSubtests: map[string][]string{"f1": {flakeT1 + "/plain", flakeSub}},
	}
	var out bytes.Buffer
	if err := runFlakeRegister(context.Background(), &out, io.Discard, flakeRegisterFlags{Issue: 465, Tests: []string{flakeSub}}, w.deps()); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(w.cards[0].Tests, []string{flakeSub}) || !strings.Contains(out.String(), "failed in attempt f1") {
		t.Fatalf("card %v out %s", w.cards[0].Tests, out.String())
	}
	sibling := flakeT1 + "/plain/other.test.sh"
	if err := runFlakeRegister(context.Background(), io.Discard, io.Discard, flakeRegisterFlags{Issue: 465, Tests: []string{sibling}}, w.deps()); err == nil || !strings.Contains(err.Error(), "no same-head") {
		t.Fatalf("a subtest the failed log does not name registered: %v", err)
	}
	w.logSubtests = nil
	w.cards[0].Tests = nil
	if err := runFlakeRegister(context.Background(), io.Discard, io.Discard, flakeRegisterFlags{Issue: 465, Tests: []string{flakeSub}}, w.deps()); err == nil {
		t.Fatal("a subtest registered without its failed log")
	}
	// One issue owns a test and its subtests.
	w.cards = []flakes.Card{{ID: "000465", Status: "working", Tests: []string{flakeT1}}, {ID: "000466", Status: "working"}}
	w.logSubtests = map[string][]string{"f1": {flakeSub}}
	if err := runFlakeRegister(context.Background(), io.Discard, io.Discard, flakeRegisterFlags{Issue: 466, Tests: []string{flakeSub}}, w.deps()); err == nil || !strings.Contains(err.Error(), "overlaps") {
		t.Fatalf("an overlapping registration on another issue: %v", err)
	}
}

func TestFlakeUnregisterAndNarrowVerbs(t *testing.T) {
	w := &flakeWorld{cards: []flakes.Card{{ID: "000465", Status: "working", Tests: []string{flakeT1, flakeT2}}}}
	ctx := context.Background()
	if err := runFlakeEdit(ctx, io.Discard, io.Discard, flakeEditFlags{Issue: 465, Test: flakeT1, To: flakeSub}, true, w.deps()); err != nil {
		t.Fatal(err)
	}
	if err := runFlakeEdit(ctx, io.Discard, io.Discard, flakeEditFlags{Issue: 465, Test: flakeT2}, false, w.deps()); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(w.cards[0].Tests, []string{flakeSub}) {
		t.Fatalf("card %v", w.cards[0].Tests)
	}
	for name, f := range map[string]flakeEditFlags{
		"no issue":   {Test: flakeT1},
		"a bad name": {Issue: 465, Test: "TestX"},
		"a bad --to": {Issue: 465, Test: flakeT1, To: "x; rm"},
	} {
		if err := runFlakeEdit(ctx, io.Discard, io.Discard, f, f.To != "", w.deps()); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}

// The verbs on real tracker cards: owner-only (anyone while unclaimed), a
// concurrent registration kept, and a repeat that writes nothing.
func TestFlakeEditOnTrackerCards(t *testing.T) {
	unclaimedPath, unclaimed, unclaimedDetailPath, unclaimedDetail := seededIssue(t, "000071", "broad")
	foreignPath, foreign, foreignDetailPath, foreignDetail := seededIssue(t, "000072", "theirs")
	unclaimed = strings.Replace(unclaimed, "status: open", "status: working\nflake_tests: ["+flakeT1+", "+flakeT2+"]", 1)
	foreign = strings.Replace(foreign, "status: open", "status: working\nflake_tests: ["+flakeT2+"]\nclaimant:\n    operator: Someone Else\n    machine: 0000000000000000000000000000beef\n    machine_name: elsewhere\n    worktree: /elsewhere/repo\n    repository: github.com/other/repo", 1)
	newTrackerRepo(t, map[string]string{unclaimedPath: unclaimed, foreignPath: foreign},
		map[string]string{unclaimedDetailPath: unclaimedDetail, foreignDetailPath: foreignDetail})
	ctx := withIssueRecordsScope(context.Background())
	d := defaultFlakeDeps("workshop/issues")
	if err := runFlakeEdit(ctx, io.Discard, io.Discard, flakeEditFlags{Issue: 71, Test: flakeT1, To: flakeSub}, true, d); err != nil {
		t.Fatalf("narrow on an unclaimed card: %v", err)
	}
	if err := runFlakeEdit(ctx, io.Discard, io.Discard, flakeEditFlags{Issue: 71, Test: flakeT1, To: flakeSub}, true, d); err != nil {
		t.Fatalf("a repeated narrow: %v", err)
	}
	invalidateIssueRecords(ctx)
	reg, _, err := loadFlakeRegistry(ctx, "workshop/issues")
	if err == nil {
		t.Fatalf("flakeT2 on two issues should be ambiguous; registry %+v", reg.Entries())
	}
	if err := runFlakeEdit(ctx, io.Discard, io.Discard, flakeEditFlags{Issue: 72, Test: flakeT2}, false, d); err == nil || !strings.Contains(err.Error(), "only its owner unregisters or narrows") {
		t.Fatalf("an edit of another workspace's card: %v", err)
	}
	if err := runFlakeEdit(ctx, io.Discard, io.Discard, flakeEditFlags{Issue: 71, Test: flakeT2}, false, d); err != nil {
		t.Fatalf("unregister on an unclaimed card: %v", err)
	}
	invalidateIssueRecords(ctx)
	if reg, _, err = loadFlakeRegistry(ctx, "workshop/issues"); err != nil {
		t.Fatal(err)
	}
	if id, ok := reg.Lookup(flakeSub); !ok || id != "000071" {
		t.Fatalf("narrowed name: %q %v", id, ok)
	}
	if _, ok := reg.Lookup(flakeT1); ok {
		t.Fatal("the broad name survived narrowing")
	}
	if id, ok := reg.Lookup(flakeT2); !ok || id != "000072" {
		t.Fatalf("the other card's test: %q %v", id, ok)
	}
	// The last name unregistered clears the field: the card stays valid.
	if err := runFlakeEdit(ctx, io.Discard, io.Discard, flakeEditFlags{Issue: 71, Test: flakeSub}, false, d); err != nil {
		t.Fatalf("unregister the last name: %v", err)
	}
	invalidateIssueRecords(ctx)
	if reg, _, err = loadFlakeRegistry(ctx, "workshop/issues"); err != nil {
		t.Fatalf("a card with its last flake removed is unreadable: %v", err)
	}
	if _, ok := reg.Lookup(flakeSub); ok {
		t.Fatal("the last name survived")
	}
}
