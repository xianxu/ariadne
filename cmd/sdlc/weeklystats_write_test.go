package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestPublishBrainSnapshotImmutable(t *testing.T) {
	root := t.TempDir()
	end := time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)
	want := filepath.Join(root, "memory", "work", "devinfra", "test-economy", "weekly", "2026-10-12.json")
	path, err := publishBrainSnapshot(root, end, []byte(`{"schema_version":1}`))
	if err != nil || path != want {
		t.Fatalf("publish path=%q err=%v, want %q", path, err, want)
	}
	if got, err := os.ReadFile(want); err != nil || !bytes.Equal(got, []byte(`{"schema_version":1}`)) {
		t.Fatalf("published bytes=%q err=%v", got, err)
	}
	if path, err := publishBrainSnapshot(root, end, []byte(`{"schema_version":1}`)); err != nil || path != want {
		t.Fatalf("identical rerun path=%q err=%v", path, err)
	}
	if _, err := publishBrainSnapshot(root, end, []byte(`{"schema_version":2}`)); err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("different rerun should name conflict path, got %v", err)
	}
	got, err := os.ReadFile(want)
	if err != nil || !bytes.Equal(got, []byte(`{"schema_version":1}`)) {
		t.Fatalf("conflict changed published bytes=%q err=%v", got, err)
	}
}

func TestPublishBrainSnapshotConcurrentDifferentWriters(t *testing.T) {
	root := t.TempDir()
	end := time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for _, data := range [][]byte{[]byte(`{"writer":1}`), []byte(`{"writer":2}`)} {
		wg.Add(1)
		go func(data []byte) {
			defer wg.Done()
			_, err := publishBrainSnapshot(root, end, data)
			results <- err
		}(data)
	}
	wg.Wait()
	close(results)
	successes, conflicts := 0, 0
	for err := range results {
		if err == nil {
			successes++
		} else if strings.Contains(err.Error(), "already exists") {
			conflicts++
		} else {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("successes=%d conflicts=%d", successes, conflicts)
	}
	got, err := os.ReadFile(filepath.Join(root, "memory/work/devinfra/test-economy/weekly/2026-10-12.json"))
	if err != nil || (!bytes.Equal(got, []byte(`{"writer":1}`)) && !bytes.Equal(got, []byte(`{"writer":2}`))) {
		t.Fatalf("published bytes=%q err=%v", got, err)
	}
}

func TestPublishBrainSnapshotCleansOnlyStaleOwnedTemps(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "memory/work/devinfra/test-economy/weekly")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	stale := filepath.Join(dir, ".weekly-snapshot-stale.tmp")
	fresh := filepath.Join(dir, ".weekly-snapshot-fresh.tmp")
	unowned := filepath.Join(dir, "other-stale.tmp")
	for _, path := range []string{stale, fresh, unowned} {
		if err := os.WriteFile(path, []byte("interrupted"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	old := time.Now().Add(-48 * time.Hour)
	for _, path := range []string{stale, unowned} {
		if err := os.Chtimes(path, old, old); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := publishBrainSnapshot(root, time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC), []byte("snapshot")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Fatalf("owned stale temp remains: %v", err)
	}
	for _, path := range []string{fresh, unowned} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("unexpected cleanup of %q: %v", path, err)
		}
	}
}
