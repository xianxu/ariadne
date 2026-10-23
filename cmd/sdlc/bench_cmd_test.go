package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBenchCLIStatusAndImmutableObservations(t *testing.T) {
	s, err := createBenchStore(filepath.Join(t.TempDir(), "run"), benchFixtureTask(t))
	if err != nil {
		t.Fatal(err)
	}
	if err := s.putData("fixture", map[string]string{"value": "original"}); err != nil {
		t.Fatal(err)
	}
	if err := s.putData("fixture", map[string]string{"value": "original"}); err != nil {
		t.Fatal(err)
	}
	if err := s.putData("fixture", map[string]string{"value": "changed"}); err == nil {
		t.Fatal("immutable observation replaced")
	}
	var output bytes.Buffer
	cmd := NewBenchCmd()
	cmd.SetOut(&output)
	cmd.SetArgs([]string{"status", "--run-dir", s.root, "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	var status map[string]any
	if err := json.Unmarshal(output.Bytes(), &status); err != nil {
		t.Fatal(err)
	}
	if status["phase"] != "new" || status["run_id"] == "" {
		t.Fatalf("bad status %s", output.String())
	}
	for _, private := range []string{"source_repo", "landing_oid", "acceptance", "original_issue_oid"} {
		if strings.Contains(output.String(), private) {
			t.Fatalf("private descriptor leaked: %s", output.String())
		}
	}
}

func TestBenchCLIRegistersWholeReplayAndPublicProtocol(t *testing.T) {
	for _, name := range []string{"prepare", "status", "run", "score", "record", "cancel", "clean", "decision", "finish"} {
		cmd := NewBenchCmd()
		found, _, err := cmd.Find([]string{name})
		if err != nil || found == cmd || found.Name() != name {
			t.Errorf("missing replay command %s: %v", name, err)
		}
	}
}

func TestBenchCLIPrepareChecksRetentionBeforeCreatingRun(t *testing.T) {
	parent := t.TempDir()
	task := benchFixtureTask(t)
	for i := 0; i < benchRetainedRunLimit; i++ {
		if _, err := createBenchStore(filepath.Join(parent, fmt.Sprintf("run-%d", i)), task); err != nil {
			t.Fatal(err)
		}
	}
	raw, err := json.Marshal(task)
	if err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(t.TempDir(), "task.json")
	if err := os.WriteFile(p, raw, 0600); err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(parent, "next-run")
	cmd := NewBenchCmd()
	cmd.SetArgs([]string{"prepare", "--task", p, "--run-dir", dest})
	if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "50-run cap") {
		t.Fatal("retention capacity not checked before effects", err)
	}
	if _, err := os.Lstat(dest); !os.IsNotExist(err) {
		t.Fatal("refused admission created a run", err)
	}
}
