package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
)

func TestBoundaryReviewPreparationAndInvalidEvidence(t *testing.T) {
	p := boundaryWindow(t)
	p.PlansDir = t.TempDir()
	os.MkdirAll(".sdlc", 0755)
	if err := os.WriteFile(".sdlc/review-prepare.json", []byte(`{"version":1,"command":["sh","-c","echo generated > generated"]}`), 0644); err != nil {
		t.Fatal(err)
	}
	git(t, "", "add", ".sdlc/review-prepare.json")
	git(t, "", "commit", "-qm", "declare prep")
	p.Head = strings.TrimSpace(captureGit(t, "rev-parse", "HEAD"))
	old := judge.Run
	t.Cleanup(func() { judge.Run = old })
	calls := 0
	judge.Run = func(ctx context.Context, onStart func(int), name string, args ...string) (judge.ProcessOutput, error) {
		_, _, dir := unwrapReviewArgv(t, name, args)
		if _, err := os.Stat(filepath.Join(dir, "generated")); err != nil {
			t.Error("asset missing before reviewer", err)
		}
		calls++
		return judge.ProcessOutput{Stdout: []byte("VERDICT: REWORK\n```findings\n{}\n```\n")}, nil
	}
	r := dispatchBoundaryReview(&bytes.Buffer{}, &bytes.Buffer{}, p)
	if calls != 2 || r.DispatchError == nil || r.Round != nil || r.Output != "" {
		t.Fatalf("calls=%d result=%+v", calls, r)
	}
	data, err := os.ReadFile(r.SidecarPath)
	if err != nil || !strings.Contains(string(data), "Preparation attempt 2") || !strings.Contains(string(data), "Protocol attempt 2") {
		t.Fatalf("diagnostics: %v %s", err, data)
	}
	if _, err := os.Stat("generated"); !os.IsNotExist(err) {
		t.Fatalf("preparation escaped disposable checkout: %v", err)
	}
}
