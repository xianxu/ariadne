package reviewexport

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func TestParseMode(t *testing.T) {
	m, err := ParseMode("---\nname: a\norder: 2\nscope: whole-doc\nfrontier: off\n---\nBrief.\n", "a")
	if err != nil || m != (Mode{Name: "a", Order: 2, Scope: "whole-doc", Frontier: "off"}) {
		t.Fatalf("got %+v, %v", m, err)
	}
	// Omitted flags take parley's defaults; order is optional.
	m, err = ParseMode("---\nname: b\n---\nBrief.\n", "b")
	if err != nil || m != (Mode{Name: "b", Scope: "markers-only", Frontier: "on"}) {
		t.Fatalf("defaults: got %+v, %v", m, err)
	}
	for _, c := range []struct{ name, text, want string }{
		{"no frontmatter", "name: a\n", "missing frontmatter"},
		{"unterminated", "---\nname: a\n", "unterminated"},
		{"name mismatch", "---\nname: b\n---\n", "basename"},
		{"bad flag", "---\nname: a\nscope: everything\n---\n", "scope="},
		{"retired deletion flag", "---\nname: a\ndeletions: propose-strike\n---\n", "deletions"},
		{"bad order", "---\nname: a\norder: first\n---\n", "order"},
		// Parity with parley's parser at the edges.
		{"crlf", "---\r\nname: a\r\n---\r\n", "missing frontmatter"},
		{"quoted value is literal", "---\nname: \"a\"\n---\n", "basename"},
		{"empty frontmatter", "---\n---\nBody\n", "basename"},
	} {
		if _, err := ParseMode(c.text, "a"); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: err %v, want %q", c.name, err, c.want)
		}
	}
}

func fixture() fstest.MapFS {
	f := func(s string) *fstest.MapFile { return &fstest.MapFile{Data: []byte(s)} }
	return fstest.MapFS{
		"round.md":               f("round\n"),
		"review-convention.md":   f("grammar\n"),
		"collaboration-modes.md": f("collaboration modes\n"),
		"adapters/file-edit.md":  f("adapter\n"),
		"modes/zeta.md":          f("---\nname: zeta\n---\nZ\n"),
		"modes/beta.md":          f("---\nname: beta\norder: 2\n---\nB\n"),
		"modes/alpha.md":         f("---\nname: alpha\norder: 2\n---\nA\n"),
		"modes/first.md":         f("---\nname: first\norder: 1\n---\nF\n"),
	}
}

// The order is total and matches parley's listing: role, then mode order
// with unordered modes last, then name.
func TestBuildOrder(t *testing.T) {
	m, err := Build(fixture())
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, mem := range m.Members {
		got = append(got, mem.Role+":"+mem.Path)
	}
	want := "round:round.md grammar:review-convention.md collaboration-modes:collaboration-modes.md adapter:adapters/file-edit.md mode:modes/first.md mode:modes/alpha.md mode:modes/beta.md mode:modes/zeta.md"
	if strings.Join(got, " ") != want {
		t.Errorf("order\n got %s\nwant %s", strings.Join(got, " "), want)
	}
	if g := m.Members[1]; g.Source != GrammarSource || g.Bytes != len("grammar\n") {
		t.Errorf("grammar member %+v", g)
	}
	if z := m.Members[7]; z.Scope != "markers-only" || z.Order != 0 {
		t.Errorf("an unordered mode records its defaults: %+v", z)
	}
}

// The move cases' files are exported as fixtures, for parley's mover
// (#424); the other examples are not. Fixtures sort last, by case, then path.
func TestBuildFixtures(t *testing.T) {
	fs := fixture()
	f := func(s string) *fstest.MapFile { return &fstest.MapFile{Data: []byte(s)} }
	fs["examples/move-b/case.md"] = f("submitted: d.md\noutcome: ok\nmove: d.md e.md\n")
	fs["examples/move-b/before/d.md"] = f("D\n")
	fs["examples/move-b/after/e.md"] = f("D\n")
	fs["examples/move-a/case.md"] = f("move: x.md y.md\n")
	fs["examples/chat/case.md"] = f("submitted: s.md\noutcome: ok\n")
	fs["examples/chat/before/s.md"] = f("S\n")
	m, err := Build(fs)
	if err != nil {
		t.Fatal(err)
	}
	if m.Version != 3 {
		t.Errorf("version %d, want 3 (mode deletion flag retired)", m.Version)
	}
	var got []string
	for _, mem := range m.Members {
		if mem.Role == "fixture" {
			got = append(got, mem.Name+":"+mem.Path)
		}
	}
	want := "move-a:examples/move-a/case.md move-b:examples/move-b/after/e.md move-b:examples/move-b/before/d.md move-b:examples/move-b/case.md"
	if strings.Join(got, " ") != want {
		t.Errorf("fixtures\n got %s\nwant %s", strings.Join(got, " "), want)
	}
	if last := m.Members[len(m.Members)-1]; last.Role != "fixture" {
		t.Errorf("fixtures do not sort last: %+v", last)
	}
}

func TestV3ExportOmitsModeDeletionKnob(t *testing.T) {
	m, err := Build(fixture())
	if err != nil {
		t.Fatal(err)
	}
	data, err := Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(data, []byte(`"deletions"`)) || !bytes.Contains(data, []byte(`"version": 3`)) {
		t.Fatalf("v3 still exports the retired deletion knob: %s", data)
	}
}

func TestBuildRefusesMissingMembers(t *testing.T) {
	for _, drop := range []string{"round.md", "review-convention.md", "collaboration-modes.md", "adapters/file-edit.md"} {
		fs := fixture()
		delete(fs, drop)
		if _, err := Build(fs); err == nil {
			t.Errorf("built without %s", drop)
		}
	}
}

func TestMarshalDeterministic(t *testing.T) {
	a, _ := Build(fixture())
	b, _ := Build(fixture())
	ja, err := Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	jb, _ := Marshal(b)
	if !bytes.Equal(ja, jb) || !bytes.HasSuffix(ja, []byte("}\n")) {
		t.Errorf("two builds differ or lack the trailing newline:\n%s\n%s", ja, jb)
	}
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	out, err := exec.Command("go", "env", "GOMOD").Output()
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Dir(strings.TrimSpace(string(out)))
}

// The committed export.json is the build of the committed skill directory.
// Regenerate it with UPDATE_REVIEW_EXPORT=1 go test ./pkg/reviewexport.
func TestExportCurrent(t *testing.T) {
	root := moduleRoot(t)
	dir := filepath.Join(root, "construct", "local", "universal-review")
	m, err := Build(os.DirFS(dir))
	if err != nil {
		t.Fatal(err)
	}
	want, err := Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "export.json")
	if os.Getenv("UPDATE_REVIEW_EXPORT") == "1" {
		if err := os.WriteFile(path, want, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("%v (generate it with UPDATE_REVIEW_EXPORT=1 go test ./pkg/reviewexport)", err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("export.json is stale; regenerate it with UPDATE_REVIEW_EXPORT=1 go test ./pkg/reviewexport")
	}
	// The grammar is the one authored copy, linked rather than duplicated.
	link := filepath.Join(dir, "review-convention.md")
	if target, err := os.Readlink(link); err != nil || target != "../fix/review-convention.md" {
		t.Errorf("review-convention.md must link ../fix/review-convention.md: %q, %v", target, err)
	}
	canon, err := os.ReadFile(filepath.Join(root, GrammarSource))
	if err != nil {
		t.Fatal(err)
	}
	if m.Members[1].Bytes != len(canon) {
		t.Errorf("grammar member covers %d bytes, the source has %d", m.Members[1].Bytes, len(canon))
	}
	var modes []string
	collab := 0
	for _, mem := range m.Members {
		switch mem.Role {
		case "mode":
			modes = append(modes, mem.Name)
		case "collaboration-modes":
			collab++
		}
		// Only the move cases are exported, as parley's fixtures (#424).
		if strings.HasPrefix(mem.Path, "examples/") != (mem.Role == "fixture") || mem.Role == "fixture" && !strings.HasPrefix(mem.Name, "move-") {
			t.Errorf("%s exported as %s/%s; only the move cases are, as fixtures", mem.Path, mem.Role, mem.Name)
		}
	}
	// #412: parley's sync requires exactly one collaboration-modes member.
	if collab != 1 {
		t.Errorf("%d collaboration-modes members, want exactly one", collab)
	}
	if got := strings.Join(modes, ","); got != "developmental,line-editing,copy-editing,proofreading,fact-check,free-form" {
		t.Errorf("stock modes in order: %s", got)
	}
}

// These are the two agent-facing adapters for the same Pair edit engine.
// Guard the migration's refusal and recovery paths, not just manifest hashes.
func TestReviewSkillApplyContract(t *testing.T) {
	root := moduleRoot(t)
	read := func(relative string) string {
		t.Helper()
		b, err := os.ReadFile(filepath.Join(root, relative))
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	adapter := read("construct/local/universal-review/adapters/file-edit.md")
	workbench := read("construct/local/fix/SKILL.md")
	round := read("construct/local/universal-review/round.md")
	for _, tc := range []struct {
		name, body string
		want       []string
		obsolete   []string
	}{
		{"tracked adapter", adapter,
			[]string{"couch --session-sync apply EDITS.json FILE", "couch --session-sync round FILE --json", "edits-invalid", "propose-too-large", "propose-unencodable", "Exit 3", "Exit 6", "next submit"},
			[]string{"couch --session-sync write --all BODYFILE FILE"}},
		{"Pair workbench", workbench,
			[]string{"pair review apply EDITS.json FILE", "docflow round --side agent", "propose-too-large", "operator-lock", "stale", "Body"},
			[]string{"PAIR_REVIEW_HANDOFF_PATH", "PAIR_REVIEW_LANDED_PATH", "[reconcile —", "applied N edits"}},
		{"review round", round,
			[]string{"presentation:", "highlight", "propose", "never type"},
			[]string{"deletions:", "propose-strike"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, phrase := range tc.want {
				if !strings.Contains(tc.body, phrase) {
					t.Errorf("missing %q", phrase)
				}
			}
			for _, phrase := range tc.obsolete {
				if strings.Contains(tc.body, phrase) {
					t.Errorf("retired instruction %q remains", phrase)
				}
			}
		})
	}
}
