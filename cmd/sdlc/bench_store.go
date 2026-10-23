package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unicode/utf8"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/bench"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/flock"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/jsonstrict"
)

const benchRecordLimit = 1 << 20
const benchEventLimit = 2000

// The raw record is the sole durable run authority. Accepted events retain
// their original JSON (including future members); Step reconstructs its phase.
// Private evidence is not stored inside either agent's checkout.
type benchRecord struct {
	Version   int                        `json:"version"`
	ID        string                     `json:"id"`
	Root      string                     `json:"root"`
	CreatedAt string                     `json:"created_at"`
	Task      bench.Task                 `json:"task"`
	Events    []json.RawMessage          `json:"events"`
	Data      map[string]json.RawMessage `json:"data"`
	Unknown   map[string]json.RawMessage `json:"-"`
}

type benchStore struct{ root string }

func createBenchStore(root string, task bench.Task) (_ *benchStore, retErr error) {
	if err := bench.ValidateTask(task); err != nil {
		return nil, err
	}
	if !filepath.IsAbs(root) || filepath.Clean(root) != root || root == string(filepath.Separator) {
		return nil, errors.New("bench run directory must be an absolute clean path")
	}
	parent, err := filepath.EvalSymlinks(filepath.Dir(root))
	if err != nil {
		return nil, fmt.Errorf("bench run parent must already exist: %w", err)
	}
	root = filepath.Join(parent, filepath.Base(root))
	source, err := filepath.EvalSymlinks(task.SourceRepo)
	if err != nil {
		return nil, fmt.Errorf("resolve bench source before creating a run: %w", err)
	}
	if root == source || strings.HasPrefix(root, source+string(filepath.Separator)) {
		return nil, errors.New("bench run must be outside its source repository")
	}
	var token [16]byte
	if _, err := rand.Read(token[:]); err != nil {
		return nil, err
	}
	r := benchRecord{Version: 1, ID: hex.EncodeToString(token[:]), Root: root,
		CreatedAt: time.Now().UTC().Format(time.RFC3339Nano), Task: task,
		Events: []json.RawMessage{}, Data: map[string]json.RawMessage{}}
	// The enclosing record has overhead beyond the task's own bound. Validate
	// the writer envelope before creating any disposable state.
	if _, err := encodeBenchRecord(r); err != nil {
		return nil, err
	}
	// Exclusive creation makes a collision (including a symlink) a refusal,
	// never adoption of an arbitrary directory as an owned disposable run.
	if err := os.Mkdir(root, 0700); err != nil {
		return nil, fmt.Errorf("create new bench run: %w", err)
	}
	owned, err := os.Lstat(root)
	if err != nil {
		return nil, err
	}
	defer func() {
		if retErr == nil {
			return
		}
		current, err := os.Lstat(root)
		if err != nil || !os.SameFile(owned, current) {
			retErr = errors.Join(retErr, errors.New("bench setup root changed; partial state retained"))
			return
		}
		retErr = errors.Join(retErr, os.RemoveAll(root))
	}()
	s := &benchStore{root: root}
	if err := os.Mkdir(filepath.Join(root, "observer"), 0700); err != nil {
		return nil, err
	}
	if err := s.write(r); err != nil {
		return nil, err
	}
	return s, nil
}

func openBenchStore(root string) (*benchStore, error) {
	if !filepath.IsAbs(root) || filepath.Clean(root) != root {
		return nil, errors.New("bench run directory must be an absolute clean path")
	}
	s := &benchStore{root: root}
	if _, _, err := s.read(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *benchStore) recordPath() string { return filepath.Join(s.root, "observer", "run.json") }

func (s *benchStore) check() error {
	for _, dir := range []string{s.root, filepath.Join(s.root, "observer")} {
		info, err := os.Lstat(dir)
		if err != nil {
			return err
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("bench path is not an owned directory: %s", dir)
		}
		physical, err := filepath.EvalSymlinks(dir)
		if err != nil || physical != dir {
			return fmt.Errorf("bench path changed or has symlink ancestry: %s", dir)
		}
	}
	return nil
}

func (s *benchStore) read() (benchRecord, bench.Run, error) {
	if err := s.check(); err != nil {
		return benchRecord{}, bench.Run{}, err
	}
	f, err := os.OpenFile(s.recordPath(), os.O_RDONLY|syscall.O_NOFOLLOW, 0)
	if err != nil {
		return benchRecord{}, bench.Run{}, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() {
		return benchRecord{}, bench.Run{}, errors.New("bench record is not a regular file")
	}
	raw, err := io.ReadAll(io.LimitReader(f, benchRecordLimit+1))
	if err != nil {
		return benchRecord{}, bench.Run{}, err
	}
	r, state, err := decodeBenchRecord(raw)
	if err == nil && r.Root != s.root {
		err = errors.New("bench record belongs to a different run directory")
	}
	return r, state, err
}

func decodeBenchRecord(raw []byte) (benchRecord, bench.Run, error) {
	bad := func(err error) (benchRecord, bench.Run, error) { return benchRecord{}, bench.Run{}, err }
	if len(raw) > benchRecordLimit || !utf8.Valid(raw) {
		return bad(errors.New("bench record exceeds 1 MiB or has invalid UTF8"))
	}
	if err := jsonstrict.RejectDuplicateKeys(raw); err != nil {
		return bad(err)
	}
	var r benchRecord
	if err := json.Unmarshal(raw, &r); err != nil {
		return bad(err)
	}
	if r.Version != 1 || r.Events == nil || r.Data == nil || len(r.Events) > benchEventLimit {
		return bad(errors.New("invalid bench record version, collections or event budget"))
	}
	if _, err := time.Parse(time.RFC3339Nano, r.CreatedAt); err != nil {
		return bad(fmt.Errorf("invalid bench creation timestamp: %w", err))
	}
	if err := bench.ValidateTask(r.Task); err != nil {
		return bad(err)
	}
	state, err := bench.NewRun(r.ID)
	if err != nil {
		return bad(err)
	}
	for _, rawEvent := range r.Events {
		var event bench.Event
		if err := json.Unmarshal(rawEvent, &event); err != nil {
			return bad(err)
		}
		state, _, err = bench.Step(state, event)
		if err != nil {
			return bad(fmt.Errorf("invalid bench event history: %w", err))
		}
	}
	if err := json.Unmarshal(raw, &r.Unknown); err != nil {
		return bad(err)
	}
	for _, known := range []string{"version", "id", "root", "created_at", "task", "events", "data"} {
		for key := range r.Unknown {
			if strings.EqualFold(key, known) && key != known {
				return bad(fmt.Errorf("bench field must use canonical spelling: %s", known))
			}
		}
		delete(r.Unknown, known)
	}
	return r, state, nil
}

func (s *benchStore) lock(name string) (func(), error) {
	if err := s.check(); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(filepath.Join(s.root, "observer", name), os.O_CREATE|os.O_RDWR|syscall.O_NOFOLLOW, 0600)
	if err != nil {
		return nil, err
	}
	if err := flock.Exclusive(f, true); err != nil {
		f.Close()
		return nil, fmt.Errorf("bench owner is active (%s): %w", name, err)
	}
	return func() { _ = f.Close() }, nil
}

func (s *benchStore) acquireRun() (func(), error) { return s.lock("owner.lock") }

func (s *benchStore) apply(event bench.Event) (bench.Run, []bench.Effect, error) {
	release, err := s.lock("transaction.lock")
	if err != nil {
		return bench.Run{}, nil, err
	}
	defer release()
	r, state, err := s.read()
	if err != nil {
		return bench.Run{}, nil, err
	}
	next, effects, err := bench.Step(state, event)
	if err != nil {
		return state, nil, err
	}
	for _, raw := range r.Events {
		var previous bench.Event
		if err := json.Unmarshal(raw, &previous); err != nil {
			return state, nil, err
		}
		if previous.ID == event.ID {
			return next, effects, nil // Step already checked exact identity.
		}
	}
	if len(r.Events) >= benchEventLimit {
		return state, nil, errors.New("bench event budget exhausted; retain run for inspection")
	}
	raw, err := json.Marshal(event)
	if err != nil {
		return state, nil, err
	}
	r.Events = append(r.Events, raw)
	if err := s.write(r); err != nil {
		return state, nil, err // No caller may execute unpersisted effects.
	}
	return next, effects, nil
}

func encodeBenchRecord(r benchRecord) ([]byte, error) {
	raw, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil {
		return nil, err
	}
	for key, value := range r.Unknown {
		for known := range object {
			if strings.EqualFold(key, known) {
				return nil, errors.New("bench extension shadows a known field")
			}
		}
		object[key] = value
	}
	raw, err = json.Marshal(object)
	if err != nil {
		return nil, err
	}
	if _, _, err := decodeBenchRecord(raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func (s *benchStore) write(r benchRecord) error {
	if err := s.check(); err != nil {
		return err
	}
	raw, err := encodeBenchRecord(r)
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(s.recordPath()), ".record-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	_, writeErr := f.Write(raw)
	if writeErr == nil {
		writeErr = f.Sync()
	}
	err = errors.Join(writeErr, f.Close())
	if err != nil {
		return err
	}
	return os.Rename(f.Name(), s.recordPath())
}
