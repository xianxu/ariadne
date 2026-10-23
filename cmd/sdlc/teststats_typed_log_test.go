package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flakes"
)

func typedLogFrames(t *testing.T, unit, label string, exit int) string {
	t.Helper()
	events := []flakes.UnitEvent{{Version: 1, Invocation: strings.Repeat("a", 32), Unit: unit, Event: "start"}}
	if label != "" {
		events = append(events, flakes.UnitEvent{Version: 1, Invocation: strings.Repeat("a", 32), Unit: unit, Event: "case", Case: label, Result: "fail"})
	}
	events = append(events, flakes.UnitEvent{Version: 1, Invocation: strings.Repeat("a", 32), Unit: unit, Event: "finish", Exit: &exit, Complete: true})
	var log strings.Builder
	for _, event := range events {
		b, err := json.Marshal(event)
		if err != nil {
			t.Fatal(err)
		}
		log.Write(b)
		log.WriteByte('\n')
	}
	return log.String()
}

func TestTypedSuiteLogFacts(t *testing.T) {
	for _, tc := range []struct {
		name, unit, label, want string
		exit                    int
	}{
		{"script timeout", "script:tests/route.sh", "", "script:tests/route.sh", 124},
		{"Lua failure", "lua:tests/native_spec.lua", "native retry (#408)", "lua:tests/native_spec.lua#native retry (#408)", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			facts := parseSuiteLogFacts(typedLogFrames(t, tc.unit, tc.label, tc.exit) + "    --- FAIL: TestDiagnostic/child (0.00s)\n")
			if !facts.Failed || len(facts.Failures) != 1 || facts.Failures[0] != tc.want || len(facts.Unattributed) != 0 || len(facts.Warnings) != 0 {
				t.Fatalf("facts=%+v", facts)
			}
		})
	}
}

func TestTypedSuiteLogDoesNotBorrowPackageProof(t *testing.T) {
	facts := parseSuiteLogFacts(typedLogFrames(t, "script:tests/route.sh", "", 124) + "FAIL example.org/pkg 0.1s\n")
	if len(facts.Unattributed) != 1 || !strings.Contains(facts.Unattributed[0], "package example.org/pkg") {
		t.Fatalf("facts=%+v", facts)
	}
}

func TestTypedSuiteLogMalformedFrameFailsClosed(t *testing.T) {
	facts := parseSuiteLogFacts("{\"sdlc_unit\":1,\"event\":\"finish\"}\n")
	if !facts.Failed || len(facts.Unattributed) == 0 {
		t.Fatalf("facts=%+v", facts)
	}
}

func TestTypedSuiteLogDiagnosticFramesIgnored(t *testing.T) {
	frames := typedLogFrames(t, "script:tests/route.sh", "", 124)
	var log strings.Builder
	for _, line := range strings.Split(strings.TrimSpace(frames), "\n") {
		log.WriteString("    " + line + "\n")
	}
	nested, _ := json.Marshal(map[string]any{"Action": "output", "Package": "example.org/pkg", "Output": frames})
	log.Write(nested)
	facts := parseSuiteLogFacts(log.String())
	if facts.Failed || len(facts.Failures) != 0 || len(facts.Unattributed) != 0 {
		t.Fatalf("facts=%+v", facts)
	}
}
