package flakes

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
)

const (
	tA = "github.com/x/r/pkg/a.TestA"
	tB = "github.com/x/r/pkg/b.TestB"
	tC = "github.com/x/r/pkg/a.TestC"
)

func open(s string) bool { return s == "open" || s == "working" }

func registry(t *testing.T) Registry {
	t.Helper()
	r, err := BuildRegistry([]Card{
		{ID: "000500", Status: "working", Tests: []string{tA}},
		{ID: "000501", Status: "open", Tests: []string{tB}},
		{ID: "000502", Status: "done", Tests: []string{tC}}, // closed: unregistered
	}, open)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestBuildRegistry(t *testing.T) {
	r := registry(t)
	if id, ok := r.Lookup(tA); !ok || id != "000500" {
		t.Errorf("TestA: %q %v", id, ok)
	}
	if _, ok := r.Lookup(tC); ok {
		t.Error("a closed issue's test is registered")
	}
	for name, cards := range map[string][]Card{
		"ambiguous":      {{ID: "1", Status: "open", Tests: []string{tA}}, {ID: "2", Status: "open", Tests: []string{tA}}},
		"a regex":        {{ID: "1", Status: "open", Tests: []string{"github.com/x/r/pkg.Test.*"}}},
		"no package":     {{ID: "1", Status: "open", Tests: []string{"TestA"}}},
		"a subtest":      {{ID: "1", Status: "open", Tests: []string{"github.com/x/r/pkg.TestA/sub"}}},
		"not a Test fn":  {{ID: "1", Status: "open", Tests: []string{"github.com/x/r/pkg.Helper"}}},
		"shell metachar": {{ID: "1", Status: "open", Tests: []string{"github.com/x/r/p;rm.TestA"}}},
	} {
		if _, err := BuildRegistry(cards, open); err == nil {
			t.Errorf("%s: no refusal", name)
		}
	}
	if _, err := BuildRegistry([]Card{{ID: "1", Status: "done", Tests: []string{tA}}, {ID: "2", Status: "open", Tests: []string{tA}}}, open); err != nil {
		t.Errorf("a closed issue and an open one naming the same test: %v", err)
	}
}

// Every row of the plan's state table.
func TestClassify(t *testing.T) {
	r := registry(t)
	fail := func(names ...string) Result { return Result{Failures: names} }
	pass := &Result{Passed: true}
	for _, tc := range []struct {
		name string
		in   Input
		kind Kind
		why  string
	}{
		{"a pass passes", Input{Attempt: Result{Passed: true}, Registry: r}, Passed, ""},
		{"registered failures: rerun them", Input{Attempt: fail(tA, tB), Registry: r}, NeedRerun, "rerunning " + tA},
		{"the rerun passed", Input{Attempt: fail(tA), Registry: r, Rerun: pass}, KnownFlakePassed, ""},
		{"the rerun failed", Input{Attempt: fail(tA), Registry: r, Rerun: &Result{Failures: []string{tA}}}, KnownFlakeTwice, "failed twice"},
		{"policy accept", Input{Attempt: fail(tA), Registry: r, Policy: PolicyAccept}, KnownFlakeAccept, ""},
		{"an unregistered failure blocks", Input{Attempt: fail(tA, tC), Registry: r}, Unknown, "unknown failure: " + tC},
		{"unattributable blocks", Input{Attempt: Result{Failures: []string{tA}, Unattributed: []string{"shard1 exited 2 with no failing test"}}, Registry: r}, Unattributable, "shard1"},
		{"a failure with no names blocks", Input{Attempt: Result{}, Registry: r}, Unattributable, "without naming"},
		{"also failing on main", Input{Attempt: fail(tC), Registry: r, MainBaseline: &Result{Failures: []string{tC}}}, PersistentMain, "persistent main failure: " + tC},
		{"main passed: just unknown", Input{Attempt: fail(tC), Registry: r, MainBaseline: pass}, Unknown, ""},
		{"an unreadable registry is not empty", Input{Attempt: fail(tA), RegistryErr: errors.New("tracker unreachable")}, RegistryUnusable, "tracker unreachable"},
		{"policy off blocks", Input{Attempt: fail(tA), Registry: r, Policy: PolicyOff}, RuleOff, ""},
	} {
		got := Classify(tc.in)
		if got.Kind != tc.kind || !strings.Contains(got.Reason, tc.why) {
			t.Errorf("%s: %s %q, want %s containing %q", tc.name, got.Kind, got.Reason, tc.kind, tc.why)
		}
		if got.Kind.Lands() != (tc.kind == Passed || tc.kind == KnownFlakePassed || tc.kind == KnownFlakeAccept) {
			t.Errorf("%s: Lands() wrong", tc.name)
		}
	}
	many := make([]string, 0, MaxRerun+1)
	var cards []Card
	for i := 0; i <= MaxRerun; i++ {
		name := "github.com/x/r/p.TestN" + string(rune('a'+i))
		many = append(many, name)
		cards = append(cards, Card{ID: "9", Status: "open", Tests: []string{name}})
	}
	big, err := BuildRegistry(cards, open)
	if err != nil {
		t.Fatal(err)
	}
	if got := Classify(Input{Attempt: fail(many...), Registry: big}); got.Kind != TooManyKnownFlaky {
		t.Errorf("more than %d registered failures: %s", MaxRerun, got.Kind)
	}
	if got := Classify(Input{Attempt: fail(tA, tB), Registry: r}); !reflect.DeepEqual(got.Issues(), []string{"000500", "000501"}) {
		t.Errorf("issues %v", got.Issues())
	}
}

func TestEvidence(t *testing.T) {
	at := func(m int) time.Time { return time.Date(2026, 10, 18, 12, m, 0, 0, time.UTC) }
	failed := Attempt{ID: "1", SHA: "s1", Failures: []string{tC}, FinishedAt: at(1)}
	attempts := []Attempt{
		{ID: "0", SHA: "s1", Passed: true, FinishedAt: at(0)}, // an earlier pass proves nothing
		failed,
		{ID: "b", SHA: "s1", Passed: true, Baseline: true, FinishedAt: at(2)}, // a baseline is not this head's run
		{ID: "o", SHA: "s2", Passed: true, FinishedAt: at(3)},                 // another head
	}
	if _, _, ok := SameHeadFailPass(attempts, tC); ok {
		t.Error("evidence without a later same-head pass")
	}
	attempts = append(attempts, Attempt{ID: "2", SHA: "s1", Passed: true, FinishedAt: at(4)})
	if f, p, ok := SameHeadFailPass(attempts, tC); !ok || f != "1" || p != "2" {
		t.Errorf("evidence %q %q %v", f, p, ok)
	}

	r := registry(t)
	if u := Unexplained(attempts, "s1", r); len(u) != 1 || u[0].ID != "1" {
		t.Errorf("an unknown failure followed by a full rerun pass is still unexplained: %+v", u)
	}
	r2, _ := BuildRegistry([]Card{{ID: "7", Status: "open", Tests: []string{tC}}}, open)
	if u := Unexplained(attempts, "s1", r2); len(u) != 0 {
		t.Errorf("registered after the fact: %+v", u)
	}
	known := append(attempts, Attempt{ID: "3", SHA: "s3", Failures: []string{tA}, Kind: KnownFlakePassed, FinishedAt: at(5)},
		Attempt{ID: "4", SHA: "s4", Failures: []string{tA}, Kind: Unattributable, FinishedAt: at(6)})
	if u := Unexplained(known, "s3", r); len(u) != 0 {
		t.Errorf("a recorded known-flake pass: %+v", u)
	}
	if u := Unexplained(known, "s4", r); len(u) != 1 {
		t.Errorf("unattributable is never explained by registration: %+v", u)
	}
}

func TestRerunCommand(t *testing.T) {
	got, err := RerunCommand([]string{tA, tC, tB})
	want := []string{"go", "test", "-count=1", "-run", "^(TestA|TestC|TestB)$", "github.com/x/r/pkg/a", "github.com/x/r/pkg/b"}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Errorf("%v %v", got, err)
	}
	for _, bad := range [][]string{nil, {"TestA"}, {"github.com/x.Test$(rm)"}} {
		if _, err := RerunCommand(bad); err == nil {
			t.Errorf("%q: no refusal", bad)
		}
	}
	if p, _ := ParsePolicy(""); p != PolicyRerun {
		t.Errorf("default policy %q", p)
	}
	if _, err := ParsePolicy("sometimes"); err == nil {
		t.Error("a bad policy parsed")
	}
}
