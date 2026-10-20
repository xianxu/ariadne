package queuegate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flock"
)

// Store is a portable machine queue backend. Nothing here discovers processes
// or topology: admission must never pay for the status-only legacy overlay.
type Store struct {
	Root string
	Now  func() time.Time
	// BootID is status-only evidence for retiring unresolved work from a prior
	// machine boot. An unavailable identity preserves uncertainty.
	BootID func() (string, error)
	// onLeaseRead measures actual registry reads for the counted complexity
	// guard; nil in production. It does not substitute observations.
	onLeaseRead func(string)
}

type PausedError struct{ Owners []PauseOwner }

func (e *PausedError) Error() string {
	var reasons []string
	for _, o := range e.Owners {
		reasons = append(reasons, fmt.Sprintf("%s: %q", o.Owner, o.Reason))
	}
	return "queue paused: " + strings.Join(reasons, "; ")
}

func (s *Store) now() time.Time {
	if s.Now != nil {
		return s.Now().UTC()
	}
	return time.Now().UTC()
}

// WithGate is the one ordering boundary shared with the Python lock helper.
// The caller can take an existing queue allocation lock inside fn, never in
// the reverse order. Context cancellation bounds a busy control mutex.
func (s *Store) WithGate(ctx context.Context, fn func() error) error {
	if s.Root == "" || !filepath.IsAbs(s.Root) {
		return errors.New("queue state root must be an absolute path")
	}
	if err := os.MkdirAll(s.Root, 0o700); err != nil {
		return err
	}
	f, err := os.OpenFile(filepath.Join(s.Root, "control.lock"), os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		err = flock.Exclusive(f, true)
		if err == nil {
			if err := fn(); err != nil {
				return err
			}
			return ctx.Err()
		}
		if !errors.Is(err, syscall.EWOULDBLOCK) {
			return err
		}
		timer := time.NewTimer(10 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}

func readBounded(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, MaxJSON+1))
	if len(b) > MaxJSON {
		return nil, errors.New("queue record exceeds 1 MiB")
	}
	return b, err
}

func (s *Store) control() (Control, error) {
	b, err := readBounded(filepath.Join(s.Root, "pause.json"))
	if errors.Is(err, os.ErrNotExist) {
		return NewControl(), nil
	}
	if err != nil {
		return Control{}, err
	}
	return DecodeControl(b)
}

func atomicWrite(path string, data []byte) error {
	if len(data) > MaxJSON {
		return errors.New("queue record exceeds 1 MiB")
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, ".queue-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	defer f.Close()
	if _, err := f.Write(data); err != nil {
		return err
	}
	if err := f.Sync(); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := os.Rename(f.Name(), path); err != nil {
		return err
	}
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	return d.Sync()
}

func (s *Store) mutate(ctx context.Context, action Action) error {
	return s.WithGate(ctx, func() error {
		c, err := s.control()
		if err != nil {
			return fmt.Errorf("queue control unavailable; admission remains held: %w", err)
		}
		next, err := Transition(c, action)
		if err != nil {
			return err
		}
		data, err := EncodeControl(next)
		if err != nil {
			return err
		}
		if err := atomicWrite(filepath.Join(s.Root, "pause.json"), data); err != nil {
			return err
		}
		if action.Kind != "pause" || !action.CancelRunning {
			return nil
		}
		return s.requestInterruptions(ctx, action.Reason)
	})
}

func (s *Store) Pause(ctx context.Context, owner, reason string, fast bool) error {
	return s.mutate(ctx, Action{Kind: "pause", Owner: owner, Reason: reason, CancelRunning: fast})
}

func (s *Store) Resume(ctx context.Context, owner string) error {
	return s.mutate(ctx, Action{Kind: "resume", Owner: owner})
}

// operationFiles bounds directory enumeration as well as record decoding.
func operationFiles(dir string) ([]os.DirEntry, error) {
	f, err := os.Open(dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	rows, err := f.ReadDir(MaxOperations + 1)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}
	if len(rows) > MaxOperations {
		return nil, errors.New("queue registry exceeds operation limit")
	}
	return rows, nil
}

func safeID(id string) bool {
	if id == "" || len(id) > 128 {
		return false
	}
	for _, c := range id {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-' || c == '_') {
			return false
		}
	}
	return true
}

func (s *Store) intentPath(id string) string { return filepath.Join(s.Root, "cancel", id+".json") }

func (s *Store) requestInterruptions(ctx context.Context, reason string) error {
	rows, err := operationFiles(filepath.Join(s.Root, "operations"))
	if err != nil {
		return fmt.Errorf("pause applied; cancellation scan incomplete: %w", err)
	}
	intents, err := operationFiles(filepath.Join(s.Root, "cancel"))
	if err != nil {
		return fmt.Errorf("pause applied; cancellation registry unavailable: %w", err)
	}
	intentCount := len(intents)
	var errs []error
	for _, row := range rows {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("pause applied; interruption publication incomplete: %w", err)
		}
		if row.IsDir() || !strings.HasSuffix(row.Name(), ".json") {
			continue
		}
		path := filepath.Join(s.Root, "operations", row.Name())
		if s.onLeaseRead != nil {
			s.onLeaseRead(path)
		}
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("pause applied; interruption publication incomplete: %w", err)
		}
		data, err := readBounded(path)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		op, err := DecodeOperation(data)
		if err != nil || !safeID(op.ID) || row.Name() != op.ID+".json" {
			errs = append(errs, fmt.Errorf("unreadable operation %s: cancellation not authorized", row.Name()))
			continue
		}
		if op.Phase != "running" && op.Phase != "interrupting" {
			continue
		}
		isSuite := false
		for _, kind := range op.Kinds {
			isSuite = isSuite || kind == "suite"
		}
		if !isSuite {
			continue
		} // Fast pause cancels suites, not an entire merge.
		f, err := os.OpenFile(path, os.O_RDWR, 0)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		lockErr := flock.Exclusive(f, true)
		if !errors.Is(lockErr, syscall.EWOULDBLOCK) {
			f.Close()
			continue // No lifetime holder: never ask a recycled process to stop.
		}
		_, present, intentErr := s.readIntent(op.ID)
		err = intentErr
		if err == nil && !present {
			if intentCount >= MaxOperations {
				f.Close()
				return errors.New("pause applied; cancellation registry full; repair required")
			}
			if err := ctx.Err(); err != nil {
				f.Close()
				return err
			}
			intent, _ := json.Marshal(map[string]any{"schema_version": 1, "id": op.ID, "reason": reason})
			err = atomicWrite(s.intentPath(op.ID), intent)
			if err == nil {
				intentCount++
			}
		}
		if err == nil {
			op, err = TransitionOperation(op, "cancel")
			if err == nil {
				err = rewriteOperation(f, op)
			}
		}
		f.Close()
		if err != nil {
			errs = append(errs, err)
		}
	}
	if err := errors.Join(errs...); err != nil {
		return fmt.Errorf("pause applied; cancellation incomplete: %w", err)
	}
	return nil
}

func unknownOperation(identity, detail string) Operation {
	sum := sha256.Sum256([]byte(identity))
	return Operation{SchemaVersion: 1, ID: "unreadable-" + hex.EncodeToString(sum[:16]), Kinds: []string{"suite"}, Phase: "unknown", Detail: detail, StartedAt: time.Unix(1, 0).UTC()}
}

// Snapshot is the new-registry observation only. The CLI adds a read-only
// legacy overlay. Stale active leases become persistent unknown evidence;
// removing a file whose lock died must not fabricate descendant completion.
func (s *Store) Snapshot(ctx context.Context) (Snapshot, error) {
	result, _, err := s.Observe(ctx)
	return result, err
}

// Observe also exposes enrolled bindings to the CLI's legacy overlay, so a
// newly registered waiter is never mistaken for an active pre-gate merge.
func (s *Store) Observe(ctx context.Context) (Snapshot, []Operation, error) {
	var result Snapshot
	var observed []Operation
	err := s.WithGate(ctx, func() error {
		var details []string
		// These names are created only while holding control.lock and before
		// publication. No admitted operation can depend on an unlocked temp.
		for _, family := range []string{"", "operations", "cancel", "unresolved", "interrupted"} {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := s.sweepUnpublished(ctx, filepath.Join(s.Root, family)); err != nil {
				details = append(details, "unpublished queue residue: "+err.Error())
			}
		}
		control, err := s.control()
		if err != nil {
			control = NewControl()
			details = append(details, "pause owners unavailable: "+err.Error())
		}
		var ops []Operation
		for _, family := range []string{"unresolved", "operations"} {
			if err := ctx.Err(); err != nil {
				return err
			}
			rows, scanErr := operationFiles(filepath.Join(s.Root, family))
			if scanErr != nil {
				details = append(details, family+": "+scanErr.Error())
				continue
			}
			for _, row := range rows {
				if err := ctx.Err(); err != nil {
					return err
				}
				if row.IsDir() || !strings.HasSuffix(row.Name(), ".json") {
					details = append(details, "unexpected queue registry entry "+row.Name())
					continue
				}
				path := filepath.Join(s.Root, family, row.Name())
				f, openErr := os.OpenFile(path, os.O_RDWR, 0)
				if errors.Is(openErr, os.ErrNotExist) {
					continue
				}
				if openErr != nil {
					details = append(details, "unreadable queue lease "+row.Name())
					continue
				}
				lockErr := flock.Exclusive(f, true)
				held := errors.Is(lockErr, syscall.EWOULDBLOCK)
				if s.onLeaseRead != nil {
					s.onLeaseRead(path)
				}
				if err := ctx.Err(); err != nil {
					f.Close()
					return err
				}
				data, readErr := io.ReadAll(io.LimitReader(f, MaxJSON+1))
				op, decodeErr := DecodeOperation(data)
				if readErr != nil || decodeErr != nil || len(data) > MaxJSON || lockErr != nil && !held || row.Name() != op.ID+".json" {
					ops = append(ops, unknownOperation("unreadable-"+row.Name(), "queue lease "+row.Name()+" is unreadable; owner lifetime unknown"))
					f.Close()
					continue
				}
				reason, intent, intentErr := s.readIntent(op.ID)
				if intentErr != nil {
					details = append(details, "cancellation evidence "+op.ID+" unavailable: "+intentErr.Error())
				} else if intent && op.Phase == "running" {
					op, err = TransitionOperation(op, "cancel")
					if err != nil {
						f.Close()
						return err
					}
					if held {
						if err := rewriteOperation(f, op); err != nil {
							details = append(details, "intent applied; phase repair unavailable: "+err.Error())
						}
					}
				}
				interrupted := intentErr == nil && (intent || op.Phase == "interrupting")
				if !held && family == "operations" {
					if (op.Phase == "waiting" || op.Phase == "finished" || op.Phase == "interrupted") && !intent {
						_ = os.Remove(path)
						_ = os.Remove(filepath.Join(s.Root, "cancel", op.ID+".json"))
						f.Close()
						continue
					}
					if interrupted {
						if err := ctx.Err(); err != nil {
							f.Close()
							return err
						}
						if err := s.retainInterruption(ctx, op, reason); err != nil {
							details = append(details, err.Error())
							f.Close()
							ops = append(ops, op)
							continue
						}
					}
					op, err = TransitionOperation(op, "unknown")
					if err != nil {
						f.Close()
						details = append(details, err.Error())
						continue
					}
					op.Detail = "operation " + op.ID + " lost its lifetime lease; descendants require reconciliation"
					if err := s.retainUnknown(ctx, op); err != nil {
						details = append(details, err.Error())
					} else {
						_ = os.Remove(path)
						if interrupted {
							_ = os.Remove(s.intentPath(op.ID))
						}
					}
				}
				if family == "unresolved" && interrupted {
					if err := ctx.Err(); err != nil {
						f.Close()
						return err
					}
					if err := s.retainInterruption(ctx, op, reason); err != nil {
						details = append(details, err.Error())
					} else {
						_ = os.Remove(s.intentPath(op.ID))
					}
				}
				if family == "unresolved" && s.BootID != nil && op.BootID != "" {
					boot, bootErr := s.BootID()
					if bootErr == nil && boot != "" && boot != op.BootID {
						_ = os.Remove(path)
						_ = os.Remove(s.intentPath(op.ID))
						f.Close()
						continue
					}
				}
				f.Close()
				ops = append(ops, op)
			}
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		orphans, orphanDetails, orphanErr := s.reconcileOrphanIntents(ctx)
		if orphanErr != nil {
			return orphanErr
		}
		ops = append(ops, orphans...)
		details = append(details, orphanDetails...)
		result = BuildSnapshot(control, ops, details)
		observed = ops
		return nil
	})
	return result, observed, err
}

func (s *Store) sweepUnpublished(ctx context.Context, dir string) error {
	rows, err := operationFiles(dir)
	if err != nil {
		return err
	}
	for _, row := range rows {
		if err := ctx.Err(); err != nil {
			return err
		}
		if !strings.HasPrefix(row.Name(), ".queue-") && !strings.HasPrefix(row.Name(), ".enroll-") {
			continue
		}
		path := filepath.Join(dir, row.Name())
		f, err := os.OpenFile(path, os.O_RDWR, 0)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return err
		}
		err = flock.Exclusive(f, true)
		if err == nil {
			err = os.Remove(path)
		}
		f.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) retainUnknown(ctx context.Context, op Operation) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	path := filepath.Join(s.Root, "unresolved", op.ID+".json")
	if old, err := readBounded(path); err == nil {
		previous, decodeErr := DecodeOperation(old)
		if decodeErr != nil || previous.ID != op.ID {
			return errors.New("existing unresolved evidence is unreadable or mismatched")
		}
		return nil // Immutable recovery evidence preserves unknown additive fields.
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	rows, err := operationFiles(filepath.Join(s.Root, "unresolved"))
	if err != nil || len(rows) >= MaxOperations {
		return errors.New("unresolved queue registry full/unreadable; repair required")
	}
	data, err := EncodeOperation(op)
	if err != nil {
		return err
	}
	return atomicWrite(path, data)
}

func (s *Store) retainInterruption(ctx context.Context, op Operation, reason string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	path := filepath.Join(s.Root, "interrupted", op.ID+".json")
	if old, err := readBounded(path); err == nil {
		previous, decodeErr := DecodeInterruptionRecord(old)
		if decodeErr != nil || previous.ID != op.ID {
			return errors.New("existing interrupted evidence is unreadable or mismatched")
		}
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if reason == "" {
		reason = "queue cancellation interrupted this operation"
	}
	data, err := EncodeInterruptionRecord(InterruptionRecord{SchemaVersion: 1, ID: op.ID, Outcome: "interrupted", Operation: op, Reason: reason, RecordedAt: s.now()})
	if err != nil {
		return err
	}
	if err := s.pruneInterruptions(ctx, 1); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := atomicWrite(path, data); err != nil {
		return err
	}
	return nil
}

func (s *Store) pruneInterruptions(ctx context.Context, reserve int) error {
	dir := filepath.Join(s.Root, "interrupted")
	rows, err := operationFiles(dir)
	if err != nil {
		return err
	}
	type dated struct {
		name string
		at   time.Time
	}
	var keep []dated
	for _, row := range rows {
		if err := ctx.Err(); err != nil {
			return err
		}
		if row.IsDir() {
			continue
		}
		info, err := row.Info()
		if err != nil {
			return err
		}
		if info.ModTime().Before(s.now().Add(-30 * 24 * time.Hour)) {
			if err := os.Remove(filepath.Join(dir, row.Name())); err != nil {
				return err
			}
		} else {
			keep = append(keep, dated{row.Name(), info.ModTime()})
		}
	}
	sort.Slice(keep, func(i, j int) bool { return keep[i].at.After(keep[j].at) })
	for _, row := range keep[min(len(keep), MaxOperations-reserve):] {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := os.Remove(filepath.Join(dir, row.name)); err != nil {
			return err
		}
	}
	return nil
}
