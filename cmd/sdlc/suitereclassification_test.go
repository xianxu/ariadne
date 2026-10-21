package main

import (
	"strings"
	"testing"
)

func TestSuiteReclassificationEligibility(t *testing.T) {
	log := []byte(recordedCacheFailure(t))
	original := storedAttemptFixture("a1")
	original.Slot, original.Failures = "ariadne:5", nil
	original.LogSHA256 = hashBytes(log)
	original.LogPath = "failed/a1.log"
	for _, tc := range []struct {
		name, actor, repo, reason string
		mutate                    func(*SuiteAttempt)
		output                    []byte
		ok                        bool
	}{
		{name: "owner", actor: "ariadne:5", ok: true},
		{name: "steward", actor: "ariadne:0", ok: true},
		{name: "foreign slot", actor: "ariadne:4"},
		{name: "foreign repo", actor: "pair:0"},
		{name: "unknown actor", actor: ""},
		{name: "unknown owner steward", actor: "ariadne:0", mutate: func(a *SuiteAttempt) { a.Slot = "" }, ok: true},
		{name: "unknown owner worker", actor: "ariadne:5", mutate: func(a *SuiteAttempt) { a.Slot = "" }},
		{name: "other repository", actor: "ariadne:5", repo: "other/repo"},
		{name: "blank reason", actor: "ariadne:5", reason: " "},
		{name: "oversized reason", actor: "ariadne:5", reason: strings.Repeat("r", 4097)},
		{name: "pass", actor: "ariadne:5", mutate: func(a *SuiteAttempt) { a.Exit = 0 }},
		{name: "unknown exit", actor: "ariadne:5", mutate: func(a *SuiteAttempt) { a.Exit = -1 }},
		{name: "timeout", actor: "ariadne:5", mutate: func(a *SuiteAttempt) { a.TimedOut = true }},
		{name: "other check", actor: "ariadne:5", mutate: func(a *SuiteAttempt) { a.Check = "lint" }},
		{name: "baseline", actor: "ariadne:5", mutate: func(a *SuiteAttempt) { a.MainBaseline = true }, ok: true},
		{name: "affected", actor: "ariadne:5", mutate: func(a *SuiteAttempt) { a.Check = affectedCheckName }, ok: true},
		{name: "hash mismatch", actor: "ariadne:5", mutate: func(a *SuiteAttempt) { a.LogSHA256 = strings.Repeat("0", 64) }},
		{name: "mutable log", actor: "ariadne:5", mutate: func(a *SuiteAttempt) { a.LogPath = "local-checks/" + a.SHA + "/suite.log" }},
		{name: "empty historical log", actor: "ariadne:5", output: []byte{}},
		{name: "test started", actor: "ariadne:5", output: []byte("=== RUN   TestActual\n" + string(log))},
		{name: "truncated", actor: "ariadne:5", output: log[:len(log)-1]},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := original
			if tc.mutate != nil {
				tc.mutate(&a)
			}
			output := log
			if tc.output != nil {
				output = tc.output
				a.LogSHA256 = hashBytes(output)
			}
			repo, reason := original.Repo, "storage repaired"
			if tc.repo != "" {
				repo = tc.repo
			}
			if tc.reason != "" {
				reason = tc.reason
			}
			diagnosis, err := planSuiteReclassification(a, repo, tc.actor, reason, output)
			if (err == nil) != tc.ok {
				t.Fatalf("eligible=%v diagnosis=%+v err=%v", tc.ok, diagnosis, err)
			}
			if err == nil && (diagnosis.Resource == "" || diagnosis.Path == "") {
				t.Fatalf("incomplete diagnosis: %+v", diagnosis)
			}
		})
	}
}
