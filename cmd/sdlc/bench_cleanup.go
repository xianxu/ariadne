package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/bench"
)

const benchRetainedRunLimit = 50
const benchRetentionChildLimit = 1000
const benchRetentionAge = 30 * 24 * time.Hour

// cleanBenchRun retires only disposable public material. Raw evidence and the
// external ledger stay intact; uncertainty is never translated into deletion.
func cleanBenchRun(ctx context.Context, s *benchStore, d benchCouchDeps) error {
	release, err := s.acquireRun()
	if err != nil {
		return err
	}
	defer release()
	r, state, err := s.read()
	if err != nil {
		return err
	}
	if state.Phase() == bench.PhaseCleaned {
		if err := guardBenchPublicAbsent(s); err != nil {
			return err
		}
		if err := guardBenchScoringCustody(s); err != nil {
			return err
		}
		return verifyBenchRetirement(ctx, s, d)
	}
	if !state.CanClean() {
		return fmt.Errorf("bench %s retained at %s; record a confirmed terminal run or reconcile exact owned processes before clean", s.root, state.Phase())
	}
	public := filepath.Join(s.root, "public")
	info, err := snapshotBenchPublic(s)
	if err != nil {
		return err
	}
	if err := guardBenchDeleteSource(r, public); err != nil {
		return err
	}
	if err := guardBenchScoringCustody(s); err != nil {
		return err
	}
	if err := verifyBenchRetirement(ctx, s, d); err != nil {
		return err
	}
	h, w := state.Bindings()
	op := "clean-" + r.ID
	_, effects, err := s.apply(bench.Event{ID: op + "-begin", RunID: r.ID, Kind: bench.BeginClean, OperationID: op, Hub: h, Worker: w})
	if err != nil {
		return err
	}
	if len(effects) != 1 || effects[0].Kind != bench.Clean {
		return errors.New("bench cleanup effect was not authorized")
	}
	removeErr := checkBenchPublicIdentity(s, info)
	if removeErr == nil {
		removeErr = guardBenchScoringCustody(s)
	}
	if removeErr == nil {
		removeErr = ctx.Err()
	}
	if removeErr == nil && info != nil {
		removeErr = os.RemoveAll(public)
	}
	kind := bench.CleanConfirmed
	reason := ""
	if removeErr != nil {
		kind = bench.CleanUnconfirmed
		reason = removeErr.Error()
	}
	_, _, stateErr := s.apply(bench.Event{ID: op + "-" + string(kind), RunID: r.ID, Kind: kind, OperationID: op, Hub: h, Worker: w, Reason: reason})
	if removeErr != nil || stateErr != nil {
		return fmt.Errorf("bench cleanup unconfirmed at %s; retain and inspect the exact clean operation: %w", s.root, errors.Join(removeErr, stateErr))
	}
	return nil
}

func snapshotBenchPublic(s *benchStore) (os.FileInfo, error) {
	if err := s.check(); err != nil {
		return nil, err
	}
	p := filepath.Join(s.root, "public")
	info, err := os.Lstat(p)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	physical, err := filepath.EvalSymlinks(p)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 || physical != p {
		return nil, fmt.Errorf("bench public path is substituted or not an owned directory: %s", p)
	}
	return info, nil
}

func checkBenchPublicIdentity(s *benchStore, expected os.FileInfo) error {
	actual, err := snapshotBenchPublic(s)
	if err != nil {
		return err
	}
	if expected == nil && actual == nil {
		return nil
	}
	if expected == nil || actual == nil || !os.SameFile(expected, actual) {
		return errors.New("bench public directory identity changed after cleanup admission")
	}
	return nil
}

func guardBenchPublicAbsent(s *benchStore) error {
	if err := s.check(); err != nil {
		return err
	}
	_, err := os.Lstat(filepath.Join(s.root, "public"))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	return fmt.Errorf("cleaned bench %s has recreated public material; reconcile it before retention", s.root)
}

func guardBenchDeleteSource(r benchRecord, deleting string) error {
	source, err := filepath.EvalSymlinks(r.Task.SourceRepo)
	if os.IsNotExist(err) {
		source = r.Task.SourceRepo
	} else if err != nil {
		return err
	}
	if benchPathWithin(deleting, source) || benchPathWithin(source, deleting) {
		return errors.New("bench deletion overlaps the original source repository")
	}
	return nil
}

// A cancelled scorer owns more than the retired agents. Its independently
// locked inputs must survive until that managed owner confirms its cleanup.
func guardBenchScoringCustody(s *benchStore) error {
	observer := filepath.Join(s.root, "observer")
	entries, err := readBenchDirectory(observer, benchRetentionChildLimit)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), "score-") {
			continue
		}
		dir := filepath.Join(observer, entry.Name())
		info, err := os.Lstat(dir)
		if err != nil {
			return err
		}
		physical, resolveErr := filepath.EvalSymlinks(dir)
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 || resolveErr != nil || physical != dir {
			return fmt.Errorf("unrecognized scoring custody at %s; reconcile before cleanup", dir)
		}
		for _, role := range []string{"parent", "landing", "candidate"} {
			for _, leaf := range []string{"anchor", "tree"} {
				p := filepath.Join(dir, role, leaf)
				_, err := os.Lstat(p)
				if err == nil {
					return fmt.Errorf("retained scorer inputs %s; reconcile the native scoring lease before cleanup", p)
				}
				if !os.IsNotExist(err) {
					return err
				}
			}
		}
	}
	return nil
}

func readBenchDirectory(root string, limit int) ([]os.DirEntry, error) {
	f, err := os.Open(root)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	entries, err := f.ReadDir(limit + 1)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}
	if len(entries) > limit {
		return nil, fmt.Errorf("bench directory %s exceeds %d immediate children; narrow or reconcile the inventory before admission", root, limit)
	}
	return entries, nil
}

type benchRetentionCandidate struct {
	root, id string
	phase    bench.Phase
	created  time.Time
}

func sweepBenchRuns(ctx context.Context, parent string, d benchCouchDeps) error {
	return sweepBenchRunsAt(ctx, parent, d, time.Now().UTC())
}

// CreatedAt is the existing record clock; no parallel registry or retention
// state is minted. Reserve one of the 50 slots for the next admitted run.
func sweepBenchRunsAt(ctx context.Context, parent string, d benchCouchDeps, now time.Time) error {
	physical, err := filepath.EvalSymlinks(parent)
	if err != nil || !filepath.IsAbs(parent) || filepath.Clean(parent) != parent || physical != parent || parent == string(filepath.Separator) {
		return errors.New("bench retention parent must be a canonical physical directory")
	}
	entries, err := readBenchDirectory(parent, benchRetentionChildLimit)
	if err != nil {
		return err
	}
	var candidates []benchRetentionCandidate
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return err
		}
		if !entry.IsDir() || entry.Type()&os.ModeSymlink != 0 {
			continue
		}
		root := filepath.Join(parent, entry.Name())
		s := &benchStore{root: root}
		r, state, err := s.read()
		if err != nil {
			continue
		} // Unknown folders are never adopted or removed.
		created, err := time.Parse(time.RFC3339Nano, r.CreatedAt)
		if err != nil {
			continue
		}
		candidates = append(candidates, benchRetentionCandidate{root: root, id: r.ID, phase: state.Phase(), created: created})
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].created.Equal(candidates[j].created) {
			return candidates[i].root < candidates[j].root
		}
		return candidates[i].created.Before(candidates[j].created)
	})
	remaining := len(candidates)
	removedRoots := map[string]bool{}
	for _, candidate := range candidates {
		if candidate.phase != bench.PhaseCleaned || (!candidate.created.Before(now.Add(-benchRetentionAge)) && remaining < benchRetainedRunLimit) {
			continue
		}
		removed, err := pruneBenchCleanedRun(ctx, candidate, d)
		if err != nil {
			return fmt.Errorf("bench retention kept %s; reconcile its exact evidence before admission: %w", candidate.root, err)
		}
		if removed {
			remaining--
			removedRoots[candidate.root] = true
		}
	}
	if remaining >= benchRetainedRunLimit {
		var guidance []string
		for _, c := range candidates {
			if !removedRoots[c.root] {
				guidance = append(guidance, fmt.Sprintf("%s (%s)", c.root, c.phase))
				if len(guidance) == 5 {
					break
				}
			}
		}
		return fmt.Errorf("bench admission refused: %d retained valid runs reach the %d-run cap; clean recorded runs or reconcile active/uncertain custody first: %s", remaining, benchRetainedRunLimit, strings.Join(guidance, "; "))
	}
	return nil
}

func pruneBenchCleanedRun(ctx context.Context, c benchRetentionCandidate, d benchCouchDeps) (bool, error) {
	s := &benchStore{root: c.root}
	release, err := s.acquireRun()
	if err != nil {
		return false, nil
	}
	defer release() // A busy owner is never displaced.
	transaction, err := s.lock("transaction.lock")
	if err != nil {
		return false, nil
	}
	defer transaction()
	r, state, err := s.read()
	if err != nil {
		return false, err
	}
	if r.ID != c.id || state.Phase() != bench.PhaseCleaned {
		return false, nil
	}
	info, err := os.Lstat(s.root)
	if err != nil {
		return false, err
	}
	if err := guardBenchPublicAbsent(s); err != nil {
		return false, err
	}
	if err := guardBenchScoringCustody(s); err != nil {
		return false, err
	}
	if err := guardBenchDeleteSource(r, s.root); err != nil {
		return false, err
	}
	if err := verifyBenchRetirement(ctx, s, d); err != nil {
		return false, err
	}
	// The fresh native census is an external boundary. Recheck filesystem
	// custody after it returns, rather than deleting material created while
	// its response was in flight.
	if err := guardBenchPublicAbsent(s); err != nil {
		return false, err
	}
	if err := guardBenchScoringCustody(s); err != nil {
		return false, err
	}
	if err := guardBenchDeleteSource(r, s.root); err != nil {
		return false, err
	}
	if err := ctx.Err(); err != nil {
		return false, err
	}
	current, err := os.Lstat(s.root)
	if err != nil {
		return false, err
	}
	if !current.IsDir() || current.Mode()&os.ModeSymlink != 0 || !os.SameFile(info, current) {
		return false, errors.New("bench run directory identity changed during retention admission")
	}
	if err := s.check(); err != nil {
		return false, err
	}
	if err := os.RemoveAll(s.root); err != nil {
		return false, err
	}
	return true, nil
}
