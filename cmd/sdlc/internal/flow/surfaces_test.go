package flow

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestParseSurfaces: one pattern per line, blank lines and # comments skipped,
// a malformed pattern an error naming its line.
func TestParseSurfaces(t *testing.T) {
	s, err := ParseSurfaces("# shared surfaces\n\nconstruct/vocabulary/*.cue\npkg/vocab/\n  AGENTS.base.md  \n")
	if err != nil {
		t.Fatal(err)
	}
	for path, want := range map[string]bool{
		"construct/vocabulary/issue.cue":     true,
		"construct/vocabulary/sub/issue.cue": false, // * does not cross a /
		"pkg/vocab/vocab.go":                 true,  // trailing / means anything under
		"pkg/vocab/deep/x.go":                true,
		"pkg/vocabulary/x.go":                false,
		"AGENTS.base.md":                     true,
		"cmd/sdlc/close.go":                  false,
		DeclarationPath:                      true, // always: a branch cannot edit its own shell away
	} {
		if got := s.Match(path); got != want {
			t.Errorf("Match(%q) = %v, want %v", path, got, want)
		}
	}
	if _, err := ParseSurfaces("ok/*.go\nbad/[x\n"); err == nil || !strings.Contains(err.Error(), "line 2") {
		t.Errorf("bad pattern: err = %v, want one naming line 2", err)
	}
}

// TestSurfacesEmptyStillGuardsItself: a repo with no declaration still counts
// an edit to the declaration file as touching a shared surface.
func TestSurfacesEmptyStillGuardsItself(t *testing.T) {
	var s Surfaces
	if !s.Match(DeclarationPath) || s.Match("cmd/x.go") {
		t.Error("the empty declaration must match only itself")
	}
}

// TestSurfacesUnion: close reads the declaration at the window base AND at HEAD
// and takes the union, so a branch that deletes a pattern still sees it (#231 PQ-5).
func TestSurfacesUnion(t *testing.T) {
	base, _ := ParseSurfaces("pkg/vocab/\nconstruct/base.manifest\n")
	head, _ := ParseSurfaces("construct/base.manifest\ncmd/sdlc/internal/judge/architecture.md\n")
	u := Union(base, head)
	for _, p := range []string{"pkg/vocab/x.go", "construct/base.manifest", "cmd/sdlc/internal/judge/architecture.md"} {
		if !u.Match(p) {
			t.Errorf("union does not match %q", p)
		}
	}
}

// FuzzParseSurfaces: the declaration is repo-authored text. Parsing never
// panics, and a declaration that parses never panics when matched.
func FuzzParseSurfaces(f *testing.F) {
	for _, seed := range []string{"pkg/vocab/\n", "a/*.go\n# c\n", "[\n", "\\\n", "a/**/b\n", "/abs/path\n"} {
		f.Add(seed, "pkg/vocab/x.go")
	}
	f.Fuzz(func(t *testing.T, text, path string) {
		s, err := ParseSurfaces(text)
		if err != nil {
			return
		}
		_ = s.Match(path)
	})
}

// TestRepoDeclarationParses: ariadne's own declaration is valid, so a malformed
// edit to it fails here rather than at some close as an "unreadable" crossing.
func TestRepoDeclarationParses(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "..", "..", "..", DeclarationPath))
	if err != nil {
		t.Fatal(err)
	}
	s, err := ParseSurfaces(string(b))
	if err != nil {
		t.Fatal(err)
	}
	if !s.Match("construct/vocabulary/issue.cue") || !s.Match("pkg/vocab/vocab.go") {
		t.Error("the declaration no longer covers the vocabulary")
	}
}

// TestParseSurfacesRejectsGlobInDirectory: `lua/*/` would pass a glob check and
// then, compared as a literal prefix, match nothing — so it is refused.
func TestParseSurfacesRejectsGlobInDirectory(t *testing.T) {
	for _, bad := range []string{"lua/*/\n", "a/[bc]/\n", "x/?/\n", "lua/parley/**\n", "**/x.go\n", "!pkg/x.go\n"} {
		if _, err := ParseSurfaces(bad); err == nil {
			t.Errorf("ParseSurfaces(%q): want an error", bad)
		}
	}
}

// TestRepoDeclarationCoversTheShell: a guard that claims a branch cannot
// loosen its own shell must cover every file that DEFINES the shell. In ariadne
// sdlc is built from the branch's own checkout, so that set is sdlc's real build
// closure — derived here from `go list -deps` within this module (every Go file
// and every //go:embed'ed file of each package) plus go.mod and go.sum — never a
// directory someone picked (#231 BR-19, BR-25).
func TestRepoDeclarationCoversTheShell(t *testing.T) {
	root := filepath.Join("..", "..", "..", "..")
	b, err := os.ReadFile(filepath.Join(root, DeclarationPath))
	if err != nil {
		t.Fatal(err)
	}
	s, err := ParseSurfaces(string(b))
	if err != nil {
		t.Fatal(err)
	}
	const mod = "github.com/xianxu/ariadne"
	cmd := exec.Command("go", "list", "-deps", "-f",
		`{{if and .Module (eq .Module.Path "`+mod+`")}}{{.Dir}}{{range .GoFiles}}|{{.}}{{end}}{{range .EmbedFiles}}|{{.}}{{end}}{{end}}`,
		"./cmd/sdlc")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		// A guard fails when the derivation it rests on fails — a skip would
		// silently switch the guard off (#231 M2 review).
		t.Fatalf("go list -deps ./cmd/sdlc: %v", err)
	}
	absRoot, _ := filepath.Abs(root)
	// Every input the go command reads in module mode, including ones that do
	// not exist yet: a branch that ADDS go.work with a replace, or a vendor/
	// tree, changes the binary without touching a package directory.
	shell := []string{"go.mod", "go.sum", "go.work", "go.work.sum", "vendor/modules.txt"}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		dir, err := filepath.Rel(absRoot, parts[0])
		if err != nil {
			t.Fatal(err)
		}
		for _, f := range parts[1:] {
			shell = append(shell, filepath.ToSlash(filepath.Join(dir, f)))
		}
	}
	if len(shell) < 50 {
		t.Fatalf("derived only %d files from sdlc's build closure — the derivation is broken, not the declaration", len(shell))
	}
	for _, f := range shell {
		if !s.Match(f) {
			t.Errorf("%s is compiled into sdlc but is not a declared shared surface", f)
		}
	}
}

// TestSurfaceForms: every form ParseSurfaces accepts takes effect in Match — a
// table of each form against paths inside, beside and beyond it (#231 BR-26).
func TestSurfaceForms(t *testing.T) {
	for _, c := range []struct {
		pattern string
		match   map[string]bool
	}{
		{"pkg/vocab", map[string]bool{"pkg/vocab": true, "pkg/vocab/x.go": true, "pkg/vocab/a/b.go": true, "pkg/vocabulary/x.go": false, "pkg/x.go": false}},
		{"pkg/vocab/", map[string]bool{"pkg/vocab/x.go": true, "pkg/vocab/a/b.go": true, "pkg/vocabulary/x.go": false}},
		{"AGENTS.base.md", map[string]bool{"AGENTS.base.md": true, "AGENTS.base.md.bak": false, "x/AGENTS.base.md": false}},
		{"construct/vocabulary/*.cue", map[string]bool{"construct/vocabulary/issue.cue": true, "construct/vocabulary/a/b.cue": false, "construct/vocabulary/x.go": false}},
		{"lua/parley/*.lua", map[string]bool{"lua/parley/keys.lua": true, "lua/parley/a/keys.lua": false}},
	} {
		s, err := ParseSurfaces(c.pattern + "\n")
		if err != nil {
			t.Fatalf("%q: %v", c.pattern, err)
		}
		for p, want := range c.match {
			if got := s.Match(p); got != want {
				t.Errorf("%q.Match(%q) = %v, want %v", c.pattern, p, got, want)
			}
		}
	}
}
