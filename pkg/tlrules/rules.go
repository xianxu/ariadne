// Package tlrules reads the tl-worker rule table (cmd/tl-worker/testdata/rules.tsv).
//
// The table is the one list of worker rules: tl-worker's tests check that its
// help states each one, and tl-lead's tests check that the lead's help does not
// restate the ones only the worker owns (#350).
package tlrules

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// Scope says who may state a rule's phrase.
type Scope string

const (
	// Own: only tl-worker states it.
	Own Scope = "own"
	// Cited: a command the lead's procedure names too.
	Cited Scope = "cited"
)

// Rule is one row: a rule name, its scope, and the phrase the worker help carries.
type Rule struct {
	Name   string
	Scope  Scope
	Phrase string
}

// Parse reads tab-separated rows (`name<TAB>scope<TAB>phrase`), skipping blank
// lines and `#` comments. A malformed row or an unknown scope is an error.
func Parse(r io.Reader) ([]Rule, error) {
	var rules []Rule
	sc := bufio.NewScanner(r)
	for n := 1; sc.Scan(); n++ {
		line := sc.Text()
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") {
			continue
		}
		f := strings.Split(line, "\t")
		if len(f) != 3 || f[0] == "" || f[2] == "" {
			return nil, fmt.Errorf("rules line %d: want name<TAB>scope<TAB>phrase, got %q", n, line)
		}
		s := Scope(f[1])
		if s != Own && s != Cited {
			return nil, fmt.Errorf("rules line %d: unknown scope %q", n, f[1])
		}
		rules = append(rules, Rule{Name: f[0], Scope: s, Phrase: f[2]})
	}
	return rules, sc.Err()
}
