package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
)

// landingConfigPath is the landing train's setting (#449), read from main's
// tree like the local-check declaration: a branch never sets its own train.
// It sits under .sdlc/, which local-gates-guard protects, because the
// interval decides how often the authoritative suite runs.
const landingConfigPath = ".sdlc/landing.json"

// landingConfig is the file's shape. Absent, or an interval of 0, is the
// serial line of #393. Unknown members are ignored: sdlc only reads this file.
type landingConfig struct {
	TrainIntervalMinutes *int `json:"train_interval_minutes"`
	TrainMaxCars         *int `json:"train_max_cars"`
	// TrainMaxWaitMinutes bounds a ready car's wait (#475): a train leaves
	// once the oldest ready car has waited this long. Absent or 0: no bound.
	TrainMaxWaitMinutes *int `json:"train_max_wait_minutes"`
}

const (
	maxTrainInterval = 24 * 60
	maxTrainCars     = 16
)

// parseLandingConfig turns the file into a schedule, refusing out-of-range
// values with the path named rather than clamping them.
func parseLandingConfig(data []byte) (landingline.Schedule, error) {
	fail := func(format string, a ...any) (landingline.Schedule, error) {
		return landingline.Schedule{}, fmt.Errorf("%s: "+format, append([]any{landingConfigPath}, a...)...)
	}
	var c landingConfig
	dec := json.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(&c); err != nil {
		return fail("not a JSON object: %v", err)
	}
	if dec.More() {
		return fail("trailing data after the object")
	}
	s := landingline.Schedule{MaxCars: landingline.DefaultMaxCars}
	if v := c.TrainIntervalMinutes; v != nil {
		if *v < 0 || *v > maxTrainInterval {
			return fail("train_interval_minutes %d is outside 0..%d", *v, maxTrainInterval)
		}
		s.Interval = time.Duration(*v) * time.Minute
	}
	if v := c.TrainMaxCars; v != nil {
		if *v < 1 || *v > maxTrainCars {
			return fail("train_max_cars %d is outside 1..%d", *v, maxTrainCars)
		}
		s.MaxCars = *v
	}
	if v := c.TrainMaxWaitMinutes; v != nil {
		if *v < 0 || *v > maxTrainInterval {
			return fail("train_max_wait_minutes %d is outside 0..%d", *v, maxTrainInterval)
		}
		s.MaxWait = time.Duration(*v) * time.Minute
	}
	return s, nil
}

// readLandingSchedule reads the train setting from rev's tree; no file is the
// serial line.
func readLandingSchedule(git func(args ...string) (string, error), rev string) (landingline.Schedule, error) {
	out, err := git("ls-tree", "--name-only", rev, landingConfigPath)
	if err != nil {
		return landingline.Schedule{}, fmt.Errorf("read %s's tree for %s: %w", rev, landingConfigPath, err)
	}
	if strings.TrimSpace(out) == "" {
		return landingline.Schedule{MaxCars: landingline.DefaultMaxCars}, nil
	}
	data, err := git("show", rev+":"+landingConfigPath)
	if err != nil {
		return landingline.Schedule{}, fmt.Errorf("read %s on %s: %w", landingConfigPath, rev, err)
	}
	return parseLandingConfig([]byte(data))
}

// effectiveSchedule is the one schedule every consumer acts on (#475):
// main's train setting, placed on the machine plan's grid when the plan has
// an entry for repo. planNote says which plan applied (or why none did), for
// status; a plan that cannot be read is planErr, and the schedule keeps the
// interval. A setting that cannot be read is err.
func effectiveSchedule(git func(args ...string) (string, error), mainRef, repo string) (s landingline.Schedule, planNote string, planErr, err error) {
	s, err = readLandingSchedule(git, mainRef)
	if err != nil || s.Off() {
		return s, "", nil, err
	}
	path := machinePlanPathFn()
	p, perr := readMachinePlan(path)
	switch {
	case errors.Is(perr, fs.ErrNotExist):
		return s, "no machine plan (" + path + "): departing one interval after the last train", nil, nil
	case perr != nil:
		return s, "machine plan unreadable: departing one interval after the last train", perr, nil
	}
	off, ok := p.Offset(repo, s.Interval)
	if !ok {
		return s, "the machine plan has no entry for " + repo + ": departing one interval after the last train", nil, nil
	}
	s.Offset = &off
	return s, fmt.Sprintf("machine plan: slot :%02d", int(off/time.Minute)), nil, nil
}

// machinePlanPathFn locates the plan; tests point it at their own state.
var machinePlanPathFn = machinePlanPath

// machinePlanPath is the machine plan's file (#475): beside the suite lock,
// which scripts/test-lock.py puts at $WF_TEST_LOCK_FILE, else
// ${XDG_STATE_HOME:-~/.local/state}/ariadne/test-suite.lock.
func machinePlanPath() string {
	if lock := os.Getenv("WF_TEST_LOCK_FILE"); lock != "" {
		return filepath.Join(filepath.Dir(lock), "landing-plan.json")
	}
	state := os.Getenv("XDG_STATE_HOME")
	if state == "" {
		home, _ := os.UserHomeDir()
		state = filepath.Join(home, ".local", "state")
	}
	return filepath.Join(state, "ariadne", "landing-plan.json")
}

// readMachinePlan reads and checks the plan file; a missing file is
// fs.ErrNotExist (no plan: today's clock), anything else names the file.
func readMachinePlan(path string) (landingline.MachinePlan, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return landingline.MachinePlan{}, err
	}
	p, err := landingline.DecodePlan(data)
	if err != nil {
		return landingline.MachinePlan{}, fmt.Errorf("%s: %w", path, err)
	}
	return p, nil
}
