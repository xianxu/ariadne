package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestReadPreviousResourceSectionExactWeek(t *testing.T) {
	root := t.TempDir()
	end := time.Date(2026, 10, 19, 0, 0, 0, 0, time.UTC)
	window, err := newWeeklyWindow(end, end)
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(root, "memory/work/devinfra/test-economy/weekly")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	priorPath := filepath.Join(dir, "2026-10-12.json")
	if got := readPreviousResourceSection(root, window); got.State != "baseline" || !strings.Contains(got.Reason, "missing") {
		t.Fatalf("missing exact prior = %+v", got)
	}
	write := func(body string) {
		t.Helper()
		if err := os.WriteFile(priorPath, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(`{"schema_version":1,"window":{"start":"2026-10-05T00:00:00Z","end":"2026-10-12T00:00:00Z"},"repos":[]}`)
	if got := readPreviousResourceSection(root, window); got.State != "baseline" || !strings.Contains(got.Reason, "v1") {
		t.Fatalf("v1 has no resource baseline: %+v", got)
	}
	write(`{"schema_version":2,"window":{"start":"2026-10-04T00:00:00Z","end":"2026-10-11T00:00:00Z"},"resources":{"families":[]}}`)
	if got := readPreviousResourceSection(root, window); got.State != "unknown" || !strings.Contains(got.Reason, "window") {
		t.Fatalf("mismatched window = %+v", got)
	}
	write(`{"schema_version":2,"window":{"start":"2026-10-05T00:00:00Z","end":"2026-10-12T00:00:00Z"},"resources":{"families":[{"kind":"couch-ledgers","definition":"v1","bytes":100,"complete":true}]}}`)
	got := readPreviousResourceSection(root, window)
	if got.State != "complete" || got.Section == nil || len(got.Section.Families) != 1 || got.Section.Families[0].Bytes != 100 {
		t.Fatalf("valid predecessor = %+v", got)
	}
	write(`{`)
	if got := readPreviousResourceSection(root, window); got.State != "unknown" || !strings.Contains(got.Reason, "decode") {
		t.Fatalf("malformed predecessor = %+v", got)
	}
}
