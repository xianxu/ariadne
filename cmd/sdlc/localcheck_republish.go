// localcheck_republish.go — publish a completed local check's saved result
// without rerunning it (#431). A run saves its attempt (teststats_store.go)
// before it posts; when the post cannot reach GitHub, the result is still on
// disk, bound to its exact commit and log bytes.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// suiteStoreRoot is the attempt store beside the per-head logs.
func suiteStoreRoot(logDir string) string {
	return filepath.Join(filepath.Dir(logDir), "test-stats")
}

// savedAttemptRecord turns a saved attempt into the record its run would have
// posted, refusing unless it is this repository's complete attempt of this
// check on this exact commit and its log still hashes to the recorded value.
func savedAttemptRecord(a SuiteAttempt, logSHA string, tail []string, repo, sha, check, host string) (localRunRecord, error) {
	switch {
	case a.Repo != repo:
		return localRunRecord{}, fmt.Errorf("attempt %s is for %s, not %s", a.ID, a.Repo, repo)
	case a.SHA != sha:
		return localRunRecord{}, fmt.Errorf("attempt %s is for commit %s, not %s", a.ID, abbrevSHA(a.SHA), abbrevSHA(sha))
	case a.Check != check:
		return localRunRecord{}, fmt.Errorf("attempt %s is check %s, not %s", a.ID, a.Check, check)
	case a.MainBaseline:
		return localRunRecord{}, fmt.Errorf("attempt %s is a diagnostic baseline on main, never recorded", a.ID)
	case a.WallSeconds == nil || a.LogSHA256 == "":
		return localRunRecord{}, fmt.Errorf("attempt %s is incomplete (no duration or log hash)", a.ID)
	case logSHA != a.LogSHA256:
		return localRunRecord{}, fmt.Errorf("attempt %s: its log no longer hashes to the recorded sha256 %s", a.ID, abbrevSHA(a.LogSHA256))
	}
	return localRunRecord{Name: check, SHA: sha, Exit: a.Exit, TimedOut: a.TimedOut,
		Duration: time.Duration(*a.WallSeconds * float64(time.Second)), LogSHA256: a.LogSHA256, Host: host, Tail: tail}, nil
}

var (
	errNoSavedAttempt      = errors.New("no saved attempt")
	errSavedAttemptRefused = errors.New("no saved attempt verifies")
)

func hostName() string {
	h, _ := os.Hostname()
	return h
}

// loadSavedAttempt finds the attempt to publish: the one named by id, else the
// newest of this check on this commit whose log still verifies. newest reports
// whether no later attempt of it exists (one that failed to verify, say) and the
// store was read whole, which merge requires before a saved pass stands in.
func loadSavedAttempt(storeRoot, repo, sha, check, id, host string) (rec localRunRecord, newest bool, err error) {
	attempts, warnings, err := readSuiteAttempts(storeRoot, testStatsRecordLimit)
	if err != nil {
		return localRunRecord{}, false, err
	}
	// An incomplete read (a limit hit, an unreadable record) leaves the newest
	// attempt unknown, not absent: nothing then counts as newest.
	complete := len(warnings) == 0
	var best *localRunRecord
	var bestAt, latest time.Time
	var refusals []string
	for _, a := range attempts {
		if id != "" && a.ID != id || id == "" && (a.SHA != sha || a.Check != check) {
			continue
		}
		if a.FinishedAt.After(latest) {
			latest = a.FinishedAt
		}
		path := suiteAttemptLogFile(storeRoot, a)
		if path == "" {
			refusals = append(refusals, "attempt "+a.ID+" names no log")
			continue
		}
		logSHA, tail, err := hashAndTail(path, localCheckTailLines)
		if err != nil {
			refusals = append(refusals, fmt.Sprintf("attempt %s: log %s: %v", a.ID, path, err))
			continue
		}
		rec, err := savedAttemptRecord(a, logSHA, tail, repo, sha, check, host)
		if err != nil {
			refusals = append(refusals, err.Error())
			continue
		}
		if best == nil || a.FinishedAt.After(bestAt) {
			rec.LogPath = path
			best, bestAt = &rec, a.FinishedAt
		}
	}
	if best != nil {
		return *best, complete && !latest.After(bestAt), nil
	}
	if len(refusals) > 0 {
		return localRunRecord{}, false, fmt.Errorf("%w for %s on %s: %s", errSavedAttemptRefused, check, abbrevSHA(sha), strings.Join(refusals, "; "))
	}
	return localRunRecord{}, false, fmt.Errorf("%w for %s on %s", errNoSavedAttempt, check, abbrevSHA(sha))
}

// republishCommand is the recovery a failed post names.
func republishCommand(check, sha string) string {
	return fmt.Sprintf("sdlc local-check republish --name %s --rev %s", check, sha)
}

// localCheckPostBackoff spaces the status post's retries: the direct path to
// GitHub has outages (#431). The suite slot is already released by then; under
// merge the wait holds the landing turn, which still beats a 12-minute rerun.
// Tests shrink it.
var localCheckPostBackoff = []time.Duration{15 * time.Second, 45 * time.Second, 90 * time.Second}

// postLocalCheckRecord posts rec as its commit status, retrying through
// localCheckPostBackoff, then a failed run's PR comment (best effort).
func postLocalCheckRecord(ctx context.Context, d localRunDeps, rec localRunRecord) error {
	err := d.GH.PostStatus(ctx, d.Repo, rec.SHA, localStatusContext(rec.Name), rec.state(), rec.statusDescription())
	for _, wait := range localCheckPostBackoff {
		if err == nil {
			break
		}
		fmt.Fprintf(d.Stderr, "[!] local check %s: status post failed (%v); retrying in %s\n", rec.Name, err, wait)
		select {
		case <-ctx.Done():
			return fmt.Errorf("%w (then %w)", err, ctx.Err())
		case <-time.After(wait):
		}
		err = d.GH.PostStatus(ctx, d.Repo, rec.SHA, localStatusContext(rec.Name), rec.state(), rec.statusDescription())
	}
	if err != nil {
		return err
	}
	if !rec.passed() && d.PR > 0 {
		if err := d.GH.PRComment(ctx, d.Repo, d.PR, rec.failureComment()); err != nil {
			fmt.Fprintf(d.Stderr, "[!] local check %s: failure recorded, but the PR comment failed: %v\n", rec.Name, err)
		}
	}
	return nil
}

// republishSavedPass is merge's first resort for a missing check: a saved,
// verified pass of this check on this exact head is posted instead of rerun.
// ok is false when there is none (merge then runs the check).
func republishSavedPass(ctx context.Context, d localRunDeps, check, sha string) (localRunRecord, bool, error) {
	logDir, err := localCheckLogDir(d)
	if err != nil {
		return localRunRecord{}, false, nil
	}
	host := hostName()
	rec, newest, err := loadSavedAttempt(suiteStoreRoot(logDir), d.Repo, sha, check, "", host)
	switch {
	case errors.Is(err, errSavedAttemptRefused):
		fmt.Fprintf(d.Stderr, "[!] local check %s: not reusing a saved attempt: %v\n", check, err)
	case err == nil && rec.passed() && !newest:
		fmt.Fprintf(d.Stderr, "[!] local check %s: not reusing the saved pass on %s: a later attempt exists and does not verify\n", check, abbrevSHA(sha))
	}
	if err != nil || !rec.passed() || !newest {
		return localRunRecord{}, false, nil
	}
	fmt.Fprintf(d.Stderr, "==> local check %s: publishing the saved pass on %s instead of rerunning (log %s)\n", check, abbrevSHA(sha), rec.LogPath)
	if err := postLocalCheckRecord(ctx, d, rec); err != nil {
		return rec, true, fmt.Errorf("local check %s: the saved pass on %s was NOT recorded: %w; publish it with %q (no rerun needed)", check, abbrevSHA(sha), err, republishCommand(check, sha))
	}
	fmt.Fprintf(d.Stderr, "==> local check %s: saved pass recorded as %s on %s\n", check, localStatusContext(check), abbrevSHA(sha))
	return rec, true, nil
}
