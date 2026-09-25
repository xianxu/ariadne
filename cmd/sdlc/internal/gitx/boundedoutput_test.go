package gitx

import (
	"context"
	"errors"
	"testing"
)

func TestBoundedOutputCancelsAndNeverExceedsBudget(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	w := &boundedOutput{limit: 3, cancel: cancel}
	if n, err := w.Write([]byte("ab")); n != 2 || err != nil {
		t.Fatalf("first write: %d, %v", n, err)
	}
	if _, err := w.Write([]byte("cdef")); !errors.Is(err, ErrOutputLimit) {
		t.Fatalf("overflow: %v", err)
	}
	if ctx.Err() == nil || w.Len() > 3 {
		t.Fatalf("cancel=%v size=%d", ctx.Err(), w.Len())
	}
}

func TestBoundedGitOutput(t *testing.T) {
	repo, _ := trunkFixture(t, "seed\n")
	_, _, err := runGitBoundedInputContext(context.Background(), repo, nil, nil, 1, "rev-parse", "HEAD")
	if !errors.Is(err, ErrOutputLimit) {
		t.Fatalf("expected output budget error, got %v", err)
	}
}
