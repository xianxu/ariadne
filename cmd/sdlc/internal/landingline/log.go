package landingline

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
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
)

// Event is one line of the audit and pilot log.
type Event struct {
	Kind      string    `json:"kind"`
	At        time.Time `json:"at"`
	Seq       int64     `json:"seq"`
	Issue     int       `json:"issue,omitempty"`
	Branch    string    `json:"branch,omitempty"`
	PR        int       `json:"pr,omitempty"`
	Displaced []int64   `json:"displaced,omitempty"`
	Detail    string    `json:"detail,omitempty"`
}

// Record appends e, stamping it with the line's clock when e.At is zero.
func (l *Line) Record(e Event) error {
	if e.At.IsZero() {
		e.At = l.o.Now()
	}
	return l.appendLog(EventsFile, e, func(raw []byte) (time.Time, bool) {
		var e Event
		return e.At, json.Unmarshal(raw, &e) == nil
	})
}

// AddGrant appends a grant or a revoke, stamping Granted with the line's clock
// when it is zero. The caller authorizes it first; the read side (ValidGrants)
// checks the grantor again. Retention keys on the later of Granted and Expires,
// so a revoke outlives the grant it revokes.
func (l *Line) AddGrant(g Grant) error {
	if g.Granted.IsZero() {
		g.Granted = l.o.Now()
	}
	return l.appendLog(GrantsFile, g, func(raw []byte) (time.Time, bool) {
		var g Grant
		if json.Unmarshal(raw, &g) != nil {
			return time.Time{}, false
		}
		if g.Expires.After(g.Granted) {
			return g.Expires, true
		}
		return g.Granted, true
	})
}

func (l *Line) Events() ([]Event, int, error) { return readLog[Event](filepath.Join(l.dir, EventsFile)) }

func (l *Line) GrantEntries() ([]Grant, int, error) {
	return readLog[Grant](filepath.Join(l.dir, GrantsFile))
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

func (l *Line) appendLog(name string, v any, at func([]byte) (time.Time, bool)) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return l.withMutex(func() error {
		path := filepath.Join(l.dir, name)
		if err := l.sweep(path, at); err != nil {
			return err
		}
		f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return err
		}
		_, err = f.Write(append(data, '\n'))
		if cerr := f.Close(); err == nil {
			err = cerr
		}
		return err
	})
}

// sweep keeps the entries dated within Retention and not in the future, at most
// once a day per log (its .swept stamp). A line it can't parse is dropped, and
// the stamp records how many ("malformed=N"), so a damaged log is visible.
func (l *Line) sweep(path string, at func([]byte) (time.Time, bool)) error {
	now := l.o.Now()
	stamp := path + ".swept"
	if fi, err := os.Stat(stamp); err == nil && now.Sub(fi.ModTime()) < 24*time.Hour && !fi.ModTime().After(now) {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	var kept bytes.Buffer
	malformed := 0
	for _, raw := range bytes.Split(data, []byte("\n")) {
		if len(bytes.TrimSpace(raw)) == 0 {
			continue
		}
		t, ok := at(raw)
		if !ok {
			malformed++
			continue
		}
		if age := now.Sub(t); age >= 0 && age <= Retention {
			kept.Write(raw)
			kept.WriteByte('\n')
		}
	}
	if len(data) > 0 {
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
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	for sc.Scan() {
		if len(bytes.TrimSpace(sc.Bytes())) == 0 {
			continue
		}
		var v T
		if json.Unmarshal(sc.Bytes(), &v) != nil {
			skipped++
			continue
		}
		out = append(out, v)
	}
	return out, skipped, sc.Err()
}
