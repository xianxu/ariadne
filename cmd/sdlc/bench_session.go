package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"
	"unicode/utf8"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/bench"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/flock"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/jsonstrict"
)

const benchRecordLimit = 1 << 20
const benchEventLimit = 2000

// Facts are observations for this invocation, not an independently replayable
// execution record. Metis owns experiment identity, execution and its record.
type benchSessionFacts struct {
	ID        string
	Root      string
	CreatedAt string
	Task      bench.Task
	Data      map[string]json.RawMessage
}

type benchSession struct {
	root                   string
	rootInfo, observerInfo os.FileInfo
	mu                     sync.Mutex
	facts                  benchSessionFacts
	state                  bench.Run
}

func createBenchSession(root string, task bench.Task, identity ...string) (*benchSession, error) {
	if err := bench.ValidateTask(task); err != nil {
		return nil, err
	}
	// Snapshot nested caller-owned slices/maps before retaining task evidence.
	rawTask, err := bench.EncodeTask(task)
	if err != nil {
		return nil, err
	}
	task, err = bench.DecodeTask(rawTask)
	if err != nil {
		return nil, err
	}
	if !filepath.IsAbs(root) || filepath.Clean(root) != root || root == string(filepath.Separator) {
		return nil, errors.New("bench session directory must be an absolute clean path")
	}
	parent, err := filepath.EvalSymlinks(filepath.Dir(root))
	if err != nil {
		return nil, err
	}
	if parent != filepath.Dir(root) {
		return nil, errors.New("bench session parent has symlink ancestry")
	}
	source, err := filepath.EvalSymlinks(task.SourceRepo)
	if errors.Is(err, os.ErrNotExist) {
		// A prepared immutable archive needs only the source provenance, not a
		// surviving host checkout. There is no source read at this boundary.
		source = task.SourceRepo
	} else if err != nil {
		return nil, err
	}
	if root == source || strings.HasPrefix(root, source+string(filepath.Separator)) {
		return nil, errors.New("bench session must be outside its source repository")
	}
	if len(identity) > 1 {
		return nil, errors.New("bench session requires one Metis repeat identity")
	}
	digest := sha256.Sum256([]byte(root))
	id := hex.EncodeToString(digest[:16])
	if len(identity) == 1 {
		id = identity[0]
	}
	if !regexp.MustCompile(`^[0-9a-f]{32}$`).MatchString(id) {
		return nil, errors.New("bench session identity must be 32 lowercase hexadecimal characters")
	}
	state, err := bench.NewRun(id)
	if err != nil {
		return nil, err
	}
	if err = os.Mkdir(root, 0700); err != nil {
		return nil, fmt.Errorf("create fresh bench session: %w", err)
	}
	// Any subsequent failure retains the partial directory as evidence. It can
	// never be adopted by a later process as resumable execution authority.
	rootInfo, err := os.Lstat(root)
	if err != nil {
		return nil, err
	}
	if err = os.Mkdir(filepath.Join(root, "observer"), 0700); err != nil {
		return nil, err
	}
	observerInfo, err := os.Lstat(filepath.Join(root, "observer"))
	if err != nil {
		return nil, err
	}
	return &benchSession{root: root, rootInfo: rootInfo, observerInfo: observerInfo, state: state, facts: benchSessionFacts{ID: id, Root: root, CreatedAt: time.Now().UTC().Format(time.RFC3339Nano), Task: task, Data: map[string]json.RawMessage{}}}, nil
}

func (s *benchSession) check() error {
	for _, entry := range []struct {
		path  string
		owned os.FileInfo
	}{{s.root, s.rootInfo}, {filepath.Join(s.root, "observer"), s.observerInfo}} {
		info, err := os.Lstat(entry.path)
		if err != nil {
			return err
		}
		if entry.owned == nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 || !os.SameFile(entry.owned, info) {
			return fmt.Errorf("bench session directory identity changed: %s", entry.path)
		}
		physical, err := filepath.EvalSymlinks(entry.path)
		if err != nil || physical != entry.path {
			return fmt.Errorf("bench session path has changed ancestry: %s", entry.path)
		}
	}
	return nil
}

func (s *benchSession) read() (benchSessionFacts, bench.Run, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.check(); err != nil {
		return benchSessionFacts{}, bench.Run{}, err
	}
	raw, err := json.Marshal(s.facts)
	if err != nil {
		return benchSessionFacts{}, bench.Run{}, err
	}
	var facts benchSessionFacts
	if err = json.Unmarshal(raw, &facts); err != nil {
		return benchSessionFacts{}, bench.Run{}, err
	}
	return facts, s.state, nil
}

func canonicalBenchFact(raw []byte) ([]byte, error) {
	if len(raw) > benchRecordLimit || !utf8.Valid(raw) {
		return nil, errors.New("bench observation exceeds 1 MiB or has invalid UTF8")
	}
	if err := jsonstrict.RejectDuplicateKeys(raw); err != nil {
		return nil, err
	}
	var v any
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	if err := d.Decode(&v); err != nil {
		return nil, err
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return nil, errors.New("bench observation has trailing data")
	}
	return json.Marshal(v)
}

func (s *benchSession) putData(key string, value any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !regexp.MustCompile(`^[a-z][a-z0-9_-]{0,127}$`).MatchString(key) {
		return errors.New("invalid bench observation key")
	}
	if err := s.check(); err != nil {
		return err
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	raw, err = canonicalBenchFact(raw)
	if err != nil {
		return err
	}
	if previous, present := s.facts.Data[key]; present {
		if !bytes.Equal(previous, raw) {
			return fmt.Errorf("bench observation %s already has different evidence", key)
		}
		actual, err := readBenchInput(filepath.Join(s.root, "observer", key+".json"), benchRecordLimit)
		if err != nil {
			return err
		}
		canonical, err := canonicalBenchFact(actual)
		if err != nil || !bytes.Equal(canonical, raw) {
			return fmt.Errorf("bench observation %s changed on disk", key)
		}
		return nil
	}
	if len(s.facts.Data) >= benchEventLimit {
		return errors.New("bench observation budget exhausted")
	}
	if err = s.writeExclusive(key+".json", raw); err != nil {
		return err
	}
	s.facts.Data[key] = raw
	return nil
}

func (s *benchSession) writeExclusive(name string, raw []byte) error {
	if err := s.check(); err != nil {
		return err
	}
	f, err := os.OpenFile(filepath.Join(s.root, "observer", name), os.O_CREATE|os.O_EXCL|os.O_WRONLY|syscall.O_NOFOLLOW, 0600)
	if err != nil {
		return err
	}
	_, writeErr := f.Write(raw)
	if writeErr == nil {
		writeErr = f.Sync()
	}
	return errors.Join(writeErr, f.Close())
}

func (s *benchSession) apply(event bench.Event) (bench.Run, []bench.Effect, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.check(); err != nil {
		return s.state, nil, err
	}
	next, effects, err := bench.Step(s.state, event)
	if err != nil {
		return s.state, nil, err
	}
	for _, effect := range effects {
		if effect.Kind == bench.Dispatch {
			raw, err := json.Marshal(event)
			if err != nil {
				return s.state, nil, err
			}
			if err = s.writeExclusive("dispatch-intent.json", raw); err != nil {
				return s.state, nil, fmt.Errorf("dispatch intent already exists or cannot be established; reconcile and retire this session, then start a fresh Metis repeat: %w", err)
			}
		}
	}
	s.state = next
	return next, effects, nil
}

func (s *benchSession) acquireRun() (func(), error) {
	if err := s.check(); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(filepath.Join(s.root, "observer", "owner.lock"), os.O_CREATE|os.O_RDWR|syscall.O_NOFOLLOW, 0600)
	if err != nil {
		return nil, err
	}
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() {
		f.Close()
		return nil, errors.New("bench session owner lock is not a regular file")
	}
	if err = flock.Exclusive(f, true); err != nil {
		f.Close()
		return nil, fmt.Errorf("bench session owner is active: %w", err)
	}
	if err = s.check(); err != nil {
		f.Close()
		return nil, err
	}
	return func() { _ = f.Close() }, nil
}
