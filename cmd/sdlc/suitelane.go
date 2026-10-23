// suitelane.go — the short-suite light lane (#559): which recorded suites may
// run in the machine's one light slot, and the order of one run that tries it.
// Eligibility is measured from the repo's own recorded attempts, never declared.
package main

import (
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"
)

const (
	suiteLaneLight     = "light"
	suiteLaneSamples   = 3                 // newest recorded suites that must all be short
	suiteLaneThreshold = 120 * time.Second // each of them ran under this
	suiteLaneBound     = 240 * time.Second // a light run is stopped here and retried normally
	suiteLaneNice      = "10"
)

// suiteLaneEligibility is the measurement behind a lane decision; Reason
// names it for the run's announcement.
type suiteLaneEligibility struct {
	Eligible bool
	Samples  []float64 // wall seconds, newest first
	Reason   string
}

// laneEligibility reads repo's newest recorded suite attempts. A run stopped
// in the light lane leaves its overrun on the attempt that retried it, so the
// repo leaves the lane by measurement and returns after short runs push it out.
func laneEligibility(attempts []SuiteAttempt, repo string) suiteLaneEligibility {
	var suites []SuiteAttempt
	for _, a := range attempts {
		if a.Repo == repo && a.Check == suiteCheckName && !a.MainBaseline && !a.Legacy && a.WallSeconds != nil {
			suites = append(suites, a)
		}
	}
	sort.SliceStable(suites, func(i, j int) bool { return suites[i].FinishedAt.After(suites[j].FinishedAt) })
	if len(suites) < suiteLaneSamples {
		return suiteLaneEligibility{Reason: fmt.Sprintf("needs %d recorded suites, has %d", suiteLaneSamples, len(suites))}
	}
	e := suiteLaneEligibility{Eligible: true}
	var walls []string
	for _, a := range suites[:suiteLaneSamples] {
		wall := *a.WallSeconds
		e.Samples = append(e.Samples, wall)
		walls = append(walls, fmt.Sprintf("%.0fs", wall))
		switch {
		case a.LaneOverrunSeconds != nil && e.Eligible:
			e.Eligible, e.Reason = false, fmt.Sprintf("a recent light run was stopped at %.0fs", *a.LaneOverrunSeconds)
		case (a.TimedOut || wall >= suiteLaneThreshold.Seconds()) && e.Eligible:
			e.Eligible, e.Reason = false, fmt.Sprintf("a recent suite ran %.0fs (limit %.0fs)", wall, suiteLaneThreshold.Seconds())
		}
	}
	if e.Eligible {
		e.Reason = fmt.Sprintf("its last %d suites ran %s, each under %.0fs", suiteLaneSamples, strings.Join(walls, ", "), suiteLaneThreshold.Seconds())
	}
	return e
}

// suiteLaneState is where one recorded-suite run stands in its lane order.
type suiteLaneState int

const (
	laneStart suiteLaneState = iota
	laneWaitLight
	laneRunLight
	laneWaitNormal
	laneRunNormal
	laneDone
)

type suiteLaneEvent int

const (
	laneEligible   suiteLaneEvent = iota // measured short and the helper has the lane
	laneIneligible                       // otherwise
	laneAcquired                         // the requested slot was granted
	laneExited                           // the suite finished before any deadline
	laneOverran                          // the light deadline stopped it
	laneCancelled                        // the run was cancelled or interrupted
)

var errLaneEvent = errors.New("suite lane: event not legal in this state")

// nextSuiteLaneStep is the whole order of one run (#559): at most one light
// attempt, and an overrun releases the light slot before the normal wait, so
// no run holds both kinds of slot or tries the light lane twice.
func nextSuiteLaneStep(s suiteLaneState, e suiteLaneEvent) (suiteLaneState, error) {
	switch {
	case s == laneStart && e == laneEligible:
		return laneWaitLight, nil
	case s == laneStart && e == laneIneligible:
		return laneWaitNormal, nil
	case s == laneWaitLight && e == laneAcquired:
		return laneRunLight, nil
	case s == laneWaitNormal && e == laneAcquired:
		return laneRunNormal, nil
	case s == laneRunLight && e == laneOverran:
		return laneWaitNormal, nil
	case (s == laneRunLight || s == laneRunNormal) && e == laneExited:
		return laneDone, nil
	case s != laneStart && s != laneDone && e == laneCancelled:
		return laneDone, nil
	}
	return s, errLaneEvent
}

// runSuiteLanes drives one recorded suite through nextSuiteLaneStep. run
// acquires the named lane's slot and runs the suite; a light run the lane's
// deadline stopped is retried once in the normal line, and the overrun's
// length rides on the run that retried it.
func runSuiteLanes(light bool, stderr io.Writer, run func(lane string) (suiteSlotRun, error)) (suiteSlotRun, *float64, error) {
	state, _ := nextSuiteLaneStep(laneStart, map[bool]suiteLaneEvent{true: laneEligible, false: laneIneligible}[light])
	var overrun *float64
	for {
		lane := ""
		if state == laneWaitLight {
			lane = suiteLaneLight
		}
		r, err := run(lane)
		if err != nil {
			return r, overrun, err // cancelled or could not run: nothing more to try
		}
		if state, err = nextSuiteLaneStep(state, laneAcquired); err != nil {
			return r, overrun, err
		}
		event := laneExited
		if r.LaneOverrun {
			event = laneOverran
		}
		if state, err = nextSuiteLaneStep(state, event); err != nil {
			return r, overrun, err
		}
		if state == laneDone {
			return r, overrun, nil
		}
		stopped := r.Duration.Seconds()
		overrun = &stopped
		fmt.Fprintf(stderr, "==> light lane: the suite was stopped at %.0fs, over the lane's bound; retrying it in the normal line\n", stopped)
	}
}
