// Package golint is the curated catalog of mechanical Go mistakes (#407): each
// rule names the incident that taught it, and a repo test runs the catalog over
// the module, so the suite (and so every landing) enforces it. Rules read only
// the syntax tree: no type-checking, no external linter.
//
// A finding a rule gets wrong for a reason is suppressed on its line, or the
// line above, with
//
//	//lint:allow <rule> — <reason>
//
// The reason is required. A suppression naming an unknown rule, or one no
// finding uses, is itself a finding, so suppressions cannot outlive their cause.
package golint

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path"
	"strings"
)

// Rule is one catalogued mistake.
type Rule struct {
	Name     string
	Incident string // the issue that taught it, cited in every finding
	Message  string // what to do instead
	TestOnly bool   // applies to _test.go files only
	Match    func(f *ast.File, context *CheckContext) []ast.Node
}

// Finding is one rule firing at one line.
type Finding struct {
	Line    int
	Rule    string
	Message string
}

func (f Finding) String() string { return fmt.Sprintf("%d: %s: %s", f.Line, f.Rule, f.Message) }

// suppressionRule names findings about the suppressions themselves.
const suppressionRule = "lint-allow"

const allowPrefix = "//lint:allow "

type suppression struct {
	line int
	rule string
	used bool
}

// Check runs rules over one Go file's source and returns its findings, with
// valid suppressions applied and invalid or unused ones reported.
func Check(filename string, src []byte, rules []Rule) ([]Finding, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, src, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	index := newWireTypeIndex(map[string]*ast.File{filename: f})
	return checkParsed(fset, filename, f, rules, &CheckContext{Path: filename, Package: path.Dir(filename), PackageName: f.Name.Name, Index: index})
}

// CheckSources shares one parsed source set and wire-type index across files.
// The caller owns the thin IO seam; this function only sees bytes.
func CheckSources(module string, sources map[string][]byte, rules []Rule) (map[string][]Finding, error) {
	fset := token.NewFileSet()
	parsed := make(map[string]*ast.File, len(sources))
	for filename, src := range sources {
		f, err := parser.ParseFile(fset, filename, src, parser.ParseComments)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", filename, err)
		}
		parsed[filename] = f
	}
	index := newWireTypeIndex(parsed)
	findings := make(map[string][]Finding, len(parsed))
	for filename, f := range parsed {
		got, err := checkParsed(fset, filename, f, rules,
			&CheckContext{Path: filename, Package: path.Dir(filename), PackageName: f.Name.Name, Module: module, Index: index})
		if err != nil {
			return nil, err
		}
		findings[filename] = got
	}
	return findings, nil
}

func checkParsed(fset *token.FileSet, filename string, f *ast.File, rules []Rule, context *CheckContext) ([]Finding, error) {
	if ast.IsGenerated(f) {
		return nil, nil // a generator's output is fixed in the generator
	}
	known := map[string]bool{}
	for _, r := range rules {
		known[r.Name] = true
	}
	var out []Finding
	var allows []*suppression
	for _, group := range f.Comments {
		for _, c := range group.List {
			text, ok := strings.CutPrefix(c.Text, allowPrefix)
			if !ok {
				continue
			}
			line := fset.Position(c.Slash).Line
			name, reason, _ := strings.Cut(strings.TrimSpace(text), " ")
			reason = strings.TrimSpace(strings.TrimLeft(strings.TrimSpace(reason), "—–-:"))
			switch {
			case !known[name]:
				out = append(out, Finding{line, suppressionRule, fmt.Sprintf("suppression names unknown rule %q", name)})
			case reason == "":
				out = append(out, Finding{line, suppressionRule, fmt.Sprintf("suppression of %s gives no reason (//lint:allow %s — <reason>)", name, name)})
			default:
				allows = append(allows, &suppression{line: line, rule: name})
			}
		}
	}
	isTest := strings.HasSuffix(filename, "_test.go")
	for _, r := range rules {
		if r.TestOnly && !isTest {
			continue
		}
		for _, n := range r.Match(f, context) {
			line := fset.Position(n.Pos()).Line
			if allowed(allows, r.Name, line) {
				continue
			}
			out = append(out, Finding{line, r.Name, r.Message + " (" + r.Incident + ")"})
		}
	}
	for _, a := range allows {
		if !a.used {
			out = append(out, Finding{a.line, suppressionRule, fmt.Sprintf("suppression of %s matches no finding; remove it", a.rule)})
		}
	}
	return out, nil
}

// allowed marks and reports a suppression of rule on line or the line above.
func allowed(allows []*suppression, rule string, line int) bool {
	for _, a := range allows {
		if a.rule == rule && (a.line == line || a.line == line-1) {
			a.used = true
			return true
		}
	}
	return false
}
