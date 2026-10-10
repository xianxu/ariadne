package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
)

func judgeStream(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("internal", "judge", "testdata", "stream", name))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// #300 end to end: a close whose reviewer wrote its verdict and findings and
// then a block-less postscript records that verdict, not "unknown", and its
// sidecar keeps the findings.
func TestCloseRecordsTheVerdictBeforeAPostscript(t *testing.T) {
	stream := judgeStream(t, "postscript.jsonl") // before closeReady changes directory
	r, _, detailPath := closeReady(t, 340)
	calls, _ := stubJudgeSeq(t, stream)
	_, stderr, err := executeSDLCTestCommand("close", "--issue", "340", "--verified", "e2e", "--actual", "1", "--no-atlas", "--no-ledger", "--no-project")
	if err != nil {
		t.Fatalf("close: %v\n%s", err, stderr)
	}
	if *calls != 1 || !strings.Contains(stderr, "FIX-THEN-SHIP") {
		t.Fatalf("%d dispatches; want the earlier verdict:\n%s", *calls, stderr)
	}
	sidecar := readSidecar(t, r.root, detailPath)
	if !strings.Contains(sidecar, "FIX-THEN-SHIP") || !strings.Contains(sidecar, "```findings") {
		t.Fatalf("sidecar lost the verdict or findings:\n%s", sidecar)
	}
}

// A reviewer that never gives a verdict is dispatched twice; the close keeps
// its fail-safe (does not finalize) and the sidecar holds both runs.
func TestCloseWithoutAVerdictKeepsBothRuns(t *testing.T) {
	stream := judgeStream(t, "background_wait.jsonl")
	r, cardPath, detailPath := closeReady(t, 341)
	calls, _ := stubJudgeSeq(t, stream)
	_, stderr, err := executeSDLCTestCommand("close", "--issue", "341", "--verified", "e2e", "--actual", "1", "--no-atlas", "--no-ledger", "--no-project")
	if err == nil {
		t.Fatalf("a close without a verdict finalized:\n%s", stderr)
	}
	if *calls != 2 || strings.Contains(r.card(cardPath), "status: codecomplete") {
		t.Fatalf("%d dispatches; card:\n%s", *calls, r.card(cardPath))
	}
	if sidecar := readSidecar(t, r.root, detailPath); strings.Count(sidecar, "I'll wait") != 2 || !strings.Contains(sidecar, "## Attempt 2") {
		t.Fatalf("sidecar does not hold both runs:\n%s", sidecar)
	}
}

func readSidecar(t *testing.T, root, detailPath string) string {
	t.Helper()
	path := sidecarPath(filepath.Join(root, "workshop", "plans"), filepath.Base(detailPath), "")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("sidecar %s: %v", path, err)
	}
	return string(b)
}

// D3 end to end: a close over a large window gives its reviewer the sized
// limit; WF_REVIEW_TIMEOUT still overrides it.
func TestCloseSizesTheReviewTimeout(t *testing.T) {
	for _, c := range []struct {
		env      string
		min, max time.Duration
	}{
		{"", 55 * time.Minute, 61 * time.Minute}, // ~2,000 added lines → 60m
		{"10m", 0, 10 * time.Minute},
	} {
		t.Run("env="+c.env, func(t *testing.T) {
			t.Setenv("WF_REVIEW_TIMEOUT", c.env)
			r, _, _ := closeReady(t, 342)
			writeRepoFile(t, r.root, "cmd/big.go", "package big\n"+strings.Repeat("// line\n", 2000))
			r.git("add", "cmd/big.go")
			r.git("commit", "-qm", "#342: a large change")
			orig := judge.Run
			t.Cleanup(func() { judge.Run = orig })
			var left time.Duration
			judge.Run = func(ctx context.Context, onStart func(int), name string, args ...string) (judge.ProcessOutput, error) {
				if d, ok := ctx.Deadline(); ok {
					left = time.Until(d)
				}
				return judge.ProcessOutput{Stdout: []byte("VERDICT: SHIP (confidence: high)\n")}, nil
			}
			_, stderr, _ := executeSDLCTestCommand("close", "--issue", "342", "--verified", "e2e", "--actual", "1", "--no-atlas", "--no-ledger", "--no-project")
			if left <= c.min || left > c.max {
				t.Fatalf("the reviewer had %v, want (%v, %v]:\n%s", left, c.min, c.max, stderr)
			}
		})
	}
}
