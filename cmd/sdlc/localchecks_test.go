package main

import (
	"encoding/json"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/flakes"
	"strings"
	"testing"
)

func TestParseLocalChecks(t *testing.T) {
	good := `{"version":1,"checks":[{"name":"zellij","command":["make","-f","Makefile.local","test-couch-zellij-live"],"substitutes":["couch-zellij-conformance / conformance"]},{"name":"e2e","command":["make","e2e"],"required":true,"timeout_minutes":90}]}`
	d, err := parseLocalChecks([]byte(good))
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Checks) != 2 || d.Checks[0].TimeoutMinutes != 30 || d.Checks[1].TimeoutMinutes != 90 || !d.Checks[1].Required {
		t.Fatalf("%+v", d)
	}
	if c, ok := d.substituteFor("couch-zellij-conformance / conformance"); !ok || c.Name != "zellij" {
		t.Fatalf("substituteFor: %+v %v", c, ok)
	}
	if _, ok := d.substituteFor("other"); ok {
		t.Fatal("substituteFor matched an undeclared label")
	}
	for _, empty := range []string{"", "  \n"} {
		if d, err := parseLocalChecks([]byte(empty)); err != nil || len(d.Checks) != 0 {
			t.Fatalf("empty %q: %+v %v", empty, d, err)
		}
	}
	for _, tc := range []struct{ name, in, want string }{
		{"unknown field", `{"version":1,"checks":[{"name":"a","command":["x"],"substitute":["y"]}]}`, "substitute"},
		{"version", `{"version":3,"checks":[]}`, "version"},
		{"no version", `{"checks":[]}`, "version"},
		{"bad name space", `{"version":1,"checks":[{"name":"A B","command":["x"]}]}`, "name"},
		{"bad name path", `{"version":1,"checks":[{"name":"../x","command":["x"]}]}`, "name"},
		{"duplicate name", `{"version":1,"checks":[{"name":"a","command":["x"]},{"name":"a","command":["y"]}]}`, "twice"},
		{"empty command", `{"version":1,"checks":[{"name":"a","command":[]}]}`, "command"},
		{"empty argv element", `{"version":1,"checks":[{"name":"a","command":["make",""]}]}`, "command"},
		{"empty label", `{"version":1,"checks":[{"name":"a","command":["x"],"substitutes":[""]}]}`, "label"},
		{"label twice", `{"version":1,"checks":[{"name":"a","command":["x"],"substitutes":["j"]},{"name":"b","command":["x"],"substitutes":["j"]}]}`, "j"},
		{"timeout zero", `{"version":1,"checks":[{"name":"a","command":["x"],"timeout_minutes":-1}]}`, "timeout"},
		{"timeout high", `{"version":1,"checks":[{"name":"a","command":["x"],"timeout_minutes":121}]}`, "timeout"},
		{"trailing data", `{"version":1,"checks":[]} {}`, "trailing"},
		{"reserved rerun name", `{"version":1,"checks":[{"name":"suite-rerun","command":["x"]}]}`, "reserved"},
		{"reserved affected name", `{"version":1,"checks":[{"name":"affected","command":["x"]}]}`, "declared in .sdlc/affected.json"},
		{"bad known_flakes", `{"version":1,"checks":[{"name":"suite","command":["x"],"known_flakes":"sometimes"}]}`, "known_flakes"},
		{"known_flakes off the suite", `{"version":1,"checks":[{"name":"e2e","command":["x"],"known_flakes":"accept"}]}`, "the \"suite\" check only"},
		{"not json", `nope`, "local-checks"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parseLocalChecks([]byte(tc.in))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err %v, want it to mention %q", err, tc.want)
			}
		})
	}
	if d, err := parseLocalChecks([]byte(`{"version":1,"checks":[{"name":"suite","command":["x"],"known_flakes":"accept"}]}`)); err != nil || d.Checks[0].knownFlakePolicy() != flakes.PolicyAccept {
		t.Fatalf("a suite's known_flakes (#394): %+v %v", d, err)
	}
	if d, _ := parseLocalChecks([]byte(`{"version":1,"checks":[{"name":"suite","command":["x"]}]}`)); d.Checks[0].knownFlakePolicy() != flakes.PolicyRerun {
		t.Fatalf("the default policy is rerun: %+v", d)
	}
	if got := localStatusContext("zellij"); got != "sdlc-local/zellij" {
		t.Fatal(got)
	}
}

func TestJobRunnerless(t *testing.T) {
	decode := func(s string) jobRun {
		t.Helper()
		var j jobRun
		if err := json.Unmarshal([]byte(s), &j); err != nil {
			t.Fatal(err)
		}
		return j
	}
	for _, tc := range []struct {
		in   string
		want bool
	}{
		// Measured: pair job 115490631095 (attempt 1 of run 38453641991).
		{`{"status":"completed","conclusion":"cancelled","runner_name":"","steps":[]}`, true},
		{`{"status":"completed","conclusion":"cancelled","runner_name":null,"steps":[]}`, true},
		{`{"status":"completed","conclusion":"cancelled","runner_name":"GitHub Actions 1","steps":[]}`, false},
		{`{"status":"completed","conclusion":"cancelled","runner_name":"","steps":[{},{},{}]}`, false},
		{`{"status":"completed","conclusion":"failure","runner_name":"","steps":[]}`, false},
		{`{"status":"queued","conclusion":null,"runner_name":"","steps":[]}`, false},
		// Unread is not runnerless: a body missing the fields never substitutes.
		{`{"status":"completed","conclusion":"cancelled"}`, false},
		{`{"status":"completed","conclusion":"cancelled","runner_name":""}`, false},
	} {
		if got := decode(tc.in).runnerless(); got != tc.want {
			t.Errorf("%s: runnerless %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestJobIDFromLink(t *testing.T) {
	if id, ok := jobIDFromLink("https://github.com/xianxu/pair/actions/runs/38453641991/job/115490631095"); !ok || id != 115490631095 {
		t.Fatalf("%d %v", id, ok)
	}
	for _, l := range []string{"", "https://github.com/o/r/pull/3/checks", "https://github.com/o/r/actions/runs/1/job/abc", "https://example.com/actions/runs/1/job/2", "https://github.com/o/r/actions/runs/1/job/2/extra"} {
		if _, ok := jobIDFromLink(l); ok {
			t.Errorf("%q parsed as a job link", l)
		}
	}
}
