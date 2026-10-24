// localchecks_make.go — facts about a repository's make test, for the default
// suite (#332): whether a test target exists and the suite's weight, both read
// from the selected policy tree. Existing suites remain main-authoritative;
// localchecks_bootstrap.go permits an exact-subject default only on proven absence.
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/processgroup"
)

// makeTestFactsFn is the seam: tests answer without running make.
var makeTestFactsFn = makeTestFactsContext

// suiteWeightProbe prints the Makefile's suite weight under a goal that is not
// `test`, so Makefile.workflow never takes the machine lock at parse time.
var suiteProbeTimeout = 2 * time.Minute // a hanging $(shell) in main's Makefile must not stall merge

const suiteWeightProbe = "__sdlc_suite_weight: ; @echo __SDLC_SUITE_WEIGHT=$(or $(WF_TEST_WEIGHT),1)\n"

var (
	suiteWeightLine = regexp.MustCompile(`^__SDLC_SUITE_WEIGHT=(\d+)$`)
	testRuleLine    = regexp.MustCompile(`^test:`)
)

// withDefaultSuite adds the default suite to decl when it declares none,
// probing rev's Makefile (rev is main: a branch cannot drop the suite it owes).
// The one derivation the CI gate and `sdlc local-check run` share.

func withDefaultSuiteContext(ctx context.Context, git func(args ...string) (string, error), rev string, decl localCheckDecl) (localCheckDecl, error) {
	if _, declared := decl.check(suiteCheckName); declared {
		return effectiveLocalChecks(decl, false, 0), nil
	}
	has, weight, err := makeTestFactsFn(ctx, git, rev)
	if err != nil {
		return localCheckDecl{}, err
	}
	return effectiveLocalChecks(decl, has, weight), nil
}

// makeTestFacts probes rev's Makefile in a detached worktree: has is whether
// make's rule database (make -p) lists a test target, weight is
// WF_TEST_WEIGHT (default 1). No Makefile: has is false.

func makeTestFactsContext(ctx context.Context, git func(args ...string) (string, error), rev string) (has bool, weight int, err error) {
	if err := ctx.Err(); err != nil {
		return false, 0, err
	}
	wt, cleanup, err := detachedWorktree(git, rev)
	if err != nil {
		return false, 0, fmt.Errorf("probe make test on %s: %w", abbrevSHA(rev), err)
	}
	defer cleanup()
	has, weight, err = probeMakeTestFactsContext(ctx, wt)
	if err != nil {
		return false, 0, fmt.Errorf("probe make test on %s: %w", abbrevSHA(rev), err)
	}
	return has, weight, nil
}

// probeMakeTestFacts asks make in dir, under a goal that is not test.

func probeMakeTestFactsContext(ctx context.Context, dir string) (has bool, weight int, err error) {
	if err := ctx.Err(); err != nil {
		return false, 0, err
	}
	if _, err := os.Stat(filepath.Join(dir, "Makefile")); err != nil {
		if os.IsNotExist(err) {
			return false, 0, nil
		}
		return false, 0, fmt.Errorf("stat Makefile: %w", err)
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
	ctx, cancel := context.WithTimeout(ctx, suiteProbeTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "make", "-s", "-p", "-f", "Makefile", "-f", probe.Name(), "__sdlc_suite_weight")
	processgroup.Configure(cmd)
	cmd.Cancel = func() error { return processgroup.Terminate(cmd, true) }
	cmd.WaitDelay = time.Second
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "WF_TEST_LOCK=off") // belt and braces: a probe never queues for a suite slot
	out, err := cmd.Output()
	// A parent can exit before children close inherited pipes. WaitDelay
	// bounds the drain; terminate any remaining group members as well.
	_ = processgroup.Terminate(cmd, true)
	// A probe that succeeded stands, even if the deadline passed after it (#599).
	if err != nil && ctx.Err() != nil {
		return false, 0, fmt.Errorf("make probe interrupted: %w", ctx.Err())
	}
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
	prev := ""
	for sc.Scan() {
		line := sc.Text()
		if m := suiteWeightLine.FindStringSubmatch(line); m != nil {
			weight, _ = strconv.Atoi(m[1])
		}
		// make -p lists a file named only as a prerequisite under "# Not a
		// target:"; that is not a test rule.
		if testRuleLine.MatchString(line) && prev != "# Not a target:" {
			has = true
		}
		prev = line
	}
	if err := sc.Err(); err != nil {
		return false, 0, err
	}
	if weight == 0 {
		return false, 0, fmt.Errorf("probe make test: the weight line is missing")
	}
	return has, weight, nil
}
