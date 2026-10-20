package queuegate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"
)

// InterruptionRecord is diagnostic evidence, not a claim that all descendants
// are gone. Operation retains its observed phase; unresolved activity remains
// independently visible until its lifetime is proved.
type InterruptionRecord struct {
	SchemaVersion int
	ID            string
	Outcome       string
	Operation     Operation
	Reason        string
	RecordedAt    time.Time
	Extra         map[string]json.RawMessage
}

func EncodeInterruptionRecord(r InterruptionRecord) ([]byte, error) {
	if r.SchemaVersion != 1 || r.Outcome != "interrupted" || r.ID != r.Operation.ID || r.RecordedAt.IsZero() || r.Reason == "" || len(r.Reason) > MaxText || !utf8.ValidString(r.Reason) {
		return nil, errors.New("invalid interruption record")
	}
	op, err := EncodeOperation(r.Operation)
	if err != nil {
		return nil, err
	}
	return encodeObject(map[string]json.RawMessage{"schema_version": rawValue(r.SchemaVersion), "id": rawValue(r.ID), "outcome": rawValue(r.Outcome), "operation": op, "reason": rawValue(r.Reason), "recorded_at": rawValue(r.RecordedAt)}, r.Extra)
}

func DecodeInterruptionRecord(data []byte) (InterruptionRecord, error) {
	var r InterruptionRecord
	var op json.RawMessage
	var err error
	r.Extra, err = decodeObject(data, map[string]any{"schema_version": &r.SchemaVersion, "id": &r.ID, "outcome": &r.Outcome, "operation": &op, "reason": &r.Reason, "recorded_at": &r.RecordedAt}, nil)
	if err != nil {
		return r, err
	}
	r.Operation, err = DecodeOperation(op)
	if err != nil {
		return r, err
	}
	_, err = EncodeInterruptionRecord(r)
	return r, err
}

func (s *Store) readIntent(id string) (reason string, present bool, err error) {
	data, err := readBounded(s.intentPath(id))
	if errors.Is(err, os.ErrNotExist) {
		return "", false, nil
	}
	if err != nil {
		return "", true, err
	}
	var version int
	var actual string
	_, err = decodeObject(data, map[string]any{"schema_version": &version, "id": &actual, "reason": &reason}, nil)
	if err != nil {
		return "", true, err
	}
	if version != 1 || actual != id || !safeID(actual) || len(reason) > MaxText || !utf8.ValidString(reason) {
		return "", true, errors.New("cancellation intent has invalid version, identity or reason")
	}
	return reason, true, nil
}

// An intent's last reader is either its retiring owner or stale-owner
// recovery. A crash after evidence publication but before intent removal is
// idempotently collected; missing/corrupt owner evidence remains unknown.
func (s *Store) reconcileOrphanIntents(ctx context.Context) ([]Operation, []string, error) {
	rows, err := operationFiles(filepath.Join(s.Root, "cancel"))
	if err != nil {
		return nil, []string{"cancellation registry unavailable: " + err.Error()}, nil
	}
	var unknown []Operation
	var details []string
	for _, row := range rows {
		if err := ctx.Err(); err != nil {
			return nil, nil, err
		}
		if row.IsDir() || !strings.HasSuffix(row.Name(), ".json") {
			details = append(details, "unexpected cancellation registry entry "+row.Name())
			continue
		}
		id := strings.TrimSuffix(row.Name(), ".json")
		if !safeID(id) {
			unknown = append(unknown, unknownOperation(row.Name(), "unreadable cancellation identity "+row.Name()))
			continue
		}
		referenced := false
		for _, family := range []string{"operations", "unresolved"} {
			_, statErr := os.Stat(filepath.Join(s.Root, family, row.Name()))
			if statErr == nil {
				referenced = true
				break
			}
			if !errors.Is(statErr, os.ErrNotExist) {
				referenced = true
				details = append(details, statErr.Error())
				break
			}
		}
		if referenced {
			continue
		}
		_, present, intentErr := s.readIntent(id)
		if intentErr != nil || !present {
			unknown = append(unknown, unknownOperation(id, fmt.Sprintf("orphan cancellation %s unreadable; repair required", id)))
			continue
		}
		data, readErr := readBounded(filepath.Join(s.Root, "interrupted", row.Name()))
		if readErr == nil {
			r, decodeErr := DecodeInterruptionRecord(data)
			if decodeErr == nil && r.ID == id {
				if err := os.Remove(s.intentPath(id)); err != nil && !errors.Is(err, os.ErrNotExist) {
					details = append(details, err.Error())
				}
				continue
			}
		}
		unknown = append(unknown, unknownOperation(id, "orphan cancellation "+id+" has no readable operation/interruption evidence; repair required"))
	}
	return unknown, details, nil
}
