package golint

import (
	"go/ast"
	"go/token"
	"strconv"
	"strings"
	"time"
)

// Catalog is the curated list, in the order findings are reported. Add a rule
// with the fix for the mistake class it catches (tl-worker, tl-lead).
var Catalog = []Rule{
	{
		Name:     "test-shrunk-deadline",
		Incident: "ariadne#434",
		Message:  "a sub-second deadline races fixture startup under load; fire it on cue after readiness (testfix.DeadlineOnCue), or suppress where the test only asserts the bound",
		TestOnly: true,
		Match:    shrunkDeadlines,
	},
	{
		Name:     "tmp-literal",
		Incident: "ariadne#342",
		Message:  "a hardcoded /tmp path ignores TMPDIR, which the macOS sandbox requires; use os.TempDir, t.TempDir or os.MkdirTemp(\"\", …)",
		Match:    tmpLiterals,
	},
	{
		Name:     "test-go-build",
		Incident: "pair#602",
		Message:  "building a binary per test costs a compile each run; build it once per package (TestMain) or test the code in-process",
		TestOnly: true,
		Match:    goBuilds,
	},
	{
		Name:     "test-fixed-sleep",
		Incident: "2026-10-15 flakes",
		Message:  "a fixed sleep guesses at timing; wait for an event (testfix.Watch, AwaitPIDFile, a channel) bounded by testfix.HangCeiling",
		TestOnly: true,
		Match:    fixedSleeps,
	},
}

// constDuration evaluates a constant time expression: time.Millisecond,
// 50 * time.Millisecond, time.Second * 2. ok is false for anything else.
func constDuration(e ast.Expr) (time.Duration, bool) {
	switch v := e.(type) {
	case *ast.ParenExpr:
		return constDuration(v.X)
	case *ast.SelectorExpr:
		if pkg, ok := v.X.(*ast.Ident); !ok || pkg.Name != "time" {
			return 0, false
		}
		unit, ok := map[string]time.Duration{"Nanosecond": time.Nanosecond, "Microsecond": time.Microsecond,
			"Millisecond": time.Millisecond, "Second": time.Second, "Minute": time.Minute, "Hour": time.Hour}[v.Sel.Name]
		return unit, ok
	case *ast.BinaryExpr:
		if v.Op != token.MUL {
			return 0, false
		}
		if n, ok := intLit(v.X); ok {
			d, ok := constDuration(v.Y)
			return time.Duration(n) * d, ok
		}
		if n, ok := intLit(v.Y); ok {
			d, ok := constDuration(v.X)
			return time.Duration(n) * d, ok
		}
	}
	return 0, false
}

func intLit(e ast.Expr) (int64, bool) {
	lit, ok := e.(*ast.BasicLit)
	if !ok || lit.Kind != token.INT {
		return 0, false
	}
	n, err := strconv.ParseInt(lit.Value, 0, 64)
	return n, err == nil
}

func stringLit(e ast.Expr) (string, bool) {
	lit, ok := e.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return "", false
	}
	s, err := strconv.Unquote(lit.Value)
	return s, err == nil
}

// calls visits every call whose function is pkg.<one of names>; empty names
// means any function of pkg, and an empty pkg means any call at all.
func calls(f *ast.File, pkg string, names []string, visit func(*ast.CallExpr)) {
	ast.Inspect(f, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if pkg == "" {
			visit(call)
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if id, ok := sel.X.(*ast.Ident); ok && id.Name == pkg {
			if len(names) == 0 {
				visit(call)
			}
			for _, name := range names {
				if sel.Sel.Name == name {
					visit(call)
				}
			}
		}
		return true
	})
}

func shrunkDeadlines(f *ast.File) []ast.Node {
	var out []ast.Node
	ast.Inspect(f, func(n ast.Node) bool {
		as, ok := n.(*ast.AssignStmt)
		if !ok || as.Tok != token.ASSIGN || len(as.Lhs) != 1 || len(as.Rhs) != 1 {
			return true
		}
		id, ok := as.Lhs[0].(*ast.Ident)
		if !ok || !strings.Contains(id.Name, "Timeout") && !strings.Contains(id.Name, "Grace") && !strings.Contains(id.Name, "Deadline") {
			return true
		}
		if d, ok := constDuration(as.Rhs[0]); ok && d < time.Second {
			out = append(out, as)
		}
		return true
	})
	return out
}

func tmpLiterals(f *ast.File) []ast.Node {
	var out []ast.Node
	visit := func(call *ast.CallExpr) {
		for _, arg := range call.Args {
			if s, ok := stringLit(arg); ok && (s == "/tmp" || strings.HasPrefix(s, "/tmp/")) {
				out = append(out, arg)
			}
		}
	}
	// filepath.Join builds a path to use; FromSlash, Base and the like only
	// transform a string, so test data may name /tmp freely.
	calls(f, "os", nil, visit)
	calls(f, "filepath", []string{"Join"}, visit)
	calls(f, "ioutil", nil, visit)
	return out
}

func goBuilds(f *ast.File) []ast.Node {
	var out []ast.Node
	calls(f, "", nil, func(call *ast.CallExpr) {
		for i := 0; i+1 < len(call.Args); i++ {
			a, okA := stringLit(call.Args[i])
			b, okB := stringLit(call.Args[i+1])
			if okA && okB && a == "go" && b == "build" {
				out = append(out, call)
				return
			}
		}
	})
	return out
}

func fixedSleeps(f *ast.File) []ast.Node {
	var out []ast.Node
	calls(f, "time", []string{"Sleep"}, func(call *ast.CallExpr) {
		if len(call.Args) != 1 {
			return
		}
		// Polling ticks under 100ms are an event wait's cadence; a sleep of a
		// minute or more is the block-until-killed idiom of a child process.
		if d, ok := constDuration(call.Args[0]); ok && d >= 100*time.Millisecond && d < time.Minute {
			out = append(out, call)
		}
	})
	return out
}
