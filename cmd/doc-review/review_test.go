package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantAgent AgentCLI
		wantFile  string
		wantErr   bool
	}{
		{"file only defaults to codex", []string{"x.md"}, AgentCodex, "x.md", false},
		{"agent + file", []string{"gemini", "x.md"}, AgentGemini, "x.md", false},
		{"claude + file", []string{"claude", "a/b.md"}, AgentClaude, "a/b.md", false},
		{"unknown agent", []string{"bard", "x.md"}, "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, f, err := parseArgs(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got agent=%q file=%q", a, f)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if a != tt.wantAgent || f != tt.wantFile {
				t.Fatalf("got (%q,%q), want (%q,%q)", a, f, tt.wantAgent, tt.wantFile)
			}
		})
	}
}

func TestReportPath(t *testing.T) {
	tests := []struct {
		file  string
		agent AgentCLI
		want  string
	}{
		{"/a/b/tod.md", AgentCodex, "/a/b/tod-codex-check.md"},
		{"/a/b/x.md", AgentGemini, "/a/b/x-gemini-check.md"},
		{"notes.md", AgentClaude, "notes-claude-check.md"},
		{"/a/no-ext", AgentCodex, "/a/no-ext-codex-check.md"},
	}
	for _, tt := range tests {
		if got := reportPath(tt.file, tt.agent); got != tt.want {
			t.Errorf("reportPath(%q,%q)=%q, want %q", tt.file, tt.agent, got, tt.want)
		}
	}
}

func TestBuildArgs_ReadOnly(t *testing.T) {
	const prompt = "REVIEW THIS"

	// codex: read-only sandbox + captures via -o tmp file.
	name, args, fromFile, tmp := buildArgs(AgentCodex, prompt)
	if name != "codex" {
		t.Fatalf("codex name = %q", name)
	}
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "exec") || !strings.Contains(joined, "--sandbox read-only") {
		t.Errorf("codex args missing read-only sandbox: %v", args)
	}
	if !fromFile || tmp == "" {
		t.Errorf("codex should capture from a temp file; fromFile=%v tmp=%q", fromFile, tmp)
	}
	os.Remove(tmp)
	if args[len(args)-1] != prompt {
		t.Errorf("codex prompt must be last arg, got %q", args[len(args)-1])
	}

	// gemini: non-interactive -p (no --yolo ⇒ cannot write), stdout capture.
	name, args, fromFile, _ = buildArgs(AgentGemini, prompt)
	if name != "gemini" || fromFile {
		t.Errorf("gemini: name=%q fromFile=%v", name, fromFile)
	}
	if strings.Contains(strings.Join(args, " "), "--yolo") {
		t.Errorf("gemini must NOT get --yolo (that enables writes): %v", args)
	}

	// claude: read-only allowlist — must not grant Edit/Write/Bash.
	name, args, _, _ = buildArgs(AgentClaude, prompt)
	if name != "claude" {
		t.Fatalf("claude name = %q", name)
	}
	joined = strings.Join(args, " ")
	for _, banned := range []string{"Edit", "Write", "Bash"} {
		if strings.Contains(joined, banned) {
			t.Errorf("claude allowlist must not contain %q: %v", banned, args)
		}
	}
	if !strings.Contains(joined, "Read") {
		t.Errorf("claude allowlist should grant Read: %v", args)
	}
}

func TestBuildPrompt(t *testing.T) {
	got := buildPrompt("/abs/path/doc.md")
	if !strings.Contains(got, "/abs/path/doc.md") {
		t.Errorf("prompt should embed the file path")
	}
	if !strings.Contains(got, "DO NOT modify") {
		t.Errorf("prompt should forbid edits")
	}
	if !strings.Contains(strings.ToLower(got), "reference") {
		t.Errorf("prompt should ask about references")
	}
	if strings.Contains(got, "{{FILE}}") {
		t.Errorf("prompt placeholder not substituted")
	}
}

func TestWrapReport(t *testing.T) {
	got := wrapReport("tod.md", AgentCodex, "## Summary\nLooks fine.")
	if !strings.HasPrefix(got, "---\n") {
		t.Errorf("report should start with frontmatter")
	}
	if !strings.Contains(got, "reviews: tod.md") || !strings.Contains(got, "reviewer: codex") {
		t.Errorf("frontmatter should record doc + reviewer:\n%s", got)
	}
	if !strings.Contains(got, "## Summary") {
		t.Errorf("report body should be preserved")
	}
}

// TestRunReview_EndToEnd stubs the agent and verifies runReview writes a wrapped
// report to the computed path and emits the triage instruction.
func TestRunReview_EndToEnd(t *testing.T) {
	dir := t.TempDir()
	doc := filepath.Join(dir, "tod.md")
	if err := os.WriteFile(doc, []byte("# Doc\nclaim [1].\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	orig := runAgent
	defer func() { runAgent = orig }()
	var gotName string
	var gotArgs []string
	runAgent = func(_ context.Context, name string, args ...string) ([]byte, error) {
		gotName, gotArgs = name, args
		// codex mode captures from the -o temp file: write the report there.
		for i, a := range args {
			if a == "-o" && i+1 < len(args) {
				_ = os.WriteFile(args[i+1], []byte("## Summary\nVerdict: Supported."), 0o644)
			}
		}
		return []byte("ignored stdout"), nil
	}

	var out, errb bytes.Buffer
	f := &reviewFlags{Agent: AgentCodex, File: doc}
	if err := runReview(&out, &errb, f); err != nil {
		t.Fatalf("runReview: %v", err)
	}

	if gotName != "codex" {
		t.Errorf("dispatched %q, want codex", gotName)
	}
	if !strings.Contains(strings.Join(gotArgs, " "), "--sandbox read-only") {
		t.Errorf("codex not invoked read-only: %v", gotArgs)
	}

	want := filepath.Join(dir, "tod-codex-check.md")
	b, err := os.ReadFile(want)
	if err != nil {
		t.Fatalf("expected report at %s: %v", want, err)
	}
	if !strings.Contains(string(b), "Verdict: Supported") {
		t.Errorf("report missing reviewer body:\n%s", b)
	}
	if !strings.Contains(string(b), "reviewer: codex") {
		t.Errorf("report missing provenance frontmatter")
	}
	if !strings.Contains(out.String(), "triage") {
		t.Errorf("stdout should instruct the main agent to triage:\n%s", out.String())
	}

	// Source doc must be untouched.
	src, _ := os.ReadFile(doc)
	if string(src) != "# Doc\nclaim [1].\n" {
		t.Errorf("source document was modified: %q", src)
	}
}

func TestRunReview_EmptyReportFails(t *testing.T) {
	dir := t.TempDir()
	doc := filepath.Join(dir, "d.md")
	os.WriteFile(doc, []byte("x"), 0o644)

	orig := runAgent
	defer func() { runAgent = orig }()
	runAgent = func(_ context.Context, name string, args ...string) ([]byte, error) {
		return []byte(""), nil // launched fine, but produced nothing
	}

	var out, errb bytes.Buffer
	err := runReview(&out, &errb, &reviewFlags{Agent: AgentGemini, File: doc})
	if err == nil {
		t.Fatal("expected error on empty report")
	}
	if !strings.Contains(err.Error(), "empty report") {
		t.Errorf("error should mention empty report: %v", err)
	}
}

func TestRunReview_DryRun(t *testing.T) {
	dir := t.TempDir()
	doc := filepath.Join(dir, "d.md")
	os.WriteFile(doc, []byte("x"), 0o644)

	orig := runAgent
	defer func() { runAgent = orig }()
	called := false
	runAgent = func(_ context.Context, name string, args ...string) ([]byte, error) {
		called = true
		return nil, nil
	}

	var out, errb bytes.Buffer
	if err := runReview(&out, &errb, &reviewFlags{Agent: AgentCodex, File: doc, DryRun: true}); err != nil {
		t.Fatalf("dry-run: %v", err)
	}
	if called {
		t.Error("dry-run must not invoke the agent")
	}
	if !strings.Contains(out.String(), "d-codex-check.md") {
		t.Errorf("dry-run should print the report path:\n%s", out.String())
	}
}
