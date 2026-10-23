package main

import (
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
	"syscall"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/bench"
)

// benchRecordDeps is a private failure-injection seam. The public command always
// uses the real append owner and records production evidence as nonsynthetic.
type benchRecordDeps struct {
	Couch     benchCouchDeps
	Append    func(context.Context, string, bench.LedgerRow) error
	Synthetic bool
}
type benchResultEvidence struct {
	Version      int                        `json:"version"`
	RunID        string                     `json:"run_id"`
	Task         bench.Task                 `json:"task"`
	Events       []json.RawMessage          `json:"events"`
	Measurements bench.Measurements         `json:"measurements"`
	Acceptance   bench.Acceptance           `json:"acceptance"`
	Observations map[string]json.RawMessage `json:"observations"`
}

func recordBenchRun(ctx context.Context, s *benchStore, ledger string, d benchCouchDeps) error {
	return recordBenchRunWith(ctx, s, ledger, benchRecordDeps{Couch: d})
}

func recordBenchRunWith(ctx context.Context, s *benchStore, ledger string, d benchRecordDeps) error {
	if ctx == nil || s == nil {
		return errors.New("bench record requires context and store")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	release, err := s.acquireRun()
	if err != nil {
		return err
	}
	defer release()
	r, run, err := s.read()
	if err != nil {
		return err
	}
	switch run.Phase() {
	case bench.PhaseStopped, bench.PhaseScored, bench.PhaseAcceptanceUnavailable, bench.PhaseRecording, bench.PhaseRecordUnknown, bench.PhaseRecorded:
	default:
		return fmt.Errorf("bench cannot record in %s", run.Phase())
	}
	if err := verifyBenchRetirement(ctx, s, d.Couch); err != nil {
		return err
	}
	runtime, err := benchRetirementSubject(r, run)
	if err != nil {
		return err
	}
	if retained, ok := r.Data["record_ledger"]; ok {
		var previous string
		if err := json.Unmarshal(retained, &previous); err != nil {
			return err
		}
		if ledger != "" && ledger != previous {
			return errors.New("bench ledger flag differs from immutable recording destination")
		}
		ledger = previous
	}
	if runtime.Ledger != "" {
		if ledger != "" && ledger != runtime.Ledger {
			return errors.New("bench ledger flag differs from retained runtime")
		}
		ledger = runtime.Ledger
	}
	if _, err := benchLedgerParent(ledger); err != nil {
		return err
	}
	if benchPathWithin(s.root, ledger) || benchPathWithin(r.Task.SourceRepo, ledger) {
		return errors.New("bench ledger must be outside run and source repository")
	}
	if info, err := os.Lstat(ledger); err == nil {
		if !info.Mode().IsRegular() {
			return errors.New("bench ledger is not a regular file")
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	var measurements bench.Measurements
	var acceptance bench.Acceptance
	if err := json.Unmarshal(r.Data["measurements"], &measurements); err != nil {
		return fmt.Errorf("retained measurements: %w", err)
	}
	if err := json.Unmarshal(r.Data["acceptance"], &acceptance); err != nil {
		return fmt.Errorf("retained acceptance: %w", err)
	}
	evidence := benchResultEvidence{Version: 1, RunID: r.ID, Task: r.Task, Events: []json.RawMessage{}, Measurements: measurements, Acceptance: acceptance, Observations: map[string]json.RawMessage{}}
	for _, raw := range r.Events {
		var event bench.Event
		if err := json.Unmarshal(raw, &event); err != nil {
			return err
		}
		if event.Kind == bench.BeginRecord {
			break
		}
		evidence.Events = append(evidence.Events, raw)
	}
	for key, raw := range r.Data {
		if key != "record_row" && key != "record_ledger" && key != "measurements" && key != "acceptance" {
			evidence.Observations[key] = raw
		}
	}
	raw, err := json.Marshal(evidence)
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	if len(raw) > benchRecordLimit {
		return errors.New("bench result evidence exceeds 1 MiB")
	}
	digest := sha256.Sum256(raw)
	digestText := hex.EncodeToString(digest[:])
	row, err := bench.BuildRow(r.Task, run, measurements, acceptance, digestText, d.Synthetic)
	if err != nil {
		return err
	}
	rowRaw, err := bench.EncodeLedgerRow(row)
	if err != nil {
		return err
	}
	// An existing snapshot is compared byte-for-byte; a partial prior write is
	// retained for reconciliation and can never be silently filled or replaced.
	if err := retainBenchResultEvidence(ctx, s, raw); err != nil {
		return err
	}
	if err := s.putData("record_ledger", ledger); err != nil {
		return err
	}
	if previous, ok := r.Data["record_row"]; ok {
		prior, err := bench.DecodeLedgerRow(previous)
		if err != nil {
			return err
		}
		encoded, err := bench.EncodeLedgerRow(prior)
		if err != nil {
			return err
		}
		if !bytes.Equal(encoded, rowRaw) {
			return errors.New("bench immutable row differs from retained result")
		}
	}
	if err := s.putData("record_row", json.RawMessage(rowRaw)); err != nil {
		return err
	}
	if run.Phase() == bench.PhaseRecorded {
		rows, err := readBenchLedger(ctx, ledger)
		if err != nil {
			return err
		}
		for _, existing := range rows {
			if existing.RunID == row.RunID {
				encoded, err := bench.EncodeLedgerRow(existing)
				if err != nil {
					return err
				}
				if bytes.Equal(encoded, rowRaw) {
					return nil
				}
				return errors.New("recorded bench ledger row conflicts with retained snapshot")
			}
		}
		return errors.New("confirmed bench row is missing from retained ledger")
	}
	op := "record-" + digestText[:24]
	h, w := run.Bindings()
	event := func(kind bench.EventKind, suffix, reason string) bench.Event {
		return bench.Event{ID: op + suffix, RunID: r.ID, Kind: kind, OperationID: op, Hub: h, Worker: w, Reason: reason}
	}
	if run.Phase() != bench.PhaseRecording && run.Phase() != bench.PhaseRecordUnknown {
		if _, _, err := s.apply(event(bench.BeginRecord, "-begin", "immutable result retained")); err != nil {
			return err
		}
	} else {
		// The retained authorization must belong to this exact row operation.
		found := false
		for _, raw := range r.Events {
			var e bench.Event
			if err := json.Unmarshal(raw, &e); err != nil {
				return err
			}
			if e.Kind == bench.BeginRecord && e.OperationID == op {
				found = true
			}
		}
		if !found {
			return errors.New("recording authorization differs from retained row")
		}
	}
	appendRow := d.Append
	if appendRow == nil {
		appendRow = recordBenchRow
	}
	if err := appendRow(ctx, ledger, row); err != nil {
		_, _, retainErr := s.apply(event(bench.RecordUnconfirmed, "-unconfirmed", "ledger append not confirmed; reconcile retained row and original ledger"))
		return errors.Join(err, retainErr)
	}
	_, _, err = s.apply(event(bench.RecordConfirmed, "-confirmed", "exact row durably present in retained ledger"))
	return err
}

func retainBenchResultEvidence(ctx context.Context, s *benchStore, raw []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := s.check(); err != nil {
		return err
	}
	path := filepath.Join(s.root, "observer", "result-evidence.json")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	created := err == nil
	if errors.Is(err, os.ErrExist) {
		info, statErr := os.Lstat(path)
		if statErr != nil {
			return statErr
		}
		if !info.Mode().IsRegular() {
			return errors.New("bench result evidence is not a regular file")
		}
		f, err = os.OpenFile(path, os.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0)
	}
	if err != nil {
		return err
	}
	var operationErr error
	if created {
		n, writeErr := f.Write(raw)
		if writeErr == nil && n != len(raw) {
			writeErr = io.ErrShortWrite
		}
		operationErr = writeErr
	} else {
		info, statErr := f.Stat()
		if statErr != nil {
			operationErr = statErr
		} else if !info.Mode().IsRegular() {
			operationErr = errors.New("bench result evidence is not a regular file")
		} else {
			previous, readErr := io.ReadAll(io.LimitReader(f, benchRecordLimit+1))
			operationErr = readErr
			if readErr == nil && !bytes.Equal(previous, raw) {
				operationErr = errors.New("immutable bench result evidence differs; reconcile retained bytes")
			}
		}
	}
	if operationErr == nil {
		operationErr = f.Sync()
	}
	if err := errors.Join(operationErr, f.Close()); err != nil {
		return err
	}
	return syncBenchLedgerParent(ctx, path)
}
