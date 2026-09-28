package gitx

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestTrunkContextPropagatesThroughTransaction(t *testing.T) {
	repo, _ := trunkFixture(t, "seed\n")
	type key struct{}
	ctx := context.WithValue(context.Background(), key{}, "operation")
	tf, err := NewTrunkFileContext(ctx, repo, "origin", "main")
	if err != nil {
		t.Fatal(err)
	}
	original := runGitInContext
	t.Cleanup(func() { runGitInContext = original })
	seen := map[string]bool{}
	runGitInContext = func(got context.Context, dir string, env []string, args ...string) ([]byte, []byte, error) {
		if got.Value(key{}) != "operation" {
			t.Errorf("%s lost operation context", args[0])
		}
		seen[args[0]] = true
		return original(got, dir, env, args...)
	}
	err = tf.Update("issue.md", "context transaction", func(old []byte) ([]byte, error) { return append(old, []byte("next\n")...), nil })
	if err != nil {
		t.Fatal(err)
	}
	if _, err = tf.Read("issue.md"); err != nil {
		t.Fatal(err)
	}
	for _, command := range []string{"fetch", "ls-tree", "cat-file", "read-tree", "hash-object", "update-index", "write-tree", "commit-tree", "push"} {
		if !seen[command] {
			t.Errorf("transaction did not exercise %s", command)
		}
	}
}

func TestTrunkContextCancellationCannotBecomeAbsence(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	tf, err := NewTrunkFileContext(ctx, t.TempDir(), "origin", "issue-tracker")
	if err != nil {
		t.Fatal(err)
	}
	_, err = tf.ViewOf(strings.Repeat("a", 40)).Exists("issue.md")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Exists cancellation = %v", err)
	}
	called := false
	err = tf.UpdateMany("cancelled", func(*TrunkView) (TrunkWrite, error) { called = true; return TrunkWrite{}, nil })
	if !errors.Is(err, context.Canceled) || called {
		t.Fatalf("UpdateMany error=%v, prepared=%v", err, called)
	}
}

func TestTrunkContextCancellationDuringPushIsUncertain(t *testing.T) {
	repo, _ := trunkFixture(t, "seed\n")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tf, err := NewTrunkFileContext(ctx, repo, "origin", "main")
	if err != nil {
		t.Fatal(err)
	}
	original := runGitInContext
	t.Cleanup(func() { runGitInContext = original })
	runGitInContext = func(ctx context.Context, dir string, env []string, args ...string) ([]byte, []byte, error) {
		if args[0] == "push" {
			cancel()
			return nil, nil, context.Canceled
		}
		return original(ctx, dir, env, args...)
	}
	err = tf.Update("issue.md", "cancel push", func(old []byte) ([]byte, error) { return []byte("new\n"), nil })
	if !errors.Is(err, ErrPublicationUncertain) {
		t.Fatalf("push cancellation = %v", err)
	}
}

func TestTrunkContextRequiresContext(t *testing.T) {
	if _, err := NewTrunkFileContext(nil, t.TempDir(), "origin", "main"); err == nil {
		t.Fatal("nil context accepted")
	}
}

func TestPreparedTransactionPersistsCandidateBeforePush(t *testing.T) {
	repo, origin := trunkFixture(t, "seed\n")
	tf, err := NewTrunkFileContext(context.Background(), repo, "origin", "main")
	if err != nil {
		t.Fatal(err)
	}
	before := commitCount(t, origin)
	blocked := errors.New("receipt persistence failed")
	var candidate string
	err = tf.UpdateManyPrepared("prepare receipt", func(*TrunkView) (TrunkWrite, error) {
		return TrunkWrite{Write: map[string][]byte{"receipt.md": []byte("payload")}}, nil
	}, func(base, next string) error {
		if base == next || len(base) != 40 || len(next) != 40 {
			t.Fatalf("invalid prepared identity %s %s", base, next)
		}
		candidate = next
		return blocked
	})
	if !errors.Is(err, blocked) || candidate == "" {
		t.Fatalf("candidate=%q error=%v", candidate, err)
	}
	if got := commitCount(t, origin); got != before {
		t.Fatalf("published before receipt: commits %d → %d", before, got)
	}
}
