package queuegate

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flock"
)

// Lease owns one registry inode throughout its lifetime. Pause requests rewrite
// metadata under the global gate; readers never infer liveness from its PID.
type Lease struct {
	store *Store
	file  *os.File
	id    string
	mu    sync.Mutex
}

func (s *Store) Enroll(ctx context.Context, op Operation) (*Lease, error) {
	if !safeID(op.ID) || op.Phase != "waiting" {
		return nil, errors.New("queue enrollment needs a safe id and waiting phase")
	}
	data, err := EncodeOperation(op)
	if err != nil {
		return nil, err
	}
	var lease *Lease
	err = s.WithGate(ctx, func() error {
		for _, family := range []string{"operations", "unresolved", "cancel"} {
			rows, err := operationFiles(filepath.Join(s.Root, family))
			if err != nil {
				return err
			}
			if len(rows) >= MaxOperations {
				return errors.New("queue registry full; repair required")
			}
		}
		dir := filepath.Join(s.Root, "operations")
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return err
		}
		f, err := os.CreateTemp(dir, ".enroll-*")
		if err != nil {
			return err
		}
		ok := false
		defer func() {
			if !ok {
				f.Close()
				os.Remove(f.Name())
			}
		}()
		if err := flock.Exclusive(f, true); err != nil {
			return err
		}
		if _, err := f.Write(data); err != nil {
			return err
		}
		if err := f.Sync(); err != nil {
			return err
		}
		path := filepath.Join(dir, op.ID+".json")
		for _, family := range []string{"operations", "unresolved", "interrupted", "cancel"} {
			if _, err := os.Lstat(filepath.Join(s.Root, family, op.ID+".json")); !errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("queue operation id already exists/unreadable: %s", op.ID)
			}
		}
		if err := os.Rename(f.Name(), path); err != nil {
			return err
		}
		lease = &Lease{store: s, file: f, id: op.ID}
		ok = true
		return nil
	})
	return lease, err
}

func rewriteOperation(f *os.File, op Operation) error {
	b, err := EncodeOperation(op)
	if err != nil {
		return err
	}
	if err := f.Truncate(0); err != nil {
		return err
	}
	if _, err := f.WriteAt(b, 0); err != nil {
		return err
	}
	return f.Sync()
}

func (l *Lease) operation() (Operation, error) {
	if l.file == nil {
		return Operation{}, errors.New("queue lease already retired")
	}
	if _, err := l.file.Seek(0, io.SeekStart); err != nil {
		return Operation{}, err
	}
	b, err := io.ReadAll(io.LimitReader(l.file, MaxJSON+1))
	if err != nil {
		return Operation{}, err
	}
	op, err := DecodeOperation(b)
	if err == nil && op.ID != l.id {
		return Operation{}, errors.New("queue lease identity changed")
	}
	return op, err
}

// TryAdmit runs a nonblocking resource acquisition only while the pause set is
// known empty. A true result and active lease are one pause-serialized effect.
func (l *Lease) TryAdmit(ctx context.Context, resource func() (bool, error)) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	admitted := false
	err := l.store.WithGate(ctx, func() error {
		c, err := l.store.control()
		if err != nil {
			return fmt.Errorf("queue control unavailable; admission held: %w", err)
		}
		if len(c.PauseOwners) > 0 {
			return &PausedError{Owners: c.PauseOwners}
		}
		op, err := l.operation()
		if err != nil {
			return err
		}
		if op.Phase != "waiting" {
			return fmt.Errorf("queue operation %s is %s, not waiting", op.ID, op.Phase)
		}
		if resource != nil {
			ok, err := resource()
			if err != nil {
				return err
			}
			if !ok {
				return nil
			}
		}
		op, err = TransitionOperation(op, "admit")
		if err != nil {
			return err
		}
		if err := rewriteOperation(l.file, op); err != nil {
			return err
		}
		admitted = true
		return nil
	})
	return admitted, err
}

func (l *Lease) CancellationRequested(ctx context.Context) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	var requested bool
	err := l.store.WithGate(ctx, func() error {
		op, err := l.operation()
		if err != nil {
			return err
		}
		_, present, err := l.store.readIntent(op.ID)
		if err != nil {
			return fmt.Errorf("queue cancellation evidence unavailable: %w", err)
		}
		requested = present || op.Phase == "interrupting"
		if present && op.Phase == "running" {
			op, err = TransitionOperation(op, "cancel")
			if err != nil {
				return err
			}
			return rewriteOperation(l.file, op)
		}
		return nil
	})
	return requested, err
}

func (l *Lease) Retire(ctx context.Context, reason string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file == nil {
		return nil
	}
	return l.store.WithGate(ctx, func() error {
		op, err := l.operation()
		if err != nil {
			return err
		}
		intentReason, intent, intentErr := l.store.readIntent(op.ID)
		if intentErr != nil {
			return intentErr
		}
		if reason == "" && intent {
			reason = intentReason
		}
		if op.Phase == "interrupting" || reason != "" || intent {
			if err := l.store.retainInterruption(ctx, op, reason); err != nil {
				return err
			}
			if !op.Managed && op.Phase != "waiting" {
				op, err = TransitionOperation(op, "unknown")
				if err != nil {
					return err
				}
				op.Detail = "unmanaged operation " + op.ID + ": " + reason + "; descendants require operator reconciliation"
				if err := l.store.retainUnknown(ctx, op); err != nil {
					return err
				}
			}
		}
		if err := os.Remove(filepath.Join(l.store.Root, "operations", l.id+".json")); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		_ = os.Remove(filepath.Join(l.store.Root, "cancel", l.id+".json"))
		err = l.file.Close()
		l.file = nil
		return err
	})
}

// Close only releases the lifetime lock. If a caller did not prove completion
// through Retire, a later snapshot must conservatively reconcile the residue.
func (l *Lease) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file == nil {
		return nil
	}
	err := l.file.Close()
	l.file = nil
	return err
}
