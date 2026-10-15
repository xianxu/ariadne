package main

import (
	"math"
	"regexp"
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
}

var (
	shardSummary = regexp.MustCompile(`^(packages|shard[0-9]+)\s+exit\s+[0-9]+\s+([0-9]+(?:\.[0-9]+)?)s\s+[0-9]+\s+failed$`)
	goPackageOK  = regexp.MustCompile(`^ok\s+(\S+)\s+([0-9]+(?:\.[0-9]+)?)s(?:\s|$)`)
	shardWall    = regexp.MustCompile(`^wall\s+([0-9]+(?:\.[0-9]+)?)s(?:\s|$)`)
)

// parseSuiteLogFacts reads the two existing suite output formats. Failure
// naming stays centralized in failingTests, shared with baseline diagnosis.
func parseSuiteLogFacts(log string) SuiteLogFacts {
	facts := SuiteLogFacts{Failures: failingTests(log)}
	for _, raw := range strings.Split(log, "\n") {
		line := strings.TrimSpace(raw)
		if m := shardSummary.FindStringSubmatch(line); m != nil {
			if seconds, ok := suiteSeconds(m[2]); ok {
				kind := "shard"
				if m[1] == "packages" {
					kind = "package-group"
				}
				facts.Units = append(facts.Units, SuiteDuration{Kind: kind, Name: m[1], Seconds: seconds})
			}
			continue
		}
		if m := goPackageOK.FindStringSubmatch(line); m != nil && strings.ContainsAny(m[1], "/.") {
			if seconds, ok := suiteSeconds(m[2]); ok {
				facts.Units = append(facts.Units, SuiteDuration{Kind: "package", Name: m[1], Seconds: seconds})
			}
			continue
		}
		if m := shardWall.FindStringSubmatch(line); m != nil {
			if seconds, ok := suiteSeconds(m[1]); ok {
				facts.WallSeconds = &seconds
			}
		}
	}
	return facts
}

func suiteSeconds(raw string) (float64, bool) {
	seconds, err := strconv.ParseFloat(raw, 64)
	return seconds, err == nil && !math.IsNaN(seconds) && !math.IsInf(seconds, 0) && seconds >= 0
}
