package testfix

import (
	"fmt"
	"testing"
)

func TestWatchSignalsOnNeedleAcrossWrites(t *testing.T) {
	w := NewWatch("lock held")
	fmt.Fprint(w, "waiting for sdlc repo ")
	fmt.Fprint(w, "lock ")
	select {
	case <-w.Seen():
		t.Fatal("signalled before the needle was complete")
	default:
	}
	fmt.Fprint(w, "held by pid 1\n")
	fmt.Fprint(w, "more\n") // a second match must not close twice
	select {
	case <-w.Seen():
	default:
		t.Fatal("needle split across writes was not seen")
	}
	if got := w.String(); got != "waiting for sdlc repo lock held by pid 1\nmore\n" {
		t.Errorf("output not kept: %q", got)
	}
}
