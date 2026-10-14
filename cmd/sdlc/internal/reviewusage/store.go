package reviewusage

import (
	"bytes"
	"context"
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

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/repolock"
)

const MaxRecords = 4096

// Store is a thin caller-context adapter over the shared Git CAS/object store.
// Its dedicated admission mutex is held only while creating a new receipt,
// never during an external reviewer or a gate's semantic lock boundary.
type Store struct {
	root  string
	limit int
}
type Entry struct {
	OID    string
	Record Record
}

func NewStore(root string) (*Store, error) {
	if root == "" {
		return nil, errors.New("review usage store requires repository")
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	return &Store{root: abs, limit: MaxRecords}, nil
}
func (s *Store) raw(ctx context.Context) (*gitx.RecoveryStore, error) {
	return gitx.NewRecoveryStoreFor(ctx, s.root, "review-usage")
}
func (s *Store) Load(ctx context.Context, id string) (Entry, error) {
	raw, err := s.raw(ctx)
	if err != nil {
		return Entry{}, err
	}
	entry, err := raw.Load(id)
	if err != nil {
		return Entry{}, err
	}
	r, err := Parse(entry.Data.Document)
	if err != nil {
		return Entry{}, err
	}
	if r.Intent.ID != entry.Token {
		return Entry{}, errors.New("review usage ref identity mismatch")
	}
	return Entry{OID: entry.OID, Record: r}, nil
}
func (s *Store) List(ctx context.Context) ([]Entry, error) {
	raw, err := s.raw(ctx)
	if err != nil {
		return nil, err
	}
	refs, err := raw.List()
	if err != nil {
		return nil, err
	}
	if len(refs) > s.limit {
		return nil, fmt.Errorf("review usage refs exceed limit %d", s.limit)
	}
	out := make([]Entry, 0, len(refs))
	for _, ref := range refs {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		entry, err := s.Load(ctx, ref.Token)
		if err != nil {
			return nil, err
		}
		out = append(out, entry)
	}
	return out, nil
}
func (s *Store) Delete(ctx context.Context, id, expectedOID string) error {
	raw, err := s.raw(ctx)
	if err != nil {
		return err
	}
	return raw.Delete(id, expectedOID)
}
func (s *Store) SaveRequested(ctx context.Context, intent Intent) (string, error) {
	r, err := NewRequested(intent)
	if err != nil {
		return "", err
	}
	return s.Save(ctx, r, "")
}
func (s *Store) Save(ctx context.Context, r Record, expectedOID string) (string, error) {
	doc, err := Marshal(r)
	if err != nil {
		return "", err
	}
	raw, err := s.raw(ctx)
	if err != nil {
		return "", err
	}
	if expectedOID != "" {
		prior, err := s.Load(ctx, r.Intent.ID)
		if err != nil {
			return "", err
		}
		if prior.OID != expectedOID {
			return "", gitx.ErrRecoveryChanged
		}
		if err := Advance(prior.Record, r); err != nil {
			return "", err
		}
		if prior.Record == r {
			return prior.OID, nil
		}
		return raw.Save(r.Intent.ID, expectedOID, gitx.RecoveryData{Document: doc, Objects: []string{r.Intent.ReviewedHEAD}})
	}
	if r.Phase != Requested {
		return "", errors.New("new review usage receipt must be requested")
	}
	out, diag, err := gitx.RunGitContext(ctx, "-C", s.root, "rev-parse", "--git-common-dir")
	if err != nil {
		return "", fmt.Errorf("review usage common directory: %w: %s", err, diag)
	}
	common := strings.TrimSpace(string(out))
	if common == "" {
		return "", errors.New("empty review usage common directory")
	}
	if !filepath.IsAbs(common) {
		common = filepath.Join(s.root, common)
	}
	admission := filepath.Join(common, "sdlc-review-usage-admission")
	if err := os.MkdirAll(admission, 0700); err != nil {
		return "", err
	}
	host, err := os.Hostname()
	if err != nil {
		return "", fmt.Errorf("review usage admission hostname: %w", err)
	}
	lock, err := repolock.Acquire(ctx, repolock.Options{
		GitCommonDir: admission, Command: "review usage admission", CallerHandlesSignals: true, Stderr: io.Discard,
		PID: os.Getpid(), Hostname: host, CWD: s.root, WaitTimeout: 10 * time.Second,
		ProcessAlive: func(pid int) bool { return pid > 0 && !errors.Is(syscall.Kill(pid, 0), syscall.ESRCH) },
	})
	if err != nil {
		return "", err
	}
	defer lock.Release()
	refs, err := raw.List()
	if err != nil {
		return "", err
	}
	if len(refs) >= s.limit {
		return "", fmt.Errorf("review usage store full (%d retained refs); publish identical attempt evidence and run confirmed-reachability cleanup before dispatch", len(refs))
	}
	return raw.Save(r.Intent.ID, "", gitx.RecoveryData{Document: doc, Objects: []string{r.Intent.ReviewedHEAD}})
}

// Record supports a recorder callback while preserving per-record CAS. A
// concurrent writer is reported, never silently overwritten or retried.
func (s *Store) Record(ctx context.Context, r Record) error {
	prior, err := s.Load(ctx, r.Intent.ID)
	if errors.Is(err, gitx.ErrRecoveryAbsent) {
		_, err = s.Save(ctx, r, "")
		return err
	}
	if err != nil {
		return err
	}
	_, err = s.Save(ctx, r, prior.OID)
	return err
}
func Marshal(r Record) ([]byte, error) {
	if err := Validate(r); err != nil {
		return nil, err
	}
	return json.Marshal(r)
}

func uniqueKeys(d *json.Decoder, depth int) error {
	if depth > 16 {
		return errors.New("review usage JSON nesting exceeds limit")
	}
	tok, err := d.Token()
	if err != nil {
		return err
	}
	delim, ok := tok.(json.Delim)
	if !ok {
		return nil
	}
	switch delim {
	case '{':
		seen := map[string]bool{}
		for d.More() {
			key, err := d.Token()
			if err != nil {
				return err
			}
			name, ok := key.(string)
			if !ok || seen[name] {
				return errors.New("duplicate review usage JSON field")
			}
			seen[name] = true
			if err := uniqueKeys(d, depth+1); err != nil {
				return err
			}
		}
	case '[':
		for d.More() {
			if err := uniqueKeys(d, depth+1); err != nil {
				return err
			}
		}
	default:
		return errors.New("invalid review usage JSON delimiter")
	}
	_, err = d.Token()
	return err
}
func objectShape(raw []byte, required, optional string) (map[string]json.RawMessage, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil || fields == nil {
		return nil, errors.New("review usage JSON requires object")
	}
	allowed := map[string]bool{}
	for _, k := range strings.Fields(required) {
		allowed[k] = true
		if _, ok := fields[k]; !ok {
			return nil, fmt.Errorf("missing review usage field %s", k)
		}
	}
	for _, k := range strings.Fields(optional) {
		allowed[k] = true
	}
	for k, v := range fields {
		if !allowed[k] || bytes.Equal(bytes.TrimSpace(v), []byte("null")) {
			return nil, fmt.Errorf("invalid review usage field %s", k)
		}
	}
	return fields, nil
}
func Parse(raw []byte) (Record, error) {
	if len(raw) == 0 || len(raw) > MaxBytes || !utf8.Valid(raw) {
		return Record{}, errors.New("invalid review usage document encoding/size")
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	if err := uniqueKeys(d, 0); err != nil {
		return Record{}, err
	}
	if _, err := d.Token(); err != io.EOF {
		return Record{}, errors.New("trailing review usage JSON")
	}
	fields, err := objectShape(raw, "version intent phase", "pid started_at finished_at session_id outcome diagnostic")
	if err != nil {
		return Record{}, err
	}
	if _, err := objectShape(fields["intent"], "id repository issue_id recipe round attempt agent root reviewed_head slot requested_at", "disposable_root"); err != nil {
		return Record{}, err
	}
	var r Record
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&r); err != nil {
		return r, err
	}
	return r, Validate(r)
}
