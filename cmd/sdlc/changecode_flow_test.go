package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flow"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
)

const flowIssue = "---\nid: 000231\nstatus: working\nestimate_hours:\n---\n\n# T\n\n## Spec\n\nthe contract\n\n" +
	"## Done when\n\n- it works\n\n## Plan\n\n- [ ] do it\n\n## Log\n"

func withPlanRows(content, rows string) string {
	return strings.Replace(content, "- [ ] do it\n", rows, 1)
}

func recordedFlow(t *testing.T, content string) flow.Flow {
	t.Helper()
	fm, _, err := issue.Parse(content)
	if err != nil {
		t.Fatal(err)
	}
	f, recorded, err := flow.FromFrontmatter(fm)
	if err != nil || !recorded {
		t.Fatalf("no valid flow record in:\n%s\n(recorded=%v err=%v)", content, recorded, err)
	}
	return f
}

// TestDecideChangeCodeFlow drives change-code's flow decision over issue text:
// what it infers, what it writes, and what it refuses (#231).
func TestDecideChangeCodeFlow(t *testing.T) {
	mx := withPlanRows(flowIssue, "- [ ] M1 — a\n- [ ] M2 — b\n")
	cases := []struct {
		name       string
		content    string
		plan, pin  string
		kind       flow.Kind
		provenance flow.Provenance
	}{
		{"no plan, no Mx → quick", flowIssue, "", "", flow.Quick, flow.Inferred},
		{"Mx rows → full", mx, "", "", flow.Full, flow.Inferred},
		{"durable plan → full", flowIssue, "# the plan", "", flow.Full, flow.Inferred},
		{"pin full on a small issue", flowIssue, "", "full", flow.Full, flow.Operator},
		{"pin quick despite a plan", flowIssue, "# the plan", "quick", flow.Quick, flow.Operator},
	}
	for _, c := range cases {
		d, err := decideChangeCodeFlow(c.content, c.plan, c.pin)
		if err != nil {
			t.Errorf("%s: %v", c.name, err)
			continue
		}
		got := recordedFlow(t, d.content)
		if got.Kind != c.kind || got.Provenance != c.provenance || d.flow != got {
			t.Errorf("%s: wrote %+v (returned %+v), want %s/%s", c.name, got, d.flow, c.kind, c.provenance)
		}
		if (got.Spec != "") != (c.kind == flow.Quick) {
			t.Errorf("%s: contract hashes present=%v, want them exactly on quick", c.name, got.Spec != "")
		}
		if d.rule == "" {
			t.Errorf("%s: no rule reported for the flow", c.name)
		}
	}

	// A milestone row quoted inside a fenced example is not a milestone: the
	// inference must read the fence-filtered Plan (#231 BR-4).
	fenced := withPlanRows(flowIssue, "- [ ] do it\n\n```markdown\n- [ ] M1 — example row\n```\n")
	if d, err := decideChangeCodeFlow(fenced, "", ""); err != nil || d.flow.Kind != flow.Quick {
		t.Errorf("fenced Mx example: got %+v (err %v), want quick — a quoted row is not a milestone", d.flow, err)
	}

	if _, err := decideChangeCodeFlow(mx, "", "quick"); err == nil {
		t.Error("--flow quick on a Plan with Mx rows: want a refusal")
	}

	// A re-run refreshes the anchor: the contract moved, the hashes follow.
	first, _ := decideChangeCodeFlow(flowIssue, "", "")
	reframed := strings.Replace(first.content, "the contract", "a reframed contract", 1)
	second, err := decideChangeCodeFlow(reframed, "", "")
	if err != nil || second.flow.Spec == first.flow.Spec {
		t.Errorf("re-run after a reframe kept spec %q (err %v) — the anchor must move", second.flow.Spec, err)
	}

	// quick/operator gains an Mx row → full/inferred (a crossed shell beats a pin).
	pinned, _ := decideChangeCodeFlow(flowIssue, "", "quick")
	grown := withPlanRows(pinned.content, "- [ ] M1 — a\n")
	d, err := decideChangeCodeFlow(grown, "", "")
	if err != nil || d.flow.Kind != flow.Full || d.flow.Provenance != flow.Inferred {
		t.Errorf("quick/operator + Mx row: got %+v (err %v), want full/inferred", d.flow, err)
	}

	// No frontmatter: nowhere to record a flow, so change-code refuses rather than
	// deciding one itself outside Decide (#231 BR-14).
	if _, err := decideChangeCodeFlow("# T\n\n## Spec\n\nx\n", "", ""); err == nil {
		t.Error("no frontmatter: want a refusal")
	}

	// A malformed record resolves to full and is rewritten well-formed, with a warning.
	bad := strings.Replace(flowIssue, "estimate_hours:", "estimate_hours:\nflow: {kind: quikc}", 1)
	d, err = decideChangeCodeFlow(bad, "", "")
	if err != nil || d.flow.Kind != flow.Full || d.warning == "" {
		t.Errorf("malformed record: got %+v warning=%q err=%v, want full with a warning", d.flow, d.warning, err)
	}
	_ = recordedFlow(t, d.content) // and it now parses
}

// TestRecordChangeCodeFlowWritesUnlessDryRun: reporting writes nothing, the
// record step writes the record, and --dry-run writes nothing.
func TestRecordChangeCodeFlowWritesUnlessDryRun(t *testing.T) {
	for _, dry := range []bool{false, true} {
		path := filepath.Join(t.TempDir(), "000231-x.md")
		if err := os.WriteFile(path, []byte(flowIssue), 0o644); err != nil {
			t.Fatal(err)
		}
		f := &changeCodeFlags{DryRun: dry, PlansDir: t.TempDir()}
		d := reportChangeCodeFlow(ioDiscard(), f, flowIssue, "")
		if on, _ := os.ReadFile(path); string(on) != flowIssue {
			t.Fatalf("dry-run=%v: reporting the flow wrote the issue", dry)
		}
		recordChangeCodeFlow(ioDiscard(), f, path, "000231-x", d)
		fl := d.flow
		on, _ := os.ReadFile(path)
		wrote := string(on) != flowIssue
		if wrote == dry {
			t.Errorf("dry-run=%v: file written=%v", dry, wrote)
		}
		if fl.Kind != flow.Quick {
			t.Errorf("dry-run=%v: flow %+v, want quick", dry, fl)
		}
	}
}

// TestActiveChangeCodeGates: on quick the run executes none of change-code's
// gates; on full, all of them in declaration order. changeCodeGateOrder keeps
// returning all five either way, so the ordering guards keep their strength.
func TestActiveChangeCodeGates(t *testing.T) {
	quick := &changeCodeCtx{f: &changeCodeFlags{}, flow: flow.Flow{Kind: flow.Quick, Provenance: flow.Inferred}}
	if gs := activeChangeCodeGates(quick); len(gs) != 0 {
		t.Errorf("quick runs %d gates, want 0", len(gs))
	}
	full := &changeCodeCtx{f: &changeCodeFlags{}, flow: flow.Flow{Kind: flow.Full, Provenance: flow.Inferred}}
	var names []string
	for _, g := range activeChangeCodeGates(full) {
		names = append(names, g.name)
	}
	if strings.Join(names, ",") != strings.Join(changeCodeGateOrder(), ",") {
		t.Errorf("full runs %v, want %v", names, changeCodeGateOrder())
	}
}

// TestPlanGateContentIgnoresFlow: writing (or re-pinning) the flow record must
// not bust the plan-quality pass-through cache and re-dispatch a judge.
func TestPlanGateContentIgnoresFlow(t *testing.T) {
	d, _ := decideChangeCodeFlow(flowIssue, "", "")
	if planGateContent(flowIssue) != planGateContent(d.content) {
		t.Error("recording the flow changed the plan-gate content")
	}
	pinned := strings.Replace(d.content, "provenance: inferred", "provenance: operator", 1)
	if planGateContent(d.content) != planGateContent(pinned) {
		t.Error("re-pinning the flow changed the plan-gate content")
	}
}

// TestFlowInfoLineNoGatesigCollision: the flow line change-code prints must not
// read as a gate bypass or refusal to the friction instrument (#172).
func TestFlowInfoLineNoGatesigCollision(t *testing.T) {
	for _, fl := range []flow.Flow{{Kind: flow.Quick, Provenance: flow.Inferred}, {Kind: flow.Full, Provenance: flow.Operator}} {
		assertNoGatesigCollision(t, "\x1b[1;36m==>\x1b[0m "+flowInfoLine(fl, flow.RuleNoPlan))
	}
}

// flowWirings: runChangeCode is not in-process drivable (#191), so the two
// call sites that make the flow real are asserted at the source.
var flowWirings = []wiring{
	{"changecode.go", "runChangeCode", "reportChangeCodeFlow",
		"the flow is decided before the gates run, which it decides between (#231)"},
	{"changecode.go", "runChangeCode", "recordChangeCodeFlow",
		"the decided flow is written to the issue (#231)"},
	{"changecode.go", "runChangeCode", "activeChangeCodeGates",
		"the gate loop iterates the flow-aware list, so quick runs no change-code gate (#231)"},
}

func TestChangeCodeWiresTheFlow(t *testing.T) {
	assertWiring(t, flowWirings)
}

// TestChangeCodeHelpShowsTheShell: the help prints the shell from its single
// source, so changing a limit cannot leave the help behind.
func TestChangeCodeHelpShowsTheShell(t *testing.T) {
	if got := renderLong("change-code"); !strings.Contains(got, flow.ShellSummary()) {
		t.Errorf("change-code help does not carry flow.ShellSummary():\n%s", got)
	}
}

// TestRunChangeCodeRecordsFlowAfterGates pins the ORDER the wiring test cannot:
// in runChangeCode the flow is recorded after the gate loop, after the dry-run
// return, and before the sync commit. A refused or dry run therefore leaves the
// issue byte-identical, and the record lands in the same commit as the design
// (#231 BR-5). runChangeCode is not in-process drivable (#191), so the order is
// asserted at the source.
func TestRunChangeCodeRecordsFlowAfterGates(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "changecode.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var run *ast.FuncDecl
	for _, d := range file.Decls {
		if fn, ok := d.(*ast.FuncDecl); ok && fn.Name.Name == "runChangeCode" {
			run = fn
		}
	}
	if run == nil {
		t.Fatal("runChangeCode not found")
	}
	pos := map[string]token.Pos{}
	var dryRunIf token.Pos
	ast.Inspect(run.Body, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.CallExpr:
			if id, ok := x.Fun.(*ast.Ident); ok && pos[id.Name] == 0 {
				pos[id.Name] = x.Pos()
			}
		case *ast.IfStmt:
			if sel, ok := x.Cond.(*ast.SelectorExpr); ok && sel.Sel.Name == "DryRun" && dryRunIf == 0 {
				dryRunIf = x.End()
			}
		}
		return true
	})
	order := []struct {
		name string
		at   token.Pos
	}{
		{"activeChangeCodeGates", pos["activeChangeCodeGates"]},
		{"the dry-run return", dryRunIf},
		{"recordChangeCodeFlow", pos["recordChangeCodeFlow"]},
		{"syncIssue", pos["syncIssue"]},
	}
	for i, o := range order {
		if o.at == 0 {
			t.Fatalf("%s not found in runChangeCode", o.name)
		}
		if i > 0 && o.at <= order[i-1].at {
			t.Errorf("%s comes before %s in runChangeCode — the flow record must be written after the gates pass "+
				"and past the dry-run return, and before the sync commit", o.name, order[i-1].name)
		}
	}
}

// TestRecordChangeCodeFlowKeepsConcurrentEdit: the gates can run for minutes,
// and the issue may be edited meanwhile (an editor, the agent). The record is
// re-derived from the file as it is at record time, so the edit survives and
// the contract hashes describe the edited text (#231 BR-11).
func TestRecordChangeCodeFlowKeepsConcurrentEdit(t *testing.T) {
	path := filepath.Join(t.TempDir(), "000231-x.md")
	os.WriteFile(path, []byte(flowIssue), 0o644)
	f := &changeCodeFlags{PlansDir: t.TempDir()}
	d := reportChangeCodeFlow(ioDiscard(), f, flowIssue, "")

	edited := strings.Replace(flowIssue, "the contract", "the contract, sharpened while the gates ran", 1)
	os.WriteFile(path, []byte(edited), 0o644)
	recordChangeCodeFlow(ioDiscard(), f, path, "000231-x", d)

	on, _ := os.ReadFile(path)
	if !strings.Contains(string(on), "sharpened while the gates ran") {
		t.Fatalf("the concurrent edit was clobbered:\n%s", on)
	}
	_, body, _ := issue.Parse(string(on))
	spec, done := flow.ContractHashes(body)
	if got := recordedFlow(t, string(on)); got.Spec != spec || got.Done != done {
		t.Errorf("recorded hashes %s/%s describe the stale text, want %s/%s", got.Spec, got.Done, spec, done)
	}
}

// TestRecordChangeCodeFlowRefusesFlowChangingEdit: an edit during the gates
// that changes the flow itself (here, Mx rows appear) means the gates ran for
// the wrong flow. Recording either answer would be wrong, so it refuses and the
// file is left as the editor left it.
func TestRecordChangeCodeFlowRefusesFlowChangingEdit(t *testing.T) {
	path := filepath.Join(t.TempDir(), "000231-x.md")
	os.WriteFile(path, []byte(flowIssue), 0o644)
	f := &changeCodeFlags{PlansDir: t.TempDir()}
	d := reportChangeCodeFlow(ioDiscard(), f, flowIssue, "")
	if d.flow.Kind != flow.Quick {
		t.Fatalf("precondition: %+v, want quick", d.flow)
	}

	grown := withPlanRows(flowIssue, "- [ ] M1 — a\n- [ ] M2 — b\n")
	os.WriteFile(path, []byte(grown), 0o644)
	msg, died := expectDie(t, func() { recordChangeCodeFlow(ioDiscard(), f, path, "000231-x", d) })
	if !died || !strings.Contains(msg, "re-run") {
		t.Errorf("flow-changing edit: died=%v msg=%q, want a refusal naming the re-run", died, msg)
	}
	if on, _ := os.ReadFile(path); string(on) != grown {
		t.Errorf("the refused record still wrote the file:\n%s", on)
	}
}

// TestOnlyFlowPackageBuildsFlowValues: Decide is the only producer of a flow
// decision, and Parse the only reader of a record. A flow.Flow literal or a
// flow.Kind / flow.Provenance / flow.Rule conversion anywhere else in cmd/sdlc
// is a decision made outside them — the #231 BR-7/BR-14 family, fixed as a rule.
func TestOnlyFlowPackageBuildsFlowValues(t *testing.T) {
	fset := token.NewFileSet()
	err := filepath.WalkDir(".", func(path string, de fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if de.IsDir() {
			if path == filepath.Join("internal", "flow") || de.Name() == "testdata" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		file, perr := parser.ParseFile(fset, path, nil, 0)
		if perr != nil {
			return perr
		}
		ast.Inspect(file, func(n ast.Node) bool {
			var typ ast.Expr
			switch x := n.(type) {
			case *ast.CompositeLit:
				typ = x.Type
			case *ast.CallExpr:
				typ = x.Fun
			default:
				return true
			}
			sel, ok := typ.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			if pkg, ok := sel.X.(*ast.Ident); ok && pkg.Name == "flow" {
				switch sel.Sel.Name {
				case "Flow", "Kind", "Provenance", "Rule":
					t.Errorf("%s: builds a flow.%s outside package flow — decide through flow.Decide "+
						"(or read through flow.Parse) instead", fset.Position(n.Pos()), sel.Sel.Name)
				}
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
