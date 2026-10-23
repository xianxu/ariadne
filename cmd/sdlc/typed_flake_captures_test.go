package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flakes"
)

func TestTypedHistoricalCapturesRemainDiagnostic(t *testing.T) {
	for _, tc := range []struct{ file, test string }{
		{"pair-773-timeout.log", "script:tests/nvim-route-poke-admission-test.sh"},
		{"parley-410-o-retry.log", "lua:tests/integration/session_sync_native_spec.lua#session-sync native edits retries refused admission with native o entry (#408 retry)"},
	} {
		t.Run(tc.file, func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join("testdata", "typed-flakes", tc.file))
			if err != nil {
				t.Fatal(err)
			}
			facts := parseSuiteLogFacts(string(raw))
			if facts.Typed.Seen || len(facts.Failures) != 0 || len(flakes.RerunMissingProof(string(raw), []string{tc.test})) != 1 {
				t.Fatalf("historical diagnostics manufactured typed proof: %+v", facts)
			}
		})
	}
}
