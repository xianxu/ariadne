package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/queuegate"
	"github.com/xianxu/ariadne/pkg/workspace"
)

type queueRuntimeKey struct{}
type queueAdmittedKey struct{}
type queueRuntime struct {
	store       *queuegate.Store
	lease       *queuegate.Lease
	op          queuegate.Operation
	mu          sync.Mutex
	retired     bool
	interrupted bool
}

func queueRuntimeFrom(ctx context.Context) *queueRuntime {
	q, _ := ctx.Value(queueRuntimeKey{}).(*queueRuntime)
	return q
}
func withQueueAdmitted(ctx context.Context) context.Context {
	return context.WithValue(ctx, queueAdmittedKey{}, true)
}
func isQueueAdmitted(ctx context.Context) bool {
	yes, _ := ctx.Value(queueAdmittedKey{}).(bool)
	return yes
}

func startQueueOperation(ctx context.Context, root, head string, kinds []string) (context.Context, *queueRuntime, error) {
	root = canonicalRoot(root)
	state, err := queueStateRoot()
	if err != nil {
		return ctx, nil, err
	}
	var nonce [16]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return ctx, nil, err
	}
	started, err := processStartTime(os.Getpid())
	if err != nil {
		return ctx, nil, fmt.Errorf("queue process identity: %w", err)
	}
	address := ""
	if identity, err := workspace.Resolve(queueGitReader{ctx}, root, ""); err == nil && identity.Address != nil {
		address = *identity.Address
	}
	boot, _ := queueBootID()
	op := queuegate.Operation{SchemaVersion: 1, ID: "op-" + hex.EncodeToString(nonce[:]), Workspace: address, Head: head, Kinds: kinds, Phase: "waiting", PID: os.Getpid(), PIDStarted: started, Managed: true, StartedAt: time.Now().UTC(), BootID: boot, Root: root}
	s := &queuegate.Store{Root: state, BootID: queueBootID}
	lease, err := s.Enroll(ctx, op)
	if err != nil {
		if lease != nil {
			_ = lease.Close()
		}
		return ctx, nil, err
	}
	q := &queueRuntime{store: s, lease: lease, op: op}
	return context.WithValue(ctx, queueRuntimeKey{}, q), q, nil
}

func (q *queueRuntime) admit(ctx context.Context, resource func() (bool, error)) (bool, error) {
	return q.lease.TryAdmit(ctx, resource)
}
func (q *queueRuntime) waitAdmission(ctx context.Context, stderr io.Writer) error {
	previous := ""
	for {
		ok, err := q.admit(ctx, nil)
		if ok && err == nil {
			return nil
		}
		var paused *queuegate.PausedError
		if !errors.As(err, &paused) {
			return err
		}
		if message := err.Error(); message != previous {
			fmt.Fprintln(stderr, message)
			previous = message
		}
		timer := time.NewTimer(200 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}

func (q *queueRuntime) finish(reason string) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.retired {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	interrupted, err := q.lease.RetireOutcome(ctx, reason)
	q.interrupted = interrupted
	if err != nil {
		_ = q.lease.Close()
	}
	q.retired = true
	return err
}

// The owner consumes its own durable request. Joining this reader precedes
// retirement, so no cancellation goroutine outlives the lease it observes.
func consumeQueueCancellation(ctx context.Context, q *queueRuntime, cancel context.CancelFunc) func() {
	watch, stop := context.WithCancel(ctx)
	done := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for {
			probe, end := context.WithTimeout(watch, time.Second)
			requested, err := q.lease.CancellationRequested(probe)
			end()
			if watch.Err() != nil {
				return
			}
			if err != nil || requested {
				cancel()
				return
			}
			select {
			case <-watch.Done():
				return
			case <-ticker.C:
			}
		}
	}()
	return func() { stop(); <-done }
}
