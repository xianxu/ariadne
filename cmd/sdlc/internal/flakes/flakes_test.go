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
		"ambiguous":                 {{ID: "1", Status: "open", Tests: []string{tA}}, {ID: "2", Status: "open", Tests: []string{tA}}},
		"a regex":                   {{ID: "1", Status: "open", Tests: []string{"github.com/x/r/pkg.Test.*"}}},
		"no package":                {{ID: "1", Status: "open", Tests: []string{"TestA"}}},
		"a subtest with a space":    {{ID: "1", Status: "open", Tests: []string{"github.com/x/r/pkg.TestA/a b"}}},
		"a subtest parent segment":  {{ID: "1", Status: "open", Tests: []string{"github.com/x/r/pkg.TestA/../TestB"}}},
		"a subtest with shell text": {{ID: "1", Status: "open", Tests: []string{"github.com/x/r/pkg.TestA/$(rm)"}}},
		"an empty subtest segment":  {{ID: "1", Status: "open", Tests: []string{"github.com/x/r/pkg.TestA//sub"}}},
		"not a Test fn":             {{ID: "1", Status: "open", Tests: []string{"github.com/x/r/pkg.Helper"}}},
		"shell metachar":            {{ID: "1", Status: "open", Tests: []string{"github.com/x/r/p;rm.TestA"}}},
		"a flag":                    {{ID: "1", Status: "open", Tests: []string{"-exec.TestA"}}},
		"a flag segment":            {{ID: "1", Status: "open", Tests: []string{"github.com/-exec.TestA"}}},
		"a relative pkg":            {{ID: "1", Status: "open", Tests: []string{"./x.TestB"}}},
		"a parent seg":              {{ID: "1", Status: "open", Tests: []string{"github.com/x/../y.TestB"}}},
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
	for _, k := range []Kind{KnownFlakeTwice, TooManyKnownFlaky, RuleOff} {
		judged := append(attempts, Attempt{ID: "5", SHA: "s5", Failures: []string{tA}, Kind: k, FinishedAt: at(7)}, Attempt{ID: "6", SHA: "s5", Passed: true, FinishedAt: at(8)})
		if u := Unexplained(judged, "s5", r); len(u) != 1 {
			t.Errorf("%s then a full rerun pass: a rejected failure is explained by registration: %+v", k, u)
		}
	}
}

func TestRerunCommand(t *testing.T) {
	got, err := RerunCommand([]string{tA, tC, tB})
	want := []string{"go", "test", "-count=1", "-v", "-run", "^(TestA|TestC|TestB)$", "github.com/x/r/pkg/a", "github.com/x/r/pkg/b"}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Errorf("%v %v", got, err)
	}
	for _, bad := range [][]string{nil, {"TestA"}, {"github.com/x.Test$(rm)"}, {"-exec.TestA"}, {"./x.TestB"}, {"x/.hidden.TestC"}} {
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

// A landing is rechecked just before it is posted (#394 M2 review, BR-12).
func TestRevalidate(t *testing.T) {
	r := registry(t)
	landed := Outcome{Kind: KnownFlakePassed, Known: []Entry{{Test: tA, Issue: "000500"}}}
	if got := Revalidate(landed, r, nil, PolicyRerun); got.Kind != KnownFlakePassed {
		t.Errorf("still valid: %s", got.Kind)
	}
	closed, _ := BuildRegistry(nil, open)
	if got := Revalidate(landed, closed, nil, PolicyRerun); got.Kind != Unknown || !strings.Contains(got.Reason, "no longer a known flake") {
		t.Errorf("the issue closed during the run: %s %q", got.Kind, got.Reason)
	}
	if got := Revalidate(landed, r, nil, PolicyOff); got.Kind != RuleOff {
		t.Errorf("the rule turned off: %s", got.Kind)
	}
	if got := Revalidate(landed, r, errors.New("tracker down"), PolicyRerun); got.Kind != RegistryUnusable {
		t.Errorf("an unreadable registry: %s", got.Kind)
	}
	accepted := Outcome{Kind: KnownFlakeAccept, Known: landed.Known}
	if got := Revalidate(accepted, r, nil, PolicyRerun); got.Kind != Unknown {
		t.Errorf("accept no longer allowed: %s", got.Kind)
	}
	if got := Revalidate(Outcome{Kind: Unknown}, r, nil, PolicyRerun); got.Kind != Unknown {
		t.Errorf("a non-landing passes through: %s", got.Kind)
	}
}

// A rerun proves only what its log shows passing (#394 M2 review).
func TestRerunEvidence(t *testing.T) {
	log := "=== RUN   TestA\n--- PASS: TestA (0.01s)\nok  github.com/x/r/pkg/a 0.1s\n"
	if m := RerunMissing(log, []string{tA}); len(m) != 0 {
		t.Errorf("TestA passed: missing %v", m)
	}
	if m := RerunMissing("ok  github.com/x/r/pkg/a 0.1s [no tests to run]\n", []string{tA}); len(m) != 1 {
		t.Errorf("no tests ran: missing %v", m)
	}
	got := Classify(Input{Attempt: Result{Failures: []string{tA}}, Registry: registry(t), Rerun: &Result{Passed: true, Unattributed: []string{"the rerun timed out"}}})
	if got.Kind != RerunInconclusive || got.Kind.Lands() {
		t.Errorf("an inconclusive rerun: %s", got.Kind)
	}
}

func TestParseName(t *testing.T) {
	for name, want := range map[string]Name{
		"github.com/x/r/pkg.TestA": {Pkg: "github.com/x/r/pkg", Test: "TestA"},
		"github.com/xianxu/ariadne/scripts/test.TestConstructScripts/plain/dev-aliases.test.sh": {Pkg: "github.com/xianxu/ariadne/scripts/test", Test: "TestConstructScripts", Sub: "plain/dev-aliases.test.sh"},
		"gopkg.in/yaml.v3.TestUnmarshal/case#01":                                                {Pkg: "gopkg.in/yaml.v3", Test: "TestUnmarshal", Sub: "case#01"},
		// A subtest that looks like a test never moves the package boundary.
		"github.com/x/r/pkg.TestA/sub.TestB": {Pkg: "github.com/x/r/pkg", Test: "TestA", Sub: "sub.TestB"},
	} {
		got, ok := ParseName(name)
		if !ok || got != want || got.String() != name {
			t.Errorf("%s: %+v %v", name, got, ok)
		}
	}
	n, _ := ParseName("github.com/x/r/pkg.TestA/plain/x.sh")
	if n.Top() != "github.com/x/r/pkg.TestA" || n.Run() != "TestA/plain/x.sh" {
		t.Errorf("top %q run %q", n.Top(), n.Run())
	}
	if _, _, ok := SplitTest("github.com/x/r/pkg.TestA/sub"); ok {
		t.Error("SplitTest accepted a subtest; its callers need a whole test")
	}
	if err := ValidTest("github.com/x/r/pkg.TestA/sub"); err != nil {
		t.Errorf("a subtest name refused: %v", err)
	}
}

func TestCoversAndLeaves(t *testing.T) {
	const top = "p.TestX"
	for _, c := range []struct {
		r, f string
		want bool
	}{
		{top, top, true},
		{top, top + "/plain/a.sh", true},
		{top + "/plain/a.sh", top + "/plain/a.sh", true},
		{top + "/plain", top + "/plain/a.sh", true},
		{top + "/plain/a.sh", top + "/plain/b.sh", false}, // a sibling
		{top + "/plain/a.sh", top + "/plain", false},      // its parent
		{top + "/plain/a.sh", top, false},
		{top + "/pl", top + "/plain/a.sh", false}, // a prefix that is not a segment
	} {
		if got := Covers(c.r, c.f); got != c.want {
			t.Errorf("Covers(%q, %q) = %v", c.r, c.f, got)
		}
	}
	failing := []string{top + "/plain", top + "/plain/b.sh", top + "/plain/a.sh", "p.TestY/z"}
	if got := Leaves(top, failing); strings.Join(got, ",") != top+"/plain/a.sh,"+top+"/plain/b.sh" {
		t.Errorf("leaves %v", got)
	}
	if got := Leaves("p.TestZ", failing); len(got) != 1 || got[0] != "p.TestZ" {
		t.Errorf("a test with no listed subtests is its own leaf: %v", got)
	}
}

// #466: a registered subtest explains its own failure and never a sibling's;
// the rerun is its whole top-level test and must print the subtest's PASS.
func TestClassifySubtests(t *testing.T) {
	const top = "github.com/x/r/scripts/test.TestConstructScripts"
	const dev, sib = top + "/plain/dev-aliases.test.sh", top + "/plain/other.test.sh"
	r, err := BuildRegistry([]Card{{ID: "000465", Status: "working", Tests: []string{dev}}}, open)
	if err != nil {
		t.Fatal(err)
	}
	attempt := func(subs ...string) Result {
		return Result{Failures: []string{top}, Subtests: append([]string{top + "/plain"}, subs...)}
	}
	o := Classify(Input{Attempt: attempt(dev), Registry: r})
	if o.Kind != NeedRerun || len(o.Known) != 1 || o.Known[0].Test != dev || o.Known[0].Issue != "000465" {
		t.Fatalf("registered subtest: %+v", o)
	}
	if got := RerunTests(o.Known); len(got) != 1 || got[0] != top {
		t.Fatalf("rerun tests %v: a subtest reruns its whole top-level test", got)
	}
	o = Classify(Input{Attempt: attempt(sib), Registry: r})
	if o.Kind != Unknown || len(o.Unknown) != 1 || o.Unknown[0] != sib || !strings.Contains(o.Reason, "other.test.sh") {
		t.Fatalf("sibling: %+v", o)
	}
	if o := Classify(Input{Attempt: attempt(dev, sib), Registry: r}); o.Kind != Unknown {
		t.Fatalf("registered plus sibling: %+v", o)
	}
	// A top-level failure whose log names no subtest is its own leaf: not
	// covered by a registered subtest.
	if o := Classify(Input{Attempt: Result{Failures: []string{top}}, Registry: r}); o.Kind != Unknown {
		t.Fatalf("parent alone: %+v", o)
	}
	// The rerun: passing with the exact PASS line lands; a sibling failing in
	// the rerun is named as unknown, not "failed twice".
	in := Input{Attempt: attempt(dev), Registry: r}
	in.Rerun = &Result{Passed: true}
	if o := Classify(in); o.Kind != KnownFlakePassed {
		t.Fatalf("rerun pass: %+v", o)
	}
	in.Rerun = &Result{Failures: []string{top}, Subtests: []string{top + "/plain", sib}}
	if o := Classify(in); o.Kind != Unknown || !strings.Contains(o.Reason, "rerun failed elsewhere") {
		t.Fatalf("rerun sibling failure: %+v", o)
	}
	in.Rerun = &Result{Failures: []string{top}, Subtests: []string{top + "/plain", dev}}
	if o := Classify(in); o.Kind != KnownFlakeTwice {
		t.Fatalf("rerun same failure: %+v", o)
	}
	log := "--- PASS: TestConstructScripts/plain/dev-aliases.test.sh (0.10s)\n--- PASS: TestConstructScripts (1.00s)\n"
	if m := RerunMissing(log, []string{dev}); len(m) != 0 {
		t.Fatalf("missing %v", m)
	}
	if m := RerunMissing("--- PASS: TestConstructScripts (1.00s)\n", []string{dev}); len(m) != 1 {
		t.Fatal("a rerun without the subtest's PASS line proved it")
	}
	// A registered top-level test still covers every subtest.
	whole, _ := BuildRegistry([]Card{{ID: "1", Status: "working", Tests: []string{top}}}, open)
	if o := Classify(Input{Attempt: attempt(sib), Registry: whole}); o.Kind != NeedRerun || o.Known[0].Test != top {
		t.Fatalf("whole test: %+v", o)
	}
	// Overlapping names on two issues are ambiguous.
	if _, err := BuildRegistry([]Card{{ID: "1", Status: "working", Tests: []string{top}}, {ID: "2", Status: "working", Tests: []string{dev}}}, open); err == nil {
		t.Fatal("a test and its subtest on two issues registered")
	}
}

// BR-1 (#466 close review): the success guard's Unexplained judges coverage
// the gate's way. A failure whose failing leaf is a registered subtest is
// explained; a sibling is not; without the log's subtests, a registered
// subtest never covers its top-level test (fail closed).
func TestUnexplainedUsesLeafCoverage(t *testing.T) {
	const top = "github.com/x/r/scripts/test.TestConstructScripts"
	const dev, sib = top + "/plain/dev.sh", top + "/plain/sib.sh"
	r, _ := BuildRegistry([]Card{{ID: "000465", Status: "working", Tests: []string{dev}}}, open)
	at := func(id string, subs ...string) Attempt {
		return Attempt{ID: id, SHA: "abc", Failures: []string{top}, Subtests: subs, Kind: Unknown}
	}
	for name, c := range map[string]struct {
		a    Attempt
		want int
	}{
		"registered subtest":   {at("a", top+"/plain", dev), 0},
		"a sibling":            {at("b", top+"/plain", sib), 1},
		"no log subtests":      {at("c"), 1},
		"registered + sibling": {at("d", dev, sib), 1},
	} {
		if got := Unexplained([]Attempt{c.a}, "abc", r); len(got) != c.want {
			t.Errorf("%s: %d unexplained, want %d", name, len(got), c.want)
		}
	}
	if !Explained([]string{top}, []string{dev}, r) || Explained(nil, nil, r) {
		t.Error("Explained")
	}
}
