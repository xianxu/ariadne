// localcheck_lease.go — a local-check run's lease (#400): the record that lets
// `sdlc local-check cancel` find and stop only its own checkout's runs. The run
// holds an exclusive flock on its lease for its whole life, so a held lock
// proves the owner is alive; ownership is never inferred from the process table.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flock"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/processgroup"
)

// runLease is what a run records about itself.
type runLease struct {
	Root  string `json:"root"` // the checkout the run was started from
	Check string `json:"check"`
	SHA   string `json:"sha"`
	PID   int    `json:"pid"`
	// PIDStarted is the process's start time as processStartTime reads it; cancel
	// compares it before signalling, so a reused PID is not mistaken for the run.
	PIDStarted string    `json:"pid_started"`
	Started    time.Time `json:"started"`
}

type leaseVerdict int

const (
	leaseStale      leaseVerdict = iota // no holder: the run is gone
	leaseForeign                        // another checkout's live run
	leaseOtherCheck                     // ours, but not the check named
	leaseTarget                         // ours to stop
)

// classifyRunLease decides what cancel may do with one lease.
func classifyRunLease(rec runLease, held bool, root, name string) leaseVerdict {
	switch {
	case !held:
		return leaseStale
	case rec.Root != root:
		return leaseForeign
	case name != "" && rec.Check != name:
		return leaseOtherCheck
	}
	return leaseTarget
}

// localCheckRunsDir holds the leases: beside the per-head logs, never among
// them, so log pruning cannot remove a live run's lease.
func localCheckRunsDir(logDir string) string {
	return filepath.Join(filepath.Dir(logDir), "local-check-runs")
}

// processStartTime reads a process's start time from the kernel (no ps: it
// would need a subprocess a sandboxed agent may not be allowed to run).
var processStartTime = processgroup.StartTime

// canonicalRoot makes two spellings of one checkout compare equal.
func canonicalRoot(root string) string {
	if r, err := filepath.EvalSymlinks(root); err == nil {
		return r
	}
	return filepath.Clean(root)
}

// holdRunLease publishes rec in dir, already locked: it is written and locked
// under a temporary name that cancel ignores and only then renamed into place,
// so no reader ever sees an unlocked lease of a live run. The release removes it.
func holdRunLease(dir string, rec runLease) (string, func(), error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", nil, err
	}
	f, err := os.CreateTemp(dir, ".lease-*.tmp")
	if err != nil {
		return "", nil, err
	}
	abandon := func(err error) (string, func(), error) {
		_ = os.Remove(f.Name())
		_ = f.Close()
		return "", nil, err
	}
	if err := flock.Exclusive(f, true); err != nil {
		return abandon(err)
	}
	if err := json.NewEncoder(f).Encode(rec); err != nil {
		return abandon(err)
	}
	id := strings.TrimSuffix(strings.TrimPrefix(filepath.Base(f.Name()), "."), ".tmp")
	final := filepath.Join(dir, id+".json")
	if err := os.Rename(f.Name(), final); err != nil {
		return abandon(err)
	}
	var once sync.Once
	return id, func() {
		once.Do(func() {
			_ = os.Remove(final)
			_ = f.Close()
		})
	}, nil
}

// openRunLease observes the existing lifetime lock without taking authority
// from an unlocked record. Both notices and cancellation use this same reader.
func openRunLease(p string) (*os.File, runLease, bool, error) {
	var rec runLease
	f, err := os.OpenFile(p, os.O_RDWR, 0)
	if err != nil {
		return nil, rec, false, err
	}
	lockErr := flock.Exclusive(f, true)
	held := errors.Is(lockErr, syscall.EWOULDBLOCK)
	if lockErr != nil && !held {
		f.Close()
		return nil, rec, false, lockErr
	}
	if err := json.NewDecoder(f).Decode(&rec); err != nil && held {
		f.Close()
		return nil, rec, false, err
	}
	return f, rec, held, nil
}

func noticeDuplicateRunLeases(dir, id string, current runLease, out io.Writer, read func(string) (*os.File, runLease, bool, error)) {
	paths, _ := filepath.Glob(filepath.Join(dir, "*.json"))
	for _, p := range paths {
		f, rec, held, err := read(p)
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				fmt.Fprintf(out, "[!] lease %s: %v; duplicate check unavailable\n", p, err)
			}
			continue
		}
		f.Close()
		if classifyRunLease(rec, held, current.Root, current.Check) == leaseTarget && rec.SHA != current.SHA {
			existing := strings.TrimSuffix(filepath.Base(p), ".json")
			fmt.Fprintf(out, "[!] local check %s run %s on %s also has existing run %s on %s; to replace it explicitly: sdlc local-check cancel --run %s\n", current.Check, id, abbrevSHA(current.SHA), existing, abbrevSHA(rec.SHA), existing)
		}
	}
}

// sweepStaleRunLeases removes the leases of runs that died without releasing
// them (no holder), and temporary names a run killed before its rename left
// behind. Every run and every cancel sweeps, so the directory holds about one
// file per live run.
func sweepStaleRunLeases(dir string) {
	paths, _ := filepath.Glob(filepath.Join(dir, "*.json"))
	for _, p := range paths {
		f, err := os.OpenFile(p, os.O_RDWR, 0)
		if err != nil {
			continue
		}
		if flock.Exclusive(f, true) == nil {
			_ = os.Remove(p) // removed while locked: no run holds it
		}
		_ = f.Close()
	}
	tmps, _ := filepath.Glob(filepath.Join(dir, ".lease-*.tmp"))
	for _, p := range tmps {
		if info, err := os.Stat(p); err == nil && time.Since(info.ModTime()) > localCheckStaleAge {
			_ = os.Remove(p)
		}
	}
}

// cancelLocalChecks selects one runID, one check name, or all this checkout's
// live runs, sweeps leases whose run is gone, and waits up to wait for each
// stopped run to release its lease. Another checkout's runs are never touched.
func cancelLocalChecks(dir, root, name, runID string, out io.Writer, wait time.Duration) (int, error) {
	if runID != "" && (strings.ContainsAny(runID, "/\\") || runID == "." || runID == "..") {
		return 0, fmt.Errorf("local-check cancel: invalid run ID %q", runID)
	}
	root = canonicalRoot(root)
	sweepStaleRunLeases(dir)
	paths, _ := filepath.Glob(filepath.Join(dir, "*.json"))
	if runID != "" {
		paths = []string{filepath.Join(dir, runID+".json")}
	}
	type stopping struct {
		f   *os.File
		rec runLease
	}
	var stops []stopping
	foreign := 0
	for _, p := range paths {
		f, rec, held, err := openRunLease(p)
		if err != nil {
			if runID != "" {
				return 0, fmt.Errorf("local-check cancel: run %q unavailable: %w", runID, err)
			}
			if !errors.Is(err, os.ErrNotExist) {
				fmt.Fprintf(out, "[!] lease %s: %v; left alone\n", p, err)
			}
			continue
		}
		switch classifyRunLease(rec, held, root, name) {
		case leaseStale:
			if runID != "" {
				f.Close()
				return 0, fmt.Errorf("local-check cancel: run %q is no longer live", runID)
			}
			_ = os.Remove(p)
		case leaseForeign:
			if runID != "" {
				f.Close()
				return 0, fmt.Errorf("local-check cancel: run %q belongs to another checkout: %s", runID, rec.Root)
			}
			foreign++
		case leaseTarget:
			if rec.PID <= 1 { // 0 and -1 would signal a group or every process
				fmt.Fprintf(out, "[!] lease %s names pid %d; not signalled\n", p, rec.PID)
				break
			}
			if started, err := processStartTime(rec.PID); err != nil || started == "" || started != rec.PIDStarted {
				fmt.Fprintf(out, "[!] lease %s names pid %d, which is no longer that run; not signalled\n", p, rec.PID)
				break
			}
			if err := syscall.Kill(rec.PID, syscall.SIGTERM); err != nil {
				fmt.Fprintf(out, "[!] local check %s (pid %d): %v\n", rec.Check, rec.PID, err)
				break
			}
			stops = append(stops, stopping{f, rec})
			continue // keep f open to watch its lock
		}
		_ = f.Close()
	}
	if foreign > 0 {
		fmt.Fprintf(out, "    %d run(s) of other checkouts left running\n", foreign)
	}
	if len(stops) == 0 {
		fmt.Fprintf(out, "==> no local-check run of %s to cancel\n", root)
		return 0, nil
	}
	deadline := time.Now().Add(wait)
	var stuck []string
	for _, s := range stops {
		released := flock.Exclusive(s.f, true) == nil
		for !released && time.Now().Before(deadline) {
			time.Sleep(50 * time.Millisecond)
			released = flock.Exclusive(s.f, true) == nil
		}
		_ = s.f.Close()
		if !released {
			stuck = append(stuck, fmt.Sprintf("%s (pid %d)", s.rec.Check, s.rec.PID))
			continue
		}
		fmt.Fprintf(out, "==> cancelled local check %s on %s (pid %d)\n", s.rec.Check, abbrevSHA(s.rec.SHA), s.rec.PID)
	}
	if len(stuck) > 0 {
		return len(stops) - len(stuck), fmt.Errorf("local-check cancel: still running after %s: %s", wait, strings.Join(stuck, ", "))
	}
	return len(stops), nil
}
