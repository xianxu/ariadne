package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/estimate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"go.yaml.in/yaml/v3"
)

// partialMeasurement is a measured actual with 20 unattributed minutes of an
// unknown branch beside its 60 attributed ones.
func partialMeasurement() actualResult {
	inputs := measuredInputsFixture()
	inputs.Unattributed = []gatestate.MeasurementUnattributed{{Kind: gatestate.UnattributedUnknownBranch, Source: "codex/main", Start: "2026-10-13T18:10:00Z", End: "2026-10-13T18:30:00Z", Bounded: true, Reason: "concurrent claims have unknown, detached or ambiguous active branch: #42 # not a comment"}}
	inputs.UnattributedMinutes = 20
	return actualResult{Status: actualMeasured, Hours: 1, Window: "2026-10-13T18:00:00Z → HEAD", Inputs: inputs}
}

func stubCloseMeasurement(t *testing.T, res actualResult) {
	t.Helper()
	orig := computeActualForCloseFn
	t.Cleanup(func() { computeActualForCloseFn = orig })
	computeActualForCloseFn = func(context.Context, string, string, string, ...string) actualResult { return res }
}

// #603: a close in a repository without an issue tracker adopts every
// measurement outcome. The frontmatter it writes stays valid YAML and carries
// the bare confidence token; calibration gets the confidence columns; an
// unmeasured close writes N/A, no row, and no --no-actual acknowledgement.
func TestLegacyCloseRecordsEveryConfidence(t *testing.T) {
	for _, tc := range []struct {
		name, actual, token string
		res                 actualResult
		row                 bool
	}{
		{"exact", "1", "exact", actualResult{Status: actualMeasured, Hours: 1, Inputs: measuredInputsFixture()}, true},
		{"partial", "1", "partial 0.75", partialMeasurement(), true},
		{"unmeasured", issue.ActualNotApplicableSentinel, "unmeasured", actualResult{Status: actualTelemetryGap, Detail: "window has commits: but no events # here"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			issuesDir := closeRepo(t, 73)
			ledger := filepath.Join(t.TempDir(), "l.tsv")
			t.Setenv("WF_CALIB_LEDGER", ledger)
			stubCloseMeasurement(t, tc.res)
			var stderr bytes.Buffer
			f := &closeFlags{Issue: 73, Verified: "done", NoAtlas: true, IssuesDir: issuesDir, BrainDir: "../nonexistent-brain"}
			if err := runClose(io.Discard, &stderr, f); err != nil {
				t.Fatalf("close refused a %s measurement: %v\n%s", tc.name, err, stderr.String())
			}
			raw, err := os.ReadFile(mustIssuePath(t, issuesDir, 73))
			if err != nil {
				t.Fatal(err)
			}
			fm, _, err := issue.Parse(string(raw))
			if err != nil {
				t.Fatal(err)
			}
			var parsed map[string]any
			if err := yaml.Unmarshal([]byte(fm), &parsed); err != nil {
				t.Fatalf("frontmatter is not valid YAML: %v\n%s", err, fm)
			}
			if got := fmt.Sprint(parsed["actual_hours"]); got != tc.actual {
				t.Fatalf("actual_hours = %q, want %s", got, tc.actual)
			}
			if parsed["actual_confidence"] != tc.token {
				t.Fatalf("actual_confidence = %v, want %q", parsed["actual_confidence"], tc.token)
			}
			if strings.Contains(stderr.String(), "--no-actual (or --force): closing with actual_hours") {
				t.Fatalf("a measured close printed the --no-actual acknowledgement:\n%s", stderr.String())
			}
			data, _ := os.ReadFile(ledger)
			rows := estimate.ParseRows(string(data))
			if !tc.row {
				if len(rows) != 0 {
					t.Fatalf("an unmeasured close wrote a calibration row: %q", data)
				}
				return
			}
			if len(rows) != 1 || rows[0].Confidence != string(classifyActual(tc.res).Level) || rows[0].AttributedShare != classifyActual(tc.res).Share() {
				t.Fatalf("calibration row = %+v", rows)
			}
		})
	}
}

// #603: a tracker close carries the confidence in its evidence commit's
// trailer, beside Close-Actual, and an unmeasured close proceeds with N/A.
func TestTrackerCloseCarriesConfidenceTrailer(t *testing.T) {
	for _, tc := range []struct {
		name, actual, trailer string
		res                   actualResult
	}{
		{"partial", "Close-Actual: 1.00", "Close-Actual-Confidence: partial 0.75 (20m unattributed: concurrent claims", partialMeasurement()},
		{"unmeasured", "Close-Actual: N/A", "Close-Actual-Confidence: unmeasured (telemetry gap", actualResult{Status: actualTelemetryGap}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, cardPath, _ := closeReady(t, 302)
			res := tc.res
			if res.Inputs != nil {
				res.Inputs.CodeHead = r.git("rev-parse", "HEAD")
			}
			stubCloseMeasurement(t, res)
			stubJudge(t, "VERDICT: SHIP (confidence: high)\n\nfine\n"+emptyBoundaryFindings)
			_, stderr, err := executeSDLCTestCommand("close", "--issue", "302", "--verified", "e2e", "--no-atlas")
			if err != nil {
				t.Fatalf("close: %v\n%s", err, stderr)
			}
			msg := r.git("log", "-1", "--format=%B", evidenceRev)
			for _, want := range []string{tc.actual, tc.trailer} {
				if !strings.Contains(msg, want) {
					t.Errorf("evidence message lacks %q:\n%s", want, msg)
				}
			}
			if !strings.Contains(r.card(cardPath), "status: codecomplete") {
				t.Fatal("the close did not complete")
			}
		})
	}
}

// #172-instrument guard over #603's renderings: neither the confidence lines
// nor the unmeasured explanation reads as a gate acknowledgement or refusal.
func TestConfidenceLinesNoGatesigCollision(t *testing.T) {
	for _, res := range []actualResult{{Status: actualMeasured, Hours: 1, Inputs: measuredInputsFixture()}, partialMeasurement(), {Status: actualTelemetryGap}, {Status: actualError, Detail: "queue evidence incomplete"}} {
		var out bytes.Buffer
		f := &closeFlags{Issue: 9}
		stubCloseMeasurement(t, res)
		adoptOmittedActual(&out, f, "9", "issue")
		for _, line := range strings.Split(out.String(), "\n") {
			assertNoGatesigCollision(t, line)
		}
	}
}
