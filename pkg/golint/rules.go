package golint

import (
	"go/ast"
	"go/token"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Catalog is the curated list, in the order findings are reported. Add a rule
// with the fix for the mistake class it catches (tl-worker, tl-lead).
var Catalog = []Rule{
	{
		Name:     "test-hardcoded-arch-count",
		Incident: "ariadne#454",
		Message:  "derive architecture header counts from the registry or assert the shared ArchitectureBlock; a fixed count breaks when a principle is added",
		TestOnly: true,
		Match:    hardcodedArchCounts,
	},
	{
		Name:     "test-batch-shell-runner",
		Incident: "ariadne#448",
		Message:  "the construct batch runner shares one deadline across independent scripts; list with --list and bound each script separately",
		TestOnly: true,
		Match:    batchShellRunner,
	},
	{
		Name:     "wire-strict-decode",
		Incident: "pair#594",
		Message:  "strict JSON decoding rejects newer fields in a wire message; use a tolerant decoder for annotated wire types",
		Match:    wireStrictDecode,
	},
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
		Message:  "building a binary per test costs a compile each run; use testfix.BuiltBinary (one build per test process) or test the code in-process",
		TestOnly: true,
		Match:    goBuilds,
	},
	{
		Name:     "test-tight-wallclock",
		Incident: "2026-10-15 flakes",
		Message:  "a sub-second upper bound on elapsed wall-clock time fails under load; assert the event or outcome, or bound only a hang (1s or more)",
		TestOnly: true,
		Match:    tightWallclock,
	},
	{
		Name:     "test-fixed-sleep",
		Incident: "2026-10-15 flakes",
		Message:  "a fixed sleep guesses at timing; wait for an event (testfix.Watch, AwaitPIDFile, a channel) bounded by testfix.HangCeiling",
		TestOnly: true,
		Match:    fixedSleeps,
	},
}

var archEntryCountRE = regexp.MustCompile(`\beach of the [0-9]+ entries\b`)

// Catch the brittle rendered-header assertion, not intentional registry-policy
// tripwires or synthetic source snippets in catalog tests.
func hardcodedArchCounts(f *ast.File, _ *CheckContext) []ast.Node {
	var out []ast.Node
	calls(f, "strings", []string{"Contains"}, func(call *ast.CallExpr) {
		if len(call.Args) != 2 {
			return
		}
		if s, ok := stringLit(call.Args[1]); ok && archEntryCountRE.MatchString(s) {
			out = append(out, call)
		}
	})
	return out
}

// Protect the known batch entrypoint; Go syntax cannot infer arbitrary shell
// loops. Selection is safe, but executing this directory under one test command
// accumulates healthy progress against a hang ceiling.
func batchShellRunner(f *ast.File, _ *CheckContext) []ast.Node {
	var out []ast.Node
	calls(f, "os/exec", []string{"Command", "CommandContext"}, func(call *ast.CallExpr) {
		batch, list := false, false
		for _, arg := range call.Args {
			if s, ok := stringLit(arg); ok {
				batch = batch || path.Base(s) == "construct-scripts.test.sh"
				list = list || s == "--list"
			}
		}
		if batch && !list {
			out = append(out, call)
		}
	})
	return out
}

// constDuration evaluates a constant time expression: time.Millisecond,
// 50 * time.Millisecond, time.Second * 2 (timePkg is the file's name for the
// time package). ok is false for anything else.
func constDuration(timePkg string, e ast.Expr) (time.Duration, bool) {
	switch v := e.(type) {
	case *ast.ParenExpr:
		return constDuration(timePkg, v.X)
	case *ast.SelectorExpr:
		if pkg, ok := v.X.(*ast.Ident); !ok || timePkg == "" || pkg.Name != timePkg {
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
			d, ok := constDuration(timePkg, v.Y)
			return time.Duration(n) * d, ok
		}
		if n, ok := intLit(v.Y); ok {
			d, ok := constDuration(timePkg, v.X)
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

// localName is the name f refers to an imported package by: its alias, or the
// last element of its path. Empty when f does not import it.
func localName(f *ast.File, path string) string {
	for _, imp := range f.Imports {
		if p, err := strconv.Unquote(imp.Path.Value); err != nil || p != path {
			continue
		}
		if imp.Name != nil {
			return imp.Name.Name
		}
		return path[strings.LastIndex(path, "/")+1:]
	}
	return ""
}

// calls visits every call whose function is <the package at path>.<one of
// names>, by the name the file imports it under; empty names means any of its
// functions, and an empty path means any call at all.
func calls(f *ast.File, path string, names []string, visit func(*ast.CallExpr)) {
	pkg := ""
	if path != "" {
		if pkg = localName(f, path); pkg == "" {
			return
		}
	}
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

// shrunkDeadlines flags an override (=, not a declaration) of a variable or
// field named for a deadline. Without type-checking a local cannot be told
// from a package variable; locals so named are rare, and suppressible.
func shrunkDeadlines(f *ast.File, _ *CheckContext) []ast.Node {
	timePkg := localName(f, "time")
	var out []ast.Node
	ast.Inspect(f, func(n ast.Node) bool {
		as, ok := n.(*ast.AssignStmt)
		if !ok || as.Tok != token.ASSIGN || len(as.Lhs) != 1 || len(as.Rhs) != 1 {
			return true
		}
		var name string
		switch lhs := as.Lhs[0].(type) {
		case *ast.Ident:
			name = lhs.Name
		case *ast.SelectorExpr:
			name = lhs.Sel.Name
		}
		if !strings.Contains(name, "Timeout") && !strings.Contains(name, "Grace") && !strings.Contains(name, "Deadline") {
			return true
		}
		if d, ok := constDuration(timePkg, as.Rhs[0]); ok && d < time.Second {
			out = append(out, as)
		}
		return true
	})
	return out
}

func tmpLiterals(f *ast.File, _ *CheckContext) []ast.Node {
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
	calls(f, "path/filepath", []string{"Join"}, visit)
	calls(f, "io/ioutil", nil, visit)
	return out
}

func goBuilds(f *ast.File, _ *CheckContext) []ast.Node {
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

func fixedSleeps(f *ast.File, _ *CheckContext) []ast.Node {
	timePkg := localName(f, "time")
	var out []ast.Node
	calls(f, "time", []string{"Sleep"}, func(call *ast.CallExpr) {
		if len(call.Args) != 1 {
			return
		}
		// Polling ticks under 100ms are an event wait's cadence; a sleep of a
		// minute or more is the block-until-killed idiom of a child process.
		if d, ok := constDuration(timePkg, call.Args[0]); ok && d >= 100*time.Millisecond && d < time.Minute {
			out = append(out, call)
		}
	})
	return out
}

// tightWallclock flags an upper bound under 1s on elapsed time: time.Since(x)
// itself, or a variable assigned from it, compared against a constant. Lower
// bounds (proof that a wait happened) and loose hang bounds are fine.
func tightWallclock(f *ast.File, _ *CheckContext) []ast.Node {
	timePkg := localName(f, "time")
	if timePkg == "" {
		return nil
	}
	isSince := func(e ast.Expr) bool {
		call, ok := e.(*ast.CallExpr)
		if !ok {
			return false
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return false
		}
		pkg, ok := sel.X.(*ast.Ident)
		return ok && pkg.Name == timePkg && sel.Sel.Name == "Since"
	}
	elapsed := map[string]bool{}
	ast.Inspect(f, func(n ast.Node) bool {
		if as, ok := n.(*ast.AssignStmt); ok && len(as.Lhs) == 1 && len(as.Rhs) == 1 && isSince(as.Rhs[0]) {
			if id, ok := as.Lhs[0].(*ast.Ident); ok {
				elapsed[id.Name] = true
			}
		}
		return true
	})
	measured := func(e ast.Expr) bool {
		if id, ok := e.(*ast.Ident); ok {
			return elapsed[id.Name]
		}
		return isSince(e)
	}
	var out []ast.Node
	ast.Inspect(f, func(n ast.Node) bool {
		b, ok := n.(*ast.BinaryExpr)
		if !ok {
			return true
		}
		var bound ast.Expr
		switch {
		case (b.Op == token.GTR || b.Op == token.GEQ) && measured(b.X):
			bound = b.Y
		case (b.Op == token.LSS || b.Op == token.LEQ) && measured(b.Y):
			bound = b.X
		default:
			return true
		}
		if d, ok := constDuration(timePkg, bound); ok && d < time.Second {
			out = append(out, b)
		}
		return true
	})
	return out
}
