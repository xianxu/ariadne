// Package freshness is the one definition of "this branch has caught up with
// main" (ariadne-robustness-1 requirement 6): the branch contains every
// first-parent commit of main that landed at or before the most recent noon in
// the repository's timezone. Close and milestone-close apply it before a
// boundary review (#269); start-plan applies it when a catch-up before new work
// fails (#303).
//
// A fixed past cutoff is deterministic and does not chase a moving main: a push
// after the cutoff cannot move the requirement, and integrating once a day
// always satisfies it.
package freshness

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
)

// ConfigPath is the repository's freshness configuration, relative to its root.
const ConfigPath = ".sdlc/freshness.json"

// Cutoff is the most recent noon at or before now, in loc. Pure.
func Cutoff(now time.Time, loc *time.Location) time.Time {
	local := now.In(loc)
	noon := time.Date(local.Year(), local.Month(), local.Day(), 12, 0, 0, 0, loc)
	if noon.After(local) {
		noon = time.Date(local.Year(), local.Month(), local.Day()-1, 12, 0, 0, 0, loc)
	}
	return noon
}

type config struct {
	Version  int    `json:"version"`
	Timezone string `json:"timezone"`
}

// LoadLocation reads the repository's timezone from ConfigPath under root. An
// absent file means the machine's local zone. A file that cannot be read or
// names an unknown zone is an error: a wrong zone silently moves the cutoff, so
// it must not fall back. source names where the zone came from, for messages.
func LoadLocation(root string) (loc *time.Location, source string, err error) {
	path := filepath.Join(root, ConfigPath)
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return time.Local, "the local zone; no " + ConfigPath, nil
	}
	if err != nil {
		return nil, "", fmt.Errorf("read %s: %w", ConfigPath, err)
	}
	var c config
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, "", fmt.Errorf("parse %s: %w", ConfigPath, err)
	}
	if c.Version != 1 || c.Timezone == "" {
		return nil, "", fmt.Errorf("%s: want {\"version\":1,\"timezone\":\"<IANA zone>\"}", ConfigPath)
	}
	loc, err = time.LoadLocation(c.Timezone)
	if err != nil {
		return nil, "", fmt.Errorf("%s: timezone %q: %w", ConfigPath, c.Timezone, err)
	}
	return loc, ConfigPath, nil
}

// State is the outcome of a freshness probe.
type State int

const (
	// Inconclusive: the probe could not decide (no main ref, a git error). Never
	// to be reported as up to date.
	Inconclusive State = iota
	// Fresh: the branch contains main as of the cutoff.
	Fresh
	// DocsOnly: the branch lacks main's changes before the cutoff, but none of
	// them has code surface.
	DocsOnly
	// Stale: the branch lacks code changes main had before the cutoff.
	Stale
)

// Result is a probe's verdict and the facts behind it.
type Result struct {
	State    State
	Cutoff   time.Time
	Required string   // newest first-parent commit of main at or before Cutoff ("" when none)
	Code     []string // missing paths with code surface (Stale)
	Docs     int      // missing paths without code surface
	Why      string   // what made the probe Inconclusive
}

// Git runs `git <args>` and returns stdout. A non-zero exit is an error whose
// chain carries the *exec.ExitError, so callers can read the exit code.
type Git func(args ...string) ([]byte, error)

// Check decides whether head contains mainRef as of cutoff. isCode classifies
// paths as code surface (the caller owns that predicate, so the publish gate and
// this probe agree on what counts as code).
func Check(git Git, mainRef, head string, cutoff time.Time, isCode func([]string) bool) Result {
	r := Result{Cutoff: cutoff}
	inconclusive := func(what string, err error) Result {
		r.State, r.Why = Inconclusive, fmt.Sprintf("%s: %v", what, err)
		return r
	}
	out, err := git("rev-list", "-n1", "--first-parent", "--before="+strconv.FormatInt(cutoff.Unix(), 10), mainRef, "--")
	if err != nil {
		return inconclusive("read "+mainRef, err)
	}
	r.Required = strings.TrimSpace(string(out))
	if r.Required == "" {
		r.State = Fresh // nothing on main is old enough to require
		return r
	}
	switch _, err := git("merge-base", "--is-ancestor", r.Required, head); {
	case err == nil:
		r.State = Fresh
		return r
	case exitCode(err) != 1:
		return inconclusive("ancestry of "+short(r.Required), err)
	}
	base, err := git("merge-base", r.Required, head)
	if err != nil {
		return inconclusive("merge-base with "+short(r.Required), err)
	}
	names, err := git(append(gitx.ChangedPathArgs(strings.TrimSpace(string(base)), r.Required), "--")...)
	if err != nil {
		return inconclusive("diff to "+short(r.Required), err)
	}
	for _, p := range strings.Split(string(names), "\x00") {
		switch {
		case p == "":
		case isCode([]string{p}):
			r.Code = append(r.Code, p)
		default:
			r.Docs++
		}
	}
	switch {
	case len(r.Code) > 0:
		r.State = Stale
	case r.Docs > 0:
		r.State = DocsOnly
	default:
		r.State = Fresh // the missing landings net to no change (say, landed and reverted)
	}
	return r
}

func exitCode(err error) int {
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return ee.ExitCode()
	}
	return -1
}

func short(sha string) string {
	if len(sha) > 8 {
		return sha[:8]
	}
	return sha
}
