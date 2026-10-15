package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func hashBytes(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func storedAttemptFixture(id string) SuiteAttempt {
	now := time.Date(2026, 10, 14, 20, 0, 0, 0, time.UTC)
	return SuiteAttempt{Version: 1, ID: id, Repo: "x/ariadne", SHA: strings.Repeat("a", 40),
		Check: "suite", StartedAt: now, FinishedAt: now.Add(time.Second), Exit: 2,
		LogSHA256: strings.Repeat("b", 64), Failures: []string{"x.TestA"}}
}

func TestDecodeSuiteAttemptRejectsUntrustedRecords(t *testing.T) {
	valid := `{"version":1,"id":"a1","repo":"x/ariadne","sha":"` + strings.Repeat("a", 40) + `","check":"suite","started_at":"2026-10-14T20:00:00Z","finished_at":"2026-10-14T20:00:01Z","exit":2,"log_sha256":"` + strings.Repeat("b", 64) + `","log_path":"failed/a1.log"}`
	if got, err := decodeSuiteAttempt([]byte(valid)); err != nil || got.ID != "a1" {
		t.Fatalf("valid record: %+v %v", got, err)
	}
	for name, raw := range map[string]string{
		"truncated":     valid[:len(valid)-1],
		"version":       strings.Replace(valid, `"version":1`, `"version":2`, 1),
		"traversal":     strings.Replace(valid, `failed/a1.log`, `../../outside`, 1),
		"negative wall": strings.Replace(valid, `"exit":2`, `"wall_seconds":-1,"exit":2`, 1),
		"false success": strings.Replace(valid, `"exit":2`, `"exit":0,"timed_out":true`, 1),
		"unknown":       strings.Replace(valid, `"version":1`, `"version":1,"surprise":true`, 1),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := decodeSuiteAttempt([]byte(raw)); err == nil {
				t.Fatal("accepted malformed/untrusted record")
			}
		})
	}
	if _, err := decodeSuiteAttempt([]byte(valid + strings.Repeat(" ", 1<<20))); err == nil {
		t.Fatal("accepted oversized record")
	}
}

func TestPublishSuiteAttemptKeepsFailureAcrossRetry(t *testing.T) {
	root := t.TempDir()
	one, err := beginSuiteAttempt(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(one.LogPath, []byte("first failure\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	first := storedAttemptFixture(one.ID)
	first.LogSHA256 = hashBytes([]byte("first failure\n"))
	if _, err := publishSuiteAttempt(root, one, first); err != nil {
		t.Fatal(err)
	}
	two, err := beginSuiteAttempt(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(two.LogPath, []byte("retry passed\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	second := storedAttemptFixture(two.ID)
	second.Exit = 0
	second.FinishedAt = first.FinishedAt.Add(time.Second)
	second.LogSHA256 = hashBytes([]byte("retry passed\n"))
	if _, err := publishSuiteAttempt(root, two, second); err != nil {
		t.Fatal(err)
	}
	stored, warnings, err := readSuiteAttempts(root, 10)
	if err != nil || len(warnings) != 0 || len(stored) != 2 {
		t.Fatalf("read: %d %+v %v", len(stored), warnings, err)
	}
	failedLog := filepath.Join(root, "failed", one.ID+".log")
	if b, err := os.ReadFile(failedLog); err != nil || string(b) != "first failure\n" {
		t.Fatalf("failed log: %q %v", b, err)
	}
	if _, err := os.Stat(filepath.Join(root, "attempts", one.ID+".json")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "attempts", two.ID+".json")); err != nil {
		t.Fatal(err)
	}
}

func TestPublishSuiteAttemptNeverReplacesAnExistingFailedLog(t *testing.T) {
	root := t.TempDir()
	w, err := beginSuiteAttempt(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(w.LogPath, []byte("new"), 0o600); err != nil {
		t.Fatal(err)
	}
	failed := filepath.Join(root, "failed", w.ID+".log")
	if err := os.MkdirAll(filepath.Dir(failed), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(failed, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	a := storedAttemptFixture(w.ID)
	a.LogSHA256 = hashBytes([]byte("new"))
	if _, err := publishSuiteAttempt(root, w, a); err == nil {
		t.Fatal("overwrote failed log")
	}
	if b, err := os.ReadFile(failed); err != nil || string(b) != "old" {
		t.Fatalf("existing failure lost: %q %v", b, err)
	}
}

func TestPublishSuiteAttemptConcurrentPruneAndResumeAfterFailureLink(t *testing.T) {
	root := t.TempDir()
	first, err := beginSuiteAttempt(root)
	if err != nil {
		t.Fatal(err)
	}
	second, err := beginSuiteAttempt(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, w := range []*suiteAttemptWriter{first, second} {
		if err := os.WriteFile(w.LogPath, []byte(w.ID), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(filepath.Join(root, "failed"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Link(first.LogPath, filepath.Join(root, "failed", first.ID+".log")); err != nil {
		t.Fatal(err)
	}
	b := storedAttemptFixture(second.ID)
	b.LogSHA256 = hashBytes([]byte(second.ID))
	if _, err := publishSuiteAttempt(root, second, b); err != nil {
		t.Fatal(err)
	}
	if err := pruneSuiteAttempts(root, time.Now().Add(48*time.Hour)); err != nil {
		t.Fatal(err)
	}
	a := storedAttemptFixture(first.ID)
	a.LogSHA256 = hashBytes([]byte(first.ID))
	if _, err := publishSuiteAttempt(root, first, a); err != nil {
		t.Fatalf("same attempt cannot resume publication: %v", err)
	}
	rows, warnings, err := readSuiteAttempts(root, 10)
	if err != nil || len(warnings) != 0 || len(rows) != 2 {
		t.Fatalf("records: %d %v %v", len(rows), warnings, err)
	}
	for _, w := range []*suiteAttemptWriter{first, second} {
		if b, err := os.ReadFile(filepath.Join(root, "failed", w.ID+".log")); err != nil || string(b) != w.ID {
			t.Fatalf("failure log lost: %q %v", b, err)
		}
	}
}

func TestPruneSuiteAttemptsLeavesActiveWriter(t *testing.T) {
	root := t.TempDir()
	active, err := beginSuiteAttempt(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(active.LogPath, []byte("running"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := pruneSuiteAttempts(root, time.Now().Add(48*time.Hour)); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(active.LogPath); err != nil {
		t.Fatalf("active writer removed: %v", err)
	}
}

func TestPruneSuiteAttemptsPairsExpiredRecordsAndOrphans(t *testing.T) {
	root := t.TempDir()
	old := time.Now().Add(-32 * 24 * time.Hour)
	id := "abc123"
	for _, name := range []string{filepath.Join("attempts", id+".json"), filepath.Join("failed", id+".log")} {
		p := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("old"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(p, old, old); err != nil {
			t.Fatal(err)
		}
	}
	if err := pruneSuiteAttempts(root, time.Now()); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{filepath.Join("attempts", id+".json"), filepath.Join("failed", id+".log")} {
		if _, err := os.Stat(filepath.Join(root, name)); !os.IsNotExist(err) {
			t.Fatalf("retained expired %s: %v", name, err)
		}
	}
	// A crash after moving the failed log but before publishing the marker
	// leaves a failed log plus a dead owner's in-flight directory.
	orphan := "def456"
	inflight := filepath.Join(root, "inflight", orphan)
	if err := os.MkdirAll(inflight, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(inflight, "owner"), []byte("999999999"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(inflight, old, old); err != nil {
		t.Fatal(err)
	}
	orphanLog := filepath.Join(root, "failed", orphan+".log")
	if err := os.WriteFile(orphanLog, []byte("crash"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(orphanLog, old, old); err != nil {
		t.Fatal(err)
	}
	if err := pruneSuiteAttempts(root, time.Now()); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(orphanLog); !os.IsNotExist(err) {
		t.Fatalf("retained dead-owner orphan log: %v", err)
	}
	if _, err := os.Stat(inflight); !os.IsNotExist(err) {
		t.Fatalf("retained dead-owner inflight: %v", err)
	}
}

func TestReadSuiteAttemptsCapsTotalBytes(t *testing.T) {
	root := t.TempDir()
	for _, id := range []string{"aa11", "bb22"} {
		a := storedAttemptFixture(id)
		a.LogPath = filepath.Join("failed", id+".log")
		data, err := json.Marshal(a)
		if err != nil {
			t.Fatal(err)
		}
		p := filepath.Join(root, "attempts", id+".json")
		if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	rows, warnings, err := readSuiteAttemptsBudget(root, 10, 500)
	if err != nil || len(rows) != 1 || len(warnings) == 0 {
		t.Fatalf("unbounded intake: %d %v %v", len(rows), warnings, err)
	}
}
