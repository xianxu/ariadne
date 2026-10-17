package main

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/pkg/layergraph"
	"github.com/xianxu/ariadne/pkg/vocab"
)

const shapeRow = "| `collaboration-mode` | yes | `discussion`: the default (ariadne#415). |\n"

func proto(rows string) string {
	return "---\ntype: type\n---\n\n# x\n\n## Frontmatter shape\n\n| Field | Required | Notes |\n|---|---|---|\n| `type` | yes | `x` |\n" + rows + "\n## Body skeleton\n\n| `collaboration-mode` | yes | `chat` |\n"
}

func TestCollaborationModeDefault(t *testing.T) {
	if got, err := collaborationModeDefault(proto(shapeRow)); err != nil || got != "discussion" {
		t.Fatalf("got %q, %v", got, err)
	}
	if got, err := collaborationModeDefault(strings.ReplaceAll(proto(shapeRow), "\n", "\r\n")); err != nil || got != "discussion" {
		t.Fatalf("CRLF: got %q, %v", got, err)
	}
	for name, c := range map[string]struct{ text, want string }{
		// A row outside the Frontmatter shape section doesn't count.
		"no row":           {proto(""), "no collaboration-mode row"},
		"not a mode":       {proto("| `collaboration-mode` | yes | `debate` |\n"), "names no default"},
		"no value":         {proto("| `collaboration-mode` | yes | the default |\n"), "names no default"},
		"not at the start": {proto("| `collaboration-mode` | yes | see `discussion` |\n"), "names no default"},
		"no shape at all":  {"---\n---\n# x\n", "no Frontmatter shape"},
	} {
		if _, err := collaborationModeDefault(c.text); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: err %v, want %q", name, err, c.want)
		}
	}
}

func TestLintTypesNamesTheTypeWithoutADefault(t *testing.T) {
	dir := t.TempDir()
	good, bad := filepath.Join(dir, "good.md"), filepath.Join(dir, "bad.md")
	if os.WriteFile(good, []byte(proto(shapeRow)), 0o644) != nil || os.WriteFile(bad, []byte(proto("")), 0o644) != nil {
		t.Fatal("write")
	}
	got := lintTypes(layergraph.OSFS{}, []TypeProto{{Name: "good", BodyPath: good}, {Name: "bad", BodyPath: bad}})
	if len(got) != 1 || !strings.HasPrefix(got[0], "bad ") {
		t.Errorf("problems %q, want only bad", got)
	}
}

// #415: every type this repository's registry resolves declares a default,
// so a new prototype without one fails here (and `datatype lint` in any
// derived repository).
func TestEveryTypeDeclaresACollaborationMode(t *testing.T) {
	protos, err := resolveTypes()
	if err != nil {
		t.Fatal(err)
	}
	if len(protos) < 10 {
		t.Fatalf("%d types resolved, want the 10 construct/datatype prototypes", len(protos))
	}
	for _, p := range lintTypes(layergraph.OSFS{}, protos) {
		t.Error(p)
	}
	// project is also written by `sdlc project new` from its model; the two
	// declarations must agree.
	i := slices.IndexFunc(protos, func(p TypeProto) bool { return p.Name == "project" })
	if i < 0 {
		t.Fatal("the project type did not resolve")
	}
	b, err := os.ReadFile(protos[i].BodyPath)
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := collaborationModeDefault(string(b)); got != vocab.Project().CollaborationMode() {
		t.Errorf("project.md declares %q, project.cue's scaffold %q", got, vocab.Project().CollaborationMode())
	}
}

// #415: a durable plan, written from the writing-plans skill's header,
// declares its default collaboration mode.
func TestPlanHeaderDeclaresACollaborationMode(t *testing.T) {
	root, err := findRepoRoot(".")
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(root, "construct", "adapted", "superpowers-writing-plans", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	m := regexp.MustCompile("```markdown\n---\ncollaboration-mode: ([a-z]+)\n---\n\n# \\[Feature Name\\] Implementation Plan").FindSubmatch(b)
	if m == nil || !slices.Contains(vocab.CollaborationModes, string(m[1])) {
		t.Errorf("the plan header declares no valid collaboration mode")
	}
}
