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

// nativeSupplierFixture separates the ticket owner from the actual helper
// supplier. Git's real refs/index/worktree topology back every observation.
func nativeSupplierFixture(t *testing.T) (owner, supplier, sha string) {
	t.Helper()
	source, sha, _, _ := localRunRepo(t, "acquired")
	consumer, _, _, _ := localRunRepo(t, "acquired")
	link := filepath.Join(consumer.Root, "Makefile.workflow")
	if err := os.Remove(link); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(source.Root, "Makefile.workflow"), link); err != nil {
		t.Fatal(err)
	}
	return consumer.Root, source.Root, sha
}

func nativeSupplierTickets(t *testing.T, owners ...string) suiteQueueObservation {
	t.Helper()
	obs := suiteQueueObservation{Available: true, PolicySeconds: 3600}
	for i, owner := range owners {
		raw, err := json.Marshal(map[string]any{"repo": "native-legacy", "worktree": owner})
		if err != nil {
			t.Fatal(err)
		}
		obs.Tickets = append(obs.Tickets, suiteQueueTicket{Number: int64(i + 1), Position: i + 1, Reason: "legacy", Record: raw})
	}
	return obs
}

func TestSuiteQueueSupplierNativeRevisionsAndLaterFastForward(t *testing.T) {
	owner1, supplier1, old := nativeSupplierFixture(t)
	owner2, supplier2, _ := nativeSupplierFixture(t)
	writeRepoFile(t, supplier2, "distinct", "second supplier\n")
	git(t, supplier2, "add", "distinct")
	git(t, supplier2, "commit", "-qm", "distinct")
	second := strings.TrimSpace(git(t, supplier2, "rev-parse", "HEAD"))
	obs := nativeSupplierTickets(t, owner1, owner2)
	io := suiteSupplierIO{resolve: resolveQueueSupplier, read: readQueueSupplier}
	before := annotateSuiteSuppliers(context.Background(), obs, io)
	if first, last := before.Suppliers[1], before.Suppliers[2]; first.Checkout != supplier1 || first.Revision != old || last.Checkout != supplier2 || last.Revision != second || first.Revision == last.Revision || first.Reason != "" || last.Reason != "" {
		t.Fatalf("distinct real suppliers not observed: %+v", before.Suppliers)
	}
	ticket := string(obs.Tickets[0].Record)
	git(t, supplier1, "switch", "-qc", "future")
	writeRepoFile(t, supplier1, "advance", "new source after old waiter started\n")
	git(t, supplier1, "add", "advance")
	git(t, supplier1, "commit", "-qm", "future")
	next := strings.TrimSpace(git(t, supplier1, "rev-parse", "HEAD"))
	git(t, supplier1, "switch", "-q", "main")
	git(t, supplier1, "merge", "--ff-only", "future")
	after := annotateSuiteSuppliers(context.Background(), obs, io)
	if after.Suppliers[1].Revision != next || string(obs.Tickets[0].Record) != ticket || after.Suppliers[2].Revision != second {
		t.Fatalf("stale cache or modified waiter: %+v", after.Suppliers)
	}
	var out bytes.Buffer
	renderSuiteQueue(&out, after)
	if !strings.Contains(out.String(), next) || strings.Count(out.String(), "launch revision unknown") != 2 {
		t.Fatal(out.String())
	}
}

func TestSuiteQueueSupplierNativeUnavailableOwners(t *testing.T) {
	for _, kind := range []string{"missing", "nested", "redirected", "git-failure"} {
		t.Run(kind, func(t *testing.T) {
			owner, supplier, _ := nativeSupplierFixture(t)
			switch kind {
			case "missing":
				owner = filepath.Join(t.TempDir(), "missing")
			case "nested":
				owner = filepath.Join(owner, "nested")
				if err := os.Mkdir(owner, 0700); err != nil {
					t.Fatal(err)
				}
			case "redirected":
				alias := filepath.Join(t.TempDir(), "owner")
				if err := os.Symlink(owner, alias); err != nil {
					t.Fatal(err)
				}
				owner = alias
			case "git-failure":
				resolved, err := resolveQueueSupplier(context.Background(), owner)
				if err != nil || resolved != supplier {
					t.Fatalf("fixture resolver: %s %v", resolved, err)
				}
				if err := os.WriteFile(filepath.Join(supplier, ".git", "HEAD"), []byte("ref: refs/heads/missing\n"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			obs := annotateSuiteSuppliers(context.Background(), nativeSupplierTickets(t, owner), suiteSupplierIO{resolve: resolveQueueSupplier, read: readQueueSupplier})
			row := obs.Suppliers[1]
			if row.Reason == "" || row.Revision != "" {
				t.Fatalf("%s manufactured supplier proof: %+v", kind, row)
			}
			var out bytes.Buffer
			renderSuiteQueue(&out, obs)
			if !strings.Contains(out.String(), "revision unavailable") || !strings.Contains(out.String(), "launch revision unknown") {
				t.Fatal(out.String())
			}
		})
	}
}

func TestSuiteQueueSupplierNativeFollowsHelperSymlink(t *testing.T) {
	owner, first, _ := nativeSupplierFixture(t)
	_, actual, sha := nativeSupplierFixture(t)
	path := filepath.Join(first, "scripts", "test-lock.py")
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(actual, "scripts", "test-lock.py"), path); err != nil {
		t.Fatal(err)
	}
	obs := annotateSuiteSuppliers(context.Background(), nativeSupplierTickets(t, owner), suiteSupplierIO{resolve: resolveQueueSupplier, read: readQueueSupplier})
	row := obs.Suppliers[1]
	if row.Reason != "" || row.Checkout != actual || row.Revision != sha {
		t.Fatalf("workflow source mistaken for helper source: %+v", row)
	}
}
