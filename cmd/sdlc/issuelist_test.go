package main

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestIssueListFilter(t *testing.T) {
	web := IssueState{ID: "1", Status: "open", Area: []string{"couch/broadcast/web"}, Severity: "high"}
	caster := IssueState{ID: "2", Status: "working", Area: []string{"couch/broadcaster"}, Severity: "low"}
	multi := IssueState{ID: "3", Status: "open", Area: []string{"sdlc/merge", "couch/session"}}
	none := IssueState{ID: "4", Status: "open", Kind: "bug"}
	all := []IssueState{web, caster, multi, none}
	ids := func(f issueListFilter) string {
		var out []string
		for _, is := range all {
			if f.matches(is) {
				out = append(out, is.ID)
			}
		}
		return strings.Join(out, ",")
	}
	for _, tc := range []struct {
		name              string
		status            string
		areas, severities []string
		want              string
	}{
		{"none", "", nil, nil, "1,2,3,4"},
		{"segment prefix", "", []string{"couch/broadcast"}, nil, "1"},
		{"exact", "", []string{"couch/broadcast/web"}, nil, "1"},
		{"top level", "", []string{"couch"}, nil, "1,2,3"},
		{"areas OR", "", []string{"sdlc", "couch/broadcaster"}, nil, "2,3"},
		{"severity list", "", nil, []string{"high", "low"}, "1,2"},
		{"AND across flags", "open", []string{"couch"}, []string{"high"}, "1"},
		{"status", "working", nil, nil, "2"},
	} {
		f, err := newIssueListFilter(tc.status, tc.areas, tc.severities, nil)
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		if got := ids(f); got != tc.want {
			t.Errorf("%s: %s, want %s", tc.name, got, tc.want)
		}
	}
	// #633: --kind filters by the triage mark, and refuses a value outside it.
	if f, err := newIssueListFilter("", nil, nil, []string{"bug"}); err != nil || ids(f) != "4" {
		t.Errorf("--kind bug: %v %s", err, ids(f))
	}
	for _, kinds := range [][]string{{"chore"}, {""}} {
		if _, err := newIssueListFilter("", nil, nil, kinds); err == nil || !strings.Contains(err.Error(), "--kind") {
			t.Errorf("--kind %q: %v", kinds, err)
		}
	}
	for _, bad := range []struct {
		status            string
		areas, severities []string
		want              string
	}{
		{"waiting", nil, nil, "invalid status"},
		{"", []string{"Bad/Path"}, nil, "--area"},
		{"", nil, []string{"urgent"}, "critical, high, medium, low"},
	} {
		if _, err := newIssueListFilter(bad.status, bad.areas, bad.severities, nil); err == nil || !strings.Contains(err.Error(), bad.want) {
			t.Errorf("%+v: %v", bad, err)
		}
	}
}

func TestIssueListByAreaAndSeverity(t *testing.T) {
	p1, c1, d1p, d1 := seededIssue(t, "000001", "web")
	p2, c2, d2p, d2 := seededIssue(t, "000002", "plain")
	c1 = strings.Replace(c1, "status: open", "status: open\narea: [couch/broadcast/web]\nseverity: high", 1)
	d1 = strings.Replace(d1, "status: open", "status: open\narea: [couch/broadcast/web]\nseverity: high", 1)
	newTrackerRepo(t, map[string]string{p1: c1, p2: c2}, map[string]string{d1p: d1, d2p: d2})

	var out, errs bytes.Buffer
	if err := runIssueList(context.Background(), &out, &errs, &issueListFlags{Areas: []string{"couch/broadcast"}, IssuesDir: "workshop/issues"}); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(out.String()); !strings.HasPrefix(got, "000001") || strings.Contains(got, "000002") || !strings.Contains(got, "high") || !strings.Contains(got, "couch/broadcast/web") {
		t.Fatalf("filtered list:\n%s", got)
	}
	// Unclassified output keeps the three-column shape.
	out.Reset()
	if err := runIssueList(context.Background(), &out, &errs, &issueListFlags{Status: "open", IssuesDir: "workshop/issues"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "000002  open        -         -") {
		t.Fatalf("mixed list should pad the unclassified row:\n%s", out.String())
	}
	// state's JSON carries both fields.
	states, _, err := listIssueStates(context.Background(), "workshop/issues")
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(states)
	if !strings.Contains(string(raw), `"area":["couch/broadcast/web"]`) || !strings.Contains(string(raw), `"severity":"high"`) {
		t.Fatalf("state json: %s", raw)
	}
}
