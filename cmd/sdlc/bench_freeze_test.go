package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/bench"
)

func TestRunBenchFreeze(t *testing.T) {
	root := t.TempDir()
	issuesDir := filepath.Join(root, "workshop", "issues")
	if err := os.MkdirAll(issuesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(issuesDir, "000119-demo.md"),
		[]byte("---\nid: 000119\nstatus: working\n---\n\n# Demo\n\n## Spec\n\nDo the thing.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	orig := headSHA
	headSHA = func() (string, error) { return "deadbeef", nil }
	defer func() { headSHA = orig }()

	var out, errBuf bytes.Buffer
	if err := runBenchFreeze(&out, &errBuf, benchFreezeFlags{Issue: 119, Repo: "ariadne", Root: root}); err != nil {
		t.Fatal(err)
	}
	got, err := bench.NewStore(filepath.Join(root, "workshop", "benchmarks")).ReadTask("119-demo")
	if err != nil {
		t.Fatal(err)
	}
	if got.BaseSHA != "deadbeef" {
		t.Errorf("base_sha = %q, want deadbeef", got.BaseSHA)
	}
	if !strings.Contains(got.Spec, "Do the thing") {
		t.Errorf("spec = %q", got.Spec)
	}
	if got.SourceIssue != "119" {
		t.Errorf("source_issue = %q", got.SourceIssue)
	}
}

func TestRunBenchFreezeRequiresIssue(t *testing.T) {
	var out, errBuf bytes.Buffer
	if err := runBenchFreeze(&out, &errBuf, benchFreezeFlags{}); err == nil {
		t.Fatal("expected error when --issue is 0")
	}
}

func TestRunBenchFreezeWarnsOnSubheading(t *testing.T) {
	root := t.TempDir()
	issuesDir := filepath.Join(root, "workshop", "issues")
	os.MkdirAll(issuesDir, 0o755)
	os.WriteFile(filepath.Join(issuesDir, "000200-sub.md"),
		[]byte("---\nid: 000200\n---\n\n# Sub\n\n## Spec\n\nintro\n\n## Inner\n\nbody\n"), 0o644)
	orig := headSHA
	headSHA = func() (string, error) { return "cafe", nil }
	defer func() { headSHA = orig }()
	var out, errBuf bytes.Buffer
	if err := runBenchFreeze(&out, &errBuf, benchFreezeFlags{Issue: 200, Root: root}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(errBuf.String(), "warning") {
		t.Errorf("expected a ## subheading warning, got stderr=%q", errBuf.String())
	}
}
