package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// readExistingWeeklySnapshot recognizes only a complete artifact for the exact
// requested week. Returning it before collection preserves its observed_at and
// every other immutable byte on repeat invocations.
func readExistingWeeklySnapshot(brainRoot string, window WeeklyWindow) (string, bool, error) {
	path := weeklySnapshotPath(brainRoot, window.End)
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("open existing weekly snapshot %s: %w", path, err)
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || info.Size() > weeklyPreviousLimit {
		return "", false, fmt.Errorf("existing weekly snapshot %s unreadable or exceeds 4 MiB", path)
	}
	data, err := io.ReadAll(io.LimitReader(f, weeklyPreviousLimit+1))
	if err != nil || len(data) > weeklyPreviousLimit {
		return "", false, fmt.Errorf("existing weekly snapshot %s unreadable or exceeds 4 MiB", path)
	}
	var snapshot WeeklySnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return "", false, fmt.Errorf("existing weekly snapshot %s invalid JSON: %w", path, err)
	}
	if snapshot.SchemaVersion != weeklySnapshotSchemaVersion || !snapshot.Window.Start.Equal(window.Start) || !snapshot.Window.End.Equal(window.End) || len(snapshot.Repos) == 0 {
		return "", false, fmt.Errorf("existing weekly snapshot %s has incompatible schema, window or fleet coverage", path)
	}
	seenRepos := map[string]bool{}
	for _, repo := range snapshot.Repos {
		if repo.Repo == "" || seenRepos[repo.Repo] {
			return "", false, fmt.Errorf("existing weekly snapshot %s has missing or duplicate repo identity", path)
		}
		seenRepos[repo.Repo] = true
	}
	if err := validatePersistedWeeklyResources(data, snapshot.Resources); err != nil {
		return "", false, fmt.Errorf("existing weekly snapshot %s invalid resources: %w", path, err)
	}
	return path, true, nil
}
