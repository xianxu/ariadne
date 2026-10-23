package landingline

import (
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

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flock"
)

// A merge intent (#565) is the one landing record that outlives its process:
// intents/<key> names a branch whose sdlc merge started, flock-held by that
// merge. A terminal outcome removes it; an interruption or a kill leaves it
// unlocked, so after a full shutdown the next status names the branch and
// the one command that re-enters its landing.

// IntentVersion is the record format; a reader refuses another.
const IntentVersion = 1

// IntentMaxAge bounds an orphan nobody re-enters (the events log's span).
const IntentMaxAge = 30 * 24 * time.Hour

// Intent is one in-flight merge.
type Intent struct {
	Version   int                        `json:"version"`
	Branch    string                     `json:"branch"`
	Issue     int                        `json:"issue,omitempty"`
	Head      string                     `json:"head"`
	Workspace string                     `json:"workspace,omitempty"`
	Root      string                     `json:"root"`
	PID       int                        `json:"pid"`
	StartedAt time.Time                  `json:"started_at"`
	Extra     map[string]json.RawMessage `json:"-"`
}

func intentFields(in *Intent) map[string]any {
	return map[string]any{"version": &in.Version, "branch": &in.Branch, "issue": &in.Issue, "head": &in.Head,
		"workspace": &in.Workspace, "root": &in.Root, "pid": &in.PID, "started_at": &in.StartedAt}
}

// DecodeIntent parses an intent with the line's strict, preserving codec and
// validates it: the file sits in a shared directory, so nothing in it is
// trusted until it parses.
func DecodeIntent(data []byte) (Intent, error) {
	var in Intent
	extra, err := decodeStrict(data, intentFields(&in))
	if err != nil {
		return Intent{}, fmt.Errorf("intent: %w", err)
	}
	in.Extra = extra
	if err := in.Valid(); err != nil {
		return Intent{}, err
	}
	return in, nil
}

func (in Intent) MarshalJSON() ([]byte, error) {
	type known Intent
	base, err := json.Marshal(known(in))
	if err != nil {
		return nil, err
	}
	fields := intentFields(&in)
	return encodeWithExtra(base, in.Extra, func(k string) bool { _, ok := fields[k]; return ok }), nil
}

// Valid reports whether in is a well-formed intent.
func (in Intent) Valid() error {
	switch {
	case in.Version != IntentVersion:
		return fmt.Errorf("intent version %d, want %d", in.Version, IntentVersion)
	case !validBranch(in.Branch):
		return fmt.Errorf("intent branch %q is not a branch name", in.Branch)
	case !validHead(in.Head):
		return fmt.Errorf("intent head %q is not a commit id", in.Head)
	case !filepath.IsAbs(in.Root) || filepath.Clean(in.Root) != in.Root:
		return fmt.Errorf("intent root %q is not a clean absolute path", in.Root)
	case in.PID <= 0 || in.Issue < 0:
		return errors.New("intent pid or issue out of range")
	case in.StartedAt.IsZero():
		return errors.New("intent without a start time")
	}
	return nil
}

// validBranch is the subset of git's ref rules a re-entry command relies on:
// printable, no shell or revision syntax, not an option.
func validBranch(b string) bool {
	if b == "" || len(b) > 255 || strings.HasPrefix(b, "-") || strings.HasPrefix(b, "/") || strings.HasSuffix(b, "/") ||
		strings.HasSuffix(b, ".lock") || strings.Contains(b, "..") || strings.Contains(b, "//") || strings.Contains(b, "@{") {
		return false
	}
	for _, r := range b {
		if r <= ' ' || r == 0x7f || strings.ContainsRune(`~^:?*[\'"$`+"`;&|<>(){}!#", r) {
			return false
		}
	}
	return true
}

func validHead(h string) bool {
	if len(h) != 40 && len(h) != 64 {
		return false
	}
	_, err := hex.DecodeString(h)
	return err == nil && strings.ToLower(h) == h
}

// IntentState is what a reader can know about an intent from its lock and
// its branch.
type IntentState string

const (
	IntentLive     IntentState = "live"     // its merge holds the lock
	IntentOrphaned IntentState = "orphaned" // no merge holds it and the branch remains: re-enter
	IntentStale    IntentState = "stale"    // no merge holds it and the branch is gone: sweep
)

// ClassifyIntent is the whole reading rule.
func ClassifyIntent(locked, branchExists bool) IntentState {
	switch {
	case locked:
		return IntentLive
	case branchExists:
		return IntentOrphaned
	}
	return IntentStale
}

// Intents is a line directory's intents/ store.
type Intents struct{ dir string }

// OpenIntents returns the store under a landing line directory.
func OpenIntents(lineDir string) Intents { return Intents{dir: filepath.Join(lineDir, "intents")} }

func (s Intents) path(branch string) string {
	sum := sha256.Sum256([]byte(branch))
	return filepath.Join(s.dir, hex.EncodeToString(sum[:12]))
}

// IntentHeldError is Hold's refusal while another merge holds the branch's intent.
type IntentHeldError struct{ Intent Intent }

func (e *IntentHeldError) Error() string {
	if e.Intent.PID == 0 {
		return fmt.Sprintf("a merge of %s is already running; wait for it or stop it, then rerun", e.Intent.Branch)
	}
	return fmt.Sprintf("a merge of %s is already running (pid %d, since %s); wait for it or stop it, then rerun",
		e.Intent.Branch, e.Intent.PID, e.Intent.StartedAt.Format(time.RFC3339))
}

// IntentHold is a running merge's hold on its intent.
type IntentHold struct {
	path string
	file *os.File
}

// IntentHoldPatience bounds how long Hold waits out a brief lock before
// refusing: readers hold a shared lock only while they read, and a sweep an
// exclusive one only while it removes, so a lock that outlasts this is a
// running merge's.
var IntentHoldPatience = 2 * time.Second

// onHoldWait is a test seam: called each time Hold finds the lock taken and
// waits.
var onHoldWait = func() {}

// Hold records in and locks it for this process's lifetime. An orphan for
// the same branch is taken over in place, keeping a later sdlc's members;
// a live one refuses with *IntentHeldError. An unreadable record (corrupt,
// or a later sdlc's version) is never overwritten: Hold refuses, naming it.
func (s Intents) Hold(in Intent) (*IntentHold, error) {
	in.Version = IntentVersion
	if err := in.Valid(); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return nil, err
	}
	path := s.path(in.Branch)
	deadline := time.Now().Add(IntentHoldPatience)
	for {
		f, err := os.OpenFile(path, os.O_RDWR, 0)
		if errors.Is(err, os.ErrNotExist) {
			// A first record appears already locked and complete: it is
			// written under a temporary name, then linked into place.
			h, err := s.create(path, in)
			if errors.Is(err, os.ErrExist) {
				continue // another merge of this branch linked first
			}
			return h, err
		}
		if err != nil {
			return nil, err
		}
		if err := flock.Exclusive(f, true); err != nil {
			if time.Now().Before(deadline) {
				f.Close()
				onHoldWait()
				time.Sleep(20 * time.Millisecond)
				continue
			}
			prev, err := readIntent(f) // the holder may still be writing it
			if err != nil {
				prev = Intent{Branch: in.Branch}
			}
			f.Close()
			return nil, &IntentHeldError{Intent: prev}
		}
		if !samePath(f, path) { // a sweep removed it between open and lock
			f.Close()
			if time.Now().After(deadline) {
				return nil, errors.New("record merge intent: the intent file kept changing under its lock")
			}
			continue
		}
		prev, err := readIntent(f)
		if err == nil && prev.Branch != in.Branch {
			err = fmt.Errorf("it records branch %q", prev.Branch)
		}
		if err != nil {
			f.Close()
			return nil, fmt.Errorf("unreadable merge intent %q: %v; inspect it, then remove it and rerun", path, err)
		}
		in.Extra = mergeExtra(prev.Extra, in.Extra)
		data, err := json.Marshal(in)
		if err == nil {
			err = f.Truncate(0)
		}
		if err == nil {
			_, err = f.WriteAt(data, 0)
		}
		if err == nil {
			err = f.Sync()
		}
		if err != nil {
			f.Close()
			return nil, fmt.Errorf("record merge intent: %w", err)
		}
		return &IntentHold{path: path, file: f}, nil
	}
}

// create writes a first record to a locked temporary file and links it to
// path, so no reader ever finds it unlocked or incomplete. os.ErrExist means
// another merge linked first.
func (s Intents) create(path string, in Intent) (*IntentHold, error) {
	data, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}
	f, err := os.CreateTemp(s.dir, ".tmp-")
	if err != nil {
		return nil, err
	}
	if err = flock.Exclusive(f, true); err == nil {
		if _, err = f.Write(data); err == nil {
			err = f.Sync()
		}
	}
	if err == nil {
		err = os.Link(f.Name(), path)
	}
	os.Remove(f.Name())
	if err != nil {
		f.Close()
		return nil, err
	}
	return &IntentHold{path: path, file: f}, nil
}

// Release ends the hold. A terminal outcome removes the intent; an
// interrupted one (keep) only unlocks it, leaving the orphan to re-enter.
func (h *IntentHold) Release(keep bool) error {
	var err error
	if !keep {
		if err = os.Remove(h.path); errors.Is(err, os.ErrNotExist) {
			err = nil
		}
	}
	if cerr := h.file.Close(); err == nil {
		err = cerr
	}
	return err
}

// ObservedIntent is one intent as a reader found it. Err is set (and Intent
// zero) for a record that does not parse; it is reported, never acted on.
type ObservedIntent struct {
	Intent  Intent
	Path    string
	Locked  bool
	ModTime time.Time
	Err     error
}

// List observes every intent. A reader's shared lock fails only while a
// merge holds the intent exclusively, so readers never mistake each other
// for a running merge, nor block one for longer than a read.
func (s Intents) List() ([]ObservedIntent, error) {
	paths, err := s.files()
	var out []ObservedIntent
	for _, path := range paths {
		if o, ok := observeIntent(path); ok {
			out = append(out, o)
		}
	}
	return out, err
}

func (s Intents) files() ([]string, error) {
	entries, err := os.ReadDir(s.dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	var out []string
	for _, e := range entries {
		if e.Type().IsRegular() && !strings.HasPrefix(e.Name(), ".") {
			out = append(out, filepath.Join(s.dir, e.Name()))
		}
	}
	return out, err
}

func observeIntent(path string) (ObservedIntent, bool) {
	o := ObservedIntent{Path: path}
	f, err := os.Open(path)
	if err != nil {
		return o, false // gone meanwhile
	}
	defer f.Close()
	if info, err := f.Stat(); err == nil {
		o.ModTime = info.ModTime()
	}
	o.Locked = flock.Shared(f, true) != nil
	o.Intent, o.Err = readIntent(f)
	return o, true
}

// Sweep removes unlocked intents whose branch is gone (stale) or that are
// older than maxAge, and returns what it removed. It decides from a read,
// with no lock held across branchExists, then removes under a momentary
// exclusive lock only if the record is unchanged. Unreadable records are
// left for the operator: status names their path.
func (s Intents) Sweep(branchExists func(string) (bool, error), now time.Time, maxAge time.Duration) ([]Intent, error) {
	s.sweepTemps()
	paths, err := s.files()
	errs := []error{err}
	var removed []Intent
	for _, path := range paths {
		o, ok := observeIntent(path)
		if !ok || o.Locked || o.Err != nil {
			continue
		}
		exists, err := branchExists(o.Intent.Branch)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if ClassifyIntent(false, exists) != IntentStale && now.Sub(o.Intent.StartedAt) <= maxAge {
			continue
		}
		if removeUnchanged(path, o.Intent) {
			removed = append(removed, o.Intent)
		}
	}
	return removed, errors.Join(errs...)
}

// sweepTemps removes temporary records whose writer died between creating
// and linking them: a live writer holds its temp file's lock.
func (s Intents) sweepTemps() {
	entries, _ := os.ReadDir(s.dir)
	for _, e := range entries {
		if !strings.HasPrefix(e.Name(), ".tmp-") {
			continue
		}
		path := filepath.Join(s.dir, e.Name())
		if f, err := os.Open(path); err == nil {
			if flock.Exclusive(f, true) == nil && samePath(f, path) {
				os.Remove(path)
			}
			f.Close()
		}
	}
}

// removeUnchanged unlinks path only while no merge holds it and it still
// records the same run as seen, so a re-entry meanwhile is never deleted.
func removeUnchanged(path string, seen Intent) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	if flock.Exclusive(f, true) != nil || !samePath(f, path) {
		return false
	}
	now, err := readIntent(f)
	if err != nil || now.Branch != seen.Branch || now.PID != seen.PID || !now.StartedAt.Equal(seen.StartedAt) {
		return false
	}
	return os.Remove(path) == nil
}

func readIntent(f *os.File) (Intent, error) {
	data, err := io.ReadAll(io.NewSectionReader(f, 0, maxTicket+1))
	if err != nil {
		return Intent{}, err
	}
	if len(data) > maxTicket {
		return Intent{}, errors.New("intent: record too large")
	}
	if len(data) == 0 {
		return Intent{}, errors.New("intent: empty record")
	}
	return DecodeIntent(data)
}

// samePath reports whether f is still the file at path (not unlinked or
// replaced since it was opened).
func samePath(f *os.File, path string) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	pi, err := os.Stat(path)
	return err == nil && os.SameFile(fi, pi)
}
