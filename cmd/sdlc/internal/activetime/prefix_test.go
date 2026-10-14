package activetime

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestComputeUsesPinnedSourcePrefix(t *testing.T) {
	dir := eventsDir(t,
		`{"timestamp":"2026-01-01T00:00:00Z","type":"user","message":{"content":"#8"}}`,
		`{"timestamp":"2026-01-01T00:10:00Z","type":"user","message":{"content":"#8"}}`,
	)
	path := filepath.Join(dir, "s.jsonl")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	prefix := SourcePrefix{Bytes: int64(len(data)), Digest: fmt.Sprintf("%x", sha256.Sum256(data))}
	withGitRun(t, func(repo string, args ...string) ([]byte, error) { return nil, nil })
	options := Options{Files: []string{path}, GitRepo: "/repo", SinceISO: wideSince, UntilISO: wideUntil, Issues: []string{"8"}, CommitWeight: 1, ThresholdMin: 15, IncludeAssistant: true, SourcePrefixes: map[string]SourcePrefix{path: prefix}}
	if err := os.WriteFile(path, append(data, []byte("{\"timestamp\":\"2026-01-01T00:20:00Z\",\"type\":\"user\",\"message\":{\"content\":\"#8\"}}\n")...), 0600); err != nil {
		t.Fatal(err)
	}
	result, err := Compute(options)
	if err != nil || result.PerIssue["8"] != 10 {
		t.Fatalf("later events changed pinned measurement: %+v %v", result, err)
	}
	data[0] = ' '
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Compute(options); err == nil {
		t.Fatal("changed evidence prefix was silently measured")
	}
}
