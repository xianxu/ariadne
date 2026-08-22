package main

import (
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

// fixerFacingLiteral reports whether a message string is a fixer-facing findings
// refusal: it says findings exist (or need disposing) AND directs the reader to
// act on them. Both halves are required — "no valid ```findings block" reports a
// parse fault with nothing to fix, and "Fix the ref (or fence it)" directs without
// any findings involved.
func fixerFacingLiteral(s string) bool {
	l := strings.ToLower(s)
	hasFindings := strings.Contains(l, "finding") || strings.Contains(l, "dispose")
	hasDirective := strings.Contains(l, "fix ") || strings.Contains(l, "address ") ||
		strings.Contains(l, "review above")
	return hasFindings && hasDirective
}

// TestEveryFixerFacingSiteRoutes is the ENUMERATION guard (#203). A table over the
// known emitting funcs would be structurally blind to a ninth site — which is not
// hypothetical: the first draft of this issue's plan listed four sites and the tree
// held eight. So this scans the sources for the CLASS signature instead, and every
// match must route through fixTheClassLine.
//
// Granularity is the call, not the function: finalizeBoundaryReview emits two
// separate arms (open-blocking and REWORK), and a function-level check would pass
// with only one of them routed.
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

		// Pass 1: cwarn/die calls own the literals inside them. Each such call
		// must reference the routing line itself.
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
			matched := false
			for _, lit := range stringLiterals(call) {
				claimed[lit.node] = true
				if fixerFacingLiteral(lit.text) {
					matched = true
				}
			}
			if matched && !referencesRoutingLine(call) {
				violations = append(violations, describe(fset, call.Pos(), path))
			}
			return true
		})

		// Pass 2: anything else — a string builder like formatFixThenShipProtocol,
		// which never calls cwarn itself. Its enclosing func must route.
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			// The one reasoned exclusion: the routing line's own definition
			// matches the signature it describes, and cannot route through itself.
			if fn.Name.Name == "fixTheClassLine" {
				continue
			}
			for _, lit := range stringLiterals(fn.Body) {
				if claimed[lit.node] || !fixerFacingLiteral(lit.text) {
					continue
				}
				if !referencesRoutingLine(fn.Body) {
					violations = append(violations, describe(fset, lit.node.Pos(), path)+" (in "+fn.Name.Name+")")
				}
			}
		}
	}
	if len(violations) > 0 {
		t.Errorf("%d fixer-facing findings refusal(s) do not route through fixTheClassLine():\n  %s\n"+
			"Every surface that hands findings to the fixer must say to fix the CLASS, not the site (ARCH-PURPOSE).",
			len(violations), strings.Join(violations, "\n  "))
	}
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
		if ident, ok := node.(*ast.Ident); ok && ident.Name == "fixTheClassLine" {
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
