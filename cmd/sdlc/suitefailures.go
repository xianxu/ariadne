package main

import (
	"sort"
	"strings"
)

// failingTests extracts top-level test names from plain go test output and
// test-shard.py's failure summary. The latter names the package before each
// test's raw output; plain go test names it in the trailing FAIL line.
func failingTests(log string) []string {
	names := make(map[string]struct{})
	var pending []string
	var summaryPackage string
	add := func(pkg, test string) {
		if test == "" || strings.Contains(test, "/") {
			return // subtests are represented by their parent
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
	}

	for _, line := range strings.Split(log, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		if len(fields) >= 3 && fields[0] == "---" && fields[1] == "FAIL:" {
			pending = append(pending, fields[2])
			continue
		}
		if len(fields) >= 3 && fields[0] == "---" && fields[1] == "FAIL" {
			flush(summaryPackage)
			summaryPackage = fields[2]
			if len(fields) == 4 { // a package-only failure has no test name
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
