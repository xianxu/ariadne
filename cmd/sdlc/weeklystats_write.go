package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const weeklySnapshotTempPrefix = ".weekly-snapshot-"

// publishBrainSnapshot publishes final JSON bytes once. A rerun with identical
// bytes is harmless; a differing rerun cannot replace the first publication.
func publishBrainSnapshot(brainRoot string, end time.Time, data []byte) (string, error) {
	if brainRoot == "" {
		return "", errors.New("brain root is empty")
	}
	if end.IsZero() || end.Location() != time.UTC || end.Hour() != 0 || end.Minute() != 0 || end.Second() != 0 || end.Nanosecond() != 0 {
		return "", errors.New("snapshot end must be midnight UTC")
	}
	dir := filepath.Join(brainRoot, "memory", "work", "devinfra", "test-economy", "weekly")
	path := filepath.Join(dir, end.Format("2006-01-02")+".json")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create weekly snapshot directory: %w", err)
	}
	if err := cleanupStaleWeeklySnapshotTemps(dir, time.Now()); err != nil {
		return "", err
	}
	// CreateTemp uses O_EXCL and keeps each concurrent writer's bytes separate.
	tmp, err := os.CreateTemp(dir, weeklySnapshotTempPrefix+"*.tmp")
	if err != nil {
		return "", fmt.Errorf("create weekly snapshot temp: %w", err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return "", fmt.Errorf("write weekly snapshot temp: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return "", fmt.Errorf("sync weekly snapshot temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return "", fmt.Errorf("close weekly snapshot temp: %w", err)
	}
	if err := os.Link(tmp.Name(), path); err != nil {
		if errors.Is(err, os.ErrExist) {
			existing, readErr := os.ReadFile(path)
			if readErr != nil {
				return "", fmt.Errorf("read existing weekly snapshot %s: %w", path, readErr)
			}
			if bytes.Equal(existing, data) {
				return path, nil
			}
			return "", fmt.Errorf("weekly snapshot %s already exists with different bytes", path)
		}
		return "", fmt.Errorf("publish weekly snapshot %s: %w", path, err)
	}
	if err := syncWeeklySnapshotDir(dir); err != nil {
		return "", fmt.Errorf("sync weekly snapshot directory %s: %w", dir, err)
	}
	return path, nil
}

func cleanupStaleWeeklySnapshotTemps(dir string, now time.Time) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read weekly snapshot directory: %w", err)
	}
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), weeklySnapshotTempPrefix) || !strings.HasSuffix(entry.Name(), ".tmp") || !entry.Type().IsRegular() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return fmt.Errorf("stat weekly snapshot temp %s: %w", entry.Name(), err)
		}
		if now.Sub(info.ModTime()) <= 24*time.Hour {
			continue
		}
		if err := os.Remove(filepath.Join(dir, entry.Name())); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("remove stale weekly snapshot temp %s: %w", entry.Name(), err)
		}
	}
	return nil
}

func syncWeeklySnapshotDir(dir string) error {
	f, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer f.Close()
	return f.Sync()
}
