package queuegate

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flock"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/jsonstrict"
)

// ReconcileRequest is an operator attestation, not a process inventory proof.
type ReconcileRequest struct{ ID, Digest, Actor, Note string }

func (r ReconcileRequest) validate() error {
	if !safeID(r.ID) {
		return errors.New("reconcile needs a safe --id")
	}
	if b, err := hex.DecodeString(r.Digest); err != nil || len(b) != sha256.Size || strings.ToLower(r.Digest) != r.Digest {
		return errors.New("reconcile needs a lowercase SHA256 --digest")
	}
	if strings.TrimSpace(r.Note) == "" || !boundedText(r.Note, MaxText) {
		return errors.New("reconcile needs a nonempty --note (at most 4096 bytes) attesting no descendants remain")
	}
	if strings.TrimSpace(r.Actor) == "" || !boundedText(r.Actor, MaxText) {
		return errors.New("reconcile actor unavailable or oversized")
	}
	return nil
}

type reconcileRecord struct {
	SchemaVersion int       `json:"schema_version"`
	ID            string    `json:"id"`
	Digest        string    `json:"digest"`
	Original      []byte    `json:"original"`
	Actor         string    `json:"actor"`
	Note          string    `json:"note"`
	At            time.Time `json:"at"`
}

// Base64 plus metadata can exceed the operation's 1 MiB limit. The log is
// streamed, never rewritten: unknown audit members and original bytes survive.
const maxReconcileLine = 2 * MaxJSON

func decodeReconcileRecord(raw []byte) (reconcileRecord, error) {
	var r reconcileRecord
	if len(raw) > maxReconcileLine || !utf8.Valid(raw) {
		return r, errors.New("reconciliation audit line exceeds its bound or is not UTF-8")
	}
	if err := jsonstrict.RejectDuplicateKeys(raw); err != nil {
		return r, err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return r, err
	}
	for _, key := range []string{"schema_version", "id", "digest", "original", "actor", "note", "at"} {
		if v, ok := fields[key]; !ok || bytes.Equal(bytes.TrimSpace(v), []byte("null")) {
			return r, fmt.Errorf("reconciliation audit missing/null %s", key)
		}
	}
	if err := json.Unmarshal(raw, &r); err != nil {
		return r, err
	}
	if r.SchemaVersion != 1 || r.At.IsZero() {
		return r, errors.New("invalid reconciliation audit version or time")
	}
	if err := (ReconcileRequest{ID: r.ID, Digest: r.Digest, Actor: r.Actor, Note: r.Note}).validate(); err != nil {
		return r, err
	}
	op, err := DecodeOperation(r.Original)
	if err != nil || op.ID != r.ID || op.Phase != "unknown" || op.PID <= 1 || op.PIDStarted == "" {
		return r, errors.New("invalid original operation in reconciliation audit")
	}
	sum := sha256.Sum256(r.Original)
	if hex.EncodeToString(sum[:]) != r.Digest {
		return r, errors.New("reconciliation audit original digest mismatch")
	}
	return r, nil
}

func reconcileLines(data []byte, atEOF bool) (int, []byte, error) {
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		return i + 1, data[:i], nil
	}
	if atEOF && len(data) > 0 {
		return 0, nil, errors.New("torn reconciliation audit line; inspect the log before retrying")
	}
	return 0, nil, nil
}

func regularReconcileFile(path string, flags int) (*os.File, error) {
	before, err := os.Lstat(path)
	if err != nil && !(errors.Is(err, os.ErrNotExist) && flags&os.O_CREATE != 0) {
		return nil, err
	}
	if before != nil && !before.Mode().IsRegular() {
		return nil, fmt.Errorf("reconciliation path is not a regular file: %s", path)
	}
	f, err := os.OpenFile(path, flags, 0600)
	if err != nil {
		return nil, err
	}
	after, err := f.Stat()
	if err != nil || !after.Mode().IsRegular() || before != nil && !os.SameFile(before, after) {
		f.Close()
		return nil, fmt.Errorf("reconciliation file identity changed: %s", path)
	}
	return f, nil
}

func (s *Store) findReconciliation(ctx context.Context, req ReconcileRequest) (*reconcileRecord, error) {
	f, err := regularReconcileFile(filepath.Join(s.Root, "reconcile.jsonl"), os.O_RDONLY)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scan := bufio.NewScanner(f)
	scan.Buffer(make([]byte, 4096), maxReconcileLine)
	scan.Split(reconcileLines)
	var found *reconcileRecord
	for scan.Scan() {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		r, err := decodeReconcileRecord(scan.Bytes())
		if err != nil {
			return nil, fmt.Errorf("reconciliation audit unreadable; inspect before retrying: %w", err)
		}
		if r.ID == req.ID {
			if r.Digest != req.Digest {
				return nil, errors.New("reconciliation audit has another digest for this id")
			}
			found = &r
		}
	}
	return found, scan.Err()
}

// Reconcile removes only the exact unknown tombstone after durable audit.
// The caller supplies kernel incarnation evidence; descendants are the note's
// explicit operator attestation. No process or topology scan enters admission.
func (s *Store) Reconcile(ctx context.Context, req ReconcileRequest, absent func(Operation) (bool, error)) error {
	if err := req.validate(); err != nil {
		return err
	}
	if absent == nil {
		return errors.New("original owner incarnation probe unavailable")
	}
	writeFile, syncFile := (*os.File).Write, (*os.File).Sync
	if s.reconcileWrite != nil {
		writeFile = s.reconcileWrite
	}
	if s.reconcileSync != nil {
		syncFile = s.reconcileSync
	}
	return s.WithGate(ctx, func() error {
		if _, err := os.Lstat(filepath.Join(s.Root, "operations", req.ID+".json")); !errors.Is(err, os.ErrNotExist) {
			return errors.New("operation record or lease exists or is unreadable; wait or inspect before reconciling")
		}
		path := filepath.Join(s.Root, "unresolved", req.ID+".json")
		f, err := regularReconcileFile(path, os.O_RDONLY)
		missing := errors.Is(err, os.ErrNotExist)
		var raw []byte
		if err != nil && !missing {
			return err
		}
		if !missing {
			defer f.Close()
			if err := flock.Exclusive(f, true); err != nil {
				return fmt.Errorf("unresolved operation lease is live or unavailable: %w", err)
			}
			raw, err = io.ReadAll(io.LimitReader(f, MaxJSON+1))
			if err != nil {
				return err
			}
			sum := sha256.Sum256(raw)
			if hex.EncodeToString(sum[:]) != req.Digest {
				return errors.New("tombstone raw digest differs; inspect and use its current digest")
			}
		}
		previous, err := s.findReconciliation(ctx, req)
		if err != nil {
			return err
		}
		if missing {
			if previous == nil {
				return errors.New("tombstone absent without matching reconciliation evidence")
			}
			raw = previous.Original
		}
		op, err := DecodeOperation(raw)
		if err != nil || op.ID != req.ID || op.Phase != "unknown" || op.PID <= 1 || op.PIDStarted == "" {
			return errors.New("reconcile requires an exact unknown operation with its original PID incarnation")
		}
		ended, err := absent(op)
		if err != nil {
			return fmt.Errorf("original owner incarnation unknown: %w", err)
		}
		if !ended {
			return errors.New("original owner incarnation is live; wait before reconciling")
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		reason, intent, err := s.readIntent(req.ID)
		if err != nil {
			return err
		}
		if intent {
			if err := s.retainInterruption(ctx, op, reason); err != nil {
				return err
			}
		}
		log, err := regularReconcileFile(filepath.Join(s.Root, "reconcile.jsonl"), os.O_CREATE|os.O_RDWR|os.O_APPEND)
		if err != nil {
			return err
		}
		defer log.Close()
		if previous == nil {
			line, err := json.Marshal(reconcileRecord{SchemaVersion: 1, ID: req.ID, Digest: req.Digest, Original: raw, Actor: req.Actor, Note: req.Note, At: s.now()})
			if err != nil {
				return err
			}
			if _, err := decodeReconcileRecord(line); err != nil {
				return err
			}
			if _, err := writeFile(log, append(line, '\n')); err != nil {
				return err
			}
		}
		if err := syncFile(log); err != nil {
			return err
		}
		if err := syncDirectoryWith(s.Root, syncFile); err != nil {
			return err
		}
		if s.afterReconcileAudit != nil {
			if err := s.afterReconcileAudit(); err != nil {
				return err
			}
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if !missing {
			if err := os.Remove(path); err != nil {
				return err
			}
		}
		return syncDirectoryWith(filepath.Dir(path), syncFile)
	})
}
