package flakes

import (
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/jsonstrict"
)

// RerunCommands selects exactly the requested whole tests in each package.
// Registered subtests promote to their parent test so a failing sibling blocks
// the rerun (#466). Sorting and deduplication make the argv independent of the
// registry's input order; names never leak into another package's pattern.
func RerunCommands(tests []string) ([][]string, error) {
	if len(tests) == 0 {
		return nil, fmt.Errorf("no tests to rerun")
	}
	groups := make(map[string]map[string]bool)
	for _, test := range tests {
		name, ok := ParseName(test)
		if !ok {
			return nil, fmt.Errorf("%q is not a rerunnable test name", test)
		}
		if groups[name.Pkg] == nil {
			groups[name.Pkg] = make(map[string]bool)
		}
		groups[name.Pkg][name.Test] = true
	}
	var packages []string
	for pkg := range groups {
		packages = append(packages, pkg)
	}
	slices.Sort(packages)
	commands := make([][]string, 0, len(packages))
	for _, pkg := range packages {
		var names []string
		for name := range groups[pkg] {
			names = append(names, regexp.QuoteMeta(name))
		}
		slices.Sort(names)
		commands = append(commands, []string{"go", "test", "-count=1", "-json", "-run", "^(" + strings.Join(names, "|") + ")$", pkg})
	}
	return commands, nil
}

// RerunMissingJSON lists names without complete package-bound go test -json
// proof. Each required test, its ancestors, and its package must pass. Output
// text is never evidence. Conflicting terminal outcomes cannot be repaired by
// a later pass, and ambiguous JSON invalidates the log rather than authorizing
// a partial prefix. Plain command announcements are allowed between events.
// RerunMissing remains the separate reader for supported legacy text logs.
func RerunMissingJSON(log string, tests []string) []string {
	type identity struct{ pkg, test string }
	type terminal struct{ passed, blocked bool }
	outcomes := make(map[identity]terminal)
	invalid := false
	for line := range strings.SplitSeq(log, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "{") {
			continue
		}
		raw := []byte(line)
		if err := jsonstrict.RejectDuplicateKeys(raw); err != nil {
			invalid = true
			break
		}
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(raw, &fields); err != nil {
			invalid = true
			break
		}
		// Decode the interpreted fields as strings explicitly: encoding/json's
		// default string decoding accepts null, which would turn Test:null into a
		// package outcome. Unknown toolchain fields do not affect this proof.
		values := make(map[string]string)
		malformed := false
		for field, value := range fields {
			for _, known := range []string{"Action", "Package", "Test"} {
				if !strings.EqualFold(field, known) {
					continue
				}
				var text string
				if string(value) == "null" || json.Unmarshal(value, &text) != nil {
					malformed = true
					break
				}
				values[known] = text
			}
		}
		if malformed {
			invalid = true
			break
		}
		action, pkg, test := values["Action"], values["Package"], values["Test"]
		if action != "pass" && action != "fail" && action != "skip" {
			continue
		}
		if !importPath.MatchString(pkg) {
			invalid = true
			break
		}
		if test != "" {
			name, ok := ParseName(pkg + "." + test)
			if !ok || name.Pkg != pkg || name.Run() != test {
				invalid = true
				break
			}
		}
		key := identity{pkg, test}
		outcome := outcomes[key]
		outcome.passed = outcome.passed || action == "pass"
		outcome.blocked = outcome.blocked || action != "pass"
		outcomes[key] = outcome
	}
	var missing []string
	for _, test := range tests {
		name, ok := ParseName(test)
		complete := ok && !invalid
		if complete {
			outcome := outcomes[identity{name.Pkg, ""}]
			complete = outcome.passed && !outcome.blocked
			run := name.Run()
			for {
				outcome = outcomes[identity{name.Pkg, run}]
				complete = complete && outcome.passed && !outcome.blocked
				index := strings.LastIndexByte(run, '/')
				if index < 0 {
					break
				}
				run = run[:index]
			}
		}
		if !complete {
			missing = append(missing, test)
		}
	}
	return missing
}
