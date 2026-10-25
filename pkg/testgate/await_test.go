package testgate

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"testing"
	"time"
)

// fatalTB records the first Fatalf and stops its goroutine, as testing does.
type fatalTB struct {
	testing.TB
	msg string
}

func (f *fatalTB) Helper() {}
func (f *fatalTB) Fatalf(format string, args ...any) {
	f.msg = fmt.Sprintf(format, args...)
	runtime.Goexit()
}

// #622: a worker that ends before the condition fails the wait at once, and
// its result stays on the channel for a later reader.
func TestAwaitConditionFailsFastAndKeepsTheResult(t *testing.T) {
	worker := make(chan error, 1)
	worker <- errors.New("landed early")
	tb := &fatalTB{TB: t}
	done := make(chan struct{})
	go func() {
		defer close(done)
		AwaitCondition(tb, "queued", func() bool { return false }, []chan error{worker})
	}()
	select {
	case <-done:
	case <-time.After(HangCeiling):
		t.Fatal("AwaitCondition did not fail on an ended worker")
	}
	if !strings.Contains(tb.msg, "queued: worker 0 ended first: landed early") {
		t.Fatalf("message %q", tb.msg)
	}
	select {
	case err := <-worker:
		if err == nil || err.Error() != "landed early" {
			t.Fatalf("result changed: %v", err)
		}
	default:
		t.Fatal("the worker's result was consumed")
	}
}

func TestAwaitConditionReturnsWhenTheConditionHolds(t *testing.T) {
	polls := 0
	worker := make(chan error, 1) // still running
	AwaitCondition(t, "third poll", func() bool { polls++; return polls == 3 }, []chan error{worker})
	if polls != 3 {
		t.Fatalf("polls = %d", polls)
	}
}

func TestAwaitConditionRefusesUnbufferedOrClosedWorkers(t *testing.T) {
	closed := make(chan error, 1)
	close(closed)
	for name, w := range map[string]chan error{"unbuffered": make(chan error), "closed": closed} {
		tb := &fatalTB{TB: t}
		done := make(chan struct{})
		go func() {
			defer close(done)
			AwaitCondition(tb, "x", func() bool { return false }, []chan error{w})
		}()
		select {
		case <-done:
		case <-time.After(HangCeiling):
			t.Fatalf("%s: AwaitCondition hung", name)
		}
		if !strings.Contains(tb.msg, name) {
			t.Fatalf("%s: message %q", name, tb.msg)
		}
	}
}

func TestReturnsWaitsForTheCall(t *testing.T) {
	ran := false
	Returns(t, "the call", func() { ran = true })
	if !ran {
		t.Fatal("Returns came back before the call finished")
	}
}
