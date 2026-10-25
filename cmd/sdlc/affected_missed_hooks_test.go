package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/affected"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/flakes"
)

// #616: the merge's suite comparison records missed mappings only when main passed and the merge introduced
// failures, and the gate's error is the original one in every case.
func TestDiagnoseSuiteFailureRecordsMissedMappings(t *testing.T) {
	const failing = "--- FAIL github.com/x/m/leaf TestA\nFAIL\n"
	cases := []struct {
		name     string
		mainCmd  []string
		recorder func(ctx context.Context, main, head string, introduced []string) (string, error)
		called   bool
		label    string
	}{
		{"main passed: recorded", []string{"true"}, func(_ context.Context, _, _ string, in []string) (string, error) {
			return "missed mappings: recorded 1 (sdlc affected missed list)", nil
		}, true, "missed mappings: recorded 1"},
		{"a recorder error is a label", []string{"true"}, func(context.Context, string, string, []string) (string, error) {
			return "", errors.New("boom")
		}, true, "missed mappings: not recorded (boom)"},
		{"nothing to say adds nothing", []string{"true"}, func(context.Context, string, string, []string) (string, error) { return "", nil }, true, ""},
		{"a red main records nothing", []string{"false"}, nil, false, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			d, sha, _, _ := localRunRepo(t, "acquired")
			log := filepath.Join(t.TempDir(), "merged.log")
			if err := os.WriteFile(log, []byte(failing), 0o600); err != nil {
				t.Fatal(err)
			}
			var got []string
			called := false
			d.RecordMissed = func(ctx context.Context, main, head string, introduced []string) (string, error) {
				called, got = true, introduced
				if c.recorder == nil {
					return "", nil
				}
				return c.recorder(ctx, main, head, introduced)
			}
			check := localCheck{Name: suiteCheckName, Command: c.mainCmd, TimeoutMinutes: 1}
			gateErr := &localCheckFailure{Check: check, Run: localRunRecord{Name: suiteCheckName, SHA: "head", Exit: 1, LogPath: log}, Head: "head"}
			err := diagnoseSuiteFailure(context.Background(), gateErr, sha, d)
			if !errors.Is(err, error(gateErr)) {
				t.Fatalf("the gate's error was replaced: %v", err)
			}
			if called != c.called || (called && !reflect.DeepEqual(got, []string{"github.com/x/m/leaf.TestA"})) {
				t.Fatalf("recorder called %v with %q", called, got)
			}
			if c.label != "" && !strings.Contains(err.Error(), c.label) || c.label == "" && strings.Contains(err.Error(), "missed mappings") {
				t.Fatalf("label: %v", err)
			}
		})
	}
}

// The production recorder plans the head's delta from main's affected.json over the head's graph: a change to a file
// one package owns that breaks another package's test is a missed mapping; a repo without affected.json records nothing.
func TestMissedMergeRecorderRealGit(t *testing.T) {
	d, _, change, _, _ := affectedFixture(t)
	must := func(args ...string) string {
		t.Helper()
		o, err := d.Git(args...)
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, o)
		}
		return strings.TrimSpace(o)
	}
	none := func(context.Context) (flakes.Registry, error) { return flakes.Registry{}, nil }
	record := missedMergeRecorder(d.Root, none)
	main := must("rev-parse", "main")
	head := change(map[string]string{"leaf/data.txt": "read by other\n"})
	if line, err := record(context.Background(), main, head, []string{"example.com/m/other.TestReads"}); err != nil || line != "" {
		t.Fatalf("without affected.json: %q %v", line, err)
	}
	must("switch", "-q", "main")
	if err := os.MkdirAll(filepath.Join(d.Root, ".sdlc"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(d.Root, ".sdlc", "affected.json"), []byte(`{"version":1,"command":["make","test-affected"],"full":["go.mod"]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	must("add", "-A")
	must("commit", "-q", "-m", "declare affected")
	main = must("rev-parse", "HEAD")
	must("switch", "-q", "br")
	must("merge", "-q", "--no-edit", "main")
	head = must("rev-parse", "HEAD")
	line, err := record(context.Background(), main, head, []string{"example.com/m/other.TestReads", "example.com/m/leaf.TestLeaf"})
	if err != nil || !strings.Contains(line, "recorded 1") {
		t.Fatalf("record: %q %v", line, err)
	}
	store, err := missedStoreFor(d.Root)
	if err != nil {
		t.Fatal(err)
	}
	l, err := store.read()
	if err != nil {
		t.Fatal(err)
	}
	want := affected.MissID("example.com/m/other", []string{"leaf/data.txt"})
	if len(l.Entries) != 1 || l.Entries[0].ID != want || l.Entries[0].Test != "example.com/m/other.TestReads" || l.Entries[0].Source != "merge" || l.Entries[0].Head != head {
		t.Fatalf("ledger: %+v", l.Entries)
	}
}
