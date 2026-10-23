package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBenchAuditRecordsRefusalsAndGuardsBeforeEffect(t *testing.T) {
	c := benchContextFixture(t)
	ctx := context.Background()
	failure := errors.New("real gate refused")
	called := 0
	if err := withBenchCommand(ctx, &c, c.WorkRoot, []string{"close"}, func() error { called++; return failure }); !errors.Is(err, failure) {
		t.Fatal(err)
	}
	if called != 1 {
		t.Fatal("gate not called")
	}
	events, err := readBenchAudit(ctx, &c)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 || events[0].Kind != "command-start" || events[1].Kind != "command-end" || events[1].ParentID != events[0].ID || events[1].Success == nil || *events[1].Success || events[1].Error != "real gate refused" {
		t.Fatalf("refusal lost: %+v", events)
	}
	if err := withBenchCommand(ctx, &c, c.WorkRoot, []string{"merge"}, func() error { called++; return nil }); err == nil {
		t.Fatal("merge allowed")
	}
	if called != 1 {
		t.Fatal("forbidden effect executed")
	}
	events, err = readBenchAudit(ctx, &c)
	if err != nil || len(events) != 4 || events[3].Success == nil || *events[3].Success {
		t.Fatalf("guard refusal not recorded: %+v %v", events, err)
	}
}

func TestBenchAuditPartialTailAndCapPreventNewEffects(t *testing.T) {
	for _, kind := range []string{"partial-tail", "oversize", "symlink"} {
		t.Run(kind, func(t *testing.T) {
			c := benchContextFixture(t)
			ctx := context.Background()
			if err := withBenchCommand(ctx, &c, c.WorkRoot, []string{"actual"}, func() error { return nil }); err != nil {
				t.Fatal(err)
			}
			path := benchAuditPath(&c)
			switch kind {
			case "partial-tail":
				f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
				if err != nil {
					t.Fatal(err)
				}
				_, err = f.WriteString(`{"version":`)
				if err != nil {
					t.Fatal(err)
				}
				if err := f.Close(); err != nil {
					t.Fatal(err)
				}
			case "oversize":
				if err := os.WriteFile(path, []byte(strings.Repeat(" ", benchRecordLimit+1)), 0600); err != nil {
					t.Fatal(err)
				}
			case "symlink":
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
				target := filepath.Join(t.TempDir(), "victim")
				if err := os.WriteFile(target, []byte("untouched"), 0600); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(target, path); err != nil {
					t.Fatal(err)
				}
			}
			called := false
			if err := withBenchCommand(ctx, &c, c.WorkRoot, []string{"claim"}, func() error { called = true; return nil }); err == nil || called {
				t.Fatalf("audit failure allowed effect: %v %v", called, err)
			}
		})
	}
}

func TestBenchAuditProposalsBindRunAndPreserveFutureFields(t *testing.T) {
	c := benchContextFixture(t)
	ctx := context.Background()
	e := benchAuditEvent{Version: 1, ID: "proposal-1", RunID: c.RunID, Kind: "decision", At: "2026-10-23T20:00:00Z", Root: c.HubRoot, Slot: "ops-bench:0", Session: "s", Generation: "g", DecisionKind: "clarify", Reason: "original requirement"}
	if err := appendBenchAudit(ctx, &c, e); err != nil {
		t.Fatal(err)
	}
	if err := appendBenchAudit(ctx, &c, e); err != nil {
		t.Fatal("identical proposal not idempotent", err)
	}
	e.Reason = "different"
	if err := appendBenchAudit(ctx, &c, e); err == nil {
		t.Fatal("conflicting proposal discarded")
	}
	e.ID = "proposal-2"
	e.RunID = "different"
	if err := appendBenchAudit(ctx, &c, e); err == nil {
		t.Fatal("wrong run proposal accepted")
	}
	events, err := readBenchAudit(ctx, &c)
	if err != nil || len(events) != 1 {
		t.Fatalf("bad proposal history %+v %v", events, err)
	}
}

func TestBenchAuditLegacyExitRecordsARefusal(t *testing.T) {
	c := benchContextFixture(t)
	func() {
		defer func() {
			if recover() == nil {
				t.Error("fixture did not exit")
			}
		}()
		_ = withBenchCommand(context.Background(), &c, c.WorkRoot, []string{"close"}, func() error { runDieCleanups(); panic("simulated process exit") })
	}()
	events, err := readBenchAudit(context.Background(), &c)
	if err != nil || len(events) != 2 || events[1].Success == nil || *events[1].Success {
		t.Fatalf("fatal refusal lost: %+v %v", events, err)
	}
}
