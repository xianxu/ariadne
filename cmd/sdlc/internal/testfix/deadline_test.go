package testfix

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDeadlineOnCue(t *testing.T) {
	ctx, expire := DeadlineOnCue()
	derived, cancel := context.WithTimeout(ctx, time.Hour)
	defer cancel()
	select {
	case <-derived.Done():
		t.Fatal("expired before its cue")
	default:
	}
	if ctx.Err() != nil || derived.Err() != nil {
		t.Fatalf("live before its cue: %v %v", ctx.Err(), derived.Err())
	}
	expire()
	expire() // idempotent
	select {
	case <-derived.Done():
	case <-time.After(HangCeiling):
		t.Fatal("a derived context never saw the cue")
	}
	if !errors.Is(derived.Err(), context.DeadlineExceeded) || !errors.Is(ctx.Err(), context.DeadlineExceeded) {
		t.Fatalf("want DeadlineExceeded, got %v / %v", derived.Err(), ctx.Err())
	}
}
