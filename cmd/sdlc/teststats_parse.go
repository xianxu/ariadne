package main

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
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
	facts := SuiteLogFacts{Failures: failingTests(log)}
	failures := make(map[string]struct{}, len(facts.Failures))
	for _, name := range facts.Failures {
		failures[name] = struct{}{}
	}
	for index, raw := range strings.Split(log, "\n") {
		line := strings.TrimSpace(raw)
		if strings.HasPrefix(line, "{") && strings.Contains(line, `"Action"`) {
			var event goTestJSONEvent
			if err := json.Unmarshal([]byte(line), &event); err != nil {
				appendSuiteParseWarning(&facts, fmt.Sprintf("line %d: malformed go test JSON event: %v", index+1, err))
				continue
			}
			if event.Action == "fail" {
				facts.Failed = true
			}
			if event.Action == "fail" && event.Package != "" && event.Test != "" && !strings.Contains(event.Test, "/") {
				failures[event.Package+"."+event.Test] = struct{}{}
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
	if len(failures) != 0 {
		facts.Failures = facts.Failures[:0]
		for name := range failures {
			facts.Failures = append(facts.Failures, name)
		}
		sort.Strings(facts.Failures)
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
