package transcripts

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverEvidencePinsCodexIdentityAndPrefix(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "child.jsonl")
	data := `{"timestamp":"2026-10-14T00:15:00Z","type":"session_meta","payload":{"id":"child","session_id":"ancestor","cwd":"/repo","source":{"subagent":{"thread_spawn":{"parent_thread_id":"parent"}}}}}` + "\n" + `{"type":"event_msg","payload":{"text":"work"}}` + "\n"
	if err := os.WriteFile(path, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	got, _, err := DiscoverEvidence(context.Background(), []string{"/repo"}, []Harness{CodexHarness(root)})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("sessions=%v", got)
	}
	s := got[0]
	if s.ID != "child" || s.ParentID != "parent" || !s.Spawned || !s.Started.Equal(evidenceTime(15)) || s.Length != int64(len(data)) || s.Digest != fmt.Sprintf("%x", sha256.Sum256([]byte(data))) {
		t.Fatalf("evidence=%+v", s)
	}
}

func TestDiscoverEvidenceClaudeDoesNotRecurseSubagents(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, cwdToClaudeDir("/repo"))
	if err := os.MkdirAll(filepath.Join(dir, "subagents"), 0700); err != nil {
		t.Fatal(err)
	}
	data := []byte(`{"type":"user","timestamp":"2026-10-14T00:00:00Z","sessionId":"root","cwd":"/repo"}` + "\n")
	for _, path := range []string{filepath.Join(dir, "root.jsonl"), filepath.Join(dir, "subagents", "child.jsonl")} {
		if err := os.WriteFile(path, data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	got, _, err := DiscoverEvidence(context.Background(), []string{"/repo"}, []Harness{ClaudeHarness(root)})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ID != "root" || got[0].Harness != "claude" {
		t.Fatalf("got=%+v", got)
	}
}

func TestDiscoverEvidenceRejectsMalformedRelevantAndHonorsCancellation(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, cwdToClaudeDir("/repo"))
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "bad.jsonl"), []byte("broken\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := DiscoverEvidence(context.Background(), []string{"/repo"}, []Harness{ClaudeHarness(root)}); err == nil {
		t.Fatal("malformed relevant source accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, err := DiscoverEvidence(ctx, []string{"/repo"}, []Harness{CodexHarness(root)}); err == nil {
		t.Fatal("cancellation ignored")
	}
}

func TestDiscoverEvidenceRejectsMalformedRelevantCodexMetadata(t *testing.T) {
	root := t.TempDir()
	data := []byte(`{"timestamp":"invalid","type":"session_meta","payload":{"id":"root","cwd":"/repo"}}` + "\n")
	if err := os.WriteFile(filepath.Join(root, "bad.jsonl"), data, 0600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := DiscoverEvidence(context.Background(), []string{"/repo"}, []Harness{CodexHarness(root)}); err == nil {
		t.Fatal("relevant malformed Codex session accepted")
	}
}
