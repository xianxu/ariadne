package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/affected"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/flakes"
)

// The missed-mapping ledger (#616) is one file per clone, in the common git dir, so every worktree's merges and
// trains record into the one list a curator reads.
const (
	missedLedgerFile  = "affected-missed.json"
	missedLedgerLock  = "affected-missed.lock"
	missedTempPrefix  = ".affected-missed-"
	missedTempMaxAge  = time.Hour
	missedRecordLimit = 2 * time.Minute
)

// missedLedgerStore is the ledger's thin IO shell: every operation is one read-modify-write under an exclusive flock,
// written by temporary file and rename, so concurrent writers never lose each other's entries and a crash leaves the
// previous file.
type missedLedgerStore struct{ dir string }

func missedStoreFor(root string) (missedLedgerStore, error) {
	out, err := execGitRunner{}.GitInDir(root, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return missedLedgerStore{}, fmt.Errorf("locate %s's git directory: %v", root, err)
	}
	return missedLedgerStore{dir: filepath.Join(strings.TrimSpace(string(out)), "sdlc")}, nil
}

func (s missedLedgerStore) path() string { return filepath.Join(s.dir, missedLedgerFile) }

// update applies op to the current ledger and writes the result; op's error leaves the file unchanged. A ledger that
// does not decode is refused, never overwritten: a curator repairs or removes it.
func (s missedLedgerStore) update(ctx context.Context, now time.Time, op func(*affected.MissedLedger) error) error {
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return err
	}
	lock, err := os.OpenFile(filepath.Join(s.dir, missedLedgerLock), os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return err
	}
	defer lock.Close() // closing releases the flock
	if err := acquireResumeFlock(ctx, lock, false); err != nil {
		return fmt.Errorf("lock the missed-mapping ledger: %w", err)
	}
	s.sweepTemps(now)
	l, err := s.read()
	if err != nil {
		return err
	}
	if err := op(&l); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(s.dir, missedTempPrefix+"*")
	if err != nil {
		return err
	}
	_, err = tmp.Write(append(raw, '\n'))
	if cerr := tmp.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		err = os.Rename(tmp.Name(), s.path())
	}
	if err != nil {
		os.Remove(tmp.Name())
	}
	return err
}

// read decodes the ledger; a missing file is an empty one.
func (s missedLedgerStore) read() (affected.MissedLedger, error) {
	raw, err := os.ReadFile(s.path())
	if errors.Is(err, fs.ErrNotExist) {
		return affected.MissedLedger{Version: affected.MissedLedgerVersion}, nil
	}
	if err != nil {
		return affected.MissedLedger{}, err
	}
	l, err := affected.DecodeMissedLedger(raw)
	if err != nil {
		return affected.MissedLedger{}, fmt.Errorf("%s: %w (repair or remove it)", s.path(), err)
	}
	return l, nil
}

// sweepTemps removes temporary files a crashed writer left, once they are older than any live write could be. It runs
// under the lock, so it never races a writer's rename.
func (s missedLedgerStore) sweepTemps(now time.Time) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if !strings.HasPrefix(e.Name(), missedTempPrefix) {
			continue
		}
		if info, err := e.Info(); err == nil && info.Mode().IsRegular() && now.Sub(info.ModTime()) > missedTempMaxAge {
			os.Remove(filepath.Join(s.dir, e.Name()))
		}
	}
}

// record stores misses under ev and reports what happened, for a gate's diagnostic line.
func (s missedLedgerStore) record(ctx context.Context, misses []affected.Miss, ev affected.MissEvidence) (string, error) {
	if len(misses) == 0 {
		return "missed mappings: none", nil
	}
	var recorded, dropped int
	err := s.update(ctx, time.Now(), func(l *affected.MissedLedger) error {
		recorded, dropped = l.Record(misses, ev)
		return nil
	})
	if err != nil {
		return "", err
	}
	line := fmt.Sprintf("missed mappings: recorded %d (sdlc affected missed list)", recorded)
	if dropped > 0 {
		line += fmt.Sprintf("; the ledger is at its cap, dropped the %d oldest", dropped)
	}
	return line, nil
}

// missedMergeRecorder is the merge gate's recorder (#616): when the merged head's suite fails where main's passes, it
// plans the head's delta from main as an affected run would (main's config and declarations over the head's graph)
// and records each introduced failure that selection would not have run. A repository without affected.json has
// nothing to miss. It runs under its own budget; the gate's outcome never depends on it.
func missedMergeRecorder(root string, registry func(context.Context) (flakes.Registry, error)) func(ctx context.Context, main, head string, introduced []string) (string, error) {
	return func(ctx context.Context, main, head string, introduced []string) (string, error) {
		ctx, cancel := context.WithTimeout(ctx, missedRecordLimit)
		defer cancel()
		r := boundedLandingGit(ctx, execGitRunner{})
		git := func(args ...string) (string, error) { return landingGit(r, root, args...) }
		paths, err := changedPathsBetween(git, main, head)
		if err != nil {
			return "", fmt.Errorf("changed paths: %w", err)
		}
		cleanupGit, retire := retiringCleanupGit(ctx, root)
		defer retire()
		at, cleanup, err := affectedSelectionAt(ctx, root, git, cleanupGit, main, head, paths)
		if cleanup != nil {
			defer cleanup()
		}
		if errors.Is(err, errAffectedUndeclared) {
			return "", nil
		}
		if err != nil {
			return "", err
		}
		if registry == nil {
			return "", errors.New("no flake registry")
		}
		reg, err := registry(withIssueRecordsScope(ctx))
		if err != nil {
			return "", fmt.Errorf("flake registry: %w", err)
		}
		misses := affected.Missed(at.sel, introduced, func(f string) bool { _, ok := reg.Cover(f); return ok })
		if len(misses) == 0 {
			return "", nil
		}
		store, err := missedStoreFor(root)
		if err != nil {
			return "", err
		}
		return store.record(ctx, misses, affected.MissEvidence{Paths: paths, Head: head, Main: main, Source: "merge", At: time.Now()})
	}
}
