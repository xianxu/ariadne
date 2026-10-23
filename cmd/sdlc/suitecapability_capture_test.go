package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestSuiteCapabilityCapturedFailuresRemainFailures(t *testing.T) {
	raw, err := os.ReadFile("testdata/suite-sandbox-mixed.log")
	if err != nil {
		t.Fatal(err)
	}
	// Original #561 bytes, including the non-denial invalid-attempt failure.
	if fmt.Sprintf("%x", sha256.Sum256(raw)) != "b59b0ca759609006000b60cfa623460af4aa9821e76c550e1237fec0806ae646" {
		t.Fatal("original capture changed")
	}
	// This projection is deliberately synthetic, not claimed as the original
	// run. Even entirely denial-looking output cannot prove the host was denied.
	allDenial := "--- FAIL example.org/p TestPermission\n=== RUN   TestPermission\n    permission_test.go:3: operation not permitted\n--- FAIL: TestPermission (0.01s)\n--- FAIL example.org/p\nFAIL\n"
	for _, output := range []string{string(raw), allDenial} {
		d, sha, gh, stderr := localRunRepo(t, "skipped")
		d.suiteCapabilities = func(context.Context) error { return nil }
		rec, err := runLocalCheck(context.Background(), d, infrastructureCheck(t, output), sha)
		if err != nil || rec.passed() {
			t.Fatalf("rec=%+v err=%v\n%s", rec, err, stderr)
		}
		if len(gateAttempts(t, d)) != 1 || gh.statuses[sha]["sdlc-local/suite"] != "failure" {
			t.Fatalf("log wording erased real failure: %s", stderr)
		}
		if strings.HasSuffix(rec.LogPath, ".environment.log") {
			t.Fatal("log text granted environment authority")
		}
		before := capabilityEvidenceFiles(t, d)
		d.suiteCapabilities = func(context.Context) error { return suiteCapabilityDenied() }
		_, err = runLocalCheck(context.Background(), d, suiteCheck("true"), sha)
		assertCapabilityRefusal(t, err, suiteCheckName, sha)
		if len(gateAttempts(t, d)) != 1 || gh.statuses[sha]["sdlc-local/suite"] != "failure" || len(before) != len(capabilityEvidenceFiles(t, d)) {
			t.Fatal("later environment refusal erased the genuine failure")
		}
	}
}
