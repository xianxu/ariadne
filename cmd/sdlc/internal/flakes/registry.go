// Package flakes is the pure core of issue-backed known-flake landing (#394):
// the registry of tests that open flake issues name, the classification of a
// failed suite attempt against it, the same-head fail→pass evidence a
// registration needs, and the targeted rerun's command. It does no IO.
package flakes

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// testName is a top-level Go test as the suite log names it: an import path,
// a dot, and a Test function. Subtests are represented by their parent.
var (
	testFunc = regexp.MustCompile(`^Test[A-Za-z0-9_]*$`)
	// Each segment starts with a letter, digit or underscore: never "-" (a
	// go test flag in the rerun argv) and never "." or ".." (a relative path).
	importPath = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_.~-]*(/[A-Za-z0-9_][A-Za-z0-9_.~-]*)*$`)
)

// SplitTest splits "<import path>.<TestName>" at its last dot. ok is false
// for anything else, so only an exact, rerunnable name can be registered.
func SplitTest(name string) (pkg, test string, ok bool) {
	i := strings.LastIndex(name, ".")
	if i <= 0 {
		return "", "", false
	}
	pkg, test = name[:i], name[i+1:]
	if !importPath.MatchString(pkg) || !testFunc.MatchString(test) {
		return "", "", false
	}
	return pkg, test, true
}

// ValidTest refuses a name that is not an exact top-level test.
func ValidTest(name string) error {
	if _, _, ok := SplitTest(name); !ok {
		return fmt.Errorf("%q is not a test name like github.com/x/y/pkg.TestName", name)
	}
	return nil
}

// Card is the part of an issue card the registry reads.
type Card struct {
	ID      string
	Status  string
	Tests   []string
	Owner   string // the claimant's workspace or operator; "" when unclaimed
	Created string // the issue's created date, for the registration's age
}

// Registry maps each registered test to the open issue that owns it.
type Registry struct{ owner map[string]string }

// BuildRegistry registers the tests of every card whose status active
// accepts. A test that two such issues name is ambiguous and refuses: which
// issue a failure belongs to must not depend on map order. A malformed name
// refuses too, rather than registering something no failure can match.
func BuildRegistry(cards []Card, active func(status string) bool) (Registry, error) {
	r := Registry{owner: map[string]string{}}
	for _, c := range cards {
		if !active(c.Status) {
			continue
		}
		for _, t := range c.Tests {
			if err := ValidTest(t); err != nil {
				return Registry{}, fmt.Errorf("issue %s flake_tests: %v", c.ID, err)
			}
			if prior, ok := r.owner[t]; ok && prior != c.ID {
				return Registry{}, fmt.Errorf("%s is registered by both issue %s and issue %s", t, prior, c.ID)
			}
			r.owner[t] = c.ID
		}
	}
	return r, nil
}

// Lookup returns the issue that registered test.
func (r Registry) Lookup(test string) (string, bool) {
	id, ok := r.owner[test]
	return id, ok
}

// Entry is one registered test and its issue.
type Entry struct {
	Test  string `json:"test"`
	Issue string `json:"issue"`
}

// Entries lists the registry, sorted by test.
func (r Registry) Entries() []Entry {
	out := make([]Entry, 0, len(r.owner))
	for t, id := range r.owner {
		out = append(out, Entry{Test: t, Issue: id})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Test < out[j].Test })
	return out
}
