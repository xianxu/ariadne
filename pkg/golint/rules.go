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
		Name:     "session-move-batched-reads",
		Incident: "ariadne#504",
		Message:  "session move must use RegistrationSnapshot and local lockfiles; per-document LookupRegistration or Status subprocesses made the 196-document live migration take about 20 minutes",
		Match:    sessionMoveBatchedReads,
	},
	{
		Name:     "session-move-raw-fields",
		Incident: "ariadne#504",
		Message:  "session move classification must decode YAML with moveMetadataFields; sidedoc.Fields preserves quotes and tags and can miss nested session registrations",
		Match:    sessionMoveRawFields,
	},
	{
		Name:     "suite-line-provenance",
		Incident: "ariadne#476",
		Message:  "suite scanners must preserve indentation: check suiteIndentLevel or suiteControlLine on the raw line before Fields, and never trim leading whitespace",
		Match:    suiteLineProvenance,
	},
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
		Incident: "ariadne#342, ariadne#468",
		Message:  "a hardcoded temp path (/tmp, /private/tmp, /var/folders, /Volumes/AgentScratch) ignores TMPDIR, which the macOS sandbox requires and which points suites at the TestScratch RAM disk; use os.TempDir, t.TempDir or os.MkdirTemp(\"\", …)",
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
		Name:     "test-reads-workshop",
		Incident: "ariadne#469",
		Message:  "a test that reads the repository's workshop/plans or workshop/issues breaks when the issue lands and its files move to workshop/history; read the atlas or a fixture",
		TestOnly: true,
		Match:    workshopReads,
	},
	{
		Name:     "json-recursive-scanner",
		Incident: "ariadne#489",
		Message:  "a hand-rolled recursive duplicate-key JSON scanner repeats jsonstrict's rule and drifts from it; call jsonstrict.RejectDuplicateKeys or jsonstrict.Scan (cmd/sdlc/internal/jsonstrict), which folds keys as encoding/json matches them",
		Match:    recursiveJSONScanners,
	},
	{
		Name:     "test-fixed-sleep",
		Incident: "2026-10-15 flakes",
		Message:  "a fixed sleep guesses at timing; wait for an event (testfix.Watch, AwaitPIDFile, a channel) bounded by testfix.HangCeiling",
		TestOnly: true,
		Match:    fixedSleeps,
	},
}

// These two owners use the migration's batch seam exclusively. Match selector
// names regardless of receiver: SessionSync is injected dynamically, so package
// alias or shadow resolution would not establish the receiver's concrete type.
func sessionMoveBatchedReads(f *ast.File, context *CheckContext) []ast.Node {
	filename := path.Clean(context.Path)
	owned := false
	for _, owner := range []string{"cmd/tl-lead/status_move_apply.go", "cmd/tl-lead/status_move_snapshot.go"} {
		if filename == owner || strings.HasSuffix(filename, "/"+owner) {
			owned = true
			break
		}
	}
	if !owned {
		return nil
	}
	var out []ast.Node
	ast.Inspect(f, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if ok && (sel.Sel.Name == "LookupRegistration" || sel.Sel.Name == "Status") {
			out = append(out, call)
		}
		return true
	})
	return out
}

// Protect the session move classifier, not arbitrary side-doc readers: Fields
// intentionally returns raw YAML elsewhere. Banning the raw reader in this
// owner avoids approximating YAML semantics or map-value dataflow in the lint.
func sessionMoveRawFields(f *ast.File, context *CheckContext) []ast.Node {
	const owner = "cmd/tl-lead/status_move_snapshot.go"
	filename := path.Clean(context.Path)
	if filename != owner && !strings.HasSuffix(filename, "/"+owner) {
		return nil
	}
	pkg := localName(f, "github.com/xianxu/ariadne/pkg/sidedoc")
	if pkg == "" || pkg == "_" {
		return nil
	}
	var out []ast.Node
	ast.Inspect(f, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		match := isPackageCall(call, pkg, "Fields")
		if pkg == "." {
			id, ok := call.Fun.(*ast.Ident)
			match = ok && id.Name == "Fields" && id.Obj == nil
		}
		if match {
			out = append(out, call)
		}
		return true
	})
	return out
}

// Guard the known text-suite scanners, not arbitrary parsers. This is a
// lexical ordering tripwire; behavioral fixtures own the provenance policy.
func suiteLineProvenance(f *ast.File, _ *CheckContext) []ast.Node {
	var out []ast.Node
	pkg := localName(f, "strings")
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || (fn.Name.Name != "failingNames" && fn.Name.Name != "failingNameLines" && fn.Name.Name != "parseSuiteLogFacts" && fn.Name.Name != "unattributedFailures") {
			continue
		}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			loop, ok := n.(*ast.RangeStmt)
			if !ok {
				return true
			}
			raw, ok := loop.Value.(*ast.Ident)
			if !ok {
				raw, ok = loop.Key.(*ast.Ident) // one-value range over a line iterator
			}
			if !ok {
				return true
			}
			checked := false
			ast.Inspect(loop.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				if fun, ok := call.Fun.(*ast.Ident); ok && (fun.Name == "suiteControlLine" || fun.Name == "suiteIndentLevel") && len(call.Args) == 1 {
					if arg, ok := call.Args[0].(*ast.Ident); ok && arg.Name == raw.Name {
						checked = true
					}
				}
				if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
					if id, ok := sel.X.(*ast.Ident); ok && pkg != "" && id.Name == pkg &&
						(sel.Sel.Name == "TrimSpace" || strings.HasPrefix(sel.Sel.Name, "TrimLeft") || sel.Sel.Name == "Fields" && !checked) {
						out = append(out, call)
					}
				}
				return true
			})
			return false
		})
	}
	return out
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
			if s, ok := stringLit(arg); ok && isTempLiteral(s) {
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

// tempRoots are the temp folders a path must reach through TMPDIR instead:
// the classic one, its macOS real path, the system's per-user folder and the
// agents' scratch volume (#468: all on the SSD).
var tempRoots = []string{"/tmp", "/private/tmp", "/var/folders", "/private/var/folders", "/Volumes/AgentScratch"}

func isTempLiteral(s string) bool {
	for _, r := range tempRoots {
		if s == r || strings.HasPrefix(s, r+"/") {
			return true
		}
	}
	return false
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

// workshopRE is a relative path that climbs from a package straight into
// the repository's live workshop artifacts ("../../workshop/plans/…"). A
// fixture path inside a temp repo ("workshop/issues/…", "../dst/workshop/…")
// does not climb straight there, so it is not matched.
var workshopRE = regexp.MustCompile(`^(\.\./)+workshop/(plans|issues)(/|$)`)

// workshopReads flags such a path passed to any call (a read, a Join, a
// glob), whole or split into Join segments: "..", "workshop", "plans".
func workshopReads(f *ast.File, _ *CheckContext) []ast.Node {
	var out []ast.Node
	calls(f, "", nil, func(call *ast.CallExpr) {
		lits := make([]string, len(call.Args))
		for i, arg := range call.Args {
			lits[i], _ = stringLit(arg)
		}
		for i, s := range lits {
			split := i > 0 && i+1 < len(lits) && s == "workshop" && (lits[i+1] == "plans" || lits[i+1] == "issues") &&
				(lits[i-1] == ".." || strings.HasSuffix(lits[i-1], "/.."))
			if workshopRE.MatchString(s) || split {
				out = append(out, call.Args[i])
			}
		}
	})
	return out
}

// recursiveJSONScanners flags a function, or a closure assigned to a variable,
// that calls itself while walking json.Decoder tokens (Token and More) and
// keeps a string-keyed set: the shape of every hand-rolled any-depth
// duplicate-key scanner #489 folded into jsonstrict, which alone may keep it.
// A schema decoder walking one object's members without recursing, or a
// recursive value skipper with no key set, is not this shape. An iterative
// walk over an explicit stack is beyond a syntactic rule; review catches it.
func recursiveJSONScanners(f *ast.File, _ *CheckContext) []ast.Node {
	if f.Name.Name == "jsonstrict" {
		return nil
	}
	walks := func(body *ast.BlockStmt, self string) bool {
		var tok, more, recurse, keySet bool
		ast.Inspect(body, func(n ast.Node) bool {
			if m, ok := n.(*ast.MapType); ok {
				if id, ok := m.Key.(*ast.Ident); ok && id.Name == "string" {
					keySet = true
				}
			}
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			switch fn := call.Fun.(type) {
			case *ast.SelectorExpr:
				tok = tok || fn.Sel.Name == "Token" && len(call.Args) == 0
				more = more || fn.Sel.Name == "More" && len(call.Args) == 0
				recurse = recurse || fn.Sel.Name == self // a method calling itself through its receiver
			case *ast.Ident:
				recurse = recurse || fn.Name == self
			}
			return true
		})
		return tok && more && recurse && keySet
	}
	var out []ast.Node
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			if x.Body != nil && walks(x.Body, x.Name.Name) {
				out = append(out, x)
			}
		case *ast.AssignStmt:
			for i, rhs := range x.Rhs {
				lit, ok := rhs.(*ast.FuncLit)
				if !ok || i >= len(x.Lhs) {
					continue
				}
				if id, ok := x.Lhs[i].(*ast.Ident); ok && walks(lit.Body, id.Name) {
					out = append(out, x)
				}
			}
		}
		return true
	})
	return out
}
