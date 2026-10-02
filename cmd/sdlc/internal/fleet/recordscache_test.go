package fleet

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

// #290: callers of one repository share one load and see its result; a
// caller never waits on another repository's load. The gate holds every load
// open until the test releases it, so overlap is proven by state, not timing.
func TestRecordsCacheIsPerKeyOnce(t *testing.T) {
	var loads sync.Map // root -> *int32
	gate := make(chan struct{})
	started := make(chan string, 8)
	c := newRecordsCache(func(_ context.Context, root string) (tracker.Records, error) {
		n, _ := loads.LoadOrStore(root, new(int32))
		atomic.AddInt32(n.(*int32), 1)
		started <- root
		<-gate
		return tracker.Records{Ref: root}, nil
	})
	var wg sync.WaitGroup
	results := make(chan string, 3)
	for _, root := range []string{"/a", "/a", "/b"} {
		wg.Add(1)
		go func(root string) {
			defer wg.Done()
			rs, err := c.get(context.Background(), root)
			if err != nil {
				t.Error(err)
			}
			results <- root + "=" + rs.Ref
		}(root)
	}
	// Both repositories' loads are in flight at once: /b did not wait for /a.
	seen := map[string]bool{<-started: true, <-started: true}
	if !seen["/a"] || !seen["/b"] {
		t.Fatalf("loads in flight: %v, want /a and /b concurrently", seen)
	}
	close(gate)
	wg.Wait()
	close(results)
	for r := range results {
		if r != "/a=/a" && r != "/b=/b" {
			t.Fatalf("a caller got another repository's records: %s", r)
		}
	}
	for _, root := range []string{"/a", "/b"} {
		if n, _ := loads.Load(root); atomic.LoadInt32(n.(*int32)) != 1 {
			t.Fatalf("%s loaded %d times, want once", root, *n.(*int32))
		}
	}
}

// #290: a load that ran out of time is kept, so the row walk does not wait for
// it again; any other failure is retried by the next caller.
func TestRecordsCacheKeepsTimeoutsRetriesOtherErrors(t *testing.T) {
	var calls int32
	c := newRecordsCache(func(_ context.Context, root string) (tracker.Records, error) {
		atomic.AddInt32(&calls, 1)
		if root == "/slow" {
			return tracker.Records{}, fmt.Errorf("read: %w", context.DeadlineExceeded)
		}
		return tracker.Records{}, errors.New("transient")
	})
	for i := 0; i < 2; i++ {
		_, _ = c.get(context.Background(), "/slow")
		_, _ = c.get(context.Background(), "/flaky")
	}
	if calls != 3 {
		t.Fatalf("loads: %d, want 1 (timeout kept) + 2 (error retried)", calls)
	}
}

// #290: the warm-up loads every root once with at most limit in flight.
func TestWarmRecordsIsBounded(t *testing.T) {
	const limit = 3
	var inFlight, peak int32
	release := make(chan struct{})
	entered := make(chan struct{}, 20)
	var mu sync.Mutex
	loaded := map[string]int{}
	get := func(_ context.Context, root string) (tracker.Records, error) {
		n := atomic.AddInt32(&inFlight, 1)
		for p := atomic.LoadInt32(&peak); n > p && !atomic.CompareAndSwapInt32(&peak, p, n); p = atomic.LoadInt32(&peak) {
		}
		entered <- struct{}{}
		<-release
		atomic.AddInt32(&inFlight, -1)
		mu.Lock()
		loaded[root]++
		mu.Unlock()
		return tracker.Records{}, nil
	}
	roots := []string{"/1", "/2", "/3", "/4", "/5", "/6", "/7"}
	done := make(chan struct{})
	go func() { warmRecords(context.Background(), roots, limit, get); close(done) }()
	for i := 0; i < limit; i++ {
		<-entered // the bound is reached before anything is released
	}
	select {
	case <-entered:
		t.Fatal("a load started beyond the bound")
	default:
	}
	close(release)
	<-done
	if peak != limit || len(loaded) != len(roots) {
		t.Fatalf("peak %d (want %d), loaded %v", peak, limit, loaded)
	}
	for r, n := range loaded {
		if n != 1 {
			t.Fatalf("%s loaded %d times", r, n)
		}
	}
}
