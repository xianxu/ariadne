package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func policyProjection(t *testing.T, marker any, granted bool) suiteQueueObservation {
	t.Helper()
	record := map[string]any{"repo": "legacy", "worktree": "/private/legacy"}
	if marker != nil {
		record["ordering_policy"] = marker
	}
	data, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	return suiteQueueObservation{Version: 1, Available: true, PolicySeconds: 3600, Tickets: []suiteQueueTicket{
		{Number: 32, Position: 1, Reason: "unknown_age_fifo", Record: data},
		{Number: 37, Position: 2, Reason: "grant_priority", Granted: granted, Record: json.RawMessage(`{"repo":"granted","worktree":"/private/granted","ordering_policy":"fifo-grant-aging-v1","subject":{"common_dir":"/private/pair.git","issue":832}}`)},
	}}
}

func TestSuiteQueueCompatibilityCountedGrowthAndTrainHints(t *testing.T) {
	for _, n := range []int{16, 32} {
		visits := 0
		result := assessSuiteQueueCompatibility(func(yield func(suiteQueueTicket) bool) {
			for i := 0; i < n; i++ {
				visits++
				if !yield(suiteQueueTicket{Number: int64(i + 1), Granted: true, Record: json.RawMessage(`{"repo":"train","worktree":"/private/train","subject":{"common_dir":"/private/pair's repo.git","issue":0,"issues":[832,833,832]}}`)}) {
					return
				}
			}
		})
		if visits != n || len(result.Incompatible) != n || len(result.Revoke) != 2 || !result.Granted {
			t.Fatalf("nonlinear/incorrect projection: visits=%d result=%+v", visits, result)
		}
		if !strings.Contains(result.Revoke[1], "--issue 833") || !strings.Contains(result.Revoke[1], quoteHelperShell("/private/pair's repo.git")) {
			t.Fatalf("train/literal path recovery lost: %+v", result)
		}
	}
}

func TestLandingGrantCompatibilityLifecycle(t *testing.T) {
	repo := landingRepo(t)
	d := testLandingDeps(slotPtr(0), slotPtr(0))
	for _, tc := range []struct {
		name       string
		projection suiteQueueObservation
		allowed    bool
	}{
		{"empty", suiteQueueObservation{Available: true}, true},
		{"compatible", policyProjection(t, "fifo-grant-aging-v1", false), true},
		{"legacy", policyProjection(t, nil, false), false},
		{"future", policyProjection(t, "future-policy", false), false},
		{"unavailable", suiteQueueObservation{Reason: "allocation_busy"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			d.suite = func(ctx context.Context, root string) suiteQueueObservation {
				calls++
				if root != repo {
					t.Fatal("inspected different repository")
				}
				if deadline, ok := ctx.Deadline(); !ok || time.Until(deadline) > 3*time.Second {
					t.Fatal("missing bounded observation")
				}
				return tc.projection
			}
			before := repoGrants(t, repo, d.now())
			_, err := runLanding(t, d, "grant", "--repo", repo, "--issue", "832", "--reason", "private fixture")
			if (err == nil) != tc.allowed || calls != 1 {
				t.Fatalf("allowed=%v calls=%d err=%v", tc.allowed, calls, err)
			}
			if !tc.allowed && len(repoGrants(t, repo, d.now())) != len(before) {
				t.Fatal("refusal changed grants")
			}
			d.suite = func(context.Context, string) suiteQueueObservation {
				t.Fatal("revoke inspected")
				return suiteQueueObservation{}
			}
			if _, err := runLanding(t, d, "revoke", "--repo", repo, "--issue", "832"); err != nil {
				t.Fatal(err)
			}
		})
	}
	// Cancelling a successful observation must still prevent the grant append.
	ctx, cancel := context.WithCancel(context.Background())
	d.suite = func(context.Context, string) suiteQueueObservation {
		cancel()
		return suiteQueueObservation{Available: true}
	}
	cmd := newLandingCmd(d)
	cmd.SetArgs([]string{"grant", "--repo", repo, "--issue", "833", "--reason", "private fixture"})
	if err := cmd.ExecuteContext(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("lost cancellation: %v", err)
	}
	if len(repoGrants(t, repo, d.now())) != 0 {
		t.Fatal("cancelled grant was written")
	}
}

func TestSuiteQueueDecodeRejectsDuplicatePolicyAuthority(t *testing.T) {
	raw := strings.Replace(suiteQueueFixture, `"repo":"first"`, `"repo":"first","ordering_policy":"old","ordering_policy":"fifo-grant-aging-v1"`, 1)
	if _, err := decodeSuiteQueue([]byte(raw)); err == nil {
		t.Fatal("last policy value manufactured compatibility")
	}
}

func TestSuiteQueueMixedPolicyRisk(t *testing.T) {
	for _, marker := range []any{nil, "old", true, 1, map[string]any{"version": 1}} {
		t.Run(fmt.Sprint(marker), func(t *testing.T) {
			var out bytes.Buffer
			renderSuiteQueue(&out, policyProjection(t, marker, true))
			for _, want := range []string{"legacy/unknown ordering", "ticket 32", "/private/legacy", "deadlock risk", "sdlc landing revoke --issue 832", "--repo"} {
				if !strings.Contains(out.String(), want) {
					t.Errorf("missing %q: %s", want, out.String())
				}
			}
		})
	}
	for _, tc := range []struct {
		marker any
		grant  bool
	}{{nil, false}, {"fifo-grant-aging-v1", true}} {
		var out bytes.Buffer
		renderSuiteQueue(&out, policyProjection(t, tc.marker, tc.grant))
		if strings.Contains(out.String(), "deadlock risk") {
			t.Fatalf("invented risk: %s", out.String())
		}
	}
}

func TestLandingGrantRefusesLegacySupplierSnapshot(t *testing.T) {
	raw, err := json.Marshal(policyProjection(t, nil, false))
	if err != nil {
		t.Fatal(err)
	}
	repo := suiteQueueHelperFixture(t, "print("+quoteHelperShell(string(raw))+")\n")
	git(t, "", "init", "-q", repo)
	d := testLandingDeps(slotPtr(0), slotPtr(0))
	d.suite = observeSuiteQueue
	_, err = runLanding(t, d, "grant", "--repo", repo, "--issue", "832", "--reason", "private fixture")
	if err == nil || !strings.Contains(err.Error(), "ticket 32") || !strings.Contains(err.Error(), "/private/legacy") {
		t.Fatalf("legacy grant not refused with identity: %v", err)
	}
	if grants := repoGrants(t, repo, d.now()); len(grants) != 0 {
		t.Fatalf("refused grant written: %+v", grants)
	}
	// A recovery revoke must not invoke even an unavailable supplying inspector.
	helper, err := filepath.EvalSymlinks(filepath.Join(repo, "Makefile.workflow"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(filepath.Dir(helper), "scripts/test-lock.py"), []byte("raise Exception('revoke inspected')\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := runLanding(t, d, "revoke", "--repo", repo, "--issue", "832"); err != nil {
		t.Fatal(err)
	}
}
