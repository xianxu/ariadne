package main

import (
	"iter"
	"sort"
	"strings"
	"unicode"
)

// Suite control records start at column zero. Quoted subprocess output is
// indented by testing and must not become package or outcome evidence (#476).
func suiteControlLine(line string) bool {
	return suiteIndentLevel(line) == 0
}

// Go's testing indenter adds four spaces per actual level, independently of
// slashes within a t.Run name. Keep the raw line; this only measures its indent.
func suiteIndentLevel(line string) int {
	spaces := 0
	for _, r := range line {
		if !unicode.IsSpace(r) {
			if spaces%4 == 0 {
				return spaces / 4
			}
			return -1
		}
		if r != ' ' {
			return -1
		}
		spaces++
	}
	return -1
}

// failingTests extracts top-level test names from plain go test output and
// test-shard.py's failure summary. The latter names the package before each
// test's raw output; plain go test names it in the trailing FAIL line.
// Subtests are represented by their parent; failingSubtests lists them.
func failingTests(log string) []string { return failingNames(log, false) }

// failingSubtests is the failing subtests the same log names (#466), as
// "<pkg>.<Test>/<sub>…": the known-flake gate judges a failing test by them.
func failingSubtests(log string) []string { return failingNames(log, true) }

// failingNames is the one scanner: top-level names, or only subtest names.
func failingNames(log string, subtests bool) []string {
	return failingNameLines(strings.Lines(log), subtests)
}

// The iterator keeps one traversal explicit and lets tests count line visits.
func failingNameLines(lines iter.Seq[string], subtests bool) []string {
	names := make(map[string]struct{})
	var pending []string
	var summaryPackage string
	var ancestors []string
	add := func(pkg, test string) {
		if test == "" || strings.Contains(test, "/") != subtests {
			return
		}
		if pkg != "" {
			test = pkg + "." + test
		}
		names[test] = struct{}{}
	}
	flush := func(pkg string) {
		for _, test := range pending {
			add(pkg, test)
		}
		pending = nil
		ancestors = ancestors[:0]
	}

	for line := range lines {
		level := suiteIndentLevel(line)
		if level < 0 {
			continue
		}
		control := level == 0
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		if !control {
			if len(fields) < 3 || fields[0] != "---" || fields[1] != "FAIL:" ||
				level <= 0 || level > len(ancestors) || !strings.HasPrefix(fields[2], ancestors[level-1]+"/") {
				continue
			}
			ancestors = append(ancestors[:level], fields[2])
		}
		if len(fields) >= 3 && fields[0] == "---" && fields[1] == "FAIL:" {
			if control {
				ancestors = append(ancestors[:0], fields[2])
			}
			pending = append(pending, fields[2])
			continue
		}
		if len(fields) >= 3 && fields[0] == "---" && fields[1] == "FAIL" {
			flush(summaryPackage)
			summaryPackage = fields[2]
			if len(fields) == 4 { // a package-only failure has no test name
				top, _, _ := strings.Cut(fields[3], "/")
				ancestors = append(ancestors[:0], top)
				add(summaryPackage, fields[3])
			}
			continue
		}
		if len(fields) >= 2 && fields[0] == "FAIL" {
			flush(fields[1])
			summaryPackage = ""
		}
	}
	flush(summaryPackage)
	if len(names) == 0 {
		return nil
	}
	result := make([]string, 0, len(names))
	for name := range names {
		result = append(result, name)
	}
	sort.Strings(result)
	return result
}

// classifyFailures reports which merged-head failures are new relative to
// main's baseline and which appeared in both runs.
func classifyFailures(merged, baseline []string) (introduced, preexisting []string) {
	base := make(map[string]struct{}, len(baseline))
	for _, name := range baseline {
		base[name] = struct{}{}
	}
	seen := make(map[string]struct{}, len(merged))
	for _, name := range merged {
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		if _, ok := base[name]; ok {
			preexisting = append(preexisting, name)
		} else {
			introduced = append(introduced, name)
		}
	}
	sort.Strings(introduced)
	sort.Strings(preexisting)
	return introduced, preexisting
}
