package main

import (
	"errors"
	"fmt"
	"sort"
)

type resumeBatchPhase uint8

const (
	batchPending resumeBatchPhase = iota
	batchRemaining
	batchStopped
	batchComplete
)

type resumeBatchEventKind uint8

const (
	batchAdmit resumeBatchEventKind = iota
	batchResult
	batchCancel
)

type resumeBatchEvent struct {
	kind resumeBatchEventKind
	addr string
	err  error
}

type resumeLaunch struct {
	Addr string
	Wave int
}

type resumeBatchResult struct {
	Addr string
	Err  error
}

type resumeFailureKey struct{ Status, Code, Diagnostic string }
type codedResumeFailure interface {
	error
	failureKey() (resumeFailureKey, bool)
}

// ResumeBatch owns dispatch state. Only advanceResumeBatch changes it after
// construction. IO executes each declared launch and returns its result event;
// it never changes the cursor, phase, active set or failure counts itself.
type ResumeBatch struct {
	phase      resumeBatchPhase
	limit      int
	todo       []resumeLaunch
	next       int
	pendingEnd int
	active     map[string]int
	results    []resumeBatchResult
	failures   map[resumeFailureKey]int
	stopCause  error
}

func newResumeBatch(rows []resumeRow, limit int) (*ResumeBatch, error) {
	if limit <= 0 {
		return nil, fmt.Errorf("concurrency must be a positive integer")
	}
	b := &ResumeBatch{limit: limit, active: map[string]int{}, failures: map[resumeFailureKey]int{}}
	seen := map[string]bool{}
	for _, row := range rows {
		if row.Class != classResume {
			continue
		}
		if row.Addr == "" || seen[row.Addr] {
			return nil, fmt.Errorf("duplicate or empty resume address %q", row.Addr)
		}
		seen[row.Addr] = true
		wave := 2
		if row.Pending {
			wave = 1
			b.pendingEnd++
		}
		b.todo = append(b.todo, resumeLaunch{Addr: row.Addr, Wave: wave})
	}
	sort.Slice(b.todo, func(i, j int) bool {
		if b.todo[i].Wave != b.todo[j].Wave {
			return b.todo[i].Wave < b.todo[j].Wave
		}
		return b.todo[i].Addr < b.todo[j].Addr
	})
	if len(b.todo) == 0 {
		b.phase = batchComplete
	}
	return b, nil
}

// Result events never launch work. The coordinator first drains ready results
// and observes cancellation, then requests at most one new launch per Admit.
func advanceResumeBatch(b *ResumeBatch, e resumeBatchEvent) []resumeLaunch {
	if b.phase == batchComplete {
		return nil
	}
	switch e.kind {
	case batchCancel:
		if b.phase != batchStopped {
			b.phase, b.stopCause = batchStopped, e.err
		}
	case batchResult:
		if _, active := b.active[e.addr]; !active {
			return nil
		} // duplicate/foreign result
		delete(b.active, e.addr)
		b.results = append(b.results, resumeBatchResult{Addr: e.addr, Err: e.err})
		if e.err != nil && b.phase != batchStopped {
			var coded codedResumeFailure
			var key resumeFailureKey
			stable := false
			if errors.As(e.err, &coded) {
				key, stable = coded.failureKey()
			}
			if stable {
				b.failures[key]++
			}
			if !stable || b.failures[key] >= 2 {
				b.phase, b.stopCause = batchStopped, e.err
			}
		}
	case batchAdmit:
		if b.phase == batchStopped || len(b.active) >= b.limit {
			return nil
		}
	}
	if b.phase == batchStopped {
		return nil
	}
	if b.phase == batchPending && b.next == b.pendingEnd && len(b.active) == 0 {
		b.phase = batchRemaining
	}
	if b.phase == batchRemaining && b.next == len(b.todo) && len(b.active) == 0 {
		b.phase = batchComplete
		return nil
	}
	if e.kind != batchAdmit || b.next == len(b.todo) || (b.phase == batchPending && b.next == b.pendingEnd) {
		return nil
	}
	launch := b.todo[b.next]
	b.next++
	b.active[launch.Addr] = launch.Wave
	return []resumeLaunch{launch}
}

func (b *ResumeBatch) finished() bool {
	return b.phase == batchComplete || (b.phase == batchStopped && len(b.active) == 0)
}

func (b *ResumeBatch) stopped() bool { return b.phase == batchStopped }

func (b *ResumeBatch) untouched() []string {
	addresses := make([]string, 0, len(b.todo)-b.next)
	for _, launch := range b.todo[b.next:] {
		addresses = append(addresses, launch.Addr)
	}
	return addresses
}

func (b *ResumeBatch) resultError() error {
	var failures []error
	for _, result := range b.results {
		if result.Err != nil {
			failures = append(failures, fmt.Errorf("%s: %w", result.Addr, result.Err))
		}
	}
	joined := errors.Join(failures...)
	if b.stopCause != nil && !errors.Is(joined, b.stopCause) {
		joined = errors.Join(joined, b.stopCause)
	}
	return joined
}
