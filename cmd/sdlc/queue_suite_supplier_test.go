package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSuiteQueueSupplierCurrentRevisionIsNotLaunchProof(t *testing.T) {
	d, sha, _, _ := localRunRepo(t, "acquired")
	raw, _ := json.Marshal(map[string]any{"version": 1, "available": true, "policy_seconds": 3600, "protected_through": 0, "priority_overridden": false, "tickets": []any{
		map[string]any{"number": 1, "position": 1, "reason": "legacy", "record": map[string]any{"repo": "old", "worktree": d.Root}},
	}})
	helper := filepath.Join(d.Root, "scripts/test-lock.py")
	if err := os.WriteFile(helper, []byte("print("+quoteHelperShell(string(raw))+")\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(d.Root)
	var out bytes.Buffer
	renderSuiteQueue(&out, defaultSuiteQueueObservation(context.Background()))
	for _, want := range []string{d.Root, sha, "current supplying checkout", "launch revision unknown", "dirty"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("missing %q: %s", want, out.String())
		}
	}
}

func supplierTickets(n int, same bool) suiteQueueObservation {
	out := suiteQueueObservation{Available: true, PolicySeconds: 3600}
	for i := 0; i < n; i++ {
		owner := i
		if same {
			owner = 0
		}
		raw, _ := json.Marshal(map[string]any{"repo": "old", "worktree": fmt.Sprintf("/private/owner%d", owner)})
		out.Tickets = append(out.Tickets, suiteQueueTicket{Number: int64(i + 1), Position: i + 1, Reason: "legacy", Record: raw})
	}
	return out
}

func TestSuiteQueueSupplierCountedGrowth(t *testing.T) {
	for _, n := range []int{8, 16} {
		for _, mode := range []string{"same-owner", "same-supplier", "distinct"} {
			t.Run(fmt.Sprintf("%s/%d", mode, n), func(t *testing.T) {
				resolves, reads := 0, 0
				io := suiteSupplierIO{
					resolve: func(_ context.Context, owner string) (string, error) {
						resolves++
						if mode == "distinct" {
							return owner + "/supplier", nil
						}
						return "/private/shared", nil
					},
					read: func(_ context.Context, path string) suiteQueueSupplier {
						reads++
						return suiteQueueSupplier{Checkout: path, Revision: strings.Repeat("a", 40)}
					},
				}
				got := annotateSuiteSuppliers(context.Background(), supplierTickets(n, mode == "same-owner"), io)
				wantResolve, wantRead := n, 1
				if mode == "same-owner" {
					wantResolve = 1
				}
				if mode == "distinct" {
					wantRead = n
				}
				if resolves != wantResolve || reads != wantRead || len(got.Suppliers) != n {
					t.Fatalf("nonlinear/incorrect: resolve=%d read=%d rows=%d", resolves, reads, len(got.Suppliers))
				}
			})
		}
	}
}

func TestSuiteQueueSupplierUnavailableAndCompatibleControls(t *testing.T) {
	for _, owner := range []string{"", "relative", "/private/../escape", "/private/a\nb"} {
		obs := supplierTickets(1, true)
		obs.Tickets[0].Record, _ = json.Marshal(map[string]any{"worktree": owner})
		io := suiteSupplierIO{resolve: func(context.Context, string) (string, error) { t.Fatal("invalid owner probed"); return "", nil }}
		got := annotateSuiteSuppliers(context.Background(), obs, io)
		if got.Suppliers[1].Reason == "" {
			t.Fatalf("invalid owner fabricated proof: %+v", got)
		}
	}
	obs := supplierTickets(1, true)
	obs.Tickets[0].Record = json.RawMessage(`{"ordering_policy":"fifo-grant-aging-v1","worktree":"/private/owner"}`)
	got := annotateSuiteSuppliers(context.Background(), obs, suiteSupplierIO{})
	if len(got.Suppliers) != 0 {
		t.Fatal("current policy did supplier work")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	got = annotateSuiteSuppliers(ctx, supplierTickets(16, false), suiteSupplierIO{})
	for _, row := range got.Suppliers {
		if row.Reason != "status deadline exhausted" {
			t.Fatalf("deadline not labelled: %+v", row)
		}
	}
	// Malformed known fields must follow the same legacy classification as the grant gate.
	obs.Tickets[0].Record = json.RawMessage(`{"ordering_policy":"fifo-grant-aging-v1","subject":"malformed","worktree":"/private/owner"}`)
	got = annotateSuiteSuppliers(context.Background(), obs, suiteSupplierIO{resolve: func(context.Context, string) (string, error) { return "", fmt.Errorf("unavailable") }})
	var out bytes.Buffer
	renderSuiteQueue(&out, got)
	if !strings.Contains(out.String(), "revision unavailable") || !strings.Contains(out.String(), "launch revision unknown") {
		t.Fatal(out.String())
	}
}
