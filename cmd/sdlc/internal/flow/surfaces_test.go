package flow

import (
	"os"
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
	for _, bad := range []string{"lua/*/\n", "a/[bc]/\n", "x/?/\n"} {
		if _, err := ParseSurfaces(bad); err == nil {
			t.Errorf("ParseSurfaces(%q): want an error", bad)
		}
	}
}

// TestRepoDeclarationCoversTheShell: a guard that claims a branch cannot loosen
// its own shell must cover every file that DEFINES the shell — and the set is
// derived here, not listed (#231 BR-19). In ariadne sdlc is built from the
// branch's own checkout, so the shell's definition is: every production file
// that imports this package, and the packages that decide what it measures and
// how it reviews (flow, churn, judge).
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
	const importPath = `"github.com/xianxu/ariadne/cmd/sdlc/internal/flow"`
	var shell []string
	sdlc := filepath.Join(root, "cmd", "sdlc")
	err = filepath.WalkDir(sdlc, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(p, ".go") || strings.HasSuffix(p, "_test.go") {
			return err
		}
		rel, _ := filepath.Rel(root, p)
		rel = filepath.ToSlash(rel)
		src, rerr := os.ReadFile(p)
		if rerr != nil {
			return rerr
		}
		for _, pkg := range []string{"cmd/sdlc/internal/flow/", "cmd/sdlc/internal/churn/", "cmd/sdlc/internal/judge/"} {
			if strings.HasPrefix(rel, pkg) {
				shell = append(shell, rel)
				return nil
			}
		}
		if strings.Contains(string(src), importPath) {
			shell = append(shell, rel)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(shell) < 10 {
		t.Fatalf("derived only %d shell files — the derivation is broken, not the declaration", len(shell))
	}
	for _, f := range shell {
		if !s.Match(f) {
			t.Errorf("%s defines the quick-flow shell but is not a declared shared surface", f)
		}
	}
}
