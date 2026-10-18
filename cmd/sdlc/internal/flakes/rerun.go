package flakes

import (
	"fmt"
	"slices"
	"strings"
)

// RerunCommand is the argv of the targeted rerun: one `go test` over the
// packages of tests, matching exactly their names. Only validated names reach
// the -run pattern, as an argv element, never through a shell. A name that
// exists in more than one listed package also runs in the others; any failure
// there fails the rerun, which then blocks, so the rule stays conservative.
func RerunCommand(tests []string) ([]string, error) {
	if len(tests) == 0 {
		return nil, fmt.Errorf("no tests to rerun")
	}
	var pkgs, funcs []string
	for _, t := range tests {
		pkg, fn, ok := SplitTest(t)
		if !ok {
			return nil, fmt.Errorf("%q is not a rerunnable test name", t)
		}
		if !slices.Contains(pkgs, pkg) {
			pkgs = append(pkgs, pkg)
		}
		if !slices.Contains(funcs, fn) {
			funcs = append(funcs, fn)
		}
	}
	argv := []string{"go", "test", "-count=1", "-run", "^(" + strings.Join(funcs, "|") + ")$"}
	return append(argv, pkgs...), nil
}
