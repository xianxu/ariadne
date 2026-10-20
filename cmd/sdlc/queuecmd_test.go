package main

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/queuegate"
)

func executeQueue(t *testing.T, d queueCommandDeps, args ...string) (string, error) {
	t.Helper()
	cmd := newQueueCmd(d)
	var out, diagnostics bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&diagnostics)
	cmd.SetArgs(args)
	err := cmd.Execute()
	if len(args) > 0 && args[0] == "status" && !strings.HasPrefix(out.String(), "{") {
		t.Fatalf("status stdout is not JSON: %q; diagnostic %q", out.String(), diagnostics.String())
	}
	return out.String(), err
}

func queueFixture(t *testing.T) queueCommandDeps {
	t.Helper()
	return queueCommandDeps{
		store: func() (*queuegate.Store, error) { return &queuegate.Store{Root: t.TempDir()}, nil },
		authority: landingDeps{
			caller:    func() (string, *int, bool) { n := 0; return "repo:0", &n, true },
			couchTree: func() (*int, bool) { return nil, false },
		},
		legacy: func(context.Context, []queuegate.Operation) ([]queuegate.Activity, []string) { return nil, nil },
	}
}

func TestQueueCLIFrozenProtocol(t *testing.T) {
	d := queueFixture(t)
	s := &queuegate.Store{Root: t.TempDir()}
	d.store = func() (*queuegate.Store, error) { return s, nil }
	for _, args := range [][]string{
		{"pause", "--reason", "reason is data; $(false)", "--owner", "couch-drain:a"},
		{"pause", "--reason", "second", "--owner", "couch-drain:b", "--cancel-running"},
		{"resume", "--owner", "couch-drain:a"},
	} {
		if _, err := executeQueue(t, d, args...); err != nil {
			t.Fatal(err)
		}
	}
	out, err := executeQueue(t, d, "status", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var snapshot queuegate.Snapshot
	if err := json.Unmarshal([]byte(out), &snapshot); err != nil {
		t.Fatal(err)
	}
	if err := queuegate.ValidateSnapshot(snapshot); err != nil {
		t.Fatal(err)
	}
	if snapshot.State != "paused" || !snapshot.Complete || len(snapshot.PauseOwners) != 1 || snapshot.PauseOwners[0].Owner != "couch-drain:b" || !snapshot.PauseOwners[0].CancelRunning {
		t.Fatalf("owner-scoped protocol: %s", out)
	}
}

func TestQueueAuthority(t *testing.T) {
	d := queueFixture(t)
	d.authority.caller = func() (string, *int, bool) { n := 3; return "repo:3", &n, true }
	called := false
	d.store = func() (*queuegate.Store, error) { called = true; return &queuegate.Store{Root: t.TempDir()}, nil }
	if _, err := executeQueue(t, d, "pause", "--owner", "forbidden"); err == nil || called {
		t.Fatalf("worker mutation reached storage: called=%v err=%v", called, err)
	}
	if _, err := executeQueue(t, d, "status", "--json"); err != nil || !called {
		t.Fatalf("worker cannot inspect: called=%v err=%v", called, err)
	}
}

func TestQueueAdmissionNeverScansLegacy(t *testing.T) {
	for _, n := range []int{8, 16} {
		d := queueFixture(t)
		s := &queuegate.Store{Root: t.TempDir()}
		d.store = func() (*queuegate.Store, error) { return s, nil }
		scans := 0
		d.legacy = func(context.Context, []queuegate.Operation) ([]queuegate.Activity, []string) {
			scans++
			return nil, nil
		}
		for i := 0; i < n; i++ {
			if _, err := executeQueue(t, d, "pause", "--owner", "owner"); err != nil {
				t.Fatal(err)
			}
			if _, err := executeQueue(t, d, "resume", "--owner", "owner"); err != nil {
				t.Fatal(err)
			}
		}
		if scans != 0 {
			t.Fatalf("%d mutations performed %d legacy scans", n, scans)
		}
		if _, err := executeQueue(t, d, "status", "--json"); err != nil {
			t.Fatal(err)
		}
		if scans != 1 {
			t.Fatalf("status did %d scans, want one", scans)
		}
	}
}
