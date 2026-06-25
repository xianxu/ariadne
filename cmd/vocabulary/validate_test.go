package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// fakeCue runs the validator tests without the `cue` binary: VetInstance returns
// a canned output, so the pure diagnostic transform is exercised in isolation.
type fakeCue struct {
	out string
	err error
}

func (f fakeCue) Vet(string) error                           { return nil }
func (f fakeCue) Export(string) ([]byte, error)              { return nil, nil }
func (f fakeCue) VetInstance(_, _, _ string) (string, error) { return f.out, f.err }

// ── Captured cue v0.16.1 `cue vet -d` stderr fixtures. FIXTURE-COUPLED: a cue
// upgrade may reword these; re-capture from a real run if a bump breaks the parse. ──

const fxBadStatus = `status: 6 errors in empty disjunction:
status: conflicting values "blocked" and "in-progress":
    ./construct/vocabulary/issue.cue:15:24
    ../../tmp/badstatus.yaml:2:9
status: conflicting values "done" and "in-progress":
    ./construct/vocabulary/issue.cue:16:13
status: conflicting values "open" and "in-progress":
    ./construct/vocabulary/issue.cue:14:13
status: conflicting values "punt" and "in-progress":
status: conflicting values "wontfix" and "in-progress":
status: conflicting values "working" and "in-progress":
`

const fxDoneMissingActual = `actual_hours: field is required but not present
`

const fxStatusAbsent = `unresolved disjunction "open" | "working" | "blocked" | "done" | "wontfix" | "punt" (type string):
    ./construct/vocabulary/issue.cue:40:5
`

const fxFieldNotAllowed = `target: field not allowed:
    ./x.yaml:3:1
`

func TestParseCueDiagnostics_BadStatusEnum(t *testing.T) {
	diags := parseCueDiagnostics(fxBadStatus)
	if len(diags) != 1 {
		t.Fatalf("want 1 collapsed diagnostic for the enum conflict, got %d: %v", len(diags), diags)
	}
	d := diags[0]
	if d.Field != "status" {
		t.Errorf("field = %q, want status", d.Field)
	}
	if !strings.Contains(d.Message, `"in-progress" is not valid`) {
		t.Errorf("message should name the bad value: %q", d.Message)
	}
	for _, want := range []string{"open", "working", "blocked", "done", "wontfix", "punt"} {
		if !strings.Contains(d.Message, want) {
			t.Errorf("message should list valid value %q: %q", want, d.Message)
		}
	}
}

func TestParseCueDiagnostics_RequiredMissing(t *testing.T) {
	diags := parseCueDiagnostics(fxDoneMissingActual)
	if len(diags) != 1 || diags[0].Field != "actual_hours" {
		t.Fatalf("want one actual_hours diagnostic, got %v", diags)
	}
	if !strings.Contains(diags[0].Message, "required field is missing") {
		t.Errorf("message = %q", diags[0].Message)
	}
}

func TestParseCueDiagnostics_StatusAbsentDisjunction(t *testing.T) {
	diags := parseCueDiagnostics(fxStatusAbsent)
	if len(diags) != 1 {
		t.Fatalf("want one diagnostic, got %v", diags)
	}
	m := diags[0].Message
	if !strings.Contains(m, "required field is missing") || !strings.Contains(m, "open|working|blocked|done|wontfix|punt") {
		t.Errorf("message should flag a missing required enum field with its values: %q", m)
	}
}

// #124 M1-review steer: when cue NAMES the field on an `incomplete value` line
// (some schema shapes do), the diagnostic must preserve the field name rather than
// genericize it. (The real #Issue's missing-status case emits the field-less
// `unresolved disjunction` shape above; this covers the named variant.)
func TestParseCueDiagnostics_IncompleteValueNamesField(t *testing.T) {
	const fx = `status: incomplete value "open" | "working" | "blocked" | "done" | "wontfix" | "punt":
    ./x.cue:6:11
`
	diags := parseCueDiagnostics(fx)
	if len(diags) != 1 {
		t.Fatalf("want one diagnostic, got %v", diags)
	}
	if diags[0].Field != "status" {
		t.Errorf("field = %q, want status (name preserved)", diags[0].Field)
	}
	if !strings.Contains(diags[0].Message, "open|working|blocked|done|wontfix|punt") {
		t.Errorf("message should list the valid values: %q", diags[0].Message)
	}
}

func TestParseCueDiagnostics_FieldNotAllowed(t *testing.T) {
	diags := parseCueDiagnostics(fxFieldNotAllowed)
	if len(diags) != 1 || diags[0].Field != "target" {
		t.Fatalf("want one target diagnostic, got %v", diags)
	}
	if !strings.Contains(diags[0].Message, "unknown field") {
		t.Errorf("message = %q", diags[0].Message)
	}
}

// An unrecognized cue line must surface, not be silently dropped.
func TestParseCueDiagnostics_UnknownShapePassesThrough(t *testing.T) {
	diags := parseCueDiagnostics("some.field: a brand new cue error shape we don't parse yet\n")
	if len(diags) != 1 || !strings.Contains(diags[0].Message, "brand new cue error shape") {
		t.Errorf("unrecognized line should pass through verbatim, got %v", diags)
	}
}

func TestParseCueDiagnostics_Clean(t *testing.T) {
	if diags := parseCueDiagnostics(""); len(diags) != 0 {
		t.Errorf("empty cue output → no diagnostics, got %v", diags)
	}
}

// validateInstanceFile with a fake runner: the file read + frontmatter split are
// real; the cue verdict is injected.
func TestValidateInstanceFile_FakeRunner(t *testing.T) {
	md := filepath.Join(t.TempDir(), "x.md")
	if err := os.WriteFile(md, []byte("---\nid: \"000001\"\nstatus: working\n---\nbody\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Clean verdict → no diagnostics.
	got, err := validateInstanceFile(fakeCue{out: ""}, md, "schema.cue", "issue")
	if err != nil || len(got) != 0 {
		t.Fatalf("clean: got %v, err %v", got, err)
	}
	// Bad verdict → parsed diagnostics.
	got, err = validateInstanceFile(fakeCue{out: fxBadStatus}, md, "schema.cue", "issue")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Field != "status" {
		t.Errorf("want the status diagnostic, got %v", got)
	}
}

func TestValidateInstanceFile_NoFrontmatter(t *testing.T) {
	md := filepath.Join(t.TempDir(), "x.md")
	os.WriteFile(md, []byte("no fence here\n"), 0o644)
	if _, err := validateInstanceFile(fakeCue{}, md, "s.cue", "issue"); err == nil {
		t.Error("a file without a frontmatter fence should error")
	}
}

// ── Integration tests against the real `cue` binary + the real issue.cue. ──

func repoRootT(t *testing.T) string {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root, err := findRepoRoot(cwd)
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func issueSchemaT(t *testing.T) string {
	t.Helper()
	paths, err := resolveVocab()
	if err != nil {
		t.Fatal(err)
	}
	s, ok := paths["issue"]
	if !ok {
		t.Fatalf("no issue noun resolved (have %v)", sortedKeys(paths))
	}
	return s
}

// TestValidateInstance_ActiveCorpusPasses is the M1 contract: EVERY active
// workshop/issues/*.md (the set the merge gate actually validates) conforms to
// #Issue's frontmatter. (History holds legacy done-without-actuals files that
// predate the actuals requirement — covered separately; they're immutable and
// never enter a change window.)
func TestValidateInstance_ActiveCorpusPasses(t *testing.T) {
	if _, err := exec.LookPath("cue"); err != nil {
		t.Skip("cue not on PATH")
	}
	root := repoRootT(t)
	schema := issueSchemaT(t)
	files, _ := filepath.Glob(filepath.Join(root, "workshop", "issues", "*.md"))
	if len(files) == 0 {
		t.Fatal("no active issue files found")
	}
	for _, f := range files {
		diags, err := validateInstanceFile(osCue{}, f, schema, "issue")
		if err != nil {
			t.Errorf("%s: run error: %v", filepath.Base(f), err)
			continue
		}
		if len(diags) > 0 {
			t.Errorf("%s: NOT frontmatter-conformant: %v", filepath.Base(f), diags)
		}
	}
}

// The motivating scenario + the other rejection cases, end-to-end through real cue.
func TestValidateInstance_RejectsMalformations(t *testing.T) {
	if _, err := exec.LookPath("cue"); err != nil {
		t.Skip("cue not on PATH")
	}
	schema := issueSchemaT(t)
	cases := []struct {
		name      string
		fm        string
		wantField string
		wantMsg   string
	}{
		{"bad status value", "id: \"000001\"\nstatus: in-progress\n", "status", "not valid"},
		{"statuss typo (status absent)", "id: \"000001\"\nstatuss: working\n", "", "required field is missing"},
		{"done missing actual_hours", "id: \"000001\"\nstatus: done\n", "actual_hours", "required field is missing"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			md := filepath.Join(t.TempDir(), "x.md")
			os.WriteFile(md, []byte("---\n"+c.fm+"---\nbody\n"), 0o644)
			diags, err := validateInstanceFile(osCue{}, md, schema, "issue")
			if err != nil {
				t.Fatal(err)
			}
			if len(diags) == 0 {
				t.Fatalf("expected a rejection, got none")
			}
			found := false
			for _, d := range diags {
				if (c.wantField == "" || d.Field == c.wantField) && strings.Contains(d.Message, c.wantMsg) {
					found = true
				}
			}
			if !found {
				t.Errorf("want field=%q msg~=%q, got %v", c.wantField, c.wantMsg, diags)
			}
		})
	}
}

// A fully valid issue passes end-to-end.
func TestValidateInstance_ValidPasses(t *testing.T) {
	if _, err := exec.LookPath("cue"); err != nil {
		t.Skip("cue not on PATH")
	}
	schema := issueSchemaT(t)
	md := filepath.Join(t.TempDir(), "x.md")
	os.WriteFile(md, []byte("---\nid: \"000001\"\nstatus: working\nestimate_hours: 2.0\ntarget: foo\n---\n# T\n"), 0o644)
	diags, err := validateInstanceFile(osCue{}, md, schema, "issue")
	if err != nil {
		t.Fatal(err)
	}
	if len(diags) != 0 {
		t.Errorf("valid issue should pass, got %v", diags)
	}
}

func pensiveSchemaT(t *testing.T) string {
	t.Helper()
	paths, err := resolveVocab()
	if err != nil {
		t.Fatal(err)
	}
	s, ok := paths["pensive"]
	if !ok {
		t.Fatalf("no pensive noun resolved (have %v)", sortedKeys(paths))
	}
	return s
}

// TestValidateInstance_PensiveGeneralizes is #124 M3's genericity proof: the SAME
// validateInstanceFile engine validates a second datatype against #Pensive — the
// only per-datatype addition is construct/vocabulary/pensive.cue. Every real
// workshop/pensive/*.md passes; a bad `mode` enum is rejected with a named diagnostic.
func TestValidateInstance_PensiveGeneralizes(t *testing.T) {
	if _, err := exec.LookPath("cue"); err != nil {
		t.Skip("cue not on PATH")
	}
	root := repoRootT(t)
	schema := pensiveSchemaT(t)

	files, _ := filepath.Glob(filepath.Join(root, "workshop", "pensive", "*.md"))
	if len(files) == 0 {
		t.Fatal("no pensive files found")
	}
	for _, f := range files {
		diags, err := validateInstanceFile(osCue{}, f, schema, "pensive")
		if err != nil {
			t.Errorf("%s: run error: %v", filepath.Base(f), err)
			continue
		}
		if len(diags) > 0 {
			t.Errorf("%s: NOT pensive-conformant: %v", filepath.Base(f), diags)
		}
	}

	// A bad `mode` enum is rejected, naming the field — same engine, same diagnostic quality.
	md := filepath.Join(t.TempDir(), "p.md")
	os.WriteFile(md, []byte("---\ntype: pensive\ndate: 2026-06-25\ntopic: x\nmode: musing\ndescription: y\n---\nbody\n"), 0o644)
	diags, err := validateInstanceFile(osCue{}, md, schema, "pensive")
	if err != nil {
		t.Fatal(err)
	}
	rejectedMode := false
	for _, d := range diags {
		if d.Field == "mode" && strings.Contains(d.Message, "not valid") {
			rejectedMode = true
		}
	}
	if !rejectedMode {
		t.Errorf("expected a `mode` rejection for mode: musing, got %v", diags)
	}
}
