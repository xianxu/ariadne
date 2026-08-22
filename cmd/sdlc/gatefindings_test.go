package main

import (
	"bytes"
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
)

// TestFixTheClassLine_RoutesToArchPrinciples is the drift guard, modelled on
// judge.TestArchitecture_NarrativeRoutesToArchPrinciples (#128): a surface that
// wants to convey an ARCH principle ROUTES to it and cites the marker; it does not
// restate the definition and silently drift from it. So this asserts the ROUTING,
// never the wording — asserting the wording is what would let the line become a
// second copy of the principle.
func TestFixTheClassLine_RoutesToArchPrinciples(t *testing.T) {
	line := fixTheClassLine()
	if !strings.Contains(line, "ARCH-PURPOSE") {
		t.Error("the routing line must cite the marker, so the reader knows which principle governs")
	}
	if !strings.Contains(line, "sdlc arch-principles") {
		t.Error("the routing line must name the PULL command; a bare marker is a dangling pointer")
	}

	// The other half of the route: the destination still has to carry what the
	// citation promises. If ARCH-PURPOSE loses the class-vs-instance discipline,
	// every refusal is pointing at nothing.
	arch := judge.ArchitectureRegistry
	start := strings.Index(arch, "## ARCH-PURPOSE")
	if start < 0 {
		t.Fatal("ARCH-PURPOSE missing from the registry — the routing line points nowhere")
	}
	entry := arch[start:]
	if next := strings.Index(entry[3:], "\n## "); next >= 0 {
		entry = entry[:next+3]
	}
	for _, want := range []string{"CLASS", "enumeration", "family:"} {
		if !strings.Contains(entry, want) {
			t.Errorf("ARCH-PURPOSE no longer carries %q — the class discipline the gate refusals route to is gone", want)
		}
	}
}

// fixerFacingMessage reports whether a message is a fixer-facing findings refusal:
// it says findings exist (or need disposing) AND directs the reader to act on them.
// Both halves are required — "no valid ```findings block" reports a parse fault with
// nothing to fix, and "Fix the ref (or fence it)" directs without any findings.
//
// The directive vocabulary is deliberately wider than the three verbs the first
// draft used (#203 BR-2c): a signature narrower than the class it claims is the
// same defect this whole issue is about, one level down.
func fixerFacingMessage(s string) bool {
	l := strings.ToLower(s)
	hasFindings := strings.Contains(l, "finding") || strings.Contains(l, "dispose")
	if !hasFindings {
		return false
	}
	for _, verb := range []string{
		"fix ", "address ", "triage", "resolve ", "act on",
		"review above", "before crossing", "before committing",
	} {
		if strings.Contains(l, verb) {
			return true
		}
	}
	return false
}

// TestEveryFixerFacingSiteRoutes is the ENUMERATION guard (#203). A table over the
// known emitting funcs would be structurally blind to a ninth site — not
// hypothetical: this issue's first plan listed four sites and the tree held eight.
// So it scans the sources for the CLASS signature instead.
//
// SCAN BOUNDARY (#203 BR-6): cmd/sdlc's own package directory, non-test files.
// Subpackages are excluded deliberately — they emit no fixer-facing refusals (the
// gates live here), and internal/judge holds the reviewer-side prompts, which are a
// different audience. The one sibling binary with the same framing,
// cmd/doc-review/review.go's "triage each finding", is ruled OUT in the issue's
// Non-goals: different binary, no family ledger, explicitly advisory/read-only, and
// the ARCH-* registry is not delivered to it.
//
// Granularity is the CALL, not the function: finalizeBoundaryReview emits two
// separate arms (open-blocking and REWORK), and a function-level check would pass
// with only one routed. Where a literal is not inside a cwarn/die call — a string
// builder like formatFixThenShipProtocol — pass 2 is COUNT-based for the same
// reason (#203 BR-2d): requiring merely "the func routes somewhere" would let a
// second unrouted line ship inside an already-routing builder.
func TestEveryFixerFacingSiteRoutes(t *testing.T) {
	files, err := filepath.Glob(filepath.Join(".", "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	var violations []string
	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, src, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}

		// Pass 1: cwarn/die calls own every literal inside them. The call's
		// literals are CONCATENATED before matching (#203 BR-2a) — the tree's
		// prevailing style splits one message across adjacent literals, so
		// matching each in isolation misses a message whose "findings" and its
		// directive verb land in different pieces.
		claimed := map[ast.Node]bool{}
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			ident, ok := call.Fun.(*ast.Ident)
			if !ok || (ident.Name != "cwarn" && ident.Name != "die") {
				return true
			}
			var joined strings.Builder
			for _, lit := range stringLiterals(call) {
				claimed[lit.node] = true
				joined.WriteString(lit.text)
			}
			if fixerFacingMessage(joined.String()) && !referencesRoutingLine(call) {
				violations = append(violations, describe(fset, call.Pos(), path))
			}
			return true
		})

		// Pass 2: every remaining literal in the file — including package-level
		// const/var, which the first draft never walked (#203 BR-2b).
		for _, lit := range stringLiterals(file) {
			if claimed[lit.node] || !fixerFacingMessage(lit.text) {
				continue
			}
			fn := enclosingFunc(file, lit.node)
			if fn == nil {
				violations = append(violations, describe(fset, lit.node.Pos(), path)+" (package-level)")
				continue
			}
			// The one reasoned exclusion: the routing line's own definition matches
			// the signature it describes and cannot route through itself.
			if fn.Name.Name == "fixTheClassLine" || fn.Name.Name == "fixTheClassNote" {
				continue
			}
			if countMatchingLiterals(fn.Body, claimed) > countRoutingRefs(fn.Body) {
				violations = append(violations, describe(fset, lit.node.Pos(), path)+" (in "+fn.Name.Name+")")
			}
		}
	}
	if len(violations) > 0 {
		t.Errorf("%d fixer-facing findings refusal(s) do not route through fixTheClassLine():\n  %s\n"+
			"Every surface that hands findings to the fixer must say to fix the CLASS, not the site (ARCH-PURPOSE).",
			len(violations), strings.Join(violations, "\n  "))
	}
}

// TestEveryFixerFacingHelptextRoutes is the same guard for the DOC class (#203
// BR-1). The Go class got a mechanical scan that found eight where a hand pass
// found four; the doc class got a hand pass — which is the defect this issue is
// about, committed inside the issue about it. So the docs get a scan too, and it
// also pins the helptext ARCH-PURPOSE citations that were otherwise unguarded
// (BR-3).
//
// Granularity is the PARAGRAPH, routed if it or the paragraph immediately after
// cites the marker — matching how these files are written (a statement, then its
// elaboration). Blocks indented >=6 spaces are skipped: those are quoted tool
// output, not prose directing the reader. That exclusion is what correctly keeps
// close.md's convergence-line examples out — they quote the rule rather than
// instructing anyone.
func TestEveryFixerFacingHelptextRoutes(t *testing.T) {
	files, err := filepath.Glob(filepath.Join("helptext", "*.md"))
	if err != nil {
		t.Fatal(err)
	}
	var violations []string
	for _, path := range files {
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		paras := strings.Split(string(body), "\n\n")
		line, section := 1, ""
		for i, para := range paras {
			start := line
			line += strings.Count(para, "\n") + 2
			if h := sectionHeading(para); h != "" {
				section = h
			}
			if !fixerFacingMessage(para) || isQuotedOutput(para) || referenceSections[section] {
				continue
			}
			routed := strings.Contains(para, "ARCH-PURPOSE")
			if !routed && i+1 < len(paras) {
				routed = strings.Contains(paras[i+1], "ARCH-PURPOSE")
			}
			if !routed {
				violations = append(violations, path+":"+strconv.Itoa(start))
			}
		}
	}
	if len(violations) > 0 {
		t.Errorf("%d helptext passage(s) hand findings to the fixer without routing to ARCH-PURPOSE:\n  %s",
			len(violations), strings.Join(violations, "\n  "))
	}
}

// referenceSections are the helptext sections that DESCRIBE the tool rather than
// instruct the reader: what it accepts, what isn't built yet, where to read more.
// A passage there can mention findings without handing any to anyone —
// `judge.md`'s `--tools` flag note and its DEFERRED roadmap both do. This is the
// guard's explicit, reasoned exclusion; it is a structural class, not a list of
// line numbers, so it does not rot as the files are edited. A NEW reference
// section carrying a real directive fires the guard, which is correct: the author
// then routes it or names the section here.
var referenceSections = map[string]bool{
	"FLAGS":                        true,
	"USAGE":                        true,
	"EXAMPLES":                     true,
	"EXIT CODES":                   true,
	"ENVIRONMENT":                  true,
	"RELATED":                      true,
	"DEEP-DIVE REFERENCES":         true,
	"DEFERRED TO LATER MILESTONES": true,
}

// sectionHeading returns the ALL-CAPS heading a paragraph declares, or "".
func sectionHeading(para string) string {
	t := strings.TrimSpace(para)
	if t == "" || strings.Contains(t, "\n") {
		return ""
	}
	// Trim a trailing parenthetical like "THE GATE LEDGER (#194)".
	if i := strings.Index(t, " ("); i > 0 {
		t = t[:i]
	}
	if t != strings.ToUpper(t) {
		return ""
	}
	for _, r := range t {
		if r >= 'A' && r <= 'Z' {
			return t
		}
	}
	return ""
}

// isQuotedOutput reports whether every non-empty line is indented >=6 spaces —
// the helptext convention for verbatim tool output, as against 2-space prose and
// 4-to-5-space list items.
func isQuotedOutput(para string) bool {
	any := false
	for _, l := range strings.Split(para, "\n") {
		if strings.TrimSpace(l) == "" {
			continue
		}
		any = true
		if !strings.HasPrefix(l, "      ") {
			return false
		}
	}
	return any
}

func enclosingFunc(file *ast.File, target ast.Node) *ast.FuncDecl {
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		if target.Pos() >= fn.Body.Pos() && target.End() <= fn.Body.End() {
			return fn
		}
	}
	return nil
}

func countMatchingLiterals(n ast.Node, claimed map[ast.Node]bool) int {
	count := 0
	for _, lit := range stringLiterals(n) {
		if !claimed[lit.node] && fixerFacingMessage(lit.text) {
			count++
		}
	}
	return count
}

func countRoutingRefs(n ast.Node) int {
	count := 0
	ast.Inspect(n, func(node ast.Node) bool {
		if ident, ok := node.(*ast.Ident); ok &&
			(ident.Name == "fixTheClassLine" || ident.Name == "fixTheClassNote") {
			count++
		}
		return true
	})
	return count
}

type litRef struct {
	node ast.Node
	text string
}

func stringLiterals(n ast.Node) []litRef {
	var out []litRef
	ast.Inspect(n, func(node ast.Node) bool {
		lit, ok := node.(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}
		if s, err := strconv.Unquote(lit.Value); err == nil {
			out = append(out, litRef{node: lit, text: s})
		}
		return true
	})
	return out
}

func referencesRoutingLine(n ast.Node) bool {
	found := false
	ast.Inspect(n, func(node ast.Node) bool {
		// Either spelling routes: fixTheClassNote is fixTheClassLine pre-joined
		// for the common one-line-refusal case (#203 BR-4).
		if ident, ok := node.(*ast.Ident); ok &&
			(ident.Name == "fixTheClassLine" || ident.Name == "fixTheClassNote") {
			found = true
		}
		return !found
	})
	return found
}

func describe(fset *token.FileSet, pos token.Pos, path string) string {
	p := fset.Position(pos)
	return path + ":" + strconv.Itoa(p.Line)
}

// TestGatePathStderrCarriesRoutingLine closes the seam the two scan guards leave
// open (#203 BR-4). They assert the SOURCE routes; neither proves the line
// survives to an operator's stderr — a formatter whose output were dropped, or
// swallowed by a wrapper, would pass both scans and reach nobody. So this drives a
// real gate path end-to-end with a judge that reports findings, and reads stderr.
//
// runPlanQualityJudge is the seam changecode_test.go already drives directly; it
// is chosen for that reason rather than for being special. It is one path, not the
// class — the class is what the scans cover.
func TestGatePathStderrCarriesRoutingLine(t *testing.T) {
	orig := judge.Run
	t.Cleanup(func() { judge.Run = orig })
	judge.Run = func(ctx context.Context, onStart func(pid int), name string, args ...string) ([]byte, error) {
		// No ```findings block, so the gate takes classifyFallback's verdict-token
		// path — one of the eight routed sites.
		return []byte("VERDICT: FINDINGS (confidence: high)\n"), nil
	}

	var stderr bytes.Buffer
	f := &changeCodeFlags{PlansDir: t.TempDir()}
	err := runPlanQualityJudge(&bytes.Buffer{}, &stderr, f, "issue", "000001-issue.md", "## Spec\n\nx", "")
	if err == nil {
		t.Fatal("expected the plan-quality gate to fail on a findings verdict")
	}
	if got := stderr.String(); !strings.Contains(got, fixTheClassLine()) {
		t.Errorf("gate stderr does not carry the routing line.\ngot:\n%s\nwant substring:\n%s",
			got, fixTheClassLine())
	}
}
