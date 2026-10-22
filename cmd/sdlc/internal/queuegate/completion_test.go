package queuegate

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

var directSuiteFixtureIDs = []string{
	"65dbdc31d6444c45a8692456d1e0a595",
	"f70ce4520db34fdaa8208eb03bf778c0",
	"b54088501a804c8eab8d7bab2de27e79",
}

func directSuiteFixture(t *testing.T, id string) ([]byte, Operation) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("../../../../scripts/test/fixtures/direct-suite-completion", id+".json"))
	if err != nil {
		t.Fatal(err)
	}
	op, err := DecodeOperation(raw)
	if err != nil {
		t.Fatal(err)
	}
	if op.ID != id || op.Managed || op.Phase != "unknown" {
		t.Fatalf("not an original direct-suite tombstone: %+v", op)
	}
	return raw, op
}

func directSuiteExitEvidence(pid int) map[string]json.RawMessage {
	group, _ := json.Marshal(pid)
	return map[string]json.RawMessage{
		"owner_observation":   json.RawMessage(`"exact"`),
		"status_observation":  json.RawMessage(`"observed"`),
		"exit_status":         json.RawMessage(`0`),
		"exit_observed_at":    json.RawMessage(`"2026-10-22T05:00:00.123456+00:00"`),
		"process_group":       group,
		"group_observation":   json.RawMessage(`"empty"`),
		"group_observed_at":   json.RawMessage(`"2026-10-22T05:00:00.124000+00:00"`),
		"survivor_pids":       json.RawMessage(`[]`),
		"survivors_truncated": json.RawMessage(`false`),
		"future_completion":   json.RawMessage(`{ "nested" : [null, {"number":1e99999}] }`),
	}
}

func assertDirectSuiteEvidence(t *testing.T, op Operation, expected map[string]json.RawMessage) {
	t.Helper()
	for key, want := range expected {
		if !bytes.Equal(op.Extra[key], want) {
			t.Fatalf("completion evidence %s changed: got %s, want %s", key, op.Extra[key], want)
		}
	}
}

func TestDirectSuiteCompletionCrashCutsFreshObserver(t *testing.T) {
	for _, phase := range []string{"running", "finished"} {
		t.Run(phase, func(t *testing.T) {
			_, op := directSuiteFixture(t, directSuiteFixtureIDs[1])
			// Derive the two publication cuts from the real incident identity.
			// Even complete-looking diagnostics are not a terminal phase commit.
			op.Phase = phase
			op.Extra = directSuiteExitEvidence(op.PID)
			root := t.TempDir()
			path := filepath.Join(root, "operations", op.ID+".json")
			raw, err := EncodeOperation(op)
			if err != nil {
				t.Fatal(err)
			}
			if err := atomicWrite(path, raw); err != nil {
				t.Fatal(err)
			}
			for pass := 0; pass < 2; pass++ {
				fresh := &Store{Root: root}
				x, observed, err := fresh.Observe(context.Background())
				if err != nil {
					t.Fatal(err)
				}
				if phase == "finished" {
					if !x.Complete || len(x.Active) != 0 || len(observed) != 0 {
						t.Fatalf("durable completion did not clean up: %+v %+v", x, observed)
					}
				} else {
					if x.Complete || len(observed) != 1 || observed[0].Phase != "unknown" {
						t.Fatalf("uncommitted completion promoted to success: %+v %+v", x, observed)
					}
					assertDirectSuiteEvidence(t, observed[0], op.Extra)
				}
			}
			if _, err := os.Stat(path); !os.IsNotExist(err) {
				t.Fatalf("crashed operation residue retained: %v", err)
			}
			_, err = os.Stat(filepath.Join(root, "unresolved", op.ID+".json"))
			if phase == "running" && err != nil || phase == "finished" && !os.IsNotExist(err) {
				t.Fatalf("wrong tombstone disposition for %s: %v", phase, err)
			}
		})
	}
}

func TestDirectSuiteOriginalTombstonesNeverPromotedByFreshObserver(t *testing.T) {
	for _, id := range directSuiteFixtureIDs {
		t.Run(id, func(t *testing.T) {
			raw, op := directSuiteFixture(t, id)
			root := t.TempDir()
			path := filepath.Join(root, "unresolved", id+".json")
			if err := atomicWrite(path, raw); err != nil {
				t.Fatal(err)
			}
			for pass := 0; pass < 2; pass++ {
				fresh := &Store{Root: root, BootID: func() (string, error) { return op.BootID, nil }}
				x, observed, err := fresh.Observe(context.Background())
				if err != nil || x.Complete || len(observed) != 1 || observed[0].Phase != "unknown" {
					t.Fatalf("historical uncertainty erased: %+v %+v %v", x, observed, err)
				}
				retained, err := os.ReadFile(path)
				if err != nil || !bytes.Equal(retained, raw) {
					t.Fatalf("historical tombstone rewritten: %s %v", retained, err)
				}
			}
		})
	}
}
