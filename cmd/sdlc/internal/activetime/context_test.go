package activetime

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestComputeCanceledContextReturnsNoMeasurement(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	got, err := Compute(Options{Context: ctx, GitRepo: "/does-not-exist"})
	if !errors.Is(err, context.Canceled) || !reflect.DeepEqual(got, Result{}) {
		t.Fatalf("canceled computation returned result: %+v %v", got, err)
	}
}

// Deterministically cancel after reads have begun, without machine-speed timing.
type cancelAfterChecks struct {
	context.Context
	cancel     context.CancelFunc
	checks, at int
}

func (c *cancelAfterChecks) Err() error {
	c.checks++
	if c.checks == c.at {
		c.cancel()
	}
	return c.Context.Err()
}

func TestPinnedLargePrefixCancellationDuringRead(t *testing.T) {
	data := []byte(strings.Repeat("{\"type\":\"irrelevant\"}\n", 50000))
	path := filepath.Join(t.TempDir(), "session.jsonl")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	base, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx := &cancelAfterChecks{Context: base, cancel: cancel, at: 5}
	got, err := readSessionPrefixContext(ctx, path, []SourcePrefix{{Bytes: int64(len(data)), Digest: fmt.Sprintf("%x", sha256.Sum256(data))}})
	if !errors.Is(err, context.Canceled) || got != nil || ctx.checks < ctx.at {
		t.Fatalf("partial prefix yielded evidence: bytes=%d checks=%d err=%v", len(got), ctx.checks, err)
	}
}

func TestTranscriptCancellationIsNotSkippedAsUnreadable(t *testing.T) {
	path := filepath.Join(t.TempDir(), "session.jsonl")
	if err := os.WriteFile(path, []byte(strings.Repeat("{}\n", 50000)), 0644); err != nil {
		t.Fatal(err)
	}
	base, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx := &cancelAfterChecks{Context: base, cancel: cancel, at: 5}
	events, spans, err := loadEventsWithFilesContext(ctx, nil, []string{path}, newMentionScope("repo", nil), true, "", "", nil)
	if !errors.Is(err, context.Canceled) || events != nil || spans != nil {
		t.Fatalf("cancellation swallowed: %d %d %v", len(events), len(spans), err)
	}
}

func TestWindowCommitCancellationPropagates(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	orig := gitRunContext
	t.Cleanup(func() { gitRunContext = orig })
	gitRunContext = func(ctx context.Context, repo string, args ...string) ([]byte, error) {
		cancel()
		return nil, ctx.Err()
	}
	commits, err := loadWindowCommitsContext(ctx, "/repo", "", "", Scope{})
	if !errors.Is(err, context.Canceled) || commits != nil {
		t.Fatalf("canceled commit read: %+v %v", commits, err)
	}
}
