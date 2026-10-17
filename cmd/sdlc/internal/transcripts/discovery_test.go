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

func TestDiscoverEvidencePinsResumedCodexSessionInOriginalDateFolder(t *testing.T) {
	root := t.TempDir()
	day := filepath.Join(root, "2026", "10", "13")
	if err := os.MkdirAll(day, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(day, "rollout-root.jsonl")
	first := `{"timestamp":"2026-10-13T22:56:13Z","type":"session_meta","payload":{"id":"root","cwd":"/repo","source":"cli"}}` + "\n"
	if err := os.WriteFile(path, []byte(first), 0600); err != nil {
		t.Fatal(err)
	}
	appended := `{"timestamp":"2026-10-17T09:59:24Z","type":"response_item","payload":{"type":"custom_tool_call","name":"exec","call_id":"suite","input":"text(await tools.exec_command({cmd:\"sdlc local-check run --name suite\"}))"}}` + "\n"
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(appended); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	got, _, err := DiscoverEvidence(context.Background(), []string{"/repo"}, []Harness{CodexHarness(root)})
	if err != nil {
		t.Fatal(err)
	}
	all := first + appended
	if len(got) != 1 || got[0].File != path || got[0].ID != "root" || got[0].Length != int64(len(all)) || got[0].Digest != fmt.Sprintf("%x", sha256.Sum256([]byte(all))) {
		t.Fatalf("resumed source in old date folder not pinned with appended bytes: %+v", got)
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
