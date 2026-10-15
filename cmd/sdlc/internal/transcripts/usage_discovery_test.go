package transcripts

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func usageFixture(t *testing.T, root, name, body string) string {
	t.Helper()
	p := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	return p
}
func usageMeta(id, cwd, parent string) string {
	source := json.RawMessage(`"cli"`)
	if parent != "" {
		source = json.RawMessage(`{"subagent":{"thread_spawn":{"parent_thread_id":"` + parent + `"}}}`)
	}
	b, _ := json.Marshal(map[string]any{"type": "session_meta", "timestamp": "2026-10-14T00:00:00Z", "payload": map[string]any{"id": id, "session_id": "copied-parent", "cwd": cwd, "source": source}})
	return string(b) + "\n"
}
func TestUsageDiscoveryFirstMetadataUnknownAndNoHash(t *testing.T) {
	root := t.TempDir()
	usageFixture(t, root, "child.jsonl", usageMeta("child", "/unknown/project", "parent")+usageMeta("copied-parent", "/parent", "")+strings.Repeat("x", 2<<20))
	unknown := usageFixture(t, root, "unknown.jsonl", "{broken\n")
	got, err := DiscoverUsageEvidence(context.Background(), []Harness{CodexHarness(root)}, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Sources) != 2 || len(got.UnknownFiles) != 2 || got.UnknownFiles[1] != unknown {
		t.Fatalf("unknown spend hidden: %+v", got)
	}
	var child UsageSource
	for _, s := range got.Sources {
		if strings.HasSuffix(s.File, "child.jsonl") {
			child = s
		}
	}
	if child.ID != "child" || child.CWD != "/unknown/project" || child.ParentID != "parent" || !child.Spawned || !child.MetadataComplete || child.Digest != "" || child.Length < 2<<20 {
		t.Fatalf("copied metadata replaced identity / whole source hashed: %+v", child)
	}
	if len(got.Sessions) != 0 {
		t.Fatalf("missing parent admitted: %+v", got.Sessions)
	}
}
func TestUsageDiscoveryClaudeChildDispatchAndOldParent(t *testing.T) {
	root := t.TempDir()
	dir := "encoded-project"
	parent := usageFixture(t, root, filepath.Join(dir, "parent.jsonl"), `{"type":"user","timestamp":"2026-10-14T00:00:00Z","sessionId":"parent","cwd":"/slot3/repo"}`+"\n"+`{"type":"assistant","timestamp":"2026-10-14T00:12:00Z","sessionId":"parent","message":{"content":[{"type":"tool_use","name":"Task","id":"tool-1","input":{}}]}}`+"\n")
	child := usageFixture(t, root, filepath.Join(dir, "parent", "subagents", "agent-worker.jsonl"), `{"type":"assistant","timestamp":"2026-10-14T00:13:00Z","sessionId":"parent","agentId":"worker","cwd":"/different/child","isSidechain":true}`+"\n")
	usageFixture(t, root, strings.TrimSuffix(strings.TrimPrefix(child, root+string(os.PathSeparator)), ".jsonl")+".meta.json", `{"toolUseId":"tool-1"}`)
	old := time.Date(2026, 10, 13, 0, 0, 0, 0, time.UTC)
	if err := os.Chtimes(parent, old, old); err != nil {
		t.Fatal(err)
	}
	got, err := DiscoverUsageEvidence(context.Background(), []Harness{ClaudeHarness(root)}, time.Date(2026, 10, 14, 0, 10, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Sources) != 1 || len(got.Sessions) != 2 {
		t.Fatalf("old parent proof unavailable / pruning broken: %+v", got)
	}
	s := got.Sources[0]
	if s.File != child || s.ID != "parent/agent/worker" || s.ParentID != "parent" || s.CWD != "/slot3/repo" || !s.Spawned || !s.Started.Equal(time.Date(2026, 10, 14, 0, 12, 0, 0, time.UTC)) || s.LineageStatus != "verified" {
		t.Fatalf("child lacks explicit dispatch lineage: %+v", s)
	}
}
func TestUsageDiscoveryUnboundClaudeChildStaysUnknown(t *testing.T) {
	root := t.TempDir()
	usageFixture(t, root, "project/parent.jsonl", `{"type":"user","timestamp":"2026-10-14T00:00:00Z","sessionId":"parent","cwd":"/repo"}`+"\n")
	usageFixture(t, root, "project/parent/subagents/agent-worker.jsonl", `{"type":"assistant","timestamp":"2026-10-14T00:13:00Z","sessionId":"parent","agentId":"worker","cwd":"/repo"}`+"\n")
	got, err := DiscoverUsageEvidence(context.Background(), []Harness{ClaudeHarness(root)}, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Sources) != 2 || len(got.Sessions) != 1 || len(got.UnknownFiles) != 1 || got.Diagnostics["invalid_lineage"] != 1 {
		t.Fatalf("unbound child guessed from cwd: %+v", got)
	}
}
func TestUsageDiscoveryCancellationAndBoundedExamples(t *testing.T) {
	root := t.TempDir()
	for i := 0; i < 6; i++ {
		usageFixture(t, root, strings.Repeat("a", i+1)+".jsonl", "bad\n")
	}
	got, err := DiscoverUsageEvidence(context.Background(), []Harness{CodexHarness(root)}, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if got.Diagnostics["unknown_metadata"] != 6 || len(got.Examples["unknown_metadata"]) != 3 || len(got.Sources) != 6 {
		t.Fatalf("diagnostics unbounded or unknowns dropped: %+v", got)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := DiscoverUsageEvidence(ctx, []Harness{CodexHarness(root)}, time.Time{}); err == nil {
		t.Fatal("cancelled discovery continued")
	}
}
func TestUsageDiscoveryCopiedParentMetadataKeepsCodexChildBound(t *testing.T) {
	root := t.TempDir()
	usageFixture(t, root, "parent.jsonl", usageMeta("parent", "/repo", ""))
	usageFixture(t, root, "child.jsonl", usageMeta("child", "/brain", "parent")+usageMeta("parent", "/repo", ""))
	got, err := DiscoverUsageEvidence(context.Background(), []Harness{CodexHarness(root)}, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Sessions) != 2 {
		t.Fatalf("valid lineage rejected: %+v", got)
	}
	for _, s := range got.Sessions {
		if strings.HasSuffix(s.File, "child.jsonl") && s.ID != "child" {
			t.Fatalf("copied parent identity won: %+v", s)
		}
	}
}

func TestUsageDiscoveryNestedClaudeLineageAndDuplicateDispatch(t *testing.T) {
	root := t.TempDir()
	usageFixture(t, root, "project/parent.jsonl", `{"type":"user","timestamp":"2026-10-14T00:00:00Z","sessionId":"parent","cwd":"/repo"}`+"\n"+strings.Repeat(`{"type":"assistant","timestamp":"2026-10-14T00:02:00Z","message":{"content":[{"type":"tool_use","name":"Agent","id":"tool-parent"}]}}`+"\n", 2))
	usageFixture(t, root, "project/parent/subagents/agent-first.jsonl", `{"type":"assistant","timestamp":"2026-10-14T00:03:00Z","sessionId":"parent","agentId":"first"}`+"\n"+`{"type":"assistant","timestamp":"2026-10-14T00:04:00Z","message":{"content":[{"type":"tool_use","name":"Agent","id":"tool-child"}]}}`+"\n")
	usageFixture(t, root, "project/parent/subagents/agent-first.meta.json", `{"toolUseId":"tool-parent"}`)
	usageFixture(t, root, "project/parent/subagents/agent-second.jsonl", `{"type":"assistant","timestamp":"2026-10-14T00:05:00Z","sessionId":"parent","agentId":"second"}`+"\n")
	usageFixture(t, root, "project/parent/subagents/agent-second.meta.json", `{"toolUseId":"tool-child"}`)
	got, err := DiscoverUsageEvidence(context.Background(), []Harness{ClaudeHarness(root)}, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Sessions) != 3 {
		t.Fatalf("nested lineage lost: %+v", got)
	}
	for _, s := range got.Sessions {
		if s.ID == "parent/agent/second" && (s.ParentID != "parent/agent/first" || s.CWD != "/repo" || s.Started.Minute() != 4) {
			t.Fatalf("nested dispatch guessed: %+v", s)
		}
	}
}
func TestUsageDiscoveryRejectsCyclesConflictsAndFirstMalformedMetadata(t *testing.T) {
	root := t.TempDir()
	usageFixture(t, root, "a.jsonl", usageMeta("a", "/repo", "b"))
	usageFixture(t, root, "b.jsonl", usageMeta("b", "/repo", "a"))
	malformed := usageFixture(t, root, "invalid.jsonl", `{"type":"session_meta","timestamp":"bad","payload":{"id":"bad","cwd":"/repo"}}`+"\n"+usageMeta("good", "/repo", ""))
	usageFixture(t, root, "copy1.jsonl", usageMeta("duplicate", "/repo", ""))
	usageFixture(t, root, "copy2.jsonl", usageMeta("duplicate", "/elsewhere", ""))
	got, err := DiscoverUsageEvidence(context.Background(), []Harness{CodexHarness(root)}, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Sessions) != 0 || len(got.Sources) != 5 || len(got.UnknownFiles) != 5 {
		t.Fatalf("conflicting source admitted: %+v", got)
	}
	for _, s := range got.Sources {
		if s.File == malformed && (!strings.HasPrefix(s.ID, "file:") || s.MetadataComplete) {
			t.Fatalf("later metadata repaired original identity: %+v", s)
		}
	}
}
func TestUsageDiscoveryCandidateAndLineLimits(t *testing.T) {
	root := t.TempDir()
	for _, n := range []string{"a", "b", "c"} {
		usageFixture(t, root, n+".jsonl", usageMeta(n, "/repo", ""))
	}
	got, err := discoverUsageEvidence(context.Background(), []Harness{CodexHarness(root)}, time.Time{}, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Sources) != 2 || got.Diagnostics["candidate_limit"] != 1 || got.Diagnostics["candidates"] != 2 {
		t.Fatalf("candidate cap ignored: %+v", got)
	}
	bigroot := t.TempDir()
	usageFixture(t, bigroot, "huge.jsonl", strings.Repeat("x", UsageLineLimit+1)+"\n")
	got, err = DiscoverUsageEvidence(context.Background(), []Harness{CodexHarness(bigroot)}, time.Time{})
	if err != nil || len(got.Sources) != 1 || got.Diagnostics["unknown_metadata"] != 1 {
		t.Fatalf("large line hid source / bypassed bounds: %+v %v", got, err)
	}
}

func TestUsageDiscoveryOldCodexParentAndMetaFileBounds(t *testing.T) {
	root := t.TempDir()
	parent := usageFixture(t, root, "parent.jsonl", usageMeta("parent", "/repo", ""))
	usageFixture(t, root, "child.jsonl", usageMeta("child", "/brain", "parent"))
	old := time.Date(2026, 10, 13, 0, 0, 0, 0, time.UTC)
	if err := os.Chtimes(parent, old, old); err != nil {
		t.Fatal(err)
	}
	got, err := DiscoverUsageEvidence(context.Background(), []Harness{CodexHarness(root)}, old.Add(24*time.Hour))
	if err != nil || len(got.Sources) != 1 || len(got.Sessions) != 2 {
		t.Fatalf("required old parent dropped: %+v %v", got, err)
	}
	cr := t.TempDir()
	usageFixture(t, cr, "project/parent.jsonl", `{"type":"user","timestamp":"2026-10-14T00:00:00Z","sessionId":"parent","cwd":"/repo"}`+"\n"+`{"type":"assistant","timestamp":"2026-10-14T00:12:00Z","message":{"content":[{"type":"tool_use","name":"Task","id":"tool-1"}]}}`+"\n")
	usageFixture(t, cr, "project/parent/subagents/agent-worker.jsonl", `{"type":"assistant","timestamp":"2026-10-14T00:13:00Z","sessionId":"parent","agentId":"worker","cwd":"/repo"}`+"\n")
	usageFixture(t, cr, "project/parent/subagents/agent-worker.meta.json", `{"toolUseId":"tool-1"}`+strings.Repeat(" ", 64<<10))
	got, err = DiscoverUsageEvidence(context.Background(), []Harness{ClaudeHarness(cr)}, time.Time{})
	if err != nil || len(got.Sessions) != 1 || got.Diagnostics["invalid_lineage"] != 1 {
		t.Fatalf("oversized meta accepted: %+v %v", got, err)
	}
}

func TestUsageDiscoveryRejectsConflictingAndFutureClaudeDispatch(t *testing.T) {
	for _, stamps := range [][]string{{"2026-10-14T00:12:00Z", "2026-10-14T00:14:00Z"}, {"2026-10-14T00:14:00Z"}} {
		root := t.TempDir()
		body := `{"type":"user","timestamp":"2026-10-14T00:00:00Z","sessionId":"parent","cwd":"/repo"}` + "\n"
		for _, stamp := range stamps {
			body += `{"type":"assistant","timestamp":"` + stamp + `","message":{"content":[{"type":"tool_use","name":"Agent","id":"tool-1"}]}}` + "\n"
		}
		usageFixture(t, root, "project/parent.jsonl", body)
		usageFixture(t, root, "project/parent/subagents/agent-worker.jsonl", `{"type":"assistant","timestamp":"2026-10-14T00:13:00Z","sessionId":"parent","agentId":"worker","cwd":"/repo"}`+"\n")
		usageFixture(t, root, "project/parent/subagents/agent-worker.meta.json", `{"toolUseId":"tool-1"}`)
		got, err := DiscoverUsageEvidence(context.Background(), []Harness{ClaudeHarness(root)}, time.Time{})
		if err != nil || len(got.Sessions) != 1 || got.Diagnostics["invalid_lineage"] != 1 {
			t.Fatalf("conflicting/future dispatch accepted: %+v %v", got, err)
		}
	}
}

func TestUsageDiscoveryPinsFileIdentityAndModificationTime(t *testing.T) {
	root := t.TempDir()
	path := usageFixture(t, root, "session.jsonl", usageMeta("session", "/repo", ""))
	got, err := DiscoverUsageEvidence(context.Background(), []Harness{CodexHarness(root)}, time.Time{})
	if err != nil || len(got.Sources) != 1 {
		t.Fatalf("discover source: %+v %v", got, err)
	}
	source := got.Sources[0]
	initial, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if source.Info == nil || !os.SameFile(source.Info, initial) || source.Info.Size() != source.Length || !source.Info.ModTime().Equal(initial.ModTime()) {
		t.Fatalf("source stat identity not pinned: %+v", source)
	}
	// Equal-size rewrites keep the inode, but sampled mtime still detects them.
	if err := os.WriteFile(path, []byte(usageMeta("session", "/work", "")), 0600); err != nil {
		t.Fatal(err)
	}
	changedAt := initial.ModTime().Add(time.Second)
	if err := os.Chtimes(path, changedAt, changedAt); err != nil {
		t.Fatal(err)
	}
	rewritten, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if !os.SameFile(source.Info, rewritten) || rewritten.Size() != source.Length || source.Info.ModTime().Equal(rewritten.ModTime()) || !source.Info.ModTime().Equal(initial.ModTime()) {
		t.Fatal("same-size rewrite cannot be detected from pinned mtime")
	}

	// Appends keep identity; the captured prefix size and timestamp stay fixed.
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.WriteString("{}\n"); err != nil {
		t.Fatal(err)
	}
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
	appended, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if !os.SameFile(source.Info, appended) || appended.Size() <= source.Length || source.Info.Size() != initial.Size() {
		t.Fatal("append changed pinned descriptor / lost source identity")
	}
	// Replacement with equal bytes must not be mistaken for the pinned inode.
	replacement := usageFixture(t, root, "replacement.jsonl", usageMeta("session", "/repo", ""))
	if err := os.Rename(replacement, path); err != nil {
		t.Fatal(err)
	}
	replaced, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if os.SameFile(source.Info, replaced) || replaced.Size() != source.Length {
		t.Fatal("equal-sized replacement cannot be detected")
	}
}

func TestUsageDiscoveryClaudeQueueLifecycleBeforeActualMetadata(t *testing.T) {
	root := t.TempDir()
	usageFixture(t, root, "project/session.jsonl",
		`{"type":"queue-operation","operation":"enqueue","timestamp":"2026-10-14T00:00:00Z","sessionId":"queue-session"}`+"\n"+
			`{"type":"user","timestamp":"2026-10-14T00:01:00Z","sessionId":"actual-session","cwd":"/repo"}`+"\n"+
			`{"type":"assistant","timestamp":"2026-10-14T00:02:00Z","sessionId":"copied-session","cwd":"/other","message":{"id":"message","model":"model","usage":{"input_tokens":1,"output_tokens":2}}}`+"\n")
	got, err := DiscoverUsageEvidence(context.Background(), []Harness{ClaudeHarness(root)}, time.Time{})
	if err != nil || len(got.Sources) != 1 || len(got.Sessions) != 1 {
		t.Fatalf("queue lifecycle hid actual session: %+v %v", got, err)
	}
	s := got.Sources[0]
	if !s.MetadataComplete || s.LineageStatus != "verified" || s.ID != "actual-session" || s.CWD != "/repo" || s.Started.Minute() != 1 {
		t.Fatalf("first actual metadata not immutable: %+v", s)
	}
}

func TestUsageDiscoveryClaudeMalformedActualMetadataCannotBeRepaired(t *testing.T) {
	root := t.TempDir()
	usageFixture(t, root, "project/session.jsonl",
		`{"type":"queue-operation","operation":"enqueue","timestamp":"2026-10-14T00:00:00Z","sessionId":"session"}`+"\n"+
			`{"type":"user","timestamp":"2026-10-14T00:01:00Z","sessionId":"session"}`+"\n"+
			`{"type":"user","timestamp":"2026-10-14T00:02:00Z","sessionId":"session","cwd":"/repo"}`+"\n")
	got, err := DiscoverUsageEvidence(context.Background(), []Harness{ClaudeHarness(root)}, time.Time{})
	if err != nil || len(got.Sources) != 1 || len(got.Sessions) != 0 || got.Sources[0].MetadataComplete {
		t.Fatalf("malformed first actual metadata repaired by later record: %+v %v", got, err)
	}
}

func TestUsageDiscoveryScopesProviderFailures(t *testing.T) {
	qoderRoot, codexRoot := t.TempDir(), t.TempDir()
	if err := os.Symlink(filepath.Join(qoderRoot, "absent"), filepath.Join(qoderRoot, "unavailable.jsonl")); err != nil {
		t.Fatal(err)
	}
	usageFixture(t, qoderRoot, "malformed.jsonl", "not-json\n")
	usageFixture(t, codexRoot, "valid.jsonl", usageMeta("codex-valid", "/repo", ""))
	got, err := DiscoverUsageEvidence(context.Background(), []Harness{QoderHarness(qoderRoot), CodexHarness(codexRoot)}, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if got.Diagnostics["unavailable_source"] != 1 || got.Diagnostics["unknown_metadata"] != 1 {
		t.Fatalf("aggregate diagnostics=%v", got.Diagnostics)
	}
	if got.AgentDiagnostics["qoder"]["unavailable_source"] != 1 || got.AgentDiagnostics["qoder"]["unknown_metadata"] != 1 {
		t.Fatalf("Qoder gaps missing: %v", got.AgentDiagnostics)
	}
	if got.AgentDiagnostics["codex"]["unavailable_source"] != 0 || got.AgentDiagnostics["codex"]["unknown_metadata"] != 0 || len(got.GlobalDiagnostics) != 0 {
		t.Fatalf("provider gap leaked globally: agents=%v global=%v", got.AgentDiagnostics, got.GlobalDiagnostics)
	}
}
func TestUsageDiscoveryCandidateBoundScopesUnscannedProviders(t *testing.T) {
	claudeRoot, codexRoot, qoderRoot := t.TempDir(), t.TempDir(), t.TempDir()
	usageFixture(t, claudeRoot, "one.jsonl", "{}\n")
	usageFixture(t, codexRoot, "one.jsonl", usageMeta("one", "/repo", ""))
	usageFixture(t, codexRoot, "two.jsonl", usageMeta("two", "/repo", ""))
	qoderFixture(t, qoderRoot, qoderRows("/repo"))
	// Claude completes before the cap. Codex reaches it, and Qoder is never walked.
	got, err := discoverUsageEvidence(context.Background(), []Harness{ClaudeHarness(claudeRoot), CodexHarness(codexRoot), QoderHarness(qoderRoot)}, time.Time{}, 2)
	if err != nil {
		t.Fatal(err)
	}
	if got.Diagnostics["candidate_limit"] != 1 || got.AgentDiagnostics["codex"]["candidate_limit"] != 1 || got.AgentDiagnostics["qoder"]["candidate_limit"] != 1 {
		t.Fatalf("unscanned providers lost bound: aggregate=%v agents=%v", got.Diagnostics, got.AgentDiagnostics)
	}
	if got.AgentDiagnostics["claude"]["candidate_limit"] != 0 || got.GlobalDiagnostics["candidate_limit"] != 0 {
		t.Fatalf("completed provider contaminated: agents=%v global=%v", got.AgentDiagnostics, got.GlobalDiagnostics)
	}
}

func TestUsageDiscoveryUnsupportedHarnessRemainsGlobal(t *testing.T) {
	got, err := DiscoverUsageEvidence(context.Background(), []Harness{fakeHarness{name: "future-provider"}}, time.Time{})
	if err != nil || got.Diagnostics["unsupported_harness"] != 1 || got.GlobalDiagnostics["unsupported_harness"] != 1 || len(got.AgentDiagnostics) != 0 {
		t.Fatalf("unknown provider diagnostic vanished: %+v %v", got, err)
	}
}
