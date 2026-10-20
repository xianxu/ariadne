// Package flakes is the pure core of issue-backed known-flake landing (#394):
// the registry of tests that open flake issues name, the classification of a
// failed suite attempt against it, the same-head fail→pass evidence a
// registration needs, and the targeted rerun's command. It does no IO.
package flakes

import (
	"fmt"
	"iter"
	"maps"
	"regexp"
	"sort"
	"strings"
)

// A test name is a Go test as the suite log names it: an import path, a dot,
// a Test function, and optionally its subtest path (#466): the segments
// go test prints after rewriting spaces, joined by "/".
var (
	testFunc = regexp.MustCompile(`^Test[A-Za-z0-9_]*$`)
	// Each segment starts with a letter, digit or underscore: never "-" (a
	// go test flag in the rerun argv) and never "." or ".." (a relative path).
	importPath = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_.~-]*(/[A-Za-z0-9_][A-Za-z0-9_.~-]*)*$`)
	// A subtest segment: argv-safe printable text with no "/" or space; the
	// rerun quotes it for a regexp wherever it reaches -run.
	subSegment = regexp.MustCompile(`^[A-Za-z0-9_.+=,:@#%~-]{1,128}$`)
)

// Name is a parsed test name.
type Name struct {
	Pkg, Test string // the import path and the top-level Test function
	Sub       string // the subtest path, "" for a top-level test
}

// Top is the top-level test's full name.
func (n Name) Top() string { return n.Pkg + "." + n.Test }

// String is the full name, subtests included.
func (n Name) String() string {
	if n.Sub == "" {
		return n.Top()
	}
	return n.Top() + "/" + n.Sub
}

// Run is the name as go test prints it after "--- PASS: ": the test and its
// subtest path, without the package.
func (n Name) Run() string {
	if n.Sub == "" {
		return n.Test
	}
	return n.Test + "/" + n.Sub
}

// ParseName parses "<import path>.<TestName>[/<sub>...]". The package
// boundary is the leftmost dot whose left side is an import path and whose
// right side, up to the first "/", is a Test function: a dot inside a
// subtest ("dev-aliases.test.sh"), or a subtest that looks like a test
// ("TestA/sub.TestB"), never moves it. ok is false for anything else.
func ParseName(name string) (Name, bool) {
	for i := 0; i < len(name); i++ {
		if name[i] != '.' {
			continue
		}
		pkg, rest := name[:i], name[i+1:]
		if !importPath.MatchString(pkg) {
			continue
		}
		test, sub, _ := strings.Cut(rest, "/")
		if !testFunc.MatchString(test) {
			continue
		}
		if strings.Contains(rest, "/") {
			for _, seg := range strings.Split(sub, "/") {
				if !subSegment.MatchString(seg) || seg == "." || seg == ".." {
					return Name{}, false
				}
			}
		}
		return Name{Pkg: pkg, Test: test, Sub: sub}, true
	}
	return Name{}, false
}

// SplitTest splits a top-level "<import path>.<TestName>"; ok is false for a
// subtest name or anything else, so callers that need a whole test get one.
func SplitTest(name string) (pkg, test string, ok bool) {
	n, ok := ParseName(name)
	if !ok || n.Sub != "" {
		return "", "", false
	}
	return n.Pkg, n.Test, true
}

// ValidTest refuses a name that is not an exact test or subtest name.
func ValidTest(name string) error {
	if _, ok := ParseName(name); !ok {
		return fmt.Errorf("%q is not a test name like github.com/x/y/pkg.TestName or …/pkg.TestName/subtest", name)
	}
	return nil
}

// Covers reports whether registered name r explains failing name f: f is r,
// or one of r's subtests. A registered subtest never covers a sibling or its
// parent; a registered top-level test covers all of its subtests.
func Covers(r, f string) bool { return f == r || strings.HasPrefix(f, r+"/") }

// Nearest is the name among names that covers failing name f most closely:
// f itself or its nearest ancestor; "" when none does. Two names covering one
// leaf are both its ancestors, so they differ in length and one is nearest.
func Nearest(names iter.Seq[string], f string) string {
	best := ""
	for name := range names {
		if Covers(name, f) && len(name) > len(best) {
			best = name
		}
	}
	return best
}

// Leaves are a failing top-level test's failing leaf subtests: those listed
// under it with no failing subtest of their own. With none listed, the test
// itself is the leaf (it failed on its own, or its log names no subtests).
func Leaves(top string, failing []string) []string {
	var under []string
	for _, f := range failing {
		if strings.HasPrefix(f, top+"/") {
			under = append(under, f)
		}
	}
	var leaves []string
	for _, f := range under {
		leaf := true
		for _, g := range under {
			if g != f && strings.HasPrefix(g, f+"/") {
				leaf = false
				break
			}
		}
		if leaf {
			leaves = append(leaves, f)
		}
	}
	if len(leaves) == 0 {
		return []string{top}
	}
	sort.Strings(leaves)
	return leaves
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
			for other, prior := range r.owner {
				if prior != c.ID && (Covers(other, t) || Covers(t, other)) {
					// Which issue a failure belongs to must be one answer: a
					// test and its own subtest on two issues would be two.
					return Registry{}, fmt.Errorf("%s (issue %s) and %s (issue %s) overlap: one issue must own both", t, c.ID, other, prior)
				}
			}
			r.owner[t] = c.ID
		}
	}
	return r, nil
}

// Lookup returns the issue that registered exactly test.
func (r Registry) Lookup(test string) (string, bool) {
	id, ok := r.owner[test]
	return id, ok
}

// Cover returns the registered name that covers failing name f: f itself or
// its nearest registered ancestor (#466). A registered sibling never covers.
func (r Registry) Cover(f string) (Entry, bool) {
	best := Nearest(maps.Keys(r.owner), f)
	if best == "" {
		return Entry{}, false
	}
	return Entry{Test: best, Issue: r.owner[best]}, true
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
