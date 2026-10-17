package fleet

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flock"
)

// stagedRecordName is the per-worktree file `sdlc stage` (#344) writes when it
// detaches a checkout at another slot's commit for a smoke test. Since #421 it
// is also the stage's lease: it names the holder for the stage's lifetime.
const stagedRecordName = "sdlc-staged"

// StagedRecordPath is the staged record's path in the worktree's own git dir
// (never the shared one, so each slot has its own).
func StagedRecordPath(git GitReader, tree string) (string, error) {
	return gitPath(git, tree, stagedRecordName)
}

func gitPath(git GitReader, tree, name string) (string, error) {
	out, err := git.GitInDir(tree, "rev-parse", "--git-path", name)
	if err != nil {
		return "", err
	}
	p := gitLine(out)
	if !filepath.IsAbs(p) {
		p = filepath.Join(tree, p)
	}
	return p, nil
}

// StageLease is the staged record (#421). Holder "" is no holder: a legacy
// one-line record (a commit only) or a lease not yet taken.
type StageLease struct {
	Commit, Holder, Branch, Slot string
	At                           time.Time
}

var leaseKeys = []string{"commit", "holder", "branch", "slot", "at"}

// Format renders the lease as one `key value` line per field.
func (l StageLease) Format() string {
	at := ""
	if !l.At.IsZero() {
		at = l.At.Format(time.RFC3339)
	}
	var b strings.Builder
	for i, v := range []string{l.Commit, l.Holder, l.Branch, l.Slot, at} {
		fmt.Fprintf(&b, "%s %s\n", leaseKeys[i], v)
	}
	return b.String()
}

// ParseStageLease reads a staged record. A single token is a legacy record (a
// commit, no holder). Anything else must be exactly the five keys: a damaged
// lease is an error, never a free one.
func ParseStageLease(raw string) (StageLease, error) {
	text := strings.TrimSpace(raw)
	if text == "" {
		return StageLease{}, errors.New("the staged record is empty")
	}
	if !strings.ContainsAny(text, " \n") {
		return StageLease{Commit: text}, nil
	}
	lines := strings.Split(text, "\n")
	if len(lines) != len(leaseKeys) {
		return StageLease{}, fmt.Errorf("the staged record has %d lines, want %d", len(lines), len(leaseKeys))
	}
	vals := make([]string, len(lines))
	for i, line := range lines {
		key, val, _ := strings.Cut(line, " ")
		if key != leaseKeys[i] {
			return StageLease{}, fmt.Errorf("line %d of the staged record is %q, want key %q", i+1, line, leaseKeys[i])
		}
		vals[i] = strings.TrimSpace(val)
	}
	l := StageLease{Commit: vals[0], Holder: vals[1], Branch: vals[2], Slot: vals[3]}
	if vals[4] != "" {
		at, err := time.Parse(time.RFC3339, vals[4])
		if err != nil {
			return StageLease{}, fmt.Errorf("the staged record's time %q: %v", vals[4], err)
		}
		l.At = at
	}
	return l, nil
}

// ReadStageLease reads tree's lease; a missing record is the zero lease.
func ReadStageLease(git GitReader, tree string) (StageLease, error) {
	p, err := StagedRecordPath(git, tree)
	if err != nil {
		return StageLease{}, err
	}
	raw, err := os.ReadFile(p)
	if errors.Is(err, os.ErrNotExist) {
		return StageLease{}, nil
	}
	if err != nil {
		return StageLease{}, err
	}
	l, err := ParseStageLease(string(raw))
	if err != nil {
		return StageLease{}, fmt.Errorf("%s: %w", p, err)
	}
	return l, nil
}

// WriteStageLease replaces tree's lease atomically (a temp file, then rename).
func WriteStageLease(git GitReader, tree string, l StageLease) error {
	p, err := StagedRecordPath(git, tree)
	if err != nil {
		return err
	}
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, []byte(l.Format()), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

// LockStage takes the non-blocking flock that makes reading, deciding on and
// writing tree's lease one step (#421). The kernel drops it if the holder dies.
func LockStage(git GitReader, tree string) (unlock func(), err error) {
	p, err := gitPath(git, tree, stagedRecordName+".lock")
	if err != nil {
		return nil, err
	}
	f, err := os.OpenFile(p, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, err
	}
	if err := flock.Exclusive(f, true); err != nil {
		f.Close()
		return nil, fmt.Errorf("another sdlc stage or move is changing %s now (%v); rerun when it finishes", tree, err)
	}
	return func() { f.Close() }, nil
}

// StagedCommit is the commit sdlc last staged in tree ("" when none). A
// record is trusted only where it equals the detached HEAD exactly; a stale
// one (`:0` switched by hand) never vouches for another commit.
func StagedCommit(git GitReader, tree string) string {
	l, err := ReadStageLease(git, tree)
	if err != nil {
		return ""
	}
	return l.Commit
}
