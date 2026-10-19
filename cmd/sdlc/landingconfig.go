package main

import (
	"bytes"
	"encoding/json"
	"fmt"
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
