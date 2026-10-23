package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flakes"
)

func typedPassingLog(t *testing.T, unit, label string) string {
	t.Helper()
	var log strings.Builder
	zero := 0
	events := []flakes.UnitEvent{{Version: 1, Invocation: strings.Repeat("b", 32), Unit: unit, Event: "start"}}
	if label != "" {
		events = append(events, flakes.UnitEvent{Version: 1, Invocation: strings.Repeat("b", 32), Unit: unit, Event: "case", Case: label, Result: "pass"})
	}
	events = append(events, flakes.UnitEvent{Version: 1, Invocation: strings.Repeat("b", 32), Unit: unit, Event: "finish", Exit: &zero, Complete: true})
	for _, e := range events {
		b, err := json.Marshal(e)
		if err != nil {
			t.Fatal(err)
		}
		log.Write(b)
		log.WriteByte('\n')
	}
	return log.String()
}

func typedResumeFixture(t *testing.T) ResumeProofInput {
	t.Helper()
	in := resumeCorePassed(t)
	name := "lua:tests/native_spec.lua#native retry"
	in.Parent.Version = 4
	in.Parent.RunnerContractSHA256 = strings.Repeat("f", 64)
	in.Parent.Failures = []string{name}
	in.ParentFacts = parseSuiteLogFacts(typedLogFrames(t, "lua:tests/native_spec.lua", "native retry", 1))
	in.RunnerContractSHA256 = in.Parent.RunnerContractSHA256
	in.Child.Version = 4
	in.Child.RunnerContractSHA256 = in.Parent.RunnerContractSHA256
	in.ChildLog = typedPassingLog(t, "lua:tests/native_spec.lua", "native retry")
	in.ChildLogSHA256 = resumeRecordDigest([]byte(in.ChildLog))
	in.Child.LogSHA256 = in.ChildLogSHA256
	var err error
	in.Registry, err = flakes.BuildRegistry([]flakes.Card{{ID: "556", Tests: []string{name}}}, func(string) bool { return true })
	if err != nil {
		t.Fatal(err)
	}
	return in
}

func TestTypedResumeProofRequiresExactContractAndCase(t *testing.T) {
	in := typedResumeFixture(t)
	if p := resolveResumeProof(in); p.State != ResumePassed {
		t.Fatalf("valid typed resume: %+v", p)
	}
	for name, mutate := range map[string]func(*ResumeProofInput){
		"current contract absent":  func(i *ResumeProofInput) { i.RunnerContractSHA256 = "" },
		"current contract changed": func(i *ResumeProofInput) { i.RunnerContractSHA256 = strings.Repeat("e", 64) },
		"child contract changed":   func(i *ResumeProofInput) { i.Child.RunnerContractSHA256 = strings.Repeat("e", 64) },
		"legacy child":             func(i *ResumeProofInput) { i.Child.Version = 2; i.Child.RunnerContractSHA256 = "" },
		"other case": func(i *ResumeProofInput) {
			i.ChildLog = typedPassingLog(t, "lua:tests/native_spec.lua", "other case")
			i.ChildLogSHA256 = resumeRecordDigest([]byte(i.ChildLog))
			i.Child.LogSHA256 = i.ChildLogSHA256
		},
		"missing finish": func(i *ResumeProofInput) {
			i.ChildLog = strings.Split(i.ChildLog, "\n")[0] + "\n"
			i.ChildLogSHA256 = resumeRecordDigest([]byte(i.ChildLog))
			i.Child.LogSHA256 = i.ChildLogSHA256
		},
	} {
		t.Run(name, func(t *testing.T) {
			i := typedResumeFixture(t)
			mutate(&i)
			if p := resolveResumeProof(i); p.State == ResumePassed || p.Outcome.Kind.Lands() {
				t.Fatalf("unsafe: %+v", p)
			}
		})
	}
}

func TestTypedRejectedOutcomeCannotPassAtZeroExit(t *testing.T) {
	for _, kind := range []flakes.Kind{flakes.Unattributable, flakes.RerunInconclusive, flakes.KnownFlakeTwice} {
		if (localRunRecord{Exit: 0, Outcome: kind}).passed() {
			t.Fatalf("rejected %s passed", kind)
		}
	}
}
