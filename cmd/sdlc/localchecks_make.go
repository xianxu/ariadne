// localchecks_make.go — facts about a repository's make test, for the default
// suite (#332): whether a test target exists and the suite's weight, both read
// from main's Makefile (never the branch under test, so a branch cannot drop
// or lighten the suite it owes).
package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// makeTestFactsFn is the seam: tests answer without running make.
var makeTestFactsFn = makeTestFacts

// suiteWeightProbe prints the Makefile's suite weight under a goal that is not
// `test`, so Makefile.workflow never takes the machine lock at parse time.
const suiteWeightProbe = "__sdlc_suite_weight: ; @echo __SDLC_SUITE_WEIGHT=$(or $(WF_TEST_WEIGHT),1)\n"

var (
	suiteWeightLine = regexp.MustCompile(`^__SDLC_SUITE_WEIGHT=(\d+)$`)
	testRuleLine    = regexp.MustCompile(`^test:`)
)

// makeTestFacts probes rev's Makefile in a detached worktree: has is whether
// make's rule database (make -p) lists a test target, weight is
// WF_TEST_WEIGHT (default 1). No Makefile: has is false.
func makeTestFacts(git func(args ...string) (string, error), rev string) (has bool, weight int, err error) {
	wt, cleanup, err := detachedWorktree(git, rev)
	if err != nil {
		return false, 0, fmt.Errorf("probe make test on %s: %w", abbrevSHA(rev), err)
	}
	defer cleanup()
	has, weight, err = probeMakeTestFacts(wt)
	if err != nil {
		return false, 0, fmt.Errorf("probe make test on %s: %w", abbrevSHA(rev), err)
	}
	return has, weight, nil
}

// probeMakeTestFacts asks make in dir, under a goal that is not test.
func probeMakeTestFacts(dir string) (has bool, weight int, err error) {
	if _, err := os.Stat(filepath.Join(dir, "Makefile")); err != nil {
		return false, 0, nil
	}
	probe, err := os.CreateTemp("", "sdlc-suite-probe-*.mk")
	if err != nil {
		return false, 0, err
	}
	defer os.Remove(probe.Name())
	if _, err := probe.WriteString(suiteWeightProbe); err != nil {
		probe.Close()
		return false, 0, err
	}
	probe.Close()
	cmd := exec.Command("make", "-s", "-p", "-f", "Makefile", "-f", probe.Name(), "__sdlc_suite_weight")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "WF_TEST_LOCK=off") // belt and braces: a probe never queues for a suite slot
	out, err := cmd.Output()
	if err != nil {
		return false, 0, err
	}
	return parseMakeTestFacts(string(out))
}

// parseMakeTestFacts reads the probe's output: make's database plus the
// weight line.
func parseMakeTestFacts(out string) (has bool, weight int, err error) {
	weight = 0
	sc := bufio.NewScanner(strings.NewReader(out))
	sc.Buffer(make([]byte, 64<<10), 4<<20)
	for sc.Scan() {
		line := sc.Text()
		if m := suiteWeightLine.FindStringSubmatch(line); m != nil {
			weight, _ = strconv.Atoi(m[1])
		}
		if testRuleLine.MatchString(line) {
			has = true
		}
	}
	if err := sc.Err(); err != nil {
		return false, 0, err
	}
	if weight == 0 {
		return false, 0, fmt.Errorf("probe make test: the weight line is missing")
	}
	return has, weight, nil
}
