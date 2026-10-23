package landingline

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// The line's two append-only logs. Both are written under the line mutex, so
// a sweep's rewrite never loses a concurrent append, and read without it: a
// torn or hand-edited line is skipped and counted, never trusted.
const (
	EventsFile = "events.jsonl"
	GrantsFile = "grants.jsonl"
	// Retention matches test-stats' local records (ARCH-FUNERAL).
	Retention = 30 * 24 * time.Hour
)

const (
	EventEnqueue = "enqueue"
	EventAdmit   = "admit"
	EventLanded  = "landed"
	EventRefused = "refused"
	// EventResumed: admitted, the merge found its PR already landed by another
	// invocation, and finished cleanup. Not a landing, so Delays skips it.
	EventResumed = "resumed"
	// EventTrain (#449): a train's outcome, written by its conductor. Its
	// NextDue is the schedule (DepartureDue reads the latest one).
	EventTrain = "train"
	// EventCar (#449): one car's outcome on a train (landed, held, released).
	EventCar = "car"
	// EventTrainMeasure (#561): a train's measurement record (start or
	// final), evidence only; Measurement holds its versioned payload.
	EventTrainMeasure = "train_measure"
)

// Event is one line of the audit and pilot log.
type Event struct {
	Wait      *WaitFact `json:"wait,omitempty"`
	Kind      string    `json:"kind"`
	At        time.Time `json:"at"`
	Seq       int64     `json:"seq"`
	Issue     int       `json:"issue,omitempty"`
	Branch    string    `json:"branch,omitempty"`
	PR        int       `json:"pr,omitempty"`
	Displaced []int64   `json:"displaced,omitempty"`
	Detail    string    `json:"detail,omitempty"`
	// Train fields (#449), additive: a reader before them ignores them, and
	// the sweep keeps each line's raw bytes, so a rewrite never drops them.
	Train   int64     `json:"train,omitempty"` // the conductor's ticket
	Cars    []int64   `json:"cars,omitempty"`
	Result  string    `json:"result,omitempty"`
	Splits  int       `json:"splits,omitempty"`
	Culprit int64     `json:"culprit,omitempty"`
	NextDue time.Time `json:"next_due,omitzero"`
	// Departed is when a train (or a lone landing) left (#475), additive: the
	// completion time At is not the departure.
	Departed time.Time `json:"departed,omitzero"`
	// Measurement is a train_measure event's payload (#561), kept as raw
	// bytes: AggregateTrainMeasurements decodes it by version.
	Measurement json.RawMessage `json:"measurement,omitempty"`
}

// Record appends e, stamping it with the line's clock when e.At is zero.
func (l *Line) Record(e Event) error {
	if e.At.IsZero() {
		e.At = l.o.Now()
	}
	return appendLog(l, EventsFile, e, KeepEvents)
}

// AddGrant appends a grant or a revoke, stamping Granted with the line's clock
// when it is zero (a revoke's Granted is when it was revoked). The caller
// authorizes it first; the read side (ValidGrants) checks the grantor again.
func (l *Line) AddGrant(g Grant) error {
	if g.Granted.IsZero() {
		g.Granted = l.o.Now()
	}
	if !g.Revoked && !g.boundedExpiry() {
		return fmt.Errorf("landing grant: expiry must be after grant time and within %s", MaxGrant)
	}
	return appendLog(l, GrantsFile, g, KeepGrants)
}

func (l *Line) Events() ([]Event, int, error) {
	return readLog[Event](filepath.Join(l.dir, EventsFile))
}

func (l *Line) GrantEntries() ([]Grant, int, error) {
	return ReadGrantEntries(l.dir)
}

// GrantsNow is the one answer to "who holds priority now": the valid grants
// at the line's clock. An unreadable log is an error for the caller to
// treat as no priority; every caller treats it so (none is guessed).
func (l *Line) GrantsNow() (map[int]Grant, error) {
	entries, _, err := l.GrantEntries()
	if err != nil {
		return nil, err
	}
	return ValidGrants(entries, l.o.Now()), nil
}

// ReadGrantEntries reads an existing grant log without opening or creating a line.
// Missing logs are empty; malformed entries are counted for the caller.
func ReadGrantEntries(dir string) ([]Grant, int, error) {
	return readLog[Grant](filepath.Join(dir, GrantsFile))
}

// Delays are the ready-to-landed delays in seconds, enqueue to landed, for the
// landings whose enqueue is still in the log.
func Delays(events []Event) []float64 {
	enqueued := map[int64]time.Time{}
	var out []float64
	for _, e := range events {
		switch e.Kind {
		case EventEnqueue:
			enqueued[e.Seq] = e.At
		case EventLanded:
			if at, ok := enqueued[e.Seq]; ok {
				out = append(out, e.At.Sub(at).Seconds())
			}
		}
	}
	return out
}

func appendLog[T any](l *Line, name string, v T, keep func([]T, time.Time) []bool) error {
	return l.withMutex(func() error { return appendLogLocked(l, name, v, keep) })
}

// appendLogLocked requires the existing line mutex, including for wait facts.
func appendLogLocked[T any](l *Line, name string, v T, keep func([]T, time.Time) []bool) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	path := filepath.Join(l.dir, name)
	if err := sweep(l, path, keep); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	_, err = f.Write(append(data, '\n'))
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	return err
}

// sweep rewrites the log with the entries keep marks (retention.go owns that
// rule), at most once a day per log (its .swept stamp). A line it can't parse
// is dropped, and the stamp records how many ("malformed=N").
func sweep[T any](l *Line, path string, keep func([]T, time.Time) []bool) error {
	now := l.o.Now()
	stamp := path + ".swept"
	if fi, err := os.Stat(stamp); err == nil && now.Sub(fi.ModTime()) < 24*time.Hour && !fi.ModTime().After(now) {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	var raws [][]byte
	var entries []T
	malformed := 0
	for _, raw := range bytes.Split(data, []byte("\n")) {
		if len(bytes.TrimSpace(raw)) == 0 {
			continue
		}
		var v T
		if json.Unmarshal(raw, &v) != nil {
			malformed++
			continue
		}
		raws, entries = append(raws, raw), append(entries, v)
	}
	if len(data) > 0 {
		var kept bytes.Buffer
		for i, k := range keep(entries, now) {
			if k {
				kept.Write(raws[i])
				kept.WriteByte('\n')
			}
		}
		tmp := path + ".tmp"
		if err := os.WriteFile(tmp, kept.Bytes(), 0o644); err != nil {
			return err
		}
		if err := os.Rename(tmp, path); err != nil {
			return err
		}
	}
	if err := os.WriteFile(stamp, []byte(fmt.Sprintf("malformed=%d\n", malformed)), 0o644); err != nil {
		return err
	}
	return os.Chtimes(stamp, now, now)
}

func readLog[T any](path string) ([]T, int, error) {
	f, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, 0, nil
	}
	if err != nil {
		return nil, 0, err
	}
	defer f.Close()
	var out []T
	skipped := 0
	r := bufio.NewReader(f) // no line limit, matching sweep's: a long line is an entry, not an error
	for {
		raw, err := r.ReadBytes('\n')
		if line := bytes.TrimSpace(raw); len(line) > 0 {
			var v T
			if json.Unmarshal(line, &v) != nil {
				skipped++
			} else {
				out = append(out, v)
			}
		}
		if errors.Is(err, io.EOF) {
			return out, skipped, nil
		}
		if err != nil {
			return out, skipped, err
		}
	}
}
