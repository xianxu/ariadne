package main

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flakes"
)

// SuiteDuration is one measured unit of work in a suite log. Kind keeps
// sharded processes distinct from Go packages.
type SuiteDuration struct {
	Kind    string  `json:"kind"`
	Name    string  `json:"name"`
	Seconds float64 `json:"seconds"`
}

// SuiteLogFacts contains only measurements recoverable from a completed log.
// A nil wall time means the log format did not report one.
type SuiteLogFacts struct {
	Failures    []string        `json:"failures,omitempty"`
	Units       []SuiteDuration `json:"units,omitempty"`
	WallSeconds *float64        `json:"wall_seconds"`
	Failed      bool            `json:"failed"`
	Warnings    []string        `json:"warnings,omitempty"`
	// Unattributed says why some failure has no exact test name (#394): a
	// build error, a timeout, a shard or package that failed without a
	// failing test, tests that never ran. Such a failure is never a flake.
	Unattributed []string `json:"unattributed,omitempty"`
	// Subtests are the failing subtests the log names (#466). Never
	// persisted: records keep top-level failures, and a reader that needs
	// subtests reads them from the log again.
	Subtests []string `json:"-"`
	// Typed is a transient projection; raw logs remain the evidence authority.
	Typed flakes.UnitFacts `json:"-"`
}

const suiteLogLineLimit = 1 << 20

var (
	shardSummary  = regexp.MustCompile(`^(packages|shard[0-9]+)\s+exit\s+([0-9]+)\s+([0-9]+(?:\.[0-9]+)?)s\s+[0-9]+\s+failed$`)
	goPackageOK   = regexp.MustCompile(`^ok\s+(\S+)\s+([0-9]+(?:\.[0-9]+)?)s(?:\s|$)`)
	goPackageFAIL = regexp.MustCompile(`^FAIL\s+(\S+)\s+([0-9]+(?:\.[0-9]+)?)s(?:\s|$)`)
	shardWall     = regexp.MustCompile(`^wall\s+([0-9]+(?:\.[0-9]+)?)s(?:\s|$)`)
)

// wire: go test -json events may add fields in a newer Go toolchain.
type goTestJSONEvent struct {
	Action  string   `json:"Action"`
	Package string   `json:"Package"`
	Test    string   `json:"Test"`
	Elapsed *float64 `json:"Elapsed"`
}

// parseSuiteLogFacts reads the existing text and go test -json suite formats. Failure
// naming stays centralized in failingTests, shared with baseline diagnosis.
func parseSuiteLogFacts(log string) SuiteLogFacts {
	// A very long output line can make both regex matching and the shared
	// failure extractor disproportionately expensive. Reject the entire log
	// before either parser sees it; partial measurements would mislead readers.
	start := 0
	for index := 0; index <= len(log); index++ {
		if index < len(log) && log[index] != '\n' {
			continue
		}
		if index-start > suiteLogLineLimit {
			return SuiteLogFacts{Warnings: []string{fmt.Sprintf("line exceeds %d bytes; suite facts unavailable", suiteLogLineLimit)}}
		}
		start = index + 1
	}
	facts := SuiteLogFacts{Failures: failingTests(log), Subtests: failingSubtests(log)}
	subtests := make(map[string]struct{}, len(facts.Subtests))
	for _, name := range facts.Subtests {
		subtests[name] = struct{}{}
	}
	failures := make(map[string]struct{}, len(facts.Failures))
	for _, name := range facts.Failures {
		failures[name] = struct{}{}
	}
	typed := flakes.NewUnitEvidence()
	var typedErrors []string
	ambiguousSubtests := false
	for index, raw := range strings.Split(log, "\n") {
		level := suiteIndentLevel(raw)
		if level != 0 {
			if level > 0 {
				fields := strings.Fields(raw)
				ambiguousSubtests = ambiguousSubtests || len(fields) >= 3 && fields[0] == "---" && fields[1] == "FAIL:" && strings.Contains(fields[2], "/")
			}
			continue
		}
		line := strings.TrimSuffix(raw, "\r")
		if event, recognized, err := flakes.DecodeUnitEvent(line); recognized {
			if err != nil {
				typedErrors = append(typedErrors, fmt.Sprintf("line %d: malformed unit event: %v", index+1, err))
			} else {
				_ = typed.Apply(event) // Apply retains every invalid transition as sticky evidence.
			}
			continue
		}
		if strings.HasPrefix(line, "{") && strings.Contains(line, `"Action"`) {
			var event goTestJSONEvent
			if err := json.Unmarshal([]byte(line), &event); err != nil {
				appendSuiteParseWarning(&facts, fmt.Sprintf("line %d: malformed go test JSON event: %v", index+1, err))
				continue
			}
			if event.Action == "fail" {
				facts.Failed = true
			}
			if event.Action == "fail" && event.Package != "" && event.Test != "" {
				if strings.Contains(event.Test, "/") {
					subtests[event.Package+"."+event.Test] = struct{}{}
				} else {
					failures[event.Package+"."+event.Test] = struct{}{}
				}
			}
			if (event.Action == "pass" || event.Action == "fail") && event.Test == "" {
				if event.Package == "" || event.Elapsed == nil || math.IsNaN(*event.Elapsed) || math.IsInf(*event.Elapsed, 0) || *event.Elapsed < 0 {
					appendSuiteParseWarning(&facts, fmt.Sprintf("line %d: package outcome lacks valid package or elapsed seconds", index+1))
					continue
				}
				facts.Units = append(facts.Units, SuiteDuration{Kind: "package", Name: event.Package, Seconds: *event.Elapsed})
			}
			continue
		}
		if m := shardSummary.FindStringSubmatch(line); m != nil {
			if seconds, ok := suiteSeconds(m[3]); ok {
				kind := "shard"
				if m[1] == "packages" {
					kind = "package-group"
				}
				facts.Units = append(facts.Units, SuiteDuration{Kind: kind, Name: m[1], Seconds: seconds})
			}
			facts.Failed = facts.Failed || m[2] != "0"
			continue
		}
		if m := goPackageOK.FindStringSubmatch(line); m != nil && strings.ContainsAny(m[1], "/.") {
			if seconds, ok := suiteSeconds(m[2]); ok {
				facts.Units = append(facts.Units, SuiteDuration{Kind: "package", Name: m[1], Seconds: seconds})
			}
			continue
		}
		if m := goPackageFAIL.FindStringSubmatch(line); m != nil && strings.ContainsAny(m[1], "/.") {
			if seconds, ok := suiteSeconds(m[2]); ok {
				facts.Units = append(facts.Units, SuiteDuration{Kind: "package", Name: m[1], Seconds: seconds})
			}
			facts.Failed = true
			continue
		}
		if m := shardWall.FindStringSubmatch(line); m != nil {
			if seconds, ok := suiteSeconds(m[1]); ok {
				facts.WallSeconds = &seconds
			}
		}
		if line == "FAIL" || strings.HasPrefix(line, "FAIL ") || strings.HasPrefix(line, "FAIL\t") {
			facts.Failed = true
		}
	}
	goFailures := len(failures) != 0
	facts.Typed = typed.Finish()
	facts.Typed.Unattributed = append(facts.Typed.Unattributed, typedErrors...)
	facts.Typed.Seen = facts.Typed.Seen || len(typedErrors) != 0
	for _, name := range facts.Typed.Failures {
		failures[name] = struct{}{}
	}
	facts.Failed = facts.Failed || len(facts.Typed.Failures) != 0 || len(facts.Typed.Unattributed) != 0
	if len(failures) != 0 {
		facts.Failures = facts.Failures[:0]
		for name := range failures {
			facts.Failures = append(facts.Failures, name)
		}
		sort.Strings(facts.Failures)
	}
	if len(subtests) != 0 {
		facts.Subtests = facts.Subtests[:0]
		for name := range subtests {
			facts.Subtests = append(facts.Subtests, name)
		}
		sort.Strings(facts.Subtests)
	}
	// Structured failure names must be complete before package attribution.
	facts.Unattributed = append(unattributedFailures(log, facts.Failures), facts.Typed.Unattributed...)
	if ambiguousSubtests && goFailures {
		facts.Warnings = append(facts.Warnings, "ambiguous indented subtest output ignored; attribution may be coarsened to a proven ancestor, which subtest-only flake registrations cannot cover")
	}
	return facts
}

func appendSuiteParseWarning(facts *SuiteLogFacts, warning string) {
	if len(facts.Warnings) < 8 {
		facts.Warnings = append(facts.Warnings, warning)
	} else if len(facts.Warnings) == 8 {
		facts.Warnings = append(facts.Warnings, "additional malformed events omitted")
	}
}

func suiteSeconds(raw string) (float64, bool) {
	seconds, err := strconv.ParseFloat(raw, 64)
	return seconds, err == nil && !math.IsNaN(seconds) && !math.IsInf(seconds, 0) && seconds >= 0
}

var (
	shardNoTest  = regexp.MustCompile(`^---\s+(\S+)\s+exited\s+[0-9]+\s+with no failing test`)
	neverRan     = regexp.MustCompile(`^---\s+[0-9]+\s+listed .* never ran`)
	buildFailed  = regexp.MustCompile(`^FAIL\s+(\S+)\s+\[(build|setup) failed\]`)
	testTimedOut = regexp.MustCompile(`^panic: test timed out after`)
)

// unattributedFailures lists why a failure in log has no exact test name,
// given the named failures. A package that failed is attributed only when one
// of its own tests failed: test-shard prints a package-level "--- FAIL pkg"
// beside every failing test, and go test a "FAIL pkg" line. A name without a
// package (a "--- FAIL: TestX" no package line claimed) is unattributed too.
func unattributedFailures(log string, failures []string) []string {
	var why []string
	add := func(s string) {
		if !slices.Contains(why, s) {
			why = append(why, s)
		}
	}
	named := map[string]bool{}
	for _, f := range failures {
		if id, err := flakes.ParseIdentity(f); err == nil && id.Kind() != flakes.KindGo {
			continue
		}
		if pkg, _, ok := flakes.SplitTest(f); ok {
			named[pkg] = true
		} else {
			add("failure " + f + " names no package")
		}
	}
	packageFailed := func(pkg string) {
		if pkg != "" && !named[pkg] {
			add("package " + pkg + " failed without a failing test")
		}
	}
	for _, raw := range strings.Split(log, "\n") {
		if !suiteControlLine(raw) {
			continue
		}
		line := strings.TrimSuffix(raw, "\r")
		switch fields := strings.Fields(line); {
		case shardNoTest.MatchString(line):
			add(shardNoTest.FindStringSubmatch(line)[1] + " exited without a failing test")
		case neverRan.MatchString(line):
			add("listed tests never ran")
		case buildFailed.MatchString(line):
			m := buildFailed.FindStringSubmatch(line)
			add("package " + m[1] + ": " + m[2] + " failed")
		case testTimedOut.MatchString(line):
			add("a test timed out")
		case len(fields) == 3 && fields[0] == "---" && fields[1] == "FAIL":
			packageFailed(fields[2])
		case goPackageFAIL.MatchString(line):
			packageFailed(goPackageFAIL.FindStringSubmatch(line)[1])
		case strings.HasPrefix(line, "{") && strings.Contains(line, `"Action":"fail"`):
			var event goTestJSONEvent
			if json.Unmarshal([]byte(line), &event) == nil && event.Test == "" {
				packageFailed(event.Package)
			}
		}
	}
	return why
}
