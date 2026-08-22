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

// fixerFacingSurfaces DECLARES the class this file guards: every surface that
// hands findings to a fixer and says what to do with them (#203 BR-9). Each guard
// below globs its own slice of it; what was missing was the SET — with no declared
// set, siblings got ruled in or out from memory, which is how
// superpowers-receiving-code-review sat there telling the agent to implement
// findings "one item at a time" while this very issue existed to stop that.
//
// Guarded here:
//
//	cmd/sdlc/*.go            the eight gate emissions  (TestEveryFixerFacingSiteRoutes)
//	cmd/sdlc/helptext/*.md   the help those gates print (TestEveryFixerFacingHelptextRoutes)
//	construct/adapted/superpowers-receiving-code-review/SKILL.md
//	                         the canonical reception skill — the surface invoked at
//	                         exactly the moment a gate hands findings over
//	                         (TestReceivingCodeReviewSkillGeneralizes)
//
// Ruled OUT, with reasons:
//
//	cmd/sdlc/internal/...    reviewer-side prompts — a different audience; the judges
//	                         are already told to slug the rule, not the symptom
//	cmd/doc-review           different binary, no family ledger, explicitly advisory
//	                         and read-only over a document the agent owns, and the
//	                         ARCH-* registry is not delivered to it. A separable
//	                         extension in ARCH-PURPOSE's sense — revisit if it grows
//	                         a ledger
//	AGENTS.base.md           already routes to `sdlc arch-principles` and is guarded
//	                         by judge.TestArchitecture_NarrativeRoutesToArchPrinciples,
//	                         so extending ARCH-PURPOSE reaches it for free
//
// A new instruction surface belongs in this comment AND under a guard. Adding one
// to the tree without adding it here is the omission BR-9 names.

// TestReceivingCodeReviewSkillGeneralizes pins the third surface. The skill is
// prose, not Go, and it deliberately says "feedback"/"items" rather than
// "findings" — so it escapes fixerFacingMessage and needs its own assertion. What
// matters is that its response pattern tells the reader to generalize before
// implementing, and routes to the principle.
func TestReceivingCodeReviewSkillGeneralizes(t *testing.T) {
	path := filepath.Join("..", "..", "construct", "adapted",
		"superpowers-receiving-code-review", "SKILL.md")
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	skill := string(body)
	for _, want := range []string{"GENERALIZE", "ARCH-PURPOSE", "family:"} {
		if !strings.Contains(skill, want) {
			t.Errorf("%s no longer carries %q — the findings-reception surface stopped telling the reader to fix the class", path, want)
		}
	}
	if strings.Contains(skill, "6. IMPLEMENT: One item at a time") {
		t.Errorf("%s restored the per-item implement step that #203 replaced with GENERALIZE-then-sweep", path)
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
// THE RULE (arrived at after three rounds of widening to whatever shape the last
// finding named — #203 BR-7): match every fixer-facing message as a whole
// string-valued EXPRESSION, and attribute each match to a routing reference in its
// own STATEMENT.
//
// Both halves are load-bearing, and each collapses a family of shapes that earlier
// drafts handled one at a time:
//
//   - Whole-expression matching folds `+`-joined chains wherever they occur. The
//     tree splits one message across adjacent literals as a matter of style
//     (close.go's boundary-gate refusal is three pieces), so per-literal matching
//     misses any message whose "findings" and its directive verb land in different
//     pieces. Folding at the expression — not inside two hardcoded emitter names —
//     also covers the ~200 fmt.Fprint* calls that are not cwarn/die.
//   - Statement attribution means a routing reference credits only the message it
//     is joined to. Counting references per FUNCTION let an already-routing func
//     carry a second, unrouted line for free.
//
// SCAN BOUNDARY (#203 BR-6): cmd/sdlc's own package directory, non-test files.
// The full set of surfaces in this class — and the ruling on each — is
// fixerFacingSurfaces below.
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
		for _, m := range fixerFacingMatches(file) {
			if m.fn != nil && (m.fn.Name.Name == "fixTheClassLine" || m.fn.Name.Name == "fixTheClassNote") {
				// The one reasoned exclusion: the routing line's own definition
				// matches the signature it describes and cannot route through itself.
				continue
			}
			if referencesRoutingLine(m.stmt) {
				continue
			}
			where := describe(fset, m.expr.Pos(), path)
			if m.fn != nil {
				where += " (in " + m.fn.Name.Name + ")"
			} else {
				where += " (package-level)"
			}
			violations = append(violations, where)
		}
	}
	if len(violations) > 0 {
		t.Errorf("%d fixer-facing findings refusal(s) do not route through fixTheClassLine():\n  %s\n"+
			"Every surface that hands findings to the fixer must say to fix the CLASS, not the site (ARCH-PURPOSE).",
			len(violations), strings.Join(violations, "\n  "))
	}
}

// match is one fixer-facing message: the maximal string expression carrying it,
// the statement it belongs to (the unit a routing reference must share), and the
// enclosing func if any.
type match struct {
	expr ast.Expr
	stmt ast.Node
	fn   *ast.FuncDecl
}

// fixerFacingMatches walks the file once, tracking the innermost enclosing
// statement and func, and yields every MAXIMAL string-valued expression whose
// folded text is a fixer-facing message. Maximal matters: descending into a
// matched expression would re-report its own literal halves.
func fixerFacingMatches(file *ast.File) []match {
	var out []match
	var stmts []ast.Node
	var fns []*ast.FuncDecl

	var walk func(n ast.Node)
	walk = func(n ast.Node) {
		if n == nil {
			return
		}
		if expr, ok := n.(ast.Expr); ok && isStringExpr(expr) {
			if fixerFacingMessage(foldStringExpr(expr)) {
				m := match{expr: expr}
				if len(stmts) > 0 {
					m.stmt = stmts[len(stmts)-1]
				} else {
					m.stmt = expr // package-level const/var: the expression is its own unit
				}
				if len(fns) > 0 {
					m.fn = fns[len(fns)-1]
				}
				out = append(out, m)
			}
			return // maximal: do not descend
		}
		if fn, ok := n.(*ast.FuncDecl); ok {
			fns = append(fns, fn)
			defer func() { fns = fns[:len(fns)-1] }()
		}
		if stmt, ok := n.(ast.Stmt); ok {
			stmts = append(stmts, stmt)
			defer func() { stmts = stmts[:len(stmts)-1] }()
		}
		ast.Inspect(n, func(c ast.Node) bool {
			if c == n || c == nil {
				return c == n
			}
			walk(c)
			return false
		})
	}
	walk(file)
	return out
}

// isStringExpr reports whether e is a string literal or a +-joined chain
// containing one.
func isStringExpr(e ast.Expr) bool {
	switch v := e.(type) {
	case *ast.BasicLit:
		return v.Kind == token.STRING
	case *ast.BinaryExpr:
		return v.Op == token.ADD && (isStringExpr(v.X) || isStringExpr(v.Y))
	}
	return false
}

// foldStringExpr concatenates the literal parts of a string expression. Non-literal
// operands (a helper call, a variable) contribute nothing — they cannot be read
// statically, and a message that depends on one is matched on what IS visible.
func foldStringExpr(e ast.Expr) string {
	switch v := e.(type) {
	case *ast.BasicLit:
		if v.Kind != token.STRING {
			return ""
		}
		if s, err := strconv.Unquote(v.Value); err == nil {
			return s
		}
	case *ast.BinaryExpr:
		if v.Op == token.ADD {
			return foldStringExpr(v.X) + foldStringExpr(v.Y)
		}
	}
	return ""
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
